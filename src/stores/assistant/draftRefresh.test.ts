import { describe, expect, it, vi } from 'vitest';
import { mocks, response, deferred, local, store, setLocal } from './testFixture.ts';

describe('assistant draft refresh', () => {
    it.each(['open', 'initialize'] as const)('sends the visible new draft when %s runs before draft persistence finishes', async refresh => {
        await store.initialize();
        await store.open('1');
        await store.setDraft('old saved draft');
        const owner = mocks.local.mock.results[0]!.value;
        const held = deferred<typeof local>();
        vi.spyOn(owner, 'saveDraft').mockReturnValueOnce(held.promise);
        const editing = store.setDraft('new visible question');
        if (refresh === 'open') await store.open('1');
        else await store.initialize();
        expect(store.draft).toBe('new visible question');
        mocks.api['send']!.mockResolvedValue(response({ status: 'accepted', conversationId: '1' }));
        await store.send();
        expect(mocks.api['send']).toHaveBeenCalledWith(expect.objectContaining({ content: 'new visible question' }));
        held.resolve(local);
        await editing;
        expect(store.draft).toBe('');
    });

    it('hydrates saved text after an early empty restoration before identity storage is ready', async () => {
        await store.open('1');
        expect(store.draft).toBe('');
        setLocal({ ...local, drafts: [{ id: 'persisted', editorId: 'editor', deviceId: 'device',
            revision: 1, conflict: false, conversationId: '1', text: 'saved before reload', expectedDeletionEpoch: 1 }] });
        await store.initialize();
        expect(store.draft).toBe('saved before reload');
    });

    it('restores the appropriate saved draft when navigating to another conversation', async () => {
        await store.initialize();
        await store.open('1');
        await store.setDraft('conversation one');
        await store.open('2');
        expect(store.draft).toBe('');
        await store.open('1');
        expect(store.draft).toBe('conversation one');
    });
});
