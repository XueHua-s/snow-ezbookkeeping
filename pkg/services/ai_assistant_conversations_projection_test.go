package services

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func TestAssistantSnapshotExcludesPrivateGenerationData(t *testing.T) {
	m := &models.AssistantMessage{Id: 9007199254740993, Uid: 7, ConversationId: 19, UserMessageId: 20, SchemaVersion: 1, SnapshotRevision: 1, AttemptNumber: 1, CreatedAt: 1700000000000, RunState: models.AssistantRunCompleted, Content: "原始 **Markdown**\n", Owner: "private-owner-canary", Continuation: "private-cursor-canary", InputDependenciesJSON: "private-dependencies-canary", LimitsJSON: "private-limits-canary", QueriesJSON: "[]", ReferencesJSON: `[{"id":"7","sourceAmountMinor":"9007199254740993","currency":"CNY"}]`}
	s, err := AssistantMessageSnapshot(m)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "canary") {
		t.Fatalf("private worker data leaked: %s", payload)
	}
	if s.RunId != "9007199254740993" || s.References[0].SourceAmountMinor != "9007199254740993" || s.Content != m.Content {
		t.Fatal("exact IDs, amounts or original Markdown changed")
	}
}
func TestAssistantSnapshotRejectsCorruptEvidence(t *testing.T) {
	for _, m := range []*models.AssistantMessage{{SchemaVersion: 2}, {SchemaVersion: 1, QueriesJSON: "{"}, {SchemaVersion: 1, ReferencesJSON: "["}} {
		if _, err := AssistantMessageSnapshot(m); err == nil {
			t.Fatal("corrupt persisted evidence accepted")
		}
	}
}
