package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
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

func TestLoadLLMConfiguration_ChatCompletionsStream(t *testing.T) {
	configFile, err := ini.Load([]byte("[llm_test]\nllm_provider = openai\nchat_completions_stream = true\n"))
	assert.NoError(t, err)

	llmConfig, err := loadLLMConfiguration(configFile, "llm_test")
	assert.NoError(t, err)
	assert.True(t, llmConfig.ChatCompletionsStream)
}

func TestAIAssistantIndependentEmbeddingConfiguration(t *testing.T) {
	configFile, err := ini.Load([]byte(`[llm_assistant]
llm_provider = openai
openai_base_url = https://chat.example/v1/
openai_api_key = chat-key
openai_model_id = gpt-6-astra
[llm_assistant_embedding]
llm_provider = openai
openai_base_url = https://embedding.example/compatible-mode/v1/
openai_api_key = embedding-key
openai_embedding_model_id = qwen3.7-text-embedding-flash
`))
	assert.NoError(t, err)
	chat, err := loadLLMConfiguration(configFile, "llm_assistant")
	assert.NoError(t, err)
	embedding, err := loadLLMConfiguration(configFile, "llm_assistant_embedding")
	assert.NoError(t, err)
	config := &Config{AIAssistantLLMConfig: chat, AIAssistantEmbeddingLLMConfig: embedding}
	assert.Equal(t, "embedding-key", config.GetAIAssistantEmbeddingLLMConfig().OpenAIAPIKey)
	assert.Equal(t, "https://embedding.example/compatible-mode/v1/embeddings", config.GetAIAssistantEmbeddingLLMConfig().GetOpenAIEndpointURL("embeddings"))
	assert.Equal(t, "chat-key", config.AIAssistantLLMConfig.OpenAIAPIKey)
	config.AIAssistantEmbeddingLLMConfig = nil
	assert.Same(t, chat, config.GetAIAssistantEmbeddingLLMConfig())
}

func TestLoadConfiguration_IndependentEmbeddingProvider(t *testing.T) {
	for _, tc := range []struct {
		name, provider, model string
		wantError             error
	}{
		{name: "legacy"},
		{name: "independent", provider: "openai", model: "qwen3.7-text-embedding-flash"},
		{name: "unsupported provider", provider: "anthropic", model: "test", wantError: errs.ErrInvalidLLMProvider},
		{name: "missing model", provider: "openai", wantError: errs.ErrAIAssistantEmbeddingModelInvalid},
		{name: "blank model", provider: "openai", model: "   ", wantError: errs.ErrAIAssistantEmbeddingModelInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, err := ini.Load("../../conf/ezbookkeeping.ini")
			require.NoError(t, err)
			file.Section("llm_assistant").Key("llm_provider").SetValue("openai")
			section := file.Section("llm_assistant_embedding")
			section.Key("llm_provider").SetValue(tc.provider)
			section.Key("openai_base_url").SetValue("https://embedding.example/compatible-mode/v1/")
			section.Key("openai_embedding_model_id").SetValue(tc.model)
			dir := t.TempDir()
			t.Setenv(ebkWorkDirEnvName, dir)
			file.Section("server").Key("static_root_path").SetValue(dir)
			file.Section("storage").Key("local_filesystem_path").SetValue(dir)
			keyPath := filepath.Join(dir, "embedding-key")
			require.NoError(t, os.WriteFile(keyPath, []byte("test-embedding-key\n"), 0600))
			t.Setenv(getConfigItemFilePathEnvironmentKey("llm_assistant_embedding", "openai_api_key"), keyPath)
			configPath := filepath.Join(dir, "config.ini")
			require.NoError(t, file.SaveTo(configPath))
			config, err := LoadConfiguration(configPath)
			if tc.wantError != nil {
				require.ErrorIs(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			if tc.provider == "" {
				require.Nil(t, config.AIAssistantEmbeddingLLMConfig)
				require.Same(t, config.AIAssistantLLMConfig, config.GetAIAssistantEmbeddingLLMConfig())
			} else {
				embedding := config.GetAIAssistantEmbeddingLLMConfig()
				require.NotSame(t, config.AIAssistantLLMConfig, embedding)
				require.Equal(t, "test-embedding-key", embedding.OpenAIAPIKey)
				require.Equal(t, tc.model, embedding.OpenAIEmbeddingModelID)
			}
		})
	}
}
