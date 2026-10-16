package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantQueryValidationCorrectionsPreserveScope(t *testing.T) {
	for _, tc := range []struct {
		name, bad, good, field string
		completed              bool
	}{
		{"currency", `{"filter":{"currency":"BAD!","account_ids":["11"]}}`, `{"filter":{"currency":"CNY","account_ids":["11"]}}`, "filter.currency", true},
		{"time", `{"filter":{"start_time":"wrong","account_ids":["11"]}}`, `{"filter":{"start_time":"2020-01-01T00:00:00Z","account_ids":["11"]}}`, "filter.start_time", true},
		{"schema", `{"filter":{"types":["invalid"],"account_ids":["11"]}}`, `{"filter":{"types":["expense"],"account_ids":["11"]}}`, "arguments.filter.types", true},
		{"amount", `{"filter":{"currency":"CNY","amount_min":"oops","account_ids":["11"]}}`, `{"filter":{"currency":"CNY","amount_min":"0","account_ids":["11"]}}`, "filter.amount_min", true},
		{"ids", `{"filter":{"account_ids":["invalid"],"currency":"CNY"}}`, `{"filter":{"account_ids":["11"],"currency":"CNY"}}`, "filter.account_ids", true},
		{"changed-scope", `{"filter":{"currency":"BAD!","account_ids":["11"]}}`, `{"filter":{"currency":"CNY"}}`, "filter.currency", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, _ := assistantServiceFixture(t)
			requests := 0
			p := compactionTimeoutProvider(func(_ core.Context, input data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
				requests++
				if requests == 2 {
					result := input.Messages[len(input.Messages)-1].ToolResult
					if result == nil || result.Error == nil || result.Error.Field != tc.field {
						t.Fatalf("imprecise validation: %+v", result)
					}
				}
				if requests <= 2 {
					args, id := tc.bad, "bad"
					if requests == 2 {
						args, id = tc.good, "good"
					}
					return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{{ID: id, Name: data.ToolQueryTransactions, Arguments: json.RawMessage(args)}}}, nil
				}
				result := input.Messages[len(input.Messages)-1].ToolResult
				if result == nil || !result.OK {
					t.Fatalf("corrected query did not execute: %+v", result)
				}
				if emit != nil {
					if err := emit(data.DisplayEvent{ReplyDelta: "answer"}); err != nil {
						return data.TurnResponse{}, err
					}
				}
				return data.TurnResponse{Completed: true, Content: "answer"}, nil
			})
			var deltas string
			result, err := llm.NewAssistantRunner(p, nil, assistantContextTools{AssistantQueryService: query}).Run(core.NewNullContext(), llm.AssistantRunInput{UID: 7, Messages: []data.ConversationMessage{{Role: data.ConversationRoleUser, Content: "query account 11"}}, MaxInputTokens: 16000, MaxRequestBytes: 128 * 1024, MaxOutputTokens: 4096, MaxRounds: 4}, func(event llm.AssistantEvent) error {
				if event.Type == llm.AssistantEventReplyDelta {
					deltas += event.Delta
				}
				return nil
			})
			if tc.completed {
				if err != nil || result.State != llm.AssistantStateCompleted || result.Reply != "answer" {
					t.Fatalf("correction failed: %+v %v", result, err)
				}
			} else {
				var controlled *data.ToolError
				if !errors.As(err, &controlled) || controlled.Code != data.ToolErrorInvalidArgument || result.Reply != "" || deltas != "" {
					t.Fatalf("scope change hid failure: %+v %v %q", result, err, deltas)
				}
			}
		})
	}
}

func TestAssistantWorkerRetainsInaccessibleQueryFailure(t *testing.T) {
	worker, _ := assistantRunFixture(t, settings.Sqlite3DbType)
	query, _ := assistantServiceFixture(t)
	requests := 0
	p := compactionTimeoutProvider(func(_ core.Context, _ data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
		requests++
		if requests == 1 {
			return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{
				{ID: "missing", Name: data.ToolGetTransaction, Arguments: json.RawMessage(`{"transaction_id":"999"}`)},
				{ID: "available", Name: data.ToolGetTransaction, Arguments: json.RawMessage(`{"transaction_id":"100"}`)},
			}}, nil
		}
		if emit != nil {
			if err := emit(data.DisplayEvent{ReplyDelta: "unverified answer"}); err != nil {
				return data.TurnResponse{}, err
			}
		}
		return data.TurnResponse{Completed: true, Content: "unverified answer"}, nil
	})
	worker.options.Prepare = func(_ core.Context, run *models.AssistantMessage) (*AssistantPreparedRun, error) {
		return &AssistantPreparedRun{Runner: llm.NewAssistantRunner(p, nil, assistantContextTools{AssistantQueryService: query}), Input: llm.AssistantRunInput{UID: 7, Messages: []data.ConversationMessage{{Role: data.ConversationRoleUser, Content: "get both transactions"}}, MaxInputTokens: 16000, MaxRequestBytes: 128 * 1024, MaxOutputTokens: 4096, MaxRounds: 4}, Dependencies: models.AssistantInputDependencies{DeletionEpoch: run.DeletionEpoch}}, nil
	}
	run, err := worker.claim(core.NewNullContext(), 7, 3)
	if err != nil {
		t.Fatal(err)
	}
	worker.execute(context.Background(), run)
	// A query's not_accessible reason is snapshot data, not a failed snapshot
	// lookup: the current conversation remains readable and can be retried.
	snapshot, err := worker.GetRunSnapshot(core.NewNullContext(), 7, "3")
	if err != nil || snapshot.ReasonCode == nil || *snapshot.ReasonCode != models.AssistantReasonNotAccessible || snapshot.State != models.AssistantRunIncomplete || snapshot.Content != "" {
		t.Fatalf("wrong persisted query failure: %+v %v", snapshot, err)
	}
}
