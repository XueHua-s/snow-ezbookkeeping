package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestGetConfigWithoutSensitiveData_EmbeddingKey(t *testing.T) {
	config := &settings.Config{DatabaseConfig: &settings.DatabaseConfig{}, SMTPConfig: &settings.SMTPConfig{}, MinIOConfig: &settings.MinIOConfig{}, S3Config: &settings.S3Config{}, AIAssistantEmbeddingLLMConfig: &settings.LLMConfig{OpenAIAPIKey: "embedding-secret"}, CodingAssistantLLMConfig: &settings.LLMConfig{OpenAIAPIKey: "coding-secret"}}
	redacted := getConfigWithoutSensitiveData(config)
	assert.Equal(t, "****", redacted.AIAssistantEmbeddingLLMConfig.OpenAIAPIKey)
	assert.Equal(t, "embedding-secret", config.AIAssistantEmbeddingLLMConfig.OpenAIAPIKey)
	assert.Equal(t, "****", redacted.CodingAssistantLLMConfig.OpenAIAPIKey)
	assert.Equal(t, "coding-secret", config.CodingAssistantLLMConfig.OpenAIAPIKey)
}
