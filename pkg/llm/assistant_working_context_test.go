package llm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

func TestAssistantWorkingContextPreservesEveryCellAndEvidence(t *testing.T) {
	raw := json.RawMessage(`[{"transaction_id":"9223372036854775807","amount":"-123456789.01","memo":null,"integer":9223372036854775807},{"transaction_id":"9223372036854775806","amount":"0.01","memo":"same","integer":9223372036854775806}]`)
	raw = json.RawMessage("[" + strings.TrimSuffix(strings.TrimPrefix(string(raw), "["), "]") + "," + strings.TrimSuffix(strings.TrimPrefix(string(raw), "["), "]") + "]")
	result := runnerSuccess("query")
	result.Data, result.Totals = raw, json.RawMessage(`{"CNY":"-123456789.00"}`)
	result.HasMore, result.NextCursor = true, "signed-cursor"
	original := data.ToolTurnInput{Messages: []data.ConversationMessage{{Role: data.ConversationRoleUser, Content: "exact date and account"}, {Role: data.ConversationRoleAssistant, ToolCalls: []data.ToolCall{runnerCall("a", `{}`)}}, {Role: data.ConversationRoleTool, ToolCallID: "a", ToolResult: &result}}, State: "private"}
	compact := compactAssistantWorkingContext(original)
	if compact.State != nil || len(compact.Messages) != 3 || compact.Messages[0].Content != original.Messages[0].Content {
		t.Fatal("lost user constraint/pairing")
	}
	out := compact.Messages[2].ToolResult
	if string(result.Data) != string(raw) || out.TotalsCoverage != result.TotalsCoverage || string(out.Totals) != string(result.Totals) || out.NextCursor != result.NextCursor || !out.HasMore {
		t.Fatal("mutated facts/evidence")
	}
	var table struct {
		Format  string
		Columns []string
		Rows    [][]json.RawMessage
	}
	if err := json.Unmarshal(out.Data, &table); err != nil || table.Format != "column_rows" || len(table.Rows) != 4 {
		t.Fatal(string(out.Data), err)
	}
	var records []map[string]json.RawMessage
	_ = json.Unmarshal(raw, &records)
	for i, row := range table.Rows {
		for j, cell := range row {
			if string(cell) != string(records[i][table.Columns[j]]) {
				t.Fatal("changed cell", i, j, string(cell))
			}
		}
	}
	again := compactAssistantWorkingContext(compact)
	if string(again.Messages[2].ToolResult.Data) != string(out.Data) {
		t.Fatal("not idempotent")
	}
}

func TestAssistantWorkingContextKeepsMissingDistinctFromNull(t *testing.T) {
	raw := json.RawMessage(`[{"long_column_name":null},{"other_column_name":null}]`)
	if string(compactAssistantData(raw)) != string(raw) {
		t.Fatal("different keys collapsed")
	}
}

func TestAssistantRunnerUpstreamContextRetryIsBoundedAndPreservesFacts(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		visible, local, always bool
		wantCalls              int
	}{{"recover", false, false, false, 2}, {"bounded", false, false, true, 2}, {"visible", true, false, true, 1}, {"local", false, true, true, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			model := runnerModel(func(_ core.Context, _ int64, in data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
				calls++
				if in.Messages[0].Content != "question" {
					t.Fatal("user lost")
				}
				if calls == 2 && in.MaxOutputTokens != 1024 {
					t.Fatal("output reserve unchanged")
				}
				if calls == 1 || tc.always {
					if tc.visible {
						_ = emit(data.DisplayEvent{ReplyDelta: "partial"})
					}
					return data.TurnResponse{}, &data.ToolError{Code: data.ToolErrorBudgetExhausted, UpstreamContext: !tc.local}
				}
				return data.TurnResponse{Completed: true, Content: "complete"}, nil
			})
			input := runnerInput()
			input.MaxOutputTokens = 2048
			var observation AssistantObservation
			input.Observe = func(o AssistantObservation) { observation = o }
			result, err := NewAssistantRunner(model, nil, &runnerExecutor{}).Run(runnerContext(), input, func(AssistantEvent) error { return nil })
			if tc.name == "recover" && (observation.ContextRecoveryAttempts != 1 || observation.ContextFailureSource != "") {
				t.Fatal("missing successful recovery telemetry", observation)
			}
			if tc.name == "bounded" && (observation.ContextRecoveryAttempts != 1 || observation.ContextFailureSource != "upstream_input") {
				t.Fatal("missing rejection source", observation)
			}
			if calls != tc.wantCalls {
				t.Fatal(calls, err)
			}
			if tc.name == "recover" && (err != nil || result.Reply != "complete") {
				t.Fatal(result, err)
			}
		})
	}
}

type exactBudgetModel struct{ runnerModel }

func (p exactBudgetModel) CheckToolTurnInput(data.ToolTurnInput) error { return nil }
func TestAssistantRunnerUsesProviderWireAuthority(t *testing.T) {
	input := runnerInput()
	input.Messages[0].Content = strings.Repeat("x", 100)
	input.MaxInputTokens = 1
	model := exactBudgetModel{runnerModel(func(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
		return data.TurnResponse{Completed: true, Content: "wire accepted"}, nil
	})}
	result, err := NewAssistantRunner(model, nil, &runnerExecutor{}).Run(runnerContext(), input, nil)
	if err != nil || result.Reply != "wire accepted" {
		t.Fatal("neutral serialization overruled provider", err)
	}
}

func TestAssistantWorkingContextCompactsAllBusinessProjections(t *testing.T) {
	row := `{"period_name":"baseline","currency":"CNY","signed_exact_amount":"-123456789.01","transaction_count":"9183"}`
	raw := json.RawMessage("[" + strings.TrimSuffix(strings.Repeat(row+",", 56), ",") + "]")
	result := runnerSuccess("query")
	result.Data, result.Comparison, result.Totals = raw, raw, raw
	input := data.ToolTurnInput{Messages: []data.ConversationMessage{{Role: data.ConversationRoleTool, ToolCallID: "a", ToolResult: &result}}}
	output := compactAssistantWorkingContext(input).Messages[0].ToolResult
	for _, projection := range []json.RawMessage{output.Data, output.Comparison, output.Totals} {
		var table struct {
			Columns []string
			Rows    [][]json.RawMessage
		}
		if err := json.Unmarshal(projection, &table); err != nil || len(table.Rows) != 56 {
			t.Fatal("projection not compacted", err)
		}
		for _, cells := range table.Rows {
			restored := map[string]json.RawMessage{}
			for j, key := range table.Columns {
				restored[key] = cells[j]
			}
			var original map[string]json.RawMessage
			_ = json.Unmarshal([]byte(row), &original)
			for key, value := range original {
				if string(restored[key]) != string(value) {
					t.Fatal("lost exact business value", key)
				}
			}
		}
	}
	if string(result.Data) != string(raw) || string(result.Comparison) != string(raw) || string(result.Totals) != string(raw) {
		t.Fatal("original evidence mutated")
	}
}

func TestAssistantWorkingCheckpointPreservesExactFactsAndBoundsRecovery(t *testing.T) {
	result := runnerSuccess("query")
	result.Data = json.RawMessage(`{"rows":"` + strings.Repeat("x", 1200) + `"}`)
	result.Comparison = json.RawMessage(`[{"delta":"-0.01"}]`)
	result.Totals = json.RawMessage(`[{"currency":"CNY","amount":"-123456789.01"},{"currency":"USD","amount":"2000.00"}]`)
	result.MatchedCount = "9183"
	result.Warnings = []string{"transfers are logical pairs"}
	original, _ := json.Marshal(result)
	input := data.ToolTurnInput{State: "opaque", Messages: []data.ConversationMessage{
		{Role: data.ConversationRoleSystem, Content: "exact policy"},
		{Role: data.ConversationRoleUser, Content: "2022 vs 2023; account 42; USD"},
		{Role: data.ConversationRoleAssistant, Content: strings.Repeat("dated answer", 100)},
		{Role: data.ConversationRoleUser, Content: "continue"},
		{Role: data.ConversationRoleAssistant, ToolCalls: []data.ToolCall{runnerCall("a", `{}`)}},
		{Role: data.ConversationRoleTool, ToolCallID: "a", ToolResult: &result},
	}}
	run := &assistantRun{c: runnerContext(), ids: map[string]assistantCallIdentity{"a": {key: [32]byte{1}, result: result}}}
	rebuilt, changed := run.recoverWorkingContext(input)
	if !changed || rebuilt.State != nil || len(rebuilt.Messages) != len(input.Messages) {
		t.Fatal("checkpoint missing")
	}
	out := rebuilt.Messages[5].ToolResult
	if string(out.Totals) != string(result.Totals) || out.Scope != result.Scope || out.QueryID != result.QueryID || out.MatchedCount != "9183" || out.Coverage != result.Coverage || len(out.Warnings) != 1 || string(out.Data) != string(assistantArchivedProjection) || out.Comparison != nil {
		t.Fatal("checkpoint lost exact facts", out)
	}
	for _, i := range []int{0, 1, 3, 4} {
		if rebuilt.Messages[i].Content != input.Messages[i].Content {
			t.Fatal("constraint lost")
		}
	}
	after, _ := json.Marshal(result)
	if string(after) != string(original) {
		t.Fatal("original evidence mutated")
	}
	if _, changed = run.recoverWorkingContext(rebuilt); changed {
		t.Fatal("idempotent checkpoint consumed generation")
	}
	rebuilt.Messages[5].ToolResult = &result
	if _, changed = run.recoverWorkingContext(rebuilt); !changed {
		t.Fatal("second generation missing")
	}
	if _, changed = run.recoverWorkingContext(input); changed {
		t.Fatal("unbounded recovery")
	}
}
