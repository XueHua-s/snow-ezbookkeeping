package openai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/httpclient"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// These transport limits bound untrusted upstream buffering, not ledger queries.
const (
	assistantResponseBytes       = 2 * 1024 * 1024
	assistantEventBytes          = 256 * 1024
	assistantArgumentBytes       = 96 * 1024
	assistantDefaultRequestBytes = 128 * 1024
)

type assistantToolProvider struct {
	provider.LargeLanguageModelProvider
	client       *http.Client
	buildRequest func(core.Context, int64) (*http.Request, error)
	model        string
	responses    bool
	stream       bool
	thinking     settings.LLMThinkingLevel
}

// responsesToolState is immutable and belongs to one adapter and neutral prefix.
// No remote response ID or server-side storage is used. Rebuilding a fallback
// deliberately starts from neutral messages with State=nil.
type responsesToolState struct {
	owner        *assistantToolProvider
	prefixLength int
	prefixHash   [32]byte
	items        []json.RawMessage
}

func newAssistantToolProvider(base provider.LargeLanguageModelProvider, cfg *settings.LLMConfig, build func(core.Context, int64) (*http.Request, error), model string, responses bool) *assistantToolProvider {
	return &assistantToolProvider{LargeLanguageModelProvider: base, client: httpclient.NewHttpClient(cfg.LargeLanguageModelAPIRequestTimeout, cfg.LargeLanguageModelAPIProxy, cfg.LargeLanguageModelAPISkipTLSVerify, core.GetOutgoingUserAgent(), false, nil), buildRequest: build, model: model, responses: responses, stream: !responses && cfg.ChatCompletionsStream, thinking: cfg.EnableThinking}
}

func assistantError(code data.ToolErrorCode) error {
	messages := map[data.ToolErrorCode]string{
		data.ToolErrorProtocol:            "The model turn did not complete a valid protocol response.",
		data.ToolErrorProviderUnsupported: "The provider cannot represent this tool conversation.",
		data.ToolErrorBudgetExhausted:     "The model context budget was exhausted.",
		data.ToolErrorUpstreamFailed:      "The model provider request failed.",
	}
	return &data.ToolError{Code: code, Message: messages[code], Retryable: code == data.ToolErrorUpstreamFailed || code == data.ToolErrorProtocol}
}

func assistantContextError() error {
	e := assistantError(data.ToolErrorBudgetExhausted).(*data.ToolError)
	e.UpstreamContext = true
	return e
}

func (p *assistantToolProvider) RunToolTurn(c core.Context, uid int64, input data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
	var zero data.TurnResponse
	if err := c.Err(); err != nil {
		return zero, err
	}
	body, items, err := p.toolRequestBody(input, emit != nil)
	if err != nil {
		return zero, err
	}
	req, err := p.buildRequest(c, uid)
	if err != nil {
		return zero, assistantError(data.ToolErrorUpstreamFailed)
	}
	req = req.WithContext(c)
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return zero, assistantTransportError(c, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, readErr := readAssistantBody(resp.Body)
		if readErr != nil {
			return zero, assistantTransportError(c, readErr)
		}
		if isContextLimit(b) {
			return zero, assistantContextError()
		}
		return zero, assistantError(data.ToolErrorUpstreamFailed)
	}
	var turn data.TurnResponse
	var outputs []json.RawMessage
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/event-stream") || p.stream {
		if p.responses {
			turn, outputs, err = readResponsesToolStream(resp.Body, emit)
		} else {
			turn, err = readChatToolStream(resp.Body, emit)
		}
	} else {
		var b []byte
		b, err = readAssistantBody(resp.Body)
		if err == nil {
			if isContextLimit(b) {
				err = assistantContextError()
			} else if p.responses {
				turn, outputs, err = parseResponsesToolTurn(b)
			} else {
				turn, err = parseChatToolTurn(b)
			}
		}
	}
	if err != nil {
		if turn.OutputLimited && c.Err() == nil {
			return turn, assistantTransportError(c, err)
		}
		return zero, assistantTransportError(c, err)
	}
	if err = c.Err(); err != nil {
		return zero, err
	}
	if p.responses {
		prefix := append(append([]data.ConversationMessage(nil), input.Messages...), data.ConversationMessage{Role: data.ConversationRoleAssistant, Content: turn.Content, ToolCalls: turn.ToolCalls})
		b, _ := json.Marshal(prefix)
		turn.State = &responsesToolState{owner: p, prefixLength: len(prefix), prefixHash: sha256.Sum256(b), items: append(items, outputs...)}
	}
	return turn, nil
}

type assistantSinkError struct{ error }

func assistantTransportError(c core.Context, err error) error {
	var sinkError *assistantSinkError
	if errors.As(err, &sinkError) {
		return sinkError.error
	}
	if c.Err() != nil {
		return c.Err()
	}
	var controlled *data.ToolError
	if errors.As(err, &controlled) {
		return err
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return assistantError(data.ToolErrorUpstreamFailed)
}

func readAssistantBody(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, assistantResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > assistantResponseBytes {
		return nil, assistantError(data.ToolErrorProtocol)
	}
	return b, nil
}

// Only explicit provider error fields are considered; user text cannot trigger
// budget normalization and upstream descriptions are never returned or logged.
func isContextLimit(b []byte) bool {
	var envelope struct {
		Type    string `json:"type"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Error   *struct {
			Code    string `json:"code"`
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
		Response json.RawMessage `json:"response"`
	}
	if json.Unmarshal(b, &envelope) != nil {
		return false
	}
	if envelope.Error == nil {
		if envelope.Type == "error" {
			return contextLimitFields(envelope.Code, "", envelope.Message)
		}
		if len(envelope.Response) > 0 {
			return isContextLimit(envelope.Response)
		}
		return false
	}
	return contextLimitFields(envelope.Error.Code, envelope.Error.Type, envelope.Error.Message)
}

func contextLimitFields(code, kind, message string) bool {
	for _, s := range []string{code, kind} {
		switch s {
		case "context_length_exceeded", "context_window_exceeded", "max_context_length_exceeded", "input_too_long":
			return true
		}
	}
	msg := strings.ToLower(message)
	return strings.Contains(msg, "maximum context length") || strings.Contains(msg, "context window") && (strings.Contains(msg, "exceed") || strings.Contains(msg, "too long"))
}

func (p *assistantToolProvider) toolRequestBody(input data.ToolTurnInput, display bool) ([]byte, []json.RawMessage, error) {
	if err := validateToolConversation(input.Messages); err != nil {
		return nil, nil, err
	}
	tools := make([]any, 0, len(input.Tools))
	for _, def := range input.Tools {
		var schema map[string]any
		if json.Unmarshal(def.Parameters, &schema) != nil || schema == nil {
			return nil, nil, assistantError(data.ToolErrorProviderUnsupported)
		}
		strictToolSchema(schema)
		fn := map[string]any{"name": def.Name, "description": def.Description, "parameters": schema, "strict": true}
		if p.responses {
			fn["type"] = "function"
			tools = append(tools, fn)
		} else {
			tools = append(tools, map[string]any{"type": "function", "function": fn})
		}
	}
	body := map[string]any{"model": p.model, "tools": tools, "stream": p.stream || display}
	if input.MaxOutputTokens > 0 {
		if p.responses {
			body["max_output_tokens"] = input.MaxOutputTokens
		} else {
			body["max_completion_tokens"] = input.MaxOutputTokens
		}
	}
	if p.responses {
		reasoning := map[string]any{}
		if effort, ok := openAIResponsesRequestReasoningEffortsMapping[p.thinking]; ok {
			reasoning["effort"] = effort
		}
		if display && p.thinking != settings.LLMThinkingDisabled {
			reasoning["summary"] = openAIResponsesReasoningSummaryLevel
		}
		if len(reasoning) > 0 {
			body["reasoning"] = reasoning
		}
	} else if effort, ok := openAIChatCompletionsRequestReasoningEffortsMapping[p.thinking]; ok {
		body["reasoning"] = map[string]any{"effort": effort}
	}
	var items []json.RawMessage
	start := 0
	if input.State != nil {
		state, ok := input.State.(*responsesToolState)
		if !p.responses || !ok || state.owner != p || len(input.Messages) < state.prefixLength {
			return nil, nil, assistantError(data.ToolErrorProviderUnsupported)
		}
		b, _ := json.Marshal(input.Messages[:state.prefixLength])
		if sha256.Sum256(b) != state.prefixHash {
			return nil, nil, assistantError(data.ToolErrorProviderUnsupported)
		}
		items = append(items, state.items...)
		start = state.prefixLength
	}
	for _, m := range input.Messages[start:] {
		encoded, err := encodeToolMessage(m, p.responses)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, encoded...)
	}
	if p.responses {
		body["input"] = items
		body["store"] = false
		body["include"] = []string{"reasoning.encrypted_content"}
	} else {
		body["messages"] = items
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, nil, assistantError(data.ToolErrorProviderUnsupported)
	}
	limit := input.MaxRequestBytes
	if limit <= 0 {
		limit = assistantDefaultRequestBytes
	}
	// Check the actual wire body, including private Responses continuation state.
	if !data.AssistantRequestFits(b, input.MaxInputTokens, limit) {
		return nil, nil, assistantError(data.ToolErrorBudgetExhausted)
	}
	return b, items, nil
}

func strictToolSchema(schema map[string]any) {
	// Walk subschemas rather than arbitrary strings, preserving domain constraints.
	for _, key := range []string{"items", "additionalProperties"} {
		if sub, ok := schema[key].(map[string]any); ok {
			strictToolSchema(sub)
		}
	}
	for _, key := range []string{"$defs", "definitions"} {
		if defs, ok := schema[key].(map[string]any); ok {
			for _, v := range defs {
				if sub, ok := v.(map[string]any); ok {
					strictToolSchema(sub)
				}
			}
		}
	}
	for _, key := range []string{"anyOf", "oneOf", "allOf"} {
		if subs, ok := schema[key].([]any); ok {
			for _, v := range subs {
				if sub, ok := v.(map[string]any); ok {
					strictToolSchema(sub)
				}
			}
		}
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		if schema["type"] == "object" {
			schema["additionalProperties"] = false
			schema["required"] = []string{}
		}
		return
	}
	required := map[string]bool{}
	if list, ok := schema["required"].([]any); ok {
		for _, v := range list {
			if key, ok := v.(string); ok {
				required[key] = true
			}
		}
	}
	names := make([]string, 0, len(props))
	for name, v := range props {
		sub, ok := v.(map[string]any)
		if !ok {
			continue
		}
		strictToolSchema(sub)
		if !required[name] {
			props[name] = map[string]any{"anyOf": []any{sub, map[string]any{"type": "null"}}}
		}
		names = append(names, name)
	}
	sort.Strings(names)
	schema["required"] = names
	schema["additionalProperties"] = false
}

func validateToolConversation(messages []data.ConversationMessage) error {
	pending := map[string]bool{}
	for _, m := range messages {
		switch m.Role {
		case data.ConversationRoleSystem, data.ConversationRoleUser, data.ConversationRoleAssistant:
			if len(pending) > 0 || m.ToolResult != nil || m.ToolCallID != "" || (m.Role != data.ConversationRoleAssistant && len(m.ToolCalls) > 0) {
				return assistantError(data.ToolErrorProviderUnsupported)
			}
			for _, call := range m.ToolCalls {
				if call.ID == "" || call.Name == "" || !validArguments(call.Arguments) || pending[call.ID] {
					return assistantError(data.ToolErrorProviderUnsupported)
				}
				pending[call.ID] = true
			}
		case data.ConversationRoleTool:
			if !pending[m.ToolCallID] || m.ToolResult == nil || m.Content != "" || len(m.ToolCalls) > 0 {
				return assistantError(data.ToolErrorProviderUnsupported)
			}
			delete(pending, m.ToolCallID)
		default:
			return assistantError(data.ToolErrorProviderUnsupported)
		}
	}
	if len(pending) > 0 {
		return assistantError(data.ToolErrorProviderUnsupported)
	}
	return nil
}

func encodeToolMessage(m data.ConversationMessage, responses bool) ([]json.RawMessage, error) {
	var values []any
	if m.Role == data.ConversationRoleTool {
		result, err := json.Marshal(m.ToolResult)
		if err != nil {
			return nil, assistantError(data.ToolErrorProviderUnsupported)
		}
		if responses {
			values = append(values, map[string]any{"type": "function_call_output", "call_id": m.ToolCallID, "output": string(result)})
		} else {
			values = append(values, map[string]any{"role": "tool", "tool_call_id": m.ToolCallID, "content": string(result)})
		}
	} else if responses {
		if m.Content != "" || len(m.ToolCalls) == 0 {
			values = append(values, map[string]any{"role": m.Role, "content": m.Content})
		}
		for _, call := range m.ToolCalls {
			values = append(values, map[string]any{"type": "function_call", "call_id": call.ID, "name": call.Name, "arguments": string(call.Arguments)})
		}
	} else {
		msg := map[string]any{"role": m.Role, "content": m.Content}
		if len(m.ToolCalls) > 0 {
			calls := make([]any, 0, len(m.ToolCalls))
			for _, call := range m.ToolCalls {
				calls = append(calls, map[string]any{"type": "function", "id": call.ID, "function": map[string]any{"name": call.Name, "arguments": string(call.Arguments)}})
			}
			msg["tool_calls"] = calls
		}
		values = append(values, msg)
	}
	out := make([]json.RawMessage, 0, len(values))
	for _, v := range values {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, assistantError(data.ToolErrorProviderUnsupported)
		}
		out = append(out, b)
	}
	return out, nil
}

func validArguments(b []byte) bool {
	trimmed := bytes.TrimSpace(b)
	return len(b) <= assistantArgumentBytes && len(trimmed) > 1 && trimmed[0] == '{' && json.Valid(trimmed)
}

var _ provider.LargeLanguageModelToolProvider = (*assistantToolProvider)(nil)

// CheckToolTurnInput measures the exact wire envelope, including adapter-owned
// state and JSON escaping, without sending an upstream request.
func (p *assistantToolProvider) CheckToolTurnInput(input data.ToolTurnInput) error {
	_, _, err := p.toolRequestBody(input, false)
	if err != nil {
		return err
	}
	_, _, err = p.toolRequestBody(input, true)
	return err
}
