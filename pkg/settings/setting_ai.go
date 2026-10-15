package settings

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"gopkg.in/ini.v1"
	"os"
	"strconv"
	"strings"
)

func loadAIConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	config.TransactionFromAITextRecognition = getConfigItemBoolValue(configFile, sectionName, "transaction_from_ai_text_recognition", getConfigItemBoolValue(configFile, "llm", "transaction_from_ai_text_recognition", false))
	config.TransactionFromAIImageRecognition = getConfigItemBoolValue(configFile, sectionName, "transaction_from_ai_image_recognition", getConfigItemBoolValue(configFile, "llm", "transaction_from_ai_image_recognition", false))
	config.InsightsExplorerCodingAssistant = getConfigItemBoolValue(configFile, sectionName, "insights_explorer_coding_assistant", false)
	config.MaxAIRecognitionPictureFileSize = getConfigItemUint32Value(configFile, sectionName, "max_ai_recognition_picture_size", getConfigItemUint32Value(configFile, "llm", "max_ai_recognition_picture_size", defaultAIRecognitionPictureMaxSize))
	config.EnableAIAssistant = getConfigItemBoolValue(configFile, sectionName, "enable_ai_assistant", getConfigItemBoolValue(configFile, "llm", "enable_ai_assistant", false))

	return nil
}

func loadAIAssistantRuntimeConfiguration(configFile *ini.File) (AIAssistantRuntimeConfig, error) {
	config := AIAssistantRuntimeConfig{}
	for _, item := range []struct {
		name         string
		defaultValue uint32
		target       *uint32
	}{
		{"run_timeout", defaultAIAssistantRunTimeout, &config.RunTimeout},
		{"query_timeout", defaultAIAssistantQueryTimeout, &config.QueryTimeout},
		{"max_rounds", defaultAIAssistantMaxRounds, &config.MaxRounds},
	} {
		value, err := getPositiveConfigItemUint32Value(configFile, "llm_assistant", item.name, item.defaultValue)
		if err != nil {
			return AIAssistantRuntimeConfig{}, err
		}
		*item.target = value
	}
	return config, nil
}

// getPositiveConfigItemUint32Value rejects invalid configured budgets instead of
// falling back silently. File overrides take precedence over direct env and INI.
func getPositiveConfigItemUint32Value(configFile *ini.File, sectionName string, itemName string, defaultValue uint32) (uint32, error) {
	var rawValue string
	filePath := os.Getenv(getConfigItemFilePathEnvironmentKey(sectionName, itemName))
	if filePath != "" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return 0, fmt.Errorf("invalid [%s] %s: cannot read environment override file", sectionName, itemName)
		}
		rawValue = strings.TrimRight(string(content), "\r\n")
	} else if value, exists := os.LookupEnv(getConfigItemValueEnvironmentKey(sectionName, itemName)); exists {
		rawValue = value
	} else if section := configFile.Section(sectionName); section.HasKey(itemName) {
		rawValue = section.Key(itemName).Value()
	} else {
		return defaultValue, nil
	}

	value, err := strconv.ParseUint(rawValue, 10, 32)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("invalid [%s] %s: must be an integer between 1 and 4294967295", sectionName, itemName)
	}
	return uint32(value), nil
}

func loadLLMConfiguration(configFile *ini.File, sectionName string) (*LLMConfig, error) {
	llmConfig := &LLMConfig{}
	llmProvider := getConfigItemStringValue(configFile, sectionName, "llm_provider")

	if llmProvider == "" {
		llmConfig.LLMProvider = ""
	} else if llmProvider == OpenAILLMProvider {
		llmConfig.LLMProvider = OpenAILLMProvider
	} else if llmProvider == OpenAICompatibleLLMProvider {
		llmConfig.LLMProvider = OpenAICompatibleLLMProvider
	} else if llmProvider == OpenAIResponsesCompatibleLLMProvider {
		llmConfig.LLMProvider = OpenAIResponsesCompatibleLLMProvider
	} else if llmProvider == AnthropicLLMProvider {
		llmConfig.LLMProvider = AnthropicLLMProvider
	} else if llmProvider == AnthropicCompatibleLLMProvider {
		llmConfig.LLMProvider = AnthropicCompatibleLLMProvider
	} else if llmProvider == OpenRouterLLMProvider {
		llmConfig.LLMProvider = OpenRouterLLMProvider
	} else if llmProvider == OllamaLLMProvider {
		llmConfig.LLMProvider = OllamaLLMProvider
	} else if llmProvider == LMStudioLLMProvider {
		llmConfig.LLMProvider = LMStudioLLMProvider
	} else if llmProvider == GoogleAILLMProvider {
		llmConfig.LLMProvider = GoogleAILLMProvider
	} else {
		return nil, errs.ErrInvalidLLMProvider
	}

	thinkingLevel := getConfigItemStringValue(configFile, sectionName, "enable_thinking")

	if thinkingLevel == string(LLMThinkingDefault) {
		llmConfig.EnableThinking = LLMThinkingDefault
	} else if thinkingLevel == string(LLMThinkingDisabled) {
		llmConfig.EnableThinking = LLMThinkingDisabled
	} else if thinkingLevel == string(LLMThinkingEnabled) {
		llmConfig.EnableThinking = LLMThinkingEnabled
	} else if thinkingLevel == string(LLMThinkingLow) {
		llmConfig.EnableThinking = LLMThinkingLow
	} else if thinkingLevel == string(LLMThinkingMedium) {
		llmConfig.EnableThinking = LLMThinkingMedium
	} else if thinkingLevel == string(LLMThinkingHigh) {
		llmConfig.EnableThinking = LLMThinkingHigh
	} else if thinkingLevel == string(LLMThinkingXHigh) {
		llmConfig.EnableThinking = LLMThinkingXHigh
	} else {
		return nil, errs.ErrInvalidLLMThinkingLevel
	}

	llmConfig.OpenAIAPIKey = getConfigItemStringValue(configFile, sectionName, "openai_api_key")
	llmConfig.OpenAIBaseURL = getConfigItemStringValue(configFile, sectionName, "openai_base_url")
	llmConfig.OpenAIModelID = getConfigItemStringValue(configFile, sectionName, "openai_model_id")
	llmConfig.ChatCompletionsStream = getConfigItemBoolValue(configFile, sectionName, "chat_completions_stream", false)

	llmConfig.OpenAICompatibleBaseURL = getConfigItemStringValue(configFile, sectionName, "openai_compatible_base_url")
	llmConfig.OpenAICompatibleAPIKey = getConfigItemStringValue(configFile, sectionName, "openai_compatible_api_key")
	llmConfig.OpenAICompatibleModelID = getConfigItemStringValue(configFile, sectionName, "openai_compatible_model_id")

	llmConfig.AnthropicAPIKey = getConfigItemStringValue(configFile, sectionName, "anthropic_api_key")
	llmConfig.AnthropicModelID = getConfigItemStringValue(configFile, sectionName, "anthropic_model_id")
	llmConfig.AnthropicMaxTokens = getConfigItemUint32Value(configFile, sectionName, "anthropic_max_tokens", defaultAnthropicLargeLanguageModelAPIMaximumTokens)
	llmConfig.AnthropicThinkingBudgetTokens = getConfigItemUint32Value(configFile, sectionName, "anthropic_thinking_budget_tokens", defaultAnthropicLargeLanguageModelThinkingBudget)

	llmConfig.AnthropicCompatibleBaseURL = getConfigItemStringValue(configFile, sectionName, "anthropic_compatible_base_url")
	llmConfig.AnthropicCompatibleAPIVersion = getConfigItemStringValue(configFile, sectionName, "anthropic_compatible_api_version")
	llmConfig.AnthropicCompatibleAPIKey = getConfigItemStringValue(configFile, sectionName, "anthropic_compatible_api_key")
	llmConfig.AnthropicCompatibleModelID = getConfigItemStringValue(configFile, sectionName, "anthropic_compatible_model_id")
	llmConfig.AnthropicCompatibleMaxTokens = getConfigItemUint32Value(configFile, sectionName, "anthropic_compatible_max_tokens", defaultAnthropicLargeLanguageModelAPIMaximumTokens)
	llmConfig.AnthropicCompatibleThinkingBudgetTokens = getConfigItemUint32Value(configFile, sectionName, "anthropic_compatible_thinking_budget_tokens", defaultAnthropicLargeLanguageModelThinkingBudget)

	llmConfig.OpenRouterAPIKey = getConfigItemStringValue(configFile, sectionName, "openrouter_api_key")
	llmConfig.OpenRouterModelID = getConfigItemStringValue(configFile, sectionName, "openrouter_model_id")

	llmConfig.OllamaServerURL = getConfigItemStringValue(configFile, sectionName, "ollama_server_url")
	llmConfig.OllamaModelID = getConfigItemStringValue(configFile, sectionName, "ollama_model_id")

	llmConfig.LMStudioServerURL = getConfigItemStringValue(configFile, sectionName, "lm_studio_server_url")
	llmConfig.LMStudioToken = getConfigItemStringValue(configFile, sectionName, "lm_studio_token")
	llmConfig.LMStudioModelID = getConfigItemStringValue(configFile, sectionName, "lm_studio_model_id")

	llmConfig.GoogleAIAPIKey = getConfigItemStringValue(configFile, sectionName, "google_ai_api_key")
	llmConfig.GoogleAIModelID = getConfigItemStringValue(configFile, sectionName, "google_ai_model_id")

	llmConfig.LargeLanguageModelAPIProxy = getConfigItemStringValue(configFile, sectionName, "proxy", "system")
	llmConfig.LargeLanguageModelAPIRequestTimeout = getConfigItemUint32Value(configFile, sectionName, "request_timeout", defaultLargeLanguageModelAPIRequestTimeout)
	llmConfig.LargeLanguageModelAPISkipTLSVerify = getConfigItemBoolValue(configFile, sectionName, "skip_tls_verify", false)

	return llmConfig, nil
}
