import { expect, it, vi } from 'vitest';
import { prepareAssistantDataClear } from './dataClear.ts';
import type { createAssistantLocalStorage, AssistantLocalSnapshot } from '@/lib/assistant_local_storage.ts';
import type { AssistantCapabilities } from '@/models/ai_assistant_conversation.ts';

vi.mock('@/lib/services/assistant_conversations.ts', () => ({ assistantConversationServices: { command: vi.fn() } }));

type Storage = ReturnType<typeof createAssistantLocalStorage>;
it('rechecks identity after the nested checkIdentity helper resolves', async () => {
    const local: AssistantLocalSnapshot = { drafts: [], pending: [], scope: 'page', version: 0 };
    const savePending = vi.fn(async () => local);
    const oldOwner = { savePending } as unknown as Storage;
    const newOwner = { savePending } as unknown as Storage;
    let state = { local, storage: oldOwner, capabilities: { deletionEpoch: 1 } as AssistantCapabilities, epoch: 1, clockOffset: 0 };
    const reconcilePending = vi.fn(async () => undefined);
    const checking = prepareAssistantDataClear({
        state: () => state,
        checkIdentity: async () => {
            queueMicrotask(() => { state = { ...state, storage: newOwner, epoch: 2 }; });
            return true;
        },
        reconcilePending,
        applyLocal: vi.fn()
    }, false);
    await expect(checking).rejects.toThrow('Assistant identity changed');
    expect(reconcilePending).not.toHaveBeenCalled();
    expect(savePending).not.toHaveBeenCalled();
});
