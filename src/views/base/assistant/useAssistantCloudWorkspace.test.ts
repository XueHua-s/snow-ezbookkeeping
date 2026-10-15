// @vitest-environment happy-dom
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { reactive } from 'vue';
import { useAssistantCloudWorkspace } from './useAssistantCloudWorkspace.ts';
import type { AssistantMessageDTO, AssistantConversationDTO } from '@/models/ai_assistant_conversation.ts';
const mocks = vi.hoisted(() => ({ store: {} as Record<string, unknown> }));
vi.mock('@/stores/assistant.ts', () => ({ useAssistantStore: () => mocks.store }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));
const message = { id: 'question', content: 'Original', versions: [], selectedAnswer: { assistantMessageId: 'answer', state: 'completed' } } as unknown as AssistantMessageDTO;
beforeEach(() => {
    mocks.store = reactive({ draft: 'Untouched draft', detail: { messages: [message], conversation: { archived: false } }, viewedVersions: {}, online: true,
        capabilities: { generationAvailable: true }, busy: false, pending: [], activePending: [], isLatestMessage: vi.fn().mockReturnValue(true), memories: { memoryUseEpoch: 4 },
        send: vi.fn().mockResolvedValue({}), branch: vi.fn().mockResolvedValue({ conversationId: 'branch' }), execute: vi.fn().mockResolvedValue({}),
        setDraft: vi.fn(), viewVersion: vi.fn(), selectVersion: vi.fn(), search: vi.fn(), loadHistory: vi.fn() });
});
describe('shared cloud product commands', () => {
    it('creates an edit branch without sending, preserving the original draft', async () => {
        const ui = useAssistantCloudWorkspace(); ui.edit(message); ui.input.value = 'Edited question'; await ui.send();
        expect(mocks.store['branch']).toHaveBeenCalledWith('question', 'edit', 'Edited question');
        expect(mocks.store['send']).not.toHaveBeenCalled(); expect(mocks.store['draft']).toBe('Untouched draft'); expect(ui.editing.value).toBeUndefined();
    });
    it('keeps edit text after a branch creation failure', async () => {
        vi.mocked(mocks.store['branch'] as () => Promise<unknown>).mockRejectedValue(new Error('offline'));
        const ui = useAssistantCloudWorkspace(); ui.edit(message); ui.input.value = 'Edited question'; await ui.send();
        expect(ui.editing.value?.id).toBe('question'); expect(ui.input.value).toBe('Edited question'); expect(ui.error.value).toBe('Assistant request failed');
    });
    it('submits next page bound to displayed answer version', async () => {
        const ui = useAssistantCloudWorkspace(); await ui.nextPage(message);
        expect(mocks.store['send']).toHaveBeenCalledWith(expect.objectContaining({ mode: 'next_page', sourceMessageId: 'answer' }));
    });
    it('requires confirmation for clear and delete', async () => {
        const confirm = vi.fn().mockResolvedValue(false); const ui = useAssistantCloudWorkspace(confirm);
        await ui.clear(); await ui.remove({ id: 'c', metadataVersion: 7 } as AssistantConversationDTO);
        expect(mocks.store['execute']).not.toHaveBeenCalled(); confirm.mockResolvedValue(true); await ui.clear();
        expect(mocks.store['execute']).toHaveBeenCalledWith('clear', {});
    });
    it('distinguishes memory add/correct/confirm and includes server epoch', async () => {
        const ui = useAssistantCloudWorkspace(); await ui.saveMemory(undefined, 'Budget preference');
        expect(mocks.store['execute']).toHaveBeenCalledWith('saveMemory', expect.objectContaining({ intent: 'add', expectedMemoryUseEpoch: 4, text: 'Budget preference' }));
    });
    it('summary confirms its period without treating unsent draft as a filter', async () => {
        const ui = useAssistantCloudWorkspace(); ui.summaryPeriod.value = '2026-01-01 to 2026-01-31'; await ui.summary();
        expect(mocks.store['send']).toHaveBeenCalledWith({ mode: 'summary', content: 'Generate AI Summary: 2026-01-01 to 2026-01-31' });
        expect(mocks.store['draft']).toBe('Untouched draft');
    });
    it('treats the end of a historical window as a branch, not the end of the server conversation', async () => {
        vi.mocked(mocks.store['isLatestMessage'] as () => boolean).mockReturnValue(false);
        const ui = useAssistantCloudWorkspace(async () => true); await ui.nextPage(message);
        expect(mocks.store['branch']).toHaveBeenCalledWith('answer', 'continue'); expect(mocks.store['send']).not.toHaveBeenCalled();
    });
    it('restart query is explicit and retains the selected answer source', async () => {
        const ui = useAssistantCloudWorkspace(); ui.retryQuerySource.value = 'answer'; await ui.restartQuery();
        expect(mocks.store['send']).toHaveBeenCalledWith(expect.objectContaining({ mode: 'next_page', sourceMessageId: 'answer', restartQuery: true }));
    });
    it('renders controlled server reasons and never leaks arbitrary provider messages', async () => {
        vi.mocked(mocks.store['send'] as () => Promise<unknown>).mockRejectedValue({ context: { reasonCode: 'cursor_expired' }, message: 'private provider diagnostic' });
        const ui = useAssistantCloudWorkspace(); await ui.send(); expect(ui.error.value).not.toContain('private provider'); expect(ui.error.value).not.toBe('Assistant request failed');
    });
    it('keeps editing possible when generation unavailable but explains send disabled', () => {
        (mocks.store['capabilities'] as { generationAvailable: boolean }).generationAvailable = false;
        const ui = useAssistantCloudWorkspace(); expect(ui.disabledReason.value).toBe('AI assistant is disabled'); ui.input.value = 'New local draft'; expect(mocks.store['setDraft']).toHaveBeenCalledWith('New local draft');
    });
});
