import { expect, it, vi } from 'vitest';
import type { AssistantLocalPending, AssistantLocalSnapshot } from '@/lib/assistant_local_storage.ts';
import { mocks, deferred, local, setLocal, response, store } from './testFixture.ts';

it('marks normal sending before persistence publishes and throughout draft persistence and network waits', async () => {
    await store.initialize(); await store.setDraft('question');
    const persist = deferred<AssistantLocalSnapshot>();
    const draft = deferred<AssistantLocalSnapshot>();
    const network = deferred<unknown>();
    mocks.save.mockReturnValueOnce(persist.promise);
    const storage = mocks.local.mock.results[0]!.value;
    vi.spyOn(storage, 'saveDraft').mockReturnValueOnce(draft.promise);
    mocks.api['send']!.mockReturnValueOnce(network.promise);
    const sending = store.send();
    const item = mocks.save.mock.calls[0]![0] as AssistantLocalPending;
    expect(store.delivery[item.clientRequestId]).toBe('sending');
    expect(store.submitting).toBe(true);
    await expect(store.send()).resolves.toBeUndefined();
    setLocal({ ...local, pending: [item] }); persist.resolve(local);
    await vi.waitFor(() => expect(store.pending).toHaveLength(1));
    expect(store.delivery[item.clientRequestId]).toBe('sending');
    expect(mocks.api['send']).not.toHaveBeenCalled();
    await store.synchronize();
    expect(mocks.api['command']).not.toHaveBeenCalled();
    draft.resolve(local);
    await vi.waitFor(() => expect(mocks.api['send']).toHaveBeenCalledTimes(1));
    await store.synchronize();
    expect(store.delivery[item.clientRequestId]).toBe('sending');
    expect(mocks.api['command']).not.toHaveBeenCalled();
    network.reject(new Error('lost response'));
    await expect(sending).rejects.toThrow('lost response');
    expect(store.delivery[item.clientRequestId]).toBe('unsent');
    expect(store.pending).toHaveLength(1);
    expect(store.submitting).toBe(false);
});

it('removes transient sending state after persistence failure', async () => {
    await store.initialize(); await store.setDraft('question');
    mocks.save.mockRejectedValueOnce(new Error('disk failure'));
    await expect(store.send()).rejects.toThrow('disk failure');
    expect(store.delivery).toEqual({});
    expect(store.submitting).toBe(false);
    expect(mocks.api['send']).not.toHaveBeenCalled();
});

it('does not retain sending state after session invalidation during persistence', async () => {
    await store.initialize(); await store.setDraft('question');
    const persist = deferred<AssistantLocalSnapshot>();
    mocks.save.mockReturnValueOnce(persist.promise);
    const sending = store.send();
    store.reset(false);
    persist.resolve(local);
    await sending;
    expect(store.delivery).toEqual({});
    expect(mocks.api['send']).not.toHaveBeenCalled();
});

it('keeps submission active while the newly accepted conversation loads', async () => {
    await store.initialize(); await store.setDraft('question');
    const loading = deferred<unknown>();
    mocks.api['send']!.mockResolvedValueOnce(response({ status: 'accepted', conversationId: 'new' }));
    mocks.api['get']!.mockReturnValueOnce(loading.promise);
    const sending = store.send();
    await vi.waitFor(() => expect(store.loading).toBe(true));
    expect(store.submitting).toBe(true);
    loading.resolve(response({ conversation: { id: 'new' }, messages: [] }));
    await sending;
    expect(store.submitting).toBe(false);
    expect(store.loading).toBe(false);
    expect(store.delivery).toEqual({});
});

it('keeps explicit retries recoverable when persisting a structured rejection fails', async () => {
    await store.initialize(); await store.setDraft('question');
    await expect(store.send()).rejects.toThrow('lost response');
    const id = store.pending[0]!.clientRequestId;
    expect(store.delivery[id]).toBe('unsent');
    const storage = mocks.local.mock.results[0]!.value;
    vi.spyOn(storage, 'markPendingRejected').mockRejectedValueOnce(new Error('rejection persistence failed'));
    mocks.api['send']!.mockRejectedValueOnce({ response: { status: 409, data: {
        success: false, errorCode: 1, context: { reasonCode: 'context_changed' }
    } } });
    await expect(store.retryPending(id)).rejects.toThrow('rejection persistence failed');
    expect(store.pending).toHaveLength(1);
    expect(store.delivery[id]).toBeUndefined();
    const priorChecks = mocks.api['command']!.mock.calls.length;
    await store.synchronize();
    expect(mocks.api['command']!.mock.calls.length).toBeGreaterThan(priorChecks);
    expect(store.delivery[id]).toBe('unsent');
});

it('does not let an old generation clear a recovered retry with the same request id', async () => {
    await store.initialize(); await store.setDraft('question');
    const oldRequest = deferred<unknown>();
    mocks.api['send']!.mockReturnValueOnce(oldRequest.promise);
    const original = store.send();
    await vi.waitFor(() => expect(mocks.api['send']).toHaveBeenCalledTimes(1));
    const id = store.pending[0]!.clientRequestId;
    store.reset(false);
    await store.initialize();
    await store.synchronize();
    expect(store.delivery[id]).toBe('unsent');
    const newRequest = deferred<unknown>();
    mocks.api['send']!.mockReturnValueOnce(newRequest.promise);
    const retry = store.retryPending(id);
    await vi.waitFor(() => expect(mocks.api['send']).toHaveBeenCalledTimes(2));
    expect(store.delivery[id]).toBe('sending');
    oldRequest.resolve(response({ status: 'accepted', conversationId: 'old-generation' }));
    await original;
    expect(store.delivery[id]).toBe('sending');
    const checks = mocks.api['command']!.mock.calls.length;
    await store.synchronize();
    expect(mocks.api['command']!.mock.calls.length).toBe(checks);
    newRequest.reject(new Error('retry lost response'));
    await expect(retry).rejects.toThrow('retry lost response');
    expect(store.delivery[id]).toBe('unsent');
});
