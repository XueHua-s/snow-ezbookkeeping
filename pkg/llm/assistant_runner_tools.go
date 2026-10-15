package llm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

func (run *assistantRun) execute(call data.ToolCall, key [32]byte, allowed bool) (data.ToolResult, error) {
	start := run.runner.now()
	var result data.ToolResult
	cached := false
	if previous, exists := run.ids[call.ID]; exists {
		result = previous.result
		cached = true
	} else if previous := run.cache[key]; previous != nil {
		result = previous.result
		cached = true
	}
	if cached {
		previous := run.cache[key]
		if result.OK && previous != nil && previous.reuses >= 1 {
			return result, run.stop(data.ToolErrorBudgetExhausted, AssistantReasonNoProgress)
		}
		if previous != nil {
			previous.reuses++
		}
	} else if !allowed {
		result = data.ToolResult{Error: assistantRunnerError(data.ToolErrorInvalidArgument)}
	} else {
		ctx, cancel := context.WithTimeout(run.c, run.input.QueryTimeout)
		queryContext := core.NewAssistantContext(ctx, run.c.GetContextId(), run.c.GetClientLocale(), run.c.ClientIP())
		result = run.runner.executor.Execute(queryContext, call)
		queryErr := ctx.Err()
		cancel()
		if queryErr != nil {
			result = data.ToolResult{Error: assistantRunnerError(data.ToolErrorQueryTimeout), Metrics: result.Metrics}
			if errors.Is(queryErr, context.Canceled) {
				result.Error = assistantRunnerError(data.ToolErrorCanceled)
			}
		}
	}
	if result.OK && result.Error != nil || !result.OK && result.Error == nil {
		result = data.ToolResult{Error: assistantRunnerError(data.ToolErrorProtocol), Metrics: result.Metrics}
	}
	encoded, marshalErr := json.Marshal(result)
	byteCount := int64(len(encoded))
	reason := AssistantReasonCompleted
	if result.Error != nil {
		reason = assistantReason(result.Error.Code)
	}
	if marshalErr != nil {
		reason = AssistantReasonProtocolError
	} else if byteCount > AssistantMaxToolResultBytes {
		reason = AssistantReasonResultBudget
	}
	observation := AssistantObservation{Kind: AssistantObservationTool, RequestID: run.c.GetContextId(), SerializedBytes: byteCount, DurationMilliseconds: run.runner.now().Sub(start).Milliseconds(), Reason: reason}
	if assistantKnownTool(call.Name) {
		observation.Tool = call.Name
	}
	if cached {
		observation.CacheHits = 1
		run.observation.CacheHits++
	} else if allowed {
		observation.Executions = 1
		observation.RowsRead = result.Metrics.RowsRead
		observation.ResultRecords = result.Metrics.ResultRecords
		run.observation.Executions++
		run.observation.RowsRead += observation.RowsRead
		run.observation.ResultRecords += observation.ResultRecords
	}
	run.observation.SerializedBytes += byteCount
	if run.input.Observe != nil {
		run.input.Observe(observation)
	}
	if err := run.contextError(); err != nil {
		return result, err
	}
	if marshalErr != nil {
		return result, run.stop(data.ToolErrorProtocol, AssistantReasonProtocolError)
	}
	if result.OK {
		run.evidence(call.Name, result)
	}
	if byteCount > AssistantMaxToolResultBytes {
		return result, run.stop(data.ToolErrorResultTooLarge, AssistantReasonResultBudget)
	}
	run.ids[call.ID] = assistantCallIdentity{key: key, result: result}
	if result.OK {
		if run.cache[key] == nil {
			run.cache[key] = &assistantCachedResult{result: result}
		}
		delete(run.unresolved, call.Name)
		run.lastFailure = ""
	} else {
		code := assistantErrorCode(result.Error)
		run.unresolved[call.Name] = assistantRunnerError(code)
		failure := string(key[:]) + string(code)
		if failure == run.lastFailure {
			return result, run.stop(data.ToolErrorBudgetExhausted, AssistantReasonRepeatedError)
		}
		run.lastFailure = failure
		switch code {
		case data.ToolErrorCanceled, data.ToolErrorQueryTimeout, data.ToolErrorResultTooLarge, data.ToolErrorBudgetExhausted, data.ToolErrorProtocol, data.ToolErrorDataInconsistent:
			return result, run.stop(code, assistantReason(code))
		}
	}
	return result, nil
}

func (run *assistantRun) evidence(tool data.ToolName, result data.ToolResult) {
	if result.QueryID != "" && result.Scope != nil && !run.queryIDs[result.QueryID] {
		run.queryIDs[result.QueryID] = true
		run.result.Queries = append(run.result.Queries, data.QueryEvidence{QueryID: result.QueryID, Tool: tool, Scope: *result.Scope, Coverage: result.Coverage, QueriedAt: result.QueriedAt, Warnings: result.Warnings})
	}
	for _, reference := range result.References {
		if reference.ID != "" && !run.referenceIDs[reference.ID] && len(run.result.References) < AssistantMaxReferences {
			run.referenceIDs[reference.ID] = true
			run.result.References = append(run.result.References, reference)
		}
	}
	if tool != data.ToolGetTransaction {
		run.result.Continuation = nil
		if result.HasMore && result.NextCursor != "" {
			run.result.Continuation = &data.ToolContinuation{Tool: tool, Cursor: result.NextCursor}
		}
	}
}

func (r *AssistantRunner) callKey(call data.ToolCall) [32]byte {
	arguments := call.Arguments
	var normalizedDomain byte
	if normalizer, ok := r.executor.(assistantArgumentNormalizer); ok {
		if normalized, err := normalizer.NormalizeToolArguments(call); err == nil {
			arguments = normalized
			normalizedDomain = 1
		}
	}
	// JSON v1 accepts duplicate object keys. Detect them before canonicalizing so
	// invalid input never aliases a previously authorized cached success.
	if canonical, err := assistantCanonicalArguments(arguments); err == nil {
		arguments = canonical
	}
	// Failed domain validation must never collide with a successful normalized
	// key, even if hostile raw arguments reproduce the normalizer's output.
	prefix := append([]byte(call.Name), 0, normalizedDomain)
	return sha256.Sum256(append(prefix, arguments...))
}
