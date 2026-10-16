import { computed, shallowRef, watch } from 'vue';
import { useAssistantWorkspaceScope } from './useAssistantWorkspaceScope.ts';
import { useAssistantStore } from '@/stores/assistant.ts';
import type {
    AssistantConversationDTO,
    AssistantMemoryDTO,
    AssistantMessageDTO
} from '@/models/ai_assistant_conversation.ts';
import { assistantReasonLabels } from '@/components/common/assistant/assistantPresentation.ts';
import type { AssistantReasonCode } from '@/models/ai_assistant_conversation.ts';
import { useI18n } from '@/locales/helpers.ts';

/** Product commands shared by both platform shells. Persistence and reconciliation belong to the store. */
export function useAssistantCloudWorkspace(
    confirm: (message: string) => Promise<boolean> = async (message) => window.confirm(message),
    routeKey: () => unknown = () => undefined
) {
    const store = useAssistantStore();
    const { tt } = useI18n();
    const error = shallowRef('');
    const working = shallowRef(false);
    const historyOpen = shallowRef(false);
    const memoryOpen = shallowRef(false);
    const archived = shallowRef(false);
    const query = shallowRef('');
    const historySearching = shallowRef(false);
    const includeVersions = shallowRef(false);
    const editing = shallowRef<AssistantMessageDTO>();
    const editText = shallowRef('');
    const summaryOpen = shallowRef(false);
    const retryQuerySource = shallowRef('');
    const summaryPeriod = shallowRef(tt('This month'));
    let pendingCount = 0;
    const captureScope = useAssistantWorkspaceScope(
        () => [store.capabilities?.ownerScope, store.capabilities?.deletionEpoch, store.conversationId, routeKey()],
        () => {
            // External route changes also dismiss the view; filters remain for its next open.
            historyOpen.value = false;
            historySearching.value = false;
            pendingCount = 0;
            working.value = false;
            error.value = '';
            editing.value = undefined;
            editText.value = '';
            retryQuerySource.value = '';
            summaryOpen.value = false;
        }
    );
    watch(
        [() => store.capabilities?.ownerScope, () => store.capabilities?.deletionEpoch],
        () => {
            historyOpen.value = false;
            memoryOpen.value = false;
            query.value = '';
            includeVersions.value = false;
            archived.value = false;
            summaryPeriod.value = tt('This month');
        },
        { flush: 'sync' }
    );
    async function requestConfirmation(message: string) {
        const current = captureScope();
        // Return a lease, not a boolean: callers recheck immediately before the command.
        return (await confirm(message)) ? current : () => false;
    }
    const messages = computed(() =>
        (store.detail?.messages ?? []).map((message) => ({
            ...message,
            selectedAnswer: store.viewedVersions[message.id] ?? message.selectedAnswer
        }))
    );
    const input = computed({
        get: () => (editing.value ? editText.value : store.draft),
        set: (value) => {
            if (editing.value) editText.value = value;
            else void store.setDraft(value);
        }
    });
    const disabledReason = computed(() =>
        !store.online
            ? 'Offline. Drafts are saved on this device.'
            : !store.capabilities?.generationAvailable
              ? (assistantReasonLabels[store.capabilities?.unavailableReason as AssistantReasonCode] ??
                'AI assistant is disabled')
              : store.conversationId && !store.detail
                ? 'This conversation does not exist or is not accessible.'
                : store.detail?.conversation.archived
                  ? 'Unarchive this conversation to continue.'
                  : store.activePending.length
                    ? 'Confirming delivery'
                    : undefined
    );
    async function run<T>(action: (current: () => boolean) => Promise<T>): Promise<T | undefined> {
        const current = captureScope();
        error.value = '';
        pendingCount++;
        working.value = true;
        try {
            const result = await action(current);
            return current() ? result : undefined;
        } catch (failure) {
            if (current()) {
                const value = failure as {
                    message?: string;
                    context?: { reasonCode?: string };
                    response?: { data?: { context?: { reasonCode?: string } } };
                };
                const code = value?.response?.data?.context?.reasonCode ?? value?.context?.reasonCode ?? value?.message;
                error.value = tt(assistantReasonLabels[code as AssistantReasonCode] ?? 'Assistant request failed');
            }
        } finally {
            if (current()) {
                pendingCount--;
                working.value = pendingCount > 0;
            }
        }
        return undefined;
    }
    async function refreshHistory(cursor?: string) {
        await run(() =>
            query.value
                ? store.search(query.value, cursor, includeVersions.value, archived.value)
                : store.loadHistory(cursor, archived.value)
        );
    }
    // Filters survive the view; reopening retries a search canceled by navigation.
    watch([query, includeVersions, historyOpen], (_, previous, onCleanup) => {
        // The store refreshes the unfiltered history when the panel becomes visible.
        if (!historyOpen.value || (!query.value && !previous[2])) return;
        const current = captureScope();
        historySearching.value = true;
        let active = true;
        const timer = setTimeout(async () => {
            if (!current()) return;
            await refreshHistory();
            if (active && current()) historySearching.value = false;
        }, 300);
        onCleanup(() => {
            active = false;
            historySearching.value = false;
            clearTimeout(timer);
        });
    });

    function rename(item: AssistantConversationDTO, title: string) {
        return run(() =>
            store.execute('update', { id: item.id, expectedMetadataVersion: item.metadataVersion, title })
        );
    }
    function archive(item: AssistantConversationDTO) {
        return run(() =>
            store.execute('update', {
                id: item.id,
                expectedMetadataVersion: item.metadataVersion,
                archived: !item.archived
            })
        );
    }
    async function remove(item: AssistantConversationDTO) {
        const authorized = await requestConfirmation(
            tt(
                'Delete this conversation? Its tasks will stop and related summaries and memories will be invalidated. Backups follow the server retention policy.'
            )
        );
        if (!authorized()) return;
        return run(() => store.execute('delete', { id: item.id, expectedMetadataVersion: item.metadataVersion }));
    }
    async function clear() {
        const authorized = await requestConfirmation(
            tt(
                'Clear all assistant data, including conversations, archives, memories and local drafts? Your bills will not be deleted.'
            )
        );
        if (!authorized()) return;
        return run(() => store.execute('clear', {}));
    }
    function edit(message: AssistantMessageDTO) {
        if (store.busy) return;
        editing.value = message;
        editText.value = message.content;
    }
    async function send() {
        if (editing.value) {
            const target = editing.value;
            const text = editText.value;
            await run(async (current) => {
                const receipt = await store.branch(target.id, 'edit', text);
                if (receipt && current()) editing.value = undefined;
            });
        } else await run(() => store.send());
    }
    function branch(message: AssistantMessageDTO) {
        return run(() => store.branch(message.id, 'regenerate', message.content));
    }
    function continueConversation() {
        return run(async (current) => {
            await store.loadLatest();
            if (!current()) return;
            const last = store.detail?.messages.at(-1);
            if (last) await store.branch(last.id, 'continue');
        });
    }
    function retry(message: AssistantMessageDTO, intent = 'retry') {
        return store.isLatestMessage(message.id) ? run(() => store.retry(message.id, intent)) : branch(message);
    }
    async function nextPage(message: AssistantMessageDTO) {
        const answer = message.selectedAnswer;
        retryQuerySource.value = answer?.assistantMessageId ?? '';
        if (!answer) return;
        if (!store.isLatestMessage(message.id)) {
            const authorized = await requestConfirmation(tt('Continue this query in a new conversation?'));
            if (!authorized()) return;
            await run(() => store.branch(answer.assistantMessageId, 'continue'));
        } else
            await run(() =>
                store.send({
                    mode: 'next_page',
                    content: tt('Show the next page of the current query results'),
                    sourceMessageId: answer.assistantMessageId
                })
            );
    }
    function restartQuery() {
        return run(() =>
            store.send({
                mode: 'next_page',
                content: tt('Query these bills again'),
                sourceMessageId: retryQuerySource.value,
                restartQuery: true
            })
        );
    }
    async function version(message: AssistantMessageDTO, id: string) {
        await run(async (current) => {
            await store.viewVersion(id);
            if (!current()) return;
            const answer = store.viewedVersions[message.id];
            if (store.isLatestMessage(message.id) && answer?.state === 'completed')
                await store.selectVersion(message.id, id);
        });
    }
    function summary() {
        return run(async (current) => {
            await store.send({ mode: 'summary', content: `${tt('Generate AI Summary')}: ${summaryPeriod.value}` });
            if (current()) summaryOpen.value = false;
        });
    }
    async function removeLocalDraft(id: string) {
        const authorized = await requestConfirmation(tt('Delete this local draft?'));
        if (authorized()) await run(() => store.removeDraft(id));
    }
    function showMemories() {
        memoryOpen.value = true;
        return run(() => store.loadMemories());
    }
    function saveMemory(item: AssistantMemoryDTO | undefined, text: string) {
        return run(() =>
            store.execute('saveMemory', {
                id: item?.id,
                expectedVersion: item?.version ?? 0,
                expectedMemoryUseEpoch: store.memories?.memoryUseEpoch ?? 0,
                text,
                key: item?.key ?? crypto.randomUUID(),
                type: item?.type ?? 'preference',
                intent: item?.status === 'candidate' ? 'confirm' : item ? 'correct' : 'add'
            })
        );
    }
    async function forget(item?: AssistantMemoryDTO) {
        const authorized = await requestConfirmation(
            tt(
                item
                    ? 'Forget this memory? Original conversations remain.'
                    : 'Forget all memories? Original conversations remain.'
            )
        );
        if (!authorized()) return;
        return run(() =>
            store.execute('deleteMemory', {
                id: item?.id,
                all: !item,
                expectedVersion: item?.version ?? 0,
                expectedMemoryUseEpoch: store.memories?.memoryUseEpoch ?? 0
            })
        );
    }
    return {
        store,
        tt,
        captureScope,
        removeLocalDraft,
        error,
        working,
        historyOpen,
        memoryOpen,
        archived,
        query,
        historySearching,
        includeVersions,
        editing,
        input,
        summaryOpen,
        summaryPeriod,
        retryQuerySource,
        restartQuery,
        messages,
        disabledReason,
        run,
        refreshHistory,
        rename,
        archive,
        remove,
        clear,
        edit,
        send,
        branch,
        continueConversation,
        retry,
        nextPage,
        version,
        summary,
        showMemories,
        saveMemory,
        forget
    };
}
