import { computed, shallowRef, watch } from 'vue';
import { useAssistantWorkspaceScope } from './useAssistantWorkspaceScope.ts';
import { useAssistantStore } from '@/stores/assistant.ts';
import type {
    AssistantConversationDTO,
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
    const archived = shallowRef(false);
    const query = shallowRef('');
    const historySearching = shallowRef(false);
    const includeVersions = shallowRef(false);
    const editing = shallowRef<AssistantMessageDTO>();
    const editText = shallowRef('');
    const retryQuerySource = shallowRef('');
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
        }
    );
    watch(
        [() => store.capabilities?.ownerScope, () => store.capabilities?.deletionEpoch],
        () => {
            historyOpen.value = false;
            query.value = '';
            includeVersions.value = false;
            archived.value = false;
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
    async function removeLocalDraft(id: string) {
        const authorized = await requestConfirmation(tt('Delete this local draft?'));
        if (authorized()) await run(() => store.removeDraft(id));
    }
    return {
        store,
        tt,
        captureScope,
        removeLocalDraft,
        error,
        working,
        historyOpen,
        archived,
        query,
        historySearching,
        includeVersions,
        editing,
        input,
        retryQuerySource,
        restartQuery,
        messages,
        disabledReason,
        run,
        refreshHistory,
        rename,
        archive,
        remove,
        edit,
        send,
        branch,
        continueConversation,
        retry,
        nextPage,
        version
    };
}
