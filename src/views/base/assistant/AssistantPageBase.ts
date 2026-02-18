import { ref, computed } from 'vue';

import type {
    AIAssistantChatRequest,
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
    readonly thinking?: string;
    readonly createdAt: number;
    readonly references?: AIAssistantReferencedTransaction[];
}

const maxHistoryMessageCount = 12;

export function useAssistantPageBase() {
    const enabled = computed<boolean>(() => isAIAssistantEnabled());
    const messages = ref<AIAssistantConversationMessage[]>([]);
    const messageInput = ref<string>('');
    const requesting = ref<boolean>(false);
    const rendering = ref<boolean>(false);
    const cancelableUuid = ref<string | undefined>(undefined);

    const canSendMessage = computed<boolean>(() => {
        return !!messageInput.value.trim() && !requesting.value && !rendering.value;
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

    function patchConversationMessage(messageId: string, patch: Partial<AIAssistantConversationMessage>): void {
        messages.value = messages.value.map(message => {
            if (message.id !== messageId) {
                return message;
            }

            return {
                ...message,
                ...patch
            };
        });
    }

    function appendUserMessage(content: string): void {
        appendConversationMessage({
            id: generateRandomUUID(),
            role: 'user',
            content: content,
            createdAt: Date.now()
        });
    }

    async function appendAssistantMessageByStream(req: AIAssistantChatRequest): Promise<void> {
        const messageId = generateRandomUUID();
        let latestReply = '';
        let latestThinking = '';

        appendConversationMessage({
            id: messageId,
            role: 'assistant',
            content: '',
            thinking: '',
            createdAt: Date.now()
        });

        rendering.value = true;

        try {
            await requestAIAssistantStream(req, {
                onThinkingDelta: delta => {
                    latestThinking += delta;
                    patchConversationMessage(messageId, {
                        thinking: latestThinking
                    });
                },
                onReplyDelta: delta => {
                    latestReply += delta;
                    patchConversationMessage(messageId, {
                        content: latestReply
                    });
                },
                onReferences: references => {
                    patchConversationMessage(messageId, {
                        references
                    });
                },
                onDone: chunk => {
                    const doneReply = chunk.reply || latestReply;
                    const doneThinking = chunk.thinking || latestThinking;

                    if (doneReply) {
                        latestReply = doneReply;
                    }

                    if (doneThinking) {
                        latestThinking = doneThinking;
                    }

                    patchConversationMessage(messageId, {
                        content: latestReply,
                        thinking: latestThinking
                    });
                }
            });

            if (!latestReply) {
                patchConversationMessage(messageId, {
                    content: ''
                });
            }
        } finally {
            rendering.value = false;
        }
    }

    async function requestAIAssistantStream(req: AIAssistantChatRequest, callbacks: {
        onThinkingDelta?: (delta: string) => void;
        onReplyDelta?: (delta: string) => void;
        onReferences?: (references: AIAssistantReferencedTransaction[] | undefined) => void;
        onDone?: (chunk: {
            reply?: string;
            thinking?: string;
        }) => void;
    }): Promise<void> {
        cancelableUuid.value = generateRandomUUID();
        requesting.value = true;

        try {
            await services.chatWithAIAssistantStream({
                req,
                cancelableUuid: cancelableUuid.value,
                callbacks
            });
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
                throw { message: typedError.response.data.errorMessage };
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
        if (requesting.value || rendering.value) {
            return;
        }

        const message = messageInput.value.trim();

        if (!message) {
            return;
        }

        const history = getHistoryPayload();
        appendUserMessage(message);
        messageInput.value = '';

        await appendAssistantMessageByStream({
            mode: 'chat',
            message: message,
            history
        });
    }

    async function generateSummary(): Promise<void> {
        if (requesting.value || rendering.value) {
            return;
        }

        await appendAssistantMessageByStream({
            mode: 'summary',
            history: getHistoryPayload()
        });
    }

    return {
        enabled,
        messages,
        messageInput,
        requesting,
        rendering,
        canSendMessage,
        clearConversation,
        cancelCurrentRequest,
        sendMessage,
        generateSummary
    };
}
