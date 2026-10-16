import { describe, expect, it, vi } from 'vitest';
import type { AssistantLocalSnapshot } from '@/lib/assistant_local_storage.ts';
import { mocks, response, capability, conversation, snapshot, deferred, local, store, setLocal } from './testFixture.ts';

async function switchIdentity() {
    store.reset(false);
    setLocal({ drafts: [], pending: [], scope: 'device', version: 0 });
    mocks.api['capabilities']!.mockResolvedValue(response(capability({ ownerScope: 'B' })));
    await store.initialize();
}

describe('assistant operation authority', () => {
    it.each(['navigation', 'identity'])('rejects delayed 404 cleanup after %s', async boundary => {
        await store.initialize();
        await store.open('1');
        const started = deferred<void>();
        const held = deferred<AssistantLocalSnapshot>();
        const owner = mocks.local.mock.results[0]!.value;
        vi.spyOn(owner, 'clearConversation').mockImplementation(() => { started.resolve(); return held.promise; });
        mocks.api['state']!.mockRejectedValueOnce({ response: { status: 404 } });
        const syncing = store.synchronize();
        await started.promise;
        if (boundary === 'identity') await switchIdentity();
        await store.open('2');
        await store.setDraft('new draft');
        const current = store.local;
        held.resolve({ drafts: [], pending: [], scope: 'page', version: 999 });
        await syncing;
        expect(store.local).toBe(current);
        expect(store.draft).toBe('new draft');
        expect(store.conversationId).toBe('2');
    });
    it('ignores old initialize 401 after a new owner is established', async () => {
        const held = deferred<unknown>();
        mocks.api['capabilities']!.mockReturnValueOnce(held.promise);
        const initializing = store.initialize();
        await switchIdentity();
        held.reject({ response: { status: 401 } });
        await initializing;
        expect(store.authenticated).toBe(true);
        expect(store.capabilities?.ownerScope).toBe('B');
        expect(store.error).toBeUndefined();
    });
    it('keeps a new send submitting when an old send finishes', async () => {
        await store.initialize();
        const oldSave = deferred<AssistantLocalSnapshot>();
        mocks.save.mockReturnValueOnce(oldSave.promise);
        const oldSend = store.send({ content: 'old' });
        await switchIdentity();
        const newSave = deferred<AssistantLocalSnapshot>();
        mocks.save.mockReturnValueOnce(newSave.promise);
        const newSend = store.send({ content: 'new' });
        oldSave.resolve(local);
        await oldSend;
        expect(store.submitting).toBe(true);
        store.reset(false);
        newSave.resolve(local);
        await newSend;
    });
    it('reset clears loading and error immediately without waiting for reads', async () => {
        await store.initialize();
        const held = deferred<unknown>();
        mocks.api['get']!.mockReturnValueOnce(held.promise);
        const reading = store.open('1');
        store.error = new Error('old error');
        store.reset(false);
        expect(store.loading).toBe(false);
        expect(store.error).toBeUndefined();
        held.reject(new Error('late error'));
        await reading;
        expect(store.error).toBeUndefined();
    });
    it('retains newer same-owner local snapshots when an older cleanup finishes late', async () => {
        await store.initialize();
        const owner = mocks.local.mock.results[0]!.value;
        const held = deferred<AssistantLocalSnapshot>();
        vi.spyOn(owner, 'removeDraft').mockReturnValueOnce(held.promise);
        const removing = store.removeDraft('old');
        vi.spyOn(owner, 'saveDraft').mockResolvedValueOnce({ ...local, sequence: 12, version: 12 });
        await store.setDraft('new draft');
        const current = store.local;
        held.resolve({ drafts: [], pending: [], scope: 'device', sequence: 11, version: 11 });
        await removing;
        expect(store.local).toBe(current);
    });
    it('accepts a newer snapshot even when fallback recovery lowers its persisted version', async () => {
        await store.initialize();
        const owner = mocks.local.mock.results[0]!.value;
        vi.spyOn(owner, 'saveDraft').mockResolvedValueOnce({ ...local, scope: 'page', sequence: 12, version: 100 });
        await store.setDraft('offline edit');
        vi.spyOn(owner, 'saveDraft').mockResolvedValueOnce({ ...local, scope: 'page', sequence: 13, version: 5 });
        await store.setDraft('cleanup retained edit');
        expect(store.local.version).toBe(5);
        expect(store.local.sequence).toBe(13);
    });
    it('does not dispatch after a microtask identity change between draft persistence and caller continuation', async () => {
        await store.initialize();
        await store.setDraft('old');
        const owner = mocks.local.mock.results[0]!.value;
        vi.spyOn(owner, 'saveDraft').mockImplementationOnce(async () => {
            queueMicrotask(() => queueMicrotask(() => store.reset(false)));
            return local;
        });
        await store.send();
        expect(mocks.api['send']).not.toHaveBeenCalled();
    });
    it('does not continue receipt cleanup after a queued owner change', async () => {
        await store.initialize();
        const owner = mocks.local.mock.results[0]!.value;
        const removing = vi.spyOn(owner, 'removePending');
        vi.spyOn(owner, 'migrateDraft').mockImplementationOnce(async () => {
            queueMicrotask(() => store.reset(false));
            return local;
        });
        mocks.api['send']!.mockResolvedValueOnce(response({ status: 'accepted', conversationId: 'new' }));
        await store.send({ content: 'old' });
        expect(removing).not.toHaveBeenCalled();
    });
    it('restarts hydration after same-owner deletion epoch changes', async () => {
        await store.initialize();
        await store.open('1');
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ deletionEpoch: 2 })));
        mocks.api['list']!.mockResolvedValueOnce(response({ items: [{ id: 'fresh' }] }));
        await store.initialize();
        expect(store.capabilities?.deletionEpoch).toBe(2);
        expect(store.conversationId).toBe('');
        expect(store.history.items[0]?.id).toBe('fresh');
    });
    it('disposal closes owner/listener and prevents late initialization from recreating resources', async () => {
        await store.initialize();
        const owner = mocks.local.mock.results[0]!.value;
        const held = deferred<unknown>();
        mocks.api['capabilities']!.mockReturnValueOnce(held.promise);
        const initializing = store.initialize();
        store.$dispose();
        held.resolve(response(capability()));
        await initializing;
        store.attach();
        expect(owner.close).toHaveBeenCalledTimes(1);
        expect(mocks.local).toHaveBeenCalledTimes(1);
        expect(store.authenticated).toBe(false);
        expect(mocks.api['subscribe']).not.toHaveBeenCalled();
    });
    it('disposal aborts a stream without canceling its server run', async () => {
        await store.initialize();
        mocks.api['get']!.mockResolvedValueOnce(response({ ...conversation(), activeRun: snapshot() }));
        let signal: AbortSignal | undefined;
        mocks.api['subscribe']!.mockImplementation((_id, _revision, _callbacks, value: AbortSignal) => {
            signal = value;
            return new Promise(() => undefined);
        });
        store.attach();
        await store.open('1');
        store.$dispose();
        expect(signal?.aborted).toBe(true);
        expect(mocks.api['cancel']).not.toHaveBeenCalled();
    });
    it('refreshes visible history after a clear receipt advances the deletion epoch', async () => {
        await store.initialize();
        store.setHistoryVisible(true);
        await store.loadHistory();
        mocks.api['clear']!.mockResolvedValueOnce(response({ status: 'accepted', deletionEpoch: 2 }));
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ deletionEpoch: 2 })));
        mocks.api['list']!.mockResolvedValue(response({ items: [{ id: 'fresh-after-clear' }] }));
        await store.execute('clear', {});
        expect(store.capabilities?.deletionEpoch).toBe(2);
        expect(store.history.items[0]?.id).toBe('fresh-after-clear');
    });
    it('does not navigate from a receipt helper after an intervening microtask resets identity', async () => {
        await store.initialize();
        await store.open('1');
        const owner = mocks.local.mock.results[0]!.value;
        let crossedBoundary = false;
        vi.spyOn(owner, 'resolveBranch').mockImplementationOnce(() => Promise.resolve(local).then(saved => {
            queueMicrotask(() => queueMicrotask(() => {
                store.reset(false);
                crossedBoundary = true;
            }));
            return saved;
        }));
        const navigations: boolean[] = [];
        mocks.api['get']!.mockImplementation(async ({ id }: { id: string }) => {
            navigations.push(crossedBoundary);
            return response(conversation(id));
        });
        mocks.api['branch']!.mockResolvedValueOnce(response({ status: 'accepted', conversationId: 'target' }));
        await store.branch('message', 'edit', 'text');
        expect(navigations).not.toContain(true);
    });

    it('persists a late branch receipt without navigating away from newer route intent', async () => {
        await store.initialize();
        await store.open('1');
        const started = deferred<void>();
        const held = deferred<unknown>();
        mocks.api['branch']!.mockImplementationOnce(() => { started.resolve(); return held.promise; });
        const branching = store.branch('message', 'edit', 'text');
        await started.promise;
        await store.open();
        held.resolve(response({ status: 'accepted', conversationId: 'target' }));
        await branching;
        expect(store.conversationId).toBe('');
        expect(store.local.pending[0]?.resolvedConversationId).toBe('target');
    });

});
