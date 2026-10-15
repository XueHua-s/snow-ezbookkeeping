package datastore

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"xorm.io/xorm"
)

// NewAssistantSession reads the authoritative primary and suppresses sensitive
// SQL parameters. The caller owns Close and any short transaction lifecycle.
func (db *Database) NewAssistantSession(ctx core.Context) *xorm.Session {
	return db.engineGroup.Master().NewSession().Context(NewXOrmContextAdapter(ctx)).MustLogSQL(false).NoCache()
}
