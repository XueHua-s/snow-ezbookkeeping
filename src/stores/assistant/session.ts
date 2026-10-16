import { ref, shallowRef } from 'vue';
import { assistantConversationServices as api } from '@/lib/services/assistant_conversations.ts';
import { createAssistantLocalStorage, type AssistantLocalSnapshot } from '@/lib/assistant_local_storage.ts';
import type { AssistantCapabilities } from '@/models/ai_assistant_conversation.ts';
import { result, statusCode } from './types.ts';

/** Identity and browser recovery authority. Leases must be checked at each caller's side-effect boundary. */
export function useAssistantSession() {
    const capabilities = shallowRef<AssistantCapabilities>();
    const authenticated = ref(false);
    const legacyBackend = ref(false);
    const error = shallowRef<unknown>();
    const local = shallowRef<AssistantLocalSnapshot>({ drafts: [], pending: [], scope: 'page', version: 0 });
    const draft = ref('');
    const conversationId = ref('');
    const recentConversationId = ref('');
    const generation = ref(0);
    let storage: ReturnType<typeof createAssistantLocalStorage> | undefined;
    let unsubscribe: (() => void) | undefined;
    let draftId: string | undefined;
    let draftRevision: number | undefined;
    let identityTicket = 0;
    let identityRevision = 0;
    let disposed = false;
    let clockOffset = 0;

    function capture() {
        const epoch = generation.value;
        return () => !disposed && epoch === generation.value;
    }
    function applyLocal(value: AssistantLocalSnapshot) {
        // Row versions can decrease during page fallback cleanup/recovery. The owner-local
        // sequence orders completed snapshots across those transitions, without comparing owners.
        if (value.sequence !== undefined && local.value.sequence !== undefined && value.sequence < local.value.sequence)
            return;
        if (value.sequence === undefined && value.scope === local.value.scope && value.version < local.value.version)
            return;
        local.value = value;
    }
    function reset(clearLocal = true) {
        const old = storage;
        storage = undefined;
        unsubscribe?.();
        unsubscribe = undefined;
        if (clearLocal && capabilities.value) {
            try {
                localStorage.removeItem('assistant-recent:' + capabilities.value.ownerScope);
            } catch {
                /* Optional hint. */
            }
        }
        if (old) {
            if (clearLocal)
                void old
                    .clearOwner()
                    .catch(() => undefined)
                    .finally(() => old.close());
            else old.close();
        }
        capabilities.value = undefined;
        authenticated.value = false;
        legacyBackend.value = false;
        error.value = undefined;
        conversationId.value = '';
        recentConversationId.value = '';
        local.value = { drafts: [], pending: [], scope: 'page', version: 0 };
        draft.value = '';
        draftId = undefined;
        draftRevision = undefined;
        clockOffset = 0;
        identityTicket++;
        identityRevision++;
        generation.value++;
    }
    function fail(failure: unknown) {
        if (disposed) return;
        if (statusCode(failure) === 401) reset(false);
        error.value = failure;
    }
    async function checkIdentity() {
        if (disposed) return undefined;
        const epoch = generation.value;
        let ticket = ++identityTicket;
        let next: AssistantCapabilities;
        try {
            next = await result(api.capabilities());
        } catch (failure) {
            if (disposed || epoch !== generation.value || ticket !== identityTicket) return undefined;
            // Capability support belongs to this request, including foreground recovery.
            legacyBackend.value = statusCode(failure) === 404;
            throw failure;
        }
        if (disposed || epoch !== generation.value || ticket !== identityTicket) return undefined;
        if (capabilities.value && capabilities.value.ownerScope !== next.ownerScope) {
            reset();
            // Establish the new identity, but invalidate every old caller continuation.
            ticket = ++identityTicket;
        }
        if (!storage) {
            storage = createAssistantLocalStorage({ ownerScope: next.ownerScope });
            const owner = storage;
            unsubscribe = owner.subscribe(() => {
                if (storage === owner && !disposed) void refreshLocal().catch(fail);
            });
        }
        const owner = storage;
        const valid = capture();
        const changed = capabilities.value && next.deletionEpoch > capabilities.value.deletionEpoch;
        const loaded = await owner.reconcileEpoch(next.deletionEpoch).catch((failure) => {
            if (valid() && ticket === identityTicket) throw failure;
        });
        if (!loaded || !valid() || ticket !== identityTicket) return undefined;
        if (changed) {
            conversationId.value = '';
            draft.value = '';
            generation.value++;
        }
        const serverTime = Date.parse(next.serverTime);
        if (Number.isFinite(serverTime)) clockOffset = serverTime - Date.now();
        capabilities.value = next;
        applyLocal(loaded);
        authenticated.value = true;
        legacyBackend.value = false;
        try {
            recentConversationId.value = localStorage.getItem('assistant-recent:' + next.ownerScope) ?? '';
        } catch {
            /* Optional hint. */
        }
        return { valid: capture() };
    }
    async function refreshLocal() {
        const owner = storage;
        const valid = capture();
        if (!owner) return;
        try {
            const next = await owner.load();
            if (valid()) applyLocal(next);
        } catch (failure) {
            if (valid()) fail(failure);
        }
    }
    function restoreDraft() {
        const choices = local.value.drafts.filter((item) => item.conversationId === conversationId.value);
        const found = choices.find((item) => item.editorId === storage?.editorId) ?? choices[0];
        draft.value = found?.text ?? '';
        draftId = found?.id;
        draftRevision = found?.revision;
    }
    async function setDraft(text: string) {
        if (disposed) return;
        draft.value = text;
        const owner = storage;
        const valid = capture();
        const id = conversationId.value;
        if (!owner || !capabilities.value) return;
        try {
            const saved = await owner.saveDraft({
                conversationId: id,
                text,
                expectedDeletionEpoch: capabilities.value.deletionEpoch,
                id: draftId,
                revision: draftRevision,
            });
            if (!valid()) return;
            applyLocal(saved);
            if (conversationId.value === id && draft.value === text) {
                const own = local.value.drafts.find(
                    (item) => item.conversationId === id && item.editorId === owner.editorId,
                );
                draftId = own?.id;
                draftRevision = own?.revision;
            }
        } catch (failure) {
            if (valid()) fail(failure);
        }
    }
    async function retryLocalPersistence() {
        const owner = storage;
        const valid = capture();
        if (!owner) return;
        const saved = await owner.retryPersistence();
        if (valid()) applyLocal(saved);
    }
    async function removeDraft(id: string) {
        const owner = storage;
        const valid = capture();
        if (!owner) return;
        const saved = await owner.removeDraft(id);
        if (valid()) applyLocal(saved);
    }
    function chooseDraft(id: string) {
        const selected = local.value.drafts.find(
            (item) => item.id === id && item.conversationId === conversationId.value,
        );
        if (!selected || disposed) return;
        draftId = selected.id;
        draftRevision = selected.revision;
        draft.value = selected.text;
    }
    function dispose() {
        disposed = true;
        reset(false);
    }
    return {
        capabilities,
        authenticated,
        legacyBackend,
        error,
        local,
        draft,
        conversationId,
        recentConversationId,
        generation,
        get identityRevision() {
            return identityRevision;
        },
        get storage() {
            return storage;
        },
        get clockOffset() {
            return clockOffset;
        },
        get disposed() {
            return disposed;
        },
        capture,
        applyLocal,
        reset,
        fail,
        checkIdentity,
        restoreDraft,
        setDraft,
        retryLocalPersistence,
        removeDraft,
        chooseDraft,
        dispose,
    };
}
export type AssistantSession = ReturnType<typeof useAssistantSession>;
