package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

type assistantRunBuffer struct {
	sync.Mutex
	content   string
	progress  string
	dirty     bool
	stopCause error
	summary   string
}

func (s *AssistantRunService) execute(parent context.Context, claimed *models.AssistantMessage) {
	deadline, cancel := context.WithDeadline(parent, time.UnixMilli(claimed.Deadline))
	defer cancel()
	ctx := core.NewAssistantContext(deadline, claimed.ClientRequestId, claimed.Locale, "")
	buffer := &assistantRunBuffer{}
	finished := make(chan struct{})
	monitorError := make(chan error, 1)
	var monitor sync.WaitGroup
	monitor.Add(1)
	go func() {
		defer monitor.Done()
		tick := time.NewTicker(time.Duration(s.policy(claimed).Experience.CheckpointMilliseconds) * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-finished:
				return
			case <-deadline.Done():
				return
			case <-tick.C:
				if err := s.checkpoint(ctx, claimed, buffer, nil, false); err != nil {
					monitorError <- err
					cancel()
					return
				}
			}
		}
	}()
	var result llm.AssistantRunResult
	var runErr error
	var prepared *AssistantPreparedRun
	if s.options.Prepare == nil {
		runErr = models.AssistantReasonProviderUnavailable
	} else {
		prepared, runErr = s.options.Prepare(ctx, claimed)
	}
	if runErr == nil && (prepared == nil || prepared.Runner == nil) {
		runErr = models.AssistantReasonProviderUnavailable
	}
	if runErr == nil {
		runErr = s.checkpoint(ctx, claimed, buffer, &prepared.Dependencies, true)
	}
	if runErr == nil {
		var limits models.AssistantLimits
		_ = json.Unmarshal([]byte(claimed.LimitsJSON), &limits)
		result, runErr = prepared.Runner.Run(ctx, prepared.Input, func(event llm.AssistantEvent) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			buffer.Lock()
			defer buffer.Unlock()
			if event.Type == llm.AssistantEventReplyReset {
				buffer.content = ""
				buffer.dirty = true
			}
			if event.Type == llm.AssistantEventReplyDelta {
				if len(buffer.content)+len(event.Delta) > limits.ResponseBytes {
					remaining := limits.ResponseBytes - len(buffer.content)
					chunk := event.Delta[:remaining]
					for !utf8.ValidString(chunk) {
						chunk = chunk[:len(chunk)-1]
					}
					buffer.content += chunk
					buffer.dirty = true
					buffer.stopCause = models.AssistantReasonOutputLimit
					return buffer.stopCause
				}
				buffer.content += event.Delta
				buffer.dirty = true
			}
			if event.Type == llm.AssistantEventThinkingDelta && len(buffer.summary)+len(event.Delta) <= limits.ProcessSummaryBytes {
				buffer.summary += event.Delta
				buffer.dirty = true
			}
			if event.Type == llm.AssistantEventStatus {
				buffer.progress = string(event.Status)
				buffer.dirty = true
			}
			return nil
		})
	}
	buffer.Lock()
	if buffer.stopCause != nil {
		runErr = buffer.stopCause
	}
	buffer.Unlock()
	close(finished)
	monitor.Wait()
	select {
	case err := <-monitorError:
		runErr = err
	default:
	}
	// A canceled application/provider context cannot prevent the final database
	// transaction. This fresh bounded context never resumes upstream generation.
	finalCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	c := core.NewAssistantContext(finalCtx, claimed.ClientRequestId, claimed.Locale, "")
	err := s.storage.WithUserTransaction(c, claimed.Uid, func(tx *AssistantStorageTx) error {
		run, err := assistantConversationMessage(tx.Session, claimed.Uid, claimed.Id)
		if err != nil {
			return err
		}
		if run.RunState.IsTerminal() || run.Owner != claimed.Owner || run.LeaseEpoch != claimed.LeaseEpoch || run.LeaseExpiresAt <= tx.Now.UnixMilli() {
			return nil
		}
		if run.RunState == models.AssistantRunCancelRequested || run.CancelRequested {
			return s.finish(tx, run, models.AssistantRunCanceled, run.ReasonCode)
		}
		if err = s.validate(tx, run); err != nil {
			return s.finish(tx, run, models.AssistantRunCanceled, models.AssistantReasonContextChanged)
		}
		if s.options.GenerationAvailable == nil || !s.options.GenerationAvailable(c, run.Uid) {
			return s.finish(tx, run, models.AssistantRunCanceled, models.AssistantReasonGenerationUnavailable)
		}
		if parent.Err() != nil {
			return s.finish(tx, run, models.AssistantRunInterrupted, models.AssistantReasonInterrupted)
		}
		state, reason := models.AssistantRunCompleted, models.AssistantReasonCode("")
		savedContent, savedSummary := run.Content, run.ProcessSummary
		savedQueries, savedReferences, savedContinuation := run.QueriesJSON, run.ReferencesJSON, run.Continuation
		if err := assistantRunEvidence(run, result); err != nil {
			runErr = err
		}
		buffer.Lock()
		run.Content = buffer.content
		run.ProcessSummary = buffer.summary
		buffer.Unlock()
		if runErr != nil {
			state = models.AssistantRunFailed
			reason = models.AssistantReasonProviderUnavailable
			var controlled models.AssistantReasonCode
			if errors.As(runErr, &controlled) {
				reason = controlled
			}
			var toolError *data.ToolError
			if errors.As(runErr, &toolError) {
				switch toolError.Code {
				case data.ToolErrorBudgetExhausted, data.ToolErrorResultTooLarge:
					reason = models.AssistantReasonExecutionLimit
				case data.ToolErrorQueryTimeout:
					reason = models.AssistantReasonRunTimeout
				case data.ToolErrorInvalidCursor:
					reason = models.AssistantReasonCursorExpired
				}
			}
			if result.StopReason == llm.AssistantReasonOutputBudget {
				reason = models.AssistantReasonOutputLimit
			}
			if result.StopReason == llm.AssistantReasonContextBudget {
				reason = models.AssistantReasonContextLimit
			}
			if errors.Is(runErr, context.DeadlineExceeded) || deadline.Err() == context.DeadlineExceeded {
				reason = models.AssistantReasonRunTimeout
			}
			if reason == models.AssistantReasonOutputLimit || reason == models.AssistantReasonRunTimeout || result.State == llm.AssistantStateIncomplete {
				state = models.AssistantRunIncomplete
			}
			if reason == models.AssistantReasonContextChanged {
				state = models.AssistantRunCanceled
			}
		} else if strings.TrimSpace(run.Content) == "" {
			state = models.AssistantRunFailed
			reason = models.AssistantReasonEmptyResponse
		}
		question, err := assistantConversationMessage(tx.Session, run.Uid, run.UserMessageId)
		if err != nil {
			return err
		}
		n, err := AssistantAttemptStorageBytes(question, run)
		if err != nil {
			return err
		}
		if n+1024 > run.ReservedBytes {
			run.Content = savedContent
			run.ProcessSummary = savedSummary
			run.QueriesJSON, run.ReferencesJSON, run.Continuation = savedQueries, savedReferences, savedContinuation
			run.ContinuationQueryId = ""
			state = models.AssistantRunIncomplete
			reason = models.AssistantReasonOutputLimit
		}
		return s.finish(tx, run, state, reason)
	})
	if err != nil {
		s.options.Observe(c, AssistantRunObservation{RunId: claimed.Id, Reason: models.AssistantReasonStorageUnavailable})
	} else {
		snapshot, e := s.GetRunSnapshot(c, claimed.Uid, assistantConversationID(claimed.Id))
		if e == nil {
			s.options.Observe(c, AssistantRunObservation{RunId: claimed.Id, State: snapshot.State, Reason: func() models.AssistantReasonCode {
				if snapshot.ReasonCode != nil {
					return *snapshot.ReasonCode
				}
				return ""
			}(), Revision: snapshot.Revision, ElapsedMilliseconds: s.storage.now().UnixMilli() - claimed.CreatedAt})
		}
	}
}
func (s *AssistantRunService) validate(tx *AssistantStorageTx, run *models.AssistantMessage) error {
	if run.DeletionEpoch != tx.State.DeletionEpoch {
		return models.AssistantReasonContextChanged
	}
	conversation, err := assistantConversationLoad(tx.Session, run.Uid, run.ConversationId)
	if err != nil {
		return err
	}
	if run.InputDependenciesJSON == "" {
		return nil
	}
	var deps models.AssistantInputDependencies
	if json.Unmarshal([]byte(run.InputDependenciesJSON), &deps) != nil {
		return models.AssistantReasonContextChanged
	}
	var limits models.AssistantLimits
	if json.Unmarshal([]byte(run.LimitsJSON), &limits) != nil {
		return models.AssistantReasonStorageUnavailable
	}
	return tx.ValidateDependencies(conversation, &deps, limits)
}
func (s *AssistantRunService) checkpoint(ctx core.Context, claimed *models.AssistantMessage, b *assistantRunBuffer, deps *models.AssistantInputDependencies, running bool) error {
	b.Lock()
	defer b.Unlock()
	if !b.dirty && deps == nil && !running {
		needsHeartbeat := false
		err := s.storage.WithUserRead(ctx, claimed.Uid, func(session *xorm.Session) error {
			run, err := assistantConversationMessage(session, claimed.Uid, claimed.Id)
			if err != nil {
				return err
			}
			now := s.storage.now()
			if run.RunState.IsTerminal() || run.Owner != claimed.Owner || run.LeaseEpoch != claimed.LeaseEpoch || run.LeaseExpiresAt <= now.UnixMilli() {
				return models.AssistantReasonInterrupted
			}
			if run.CancelRequested || run.RunState == models.AssistantRunCancelRequested {
				return models.AssistantReasonContextChanged
			}
			if s.options.GenerationAvailable == nil || !s.options.GenerationAvailable(ctx, run.Uid) {
				return models.AssistantReasonProviderUnavailable
			}
			state := &models.AssistantUserState{}
			found, err := session.Where("uid = ?", claimed.Uid).Get(state)
			if err != nil {
				return err
			}
			if !found {
				return models.AssistantReasonContextChanged
			}
			if err = s.validate(&AssistantStorageTx{Session: session, State: state, Now: now}, run); err != nil {
				return err
			}
			policy := s.policy(run)
			needsHeartbeat = run.LeaseExpiresAt-now.UnixMilli() <= int64(policy.Experience.LeaseSeconds-policy.Experience.LeaseHeartbeatSeconds)*1000
			return nil
		})
		if err != nil || !needsHeartbeat {
			return err
		}
	}
	err := s.storage.WithUserTransaction(ctx, claimed.Uid, func(tx *AssistantStorageTx) error {
		run, err := assistantConversationMessage(tx.Session, claimed.Uid, claimed.Id)
		if err != nil {
			return err
		}
		if run.RunState.IsTerminal() || run.Owner != claimed.Owner || run.LeaseEpoch != claimed.LeaseEpoch || run.LeaseExpiresAt <= tx.Now.UnixMilli() {
			return models.AssistantReasonInterrupted
		}
		if run.CancelRequested || run.RunState == models.AssistantRunCancelRequested {
			return models.AssistantReasonContextChanged
		}
		if s.options.GenerationAvailable == nil || !s.options.GenerationAvailable(ctx, run.Uid) {
			return models.AssistantReasonProviderUnavailable
		}
		if deps != nil {
			raw, err := json.Marshal(deps)
			if err != nil {
				return err
			}
			run.InputDependenciesJSON = string(raw)
		}
		if err = s.validate(tx, run); err != nil {
			return err
		}
		if running {
			run.RunState = models.AssistantRunRunning
		}
		if b.dirty {
			run.Content = b.content
			run.ProcessSummary = b.summary
			run.Progress = b.progress
		}
		policy := s.policy(run)
		heartbeatDue := run.LeaseExpiresAt-tx.Now.UnixMilli() <= int64(policy.Experience.LeaseSeconds-policy.Experience.LeaseHeartbeatSeconds)*1000
		if !b.dirty && !running && deps == nil && !heartbeatDue {
			return nil
		}
		if heartbeatDue {
			run.LeaseExpiresAt = tx.Now.Add(time.Duration(policy.Experience.LeaseSeconds) * time.Second).UnixMilli()
		}
		if b.dirty || running {
			run.SnapshotRevision++
			run.SnapshotSavedAt = tx.Now.UnixMilli()
		}
		question, err := assistantConversationMessage(tx.Session, run.Uid, run.UserMessageId)
		if err != nil {
			return err
		}
		n, err := AssistantAttemptStorageBytes(question, run)
		if err != nil {
			return err
		}
		if n+1024 > run.ReservedBytes {
			return models.AssistantReasonOutputLimit
		}
		return assistantRunSave(tx, run)
	})
	if err == nil {
		b.dirty = false
	}
	return err
}
