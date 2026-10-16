package llm

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"time"
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
		observation: AssistantObservation{Kind: AssistantObservationRun, RequestID: c.GetContextId(), Rounds: input.PriorModelCalls},
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
		requestBytes := input.MaxRequestBytes
		if requestBytes <= 0 || requestBytes > AssistantMaxModelRequestBytes {
			requestBytes = AssistantMaxModelRequestBytes
		}
		turnInput := data.ToolTurnInput{Messages: messages, Tools: definitions, State: state, MaxRequestBytes: requestBytes, MaxInputTokens: input.MaxInputTokens, MaxOutputTokens: input.MaxOutputTokens}
		body, marshalErr := json.Marshal(turnInput)
		if marshalErr != nil {
			return run.result, run.stop(data.ToolErrorProtocol, AssistantReasonProtocolError)
		}
		if !data.AssistantRequestFits(body, input.MaxInputTokens, requestBytes) {
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
			if turn.OutputLimited {
				if len(turn.ToolCalls) == 0 && len(run.unresolved) == 0 && turn.Content != "" {
					run.result.Reply = turn.Content
					if e := run.emit(AssistantEvent{Type: AssistantEventReplyDelta, Delta: turn.Content}); e != nil {
						return run.result, run.writeFailure()
					}
				}
				return run.result, run.stop(data.ToolErrorBudgetExhausted, AssistantReasonOutputBudget)
			}
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
			var fits func(data.ToolResult) bool
			if input.MaxInputTokens > 0 {
				fits = func(result data.ToolResult) bool {
					candidate := append(append([]data.ConversationMessage(nil), messages...), data.ConversationMessage{Role: data.ConversationRoleTool, ToolCallID: call.ID, ToolResult: &result})
					// Supply bounded placeholders for the other calls in this same assistant
					// turn so protocol validation sees a complete call/result chain.
					for _, pending := range turn.ToolCalls[i+1:] {
						placeholder := data.ToolResult{OK: false, Error: &data.ToolError{Code: data.ToolErrorBudgetExhausted, Message: "Reserved pending result"}}
						candidate = append(candidate, data.ConversationMessage{Role: data.ConversationRoleTool, ToolCallID: pending.ID, ToolResult: &placeholder})
					}
					next := data.ToolTurnInput{Messages: candidate, Tools: definitions, State: turn.State, MaxRequestBytes: requestBytes, MaxInputTokens: input.MaxInputTokens, MaxOutputTokens: input.MaxOutputTokens}
					if !assistantToolRequestFits(current, next) {
						return false
					}
					if !switched && r.fallback != nil {
						next.State = nil
						if !assistantToolRequestFits(r.fallback, next) {
							return false
						}
					}
					return true
				}
			}
			result, executionErr := run.execute(call, keys[i], allowed[call.Name], fits)
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
	run.result.StopReason = reason
	run.result.State = AssistantStateIncomplete
	switch code {
	case data.ToolErrorCanceled:
		run.result.State = AssistantStateCanceled
	case data.ToolErrorProviderUnsupported, data.ToolErrorUpstreamFailed, data.ToolErrorProtocol:
		run.result.State = AssistantStateFailed
	}
	return assistantRunnerError(code)
}

func assistantToolRequestFits(p provider.LargeLanguageModelToolProvider, input data.ToolTurnInput) bool {
	if counter, ok := p.(interface {
		CheckToolTurnInput(data.ToolTurnInput) error
	}); ok {
		err := counter.CheckToolTurnInput(input)
		return err == nil
	}
	body, e := json.Marshal(input)
	return e == nil && data.AssistantRequestFits(body, input.MaxInputTokens, input.MaxRequestBytes)
}
