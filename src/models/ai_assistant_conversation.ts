import type { AIAssistantQuery } from './large_language_model.ts';

export type AssistantRunState = 'queued' | 'preparing' | 'running' | 'cancel_requested' | 'completed' | 'incomplete' | 'failed' | 'canceled' | 'interrupted';
export type AssistantReasonCode = 'queue_timeout' | 'run_timeout' | 'execution_limit' | 'generation_unavailable' | 'input_too_long' | 'conversation_conflict' | 'data_epoch_conflict' | 'memory_conflict' | 'context_changed' | 'already_running' | 'quota_exceeded' | 'queue_full' | 'subscription_limit' | 'cursor_expired' | 'context_limit' | 'output_limit' | 'empty_response' | 'storage_unavailable' | 'provider_unavailable' | 'interrupted' | 'not_accessible';

export interface AssistantCommand {
    readonly clientRequestId: string;
    readonly requestIssuedAt: string;
    readonly expectedDeletionEpoch: number;
}

export interface AssistantSendRequest extends AssistantCommand {
    /** Explicitly restart the authorized saved query; only valid for next_page. */
    readonly restartQuery?: boolean;
    readonly conversationId?: string;
    readonly expectedContextVersion: number;
    readonly mode: string;
    readonly content: string;
    readonly sourceMessageId?: string;
    readonly locale?: string;
    readonly timezone?: string;
}

export interface AssistantRetryRequest extends AssistantCommand {
    readonly conversationId: string;
    readonly userMessageId: string;
    readonly expectedContextVersion: number;
    readonly intent: string;
}

export interface AssistantBranchRequest extends AssistantCommand {
    readonly sourceConversationId: string;
    readonly sourceMessageId: string;
    readonly expectedContextVersion: number;
    readonly kind: string;
}

export interface AssistantUpdateConversationRequest extends AssistantCommand {
    readonly id: string;
    readonly expectedMetadataVersion: number;
    readonly expectedContextVersion?: number;
    readonly title?: string;
    readonly archived?: boolean;
    readonly memoryMode?: string;
}

export interface AssistantDeleteConversationRequest extends AssistantCommand {
    readonly id: string;
    readonly expectedMetadataVersion: number;
}

export type AssistantClearRequest = AssistantCommand;

export interface AssistantSelectVersionRequest extends AssistantCommand {
    readonly conversationId: string;
    readonly userMessageId: string;
    readonly assistantMessageId: string;
    readonly expectedContextVersion: number;
}

export interface AssistantCancelRequest extends AssistantCommand {
    readonly runId: string;
}

export interface AssistantSaveMemoryRequest extends AssistantCommand {
    readonly id?: string;
    readonly expectedVersion: number;
    readonly expectedMemoryUseEpoch: number;
    readonly text: string;
    readonly key: string;
    readonly type: string;
    readonly intent: string;
    readonly sourceMessageIds?: string[];
    readonly expiresAt?: string;
}

export interface AssistantDeleteMemoryRequest extends AssistantCommand {
    readonly id?: string;
    readonly expectedVersion: number;
    readonly all?: boolean;
    readonly expectedMemoryUseEpoch: number;
}

export interface AssistantMemorySettingsRequest extends AssistantCommand {
    readonly enabled: boolean;
    readonly expectedSettingsVersion: number;
}

export interface AssistantListRequest {
    readonly cursor?: string;
    readonly limit?: number;
    readonly archived?: boolean;
}

export interface AssistantSearchRequest {
    readonly query: string;
    readonly cursor?: string;
    readonly limit?: number;
    readonly archived?: boolean;
    readonly includeVersions?: boolean;
}

export interface AssistantGetConversationRequest {
    readonly id: string;
    readonly beforeTurnSequence?: number;
    readonly aroundMessageId?: string;
    readonly limit?: number;
}

export interface AssistantExportRequest {
    /** Display-only selections for one conversation; includeVersions exports all attempts. */
    readonly viewedVersionIds?: string[];
    readonly id?: string;
    readonly all?: boolean;
    readonly format: string;
    readonly includeVersions?: boolean;
}

export interface AssistantVersionsRequest {
    readonly userMessageId: string;
    readonly cursor?: string;
    readonly limit?: number;
}

export interface AssistantRunRequest {
    readonly runId?: string;
    readonly clientRequestId?: string;
    readonly afterRevision?: number;
}

export interface AssistantVersionDTO {
    readonly id: string;
    readonly attemptNumber: number;
    readonly state: AssistantRunState;
    readonly createdAt: string;
    readonly completedAt?: string;
    readonly reasonCode?: AssistantReasonCode;
}

export interface AssistantReferenceDTO {
    readonly hideAmount?: boolean;
    readonly id: string;
    readonly time: number;
    readonly timeText?: string;
    readonly type: number;
    readonly categoryName?: string;
    readonly sourceAccountName?: string;
    readonly destinationAccountName?: string;
    readonly sourceAmountMinor: string;
    readonly destinationAmountMinor?: string;
    readonly currency?: string;
    readonly destinationCurrency?: string;
    readonly comment?: string;
}

export interface AssistantRunSnapshot {
    readonly processSummary?: string;
    readonly schemaVersion: number;
    readonly runId: string;
    readonly conversationId: string;
    readonly userMessageId: string;
    readonly assistantMessageId: string;
    readonly attemptNumber: number;
    readonly revision: number;
    readonly state: AssistantRunState;
    readonly progress?: string;
    readonly content: string;
    readonly queries: AIAssistantQuery[];
    readonly references: AssistantReferenceDTO[];
    readonly reasonCode: AssistantReasonCode | null;
    readonly savedAt: string;
    readonly pageableQueryId?: string;
}

export interface AssistantMessageDTO {
    readonly schemaVersion: number;
    readonly id: string;
    readonly conversationId: string;
    readonly turnSequence: number;
    readonly role: string;
    readonly content: string;
    readonly createdAt: string;
    readonly selectedAssistantMessageId?: string;
    readonly previousAssistantMessageId?: string;
    readonly locale?: string;
    readonly timezone?: string;
    readonly versions: AssistantVersionDTO[];
    readonly selectedAnswer?: AssistantRunSnapshot;
    readonly hasMoreVersions: boolean;
}

/** version tracks source revocation; normal append/rename/archive do not invalidate it. */
export interface AssistantSourceDependency {
    readonly conversationId: string;
    readonly messageId: string;
    readonly version: number;
}

export interface AssistantConversationDTO {
    /** Pending continue-branch saved-query answer ID; never a cursor. */
    readonly pendingQuerySourceMessageId?: string;
    readonly lastRunState?: AssistantRunState;
    readonly lastReasonCode?: AssistantReasonCode;
    readonly schemaVersion: number;
    readonly id: string;
    readonly title: string;
    readonly archived: boolean;
    readonly createdAt: string;
    readonly activityAt: string;
    readonly metadataVersion: number;
    readonly contextVersion: number;
    readonly nextTurnSequence: number;
    readonly activeRunId?: string;
    readonly usedBytes: number;
    readonly reservedBytes: number;
    readonly turnCount: number;
    readonly memoryMode: string;
    readonly sourceConversationId?: string;
    readonly sourceMessageId?: string;
    readonly branchKind?: string;
    readonly pendingContinuation: boolean;
    readonly sourceInvalidated: boolean;
}

export interface AssistantConversationDetail {
    readonly conversation: AssistantConversationDTO;
    readonly messages: AssistantMessageDTO[];
    readonly activeRun?: AssistantRunSnapshot;
    readonly nextCursor?: string;
    readonly hasMore: boolean;
}

export interface AssistantConversationState {
    readonly id: string;
    readonly metadataVersion: number;
    readonly contextVersion: number;
    readonly activeRunId?: string;
    readonly activeRunRevision: number;
    readonly selectedAssistantMessageId?: string;
    readonly archived: boolean;
    readonly memoryMode: string;
    readonly accessible: boolean;
}

export interface AssistantConversationList {
    readonly items: AssistantConversationDTO[];
    readonly nextCursor?: string;
}

export interface AssistantSearchMatch {
    readonly conversationId: string;
    readonly messageId: string;
    readonly versionId?: string;
    readonly title: string;
    readonly snippet: string;
}

export interface AssistantSearchResult {
    readonly items: AssistantSearchMatch[];
    readonly nextCursor?: string;
}

export interface AssistantVersionsResult {
    readonly items: AssistantVersionDTO[];
    readonly nextCursor?: string;
}

export interface AssistantCommandResult {
    /** Whole-user clear completion, distinct from the initial assistant write fence. */
    readonly operationCompleted?: boolean;
    readonly clientRequestId: string;
    readonly commandType: string;
    readonly status: string;
    readonly conversationId?: string;
    readonly userMessageId?: string;
    readonly assistantMessageId?: string;
    readonly runId?: string;
    readonly memoryId?: string;
    readonly metadataVersion?: number;
    readonly contextVersion?: number;
    readonly deletionEpoch: number;
    readonly settingsVersion?: number;
    readonly memoryUseEpoch?: number;
    readonly state?: AssistantRunState;
}

export interface AssistantMemoryDTO {
    readonly schemaVersion: number;
    readonly id: string;
    readonly key: string;
    readonly text: string;
    readonly type: string;
    readonly status: string;
    readonly sourceMessageIds: string[];
    readonly sources: AssistantSourceDependency[];
    readonly createdAt: string;
    readonly updatedAt: string;
    readonly expiresAt?: string;
    readonly version: number;
    readonly explicit: boolean;
}

export interface AssistantMemoryList {
    readonly items: AssistantMemoryDTO[];
    readonly enabled: boolean;
    readonly settingsVersion: number;
    readonly memoryUseEpoch: number;
}

export interface AssistantSummaryDTO {
    readonly schemaVersion: number;
    readonly topic: string;
    readonly confirmedConstraints: string[];
    readonly resolvedReferences: string[];
    readonly openQuestions: string[];
    readonly sourceMessageIds: string[];
    readonly sources: AssistantSourceDependency[];
    readonly coveredThrough: number;
    readonly contextVersion: number;
    readonly generatedAt: string;
}

export interface AssistantInputDependencies {
    readonly deletionEpoch: number;
    readonly memoryUseEpoch: number;
    readonly memoryMode: string;
    readonly memories: AssistantMemoryDependency[];
    readonly sources: AssistantSourceDependency[];
}

export interface AssistantMemoryDependency {
    readonly id: string;
    readonly version: number;
}

export interface AssistantUsage {
    readonly usedBytes: number;
    readonly reservedBytes: number;
    readonly conversationCount: number;
    readonly runningCount: number;
    readonly queuedCount: number;
    readonly activeMemoryCount: number;
    readonly candidateMemoryCount: number;
}

export interface AssistantLimits {
    readonly generationRequestsPerMinute: number;
    readonly searchExportRequestsPerMinute: number;
    readonly statusRequestsPerMinute: number;
    readonly searchExportConcurrency: number;
    readonly searchTimeoutSeconds: number;
    readonly exportTimeoutSeconds: number;
    readonly rateLimitUsers: number;

    readonly processSummaryBytes: number;
    readonly inputCharacters: number;
    readonly inputBytes: number;
    readonly titleCharacters: number;
    readonly responseBytes: number;
    readonly attemptBytes: number;
    readonly conversationBytes: number;
    readonly conversationTurns: number;
    readonly conversationHeaderBytes: number;
    readonly userBytes: number;
    readonly conversations: number;
    readonly memories: number;
    readonly memoryBytes: number;
    readonly memoryCandidates: number;
    readonly runningPerUser: number;
    readonly queuedPerUser: number;
    readonly workerConcurrency: number;
    readonly pageSize: number;
    readonly pageSizeMax: number;
    readonly versionsPageSize: number;
    readonly recentVersions: number;
    readonly pageBytes: number;
    readonly frameBytes: number;
    readonly subscriptionsPerUser: number;
    readonly subscriptionsPerRun: number;
    readonly inputTokens: number;
    readonly summaryTokens: number;
    readonly memoryTokens: number;
    readonly outputTokens: number;
    readonly requestBytes: number;
}

export interface AssistantCapabilities {
    readonly schemaVersion: number;
    readonly generationAvailable: boolean;
    readonly unavailableReason?: AssistantReasonCode;
    readonly historyAvailable: boolean;
    readonly memoryAvailable: boolean;
    readonly ownerScope: string;
    readonly serverTime: string;
    readonly deletionEpoch: number;
    readonly settingsVersion: number;
    readonly memoryUseEpoch: number;
    readonly memoryEnabled: boolean;
    readonly limits: AssistantLimits;
    readonly usage: AssistantUsage;
}
