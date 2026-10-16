package services

import (
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

// The schema's executable constraints remain identical. Repeated prose is
// collected once in the system prompt to leave room for actual conversation.
// Embedding preserves Execute and NormalizeToolArguments on the same UID service.
type assistantContextTools struct {
	*AssistantQueryService
	continuation *data.ToolContinuation
}

func (t assistantContextTools) Definitions() []data.ToolDefinition {
	definitions := t.AssistantQueryService.Definitions()
	for i := range definitions {
		var schema map[string]any
		if json.Unmarshal(definitions[i].Parameters, &schema) != nil {
			continue
		}
		assistantRemoveSchemaDescriptions(schema)
		definitions[i].Parameters, _ = json.Marshal(schema)
	}
	return definitions
}
func assistantRemoveSchemaDescriptions(v any) {
	switch value := v.(type) {
	case map[string]any:
		delete(value, "description")
		for _, child := range value {
			assistantRemoveSchemaDescriptions(child)
		}
	case []any:
		for _, child := range value {
			assistantRemoveSchemaDescriptions(child)
		}
	}
}

const assistantCompactToolRules = `Shared tool argument rules (apply to every schema): IDs are positive integer strings; null means unrestricted, never use empty ID arrays. Parent IDs expand to effective descendants. Filters default to all historical/future effective records including hidden accounts; RFC3339 times are second-precision [start,end), never infer a cutoff. tag_match defaults any; all/any require tag_ids; no_tags conflicts with tag_ids. transfer_side defaults either; account/currency/amount filters match the same side; count each transfer once. Amount bounds are inclusive signed decimals, at most two fractional digits, requiring currency. Keywords are case-insensitive literal substrings, including %, _ and backslash. Cursor is private, authenticated, tool-bound, valid 24 hours; use only for explicitly requested continuation; omit other parameters or match its original query. page_size defaults 50 (maximum 200), limiting pages only, never totals. Aggregate metrics default count/sum across the complete matching ledger, separated by currency/flow. Named periods must be unique nonempty [start,end) ranges and replace filter time bounds.`

func (t assistantContextTools) ExecuteForContext(ctx core.Context, call data.ToolCall, fits func(data.ToolResult) bool) data.ToolResult {
	resolved, e := t.resolveCall(call)
	if e != nil {
		return assistantFailure(e)
	}
	bounded := *t.AssistantQueryService
	bounded.resultFits = fits
	result := bounded.Execute(ctx, resolved)
	if result.Error != nil && result.Error.Code == data.ToolErrorResultTooLarge {
		result.Error = &data.ToolError{Code: data.ToolErrorBudgetExhausted, Message: "A complete result cannot fit the remaining model context."}
	}
	return result
}

const assistantSavedPageHandle = "@saved_page"

func (t assistantContextTools) resolveCall(call data.ToolCall) (data.ToolCall, error) {
	args, e := assistantStrictJSON(call.Arguments)
	if e != nil {
		return call, e
	}
	cursor, _ := args["cursor"].(string)
	if len(cursor) > 0 && cursor[0] == '@' {
		if cursor != assistantSavedPageHandle || t.continuation == nil || call.Name != t.continuation.Tool {
			return call, assistantCursorError()
		}
		args["cursor"] = t.continuation.Cursor
		raw, e := json.Marshal(args)
		call.Arguments = raw
		return call, e
	}
	return call, nil
}
func (t assistantContextTools) NormalizeToolArguments(call data.ToolCall) (json.RawMessage, error) {
	resolved, e := t.resolveCall(call)
	if e != nil {
		return nil, e
	}
	return t.AssistantQueryService.NormalizeToolArguments(resolved)
}
func (t assistantContextTools) Execute(ctx core.Context, call data.ToolCall) data.ToolResult {
	resolved, e := t.resolveCall(call)
	if e != nil {
		return assistantFailure(e)
	}
	return t.AssistantQueryService.Execute(ctx, resolved)
}
