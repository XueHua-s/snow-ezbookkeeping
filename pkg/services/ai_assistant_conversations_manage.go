package services

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func (s *AssistantConversationService) Update(ctx core.Context, uid int64, r models.AssistantUpdateConversationRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "conversations.update", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		var empty models.AssistantCommandResult
		id, err := assistantParseID(r.Id)
		if err != nil {
			return empty, err
		}
		c, err := assistantConversationLoad(tx.Session, uid, id)
		if err != nil {
			return empty, err
		}
		if c.MetadataVersion != r.ExpectedMetadataVersion {
			return empty, models.AssistantReasonConversationConflict
		}
		if r.Title == nil && r.Archived == nil && r.MemoryMode == nil {
			return empty, models.AssistantReasonConversationConflict
		}
		if r.Title != nil {
			if !utf8.ValidString(*r.Title) || strings.TrimSpace(*r.Title) == "" || utf8.RuneCountInString(*r.Title) > s.limits().TitleCharacters {
				return empty, models.AssistantReasonInputTooLong
			}
			c.Title = *r.Title
			c.TitleManual = true
		}
		if r.Archived != nil {
			c.Archived = *r.Archived
		}
		if r.MemoryMode != nil {
			if (*r.MemoryMode != "inherit" && *r.MemoryMode != "off") || r.ExpectedContextVersion == nil || *r.ExpectedContextVersion != c.ContextVersion {
				return empty, models.AssistantReasonConversationConflict
			}
			if c.MemoryMode != *r.MemoryMode {
				c.MemoryMode = *r.MemoryMode
				c.ContextVersion++
				tx.State.MemoryGenerationEpoch++
				if c.MemoryMode == "off" {
					if err = assistantConversationSave(tx, c); err != nil {
						return empty, err
					}
					if err = tx.InvalidateConversationMemories(c.Id); err != nil {
						return empty, err
					}
					c, err = assistantConversationLoad(tx.Session, uid, id)
					if err != nil {
						return empty, err
					}
				}
				if c.MemoryMode == "off" && c.ActiveRunId != 0 {
					run, err := assistantConversationMessage(tx.Session, uid, c.ActiveRunId)
					if err != nil {
						return empty, err
					}
					var dep models.AssistantInputDependencies
					if run.InputDependenciesJSON != "" && json.Unmarshal([]byte(run.InputDependenciesJSON), &dep) != nil {
						return empty, models.AssistantReasonStorageUnavailable
					}
					if len(dep.Memories) > 0 {
						if err = tx.cancelAssistantRun(c, run, models.AssistantReasonContextChanged); err != nil {
							return empty, err
						}
					}
				}
			}
		}
		c.MetadataVersion++
		if err = assistantSettleConversationHeader(tx, c); err != nil {
			return empty, err
		}
		if err = assistantConversationSave(tx, c); err != nil {
			return empty, err
		}
		return models.AssistantCommandResult{ConversationId: r.Id, MetadataVersion: c.MetadataVersion, ContextVersion: c.ContextVersion}, nil
	})
}
func (s *AssistantConversationService) SelectVersion(ctx core.Context, uid int64, r models.AssistantSelectVersionRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "messages.select_version", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		var empty models.AssistantCommandResult
		id, err := assistantParseID(r.ConversationId)
		if err != nil {
			return empty, err
		}
		c, err := assistantConversationLoad(tx.Session, uid, id)
		if err != nil {
			return empty, err
		}
		if c.ContextVersion != r.ExpectedContextVersion {
			return empty, models.AssistantReasonConversationConflict
		}
		if c.ActiveRunId != 0 {
			return empty, models.AssistantReasonAlreadyRunning
		}
		qid, err := assistantParseID(r.UserMessageId)
		if err != nil {
			return empty, err
		}
		q, err := assistantConversationMessage(tx.Session, uid, qid)
		if err != nil {
			return empty, err
		}
		aid, err := assistantParseID(r.AssistantMessageId)
		if err != nil {
			return empty, err
		}
		a, err := assistantConversationMessage(tx.Session, uid, aid)
		if err != nil {
			return empty, err
		}
		if q.Role != "user" || q.ConversationId != c.Id || q.TurnSequence != c.NextTurnSequence-1 || a.Role != "assistant" || a.UserMessageId != q.Id || a.RunState != models.AssistantRunCompleted {
			return empty, models.AssistantReasonConversationConflict
		}
		q.SelectedAssistantMessageId = a.Id
		c.ContextVersion++
		c.Summary = ""
		c.SummaryThrough = 0
		c.SummarySourcesJSON = ""
		if _, err = tx.Session.Where("uid = ? AND id = ?", uid, q.Id).Cols("selected_assistant_message_id").Update(q); err != nil {
			return empty, err
		}
		if err = assistantSettleConversationHeader(tx, c); err != nil {
			return empty, err
		}
		if err = assistantConversationSave(tx, c); err != nil {
			return empty, err
		}
		return assistantAccepted(c, q, a), nil
	})
}
func (s *AssistantConversationService) Delete(ctx core.Context, uid int64, r models.AssistantDeleteConversationRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "conversations.delete", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		var empty models.AssistantCommandResult
		id, err := assistantParseID(r.Id)
		if err != nil {
			return empty, err
		}
		c, err := assistantConversationLoad(tx.Session, uid, id)
		if err != nil {
			return empty, err
		}
		if c.MetadataVersion != r.ExpectedMetadataVersion {
			return empty, models.AssistantReasonConversationConflict
		}
		if err = tx.InvalidateSources(c.Id); err != nil {
			return empty, err
		}
		c, err = assistantConversationLoad(tx.Session, uid, id)
		if err != nil {
			return empty, err
		}
		if err = assistantDeleteConversation(tx, c); err != nil {
			return empty, err
		}
		return models.AssistantCommandResult{ConversationId: r.Id, MetadataVersion: c.MetadataVersion, ContextVersion: c.ContextVersion, Status: "deleted"}, nil
	})
}
func (s *AssistantConversationService) Clear(ctx core.Context, uid int64, r models.AssistantClearRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "conversations.clear", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		return models.AssistantCommandResult{}, tx.ClearAssistantData()
	})
}
func assistantDeleteConversation(tx *AssistantStorageTx, c *models.AssistantConversation) error {
	if c.ActiveRunId != 0 {
		run, err := assistantConversationMessage(tx.Session, tx.State.Uid, c.ActiveRunId)
		if err != nil {
			return err
		}
		// Deleted content cannot receive another checkpoint; reclaim its reservation
		// now while leaving cancellation persisted for the active worker to observe.
		if !run.ReservationSettled {
			if err = tx.SettleRun(c, run, run.UsedBytes); err != nil {
				return err
			}
		}
		run.CancelRequested = true
		run.RunState = models.AssistantRunCanceled
		run.CompletedAt = tx.Now.UnixMilli()
		run.SnapshotRevision++
		if _, err = tx.Session.Where("uid = ? AND id = ?", tx.State.Uid, run.Id).AllCols().Update(run); err != nil {
			return err
		}
	}
	tx.State.UsedBytes -= c.UsedBytes
	tx.State.ReservedBytes -= c.ReservedBytes
	tx.State.ConversationCount--
	c.Deleted = true
	c.DeletedAt = tx.Now.UnixMilli()
	c.MetadataVersion++
	c.ContextVersion++
	c.SourceVersion++
	c.ActiveRunId = 0
	c.UsedBytes = 0
	c.ReservedBytes = 0
	c.HeaderReservedBytes = 0
	c.HeaderUsedBytes = 0
	c.Summary = ""
	c.SeedContext = ""
	c.FrozenPrefixJSON = ""
	if err := tx.TombstoneReceipts(c.Id, 0, false); err != nil {
		return err
	}
	return assistantConversationSave(tx, c)
}
