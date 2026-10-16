package services

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
)

func assistantRunFixture(t *testing.T, kind string) (*AssistantRunService, *models.AssistantMessage) {
	storage, _ := assistantStorageFixture(t, kind)
	storage.now = time.Now
	service := NewAssistantRunService(storage, AssistantRunOptions{GenerationAvailable: func(core.Context, int64) bool { return true }})
	run := &models.AssistantMessage{Id: 3, Uid: 7, ConversationId: 1, UserMessageId: 2, Role: "assistant", AttemptNumber: 1, RunState: models.AssistantRunQueued, CreatedAt: time.Now().UnixMilli(), SnapshotRevision: 1, SchemaVersion: 1}
	require.NoError(t, storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
		conversation := &models.AssistantConversation{Id: 1, Uid: 7, ActiveRunId: 3}
		question := &models.AssistantMessage{Id: 2, Uid: 7, ConversationId: 1, UserMessageId: 2, Role: "user", Content: "question"}
		limits := settings.DefaultAIAssistantExperienceConfig().Limits
		if err := tx.ReserveConversation(conversation, &limits); err != nil {
			return err
		}
		if err := assistantSettleConversationHeader(tx, conversation); err != nil {
			return err
		}
		if err := tx.ReserveRun(conversation, run, &limits); err != nil {
			return err
		}
		_, err := tx.Session.Insert(conversation, question, run)
		return err
	}))
	return service, run
}
func TestAssistantRunDatabaseFencingAndRecovery(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s, queued := assistantRunFixture(t, kind)
			ctx := core.NewNullContext()
			second := NewAssistantRunService(s.storage, s.options)
			var wg sync.WaitGroup
			results := make(chan *models.AssistantMessage, 2)
			for _, worker := range []*AssistantRunService{s, second} {
				wg.Add(1)
				go func() {
					defer wg.Done()
					run, err := worker.claim(ctx, 7, queued.Id)
					require.NoError(t, err)
					results <- run
				}()
			}
			wg.Wait()
			close(results)
			var claimed *models.AssistantMessage
			count := 0
			for run := range results {
				if run != nil {
					claimed = run
					count++
				}
			}
			require.Equal(t, 1, count)
			owner := s
			if claimed.Owner == second.owner {
				owner = second
			}
			fake := *claimed
			fake.LeaseEpoch--
			err := owner.checkpoint(ctx, &fake, &assistantRunBuffer{content: "stale", dirty: true}, nil, false)
			require.ErrorIs(t, err, models.AssistantReasonInterrupted)
			require.NoError(t, owner.checkpoint(ctx, claimed, &assistantRunBuffer{content: "saved fragment", dirty: true}, &models.AssistantInputDependencies{}, true))
			cancel := models.AssistantCancelRequest{AssistantCommand: storageCommand(s.storage, "stop", 0), RunId: "3"}
			_, err = second.Cancel(ctx, 7, cancel)
			require.NoError(t, err)
			require.Error(t, owner.checkpoint(ctx, claimed, &assistantRunBuffer{content: "late completion", dirty: true}, nil, false))
			require.NoError(t, s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
				run, e := assistantConversationMessage(tx.Session, 7, 3)
				if e != nil {
					return e
				}
				return owner.finish(tx, run, models.AssistantRunCanceled, "")
			}))
			snapshot, err := s.GetRunSnapshot(ctx, 7, "3")
			require.NoError(t, err)
			require.Equal(t, "saved fragment", snapshot.Content)
			require.Equal(t, models.AssistantRunCanceled, snapshot.State)
			_, err = s.GetRunSnapshot(ctx, 8, "3")
			require.ErrorIs(t, err, models.AssistantReasonNotAccessible)
			_, err = second.Cancel(ctx, 7, cancel)
			require.NoError(t, err)
			state, err := s.storage.GetUserState(ctx, 7)
			require.NoError(t, err)
			require.LessOrEqual(t, state.ReservedBytes, int64(settings.DefaultAIAssistantExperienceConfig().Limits.ConversationHeaderBytes))
			require.Zero(t, state.RunningCount)
		})
	}
}
func TestAssistantRunExpiredClaimNeverReplays(t *testing.T) {
	s, run := assistantRunFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	claimed, err := s.claim(ctx, 7, run.Id)
	require.NoError(t, err)
	require.NotNil(t, claimed)
	s.storage.now = func() time.Time { return time.UnixMilli(claimed.LeaseExpiresAt + 1) }
	again, err := s.claim(ctx, 7, run.Id)
	require.NoError(t, err)
	require.Nil(t, again)
	snapshot, err := s.GetRunSnapshot(ctx, 7, "3")
	require.NoError(t, err)
	require.Equal(t, models.AssistantRunInterrupted, snapshot.State)
}
func TestAssistantRunQueueExpiresAndDependencyRevocation(t *testing.T) {
	t.Run("queue", func(t *testing.T) {
		s, run := assistantRunFixture(t, settings.Sqlite3DbType)
		s.storage.now = func() time.Time { return time.UnixMilli(run.CreatedAt).Add(time.Minute) }
		claimed, err := s.claim(core.NewNullContext(), 7, 3)
		require.NoError(t, err)
		require.Nil(t, claimed)
		snap, err := s.GetRunSnapshot(core.NewNullContext(), 7, "3")
		require.NoError(t, err)
		require.Equal(t, models.AssistantRunFailed, snap.State)
	})
	t.Run("epoch", func(t *testing.T) {
		s, _ := assistantRunFixture(t, settings.Sqlite3DbType)
		ctx := core.NewNullContext()
		claimed, err := s.claim(ctx, 7, 3)
		require.NoError(t, err)
		require.NoError(t, s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error { tx.State.DeletionEpoch++; return nil }))
		require.ErrorIs(t, s.checkpoint(ctx, claimed, &assistantRunBuffer{}, nil, false), models.AssistantReasonContextChanged)
	})
}

func TestAssistantRunAcceptancePolicySurvivesConfigReduction(t *testing.T) {
	s, run := assistantRunFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	accepted := NewAssistantRunPolicy(s.options.Runtime(), s.options.Config())
	raw, err := json.Marshal(accepted)
	require.NoError(t, err)
	require.NoError(t, s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		saved, err := assistantConversationMessage(tx.Session, 7, run.Id)
		if err != nil {
			return err
		}
		saved.RuntimeJSON = string(raw)
		return assistantRunSave(tx, saved)
	}))
	s.storage.now = func() time.Time { return time.UnixMilli(run.CreatedAt).Add(2 * time.Second) }
	s.options.Config = func() settings.AIAssistantExperienceConfig {
		config := settings.DefaultAIAssistantExperienceConfig()
		config.QueueTimeoutSeconds = 1
		config.LeaseSeconds = 2
		config.LeaseHeartbeatSeconds = 1
		return config
	}
	s.options.Runtime = func() settings.AIAssistantRuntimeConfig { return settings.AIAssistantRuntimeConfig{RunTimeout: 1000} }
	claimed, err := s.claim(ctx, 7, 3)
	require.NoError(t, err)
	require.NotNil(t, claimed)
	require.Equal(t, int64(accepted.RunTimeout), claimed.Deadline-s.storage.now().UnixMilli())
	require.Equal(t, int64(accepted.Experience.LeaseSeconds)*1000, claimed.LeaseExpiresAt-s.storage.now().UnixMilli())
	before, err := s.GetRunSnapshot(ctx, 7, "3")
	require.NoError(t, err)
	s.storage.now = func() time.Time { return time.UnixMilli(run.CreatedAt).Add(8 * time.Second) }
	require.NoError(t, s.checkpoint(ctx, claimed, &assistantRunBuffer{}, nil, false))
	after, err := s.GetRunSnapshot(ctx, 7, "3")
	require.NoError(t, err)
	require.Equal(t, before.Revision, after.Revision)
	require.Equal(t, before.SavedAt, after.SavedAt)
}
