package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

type compactionTimeoutProvider func(core.Context, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error)

func (p compactionTimeoutProvider) RunToolTurn(c core.Context, _ int64, input data.ToolTurnInput, sink data.DisplayEventSink) (data.TurnResponse, error) {
	return p(c, input, sink)
}

func TestAssistantSummaryLocalTimeoutContinuesSameRunWithOriginalHistory(t *testing.T) {
	for _, kind := range []string{"summary-deadline", "provider-deadline"} {
		t.Run(kind, func(t *testing.T) {
			t.Chdir("../..")
			conv := assistantConversationFixture(t, settings.Sqlite3DbType)
			old := conversationSend(t, conv, "old", "", 0, "账户11，人民币，保留退款负数")
			historical := strings.Repeat("中", 9000)
			conversationComplete(t, conv, old, historical)
			current := conversationSend(t, conv, "current", old.ConversationId, old.ContextVersion, "继续查询")
			run := contextBudgetRun(t, conv, current)
			query, _ := assistantServiceFixture(t)
			requests := 0
			p := compactionTimeoutProvider(func(ctx core.Context, input data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
				requests++
				if len(input.Tools) == 0 {
					if kind == "summary-deadline" {
						<-ctx.Done()
						return data.TurnResponse{}, ctx.Err()
					}
					return data.TurnResponse{}, context.DeadlineExceeded
				}
				if requests == 2 {
					return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{{ID: "total", Name: data.ToolAggregateTransactions, Arguments: json.RawMessage(`{"filter":{"account_ids":["11"],"currency":"CNY","types":["expense"]}}`)}}}, nil
				}
				if emit == nil {
					t.Fatal("streaming sink missing")
				}
				if err := emit(data.DisplayEvent{ReplyDelta: "支出"}); err != nil {
					return data.TurnResponse{}, err
				}
				if err := emit(data.DisplayEvent{ReplyDelta: " 80.00"}); err != nil {
					return data.TurnResponse{}, err
				}
				return data.TurnResponse{Completed: true, Content: "支出 80.00"}, nil
			})
			s := contextBudgetService(conv, p)
			policy, err := s.acceptedPolicy(run)
			if err != nil {
				t.Fatal(err)
			}
			policy.Experience.SummaryTimeoutSeconds = 1
			raw, _ := json.Marshal(policy)
			run.RuntimeJSON = string(raw)
			before, err := s.snapshot(core.NewNullContext(), run)
			if err != nil {
				t.Fatal(err)
			}
			var events []assistantCompactionObservation
			s.observeCompaction = func(_ core.Context, event assistantCompactionObservation) { events = append(events, event) }
			prepared, err := s.Prepare(core.NewNullContext(), run)
			if err != nil {
				t.Fatal(err)
			}
			if prepared.Input.PriorModelCalls != 1 || requests != 1 {
				t.Fatal("failed summary call not charged", requests)
			}
			found := false
			for _, message := range prepared.Input.Messages {
				if message.Content == historical {
					found = true
				}
			}
			if !found {
				t.Fatal("original history discarded")
			}
			after, err := s.snapshot(core.NewNullContext(), run)
			if err != nil {
				t.Fatal(err)
			}
			if after.Conversation.Summary != "" || after.Conversation.SummaryThrough != 0 || !reflect.DeepEqual(before.Dependencies, after.Dependencies) {
				t.Fatal("failed summary mutated dependencies/history")
			}
			deltas := 0
			result, err := llm.NewAssistantRunner(p, nil, assistantContextTools{AssistantQueryService: query}).Run(core.NewNullContext(), prepared.Input, func(event llm.AssistantEvent) error {
				if event.Type == llm.AssistantEventReplyDelta {
					deltas++
				}
				return nil
			})
			if err != nil || result.State != llm.AssistantStateCompleted || result.Reply != "支出 80.00" || deltas != 2 || requests != 3 || len(result.Queries) != 1 {
				t.Fatal(result.State, err, deltas, requests)
			}
			source := "provider_timeout"
			if kind == "summary-deadline" {
				source = "summary_timeout"
			}
			failed, recovered := false, false
			for _, event := range events {
				failed = failed || event.FailureSource == source
				recovered = recovered || event.Recovery == "original_history"
				encoded, _ := json.Marshal(event)
				if strings.Contains(string(encoded), "账户") || strings.Contains(string(encoded), historical) {
					t.Fatal("observation leaked content")
				}
			}
			if !failed || !recovered {
				t.Fatal("missing layered recovery telemetry", events)
			}
		})
	}
}

func TestAssistantSummaryParentContextAlwaysWins(t *testing.T) {
	for _, mode := range []string{"cancel", "deadline"} {
		t.Run(mode, func(t *testing.T) {
			t.Chdir("../..")
			conv := assistantConversationFixture(t, settings.Sqlite3DbType)
			old := conversationSend(t, conv, "old", "", 0, "account 11")
			conversationComplete(t, conv, old, strings.Repeat("中", 9000))
			current := conversationSend(t, conv, "current", old.ConversationId, old.ContextVersion, "continue")
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "deadline" {
				parent, cancel = context.WithTimeout(context.Background(), 20*time.Millisecond)
				defer cancel()
			}
			calls := 0
			p := compactionTimeoutProvider(func(ctx core.Context, _ data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
				calls++
				if mode == "cancel" {
					cancel()
				}
				<-ctx.Done()
				return data.TurnResponse{}, context.DeadlineExceeded
			})
			_, err := contextBudgetService(conv, p).Prepare(core.NewAssistantContext(parent, "parent-test", "", ""), contextBudgetRun(t, conv, current))
			want := context.Canceled
			if mode == "deadline" {
				want = context.DeadlineExceeded
			}
			if !errors.Is(err, want) || calls > 1 {
				t.Fatal("parent context was recovered", err, calls)
			}
		})
	}
}

func TestAssistantSummaryHardBudgetRecoveryUsesOneSmallerWholePrefix(t *testing.T) {
	for _, success := range []bool{false, true} {
		t.Run(fmt.Sprint(success), func(t *testing.T) {
			conv := assistantConversationFixture(t, settings.Sqlite3DbType)
			var last models.AssistantCommandResult
			for i := 0; i < 3; i++ {
				last = conversationSend(t, conv, fmt.Sprint("old", i), last.ConversationId, last.ContextVersion, fmt.Sprintf("constraint-%d CNY account11", i))
				conversationComplete(t, conv, last, strings.Repeat("中", 5000))
			}
			current := conversationSend(t, conv, "current", last.ConversationId, last.ContextVersion, "latest exact question")
			run := contextBudgetRun(t, conv, current)
			var inputs []data.ToolTurnInput
			p := compactionTimeoutProvider(func(_ core.Context, input data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
				inputs = append(inputs, input)
				if len(inputs) == 1 || !success {
					return data.TurnResponse{}, context.DeadlineExceeded
				}
				return data.TurnResponse{Completed: true, Content: `{"topic":"older turn","confirmedConstraints":["constraint-0 CNY account11"],"resolvedReferences":[],"openQuestions":[]}`}, nil
			})
			s := contextBudgetService(conv, p)
			snap, err := s.snapshot(core.NewNullContext(), run)
			if err != nil {
				t.Fatal(err)
			}
			changed, calls, err := s.compact(core.NewNullContext(), run, snap, p, 16000, true)
			if calls != 2 || len(inputs) != 2 || len(inputs[1].Messages) >= len(inputs[0].Messages) {
				t.Fatal("unbounded or nonshrinking recovery", calls, inputs)
			}
			encoded, _ := json.Marshal(inputs[1].Messages)
			if !strings.Contains(string(encoded), "constraint-0") || strings.Contains(string(encoded), "constraint-1") || strings.Contains(string(encoded), "latest exact question") {
				t.Fatal("wrong prefix summarized")
			}
			fresh, e := s.snapshot(core.NewNullContext(), run)
			if e != nil {
				t.Fatal(e)
			}
			if success {
				if err != nil || !changed || fresh.Conversation.SummaryThrough != 1 || len(fresh.Turns) != 3 {
					t.Fatal("uncovered constraints lost", err, changed, fresh.Conversation.SummaryThrough)
				}
			} else if err != models.AssistantReasonProviderUnavailable || changed || fresh.Conversation.Summary != "" || fresh.Conversation.SummaryThrough != 0 {
				t.Fatal("failed recovery committed", err, changed)
			}
		})
	}
}
