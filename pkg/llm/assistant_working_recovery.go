package llm

import (
	"bytes"
	"encoding/json"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
)

var assistantArchivedProjection = json.RawMessage(`{"format":"archived_checkpoint","unavailable_fields":["data","comparison","references"],"instruction":"Only retained scope, totals, counts and coverage are available evidence. Previous page rows and comparisons are archived, not currently available to cite. Repeat the original tool call to read them again; never infer missing rows."}`)

// recoverWorkingContext changes only the model working set. Exact complete
// totals and provenance survive; the execution cache and UI evidence are never
// changed. Two recovery generations and one cache reread per archived query bound
// repeated fetch/archive loops without making a failed query look successful.
func (run *assistantRun) recoverWorkingContext(input data.ToolTurnInput) (data.ToolTurnInput, bool) {
	if run.workingRecoveries >= 2 || run.c.Err() != nil {
		return input, false
	}
	messages := append([]data.ConversationMessage(nil), input.Messages...)
	changed := false
	for i, message := range messages {

		original := message.ToolResult
		if original == nil || !original.OK || bytes.Equal(original.Data, assistantArchivedProjection) || len(original.Data)+len(original.Comparison) <= len(assistantArchivedProjection) {
			continue
		}
		checkpoint := *original
		checkpoint.Data = assistantArchivedProjection
		checkpoint.Comparison = nil
		checkpoint.References = nil
		messages[i].ToolResult = &checkpoint
		if identity, ok := run.ids[message.ToolCallID]; ok {
			if run.archivedResults == nil {
				run.archivedResults = map[[32]byte]bool{}
			}
			run.archivedResults[identity.key] = true
		}
		changed = true
	}
	if !changed {
		return input, false
	}
	input.Messages, input.State = messages, nil
	run.workingRecoveries++
	run.observation.ContextRecoveryAttempts++
	return input, true
}

// executeTurnTools owns page allocation and bounded context recovery for a whole
// parallel-call group. Placeholders ensure every provider budget check sees valid
// call/result pairs, including when earlier results must be checkpointed.
func (run *assistantRun) executeTurnTools(p provider.LargeLanguageModelToolProvider, switched bool, turn data.TurnResponse, input data.ToolTurnInput, allowed map[data.ToolName]bool, keys [][32]byte) ([]data.ConversationMessage, error) {
	messages, state := input.Messages, input.State
	for i, call := range turn.ToolCalls {
		if err := run.contextError(); err != nil {
			return messages, err
		}
		complete := func(result *data.ToolResult) data.ToolTurnInput {
			candidate := append([]data.ConversationMessage(nil), messages...)
			for j, pending := range turn.ToolCalls[i:] {
				placeholder := data.ToolResult{Error: &data.ToolError{Code: data.ToolErrorBudgetExhausted, Message: "Reserved pending result"}}
				if j == 0 && result != nil {
					placeholder = *result
				}
				candidate = append(candidate, data.ConversationMessage{Role: data.ConversationRoleTool, ToolCallID: pending.ID, ToolResult: &placeholder})
			}
			request := input
			request.Messages, request.State = candidate, state
			return request
		}
		allocation := assistantAllocateToolContext(p, complete(nil), len(turn.ToolCalls)-i)
		allocated := false
		recover := func() bool {
			rebuilt, changed := run.recoverWorkingContext(complete(nil))
			if !changed {
				return false
			}
			// Drop only the synthetic pending placeholders, retaining all actual pairs.
			messages = rebuilt.Messages[:len(messages)]
			state = nil
			allocation = assistantAllocateToolContext(p, complete(nil), len(turn.ToolCalls)-i)
			allocated = false
			return true
		}
		fits := func(result data.ToolResult) bool {
			if result.OK && !allocated {
				// Complete totals and provenance are mandatory, not pageable data.
				// Reserve their envelope first, then allocate a stable row-page share.
				envelope := result
				envelope.Data, envelope.Comparison, envelope.References = nil, nil, nil
				allocation = assistantAllocateToolContext(p, complete(&envelope), len(turn.ToolCalls)-i)
				allocated = true
			}
			candidate := complete(&result)
			candidate.MaxInputTokens, candidate.MaxRequestBytes = allocation.MaxInputTokens, allocation.MaxRequestBytes
			candidate = compactAssistantWorkingContext(candidate)
			if !assistantToolRequestFits(p, candidate) {
				return false
			}
			if !switched && run.runner.fallback != nil {
				candidate.State = nil
				if !assistantToolRequestFits(run.runner.fallback, candidate) {
					return false
				}
			}
			return true
		}
		result, err := run.execute(call, keys[i], allowed[call.Name], fits, recover)
		if err != nil {
			return messages, err
		}
		messages = append(messages, data.ConversationMessage{Role: data.ConversationRoleTool, ToolCallID: call.ID, ToolResult: &result})
	}
	return messages, nil
}
