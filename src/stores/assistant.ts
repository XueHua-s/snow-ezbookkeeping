import { computed, onScopeDispose, watch } from 'vue';
import { defineStore } from 'pinia';
import { assistantConversationServices as api } from '@/lib/services/assistant_conversations.ts';
import type * as DTO from '@/models/ai_assistant_conversation.ts';
import { result, statusCode } from './assistant/types.ts';
import { createReferenceReader } from './assistant/reference.ts';
import { useAssistantHistory } from './assistant/history.ts';
import { useAssistantSession } from './assistant/session.ts';
import { useAssistantConversation } from './assistant/conversation.ts';
import { useAssistantCommands } from './assistant/commands.ts';

/** Sole cloud UI authority. Modules own identity/recovery, conversation resources and durable commands. */
export const useAssistantStore = defineStore('assistant', () => {
    const session = useAssistantSession();
    const {
        capabilities,
        conversationId,
        recentConversationId,
        legacyBackend,
        local,
        draft,
        error,
        authenticated,
        reset,
        fail,
        checkIdentity,
        restoreDraft,
        setDraft,
        chooseDraft,
        retryLocalPersistence,
        removeDraft,
        applyLocal,
    } = session;
    const historyState = useAssistantHistory(() => session.generation.value);
    const { history, searchResults, loadHistory, search } = historyState;
    const view = useAssistantConversation(session, synchronize);
    const {
        detail,
        memories,
        viewedVersions,
        loading,
        updatesAvailable,
        windowMissingLatest,
        online,
        busy,
        open,
        loadOlder,
        mergeSnapshot,
        attach,
        detach,
        viewVersion,
        isLatestMessage,
        loadMemories,
        loadVersions,
        savePosition,
        loadLatest,
        readingPosition,
    } = view;
    const commands = useAssistantCommands(session, view, historyState, synchronize);
    const {
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
    } = commands;
    const unsaved = computed(() => local.value.scope === 'page' && (!!draft.value || local.value.pending.length > 0));
    const canSend = computed(
        () =>
            authenticated.value &&
            online.value &&
            capabilities.value?.generationAvailable &&
            !detail.value?.conversation.archived &&
            !busy.value &&
            !submitting.value &&
            !activePending.value.length &&
            !!draft.value.trim(),
    );
    let syncPromise: Promise<void> | undefined;
    watch(
        session.generation,
        () => {
            historyState.reset();
            syncPromise = undefined;
        },
        { flush: 'sync' },
    );
    async function initialize() {
        const valid = session.capture();
        if (!valid()) return;
        try {
            const checked = await checkIdentity();
            if (!checked?.valid()) return;
            if (!valid()) {
                await initialize();
                return;
            }
            restoreDraft();
            await reconcilePending();
            if (valid()) await loadHistory();
            if (valid()) view.schedule();
        } catch (failure) {
            if (!valid()) return;
            fail(failure);
        }
    }
    const getReferenceTransaction = createReferenceReader(() => session.generation.value, fail);
    async function synchronize() {
        if (syncPromise) return syncPromise;
        const valid = session.capture();
        const currentRoute = view.capture();
        if (!valid()) return;
        const task = (async () => {
            try {
                if (!online.value) return;
                const checked = await checkIdentity();
                if (!checked?.valid()) return;
                if (!valid()) {
                    // Identity discovery authorizes a fresh read generation, not any old command.
                    // The returned lease also fences microtasks between helper and caller.
                    await synchronize();
                    return;
                }
                await reconcilePending();
                if (!valid() || !currentRoute()) return;
                const id = conversationId.value;
                if (id) {
                    let state: DTO.AssistantConversationState | undefined;
                    try {
                        state = await result(api.state(id));
                    } catch (failure) {
                        if (!currentRoute()) return;
                        if (statusCode(failure) !== 404) throw failure;
                    }
                    if (!currentRoute()) return;
                    if (!state?.accessible) {
                        const owner = session.storage;
                        if (owner) {
                            const saved = await owner.clearConversation(id);
                            if (!currentRoute() || owner !== session.storage) return;
                            applyLocal(saved);
                        }
                        if (currentRoute()) await open();
                        return;
                    }
                    if (
                        !detail.value ||
                        state.contextVersion !== detail.value.conversation.contextVersion ||
                        state.metadataVersion !== detail.value.conversation.metadataVersion ||
                        state.activeRunId !== detail.value.conversation.activeRunId
                    ) {
                        const next = await result(api.get({ id, limit: 30 }));
                        if (!currentRoute()) return;
                        view.applyDetail(next);
                    } else if (state.activeRunId && state.activeRunRevision > (detail.value.activeRun?.revision ?? 0)) {
                        const snapshot = await result(api.run({ runId: state.activeRunId }));
                        if (!currentRoute()) return;
                        mergeSnapshot(snapshot);
                    }
                }
                if (!currentRoute()) return;
                await historyState.refreshVisible();
                if (!currentRoute()) return;
                view.startSubscription();
                error.value = undefined;
            } catch (failure) {
                if (currentRoute()) fail(failure);
            }
        })().finally(() => {
            if (syncPromise === task) syncPromise = undefined;
            if (valid()) view.schedule();
        });
        syncPromise = task;
        return task;
    }
    async function exportData(request: DTO.AssistantExportRequest, signal?: AbortSignal) {
        const valid = session.capture();
        if (!valid()) throw new Error('Assistant identity changed');
        const response = await api.export(request, signal);
        if (!valid()) throw new Error('Assistant identity changed');
        return response.data;
    }
    function setHistoryVisible(value: boolean) {
        const valid = session.capture();
        if (valid())
            void historyState.setVisible(value).catch((failure) => {
                if (valid()) fail(failure);
            });
    }
    function pageShow(event: PageTransitionEvent) {
        if (!event.persisted) {
            view.foreground();
            return;
        }
        const id = conversationId.value;
        reset(false);
        const valid = session.capture();
        void initialize()
            .then(() => (valid() && authenticated.value ? open(id) : undefined))
            .finally(() => {
                if (valid()) view.schedule();
            });
    }
    function identityStorage(event: StorageEvent) {
        if (event.key === 'ebk_user_token' || event.key === 'ebk_user_info') {
            reset(false);
            view.foreground();
        }
    }
    function beforeUnload(event: BeforeUnloadEvent) {
        if (unsaved.value) {
            event.preventDefault();
            event.returnValue = '';
        }
    }
    if (typeof window !== 'undefined') {
        const clear = () => reset();
        const hide = () => reset(false);
        window.addEventListener('assistant-identity-cleared', clear);
        window.addEventListener('assistant-identity-hidden', hide);
        window.addEventListener('pageshow', pageShow);
        window.addEventListener('storage', identityStorage);
        window.addEventListener('beforeunload', beforeUnload);
        onScopeDispose(() => {
            window.removeEventListener('assistant-identity-cleared', clear);
            window.removeEventListener('assistant-identity-hidden', hide);
            window.removeEventListener('pageshow', pageShow);
            window.removeEventListener('storage', identityStorage);
            window.removeEventListener('beforeunload', beforeUnload);
        });
    }
    onScopeDispose(() => {
        session.dispose();
        view.dispose();
        historyState.dispose();
    });
    return {
        getReferenceTransaction,
        capabilities,
        conversationId,
        recentConversationId,
        legacyBackend,
        updatesAvailable,
        windowMissingLatest,
        loadLatest,
        exportData,
        loadVersions,
        detail,
        history,
        searchResults,
        memories,
        local,
        draft,
        error,
        loading,
        authenticated,
        online,
        submitting,
        activePending,
        discardRejected,
        retryLocalPersistence,
        removeDraft,
        delivery,
        viewedVersions,
        busy,
        pending,
        unsaved,
        canSend,
        initialize,
        reset,
        open,
        setDraft,
        chooseDraft,
        loadHistory,
        search,
        loadOlder,
        isLatestMessage,
        mergeSnapshot,
        synchronize,
        attach,
        detach,
        execute,
        send,
        retry,
        branch,
        stop,
        viewVersion,
        selectVersion,
        loadMemories,
        prepareDataClear,
        recoverBranchNavigation,
        acknowledgeBranchNavigation,
        retryPending,
        restorePending,
        reconcilePending,
        savePosition,
        readingPosition,
        setHistoryVisible,
    };
});
