package core

import (
	"context"
	"testing"
	"time"
)

func TestAssistantContext(t *testing.T) {
	type key struct{}
	parent, cancel := context.WithDeadline(context.WithValue(context.Background(), key{}, "value"), time.Now().Add(time.Minute))
	defer cancel()
	c := NewAssistantContext(parent, "request", "zh-CN", "127.0.0.2")
	if c.GetContextId() != "request" || c.GetClientLocale() != "zh-CN" || c.ClientIP() != "127.0.0.2" || c.Value(key{}) != "value" {
		t.Fatal("metadata/values lost")
	}
	deadline, ok := c.Deadline()
	expected, _ := parent.Deadline()
	if !ok || !deadline.Equal(expected) {
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
	nilParent := NewAssistantContext(nil, "", "", "")
	if nilParent.Err() != nil || nilParent.Done() != nil || nilParent.Value(key{}) != nil {
		t.Fatal("nil parent must use background")
	}
}
