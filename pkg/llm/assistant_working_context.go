package llm

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
)

// compactAssistantWorkingContext removes repeated object keys, never records or
// facts. Original results remain in the execution cache/evidence. Rebuilding the
// neutral conversation invalidates adapter-owned replay state, but preserves every
// system/user message and complete call/result pair.
func compactAssistantWorkingContext(input data.ToolTurnInput) data.ToolTurnInput {
	messages := append([]data.ConversationMessage(nil), input.Messages...)
	changed := false
	for i, message := range messages {
		if message.ToolResult == nil {
			continue
		}
		result := *message.ToolResult
		resultChanged := false
		for _, projection := range []*json.RawMessage{&result.Data, &result.Comparison, &result.Totals} {
			compact := compactAssistantData(*projection)
			if !bytes.Equal(compact, *projection) {
				*projection = compact
				resultChanged = true
			}
		}
		if resultChanged {
			messages[i].ToolResult = &result
			changed = true
		}
	}
	if changed {
		input.Messages, input.State = messages, nil
	}
	return input
}

// Tabular arrays use explicit columns and rows, with an exact positional mapping.
// Only objects with identical keys qualify, so missing keys are never conflated
// with null. json.Number retains integers beyond JavaScript's exact range.
func compactAssistantData(raw json.RawMessage) json.RawMessage {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil {
		return raw
	}
	encoded, err := json.Marshal(assistantTabularValue(value))
	if err != nil || len(encoded) >= len(raw) {
		return raw
	}
	return encoded
}

func assistantTabularValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		// Already compacted tables must not be interpreted as source records again.
		if v["format"] == "column_rows" {
			return v
		}
		for key, child := range v {
			v[key] = assistantTabularValue(child)
		}
		return v
	case []any:
		for i, child := range v {
			v[i] = assistantTabularValue(child)
		}
		if len(v) < 2 {
			return v
		}
		first, ok := v[0].(map[string]any)
		if !ok || len(first) == 0 {
			return v
		}
		columns := make([]string, 0, len(first))
		for key := range first {
			columns = append(columns, key)
		}
		sort.Strings(columns)
		rows := make([][]any, len(v))
		for i, child := range v {
			object, ok := child.(map[string]any)
			if !ok || len(object) != len(columns) {
				return v
			}
			rows[i] = make([]any, len(columns))
			for j, key := range columns {
				cell, exists := object[key]
				if !exists {
					return v
				}
				rows[i][j] = cell
			}
		}
		table := map[string]any{"format": "column_rows", "columns": columns, "rows": rows}
		original, _ := json.Marshal(v)
		compact, _ := json.Marshal(table)
		if len(compact) < len(original) {
			return table
		}
		return v
	default:
		return value
	}
}

// A page gets at most one eighth of the total input budget. This stable share
// leaves room for multi-step analysis instead of consuming most remaining space.
// Search the budget, never the result length:
// serialized BPE counts are not monotonic in the number of returned records.
func assistantAllocateToolContext(p provider.LargeLanguageModelToolProvider, input data.ToolTurnInput, remaining int) data.ToolTurnInput {
	if remaining < 1 {
		return input
	}
	input = compactAssistantWorkingContext(input)
	if !assistantToolRequestFits(p, input) {
		return input
	}
	if input.MaxInputTokens > 0 {
		limit := input.MaxInputTokens
		low, high := 1, limit
		for low < high {
			middle := low + (high-low)/2
			candidate := input
			candidate.MaxInputTokens = middle
			if assistantToolRequestFits(p, candidate) {
				high = middle
			} else {
				low = middle + 1
			}
		}
		input.MaxInputTokens = low + min(limit/8, (limit-low)/(2*remaining))
	}
	if input.MaxRequestBytes > 0 {
		limit := input.MaxRequestBytes
		low, high := 1, limit
		for low < high {
			middle := low + (high-low)/2
			candidate := input
			candidate.MaxRequestBytes = middle
			if assistantToolRequestFits(p, candidate) {
				high = middle
			} else {
				low = middle + 1
			}
		}
		input.MaxRequestBytes = low + min(limit/8, (limit-low)/(2*remaining))
	}
	return input
}
