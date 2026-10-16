package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/contexts"
)

func assistantLedgerReuseFixture(t *testing.T, kind string, count int) (*AssistantConversationService, *TransactionService, *xorm.Engine) {
	t.Helper()
	s := assistantConversationFixture(t, kind)
	session := s.storage.UserDataDB(7).NewAssistantSession(core.NewNullContext())
	engine := session.Engine()
	session.Close()
	require.NoError(t, engine.Sync2(&models.Transaction{}, &models.TransactionTagIndex{}, &models.TransactionPictureInfo{}, &models.Account{}))
	_, err := engine.Insert(&models.Account{Uid: 7, AccountId: 1, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Balance: int64(count)})
	require.NoError(t, err)
	for id := int64(1); id <= int64(count); id++ {
		_, err = engine.Insert(&models.Transaction{TransactionId: id, Uid: 7, AccountId: 1, Amount: 1, Type: models.TRANSACTION_DB_TYPE_INCOME, TransactionTime: id})
		require.NoError(t, err)
	}
	return s, &TransactionService{ServiceUsingDB: s.storage.ServiceUsingDB}, engine
}

func assistantLedgerConversation(t *testing.T, s *AssistantConversationService, id string) *models.AssistantConversation {
	t.Helper()
	cid, err := assistantParseID(id)
	require.NoError(t, err)
	var c *models.AssistantConversation
	require.NoError(t, s.storage.WithUserRead(core.NewNullContext(), 7, func(session *xorm.Session) error {
		c, err = assistantConversationLoad(session, 7, cid)
		return err
	}))
	return c
}

func TestAssistantLedgerClearReuseDatabaseContracts(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			t.Run("scan_count_independent_of_deletions", func(t *testing.T) {
				for _, count := range []int{1, 25} {
					t.Run(fmt.Sprint(count), func(t *testing.T) {
						s, ledger, engine := assistantLedgerReuseFixture(t, kind, count)
						first := conversationSend(t, s, "history", "", 0, "question")
						conversationComplete(t, s, first, "answer")
						scans := &assistantLedgerScanCounter{}
						engine.AddHook(scans)
						require.NoError(t, ledger.DeleteAllTransactionsOfAccount(core.NewNullContext(), 7, 1, 1000, s.storage))
						require.EqualValues(t, 2, scans.count.Load(), "one populated page and one terminal page, independent of ledger size")
						c := assistantLedgerConversation(t, s, first.ConversationId)
						require.EqualValues(t, 2, c.SourceVersion)
						remaining, err := engine.Where("uid = ? AND deleted = ?", 7, false).Count(&models.Transaction{})
						require.NoError(t, err)
						require.Zero(t, remaining)
					})
				}
			})
			t.Run("new_run_between_deletions", func(t *testing.T) {
				s, ledger, _ := assistantLedgerReuseFixture(t, kind, 2)
				clear := newAssistantLedgerClear(s.storage, 7)
				ctx := core.NewNullContext()
				require.NoError(t, ledger.deleteTransaction(ctx, 7, 1, clear))
				fresh := conversationSend(t, s, "interleaved", "", 0, "remaining ledger")
				require.NoError(t, ledger.deleteTransaction(ctx, 7, 2, clear))
				snapshot, err := NewAssistantRunService(s.storage, AssistantRunOptions{}).GetRunSnapshot(ctx, 7, fresh.RunId)
				require.NoError(t, err)
				require.Equal(t, models.AssistantRunCanceled, snapshot.State)
				require.EqualValues(t, 1, assistantLedgerConversation(t, s, fresh.ConversationId).ContextInvalidatedThroughTurn)
			})
			t.Run("completed_same_turn_regeneration", func(t *testing.T) {
				s, ledger, _ := assistantLedgerReuseFixture(t, kind, 2)
				first := conversationSend(t, s, "history", "", 0, "question")
				conversationComplete(t, s, first, "old answer")
				clear := newAssistantLedgerClear(s.storage, 7)
				ctx := core.NewNullContext()
				require.NoError(t, ledger.deleteTransaction(ctx, 7, 1, clear))
				before := assistantLedgerConversation(t, s, first.ConversationId)
				regenerated, err := s.Retry(ctx, 7, models.AssistantRetryRequest{
					AssistantCommand: storageCommand(s.storage, "regenerate", 0), ConversationId: first.ConversationId,
					UserMessageId: first.UserMessageId, ExpectedContextVersion: before.ContextVersion, Intent: "regenerate",
				})
				require.NoError(t, err)
				conversationComplete(t, s, regenerated, "remaining ledger answer")
				require.NoError(t, ledger.deleteTransaction(ctx, 7, 2, clear))
				after := assistantLedgerConversation(t, s, first.ConversationId)
				require.Equal(t, before.NextTurnSequence, after.NextTurnSequence)
				require.Greater(t, after.SourceVersion, before.SourceVersion)
				require.Greater(t, after.ContextVersion, regenerated.ContextVersion)
			})

			t.Run("partial_failure_keeps_prior_commit", func(t *testing.T) {
				s, ledger, engine := assistantLedgerReuseFixture(t, kind, 2)
				_, err := engine.Where("uid = ? AND transaction_id = ?", 7, 1).Cols("type", "related_account_id").Update(&models.Transaction{Type: models.TRANSACTION_DB_TYPE_TRANSFER_OUT, RelatedAccountId: 999})
				require.NoError(t, err)
				first := conversationSend(t, s, "partial", "", 0, "question")
				require.ErrorIs(t, ledger.DeleteAllTransactionsOfAccount(core.NewNullContext(), 7, 1, 1000, s.storage), errs.ErrDestinationAccountNotFound)
				var rows []models.Transaction
				require.NoError(t, engine.Asc("transaction_id").Find(&rows))
				require.Len(t, rows, 2)
				require.False(t, rows[0].Deleted)
				require.True(t, rows[1].Deleted)
				snapshot, err := NewAssistantRunService(s.storage, AssistantRunOptions{}).GetRunSnapshot(core.NewNullContext(), 7, first.RunId)
				require.NoError(t, err)
				require.Equal(t, models.AssistantRunCanceled, snapshot.State)
			})
			t.Run("rollback_and_final_cas_retry", func(t *testing.T) {
				s, ledger, engine := assistantLedgerReuseFixture(t, kind, 2)
				ctx := core.NewNullContext()
				clear := newAssistantLedgerClear(s.storage, 7)
				require.NoError(t, ledger.deleteTransaction(ctx, 7, 1, clear))
				proof := clear.committedVersion
				fresh := conversationSend(t, s, "interleaved", "", 0, "remaining ledger")
				failure := errors.New("mutation rollback")
				require.ErrorIs(t, clear.apply(ctx, func(*xorm.Session) error { return failure }), failure)
				require.Equal(t, proof, clear.committedVersion)
				snapshot, err := NewAssistantRunService(s.storage, AssistantRunOptions{}).GetRunSnapshot(ctx, 7, fresh.RunId)
				require.NoError(t, err)
				require.Equal(t, models.AssistantRunQueued, snapshot.State)
				attempts := 0
				require.NoError(t, clear.apply(ctx, func(session *xorm.Session) error {
					attempts++
					var run models.AssistantMessage
					runID, err := assistantParseID(fresh.RunId)
					require.NoError(t, err)
					found, err := session.Where("uid = ? AND id = ?", 7, runID).Get(&run)
					require.NoError(t, err)
					require.True(t, found)
					require.Equal(t, models.AssistantRunCanceled, run.RunState, "every retried attempt must reapply its rolled-back fence")
					require.Equal(t, proof, clear.committedVersion, "callbacks and retried attempts cannot publish a proof")
					var state models.AssistantUserState
					found, err = session.Where("uid = ?", 7).Get(&state)
					require.NoError(t, err)
					require.True(t, found)
					require.EqualValues(t, 1, state.MemoryGenerationEpoch, "fence epoch is committed with State only at transaction end")
					// Force the final user-state CAS to fail after the callback returns. Both
					// the ledger mutation and cancellation must roll back before its retry.
					if attempts == 1 {
						_, err = session.Where("uid = ?", 7).Incr("version", 1).Update(&models.AssistantUserState{})
						require.NoError(t, err)
					}
					_, err = session.Where("uid = ? AND transaction_id = ?", 7, 2).Cols("deleted").Update(&models.Transaction{Deleted: true})
					return err
				}))
				require.Equal(t, 2, attempts)
				state, err := s.storage.GetUserState(ctx, 7)
				require.NoError(t, err)
				require.Equal(t, state.Version, clear.committedVersion)
				require.EqualValues(t, 2, state.MemoryGenerationEpoch)
				snapshot, err = NewAssistantRunService(s.storage, AssistantRunOptions{}).GetRunSnapshot(ctx, 7, fresh.RunId)
				require.NoError(t, err)
				require.Equal(t, models.AssistantRunCanceled, snapshot.State)
				remaining, err := engine.Where("uid = ? AND deleted = ?", 7, false).Count(&models.Transaction{})
				require.NoError(t, err)
				require.Zero(t, remaining)
			})
		})
	}
}

// Count statements without logging financial SQL or enabling sensitive logs.
type assistantLedgerScanCounter struct{ count atomic.Int64 }

func (c *assistantLedgerScanCounter) BeforeProcess(h *contexts.ContextHook) (context.Context, error) {
	sql := strings.NewReplacer("`", "", "\"", "").Replace(h.SQL)
	if strings.Contains(sql, "SELECT") && strings.Contains(sql, "assistant_conversation") && strings.Contains(sql, "id >") {
		c.count.Add(1)
	}
	return h.Ctx, nil
}
func (*assistantLedgerScanCounter) AfterProcess(*contexts.ContextHook) error { return nil }
