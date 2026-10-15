import type { AIAssistantChatRequest } from '@/models/large_language_model.ts';
import type { AIAssistantChatStreamCallbacks } from '@/lib/types/assistant.ts';

export interface PendingAssistantRequest {
    req: AIAssistantChatRequest;
    cancelableUuid?: string;
    callbacks?: AIAssistantChatStreamCallbacks;
    resolve(): void;
    reject(error: unknown): void;
}
