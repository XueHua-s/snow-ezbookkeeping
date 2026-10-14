package llm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// ErrAssistantWriteFailed lets the transport cancel ongoing work when an
// asynchronous heartbeat fails, while preserving the runner's single outcome.
var ErrAssistantWriteFailed = errors.New("assistant response write failed")

// AssistantRunner owns one request's model loop. It never chooses a UID-specific
// datastore: the caller supplies an executor already bound to the authenticated UID.
// A runner is reusable; all conversation, cache and provider state stays in Run.
type AssistantRunner struct {
	primary, fallback provider.LargeLanguageModelToolProvider
	executor          ToolExecutor
	now               func() time.Time
}

// NewAssistantRunner accepts an optional clock for deterministic observations.
func NewAssistantRunner(primary, fallback provider.LargeLanguageModelToolProvider, executor ToolExecutor, clock ...func() time.Time) *AssistantRunner {
	now := time.Now
	if len(clock) > 0 && clock[0] != nil {
		now = clock[0]
	}
	return &AssistantRunner{primary: primary, fallback: fallback, executor: executor, now: now}
}

// Domain normalization is optional so the loop can run with simple executors.
// It must not access the database or replace Execute's authorization checks.
type assistantArgumentNormalizer interface {
	NormalizeToolArguments(call data.ToolCall) (json.RawMessage, error)
}

type assistantCachedResult struct {
	result data.ToolResult
	reuses int
}
type assistantCallIdentity struct {
	key    [32]byte
	result data.ToolResult
}

type assistantRun struct {
	runner       *AssistantRunner
	input        AssistantRunInput
	result       AssistantRunResult
	observation  AssistantObservation
	started      time.Time
	c            core.Context
	cancel       context.CancelFunc
	sink         EventSink
	writeErr     error
	visible      bool
	cache        map[[32]byte]*assistantCachedResult
	ids          map[string]assistantCallIdentity
	unresolved   map[data.ToolName]*data.ToolError
	queryIDs     map[string]bool
	referenceIDs map[string]bool
	lastFailure  string
}

// Run buffers assistant text until a complete tool-free terminal turn. Errors
// retain confirmed evidence; callers own serialization of the sole terminal event.
func (r *AssistantRunner) Run(c core.Context, input AssistantRunInput, sink EventSink) (result AssistantRunResult, err error) {
	defaults := (&settings.Config{}).GetAIAssistantRuntimeConfig()
	if input.MaxRounds == 0 {
		input.MaxRounds = defaults.MaxRounds
	}
	if input.RunTimeout == 0 {
		input.RunTimeout = time.Duration(defaults.RunTimeout) * time.Millisecond
	}
	if input.QueryTimeout == 0 {
		input.QueryTimeout = time.Duration(defaults.QueryTimeout) * time.Millisecond
	}
	if c == nil {
		c = core.NewAssistantContext(context.Background(), "", "", "")
	}
	parent, cancel := context.WithTimeout(c, input.RunTimeout)
	run := &assistantRun{runner: r, input: input, started: r.now(), c: core.NewAssistantContext(parent, c.GetContextId(), c.GetClientLocale(), c.ClientIP()), cancel: cancel, sink: sink,
		result:      AssistantRunResult{Mode: input.Mode, State: AssistantStateRequesting},
		observation: AssistantObservation{Kind: AssistantObservationRun, RequestID: c.GetContextId()},
		cache:       map[[32]byte]*assistantCachedResult{}, ids: map[string]assistantCallIdentity{}, unresolved: map[data.ToolName]*data.ToolError{}, queryIDs: map[string]bool{}, referenceIDs: map[string]bool{}}
	defer func() {
		if input.Finalize != nil {
			if writeErr := input.Finalize(run.result, err); writeErr != nil {
				run.writeErr = writeErr
				err = run.writeFailure()
			}
		}
		cancel()
		run.observation.DurationMilliseconds = r.now().Sub(run.started).Milliseconds()
		if input.FirstStatusWrittenAt != nil {
			if at := input.FirstStatusWrittenAt(); at != nil {
				ms := at.Sub(run.started).Milliseconds()
				if ms < 0 {
					ms = 0
				}
				run.observation.FirstStatusMilliseconds = &ms
			}
		}
		if input.Observe != nil {
			input.Observe(run.observation)
		}
		result = run.result
	}()
	if input.UID <= 0 || input.RunTimeout < 0 || input.QueryTimeout < 0 || r.executor == nil {
		return run.result, run.stop(data.ToolErrorInvalidArgument, AssistantReasonInvalidArgument)
	}
	if r.primary == nil {
		return run.result, run.stop(data.ToolErrorProviderUnsupported, AssistantReasonProviderUnsupported)
	}
	messages := append([]data.ConversationMessage(nil), input.Messages...)
	definitions := r.executor.Definitions()
	allowed := map[data.ToolName]bool{}
	for _, definition := range definitions {
		if assistantKnownTool(definition.Name) {
			allowed[definition.Name] = true
		}
	}
	current := r.primary
	switched := false
	var state any
	calls := 0
	for {
		if err = run.contextError(); err != nil {
			return run.result, err
		}
		if run.observation.Rounds >= input.MaxRounds {
			return run.result, run.stop(data.ToolErrorBudgetExhausted, AssistantReasonRoundBudget)
		}
		turnInput := data.ToolTurnInput{Messages: messages, Tools: definitions, State: state, MaxRequestBytes: AssistantMaxModelRequestBytes}
		body, marshalErr := json.Marshal(turnInput)
		if marshalErr != nil {
			return run.result, run.stop(data.ToolErrorProtocol, AssistantReasonProtocolError)
		}
		if len(body) > AssistantMaxModelRequestBytes {
			return run.result, run.stop(data.ToolErrorBudgetExhausted, AssistantReasonContextBudget)
		}
		if err = run.emit(AssistantEvent{Type: AssistantEventStatus, Status: AssistantStatusQuerying}); err != nil {
			return run.result, run.writeFailure()
		}
		var display data.DisplayEventSink
		if sink != nil {
			display = func(event data.DisplayEvent) error {
				if event.ThinkingDelta == "" {
					return nil
				}
				if e := run.emit(AssistantEvent{Type: AssistantEventThinkingDelta, Delta: event.ThinkingDelta}); e != nil {
					return e
				}
				run.result.Thinking += event.ThinkingDelta
				return nil
			}
		}
		run.observation.Rounds++
		turn, turnErr := current.RunToolTurn(run.c, input.UID, turnInput, display)
		if run.writeErr != nil {
			return run.result, run.writeFailure()
		}
		if err = run.contextError(); err != nil {
			return run.result, err
		}
		if turnErr == nil && !turn.Completed {
			turnErr = assistantRunnerError(data.ToolErrorProtocol)
		}
		if turnErr != nil {
			if errors.Is(turnErr, context.Canceled) {
				return run.result, run.stop(data.ToolErrorCanceled, AssistantReasonCanceled)
			}
			if errors.Is(turnErr, context.DeadlineExceeded) {
				// The run context is still alive: this is a provider HTTP timeout,
				// eligible for the same one-time fallback as other transport failures.
				turnErr = assistantRunnerError(data.ToolErrorUpstreamFailed)
			}
			code := assistantErrorCode(turnErr)
			if !switched && !run.visible && r.fallback != nil && (code == data.ToolErrorUpstreamFailed || code == data.ToolErrorProtocol) {
				current = r.fallback
				switched = true
				state = nil
				continue
			}
			return run.result, run.stop(code, assistantReason(code))
		}
		if len(turn.ToolCalls) == 0 {
			// A model's completed text cannot turn a failed financial query into an
			// empty ledger or a verified answer. A corrected success of that tool must
			// resolve the error before this request can complete.
			if len(run.unresolved) > 0 {
				for _, failure := range run.unresolved {
					return run.result, run.stop(failure.Code, assistantReason(failure.Code))
				}
			}
			if err = run.emit(AssistantEvent{Type: AssistantEventStatus, Status: AssistantStatusAnswering}); err != nil {
				return run.result, run.writeFailure()
			}
			if len(run.result.References) > 0 {
				if err = run.emit(AssistantEvent{Type: AssistantEventReferences, References: run.result.References}); err != nil {
					return run.result, run.writeFailure()
				}
			}
			if turn.Content != "" {
				if err = run.emit(AssistantEvent{Type: AssistantEventReplyDelta, Delta: turn.Content}); err != nil {
					return run.result, run.writeFailure()
				}
			}
			run.result.Reply = turn.Content
			run.result.State = AssistantStateCompleted
			run.observation.Reason = AssistantReasonCompleted
			return run.result, nil
		}
		if len(turn.ToolCalls) > AssistantMaxToolCalls-calls {
			return run.result, run.stop(data.ToolErrorBudgetExhausted, AssistantReasonCallBudget)
		}
		keys := make([][32]byte, len(turn.ToolCalls))
		roundIDs := map[string]bool{}
		for i, call := range turn.ToolCalls {
			keys[i] = r.callKey(call)
			prior, exists := run.ids[call.ID]
			if call.ID == "" || roundIDs[call.ID] || (exists && prior.key != keys[i]) {
				return run.result, run.stop(data.ToolErrorProtocol, AssistantReasonProtocolError)
			}
			roundIDs[call.ID] = true
		}
		calls += len(turn.ToolCalls)
		// Preserve this exact assistant prefix for same-adapter opaque state replay.
		messages = append(messages, data.ConversationMessage{Role: data.ConversationRoleAssistant, Content: turn.Content, ToolCalls: turn.ToolCalls})
		for i, call := range turn.ToolCalls {
			if err = run.contextError(); err != nil {
				return run.result, err
			}
			result, executionErr := run.execute(call, keys[i], allowed[call.Name])
			if executionErr != nil {
				return run.result, executionErr
			}
			messages = append(messages, data.ConversationMessage{Role: data.ConversationRoleTool, ToolCallID: call.ID, ToolResult: &result})
		}
		state = turn.State
	}
}

func (run *assistantRun) emit(event AssistantEvent) error {
	if run.writeErr != nil {
		return run.writeErr
	}
	if run.sink == nil {
		return nil
	}
	if err := run.sink(event); err != nil {
		run.writeErr = err
		run.cancel()
		return err
	}
	if (event.Type == AssistantEventThinkingDelta || event.Type == AssistantEventReplyDelta) && event.Delta != "" {
		run.visible = true
	}
	return nil
}
func (run *assistantRun) writeFailure() error {
	run.result.State = AssistantStateCanceled
	run.observation.Reason = AssistantReasonWriteFailed
	return run.writeErr
}
func (run *assistantRun) contextError() error {
	if err := run.c.Err(); err != nil {
		if errors.Is(context.Cause(run.c), ErrAssistantWriteFailed) {
			run.stop(data.ToolErrorCanceled, AssistantReasonWriteFailed)
		} else if errors.Is(err, context.DeadlineExceeded) {
			run.stop(data.ToolErrorQueryTimeout, AssistantReasonRunTimeout)
		} else {
			run.stop(data.ToolErrorCanceled, AssistantReasonCanceled)
		}
		return err
	}
	return nil
}
func (run *assistantRun) stop(code data.ToolErrorCode, reason AssistantStopReason) error {
	run.observation.Reason = reason
	run.result.State = AssistantStateIncomplete
	switch code {
	case data.ToolErrorCanceled:
		run.result.State = AssistantStateCanceled
	case data.ToolErrorProviderUnsupported, data.ToolErrorUpstreamFailed, data.ToolErrorProtocol:
		run.result.State = AssistantStateFailed
	}
	return assistantRunnerError(code)
}

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
func assistantCanonicalArguments(raw []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value func() (any, error)
	value = func() (any, error) {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return token, nil
		}
		switch delimiter {
		case '{':
			object := map[string]any{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, io.ErrUnexpectedEOF
				}
				if _, exists := object[key]; exists {
					return nil, io.ErrUnexpectedEOF
				}
				item, err := value()
				if err != nil {
					return nil, err
				}
				object[key] = item
			}
			_, err = decoder.Token()
			return object, err
		case '[':
			array := []any{}
			for decoder.More() {
				item, err := value()
				if err != nil {
					return nil, err
				}
				array = append(array, item)
			}
			_, err = decoder.Token()
			return array, err
		default:
			return nil, io.ErrUnexpectedEOF
		}
	}
	result, err := value()
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, io.ErrUnexpectedEOF
	}
	return json.Marshal(result)
}
func assistantKnownTool(name data.ToolName) bool {
	switch name {
	case data.ToolListAccounts, data.ToolListCategories, data.ToolListTags, data.ToolQueryTransactions, data.ToolGetTransaction, data.ToolAggregateTransactions, data.ToolGetAccountBalances:
		return true
	}
	return false
}
func assistantErrorCode(err error) data.ToolErrorCode {
	var controlled *data.ToolError
	if errors.As(err, &controlled) {
		switch controlled.Code {
		case data.ToolErrorInvalidArgument, data.ToolErrorNotFoundOrNotAccessible, data.ToolErrorQueryTimeout, data.ToolErrorCanceled, data.ToolErrorBudgetExhausted, data.ToolErrorProviderUnsupported, data.ToolErrorUpstreamFailed, data.ToolErrorResultTooLarge, data.ToolErrorInvalidCursor, data.ToolErrorDataInconsistent, data.ToolErrorProtocol:
			return controlled.Code
		}
	}
	return data.ToolErrorUpstreamFailed
}
func assistantRunnerError(code data.ToolErrorCode) *data.ToolError {
	return &data.ToolError{Code: code, Message: "The assistant could not complete this request.", Retryable: code == data.ToolErrorQueryTimeout || code == data.ToolErrorUpstreamFailed}
}
func assistantReason(code data.ToolErrorCode) AssistantStopReason {
	switch code {
	case data.ToolErrorInvalidArgument:
		return AssistantReasonInvalidArgument
	case data.ToolErrorNotFoundOrNotAccessible:
		return AssistantReasonNotAccessible
	case data.ToolErrorQueryTimeout:
		return AssistantReasonQueryTimeout
	case data.ToolErrorCanceled:
		return AssistantReasonCanceled
	case data.ToolErrorBudgetExhausted:
		return AssistantReasonContextBudget
	case data.ToolErrorProviderUnsupported:
		return AssistantReasonProviderUnsupported
	case data.ToolErrorResultTooLarge:
		return AssistantReasonResultBudget
	case data.ToolErrorInvalidCursor:
		return AssistantReasonInvalidCursor
	case data.ToolErrorDataInconsistent:
		return AssistantReasonDataInconsistent
	case data.ToolErrorProtocol:
		return AssistantReasonProtocolError
	default:
		return AssistantReasonUpstreamFailed
	}
}
