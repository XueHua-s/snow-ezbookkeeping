package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/xorm"
)

func (s *AssistantConversationService) List(ctx core.Context, uid int64, r models.AssistantListRequest) (models.AssistantConversationList, error) {
	result := models.AssistantConversationList{Items: []models.AssistantConversationDTO{}}
	limits := s.limits()
	limit := assistantConversationPageLimit(r.Limit, limits.PageSize, limits.PageSizeMax)
	scope := fmt.Sprintf("list:%t", r.Archived)
	now := s.now()
	cursor := assistantConversationCursor{UID: uid, Scope: scope, Cutoff: now.UnixMilli(), Expires: now.Add(15 * time.Minute).UnixMilli()}
	if r.Cursor != "" {
		var err error
		cursor, err = decodeAssistantConversationCursor(r.Cursor, uid, scope, s.options.Secret, now)
		if err != nil {
			return result, err
		}
	}
	err := s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		query := session.Where("uid = ? AND deleted = ? AND activity_at <= ?", uid, false, cursor.Cutoff)
		if !r.Archived {
			query = query.And("archived = ?", false)
		}
		if cursor.ID != "" {
			id, err := assistantParseID(cursor.ID)
			if err != nil {
				return err
			}
			query = query.And("(activity_at < ? OR (activity_at = ? AND id < ?))", cursor.Time, cursor.Time, id)
		}
		var rows []models.AssistantConversation
		if err := query.Desc("activity_at", "id").Limit(limit + 1).Find(&rows); err != nil {
			return err
		}
		count := len(rows)
		if count > limit {
			count = limit
		}
		bytes := 64
		for i := 0; i < count; i++ {
			dto := assistantConversationDTO(&rows[i])
			encoded, _ := json.Marshal(dto)
			if bytes+len(encoded) > limits.PageBytes {
				break
			}
			result.Items = append(result.Items, dto)
			bytes += len(encoded) + 1
			cursor.Time = rows[i].ActivityAt
			cursor.ID = dto.Id
		}
		if len(result.Items) < len(rows) {
			var err error
			result.NextCursor, err = encodeAssistantConversationCursor(cursor, s.options.Secret)
			return err
		}
		return nil
	})
	return result, err
}

func (s *AssistantConversationService) State(ctx core.Context, uid int64, id string) (models.AssistantConversationState, error) {
	var result models.AssistantConversationState
	numeric, err := assistantParseID(id)
	if err != nil {
		return result, err
	}
	err = s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		c, err := assistantConversationLoad(session, uid, numeric)
		if err != nil {
			return err
		}
		result = models.AssistantConversationState{Id: id, MetadataVersion: c.MetadataVersion, ContextVersion: c.ContextVersion, ActiveRunId: assistantConversationID(c.ActiveRunId), Archived: c.Archived, MemoryMode: c.MemoryMode, Accessible: true}
		if c.ActiveRunId != 0 {
			m, err := assistantConversationMessage(session, uid, c.ActiveRunId)
			if err != nil {
				return err
			}
			result.ActiveRunRevision = m.SnapshotRevision
		}
		var last models.AssistantMessage
		found, err := session.Where("uid = ? AND conversation_id = ? AND role = ?", uid, numeric, "user").Desc("turn_sequence").Get(&last)
		if found {
			result.SelectedAssistantMessageId = assistantConversationID(last.SelectedAssistantMessageId)
		}
		return err
	})
	return result, err
}

func (s *AssistantConversationService) GetMessage(ctx core.Context, uid int64, id string) (models.AssistantRunSnapshot, error) {
	var result models.AssistantRunSnapshot
	numeric, err := assistantParseID(id)
	if err != nil {
		return result, err
	}
	err = s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		m, err := assistantConversationMessage(session, uid, numeric)
		if err != nil {
			return err
		}
		if m.Role != "assistant" {
			return models.AssistantReasonNotAccessible
		}
		result, err = AssistantMessageSnapshot(m)
		return err
	})
	return result, err
}

func (s *AssistantConversationService) Get(ctx core.Context, uid int64, r models.AssistantGetConversationRequest) (models.AssistantConversationDetail, error) {
	result := models.AssistantConversationDetail{Messages: []models.AssistantMessageDTO{}}
	id, err := assistantParseID(r.Id)
	if err != nil {
		return result, err
	}
	limits := s.limits()
	limit := assistantConversationPageLimit(r.Limit, limits.PageSize, limits.PageSizeMax)
	err = s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		c, err := assistantConversationLoad(session, uid, id)
		if err != nil {
			return err
		}
		result.Conversation = assistantConversationDTO(c)
		before := r.BeforeTurnSequence
		if r.AroundMessageId != "" {
			targetID, err := assistantParseID(r.AroundMessageId)
			if err != nil {
				return err
			}
			target, err := assistantConversationMessage(session, uid, targetID)
			if err != nil {
				return err
			}
			if target.ConversationId != id {
				return models.AssistantReasonNotAccessible
			}
			before = target.TurnSequence + 1
		}
		query := session.Where("uid = ? AND conversation_id = ? AND role = ?", uid, id, "user")
		if before > 0 {
			query = query.And("turn_sequence < ?", before)
		}
		var rows []models.AssistantMessage
		if err := query.Desc("turn_sequence").Limit(limit + 1).Find(&rows); err != nil {
			return err
		}
		encoded, _ := json.Marshal(result)
		bytes := len(encoded) + 1024
		for i := 0; i < len(rows) && i < limit; i++ {
			dto, err := s.questionDetail(session, uid, &rows[i])
			if err != nil {
				return err
			}
			encoded, _ := json.Marshal(dto)
			if bytes+len(encoded) > limits.PageBytes {
				break
			}
			bytes += len(encoded) + 1
			result.Messages = append(result.Messages, dto)
		}
		result.HasMore = len(result.Messages) < len(rows)
		if result.HasMore && len(result.Messages) > 0 {
			result.NextCursor = strconv.FormatInt(result.Messages[len(result.Messages)-1].TurnSequence, 10)
		}
		for left, right := 0, len(result.Messages)-1; left < right; left, right = left+1, right-1 {
			result.Messages[left], result.Messages[right] = result.Messages[right], result.Messages[left]
		}
		// ActiveRun is included only if its full snapshot fits the same response budget.
		if c.ActiveRunId != 0 {
			m, err := assistantConversationMessage(session, uid, c.ActiveRunId)
			if err != nil {
				return err
			}
			snap, err := AssistantMessageSnapshot(m)
			if err != nil {
				return err
			}
			encoded, _ := json.Marshal(snap)
			if bytes+len(encoded) <= limits.PageBytes {
				result.ActiveRun = &snap
			}
		}
		return nil
	})
	return result, err
}
func (s *AssistantConversationService) questionDetail(session *xorm.Session, uid int64, m *models.AssistantMessage) (models.AssistantMessageDTO, error) {
	dto := assistantQuestionDTO(m)
	var versions []models.AssistantMessage
	limit := s.limits().RecentVersions
	if err := session.Where("uid = ? AND user_message_id = ? AND role = ?", uid, m.Id, "assistant").Desc("attempt_number").Limit(limit + 1).Find(&versions); err != nil {
		return dto, err
	}
	dto.HasMoreVersions = len(versions) > limit
	if len(versions) > limit {
		versions = versions[:limit]
	}
	for i := range versions {
		dto.Versions = append(dto.Versions, assistantVersionDTO(&versions[i]))
	}
	selected := m.SelectedAssistantMessageId
	if selected == 0 && len(versions) > 0 {
		selected = versions[0].Id
	}
	if selected != 0 {
		answer, err := assistantConversationMessage(session, uid, selected)
		if err != nil {
			return dto, err
		}
		if answer.UserMessageId != m.Id {
			return dto, models.AssistantReasonStorageUnavailable
		}
		snapshot, err := AssistantMessageSnapshot(answer)
		if err != nil {
			snapshot.Content = ""
			snapshot.State = models.AssistantRunFailed
			reason := models.AssistantReasonStorageUnavailable
			snapshot.ReasonCode = &reason
			snapshot.Queries = []models.AIAssistantQuery{}
			snapshot.References = []models.AssistantReferenceDTO{}
		}
		dto.SelectedAnswer = &snapshot
	}
	return dto, nil
}
