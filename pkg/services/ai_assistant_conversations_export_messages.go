package services

import (
	"encoding/json"
	"fmt"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// The question and its attempts form one ordered stream. A group may span SQL
// batches without accumulating versions or loading a complete conversation.
func (s *AssistantConversationService) exportMessages(ctx core.Context, uid, conversationID int64, r models.AssistantExportRequest, selected map[int64]int64, write func(string) error) error {
	afterTurn, afterAttempt := int64(0), -1
	questionOpen, answerWritten := false, false
	currentQuestionID := int64(0)
	for {
		var rows []models.AssistantMessage
		err := s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
			if _, err := assistantConversationLoad(session, uid, conversationID); err != nil {
				return err
			}
			query := session.Table(&models.AssistantMessage{}).Alias("m").Where("m.uid = ? AND m.conversation_id = ? AND (m.turn_sequence > ? OR (m.turn_sequence = ? AND m.attempt_number > ?))", uid, conversationID, afterTurn, afterTurn, afterAttempt)
			if !r.IncludeVersions {
				table := session.Engine().TableName(&models.AssistantMessage{})
				quoted := session.Engine().Quote(table)
				query = query.Join("INNER", []string{table, "q"}, "q.uid = m.uid AND q.id = m.user_message_id")
				defaultChoice := builder.Expr("(m.id = q.selected_assistant_message_id OR (q.selected_assistant_message_id = 0 AND NOT EXISTS (SELECT 1 FROM " + quoted + " newer WHERE newer.uid = m.uid AND newer.user_message_id = m.user_message_id AND newer.attempt_number > m.attempt_number)))")
				choice := builder.Or(builder.Expr("m.role = ?", "user"))
				selectedQuestions := make([]int64, 0, len(selected))
				for question, answer := range selected {
					selectedQuestions = append(selectedQuestions, question)
					choice = choice.Or(builder.Eq{"m.user_message_id": question, "m.id": answer})
				}
				if len(selectedQuestions) > 0 {
					choice = choice.Or(builder.And(builder.NotIn("q.id", selectedQuestions), defaultChoice))
				} else {
					choice = choice.Or(defaultChoice)
				}
				query = query.And(choice)
			}
			return query.Select("m.*").Asc("m.turn_sequence", "m.attempt_number").Limit(int(s.options.Config().ExportBatchSize)).Find(&rows)
		})
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			if questionOpen && r.Format == "json" {
				return write("]}")
			}
			return nil
		}
		for i := range rows {
			m := &rows[i]
			afterTurn = m.TurnSequence
			afterAttempt = m.AttemptNumber
			if m.Role == "user" {
				if r.Format == "json" {
					if questionOpen {
						if err = write("]},"); err != nil {
							return err
						}
					}
					if viewed, ok := selected[m.Id]; ok && !r.IncludeVersions {
						m.SelectedAssistantMessageId = viewed
					}
					raw, err := json.Marshal(assistantQuestionDTO(m))
					if err != nil {
						return err
					}
					if err = write(`{"question":` + string(raw) + `,"answers":[`); err != nil {
						return err
					}
				} else if err = write(fmt.Sprintf("## User · %s\n\n%s\n\n", assistantConversationTime(m.CreatedAt), m.Content)); err != nil {
					return err
				}
				questionOpen = true
				currentQuestionID = m.Id
				answerWritten = false
			} else {
				if !questionOpen || m.Role != "assistant" || m.UserMessageId != currentQuestionID {
					return models.AssistantReasonStorageUnavailable
				}
				if r.Format == "json" && answerWritten {
					if err = write(","); err != nil {
						return err
					}
				}
				if err = assistantExportAnswer(m, r.Format, write); err != nil {
					return err
				}
				answerWritten = true
			}
		}
		if len(rows) < int(s.options.Config().ExportBatchSize) {
			if questionOpen && r.Format == "json" {
				return write("]}")
			}
			return nil
		}
	}
}
func assistantExportAnswer(answer *models.AssistantMessage, format string, write func(string) error) error {
	snapshot, err := AssistantMessageSnapshot(answer)
	if err != nil {
		return err
	}
	if format == "json" {
		raw, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		return write(string(raw))
	}
	if err = write(fmt.Sprintf("## Assistant · %s · %s\n\n%s\n\n", snapshot.SavedAt, snapshot.State, snapshot.Content)); err != nil {
		return err
	}
	if len(snapshot.References) > 0 {
		if err = write("### Sources\n\n"); err != nil {
			return err
		}
		for _, reference := range snapshot.References {
			if err = write(fmt.Sprintf("- Transaction %s · %s · %s %s\n", reference.Id, reference.TimeText, reference.SourceAmountMinor, reference.Currency)); err != nil {
				return err
			}
		}
		return write("\n")
	}
	return nil
}
