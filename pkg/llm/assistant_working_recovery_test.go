package llm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

type recoveryExecutor struct {
	runnerExecutor
	budgetFailures int
}

func (e *recoveryExecutor) ExecuteForContext(c core.Context, call data.ToolCall, fits func(data.ToolResult) bool) data.ToolResult {
	if call.ID == "b" && e.budgetFailures == 0 {
		e.budgetFailures++
		return data.ToolResult{Error: assistantRunnerError(data.ToolErrorBudgetExhausted)}
	}
	return e.Execute(c, call)
}

func TestAssistantWorkingRecoveryRetriesParallelGroupAndBoundsCacheReread(t *testing.T) {
	for _, loop := range []bool{false, true} {
		t.Run(map[bool]string{false: "complete", true: "bounded-loop"}[loop], func(t *testing.T) {
			executor := &recoveryExecutor{}
			executor.execute = func(_ core.Context, call data.ToolCall) data.ToolResult {
				result := runnerSuccess(call.ID)
				result.Totals = json.RawMessage(`{"CNY":"-12.01","USD":"2000.00"}`)
				result.Data = json.RawMessage(`{"detail":"` + strings.Repeat("x", 1200) + `"}`)
				return result
			}
			rounds := 0
			model := runnerModel(func(_ core.Context, _ int64, input data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
				rounds++
				switch rounds {
				case 1:
					return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{runnerCall("a", `{"page_size":1}`)}, State: "opaque-before-recovery"}, nil
				case 2:
					return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{runnerCall("b", `{"page_size":2}`), runnerCall("c", `{"page_size":3}`)}, State: "opaque-before-recovery"}, nil
				case 3:
					if input.State != nil {
						t.Fatal("opaque state survived recovery")
					}
					if len(input.Messages) != 6 || input.Messages[2].ToolResult == nil || !strings.Contains(string(input.Messages[2].ToolResult.Data), "archived_checkpoint") {
						t.Fatal("pairs or checkpoint missing", input.Messages)
					}
					return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{runnerCall("a-read-1", `{"page_size":1}`)}}, nil
				default:
					if loop {
						return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{runnerCall("a-read-"+strings.Repeat("x", rounds), `{"page_size":1}`)}}, nil
					}
					last := input.Messages[len(input.Messages)-1].ToolResult
					if strings.Contains(string(last.Data), "archived_checkpoint") || string(last.Totals) != `{"CNY":"-12.01","USD":"2000.00"}` {
						t.Fatal("cached original was changed", last)
					}
					return data.TurnResponse{Completed: true, Content: "verified"}, nil
				}
			})
			input := runnerInput()
			input.MaxRounds = 10
			var observation AssistantObservation
			input.Observe = func(o AssistantObservation) {
				if o.Kind == AssistantObservationRun {
					observation = o
				}
			}
			result, err := NewAssistantRunner(model, nil, executor).Run(runnerContext(), input, nil)
			if len(executor.calls) != 3 || len(result.Queries) != 3 || observation.ContextRecoveryAttempts != 1 || observation.Executions != 4 {
				t.Fatal("retry/evidence accounting", len(executor.calls), len(result.Queries), observation)
			}
			if loop {
				if err == nil || result.StopReason != AssistantReasonNoProgress || rounds != 5 {
					t.Fatal("unbounded rereads", err, result.StopReason, rounds)
				}
			} else if err != nil || result.Reply != "verified" {
				t.Fatal(result, err)
			}
		})
	}
}

func TestAssistantWorkingRecoveryHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	run := &assistantRun{c: core.NewAssistantContext(ctx, "", "", "")}
	result := runnerSuccess("a")
	result.Data = json.RawMessage(`{"data":"` + strings.Repeat("x", 1000) + `"}`)
	input := data.ToolTurnInput{Messages: []data.ConversationMessage{{Role: data.ConversationRoleTool, ToolResult: &result}}, State: "original"}
	rebuilt, changed := run.recoverWorkingContext(input)
	if changed || rebuilt.State != "original" || run.workingRecoveries != 0 {
		t.Fatal("canceled request recovered")
	}
}
