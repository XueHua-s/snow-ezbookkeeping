package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"xorm.io/xorm"
)

// Bulk-clear operations supply a guard; ordinary transaction deletion
// keeps its existing behavior. Both mutations use the same user-shard session.
// No external side effects may run in apply, because CAS retries it on rollback.
func (s *TransactionService) withLedgerClearTransaction(ctx core.Context, uid int64, clear *assistantLedgerClear, apply func(*xorm.Session) error) error {
	if clear == nil {
		return s.UserDataDB(uid).DoTransaction(ctx, apply)
	}
	return clear.apply(ctx, apply)
}
