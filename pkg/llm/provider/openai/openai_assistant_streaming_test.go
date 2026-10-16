package openai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/stretchr/testify/require"
)

// The completion is deliberately withheld until the client sees actual text.
// A parser that buffers the full response deadlocks and fails this regression.
func TestAssistantHTTPTextArrivesBeforeCompletion(t *testing.T) {
	for _, responses := range []bool{false, true} {
		t.Run(map[bool]string{false: "chat", true: "responses"}[responses], func(t *testing.T) {
			release := make(chan struct{})
			defer close(release)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, true, body["stream"])
				w.Header().Set("Content-Type", "text/event-stream")
				if responses {
					io.WriteString(w, assistantSSE(`{"type":"response.output_text.delta","output_index":0,"item_id":"m1","delta":"你好"}`))
				} else {
					io.WriteString(w, assistantSSE(`{"choices":[{"index":0,"delta":{"content":"你好"}}]}`))
				}
				w.(http.Flusher).Flush()
				select {
				case <-release:
				case <-r.Context().Done():
					return
				}
				if responses {
					io.WriteString(w, assistantSSE(`{"type":"response.completed","response":{"status":"completed","output":[{"id":"m1","type":"message","role":"assistant","content":[{"type":"output_text","text":"你好世界"}]}]}}`))
				} else {
					io.WriteString(w, assistantSSE(`{"choices":[{"index":0,"delta":{"content":"世界"},"finish_reason":"stop"}]}`, "[DONE]"))
				}
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			partial := make(chan string, 2)
			done := make(chan error, 1)
			go func() {
				_, err := assistantTestProvider(server.URL, responses, false, 5000).RunToolTurn(core.NewAssistantContext(ctx, "test", "en", ""), 1, assistantTestInput(), func(e data.DisplayEvent) error {
					if e.ReplyDelta != "" {
						partial <- e.ReplyDelta
					}
					return nil
				})
				done <- err
			}()
			select {
			case text := <-partial:
				require.Equal(t, "你好", text)
			case err := <-done:
				t.Fatalf("completed before partial: %v", err)
			case <-ctx.Done():
				t.Fatal("no partial before upstream completion")
			}
			// Deferred release lets cancellation close an intentionally blocked upstream.
			cancel()
			require.Error(t, <-done)
		})
	}
}

func TestResponsesRejectsChangedStreamedText(t *testing.T) {
	input := assistantSSE(`{"type":"response.output_text.delta","output_index":0,"item_id":"m1","delta":"first"}`, `{"type":"response.completed","response":{"status":"completed","output":[{"id":"m1","type":"message","role":"assistant","content":[{"type":"output_text","text":"different"}]}]}}`)
	_, _, err := readResponsesToolStream(strings.NewReader(input), func(data.DisplayEvent) error { return nil })
	requireAssistantCode(t, err, data.ToolErrorProtocol)
}

func TestChatRetractsPreambleBeforeToolArgumentsComplete(t *testing.T) {
	input := assistantSSE(`{"choices":[{"index":0,"delta":{"content":"provisional"}}]}`, `{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call1","type":"function","function":{"name":"list_accounts","arguments":"{"}}]}}]}`)
	var events []data.DisplayEvent
	_, err := readChatToolStream(strings.NewReader(input), func(e data.DisplayEvent) error { events = append(events, e); return nil })
	require.Error(t, err)
	require.Equal(t, []data.DisplayEvent{{ReplyDelta: "provisional"}, {ReplyReset: true}}, events)
}

func TestResponsesIncompleteValidatesTextAndRetractsToolPreambles(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		reset, valid bool
	}{
		{"normal tail", `[{"id":"m1","type":"message","role":"assistant","content":[{"type":"output_text","text":"checking-more"}]}]`, false, true},
		{"wrong identity", `[{"id":"m2","type":"message","role":"assistant","content":[{"type":"output_text","text":"checking-more"}]}]`, false, false},
		{"rewritten text", `[{"id":"m1","type":"message","role":"assistant","content":[{"type":"output_text","text":"different"}]}]`, false, false},
		{"terminal tool", `[{"id":"m1","type":"message","role":"assistant","content":[{"type":"output_text","text":"checking"}]},{"id":"f1","type":"function_call","call_id":"c1","name":"list_accounts","arguments":"{"}]`, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := assistantSSE(`{"type":"response.output_text.delta","output_index":0,"item_id":"m1","delta":"checking"}`, `{"type":"response.incomplete","response":{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},"output":`+tc.output+`}}`)
			var events []data.DisplayEvent
			result, _, err := readResponsesToolStream(strings.NewReader(input), func(e data.DisplayEvent) error { events = append(events, e); return nil })
			if tc.valid {
				requireAssistantCode(t, err, data.ToolErrorBudgetExhausted)
				require.True(t, result.OutputLimited)
			} else {
				requireAssistantCode(t, err, data.ToolErrorProtocol)
				require.False(t, result.OutputLimited)
			}
			require.Equal(t, tc.reset, events[len(events)-1].ReplyReset)
			if tc.reset {
				require.Empty(t, result.Content)
			}
		})
	}
}

func TestResponsesArgumentOnlyToolRetractsAndPropagatesResetFailure(t *testing.T) {
	input := assistantSSE(`{"type":"response.output_text.delta","output_index":0,"item_id":"m1","delta":"checking"}`, `{"type":"response.function_call_arguments.delta","output_index":1,"item_id":"f1","delta":"{"}`)
	failure := errors.New("reset disconnected")
	_, _, err := readResponsesToolStream(strings.NewReader(input), func(e data.DisplayEvent) error {
		if e.ReplyReset {
			return failure
		}
		return nil
	})
	var sinkErr *assistantSinkError
	require.ErrorAs(t, err, &sinkErr)
	require.Equal(t, failure, sinkErr.error)
}
