package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"strings"
	"testing"
	"time"
)

func TestAssistantRunnerCancellationAndSinkFailure(t *testing.T) {
	t.Run("terminal write before observation", func(t *testing.T) {
		in := runnerInput()
		writeErr := errors.New("terminal write failed")
		finalized, observed := 0, 0
		in.Finalize = func(result AssistantRunResult, err error) error {
			finalized++
			if err != nil || result.State != AssistantStateCompleted || observed != 0 {
				t.Fatalf("unexpected terminal outcome: %+v %v", result, err)
			}
			return writeErr
		}
		in.Observe = func(o AssistantObservation) {
			observed++
			if finalized != 1 || o.Reason != AssistantReasonWriteFailed {
				t.Fatalf("incorrect terminal observation: %+v", o)
			}
		}
		r, err := NewAssistantRunner(runnerScript(t, runnerFinal), nil, &runnerExecutor{}).Run(runnerContext(), in, nil)
		if !errors.Is(err, writeErr) || r.State != AssistantStateCanceled || finalized != 1 || observed != 1 {
			t.Fatalf("terminal write failure lost: %+v %v", r, err)
		}
	})
	t.Run("transport cancellation cause", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(nil)
		in := runnerInput()
		var observations []AssistantObservation
		in.Observe = func(o AssistantObservation) { observations = append(observations, o) }
		m := runnerModel(func(c core.Context, _ int64, _ data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
			cancel(ErrAssistantWriteFailed)
			<-c.Done()
			return data.TurnResponse{}, c.Err()
		})
		r, err := NewAssistantRunner(m, runnerScript(t), &runnerExecutor{}).Run(core.NewAssistantContext(ctx, "id", "en", ""), in, nil)
		if !errors.Is(err, context.Canceled) || r.State != AssistantStateCanceled || len(observations) != 1 || observations[0].Reason != AssistantReasonWriteFailed {
			t.Fatalf("transport cause lost: %+v %+v %v", r, observations, err)
		}
	})
	t.Run("parent canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		r, err := NewAssistantRunner(runnerScript(t), nil, &runnerExecutor{}).Run(core.NewAssistantContext(ctx, "id", "en", ""), runnerInput(), nil)
		if !errors.Is(err, context.Canceled) || r.State != AssistantStateCanceled {
			t.Fatalf("bad cancellation %+v %v", r, err)
		}
	})
	t.Run("provider cancellation", func(t *testing.T) {
		in := runnerInput()
		in.RunTimeout = time.Millisecond
		m := runnerModel(func(c core.Context, _ int64, _ data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
			<-c.Done()
			return data.TurnResponse{}, c.Err()
		})
		r, err := NewAssistantRunner(m, runnerScript(t), &runnerExecutor{}).Run(runnerContext(), in, nil)
		if !errors.Is(err, context.DeadlineExceeded) || r.State != AssistantStateIncomplete {
			t.Fatalf("bad timeout %+v %v", r, err)
		}
	})
	t.Run("query timeout", func(t *testing.T) {
		in := runnerInput()
		in.QueryTimeout = time.Millisecond
		e := &runnerExecutor{execute: func(c core.Context, call data.ToolCall) data.ToolResult { <-c.Done(); return runnerSuccess(call.ID) }}
		r, err := NewAssistantRunner(runnerScript(t, runnerTurn(runnerCall("a", `{}`))), runnerScript(t), e).Run(runnerContext(), in, nil)
		runnerRequireCode(t, err, data.ToolErrorQueryTimeout)
		if len(r.Queries) != 0 {
			t.Fatal("timed out data accepted")
		}
	})
	for _, phase := range []AssistantEventType{AssistantEventStatus, AssistantEventThinkingDelta, AssistantEventReplyDelta, AssistantEventReferences} {
		t.Run(string(phase), func(t *testing.T) {
			writeErr := errors.New("socket secret")
			e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult {
				r := runnerSuccess(call.ID)
				r.References = []data.TransactionReference{{ID: "a"}}
				return r
			}}
			m := runnerScript(t, runnerTurn(runnerCall("a", `{}`)), func(in data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
				if phase == AssistantEventThinkingDelta {
					err := emit(data.DisplayEvent{ThinkingDelta: "thinking"})
					return data.TurnResponse{}, err
				}
				return runnerFinal(in, emit)
			})
			var observations []AssistantObservation
			in := runnerInput()
			in.Observe = func(o AssistantObservation) { observations = append(observations, o) }
			r, err := NewAssistantRunner(m, runnerScript(t), e).Run(runnerContext(), in, func(event AssistantEvent) error {
				if event.Type == phase {
					return writeErr
				}
				return nil
			})
			if !errors.Is(err, writeErr) || r.State != AssistantStateCanceled {
				t.Fatalf("write failure lost %+v %v", r, err)
			}
			finals := 0
			for _, o := range observations {
				if o.Kind == AssistantObservationRun {
					finals++
					if o.Reason != AssistantReasonWriteFailed {
						t.Fatalf("wrong reason %+v", o)
					}
				}
			}
			if finals != 1 {
				t.Fatal("terminal observation not unique")
			}
		})
	}
}

func TestAssistantRunnerObservationPrivacyAndClock(t *testing.T) {
	const canary = "PRIVATE-BALANCE-NOTE-KEY-CURSOR-STATE"
	now := time.Unix(100, 0)
	clock := func() time.Time { v := now; now = now.Add(5 * time.Millisecond); return v }
	var observations []AssistantObservation
	in := runnerInput()
	in.Messages[0].Content = canary
	in.Observe = func(o AssistantObservation) { observations = append(observations, o) }
	first := time.Unix(100, 0).Add(2 * time.Millisecond)
	in.FirstStatusWrittenAt = func() *time.Time { return &first }
	e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult {
		r := runnerSuccess(call.ID)
		r.Data = json.RawMessage(`"` + canary + `"`)
		r.NextCursor = canary
		r.HasMore = true
		r.References = []data.TransactionReference{{ID: "a", Comment: canary, SourceAmount: canary}}
		return r
	}}
	m := runnerScript(t, runnerTurn(runnerCall("a", `{"keyword":"`+canary+`"}`)), runnerTurn(runnerCall("b", `{"keyword":"`+canary+`"}`)), runnerFinal)
	_, err := NewAssistantRunner(m, nil, e, clock).Run(runnerContext(), in, func(AssistantEvent) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(observations)
	if strings.Contains(string(encoded), canary) {
		t.Fatalf("private data in telemetry %s", encoded)
	}
	if len(observations) != 3 {
		t.Fatalf("observation count %d", len(observations))
	}
	last := observations[2]
	if last.Rounds != 3 || last.Executions != 1 || last.CacheHits != 1 || last.RowsRead != 3 || last.ResultRecords != 2 || last.SerializedBytes <= 0 || last.DurationMilliseconds != 25 || last.FirstStatusMilliseconds == nil || *last.FirstStatusMilliseconds != 2 || last.RequestID != "test-request" {
		t.Fatalf("bad aggregate telemetry %+v", last)
	}
	if observations[0].DurationMilliseconds != 5 || observations[1].RowsRead != 0 || observations[1].ResultRecords != 0 {
		t.Fatal("tool measurement/cache double count")
	}
}

func TestAssistantRunnerKnowledgeAndEmptySuccess(t *testing.T) {
	for _, empty := range []bool{false, true} {
		t.Run(fmt.Sprint(empty), func(t *testing.T) {
			e := &runnerExecutor{}
			steps := []func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error){}
			if empty {
				steps = append(steps, runnerTurn(runnerCall("empty", `{}`)))
			}
			steps = append(steps, runnerFinal)
			var last AssistantObservation
			in := runnerInput()
			in.Observe = func(o AssistantObservation) { last = o }
			r, err := NewAssistantRunner(runnerScript(t, steps...), nil, e).Run(runnerContext(), in, nil)
			if err != nil || r.State != AssistantStateCompleted || last.FirstStatusMilliseconds != nil {
				t.Fatalf("knowledge/empty rejected %+v %v", r, err)
			}
			if !empty && len(e.calls) != 0 {
				t.Fatal("knowledge needs no tools")
			}
		})
	}
}

type normalizedRunnerExecutor struct{ runnerExecutor }

func (e *normalizedRunnerExecutor) NormalizeToolArguments(call data.ToolCall) (json.RawMessage, error) {
	if string(call.Arguments) == `{}` || string(call.Arguments) == `{"limit":50}` {
		return json.RawMessage(`{"normalized":{"limit":50}}`), nil
	}
	return nil, errors.New("invalid arguments")
}

func TestAssistantRunnerDomainNormalizedCache(t *testing.T) {
	e := &normalizedRunnerExecutor{}
	m := runnerScript(t, runnerTurn(runnerCall("a", `{}`)), runnerTurn(runnerCall("b", `{"limit":50}`)), runnerFinal)
	r, err := NewAssistantRunner(m, nil, e).Run(runnerContext(), runnerInput(), nil)
	if err != nil || len(e.calls) != 1 || len(r.Queries) != 1 {
		t.Fatalf("domain defaults not cached %+v %v", r, err)
	}
}

func TestAssistantRunnerProviderDeadlineFallback(t *testing.T) {
	for _, visible := range []bool{false, true} {
		t.Run(fmt.Sprint(visible), func(t *testing.T) {
			backupCalls := 0
			primary := runnerScript(t, func(_ data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
				if visible {
					if err := emit(data.DisplayEvent{ThinkingDelta: "summary"}); err != nil {
						return data.TurnResponse{}, err
					}
				}
				return data.TurnResponse{}, fmt.Errorf("provider timeout: %w", context.DeadlineExceeded)
			})
			backup := runnerModel(func(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
				backupCalls++
				return data.TurnResponse{Completed: true, Content: "recovered"}, nil
			})
			r, err := NewAssistantRunner(primary, backup, &runnerExecutor{}).Run(runnerContext(), runnerInput(), func(AssistantEvent) error { return nil })
			if visible {
				runnerRequireCode(t, err, data.ToolErrorUpstreamFailed)
				if backupCalls != 0 {
					t.Fatal("switched after thinking")
				}
			} else if err != nil || backupCalls != 1 || r.Reply != "recovered" {
				t.Fatalf("provider deadline did not switch %+v %v", r, err)
			}
		})
	}
}

func TestAssistantRunnerObservationExitReasons(t *testing.T) {
	for _, test := range []struct {
		name   string
		reason AssistantStopReason
		code   data.ToolErrorCode
	}{
		{"invalid", AssistantReasonInvalidArgument, data.ToolErrorInvalidArgument},
		{"permission", AssistantReasonNotAccessible, data.ToolErrorNotFoundOrNotAccessible},
		{"database", AssistantReasonDataInconsistent, data.ToolErrorDataInconsistent},
		{"query timeout", AssistantReasonQueryTimeout, data.ToolErrorQueryTimeout},
		{"budget", AssistantReasonContextBudget, data.ToolErrorBudgetExhausted},
		{"protocol", AssistantReasonProtocolError, data.ToolErrorProtocol},
		{"canceled", AssistantReasonCanceled, data.ToolErrorCanceled},
	} {
		t.Run(test.name, func(t *testing.T) {
			const canary = "SECRET-BILL-BALANCE-CURSOR-PROMPT-STATE"
			e := &runnerExecutor{execute: func(core.Context, data.ToolCall) data.ToolResult {
				return data.ToolResult{Error: &data.ToolError{Code: test.code, Message: canary, Field: canary}, Metrics: data.ToolExecutionMetrics{RowsRead: 7, ResultRecords: 0}}
			}}
			var records []AssistantObservation
			in := runnerInput()
			in.Observe = func(o AssistantObservation) { records = append(records, o) }
			m := runnerScript(t, runnerTurn(runnerCall("a", `{}`)), runnerFinal)
			_, err := NewAssistantRunner(m, runnerScript(t), e).Run(runnerContext(), in, nil)
			if err == nil {
				t.Fatal("error completed")
			}
			if len(records) != 2 || records[0].Kind != AssistantObservationTool || records[1].Kind != AssistantObservationRun {
				t.Fatalf("must finalize once: %+v", records)
			}
			final := records[1]
			if final.Reason != test.reason || final.Executions != 1 || final.CacheHits != 0 || final.RowsRead != 7 || final.ResultRecords != 0 || final.SerializedBytes <= 0 || final.FirstStatusMilliseconds != nil {
				t.Fatalf("bad observation %+v", final)
			}
			encoded, _ := json.Marshal(records)
			if strings.Contains(string(encoded), canary) || strings.Contains(err.Error(), canary) {
				t.Fatal("unsafe observation/error")
			}
		})
	}
}

func TestAssistantRunnerCancellationDuringTool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult { cancel(); return runnerSuccess(call.ID) }}
	var observation AssistantObservation
	in := runnerInput()
	in.Observe = func(o AssistantObservation) { observation = o }
	r, err := NewAssistantRunner(runnerScript(t, runnerTurn(runnerCall("a", `{}`), runnerCall("b", `{"other":1}`))), runnerScript(t), e).Run(core.NewAssistantContext(ctx, "id", "en", ""), in, nil)
	if !errors.Is(err, context.Canceled) || len(e.calls) != 1 || len(r.Queries) != 0 || observation.Executions != 1 || observation.Reason != AssistantReasonCanceled {
		t.Fatalf("cancellation did not stop work %+v %+v %v", r, observation, err)
	}
}

func TestAssistantRunnerInvalidArgumentsCannotAliasNormalizedCache(t *testing.T) {
	e := &normalizedRunnerExecutor{}
	m := runnerScript(t, runnerTurn(runnerCall("a", `{}`)), runnerTurn(runnerCall("b", `{"normalized":{"limit":50}}`)), runnerFinal)
	_, err := NewAssistantRunner(m, nil, e).Run(runnerContext(), runnerInput(), nil)
	if err != nil || len(e.calls) != 2 {
		t.Fatalf("invalid raw parameters reused validated cache: %v", err)
	}
}
