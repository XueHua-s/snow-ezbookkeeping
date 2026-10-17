package llm_test

import (
	"encoding/json/v2"
	"reflect"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
)

type legacyProvider struct{}

func (legacyProvider) GetJsonResponse(core.Context, int64, *data.LargeLanguageModelRequest) (*data.LargeLanguageModelTextualResponse, error) {
	return nil, nil
}
func (legacyProvider) StreamTextResponse(core.Context, int64, *data.LargeLanguageModelRequest, data.LargeLanguageModelStreamCallback) (*data.LargeLanguageModelStreamResponse, error) {
	return nil, nil
}

var _ provider.LargeLanguageModelProvider = legacyProvider{}
var _ provider.LargeLanguageModelStreamingProvider = legacyProvider{}

type toolProvider struct{}

func (toolProvider) RunToolTurn(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
	return data.TurnResponse{Completed: true}, nil
}

var _ provider.LargeLanguageModelToolProvider = toolProvider{}

type executor struct{}

func (executor) Definitions() []data.ToolDefinition { return nil }
func (executor) Execute(core.Context, data.ToolCall) data.ToolResult {
	return data.ToolResult{OK: true}
}

var _ llm.ToolExecutor = executor{}

func TestToolSupportRemainsOptional(t *testing.T) {
	var original provider.LargeLanguageModelProvider = legacyProvider{}
	if _, ok := original.(provider.LargeLanguageModelToolProvider); ok {
		t.Fatal("legacy provider unexpectedly requires tool support")
	}
	// Tool-only adapters need no receipt implementation either.
	var toolOnly provider.LargeLanguageModelToolProvider = toolProvider{}
	if _, ok := toolOnly.(provider.LargeLanguageModelProvider); ok {
		t.Fatal("tool adapter unexpectedly requires receipt support")
	}
}

func TestObservationHasExplicitNoStatusAndAllowlistedFields(t *testing.T) {
	record := llm.AssistantObservation{Kind: llm.AssistantObservationRun, RequestID: "request-1", Reason: llm.AssistantReasonCompleted}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	status, present := fields["firstStatusMilliseconds"]
	if !present || status != nil {
		t.Fatalf("JSON/no-status must be explicitly null: %s", encoded)
	}
	zero := int64(0)
	record.FirstStatusMilliseconds = &zero
	encoded, err = json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["firstStatusMilliseconds"] != float64(0) {
		t.Fatal("actual zero-latency status must differ from not applicable")
	}
	if _, present := fields["contextFailureSource"]; present {
		t.Fatal("empty context failure source must be omitted")
	}
	if count, present := fields["contextRecoveryAttempts"]; present && count != float64(0) {
		t.Fatal("unused recovery counter must be zero when emitted")
	}
	record.ContextFailureSource = "upstream_input"
	record.ContextRecoveryAttempts = 1
	encoded, err = json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["contextFailureSource"] != "upstream_input" || fields["contextRecoveryAttempts"] != float64(1) {
		t.Fatal("controlled context telemetry missing")
	}
	// Guard the complete record surface: adding arbitrary context/maps/errors
	// makes secret/financial payload logging too easy and requires explicit review.
	want := []string{"ContextFailureSource", "ContextRecoveryAttempts", "Kind", "RequestID", "Tool", "RowsRead", "ResultRecords", "SerializedBytes", "DurationMilliseconds", "Rounds", "Executions", "CacheHits", "Reason", "FirstStatusMilliseconds"}
	typ := reflect.TypeOf(record)
	if typ.NumField() != len(want) {
		t.Fatal("telemetry surface changed; review privacy allowlist")
	}
	for i, name := range want {
		if typ.Field(i).Name != name {
			t.Fatalf("unexpected telemetry field %s", typ.Field(i).Name)
		}
	}
}
