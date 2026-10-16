// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { effectScope, reactive, shallowRef } from 'vue';
import { useAssistantCloudWorkspace } from './useAssistantCloudWorkspace.ts';
import type { AssistantMessageDTO, AssistantConversationDTO } from '@/models/ai_assistant_conversation.ts';
const mocks = vi.hoisted(() => ({ store: {} as Record<string, unknown> }));
vi.mock('@/stores/assistant.ts', () => ({ useAssistantStore: () => mocks.store }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));
const scopes: ReturnType<typeof effectScope>[] = [];
afterEach(() => { scopes.splice(0).forEach(scope => scope.stop()); });
function workspace(...args: Parameters<typeof useAssistantCloudWorkspace>) {
    const scope = effectScope();
    scopes.push(scope);
    return scope.run(() => useAssistantCloudWorkspace(...args))!;
}
const message = { id: 'question', content: 'Original', versions: [], selectedAnswer: { assistantMessageId: 'answer', state: 'completed' } } as unknown as AssistantMessageDTO;
beforeEach(() => {
    mocks.store = reactive({ draft: 'Untouched draft', detail: { messages: [message], conversation: { archived: false } }, viewedVersions: {}, online: true,
        capabilities: { generationAvailable: true, ownerScope: 'owner-a', deletionEpoch: 1 }, busy: false, pending: [], activePending: [], isLatestMessage: vi.fn().mockReturnValue(true), memories: { memoryUseEpoch: 4 },
        send: vi.fn().mockResolvedValue({}), branch: vi.fn().mockResolvedValue({ conversationId: 'branch' }), execute: vi.fn().mockResolvedValue({}),
        removeDraft: vi.fn(), setDraft: vi.fn(), viewVersion: vi.fn(), selectVersion: vi.fn(), search: vi.fn(), loadHistory: vi.fn() });
});
describe('shared cloud product commands', () => {
    it('submits next page bound to displayed answer version', async () => {
        const ui = workspace(); await ui.nextPage(message);
        expect(mocks.store['send']).toHaveBeenCalledWith(expect.objectContaining({ mode: 'next_page', sourceMessageId: 'answer' }));
    });
    it('requires confirmation for deletion', async () => {
        const confirm = vi.fn().mockResolvedValue(false); const ui = workspace(confirm);
        await ui.remove({ id: 'c', metadataVersion: 7 } as AssistantConversationDTO);
        expect(mocks.store['execute']).not.toHaveBeenCalled(); confirm.mockResolvedValue(true); await ui.remove({ id: 'c', metadataVersion: 7 } as AssistantConversationDTO);
        expect(mocks.store['execute']).toHaveBeenCalledWith('delete', { id: 'c', expectedMetadataVersion: 7 });
    });
    it('treats the end of a historical window as a branch, not the end of the server conversation', async () => {
        vi.mocked(mocks.store['isLatestMessage'] as () => boolean).mockReturnValue(false);
        const ui = workspace(async () => true); await ui.nextPage(message);
        expect(mocks.store['branch']).toHaveBeenCalledWith('answer', 'continue'); expect(mocks.store['send']).not.toHaveBeenCalled();
    });
    it('restart query is explicit and retains the selected answer source', async () => {
        const ui = workspace(); ui.retryQuerySource.value = 'answer'; await ui.restartQuery();
        expect(mocks.store['send']).toHaveBeenCalledWith(expect.objectContaining({ mode: 'next_page', sourceMessageId: 'answer', restartQuery: true }));
    });
    it('renders controlled server reasons and never leaks arbitrary provider messages', async () => {
        vi.mocked(mocks.store['send'] as () => Promise<unknown>).mockRejectedValue({ context: { reasonCode: 'cursor_expired' }, message: 'private provider diagnostic' });
        const ui = workspace(); await ui.send(); expect(ui.error.value).not.toContain('private provider'); expect(ui.error.value).not.toBe('Assistant request failed');
    });
    it('keeps editing possible when generation unavailable but explains send disabled', () => {
        (mocks.store['capabilities'] as { generationAvailable: boolean }).generationAvailable = false;
        const ui = workspace(); expect(ui.disabledReason.value).toBe('AI assistant is disabled'); ui.input.value = 'New local draft'; expect(mocks.store['setDraft']).toHaveBeenCalledWith('New local draft');
    });
});

describe('workspace operation scope', () => {
    it.each(['remove', 'nextPage'] as const)('ignores %s confirmation after owner changes', async command => {
        let decide!: (accepted: boolean) => void;
        const ui = workspace(() => new Promise(resolve => { decide = resolve; }));
        vi.mocked(mocks.store['isLatestMessage'] as () => boolean).mockReturnValue(false);
        const pending = command === 'remove'
            ? ui.remove({ id: 'old', metadataVersion: 1 } as AssistantConversationDTO)
            : ui.nextPage(message);
        (mocks.store['capabilities'] as { ownerScope: string }).ownerScope = 'owner-b';
        decide(true);
        await pending;
        expect(mocks.store['execute']).not.toHaveBeenCalled();
        expect(mocks.store['branch']).not.toHaveBeenCalled();
    });
    it.each(['owner', 'route'])('ignores late errors and finally while a new %s is working', async boundary => {
        const route = shallowRef('one');
        const ui = workspace(undefined, () => route.value);
        let fail!: (error: Error) => void;
        let finish!: () => void;
        const old = ui.run(() => new Promise((_, reject) => { fail = reject; }));
        if (boundary === 'owner') (mocks.store['capabilities'] as { ownerScope: string }).ownerScope = 'owner-b';
        if (boundary === 'route') route.value = 'two';
        const current = ui.run(() => new Promise<void>(resolve => { finish = resolve; }));
        fail(new Error('old'));
        await old;
        expect(ui.error.value).toBe('');
        expect(ui.working.value).toBe(true);
        finish();
        await current;
        expect(ui.working.value).toBe(false);
    });
});

describe('workspace invalidation boundaries', () => {
    it.each(['epoch', 'route', 'dispose', 'roundtrip'])('declines a pending local draft confirmation after %s', async boundary => {
        const route = shallowRef('one');
        let decide!: (value: boolean) => void;
        const ui = workspace(() => new Promise(resolve => { decide = resolve; }), () => route.value);
        const pending = ui.removeLocalDraft('private-draft');
        if (boundary === 'epoch') (mocks.store['capabilities'] as { deletionEpoch: number }).deletionEpoch++;
        if (boundary === 'route') route.value = 'two';
        if (boundary === 'dispose') scopes.at(-1)!.stop();
        if (boundary === 'roundtrip') {
            route.value = 'two';
            route.value = 'one';
        }
        decide(true);
        await pending;
        expect(mocks.store['removeDraft']).not.toHaveBeenCalled();
    });

    it('does not let a confirmed deletion use the next deletion epoch', async () => {
        let decide!: (value: boolean) => void;
        const ui = workspace(() => new Promise(resolve => { decide = resolve; }));
        const pending = ui.remove({ id: 'c', metadataVersion: 7 } as AssistantConversationDTO);
        (mocks.store['capabilities'] as { deletionEpoch: number }).deletionEpoch++;
        decide(true);
        await pending;
        expect(mocks.store['execute']).not.toHaveBeenCalled();
    });
    it('resets owner filters but preserves same-owner history when navigating', () => {
        const route = shallowRef('one');
        const ui = workspace(undefined, () => route.value);
        ui.historyOpen.value = true;
        ui.query.value = 'Private query';
        ui.includeVersions.value = true;
        route.value = 'two';
        expect(ui.query.value).toBe('Private query');
        (mocks.store['capabilities'] as { ownerScope: string }).ownerScope = 'new-owner';
        expect(ui.historyOpen.value).toBe(false);
        expect(ui.query.value).toBe('');
        expect(ui.includeVersions.value).toBe(false);
    });
    it('does not let old same-scope completion clear working while another action remains', async () => {
        const ui = workspace();
        let finish!: () => void;
        const first = ui.run(() => new Promise<void>(resolve => { finish = resolve; }));
        await ui.run(async () => undefined);
        expect(ui.working.value).toBe(true);
        finish();
        await first;
        expect(ui.working.value).toBe(false);
    });
});

it('rechecks authority at the command boundary after the confirmation microtask', async () => {
    let decide!: (value: boolean) => void;
    const ui = workspace(() => new Promise(resolve => { decide = resolve; }));
    const pending = ui.remove({ id: 'c', metadataVersion: 7 } as AssistantConversationDTO);
    decide(true);
    queueMicrotask(() => { (mocks.store['capabilities'] as { ownerScope: string }).ownerScope = 'b'; });
    await pending;
    expect(mocks.store['execute']).not.toHaveBeenCalled();
});

it('keeps the active command and history when capabilities refresh without identity changes', async () => {
    const ui = workspace();
    ui.historyOpen.value = true;
    let finish!: () => void;
    const pending = ui.run(() => new Promise<void>(resolve => { finish = resolve; }));
    mocks.store['capabilities'] = { ...(mocks.store['capabilities'] as object) };
    expect(ui.working.value).toBe(true);
    expect(ui.historyOpen.value).toBe(true);
    finish();
    await pending;
    expect(ui.working.value).toBe(false);
});
