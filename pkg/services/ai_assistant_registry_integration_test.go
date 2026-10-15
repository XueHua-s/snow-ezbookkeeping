package services

import (
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

// This gate exercises the actual static factory registry, without replacing
// handlers. All seven implementations must work when phase 3 is integrated.
func TestAssistantRegistryRealSevenTools(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	expected := map[data.ToolName]bool{
		data.ToolListAccounts: false, data.ToolListCategories: false, data.ToolListTags: false,
		data.ToolQueryTransactions: false, data.ToolGetTransaction: false,
		data.ToolAggregateTransactions: false, data.ToolGetAccountBalances: false,
	}
	if len(s.Definitions()) != len(expected) {
		t.Fatal("registry must expose exactly seven tools")
	}
	for _, definition := range s.Definitions() {
		seen, exists := expected[definition.Name]
		if !exists || seen {
			t.Fatal("unknown or duplicate tool", definition.Name)
		}
		expected[definition.Name] = true
		t.Run(string(definition.Name), func(t *testing.T) {
			raw := `{}`
			if definition.Name == data.ToolGetTransaction {
				raw = `{"transaction_id":"100"}`
			}
			result := assistantCall(s, definition.Name, raw)
			if !result.OK || result.Error != nil {
				t.Fatalf("real registered tool failed: %+v", result.Error)
			}
			if len(result.Data) == 0 || result.QueryID == "" || result.Scope == nil || result.QueriedAt == "" {
				t.Fatal("registered tool omitted evidence", result)
			}
		})
	}
}
