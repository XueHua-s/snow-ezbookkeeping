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

const assistantCompactToolRules = `Tool arguments: positive integer-string IDs; null=unrestricted, empty ID arrays invalid; parent IDs include descendants. Times: RFC3339 seconds [start,end). tag_match:any(default)/all requires tag_ids; no_tags conflicts. transfer_side:either(default)/source/destination; account/currency/amount match the same side. Signed decimal amount bounds are inclusive (max 2 decimals), require currency. Metadata keyword searches names; filter.keyword only comments; case-insensitive literal %, _, backslash. Cursors: private, tool-bound, expire in 24h; explicit continuation only, other arguments omitted or unchanged. page_size defaults 50, max 200; never limits totals. Aggregate defaults count/sum by currency/flow. Unique named periods replace filter times.`

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
