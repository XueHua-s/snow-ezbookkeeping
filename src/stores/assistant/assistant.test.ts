import { describe, expect, it, vi } from 'vitest';
import type { AssistantLocalSnapshot, AssistantLocalPending } from '@/lib/assistant_local_storage.ts';
import { mocks, response, capability, conversation, snapshot, deferred, local, store, setLocal } from './testFixture.ts';

describe('assistant authoritative state', () => {
    it('reads current references without changing draft or reading position and fences identity changes', async () => {
        await store.initialize(); await store.setDraft('preserved'); store.savePosition({ messageId: 'm', offset: 12, following: false });
        mocks.transaction.mockResolvedValueOnce(response({ id: 'transaction' }));
        expect(await store.getReferenceTransaction('transaction')).toEqual({ id: 'transaction' });
        expect(mocks.transaction).toHaveBeenCalledWith({ id: 'transaction', withPictures: false }, { ignoreError: true });
        expect(store.draft).toBe('preserved'); expect(store.readingPosition()?.offset).toBe(12);
        const held = deferred<unknown>(); mocks.transaction.mockReturnValueOnce(held.promise);
        const pending = store.getReferenceTransaction('transaction'); store.reset(false); held.resolve(response({ id: 'transaction' }));
        await expect(pending).rejects.toMatchObject({ name: 'AbortError' });
    });
    it('keeps persisted pending commands on reference 401', async () => {
        await store.initialize(); await store.setDraft('pending'); await expect(store.send()).rejects.toThrow();
        const requestId = local.pending[0]!.clientRequestId;
        mocks.transaction.mockRejectedValueOnce({ response: { status: 401 } });
        await expect(store.getReferenceTransaction('transaction')).rejects.toMatchObject({ response: { status: 401 } });
        expect(store.authenticated).toBe(false); expect(local.pending[0]?.clientRequestId).toBe(requestId);
    });
    it('persists exact original command before HTTP and never automatically resends unknown outcomes', async () => {
        await store.initialize(); await store.setDraft('  question\n');
        const held = deferred<AssistantLocalSnapshot>(); mocks.save.mockReturnValueOnce(held.promise);
        const sending = store.send();
        expect(mocks.api['send']).not.toHaveBeenCalled();
        const item = mocks.save.mock.calls[0]![0] as AssistantLocalPending;
        setLocal({ ...local, pending: [item] }); held.resolve(local);
        await expect(sending).rejects.toThrow('lost response');
        expect(item.text).toBe('  question\n'); expect(item.payload?.['content']).toBe(item.text);
        expect(store.pending).toHaveLength(1); expect(store.delivery[item.clientRequestId]).toBe('unsent');
        await store.synchronize(); await store.initialize();
        expect(mocks.api['send']).toHaveBeenCalledTimes(1);
        await expect(store.retryPending(item.clientRequestId)).rejects.toThrow('lost response');
        expect(mocks.api['send']!.mock.calls[1]![0]).toEqual(mocks.api['send']!.mock.calls[0]![0]);
    });
    it('prevents double-click before the asynchronous persistence finishes', async () => {
        await store.initialize(); await store.setDraft('question');
        const held = deferred<AssistantLocalSnapshot>(); mocks.save.mockReturnValueOnce(held.promise);
        const first = store.send(); await expect(store.send()).resolves.toBeUndefined();
        held.resolve(local); await expect(first).rejects.toThrow(); expect(mocks.save).toHaveBeenCalledTimes(1);
    });
    it('keeps the next draft when a send fails and offers original recovery without overwriting it', async () => {
        await store.initialize(); await store.setDraft('original');
        const held = deferred<unknown>(); mocks.api['send']!.mockReturnValueOnce(held.promise);
        const sending = store.send(); await vi.waitFor(() => expect(mocks.api['send']).toHaveBeenCalled());
        await store.setDraft('next question'); held.reject(new Error('offline')); await expect(sending).rejects.toThrow();
        expect(store.draft).toBe('next question'); expect(await store.restorePending(store.pending[0]!.clientRequestId)).toBe(false);
    });
    it('ignores stale navigation and identity callbacks', async () => {
        await store.initialize(); const held = deferred<unknown>(); mocks.api['get']!.mockReturnValueOnce(held.promise);
        const first = store.open('1'); await store.open('2'); held.resolve(response(conversation('1'))); await first;
        expect(store.detail?.conversation.id).toBe('2');
        const late = deferred<unknown>(); mocks.api['get']!.mockReturnValueOnce(late.promise);
        const reading = store.open('3'); store.reset(); late.resolve(response(conversation('3'))); await reading;
        expect(store.detail).toBeUndefined(); expect(store.draft).toBe('');
    });
    it.each(['navigation', 'identity'])('fences inaccessible cleanup after %s changes', async boundary => {
        await store.initialize();
        await store.open('1');
        const held = deferred<AssistantLocalSnapshot>();
        const owner = mocks.local.mock.results[0]!.value;
        const clearing = vi.spyOn(owner, 'clearConversation').mockReturnValueOnce(held.promise);
        mocks.api['state']!.mockResolvedValueOnce(response({ id: '1', accessible: false }));
        const syncing = store.synchronize();
        await vi.waitFor(() => expect(clearing).toHaveBeenCalledWith('1'));
        if (boundary === 'identity') {
            store.reset(false);
            mocks.api['capabilities']!.mockResolvedValue(response(capability({ ownerScope: 'B' })));
            await store.initialize();
        }
        await store.open('2');
        await store.setDraft('new context draft');
        const currentLocal = store.local;
        held.resolve({ drafts: [], pending: [], scope: 'page', version: 999 });
        await syncing;
        expect(store.conversationId).toBe('2');
        expect(store.detail?.conversation.id).toBe('2');
        expect(store.draft).toBe('new context draft');
        expect(store.local).toBe(currentLocal);
    });
    it('ignores an old visible-history 401 after switching identities', async () => {
        await store.initialize();
        const held = deferred<unknown>();
        mocks.api['list']!.mockReturnValueOnce(held.promise);
        store.setHistoryVisible(true);
        store.reset(false);
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ ownerScope: 'B' })));
        await store.initialize();
        held.reject({ response: { status: 401 } });
        // Drain the full result -> history -> visibility -> store error-handler chain.
        await new Promise(resolve => setTimeout(resolve, 0));
        expect(store.authenticated).toBe(true);
        expect(store.capabilities?.ownerScope).toBe('B');
        expect(store.error).toBeUndefined();
    });
    it.each(['identity', 'new request'])('ignores a stale search failure after %s', async boundary => {
        await store.initialize();
        const held = deferred<unknown>();
        mocks.api['search']!.mockReturnValueOnce(held.promise);
        const pending = store.search('old');
        if (boundary === 'identity') {
            store.reset(false);
            await store.initialize();
        }
        mocks.api['search']!.mockResolvedValueOnce(response({ items: [{ messageId: 'new' }] }));
        await store.search('new');
        held.reject(new Error('old connection failed'));
        await expect(pending).resolves.toBeUndefined();
        expect(store.searchResults.items[0]?.messageId).toBe('new');
    });
    it('does not accept old revisions or regress a terminal snapshot', async () => {
        await store.initialize(); await store.open('1');
        store.mergeSnapshot(snapshot({ revision: 4, state: 'completed', content: 'done' }));
        store.mergeSnapshot(snapshot({ revision: 5, state: 'running' }));
        store.mergeSnapshot(snapshot({ revision: 2 }));
        expect(store.detail?.activeRun?.state).toBe('completed'); expect(store.detail?.activeRun?.content).toBe('done');
    });
    it('keeps selected historical versions read-only when later questions exist', async () => {
        await store.initialize(); await store.open('1');
        await store.selectVersion('old-question', 'old-answer'); expect(mocks.api['selectVersion']).not.toHaveBeenCalled();
    });
    it('preserves local-only storage warning, and rejects Unicode/byte over-limit without clearing text', async () => {
        await store.initialize(); const text = '😀'.repeat(4001); await store.setDraft(text);
        await expect(store.send()).rejects.toThrow('input_too_long'); expect(store.draft).toBe(text);
        setLocal({ ...local, scope: 'page', reason: 'quota' }); await store.initialize(); expect(store.unsaved).toBe(true);
    });
    it('resolves management uncertainty with the original ID and never repeats clear on reconnect', async () => {
        await store.initialize(); mocks.api['clear']!.mockRejectedValueOnce(new Error('lost clear'));
        await expect(store.execute('clear', {})).rejects.toThrow(); const item = local.pending[0]!;
        expect(item.action).toBe('clear'); expect(item.payload?.['expectedDeletionEpoch']).toBe(1);
        await store.synchronize(); expect(mocks.api['clear']).toHaveBeenCalledTimes(1);
    });
    it('does not send draft on online recovery and clears stale epoch pending commands', async () => {
        await store.initialize(); await store.setDraft('offline draft');
        await store.synchronize(); expect(mocks.api['send']).not.toHaveBeenCalled();
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ deletionEpoch: 2 })));
        await store.synchronize(); expect(store.local.pending).toHaveLength(0); expect(store.draft).toBe('');
    });
    it('creates a branch without implicitly calling the provider and retains its target draft', async () => {
        await store.initialize(); await store.open('1');
        mocks.api['branch']!.mockResolvedValue(response({ conversationId: '4', status: 'accepted' }));
        await store.branch('2', 'edit', 'edited question');
        expect(store.conversationId).toBe('4'); expect(store.draft).toBe('edited question'); expect(mocks.api['send']).not.toHaveBeenCalled();
    });
    it('retains edited text and the original branch command across unknown creation outcome', async () => {
        await store.initialize(); await store.open('1');
        mocks.api['branch']!.mockRejectedValueOnce(new Error('lost branch'));
        await expect(store.branch('2', 'edit', 'recover edited text')).rejects.toThrow();
        const saved = local.pending[0]!; expect(saved.text).toBe('recover edited text');
        mocks.api['command']!.mockResolvedValue(response({ conversationId: '4', status: 'accepted' }));
        await store.reconcilePending(); expect(store.conversationId).toBe('4'); expect(store.draft).toBe('recover edited text');
        expect(mocks.api['branch']).toHaveBeenCalledTimes(1); expect(mocks.api['send']).not.toHaveBeenCalled();
    });
    it('never dispatches a persisted command after identity reset', async () => {
        await store.initialize(); await store.setDraft('question');
        const held = deferred<AssistantLocalSnapshot>(); mocks.save.mockReturnValueOnce(held.promise);
        const sending = store.send(); store.reset(); held.resolve(local); await sending;
        expect(mocks.api['send']).not.toHaveBeenCalled(); expect(store.local.pending).toHaveLength(0);
    });
    it('clears only the deleted current conversation when its state read returns 404', async () => {
        await store.initialize(); await store.open('1'); await store.setDraft('draft');
        mocks.api['state']!.mockRejectedValueOnce({ response: { status: 404 } });
        await store.synchronize(); expect(store.conversationId).toBe(''); expect(store.detail).toBeUndefined();
    });
    it('marks evicted latest messages and reloads the real tail within the bound', async () => {
        const page = (first: number, last: number) => ({ ...conversation(), messages: Array.from({ length: last - first + 1 }, (_, index) => ({ id: String(first + index), turnSequence: first + index })), hasMore: first > 1 });
        await store.initialize(); mocks.api['get']!.mockResolvedValueOnce(response(page(91, 120))); await store.open('1');
        for (const first of [61, 31, 1]) { mocks.api['get']!.mockResolvedValueOnce(response(page(first, first + 29))); await store.loadOlder(); }
        expect(store.detail?.messages).toHaveLength(90); expect(store.detail?.messages.at(-1)?.turnSequence).toBe(90);
        expect(store.windowMissingLatest).toBe(true); expect(store.updatesAvailable).toBe(false);
        mocks.api['get']!.mockResolvedValueOnce(response(page(91, 120))); await store.loadLatest();
        expect(store.windowMissingLatest).toBe(false); expect(store.detail?.messages.at(-1)?.turnSequence).toBe(120);
    });
    it('removes a resolved branch marker only after authoritative target 404', async () => {
        local.pending.push({ clientRequestId: 'branch-id', action: 'branch', conversationId: 'source', resolvedConversationId: 'target', expectedDeletionEpoch: 1, payload: {} } as AssistantLocalPending);
        await store.initialize(); await store.open('target'); await store.setDraft('edited text');
        mocks.api['get']!.mockRejectedValueOnce(new Error('offline'));
        await store.open('target'); expect(store.local.pending).toHaveLength(1); expect(store.local.drafts).toHaveLength(1);
        mocks.api['get']!.mockRejectedValueOnce({ response: { status: 404 } });
        await store.open('target'); expect(store.local.pending).toHaveLength(0); expect(store.local.drafts).toHaveLength(0);
        expect(store.conversationId).toBe(''); expect(store.recoverBranchNavigation('source')).toBeUndefined();
    });
    it('does not let an older search or archive request replace the latest selection', async () => {
        await store.initialize(); const oldSearch = deferred<unknown>(); const oldList = deferred<unknown>();
        mocks.api['search']!.mockReturnValueOnce(oldSearch.promise).mockResolvedValueOnce(response({ items: [{ messageId: 'new' }] }));
        const first = store.search('old'); await store.search('new'); oldSearch.resolve(response({ items: [{ messageId: 'old' }] })); await first;
        expect(store.searchResults.items[0]?.messageId).toBe('new');
        mocks.api['list']!.mockReturnValueOnce(oldList.promise).mockResolvedValueOnce(response({ items: [{ id: 'archived' }] }));
        const previous = store.loadHistory(); await store.loadHistory(undefined, true); oldList.resolve(response({ items: [{ id: 'active' }] })); await previous;
        expect(store.history.items[0]?.id).toBe('archived');
    });
    it('does not clear an identical draft in another conversation while persistence waits', async () => {
        await store.initialize(); await store.open('1'); await store.setDraft('same text');
        const held = deferred<AssistantLocalSnapshot>(); mocks.save.mockReturnValueOnce(held.promise);
        const sending = store.send(); await store.open('2'); await store.setDraft('same text');
        held.resolve(local); await expect(sending).rejects.toThrow(); expect(store.draft).toBe('same text'); expect(store.conversationId).toBe('2');
    });
    it('preserves pending commands but hides all prior identity content on 401', async () => {
        await store.initialize(); await store.setDraft('important original');
        mocks.api['send']!.mockRejectedValueOnce({ response: { status: 401 } });
        await expect(store.send()).rejects.toMatchObject({ response: { status: 401 } });
        expect(store.authenticated).toBe(false); expect(store.detail).toBeUndefined(); expect(store.draft).toBe('');
        expect(local.pending[0]?.text).toBe('important original');
    });
    it('switches to same-run polling after three failed subscriptions without new sends', async () => {
        vi.useFakeTimers(); await store.initialize();
        mocks.api['get']!.mockResolvedValue(response({ ...conversation('1', { activeRunId: '3' }), activeRun: snapshot() }));
        mocks.api['state']!.mockResolvedValue(response({ accessible: true, contextVersion: 1, metadataVersion: 1, activeRunId: '3', activeRunRevision: 2 }));
        mocks.api['run']!.mockResolvedValue(response(snapshot({ revision: 2 })));
        mocks.api['subscribe']!.mockRejectedValue(new Error('disconnect'));
        await store.open('1'); store.attach(); await store.synchronize();
        for (let i = 0; i < 4; i++) { vi.setSystemTime(Date.now() + 16000); await store.synchronize(); await Promise.resolve(); }
        expect(mocks.api['subscribe']).toHaveBeenCalledTimes(3);
        expect(mocks.api['state']!.mock.calls.length).toBeGreaterThan(3); expect(mocks.api['send']).not.toHaveBeenCalled();
    });
    it('keeps the bounded older window during synchronization and explicitly restores latest', async () => {
        await store.initialize();
        const page = (from: number, to: number, version = 1) => ({ ...conversation('1', { contextVersion: version }), hasMore: true,
            messages: Array.from({ length: to - from + 1 }, (_, index) => ({ id: String(from + index), turnSequence: from + index, versions: [] })) });
        mocks.api['get']!.mockResolvedValueOnce(response(page(91, 120))).mockResolvedValueOnce(response(page(61, 90)))
            .mockResolvedValueOnce(response(page(31, 60))).mockResolvedValueOnce(response(page(1, 30)));
        await store.open('1'); await store.loadOlder(); await store.loadOlder(); await store.loadOlder();
        expect(store.detail?.messages).toHaveLength(90); expect(store.detail?.messages[0]?.id).toBe('1');
        mocks.api['state']!.mockResolvedValue(response({ accessible: true, contextVersion: 2, metadataVersion: 1 }));
        mocks.api['get']!.mockResolvedValue(response(page(92, 121, 2)));
        await store.synchronize(); expect(store.detail?.messages[0]?.id).toBe('1'); expect(store.updatesAvailable).toBe(true);
        await store.loadLatest(); expect(store.detail?.messages[0]?.id).toBe('92'); expect(store.updatesAvailable).toBe(false);
    });
    it('uses server clock and coalesces duplicate management writes before persistence', async () => {
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ serverTime: '2030-01-01T00:00:00Z' })));
        await store.initialize(); const held = deferred<unknown>(); mocks.api['clear']!.mockReturnValueOnce(held.promise);
        const first = store.execute('clear', {}); const second = store.execute('clear', {});
        await vi.waitFor(() => expect(mocks.api['clear']).toHaveBeenCalledTimes(1));
        expect(local.pending[0]?.requestIssuedAt.startsWith('2030-01-01')).toBe(true);
        held.reject(new Error('lost clear')); await expect(first).rejects.toThrow(); await expect(second).rejects.toThrow();
        expect(mocks.save).toHaveBeenCalledTimes(1);
    });
    it('persists password-free global data-clear metadata without mistaking assistant receipt for all-data success', async () => {
        await store.initialize(); const original = await store.prepareDataClear();
        expect(local.pending[0]?.action).toBe('clearAllData'); expect(local.pending[0]?.payload).toEqual(original);
        const retry = await store.prepareDataClear(); expect(retry).toEqual(original);
        mocks.api['command']!.mockResolvedValue(response({ status: 'accepted', deletionEpoch: 2 }));
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ deletionEpoch: 2 })));
        await expect(store.prepareDataClear()).rejects.toThrow('Check the remaining user data');
        expect(mocks.save).toHaveBeenCalledTimes(1);
        await store.retryPending(original.clientRequestId); expect(mocks.api['clear']).not.toHaveBeenCalled();
    });
    it('allows corrected input after a persisted structured rejection but not an untrusted proxy status', async () => {
        await store.initialize(); await store.setDraft('first question');
        mocks.api['send']!.mockRejectedValueOnce({ response: { status: 409, data: { success: false, errorCode: 209061, context: { reasonCode: 'conversation_conflict' } } } });
        await expect(store.send()).rejects.toBeDefined(); expect(store.pending[0]?.rejected).toBe(true); expect(store.activePending).toHaveLength(0);
        await store.setDraft('corrected question'); expect(store.canSend).toBe(true);
        mocks.api['send']!.mockRejectedValueOnce({ response: { status: 409 } });
        await expect(store.send()).rejects.toBeDefined(); expect(store.activePending).toHaveLength(1);
        const rejected = store.pending.find(item => item.rejected)!;
        await store.discardRejected(rejected.clientRequestId); expect(store.pending).toHaveLength(1);
    });
    it('only starts a new all-data clear after explicit confirmation of the incomplete original', async () => {
        await store.initialize(); const original = await store.prepareDataClear();
        await expect(store.prepareDataClear(true)).rejects.toThrow('Confirm the original');
        mocks.api['command']!.mockResolvedValue(response({ status: 'accepted', operationCompleted: false }));
        await expect(store.prepareDataClear()).rejects.toThrow('Check the remaining');
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ deletionEpoch: 2 })));
        const next = await store.prepareDataClear(true); expect(next.clientRequestId).not.toBe(original.clientRequestId); expect(next.expectedDeletionEpoch).toBe(2);
    });
    it('reuses an operation-completed global receipt instead of starting a new clear', async () => {
        await store.initialize(); const original = await store.prepareDataClear();
        mocks.api['command']!.mockResolvedValue(response({ status: 'accepted', operationCompleted: true }));
        expect(await store.prepareDataClear()).toEqual(original); expect(mocks.save).toHaveBeenCalledTimes(1);
    });
    it('keeps incomplete all-data confirmation across epoch advance and initialization', async () => {
        await store.initialize(); const original = await store.prepareDataClear();
        mocks.api['command']!.mockResolvedValue(response({ status: 'accepted', operationCompleted: false }));
        mocks.api['capabilities']!.mockResolvedValue(response(capability({ deletionEpoch: 2 })));
        await store.initialize(); expect(store.local.pending[0]?.clientRequestId).toBe(original.clientRequestId);
        await expect(store.prepareDataClear()).rejects.toThrow('Check the remaining');
        expect(mocks.api['clear']).not.toHaveBeenCalled();
    });
    it('persists branch target and draft before navigation, keeping a recoverable marker until route acknowledgement', async () => {
        await store.initialize(); await store.open('1');
        mocks.api['branch']!.mockResolvedValue(response({ conversationId: '4', status: 'accepted' }));
        mocks.api['get']!.mockRejectedValueOnce(new Error('navigation disconnected'));
        await store.branch('2', 'edit', 'durable edited text');
        expect(local.pending[0]?.resolvedConversationId).toBe('4');
        expect(local.drafts.find(draft => draft.conversationId === '4')?.text).toBe('durable edited text');
        await store.acknowledgeBranchNavigation('4'); expect(local.pending).toHaveLength(1);
        store.reset(false); await store.initialize();
        expect(store.recoverBranchNavigation('1')).toBe('4'); expect(store.recoverBranchNavigation('unrelated')).toBeUndefined();
        await store.open('4'); expect(store.draft).toBe('durable edited text');
        expect(local.pending).toHaveLength(1);
        await store.acknowledgeBranchNavigation('4'); expect(local.pending).toHaveLength(0);
    });
    it('leaving a page aborts only subscription, not the server task', async () => {
        await store.initialize(); mocks.api['get']!.mockResolvedValue(response({ ...conversation('1'), activeRun: snapshot() }));
        let signal: AbortSignal | undefined;
        mocks.api['subscribe']!.mockImplementation((_id, _revision, _callbacks, input: AbortSignal) => { signal = input; return new Promise(() => undefined); });
        store.attach(); await store.open('1'); store.detach();
        expect(signal?.aborted).toBe(true); expect(mocks.api['cancel']).not.toHaveBeenCalled(); expect(store.detail).toBeDefined();
    });
});
