package settings

import (
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
)

// LoadConfiguration loads setting config from given config file path
func LoadConfiguration(configFilePath string) (*Config, error) {
	var err error

	cfgFile, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, configFilePath)

	if err != nil {
		return nil, err
	}

	config := &Config{}
	config.WorkingPath, err = getWorkingPath()

	if err != nil {
		return nil, err
	}

	err = loadGlobalConfiguration(config, cfgFile, "global")

	if err != nil {
		return nil, err
	}

	err = loadServerConfiguration(config, cfgFile, "server")

	if err != nil {
		return nil, err
	}

	err = loadMCPServerConfiguration(config, cfgFile, "mcp")

	if err != nil {
		return nil, err
	}

	err = loadDatabaseConfiguration(config, cfgFile, "database")

	if err != nil {
		return nil, err
	}

	err = loadMailConfiguration(config, cfgFile, "mail")

	if err != nil {
		return nil, err
	}

	err = loadLogConfiguration(config, cfgFile, "log")

	if err != nil {
		return nil, err
	}

	err = loadStorageConfiguration(config, cfgFile, "storage")

	if err != nil {
		return nil, err
	}

	err = loadAIConfiguration(config, cfgFile, "ai")

	if err != nil {
		return nil, err
	}

	config.TextRecognitionLLMConfig, err = loadLLMConfiguration(cfgFile, "llm_text_recognition")

	if err != nil {
		return nil, err
	}

	config.ReceiptImageRecognitionLLMConfig, err = loadLLMConfiguration(cfgFile, "llm_image_recognition")

	if err != nil {
		return nil, err
	}

	config.CodingAssistantLLMConfig, err = loadLLMConfiguration(cfgFile, "llm_coding_assistant")

	if err != nil {
		return nil, err
	}

	config.ReceiptImageRecognitionFallbackLLMConfig, err = loadLLMConfiguration(cfgFile, "llm_image_recognition_fallback")
	if err != nil {
		return nil, err
	}

	config.AIAssistantLLMConfig, err = loadLLMConfiguration(cfgFile, "llm_assistant")

	if err != nil {
		return nil, err
	}

	config.AIAssistantRuntimeConfig, err = loadAIAssistantRuntimeConfiguration(cfgFile)
	if err != nil {
		return nil, err
	}

	// Legacy llm_assistant_embedding settings are deliberately ignored, including
	// environment file overrides. The assistant now queries its read-only tools.
	config.AIAssistantFallbackLLMConfig, err = loadLLMConfiguration(cfgFile, "llm_assistant_fallback")
	if err != nil {
		return nil, err
	}

	for _, assistantLLMConfig := range []*LLMConfig{config.AIAssistantLLMConfig, config.AIAssistantFallbackLLMConfig} {
		if assistantLLMConfig == nil || assistantLLMConfig.LLMProvider != OpenAILLMProvider {
			continue
		}
		if assistantLLMConfig.OpenAIModelID == "" {
			assistantLLMConfig.OpenAIModelID = defaultAIAssistantOpenAIModelID
		}
	}

	err = loadUuidConfiguration(config, cfgFile, "uuid")

	if err != nil {
		return nil, err
	}

	err = loadDuplicateCheckerConfiguration(config, cfgFile, "duplicate_checker")

	if err != nil {
		return nil, err
	}

	err = loadCronConfiguration(config, cfgFile, "cron")

	if err != nil {
		return nil, err
	}

	err = loadSecurityConfiguration(config, cfgFile, "security")

	if err != nil {
		return nil, err
	}

	err = loadAuthConfiguration(config, cfgFile, "auth")

	if err != nil {
		return nil, err
	}

	err = loadUserConfiguration(config, cfgFile, "user")

	if err != nil {
		return nil, err
	}

	err = loadDataConfiguration(config, cfgFile, "data")

	if err != nil {
		return nil, err
	}

	err = loadTipConfiguration(config, cfgFile, "tip")

	if err != nil {
		return nil, err
	}

	err = loadNotificationConfiguration(config, cfgFile, "notification")

	if err != nil {
		return nil, err
	}

	err = loadMapConfiguration(config, cfgFile, "map")

	if err != nil {
		return nil, err
	}

	err = loadExchangeRatesConfiguration(config, cfgFile, "exchange_rates")

	if err != nil {
		return nil, err
	}

	err = loadDeveloperTools(config, cfgFile, "developer")

	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetDefaultConfigFilePath returns the defaule config file path
func GetDefaultConfigFilePath() (string, error) {
	workingPath, err := getWorkingPath()

	if err != nil {
		return "", err
	}

	cfgFilePath := filepath.Join(workingPath, defaultConfigPath)
	_, err = os.Stat(cfgFilePath)

	if err != nil {
		return "", err
	}

	return cfgFilePath, nil
}
