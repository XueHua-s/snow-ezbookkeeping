package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider/openai"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// Deterministic protocol test, not a claim about a commercial model's semantic
// planning: real tools must preserve every explicit filter during topic discovery.
func TestAssistantCategoryDiscoveryOverHTTP(t *testing.T) {
	for _, mode := range []string{"topic", "explicit", "true-zero"} {
		t.Run(mode, func(t *testing.T) {
			t.Chdir("../..")
			query, engine := assistantServiceFixture(t)
			for i, name := range []string{"餐饮", "餐饮旧类", "食品", "饮料", "水果"} {
				assistantInsertEntity(t, engine, &models.TransactionCategory{CategoryId: int64(50 + i), Uid: 7, Name: name, Type: models.CATEGORY_TYPE_EXPENSE, Hidden: i < 2})
			}
			for i, amount := range []int64{2000, 3000, -500, 9900, 8800, 7700} {
				tx := &models.Transaction{TransactionId: int64(300 + i), Uid: 7, AccountId: 11, CategoryId: int64(52 + i%3), Type: models.TRANSACTION_DB_TYPE_EXPENSE, TransactionTime: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC).UnixMilli() + int64(i), Amount: amount}
				if i == 3 {
					tx.AccountId = 12
				}
				if i == 4 {
					tx.TransactionTime = time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC).UnixMilli()
				}
				assistantInsertEntity(t, engine, tx)
				if i != 5 {
					assistantInsertEntity(t, engine, &models.TransactionTagIndex{TagIndexId: int64(500 + i), Uid: 7, TransactionId: tx.TransactionId, TagId: 41})
				}
			}
			start, end := "2026-10-01T00:00:00Z", "2026-11-01T00:00:00Z"
			if mode == "true-zero" {
				start, end = "2027-10-01T00:00:00Z", "2027-11-01T00:00:00Z"
			}
			filter := map[string]any{"types": []string{"expense"}, "start_time": start, "end_time": end, "currency": "CNY", "account_ids": []string{"11"}, "tag_ids": []string{"41"}, "tag_match": "all", "amount_min": "-100.00", "amount_max": "1000.00"}
			aggregate := func(ids []string, group bool) json.RawMessage {
				copy := map[string]any{}
				for k, v := range filter {
					copy[k] = v
				}
				if ids != nil {
					copy["category_ids"] = ids
				}
				args := map[string]any{"filter": copy, "metrics": []string{"count", "sum"}}
				if group {
					args["group_by"] = []string{"category"}
				}
				return assistantJSON(args)
			}
			calls := []data.ToolCall{{ID: "names", Name: data.ToolListCategories, Arguments: json.RawMessage(`{"keyword":"餐饮","type":"expense"}`)}, {ID: "literal", Name: data.ToolAggregateTransactions, Arguments: aggregate([]string{"50", "51"}, false)}}
			if mode != "explicit" {
				calls = append(calls, data.ToolCall{ID: "discovery", Name: data.ToolAggregateTransactions, Arguments: aggregate(nil, true)})
			}
			if mode == "topic" {
				calls = append(calls, data.ToolCall{ID: "resolved", Name: data.ToolAggregateTransactions, Arguments: aggregate([]string{"52", "53", "54"}, false)})
			}
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				requests++
				if !data.AssistantRequestFits(raw, 16000, 128*1024) {
					t.Error("wire budget exceeded")
				}
				var body struct {
					Messages []struct {
						Role, Content string
						ToolCallID    string `json:"tool_call_id"`
					}
				}
				_ = json.Unmarshal(raw, &body)
				if requests == 1 {
					if !strings.Contains(body.Messages[0].Content, "Explicit category IDs and user-confirmed scopes") || !strings.Contains(body.Messages[0].Content, "do one category-grouped aggregate discovery") {
						t.Error("runtime topic policy missing")
					}
				}
				for _, message := range body.Messages {
					if message.Role != "tool" {
						continue
					}
					var result data.ToolResult
					_ = json.Unmarshal([]byte(message.Content), &result)
					if !result.OK {
						t.Error("tool failed", result.Error)
						continue
					}
					if message.ToolCallID == "names" {
						rows := expandWorkingTable(result.Data).([]any)
						if len(rows) != 2 {
							t.Error("literal lookup became semantic")
						}
						for _, row := range rows {
							if row.(map[string]any)["subtree_transaction_count"] != "0" {
								t.Error("unused category count changed")
							}
						}
						continue
					}
					scope := result.Scope
					if scope.StartTime == nil || *scope.StartTime != start || scope.EndTime == nil || *scope.EndTime != end || scope.Currency != "CNY" || !reflect.DeepEqual(scope.Types, []string{"expense"}) || !reflect.DeepEqual(scope.AccountIDs, []string{"11"}) || !reflect.DeepEqual(scope.TagIDs, []string{"41"}) || scope.TagMatch != "all" || scope.AmountMin == nil || *scope.AmountMin != "-100.00" || scope.AmountMax == nil || *scope.AmountMax != "1000.00" || scope.Keyword != "" {
						t.Error("discovery changed explicit filters", scope)
					}
					expected := "3"
					if message.ToolCallID == "literal" || mode == "true-zero" {
						expected = "0"
					}
					if result.MatchedCount != expected {
						t.Error("wrong matched records", message.ToolCallID, result.MatchedCount)
					}
					if message.ToolCallID == "literal" && !reflect.DeepEqual(scope.CategoryIDs, []string{"50", "51"}) {
						t.Error("explicit categories silently expanded")
					}
					if message.ToolCallID == "discovery" && len(scope.CategoryIDs) != 0 {
						t.Error("tentative category filter not removed")
					}
					if message.ToolCallID == "resolved" && !strings.Contains(string(result.Totals), "45.00") {
						t.Error("lost signed full totals", string(result.Totals))
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if requests <= len(calls) {
					call := calls[requests-1]
					_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"index": 0, "finish_reason": "tool_calls", "message": map[string]any{"role": "assistant", "tool_calls": []any{map[string]any{"id": call.ID, "type": "function", "function": map[string]any{"name": call.Name, "arguments": string(call.Arguments)}}}}}}})
					return
				}
				answer := "按食品、饮料和水果三个分类统计，所选范围支出 CNY 45.00。"
				if mode == "explicit" {
					answer = "你指定的两个分类合计 CNY 0.00；此口径不包含食品等其他分类。"
				}
				if mode == "true-zero" {
					answer = "所选范围没有匹配交易。"
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"index": 0, "finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": answer}}}})
			}))
			defer server.Close()
			p := openai.NewOpenAICompatibleLargeLanguageModelProvider(&settings.LLMConfig{OpenAICompatibleBaseURL: server.URL, OpenAICompatibleModelID: "fixture", LargeLanguageModelAPIRequestTimeout: 5000, LargeLanguageModelAPIProxy: "none"}, false).(provider.LargeLanguageModelToolProvider)
			question := fmt.Sprintf("%s至%s餐饮开支，CNY，账户11，标签41，金额-100到1000", start, end)
			if mode == "explicit" {
				question += "；只统计分类50和51，两者合计"
			}
			messages, err := llm.BuildAssistantMessages("chat", question, nil, time.Now(), "zh_CN", nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			messages[0].Content += "\n" + assistantCompactToolRules
			result, err := llm.NewAssistantRunner(p, nil, assistantContextTools{AssistantQueryService: query}).Run(core.NewNullContext(), llm.AssistantRunInput{UID: 7, Messages: messages, MaxInputTokens: 16000, MaxRequestBytes: 128 * 1024, MaxOutputTokens: 4096, MaxRounds: 8}, nil)
			if err != nil || result.State != llm.AssistantStateCompleted || requests != len(calls)+1 || len(result.Queries) != len(calls) {
				t.Fatal(result.State, err, requests, len(result.Queries))
			}
		})
	}
}
