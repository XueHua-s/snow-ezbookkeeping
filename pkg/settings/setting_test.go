package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLMConfig_GetOpenAIBaseURL(t *testing.T) {
	llmConfig := &LLMConfig{}
	assert.Equal(t, "https://api.openai.com/v1/", llmConfig.GetOpenAIBaseURL())

	llmConfig = &LLMConfig{OpenAIBaseURL: "https://api.example.com/v1"}
	assert.Equal(t, "https://api.example.com/v1", llmConfig.GetOpenAIBaseURL())
}

func TestLLMConfig_GetOpenAIEndpointURL(t *testing.T) {
	llmConfig := &LLMConfig{}
	assert.Equal(t, "https://api.openai.com/v1/chat/completions", llmConfig.GetOpenAIEndpointURL("chat/completions"))
	assert.Equal(t, "https://api.openai.com/v1/chat/completions", llmConfig.GetOpenAIEndpointURL("/chat/completions"))

	llmConfig = &LLMConfig{OpenAIBaseURL: "https://api.example.com/v1"}
	assert.Equal(t, "https://api.example.com/v1/chat/completions", llmConfig.GetOpenAIEndpointURL("chat/completions"))

	llmConfig = &LLMConfig{OpenAIBaseURL: "https://api.example.com/v1/"}
	assert.Equal(t, "https://api.example.com/v1/chat/completions", llmConfig.GetOpenAIEndpointURL("chat/completions"))
}

func TestGetConfigItemValueFromEnvironment_StripsTrailingLineEndingsFromFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "openai_api_key")
	err := os.WriteFile(filePath, []byte("test-key\r\n"), 0o600)
	assert.NoError(t, err)

	t.Setenv(getConfigItemFilePathEnvironmentKey("llm_assistant", "openai_api_key"), filePath)
	assert.Equal(t, "test-key", getConfigItemValueFromEnvironment("llm_assistant", "openai_api_key"))
}
