package datastore

import (
	"context"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"xorm.io/xorm/log"
)

func TestXOrmContextAdapterDelegation(t *testing.T) {
	type key struct{}
	parent, cancel := context.WithTimeout(context.WithValue(context.Background(), key{}, "value"), time.Minute)
	defer cancel()
	c := NewXOrmContextAdapter(core.NewAssistantContext(parent, "request", "", ""))
	if c.Value(key{}) != "value" || c.Value(log.SessionIDKey) != "request" {
		t.Fatal("values/session ID lost")
	}
	want, _ := parent.Deadline()
	got, ok := c.Deadline()
	if !ok || !got.Equal(want) {
		t.Fatal("deadline lost")
	}
	cancel()
	if c.Err() != context.Canceled {
		t.Fatal(c.Err())
	}
	select {
	case <-c.Done():
	default:
		t.Fatal("cancellation lost")
	}
	c = NewXOrmContextAdapter(nil)
	if c.Done() != nil || c.Err() != nil || c.Value(key{}) != nil {
		t.Fatal("nil context")
	}
	c = NewXOrmContextAdapter(core.NewAssistantContext(context.WithValue(context.Background(), log.SessionIDKey, "parent-session"), "", "", ""))
	if c.Value(log.SessionIDKey) != "parent-session" {
		t.Fatal("empty request ID must delegate")
	}
}
