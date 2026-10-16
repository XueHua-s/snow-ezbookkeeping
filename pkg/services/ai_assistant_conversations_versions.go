package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/xorm"
)

func (s *AssistantConversationService) Versions(ctx core.Context, uid int64, r models.AssistantVersionsRequest) (models.AssistantVersionsResult, error) {
	result := models.AssistantVersionsResult{Items: []models.AssistantVersionDTO{}}
	id, err := assistantParseID(r.UserMessageId)
	if err != nil {
		return result, err
	}
	limits := s.limits()
	limit := assistantConversationPageLimit(r.Limit, limits.VersionsPageSize, limits.PageSizeMax)
	scope := fmt.Sprintf("versions:%d", id)
	now := s.now()
	cursor := assistantConversationCursor{UID: uid, Scope: scope, Cutoff: now.UnixMilli(), Expires: now.Add(15 * time.Minute).UnixMilli()}
	if r.Cursor != "" {
		cursor, err = decodeAssistantConversationCursor(r.Cursor, uid, scope, s.options.Secret, now)
		if err != nil {
			return result, err
		}
	}
	err = s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		question, err := assistantConversationMessage(session, uid, id)
		if err != nil {
			return err
		}
		if question.Role != "user" {
			return models.AssistantReasonNotAccessible
		}
		query := session.Where("uid = ? AND user_message_id = ? AND role = ? AND created_at <= ?", uid, id, "assistant", cursor.Cutoff)
		if cursor.Sequence > 0 {
			query = query.And("attempt_number < ?", cursor.Sequence)
		}
		var rows []models.AssistantMessage
		if err := query.Desc("attempt_number").Limit(limit + 1).Find(&rows); err != nil {
			return err
		}
		bytes := 64
		for i := 0; i < len(rows) && i < limit; i++ {
			dto := assistantVersionDTO(&rows[i])
			encoded, _ := json.Marshal(dto)
			if bytes+len(encoded) > limits.PageBytes {
				break
			}
			bytes += len(encoded) + 1
			result.Items = append(result.Items, dto)
			cursor.Sequence = int64(rows[i].AttemptNumber)
		}
		if len(result.Items) < len(rows) {
			result.NextCursor, err = encodeAssistantConversationCursor(cursor, s.options.Secret)
			return err
		}
		return nil
	})
	return result, err
}
