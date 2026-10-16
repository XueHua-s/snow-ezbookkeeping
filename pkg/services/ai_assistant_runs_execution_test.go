package services

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
)

type assistantWorkerModel func(core.Context) (data.TurnResponse, error)

func (f assistantWorkerModel) RunToolTurn(ctx core.Context, _ int64, _ data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
	return f(ctx)
}

type assistantWorkerTools struct{}

func (assistantWorkerTools) Definitions() []data.ToolDefinition { return nil }
func (assistantWorkerTools) Execute(core.Context, data.ToolCall) data.ToolResult {
	return data.ToolResult{}
}
func assistantWorkerPrepare(model assistantWorkerModel) AssistantRunPreparer {
	return func(ctx core.Context, run *models.AssistantMessage) (*AssistantPreparedRun, error) {
		return &AssistantPreparedRun{Runner: llm.NewAssistantRunner(model, nil, assistantWorkerTools{}, time.Now), Input: llm.AssistantRunInput{UID: run.Uid, MaxRounds: 1}, Dependencies: models.AssistantInputDependencies{DeletionEpoch: run.DeletionEpoch}}, nil
	}
}
func TestAssistantWorkerRealRunnerOutputAndEmpty(t *testing.T) {
	for _, test := range []struct {
		name, reply string
		state       models.AssistantRunState
		reason      models.AssistantReasonCode
	}{{"complete", "original **Markdown**\n", models.AssistantRunCompleted, ""}, {"empty", " \n", models.AssistantRunFailed, models.AssistantReasonEmptyResponse}, {"limit", strings.Repeat("中", 12000), models.AssistantRunIncomplete, models.AssistantReasonOutputLimit}} {
		t.Run(test.name, func(t *testing.T) {
			s, _ := assistantRunFixture(t, settings.Sqlite3DbType)
			var calls atomic.Int64
			s.options.Prepare = assistantWorkerPrepare(func(core.Context) (data.TurnResponse, error) {
				calls.Add(1)
				return data.TurnResponse{Completed: true, Content: test.reply}, nil
			})
			run, err := s.claim(core.NewNullContext(), 7, 3)
			require.NoError(t, err)
			s.execute(context.Background(), run)
			snap, err := s.GetRunSnapshot(core.NewNullContext(), 7, "3")
			require.NoError(t, err)
			require.Equal(t, test.state, snap.State)
			if test.reason != "" {
				require.NotNil(t, snap.ReasonCode)
				require.Equal(t, test.reason, *snap.ReasonCode)
			}
			if test.name == "limit" {
				require.NotEmpty(t, snap.Content)
				require.LessOrEqual(t, len(snap.Content), 32768)
			}
			again, err := s.claim(core.NewNullContext(), 7, 3)
			require.NoError(t, err)
			require.Nil(t, again)
			require.EqualValues(t, 1, calls.Load())
		})
	}
}
func TestAssistantWorkerCancelAbortsProviderAndCannotComplete(t *testing.T) {
	s, _ := assistantRunFixture(t, settings.Sqlite3DbType)
	entered := make(chan struct{})
	stopped := make(chan struct{})
	s.options.Prepare = assistantWorkerPrepare(func(ctx core.Context) (data.TurnResponse, error) {
		close(entered)
		<-ctx.Done()
		close(stopped)
		return data.TurnResponse{Completed: true, Content: "late success"}, nil
	})
	run, err := s.claim(core.NewNullContext(), 7, 3)
	require.NoError(t, err)
	done := make(chan struct{})
	go func() { s.execute(context.Background(), run); close(done) }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("provider not entered")
	}
	_, err = s.Cancel(core.NewNullContext(), 7, models.AssistantCancelRequest{AssistantCommand: storageCommand(s.storage, "cancel-running", 0), RunId: "3"})
	require.NoError(t, err)
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("provider not canceled")
	}
	<-done
	snap, err := s.GetRunSnapshot(core.NewNullContext(), 7, "3")
	require.NoError(t, err)
	require.Equal(t, models.AssistantRunCanceled, snap.State)
	require.Empty(t, snap.Content)
}
func TestAssistantWorkerIdlePollDoesNotWriteSnapshots(t *testing.T) {
	s, _ := assistantRunFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	run, err := s.claim(ctx, 7, 3)
	require.NoError(t, err)
	before, err := s.storage.GetUserState(ctx, 7)
	require.NoError(t, err)
	require.NoError(t, s.checkpoint(ctx, run, &assistantRunBuffer{}, nil, false))
	after, err := s.storage.GetUserState(ctx, 7)
	require.NoError(t, err)
	require.Equal(t, before.Version, after.Version)
	snap, err := s.GetRunSnapshot(ctx, 7, "3")
	require.NoError(t, err)
	require.Equal(t, run.SnapshotRevision, snap.Revision)
}

func TestAssistantWorkerStorageFailureStopsProviderWithoutFalseCompletion(t *testing.T) {
	s, _ := assistantRunFixture(t, settings.Sqlite3DbType)
	entered, stopped := make(chan struct{}), make(chan struct{})
	s.options.Prepare = assistantWorkerPrepare(func(ctx core.Context) (data.TurnResponse, error) {
		close(entered)
		<-ctx.Done()
		close(stopped)
		return data.TurnResponse{Completed: true, Content: "not saved"}, nil
	})
	run, err := s.claim(core.NewNullContext(), 7, 3)
	require.NoError(t, err)
	done := make(chan struct{})
	go func() { s.execute(context.Background(), run); close(done) }()
	<-entered
	lock := s.storage.UserDataDB(7).NewAssistantSession(core.NewNullContext())
	defer lock.Close()
	require.NoError(t, lock.Begin())
	_, err = lock.Where("uid = ?", 7).Cols("updated_at").Update(&models.AssistantUserState{UpdatedAt: 42})
	require.NoError(t, err)
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		_ = lock.Rollback()
		t.Fatal("storage failure did not stop provider")
	}
	require.NoError(t, lock.Rollback())
	<-done
	snapshot, err := s.GetRunSnapshot(core.NewNullContext(), 7, "3")
	require.NoError(t, err)
	require.NotEqual(t, models.AssistantRunCompleted, snapshot.State)
	require.Empty(t, snapshot.Content)
}
func TestAssistantWorkerApplicationStopCancelsAndSettles(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s, _ := assistantRunFixture(t, kind)
			entered := make(chan struct{})
			s.options.Prepare = assistantWorkerPrepare(func(ctx core.Context) (data.TurnResponse, error) {
				close(entered)
				<-ctx.Done()
				return data.TurnResponse{}, ctx.Err()
			})
			app, cancel := context.WithCancel(context.Background())
			s.Start(app)
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				cancel()
				t.Fatal("worker did not start")
			}
			cancel()
			wait, stop := context.WithTimeout(context.Background(), 3*time.Second)
			defer stop()
			require.NoError(t, s.Stop(wait))
			snapshot, err := s.GetRunSnapshot(core.NewNullContext(), 7, "3")
			require.NoError(t, err)
			require.Equal(t, models.AssistantRunInterrupted, snapshot.State)
			state, err := s.storage.GetUserState(core.NewNullContext(), 7)
			require.NoError(t, err)
			require.Zero(t, state.RunningCount)
			require.LessOrEqual(t, state.ReservedBytes, int64(settings.DefaultAIAssistantExperienceConfig().Limits.ConversationHeaderBytes))
		})
	}
}

func TestAssistantWorkerProviderOutputLimitPersistsPartial(t *testing.T) {
	s, _ := assistantRunFixture(t, settings.Sqlite3DbType)
	s.options.Prepare = assistantWorkerPrepare(func(core.Context) (data.TurnResponse, error) {
		return data.TurnResponse{Content: "partial provider reply", OutputLimited: true}, &data.ToolError{Code: data.ToolErrorBudgetExhausted}
	})
	run, err := s.claim(core.NewNullContext(), 7, 3)
	require.NoError(t, err)
	s.execute(context.Background(), run)
	snapshot, err := s.GetRunSnapshot(core.NewNullContext(), 7, "3")
	require.NoError(t, err)
	require.Equal(t, models.AssistantRunIncomplete, snapshot.State)
	require.Equal(t, "partial provider reply", snapshot.Content)
	require.NotNil(t, snapshot.ReasonCode)
	require.Equal(t, models.AssistantReasonOutputLimit, *snapshot.ReasonCode)
}
