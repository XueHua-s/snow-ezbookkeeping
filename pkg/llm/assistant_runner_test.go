package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
)

type runnerModel func(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error)

func (f runnerModel) RunToolTurn(c core.Context, uid int64, in data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
	return f(c, uid, in, emit)
}

type runnerExecutor struct {
	calls   []data.ToolCall
	execute func(core.Context, data.ToolCall) data.ToolResult
}

func (e *runnerExecutor) Definitions() []data.ToolDefinition {
	var definitions []data.ToolDefinition
	for _, name := range []data.ToolName{data.ToolListAccounts, data.ToolListCategories, data.ToolListTags, data.ToolQueryTransactions, data.ToolGetTransaction, data.ToolAggregateTransactions, data.ToolGetAccountBalances} {
		definitions = append(definitions, data.ToolDefinition{Name: name, Parameters: json.RawMessage(`{"type":"object"}`)})
	}
	return definitions
}
func (e *runnerExecutor) Execute(c core.Context, call data.ToolCall) data.ToolResult {
	e.calls = append(e.calls, call)
	if e.execute != nil {
		return e.execute(c, call)
	}
	return runnerSuccess(call.ID)
}
func runnerSuccess(id string) data.ToolResult {
	return data.ToolResult{OK: true, QueryID: id, Scope: &data.QueryScope{Timezone: "UTC"}, Coverage: data.QueryCoverageAllMatching, Data: json.RawMessage(`[]`), Warnings: []string{}, Metrics: data.ToolExecutionMetrics{RowsRead: 3, ResultRecords: 2}}
}
func runnerCall(id, args string) data.ToolCall {
	return data.ToolCall{ID: id, Name: data.ToolQueryTransactions, Arguments: json.RawMessage(args)}
}
func runnerContext() core.Context {
	return core.NewAssistantContext(context.Background(), "test-request", "en", "127.0.0.1")
}
func runnerInput() AssistantRunInput {
	return AssistantRunInput{UID: 42, Mode: "chat", Messages: []data.ConversationMessage{{Role: data.ConversationRoleUser, Content: "question"}}}
}
func runnerScript(t *testing.T, steps ...func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error)) provider.LargeLanguageModelToolProvider {
	t.Helper()
	i := 0
	return runnerModel(func(c core.Context, uid int64, in data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
		if uid != 42 {
			t.Fatalf("UID changed: %d", uid)
		}
		if i >= len(steps) {
			t.Fatalf("unexpected model round %d", i)
		}
		step := steps[i]
		i++
		return step(in, emit)
	})
}
func runnerTurn(calls ...data.ToolCall) func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
	return func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
		return data.TurnResponse{Completed: true, Content: "internal tool text", ToolCalls: calls}, nil
	}
}
func runnerFinal(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
	return data.TurnResponse{Completed: true, Content: "confirmed answer"}, nil
}
func runnerRequireCode(t *testing.T, err error, code data.ToolErrorCode) {
	t.Helper()
	var controlled *data.ToolError
	if !errors.As(err, &controlled) || controlled.Code != code {
		t.Fatalf("wanted %s; got %v", code, err)
	}
}

func TestAssistantRunnerCompleteRoundsAndEvidence(t *testing.T) {
	e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult {
		r := runnerSuccess(call.ID)
		r.HasMore = true
		r.NextCursor = "cursor-" + call.ID
		for i := 0; i < 10; i++ {
			r.References = append(r.References, data.TransactionReference{ID: fmt.Sprint(i)})
		}
		return r
	}}
	state := &struct{ Secret string }{"private-state"}
	m := runnerScript(t, func(in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
		if in.MaxRequestBytes != AssistantMaxModelRequestBytes {
			t.Fatal("wire budget omitted")
		}
		return data.TurnResponse{Completed: true, Content: "internal text", ToolCalls: []data.ToolCall{runnerCall("a", `{"x":1}`), runnerCall("b", `{"x":2}`)}, State: state}, nil
	}, func(in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
		if in.State != state || len(in.Messages) != 4 {
			t.Fatalf("lost prefix/state: %+v", in)
		}
		if in.Messages[1].Content != "internal text" || len(in.Messages[1].ToolCalls) != 2 || in.Messages[2].ToolCallID != "a" || in.Messages[3].ToolCallID != "b" {
			t.Fatal("tool pairing changed")
		}
		return runnerFinal(in, nil)
	})
	var events []AssistantEvent
	r, err := NewAssistantRunner(m, nil, e).Run(runnerContext(), runnerInput(), func(event AssistantEvent) error { events = append(events, event); return nil })
	if err != nil || r.State != AssistantStateCompleted || r.Reply != "confirmed answer" || len(r.Queries) != 2 || len(r.References) != 8 || r.Continuation.Cursor != "cursor-b" {
		t.Fatalf("unexpected result: %+v %v", r, err)
	}
	for _, event := range events {
		if event.Delta == "internal text" {
			t.Fatal("tool text leaked")
		}
	}
	if len(e.calls) != 2 || e.calls[0].ID != "a" || e.calls[1].ID != "b" {
		t.Fatal("tools were not sequential")
	}
}

func TestAssistantRunnerCorrectsErrorsWithoutEmptySuccess(t *testing.T) {
	for _, correct := range []bool{false, true} {
		t.Run(fmt.Sprint(correct), func(t *testing.T) {
			e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult {
				if call.ID == "bad" {
					return data.ToolResult{Error: &data.ToolError{Code: data.ToolErrorInvalidArgument}, QueryID: "fake", Scope: &data.QueryScope{}, References: []data.TransactionReference{{ID: "fake"}}, HasMore: true, NextCursor: "fake"}
				}
				return runnerSuccess(call.ID)
			}}
			steps := []func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error){runnerTurn(runnerCall("bad", `{"date":"wrong"}`))}
			if correct {
				steps = append(steps, func(in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
					if in.Messages[2].ToolResult.OK || in.Messages[2].ToolResult.Error == nil {
						t.Fatal("error became empty success")
					}
					return runnerTurn(runnerCall("good", `{}`))(in, nil)
				})
			}
			steps = append(steps, runnerFinal)
			r, err := NewAssistantRunner(runnerScript(t, steps...), nil, e).Run(runnerContext(), runnerInput(), nil)
			if correct {
				if err != nil || len(r.Queries) != 1 {
					t.Fatalf("correction failed: %+v %v", r, err)
				}
			} else {
				runnerRequireCode(t, err, data.ToolErrorInvalidArgument)
				if r.Reply != "" || r.State != AssistantStateIncomplete {
					t.Fatal("hallucinated completion accepted")
				}
			}
			if len(r.References) != 0 || r.Continuation != nil {
				t.Fatal("failed evidence leaked")
			}
		})
	}
}

func TestAssistantRunnerWholeRoundBudgetAndIncompleteTurns(t *testing.T) {
	t.Run("whole round", func(t *testing.T) {
		var calls []data.ToolCall
		for i := 0; i < 25; i++ {
			calls = append(calls, runnerCall(fmt.Sprint(i), `{}`))
		}
		e := &runnerExecutor{}
		r, err := NewAssistantRunner(runnerScript(t, runnerTurn(calls...)), nil, e).Run(runnerContext(), runnerInput(), nil)
		runnerRequireCode(t, err, data.ToolErrorBudgetExhausted)
		if len(e.calls) != 0 || r.State != AssistantStateIncomplete {
			t.Fatal("partial round executed")
		}
	})
	t.Run("remaining quota", func(t *testing.T) {
		var calls []data.ToolCall
		for i := 0; i < 23; i++ {
			calls = append(calls, runnerCall(fmt.Sprint(i), fmt.Sprintf(`{"x":%d}`, i)))
		}
		e := &runnerExecutor{}
		_, err := NewAssistantRunner(runnerScript(t, runnerTurn(calls...), runnerTurn(runnerCall("24", `{}`), runnerCall("25", `{}`))), nil, e).Run(runnerContext(), runnerInput(), nil)
		runnerRequireCode(t, err, data.ToolErrorBudgetExhausted)
		if len(e.calls) != 23 {
			t.Fatal("partial remaining round executed")
		}
	})
	t.Run("failed partial", func(t *testing.T) {
		e := &runnerExecutor{}
		_, err := NewAssistantRunner(runnerScript(t, func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
			return data.TurnResponse{ToolCalls: []data.ToolCall{runnerCall("a", `{}`)}}, errors.New("secret")
		}), nil, e).Run(runnerContext(), runnerInput(), nil)
		runnerRequireCode(t, err, data.ToolErrorUpstreamFailed)
		if len(e.calls) != 0 {
			t.Fatal("failed turn executed")
		}
	})
	t.Run("incomplete", func(t *testing.T) {
		e := &runnerExecutor{}
		_, err := NewAssistantRunner(runnerScript(t, func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
			return data.TurnResponse{Content: "not final", ToolCalls: []data.ToolCall{runnerCall("a", `{}`)}}, nil
		}), nil, e).Run(runnerContext(), runnerInput(), nil)
		runnerRequireCode(t, err, data.ToolErrorProtocol)
		if len(e.calls) != 0 {
			t.Fatal("incomplete turn executed")
		}
	})
}

func TestAssistantRunnerCanonicalCacheAndDuplicateIDs(t *testing.T) {
	for _, sameID := range []bool{false, true} {
		t.Run(fmt.Sprint(sameID), func(t *testing.T) {
			secondID := "b"
			if sameID {
				secondID = "a"
			}
			e := &runnerExecutor{}
			m := runnerScript(t, runnerTurn(runnerCall("a", `{"b":2,"a":1}`)), runnerTurn(runnerCall(secondID, `{ "a":1, "b":2 }`)), runnerTurn(runnerCall("c", `{"a":1,"b":2}`)))
			var records []AssistantObservation
			in := runnerInput()
			in.Observe = func(o AssistantObservation) { records = append(records, o) }
			r, err := NewAssistantRunner(m, nil, e).Run(runnerContext(), in, nil)
			runnerRequireCode(t, err, data.ToolErrorBudgetExhausted)
			if len(e.calls) != 1 || len(r.Queries) != 1 {
				t.Fatal("cache reexecuted or evidence duplicated")
			}
			last := records[len(records)-1]
			if last.CacheHits != 1 || last.Executions != 1 || last.Reason != AssistantReasonNoProgress {
				t.Fatalf("bad cache telemetry %+v", last)
			}
		})
	}
	t.Run("same ID different args", func(t *testing.T) {
		e := &runnerExecutor{}
		m := runnerScript(t, runnerTurn(runnerCall("a", `{"x":1}`)), runnerTurn(runnerCall("a", `{"x":2}`), runnerCall("b", `{}`)))
		_, err := NewAssistantRunner(m, nil, e).Run(runnerContext(), runnerInput(), nil)
		runnerRequireCode(t, err, data.ToolErrorProtocol)
		if len(e.calls) != 1 {
			t.Fatal("mismatched ID round executed")
		}
	})
	t.Run("duplicate keys cannot reuse", func(t *testing.T) {
		e := &runnerExecutor{}
		m := runnerScript(t, runnerTurn(runnerCall("a", `{"x":1}`)), runnerTurn(runnerCall("b", `{"x":2,"x":1}`)), runnerFinal)
		_, err := NewAssistantRunner(m, nil, e).Run(runnerContext(), runnerInput(), nil)
		if err != nil || len(e.calls) != 2 {
			t.Fatalf("duplicate key aliased valid cache: %v", err)
		}
	})
	t.Run("repeat error", func(t *testing.T) {
		e := &runnerExecutor{execute: func(core.Context, data.ToolCall) data.ToolResult {
			return data.ToolResult{Error: &data.ToolError{Code: data.ToolErrorInvalidArgument}}
		}}
		m := runnerScript(t, runnerTurn(runnerCall("a", `{}`)), runnerTurn(runnerCall("b", `{}`)))
		_, err := NewAssistantRunner(m, nil, e).Run(runnerContext(), runnerInput(), nil)
		runnerRequireCode(t, err, data.ToolErrorBudgetExhausted)
		if len(e.calls) != 2 {
			t.Fatal("error retries missing")
		}
	})
}

func TestAssistantRunnerRequestScopedFallback(t *testing.T) {
	for _, thinking := range []bool{false, true} {
		t.Run(fmt.Sprint(thinking), func(t *testing.T) {
			e := &runnerExecutor{}
			backupCalls := 0
			primary := runnerScript(t, func(in data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
				if thinking {
					if err := emit(data.DisplayEvent{ThinkingDelta: "display summary"}); err != nil {
						return data.TurnResponse{}, err
					}
				}
				return data.TurnResponse{Completed: true, Content: "internal", ToolCalls: []data.ToolCall{runnerCall("a", `{}`)}, State: "private"}, nil
			}, func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
				return data.TurnResponse{}, errors.New("transport secret")
			})
			backup := runnerModel(func(_ core.Context, _ int64, in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
				backupCalls++
				if in.State != nil || len(in.Messages) != 3 || in.Messages[1].Content != "internal" || in.Messages[2].ToolCallID != "a" {
					t.Fatal("fallback lost neutral pair or copied private state")
				}
				return runnerFinal(in, nil)
			})
			r, err := NewAssistantRunner(primary, backup, e).Run(runnerContext(), runnerInput(), func(AssistantEvent) error { return nil })
			if thinking {
				runnerRequireCode(t, err, data.ToolErrorUpstreamFailed)
				if backupCalls != 0 || r.Thinking != "display summary" {
					t.Fatal("fallback after visible thinking")
				}
			} else if err != nil || backupCalls != 1 {
				t.Fatalf("status prevented fallback: %v", err)
			}
		})
	}
	t.Run("switch once", func(t *testing.T) {
		fail := func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
			return data.TurnResponse{}, errors.New("offline")
		}
		in := runnerInput()
		var final AssistantObservation
		in.Observe = func(o AssistantObservation) {
			if o.Kind == AssistantObservationRun {
				final = o
			}
		}
		_, err := NewAssistantRunner(runnerScript(t, fail), runnerScript(t, runnerTurn(runnerCall("a", `{}`)), fail), &runnerExecutor{}).Run(runnerContext(), in, nil)
		runnerRequireCode(t, err, data.ToolErrorUpstreamFailed)
		if final.Rounds != 3 {
			t.Fatalf("fallback rounds = %d", final.Rounds)
		}
	})
	for _, code := range []data.ToolErrorCode{data.ToolErrorBudgetExhausted, data.ToolErrorInvalidArgument, data.ToolErrorNotFoundOrNotAccessible, data.ToolErrorProviderUnsupported} {
		t.Run(string(code), func(t *testing.T) {
			m := runnerScript(t, func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
				return data.TurnResponse{}, &data.ToolError{Code: code}
			})
			_, err := NewAssistantRunner(m, runnerScript(t), &runnerExecutor{}).Run(runnerContext(), runnerInput(), nil)
			runnerRequireCode(t, err, code)
		})
	}
}

func TestAssistantRunnerBudgetsRetainEvidence(t *testing.T) {
	for _, kind := range []string{"upstream context", "round", "result", "neutral context"} {
		t.Run(kind, func(t *testing.T) {
			e := &runnerExecutor{execute: func(_ core.Context, call data.ToolCall) data.ToolResult {
				r := runnerSuccess(call.ID)
				r.HasMore = true
				r.NextCursor = "last-cursor"
				if kind == "result" {
					r.Data = json.RawMessage(`"` + strings.Repeat("x", AssistantMaxToolResultBytes) + `"`)
				}
				return r
			}}
			steps := []func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error){runnerTurn(runnerCall("a", `{}`))}
			in := runnerInput()
			if kind == "round" {
				in.MaxRounds = 1
			} else if kind == "neutral context" {
				steps[0] = func(in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
					return data.TurnResponse{Completed: true, Content: strings.Repeat("x", AssistantMaxModelRequestBytes), ToolCalls: []data.ToolCall{runnerCall("a", `{}`)}}, nil
				}
			} else if kind == "upstream context" {
				steps = append(steps, func(in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
					if len(in.Messages) != 3 || in.Messages[2].ToolCallID != "a" {
						t.Fatal("tool pair pruned")
					}
					return data.TurnResponse{}, &data.ToolError{Code: data.ToolErrorBudgetExhausted}
				})
			}
			r, err := NewAssistantRunner(runnerScript(t, steps...), runnerScript(t), e).Run(runnerContext(), in, nil)
			if err == nil || r.State != AssistantStateIncomplete || len(r.Queries) != 1 || r.Continuation.Cursor != "last-cursor" || r.Reply != "" {
				t.Fatalf("lost evidence or false completion %+v %v", r, err)
			}
		})
	}
	t.Run("initial request", func(t *testing.T) {
		in := runnerInput()
		in.Messages[0].Content = strings.Repeat("x", AssistantMaxModelRequestBytes)
		e := &runnerExecutor{}
		_, err := NewAssistantRunner(runnerScript(t), nil, e).Run(runnerContext(), in, nil)
		runnerRequireCode(t, err, data.ToolErrorBudgetExhausted)
	})
	t.Run("fallback consumes round", func(t *testing.T) {
		in := runnerInput()
		in.MaxRounds = 1
		m := runnerScript(t, func(data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
			return data.TurnResponse{}, errors.New("offline")
		})
		_, err := NewAssistantRunner(m, runnerScript(t), &runnerExecutor{}).Run(runnerContext(), in, nil)
		runnerRequireCode(t, err, data.ToolErrorBudgetExhausted)
	})
}

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
