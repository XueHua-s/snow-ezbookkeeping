package services

import (
	"encoding/json"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// assistantMemoryDTO projects the current public schema, independently of the
// stored row version. Conversation-only dependencies remain in Sources but do
// not become empty message IDs. Callers decide how to handle corrupt records.
func assistantMemoryDTO(m *models.AssistantMemory) (models.AssistantMemoryDTO, error) {
	sources := []models.AssistantSourceDependency{}
	if m.SourcesJSON != "" && json.Unmarshal([]byte(m.SourcesJSON), &sources) != nil {
		return models.AssistantMemoryDTO{}, models.AssistantReasonStorageUnavailable
	}
	if sources == nil {
		sources = []models.AssistantSourceDependency{}
	}
	dto := models.AssistantMemoryDTO{
		SchemaVersion:    models.AssistantSchemaVersion,
		Id:               assistantConversationID(m.Id),
		Key:              m.Key,
		Text:             m.Text,
		Type:             m.Type,
		Status:           m.Status,
		Sources:          sources,
		SourceMessageIds: []string{},
		Version:          m.Version,
		Explicit:         m.Explicit,
		CreatedAt:        assistantConversationTime(m.CreatedAt),
		UpdatedAt:        assistantConversationTime(m.UpdatedAt),
		ExpiresAt:        assistantConversationTime(m.ExpiresAt),
	}

	for _, source := range sources {
		if source.MessageId != "" {
			dto.SourceMessageIds = append(dto.SourceMessageIds, source.MessageId)
		}
	}
	return dto, nil
}
