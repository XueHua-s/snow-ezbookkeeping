import { describe, expect, it, vi } from 'vitest';
import { compareAssistantReference, useAssistantReference } from './useAssistantReference.ts';
import type { AssistantReferenceDTO, AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';
import type { TransactionInfoResponse } from '@/models/transaction.ts';
const reference: AssistantReferenceDTO = { id: 'bill', time: 100, type: 3, sourceAmountMinor: '12345', comment: 'Lunch' };
const current = { id: 'bill', time: 100, type: 3, sourceAmount: 12345, destinationAmount: 0, comment: 'Lunch', lastModifiedUnixTime: 100 } as TransactionInfoResponse;
const answer = { references: [reference], queries: [{ queriedAt: '2026-10-15T01:00:00Z' }], savedAt: '2026-10-15T01:00:01Z', content: 'Historical answer' } as AssistantRunSnapshot;
describe('historical bill preflight', () => {
    it('detects server modification timestamp even when the displayed fields match', () => {
        expect(compareAssistantReference(reference, { ...current, lastModifiedUnixTime: Date.parse('2026-10-15T02:00:00Z') / 1000 }, answer.queries[0]!.queriedAt)).toBe('changed');
    });
    it('detects same-second amount or note changes from preserved fields', () => {
        expect(compareAssistantReference(reference, { ...current, sourceAmount: 54321 })).toBe('changed');
        expect(compareAssistantReference(reference, { ...current, comment: 'Changed' })).toBe('changed');
    });
    it('never treats rounded unsafe legacy amounts as exact matches', () => {
        expect(compareAssistantReference({ ...reference, sourceAmountMinor: '9007199254740993' }, { ...current, sourceAmount: 9007199254740992 })).toBe('unverified');
    });
    it('treats malformed decimal snapshots as unverified instead of throwing', () => {
        expect(compareAssistantReference({ ...reference, sourceAmountMinor: '12.3' }, current)).toBe('unverified');
        expect(compareAssistantReference({ ...reference, sourceAmountMinor: 'not-a-number' }, current)).toBe('unverified');
    });
    it('does not misclassify authentication expiry as a missing bill', async () => {
        const check = useAssistantReference(async () => { throw { response: { status: 401 } }; }, vi.fn());
        await check.check('bill', answer); expect(check.state.value).toBe('failed');
    });
    it('opens unchanged current details without mutating historical query time or state', async () => {
        const open = vi.fn(); const before = JSON.stringify(answer);
        const check = useAssistantReference(async () => current, open); await check.check('bill', answer);
        expect(open).toHaveBeenCalledWith('bill'); expect(JSON.stringify(answer)).toBe(before);
    });
    it('requires an explicit view-current action for a changed bill', async () => {
        const open = vi.fn(); const check = useAssistantReference(async () => ({ ...current, comment: 'Edited' }), open);
        await check.check('bill', answer); expect(check.state.value).toBe('changed'); expect(open).not.toHaveBeenCalled(); check.viewCurrent(); expect(open).toHaveBeenCalledWith('bill');
    });
    it('gives missing and inaccessible references the same status, but distinguishes connection failure', async () => {
        const load = vi.fn().mockRejectedValueOnce({ response: { status: 400, data: { errorCode: 205001 } } }).mockRejectedValueOnce({ response: { status: 403 } }).mockRejectedValueOnce(new Error('network'));
        const check = useAssistantReference(load, vi.fn()); await check.check('bill', answer); expect(check.state.value).toBe('unavailable');
        await check.retry(); expect(check.state.value).toBe('unavailable'); await check.retry(); expect(check.state.value).toBe('failed');
    });
    it('ignores a late response after switching conversations or dismissing the notice', async () => {
        let finish!: (value: TransactionInfoResponse) => void; const open = vi.fn(); const check = useAssistantReference(() => new Promise(resolve => { finish = resolve; }), open);
        const pending = check.check('bill', answer); check.dismiss(); finish(current); await pending; expect(open).not.toHaveBeenCalled(); expect(check.state.value).toBeUndefined();
    });
});
