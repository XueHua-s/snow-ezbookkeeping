package settings

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
	"testing"
)

func TestAssistantExperienceDefaults(t *testing.T) {
	file, err := ini.Load([]byte(""))
	require.NoError(t, err)
	config, err := loadAIAssistantExperienceConfiguration(file)
	require.NoError(t, err)
	require.Equal(t, DefaultAIAssistantExperienceConfig(), config)
	require.Equal(t, 4000, config.Limits.InputCharacters)
	require.Equal(t, 104857600, config.Limits.UserBytes)
	require.Equal(t, config, (&Config{}).GetAIAssistantExperienceConfig())
}
func TestAssistantExperienceValidation(t *testing.T) {
	for _, setting := range []string{"input_bytes=0", "cleanup_hours=25", "frame_bytes=262145", "page_size=101", "receipt_retention_hours=47", "command_max_age_hours=25", "lease_seconds=5", "primary_context_tokens=4096", "input_bytes=no", "summary_tokens=16000"} {
		t.Run(setting, func(t *testing.T) {
			file, err := ini.Load([]byte("[assistant_experience]\n" + setting))
			require.NoError(t, err)
			_, err = loadAIAssistantExperienceConfiguration(file)
			require.Error(t, err)
		})
	}
}

func TestAssistantExperienceRequestBudgets(t *testing.T) {
	file, err := ini.Load([]byte(""))
	require.NoError(t, err)
	config, err := loadAIAssistantExperienceConfiguration(file)
	require.NoError(t, err)
	require.Equal(t, 30, config.Limits.GenerationRequestsPerMinute)
	require.Equal(t, 12, config.Limits.SearchExportRequestsPerMinute)
	require.Equal(t, 240, config.Limits.StatusRequestsPerMinute)
	require.Equal(t, 2, config.Limits.SearchExportConcurrency)
	require.Equal(t, 5, config.Limits.SearchTimeoutSeconds)
	require.Equal(t, 60, config.Limits.ExportTimeoutSeconds)
	require.Equal(t, 4096, config.Limits.RateLimitUsers)
	for _, key := range []string{"generation_requests_per_minute", "search_export_requests_per_minute", "status_requests_per_minute", "search_export_concurrency", "search_timeout_seconds", "export_timeout_seconds", "rate_limit_users"} {
		file, err = ini.Load([]byte("[assistant_experience]\n" + key + " = 0"))
		require.NoError(t, err)
		_, err = loadAIAssistantExperienceConfiguration(file)
		require.Error(t, err)
	}
	file, err = ini.Load([]byte("[assistant_experience]\ngeneration_requests_per_minute=15\nexport_timeout_seconds=120"))
	require.NoError(t, err)
	config, err = loadAIAssistantExperienceConfiguration(file)
	require.NoError(t, err)
	require.Equal(t, 15, config.Limits.GenerationRequestsPerMinute)
	require.Equal(t, 120, config.Limits.ExportTimeoutSeconds)
	config.Limits.SearchExportConcurrency = 0
	require.Error(t, config.Validate())
}
