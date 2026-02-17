import { ref, computed } from 'vue';

import type {
    AIAssistantChatRequest,
    AIAssistantChatResponse,
    AIAssistantHistoryItem,
    AIAssistantReferencedTransaction
} from '@/models/large_language_model.ts';

import { isAIAssistantEnabled } from '@/lib/server_settings.ts';
import services from '@/lib/services.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import logger from '@/lib/logger.ts';

export type AIAssistantConversationRole = 'user' | 'assistant';

export interface AIAssistantConversationMessage {
    readonly id: string;
    readonly role: AIAssistantConversationRole;
    readonly content: string;
    readonly createdAt: number;
    readonly references?: AIAssistantReferencedTransaction[];
}

const maxHistoryMessageCount = 12;

export function useAssistantPageBase() {
    const enabled = computed<boolean>(() => isAIAssistantEnabled());
    const messages = ref<AIAssistantConversationMessage[]>([]);
    const messageInput = ref<string>('');
    const requesting = ref<boolean>(false);
    const cancelableUuid = ref<string | undefined>(undefined);

    const canSendMessage = computed<boolean>(() => {
        return !!messageInput.value.trim() && !requesting.value;
    });

    function clearConversation(): void {
        messages.value = [];
    }

    function cancelCurrentRequest(): void {
        if (!cancelableUuid.value) {
            return;
        }

        services.cancelRequest(cancelableUuid.value);
    }

    function getHistoryPayload(): AIAssistantHistoryItem[] {
        if (!messages.value.length) {
            return [];
        }

        const history: AIAssistantHistoryItem[] = [];
        let startIndex = messages.value.length - maxHistoryMessageCount;

        if (startIndex < 0) {
            startIndex = 0;
        }

        for (let i = startIndex; i < messages.value.length; i++) {
            const message = messages.value[i] as AIAssistantConversationMessage;

            if (!message || !message.content) {
                continue;
            }

            history.push({
                role: message.role,
                content: message.content
            });
        }

        return history;
    }

    function appendConversationMessage(message: AIAssistantConversationMessage): void {
        messages.value = [...messages.value, message];
    }

    function appendUserMessage(content: string): void {
        appendConversationMessage({
            id: generateRandomUUID(),
            role: 'user',
            content: content,
            createdAt: Date.now()
        });
    }

    function appendAssistantMessage(response: AIAssistantChatResponse): void {
        appendConversationMessage({
            id: generateRandomUUID(),
            role: 'assistant',
            content: response.reply,
            createdAt: Date.now(),
            references: response.references
        });
    }

    async function requestAIAssistant(req: AIAssistantChatRequest): Promise<AIAssistantChatResponse> {
        cancelableUuid.value = generateRandomUUID();
        requesting.value = true;

        try {
            const response = await services.chatWithAIAssistant({
                req,
                cancelableUuid: cancelableUuid.value
            });
            const data = response.data;

            if (!data || !data.success || !data.result) {
                throw { message: 'Unable to get AI assistant response' };
            }

            return data.result;
        } catch (error: unknown) {
            const typedError = error as {
                canceled?: boolean;
                processed?: boolean;
                response?: {
                    data?: {
                        errorMessage?: string;
                    };
                };
            };

            if (typedError.canceled) {
                throw typedError;
            }

            logger.error('failed to request ai assistant', typedError);

            if (typedError.response && typedError.response.data && typedError.response.data.errorMessage) {
                throw { error: typedError.response.data };
            }

            if (typedError.processed) {
                throw typedError;
            }

            throw { message: 'Unable to get AI assistant response' };
        } finally {
            requesting.value = false;
            cancelableUuid.value = undefined;
        }
    }

    async function sendMessage(): Promise<void> {
        if (requesting.value) {
            return;
        }

        const message = messageInput.value.trim();

        if (!message) {
            return;
        }

        const history = getHistoryPayload();
        appendUserMessage(message);
        messageInput.value = '';

        const response = await requestAIAssistant({
            mode: 'chat',
            message: message,
            history
        });

        appendAssistantMessage(response);
    }

    async function generateSummary(): Promise<void> {
        if (requesting.value) {
            return;
        }

        const response = await requestAIAssistant({
            mode: 'summary',
            history: getHistoryPayload()
        });

        appendAssistantMessage(response);
    }

    return {
        enabled,
        messages,
        messageInput,
        requesting,
        canSendMessage,
        clearConversation,
        cancelCurrentRequest,
        sendMessage,
        generateSummary
    };
}
