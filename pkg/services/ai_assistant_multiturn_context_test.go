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

// Actual SQLite execution and HTTP wire accounting cover an extended analysis,
// not only one aggregate followed by an answer. All records are synthetic.
func TestAssistantMultiturnWorkingContextOverHTTP(t *testing.T) {
	for _, reject := range []bool{false, true} {
		t.Run(fmt.Sprint("upstream-reject=", reject), func(t *testing.T) { assistantMultiturnWorkingContextOverHTTP(t, reject) })
	}
}
func assistantMultiturnWorkingContextOverHTTP(t *testing.T, reject bool) {
	t.Chdir("../..")
	query, engine := assistantServiceFixture(t)
	for i := 0; i < 70; i++ {
		_, err := engine.Insert(&models.TransactionCategory{CategoryId: int64(2000 + i), Uid: 7, Type: models.CATEGORY_TYPE_EXPENSE, Name: fmt.Sprintf("生活消费类别%02d", i)})
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 700; i++ {
		account, amount := int64(11), int64(101)
		if i%7 == 0 {
			account = 12
		}
		if i%11 == 0 {
			amount = -37
		}
		_, err := engine.Insert(&models.Transaction{TransactionId: int64(10000 + i), Uid: 7, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: account, CategoryId: int64(2000 + i%70), TransactionTime: time.Date(2024+i%2, time.Month(1+i%12), 2, 0, 0, 0, 0, time.UTC).UnixMilli() + int64(i), Amount: amount, Comment: strings.Repeat("合成消费记录", 8)})
		if err != nil {
			t.Fatal(err)
		}
	}
	periods := `"periods":[{"name":"base","start_time":"2024-01-01T00:00:00Z","end_time":"2025-01-01T00:00:00Z"},{"name":"current","start_time":"2025-01-01T00:00:00Z","end_time":"2026-01-01T00:00:00Z"}],"comparison":{"baseline":"base","current":"current"}`
	calls := []data.ToolCall{
		{ID: "total", Name: data.ToolAggregateTransactions, Arguments: json.RawMessage(`{"metrics":["sum","count"]}`)},
		{ID: "period", Name: data.ToolAggregateTransactions, Arguments: json.RawMessage(`{` + periods + `}`)},
		{ID: "category", Name: data.ToolAggregateTransactions, Arguments: json.RawMessage(`{` + periods + `,"group_by":["category"],"page_size":200}`)},
		{ID: "month", Name: data.ToolAggregateTransactions, Arguments: json.RawMessage(`{` + periods + `,"group_by":["month","category"],"page_size":200}`)},
		{ID: "categories", Name: data.ToolListCategories, Arguments: json.RawMessage(`{"page_size":200}`)},
		{ID: "details", Name: data.ToolQueryTransactions, Arguments: json.RawMessage(`{"page_size":200}`)},
	}
	expected := map[string]data.ToolResult{}
	for _, call := range calls {
		result := query.Execute(core.NewNullContext(), call)
		if !result.OK {
			t.Fatal(call.ID, result.Error)
		}
		expected[call.ID] = result
	}
	// Independent arithmetic: 600 CNY rows (54 refunds) plus the fixture's
	// net CNY expense 80; 100 USD rows (10 refunds); one cross-currency transfer.
	if expected["total"].MatchedCount != "704" {
		t.Fatal("logical transfer counted twice", expected["total"].MatchedCount)
	}
	for _, amount := range []string{"611.48", "87.20", "40.00", "5.00"} {
		if !strings.Contains(string(expected["total"].Totals), amount) {
			t.Fatal("signed/currency total mismatch", amount, string(expected["total"].Totals))
		}
	}
	requests := 0
	seenCheckpoint := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var envelope struct{ Tools []json.RawMessage }
		_ = json.Unmarshal(raw, &envelope)
		if len(envelope.Tools) == 0 {
			w.Header().Set("Content-Type", "application/json")
			answer := `{"topic":"2022/2023消费比较","confirmedConstraints":["按币种分开","比较退款与转账"],"resolvedReferences":[],"openQuestions":["继续查询并细分原因"]}`
			_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"index": 0, "finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": answer}}}})
			return
		}
		requests++
		if !data.AssistantRequestFits(raw, 16000, 128*1024) {
			t.Error("actual HTTP request exceeds budget")
		}
		var body struct {
			Messages []struct {
				Role, Content string
				ToolCallID    string `json:"tool_call_id"`
			}
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Error(err)
		}
		for _, message := range body.Messages {
			if message.Role != "tool" {
				continue
			}
			var result data.ToolResult
			_ = json.Unmarshal([]byte(message.Content), &result)
			want := expected[message.ToolCallID]
			if !result.OK || result.MatchedCount != want.MatchedCount || result.TotalsCoverage != want.TotalsCoverage || !reflect.DeepEqual(expandWorkingTable(result.Totals), expandWorkingTable(want.Totals)) {
				t.Errorf("lost full-range totals for %s: %s vs %s", message.ToolCallID, result.Totals, want.Totals)
			}
			if strings.Contains(string(result.Data), "archived_checkpoint") {
				seenCheckpoint = true
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if reject && requests == 5 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":{"code":"context_length_exceeded"}}`)
			return
		}
		step := requests
		if reject && requests > 5 {
			step--
		}
		if step <= len(calls) {
			call := calls[step-1]
			_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"index": 0, "finish_reason": "tool_calls", "message": map[string]any{"role": "assistant", "tool_calls": []any{map[string]any{"id": call.ID, "type": "function", "function": map[string]any{"name": call.Name, "arguments": string(call.Arguments)}}}}}}})
			return
		}
		_, _ = io.WriteString(w, `{"choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"已按币种完整核对，退款保留负数，转账成对统计。"}}]}`)
	}))
	defer server.Close()
	p := openai.NewOpenAICompatibleLargeLanguageModelProvider(&settings.LLMConfig{OpenAICompatibleBaseURL: server.URL, OpenAICompatibleModelID: "fixture", LargeLanguageModelAPIRequestTimeout: 5000, LargeLanguageModelAPIProxy: "none"}, false).(provider.LargeLanguageModelToolProvider)
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	first := conversationSend(t, conv, "first", "", 0, "2022和2023比较，金额按币种分开")
	conversationComplete(t, conv, first, strings.Repeat("历史分析仅作上下文。", 600))
	second := conversationSend(t, conv, "second", first.ConversationId, first.ContextVersion, "继续比较退款与转账")
	conversationComplete(t, conv, second, strings.Repeat("历史回答不是最新余额。", 600))
	current := conversationSend(t, conv, "current", second.ConversationId, second.ContextVersion, "继续查询并细分原因")
	prepared, err := contextBudgetService(conv, p).Prepare(core.NewNullContext(), contextBudgetRun(t, conv, current))
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Input.PriorModelCalls == 0 {
		t.Fatal("long histories were not compacted before tool analysis")
	}
	var observation llm.AssistantObservation
	input := prepared.Input
	input.Observe = func(o llm.AssistantObservation) {
		if o.Kind == llm.AssistantObservationRun {
			observation = o
		}
	}
	// Conversation state and ledger are independent SQLite fixtures; the prepared
	// server-authoritative input is executed against the synthetic UID-bound ledger.
	result, err := llm.NewAssistantRunner(p, nil, assistantContextTools{AssistantQueryService: query}).Run(core.NewNullContext(), input, nil)
	wantRequests := 7
	if reject {
		wantRequests++
	}
	if err != nil || result.State != llm.AssistantStateCompleted || requests != wantRequests || len(result.Queries) != 6 {
		t.Fatal(result.State, err, requests, len(result.Queries), observation)
	}
	if reject && (!seenCheckpoint || observation.ContextRecoveryAttempts == 0 || observation.ContextRecoveryAttempts > 3) {
		t.Fatal("bounded recovery was not exercised", observation)
	}
}

func expandWorkingTable(raw json.RawMessage) any {
	var value any
	if len(raw) == 0 {
		return nil
	}
	_ = json.Unmarshal(raw, &value)
	var expand func(any) any
	expand = func(v any) any {
		switch x := v.(type) {
		case []any:
			for i := range x {
				x[i] = expand(x[i])
			}
			return x
		case map[string]any:
			if x["format"] == "column_rows" {
				columns := x["columns"].([]any)
				rows := x["rows"].([]any)
				out := make([]any, len(rows))
				for i, row := range rows {
					record := map[string]any{}
					for j, cell := range row.([]any) {
						record[columns[j].(string)] = expand(cell)
					}
					out[i] = record
				}
				return out
			}
			for key, child := range x {
				x[key] = expand(child)
			}
			return x
		default:
			return v
		}
	}
	return expand(value)
}
