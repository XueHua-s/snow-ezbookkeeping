import type { AIAssistantConversationMessage, AssistantSession, AssistantActiveRequest } from './types.ts';
import { ASSISTANT_MAX_HISTORY_CHARACTERS, ASSISTANT_MAX_HISTORY_MESSAGES } from '@/consts/assistant.ts';
import { shallowRef, readonly, computed, watch, onScopeDispose, getCurrentInstance, onDeactivated } from 'vue';

import type {
    AIAssistantChatRequest,
    AIAssistantHistoryItem,
    AIAssistantContinuation
} from '@/models/large_language_model.ts';
import type { AIAssistantRequestError } from '@/lib/types/assistant.ts';
import { useUserStore } from '@/stores/user.ts';
import { isAIAssistantEnabled } from '@/lib/server_settings.ts';
import services from '@/lib/services.ts';
import { generateRandomUUID } from '@/lib/misc.ts';

function historyContent(content: string): string {
    const characters = Array.from(content);
    return characters.length > ASSISTANT_MAX_HISTORY_CHARACTERS ? characters.slice(0, ASSISTANT_MAX_HISTORY_CHARACTERS - 1).join('') + '…' : content;
}

/** Keep complete exchanges only, so canceled/partial answers never become history facts. */
export function buildAssistantHistory(messages: readonly AIAssistantConversationMessage[]): AIAssistantHistoryItem[] {
    const history: AIAssistantHistoryItem[] = [];
    for (let i = 1; i < messages.length; i++) {
        const user = messages[i - 1];
        const assistant = messages[i];
        if (user?.role === 'user' && assistant?.role === 'assistant' && assistant.state === 'completed' && assistant.content) {
            history.push({ role: 'user', content: historyContent(user.content) }, { role: 'assistant', content: historyContent(assistant.content) });
        }
    }
    return history.slice(-ASSISTANT_MAX_HISTORY_MESSAGES);
}

export function useAssistantPageBase(): AssistantSession {
    const userStore = useUserStore();
    const enabled = computed<boolean>(() => isAIAssistantEnabled());
    const messages = shallowRef<AIAssistantConversationMessage[]>([]);
    const messageInput = shallowRef('');
    const activeRequest = shallowRef<AssistantActiveRequest>();
    const continuation = shallowRef<AIAssistantContinuation>();
    const requesting = computed(() => !!activeRequest.value);
    const latestAssistant = computed(() => messages.value.findLast(message => message.role === 'assistant'));
    const state = computed(() => latestAssistant.value?.state);
    const progress = computed(() => latestAssistant.value?.progress);
    const canSendMessage = computed(() => !!messageInput.value.trim() && !requesting.value);
    const canRequestNextPage = computed(() => !!continuation.value && !requesting.value);

    function patchMessage(id: string, patch: Partial<AIAssistantConversationMessage>): void {
        messages.value = messages.value.map(message => message.id === id ? { ...message, ...patch } : message);
    }

    function cancelCurrentRequest(): void {
        const active = activeRequest.value;
        activeRequest.value = undefined; // Invalidate callbacks before abort can synchronously notify listeners.
        continuation.value = undefined;
        if (active) {
            patchMessage(active.messageId, { state: 'canceled', progress: undefined, continuation: undefined });
            services.cancelRequest(active.uuid);
        }
    }

    function clearConversation(): void {
        cancelCurrentRequest();
        messages.value = [];
        messageInput.value = '';
    }

    const disposeConversation = clearConversation;
    // UserBasicInfo exposes the unique login name, not the internal numeric UID.
    watch(() => userStore.currentUserBasicInfo?.username, clearConversation, { flush: 'sync' });
    onScopeDispose(disposeConversation);
    if (getCurrentInstance()) {
        onDeactivated(disposeConversation);
    }

    async function request(req: AIAssistantChatRequest): Promise<void> {
        const uuid = generateRandomUUID();
        const messageId = generateRandomUUID();
        activeRequest.value = { uuid, messageId };
        continuation.value = undefined;
        messages.value = [...messages.value, {
            id: messageId, role: 'assistant', content: '', thinking: '', createdAt: Date.now(), state: 'requesting'
        }];
        const isCurrent = () => activeRequest.value?.uuid === uuid;
        let reply = '';
        let thinking = '';
        let completed = false;
        try {
            await services.chatWithAIAssistantStream({ req, cancelableUuid: uuid, callbacks: {
                onStatus: status => {
                    if (isCurrent()) {
                        patchMessage(messageId, { progress: status });
                    }
                },
                onThinkingDelta: delta => {
                    if (isCurrent()) {
                        thinking += delta;
                        patchMessage(messageId, { thinking });
                    }
                },
                onReplyDelta: delta => {
                    if (isCurrent()) {
                        reply += delta;
                        patchMessage(messageId, { content: reply });
                    }
                },
                onReferences: references => {
                    if (isCurrent()) {
                        patchMessage(messageId, { references });
                    }
                },
                onDone: chunk => {
                    if (isCurrent()) {
                        completed = true;
                        continuation.value = chunk.continuation;
                        patchMessage(messageId, {
                            content: chunk.reply ?? reply,
                            thinking: chunk.thinking ?? thinking,
                            ...(chunk.references ? { references: chunk.references } : {}),
                            queries: chunk.queries ?? [],
                            continuation: chunk.continuation,
                            state: 'completed', progress: undefined
                        });
                    }
                }
            } });
            if (isCurrent() && !completed) {
                throw { incomplete: true };
            }
        } catch (error: unknown) {
            if (!isCurrent()) {
                return; // A canceled/cleared request must not alter a newer request or show an old error.
            }
            const original = (error || {}) as AIAssistantRequestError;
            const normalized: AIAssistantRequestError = {
                ...original,
                message: original.response?.data?.errorMessage || original.message || 'Unable to get AI assistant response'
            };
            const assistant = original.response?.data?.context?.assistant;
            const nextState = original.canceled ? 'canceled' : assistant?.state ?? (original.response ? 'failed' : 'incomplete');
            continuation.value = assistant?.continuation;
            patchMessage(messageId, {
                state: nextState, progress: undefined, error: readonly(normalized),
                queries: assistant?.queries ?? [], continuation: assistant?.continuation
            });
            if (!original.canceled) {
                throw normalized;
            }
        } finally {
            if (isCurrent()) {
                activeRequest.value = undefined;
            }
        }
    }

    // Chat and pagination share transcript/continuation semantics, not ownership of the draft.
    async function sendChat(message: string): Promise<void> {
        const history = buildAssistantHistory(messages.value);
        const candidate = continuation.value;
        messages.value = [...messages.value, {
            id: generateRandomUUID(), role: 'user', content: message, createdAt: Date.now(), state: 'completed'
        }];
        await request({ mode: 'chat', message, history, continuation: candidate });
    }

    async function sendMessage(): Promise<void> {
        if (!canSendMessage.value) {
            return;
        }
        const message = messageInput.value; // Only historical content is shortened, never the current question.
        messageInput.value = '';
        await sendChat(message);
    }

    async function generateSummary(message: string): Promise<void> {
        if (requesting.value || !message.trim()) {
            return;
        }
        const history = buildAssistantHistory(messages.value);
        // Keep the localized summary action as the user half of this exchange.
        // The server still receives mode=summary without a new scope instruction.
        messages.value = [...messages.value, {
            id: generateRandomUUID(), role: 'user', content: message, createdAt: Date.now(), state: 'completed'
        }];
        await request({ mode: 'summary', history, continuation: continuation.value });
    }

    async function requestNextPage(message: string): Promise<void> {
        if (canRequestNextPage.value && message.trim()) {
            await sendChat(message);
        }
    }

    return {
        enabled, messages: readonly(messages), messageInput, requesting, state, progress, continuation: readonly(continuation),
        canSendMessage, canRequestNextPage, clearConversation, disposeConversation, cancelCurrentRequest,
        sendMessage, generateSummary, requestNextPage
    };
}
