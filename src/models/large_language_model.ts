export enum CodingAssistantType {
    InsightsExplorerCustomChart = 'insights_explorer_custom_chart'
}

export interface RecognizedTransactionResponse {
    readonly type: number;
    readonly time?: number;
    readonly categoryId?: string;
    readonly sourceAccountId?: string;
    readonly destinationAccountId?: string;
    readonly sourceAmount?: number;
    readonly destinationAmount?: number;
    readonly tagIds?: string[];
    readonly comment?: string;
}

export type RecognizedReceiptImageResponse = RecognizedTransactionResponse;

export interface RecognizedReceiptImageResultItem {
    readonly index: number;
    readonly success: boolean;
    readonly result?: RecognizedReceiptImageResponse;
    readonly error?: string;
}

export interface RecognizedReceiptImageBatchResponse {
    readonly results: RecognizedReceiptImageResultItem[];
}

export type AIAssistantMode = 'chat' | 'summary';
export type AIAssistantMessageRole = 'user' | 'assistant';

export interface AIAssistantHistoryItem {
    readonly role: AIAssistantMessageRole;
    readonly content: string;
}

export interface AIAssistantChatRequest {
    readonly mode: AIAssistantMode;
    readonly message?: string;
    readonly history?: AIAssistantHistoryItem[];
    readonly continuation?: AIAssistantContinuation;
}

export interface AIAssistantReferencedTransaction {
    readonly id: string;
    readonly time: number;
    readonly timeText?: string;
    readonly type: number;
    readonly categoryName?: string;
    readonly sourceAccountName?: string;
    readonly destinationAccountName?: string;
    readonly sourceAmount: number;
    readonly sourceAmountMinor?: string;
    readonly destinationAmountMinor?: string;
    readonly destinationAmount?: number;
    readonly currency?: string;
    readonly destinationCurrency?: string;
    readonly comment?: string;
    /** @deprecated Removed when the legacy assistant runtime is retired. */
    readonly hideAmount?: boolean;
}

export interface AIAssistantChatResponse {
    readonly mode: AIAssistantMode;
    readonly reply: string;
    readonly references?: AIAssistantReferencedTransaction[];
    readonly queries?: AIAssistantQuery[];
    readonly continuation?: AIAssistantContinuation;
}

export type AIAssistantChatStreamChunkType = 'status' | 'thinking_delta' | 'reply_delta' | 'references' | 'done';

export interface AIAssistantChatStreamChunk {
    readonly type: AIAssistantChatStreamChunkType;
    readonly status?: AIAssistantStatus;
    readonly mode?: AIAssistantMode;
    readonly delta?: string;
    readonly reply?: string;
    readonly thinking?: string;
    readonly references?: AIAssistantReferencedTransaction[];
    readonly queries?: AIAssistantQuery[];
    readonly continuation?: AIAssistantContinuation;
}

export interface CodingAssistantRequest {
    readonly type: CodingAssistantType;
    readonly userPrompt: string;
    readonly code?: string;
}

export interface CodingAssistantResponse {
    readonly code: string;
    readonly failureReason?: string;
}

export type AIAssistantState = 'requesting' | 'completed' | 'incomplete' | 'failed' | 'canceled';
export type AIAssistantStatus = 'querying' | 'aggregating' | 'answering';
export type AIAssistantQueryCoverage = 'page' | 'all_matching';
export type AIAssistantTool = 'list_accounts' | 'list_categories' | 'list_tags' | 'query_transactions' | 'get_transaction' | 'aggregate_transactions' | 'get_account_balances';
export type AIAssistantPageableTool = Exclude<AIAssistantTool, 'get_transaction'>;

export interface AIAssistantQueryPeriod {
    readonly name: string;
    readonly startTime: string;
    readonly endTime: string;
    readonly matchedCount?: string;
}

export interface AIAssistantQueryScope {
    readonly startTime: string | null;
    readonly endTime: string | null;
    readonly timezone: string;
    readonly currency?: string;
    readonly types?: string[];
    readonly accountIds?: string[];
    readonly categoryIds?: string[];
    readonly tagIds?: string[];
    readonly tagMatch?: 'any' | 'all';
    readonly noTags?: boolean;
    readonly transferSide?: 'source' | 'destination' | 'either';
    readonly amountMin?: string;
    readonly amountMax?: string;
    readonly keyword?: string;
    readonly periods?: AIAssistantQueryPeriod[];
    readonly includesHidden: boolean;
    readonly excludesDeleted: boolean;
}

export interface AIAssistantQuery {
    readonly queryId: string;
    readonly tool: AIAssistantTool;
    readonly scope: AIAssistantQueryScope;
    readonly coverage: AIAssistantQueryCoverage;
    readonly queriedAt: string;
    readonly warnings: string[];
}

// Opaque candidate context only. Keep in page memory and authenticated bodies;
// never render its cursor or assume that a new question resumes its filters.
export interface AIAssistantContinuation {
    readonly tool: AIAssistantPageableTool;
    readonly cursor: string;
}

export interface AIAssistantErrorContext {
    readonly state: 'failed' | 'incomplete';
    readonly queries: AIAssistantQuery[];
    readonly continuation?: AIAssistantContinuation;
}
