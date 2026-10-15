import { computed, ref, shallowRef, onScopeDispose } from 'vue';
import { defineStore } from 'pinia';
import { assistantConversationServices as api } from '@/lib/services/assistant_conversations.ts';
import { createAssistantLocalStorage, type AssistantLocalSnapshot, type AssistantLocalPending } from '@/lib/assistant_local_storage.ts';
import type * as DTO from '@/models/ai_assistant_conversation.ts';
import { result, terminal, statusCode, isCommandRejection, type ReadingPosition, type WriteAction, type WriteInput } from './assistant/types.ts';
import { createReferenceReader } from './assistant/reference.ts';
import { prepareAssistantDataClear } from './assistant/dataClear.ts';
/** Sole authority for cloud assistant UI; browser drafts never become an automatic send queue. */
export const useAssistantStore = defineStore('assistant', () => {
    const capabilities = shallowRef<DTO.AssistantCapabilities>();
    const conversationId = ref(''), recentConversationId = ref('');
    const legacyBackend = ref(false); const updatesAvailable = ref(false), windowMissingLatest = ref(false);
    const detail = shallowRef<DTO.AssistantConversationDetail>();
    const history = shallowRef<DTO.AssistantConversationList>({ items: [] });
    const searchResults = shallowRef<DTO.AssistantSearchResult>({ items: [] });
    const memories = shallowRef<DTO.AssistantMemoryList>();
    const local = shallowRef<AssistantLocalSnapshot>({ drafts: [], pending: [], scope: 'page', version: 0 });
    const draft = ref('');
    const error = shallowRef<unknown>();
    const loading = ref(false), submitting = ref(false);
    const authenticated = ref(false);
    const online = ref(typeof navigator === 'undefined' || navigator.onLine);
    const delivery = ref<Record<string, 'sending' | 'confirming_delivery' | 'unsent'>>({});
    const viewedVersions = shallowRef<Record<string, DTO.AssistantRunSnapshot>>({});
    const positions = new Map<string, ReadingPosition>();
    const writes = new Map<string, Promise<DTO.AssistantCommandResult | undefined>>();
    let storage: ReturnType<typeof createAssistantLocalStorage> | undefined;
    let epoch = 0, navigation = 0;
    let historyTicket = 0, searchTicket = 0;
    let olderWindow = false;
    let historyArchived = false;
    let attached = 0;
    let timer: ReturnType<typeof setTimeout> | undefined;
    let subscription: AbortController | undefined; let subscriptionRun = '';
    let syncPromise: Promise<void> | undefined;
    let localUnsubscribe: (() => void) | undefined;
    let draftId: string | undefined;
    let draftRevision: number | undefined;
    let historyVisible = false;
    let lastHistoryRefresh = 0;
    let backoff = 1000, pollUntil = 0;
    let streamFailures = 0;
    let clockOffset = 0;
    const busy = computed(() => !!detail.value?.activeRun && !terminal(detail.value.activeRun.state));
    const pending = computed(() => local.value.pending.filter(p => (p.conversationId ?? '') === conversationId.value));
    const activePending = computed(() => pending.value.filter(item => !item.rejected && !item.resolvedConversationId));
    const unsaved = computed(() => local.value.scope === 'page' && (!!draft.value || local.value.pending.length > 0));
    const canSend = computed(() => authenticated.value && online.value && capabilities.value?.generationAvailable && !detail.value?.conversation.archived && !busy.value &&
        !submitting.value && !activePending.value.length && !!draft.value.trim());
    function applyLocal(value: AssistantLocalSnapshot) { local.value = value; }
    function stopSubscription() { subscription?.abort(); subscription = undefined; subscriptionRun = ''; }
    function restoreDraft() {
        const choices = local.value.drafts.filter(d => d.conversationId === conversationId.value);
        const found = choices.find(d => d.editorId === storage?.editorId) ?? choices[0];
        draft.value = found?.text ?? ''; draftId = found?.id; draftRevision = found?.revision;
    }
    function reset(clearLocal = true) {
        epoch++; navigation++; windowMissingLatest.value = false; historyArchived = false; writes.clear(); authenticated.value = false; stopSubscription();
        if (timer) clearTimeout(timer);
        timer = undefined; syncPromise = undefined;
        if (clearLocal && capabilities.value) { try { localStorage.removeItem('assistant-recent:' + capabilities.value.ownerScope); } catch { /* ID-only hint is optional. */ } }
        const old = storage; storage = undefined; localUnsubscribe?.();
        if (old) { if (clearLocal) void old.clearOwner().finally(() => old.close()); else old.close(); }
        capabilities.value = undefined; conversationId.value = ''; recentConversationId.value = ''; legacyBackend.value = false; detail.value = undefined;
        history.value = { items: [] }; searchResults.value = { items: [] }; memories.value = undefined;
        local.value = { drafts: [], pending: [], scope: 'page', version: 0 };
        submitting.value = false; draft.value = ''; draftId = undefined; draftRevision = undefined; delivery.value = {}; viewedVersions.value = {}; positions.clear();
    }
    async function checkIdentity() {
        const generation = epoch;
        const next = await result(api.capabilities());
        if (generation !== epoch) return false;
        if (capabilities.value && capabilities.value.ownerScope !== next.ownerScope) reset();
        if (!storage) {
            storage = createAssistantLocalStorage({ ownerScope: next.ownerScope });
            localUnsubscribe = storage.subscribe(() => { void refreshLocal(); });
        }
        const owner = storage;
        const currentEpoch = epoch;
        const changed = capabilities.value && next.deletionEpoch > capabilities.value.deletionEpoch;
        const loaded = await owner.reconcileEpoch(next.deletionEpoch);
        if (currentEpoch !== epoch || owner !== storage) return false;
        if (changed) { detail.value = undefined; conversationId.value = ''; draft.value = ''; viewedVersions.value = {}; positions.clear(); stopSubscription(); }
        const serverTime = Date.parse(next.serverTime);
        if (Number.isFinite(serverTime)) clockOffset = serverTime - Date.now();
        capabilities.value = next; applyLocal(loaded); authenticated.value = true; legacyBackend.value = false;
        try { recentConversationId.value = localStorage.getItem('assistant-recent:' + next.ownerScope) ?? ''; } catch { /* Optional ID-only navigation hint. */ }
        return true;
    }
    async function refreshLocal() {
        const owner = storage; const generation = epoch;
        if (!owner) return;
        const next = await owner.load();
        if (generation === epoch && owner === storage) applyLocal(next);
    }
    async function initialize() {
        try { if (await checkIdentity()) { restoreDraft(); await reconcilePending(); await loadHistory(); } }
        catch (failure) { legacyBackend.value = statusCode(failure) === 404; fail(failure); }
    }
    const getReferenceTransaction = createReferenceReader(() => epoch, fail);
    function fail(failure: unknown) { if (statusCode(failure) === 401) reset(false); error.value = failure; }
    async function setDraft(text: string) {
        draft.value = text;
        const owner = storage; const generation = epoch; const id = conversationId.value;
        if (!owner || !capabilities.value) return;
        const saved = await owner.saveDraft({ conversationId: id, text, expectedDeletionEpoch: capabilities.value.deletionEpoch,
            id: draftId, revision: draftRevision }).catch(failure => { if (generation === epoch) fail(failure); return local.value; });
        if (generation !== epoch || owner !== storage) return;
        applyLocal(saved);
        if (conversationId.value === id && draft.value === text) {
            const own = saved.drafts.find(d => d.conversationId === id && d.editorId === owner.editorId);
            draftId = own?.id; draftRevision = own?.revision;
        }
    }
    async function retryLocalPersistence() {
        const owner = storage; const generation = epoch;
        if (!owner) return;
        const saved = await owner.retryPersistence();
        if (generation === epoch && owner === storage) applyLocal(saved);
    }
    async function removeDraft(id: string) {
        const owner = storage; const generation = epoch;
        if (!owner) return;
        const saved = await owner.removeDraft(id);
        if (generation === epoch && owner === storage) applyLocal(saved);
    }
    async function chooseDraft(id: string) {
        const selected = local.value.drafts.find(d => d.id === id && d.conversationId === conversationId.value);
        if (!selected) return;
        draftId = selected.id; draftRevision = selected.revision; draft.value = selected.text;
    }
    async function open(id = '', aroundMessageId?: string, versionId?: string) {
        const ticket = ++navigation; const generation = epoch;
        stopSubscription(); windowMissingLatest.value = false; olderWindow = false; updatesAvailable.value = false; conversationId.value = id; detail.value = undefined; viewedVersions.value = {}; restoreDraft();
        if (!id) return;
        loading.value = true; error.value = undefined;
        try {
            const next = await result(api.get({ id, aroundMessageId, limit: 30 }));
            if (ticket !== navigation || generation !== epoch) return;
            detail.value = next; recentConversationId.value = id;
            try { localStorage.setItem('assistant-recent:' + capabilities.value?.ownerScope, id); } catch { /* Navigation still works without storage. */ }
            if (versionId) await viewVersion(versionId);
            startSubscription();
        } catch (failure) {
            if (ticket !== navigation || generation !== epoch) return;
            if (statusCode(failure) === 404 && storage) { const owner = storage; const saved = await owner.clearConversation(id);
                if (ticket !== navigation || generation !== epoch || owner !== storage) return;
                applyLocal(saved); conversationId.value = ''; restoreDraft(); }
            fail(failure);
        }
        finally { if (ticket === navigation) loading.value = false; }
    }
    async function loadHistory(cursor?: string, archived = historyArchived) {
        const generation = epoch; const ticket = ++historyTicket; historyArchived = archived;
        const next = await result(api.list({ cursor, archived, limit: 30 }));
        if (generation !== epoch || ticket !== historyTicket) return;
        // The sidebar window is bounded; older pages remain available through explicit pagination.
        history.value = cursor ? { ...next, items: [...history.value.items, ...next.items].filter((item, index, items) => items.findIndex(other => other.id === item.id) === index).slice(-90) } : next; lastHistoryRefresh = Date.now();
    }
    async function search(query: string, cursor?: string, includeVersions = false, archived = false) {
        const generation = epoch; const ticket = ++searchTicket;
        const next = await result(api.search({ query, cursor, includeVersions, archived, limit: 30 }));
        if (generation === epoch && ticket === searchTicket) searchResults.value = cursor ? { ...next, items: [...searchResults.value.items, ...next.items].filter((item, index, items) => items.findIndex(other => other.messageId === item.messageId && other.versionId === item.versionId) === index).slice(-90) } : next;
    }
    async function loadOlder() {
        const current = detail.value; const generation = epoch; const ticket = navigation;
        if (!current?.hasMore) return;
        const next = await result(api.get({ id: current.conversation.id,
            beforeTurnSequence: current.messages[0]?.turnSequence, limit: 30 }));
        if (generation !== epoch || ticket !== navigation) return;
        olderWindow = true;
        const combined = [...next.messages, ...current.messages].filter((item, index, items) => items.findIndex(m => m.id === item.id) === index);
        windowMissingLatest.value ||= combined.length > 90;
        detail.value = { ...current, messages: combined.slice(0, 90), hasMore: next.hasMore, nextCursor: next.nextCursor };
    }
    function applyDetail(next: DTO.AssistantConversationDetail) {
        const previous = detail.value;
        if (!previous || previous.conversation.id !== next.conversation.id) { detail.value = next; return; }
        if (next.conversation.contextVersion < previous.conversation.contextVersion || next.conversation.metadataVersion < previous.conversation.metadataVersion) return;
        const retain = (fresh: DTO.AssistantRunSnapshot | undefined, old: DTO.AssistantRunSnapshot | undefined) =>
            fresh && old && fresh.runId === old.runId && (fresh.revision < old.revision || terminal(old.state)) ? old : fresh;
        if (olderWindow) { updatesAvailable.value = updatesAvailable.value || next.conversation.contextVersion !== previous.conversation.contextVersion;
            next = { ...next, messages: previous.messages.map(old => next.messages.find(fresh => fresh.id === old.id) ?? old), hasMore: previous.hasMore, nextCursor: previous.nextCursor }; }
        detail.value = { ...next, activeRun: retain(next.activeRun, previous.activeRun), messages: next.messages.map(message => ({
            ...message, selectedAnswer: retain(message.selectedAnswer, previous.messages.find(old => old.id === message.id)?.selectedAnswer)
        })) };
    }
    function mergeSnapshot(snapshot: DTO.AssistantRunSnapshot) {
        const current = detail.value;
        if (!current || snapshot.conversationId !== conversationId.value) return;
        const existing = current.activeRun?.runId === snapshot.runId ? current.activeRun : current.messages.find(m => m.selectedAnswer?.runId === snapshot.runId)?.selectedAnswer;
        if (existing && (existing.revision >= snapshot.revision || terminal(existing.state))) return;
        detail.value = { ...current, activeRun: snapshot,
            messages: current.messages.map(m => m.id === snapshot.userMessageId ? { ...m,
                selectedAnswer: snapshot, versions: [...m.versions.filter(v => v.id !== snapshot.assistantMessageId),
                    { id: snapshot.assistantMessageId, attemptNumber: snapshot.attemptNumber, state: snapshot.state, createdAt: snapshot.savedAt }] } : m) };
        if (terminal(snapshot.state)) { stopSubscription(); void synchronize(); }
    }
    function startSubscription() {
        const run = detail.value?.activeRun;
        if (streamFailures >= 3 || !attached || !online.value || document.hidden || !run || terminal(run.state) || subscriptionRun === run.runId || Date.now() < pollUntil) return;
        stopSubscription(); const controller = new AbortController(); subscription = controller; subscriptionRun = run.runId;
        const generation = epoch; const ticket = navigation;
        void api.subscribe(run.runId, run.revision, { onSnapshot: snapshot => {
            if (generation === epoch && ticket === navigation && !controller.signal.aborted) { backoff = 1000; streamFailures = 0; mergeSnapshot(snapshot); }
        } }, controller.signal).catch(failure => {
            if (controller.signal.aborted || generation !== epoch || ticket !== navigation) return;
            if (statusCode(failure) === 401) { fail(failure); return; }
            streamFailures++;
            const retry = (failure as { retryAfterMs?: number }).retryAfterMs ?? backoff;
            pollUntil = Date.now() + Math.max(1000, Math.min(retry, 60000)); backoff = Math.min(backoff * 2, 15000);
        }).finally(() => { if (subscription === controller) { subscription = undefined; subscriptionRun = ''; } });
    }
    async function synchronize() {
        if (syncPromise) return syncPromise;
        const syncEpoch = epoch;
        syncPromise = (async () => {
            try {
                if (!online.value || !await checkIdentity()) return;
                await reconcilePending();
                const id = conversationId.value; const generation = epoch; const ticket = navigation;
                if (id) {
                    let state: DTO.AssistantConversationState;
                    try { state = await result(api.state(id)); } catch (failure) {
                        if (statusCode(failure) === 404 && generation === epoch && ticket === navigation) {
                            const owner = storage; if (owner) { const saved = await owner.clearConversation(id); if (generation === epoch) applyLocal(saved); }
                            if (generation === epoch && ticket === navigation) await open();
                            return;
                        }
                        throw failure;
                    }
                    if (generation !== epoch || ticket !== navigation) return;
                    if (!state.accessible) {
                        if (storage) applyLocal(await storage.clearConversation(id));
                        await open();
                    } else if (!detail.value || state.contextVersion !== detail.value.conversation.contextVersion ||
                        state.metadataVersion !== detail.value.conversation.metadataVersion || state.activeRunId !== detail.value.conversation.activeRunId) {
                        const next = await result(api.get({ id, limit: 30 }));
                        if (generation === epoch && ticket === navigation) applyDetail(next);
                    } else if (state.activeRunId && state.activeRunRevision > (detail.value.activeRun?.revision ?? 0)) {
                        const snapshot = await result(api.run({ runId: state.activeRunId }));
                        if (generation === epoch && ticket === navigation) mergeSnapshot(snapshot);
                    }
                }
                if (historyVisible && Date.now() - lastHistoryRefresh >= 30000) await loadHistory(undefined, historyArchived);
                startSubscription(); error.value = undefined;
            } catch (failure) { if (syncEpoch === epoch) fail(failure); }
        })().finally(() => { if (syncEpoch === epoch) syncPromise = undefined; });
        return syncPromise;
    }
    function schedule() {
        if (timer) clearTimeout(timer);
        if (!attached || document.hidden) return;
        timer = setTimeout(() => { void synchronize().finally(schedule); }, detail.value?.activeRun && !terminal(detail.value.activeRun.state) && !subscription ? Math.max(2000, backoff) : 14000 + Math.random() * 2000);
    }
    function foreground() {
        online.value = navigator.onLine;
        if (document.hidden || !online.value) { stopSubscription(); if (timer) clearTimeout(timer); return; }
        streamFailures = 0; void synchronize().finally(schedule);
    }
    function pageShow(event: PageTransitionEvent) {
        if (event.persisted) { const id = conversationId.value; reset(false); void initialize().then(() => authenticated.value ? open(id) : undefined).finally(schedule); }
        else foreground();
    }
    function identityStorage(event: StorageEvent) {
        if (event.key === 'ebk_user_token' || event.key === 'ebk_user_info') { reset(false); foreground(); }
    }
    function beforeUnload(event: BeforeUnloadEvent) { if (unsaved.value) { event.preventDefault(); event.returnValue = ''; } }
    function attach() {
        if (++attached > 1) return;
        window.addEventListener('online', foreground); window.addEventListener('offline', foreground);
        document.addEventListener('visibilitychange', foreground); foreground();
    }
    function detach() {
        attached = Math.max(0, attached - 1); if (attached) return;
        stopSubscription(); if (timer) clearTimeout(timer);
        window.removeEventListener('online', foreground); window.removeEventListener('offline', foreground);
        document.removeEventListener('visibilitychange', foreground);
    }
    async function accept(receipt: DTO.AssistantCommandResult, item: AssistantLocalPending) {
        if (item.action === 'clearAllData' && !receipt.operationCompleted) { delivery.value[item.clientRequestId] = 'confirming_delivery'; return; }
        const owner = storage; if (!owner) return;
        const generation = epoch;
        const save = async (operation: Promise<AssistantLocalSnapshot>) => {
            const next = await operation;
            if (generation !== epoch || owner !== storage) return false;
            applyLocal(next); return true;
        };
        if (item.action === 'branch' && receipt.conversationId && receipt.status !== 'deleted') {
            if (!await save(owner.resolveBranch(item.clientRequestId, receipt.conversationId))) return;
            if (!conversationId.value || conversationId.value === item.sourceConversationId) await open(receipt.conversationId);
            return;
        }
        // A receipt describes the original command, never current deletion/version state.
        if (item.action === 'send' && receipt.conversationId && !item.conversationId) {
            if (!await save(owner.migrateDraft('', receipt.conversationId))) return;
        }
        if (generation !== epoch) return;
        if (!await save(owner.removePending(item.clientRequestId))) return;
        delete delivery.value[item.clientRequestId];
        if (item.action === 'delete' && item.payload?.['id']) {
            if (!await save(owner.clearConversation(String(item.payload['id'])))) return;
            if (conversationId.value === item.payload['id']) await open();
        }
        if (item.action === 'clear' || item.action === 'clearAllData') { await checkIdentity(); await open(); await loadHistory(); return; }
        if (receipt.status === 'deleted') {
            if (receipt.conversationId && !await save(owner.clearConversation(receipt.conversationId))) return;
            if (conversationId.value === receipt.conversationId) await open();
            return;
        }
        if (receipt.conversationId && conversationId.value === (item.conversationId ?? '')) {
            if (receipt.conversationId === conversationId.value) {
                const ticket = navigation; const next = await result(api.get({ id: receipt.conversationId, limit: 30 }));
                if (generation !== epoch || ticket !== navigation) return;
                applyDetail(next); startSubscription();
            } else await open(receipt.conversationId);
        }
        if (['update', 'delete', 'branch', 'send'].includes(item.action)) await loadHistory();
        if (['saveMemory', 'deleteMemory', 'memorySettings'].includes(item.action)) await loadMemories();
    }
    async function reconcilePending() {
        for (const item of [...local.value.pending]) {
            const generation = epoch;
            if (item.rejected || item.resolvedConversationId || (item.action !== 'clearAllData' && item.expectedDeletionEpoch !== capabilities.value?.deletionEpoch)) continue;
            try {
                const receipt = await result(api.command(item.clientRequestId));
                if (generation !== epoch) return;
                if (receipt.status === 'unknown') { delivery.value[item.clientRequestId] = 'unsent'; continue; }
                await accept(receipt, item);
            } catch (failure) {
                if (generation !== epoch) return;
                if (statusCode(failure) === 401) { fail(failure); return; }
                delivery.value[item.clientRequestId] = 'confirming_delivery';
            }
        }
    }
    async function dispatch(item: AssistantLocalPending) {
        const generation = epoch;
        delivery.value[item.clientRequestId] = 'sending';
        try {
            // All requests were fully persisted before entering dispatch. Dynamic dispatch preserves the exact original payload.
            const write = api[item.action as WriteAction] as (request: object) => ReturnType<typeof api.send>;
            const receipt = await result(write(item.payload!));
            if (generation !== epoch) return;
            await accept(receipt, item);
            return receipt;
        } catch (failure) {
            if (generation !== epoch) return;
            if (isCommandRejection(failure) && storage) {
                const rejected = await storage.markPendingRejected(item.clientRequestId);
                if (generation !== epoch) return;
                applyLocal(rejected); delivery.value[item.clientRequestId] = 'unsent'; await synchronize(); error.value = failure; throw failure;
            }
            delivery.value[item.clientRequestId] = 'confirming_delivery'; fail(failure);
            if (authenticated.value) await reconcilePending();
            throw failure;
        }
    }
    function execute<K extends WriteAction>(action: K, input: WriteInput<K>, localText?: string): Promise<DTO.AssistantCommandResult | undefined> {
        const key = JSON.stringify([action, input, localText]);
        const active = writes.get(key);
        if (active) return active;
        const task = executeOnce(action, input, localText).finally(() => { if (writes.get(key) === task) writes.delete(key); });
        writes.set(key, task); return task;
    }
    async function executeOnce<K extends WriteAction>(action: K, input: WriteInput<K>, localText?: string): Promise<DTO.AssistantCommandResult | undefined> {
        if (!authenticated.value || !storage || !capabilities.value || !online.value) throw new Error('Assistant unavailable');
        const existing = local.value.pending.find(item => !item.rejected && item.action === action && Object.entries(input).every(([key, value]) => JSON.stringify(item.payload?.[key]) === JSON.stringify(value)));
        if (existing) return retryPending(existing.clientRequestId);
        const generation = epoch; const targetNavigation = navigation; const targetConversation = conversationId.value;
        const request = { ...input, clientRequestId: crypto.randomUUID(), requestIssuedAt: new Date(Date.now() + clockOffset).toISOString(),
            expectedDeletionEpoch: capabilities.value.deletionEpoch };
        const value = input as Record<string, unknown>;
        const item: AssistantLocalPending = { clientRequestId: request.clientRequestId, requestIssuedAt: request.requestIssuedAt, expectedDeletionEpoch: request.expectedDeletionEpoch, action,
            text: localText ?? (typeof value['content'] === 'string' ? value['content'] : typeof value['text'] === 'string' ? value['text'] : undefined),
            conversationId: typeof value['conversationId'] === 'string' ? value['conversationId'] : ['update', 'delete'].includes(action) && typeof value['id'] === 'string' ? value['id'] : undefined,
            sourceConversationId: typeof value['sourceConversationId'] === 'string' ? value['sourceConversationId'] : undefined,
            payload: request as AssistantLocalPending['payload'] };
        const saved = await storage.savePending(item);
        if (generation !== epoch) return;
        applyLocal(saved);
        if (action === 'send' && targetNavigation === navigation && targetConversation === conversationId.value && draft.value === value['content']) await setDraft('');
        if (generation !== epoch) return;
        return dispatch(item);
    }
    function prepareDataClear(newOperation = false) {
        return prepareAssistantDataClear({ state: () => ({ storage, capabilities: capabilities.value, local: local.value, epoch, clockOffset }),
            checkIdentity, reconcilePending, applyLocal }, newOperation);
    }
    function recoverBranchNavigation(requestedId: string) {
        return local.value.pending.find(item => item.resolvedConversationId && (!requestedId || item.sourceConversationId === requestedId || item.resolvedConversationId === requestedId))?.resolvedConversationId;
    }
    async function acknowledgeBranchNavigation(id: string) {
        const owner = storage; const generation = epoch;
        if (!owner || local.value.scope !== 'device' || conversationId.value !== id || detail.value?.conversation.id !== id) return;
        for (const item of local.value.pending.filter(pending => pending.resolvedConversationId === id)) {
            const saved = await owner.removePending(item.clientRequestId);
            if (generation !== epoch || owner !== storage) return;
            applyLocal(saved);
        }
    }
    async function retryPending(id: string) {
        await synchronize();
        const item = local.value.pending.find(p => p.clientRequestId === id);
        if (!item || item.rejected || item.resolvedConversationId || item.action === 'clearAllData' || delivery.value[id] !== 'unsent' || item.expectedDeletionEpoch !== capabilities.value?.deletionEpoch) return;
        return dispatch(item);
    }
    async function discardRejected(id: string) {
        const owner = storage; const generation = epoch;
        if (!owner || !local.value.pending.some(item => item.clientRequestId === id && item.rejected)) return;
        const saved = await owner.removePending(id); if (generation === epoch && owner === storage) applyLocal(saved);
    }
    async function restorePending(id: string) {
        const item = local.value.pending.find(p => p.clientRequestId === id);
        if (!item?.text || draft.value.trim()) return false;
        await setDraft(item.text); return true;
    }
    async function send(options: { mode?: string; content?: string; sourceMessageId?: string; restartQuery?: boolean; locale?: string; timezone?: string } = {}) {
        const content = options.content ?? draft.value;
        const limits = capabilities.value?.limits;
        if (!content.trim() || [...content].length > (limits?.inputCharacters ?? 4000) ||
            new TextEncoder().encode(content).byteLength > (limits?.inputBytes ?? 16384)) throw new Error('input_too_long');
        if (submitting.value || busy.value || activePending.value.length || !capabilities.value?.generationAvailable) throw new Error('already_running');
        const originalId = conversationId.value;
        // Separate the next draft immediately; pending storage owns the submitted original text.
        submitting.value = true;
        try {
            return await execute('send', { conversationId: originalId || undefined,
            expectedContextVersion: detail.value?.conversation.contextVersion ?? 0, mode: options.mode ?? 'chat', content,
            sourceMessageId: options.sourceMessageId, restartQuery: options.restartQuery, locale: options.locale, timezone: options.timezone });
        } finally { submitting.value = false; }
    }
    async function retry(userMessageId: string, intent = 'retry') {
        if (busy.value || activePending.value.length || !detail.value || !isLatestMessage(userMessageId)) return;
        return execute('retry', { conversationId: conversationId.value, userMessageId,
            expectedContextVersion: detail.value.conversation.contextVersion, intent });
    }
    async function branch(sourceMessageId: string, kind: string, text?: string) {
        if (!detail.value || busy.value) return;
        // Creation and first send are separate explicit commands. The resulting target survives first-send failure.
        const receipt = await execute('branch', { sourceConversationId: conversationId.value, sourceMessageId,
            expectedContextVersion: detail.value.conversation.contextVersion, kind }, text);
        return receipt;
    }
    async function stop() { const runId = detail.value?.activeRun?.runId; if (runId && busy.value) return execute('cancel', { runId }); return undefined; }
    async function viewVersion(assistantMessageId: string) {
        const generation = epoch; const ticket = navigation;
        const answer = await result(api.message(assistantMessageId));
        if (generation !== epoch || ticket !== navigation || answer.conversationId !== conversationId.value) return;
        viewedVersions.value = Object.fromEntries([...Object.entries(viewedVersions.value).filter(([id]) => id !== answer.userMessageId).slice(-19), [answer.userMessageId, answer]]);
    }
    function isLatestMessage(id: string) { return detail.value?.messages.find(message => message.id === id)?.turnSequence === (detail.value?.conversation.nextTurnSequence ?? 0) - 1; }
    async function selectVersion(userMessageId: string, assistantMessageId: string) {
        const current = detail.value;
        if (!current || !isLatestMessage(userMessageId) || busy.value) return;
        return execute('selectVersion', { conversationId: conversationId.value, userMessageId, assistantMessageId,
            expectedContextVersion: current.conversation.contextVersion });
    }
    async function loadMemories() {
        const generation = epoch; const next = await result(api.memories());
        if (generation === epoch) memories.value = next;
    }
    async function exportData(request: DTO.AssistantExportRequest, signal?: AbortSignal) {
        const generation = epoch; const response = await api.export(request, signal);
        if (generation !== epoch) throw new Error('Assistant identity changed');
        return response.data;
    }
    async function loadVersions(userMessageId: string, cursor?: string) {
        const generation = epoch; const ticket = navigation;
        const next = await result(api.versions({ userMessageId, cursor, limit: 20 }));
        if (generation !== epoch || ticket !== navigation) return undefined;
        const current = detail.value;
        if (current) detail.value = { ...current, messages: current.messages.map(message => message.id === userMessageId ? {
            ...message, versions: next.items, hasMoreVersions: !!next.nextCursor
        } : message) };
        return next;
    }
    function savePosition(position: ReadingPosition) {
        positions.delete(conversationId.value); positions.set(conversationId.value, position);
        if (positions.size > 20) positions.delete(positions.keys().next().value!);
    }
    async function loadLatest() { await open(conversationId.value); }
    function readingPosition() { return positions.get(conversationId.value); }
    function setHistoryVisible(value: boolean) { historyVisible = value; if (value) void loadHistory().catch(fail); }
    if (typeof window !== 'undefined') {
        const clear = () => reset(); const hide = () => reset(false);
        window.addEventListener('assistant-identity-cleared', clear); window.addEventListener('assistant-identity-hidden', hide);
        window.addEventListener('pageshow', pageShow); window.addEventListener('storage', identityStorage); window.addEventListener('beforeunload', beforeUnload);
        onScopeDispose(() => {
            window.removeEventListener('assistant-identity-cleared', clear); window.removeEventListener('assistant-identity-hidden', hide);
            window.removeEventListener('pageshow', pageShow); window.removeEventListener('storage', identityStorage); window.removeEventListener('beforeunload', beforeUnload);
            attached = 1; detach(); storage?.close();
        });
    }
    return { getReferenceTransaction, capabilities, conversationId, recentConversationId, legacyBackend, updatesAvailable, windowMissingLatest, loadLatest, exportData, loadVersions, detail, history, searchResults, memories, local, draft, error, loading, authenticated,
        online, submitting, activePending, discardRejected, retryLocalPersistence, removeDraft, delivery, viewedVersions, busy, pending, unsaved, canSend, initialize, reset, open, setDraft, chooseDraft, loadHistory,
        search, loadOlder, isLatestMessage, mergeSnapshot, synchronize, attach, detach, execute, send, retry, branch, stop, viewVersion, selectVersion,
        loadMemories, prepareDataClear, recoverBranchNavigation, acknowledgeBranchNavigation, retryPending, restorePending, reconcilePending, savePosition, readingPosition, setHistoryVisible };
});
