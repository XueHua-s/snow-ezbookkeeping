package llm

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"strings"
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
	runner            *AssistantRunner
	input             AssistantRunInput
	result            AssistantRunResult
	observation       AssistantObservation
	started           time.Time
	c                 core.Context
	cancel            context.CancelFunc
	sink              EventSink
	writeErr          error
	visible           bool
	cache             map[[32]byte]*assistantCachedResult
	ids               map[string]assistantCallIdentity
	unresolved        map[[32]byte]assistantUnresolvedCall
	toolRound         int
	queryIDs          map[string]bool
	referenceIDs      map[string]bool
	lastFailure       string
	workingRecoveries int
	archivedResults   map[[32]byte]bool
	recoveredReads    map[[32]byte]bool
}

// Run streams provisional assistant text, retracting tool-turn preambles. Errors
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
		cache:       map[[32]byte]*assistantCachedResult{}, ids: map[string]assistantCallIdentity{}, unresolved: map[[32]byte]assistantUnresolvedCall{}, queryIDs: map[string]bool{}, referenceIDs: map[string]bool{}}
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
	contextRetried := false
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
		if preflightErr := assistantToolRequestError(current, turnInput); preflightErr != nil {
			if assistantErrorCode(preflightErr) != data.ToolErrorBudgetExhausted {
				return run.result, run.stop(assistantErrorCode(preflightErr), assistantReason(assistantErrorCode(preflightErr)))
			}
			turnInput = compactAssistantWorkingContext(turnInput)
			if !assistantToolRequestFits(current, turnInput) {
				turnInput, _ = run.recoverWorkingContext(turnInput)
			}
			if preflightErr = assistantToolRequestError(current, turnInput); preflightErr != nil {
				if assistantErrorCode(preflightErr) == data.ToolErrorBudgetExhausted {
					run.observation.ContextFailureSource = "local_input"
				}
				return run.result, run.stop(assistantErrorCode(preflightErr), assistantReason(assistantErrorCode(preflightErr)))
			}
			messages, state = turnInput.Messages, turnInput.State
		}
		if err = run.emit(AssistantEvent{Type: AssistantEventStatus, Status: AssistantStatusQuerying}); err != nil {
			return run.result, run.writeFailure()
		}
		var display data.DisplayEventSink
		toolTurn := false
		if sink != nil {
			display = func(event data.DisplayEvent) error {
				if event.ReplyReset {
					toolTurn = true
					if err := run.resetReply(); err != nil {
						return err
					}
				}
				if event.ThinkingDelta != "" {
					if err := run.emit(AssistantEvent{Type: AssistantEventThinkingDelta, Delta: event.ThinkingDelta}); err != nil {
						return err
					}
					run.result.Thinking += event.ThinkingDelta
				}
				if !toolTurn && len(run.unresolved) == 0 && event.ReplyDelta != "" {
					if run.result.Reply == "" {
						if err := run.emit(AssistantEvent{Type: AssistantEventStatus, Status: AssistantStatusAnswering}); err != nil {
							return err
						}
					}
					if err := run.emit(AssistantEvent{Type: AssistantEventReplyDelta, Delta: event.ReplyDelta}); err != nil {
						return err
					}
					run.result.Reply += event.ReplyDelta
				}
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
					if e := run.finishReply(turn.Content); e != nil {
						if run.writeErr == nil {
							return run.result, run.stop(data.ToolErrorProtocol, AssistantReasonProtocolError)
						}
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
			var providerError *data.ToolError
			if errors.As(turnErr, &providerError) && providerError.UpstreamContext && !contextRetried && run.result.Reply == "" && input.MaxOutputTokens > 1 {
				// Providers may count reserved output against their context window.
				// Retry once with less output reservation; never discard ledger facts.
				contextRetried = true
				run.observation.ContextRecoveryAttempts++
				run.observation.ContextFailureSource = "upstream_input"
				input.MaxOutputTokens = max(1, input.MaxOutputTokens/2)
				working := compactAssistantWorkingContext(turnInput)
				if recovered, changed := run.recoverWorkingContext(working); changed {
					working = recovered
				}
				messages, state = working.Messages, nil
				continue
			}
			if providerError != nil && providerError.UpstreamContext {
				run.observation.ContextFailureSource = "upstream_input"
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
			// empty ledger or a verified answer. A successful retry of that query must
			// resolve the error before this request can complete.
			if len(run.unresolved) > 0 {
				for _, failure := range run.unresolved {
					return run.result, run.stop(failure.failure.Code, assistantReason(failure.failure.Code))
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
			if err = run.finishReply(turn.Content); err != nil {
				if run.writeErr != nil {
					return run.result, run.writeFailure()
				}
				return run.result, run.stop(data.ToolErrorProtocol, AssistantReasonProtocolError)
			}
			run.result.State = AssistantStateCompleted
			run.observation.Reason = AssistantReasonCompleted
			run.observation.ContextFailureSource = ""
			return run.result, nil
		}
		if err = run.resetReply(); err != nil {
			return run.result, run.writeFailure()
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
		turnInput.Messages, turnInput.State = messages, turn.State
		run.toolRound++
		messages, err = run.executeTurnTools(current, switched, turn, turnInput, allowed, keys)
		if err != nil {
			return run.result, err
		}
		state = turn.State
		if run.workingRecoveries > 0 {
			state = nil
		}
		working := compactAssistantWorkingContext(data.ToolTurnInput{Messages: messages, State: state})
		messages, state = working.Messages, working.State
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
	return assistantToolRequestError(p, input) == nil
}

func assistantToolRequestError(p provider.LargeLanguageModelToolProvider, input data.ToolTurnInput) error {
	if counter, ok := p.(interface {
		CheckToolTurnInput(data.ToolTurnInput) error
	}); ok {
		return counter.CheckToolTurnInput(input)
	}
	body, err := json.Marshal(input)
	if err != nil {
		return assistantRunnerError(data.ToolErrorProtocol)
	}
	if !data.AssistantRequestFits(body, input.MaxInputTokens, input.MaxRequestBytes) {
		return assistantRunnerError(data.ToolErrorBudgetExhausted)
	}
	return nil
}

// resetReply retracts provisional tool preambles before any query executes.
func (run *assistantRun) resetReply() error {
	if run.result.Reply == "" {
		return nil
	}
	if err := run.emit(AssistantEvent{Type: AssistantEventReplyReset}); err != nil {
		return err
	}
	run.result.Reply = ""
	return nil
}

func (run *assistantRun) finishReply(content string) error {
	if !strings.HasPrefix(content, run.result.Reply) {
		return assistantRunnerError(data.ToolErrorProtocol)
	}
	tail := content[len(run.result.Reply):]
	if tail != "" {
		if err := run.emit(AssistantEvent{Type: AssistantEventReplyDelta, Delta: tail}); err != nil {
			return err
		}
	}
	run.result.Reply = content
	return nil
}
