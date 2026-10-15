package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestServerSettingsAssistantRuntimeProjection(t *testing.T) {
	previous := settings.Container.GetCurrentConfig()
	t.Cleanup(func() { settings.SetCurrentConfig(previous) })
	for _, timeout := range []uint32{0, 180000, 240000, 4294967295} {
		t.Run(fmt.Sprint(timeout), func(t *testing.T) {
			settings.SetCurrentConfig(&settings.Config{
				EnableAIAssistant: true,
				AIAssistantRuntimeConfig: settings.AIAssistantRuntimeConfig{
					RunTimeout: timeout, QueryTimeout: 30001, MaxRounds: 13,
				},
				AIAssistantLLMConfig: &settings.LLMConfig{
					LLMProvider: settings.OpenAILLMProvider, OpenAIModelID: "chat-model",
					OpenAIAPIKey: "primary-secret-canary", OpenAIBaseURL: "https://primary-private-canary.invalid/",
				},
				AIAssistantFallbackLLMConfig: &settings.LLMConfig{
					OpenAIAPIKey: "fallback-secret-canary", OpenAIBaseURL: "https://fallback-private-canary.invalid/",
				},
			})
			body, _, err := ServerSettings.ServerSettingsJavascriptHandler(nil)
			require.Nil(t, err)
			javascript := string(body)
			require.Contains(t, javascript, "_['llma']=1;")
			require.Contains(t, javascript, "_['llmam']='chat-model';")
			wantTimeout := timeout
			if wantTimeout == 0 {
				wantTimeout = 180000
			}
			require.Contains(t, javascript, fmt.Sprintf("_['llmart']=%d;", wantTimeout))
			for _, private := range []string{"canary", "query_timeout", "max_rounds", "30001", "OpenAIAPIKey", "openai_api_key"} {
				require.NotContains(t, javascript, private)
			}
		})
	}
}

func TestServerSettingsDisabledAssistantRuntime(t *testing.T) {
	previous := settings.Container.GetCurrentConfig()
	t.Cleanup(func() { settings.SetCurrentConfig(previous) })
	for _, config := range []*settings.Config{
		{EnableAIAssistant: false, AIAssistantLLMConfig: &settings.LLMConfig{LLMProvider: settings.OpenAILLMProvider}},
		{EnableAIAssistant: true},
		{EnableAIAssistant: true, AIAssistantLLMConfig: &settings.LLMConfig{}},
	} {
		settings.SetCurrentConfig(config)
		body, _, err := ServerSettings.ServerSettingsJavascriptHandler(nil)
		require.Nil(t, err)
		require.NotContains(t, string(body), "llmart")
	}
}
