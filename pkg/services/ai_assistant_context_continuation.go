package services

import (
	"context"
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/xorm"
)

func (s *AssistantContextService) prepareContinuation(ctx core.Context, run *models.AssistantMessage, snap *assistantContextSnapshot, query *AssistantQueryService) (*data.ToolContinuation, *data.QueryScope, string, error) {
	var candidate data.ToolContinuation
	var saved string
	err := s.storage.WithUserRead(ctx, run.Uid, func(session *xorm.Session) error {
		source, e := assistantConversationMessage(session, run.Uid, run.SourceMessageId)
		if e != nil {
			return e
		}
		allowed := source.ConversationId == run.ConversationId
		if !allowed {
			for _, d := range snap.Dependencies.Sources {
				if d.ConversationId == assistantConversationID(source.ConversationId) {
					for _, t := range snap.Frozen {
						if t.Answer.Id == source.Id {
							allowed = true
						}
					}
					if snap.Conversation.SourceMessageId == source.Id {
						allowed = true
					}
				}
			}
		}
		if !allowed || source.Continuation == "" {
			return models.AssistantReasonNotAccessible
		}
		if json.Unmarshal([]byte(source.Continuation), &candidate) != nil {
			return models.AssistantReasonCursorExpired
		}
		saved = source.ContinuationFilterJSON
		return nil
	})
	if err != nil {
		return nil, nil, "", err
	}
	if !run.RestartQuery {
		scope, e := query.DescribeContinuation(ctx, candidate)
		if e != nil {
			return nil, nil, "", models.AssistantReasonCursorExpired
		}
		return &candidate, scope, "", nil
	}
	var args assistantQueryArgs
	if saved == "" || json.Unmarshal([]byte(saved), &args) != nil || assistantToolDomain(candidate.Tool) == nil {
		return nil, nil, "", models.AssistantReasonCursorExpired
	}
	raw, e := json.Marshal(assistantArgsInput(candidate.Tool, args))
	if e != nil {
		return nil, nil, "", models.AssistantReasonCursorExpired
	}
	normalized, _, e := query.parseArguments(candidate.Tool, raw)
	if e != nil {
		return nil, nil, "", models.AssistantReasonCursorExpired
	}
	// Re-authorize all referenced account/category/tag IDs in today's ledger view.
	parent, cancel := context.WithTimeout(ctx, query.timeout)
	defer cancel()
	qc := core.NewAssistantContext(parent, ctx.GetContextId(), ctx.GetClientLocale(), ctx.ClientIP())
	e = query.database.WithAssistantReadView(qc, func(v datastore.AssistantReadView) error {
		return (&assistantQueryExecution{Service: query, Context: qc, View: v, Tool: candidate.Tool, Args: normalized}).prepare()
	})
	if e != nil {
		return nil, nil, "", models.AssistantReasonNotAccessible
	}
	return nil, nil, "Explicitly restart the saved query from its first page using this re-authorized tool and arguments; the old cursor is expired and must not be reused. Tool: " + string(candidate.Tool) + "; arguments: " + string(raw), nil
}
