package services

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider/openai"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// Exercise real tool results through the wire adapter: a text-only greeting does
// not demonstrate that the default context has room to answer ledger questions.
func TestAssistantDefaultContextSupportsLedgerSummary(t *testing.T) {
	t.Chdir("../..")
	query, _ := assistantServiceFixture(t)
	calls := []data.ToolCall{
		{ID: "accounts", Name: data.ToolListAccounts, Arguments: json.RawMessage(`{}`)},
		{ID: "categories", Name: data.ToolListCategories, Arguments: json.RawMessage(`{}`)},
		{ID: "totals", Name: data.ToolAggregateTransactions, Arguments: json.RawMessage(`{"filter":{"types":["expense"]}}`)},
		{ID: "distribution", Name: data.ToolAggregateTransactions, Arguments: json.RawMessage(`{"filter":{"types":["expense"]},"group_by":["category"]}`)},
		{ID: "balances", Name: data.ToolGetAccountBalances, Arguments: json.RawMessage(`{}`)},
	}
	turns := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		t.Logf("turn %d wire bytes %d", turns+1, len(raw))
		var input struct {
			Messages []struct {
				Role       string `json:"role"`
				Content    string `json:"content"`
				ToolCallID string `json:"tool_call_id"`
			} `json:"messages"`
		}
		if err := json.Unmarshal(raw, &input); err != nil {
			t.Error(err)
		}
		for _, m := range input.Messages {
			if m.Role == "tool" {
				var result data.ToolResult
				if err := json.Unmarshal([]byte(m.Content), &result); err != nil || !result.OK {
					t.Errorf("query result unavailable to model: %s", m.Content)
				}
				if m.ToolCallID == "totals" || m.ToolCallID == "distribution" {
					if result.TotalsCoverage != data.QueryCoverageAllMatching || result.MatchedCount != "2" {
						t.Errorf("full ledger evidence lost: %+v", result)
					}
					var rows []map[string]any
					if json.Unmarshal(result.Totals, &rows) != nil || len(rows) != 1 || rows[0]["flow"] != "expense" || rows[0]["sum"] != "80.00" {
						t.Errorf("signed refund total lost: %s", result.Totals)
					}
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if turns < len(calls) {
			call := calls[turns]
			turns++
			response := map[string]any{"choices": []any{map[string]any{"index": 0, "finish_reason": "tool_calls", "message": map[string]any{"role": "assistant", "tool_calls": []any{map[string]any{"id": call.ID, "type": "function", "function": map[string]any{"name": call.Name, "arguments": string(call.Arguments)}}}}}}}
			_ = json.NewEncoder(w).Encode(response)
		} else {
			turns++
			_, _ = io.WriteString(w, `{"choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"Expenses CNY 80.00, current balance CNY 2000.00."}}]}`)
		}
	}))
	defer server.Close()
	p := openai.NewOpenAICompatibleLargeLanguageModelProvider(&settings.LLMConfig{OpenAICompatibleBaseURL: server.URL, OpenAICompatibleModelID: "fixture", LargeLanguageModelAPIRequestTimeout: 2000, LargeLanguageModelAPIProxy: "none"}, false).(provider.LargeLanguageModelToolProvider)
	messages, err := llm.BuildAssistantMessages("chat", "请查一下我的支出分类和余额", nil, time.Now(), "zh-CN", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	messages[0].Content += "\n" + assistantCompactToolRules
	definitions, _ := json.Marshal((assistantContextTools{AssistantQueryService: query}).Definitions())
	t.Logf("system prompt bytes %d, tool definitions bytes %d", len(messages[0].Content), len(definitions))
	limits := settings.DefaultAIAssistantExperienceConfig()
	runner := llm.NewAssistantRunner(p, nil, assistantContextTools{AssistantQueryService: query})
	result, err := runner.Run(core.NewNullContext(), llm.AssistantRunInput{UID: 7, Mode: "chat", Messages: messages, MaxInputTokens: assistantContextBudget(limits), MaxOutputTokens: limits.Limits.OutputTokens}, nil)
	if err != nil || result.State != llm.AssistantStateCompleted || turns != len(calls)+1 {
		t.Fatalf("ledger summary stopped after %d turns: %+v, %v", turns, result, err)
	}
}
