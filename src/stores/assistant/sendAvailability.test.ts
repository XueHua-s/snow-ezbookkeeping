import { describe, expect, it } from 'vitest';
import { mocks, response, capability, conversation, snapshot, store } from './testFixture.ts';

describe('assistant send availability', () => {
    it('explains unavailable generation without claiming an answer is running', async () => {
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ generationAvailable: false })));
        await store.initialize();
        await store.setDraft('my question');
        await expect(store.send()).rejects.toThrow('generation_unavailable');
        expect(store.draft).toBe('my question');
        expect(mocks.api['send']).not.toHaveBeenCalled();
    });

    it('keeps a new draft while the previous delivery remains unconfirmed', async () => {
        await store.initialize();
        await store.setDraft('first question');
        await expect(store.send()).rejects.toThrow('lost response');
        await store.setDraft('next question');
        await expect(store.send()).resolves.toBeUndefined();
        expect(store.draft).toBe('next question');
        expect(store.pending).toHaveLength(1);
        expect(mocks.api['send']).toHaveBeenCalledTimes(1);
    });

    it('reserves the active answer prompt for an actual running answer', async () => {
        await store.initialize();
        mocks.api['get']!.mockResolvedValue(response({ ...conversation(), activeRun: snapshot() }));
        await store.open('1');
        await store.setDraft('next question');
        await expect(store.send()).rejects.toThrow('already_running');
        expect(store.draft).toBe('next question');
        expect(mocks.api['send']).not.toHaveBeenCalled();
    });
});
