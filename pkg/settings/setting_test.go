package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
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

func TestLoadAIConfiguration_AssistantEnablementCompatibility(t *testing.T) {
	cases := []struct {
		name    string
		config  string
		enabled bool
	}{
		{"legacy section", "[llm]\nenable_ai_assistant = true\n", true},
		{"current section", "[ai]\nenable_ai_assistant = true\n", true},
		{"explicit current setting wins", "[llm]\nenable_ai_assistant = true\n[ai]\nenable_ai_assistant = false\n", false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cfgFile, err := ini.Load([]byte(test.config))
			if !assert.NoError(t, err) {
				return
			}
			config := &Config{}
			assert.NoError(t, loadAIConfiguration(config, cfgFile, "ai"))
			assert.Equal(t, test.enabled, config.EnableAIAssistant)
		})
	}
}

func TestLoadAIConfiguration_LegacyRecognitionSettings(t *testing.T) {
	for _, currentSection := range []string{"", "[ai]\ntransaction_from_ai_text_recognition = false\ntransaction_from_ai_image_recognition = false\nmax_ai_recognition_picture_size = 512\n"} {
		file, err := ini.Load([]byte("[llm]\ntransaction_from_ai_text_recognition = true\ntransaction_from_ai_image_recognition = true\nmax_ai_recognition_picture_size = 1024\n" + currentSection))
		require.NoError(t, err)
		config := &Config{}
		require.NoError(t, loadAIConfiguration(config, file, "ai"))
		assert.Equal(t, currentSection == "", config.TransactionFromAITextRecognition)
		assert.Equal(t, currentSection == "", config.TransactionFromAIImageRecognition)
		if currentSection == "" {
			assert.Equal(t, uint32(1024), config.MaxAIRecognitionPictureFileSize)
		} else {
			assert.Equal(t, uint32(512), config.MaxAIRecognitionPictureFileSize)
		}
	}
}

// Retired configuration must not validate or resolve file overrides. Old
// installations can upgrade even when their embedding secrets no longer exist.
func TestLoadConfiguration_IgnoresRetiredEmbeddingConfiguration(t *testing.T) {
	for _, legacy := range []string{"absent", "invalid section", "missing secret files", "invalid file overrides"} {
		t.Run(legacy, func(t *testing.T) {
			file, err := ini.Load("../../conf/ezbookkeeping.ini")
			require.NoError(t, err)
			dir := t.TempDir()
			t.Setenv(ebkWorkDirEnvName, dir)
			file.Section("server").Key("static_root_path").SetValue(dir)
			file.Section("storage").Key("local_filesystem_path").SetValue(dir)
			file.Section("ai").Key("enable_ai_assistant").SetValue("true")
			file.Section("llm_assistant").Key("llm_provider").SetValue("openai")
			file.Section("llm_assistant").Key("openai_api_key").SetValue("active-chat-key")
			file.Section("llm_image_recognition").Key("llm_provider").SetValue("openai")
			file.Section("llm_image_recognition").Key("openai_api_key").SetValue("active-receipt-key")
			if legacy != "absent" {
				section := file.Section("llm_assistant_embedding")
				for _, key := range []string{"llm_provider", "enable_thinking", "request_timeout", "proxy", "skip_tls_verify", "openai_embedding_model_id"} {
					section.Key(key).SetValue("invalid-retired-value")
				}
				file.Section("llm_assistant").Key("openai_embedding_model_id").SetValue("invalid-retired-value")
			}
			if legacy == "missing secret files" || legacy == "invalid file overrides" {
				path := filepath.Join(dir, "missing-embedding-secret")
				if legacy == "invalid file overrides" {
					// Reading this provider/thinking override would reject startup.
					require.NoError(t, os.WriteFile(path, []byte("invalid-retired-file-canary"), 0600))
				}
				for _, key := range []string{"llm_provider", "openai_api_key", "enable_thinking", "openai_embedding_model_id"} {
					t.Setenv(getConfigItemFilePathEnvironmentKey("llm_assistant_embedding", key), path)
					t.Setenv(getConfigItemValueEnvironmentKey("llm_assistant_embedding", key), "invalid-retired-env-canary")
				}
				t.Setenv(getConfigItemFilePathEnvironmentKey("llm_assistant", "openai_embedding_model_id"), path)
			}
			path := filepath.Join(dir, "config.ini")
			require.NoError(t, file.SaveTo(path))
			config, err := LoadConfiguration(path)
			require.NoError(t, err)
			require.True(t, config.EnableAIAssistant)
			require.Equal(t, "active-chat-key", config.AIAssistantLLMConfig.OpenAIAPIKey)
			require.Equal(t, "active-receipt-key", config.ReceiptImageRecognitionLLMConfig.OpenAIAPIKey)
			require.Equal(t, AIAssistantRuntimeConfig{RunTimeout: 180000, QueryTimeout: 30000, MaxRounds: 12}, config.AIAssistantRuntimeConfig)
		})
	}
}
