import type {
    AIAssistantChatResponse,
    AIAssistantChatStreamChunk,
    AIAssistantErrorContext,
    AIAssistantStatus
} from '@/models/large_language_model.ts';

export interface AIAssistantErrorEnvelope {
    readonly success: false;
    readonly errorCode?: number;
    readonly errorMessage?: string;
    readonly path?: string;
    readonly context?: { readonly assistant?: AIAssistantErrorContext; readonly [key: string]: unknown };
}

export interface AIAssistantRequestError {
    readonly message?: string;
    readonly canceled?: boolean;
    readonly timedOut?: boolean;
    readonly incomplete?: boolean;
    readonly processed?: boolean;
    readonly response?: { readonly data?: AIAssistantErrorEnvelope };
}

export interface AIAssistantChatStreamCallbacks {
    readonly onStatus?: (status: AIAssistantStatus) => void;
    readonly onThinkingDelta?: (delta: string) => void;
    readonly onReplyReset?: () => void;
    readonly onReplyDelta?: (delta: string) => void;
    readonly onReferences?: (references: AIAssistantChatResponse['references']) => void;
    readonly onDone?: (chunk: AIAssistantChatStreamChunk) => void;
}
