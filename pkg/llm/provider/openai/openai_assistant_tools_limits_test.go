package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/httpclient"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAssistantToolContextLimitsAndRedaction(t *testing.T) {
	for _, tc := range []struct {
		name              string
		responses, stream bool
		status            int
		body              string
	}{
		{"chat-http", false, false, 400, `{"error":{"code":"context_length_exceeded","message":"private-finance-secret"}}`},
		{"responses-http", true, false, 400, `{"error":{"type":"context_window_exceeded"}}`},
		{"chat-stream", false, true, 200, assistantSSE(`{"error":{"code":"context_length_exceeded"}}`)},
		{"responses-error-event", true, false, 200, assistantSSE(`{"type":"error","code":"context_length_exceeded","message":"private"}`)},
		{"responses-stream", true, false, 200, assistantSSE(`{"type":"response.failed","response":{"status":"failed","error":{"code":"context_length_exceeded"}}}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body)
				if strings.Contains(tc.name, "stream") || strings.Contains(tc.name, "event") {
					w.Header().Set("Content-Type", "text/event-stream")
				}
				w.WriteHeader(tc.status)
				io.WriteString(w, tc.body)
			}))
			defer server.Close()
			p := assistantTestProvider(server.URL, tc.responses, tc.stream, 2000)
			logged := false
			c := httpclient.CustomHttpResponseLog(core.NewNullContext(), func([]byte) { logged = true })
			turn, err := p.RunToolTurn(c, 1, assistantTestInput(), nil)
			requireAssistantCode(t, err, data.ToolErrorBudgetExhausted)
			require.False(t, logged)
			require.NotContains(t, err.Error(), "private-finance-secret")
			require.Empty(t, turn.ToolCalls)
		})
	}
	require.False(t, isContextLimit([]byte(`{"choices":[{"message":{"content":"maximum context length"}}]}`)))
}

func TestAssistantToolBoundedBuffers(t *testing.T) {
	_, err := readAssistantBody(strings.NewReader(strings.Repeat("x", assistantResponseBytes+1)))
	requireAssistantCode(t, err, data.ToolErrorProtocol)
	for name, stream := range map[string]string{
		"event":           assistantSSE(strings.Repeat("x", assistantEventBytes+1)),
		"multiline-event": strings.Repeat("data: "+strings.Repeat("x", 1024)+"\n", 300) + "\n",
		"total":           strings.Repeat(": "+strings.Repeat("x", 1000)+"\n\n", 2200),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := readChatToolStream(strings.NewReader(stream))
			requireAssistantCode(t, err, data.ToolErrorProtocol)
		})
	}
	_, err = parseChatToolTurn([]byte(chatFixture(`{"keyword":"`+strings.Repeat("x", assistantArgumentBytes)+`"}`, "tool_calls")))
	requireAssistantCode(t, err, data.ToolErrorProtocol)
}

func TestAssistantToolProviderTimeoutAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name            string
		providerTimeout uint32
		runTimeout      time.Duration
		stream          bool
		cancel          bool
	}{
		{"provider-shorter", 30, time.Second, false, false},
		{"run-shorter", 1000, 30 * time.Millisecond, false, false},
		{"provider-stream", 30, time.Second, true, false},
		{"explicit-cancel", 1000, time.Second, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			started := make(chan struct{})
			stopped := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body)
				if tc.stream {
					w.Header().Set("Content-Type", "text/event-stream")
					io.WriteString(w, "data: {\"choices\":[]}\n\n")
					w.(http.Flusher).Flush()
				}
				close(started)
				<-r.Context().Done()
				close(stopped)
			}))
			defer server.Close()
			p := assistantTestProvider(server.URL, false, tc.stream, tc.providerTimeout)
			parent, cancel := context.WithTimeout(context.Background(), tc.runTimeout)
			defer cancel()
			if tc.cancel {
				go func() { <-started; cancel() }()
			}
			turn, err := p.RunToolTurn(core.NewAssistantContext(parent, "request", "en", "127.0.0.1"), 1, assistantTestInput(), nil)
			require.Error(t, err)
			require.Empty(t, turn.ToolCalls)
			if tc.cancel {
				require.ErrorIs(t, err, context.Canceled)
			} else {
				require.ErrorIs(t, err, context.DeadlineExceeded)
			}
			select {
			case <-stopped:
			case <-time.After(time.Second):
				t.Fatal("upstream did not stop after cancellation")
			}
		})
	}
}

func TestAssistantToolSinkFailureStopsHTTP(t *testing.T) {
	stopped := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, true, body["stream"])
		require.Equal(t, false, body["store"])
		require.Equal(t, map[string]any{"summary": "auto"}, body["reasoning"])
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, assistantSSE(`{"type":"response.reasoning_summary_text.delta","delta":"summary"}`))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(stopped)
	}))
	defer server.Close()
	p := assistantTestProvider(server.URL, true, false, 2000)
	sinkErr := errors.New("closed event sink")
	turn, err := p.RunToolTurn(core.NewNullContext(), 1, assistantTestInput(), func(data.DisplayEvent) error { return sinkErr })
	require.ErrorIs(t, err, sinkErr)
	require.Empty(t, turn.ToolCalls)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("upstream remained active")
	}
}

func TestAssistantNeutralFallbackRejectsUnrepresentableContext(t *testing.T) {
	p := assistantTestProvider("http://unused", false, false, 2000)
	input := assistantTestInput()
	for _, message := range []data.ConversationMessage{
		{Role: data.ConversationRole("developer"), Content: "elevate"},
		{Role: data.ConversationRoleTool, ToolCallID: "unknown", ToolResult: &data.ToolResult{OK: true}},
		{Role: data.ConversationRoleAssistant, ToolCalls: []data.ToolCall{{ID: "missing-result", Name: data.ToolListTags, Arguments: json.RawMessage(`{}`)}}},
	} {
		next := input
		next.Messages = append(append([]data.ConversationMessage(nil), input.Messages...), message)
		_, _, err := p.toolRequestBody(next, false)
		requireAssistantCode(t, err, data.ToolErrorProviderUnsupported)
	}
}

func TestAssistantWireBudgetPreventsNetworkAndUnknownErrorsAreSafe(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(500)
		fmt.Fprint(w, `{"error":{"message":"financial-secret"}}`)
	}))
	defer server.Close()
	p := assistantTestProvider(server.URL, false, false, 2000)
	input := assistantTestInput()
	input.MaxRequestBytes = 1
	_, err := p.RunToolTurn(core.NewNullContext(), 1, input, nil)
	requireAssistantCode(t, err, data.ToolErrorBudgetExhausted)
	require.Equal(t, int32(0), calls.Load())
	input.MaxRequestBytes = 128 * 1024
	_, err = p.RunToolTurn(core.NewNullContext(), 1, input, nil)
	requireAssistantCode(t, err, data.ToolErrorUpstreamFailed)
	require.NotContains(t, err.Error(), "financial-secret")
}

func TestAssistantResponsesInterleavedCallsAndFinalAnswer(t *testing.T) {
	item := func(id, callID, name, args string) map[string]any {
		return map[string]any{"id": id, "type": "function_call", "status": "completed", "call_id": callID, "name": name, "arguments": args}
	}
	terminal := map[string]any{"status": "completed", "output": []any{item("a", "call_a", "list_accounts", `{"keyword":"a"}`), item("b", "call_b", "list_tags", `{"keyword":"b"}`)}}
	stream := assistantSSE(
		`{"type":"response.function_call_arguments.delta","output_index":1,"item_id":"b","delta":"{\"keyword\":"}`,
		`{"type":"response.function_call_arguments.delta","output_index":0,"item_id":"a","delta":"{\"keyword\":"}`,
		`{"type":"response.function_call_arguments.delta","output_index":1,"item_id":"b","delta":"\"b\"}"}`,
		`{"type":"response.function_call_arguments.delta","output_index":0,"item_id":"a","delta":"\"a\"}"}`,
		map[string]any{"type": "response.completed", "response": terminal},
	)
	turn, _, err := readResponsesToolStream(strings.NewReader(stream), nil)
	require.NoError(t, err)
	require.Len(t, turn.ToolCalls, 2)
	require.Equal(t, "call_a", turn.ToolCalls[0].ID)
	require.JSONEq(t, `{"keyword":"b"}`, string(turn.ToolCalls[1].Arguments))
	answer := `{"status":"completed","output":[{"type":"message","id":"m","role":"assistant","status":"completed","content":[{"type":"output_text","text":"Complete answer"}]}]}`
	turn, _, err = parseResponsesToolTurn([]byte(answer))
	require.NoError(t, err)
	require.True(t, turn.Completed)
	require.Empty(t, turn.ToolCalls)
	require.Equal(t, "Complete answer", turn.Content)
	turn, err = parseChatToolTurn([]byte(`{"choices":[{"index":0,"message":{"role":"assistant","content":"Complete answer"},"finish_reason":"stop"}]}`))
	require.NoError(t, err)
	require.True(t, turn.Completed)
	require.Equal(t, "Complete answer", turn.Content)
}

func TestAssistantResponsesStateCannotMoveAdaptersOrPrefixes(t *testing.T) {
	p := assistantTestProvider("http://unused", true, false, 2000)
	other := assistantTestProvider("http://unused", true, false, 2000)
	input := assistantTestInput()
	wire, items, err := p.toolRequestBody(input, false)
	require.NoError(t, err)
	require.NotEmpty(t, wire)
	// State must be from this exact adapter and prefix, not merely same model ID.
	input.State = &responsesToolState{owner: other, prefixLength: len(input.Messages), items: items}
	_, _, err = p.toolRequestBody(input, false)
	requireAssistantCode(t, err, data.ToolErrorProviderUnsupported)
	input.State = &responsesToolState{owner: p, prefixLength: len(input.Messages), items: items}
	_, _, err = p.toolRequestBody(input, false)
	requireAssistantCode(t, err, data.ToolErrorProviderUnsupported)
	// Switching protocol without state preserves every independent role and pair.
	input.State = nil
	input.Messages = append(input.Messages, data.ConversationMessage{Role: data.ConversationRoleAssistant, Content: "lookup", ToolCalls: []data.ToolCall{{ID: "x", Name: data.ToolListTags, Arguments: json.RawMessage(`{}`)}}}, data.ConversationMessage{Role: data.ConversationRoleTool, ToolCallID: "x", ToolResult: &data.ToolResult{OK: true}})
	wire, _, err = p.toolRequestBody(input, true)
	require.NoError(t, err)
	var body map[string]any
	require.NoError(t, json.Unmarshal(wire, &body))
	require.Equal(t, true, body["stream"])
	require.Contains(t, string(wire), `"type":"function_call_output"`)
	require.Contains(t, string(wire), `"call_id":"x"`)
	require.Contains(t, string(wire), `"role":"assistant"`)
}
