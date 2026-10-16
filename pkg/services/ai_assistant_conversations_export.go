package services

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/xorm"
)

// Export never collects the user's history in memory. Each question and answer
// is authorized immediately before writing, and cancellation stops further reads.
func (s *AssistantConversationService) Export(ctx core.Context, uid int64, r models.AssistantExportRequest, writer io.Writer) error {
	if (r.Id == "" && !r.All) || (r.Id != "" && r.All) || (r.Format != "json" && r.Format != "markdown") {
		return models.AssistantReasonConversationConflict
	}
	if uid <= 0 {
		return models.AssistantReasonNotAccessible
	}
	selected, err := s.exportViewedVersions(ctx, uid, r)
	if err != nil {
		return err
	}
	if r.Id != "" {
		if _, err := s.State(ctx, uid, r.Id); err != nil {
			return err
		}
	}
	write := func(value string) error {
		if ctx != nil && ctx.Err() != nil {
			return ctx.Err()
		}
		_, err := io.WriteString(writer, value)
		return err
	}
	if r.Format == "json" {
		if err := write(`{"schemaVersion":1,"conversations":[`); err != nil {
			return err
		}
	}
	after := int64(0)
	first := true
	for {
		var conversations []models.AssistantConversation
		err := s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
			q := session.Where("uid = ? AND deleted = ? AND id > ?", uid, false, after)
			if r.Id != "" {
				id, err := assistantParseID(r.Id)
				if err != nil {
					return err
				}
				q = q.And("id = ?", id)
			}
			return q.Asc("id").Limit(int(s.options.Config().ExportBatchSize)).Find(&conversations)
		})
		if err != nil {
			return err
		}
		if len(conversations) == 0 {
			break
		}
		for i := range conversations {
			c := &conversations[i]
			after = c.Id
			if ctx != nil && ctx.Err() != nil {
				return ctx.Err()
			}
			if r.Format == "json" {
				if !first {
					if err = write(","); err != nil {
						return err
					}
				}
				raw, err := json.Marshal(assistantConversationDTO(c))
				if err != nil {
					return err
				}
				if err = write(`{"conversation":` + string(raw) + `,"messages":[`); err != nil {
					return err
				}
			} else {
				if err = write("# " + c.Title + "\n\n"); err != nil {
					return err
				}
			}
			if err = s.exportMessages(ctx, uid, c.Id, r, selected, write); err != nil {
				return err
			}
			if r.Format == "json" {
				if err = write("]}"); err != nil {
					return err
				}
			}
			first = false
		}
	}
	if r.Format == "json" {
		if err := write(`],"memories":[`); err != nil {
			return err
		}
	}
	if r.All {
		if err := s.exportMemories(ctx, uid, r.Format, write); err != nil {
			return err
		}
	}
	if r.Format == "json" {
		return write("]}")
	}
	return nil
}
func (s *AssistantConversationService) exportMemories(ctx core.Context, uid int64, format string, write func(string) error) error {
	after := int64(0)
	first := true
	for {
		var rows []models.AssistantMemory
		err := s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
			return session.Where("uid = ? AND excluded = ? AND id > ?", uid, false, after).Asc("id").Limit(20).Find(&rows)
		})
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		for i := range rows {
			m := &rows[i]
			after = m.Id
			if format == "json" {
				if !first {
					if err = write(","); err != nil {
						return err
					}
				}
				dto, err := assistantMemoryDTO(m)
				if err != nil {
					return err
				}
				raw, err := json.Marshal(dto)
				if err != nil {
					return err
				}
				if err = write(string(raw)); err != nil {
					return err
				}
			} else {
				if err = write(fmt.Sprintf("## Memory · %s\n\n%s\n\n", m.Key, m.Text)); err != nil {
					return err
				}
			}
			first = false
		}
	}
}

func (s *AssistantConversationService) exportViewedVersions(ctx core.Context, uid int64, r models.AssistantExportRequest) (map[int64]int64, error) {
	selected := map[int64]int64{}
	if len(r.ViewedVersionIds) == 0 {
		return selected, nil
	}
	if r.All || len(r.ViewedVersionIds) > s.limits().ConversationTurns {
		return nil, models.AssistantReasonConversationConflict
	}
	conversationID, err := assistantParseID(r.Id)
	if err != nil {
		return nil, err
	}
	err = s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		for _, id := range r.ViewedVersionIds {
			numeric, err := assistantParseID(id)
			if err != nil {
				return err
			}
			answer, err := assistantConversationMessage(session, uid, numeric)
			if err != nil {
				return err
			}
			if answer.Role != "assistant" || answer.ConversationId != conversationID {
				return models.AssistantReasonNotAccessible
			}
			if _, duplicate := selected[answer.UserMessageId]; duplicate {
				return models.AssistantReasonConversationConflict
			}
			selected[answer.UserMessageId] = answer.Id
		}
		return nil
	})
	return selected, err
}
