import type { AssistantCapabilities, AssistantCommand } from '@/models/ai_assistant_conversation.ts';
import type { createAssistantLocalStorage, AssistantLocalSnapshot } from '@/lib/assistant_local_storage.ts';
import { assistantConversationServices as api } from '@/lib/services/assistant_conversations.ts';
import { result, AssistantDataClearIncompleteError } from './types.ts';

type Storage = ReturnType<typeof createAssistantLocalStorage>;
interface State { storage?: Storage; capabilities?: AssistantCapabilities; local: AssistantLocalSnapshot; epoch: number; clockOffset: number }
interface Context {
    state(): State;
    checkIdentity(): Promise<boolean>;
    reconcilePending(): Promise<void>;
    applyLocal(snapshot: AssistantLocalSnapshot): void;
}
const envelope = (item: AssistantCommand): AssistantCommand => ({ clientRequestId: item.clientRequestId, requestIssuedAt: item.requestIssuedAt, expectedDeletionEpoch: item.expectedDeletionEpoch });

/** Global clearing has two completion boundaries; an assistant receipt alone cannot certify ledger deletion. */
export async function prepareAssistantDataClear(context: Context, newOperation: boolean): Promise<AssistantCommand> {
    const initial = context.state();
    const prior = initial.local.pending.find(item => item.action === 'clearAllData');
    if (prior) {
        const receipt = await result(api.command(prior.clientRequestId));
        if (initial.epoch !== context.state().epoch) throw new Error('Assistant identity changed');
        if (receipt.status !== 'unknown') {
            if (receipt.operationCompleted) return envelope(prior);
            if (!newOperation) throw new AssistantDataClearIncompleteError();
            if (initial.storage) {
                const saved = await initial.storage.removePending(prior.clientRequestId);
                if (initial.epoch !== context.state().epoch) throw new Error('Assistant identity changed');
                context.applyLocal(saved);
            }
        } else if (newOperation) throw new Error('Confirm the original clear operation before starting another one.');
    }
    if (!await context.checkIdentity()) throw new Error('Assistant unavailable');
    const checked = context.state();
    if (!checked.storage || !checked.capabilities) throw new Error('Assistant unavailable');
    if (prior && !newOperation && prior.expectedDeletionEpoch !== checked.capabilities.deletionEpoch) throw new Error('data_epoch_conflict');
    await context.reconcilePending();
    const current = context.state();
    if (current.epoch !== checked.epoch || current.storage !== checked.storage) throw new Error('Assistant identity changed');
    const existing = current.local.pending.find(item => item.action === 'clearAllData');
    if (existing) return envelope(existing);
    const command = { clientRequestId: crypto.randomUUID(), requestIssuedAt: new Date(Date.now() + current.clockOffset).toISOString(), expectedDeletionEpoch: checked.capabilities.deletionEpoch };
    const saved = await checked.storage.savePending({ ...command, action: 'clearAllData', payload: command });
    if (checked.epoch !== context.state().epoch || checked.storage !== context.state().storage) throw new Error('Assistant identity changed');
    context.applyLocal(saved); return command;
}
