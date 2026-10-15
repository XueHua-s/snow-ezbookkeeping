import type { AssistantCommand, AssistantRunState } from '@/models/ai_assistant_conversation.ts';
import type { assistantConversationServices } from '@/lib/services/assistant_conversations.ts';
import type { ApiResponsePromise } from '@/lib/services/transport.ts';

export const terminal = (state: AssistantRunState) => ['completed', 'incomplete', 'failed', 'canceled', 'interrupted'].includes(state);
export type WriteAction = 'send' | 'retry' | 'branch' | 'update' | 'delete' | 'clear' | 'selectVersion' | 'cancel' | 'saveMemory' | 'deleteMemory' | 'memorySettings';
export type WriteInput<K extends WriteAction> = Omit<Parameters<typeof assistantConversationServices[K]>[0], keyof AssistantCommand>;
export interface ReadingPosition { messageId: string; offset: number; following: boolean }
export async function result<T>(promise: ApiResponsePromise<T>): Promise<T> {
    const response = await promise;
    if (!response.data.success) throw response.data;
    return response.data.result;
}
export function statusCode(error: unknown): number | undefined {
    const value = error as { status?: number; response?: { status?: number } };
    return value?.status ?? value?.response?.status;
}

/** Only a structured application rejection proves the command was not admitted. A proxy status does not. */
export function isCommandRejection(error: unknown): boolean {
    const response = (error as { response?: { status?: number; data?: { success?: boolean; errorCode?: number; context?: { reasonCode?: string } } } })?.response;
    const data = response?.data;
    return [400, 404, 409, 413, 429].includes(response?.status ?? 0) && data?.success === false && typeof data.errorCode === 'number' &&
        ['input_too_long', 'conversation_conflict', 'data_epoch_conflict', 'memory_conflict', 'context_changed', 'already_running',
            'quota_exceeded', 'queue_full', 'not_accessible', 'cursor_expired', 'context_limit', 'generation_unavailable'].includes(data.context?.reasonCode ?? '');
}

export class AssistantDataClearIncompleteError extends Error {
    constructor() { super('Assistant data was cleared. Check the remaining user data before starting another clear operation.'); }
}
