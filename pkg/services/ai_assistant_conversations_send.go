package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

func (s *AssistantConversationService) Send(ctx core.Context, uid int64, r models.AssistantSendRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "messages.send", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		var empty models.AssistantCommandResult
		limits := s.limits()
		if s.options.GenerationAvailable == nil || !s.options.GenerationAvailable(ctx, uid) {
			return empty, models.AssistantReasonProviderUnavailable
		}
		if err := r.ValidateQueryAction(); err != nil {
			return empty, models.AssistantReasonConversationConflict
		}
		if r.Mode != "chat" && r.Mode != "summary" && r.Mode != "next_page" {
			return empty, models.AssistantReasonConversationConflict
		}
		if err := models.ValidateAssistantInput(r.Content, limits.InputCharacters, limits.InputBytes); err != nil {
			return empty, models.AssistantReasonInputTooLong
		}
		if len(r.Locale) > 32 || len(r.Timezone) > 128 {
			return empty, models.AssistantReasonConversationConflict
		}
		if r.Timezone != "" {
			if _, err := time.LoadLocation(r.Timezone); err != nil {
				return empty, models.AssistantReasonConversationConflict
			}
		}
		var c *models.AssistantConversation
		var err error
		if r.ConversationId == "" {
			if r.ExpectedContextVersion != 0 || r.Mode == "next_page" {
				return empty, models.AssistantReasonConversationConflict
			}
			title := r.Content
			if r.Mode == "summary" {
				title = "Financial summary"
				if strings.HasPrefix(r.Locale, "zh") {
					title = "财务摘要"
				}
			}
			c, err = s.newConversation(tx, title)
			if err != nil {
				return empty, err
			}
		} else {
			id, err := assistantParseID(r.ConversationId)
			if err != nil {
				return empty, err
			}
			c, err = assistantConversationLoad(tx.Session, uid, id)
			if err != nil {
				return empty, err
			}
			if c.ContextVersion != r.ExpectedContextVersion {
				return empty, models.AssistantReasonConversationConflict
			}
		}
		if c.Archived {
			return empty, models.AssistantReasonConversationConflict
		}
		if c.ActiveRunId != 0 {
			return empty, models.AssistantReasonAlreadyRunning
		}
		if c.SourceInvalidated && c.PendingContinuation {
			return empty, models.AssistantReasonContextChanged
		}
		if c.TurnCount >= limits.ConversationTurns {
			return empty, models.AssistantReasonQuotaExceeded
		}
		if err := s.validateSeed(tx, c); err != nil {
			return empty, err
		}
		previous, err := s.previousAnswer(tx, c)
		if err != nil {
			return empty, err
		}
		qid := s.options.GenerateID(uuid.UUID_TYPE_ASSISTANT_MESSAGE)
		question := &models.AssistantMessage{Id: qid, Uid: uid, ConversationId: c.Id, TurnSequence: c.NextTurnSequence, Role: "user", UserMessageId: qid, Content: r.Content, CreatedAt: tx.Now.UnixMilli(), PreviousAssistantMessageId: previous, ClientRequestId: r.ClientRequestId, Mode: r.Mode, RestartQuery: r.RestartQuery, Locale: r.Locale, Timezone: r.Timezone, SchemaVersion: models.AssistantSchemaVersion}
		var source *models.AssistantMessage
		if r.Mode == "next_page" {
			sid, err := assistantParseID(r.SourceMessageId)
			if err != nil {
				return empty, err
			}
			source, err = assistantConversationMessage(tx.Session, uid, sid)
			if err != nil {
				return empty, err
			}
			if source.Role != "assistant" || source.RunState != models.AssistantRunCompleted {
				return empty, models.AssistantReasonConversationConflict
			}
			authorized := source.ConversationId == c.Id && source.TurnSequence == c.NextTurnSequence-1 && source.Id == previous
			if !authorized && c.PendingContinuation && c.TurnCount == 0 && c.BranchKind == "continue" {
				var prefix []AssistantFrozenTurn
				if json.Unmarshal([]byte(c.FrozenPrefixJSON), &prefix) != nil {
					return empty, models.AssistantReasonStorageUnavailable
				}
				// Only the final frozen answer is the branch's current query result.
				if len(prefix) > 0 {
					authorized = prefix[len(prefix)-1].AssistantMessageId == r.SourceMessageId && prefix[len(prefix)-1].ConversationId == assistantConversationID(source.ConversationId)
				}
			}
			if !authorized {
				return empty, models.AssistantReasonConversationConflict
			}
			if err := assistantValidateQuerySource(source, r.RestartQuery, tx.Now.UnixMilli()); err != nil {
				return empty, err
			}
			question.SourceMessageId = source.Id
		} else if r.SourceMessageId != "" {
			return empty, models.AssistantReasonConversationConflict
		}
		question.UsedBytes, err = assistantQuestionStorageBytes(question)
		if err != nil {
			return empty, err
		}
		run, err := s.newRun(tx, c, question, 1, r.ClientRequestId)
		if err != nil {
			return empty, err
		}
		if source != nil {
			run.SourceMessageId = source.Id
			run.Continuation = source.Continuation
			run.ContinuationExpiresAt = source.ContinuationExpiresAt
			run.ContinuationQueryId = source.ContinuationQueryId
			run.ContinuationFilterJSON = source.ContinuationFilterJSON
		}
		c.TurnCount++
		c.NextTurnSequence++
		c.ContextVersion++
		c.PendingContinuation = false
		c.PendingQuerySourceMessageId = 0
		run.ExpectedContextVersion = c.ContextVersion
		if _, err = tx.Session.Insert(question, run); err != nil {
			return empty, err
		}
		if err = assistantConversationSave(tx, c); err != nil {
			return empty, err
		}
		return assistantAccepted(c, question, run), nil
	})
}

func (s *AssistantConversationService) Retry(ctx core.Context, uid int64, r models.AssistantRetryRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "messages.retry", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		var empty models.AssistantCommandResult
		if s.options.GenerationAvailable == nil || !s.options.GenerationAvailable(ctx, uid) {
			return empty, models.AssistantReasonProviderUnavailable
		}
		if r.Intent != "retry" && r.Intent != "regenerate" {
			return empty, models.AssistantReasonConversationConflict
		}
		id, err := assistantParseID(r.ConversationId)
		if err != nil {
			return empty, err
		}
		c, err := assistantConversationLoad(tx.Session, uid, id)
		if err != nil {
			return empty, err
		}
		if c.ContextVersion != r.ExpectedContextVersion || c.Archived {
			return empty, models.AssistantReasonConversationConflict
		}
		if c.ActiveRunId != 0 {
			return empty, models.AssistantReasonAlreadyRunning
		}
		if err = s.validateSeed(tx, c); err != nil {
			return empty, err
		}
		qid, err := assistantParseID(r.UserMessageId)
		if err != nil {
			return empty, err
		}
		q, err := assistantConversationMessage(tx.Session, uid, qid)
		if err != nil {
			return empty, err
		}
		if q.ConversationId != c.Id || q.Role != "user" || q.TurnSequence != c.NextTurnSequence-1 {
			return empty, models.AssistantReasonConversationConflict
		}
		var last models.AssistantMessage
		found, err := tx.Session.Where("uid = ? AND user_message_id = ? AND role = ?", uid, q.Id, "assistant").Desc("attempt_number").Get(&last)
		if err != nil {
			return empty, err
		}
		if !found || !last.RunState.IsTerminal() {
			return empty, models.AssistantReasonAlreadyRunning
		}
		if r.Intent == "retry" && last.RunState == models.AssistantRunCompleted {
			return empty, models.AssistantReasonConversationConflict
		}
		run, err := s.newRun(tx, c, q, last.AttemptNumber+1, r.ClientRequestId)
		if err != nil {
			return empty, err
		}
		if q.Mode == "next_page" {
			source, err := assistantConversationMessage(tx.Session, uid, q.SourceMessageId)
			if err != nil {
				return empty, err
			}
			if err := assistantValidateQuerySource(source, q.RestartQuery, tx.Now.UnixMilli()); err != nil {
				return empty, err
			}
			run.SourceMessageId = source.Id
			run.Continuation = source.Continuation
			run.ContinuationExpiresAt = source.ContinuationExpiresAt
			run.ContinuationQueryId = source.ContinuationQueryId
			run.ContinuationFilterJSON = source.ContinuationFilterJSON
		}
		c.ContextVersion++
		run.ExpectedContextVersion = c.ContextVersion
		if _, err = tx.Session.Insert(run); err != nil {
			return empty, err
		}
		if err = assistantConversationSave(tx, c); err != nil {
			return empty, err
		}
		return assistantAccepted(c, q, run), nil
	})
}
func (s *AssistantConversationService) newConversation(tx *AssistantStorageTx, title string) (*models.AssistantConversation, error) {
	runes := []rune(strings.TrimSpace(title))
	max := s.limits().TitleCharacters
	if len(runes) > max {
		runes = runes[:max]
	}
	c := &models.AssistantConversation{Id: s.options.GenerateID(uuid.UUID_TYPE_ASSISTANT_CONVERSATION), Uid: tx.State.Uid, Title: string(runes), CreatedAt: tx.Now.UnixMilli(), ActivityAt: tx.Now.UnixMilli(), MetadataVersion: 1, ContextVersion: 1, SourceVersion: 1, NextTurnSequence: 1, MemoryMode: "inherit", SchemaVersion: models.AssistantSchemaVersion}
	limits := s.limits()
	if err := tx.ReserveConversation(c, &limits); err != nil {
		return nil, err
	}
	if err := assistantSettleConversationHeader(tx, c); err != nil {
		return nil, err
	}
	if _, err := tx.Session.Insert(c); err != nil {
		return nil, err
	}
	return c, nil
}
func (s *AssistantConversationService) newRun(tx *AssistantStorageTx, c *models.AssistantConversation, q *models.AssistantMessage, attempt int, requestID string) (*models.AssistantMessage, error) {
	run := &models.AssistantMessage{Id: s.options.GenerateID(uuid.UUID_TYPE_ASSISTANT_MESSAGE), Uid: q.Uid, ConversationId: c.Id, TurnSequence: q.TurnSequence, Role: "assistant", UserMessageId: q.Id, AttemptNumber: attempt, CreatedAt: tx.Now.UnixMilli(), RunState: models.AssistantRunQueued, SnapshotRevision: 1, ClientRequestId: requestID, DeletionEpoch: tx.State.DeletionEpoch, Mode: q.Mode, RestartQuery: q.RestartQuery, Locale: q.Locale, Timezone: q.Timezone, SchemaVersion: models.AssistantSchemaVersion}
	runtimeJSON, err := json.Marshal(NewAssistantRunPolicy(s.options.Runtime(), s.options.Config()))
	if err != nil {
		return nil, err
	}
	run.RuntimeJSON = string(runtimeJSON)
	limits := s.limits()
	if err := tx.ReserveRun(c, run, &limits); err != nil {
		return nil, err
	}
	c.ActiveRunId = run.Id
	c.ActivityAt = tx.Now.UnixMilli()
	return run, nil
}
func (s *AssistantConversationService) previousAnswer(tx *AssistantStorageTx, c *models.AssistantConversation) (int64, error) {
	var last models.AssistantMessage
	found, err := tx.Session.Where("uid = ? AND conversation_id = ? AND role = ?", tx.State.Uid, c.Id, "user").Desc("turn_sequence").Get(&last)
	if err != nil || !found {
		return 0, err
	}
	if last.SelectedAssistantMessageId == 0 {
		return last.PreviousAssistantMessageId, nil
	}
	answer, err := assistantConversationMessage(tx.Session, tx.State.Uid, last.SelectedAssistantMessageId)
	if err != nil {
		return 0, err
	}
	if answer.RunState != models.AssistantRunCompleted {
		return 0, models.AssistantReasonContextChanged
	}
	return answer.Id, nil
}
func (s *AssistantConversationService) validateSeed(tx *AssistantStorageTx, c *models.AssistantConversation) error {
	if c.SourceInvalidated && c.PendingContinuation {
		return models.AssistantReasonContextChanged
	}
	var sources []models.AssistantSourceDependency
	if c.SeedSourcesJSON != "" && json.Unmarshal([]byte(c.SeedSourcesJSON), &sources) != nil {
		return models.AssistantReasonStorageUnavailable
	}
	return tx.ValidateDependencies(c, &models.AssistantInputDependencies{DeletionEpoch: tx.State.DeletionEpoch, MemoryUseEpoch: tx.State.MemoryUseEpoch, MemoryMode: c.MemoryMode, Sources: sources})
}
func assistantAccepted(c *models.AssistantConversation, q, run *models.AssistantMessage) models.AssistantCommandResult {
	return models.AssistantCommandResult{ConversationId: assistantConversationID(c.Id), UserMessageId: assistantConversationID(q.Id), AssistantMessageId: assistantConversationID(run.Id), RunId: assistantConversationID(run.Id), ContextVersion: c.ContextVersion, MetadataVersion: c.MetadataVersion, State: run.RunState}
}
func assistantSettleConversationHeader(tx *AssistantStorageTx, c *models.AssistantConversation) error {
	// Accounting counters are system-maintenance state, not recursively charged
	// application payload. Every saved header uses this one fixed projection.
	n, err := AssistantStorageBytes(struct {
		ID, Title, MemoryMode, BranchKind                                                                                                                                                   string
		Archived, TitleManual, PendingContinuation, SourceInvalidated                                                                                                                       bool
		CreatedAt, ActivityAt, MetadataVersion, ContextVersion, SourceVersion, NextTurnSequence, SourceConversationID, SourceMessageID, PendingQuerySourceID, ContextInvalidatedThroughTurn int64
		Summary, SummarySources, Seed, SeedSources, Frozen                                                                                                                                  string
		SummaryThrough, SummarySourceVersion                                                                                                                                                int64
	}{assistantConversationID(c.Id), c.Title, c.MemoryMode, c.BranchKind, c.Archived, c.TitleManual, c.PendingContinuation, c.SourceInvalidated,
		c.CreatedAt, c.ActivityAt, c.MetadataVersion, c.ContextVersion, c.SourceVersion, c.NextTurnSequence, c.SourceConversationId, c.SourceMessageId, c.PendingQuerySourceMessageId, c.ContextInvalidatedThroughTurn,
		c.Summary, c.SummarySourcesJSON, c.SeedContext, c.SeedSourcesJSON, c.FrozenPrefixJSON, c.SummaryThrough, c.SummarySourceVersion})
	if err != nil {
		return err
	}
	return tx.SettleHeader(c, n)
}

// AssistantAttemptStorageBytes includes the original question exactly once, even
// when its answer is retried; failed and partial attempts still consume capacity.
func AssistantAttemptStorageBytes(question, run *models.AssistantMessage) (int64, error) {
	snapshot, err := AssistantMessageSnapshot(run)
	if err != nil {
		return 0, err
	}
	n, err := AssistantStorageBytes(struct {
		Snapshot                                                                      models.AssistantRunSnapshot
		Dependencies, Limits, Runtime, Mode, Locale, Timezone                         string
		Continuation, ContinuationFilter, ContinuationQueryID                         string
		SourceMessageID, ExpectedContextVersion, DeletionEpoch, ContinuationExpiresAt int64
		RestartQuery                                                                  bool
	}{snapshot, run.InputDependenciesJSON, run.LimitsJSON, run.RuntimeJSON, run.Mode, run.Locale, run.Timezone,
		run.Continuation, run.ContinuationFilterJSON, run.ContinuationQueryId, run.SourceMessageId, run.ExpectedContextVersion, run.DeletionEpoch, run.ContinuationExpiresAt, run.RestartQuery})
	if run.AttemptNumber == 1 {
		n += question.UsedBytes
	}
	return n, err
}

func assistantValidateQuerySource(source *models.AssistantMessage, restart bool, now int64) error {
	if restart {
		if source.ContinuationFilterJSON == "" || !json.Valid([]byte(source.ContinuationFilterJSON)) {
			return models.AssistantReasonCursorExpired
		}
		return nil
	}
	if source.Continuation == "" || source.ContinuationExpiresAt <= now {
		return models.AssistantReasonCursorExpired
	}
	return nil
}

// Fixed-width selected-version identity reserves the same business metadata
// bytes before and after completion, without charging a question again on retry.
func assistantQuestionStorageBytes(q *models.AssistantMessage) (int64, error) {
	return AssistantStorageBytes(struct {
		ID, ConversationID, SelectedAnswerID, PreviousAnswerID, SourceMessageID string
		Content, Mode, Locale, Timezone                                         string
		CreatedAt, TurnSequence                                                 int64
		RestartQuery                                                            bool
	}{fmt.Sprintf("%020d", q.Id), fmt.Sprintf("%020d", q.ConversationId), fmt.Sprintf("%020d", q.SelectedAssistantMessageId), fmt.Sprintf("%020d", q.PreviousAssistantMessageId), fmt.Sprintf("%020d", q.SourceMessageId), q.Content, q.Mode, q.Locale, q.Timezone, q.CreatedAt, q.TurnSequence, q.RestartQuery})
}
