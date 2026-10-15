package settings

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"gopkg.in/ini.v1"
	"reflect"
)

// AIAssistantExperienceConfig is the sole authority for persistent assistant limits.
// Durations include their units; existing runner timeout settings remain authoritative.
type AIAssistantExperienceConfig struct {
	Limits                          core.AssistantLimits
	QueueTimeoutSeconds             uint32
	LeaseHeartbeatSeconds           uint32
	LeaseSeconds                    uint32
	CheckpointMilliseconds          uint32
	StreamHeartbeatSeconds          uint32
	CommandMaxAgeHours              uint32
	CommandFutureSkewSeconds        uint32
	ReceiptRetentionHours           uint32
	CleanupHours                    uint32
	MemoryCandidateDays             uint32
	MemoryIdleSeconds               uint32
	MemoryCallsPerDay               uint32
	SummaryTimeoutSeconds           uint32
	PrimaryContextTokens            uint32
	FallbackContextTokens           uint32
	ProtocolTokens                  uint32
	ScanBatchSize                   uint32
	ExportBatchSize                 uint32
	SubscriptionWriteTimeoutSeconds uint32
}

func DefaultAIAssistantExperienceConfig() AIAssistantExperienceConfig {
	return AIAssistantExperienceConfig{Limits: core.AssistantLimits{
		GenerationRequestsPerMinute:   30,
		SearchExportRequestsPerMinute: 12,
		StatusRequestsPerMinute:       240,
		SearchExportConcurrency:       2,
		SearchTimeoutSeconds:          5,
		ExportTimeoutSeconds:          60,
		RateLimitUsers:                4096,

		ProcessSummaryBytes:     4096,
		InputCharacters:         4000,
		InputBytes:              16384,
		TitleCharacters:         100,
		ResponseBytes:           32768,
		AttemptBytes:            65536,
		ConversationBytes:       2097152,
		ConversationTurns:       200,
		ConversationHeaderBytes: 32768,
		UserBytes:               104857600,
		Conversations:           1000,
		Memories:                50,
		MemoryBytes:             1024,
		MemoryCandidates:        50,
		RunningPerUser:          2,
		QueuedPerUser:           2,
		WorkerConcurrency:       4,
		PageSize:                30,
		PageSizeMax:             100,
		VersionsPageSize:        20,
		RecentVersions:          5,
		PageBytes:               262144,
		FrameBytes:              262144,
		SubscriptionsPerUser:    8,
		SubscriptionsPerRun:     4,
		InputTokens:             16000,
		SummaryTokens:           2000,
		MemoryTokens:            1000,
		OutputTokens:            4096,
		RequestBytes:            131072,
	},
		QueueTimeoutSeconds:             30,
		LeaseHeartbeatSeconds:           5,
		LeaseSeconds:                    20,
		CheckpointMilliseconds:          500,
		StreamHeartbeatSeconds:          15,
		CommandMaxAgeHours:              24,
		CommandFutureSkewSeconds:        300,
		ReceiptRetentionHours:           48,
		CleanupHours:                    24,
		MemoryCandidateDays:             7,
		MemoryIdleSeconds:               600,
		MemoryCallsPerDay:               20,
		SummaryTimeoutSeconds:           30,
		PrimaryContextTokens:            32768,
		FallbackContextTokens:           32768,
		ProtocolTokens:                  1024,
		ScanBatchSize:                   100,
		ExportBatchSize:                 30,
		SubscriptionWriteTimeoutSeconds: 10,
	}
}
func loadAIAssistantExperienceConfiguration(file *ini.File) (AIAssistantExperienceConfig, error) {
	c := DefaultAIAssistantExperienceConfig()
	var err error
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "generation_requests_per_minute", uint32(c.Limits.GenerationRequestsPerMinute)); e != nil {
		return c, e
	} else {
		c.Limits.GenerationRequestsPerMinute = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "search_export_requests_per_minute", uint32(c.Limits.SearchExportRequestsPerMinute)); e != nil {
		return c, e
	} else {
		c.Limits.SearchExportRequestsPerMinute = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "status_requests_per_minute", uint32(c.Limits.StatusRequestsPerMinute)); e != nil {
		return c, e
	} else {
		c.Limits.StatusRequestsPerMinute = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "search_export_concurrency", uint32(c.Limits.SearchExportConcurrency)); e != nil {
		return c, e
	} else {
		c.Limits.SearchExportConcurrency = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "search_timeout_seconds", uint32(c.Limits.SearchTimeoutSeconds)); e != nil {
		return c, e
	} else {
		c.Limits.SearchTimeoutSeconds = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "export_timeout_seconds", uint32(c.Limits.ExportTimeoutSeconds)); e != nil {
		return c, e
	} else {
		c.Limits.ExportTimeoutSeconds = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "rate_limit_users", uint32(c.Limits.RateLimitUsers)); e != nil {
		return c, e
	} else {
		c.Limits.RateLimitUsers = int(value)
	}

	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "process_summary_bytes", uint32(c.Limits.ProcessSummaryBytes)); e != nil {
		return c, e
	} else {
		c.Limits.ProcessSummaryBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "input_characters", uint32(c.Limits.InputCharacters)); e != nil {
		return c, e
	} else {
		c.Limits.InputCharacters = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "input_bytes", uint32(c.Limits.InputBytes)); e != nil {
		return c, e
	} else {
		c.Limits.InputBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "title_characters", uint32(c.Limits.TitleCharacters)); e != nil {
		return c, e
	} else {
		c.Limits.TitleCharacters = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "response_bytes", uint32(c.Limits.ResponseBytes)); e != nil {
		return c, e
	} else {
		c.Limits.ResponseBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "attempt_bytes", uint32(c.Limits.AttemptBytes)); e != nil {
		return c, e
	} else {
		c.Limits.AttemptBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "conversation_bytes", uint32(c.Limits.ConversationBytes)); e != nil {
		return c, e
	} else {
		c.Limits.ConversationBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "conversation_turns", uint32(c.Limits.ConversationTurns)); e != nil {
		return c, e
	} else {
		c.Limits.ConversationTurns = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "conversation_header_bytes", uint32(c.Limits.ConversationHeaderBytes)); e != nil {
		return c, e
	} else {
		c.Limits.ConversationHeaderBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "user_bytes", uint32(c.Limits.UserBytes)); e != nil {
		return c, e
	} else {
		c.Limits.UserBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "conversations", uint32(c.Limits.Conversations)); e != nil {
		return c, e
	} else {
		c.Limits.Conversations = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "memories", uint32(c.Limits.Memories)); e != nil {
		return c, e
	} else {
		c.Limits.Memories = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "memory_bytes", uint32(c.Limits.MemoryBytes)); e != nil {
		return c, e
	} else {
		c.Limits.MemoryBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "memory_candidates", uint32(c.Limits.MemoryCandidates)); e != nil {
		return c, e
	} else {
		c.Limits.MemoryCandidates = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "running_per_user", uint32(c.Limits.RunningPerUser)); e != nil {
		return c, e
	} else {
		c.Limits.RunningPerUser = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "queued_per_user", uint32(c.Limits.QueuedPerUser)); e != nil {
		return c, e
	} else {
		c.Limits.QueuedPerUser = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "worker_concurrency", uint32(c.Limits.WorkerConcurrency)); e != nil {
		return c, e
	} else {
		c.Limits.WorkerConcurrency = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "page_size", uint32(c.Limits.PageSize)); e != nil {
		return c, e
	} else {
		c.Limits.PageSize = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "page_size_max", uint32(c.Limits.PageSizeMax)); e != nil {
		return c, e
	} else {
		c.Limits.PageSizeMax = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "versions_page_size", uint32(c.Limits.VersionsPageSize)); e != nil {
		return c, e
	} else {
		c.Limits.VersionsPageSize = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "recent_versions", uint32(c.Limits.RecentVersions)); e != nil {
		return c, e
	} else {
		c.Limits.RecentVersions = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "page_bytes", uint32(c.Limits.PageBytes)); e != nil {
		return c, e
	} else {
		c.Limits.PageBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "frame_bytes", uint32(c.Limits.FrameBytes)); e != nil {
		return c, e
	} else {
		c.Limits.FrameBytes = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "subscriptions_per_user", uint32(c.Limits.SubscriptionsPerUser)); e != nil {
		return c, e
	} else {
		c.Limits.SubscriptionsPerUser = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "subscriptions_per_run", uint32(c.Limits.SubscriptionsPerRun)); e != nil {
		return c, e
	} else {
		c.Limits.SubscriptionsPerRun = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "input_tokens", uint32(c.Limits.InputTokens)); e != nil {
		return c, e
	} else {
		c.Limits.InputTokens = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "summary_tokens", uint32(c.Limits.SummaryTokens)); e != nil {
		return c, e
	} else {
		c.Limits.SummaryTokens = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "memory_tokens", uint32(c.Limits.MemoryTokens)); e != nil {
		return c, e
	} else {
		c.Limits.MemoryTokens = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "output_tokens", uint32(c.Limits.OutputTokens)); e != nil {
		return c, e
	} else {
		c.Limits.OutputTokens = int(value)
	}
	if value, e := getPositiveConfigItemUint32Value(file, "assistant_experience", "request_bytes", uint32(c.Limits.RequestBytes)); e != nil {
		return c, e
	} else {
		c.Limits.RequestBytes = int(value)
	}
	c.QueueTimeoutSeconds, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "queue_timeout_seconds", c.QueueTimeoutSeconds)
	if err != nil {
		return c, err
	}
	c.LeaseHeartbeatSeconds, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "lease_heartbeat_seconds", c.LeaseHeartbeatSeconds)
	if err != nil {
		return c, err
	}
	c.LeaseSeconds, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "lease_seconds", c.LeaseSeconds)
	if err != nil {
		return c, err
	}
	c.CheckpointMilliseconds, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "checkpoint_milliseconds", c.CheckpointMilliseconds)
	if err != nil {
		return c, err
	}
	c.StreamHeartbeatSeconds, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "stream_heartbeat_seconds", c.StreamHeartbeatSeconds)
	if err != nil {
		return c, err
	}
	c.CommandMaxAgeHours, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "command_max_age_hours", c.CommandMaxAgeHours)
	if err != nil {
		return c, err
	}
	c.CommandFutureSkewSeconds, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "command_future_skew_seconds", c.CommandFutureSkewSeconds)
	if err != nil {
		return c, err
	}
	c.ReceiptRetentionHours, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "receipt_retention_hours", c.ReceiptRetentionHours)
	if err != nil {
		return c, err
	}
	c.CleanupHours, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "cleanup_hours", c.CleanupHours)
	if err != nil {
		return c, err
	}
	c.MemoryCandidateDays, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "memory_candidate_days", c.MemoryCandidateDays)
	if err != nil {
		return c, err
	}
	c.MemoryIdleSeconds, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "memory_idle_seconds", c.MemoryIdleSeconds)
	if err != nil {
		return c, err
	}
	c.MemoryCallsPerDay, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "memory_calls_per_day", c.MemoryCallsPerDay)
	if err != nil {
		return c, err
	}
	c.SummaryTimeoutSeconds, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "summary_timeout_seconds", c.SummaryTimeoutSeconds)
	if err != nil {
		return c, err
	}
	c.PrimaryContextTokens, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "primary_context_tokens", c.PrimaryContextTokens)
	if err != nil {
		return c, err
	}
	c.FallbackContextTokens, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "fallback_context_tokens", c.FallbackContextTokens)
	if err != nil {
		return c, err
	}
	c.ProtocolTokens, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "protocol_tokens", c.ProtocolTokens)
	if err != nil {
		return c, err
	}
	c.ScanBatchSize, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "scan_batch_size", c.ScanBatchSize)
	if err != nil {
		return c, err
	}
	c.ExportBatchSize, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "export_batch_size", c.ExportBatchSize)
	if err != nil {
		return c, err
	}
	c.SubscriptionWriteTimeoutSeconds, err = getPositiveConfigItemUint32Value(file, "assistant_experience", "subscription_write_timeout_seconds", c.SubscriptionWriteTimeoutSeconds)
	if err != nil {
		return c, err
	}
	return c, c.Validate()
}
func (c AIAssistantExperienceConfig) Validate() error {
	l := c.Limits
	lv := reflect.ValueOf(l)
	for i := 0; i < lv.NumField(); i++ {
		if lv.Field(i).Int() <= 0 {
			return fmt.Errorf("assistant limits must be positive")
		}
	}
	cv := reflect.ValueOf(c)
	for i := 1; i < cv.NumField(); i++ {
		if cv.Field(i).Uint() == 0 {
			return fmt.Errorf("assistant lifecycle limits must be positive")
		}
	}
	if l.ProcessSummaryBytes > 4096 {
		return fmt.Errorf("assistant process summary limit too large")
	}
	if l.InputCharacters > 4000 || l.InputBytes > 16384 || l.TitleCharacters > 100 || l.ResponseBytes > 32768 || l.AttemptBytes > 65536 || l.PageBytes > 262144 || l.FrameBytes > 262144 || l.PageSizeMax > 100 || l.PageSize > l.PageSizeMax || l.VersionsPageSize > l.PageSizeMax || l.ResponseBytes > l.AttemptBytes || l.InputBytes > l.AttemptBytes || l.ConversationHeaderBytes > l.ConversationBytes || l.AttemptBytes > l.ConversationBytes || l.ConversationBytes > l.UserBytes {
		return fmt.Errorf("invalid assistant content or pagination limits")
	}
	if c.LeaseHeartbeatSeconds >= c.LeaseSeconds || c.CommandMaxAgeHours > 24 || c.CommandFutureSkewSeconds > 300 || c.ReceiptRetentionHours < 48 || c.ReceiptRetentionHours <= c.CommandMaxAgeHours || c.CleanupHours > 24 || c.MemoryCandidateDays > 7 || c.CheckpointMilliseconds > 1000 {
		return fmt.Errorf("invalid assistant lifecycle limits")
	}
	reserved := uint64(l.OutputTokens) + uint64(c.ProtocolTokens)
	if uint64(c.PrimaryContextTokens) <= reserved || uint64(c.FallbackContextTokens) <= reserved || l.SummaryTokens+l.MemoryTokens >= l.InputTokens {
		return fmt.Errorf("invalid assistant context budgets")
	}
	return nil
}

// GetAIAssistantExperienceConfig also supports tests and embedded callers constructing Config directly.
func (config *Config) GetAIAssistantExperienceConfig() AIAssistantExperienceConfig {
	if config.AIAssistantExperience.Limits.InputCharacters == 0 {
		return DefaultAIAssistantExperienceConfig()
	}
	return config.AIAssistantExperience
}
