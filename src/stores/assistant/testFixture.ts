import { beforeEach, afterEach, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import type { AssistantCapabilities, AssistantConversationDetail, AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';
import type { AssistantLocalSnapshot, AssistantLocalPending } from '@/lib/assistant_local_storage.ts';
import { useAssistantStore } from '../assistant.ts';

const mocks = vi.hoisted(() => ({ api: Object.fromEntries(['capabilities', 'command', 'list', 'search', 'get', 'state', 'update', 'delete', 'clear', 'branch', 'send', 'retry', 'message', 'versions', 'selectVersion', 'run', 'cancel', 'memories', 'saveMemory', 'deleteMemory', 'memorySettings', 'subscribe'].map(key => [key, vi.fn()])), save: vi.fn(), local: vi.fn(), transaction: vi.fn() }));
const preferences = vi.hoisted(() => ({ language: vi.fn(), timezone: vi.fn() }));
vi.mock('@/lib/services/assistant_conversations.ts', () => ({ assistantConversationServices: mocks.api }));
vi.mock('@/lib/services/transactions.ts', () => ({ transactionsServices: { getTransaction: mocks.transaction } }));
vi.mock('@/lib/assistant_local_storage.ts', () => ({ createAssistantLocalStorage: mocks.local }));
vi.mock('@/lib/settings.ts', async importOriginal => ({
    ...await importOriginal<typeof import('@/lib/settings.ts')>(),
    getSessionCurrentLanguageKey: preferences.language,
    getTimeZone: preferences.timezone
}));
export const response = (value: unknown) => ({ data: { success: true, result: value } });
export const capability = (overrides = {}) => ({ ownerScope: 'A', deletionEpoch: 1, generationAvailable: true, historyAvailable: true,
    limits: { inputCharacters: 4000, inputBytes: 16384 }, ...overrides }) as AssistantCapabilities;
export const conversation = (id = '1', overrides = {}): AssistantConversationDetail => ({ conversation: { id, contextVersion: 1, metadataVersion: 1, ...overrides },
    messages: [], hasMore: false }) as unknown as AssistantConversationDetail;
export const snapshot = (overrides = {}): AssistantRunSnapshot => ({ runId: '3', conversationId: '1', userMessageId: '2', assistantMessageId: '3',
    schemaVersion: 1, reasonCode: null, savedAt: '2026-10-15T00:00:00Z', attemptNumber: 1, revision: 1, state: 'running', content: 'partial', queries: [], references: [], ...overrides }) as AssistantRunSnapshot;
export function deferred<T>() { let resolve!: (value: T) => void; let reject!: (error: unknown) => void; const promise = new Promise<T>((a, b) => { resolve = a; reject = b; }); return { promise, resolve, reject }; }
export let local: AssistantLocalSnapshot;
export let store: ReturnType<typeof useAssistantStore>;
beforeEach(() => {
    preferences.language.mockReturnValue('zh_Hans');
    preferences.timezone.mockReturnValue('Asia/Shanghai');
    vi.stubGlobal('document', { hidden: false, addEventListener: vi.fn(), removeEventListener: vi.fn() });
    vi.stubGlobal('window', { addEventListener: vi.fn(), removeEventListener: vi.fn() });
    vi.stubGlobal('navigator', { onLine: true });
    local = { drafts: [], pending: [], scope: 'device', version: 0 };
    mocks.save.mockImplementation(async (item: AssistantLocalPending) => { local = { ...local, pending: [...local.pending, item] }; return local; });
    mocks.local.mockImplementation(() => ({ editorId: 'editor', load: async () => local, reconcileEpoch: async (epoch: number) => {
        local = { ...local, pending: local.pending.filter(p => p.expectedDeletionEpoch === epoch || p.action === 'clearAllData') }; return local;
    }, savePending: mocks.save, saveDraft: async ({ conversationId, text }: { conversationId: string; text: string }) => {
        local = { ...local, drafts: text ? [{ id: 'draft', editorId: 'editor', deviceId: 'device', revision: 1, conflict: false, conversationId, text, expectedDeletionEpoch: 1 }] : [] }; return local;
    }, removeDraft: async () => local, retryPersistence: async () => local, removePending: async (id: string) => { local = { ...local, pending: local.pending.filter(p => p.clientRequestId !== id) }; return local; },
    resolveBranch: async (id: string, targetId: string) => {
        const pending = local.pending.find(item => item.clientRequestId === id)!;
        local = { ...local, pending: local.pending.map(item => item.clientRequestId === id ? { ...item, resolvedConversationId: targetId } : item) };
        if (pending.text && !local.drafts.some(draft => draft.conversationId === targetId)) local.drafts.push({ id: 'branch-draft', editorId: 'editor', deviceId: 'device', revision: 1, conflict: false, conversationId: targetId, text: pending.text, expectedDeletionEpoch: 1 });
        return local;
    },
    migrateDraft: async () => local, clearConversation: async (id: string) => { local = { ...local, drafts: local.drafts.filter(item => item.conversationId !== id), pending: local.pending.filter(item => item.conversationId !== id && item.resolvedConversationId !== id) }; return local; }, clearOwner: async () => local,
    markPendingRejected: async (id: string) => { local = { ...local, pending: local.pending.map(item => item.clientRequestId === id ? { ...item, rejected: true } : item) }; return local; },
    close: vi.fn(), subscribe: () => vi.fn() }));
    mocks.api['capabilities']!.mockResolvedValue(response(capability()));
    mocks.api['list']!.mockResolvedValue(response({ items: [] }));
    mocks.api['get']!.mockImplementation(async ({ id }: { id: string }) => response(conversation(id)));
    mocks.api['state']!.mockResolvedValue(response({ id: '1', accessible: true, contextVersion: 1, metadataVersion: 1, activeRunRevision: 0 }));
    mocks.api['command']!.mockResolvedValue(response({ status: 'unknown' }));
    mocks.api['send']!.mockRejectedValue(new Error('lost response'));
    setActivePinia(createPinia()); store = useAssistantStore();
});
afterEach(() => { store.$dispose(); vi.unstubAllGlobals(); vi.useRealTimers(); });


export function setLocal(value: AssistantLocalSnapshot) { local = value; }

export { mocks, preferences };
