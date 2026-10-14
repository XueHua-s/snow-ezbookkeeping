package datastore

import (
	"context"

	"xorm.io/xorm/log"

	"github.com/mayswind/ezbookkeeping/pkg/core"
)

// XOrmContextAdapter preserves the caller's lifecycle and xorm's logging ID.
type XOrmContextAdapter struct {
	context.Context
	requestId string
}

// Value preserves xorm's SessionID while delegating all other values.
func (c *XOrmContextAdapter) Value(key any) any {
	if key == log.SessionIDKey && c.requestId != "" {
		return c.requestId
	}
	return c.Context.Value(key)
}

func NewXOrmContextAdapter(c core.Context) *XOrmContextAdapter {
	if c != nil {
		return &XOrmContextAdapter{Context: c, requestId: c.GetContextId()}
	}
	return &XOrmContextAdapter{Context: context.Background()}
}
