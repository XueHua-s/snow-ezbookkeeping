package services

import (
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"strings"
)

func (s *AssistantMemoryService) Delete(ctx core.Context, uid int64, r models.AssistantDeleteMemoryRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "memory_delete", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		result := models.AssistantCommandResult{}
		if r.ExpectedMemoryUseEpoch != tx.State.MemoryUseEpoch || r.All == (r.Id != "") {
			return result, models.AssistantReasonMemoryConflict
		}
		var records []models.AssistantMemory
		query := tx.Session.Where("uid = ? AND excluded = ?", uid, false)
		if !r.All {
			id, e := assistantParseID(r.Id)
			if e != nil {
				return result, e
			}
			query = query.And("id = ?", id)
		}
		if e := query.Limit(max(1, tx.State.ActiveMemoryCount+tx.State.CandidateMemoryCount)).Find(&records); e != nil {
			return result, e
		}
		if !r.All && (len(records) != 1 || records[0].Version != r.ExpectedVersion) {
			return result, models.AssistantReasonMemoryConflict
		}
		for i := range records {
			if e := assistantForgetMemory(tx, &records[i]); e != nil {
				return result, e
			}
		}
		tx.State.MemoryGenerationEpoch++
		tx.State.MemoryUseEpoch++
		if e := assistantRevokeMemoryRuns(tx); e != nil {
			return result, e
		}
		result.MemoryUseEpoch = tx.State.MemoryUseEpoch
		return result, nil
	})
}
func assistantForgetMemory(tx *AssistantStorageTx, m *models.AssistantMemory) error {
	tx.State.UsedBytes -= m.UsedBytes
	if m.Status == "active" {
		tx.State.ActiveMemoryCount--
	} else if m.Status == "candidate" {
		tx.State.CandidateMemoryCount--
	}
	// Keep only the key and source exclusion boundary; forgotten prose is removed.
	m.Text = ""
	m.UsedBytes = 0
	m.Excluded = true
	m.Status = "forgotten"
	m.Version++
	m.UpdatedAt = tx.Now.UnixMilli()
	if _, e := tx.Session.Where("uid = ? AND id = ?", tx.State.Uid, m.Id).AllCols().Update(m); e != nil {
		return e
	}
	return tx.TombstoneReceipts(0, m.Id, false)
}
func (s *AssistantMemoryService) Settings(ctx core.Context, uid int64, r models.AssistantMemorySettingsRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "memory_settings", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		result := models.AssistantCommandResult{}
		if r.ExpectedSettingsVersion != tx.State.SettingsVersion {
			return result, models.AssistantReasonMemoryConflict
		}
		if tx.State.MemoryEnabled != r.Enabled {
			tx.State.MemoryEnabled = r.Enabled
			tx.State.SettingsVersion++
			tx.State.MemoryUseEpoch++
			tx.State.MemoryGenerationEpoch++
			if e := assistantRevokeMemoryRuns(tx); e != nil {
				return result, e
			}
		}
		result.SettingsVersion = tx.State.SettingsVersion
		result.MemoryUseEpoch = tx.State.MemoryUseEpoch
		return result, nil
	})
}

// Only tasks whose frozen input is actually invalid are canceled. An unrelated
// preference addition increments the revalidation epoch but preserves their run.
func assistantRevokeMemoryRuns(tx *AssistantStorageTx) error {
	var runs []models.AssistantMessage
	if e := tx.Session.Where("uid = ? AND role = ?", tx.State.Uid, "assistant").In("run_state", models.AssistantRunQueued, models.AssistantRunPreparing, models.AssistantRunRunning).Limit(max(1, tx.State.RunningCount+tx.State.QueuedCount)).Find(&runs); e != nil {
		return e
	}
	for _, run := range runs {
		if run.InputDependenciesJSON == "" {
			continue
		}
		var deps models.AssistantInputDependencies
		if json.Unmarshal([]byte(run.InputDependenciesJSON), &deps) != nil {
			return models.AssistantReasonStorageUnavailable
		}
		c, e := assistantConversationLoad(tx.Session, tx.State.Uid, run.ConversationId)
		if e != nil {
			return e
		}
		if e = tx.ValidateDependencies(c, &deps); e == nil {
			continue
		} else if e != models.AssistantReasonContextChanged {
			return e
		}
		if e = tx.cancelAssistantRun(c, &run, models.AssistantReasonContextChanged); e != nil {
			return e
		}
	}
	return nil
}
func (s *AssistantMemoryService) selectContext(ctx core.Context, snap *assistantContextSnapshot, acceptedLimit ...int) (string, error) {
	if !snap.State.MemoryEnabled || snap.Conversation.MemoryMode == "off" {
		return "", nil
	}
	text := ""
	var selected []models.AssistantMemoryDependency
	var selectedSources []models.AssistantSourceDependency
	limit := s.config().GetAIAssistantExperienceConfig().Limits.MemoryTokens
	if len(acceptedLimit) > 0 {
		limit = acceptedLimit[0]
	}
	err := s.storage.WithUserTransaction(ctx, snap.State.Uid, func(tx *AssistantStorageTx) error {
		text = ""
		selected = nil
		selectedSources = nil
		if !tx.State.MemoryEnabled {
			return nil
		}
		var memories []models.AssistantMemory
		if e := tx.Session.Where("uid = ? AND excluded = ? AND status = ?", tx.State.Uid, false, "active").Desc("explicit", "updated_at").Limit(max(1, tx.State.ActiveMemoryCount)).Find(&memories); e != nil {
			return e
		}
		for _, m := range memories {
			if m.ExpiresAt > 0 && m.ExpiresAt <= tx.Now.UnixMilli() {
				continue
			}
			var sources []models.AssistantSourceDependency
			if m.SourcesJSON != "" && json.Unmarshal([]byte(m.SourcesJSON), &sources) != nil {
				continue
			}
			if assistantValidateMemorySources(tx, sources) != nil {
				continue
			}
			// Global response-style preferences apply everywhere; goals/background only
			// match explicitly mentioned keys, avoiding unrelated financial assumptions.
			if m.Type != "preference" && !strings.Contains(strings.ToLower(snap.Question.Content), strings.ToLower(m.Key)) {
				continue
			}
			item := "\n" + m.Text
			if data.EstimateAssistantTokens(text+item) > limit {
				continue
			}
			text += item
			selected = append(selected, models.AssistantMemoryDependency{Id: assistantConversationID(m.Id), Version: m.Version})
			selectedSources = append(selectedSources, sources...)
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	snap.Dependencies.Memories = append(snap.Dependencies.Memories, selected...)
	snap.Dependencies.Sources = assistantUniqueSources(append(snap.Dependencies.Sources, selectedSources...))
	return text, nil
}

// InvalidateConversationMemories changes only memory contribution, not branch
// source versions. Switching a source conversation off cannot revoke its seed.
func (tx *AssistantStorageTx) InvalidateConversationMemories(conversationID int64) error {
	var records []models.AssistantMemory
	if e := tx.Session.Where("uid = ? AND excluded = ?", tx.State.Uid, false).Limit(max(1, tx.State.ActiveMemoryCount+tx.State.CandidateMemoryCount)).Find(&records); e != nil {
		return e
	}
	changed := false
	for i := range records {
		var sources []models.AssistantSourceDependency
		if json.Unmarshal([]byte(records[i].SourcesJSON), &sources) != nil {
			return models.AssistantReasonStorageUnavailable
		}
		for _, source := range sources {
			if source.ConversationId == assistantConversationID(conversationID) {
				if e := assistantForgetMemory(tx, &records[i]); e != nil {
					return e
				}
				changed = true
				break
			}
		}
	}
	if changed {
		tx.State.MemoryUseEpoch++
		tx.State.MemoryGenerationEpoch++
		return assistantRevokeMemoryRuns(tx)
	}
	return nil
}
