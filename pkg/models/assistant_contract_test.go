package models

import (
	"encoding/json/v2"
	"strings"
	"testing"
)

func TestAssistantHTTPContractKeepsEnvelopeAndCamelCase(t *testing.T) {
	response := AIAssistantChatResponse{
		Mode: AIAssistantModeChat, Reply: "CNY 110.00",
		Queries:      []AIAssistantQuery{{QueryID: "q1", Tool: "aggregate_transactions", Scope: AIAssistantQueryScope{Timezone: "Asia/Shanghai", AccountIDs: []string{"9007199254740993"}, IncludesHidden: true, ExcludesDeleted: true}, Coverage: "all_matching", QueriedAt: "2026-10-14T12:00:00+08:00", Warnings: []string{}}},
		Continuation: &AIAssistantContinuation{Tool: "aggregate_transactions", Cursor: "opaque"},
		References:   []*AIAssistantReferencedTransaction{{Id: 9007199254740993, SourceAmount: 11000, HideAmount: true}},
	}
	encoded, err := json.Marshal(map[string]any{"success": true, "result": response})
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["success"] != true {
		t.Fatal("success envelope changed")
	}
	result := envelope["result"].(map[string]any)
	query := result["queries"].([]any)[0].(map[string]any)
	scope := query["scope"].(map[string]any)
	if query["queryId"] != "q1" || query["queriedAt"] != "2026-10-14T12:00:00+08:00" || scope["accountIds"].([]any)[0] != "9007199254740993" {
		t.Fatalf("camelCase contract changed: %s", encoded)
	}
	if scope["startTime"] != nil || scope["endTime"] != nil {
		t.Fatal("unbounded scope needs null times")
	}
	reference := result["references"].([]any)[0].(map[string]any)
	if reference["id"] != "9007199254740993" || reference["hideAmount"] != true {
		t.Fatal("reference ID/display contract changed")
	}
	for _, forbidden := range []string{"query_id", "account_ids", "queried_at", "rowsRead", "providerState"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("non-HTTP field leaked: %s", encoded)
		}
	}
}

func TestAssistantStreamAndErrorContracts(t *testing.T) {
	status, err := json.Marshal(AIAssistantChatStreamChunk{Type: AIAssistantChatStreamChunkTypeStatus, Status: AIAssistantStatusQuerying})
	if err != nil {
		t.Fatal(err)
	}
	var chunk map[string]any
	if err := json.Unmarshal(status, &chunk); err != nil {
		t.Fatal(err)
	}
	if chunk["type"] != "status" || chunk["status"] != "querying" {
		t.Fatalf("invalid status chunk: %s", status)
	}
	if _, wrapped := chunk["result"]; wrapped {
		t.Fatal("SSE must remain a direct chunk")
	}
	encoded, err := json.Marshal(map[string]any{
		"success": false, "errorCode": 123, "errorMessage": "Incomplete", "path": "/api/v1/llm/assistant/chat.json",
		"context": map[string]any{"assistant": AIAssistantErrorContext{State: AIAssistantStateIncomplete, Queries: []AIAssistantQuery{}, Continuation: &AIAssistantContinuation{Tool: "query_transactions", Cursor: "opaque"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatal(err)
	}
	details := envelope["context"].(map[string]any)["assistant"].(map[string]any)
	if envelope["success"] != false || details["state"] != "incomplete" || details["continuation"].(map[string]any)["cursor"] != "opaque" {
		t.Fatalf("error context changed: %s", encoded)
	}
	if _, ok := details["queries"].([]any); !ok {
		t.Fatal("error evidence must remain an array")
	}
}

func TestAssistantContinuationRequestAndReceiptRemainCompatible(t *testing.T) {
	var request AIAssistantChatRequest
	if err := json.Unmarshal([]byte(`{"mode":"chat","message":"next page","history":[{"role":"user","content":"all expenses"}],"continuation":{"tool":"query_transactions","cursor":"opaque"}}`), &request); err != nil {
		t.Fatal(err)
	}
	if request.Continuation == nil || request.Continuation.Tool != "query_transactions" || request.History[0].Content != "all expenses" {
		t.Fatal("continuation request failed")
	}
	// The existing receipt aliases and numeric minor-unit amounts remain intact.
	receipt := RecognizedReceiptImageResponse{SourceAmount: 12345, SourceAccountId: 9007199254740993}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"sourceAccountId":"9007199254740993"`) || !strings.Contains(string(encoded), `"sourceAmount":12345`) {
		t.Fatalf("receipt compatibility changed: %s", encoded)
	}
}
