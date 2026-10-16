package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// Only bounded operational metadata is observable. Never log history, summaries,
// tool arguments, provider errors or arbitrary response text.
type assistantCompactionObservation struct {
	Kind                 string `json:"kind"`
	Stage                string `json:"stage"`
	FailureSource        string `json:"failureSource,omitempty"`
	Recovery             string `json:"recovery,omitempty"`
	DurationMilliseconds int64  `json:"durationMilliseconds,omitempty"`
	Attempt              uint32 `json:"attempt,omitempty"`
	ModelCalls           uint32 `json:"modelCalls"`
	HistoricalTurns      int    `json:"historicalTurns,omitempty"`
}

func logAssistantCompaction(ctx core.Context, event assistantCompactionObservation) {
	raw, _ := json.Marshal(event)
	log.Infof(ctx, "[assistant] %s", raw)
}
func (s *AssistantContextService) compactionObservation(ctx core.Context, event assistantCompactionObservation) {
	event.Kind = "compaction"
	if s.observeCompaction != nil {
		s.observeCompaction(ctx, event)
	}
}
func assistantPreparationContextError(ctx core.Context, run *models.AssistantMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if run.Deadline > 0 && !time.Now().Before(time.UnixMilli(run.Deadline)) {
		return context.DeadlineExceeded
	}
	return nil
}

// Summary timeouts are local to this speculative request. Only the caller's
// context/deadline may terminate the foreground run; a local failure can leave
// the original complete history usable without committing any partial summary.
func (s *AssistantContextService) requestCompaction(ctx core.Context, run *models.AssistantMessage, p provider.LargeLanguageModelToolProvider, input data.ToolTurnInput, timeoutSeconds uint32, attempt uint32, turns int) (response data.TurnResponse, used uint32, err error) {
	started := time.Now()
	event := assistantCompactionObservation{Stage: "request_started", Attempt: attempt, HistoricalTurns: turns}
	s.compactionObservation(ctx, event)
	defer func() {
		event.Stage = "request_finished"
		event.ModelCalls = used
		event.DurationMilliseconds = time.Since(started).Milliseconds()
		s.compactionObservation(ctx, event)
	}()
	deadline := started.Add(time.Duration(timeoutSeconds) * time.Second)
	if run.Deadline > 0 && time.UnixMilli(run.Deadline).Before(deadline) {
		deadline = time.UnixMilli(run.Deadline)
	}
	child, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	if parentErr := assistantPreparationContextError(ctx, run); parentErr != nil {
		event.FailureSource = "run_context"
		return response, used, parentErr
	}
	used = 1
	response, err = p.RunToolTurn(core.NewAssistantContext(child, ctx.GetContextId(), ctx.GetClientLocale(), ctx.ClientIP()), run.Uid, input, nil)
	if parentErr := assistantPreparationContextError(ctx, run); parentErr != nil {
		event.FailureSource = "run_context"
		return response, used, parentErr
	}
	if child.Err() != nil {
		event.FailureSource = "summary_timeout"
		return response, used, models.AssistantReasonProviderUnavailable
	}
	if err != nil {
		event.FailureSource = "provider_failure"
		if errors.Is(err, context.DeadlineExceeded) {
			event.FailureSource = "provider_timeout"
		}
		var toolError *data.ToolError
		if errors.As(err, &toolError) && toolError.Code == data.ToolErrorBudgetExhausted {
			event.FailureSource = "context_budget"
			return response, used, models.AssistantReasonContextLimit
		}
		return response, used, models.AssistantReasonProviderUnavailable
	}
	var summary models.AssistantSummaryDTO
	if !response.Completed || len(response.ToolCalls) > 0 || json.Unmarshal([]byte(response.Content), &summary) != nil || strings.TrimSpace(summary.Topic) == "" {
		event.FailureSource = "invalid_response"
		return response, used, models.AssistantReasonProviderUnavailable
	}
	return response, used, nil
}
