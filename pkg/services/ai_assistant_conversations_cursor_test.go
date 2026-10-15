package services

import (
	"testing"
	"time"
)

func TestAssistantConversationCursorIsolationAndExpiry(t *testing.T) {
	now := time.Unix(1700000000, 0)
	secret := []byte("isolated-test-cursor-signing-key")
	original := assistantConversationCursor{UID: 17, Scope: "search:中文:false:false", Cutoff: now.UnixMilli(), Expires: now.Add(time.Minute).UnixMilli(), Time: now.UnixMilli(), ID: "42"}
	encoded, err := encodeAssistantConversationCursor(original, secret)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := decodeAssistantConversationCursor(encoded, 17, original.Scope, secret, now)
	if err != nil || actual != original {
		t.Fatalf("round trip: %#v %v", actual, err)
	}
	tests := []struct {
		name   string
		uid    int64
		scope  string
		secret []byte
		at     time.Time
		value  string
	}{
		{"another user", 18, original.Scope, secret, now, encoded},
		{"another filter", 17, "search:中文:true:false", secret, now, encoded},
		{"another key", 17, original.Scope, []byte("other"), now, encoded},
		{"expired", 17, original.Scope, secret, now.Add(time.Minute), encoded},
		{"malformed", 17, original.Scope, secret, now, "invalid"},
		{"tampered", 17, original.Scope, secret, now, encoded + "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := decodeAssistantConversationCursor(tt.value, tt.uid, tt.scope, tt.secret, tt.at); err == nil {
				t.Fatal("accepted cursor outside original query")
			}
		})
	}
}

func TestAssistantConversationSearchLiteralPattern(t *testing.T) {
	for input, want := range map[string]string{"中文": "%中文%", "100%_\\": "%100!%!_\\%", "a!b": "%a!!b%"} {
		if got := assistantConversationSearchPattern(input); got != want {
			t.Fatalf("%q: got %q want %q", input, got, want)
		}
	}
}
