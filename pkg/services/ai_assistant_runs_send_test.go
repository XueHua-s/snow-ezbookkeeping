package services

import (
	"context"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// Exercise the real acceptance transaction and worker settlement together. The
// fixture does not hand-build reservations, so accounting changes cannot hide
// integration regressions behind a worker-only mock.
func TestAssistantRunSendSettlementDatabaseContracts(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			for _, outcome := range []string{"completed", "failed", "canceled"} {
				t.Run(outcome, func(t *testing.T) {
					conversations := assistantConversationFixture(t, kind)
					conversations.storage.now = time.Now
					conversations.now = time.Now
					worker := NewAssistantRunService(conversations.storage, AssistantRunOptions{GenerationAvailable: func(core.Context, int64) bool { return true }})
					accepted := conversationSend(t, conversations, "real-send", "", 0, "保留原始提问\n")
					worker.options.Prepare = assistantWorkerPrepare(func(ctx core.Context) (data.TurnResponse, error) {
						if outcome == "failed" {
							return data.TurnResponse{}, &data.ToolError{Code: data.ToolErrorUpstreamFailed}
						}
						if outcome == "canceled" {
							_, err := worker.Cancel(ctx, 7, models.AssistantCancelRequest{AssistantCommand: storageCommand(conversations.storage, "cancel-before-completion", 0), RunId: accepted.RunId})
							if err != nil {
								return data.TurnResponse{}, err
							}
						}
						return data.TurnResponse{Completed: true, Content: "原始 **Markdown**\n"}, nil
					})
					runID, err := assistantParseID(accepted.RunId)
					require.NoError(t, err)
					claimed, err := worker.claim(core.NewNullContext(), 7, runID)
					require.NoError(t, err)
					require.NotNil(t, claimed)
					worker.execute(context.Background(), claimed)
					snapshot, err := worker.GetRunSnapshot(core.NewNullContext(), 7, accepted.RunId)
					require.NoError(t, err)
					require.Equal(t, models.AssistantRunState(outcome), snapshot.State)
					state, err := worker.storage.GetUserState(core.NewNullContext(), 7)
					require.NoError(t, err)
					require.Zero(t, state.RunningCount)
					require.Zero(t, state.QueuedCount)
					require.NoError(t, worker.storage.WithUserRead(core.NewNullContext(), 7, func(session *xorm.Session) error {
						run, err := assistantConversationMessage(session, 7, runID)
						if err != nil {
							return err
						}
						conversation, err := assistantConversationLoad(session, 7, run.ConversationId)
						if err != nil {
							return err
						}
						require.True(t, run.ReservationSettled)
						require.Zero(t, run.ReservedBytes)
						require.Zero(t, conversation.ActiveRunId)
						require.Equal(t, conversation.HeaderReservedBytes, state.ReservedBytes)
						require.Equal(t, conversation.HeaderReservedBytes, conversation.ReservedBytes)
						require.Equal(t, conversation.HeaderUsedBytes+run.UsedBytes, state.UsedBytes)
						require.Equal(t, conversation.UsedBytes, state.UsedBytes)
						return nil
					}))
				})
			}
		})
	}
}
