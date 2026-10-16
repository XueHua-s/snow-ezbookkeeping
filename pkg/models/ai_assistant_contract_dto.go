package models

import "github.com/mayswind/ezbookkeeping/pkg/core"

// AssistantVersionDTO is the versioned assistant wire contract.
type AssistantVersionDTO struct {
	Id            string              `json:"id"`
	AttemptNumber int                 `json:"attemptNumber"`
	State         AssistantRunState   `json:"state"`
	CreatedAt     string              `json:"createdAt"`
	CompletedAt   string              `json:"completedAt,omitempty"`
	ReasonCode    AssistantReasonCode `json:"reasonCode,omitempty"`
}

// AssistantReferenceDTO is the versioned assistant wire contract.
type AssistantReferenceDTO struct {
	HideAmount             bool   `json:"hideAmount,omitempty"`
	Id                     string `json:"id"`
	Time                   int64  `json:"time"`
	TimeText               string `json:"timeText,omitempty"`
	Type                   int    `json:"type"`
	CategoryName           string `json:"categoryName,omitempty"`
	SourceAccountName      string `json:"sourceAccountName,omitempty"`
	DestinationAccountName string `json:"destinationAccountName,omitempty"`
	SourceAmountMinor      string `json:"sourceAmountMinor"`
	DestinationAmountMinor string `json:"destinationAmountMinor,omitempty"`
	Currency               string `json:"currency,omitempty"`
	DestinationCurrency    string `json:"destinationCurrency,omitempty"`
	Comment                string `json:"comment,omitempty"`
}

// AssistantRunSnapshot is the versioned assistant wire contract.
type AssistantRunSnapshot struct {
	ProcessSummary     string                  `json:"processSummary,omitempty"`
	SchemaVersion      int                     `json:"schemaVersion"`
	RunId              string                  `json:"runId"`
	ConversationId     string                  `json:"conversationId"`
	UserMessageId      string                  `json:"userMessageId"`
	AssistantMessageId string                  `json:"assistantMessageId"`
	AttemptNumber      int                     `json:"attemptNumber"`
	Revision           int64                   `json:"revision"`
	State              AssistantRunState       `json:"state"`
	Progress           string                  `json:"progress,omitempty"`
	Content            string                  `json:"content"`
	Queries            []AIAssistantQuery      `json:"queries"`
	References         []AssistantReferenceDTO `json:"references"`
	ReasonCode         *AssistantReasonCode    `json:"reasonCode"`
	SavedAt            string                  `json:"savedAt"`
	PageableQueryId    string                  `json:"pageableQueryId,omitempty"`
}

// AssistantMessageDTO is the versioned assistant wire contract.
type AssistantMessageDTO struct {
	SchemaVersion              int                   `json:"schemaVersion"`
	Id                         string                `json:"id"`
	ConversationId             string                `json:"conversationId"`
	TurnSequence               int64                 `json:"turnSequence"`
	Role                       string                `json:"role"`
	Content                    string                `json:"content"`
	CreatedAt                  string                `json:"createdAt"`
	SelectedAssistantMessageId string                `json:"selectedAssistantMessageId,omitempty"`
	PreviousAssistantMessageId string                `json:"previousAssistantMessageId,omitempty"`
	Locale                     string                `json:"locale,omitempty"`
	Timezone                   string                `json:"timezone,omitempty"`
	Versions                   []AssistantVersionDTO `json:"versions"`
	SelectedAnswer             *AssistantRunSnapshot `json:"selectedAnswer,omitempty"`
	HasMoreVersions            bool                  `json:"hasMoreVersions"`
}

// AssistantSourceDependency freezes an immutable source prefix.
// Version is the source revocation version, NOT contextVersion: normal append,
// rename, or archive must not invalidate a branch; deletion/revocation must.
type AssistantSourceDependency struct {
	ConversationId string `json:"conversationId"`
	MessageId      string `json:"messageId"`
	Version        int64  `json:"version"`
}

// AssistantConversationDTO is the versioned assistant wire contract.
type AssistantConversationDTO struct {
	// PendingQuerySourceMessageId is exposed only for pending continue branches with an authorized saved pageable query.
	// It is an answer ID, never an opaque cursor.
	PendingQuerySourceMessageId string              `json:"pendingQuerySourceMessageId,omitempty"`
	LastRunState                AssistantRunState   `json:"lastRunState,omitempty"`
	LastReasonCode              AssistantReasonCode `json:"lastReasonCode,omitempty"`
	SchemaVersion               int                 `json:"schemaVersion"`
	Id                          string              `json:"id"`
	Title                       string              `json:"title"`
	Archived                    bool                `json:"archived"`
	CreatedAt                   string              `json:"createdAt"`
	ActivityAt                  string              `json:"activityAt"`
	MetadataVersion             int64               `json:"metadataVersion"`
	ContextVersion              int64               `json:"contextVersion"`
	NextTurnSequence            int64               `json:"nextTurnSequence"`
	ActiveRunId                 string              `json:"activeRunId,omitempty"`
	UsedBytes                   int64               `json:"usedBytes"`
	ReservedBytes               int64               `json:"reservedBytes"`
	TurnCount                   int                 `json:"turnCount"`
	MemoryMode                  string              `json:"memoryMode"`
	SourceConversationId        string              `json:"sourceConversationId,omitempty"`
	SourceMessageId             string              `json:"sourceMessageId,omitempty"`
	BranchKind                  string              `json:"branchKind,omitempty"`
	PendingContinuation         bool                `json:"pendingContinuation"`
	SourceInvalidated           bool                `json:"sourceInvalidated"`
}

// AssistantConversationDetail is the versioned assistant wire contract.
type AssistantConversationDetail struct {
	Conversation AssistantConversationDTO `json:"conversation"`
	Messages     []AssistantMessageDTO    `json:"messages"`
	ActiveRun    *AssistantRunSnapshot    `json:"activeRun,omitempty"`
	NextCursor   string                   `json:"nextCursor,omitempty"`
	HasMore      bool                     `json:"hasMore"`
}

// AssistantConversationState is the versioned assistant wire contract.
type AssistantConversationState struct {
	Id                         string `json:"id"`
	MetadataVersion            int64  `json:"metadataVersion"`
	ContextVersion             int64  `json:"contextVersion"`
	ActiveRunId                string `json:"activeRunId,omitempty"`
	ActiveRunRevision          int64  `json:"activeRunRevision"`
	SelectedAssistantMessageId string `json:"selectedAssistantMessageId,omitempty"`
	Archived                   bool   `json:"archived"`
	MemoryMode                 string `json:"memoryMode"`
	Accessible                 bool   `json:"accessible"`
}

// AssistantConversationList is the versioned assistant wire contract.
type AssistantConversationList struct {
	Items      []AssistantConversationDTO `json:"items"`
	NextCursor string                     `json:"nextCursor,omitempty"`
}

// AssistantSearchMatch is the versioned assistant wire contract.
type AssistantSearchMatch struct {
	ConversationId string `json:"conversationId"`
	MessageId      string `json:"messageId"`
	VersionId      string `json:"versionId,omitempty"`
	Title          string `json:"title"`
	Snippet        string `json:"snippet"`
}

// AssistantSearchResult is the versioned assistant wire contract.
type AssistantSearchResult struct {
	Items      []AssistantSearchMatch `json:"items"`
	NextCursor string                 `json:"nextCursor,omitempty"`
}

// AssistantVersionsResult is the versioned assistant wire contract.
type AssistantVersionsResult struct {
	Items      []AssistantVersionDTO `json:"items"`
	NextCursor string                `json:"nextCursor,omitempty"`
}

// AssistantCommandResult is the versioned assistant wire contract.
type AssistantCommandResult struct {
	// OperationCompleted is present only for user_data.clear_assistant receipts.
	// False means the assistant write fence exists but whole-user clearing is not confirmed.
	OperationCompleted *bool             `json:"operationCompleted,omitempty"`
	ClientRequestId    string            `json:"clientRequestId"`
	CommandType        string            `json:"commandType"`
	Status             string            `json:"status"`
	ConversationId     string            `json:"conversationId,omitempty"`
	UserMessageId      string            `json:"userMessageId,omitempty"`
	AssistantMessageId string            `json:"assistantMessageId,omitempty"`
	RunId              string            `json:"runId,omitempty"`
	MemoryId           string            `json:"memoryId,omitempty"`
	MetadataVersion    int64             `json:"metadataVersion,omitempty"`
	ContextVersion     int64             `json:"contextVersion,omitempty"`
	DeletionEpoch      int64             `json:"deletionEpoch"`
	SettingsVersion    int64             `json:"settingsVersion,omitempty"`
	MemoryUseEpoch     int64             `json:"memoryUseEpoch,omitempty"`
	State              AssistantRunState `json:"state,omitempty"`
}

// AssistantMemoryDTO is the versioned assistant wire contract.
type AssistantMemoryDTO struct {
	SchemaVersion    int                         `json:"schemaVersion"`
	Id               string                      `json:"id"`
	Key              string                      `json:"key"`
	Text             string                      `json:"text"`
	Type             string                      `json:"type"`
	Status           string                      `json:"status"`
	SourceMessageIds []string                    `json:"sourceMessageIds"`
	Sources          []AssistantSourceDependency `json:"sources"`
	CreatedAt        string                      `json:"createdAt"`
	UpdatedAt        string                      `json:"updatedAt"`
	ExpiresAt        string                      `json:"expiresAt,omitempty"`
	Version          int64                       `json:"version"`
	Explicit         bool                        `json:"explicit"`
}

// AssistantMemoryList is the versioned assistant wire contract.
type AssistantMemoryList struct {
	Items           []AssistantMemoryDTO `json:"items"`
	Enabled         bool                 `json:"enabled"`
	SettingsVersion int64                `json:"settingsVersion"`
	MemoryUseEpoch  int64                `json:"memoryUseEpoch"`
}

// AssistantSummaryDTO is the versioned assistant wire contract.
type AssistantSummaryDTO struct {
	SchemaVersion        int                         `json:"schemaVersion"`
	Topic                string                      `json:"topic"`
	ConfirmedConstraints []string                    `json:"confirmedConstraints"`
	ResolvedReferences   []string                    `json:"resolvedReferences"`
	OpenQuestions        []string                    `json:"openQuestions"`
	SourceMessageIds     []string                    `json:"sourceMessageIds"`
	Sources              []AssistantSourceDependency `json:"sources"`
	CoveredThrough       int64                       `json:"coveredThrough"`
	ContextVersion       int64                       `json:"contextVersion"`
	GeneratedAt          string                      `json:"generatedAt"`
}

// AssistantInputDependencies is the versioned assistant wire contract.
type AssistantInputDependencies struct {
	DeletionEpoch  int64                       `json:"deletionEpoch"`
	MemoryUseEpoch int64                       `json:"memoryUseEpoch"`
	MemoryMode     string                      `json:"memoryMode"`
	Memories       []AssistantMemoryDependency `json:"memories"`
	Sources        []AssistantSourceDependency `json:"sources"`
}

// AssistantMemoryDependency is the versioned assistant wire contract.
type AssistantMemoryDependency struct {
	Id      string `json:"id"`
	Version int64  `json:"version"`
}

// AssistantUsage is the versioned assistant wire contract.
type AssistantUsage struct {
	UsedBytes            int64 `json:"usedBytes"`
	ReservedBytes        int64 `json:"reservedBytes"`
	ConversationCount    int   `json:"conversationCount"`
	RunningCount         int   `json:"runningCount"`
	QueuedCount          int   `json:"queuedCount"`
	ActiveMemoryCount    int   `json:"activeMemoryCount"`
	CandidateMemoryCount int   `json:"candidateMemoryCount"`
}

// AssistantLimits preserves the public model contract without coupling settings to database models.
type AssistantLimits = core.AssistantLimits

// AssistantCapabilities is the versioned assistant wire contract.
type AssistantCapabilities struct {
	SchemaVersion       int                 `json:"schemaVersion"`
	GenerationAvailable bool                `json:"generationAvailable"`
	UnavailableReason   AssistantReasonCode `json:"unavailableReason,omitempty"`
	HistoryAvailable    bool                `json:"historyAvailable"`
	MemoryAvailable     bool                `json:"memoryAvailable"`
	OwnerScope          string              `json:"ownerScope"`
	ServerTime          string              `json:"serverTime"`
	DeletionEpoch       int64               `json:"deletionEpoch"`
	SettingsVersion     int64               `json:"settingsVersion"`
	MemoryUseEpoch      int64               `json:"memoryUseEpoch"`
	MemoryEnabled       bool                `json:"memoryEnabled"`
	Limits              AssistantLimits     `json:"limits"`
	Usage               AssistantUsage      `json:"usage"`
}
