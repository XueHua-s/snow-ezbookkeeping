import { ref } from 'vue';
import type { AssistantReferenceDTO, AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';
import type { TransactionInfoResponse } from '@/models/transaction.ts';

/** Compare only fields actually preserved in the historical query snapshot; never round int64 amounts. */
export function compareAssistantReference(reference: AssistantReferenceDTO, current: TransactionInfoResponse, queriedAt?: string): 'changed' | 'unchanged' | 'unverified' {
    const readAt = queriedAt ? Date.parse(queriedAt) : NaN;
    if (current.lastModifiedUnixTime && Number.isFinite(readAt) && current.lastModifiedUnixTime * 1000 > readAt) return 'changed';
    if (reference.id !== current.id || reference.time !== current.time || reference.type !== current.type ||
        (reference.comment ?? '') !== current.comment) return 'changed';
    const fields: [string | undefined, string | undefined][] = [
        [reference.categoryName, current.category?.name], [reference.sourceAccountName, current.sourceAccount?.name],
        [reference.destinationAccountName, current.destinationAccount?.name], [reference.currency, current.sourceAccount?.currency],
        [reference.destinationCurrency, current.destinationAccount?.currency]
    ];
    if (fields.some(([past, now]) => now !== undefined && (past ?? '') !== now)) return 'changed';
    const amounts: [string | undefined, number][] = [[reference.sourceAmountMinor, current.sourceAmount]];
    if (reference.destinationAmountMinor !== undefined) amounts.push([reference.destinationAmountMinor, current.destinationAmount]);
    if (amounts.some(([past]) => !/^-?\d+$/.test(past ?? ''))) return 'unverified';
    if (amounts.some(([past, now]) => Number.isSafeInteger(now) && BigInt(past ?? '0') !== BigInt(now))) return 'changed';
    return amounts.some(([, now]) => !Number.isSafeInteger(now)) ? 'unverified' : 'unchanged';
}

/** A preflight is read-only and never changes answer timestamps, history, drafts or reading position. */
export function useAssistantReference(load: (id: string) => Promise<TransactionInfoResponse>, open: (id: string) => void) {
    const state = ref<'checking' | 'changed' | 'unverified' | 'unavailable' | 'failed'>();
    const selected = ref<{ id: string; answer: AssistantRunSnapshot }>();
    let revision = 0;
    async function check(id: string, answer: AssistantRunSnapshot) {
        const ticket = ++revision;
        selected.value = { id, answer }; state.value = 'checking';
        try {
            const current = await load(id);
            if (ticket !== revision) return;
            const original = answer.references.find(reference => reference.id === id);
            const comparison = original ? compareAssistantReference(original, current, answer.queries.map(query => query.queriedAt).sort()[0]) : 'unverified';
            if (comparison === 'unchanged') { state.value = undefined; open(id); }
            else state.value = comparison;
        } catch (failure) {
            if (ticket !== revision) return;
            const error = failure as { status?: number; errorCode?: number; response?: { status?: number; data?: { errorCode?: number } } };
            const status = error?.response?.status ?? error?.status;
            // ErrTransactionNotFound is deliberately identical for missing and another owner's records.
            const code = error?.response?.data?.errorCode ?? error?.errorCode;
            state.value = status === 403 || status === 404 || code === 205001 ? 'unavailable' : 'failed';
        }
    }
    function viewCurrent() { if (selected.value && (state.value === 'changed' || state.value === 'unverified')) { const id = selected.value.id; state.value = undefined; open(id); } }
    function dismiss() { revision++; state.value = undefined; selected.value = undefined; }
    function retry() { const value = selected.value; if (value) return check(value.id, value.answer); return Promise.resolve(); }
    return { state, selected, check, viewCurrent, dismiss, retry };
}
