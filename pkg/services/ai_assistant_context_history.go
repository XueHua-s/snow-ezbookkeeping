package services

import (
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"strings"
	"xorm.io/xorm"
)

type assistantContextTurn struct{ Question, Answer *models.AssistantMessage }
type assistantContextSnapshot struct {
	Conversation *models.AssistantConversation
	Question     *models.AssistantMessage
	State        *models.AssistantUserState
	Turns        []assistantContextTurn
	Frozen       []assistantContextTurn
	Unresolved   []*models.AssistantMessage
	Dependencies models.AssistantInputDependencies
}

// Follow fixed predecessor pointers, never today's selected version of an old
// question. Failed attempts cannot enter the chain, even when their text exists.
func assistantContextChain(session *xorm.Session, uid, conversationID, predecessor int64, maxTurns int) ([]assistantContextTurn, error) {
	var reverse []assistantContextTurn
	seen := map[int64]bool{}
	for predecessor != 0 {
		if seen[predecessor] || len(reverse) >= maxTurns {
			return nil, models.AssistantReasonContextLimit
		}
		seen[predecessor] = true
		a, e := assistantConversationMessage(session, uid, predecessor)
		if e != nil {
			return nil, e
		}
		if a.ConversationId != conversationID || a.Role != "assistant" || a.RunState != models.AssistantRunCompleted || strings.TrimSpace(a.Content) == "" {
			return nil, models.AssistantReasonContextChanged
		}
		q, e := assistantConversationMessage(session, uid, a.UserMessageId)
		if e != nil {
			return nil, e
		}
		if q.ConversationId != conversationID || q.Role != "user" || q.TurnSequence != a.TurnSequence {
			return nil, models.AssistantReasonContextChanged
		}
		reverse = append(reverse, assistantContextTurn{q, a})
		predecessor = q.PreviousAssistantMessageId
	}
	for i, j := 0, len(reverse)-1; i < j; i, j = i+1, j-1 {
		reverse[i], reverse[j] = reverse[j], reverse[i]
	}
	return reverse, nil
}
func (s *AssistantContextService) snapshot(ctx core.Context, run *models.AssistantMessage) (*assistantContextSnapshot, error) {
	policy, e := s.acceptedPolicy(run)
	if e != nil {
		return nil, e
	}
	result := &assistantContextSnapshot{}
	err := s.storage.WithUserTransaction(ctx, run.Uid, func(tx *AssistantStorageTx) error {
		*result = assistantContextSnapshot{}
		c, e := assistantConversationLoad(tx.Session, run.Uid, run.ConversationId)
		if e != nil {
			return e
		}
		for _, raw := range []string{c.Summary, c.SeedContext} {
			if raw != "" {
				var summary models.AssistantSummaryDTO
				if json.Unmarshal([]byte(raw), &summary) != nil || summary.SchemaVersion != models.AssistantSchemaVersion || summary.Topic == "" {
					return models.AssistantReasonContextChanged
				}
			}
		}
		q, e := assistantConversationMessage(tx.Session, run.Uid, run.UserMessageId)
		if e != nil {
			return e
		}
		if q.ConversationId != c.Id || q.Role != "user" || run.DeletionEpoch != tx.State.DeletionEpoch {
			return models.AssistantReasonContextChanged
		}
		turns, e := assistantContextChain(tx.Session, run.Uid, c.Id, q.PreviousAssistantMessageId, max(c.TurnCount, policy.Experience.Limits.ConversationTurns))
		if e != nil {
			return e
		}
		if c.ContextInvalidatedThroughTurn > 0 {
			valid := turns[:0]
			for _, t := range turns {
				if t.Question.TurnSequence > c.ContextInvalidatedThroughTurn {
					valid = append(valid, t)
				}
			}
			turns = valid
		}
		// Failed answers cannot become facts, but their user questions still carry
		// intent (especially when the next message says only “continue”). Retain
		// exact original questions independently of the completed-answer chain.
		if e = tx.Session.Where("uid = ? AND conversation_id = ? AND role = ? AND selected_assistant_message_id = 0 AND turn_sequence < ? AND turn_sequence > ?", run.Uid, c.Id, "user", q.TurnSequence, c.ContextInvalidatedThroughTurn).Asc("turn_sequence").Find(&result.Unresolved); e != nil {
			return e
		}
		result.Conversation, result.Question, result.State, result.Turns = c, q, tx.State, turns
		result.Dependencies = models.AssistantInputDependencies{DeletionEpoch: tx.State.DeletionEpoch, MemoryUseEpoch: tx.State.MemoryUseEpoch, MemoryMode: c.MemoryMode}
		for _, raw := range []string{c.SeedSourcesJSON, c.SummarySourcesJSON} {
			if raw != "" {
				var deps []models.AssistantSourceDependency
				if json.Unmarshal([]byte(raw), &deps) != nil {
					return models.AssistantReasonContextChanged
				}
				result.Dependencies.Sources = append(result.Dependencies.Sources, deps...)
			}
		}
		if c.SourceInvalidated && c.PendingContinuation {
			return models.AssistantReasonContextChanged
		}
		if c.FrozenPrefixJSON != "" {
			var frozen []AssistantFrozenTurn
			if json.Unmarshal([]byte(c.FrozenPrefixJSON), &frozen) != nil || len(c.FrozenPrefixJSON) > policy.Experience.Limits.ConversationHeaderBytes {
				return models.AssistantReasonContextLimit
			}
			for _, pair := range frozen {
				qid, e := assistantParseID(pair.UserMessageId)
				if e != nil {
					return e
				}
				aid, e := assistantParseID(pair.AssistantMessageId)
				if e != nil {
					return e
				}
				q, e := assistantConversationMessage(tx.Session, run.Uid, qid)
				if e != nil {
					return e
				}
				a, e := assistantConversationMessage(tx.Session, run.Uid, aid)
				if e != nil {
					return e
				}
				if assistantConversationID(q.ConversationId) != pair.ConversationId || a.ConversationId != q.ConversationId || a.UserMessageId != q.Id || q.Role != "user" || a.Role != "assistant" || a.RunState != models.AssistantRunCompleted || strings.TrimSpace(a.Content) == "" {
					return models.AssistantReasonContextChanged
				}
				result.Frozen = append(result.Frozen, assistantContextTurn{q, a})
			}
		}
		return tx.ValidateDependencies(c, &result.Dependencies, policy.Experience.Limits)
	})
	return result, err
}
func assistantContextMessages(turns []assistantContextTurn) []data.ConversationMessage {
	result := make([]data.ConversationMessage, 0, 2*len(turns))
	for _, t := range turns {
		status := "Historical user message"
		if t.Answer == nil {
			status = "Earlier user question whose answer did not complete; no failed answer is evidence. Later corrections override this. User message"
		}
		result = append(result, data.ConversationMessage{Role: data.ConversationRoleUser, Content: status + " sent at " + assistantConversationTime(t.Question.CreatedAt) + " (timezone " + t.Question.Timezone + "); resolve its relative dates at that time:\n" + t.Question.Content})
		if t.Answer != nil {
			result = append(result, data.ConversationMessage{Role: data.ConversationRoleAssistant, Content: t.Answer.Content})
		}
	}
	return result
}

// Merge unanswered intent chronologically without inventing an assistant answer.
// These question-only turns can be summarized by the same bounded history path;
// SummaryThrough then prevents old failures from overriding newer corrections.
func assistantContextTurns(turns []assistantContextTurn, unresolved []*models.AssistantMessage) []assistantContextTurn {
	result := make([]assistantContextTurn, 0, len(turns)+len(unresolved))
	next := 0
	for _, turn := range turns {
		for next < len(unresolved) && unresolved[next].TurnSequence < turn.Question.TurnSequence {
			result = append(result, assistantContextTurn{Question: unresolved[next]})
			next++
		}
		result = append(result, turn)
	}
	for ; next < len(unresolved); next++ {
		result = append(result, assistantContextTurn{Question: unresolved[next]})
	}
	return result
}
