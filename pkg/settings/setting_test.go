package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
