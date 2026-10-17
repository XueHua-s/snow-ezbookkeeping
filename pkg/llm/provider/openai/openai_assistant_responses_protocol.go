package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

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
		return zero, nil, assistantContextError()
	}
	if json.Unmarshal(b, &envelope) == nil && envelope.Status == "incomplete" {
		var details struct {
			Reason string `json:"reason"`
		}
		if json.Unmarshal(envelope.IncompleteDetails, &details) == nil && details.Reason == "max_output_tokens" {
			return assistantLimitedResponse(envelope.Output), envelope.Output, assistantError(data.ToolErrorBudgetExhausted)
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
	text          string
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
	var displayed strings.Builder
	toolSeen := false
	err := readAssistantSSE(r, func(b []byte) (bool, error) {
		if isContextLimit(b) {
			return false, assistantContextError()
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
		case "error", "response.failed":
			return false, assistantError(data.ToolErrorProtocol)
		case "response.output_text.delta":
			if event.OutputIndex == nil || *event.OutputIndex < 0 || *event.OutputIndex >= 1024 || event.ItemID == "" {
				return false, assistantError(data.ToolErrorProtocol)
			}
			current := items[*event.OutputIndex]
			if current == nil {
				current = &responseStreamItem{id: event.ItemID}
				items[*event.OutputIndex] = current
			}
			if current.id != event.ItemID || current.done != nil || (current.added != nil && (current.added.Type != "message" || current.added.Role != "assistant")) {
				return false, assistantError(data.ToolErrorProtocol)
			}
			current.text += event.Delta
			if !toolSeen && event.Delta != "" {
				displayed.WriteString(event.Delta)
				if emit != nil {
					if err := emit(data.DisplayEvent{ReplyDelta: event.Delta}); err != nil {
						return false, &assistantSinkError{err}
					}
				}
			}
		case "response.reasoning_summary_text.delta":
			if emit != nil && event.Delta != "" {
				if err := emit(data.DisplayEvent{ThinkingDelta: event.Delta}); err != nil {
					return false, &assistantSinkError{err}
				}
			}
		case "response.output_item.added", "response.output_item.done", "response.function_call_arguments.delta", "response.function_call_arguments.done":
			if strings.HasPrefix(event.Type, "response.function_call_arguments.") && !toolSeen {
				toolSeen = true
				if emit != nil {
					if err := emit(data.DisplayEvent{ReplyReset: true}); err != nil {
						return false, &assistantSinkError{err}
					}
				}
			}
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
				if item.Type == "function_call" && !toolSeen {
					toolSeen = true
					if emit != nil {
						if err := emit(data.DisplayEvent{ReplyReset: true}); err != nil {
							return false, &assistantSinkError{err}
						}
					}
				}
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
		case "response.completed", "response.incomplete":
			// Terminal-only tool calls also invalidate provisional text, including
			// incomplete turns where the arguments have not finished assembling.
			var envelope struct {
				Output []json.RawMessage `json:"output"`
			}
			if json.Unmarshal(event.Response, &envelope) != nil {
				return false, assistantError(data.ToolErrorProtocol)
			}
			for _, raw := range envelope.Output {
				var item assistantResponseItem
				if json.Unmarshal(raw, &item) != nil {
					return false, assistantError(data.ToolErrorProtocol)
				}
				if item.Type == "function_call" && !toolSeen {
					toolSeen = true
					if emit != nil {
						if err := emit(data.DisplayEvent{ReplyReset: true}); err != nil {
							return false, &assistantSinkError{err}
						}
					}
				}
			}
			terminal, terminalOutputs, parseErr := parseResponsesToolTurn(event.Response)
			if parseErr != nil && !terminal.OutputLimited {
				return false, parseErr
			}
			if !toolSeen && !strings.HasPrefix(terminal.Content, displayed.String()) {
				return false, assistantError(data.ToolErrorProtocol)
			}
			indices := make([]int, 0, len(items))
			for idx := range items {
				indices = append(indices, idx)
			}
			sort.Ints(indices)
			for _, idx := range indices {
				if idx >= len(terminalOutputs) {
					return false, assistantError(data.ToolErrorProtocol)
				}
				var item assistantResponseItem
				if json.Unmarshal(terminalOutputs[idx], &item) != nil {
					return false, assistantError(data.ToolErrorProtocol)
				}
				current := items[idx]
				var itemText string
				for _, part := range item.Content {
					if part.Type == "output_text" {
						itemText += part.Text
					}
				}
				if !strings.HasPrefix(itemText, current.text) {
					return false, assistantError(data.ToolErrorProtocol)
				}
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
			result, outputs = terminal, terminalOutputs
			if parseErr != nil {
				return false, parseErr
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
