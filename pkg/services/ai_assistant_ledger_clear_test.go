package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestAssistantLedgerClearCommitFencesInterleavedRun(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			for _, mode := range []string{"all-transactions", "all-data", "account"} {
				t.Run(mode, func(t *testing.T) {
					s := assistantConversationFixture(t, kind)
					s.storage.now = time.Now
					ctx := core.NewNullContext()
					ledger := &TransactionService{ServiceUsingDB: s.storage.ServiceUsingDB}
					session := s.storage.UserDataDB(7).NewAssistantSession(ctx)
					engine := session.Engine()
					session.Close()
					require.NoError(t, engine.Sync2(&models.Transaction{}, &models.TransactionTagIndex{}, &models.TransactionPictureInfo{}, &models.Account{}))
					_, err := engine.Insert(&models.Transaction{TransactionId: 1, Uid: 7, AccountId: 1, Amount: 123, Type: models.TRANSACTION_DB_TYPE_INCOME, TransactionTime: 100}, &models.Account{AccountId: 1, Uid: 7, Balance: 123, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT})
					require.NoError(t, err)
					// This is the API's early fence. Stop the clear before the actual ledger
					// transaction, then admit and claim a new run in that exact window.
					fenced := make(chan struct{})
					proceed, releaseClear := assistantClearBarrier(t)
					cleared := make(chan error, 1)
					go func() {
						if err := s.storage.InvalidateLedgerContext(ctx, 7); err != nil {
							cleared <- err
							return
						}
						close(fenced)
						<-proceed
						if mode == "account" {
							cleared <- ledger.DeleteAllTransactionsOfAccount(ctx, 7, 1, 1000, s.storage)
						} else {
							cleared <- ledger.DeleteAllTransactions(ctx, 7, mode == "all-data", s.storage)
						}
					}()
					select {
					case <-fenced:
					case <-time.After(3 * time.Second):
						t.Fatal("early fence did not finish")
					}
					accepted := conversationSend(t, s, "in-clear-window", "", 0, "old ledger")
					worker := NewAssistantRunService(s.storage, AssistantRunOptions{GenerationAvailable: func(core.Context, int64) bool { return true }})
					id, err := assistantParseID(accepted.RunId)
					require.NoError(t, err)
					claimed, err := worker.claim(ctx, 7, id)
					require.NoError(t, err)
					require.NotNil(t, claimed)
					modelEntered, finished := make(chan struct{}), make(chan struct{})
					modelRelease, releaseModel := assistantClearBarrier(t)
					worker.options.Prepare = assistantWorkerPrepare(func(core.Context) (data.TurnResponse, error) {
						close(modelEntered)
						<-modelRelease
						return data.TurnResponse{Completed: true, Content: "stale model answer"}, nil
					})
					go func() { worker.execute(context.Background(), claimed); close(finished) }()
					select {
					case <-modelEntered:
					case <-time.After(3 * time.Second):
						t.Fatal("model did not enter")
					}
					releaseClear()
					select {
					case err := <-cleared:
						require.NoError(t, err)
					case <-time.After(3 * time.Second):
						t.Fatal("ledger clear did not finish")
					}
					snapshot, err := worker.GetRunSnapshot(ctx, 7, accepted.RunId)
					require.NoError(t, err)
					require.Equal(t, models.AssistantRunCancelRequested, snapshot.State)
					require.Error(t, worker.checkpoint(ctx, claimed, &assistantRunBuffer{dirty: true, content: "stale answer"}, nil, false))
					reclaimed, err := worker.claim(ctx, 7, id)
					require.NoError(t, err)
					require.Nil(t, reclaimed)
					releaseModel()
					select {
					case <-finished:
					case <-time.After(3 * time.Second):
						t.Fatal("worker did not finish")
					}
					snapshot, err = worker.GetRunSnapshot(ctx, 7, accepted.RunId)
					require.NoError(t, err)
					require.Equal(t, models.AssistantRunCanceled, snapshot.State)
					require.Empty(t, snapshot.Content)
					require.NoError(t, s.storage.WithUserRead(ctx, 7, func(session *xorm.Session) error {
						c, err := assistantConversationLoad(session, 7, claimed.ConversationId)
						if err != nil {
							return err
						}
						require.EqualValues(t, 1, c.ContextInvalidatedThroughTurn)
						return nil
					}))
					remaining, err := engine.Where("uid = ? AND deleted = ?", 7, false).Count(&models.Transaction{})
					require.NoError(t, err)
					require.Zero(t, remaining)
					fresh := conversationSend(t, s, "after-clear", "", 0, "new ledger")
					freshID, err := assistantParseID(fresh.RunId)
					require.NoError(t, err)
					newClaim, err := worker.claim(ctx, 7, freshID)
					require.NoError(t, err)
					require.NotNil(t, newClaim)
				})
			}
		})
	}
}

func TestAssistantLedgerClearRollbackKeepsRun(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s := assistantConversationFixture(t, kind)
			ctx := core.NewNullContext()
			accepted := conversationSend(t, s, "rollback", "", 0, "still valid")
			ledger := &TransactionService{ServiceUsingDB: s.storage.ServiceUsingDB}
			failure := errors.New("ledger mutation failed")
			err := ledger.withLedgerClearTransaction(ctx, 7, newAssistantLedgerClear(s.storage, 7), func(session *xorm.Session) error {
				_, err := session.Where("uid = ?", 7).Cols("title").Update(&models.AssistantConversation{Title: "rolled back"})
				require.NoError(t, err)
				return failure
			})
			require.ErrorIs(t, err, failure)
			worker := NewAssistantRunService(s.storage, AssistantRunOptions{})
			snapshot, err := worker.GetRunSnapshot(ctx, 7, accepted.RunId)
			require.NoError(t, err)
			require.Equal(t, models.AssistantRunQueued, snapshot.State)
			state, err := s.storage.GetUserState(ctx, 7)
			require.NoError(t, err)
			require.Equal(t, 1, state.QueuedCount)
			require.Zero(t, state.MemoryGenerationEpoch)
		})
	}
}

// Cleanup releases both workers even when an assertion fails before a barrier.
func assistantClearBarrier(t *testing.T) (<-chan struct{}, func()) {
	t.Helper()
	barrier := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(barrier) }) }
	t.Cleanup(release)
	return barrier, release
}

func TestAssistantLedgerAccountClearPartialFailure(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s := assistantConversationFixture(t, kind)
			ctx := core.NewNullContext()
			ledger := &TransactionService{ServiceUsingDB: s.storage.ServiceUsingDB}
			session := s.storage.UserDataDB(7).NewAssistantSession(ctx)
			engine := session.Engine()
			session.Close()
			require.NoError(t, engine.Sync2(&models.Transaction{}, &models.TransactionTagIndex{}, &models.TransactionPictureInfo{}, &models.Account{}))
			_, err := engine.Insert(&models.Account{AccountId: 1, Uid: 7, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Balance: 123})
			require.NoError(t, err)
			_, err = engine.Insert(
				&models.Transaction{TransactionId: 1, Uid: 7, AccountId: 1, Amount: 123, Type: models.TRANSACTION_DB_TYPE_INCOME, TransactionTime: 200},
				&models.Transaction{TransactionId: 2, Uid: 7, AccountId: 1, RelatedAccountId: 999, Amount: 50, Type: models.TRANSACTION_DB_TYPE_TRANSFER_OUT, TransactionTime: 100})
			require.NoError(t, err)
			accepted := conversationSend(t, s, "partial", "", 0, "old ledger")
			require.ErrorIs(t, ledger.DeleteAllTransactionsOfAccount(ctx, 7, 1, 1000, s.storage), errs.ErrDestinationAccountNotFound)
			var rows []models.Transaction
			require.NoError(t, engine.Asc("transaction_id").Find(&rows))
			require.Len(t, rows, 2)
			require.True(t, rows[0].Deleted)
			require.False(t, rows[1].Deleted)
			snapshot, err := NewAssistantRunService(s.storage, AssistantRunOptions{}).GetRunSnapshot(ctx, 7, accepted.RunId)
			require.NoError(t, err)
			require.Equal(t, models.AssistantRunCanceled, snapshot.State)
			state, err := s.storage.GetUserState(ctx, 7)
			require.NoError(t, err)
			require.EqualValues(t, 1, state.MemoryGenerationEpoch, "failed second deletion must roll back its fence")
			fresh := conversationSend(t, s, "ordinary-delete", "", 0, "keep existing single-delete policy")
			// Ordinary deletion still uses its original transaction path, including errors.
			require.ErrorIs(t, ledger.DeleteTransaction(ctx, 7, 2), errs.ErrDestinationAccountNotFound)
			snapshot, err = NewAssistantRunService(s.storage, AssistantRunOptions{}).GetRunSnapshot(ctx, 7, fresh.RunId)
			require.NoError(t, err)
			require.Equal(t, models.AssistantRunQueued, snapshot.State)
		})
	}
}

func TestAssistantLedgerRepeatedClearAfterSameTurnRegeneration(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s := assistantConversationFixture(t, kind)
			ctx := core.NewNullContext()
			first := conversationSend(t, s, "first", "", 0, "question")
			conversationComplete(t, s, first, "old answer")
			require.NoError(t, s.storage.InvalidateLedgerContext(ctx, 7))
			cid, err := assistantParseID(first.ConversationId)
			require.NoError(t, err)
			var before *models.AssistantConversation
			require.NoError(t, s.storage.WithUserRead(ctx, 7, func(session *xorm.Session) error {
				before, err = assistantConversationLoad(session, 7, cid)
				return err
			}))
			regenerated, err := s.Retry(ctx, 7, models.AssistantRetryRequest{
				AssistantCommand: storageCommand(s.storage, "regenerate", 0), ConversationId: first.ConversationId,
				UserMessageId: first.UserMessageId, ExpectedContextVersion: before.ContextVersion, Intent: "regenerate",
			})
			require.NoError(t, err)
			conversationComplete(t, s, regenerated, "answer from remaining ledger")
			ledger := &TransactionService{ServiceUsingDB: s.storage.ServiceUsingDB}
			require.NoError(t, ledger.withLedgerClearTransaction(ctx, 7, newAssistantLedgerClear(s.storage, 7), func(*xorm.Session) error { return nil }))
			require.NoError(t, s.storage.WithUserRead(ctx, 7, func(session *xorm.Session) error {
				after, err := assistantConversationLoad(session, 7, cid)
				if err != nil {
					return err
				}
				require.EqualValues(t, 2, after.NextTurnSequence)
				require.EqualValues(t, 1, after.ContextInvalidatedThroughTurn)
				require.Greater(t, after.SourceVersion, before.SourceVersion)
				require.Greater(t, after.ContextVersion, regenerated.ContextVersion)
				return nil
			}))
			fresh := conversationSend(t, s, "new-run-before-second-clear", "", 0, "new question")
			require.NoError(t, ledger.withLedgerClearTransaction(ctx, 7, newAssistantLedgerClear(s.storage, 7), func(*xorm.Session) error { return nil }))
			snapshot, err := NewAssistantRunService(s.storage, AssistantRunOptions{}).GetRunSnapshot(ctx, 7, fresh.RunId)
			require.NoError(t, err)
			require.Equal(t, models.AssistantRunCanceled, snapshot.State)
		})
	}
}
