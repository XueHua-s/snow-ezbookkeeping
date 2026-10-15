import type { ComputedRef, ShallowRef, DeepReadonly } from 'vue';
import type { AIAssistantReferencedTransaction, AIAssistantState, AIAssistantStatus, AIAssistantQuery, AIAssistantContinuation } from '@/models/large_language_model.ts';
import type { AIAssistantRequestError } from '@/lib/types/assistant.ts';

export type AIAssistantConversationRole = 'user' | 'assistant';

export interface AIAssistantConversationMessage {
    readonly id: string;
    readonly role: AIAssistantConversationRole;
    readonly content: string;
    readonly thinking?: string;
    readonly createdAt: number;
    readonly references?: DeepReadonly<AIAssistantReferencedTransaction[]>;
    readonly state: AIAssistantState;
    readonly progress?: AIAssistantStatus;
    readonly queries?: DeepReadonly<AIAssistantQuery[]>;
    readonly continuation?: AIAssistantContinuation;
    readonly error?: DeepReadonly<AIAssistantRequestError>;
}

export interface AssistantSession {
    readonly enabled: ComputedRef<boolean>;
    readonly messages: Readonly<ShallowRef<readonly AIAssistantConversationMessage[]>>;
    readonly messageInput: ShallowRef<string>;
    readonly requesting: ComputedRef<boolean>;
    readonly state: ComputedRef<AIAssistantState | undefined>;
    readonly progress: ComputedRef<AIAssistantStatus | undefined>;
    readonly continuation: Readonly<ShallowRef<AIAssistantContinuation | undefined>>;
    readonly canSendMessage: ComputedRef<boolean>;
    readonly canRequestNextPage: ComputedRef<boolean>;
    clearConversation(): void;
    disposeConversation(): void;
    cancelCurrentRequest(): void;
    sendMessage(): Promise<void>;
    generateSummary(message: string): Promise<void>;
    requestNextPage(message: string): Promise<void>;
}

export interface AssistantActiveRequest {
    readonly uuid: string;
    readonly messageId: string;
}

export type ReadonlyElementRef = Readonly<ShallowRef<HTMLElement | null>>;
