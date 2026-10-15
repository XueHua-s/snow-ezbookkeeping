package services

import (
	"encoding/json"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// CancelConversationRuns is shared by deletion, memory revocation and explicit
// cancellation. A worker must observe the persisted state before any late write.
func (tx *AssistantStorageTx) CancelConversationRuns(conversationID int64, reason models.AssistantReasonCode) error {
	c, err := assistantConversationLoad(tx.Session, tx.State.Uid, conversationID)
	if err != nil {
		return err
	}
	if c.ActiveRunId == 0 {
		return nil
	}
	run, err := assistantConversationMessage(tx.Session, tx.State.Uid, c.ActiveRunId)
	if err != nil {
		return err
	}
	return tx.cancelAssistantRun(c, run, reason)
}
func (tx *AssistantStorageTx) cancelAssistantRun(c *models.AssistantConversation, run *models.AssistantMessage, reason models.AssistantReasonCode) error {
	if run.RunState.IsTerminal() {
		return nil
	}
	run.CancelRequested = true
	run.ReasonCode = reason
	run.SnapshotRevision++
	run.SnapshotSavedAt = tx.Now.UnixMilli()
	if run.RunState == models.AssistantRunQueued {
		run.RunState = models.AssistantRunCanceled
		run.CompletedAt = tx.Now.UnixMilli()
		question, err := assistantConversationMessage(tx.Session, tx.State.Uid, run.UserMessageId)
		if err != nil {
			return err
		}
		actual, err := AssistantAttemptStorageBytes(question, run)
		if err != nil {
			return err
		}
		if err = tx.SettleRun(c, run, actual); err != nil {
			return err
		}
		c.ActiveRunId = 0
		c.ActivityAt = tx.Now.UnixMilli()
		c.LastRunState = run.RunState
		c.LastReasonCode = reason
	} else {
		run.RunState = models.AssistantRunCancelRequested
	}
	if _, err := tx.Session.Where("uid = ? AND id = ?", tx.State.Uid, run.Id).AllCols().Update(run); err != nil {
		return err
	}
	return assistantConversationSave(tx, c)
}

// InvalidateSources removes derived context immediately, while preserving the
// independent messages of descendant conversations. Sources are already flat,
// so this walk never recursively loads ancestral message bodies.
func (tx *AssistantStorageTx) InvalidateSources(conversationID int64) error {
	sourceID := assistantConversationID(conversationID)
	var after int64
	for {
		var rows []models.AssistantConversation
		if err := tx.Session.Where("uid = ? AND deleted = ? AND id > ?", tx.State.Uid, false, after).Asc("id").Limit(50).Find(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		for i := range rows {
			c := &rows[i]
			after = c.Id
			invalidSeed, err := assistantSourceContains(c.SeedSourcesJSON, sourceID)
			if err != nil {
				return err
			}
			invalidSummary, err := assistantSourceContains(c.SummarySourcesJSON, sourceID)
			if err != nil {
				return err
			}
			if invalidSeed || invalidSummary {
				if invalidSeed {
					c.SourceInvalidated = true
					c.PendingQuerySourceMessageId = 0
					c.SeedContext = ""
					c.FrozenPrefixJSON = ""
					c.SeedSourcesJSON = ""
				}
				c.Summary = ""
				c.SummaryThrough = 0
				c.SummarySourcesJSON = ""
				c.ContextVersion++
				if err = assistantSettleConversationHeader(tx, c); err != nil {
					return err
				}
				if err = assistantConversationSave(tx, c); err != nil {
					return err
				}
			}
			if c.ActiveRunId != 0 {
				run, err := assistantConversationMessage(tx.Session, tx.State.Uid, c.ActiveRunId)
				if err != nil {
					return err
				}
				var dependencies models.AssistantInputDependencies
				if run.InputDependenciesJSON != "" && json.Unmarshal([]byte(run.InputDependenciesJSON), &dependencies) != nil {
					return models.AssistantReasonStorageUnavailable
				}
				affected := invalidSeed || invalidSummary
				for _, source := range dependencies.Sources {
					if source.ConversationId == sourceID {
						affected = true
					}
				}
				if affected {
					if err = tx.cancelAssistantRun(c, run, models.AssistantReasonContextChanged); err != nil {
						return err
					}
				}
			}
		}
	}
	var memoryAfter int64
	for {
		var memories []models.AssistantMemory
		if err := tx.Session.Where("uid = ? AND excluded = ? AND id > ?", tx.State.Uid, false, memoryAfter).Asc("id").Limit(50).Find(&memories); err != nil {
			return err
		}
		if len(memories) == 0 {
			break
		}
		for i := range memories {
			m := &memories[i]
			memoryAfter = m.Id
			affected, err := assistantSourceContains(m.SourcesJSON, sourceID)
			if err != nil {
				return err
			}
			if !affected {
				continue
			}
			if m.Status == "active" {
				tx.State.ActiveMemoryCount--
			} else if m.Status == "candidate" {
				tx.State.CandidateMemoryCount--
			}
			tx.State.UsedBytes -= m.UsedBytes
			m.UsedBytes = 0
			m.Text = ""
			m.Excluded = true
			m.Status = "invalidated"
			m.Version++
			m.UpdatedAt = tx.Now.UnixMilli()
			tx.State.MemoryUseEpoch++
			tx.State.MemoryGenerationEpoch++
			if _, err = tx.Session.Where("uid = ? AND id = ?", tx.State.Uid, m.Id).AllCols().Update(m); err != nil {
				return err
			}
		}
	}
	// Memory invalidation can affect runs whose direct seed is unrelated.
	var active []models.AssistantConversation
	if err := tx.Session.Where("uid = ? AND deleted = ? AND active_run_id <> ?", tx.State.Uid, false, 0).Limit(4).Find(&active); err != nil {
		return err
	}
	for i := range active {
		c := &active[i]
		run, err := assistantConversationMessage(tx.Session, tx.State.Uid, c.ActiveRunId)
		if err != nil {
			return err
		}
		var dep models.AssistantInputDependencies
		if run.InputDependenciesJSON == "" {
			continue
		}
		if json.Unmarshal([]byte(run.InputDependenciesJSON), &dep) != nil {
			return models.AssistantReasonStorageUnavailable
		}
		if err = tx.ValidateDependencies(c, &dep); err == models.AssistantReasonContextChanged {
			if err = tx.cancelAssistantRun(c, run, models.AssistantReasonContextChanged); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}
func assistantSourceContains(raw, conversationID string) (bool, error) {
	if raw == "" {
		return false, nil
	}
	var sources []models.AssistantSourceDependency
	if json.Unmarshal([]byte(raw), &sources) != nil {
		return false, models.AssistantReasonStorageUnavailable
	}
	for _, source := range sources {
		if source.ConversationId == conversationID {
			return true, nil
		}
	}
	return false, nil
}
