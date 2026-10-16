package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"xorm.io/xorm"
)

// assistantLedgerClear owns one sequential bulk-clear operation. Its proof is
// local to this operation and user, never shared between requests or goroutines.
// Repeated deletions can reuse a fence only while the locked database version
// proves that no assistant writer has committed since this operation's last
// successful deletion. Reads and ordinary ledger edits cannot create context.
type assistantLedgerClear struct {
	storage          *AssistantStorageService
	uid              int64
	committedVersion int64
}

func newAssistantLedgerClear(storage *AssistantStorageService, uid int64) *assistantLedgerClear {
	if storage == nil {
		return nil
	}
	return &assistantLedgerClear{storage: storage, uid: uid}
}

// apply commits ledger changes and any required context invalidation together.
// Its callback must contain only transactional work, because CAS may retry it.
func (c *assistantLedgerClear) apply(ctx core.Context, mutation func(*xorm.Session) error) error {
	var mutationErr error
	version, err := c.storage.withUserTransactionVersion(ctx, c.uid, func(tx *AssistantStorageTx) error {
		mutationErr = nil
		if c.committedVersion == 0 || tx.previousVersion != c.committedVersion {
			if err := tx.invalidateLedgerContext(); err != nil {
				return err
			}
		}
		mutationErr = mutation(tx.Session)
		return mutationErr
	})
	if err != nil {
		// Preserve ledger validation errors; an unsuccessful transaction cannot
		// advance the proof, including when the final CAS or Commit fails.
		if mutationErr != nil {
			return mutationErr
		}
		return err
	}
	c.committedVersion = version
	return nil
}
