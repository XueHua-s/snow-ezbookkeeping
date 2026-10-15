package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
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

func readChatToolStream(r io.Reader) (data.TurnResponse, error) {
	var result data.TurnResponse
	var content strings.Builder
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

type assistantResponseItem struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	Content   []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func parseResponsesToolTurn(b []byte) (data.TurnResponse, []json.RawMessage, error) {
	var zero data.TurnResponse
	var envelope struct {
		Status            string            `json:"status"`
		Output            []json.RawMessage `json:"output"`
		Error             json.RawMessage   `json:"error"`
		IncompleteDetails json.RawMessage   `json:"incomplete_details"`
	}
	if isContextLimit(b) {
		return zero, nil, assistantError(data.ToolErrorBudgetExhausted)
	}
	if json.Unmarshal(b, &envelope) == nil && envelope.Status == "incomplete" {
		var details struct {
			Reason string `json:"reason"`
		}
		if json.Unmarshal(envelope.IncompleteDetails, &details) == nil && details.Reason == "max_output_tokens" {
			return assistantLimitedResponse(envelope.Output), nil, assistantError(data.ToolErrorBudgetExhausted)
		}
	}
	if json.Unmarshal(b, &envelope) != nil || envelope.Status != "completed" || (len(envelope.Error) > 0 && string(envelope.Error) != "null") || (len(envelope.IncompleteDetails) > 0 && string(envelope.IncompleteDetails) != "null") || len(envelope.Output) == 0 {
		return zero, nil, assistantError(data.ToolErrorProtocol)
	}
	turn := data.TurnResponse{Completed: true}
	ids := map[string]bool{}
	itemIDs := map[string]bool{}
	for _, raw := range envelope.Output {
		var item assistantResponseItem
		if json.Unmarshal(raw, &item) != nil || (item.Status != "" && item.Status != "completed") {
			return zero, nil, assistantError(data.ToolErrorProtocol)
		}
		if item.ID != "" {
			if itemIDs[item.ID] {
				return zero, nil, assistantError(data.ToolErrorProtocol)
			}
			itemIDs[item.ID] = true
		}
		switch item.Type {
		case "function_call":
			if item.CallID == "" || len(item.CallID) > 1024 || item.Name == "" || len(item.Name) > 256 || ids[item.CallID] || !validArguments([]byte(item.Arguments)) {
				return zero, nil, assistantError(data.ToolErrorProtocol)
			}
			ids[item.CallID] = true
			turn.ToolCalls = append(turn.ToolCalls, data.ToolCall{ID: item.CallID, Name: data.ToolName(item.Name), Arguments: json.RawMessage(item.Arguments)})
		case "message":
			if item.Role != "assistant" {
				return zero, nil, assistantError(data.ToolErrorProtocol)
			}
			for _, part := range item.Content {
				if part.Type != "output_text" {
					return zero, nil, assistantError(data.ToolErrorProtocol)
				}
				turn.Content += part.Text
			}
		case "reasoning": // retained opaquely, never displayed or logged
		default:
			return zero, nil, assistantError(data.ToolErrorProviderUnsupported)
		}
	}
	if len(turn.ToolCalls) == 0 && turn.Content == "" {
		return zero, nil, assistantError(data.ToolErrorProtocol)
	}
	return turn, envelope.Output, nil
}

type responseStreamItem struct {
	id            string
	arguments     string
	sawArguments  bool
	argumentsDone bool
	added         *assistantResponseItem
	done          *assistantResponseItem
}

func readResponsesToolStream(r io.Reader, emit data.DisplayEventSink) (data.TurnResponse, []json.RawMessage, error) {
	var result data.TurnResponse
	var outputs []json.RawMessage
	items := map[int]*responseStreamItem{}
	err := readAssistantSSE(r, func(b []byte) (bool, error) {
		if isContextLimit(b) {
			return false, assistantError(data.ToolErrorBudgetExhausted)
		}
		var event struct {
			Type        string          `json:"type"`
			OutputIndex *int            `json:"output_index"`
			ItemID      string          `json:"item_id"`
			Delta       string          `json:"delta"`
			Arguments   string          `json:"arguments"`
			Item        json.RawMessage `json:"item"`
			Response    json.RawMessage `json:"response"`
		}
		if json.Unmarshal(b, &event) != nil {
			return false, assistantError(data.ToolErrorProtocol)
		}
		switch event.Type {
		case "response.incomplete":
			var parseErr error
			result, outputs, parseErr = parseResponsesToolTurn(event.Response)
			return false, parseErr
		case "error", "response.failed":
			return false, assistantError(data.ToolErrorProtocol)
		case "response.reasoning_summary_text.delta":
			if emit != nil && event.Delta != "" {
				if err := emit(data.DisplayEvent{ThinkingDelta: event.Delta}); err != nil {
					return false, &assistantSinkError{err}
				}
			}
		case "response.output_item.added", "response.output_item.done", "response.function_call_arguments.delta", "response.function_call_arguments.done":
			if event.OutputIndex == nil || *event.OutputIndex < 0 || *event.OutputIndex >= 1024 {
				return false, assistantError(data.ToolErrorProtocol)
			}
			current := items[*event.OutputIndex]
			if current == nil {
				current = &responseStreamItem{}
				items[*event.OutputIndex] = current
			}
			if event.ItemID != "" {
				if current.id != "" && current.id != event.ItemID {
					return false, assistantError(data.ToolErrorProtocol)
				}
				current.id = event.ItemID
			}
			if event.Type == "response.function_call_arguments.delta" {
				if current.done != nil || current.argumentsDone || len(current.arguments)+len(event.Delta) > assistantArgumentBytes {
					return false, assistantError(data.ToolErrorProtocol)
				}
				current.arguments += event.Delta
				current.sawArguments = true
			} else if event.Type == "response.function_call_arguments.done" {
				if current.argumentsDone || current.done != nil || len(event.Arguments) > assistantArgumentBytes || (current.sawArguments && current.arguments != event.Arguments) {
					return false, assistantError(data.ToolErrorProtocol)
				}
				current.arguments = event.Arguments
				current.sawArguments = true
				current.argumentsDone = true
			} else {
				var item assistantResponseItem
				if json.Unmarshal(event.Item, &item) != nil || item.ID == "" || (current.id != "" && current.id != item.ID) {
					return false, assistantError(data.ToolErrorProtocol)
				}
				current.id = item.ID
				if current.added != nil && !sameResponseItemIdentity(*current.added, item) {
					return false, assistantError(data.ToolErrorProtocol)
				}
				if event.Type == "response.output_item.added" {
					if current.added != nil || current.done != nil {
						return false, assistantError(data.ToolErrorProtocol)
					}
					current.added = &item
				}
				if event.Type == "response.output_item.done" {
					if current.done != nil || (current.sawArguments && current.arguments != item.Arguments) {
						return false, assistantError(data.ToolErrorProtocol)
					}
					current.done = &item
				}
			}
		case "response.completed":
			var err error
			result, outputs, err = parseResponsesToolTurn(event.Response)
			if err != nil {
				return false, err
			}
			indices := make([]int, 0, len(items))
			for idx := range items {
				indices = append(indices, idx)
			}
			sort.Ints(indices)
			for _, idx := range indices {
				if idx >= len(outputs) {
					return false, assistantError(data.ToolErrorProtocol)
				}
				var item assistantResponseItem
				if json.Unmarshal(outputs[idx], &item) != nil {
					return false, assistantError(data.ToolErrorProtocol)
				}
				current := items[idx]
				if current.id != item.ID || (current.added != nil && !sameResponseItemIdentity(*current.added, item)) || (current.sawArguments && current.arguments != item.Arguments) {
					return false, assistantError(data.ToolErrorProtocol)
				}
				if current.done != nil {
					a, _ := json.Marshal(current.done)
					b, _ := json.Marshal(item)
					if !bytes.Equal(a, b) {
						return false, assistantError(data.ToolErrorProtocol)
					}
				}
			}
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		if result.OutputLimited {
			return result, nil, err
		}
		return data.TurnResponse{}, nil, err
	}
	return result, outputs, nil
}

func sameResponseItemIdentity(a, b assistantResponseItem) bool {
	return a.ID == b.ID && a.Type == b.Type && a.CallID == b.CallID && a.Name == b.Name && a.Role == b.Role
}

func assistantLimitedResponse(outputs []json.RawMessage) data.TurnResponse {
	result := data.TurnResponse{OutputLimited: true}
	for _, raw := range outputs {
		var item assistantResponseItem
		if json.Unmarshal(raw, &item) != nil {
			return data.TurnResponse{OutputLimited: true}
		}
		if item.Type == "function_call" {
			return data.TurnResponse{OutputLimited: true}
		}
		if item.Type != "message" {
			continue
		}
		if item.Role != "assistant" {
			return data.TurnResponse{OutputLimited: true}
		}
		for _, part := range item.Content {
			if part.Type != "output_text" {
				return data.TurnResponse{OutputLimited: true}
			}
			result.Content += part.Text
		}
	}
	return result
}
