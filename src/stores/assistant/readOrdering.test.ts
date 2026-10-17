import { describe, expect, it, vi } from 'vitest';
import { mocks, response, conversation, snapshot, deferred, store, capability } from './testFixture.ts';

describe('same-route assistant read intent', () => {
    it('keeps the last requested answer when versions finish in reverse order', async () => {
        await store.initialize();
        await store.open('1');
        const held = deferred<unknown>();
        mocks.api['message']!.mockReturnValueOnce(held.promise)
            .mockResolvedValueOnce(response(snapshot({ assistantMessageId: 'new' })));
        const old = store.viewVersion('old');
        await store.viewVersion('new');
        held.resolve(response(snapshot({ assistantMessageId: 'old' })));
        await old;
        expect(store.viewedVersions['2']?.assistantMessageId).toBe('new');
    });
    it('retains independent message version reads even when they finish in reverse order', async () => {
        await store.initialize();
        await store.open('1');
        const held = deferred<unknown>();
        mocks.api['message']!.mockReturnValueOnce(held.promise)
            .mockResolvedValueOnce(response(snapshot({ userMessageId: 'second', assistantMessageId: 'b' })));
        const first = store.viewVersion('a');
        await store.viewVersion('b');
        held.resolve(response(snapshot({ userMessageId: 'first', assistantMessageId: 'a' })));
        await first;
        expect(store.viewedVersions['first']?.assistantMessageId).toBe('a');
        expect(store.viewedVersions['second']?.assistantMessageId).toBe('b');
    });
    it.each(['success', 'failure'])('ignores stale memory %s after a newer refresh', async outcome => {
        await store.initialize();
        const held = deferred<unknown>();
        mocks.api['memories']!.mockReturnValueOnce(held.promise)
            .mockResolvedValueOnce(response({ items: [{ id: 'new' }], enabled: true }));
        const old = store.loadMemories();
        await store.loadMemories();
        if (outcome === 'success') held.resolve(response({ items: [{ id: 'old' }] }));
        else held.reject({ response: { status: 401 } });
        await old;
        expect(store.memories?.items[0]?.id).toBe('new');
        expect(store.authenticated).toBe(true);
    });
    it('keeps the latest version-page intent', async () => {
        await store.initialize();
        await store.open('1');
        const held = deferred<unknown>();
        mocks.api['versions']!.mockReturnValueOnce(held.promise)
            .mockResolvedValueOnce(response({ items: [{ id: 'new' }], nextCursor: 'next' }));
        const old = store.loadVersions('message', 'old-cursor');
        const fresh = await store.loadVersions('message', 'new-cursor');
        held.resolve(response({ items: [{ id: 'old' }] }));
        expect(await old).toBeUndefined();
        expect(fresh?.items[0]?.id).toBe('new');
    });
    it('does not replace a newer run snapshot with a captured pagination window', async () => {
        await store.initialize();
        mocks.api['get']!.mockResolvedValueOnce(response({ ...conversation(), hasMore: true, activeRun: snapshot() }));
        await store.open('1');
        const held = deferred<unknown>();
        mocks.api['get']!.mockReturnValueOnce(held.promise);
        const older = store.loadOlder();
        store.mergeSnapshot(snapshot({ revision: 9, content: 'new content' }));
        held.resolve(response({ ...conversation(), hasMore: false }));
        await older;
        expect(store.detail?.activeRun?.revision).toBe(9);
    });
    it('keeps latest history/search responses and suppresses obsolete errors', async () => {
        await store.initialize();
        const oldHistory = deferred<unknown>();
        mocks.api['list']!.mockReturnValueOnce(oldHistory.promise).mockResolvedValueOnce(response({ items: [{ id: 'new' }] }));
        const history = store.loadHistory();
        await store.loadHistory(undefined, true);
        oldHistory.reject({ response: { status: 401 } });
        await history;
        const oldSearch = deferred<unknown>();
        mocks.api['search']!.mockReturnValueOnce(oldSearch.promise).mockResolvedValueOnce(response({ items: [{ messageId: 'new' }] }));
        const search = store.search('old');
        await store.search('new');
        oldSearch.resolve(response({ items: [{ messageId: 'old' }] }));
        await search;
        expect(store.history.items[0]?.id).toBe('new');
        expect(store.searchResults.items[0]?.messageId).toBe('new');
    });
    it('reattaches polling after deletion epoch refresh and does not recreate it after disposal', async () => {
        vi.useFakeTimers();
        await store.initialize();
        store.attach();
        await store.synchronize();
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ deletionEpoch: 2 })));
        await store.synchronize();
        expect(vi.getTimerCount()).toBe(1);
        store.$dispose();
        expect(vi.getTimerCount()).toBe(0);
    });
    it('does not carry a failing run subscription backoff into a different run', async () => {
        vi.useFakeTimers();
        await store.initialize();
        store.attach();
        await store.synchronize();
        mocks.api['subscribe']!.mockRejectedValueOnce({ retryAfterMs: 60000 }).mockImplementation(() => new Promise(() => undefined));
        mocks.api['get']!.mockResolvedValueOnce(response({ ...conversation(), activeRun: snapshot() }));
        await store.open('1');
        // A direct microtask checkpoint completes the subscription catch/finally.
        await Promise.resolve();
        await Promise.resolve();
        mocks.api['get']!.mockResolvedValueOnce(response({ ...conversation('2'), activeRun: snapshot({ runId: 'new', conversationId: '2' }) }));
        await store.open('2');
        expect(mocks.api['subscribe']).toHaveBeenCalledTimes(2);
    });
    it('hydrates and keeps polling when capabilities discovers a different owner', async () => {
        vi.useFakeTimers();
        await store.initialize();
        store.attach();
        await store.synchronize();
        store.setHistoryVisible(true);
        await store.loadHistory();
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ ownerScope: 'B' })));
        mocks.api['list']!.mockResolvedValue(response({ items: [{ id: 'B-history' }] }));
        await store.synchronize();
        expect(store.capabilities?.ownerScope).toBe('B');
        expect(store.history.items[0]?.id).toBe('B-history');
        expect(vi.getTimerCount()).toBe(1);
    });

    it('does not let background history refresh supersede explicit pagination', async () => {
        vi.useFakeTimers();
        await store.initialize();
        store.setHistoryVisible(true);
        await store.loadHistory();
        vi.setSystemTime(Date.now() + 31000);
        const held = deferred<unknown>();
        mocks.api['list']!.mockReturnValueOnce(held.promise);
        const paging = store.loadHistory('page');
        const calls = mocks.api['list']!.mock.calls.length;
        await store.synchronize();
        expect(mocks.api['list']!.mock.calls.length).toBe(calls);
        held.resolve(response({ items: [{ id: 'older-page' }] }));
        await paging;
        expect(store.history.items[0]?.id).toBe('older-page');
    });

});

it('retains more than twenty loaded answers and refreshes duplicate metadata', async () => {
    await store.initialize();
    const versions = Array.from({ length: 40 }, (_, i) => ({ id: String(i + 1), attemptNumber: i + 1, state: 'completed' as const, createdAt: '' }));
    mocks.api['get']!.mockResolvedValueOnce(response({ ...conversation(), messages: [{ id: 'question', versions: versions.slice(20).map(version => version.id === '21' ? { ...version, state: 'running' } : version), hasMoreVersions: true }] }));
    await store.open('1');
    mocks.api['versions']!.mockResolvedValueOnce(response({ items: [...versions.slice(0, 21), { ...versions[21], state: 'running' }], nextCursor: undefined }));
    await store.loadVersions('question', 'older');
    const message = store.detail!.messages[0]!;
    expect(message.versions).toHaveLength(40);
    expect(message.versions.map(item => item.attemptNumber)).toEqual(Array.from({ length: 40 }, (_, i) => 40 - i));
    expect(message.versions.find(item => item.id === '21')?.state).toBe('completed');
    expect(message.versions.find(item => item.id === '22')?.state).toBe('completed');
    expect(message.hasMoreVersions).toBe(false);
});

it('preserves loaded answer pages and completion across server detail refreshes', async () => {
    await store.initialize();
    const versions = Array.from({ length: 45 }, (_, i) => ({ id: String(i + 1), attemptNumber: i + 1, state: 'completed' as const, createdAt: '' }));
    const detail = (metadataVersion: number) => ({ ...conversation('1', { metadataVersion }), messages: [{ id: 'question', versions: versions.slice(-5), hasMoreVersions: true }] });
    mocks.api['get']!.mockResolvedValueOnce(response(detail(1)));
    await store.open('1');
    mocks.api['versions']!.mockResolvedValueOnce(response({ items: versions.slice(25), nextCursor: 'older' }));
    await store.loadVersions('question');
    async function refresh(metadataVersion: number) {
        mocks.api['state']!.mockResolvedValueOnce(response({ id: '1', accessible: true, contextVersion: 1, metadataVersion }));
        mocks.api['get']!.mockResolvedValueOnce(response(detail(metadataVersion)));
        await store.synchronize();
    }
    await refresh(2);
    expect(store.detail!.messages[0]!.versions).toHaveLength(20);
    mocks.api['versions']!.mockResolvedValueOnce(response({ items: versions.slice(5, 25), nextCursor: 'oldest' }));
    await store.loadVersions('question', 'older');
    expect(store.detail!.messages[0]!.versions).toHaveLength(40);
    mocks.api['versions']!.mockResolvedValueOnce(response({ items: versions.slice(0, 5) }));
    await store.loadVersions('question', 'oldest');
    await refresh(3);
    expect(store.detail!.messages[0]!.versions).toHaveLength(45);
    expect(store.detail!.messages[0]!.hasMoreVersions).toBe(false);
});
