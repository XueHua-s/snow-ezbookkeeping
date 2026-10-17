package llm

import (
	"encoding/json"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

func TestAssistantRunnerUnrelatedSuccessCannotResolveFailedQuery(t *testing.T) {
	for _, code := range []data.ToolErrorCode{data.ToolErrorNotFoundOrNotAccessible, data.ToolErrorInvalidCursor, data.ToolErrorInvalidArgument} {
		for _, mode := range []string{"same-round", "cross-round", "cached-success", "multiple-failures", "exact-retry"} {
			t.Run(string(code)+"/"+mode, func(t *testing.T) {
				bad := runnerCall("bad", `{"id":"999"}`)
				bad.Name = data.ToolGetTransaction
				good := bad
				good.ID, good.Arguments = "good", json.RawMessage(`{"id":"100"}`)
				retry := bad
				retry.ID = "retry"
				e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult {
					if call.ID == "bad" || call.ID == "another-bad" {
						return data.ToolResult{Error: &data.ToolError{Code: code}}
					}
					return runnerSuccess(call.ID)
				}}
				steps := []func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error){}
				switch mode {
				case "same-round":
					steps = append(steps, runnerTurn(bad, good))
				case "cross-round":
					steps = append(steps, runnerTurn(bad), runnerTurn(good))
				case "cached-success":
					cached := good
					cached.ID = "cached"
					steps = append(steps, runnerTurn(good), runnerTurn(bad, cached))
				case "multiple-failures":
					other := bad
					other.ID, other.Arguments = "another-bad", json.RawMessage(`{"id":"888"}`)
					steps = append(steps, runnerTurn(bad, other), runnerTurn(retry))
				case "exact-retry":
					steps = append(steps, runnerTurn(bad), runnerTurn(retry))
				}
				steps = append(steps, func(_ data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
					if emit != nil {
						if err := emit(data.DisplayEvent{ReplyDelta: "unsupported claim"}); err != nil {
							return data.TurnResponse{}, err
						}
					}
					return data.TurnResponse{Completed: true, Content: "unsupported claim"}, nil
				})
				var deltas string
				result, err := NewAssistantRunner(runnerScript(t, steps...), nil, e).Run(runnerContext(), runnerInput(), func(event AssistantEvent) error {
					if event.Type == AssistantEventReplyDelta {
						deltas += event.Delta
					}
					return nil
				})
				if mode == "exact-retry" {
					if err != nil || result.State != AssistantStateCompleted {
						t.Fatalf("exact retry failed: %+v %v", result, err)
					}
				} else {
					runnerRequireCode(t, err, code)
					if result.State != AssistantStateIncomplete || result.Reply != "" || deltas != "" {
						t.Fatalf("unresolved query leaked answer: %+v %q", result, deltas)
					}
				}
			})
		}
	}
}

func TestAssistantArgumentCorrectionPreservesOtherParameters(t *testing.T) {
	for _, tc := range []struct {
		name, before, after, field string
		want                       bool
	}{
		{"remove-invalid", `{"date":"wrong"}`, `{}`, "date", true},
		{"schema-field", `{"date":"wrong","id":"999"}`, `{"id":"999","date":"today"}`, "arguments.date", true},
		{"add-missing", `{"id":"999"}`, `{"id":"999","date":"today"}`, "arguments.date", true},
		{"changed-scope", `{"date":"wrong","id":"999"}`, `{"id":"100"}`, "date", false},
		{"numeric-precision", `{"date":"wrong","id":9007199254740993}`, `{"id":9007199254740992}`, "date", false},
		{"duplicate", `{"date":"wrong","id":1,"id":2}`, `{"id":2}`, "date", false},
		{"unlabelled", `{"id":"999"}`, `{"id":"100"}`, "", false},
		{"generic", `{"id":"999"}`, `{"id":"100"}`, "arguments", false},
		{"nested", `{"periods":[{"name":"bad"}]}`, `{"periods":[]}`, "arguments.periods.name", false},
		{"generic-time", `{"start_time":"wrong"}`, `{}`, "time", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			failed := assistantUnresolvedCall{call: runnerCall("bad", tc.before), failure: data.ToolError{Code: data.ToolErrorInvalidArgument, Field: tc.field}}
			if got := assistantCorrectsArgument(failed, runnerCall("fixed", tc.after)); got != tc.want {
				t.Fatalf("correction=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestAssistantRunnerGenuineEmptyQueryCanComplete(t *testing.T) {
	e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult {
		result := runnerSuccess(call.ID)
		result.Metrics = data.ToolExecutionMetrics{}
		return result
	}}
	result, err := NewAssistantRunner(runnerScript(t, runnerTurn(runnerCall("empty", `{}`)), runnerFinal), nil, e).Run(runnerContext(), runnerInput(), nil)
	if err != nil || result.State != AssistantStateCompleted || result.Reply == "" {
		t.Fatalf("empty query rejected: %+v %v", result, err)
	}
}

func TestAssistantRunnerValidationCorrectionChains(t *testing.T) {
	for _, mode := range []string{"chain", "ambiguous", "same-round", "unrelated", "exact-precedence"} {
		t.Run(mode, func(t *testing.T) {
			first := runnerCall("first", `{"filter":{"currency":"bad","start_time":"bad"}}`)
			second := runnerCall("second", `{"filter":{"currency":"CNY","start_time":"bad"}}`)
			last := runnerCall("last", `{"filter":{"currency":"CNY","start_time":"2020-01-01"}}`)
			e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult {
				if call.ID == "first" || call.ID == "independent" {
					return data.ToolResult{Error: &data.ToolError{Code: data.ToolErrorInvalidArgument, Field: "arguments.filter.currency"}}
				}
				if call.ID == "second" {
					return data.ToolResult{Error: &data.ToolError{Code: data.ToolErrorInvalidArgument, Field: "filter.start_time"}}
				}
				return runnerSuccess(call.ID)
			}}
			steps := []func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error){}
			switch mode {
			case "chain":
				steps = append(steps, runnerTurn(first), runnerTurn(second), runnerTurn(last))
			case "ambiguous":
				other := runnerCall("independent", `{"filter":{"currency":"also bad","start_time":"bad"}}`)
				steps = append(steps, runnerTurn(first, other), runnerTurn(second), runnerTurn(last))
			case "same-round":
				last.Arguments = second.Arguments
				steps = append(steps, runnerTurn(first, last))
			case "unrelated":
				other := runnerCall("independent", `{"filter":{"currency":"bad","start_time":"different"}}`)
				steps = append(steps, runnerTurn(first, other), runnerTurn(second), runnerTurn(last))
			case "exact-precedence":
				other := runnerCall("independent", `{"filter":{"currency":"also bad","start_time":"bad"}}`)
				last.Arguments = first.Arguments
				steps = append(steps, runnerTurn(first, other), runnerTurn(last))
			}
			steps = append(steps, runnerFinal)
			result, err := NewAssistantRunner(runnerScript(t, steps...), nil, e).Run(runnerContext(), runnerInput(), nil)
			if mode == "chain" {
				if err != nil || result.State != AssistantStateCompleted {
					t.Fatalf("chain failed: %+v %v", result, err)
				}
			} else {
				runnerRequireCode(t, err, data.ToolErrorInvalidArgument)
				if result.Reply != "" {
					t.Fatal("independent validation failure discarded")
				}
			}
		})
	}
}

func TestAssistantRunnerCorrectedValidationRetainsNewReadFailure(t *testing.T) {
	for _, code := range []data.ToolErrorCode{data.ToolErrorNotFoundOrNotAccessible, data.ToolErrorInvalidCursor} {
		for _, mode := range []string{"exact-retry", "different-id", "validation-cannot-weaken-read"} {
			t.Run(string(code)+"/"+mode, func(t *testing.T) {
				call := func(id, transaction string) data.ToolCall {
					return data.ToolCall{ID: id, Name: data.ToolGetTransaction, Arguments: json.RawMessage(`{"transaction_id":"` + transaction + `"}`)}
				}
				e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult {
					switch call.ID {
					case "invalid", "invalid-again":
						return data.ToolResult{Error: &data.ToolError{Code: data.ToolErrorInvalidArgument, Field: "transaction_id"}}
					case "read-failed":
						return data.ToolResult{Error: &data.ToolError{Code: code}}
					default:
						return runnerSuccess(call.ID)
					}
				}}
				steps := []func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error){
					runnerTurn(call("invalid", "bad")), runnerTurn(call("read-failed", "100")),
				}
				if mode == "validation-cannot-weaken-read" {
					steps = append(steps, runnerTurn(call("invalid-again", "100")))
				}
				transaction := "100"
				if mode != "exact-retry" {
					transaction = "101"
				}
				steps = append(steps, runnerTurn(call("success", transaction)), runnerFinal)
				result, err := NewAssistantRunner(runnerScript(t, steps...), nil, e).Run(runnerContext(), runnerInput(), nil)
				if mode == "exact-retry" {
					if err != nil || result.State != AssistantStateCompleted {
						t.Fatalf("corrected query could not recover: %+v %v", result, err)
					}
				} else {
					runnerRequireCode(t, err, code)
					if result.State != AssistantStateIncomplete || result.Reply != "" {
						t.Fatalf("read failure was weakened: %+v", result)
					}
				}
			})
		}
	}
}
