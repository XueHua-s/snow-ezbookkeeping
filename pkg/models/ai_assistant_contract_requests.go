package models

// AssistantCommand is the versioned assistant wire contract.
type AssistantCommand struct {
	ClientRequestId       string `json:"clientRequestId"`
	RequestIssuedAt       string `json:"requestIssuedAt"`
	ExpectedDeletionEpoch int64  `json:"expectedDeletionEpoch"`
}

// AssistantSendRequest is the versioned assistant wire contract.
type AssistantSendRequest struct {
	// RestartQuery explicitly repeats the authorized saved query from its first page, never an expired cursor.
	// It is only valid with mode next_page and a server-authorized sourceMessageId.
	RestartQuery bool `json:"restartQuery,omitempty"`
	AssistantCommand
	ConversationId         string `json:"conversationId,omitempty"`
	ExpectedContextVersion int64  `json:"expectedContextVersion"`
	Mode                   string `json:"mode"`
	Content                string `json:"content"`
	SourceMessageId        string `json:"sourceMessageId,omitempty"`
	Locale                 string `json:"locale,omitempty"`
	Timezone               string `json:"timezone,omitempty"`
}

// AssistantRetryRequest is the versioned assistant wire contract.
type AssistantRetryRequest struct {
	AssistantCommand
	ConversationId         string `json:"conversationId"`
	UserMessageId          string `json:"userMessageId"`
	ExpectedContextVersion int64  `json:"expectedContextVersion"`
	Intent                 string `json:"intent"`
}

// AssistantBranchRequest is the versioned assistant wire contract.
type AssistantBranchRequest struct {
	AssistantCommand
	SourceConversationId   string `json:"sourceConversationId"`
	SourceMessageId        string `json:"sourceMessageId"`
	ExpectedContextVersion int64  `json:"expectedContextVersion"`
	Kind                   string `json:"kind"`
}

// AssistantUpdateConversationRequest is the versioned assistant wire contract.
type AssistantUpdateConversationRequest struct {
	AssistantCommand
	Id                      string  `json:"id"`
	ExpectedMetadataVersion int64   `json:"expectedMetadataVersion"`
	ExpectedContextVersion  *int64  `json:"expectedContextVersion,omitempty"`
	Title                   *string `json:"title,omitempty"`
	Archived                *bool   `json:"archived,omitempty"`
	MemoryMode              *string `json:"memoryMode,omitempty"`
}

// AssistantDeleteConversationRequest is the versioned assistant wire contract.
type AssistantDeleteConversationRequest struct {
	AssistantCommand
	Id                      string `json:"id"`
	ExpectedMetadataVersion int64  `json:"expectedMetadataVersion"`
}

// AssistantClearRequest is the versioned assistant wire contract.
type AssistantClearRequest struct {
	AssistantCommand
}

// AssistantSelectVersionRequest is the versioned assistant wire contract.
type AssistantSelectVersionRequest struct {
	AssistantCommand
	ConversationId         string `json:"conversationId"`
	UserMessageId          string `json:"userMessageId"`
	AssistantMessageId     string `json:"assistantMessageId"`
	ExpectedContextVersion int64  `json:"expectedContextVersion"`
}

// AssistantCancelRequest is the versioned assistant wire contract.
type AssistantCancelRequest struct {
	AssistantCommand
	RunId string `json:"runId"`
}

// AssistantSaveMemoryRequest is the versioned assistant wire contract.
type AssistantSaveMemoryRequest struct {
	AssistantCommand
	Id                     string   `json:"id,omitempty"`
	ExpectedVersion        int64    `json:"expectedVersion"`
	ExpectedMemoryUseEpoch int64    `json:"expectedMemoryUseEpoch"`
	Text                   string   `json:"text"`
	Key                    string   `json:"key"`
	Type                   string   `json:"type"`
	Intent                 string   `json:"intent"`
	SourceMessageIds       []string `json:"sourceMessageIds,omitempty"`
	ExpiresAt              string   `json:"expiresAt,omitempty"`
}

// AssistantDeleteMemoryRequest is the versioned assistant wire contract.
type AssistantDeleteMemoryRequest struct {
	AssistantCommand
	Id                     string `json:"id,omitempty"`
	ExpectedVersion        int64  `json:"expectedVersion"`
	All                    bool   `json:"all,omitempty"`
	ExpectedMemoryUseEpoch int64  `json:"expectedMemoryUseEpoch"`
}

// AssistantMemorySettingsRequest is the versioned assistant wire contract.
type AssistantMemorySettingsRequest struct {
	AssistantCommand
	Enabled                 bool  `json:"enabled"`
	ExpectedSettingsVersion int64 `json:"expectedSettingsVersion"`
}

// AssistantListRequest is the versioned assistant wire contract.
type AssistantListRequest struct {
	Cursor   string `json:"cursor,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Archived bool   `json:"archived,omitempty"`
}

// AssistantSearchRequest is the versioned assistant wire contract.
type AssistantSearchRequest struct {
	Query           string `json:"query"`
	Cursor          string `json:"cursor,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	Archived        bool   `json:"archived,omitempty"`
	IncludeVersions bool   `json:"includeVersions,omitempty"`
}

// AssistantGetConversationRequest is the versioned assistant wire contract.
type AssistantGetConversationRequest struct {
	Id                 string `json:"id"`
	BeforeTurnSequence int64  `json:"beforeTurnSequence,omitempty"`
	AroundMessageId    string `json:"aroundMessageId,omitempty"`
	Limit              int    `json:"limit,omitempty"`
}

// AssistantExportRequest is the versioned assistant wire contract.
type AssistantExportRequest struct {
	// ViewedVersionIds selects only the displayed answers for a single-conversation export.
	// Each ID must belong to this user/conversation, with at most one answer per turn.
	// It is invalid with All; IncludeVersions explicitly exports every version instead.
	ViewedVersionIds []string `json:"viewedVersionIds,omitempty"`
	Id               string   `json:"id,omitempty"`
	All              bool     `json:"all,omitempty"`
	Format           string   `json:"format"`
	IncludeVersions  bool     `json:"includeVersions,omitempty"`
}

// AssistantVersionsRequest is the versioned assistant wire contract.
type AssistantVersionsRequest struct {
	UserMessageId string `json:"userMessageId"`
	Cursor        string `json:"cursor,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

// AssistantRunRequest is the versioned assistant wire contract.
type AssistantRunRequest struct {
	RunId           string `json:"runId,omitempty"`
	ClientRequestId string `json:"clientRequestId,omitempty"`
	AfterRevision   int64  `json:"afterRevision,omitempty"`
}
