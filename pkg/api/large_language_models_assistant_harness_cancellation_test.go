package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider/openai"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAssistantHarnessHTTPProviderContextAndCancellation(t *testing.T) {
	a, _ := harnessAPIFixture(t)
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			started, stopped := make(chan struct{}), make(chan struct{})
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(stopped)
				assert.Equal(t, "/chat/completions", r.URL.Path)
				var body struct {
					Messages []struct{ Role, Content string } `json:"messages"`
				}
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, "system", body.Messages[0].Role)
				assert.Contains(t, body.Messages[0].Content, "2026-03-15T12:00:00+08:00")
				close(started)
				<-r.Context().Done()
			}))
			defer upstream.Close()
			p := openai.NewOpenAICompatibleLargeLanguageModelProvider(&settings.LLMConfig{OpenAICompatibleBaseURL: upstream.URL, OpenAICompatibleModelID: "mock-api-test", LargeLanguageModelAPIRequestTimeout: 5000, LargeLanguageModelAPIProxy: "none"}, false)
			a.assistant.primary = p.(provider.LargeLanguageModelToolProvider)
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			c := harnessContext(`{"message":"仅测试取消"}`, 7, parent, httptest.NewRecorder())
			finished := make(chan struct{})
			go func() {
				defer close(finished)
				if stream {
					a.AssistantChatStreamHandler(c)
				} else {
					a.AssistantChatHandler(c)
				}
			}()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("HTTP upstream never received request")
			}
			cancel()
			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Fatal("API did not join upstream")
			}
			select {
			case <-stopped:
			case <-time.After(time.Second):
				t.Fatal("HTTP request was not canceled")
			}
		})
	}
}

func TestAssistantHarnessDatabaseCancellation(t *testing.T) {
	a, e := harnessAPIFixture(t)
	harnessInsert(t, e, 100, "2026-03-01T00:00:00+08:00", 10000, 30)
	// Replace ONLY the temporary fixture's transaction table with a computationally
	// expensive view. Actual service SQL now runs in sqlite3, rather than a mock
	// executor sleeping and pretending that database cancellation worked.
	_, err := e.Exec(`ALTER TABLE "transaction" RENAME TO "assistant_seed"`)
	require.NoError(t, err)
	_, err = e.Exec(`CREATE VIEW "transaction" AS WITH RECURSIVE spin(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM spin WHERE n<1000000000) SELECT assistant_seed.* FROM assistant_seed,spin WHERE spin.n=1000000000`)
	require.NoError(t, err)
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			started := make(chan struct{})
			var rounds atomic.Int32
			var observations []llm.AssistantObservation
			a.assistant.observe = func(o llm.AssistantObservation) { observations = append(observations, o) }
			a.assistant.primary = harnessProvider(func(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
				rounds.Add(1)
				close(started)
				return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{harnessCall("expensive", data.ToolAggregateTransactions, "{}")}}, nil
			})
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			c := harnessContext(`{"message":"查询全历史"}`, 7, parent, httptest.NewRecorder())
			finished := make(chan struct{})
			go func() {
				defer close(finished)
				if stream {
					a.AssistantChatStreamHandler(c)
				} else {
					a.AssistantChatHandler(c)
				}
			}()
			<-started
			// Give the worker time to enter the recursive SQL; the 1-billion-row view
			// cannot finish naturally within this interval.
			time.Sleep(30 * time.Millisecond)
			require.Greater(t, e.DB().DB.Stats().InUse, 0)
			cancel()
			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Fatal("driver did not cancel SQL")
			}
			require.Eventually(t, func() bool { return e.DB().DB.Stats().InUse == 0 }, time.Second, time.Millisecond)
			assert.Equal(t, int32(1), rounds.Load())
			require.NotEmpty(t, observations)
			assert.Equal(t, llm.AssistantReasonCanceled, observations[len(observations)-1].Reason)
			var one int
			require.NoError(t, e.DB().DB.QueryRow("SELECT 1").Scan(&one))
			assert.Equal(t, 1, one)
		})
	}
}

func TestAssistantHarnessEvidenceSurvivesFailureAndTerminalRace(t *testing.T) {
	a, e := harnessAPIFixture(t)
	harnessInsert(t, e, 100, "2026-03-01T00:00:00+08:00", 10000, 30)
	harnessInsert(t, e, 101, "2026-03-02T00:00:00+08:00", 20000, 30)
	for _, stream := range []bool{false, true} {
		rounds := 0
		a.assistant.primary = harnessProvider(func(_ core.Context, _ int64, in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
			rounds++
			if rounds == 1 {
				return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{harnessCall("first", data.ToolQueryTransactions, `{"page_size":1}`)}}, nil
			}
			return data.TurnResponse{}, &data.ToolError{Code: data.ToolErrorBudgetExhausted}
		})
		rec, err := harnessRequest(t, a, `{"message":"查询"}`, stream)
		require.NotNil(t, err)
		assert.Contains(t, rec.Body.String(), `"state":"incomplete"`)
		assert.Contains(t, rec.Body.String(), `"queryId"`)
		assert.Contains(t, rec.Body.String(), `"continuation"`)
		assert.NotContains(t, rec.Body.String(), `"type":"done"`)
		assert.Equal(t, 1, strings.Count(rec.Body.String(), `"success":false`))
	}
	// Repeatedly race completion and client cancellation. The joined runner must
	// emit one observation and the writer may emit at most one terminal frame.
	for n := 0; n < 20; n++ {
		parent, cancel := context.WithCancel(context.Background())
		release := make(chan struct{})
		var count atomic.Int32
		a.assistant.observe = func(o llm.AssistantObservation) {
			if o.Kind == llm.AssistantObservationRun {
				count.Add(1)
			}
		}
		a.assistant.primary = harnessProvider(func(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
			<-release
			return data.TurnResponse{Completed: true, Content: "答案"}, nil
		})
		rec := httptest.NewRecorder()
		c := harnessContext(`{"message":"完成竞态"}`, 7, parent, rec)
		finished := make(chan struct{})
		go func() {
			defer close(finished)
			if err := a.AssistantChatStreamHandler(c); err != nil && parent.Err() == nil {
				utils.WriteEventStreamJsonErrorResult(c, err)
			}
		}()
		close(release)
		cancel()
		select {
		case <-finished:
		case <-time.After(time.Second):
			t.Fatal("terminal race leaked worker")
		}
		assert.LessOrEqual(t, strings.Count(rec.Body.String(), `"type":"done"`)+strings.Count(rec.Body.String(), `"success":false`), 1)
		// Cancellation can precede authentication, in which case no run started.
		assert.LessOrEqual(t, count.Load(), int32(1))
	}
}

func TestAssistantHarnessCashFlowBalanceAndTimezoneEvidence(t *testing.T) {
	a, e := harnessAPIFixture(t)
	harnessInsert(t, e, 100, "2026-03-01T00:30:00+08:00", 10000, 30)
	harnessInsert(t, e, 101, "2026-03-03T00:00:00+08:00", -2000, 30)
	harnessInsert(t, e, 102, "2026-03-04T00:00:00+08:00", 3000, 30)
	at, _ := time.Parse(time.RFC3339, "2026-03-02T00:00:00+08:00")
	_, err := e.Insert(&models.Transaction{TransactionId: 103, Uid: 7, Type: models.TRANSACTION_DB_TYPE_INCOME, AccountId: 10, CategoryId: 30, TransactionTime: at.Unix() * 1000, Amount: 100000})
	require.NoError(t, err)
	_, err = e.Exec(`UPDATE "transaction" SET timezone_utc_offset=0 WHERE transaction_id=100`)
	require.NoError(t, err)
	rounds := 0
	a.assistant.primary = harnessProvider(func(_ core.Context, _ int64, in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
		rounds++
		if rounds == 1 {
			assert.Contains(t, in.Messages[0].Content, "NOT an account balance")
			assert.Contains(t, in.Messages[0].Content, "transaction's recorded timezone")
			return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{harnessCall("flows", data.ToolAggregateTransactions, `{"filter":{"start_time":"2026-03-01T00:00:00+08:00","end_time":"2026-03-15T12:00:00+08:00"}}`), harnessCall("balances", data.ToolGetAccountBalances, "{}")}}, nil
		}
		results := harnessCalls(in)
		assert.Contains(t, string(results[0].Totals), `"1000.00"`)
		assert.Contains(t, string(results[0].Totals), `"110.00"`)
		assert.Contains(t, string(results[1].Totals), `"1700.00"`)
		assert.Equal(t, "Asia/Shanghai", results[0].Scope.Timezone)
		return data.TurnResponse{Completed: true, Content: "本期收支净额 CNY 890.00；当前净资产 CNY 1700.00。查询按 Asia/Shanghai 的 3 月范围分组，页面按交易记录时区分组时，这笔 UTC 2 月 28 日的账单可能落入不同月份。"}, nil
	})
	rec, apiErr := harnessRequest(t, a, `{"message":"净额是否余额？为何页面月支出不同？"}`, true)
	require.Nil(t, apiErr)
	assert.Contains(t, rec.Body.String(), "890.00")
	assert.Contains(t, rec.Body.String(), "1700.00")
	assert.Contains(t, rec.Body.String(), `"timezone":"Asia/Shanghai"`)
}

func TestAssistantHarnessTerminalWriteFailureIsObserved(t *testing.T) {
	a, _ := harnessAPIFixture(t)
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			var observations []llm.AssistantObservation
			a.assistant.observe = func(o llm.AssistantObservation) { observations = append(observations, o) }
			a.assistant.primary = harnessProvider(func(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
				if fail {
					return data.TurnResponse{}, errors.New("upstream failed")
				}
				return data.TurnResponse{Completed: true, Content: "回答"}, nil
			})
			rec := httptest.NewRecorder()
			writer := &harnessTerminalFailWriter{ResponseWriter: rec}
			c := harnessContext(`{"message":"终态写失败"}`, 7, context.Background(), writer)
			require.Nil(t, a.AssistantChatStreamHandler(c))
			assert.Equal(t, 1, writer.terminals)
			require.Len(t, observations, 1)
			assert.Equal(t, llm.AssistantReasonWriteFailed, observations[0].Reason)
			assert.NotContains(t, rec.Body.String(), `"type":"done"`)
		})
	}
}

type harnessTerminalFailWriter struct {
	http.ResponseWriter
	terminals int
}
