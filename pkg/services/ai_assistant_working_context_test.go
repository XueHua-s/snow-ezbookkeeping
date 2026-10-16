package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

// The first user message must succeed even when a full-ledger aggregation spans
// thousands of rows and many groups. Real SQLite and actual HTTP serialization
// exercise both the page admission decision and the following model request.
func TestAssistantCrossYearWorkingContextOverHTTP(t *testing.T) {
	t.Chdir("../..")
	query, engine := assistantServiceFixture(t)
	for i := 0; i < 56; i++ {
		_, err := engine.Insert(&models.TransactionCategory{CategoryId: int64(2000 + i), Uid: 7, Type: models.CATEGORY_TYPE_EXPENSE, Name: fmt.Sprintf("跨年生活消费类别%02d", i)})
		if err != nil {
			t.Fatal(err)
		}
	}
	transactions := make([]models.Transaction, 9183)
	for i := range transactions {
		transactions[i] = models.Transaction{TransactionId: int64(10000 + i), Uid: 7, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: 11, CategoryId: int64(2000 + i%56), TransactionTime: time.Date(2024+i%2, 1, 1, 0, 0, 0, 0, time.UTC).Unix()*1000 + int64(i), Amount: 101}
	}
	for start := 0; start < len(transactions); start += 100 {
		batch := transactions[start:min(start+100, len(transactions))]
		if _, err := engine.Insert(&batch); err != nil {
			t.Fatal(err)
		}
	}
	arguments := `{"filter":{"types":["expense"],"start_time":"2024-01-01T00:00:00Z","end_time":"2026-01-01T00:00:00Z"},"group_by":["month","category","account"],"metrics":["count","sum","min","max"],"page_size":200}`
	tools := assistantContextTools{AssistantQueryService: query}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		requests++
		if !data.AssistantRequestFits(raw, 16000, 128*1024) {
			t.Error("wire budget exceeded")
		}
		w.Header().Set("Content-Type", "application/json")
		if requests == 1 {
			call := map[string]any{"id": "aggregate", "type": "function", "function": map[string]any{"name": "aggregate_transactions", "arguments": arguments}}
			_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"index": 0, "finish_reason": "tool_calls", "message": map[string]any{"role": "assistant", "tool_calls": []any{call}}}}})
			return
		}
		var body struct {
			Messages []struct{ Role, Content string }
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Error(err)
		}
		found := false
		for _, message := range body.Messages {
			if message.Role != "tool" {
				continue
			}
			found = true
			var result data.ToolResult
			if err := json.Unmarshal([]byte(message.Content), &result); err != nil {
				t.Error(err)
			}
			if !result.OK || result.MatchedCount != "9183" || result.TotalsCoverage != data.QueryCoverageAllMatching || !strings.Contains(string(result.Totals), "9274.83") {
				t.Errorf("lost exact full totals: %+v", result)
			}
			if result.HasMore && (result.NextCursor == "" || result.Coverage != data.QueryCoveragePage) {
				t.Error("truncated groups presented as complete")
			}
			if result.ReturnedCount > 2 && !strings.Contains(string(result.Data), "column_rows") {
				t.Error("redundant records were not compacted")
			}
		}
		if !found {
			t.Error("lost tool result")
		}
		if requests == 2 {
			smallArguments := `{"filter":{"types":["expense"],"currency":"CNY","start_time":"2024-01-01T00:00:00Z","end_time":"2026-01-01T00:00:00Z"},"metrics":["sum","count"]}`
			call := map[string]any{"id": "verify_total", "type": "function", "function": map[string]any{"name": "aggregate_transactions", "arguments": smallArguments}}
			_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"index": 0, "finish_reason": "tool_calls", "message": map[string]any{"role": "assistant", "tool_calls": []any{call}}}}})
			return
		}
		if requests == 3 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":{"code":"context_length_exceeded"}}`)
			return
		}
		_, _ = io.WriteString(w, `{"choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"全量支出 CNY 9274.83。"}}]}`)
	}))
	defer server.Close()
	p := openai.NewOpenAICompatibleLargeLanguageModelProvider(&settings.LLMConfig{OpenAICompatibleBaseURL: server.URL, OpenAICompatibleModelID: "fixture", LargeLanguageModelAPIRequestTimeout: 5000, LargeLanguageModelAPIProxy: "none"}, false).(provider.LargeLanguageModelToolProvider)
	messages, err := llm.BuildAssistantMessages("chat", "请分析2024到2025消费", nil, time.Now(), "zh_CN", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	messages[0].Content += "\n" + assistantCompactToolRules
	result, err := llm.NewAssistantRunner(p, nil, tools).Run(core.NewNullContext(), llm.AssistantRunInput{UID: 7, Mode: "chat", Messages: messages, MaxInputTokens: 16000, MaxOutputTokens: 4096, MaxRounds: 5}, nil)
	if err != nil || result.State != llm.AssistantStateCompleted || requests != 4 || len(result.Queries) != 2 {
		t.Fatal(result, err, requests)
	}
}

func TestAssistantPageAdmissionMeasuresFinalMetadata(t *testing.T) {
	query, _ := assistantServiceFixture(t)
	var admitted []data.ToolResult
	query.resultFits = func(result data.ToolResult) bool { admitted = append(admitted, result); return true }
	result := assistantCall(query, data.ToolAggregateTransactions, `{"group_by":["category"],"page_size":1}`)
	if !result.OK || len(admitted) < 2 {
		t.Fatal(result, len(admitted))
	}
	for _, candidate := range admitted {
		if candidate.QueryID != result.QueryID || candidate.QueriedAt != result.QueriedAt {
			t.Fatal("admission used fictitious metadata", candidate.QueryID, result.QueryID)
		}
	}
	final, _ := json.Marshal(result)
	lastPage, _ := json.Marshal(admitted[len(admitted)-2])
	if string(final) != string(lastPage) {
		t.Fatal("accepted final page differs from execution envelope", string(lastPage), string(final))
	}
}
