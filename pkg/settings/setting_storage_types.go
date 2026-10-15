package settings

import (
	"strings"
)

// DatabaseConfig represents the database setting config
type DatabaseConfig struct {
	DatabaseType     string
	DatabaseHost     string
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string

	DatabaseSSLMode string

	DatabasePath string

	MaxIdleConnection     uint16
	MaxOpenConnection     uint16
	ConnectionMaxLifeTime uint32
}

// SMTPConfig represents the SMTP setting config
type SMTPConfig struct {
	SMTPHost          string
	SMTPUser          string
	SMTPPasswd        string
	SMTPSkipTLSVerify bool
	FromAddress       string
}

// S3Config represents the S3-compatible object storage setting config
type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	UseSSL          bool
	SkipTLSVerify   bool
	UsePathStyle    bool
	Bucket          string
	RootPath        string
}

// MinIOConfig represents the MinIO setting config
type MinIOConfig struct {
	Endpoint        string
	Location        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	SkipTLSVerify   bool
	Bucket          string
	RootPath        string
}

// WebDAVConfig represents the WebDAV setting config
type WebDAVConfig struct {
	Url            string
	Username       string
	Password       string
	RootPath       string
	RequestTimeout uint32
	Proxy          string
	SkipTLSVerify  bool
}

// AIAssistantRuntimeConfig contains execution budgets shared by all assistant models.
// Timeouts are milliseconds and limit execution, never the query's ledger scope.
type AIAssistantRuntimeConfig struct {
	RunTimeout   uint32
	QueryTimeout uint32
	MaxRounds    uint32
}

// GetAIAssistantRuntimeConfig supplies defaults for programmatically constructed
// configurations. LoadConfiguration rejects explicitly configured zero values.
func (config *Config) GetAIAssistantRuntimeConfig() AIAssistantRuntimeConfig {
	runtime := config.AIAssistantRuntimeConfig
	if runtime.RunTimeout == 0 {
		runtime.RunTimeout = defaultAIAssistantRunTimeout
	}
	if runtime.QueryTimeout == 0 {
		runtime.QueryTimeout = defaultAIAssistantQueryTimeout
	}
	if runtime.MaxRounds == 0 {
		runtime.MaxRounds = defaultAIAssistantMaxRounds
	}
	return runtime
}

// LLMConfig represents the Large Language Model setting config
type LLMConfig struct {
	LLMProvider                             string
	OpenAIBaseURL                           string
	EnableThinking                          LLMThinkingLevel
	OpenAIAPIKey                            string
	OpenAIModelID                           string
	ChatCompletionsStream                   bool
	OpenAICompatibleBaseURL                 string
	OpenAICompatibleAPIKey                  string
	OpenAICompatibleModelID                 string
	AnthropicAPIKey                         string
	AnthropicModelID                        string
	AnthropicMaxTokens                      uint32
	AnthropicThinkingBudgetTokens           uint32
	AnthropicCompatibleBaseURL              string
	AnthropicCompatibleAPIVersion           string
	AnthropicCompatibleAPIKey               string
	AnthropicCompatibleModelID              string
	AnthropicCompatibleMaxTokens            uint32
	AnthropicCompatibleThinkingBudgetTokens uint32
	OpenRouterAPIKey                        string
	OpenRouterModelID                       string
	OllamaServerURL                         string
	OllamaModelID                           string
	LMStudioServerURL                       string
	LMStudioToken                           string
	LMStudioModelID                         string
	GoogleAIAPIKey                          string
	GoogleAIModelID                         string
	LargeLanguageModelAPIRequestTimeout     uint32
	LargeLanguageModelAPIProxy              string
	LargeLanguageModelAPISkipTLSVerify      bool
}

// GetOpenAIBaseURL returns the final OpenAI API base url
func (config *LLMConfig) GetOpenAIBaseURL() string {
	baseURL := strings.TrimSpace(config.OpenAIBaseURL)
	if baseURL == "" {
		return defaultOpenAIBaseURL
	}
	return baseURL
}

// GetOpenAIEndpointURL returns the final OpenAI endpoint url with the given path
func (config *LLMConfig) GetOpenAIEndpointURL(path string) string {
	baseURL := config.GetOpenAIBaseURL()
	finalPath := strings.TrimPrefix(strings.TrimSpace(path), "/")
	if finalPath == "" {
		return baseURL
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return baseURL + finalPath
}
