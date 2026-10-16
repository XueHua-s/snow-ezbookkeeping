import { computed, ref, watch, onScopeDispose } from 'vue';
import { assistantConversationServices as api } from '@/lib/services/assistant_conversations.ts';
import type { AssistantLocalPending } from '@/lib/assistant_local_storage.ts';
import type * as DTO from '@/models/ai_assistant_conversation.ts';
import { result, statusCode, isCommandRejection, type WriteAction, type WriteInput } from './types.ts';
import { prepareAssistantDataClear } from './dataClear.ts';
import { getSessionCurrentLanguageKey, getTimeZone } from '@/lib/settings.ts';
import { getBrowserTimezoneName } from '@/lib/datetime.ts';
import type { AssistantSession } from './session.ts';
import type { AssistantConversation } from './conversation.ts';
import type { useAssistantHistory } from './history.ts';

/** Exact persisted commands and their receipts. Unknown outcomes are reconciled, never automatically replayed. */
export function useAssistantCommands(
    session: AssistantSession,
    view: AssistantConversation,
    history: ReturnType<typeof useAssistantHistory>,
    synchronize: () => Promise<void>,
) {
    const { capabilities, authenticated, local, draft, conversationId, applyLocal, setDraft, fail, error } = session;
    const { busy, detail, online, isLatestMessage, open, loadMemories } = view;
    const { loadHistory } = history;
    const submitting = ref(false);
    const delivery = ref<Record<string, 'sending' | 'confirming_delivery' | 'unsent'>>({});
    const pending = computed(() =>
        local.value.pending.filter((item) => (item.conversationId ?? '') === conversationId.value),
    );
    const activePending = computed(() =>
        pending.value.filter((item) => !item.rejected && !item.resolvedConversationId),
    );
    const writes = new Map<string, Promise<DTO.AssistantCommandResult | undefined>>();
    function reset() {
        writes.clear();
        delivery.value = {};
        submitting.value = false;
    }
    watch(session.generation, reset, { flush: 'sync' });
    onScopeDispose(reset);

    async function accept(
        receipt: DTO.AssistantCommandResult,
        item: AssistantLocalPending,
        currentRoute: () => boolean,
    ) {
        const valid = session.capture();
        const owner = session.storage;
        if (!owner || !valid()) return;
        if (item.action === 'clearAllData' && !receipt.operationCompleted) {
            delivery.value[item.clientRequestId] = 'confirming_delivery';
            return;
        }
        if (item.action === 'branch' && receipt.conversationId && receipt.status !== 'deleted') {
            const saved = await owner.resolveBranch(item.clientRequestId, receipt.conversationId);
            if (!valid()) return;
            applyLocal(saved);
            if (currentRoute() && (!conversationId.value || conversationId.value === item.sourceConversationId))
                await open(receipt.conversationId);
            return;
        }
        // A receipt describes the original command, never current deletion/version state.
        if (item.action === 'send' && receipt.conversationId && !item.conversationId) {
            const saved = await owner.migrateDraft('', receipt.conversationId);
            if (!valid()) return;
            applyLocal(saved);
        }
        const saved = await owner.removePending(item.clientRequestId);
        if (!valid()) return;
        applyLocal(saved);
        delete delivery.value[item.clientRequestId];
        if (item.action === 'delete' && item.payload?.['id']) {
            const saved = await owner.clearConversation(String(item.payload['id']));
            if (!valid()) return;
            applyLocal(saved);
            if (currentRoute() && conversationId.value === item.payload['id']) await open();
        }
        if (!valid()) return;
        if (item.action === 'clear' || item.action === 'clearAllData') {
            const identity = session.identityRevision;
            const checked = await session.checkIdentity();
            if (!checked?.valid()) return;
            if (!valid()) {
                if (identity === session.identityRevision) await synchronize();
                return;
            }
            if (currentRoute()) await open();
            if (valid()) await loadHistory();
            return;
        }
        if (receipt.status === 'deleted') {
            if (receipt.conversationId) {
                const saved = await owner.clearConversation(receipt.conversationId);
                if (!valid()) return;
                applyLocal(saved);
            }
            if (currentRoute() && conversationId.value === receipt.conversationId) await open();
            if (valid()) await loadHistory();
            return;
        }
        if (currentRoute() && receipt.conversationId && conversationId.value === (item.conversationId ?? '')) {
            if (item.action === 'retry' && receipt.status === 'accepted' && typeof item.payload?.['userMessageId'] === 'string')
                view.followLatestAnswer(item.payload['userMessageId']);
            if (receipt.conversationId === conversationId.value) {
                const next = await result(api.get({ id: receipt.conversationId, limit: 30 })).catch(failure => {
                    if (valid() && currentRoute()) throw failure;
                });
                if (!next || !valid()) return;
                if (currentRoute()) {
                    view.applyDetail(next);
                    view.startSubscription();
                }
            } else await open(receipt.conversationId);
        }
        if (!valid()) return;
        if (['update', 'delete', 'branch', 'send'].includes(item.action)) await loadHistory();
        if (!valid()) return;
        if (['saveMemory', 'deleteMemory', 'memorySettings'].includes(item.action)) await loadMemories();
    }
    async function reconcilePending() {
        const valid = session.capture();
        const currentRoute = view.capture();
        if (!valid()) return;
        for (const item of [...local.value.pending]) {
            if (!valid()) return;
            if (
                item.rejected ||
                delivery.value[item.clientRequestId] === 'sending' ||
                item.resolvedConversationId ||
                (item.action !== 'clearAllData' && item.expectedDeletionEpoch !== capabilities.value?.deletionEpoch)
            )
                continue;
            try {
                const receipt = await result(api.command(item.clientRequestId));
                if (!valid()) return;
                if (receipt.status === 'unknown') {
                    delivery.value[item.clientRequestId] = 'unsent';
                    continue;
                }
                await accept(receipt, item, currentRoute);
            } catch (failure) {
                if (!valid()) return;
                if (statusCode(failure) === 401) {
                    fail(failure);
                    return;
                }
                delivery.value[item.clientRequestId] = 'confirming_delivery';
            }
        }
    }
    async function dispatch(item: AssistantLocalPending) {
        const valid = session.capture();
        const currentRoute = view.capture();
        const owner = session.storage;
        if (!owner || !valid()) return;
        delivery.value[item.clientRequestId] = 'sending';
        try {
            // Dynamic dispatch uses the exact payload persisted before this boundary.
            const write = api[item.action as WriteAction] as (request: object) => ReturnType<typeof api.send>;
            const receipt = await result(write(item.payload!));
            if (!valid()) return;
            await accept(receipt, item, currentRoute);
            if (valid()) return receipt;
        } catch (failure) {
            if (!valid()) return;
            if (isCommandRejection(failure)) {
                const rejected = await owner.markPendingRejected(item.clientRequestId);
                if (!valid()) return;
                applyLocal(rejected);
                delivery.value[item.clientRequestId] = 'unsent';
                await synchronize();
                if (!valid()) return;
                error.value = failure;
                throw failure;
            }
            delivery.value[item.clientRequestId] = 'confirming_delivery';
            fail(failure);
            if (statusCode(failure) === 401) throw failure;
            if (valid() && authenticated.value) await reconcilePending();
            if (valid()) throw failure;
        } finally {
            // Explicit recovery retries enter dispatch without executeOnce's persistence guard.
            if (valid() && delivery.value[item.clientRequestId] === 'sending') delete delivery.value[item.clientRequestId];
        }
        return undefined;
    }
    function execute<K extends WriteAction>(
        action: K,
        input: WriteInput<K>,
        localText?: string,
    ): Promise<DTO.AssistantCommandResult | undefined> {
        const key = JSON.stringify([action, input, localText]);
        const active = writes.get(key);
        if (active) return active;
        const task = executeOnce(action, input, localText).finally(() => {
            if (writes.get(key) === task) writes.delete(key);
        });
        writes.set(key, task);
        return task;
    }
    async function executeOnce<K extends WriteAction>(
        action: K,
        input: WriteInput<K>,
        localText?: string,
    ): Promise<DTO.AssistantCommandResult | undefined> {
        const owner = session.storage;
        if (session.disposed || !authenticated.value || !owner || !capabilities.value || !online.value)
            throw new Error('Assistant unavailable');
        const existing = local.value.pending.find(
            (item) =>
                !item.rejected &&
                item.action === action &&
                Object.entries(input).every(
                    ([key, value]) => JSON.stringify(item.payload?.[key]) === JSON.stringify(value),
                ),
        );
        if (existing) return retryPending(existing.clientRequestId);
        const valid = session.capture();
        const currentRoute = view.capture();
        const request = {
            ...input,
            clientRequestId: crypto.randomUUID(),
            requestIssuedAt: new Date(Date.now() + session.clockOffset).toISOString(),
            expectedDeletionEpoch: capabilities.value.deletionEpoch,
        };
        const value = input as Record<string, unknown>;
        const item: AssistantLocalPending = {
            clientRequestId: request.clientRequestId,
            requestIssuedAt: request.requestIssuedAt,
            expectedDeletionEpoch: request.expectedDeletionEpoch,
            action,
            text:
                localText ??
                (typeof value['content'] === 'string'
                    ? value['content']
                    : typeof value['text'] === 'string'
                      ? value['text']
                      : undefined),
            conversationId:
                typeof value['conversationId'] === 'string'
                    ? value['conversationId']
                    : ['update', 'delete'].includes(action) && typeof value['id'] === 'string'
                      ? value['id']
                      : undefined,
            sourceConversationId:
                typeof value['sourceConversationId'] === 'string' ? value['sourceConversationId'] : undefined,
            payload: request as AssistantLocalPending['payload'],
        };
        // Publish delivery intent before persistence can notify subscribers about this command.
        delivery.value[item.clientRequestId] = 'sending';
        try {
            const saved = await owner.savePending(item).catch((failure) => {
                if (valid()) throw failure;
            });
            if (!saved || !valid()) return;
            applyLocal(saved);
            if (action === 'send' && currentRoute() && draft.value === value['content']) await setDraft('');
            if (!valid()) return;
            return await dispatch(item);
        } finally {
            // Persistence failure and invalidation must not leave a recoverable command hidden.
            if (valid() && delivery.value[item.clientRequestId] === 'sending') delete delivery.value[item.clientRequestId];
        }
    }
    function prepareDataClear(newOperation = false) {
        return prepareAssistantDataClear(
            {
                state: () => ({
                    storage: session.storage,
                    capabilities: capabilities.value,
                    local: local.value,
                    epoch: session.identityRevision,
                    clockOffset: session.clockOffset,
                }),
                checkIdentity: async () => !!(await session.checkIdentity()),
                reconcilePending,
                applyLocal,
            },
            newOperation,
        );
    }
    function recoverBranchNavigation(requestedId: string) {
        return local.value.pending.find(
            (item) =>
                item.resolvedConversationId &&
                (!requestedId ||
                    item.sourceConversationId === requestedId ||
                    item.resolvedConversationId === requestedId),
        )?.resolvedConversationId;
    }
    async function acknowledgeBranchNavigation(id: string) {
        const owner = session.storage;
        const valid = view.capture();
        if (
            !owner ||
            !valid() ||
            local.value.scope !== 'device' ||
            conversationId.value !== id ||
            detail.value?.conversation.id !== id
        )
            return;
        for (const item of local.value.pending.filter((pending) => pending.resolvedConversationId === id)) {
            if (!valid()) return;
            const saved = await owner.removePending(item.clientRequestId);
            if (!valid()) return;
            applyLocal(saved);
        }
    }
    async function retryPending(id: string) {
        const valid = session.capture();
        await synchronize();
        if (!valid()) return;
        const item = local.value.pending.find((pending) => pending.clientRequestId === id);
        if (
            !item ||
            item.rejected ||
            item.resolvedConversationId ||
            item.action === 'clearAllData' ||
            delivery.value[id] !== 'unsent' ||
            item.expectedDeletionEpoch !== capabilities.value?.deletionEpoch
        )
            return;
        return dispatch(item);
    }
    async function discardRejected(id: string) {
        const owner = session.storage;
        const valid = session.capture();
        if (!owner || !valid() || !local.value.pending.some((item) => item.clientRequestId === id && item.rejected))
            return;
        const saved = await owner.removePending(id);
        if (valid()) applyLocal(saved);
    }
    async function restorePending(id: string) {
        const valid = view.capture();
        const item = local.value.pending.find((pending) => pending.clientRequestId === id);
        if (!valid() || !item?.text || draft.value.trim()) return false;
        await setDraft(item.text);
        return valid();
    }
    async function send(
        options: {
            mode?: string;
            content?: string;
            sourceMessageId?: string;
            restartQuery?: boolean;
            locale?: string;
            timezone?: string;
        } = {},
    ) {
        const content = options.content ?? draft.value;
        const limits = capabilities.value?.limits;
        if (
            !content.trim() ||
            [...content].length > (limits?.inputCharacters ?? 4000) ||
            new TextEncoder().encode(content).byteLength > (limits?.inputBytes ?? 16384)
        )
            throw new Error('input_too_long');
        if (!capabilities.value?.generationAvailable) throw new Error('generation_unavailable');
        // Repeated clicks and a delivery receipt still in flight are not active model runs.
        // Keep the draft and let reconciliation finish without dispatching a duplicate command.
        if (submitting.value || activePending.value.length) return undefined;
        if (busy.value) throw new Error('already_running');
        const valid = session.capture();
        submitting.value = true;
        try {
            return await execute('send', {
                conversationId: conversationId.value || undefined,
                expectedContextVersion: detail.value?.conversation.contextVersion ?? 0,
                mode: options.mode ?? 'chat',
                content,
                sourceMessageId: options.sourceMessageId,
                restartQuery: options.restartQuery,
                // Capture the user's date context in the persisted command, so a
                // delivery retry cannot silently change its calendar boundaries.
                locale: options.locale || getSessionCurrentLanguageKey() || 'en',
                timezone: options.timezone || getTimeZone().trim() || getBrowserTimezoneName(),
            });
        } finally {
            if (valid()) submitting.value = false;
        }
    }
    async function retry(userMessageId: string, intent = 'retry') {
        if (busy.value || activePending.value.length || !detail.value || !isLatestMessage(userMessageId)) return;
        return execute('retry', {
            conversationId: conversationId.value,
            userMessageId,
            expectedContextVersion: detail.value.conversation.contextVersion,
            intent,
        });
    }
    async function branch(sourceMessageId: string, kind: string, text?: string) {
        if (!detail.value || busy.value) return;
        // Creation and first send are separate explicit commands; the target survives first-send failure.
        return execute(
            'branch',
            {
                sourceConversationId: conversationId.value,
                sourceMessageId,
                expectedContextVersion: detail.value.conversation.contextVersion,
                kind,
            },
            text,
        );
    }
    async function stop() {
        const runId = detail.value?.activeRun?.runId;
        if (runId && busy.value) return execute('cancel', { runId });
        return undefined;
    }
    async function selectVersion(userMessageId: string, assistantMessageId: string) {
        const current = detail.value;
        if (!current || !isLatestMessage(userMessageId) || busy.value) return;
        return execute('selectVersion', {
            conversationId: conversationId.value,
            userMessageId,
            assistantMessageId,
            expectedContextVersion: current.conversation.contextVersion,
        });
    }
    return {
        submitting,
        delivery,
        pending,
        activePending,
        execute,
        send,
        retry,
        branch,
        stop,
        selectVersion,
        prepareDataClear,
        recoverBranchNavigation,
        acknowledgeBranchNavigation,
        retryPending,
        discardRejected,
        restorePending,
        reconcilePending,
    };
}
