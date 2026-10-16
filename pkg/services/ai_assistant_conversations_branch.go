package services

import (
	"encoding/json"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func (s *AssistantConversationService) Branch(ctx core.Context, uid int64, r models.AssistantBranchRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "conversations.branch", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		var empty models.AssistantCommandResult
		if r.Kind != "continue" && r.Kind != "edit" && r.Kind != "regenerate" {
			return empty, models.AssistantReasonConversationConflict
		}
		id, err := assistantParseID(r.SourceConversationId)
		if err != nil {
			return empty, err
		}
		parent, err := assistantConversationLoad(tx.Session, uid, id)
		if err != nil {
			return empty, err
		}
		if parent.ContextVersion != r.ExpectedContextVersion {
			return empty, models.AssistantReasonConversationConflict
		}
		if parent.ActiveRunId != 0 {
			return empty, models.AssistantReasonAlreadyRunning
		}
		if err = s.validateSeed(tx, parent); err != nil {
			return empty, err
		}
		sid, err := assistantParseID(r.SourceMessageId)
		if err != nil {
			return empty, err
		}
		target, err := assistantConversationMessage(tx.Session, uid, sid)
		if err != nil {
			return empty, err
		}
		if target.ConversationId != parent.Id {
			return empty, models.AssistantReasonNotAccessible
		}
		querySourceID := int64(0)
		explicitAnswerID := int64(0)
		if target.Role == "assistant" {
			if target.RunState != models.AssistantRunCompleted {
				return empty, models.AssistantReasonConversationConflict
			}
			explicitAnswerID = target.Id
		}
		if r.Kind == "continue" && target.Role == "assistant" && target.ContinuationFilterJSON != "" && target.TurnSequence > parent.ContextInvalidatedThroughTurn {
			querySourceID = target.Id
		}
		if target.Role == "assistant" {
			target, err = assistantConversationMessage(tx.Session, uid, target.UserMessageId)
			if err != nil {
				return empty, err
			}
		}
		if target.Role != "user" {
			return empty, models.AssistantReasonConversationConflict
		}
		// An edit/regeneration excludes its target question. A continuation includes
		// the selected answer at its target, never a later answer.
		answerID := target.PreviousAssistantMessageId
		if r.Kind == "continue" {
			answerID = target.SelectedAssistantMessageId
			if explicitAnswerID != 0 {
				answerID = explicitAnswerID
			}
			if answerID == 0 {
				answerID = target.PreviousAssistantMessageId
			}
		}
		turns := []AssistantFrozenTurn{}
		sources := []models.AssistantSourceDependency{}
		if parent.FrozenPrefixJSON != "" && json.Unmarshal([]byte(parent.FrozenPrefixJSON), &turns) != nil {
			return empty, models.AssistantReasonStorageUnavailable
		}
		if parent.SeedSourcesJSON != "" && json.Unmarshal([]byte(parent.SeedSourcesJSON), &sources) != nil {
			return empty, models.AssistantReasonStorageUnavailable
		}
		chain := []AssistantFrozenTurn{}
		seen := map[int64]bool{}
		for answerID != 0 {
			if seen[answerID] || len(chain) > s.limits().ConversationTurns {
				return empty, models.AssistantReasonContextLimit
			}
			seen[answerID] = true
			answer, err := assistantConversationMessage(tx.Session, uid, answerID)
			if err != nil {
				return empty, err
			}
			if answer.ConversationId != parent.Id || answer.Role != "assistant" || answer.RunState != models.AssistantRunCompleted {
				return empty, models.AssistantReasonContextChanged
			}
			question, err := assistantConversationMessage(tx.Session, uid, answer.UserMessageId)
			if err != nil {
				return empty, err
			}
			if question.ConversationId != parent.Id || question.Role != "user" || question.TurnSequence > target.TurnSequence {
				return empty, models.AssistantReasonContextChanged
			}
			if question.TurnSequence <= parent.ContextInvalidatedThroughTurn {
				break
			}
			chain = append(chain, AssistantFrozenTurn{ConversationId: r.SourceConversationId, UserMessageId: assistantConversationID(question.Id), AssistantMessageId: assistantConversationID(answer.Id)})
			answerID = question.PreviousAssistantMessageId
		}
		for i := len(chain) - 1; i >= 0; i-- {
			turns = append(turns, chain[i])
		}
		sources = append(sources, models.AssistantSourceDependency{ConversationId: r.SourceConversationId, MessageId: r.SourceMessageId, Version: parent.SourceVersion})
		if len(sources) > 400 {
			return empty, models.AssistantReasonContextLimit
		}
		frozen, err := json.Marshal(turns)
		if err != nil {
			return empty, err
		}
		sourceJSON, err := json.Marshal(sources)
		if err != nil {
			return empty, err
		}
		if len(frozen)+len(sourceJSON)+len(parent.SeedContext) >= s.limits().ConversationHeaderBytes {
			return empty, models.AssistantReasonContextLimit
		}
		c, err := s.newConversation(tx, parent.Title)
		if err != nil {
			return empty, err
		}
		c.SourceConversationId = parent.Id
		c.SourceMessageId = sid
		c.BranchKind = r.Kind
		c.PendingContinuation = true
		c.PendingQuerySourceMessageId = querySourceID
		c.MemoryMode = parent.MemoryMode
		c.SeedContext = parent.SeedContext
		c.FrozenPrefixJSON = string(frozen)
		c.SeedSourcesJSON = string(sourceJSON)
		if err = assistantSettleConversationHeader(tx, c); err != nil {
			return empty, err
		}
		if err = assistantConversationSave(tx, c); err != nil {
			return empty, err
		}
		return models.AssistantCommandResult{ConversationId: assistantConversationID(c.Id), MetadataVersion: c.MetadataVersion, ContextVersion: c.ContextVersion}, nil
	})
}
