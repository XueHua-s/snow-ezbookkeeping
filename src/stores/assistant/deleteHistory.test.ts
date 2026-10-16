import { describe, expect, it, vi } from 'vitest';
import { mocks, response, capability, conversation, deferred, local, store } from './testFixture.ts';

describe('assistant deletion history refresh', () => {
    it.each(['1', '2'])('immediately removes deleted conversation %s from history', async deletedId => {
        await store.initialize();
        await store.open('1');
        mocks.api['list']!.mockResolvedValueOnce(response({ items: [conversation('1').conversation, conversation('2').conversation] }));
        await store.loadHistory();
        mocks.api['list']!.mockClear();
        const remaining = deletedId === '1' ? '2' : '1';
        mocks.api['list']!.mockResolvedValue(response({ items: [conversation(remaining).conversation] }));
        mocks.api['delete']!.mockResolvedValue(response({ status: 'deleted', conversationId: deletedId }));

        await store.execute('delete', { id: deletedId, expectedMetadataVersion: 1 });

        expect(store.history.items.map(item => item.id)).toEqual([remaining]);
        expect(mocks.api['list']).toHaveBeenCalledTimes(1);
        expect(store.conversationId).toBe(deletedId === '1' ? '' : '1');
    });

    it('does not refresh a new owner history after old deletion cleanup finishes', async () => {
        await store.initialize();
        await store.open('1');
        const owner = mocks.local.mock.results[0]!.value;
        const held = deferred<typeof local>();
        const started = deferred<void>();
        vi.spyOn(owner, 'clearConversation').mockImplementationOnce(() => {
            started.resolve();
            return held.promise;
        });
        mocks.api['delete']!.mockResolvedValue(response({ status: 'deleted', conversationId: '2' }));
        const deleting = store.execute('delete', { id: '2', expectedMetadataVersion: 1 });
        await started.promise;
        store.reset(false);
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ ownerScope: 'B' })));
        await store.initialize();
        mocks.api['list']!.mockClear();
        held.resolve(local);
        await deleting;
        expect(mocks.api['list']).not.toHaveBeenCalled();
        expect(store.capabilities?.ownerScope).toBe('B');
    });
});
