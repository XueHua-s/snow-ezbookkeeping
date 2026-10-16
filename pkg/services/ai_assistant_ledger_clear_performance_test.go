package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
)

// Measure the real account-clear path and assert its scan complexity. Wall time
// is reported, not gated: scheduler noise must not make correctness tests flaky.
func TestAssistantLedgerClearScaling(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			for _, conversations := range []int{0, 1000} {
				t.Run(fmt.Sprint(conversations), func(t *testing.T) {
					storage, engine := assistantStorageFixture(t, kind)
					require.NoError(t, engine.Sync2(&models.Transaction{}, &models.TransactionTagIndex{}, &models.TransactionPictureInfo{}, &models.Account{}))
					_, err := engine.Insert(&models.AssistantUserState{Uid: 7, Version: 1, ConversationCount: conversations, ReservedBytes: int64(conversations * storage.config().Limits.ConversationHeaderBytes)}, &models.Account{Uid: 7, AccountId: 1, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Balance: 25})
					require.NoError(t, err)
					for offset := 0; offset < conversations; offset += 50 {
						rows := make([]models.AssistantConversation, 50)
						for i := range rows {
							rows[i] = models.AssistantConversation{Uid: 7, Id: int64(offset + i + 1), Title: "Synthetic performance fixture", NextTurnSequence: 1, ContextVersion: 1, MetadataVersion: 1, SourceVersion: 1, HeaderReservedBytes: int64(storage.config().Limits.ConversationHeaderBytes), ReservedBytes: int64(storage.config().Limits.ConversationHeaderBytes)}
						}
						_, err := engine.Insert(rows)
						require.NoError(t, err)
					}
					for id := int64(1); id <= 25; id++ {
						_, err := engine.Insert(&models.Transaction{TransactionId: id, Uid: 7, AccountId: 1, Amount: 1, Type: models.TRANSACTION_DB_TYPE_INCOME, TransactionTime: id})
						require.NoError(t, err)
					}
					ledger := &TransactionService{ServiceUsingDB: storage.ServiceUsingDB}
					scans := &assistantLedgerScanCounter{}
					engine.AddHook(scans)
					start := time.Now()
					require.NoError(t, ledger.DeleteAllTransactionsOfAccount(core.NewNullContext(), 7, 1, 1000, storage))
					t.Logf("conversations=%d transactions=25 elapsed=%s scans=%d", conversations, time.Since(start), scans.count.Load())
					require.EqualValues(t, conversations/50+1, scans.count.Load(), "scan all conversations exactly once, including terminal page")
					remaining, err := engine.Where("uid = ? AND deleted = ?", 7, false).Count(&models.Transaction{})
					require.NoError(t, err)
					require.Zero(t, remaining)
				})
			}

		})
	}
}
