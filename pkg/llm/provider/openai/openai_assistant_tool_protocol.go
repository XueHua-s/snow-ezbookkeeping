package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

type assistantWireCall struct {
	Index    *int   `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}
type assistantChatMessage struct {
	Role      string              `json:"role"`
	Content   string              `json:"content"`
	ToolCalls []assistantWireCall `json:"tool_calls"`
	Refusal   string              `json:"refusal"`
}
type assistantChatChoice struct {
	Index        int                  `json:"index"`
	Message      assistantChatMessage `json:"message"`
	Delta        assistantChatMessage `json:"delta"`
	FinishReason *string              `json:"finish_reason"`
}

// Candidate zero is selected for the entire turn. Decode other candidates only
// far enough to inspect their index: their malformed or truncated content must
// never influence the selected candidate.
func selectedChatChoice(b []byte) (*assistantChatChoice, error) {
	var envelope struct {
		Choices []json.RawMessage `json:"choices"`
		Error   json.RawMessage   `json:"error"`
	}
	if json.Unmarshal(b, &envelope) != nil {
		return nil, assistantError(data.ToolErrorProtocol)
	}
	if len(envelope.Error) > 0 && string(envelope.Error) != "null" {
		if isContextLimit(b) {
			return nil, assistantError(data.ToolErrorBudgetExhausted)
		}
		return nil, assistantError(data.ToolErrorUpstreamFailed)
	}
	var selected *assistantChatChoice
	for _, raw := range envelope.Choices {
		var idx struct {
			Index *int `json:"index"`
		}
		if json.Unmarshal(raw, &idx) != nil || idx.Index == nil {
			return nil, assistantError(data.ToolErrorProtocol)
		}
		if *idx.Index != 0 {
			continue
		}
		if selected != nil {
			return nil, assistantError(data.ToolErrorProtocol)
		}
		selected = &assistantChatChoice{}
		if json.Unmarshal(raw, selected) != nil {
			return nil, assistantError(data.ToolErrorProtocol)
		}
	}
	return selected, nil
}

func finishChatToolTurn(content string, calls []assistantWireCall, finish string) (data.TurnResponse, error) {
	var zero data.TurnResponse
	if finish == "length" {
		partial := data.TurnResponse{OutputLimited: true}
		if len(calls) == 0 {
			partial.Content = content
		}
		return partial, assistantError(data.ToolErrorBudgetExhausted)
	}
	if (len(calls) > 0 && finish != "tool_calls") || (len(calls) == 0 && finish != "stop") {
		return zero, assistantError(data.ToolErrorProtocol)
	}
	turn := data.TurnResponse{Content: content, Completed: true}
	ids := map[string]bool{}
	for _, call := range calls {
		if call.ID == "" || len(call.ID) > 1024 || call.Type != "function" || call.Function.Name == "" || len(call.Function.Name) > 256 || ids[call.ID] || !validArguments([]byte(call.Function.Arguments)) {
			return zero, assistantError(data.ToolErrorProtocol)
		}
		ids[call.ID] = true
		turn.ToolCalls = append(turn.ToolCalls, data.ToolCall{ID: call.ID, Name: data.ToolName(call.Function.Name), Arguments: json.RawMessage(call.Function.Arguments)})
	}
	return turn, nil
}
func parseChatToolTurn(b []byte) (data.TurnResponse, error) {
	choice, err := selectedChatChoice(b)
	if err != nil {
		return data.TurnResponse{}, err
	}
	if choice == nil || choice.FinishReason == nil || choice.Message.Refusal != "" || choice.Message.Role != "assistant" {
		return data.TurnResponse{}, assistantError(data.ToolErrorProtocol)
	}
	return finishChatToolTurn(choice.Message.Content, choice.Message.ToolCalls, *choice.FinishReason)
}

// readAssistantSSE bounds both total upstream data and a full SSE event,
// including multiple data lines. EOF alone never substitutes for a terminal.
func readAssistantSSE(r io.Reader, consume func([]byte) (bool, error)) error {
	scanner := bufio.NewScanner(io.LimitReader(r, assistantResponseBytes+1))
	scanner.Buffer(make([]byte, 4096), assistantEventBytes+1)
	var event bytes.Buffer
	total := 0
	dispatch := func() (bool, error) {
		if event.Len() == 0 {
			return false, nil
		}
		b := bytes.TrimSuffix(event.Bytes(), []byte("\n"))
		done, err := consume(b)
		event.Reset()
		return done, err
	}
	for scanner.Scan() {
		line := scanner.Text()
		total += len(line) + 1
		if total > assistantResponseBytes {
			return assistantError(data.ToolErrorProtocol)
		}
		if line == "" {
			done, err := dispatch()
			if err != nil {
				return err
			}
			if done {
				return nil
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimPrefix(line, "data:")
			line = strings.TrimPrefix(line, " ")
			if event.Len()+len(line)+1 > assistantEventBytes {
				return assistantError(data.ToolErrorProtocol)
			}
			event.WriteString(line)
			event.WriteByte('\n')
		}
	}
	if scanner.Err() != nil {
		if errors.Is(scanner.Err(), context.DeadlineExceeded) || errors.Is(scanner.Err(), context.Canceled) {
			return scanner.Err()
		}
		return assistantError(data.ToolErrorProtocol)
	}
	done, err := dispatch()
	if err != nil {
		return err
	}
	if done {
		return nil
	}
	return assistantError(data.ToolErrorProtocol)
}

func readChatToolStream(r io.Reader, sinks ...data.DisplayEventSink) (data.TurnResponse, error) {
	var result data.TurnResponse
	var content strings.Builder
	var emit data.DisplayEventSink
	if len(sinks) > 0 {
		emit = sinks[0]
	}
	toolSeen := false
	calls := map[int]*assistantWireCall{}
	finished := false
	err := readAssistantSSE(r, func(b []byte) (bool, error) {
		if string(b) == "[DONE]" {
			if !finished {
				return false, assistantError(data.ToolErrorProtocol)
			}
			return true, nil
		}
		choice, err := selectedChatChoice(b)
		if err != nil {
			return false, err
		}
		if choice == nil {
			return false, nil
		}
		if finished {
			return false, assistantError(data.ToolErrorProtocol)
		}
		if choice.Delta.Refusal != "" || (choice.Delta.Role != "" && choice.Delta.Role != "assistant") {
			return false, assistantError(data.ToolErrorProtocol)
		}
		if len(choice.Delta.ToolCalls) > 0 && !toolSeen {
			toolSeen = true
			if emit != nil {
				if err := emit(data.DisplayEvent{ReplyReset: true}); err != nil {
					return false, &assistantSinkError{err}
				}
			}
		}
		if !toolSeen && emit != nil && choice.Delta.Content != "" {
			if err := emit(data.DisplayEvent{ReplyDelta: choice.Delta.Content}); err != nil {
				return false, &assistantSinkError{err}
			}
		}
		content.WriteString(choice.Delta.Content)
		for _, part := range choice.Delta.ToolCalls {
			if part.Index == nil || *part.Index < 0 || *part.Index >= 1024 {
				return false, assistantError(data.ToolErrorProtocol)
			}
			call := calls[*part.Index]
			if call == nil {
				call = &assistantWireCall{}
				calls[*part.Index] = call
			}
			if part.ID != "" {
				if call.ID != "" && call.ID != part.ID {
					return false, assistantError(data.ToolErrorProtocol)
				}
				call.ID = part.ID
			}
			if part.Type != "" {
				if call.Type != "" && call.Type != part.Type {
					return false, assistantError(data.ToolErrorProtocol)
				}
				call.Type = part.Type
			}
			if len(call.Function.Arguments)+len(part.Function.Arguments) > assistantArgumentBytes || len(call.Function.Name)+len(part.Function.Name) > 256 {
				return false, assistantError(data.ToolErrorProtocol)
			}
			call.Function.Name += part.Function.Name
			call.Function.Arguments += part.Function.Arguments
		}
		if choice.FinishReason != nil {
			ordered := make([]assistantWireCall, 0, len(calls))
			for i := 0; i < len(calls); i++ {
				call := calls[i]
				if call == nil {
					return false, assistantError(data.ToolErrorProtocol)
				}
				ordered = append(ordered, *call)
			}
			result, err = finishChatToolTurn(content.String(), ordered, *choice.FinishReason)
			if err != nil {
				return false, err
			}
			finished = true
		}
		return false, nil
	})
	if err != nil {
		if result.OutputLimited {
			return result, err
		}
		return data.TurnResponse{}, err
	}
	return result, nil
}
