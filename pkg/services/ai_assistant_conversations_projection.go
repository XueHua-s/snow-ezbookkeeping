package services

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// AssistantFrozenTurn records the immutable question and successful answer chosen
// when a branch is created. Later parent appends and selections cannot alter it.
type AssistantFrozenTurn struct {
	ConversationId     string `json:"conversationId"`
	UserMessageId      string `json:"userMessageId"`
	AssistantMessageId string `json:"assistantMessageId"`
}

func assistantConversationID(id int64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}
func assistantParseID(id string) (int64, error) {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil || n <= 0 {
		return 0, models.AssistantReasonNotAccessible
	}
	return n, nil
}
func assistantConversationTime(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format(time.RFC3339Nano)
}
func assistantConversationDTO(c *models.AssistantConversation) models.AssistantConversationDTO {
	return models.AssistantConversationDTO{SchemaVersion: c.SchemaVersion, Id: assistantConversationID(c.Id), Title: c.Title, BranchKind: c.BranchKind, LastRunState: c.LastRunState, LastReasonCode: c.LastReasonCode, Archived: c.Archived, CreatedAt: assistantConversationTime(c.CreatedAt), ActivityAt: assistantConversationTime(c.ActivityAt), MetadataVersion: c.MetadataVersion, ContextVersion: c.ContextVersion, NextTurnSequence: c.NextTurnSequence, ActiveRunId: assistantConversationID(c.ActiveRunId), UsedBytes: c.UsedBytes, ReservedBytes: c.ReservedBytes, TurnCount: c.TurnCount, MemoryMode: c.MemoryMode, SourceConversationId: assistantConversationID(c.SourceConversationId), SourceMessageId: assistantConversationID(c.SourceMessageId), PendingContinuation: c.PendingContinuation, PendingQuerySourceMessageId: assistantConversationID(c.PendingQuerySourceMessageId), SourceInvalidated: c.SourceInvalidated}
}
func assistantVersionDTO(m *models.AssistantMessage) models.AssistantVersionDTO {
	return models.AssistantVersionDTO{Id: assistantConversationID(m.Id), AttemptNumber: m.AttemptNumber, State: m.RunState, CreatedAt: assistantConversationTime(m.CreatedAt), CompletedAt: assistantConversationTime(m.CompletedAt), ReasonCode: m.ReasonCode}
}

// AssistantMessageSnapshot is the only public persisted-answer projection; leases,
// provider cursors and raw errors never enter history, exports or subscriptions.
func AssistantMessageSnapshot(m *models.AssistantMessage) (models.AssistantRunSnapshot, error) {
	s := models.AssistantRunSnapshot{SchemaVersion: m.SchemaVersion, RunId: assistantConversationID(m.Id), ConversationId: assistantConversationID(m.ConversationId), UserMessageId: assistantConversationID(m.UserMessageId), AssistantMessageId: assistantConversationID(m.Id), AttemptNumber: m.AttemptNumber, Revision: m.SnapshotRevision, State: m.RunState, Progress: m.Progress, Content: m.Content, ProcessSummary: m.ProcessSummary, Queries: []models.AIAssistantQuery{}, References: []models.AssistantReferenceDTO{}, SavedAt: assistantConversationTime(m.CreatedAt), PageableQueryId: m.ContinuationQueryId}
	if m.SnapshotSavedAt != 0 {
		s.SavedAt = assistantConversationTime(m.SnapshotSavedAt)
	} else if m.CompletedAt != 0 {
		s.SavedAt = assistantConversationTime(m.CompletedAt)
	}
	if m.ReasonCode != "" {
		reason := m.ReasonCode
		s.ReasonCode = &reason
	}
	if m.SchemaVersion != models.AssistantSchemaVersion {
		return s, models.AssistantReasonStorageUnavailable
	}
	if m.QueriesJSON != "" && json.Unmarshal([]byte(m.QueriesJSON), &s.Queries) != nil {
		return s, models.AssistantReasonStorageUnavailable
	}
	if m.ReferencesJSON != "" && json.Unmarshal([]byte(m.ReferencesJSON), &s.References) != nil {
		return s, models.AssistantReasonStorageUnavailable
	}
	if err := s.Validate(); err != nil {
		return s, models.AssistantReasonStorageUnavailable
	}
	return s, nil
}
func assistantQuestionDTO(m *models.AssistantMessage) models.AssistantMessageDTO {
	return models.AssistantMessageDTO{SchemaVersion: m.SchemaVersion, Id: assistantConversationID(m.Id), ConversationId: assistantConversationID(m.ConversationId), TurnSequence: m.TurnSequence, Role: m.Role, Content: m.Content, CreatedAt: assistantConversationTime(m.CreatedAt), SelectedAssistantMessageId: assistantConversationID(m.SelectedAssistantMessageId), PreviousAssistantMessageId: assistantConversationID(m.PreviousAssistantMessageId), Locale: m.Locale, Timezone: m.Timezone, Versions: []models.AssistantVersionDTO{}}
}
