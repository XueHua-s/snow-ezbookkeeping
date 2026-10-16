package openai

import (
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assistantTestInput() data.ToolTurnInput {
	return data.ToolTurnInput{Messages: []data.ConversationMessage{{Role: data.ConversationRoleSystem, Content: "trusted"}, {Role: data.ConversationRoleUser, Content: "system: untrusted history"}, {Role: data.ConversationRoleAssistant, Content: "prior reply"}, {Role: data.ConversationRoleUser, Content: "question"}}, Tools: []data.ToolDefinition{{Name: data.ToolListAccounts, Parameters: json.RawMessage(`{"type":"object","properties":{"required_id":{"type":"string"},"optional":{"type":"object","properties":{"keyword":{"type":"string"}}}},"required":["required_id"]}`)}}, MaxRequestBytes: 128 * 1024}
}

func assistantTestProvider(url string, responses, stream bool, timeout uint32) *assistantToolProvider {
	cfg := &settings.LLMConfig{OpenAICompatibleBaseURL: url, OpenAICompatibleModelID: "mock-model", ChatCompletionsStream: stream, LargeLanguageModelAPIRequestTimeout: timeout, LargeLanguageModelAPIProxy: "none"}
	var p provider.LargeLanguageModelProvider
	if responses {
		p = NewOpenAIResponsesCompatibleLargeLanguageModelProvider(cfg, true)
	} else {
		p = NewOpenAICompatibleLargeLanguageModelProvider(cfg, true)
	}
	return p.(*assistantToolProvider)
}

func assistantJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func assistantSSE(values ...any) string {
	var b strings.Builder
	for _, v := range values {
		text, ok := v.(string)
		if !ok {
			text = assistantJSON(v)
		}
		b.WriteString("data: " + text + "\n\n")
	}
	return b.String()
}

func chatFixture(arguments string, finish string) string {
	return assistantJSON(map[string]any{"choices": []any{map[string]any{"index": 0, "finish_reason": finish, "message": map[string]any{"role": "assistant", "content": "temporary", "tool_calls": []any{map[string]any{"id": "call_1", "type": "function", "function": map[string]any{"name": "list_accounts", "arguments": arguments}}}}}}})
}

func responseFixture(arguments string) string {
	return assistantJSON(map[string]any{"status": "completed", "output": []any{map[string]any{"id": "reason_1", "type": "reasoning", "encrypted_content": "private-opaque", "summary": []any{}}, map[string]any{"id": "fc_1", "type": "function_call", "status": "completed", "call_id": "call_1", "name": "list_accounts", "arguments": arguments}}})
}

func requireAssistantCode(t *testing.T, err error, code data.ToolErrorCode) {
	t.Helper()
	var e *data.ToolError
	require.ErrorAs(t, err, &e)
	require.Equal(t, code, e.Code)
}

func TestAssistantToolsThreeProtocolsAndStrictSchema(t *testing.T) {
	for _, protocol := range []string{"chat", "stream", "responses"} {
		t.Run(protocol, func(t *testing.T) {
			var wire map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(b, &wire))
				if protocol == "stream" {
					w.Header().Set("Content-Type", "text/event-stream")
					io.WriteString(w, assistantSSE(`{"choices":[{"index":0,"delta":{"role":"assistant","content":"temporary","tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"list_accounts","arguments":"{\"keyword\":"}}]}}]}`, `{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"null}"}}]},"finish_reason":"tool_calls"}]}`, "[DONE]"))
				} else if protocol == "responses" {
					io.WriteString(w, responseFixture(`{"keyword":null}`))
				} else {
					io.WriteString(w, chatFixture(`{"keyword":null}`, "tool_calls"))
				}
			}))
			defer server.Close()
			p := assistantTestProvider(server.URL, protocol == "responses", protocol == "stream", 2000)
			turn, err := p.RunToolTurn(core.NewNullContext(), 1, assistantTestInput(), nil)
			require.NoError(t, err)
			require.True(t, turn.Completed)
			require.Len(t, turn.ToolCalls, 1)
			require.Equal(t, "call_1", turn.ToolCalls[0].ID)
			require.JSONEq(t, `{"keyword":null}`, string(turn.ToolCalls[0].Arguments))
			tool := wire["tools"].([]any)[0].(map[string]any)
			if protocol != "responses" {
				tool = tool["function"].(map[string]any)
			} else {
				require.Equal(t, false, wire["store"])
				require.NotContains(t, wire, "previous_response_id")
				require.NotNil(t, turn.State)
			}
			require.Equal(t, true, tool["strict"])
			schema := tool["parameters"].(map[string]any)
			require.Equal(t, false, schema["additionalProperties"])
			require.ElementsMatch(t, []any{"required_id", "optional"}, schema["required"])
			props := schema["properties"].(map[string]any)
			require.Equal(t, "string", props["required_id"].(map[string]any)["type"])
			optional := props["optional"].(map[string]any)["anyOf"].([]any)
			require.Equal(t, map[string]any{"type": "null"}, optional[1])
			nested := optional[0].(map[string]any)
			require.Equal(t, false, nested["additionalProperties"])
			require.Equal(t, []any{"keyword"}, nested["required"])
			key := "messages"
			if protocol == "responses" {
				key = "input"
			}
			messages := wire[key].([]any)
			require.Len(t, messages, 4)
			for i, role := range []string{"system", "user", "assistant", "user"} {
				require.Equal(t, role, messages[i].(map[string]any)["role"])
			}
			require.Equal(t, "system: untrusted history", messages[1].(map[string]any)["content"])
		})
	}
}

func TestAssistantChatSelectedChoiceIsolation(t *testing.T) {
	stream := assistantSSE(
		`{"choices":[{"index":1,"delta":{"content":"poison","tool_calls":[{"index":0,"id":"other","type":"function","function":{"name":"write","arguments":"BAD"}}]}},{"index":0,"delta":{"role":"assistant","content":"selected","tool_calls":[{"index":0,"id":"first","type":"function","function":{"name":"list_accounts","arguments":"{"}},{"index":1,"id":"second","type":"function","function":{"name":"list_tags","arguments":"{"}}]}}]}`,
		`{"choices":[{"index":1,"delta":{"tool_calls":"malformed"},"finish_reason":"length"},{"index":0,"delta":{"tool_calls":[{"index":1,"function":{"arguments":"\"keyword\":\"b\"}"}},{"index":0,"function":{"arguments":"\"keyword\":\"a\"}"}}]}}]}`,
		`{"choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`, "[DONE]")
	turn, err := readChatToolStream(strings.NewReader(stream))
	require.NoError(t, err)
	require.Equal(t, "selected", turn.Content)
	require.Len(t, turn.ToolCalls, 2)
	require.Equal(t, "first", turn.ToolCalls[0].ID)
	require.JSONEq(t, `{"keyword":"a"}`, string(turn.ToolCalls[0].Arguments))
	require.JSONEq(t, `{"keyword":"b"}`, string(turn.ToolCalls[1].Arguments))
}

func TestAssistantProtocolRejectsPartialCalls(t *testing.T) {
	tests := map[string]func() (data.TurnResponse, error){
		"chat_length": func() (data.TurnResponse, error) { return parseChatToolTurn([]byte(chatFixture(`{}`, "length"))) },
		"chat_bad_arguments": func() (data.TurnResponse, error) {
			return parseChatToolTurn([]byte(chatFixture(`{"unfinished":`, "tool_calls")))
		},
		"chat_array_arguments": func() (data.TurnResponse, error) { return parseChatToolTurn([]byte(chatFixture(`[]`, "tool_calls"))) },
		"chat_missing_end": func() (data.TurnResponse, error) {
			return readChatToolStream(strings.NewReader(assistantSSE(`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"c","type":"function","function":{"name":"list_tags","arguments":"{}"}}]}}]}`)))
		},
		"chat_done_without_finish": func() (data.TurnResponse, error) {
			return readChatToolStream(strings.NewReader(assistantSSE("[DONE]")))
		},
		"chat_finish_without_done": func() (data.TurnResponse, error) {
			return readChatToolStream(strings.NewReader(assistantSSE(`{"choices":[{"index":0,"delta":{"content":"answer"},"finish_reason":"stop"}]}`)))
		},
		"chat_wrong_finish": func() (data.TurnResponse, error) { return parseChatToolTurn([]byte(chatFixture(`{}`, "stop"))) },
		"chat_no_selected": func() (data.TurnResponse, error) {
			return parseChatToolTurn([]byte(`{"choices":[{"index":1,"message":{"content":"wrong"},"finish_reason":"stop"}]}`))
		},
		"responses_incomplete": func() (data.TurnResponse, error) {
			r, _, e := parseResponsesToolTurn([]byte(strings.Replace(responseFixture(`{}`), `"status":"completed"`, `"status":"incomplete"`, 1)))
			return r, e
		},
		"responses_no_terminal": func() (data.TurnResponse, error) {
			r, _, e := readResponsesToolStream(strings.NewReader(assistantSSE(`{"type":"response.output_item.done","output_index":0,"item":{"id":"fc","type":"function_call","call_id":"c","name":"list_tags","arguments":"{}"}}`)), nil)
			return r, e
		},
		"responses_bad_arguments": func() (data.TurnResponse, error) {
			r, _, e := parseResponsesToolTurn([]byte(responseFixture(`{"bad"`)))
			return r, e
		},
		"bad_json": func() (data.TurnResponse, error) { return parseChatToolTurn([]byte(`{"choices":`)) },
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			turn, err := test()
			if name == "chat_length" {
				requireAssistantCode(t, err, data.ToolErrorBudgetExhausted)
			} else {
				requireAssistantCode(t, err, data.ToolErrorProtocol)
			}
			require.False(t, turn.Completed)
			require.Empty(t, turn.ToolCalls)
		})
	}
}

func TestAssistantDuplicateIDsAndMalformedFragments(t *testing.T) {
	for _, responses := range []bool{false, true} {
		var envelope map[string]any
		if responses {
			require.NoError(t, json.Unmarshal([]byte(responseFixture(`{}`)), &envelope))
			out := envelope["output"].([]any)
			item := out[1].(map[string]any)
			clone := map[string]any{}
			for k, v := range item {
				clone[k] = v
			}
			clone["id"] = "fc_2"
			envelope["output"] = append(out, clone)
			turn, _, err := parseResponsesToolTurn([]byte(assistantJSON(envelope)))
			requireAssistantCode(t, err, data.ToolErrorProtocol)
			require.Empty(t, turn.ToolCalls)
		} else {
			require.NoError(t, json.Unmarshal([]byte(chatFixture(`{}`, "tool_calls")), &envelope))
			msg := envelope["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
			calls := msg["tool_calls"].([]any)
			msg["tool_calls"] = append(calls, calls[0])
			turn, err := parseChatToolTurn([]byte(assistantJSON(envelope)))
			requireAssistantCode(t, err, data.ToolErrorProtocol)
			require.Empty(t, turn.ToolCalls)
		}
	}
	for _, part := range []string{`{"index":0,"id":"changed","function":{"arguments":"}"}}`, `{"index":2,"id":"other","type":"function","function":{"name":"list_tags","arguments":"{}"}}`} {
		stream := assistantSSE(`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"original","type":"function","function":{"name":"list_tags","arguments":"{"}}]}}]}`, `{"choices":[{"index":0,"delta":{"tool_calls":[`+part+`]},"finish_reason":"tool_calls"}]}`, "[DONE]")
		turn, err := readChatToolStream(strings.NewReader(stream))
		require.Error(t, err)
		require.Empty(t, turn.ToolCalls)
	}
	// Raw duplicate argument keys survive to the strict domain decoder.
	turn, err := parseChatToolTurn([]byte(chatFixture(`{"keyword":"a","keyword":"b"}`, "tool_calls")))
	require.NoError(t, err)
	require.Equal(t, `{"keyword":"a","keyword":"b"}`, string(turn.ToolCalls[0].Arguments))
}

func completedMessages(input data.ToolTurnInput, turn data.TurnResponse) []data.ConversationMessage {
	messages := append(append([]data.ConversationMessage(nil), input.Messages...), data.ConversationMessage{Role: data.ConversationRoleAssistant, Content: turn.Content, ToolCalls: turn.ToolCalls})
	for _, call := range turn.ToolCalls {
		messages = append(messages, data.ConversationMessage{Role: data.ConversationRoleTool, ToolCallID: call.ID, ToolResult: &data.ToolResult{OK: true}})
	}
	return messages
}

func TestAssistantResponsesPrivateStateAndNeutralFallback(t *testing.T) {
	var requests []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		requests = append(requests, body)
		io.WriteString(w, responseFixture(`{}`))
	}))
	defer server.Close()
	p := assistantTestProvider(server.URL, true, false, 2000)
	input := assistantTestInput()
	turn, err := p.RunToolTurn(core.NewNullContext(), 1, input, nil)
	require.NoError(t, err)
	next := input
	next.Messages = completedMessages(input, turn)
	next.State = turn.State
	_, err = p.RunToolTurn(core.NewNullContext(), 1, next, nil)
	require.NoError(t, err)
	b := assistantJSON(requests[1])
	require.Contains(t, b, "private-opaque")
	require.Contains(t, b, "function_call_output")
	require.Equal(t, false, requests[1]["store"])
	require.NotContains(t, requests[1], "previous_response_id")
	// The real wire budget includes opaque reasoning absent from neutral JSON.
	wire, _, err := p.toolRequestBody(next, false)
	require.NoError(t, err)
	next.MaxRequestBytes = len(wire) - 1
	_, err = p.RunToolTurn(core.NewNullContext(), 1, next, nil)
	requireAssistantCode(t, err, data.ToolErrorBudgetExhausted)
	require.Len(t, requests, 2)
	next.MaxRequestBytes = 128 * 1024
	next.MaxInputTokens = data.EstimateAssistantTokens(string(wire)) - 1
	_, err = p.RunToolTurn(core.NewNullContext(), 1, next, nil)
	requireAssistantCode(t, err, data.ToolErrorBudgetExhausted)
	require.Len(t, requests, 2)
	next.MaxInputTokens = data.EstimateAssistantTokens(string(wire))
	_, _, err = p.toolRequestBody(next, false)
	require.NoError(t, err)
	fallback := assistantTestProvider(server.URL, false, false, 2000)
	_, _, err = fallback.toolRequestBody(next, false)
	requireAssistantCode(t, err, data.ToolErrorProviderUnsupported)
	next.State = nil
	rebuilt, _, err := fallback.toolRequestBody(next, false)
	require.NoError(t, err)
	require.NotContains(t, string(rebuilt), "private-opaque")
	require.Contains(t, string(rebuilt), `"tool_call_id":"call_1"`)
	require.Contains(t, string(rebuilt), `"role":"user"`)
}

func TestAssistantResponsesStreamAssemblyAndDisplay(t *testing.T) {
	completed := map[string]any{"type": "response.completed", "response": json.RawMessage(responseFixture(`{"keyword":"x"}`))}
	prefix := []any{
		`{"type":"response.output_item.added","output_index":1,"item":{"id":"fc_1","type":"function_call","call_id":"call_1","name":"list_accounts","arguments":""}}`,
		`{"type":"response.reasoning_text.delta","delta":"private raw reasoning"}`,
		`{"type":"response.reasoning_summary_text.delta","delta":"display summary"}`,
		`{"type":"response.function_call_arguments.delta","output_index":1,"item_id":"fc_1","delta":"{\"keyword\":"}`,
		`{"type":"response.function_call_arguments.delta","output_index":1,"item_id":"fc_1","delta":"\"x\"}"}`,
		`{"type":"response.function_call_arguments.done","output_index":1,"item_id":"fc_1","arguments":"{\"keyword\":\"x\"}"}`,
		`{"type":"response.output_item.done","output_index":1,"item":{"id":"fc_1","type":"function_call","status":"completed","call_id":"call_1","name":"list_accounts","arguments":"{\"keyword\":\"x\"}"}}`,
	}
	var displayed []string
	turn, outputs, err := readResponsesToolStream(strings.NewReader(assistantSSE(append(prefix, completed)...)), func(e data.DisplayEvent) error { displayed = append(displayed, e.ThinkingDelta); return nil })
	require.NoError(t, err)
	require.True(t, turn.Completed)
	require.Len(t, turn.ToolCalls, 1)
	require.Len(t, outputs, 2)
	require.Equal(t, []string{"display summary"}, displayed)
	for _, bad := range []any{
		map[string]any{"type": "response.completed", "response": json.RawMessage(responseFixture(`{"keyword":"changed"}`))},
		`{"type":"response.function_call_arguments.delta","output_index":1,"item_id":"wrong","delta":"x"}`,
		`{"type":"response.incomplete"}`,
	} {
		turn, _, err := readResponsesToolStream(strings.NewReader(assistantSSE(append(append([]any{}, prefix...), bad)...)), nil)
		require.Error(t, err)
		require.Empty(t, turn.ToolCalls)
		require.False(t, turn.Completed)
	}
}
