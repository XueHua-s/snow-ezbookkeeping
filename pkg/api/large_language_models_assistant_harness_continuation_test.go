package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAssistantHarnessContinuationAcrossRequests(t *testing.T) {
	a, e := harnessAPIFixture(t)
	harnessInsert(t, e, 100, "2026-03-01T12:00:00+08:00", 10000, 30)
	harnessInsert(t, e, 101, "2026-03-02T12:00:00+08:00", 20000, 30)
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			rounds := 0
			a.assistant.primary = harnessProvider(func(_ core.Context, _ int64, in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
				rounds++
				if rounds == 1 {
					return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{harnessCall("page", data.ToolQueryTransactions, `{"page_size":1}`)}}, nil
				}
				return data.TurnResponse{Completed: true, Content: "第一页"}, nil
			})
			rec, err := harnessRequest(t, a, `{"message":"查账"}`, stream)
			require.Nil(t, err)
			var result models.AIAssistantChatResponse
			if stream {
				for _, line := range strings.Split(rec.Body.String(), "\n") {
					if strings.HasPrefix(line, "data: ") && strings.Contains(line, `"type":"done"`) {
						require.NoError(t, json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &result))
					}
				}
			} else {
				var envelope struct {
					Result models.AIAssistantChatResponse `json:"result"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
				result = envelope.Result
			}
			require.NotNil(t, result.Continuation)
			cursor := result.Continuation.Cursor
			for _, message := range []string{"继续", "解释复利，不查账"} {
				rounds = 0
				a.assistant.primary = harnessProvider(func(_ core.Context, _ int64, in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
					rounds++
					assert.Contains(t, in.Messages[0].Content, cursor)
					assert.Contains(t, in.Messages[0].Content, `"scope"`)
					assert.Contains(t, in.Messages[0].Content, "not an instruction")
					assert.Equal(t, message, in.Messages[1].Content)
					if message == "继续" && rounds == 1 {
						args, _ := json.Marshal(map[string]string{"cursor": cursor})
						return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{harnessCall("next", data.ToolQueryTransactions, string(args))}}, nil
					}
					if message == "继续" {
						results := harnessCalls(in)
						require.Len(t, results, 1)
						assert.Empty(t, results[0].NextCursor)
						assert.Contains(t, string(results[0].Data), `"transaction_id":"100"`)
					} else {
						assert.Empty(t, harnessCalls(in))
					}
					return data.TurnResponse{Completed: true, Content: "回答"}, nil
				})
				body, _ := json.Marshal(models.AIAssistantChatRequest{Message: message, Continuation: result.Continuation})
				next, err := harnessRequest(t, a, string(body), stream)
				require.Nil(t, err)
				assert.NotContains(t, next.Body.String(), cursor)
			}
			a.assistant.primary = harnessProvider(func(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
				t.Error("invalid cursor reached provider")
				return data.TurnResponse{}, nil
			})
			raw, _ := json.Marshal(models.AIAssistantChatRequest{Message: "继续", Continuation: result.Continuation})
			for _, uid := range []int64{8} {
				c := harnessContext(string(raw), uid, context.Background(), httptest.NewRecorder())
				_, err := a.AssistantChatHandler(c)
				require.NotNil(t, err)
				assert.Equal(t, errs.ErrAIAssistantInvalidCursor.Code(), err.Code())
			}
			oldNow := a.assistant.now
			a.assistant.now = func() time.Time { return oldNow().Add(25 * time.Hour) }
			_, err = harnessRequest(t, a, string(raw), stream)
			require.NotNil(t, err)
			a.assistant.now = oldNow
		})
	}
}

func TestAssistantHarnessEmptyGeneralKnowledgeAndFailures(t *testing.T) {
	a, _ := harnessAPIFixture(t)
	for _, stream := range []bool{false, true} {
		for _, mode := range []string{"empty", "general", "failure", "forecast"} {
			rounds := 0
			a.assistant.primary = harnessProvider(func(_ core.Context, _ int64, in data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
				rounds++
				if mode == "failure" {
					return data.TurnResponse{}, errors.New("SECRET provider payload")
				}
				if mode != "general" && rounds == 1 {
					return data.TurnResponse{Completed: true, ToolCalls: []data.ToolCall{harnessCall("empty", data.ToolAggregateTransactions, "{}")}}, nil
				}
				if mode != "general" {
					results := harnessCalls(in)
					require.Len(t, results, 1)
					assert.True(t, results[0].OK)
					assert.Equal(t, "0", results[0].MatchedCount)
					assert.Equal(t, "[]", string(results[0].Totals))
				}
				text := "复利是利息也产生利息。"
				if mode == "empty" {
					text = "查询成功，范围内暂无账单。"
				}
				if mode == "forecast" {
					text = "没有历史账单证据，无法预测你的下月支出。"
				}
				return data.TurnResponse{Completed: true, Content: text}, nil
			})
			rec, err := harnessRequest(t, a, `{"message":"分析"}`, stream)
			if mode == "failure" {
				require.NotNil(t, err)
				assert.Contains(t, rec.Body.String(), `"assistant":{"state":"failed"`)
				assert.NotContains(t, rec.Body.String(), `"type":"done"`)
				assert.NotContains(t, rec.Body.String(), "SECRET")
			} else {
				require.Nil(t, err)
				if mode == "general" {
					assert.NotContains(t, rec.Body.String(), `"queryId"`)
				} else {
					assert.Contains(t, rec.Body.String(), `"queryId"`)
				}
			}
		}
	}
}

// The writer catches concurrent access as well as injected errors; no Gin object
// is ever read by the model goroutine.
type harnessFailWriter struct {
	http.ResponseWriter
	writes       int
	failAt       int
	active       atomic.Bool
	concurrent   atomic.Bool
	sawHeartbeat bool
}

func (w *harnessFailWriter) Write(p []byte) (int, error) {
	if w.active.Swap(true) {
		w.concurrent.Store(true)
	}
	defer w.active.Store(false)
	w.writes++
	if bytes.HasPrefix(p, []byte(": ")) {
		w.sawHeartbeat = true
	}
	if w.writes >= w.failAt {
		return 0, io.ErrClosedPipe
	}
	return w.ResponseWriter.Write(p)
}

func (w *harnessFailWriter) Flush() { w.ResponseWriter.(http.Flusher).Flush() }

func TestAssistantHarnessStreamCancellationAndWriteFailure(t *testing.T) {
	a, _ := harnessAPIFixture(t)
	for _, kind := range []string{"cancel", "status-write", "heartbeat-write", "timeout"} {
		t.Run(kind, func(t *testing.T) {
			started := make(chan struct{})
			stopped := make(chan struct{})
			var mu sync.Mutex
			observations := []llm.AssistantObservation{}
			a.assistant.observe = func(o llm.AssistantObservation) { mu.Lock(); observations = append(observations, o); mu.Unlock() }
			a.assistant.heartbeat = 5 * time.Millisecond
			a.assistant.primary = harnessProvider(func(c core.Context, _ int64, _ data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
				close(started)
				defer close(stopped)
				<-c.Done()
				return data.TurnResponse{}, c.Err()
			})
			if kind == "timeout" {
				settings.Container.GetCurrentConfig().AIAssistantRuntimeConfig.RunTimeout = 20
			} else {
				settings.Container.GetCurrentConfig().AIAssistantRuntimeConfig.RunTimeout = 2000
			}
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			rec := httptest.NewRecorder()
			writer := &harnessFailWriter{ResponseWriter: rec, failAt: 99999}
			if kind == "status-write" {
				writer.failAt = 1
			}
			if kind == "heartbeat-write" {
				writer.failAt = 2
			}
			c := harnessContext(`{"message":"慢查询"}`, 7, parent, writer)
			finished := make(chan *errs.Error, 1)
			go func() { finished <- a.AssistantChatStreamHandler(c) }()
			if kind != "status-write" {
				select {
				case <-started:
				case <-time.After(time.Second):
					t.Fatal("provider not started")
				}
			}
			if kind == "cancel" {
				cancel()
			}
			select {
			case err := <-finished:
				if kind == "timeout" {
					err = c.GetResponseError()
					require.NotNil(t, err)
					assert.Equal(t, errs.ErrAIAssistantTimeout.Code(), err.Code())
				} else {
					require.Nil(t, err)
				}
			case <-time.After(time.Second):
				t.Fatal("handler failed to join worker")
			}
			if kind != "status-write" {
				select {
				case <-stopped:
				default:
					t.Fatal("upstream remained active")
				}
			}
			assert.False(t, writer.concurrent.Load())
			assert.NotContains(t, rec.Body.String(), `"type":"done"`)
			mu.Lock()
			defer mu.Unlock()
			require.Len(t, observations, 1)
			o := observations[0]
			assert.Equal(t, llm.AssistantObservationRun, o.Kind)
			switch kind {
			case "status-write":
				assert.Nil(t, o.FirstStatusMilliseconds)
				assert.Equal(t, llm.AssistantReasonWriteFailed, o.Reason)
			case "heartbeat-write":
				assert.True(t, writer.sawHeartbeat)
				assert.Equal(t, llm.AssistantReasonWriteFailed, o.Reason)
				assert.NotNil(t, o.FirstStatusMilliseconds)
			case "cancel":
				assert.Equal(t, llm.AssistantReasonCanceled, o.Reason)
			case "timeout":
				assert.Equal(t, llm.AssistantReasonRunTimeout, o.Reason)
			}
		})
	}
}

func TestAssistantHarnessSafeLegacyProjection(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int64
	}{{"-0.01", -1}, {"92233720368547758.07", 9223372036854775807}, {"-92233720368547758.08", -9223372036854775808}} {
		v, err := assistantMinorAmount(tc.in)
		require.NoError(t, err)
		assert.Equal(t, tc.want, v)
	}
	for _, in := range []string{"92233720368547758.08", "1e2", "1.001", "+1.00", "NaN", "1"} {
		_, err := assistantMinorAmount(in)
		require.Error(t, err)
	}
}
