package services

import (
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"strings"
	"time"
	"xorm.io/xorm"
)

type AssistantMemoryService struct {
	storage    *AssistantStorageService
	config     func() *settings.Config
	generateID func(uuid.UuidType) int64
	providers  func() (provider.LargeLanguageModelToolProvider, provider.LargeLanguageModelToolProvider, error)
}

func NewAssistantMemoryService(storage *AssistantStorageService, config func() *settings.Config) *AssistantMemoryService {
	return &AssistantMemoryService{storage: storage, config: config, generateID: uuid.Container.GenerateUuid, providers: llm.Container.GetAIAssistantToolProviders}
}
func (s *AssistantMemoryService) List(ctx core.Context, uid int64) (*models.AssistantMemoryList, error) {
	result := &models.AssistantMemoryList{Items: []models.AssistantMemoryDTO{}}
	err := s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		state := &models.AssistantUserState{SettingsVersion: 1}
		if _, e := session.Where("uid = ?", uid).Get(state); e != nil {
			return e
		}
		result.Enabled, result.SettingsVersion, result.MemoryUseEpoch = state.MemoryEnabled, state.SettingsVersion, state.MemoryUseEpoch
		var records []models.AssistantMemory
		if e := session.Where("uid = ? AND excluded = ?", uid, false).Asc("id").Limit(max(1, state.ActiveMemoryCount+state.CandidateMemoryCount)).Find(&records); e != nil {
			return e
		}
		for _, m := range records {
			dto, err := assistantMemoryDTO(&m)
			if err != nil {
				continue
			}
			result.Items = append(result.Items, dto)
		}
		return nil
	})
	return result, err
}
func (s *AssistantMemoryService) Save(ctx core.Context, uid int64, r models.AssistantSaveMemoryRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "memory_save", r.AssistantCommand, r, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		result := models.AssistantCommandResult{}
		if r.ExpectedMemoryUseEpoch != tx.State.MemoryUseEpoch {
			return result, models.AssistantReasonMemoryConflict
		}
		limits := s.config().GetAIAssistantExperienceConfig().Limits
		if strings.TrimSpace(r.Text) == "" || len(r.Text) > limits.MemoryBytes || len(r.Key) > 128 || strings.TrimSpace(r.Key) == "" || (r.Intent != "add" && r.Intent != "correct" && r.Intent != "confirm") {
			return result, models.AssistantReasonMemoryConflict
		}
		if r.Type != "preference" && r.Type != "goal" && r.Type != "background" {
			return result, models.AssistantReasonMemoryConflict
		}
		var expires int64
		if r.ExpiresAt != "" {
			t, e := time.Parse(time.RFC3339Nano, r.ExpiresAt)
			if e != nil || !t.After(tx.Now) {
				return result, models.AssistantReasonMemoryConflict
			}
			expires = t.UnixMilli()
		}
		m := &models.AssistantMemory{}
		if r.Id != "" {
			id, e := assistantParseID(r.Id)
			if e != nil {
				return result, e
			}
			found, e := tx.Session.Where("uid = ? AND id = ? AND excluded = ?", uid, id, false).Get(m)
			if e != nil {
				return result, e
			}
			if !found || m.Version != r.ExpectedVersion {
				return result, models.AssistantReasonMemoryConflict
			}
		} else {
			if r.ExpectedVersion != 0 {
				return result, models.AssistantReasonMemoryConflict
			}
			// A key has one live value. Replacing a conflicting value requires its version.
			found, e := tx.Session.Where("uid = ? AND `key` = ? AND excluded = ?", uid, strings.ToLower(strings.TrimSpace(r.Key)), false).Exist(&models.AssistantMemory{})
			if e != nil {
				return result, e
			}
			if found {
				return result, models.AssistantReasonMemoryConflict
			}
			m = &models.AssistantMemory{Id: s.generateID(uuid.UUID_TYPE_ASSISTANT_MEMORY), Uid: uid, CreatedAt: tx.Now.UnixMilli(), SchemaVersion: models.AssistantSchemaVersion}
		}
		conflict, e := tx.Session.Where("uid = ? AND `key` = ? AND status = ? AND excluded = ? AND id <> ?", uid, strings.ToLower(strings.TrimSpace(r.Key)), "active", false, m.Id).Exist(&models.AssistantMemory{})
		if e != nil {
			return result, e
		}
		if conflict {
			return result, models.AssistantReasonMemoryConflict
		}
		sources, e := assistantMemorySources(tx, r.SourceMessageIds)
		if e != nil {
			return result, e
		}
		if len(r.SourceMessageIds) == 0 && m.SourcesJSON != "" {
			if json.Unmarshal([]byte(m.SourcesJSON), &sources) != nil {
				return result, models.AssistantReasonMemoryConflict
			}
			if e = assistantValidateMemorySources(tx, sources); e != nil {
				return result, e
			}
		}
		priorBytes, priorStatus := m.UsedBytes, m.Status
		m.Text, m.Key, m.Type, m.ExpiresAt = r.Text, strings.ToLower(strings.TrimSpace(r.Key)), r.Type, expires
		m.Status = "active"
		m.Explicit = true
		m.Version++
		m.UpdatedAt = tx.Now.UnixMilli()
		m.GenerationEpoch = tx.State.MemoryGenerationEpoch
		encoded, _ := json.Marshal(sources)
		m.SourcesJSON = string(encoded)
		if priorStatus != "active" {
			if tx.State.ActiveMemoryCount >= limits.Memories {
				return result, models.AssistantReasonQuotaExceeded
			}
			tx.State.ActiveMemoryCount++
			if priorStatus == "candidate" {
				tx.State.CandidateMemoryCount--
			}
		}
		m.UsedBytes, e = assistantMemoryBytes(m)
		if e != nil {
			return result, e
		}
		delta := m.UsedBytes - priorBytes
		if tx.State.UsedBytes+tx.State.ReservedBytes+delta > int64(limits.UserBytes) {
			return result, models.AssistantReasonQuotaExceeded
		}
		tx.State.UsedBytes += delta
		if priorStatus == "" {
			_, e = tx.Session.Insert(m)
		} else {
			_, e = tx.Session.Where("uid = ? AND id = ?", uid, m.Id).AllCols().Update(m)
		}
		if e != nil {
			return result, e
		}
		tx.State.MemoryUseEpoch++
		tx.State.MemoryGenerationEpoch++
		if e = assistantRevokeMemoryRuns(tx); e != nil {
			return result, e
		}
		result.MemoryId = assistantConversationID(m.Id)
		result.MemoryUseEpoch = tx.State.MemoryUseEpoch
		return result, nil
	})
}
func assistantMemoryBytes(m *models.AssistantMemory) (int64, error) {
	return AssistantStorageBytes(struct {
		ID                               int64
		Key, Text, Type, Status, Sources string
		Version, ExpiresAt               int64
	}{m.Id, m.Key, m.Text, m.Type, m.Status, m.SourcesJSON, m.Version, m.ExpiresAt})
}
func assistantMemorySources(tx *AssistantStorageTx, ids []string) ([]models.AssistantSourceDependency, error) {
	if len(ids) > 20 {
		return nil, models.AssistantReasonContextLimit
	}
	sources := []models.AssistantSourceDependency{}
	for _, id := range ids {
		mid, e := assistantParseID(id)
		if e != nil {
			return nil, e
		}
		m, e := assistantConversationMessage(tx.Session, tx.State.Uid, mid)
		if e != nil {
			return nil, e
		}
		c, e := assistantConversationLoad(tx.Session, tx.State.Uid, m.ConversationId)
		if e != nil {
			return nil, e
		}
		if c.MemoryMode == "off" {
			return nil, models.AssistantReasonMemoryConflict
		}
		if m.Role == "user" {
			var answer models.AssistantMessage
			found, e := tx.Session.Where("uid = ? AND user_message_id = ? AND run_state = ?", tx.State.Uid, m.Id, models.AssistantRunCompleted).Get(&answer)
			if e != nil {
				return nil, e
			}
			if !found || strings.TrimSpace(answer.Content) == "" {
				return nil, models.AssistantReasonMemoryConflict
			}
		} else {
			return nil, models.AssistantReasonMemoryConflict
		}
		sources = append(sources, models.AssistantSourceDependency{ConversationId: assistantConversationID(c.Id), MessageId: id, Version: c.SourceVersion})
	}
	return assistantUniqueSources(sources), nil
}
func assistantValidateMemorySources(tx *AssistantStorageTx, sources []models.AssistantSourceDependency) error {
	for _, d := range sources {
		cid, e := assistantParseID(d.ConversationId)
		if e != nil {
			return e
		}
		c, e := assistantConversationLoad(tx.Session, tx.State.Uid, cid)
		if e != nil {
			return e
		}
		if c.SourceVersion != d.Version || c.MemoryMode == "off" {
			return models.AssistantReasonMemoryConflict
		}
	}
	return nil
}
