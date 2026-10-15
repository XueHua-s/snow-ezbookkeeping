package core

// AssistantLimits is the versioned assistant wire contract.
type AssistantLimits struct {
	GenerationRequestsPerMinute   int `json:"generationRequestsPerMinute"`
	SearchExportRequestsPerMinute int `json:"searchExportRequestsPerMinute"`
	StatusRequestsPerMinute       int `json:"statusRequestsPerMinute"`
	SearchExportConcurrency       int `json:"searchExportConcurrency"`
	SearchTimeoutSeconds          int `json:"searchTimeoutSeconds"`
	ExportTimeoutSeconds          int `json:"exportTimeoutSeconds"`
	RateLimitUsers                int `json:"rateLimitUsers"`

	ProcessSummaryBytes     int `json:"processSummaryBytes"`
	InputCharacters         int `json:"inputCharacters"`
	InputBytes              int `json:"inputBytes"`
	TitleCharacters         int `json:"titleCharacters"`
	ResponseBytes           int `json:"responseBytes"`
	AttemptBytes            int `json:"attemptBytes"`
	ConversationBytes       int `json:"conversationBytes"`
	ConversationTurns       int `json:"conversationTurns"`
	ConversationHeaderBytes int `json:"conversationHeaderBytes"`
	UserBytes               int `json:"userBytes"`
	Conversations           int `json:"conversations"`
	Memories                int `json:"memories"`
	MemoryBytes             int `json:"memoryBytes"`
	MemoryCandidates        int `json:"memoryCandidates"`
	RunningPerUser          int `json:"runningPerUser"`
	QueuedPerUser           int `json:"queuedPerUser"`
	WorkerConcurrency       int `json:"workerConcurrency"`
	PageSize                int `json:"pageSize"`
	PageSizeMax             int `json:"pageSizeMax"`
	VersionsPageSize        int `json:"versionsPageSize"`
	RecentVersions          int `json:"recentVersions"`
	PageBytes               int `json:"pageBytes"`
	FrameBytes              int `json:"frameBytes"`
	SubscriptionsPerUser    int `json:"subscriptionsPerUser"`
	SubscriptionsPerRun     int `json:"subscriptionsPerRun"`
	InputTokens             int `json:"inputTokens"`
	SummaryTokens           int `json:"summaryTokens"`
	MemoryTokens            int `json:"memoryTokens"`
	OutputTokens            int `json:"outputTokens"`
	RequestBytes            int `json:"requestBytes"`
}
