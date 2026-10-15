import { computed, ref } from 'vue';
import { useAssistantStore } from '@/stores/assistant.ts';
import type { AssistantConversationDTO, AssistantMemoryDTO, AssistantMessageDTO } from '@/models/ai_assistant_conversation.ts';
import { assistantReasonLabels } from '@/components/common/assistant/assistantPresentation.ts';
import type { AssistantReasonCode } from '@/models/ai_assistant_conversation.ts';
import { useI18n } from '@/locales/helpers.ts';

/** Product commands shared by both platform shells. Persistence and reconciliation belong to the store. */
export function useAssistantCloudWorkspace(confirm: (message: string) => Promise<boolean> = async message => window.confirm(message)) {
    const store = useAssistantStore();
    const { tt } = useI18n();
    const error = ref('');
    const working = ref(false);
    const historyOpen = ref(false);
    const memoryOpen = ref(false);
    const archived = ref(false);
    const query = ref('');
    const includeVersions = ref(false);
    const editing = ref<AssistantMessageDTO>();
    const editText = ref('');
    const summaryOpen = ref(false);
    const retryQuerySource = ref('');
    const summaryPeriod = ref(tt('This month'));
    const messages = computed(() => (store.detail?.messages ?? []).map(message => ({ ...message,
        selectedAnswer: store.viewedVersions[message.id] ?? message.selectedAnswer })));
    const input = computed({ get: () => editing.value ? editText.value : store.draft,
        set: value => { if (editing.value) editText.value = value; else void store.setDraft(value); } });
    const disabledReason = computed(() => !store.online ? 'Offline. Drafts are saved on this device.' :
        !store.capabilities?.generationAvailable ? (assistantReasonLabels[store.capabilities?.unavailableReason as AssistantReasonCode] ?? 'AI assistant is disabled') :
        store.conversationId && !store.detail ? 'This conversation does not exist or is not accessible.' :
        store.detail?.conversation.archived ? 'Unarchive this conversation to continue.' :
        store.activePending.length ? 'Confirming delivery' : undefined);
    async function run(action: () => Promise<unknown>) {
        error.value = ''; working.value = true;
        try { return await action(); }
        catch (failure) {
            const value = failure as { message?: string; context?: { reasonCode?: string }; response?: { data?: { context?: { reasonCode?: string } } } };
            const code = value?.response?.data?.context?.reasonCode ?? value?.context?.reasonCode ?? value?.message;
            error.value = tt(assistantReasonLabels[code as AssistantReasonCode] ?? 'Assistant request failed');
        }
        finally { working.value = false; }
        return undefined;
    }
    async function refreshHistory(cursor?: string) {
        await run(() => query.value ? store.search(query.value, cursor, includeVersions.value, archived.value) : store.loadHistory(cursor, archived.value));
    }
    function search(text: string, versions: boolean) { query.value = text; includeVersions.value = versions; void refreshHistory(); }
    function rename(item: AssistantConversationDTO, title: string) {
        return run(() => store.execute('update', { id: item.id, expectedMetadataVersion: item.metadataVersion, title }));
    }
    function archive(item: AssistantConversationDTO) {
        return run(() => store.execute('update', { id: item.id, expectedMetadataVersion: item.metadataVersion, archived: !item.archived }));
    }
    async function remove(item: AssistantConversationDTO) {
        if (await confirm(tt('Delete this conversation? Its tasks will stop and related summaries and memories will be invalidated. Backups follow the server retention policy.'))) {
            return run(() => store.execute('delete', { id: item.id, expectedMetadataVersion: item.metadataVersion }));
        }
        return undefined;
    }
    async function clear() {
        if (await confirm(tt('Clear all assistant data, including conversations, archives, memories and local drafts? Your bills will not be deleted.'))) return run(() => store.execute('clear', {}));
        return undefined;
    }
    function edit(message: AssistantMessageDTO) { if (store.busy) return; editing.value = message; editText.value = message.content; }
    async function send() {
        if (editing.value) {
            const target = editing.value;
            const text = editText.value;
            await run(async () => { const receipt = await store.branch(target.id, 'edit', text); if (receipt) editing.value = undefined; });
        } else await run(() => store.send());
    }
    function branch(message: AssistantMessageDTO) { return run(() => store.branch(message.id, 'regenerate', message.content)); }
    function continueConversation() {
        return run(async () => { await store.loadLatest(); const last = store.detail?.messages.at(-1); if (last) await store.branch(last.id, 'continue'); });
    }
    function retry(message: AssistantMessageDTO, intent = 'retry') { return store.isLatestMessage(message.id) ? run(() => store.retry(message.id, intent)) : branch(message); }
    async function nextPage(message: AssistantMessageDTO) {
        const answer = message.selectedAnswer;
        retryQuerySource.value = answer?.assistantMessageId ?? '';
        if (!answer) return;
        if (!store.isLatestMessage(message.id)) {
            if (!await confirm(tt('Continue this query in a new conversation?'))) return;
            await run(() => store.branch(answer.assistantMessageId, 'continue'));
        } else await run(() => store.send({ mode: 'next_page', content: tt('Show the next page of the current query results'), sourceMessageId: answer.assistantMessageId }));
    }
    function restartQuery() { return run(() => store.send({ mode: 'next_page', content: tt('Query these bills again'), sourceMessageId: retryQuerySource.value, restartQuery: true })); }
    async function version(message: AssistantMessageDTO, id: string) {
        await run(async () => {
            await store.viewVersion(id);
            const answer = store.viewedVersions[message.id];
            if (store.isLatestMessage(message.id) && answer?.state === 'completed') await store.selectVersion(message.id, id);
        });
    }
    function summary() { return run(async () => { await store.send({ mode: 'summary', content: `${tt('Generate AI Summary')}: ${summaryPeriod.value}` }); summaryOpen.value = false; }); }
    function showMemories() { memoryOpen.value = true; return run(() => store.loadMemories()); }
    function saveMemory(item: AssistantMemoryDTO | undefined, text: string) {
        return run(() => store.execute('saveMemory', { id: item?.id, expectedVersion: item?.version ?? 0,
            expectedMemoryUseEpoch: store.memories?.memoryUseEpoch ?? 0, text,
            key: item?.key ?? crypto.randomUUID(), type: item?.type ?? 'preference', intent: item?.status === 'candidate' ? 'confirm' : item ? 'correct' : 'add' }));
    }
    async function forget(item?: AssistantMemoryDTO) {
        if (!await confirm(tt(item ? 'Forget this memory? Original conversations remain.' : 'Forget all memories? Original conversations remain.'))) return;
        return run(() => store.execute('deleteMemory', { id: item?.id, all: !item, expectedVersion: item?.version ?? 0,
            expectedMemoryUseEpoch: store.memories?.memoryUseEpoch ?? 0 }));
    }
    return { store, tt, error, working, historyOpen, memoryOpen, archived, query, includeVersions, editing, input, summaryOpen,
        summaryPeriod, retryQuerySource, restartQuery, messages, disabledReason, run, refreshHistory, search, rename, archive, remove, clear, edit, send, branch,
        continueConversation, retry, nextPage, version, summary, showMemories, saveMemory, forget };
}
