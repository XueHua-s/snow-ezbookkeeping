import { afterEach, expect, it, vi } from 'vitest';
import { createAssistantLocalStorage } from './assistant_local_storage';

const stores: ReturnType<typeof createAssistantLocalStorage>[] = [];
function open(ownerScope: string, deviceId?: string) {
    const store = createAssistantLocalStorage({ ownerScope, deviceId });
    stores.push(store);
    return store;
}
const draft = (text: string) => ({ text, conversationId: 'new', expectedDeletionEpoch: 1 });
afterEach(() => {
    vi.restoreAllMocks();
    for (const store of stores.splice(0)) store.close();
    localStorage.removeItem('assistant-device-id');
});

it('restores production drafts and exact pending when the localStorage identity disappears', async () => {
    const owner = crypto.randomUUID();
    const first = open(owner);
    await first.saveDraft(draft('  preserved\n'));
    const command = { clientRequestId: crypto.randomUUID(), action: 'send', expectedDeletionEpoch: 1,
        requestIssuedAt: '2026-10-15T08:00:00Z', text: ' exact ', payload: { content: ' exact ' } };
    await first.savePending(command);
    first.close();
    localStorage.removeItem('assistant-device-id');
    const restored = await open(owner).load();
    expect(restored.scope).toBe('device');
    expect(restored.drafts.map(item => item.text)).toEqual(['  preserved\n']);
    expect(restored.pending).toMatchObject([command]);
    expect((await open(crypto.randomUUID()).load()).drafts).toEqual([]);
});

it('recovers every legacy identity without overwriting conflicting drafts', async () => {
    const owner = crypto.randomUUID();
    for (const id of ['legacy-first-tab', 'legacy-second-tab']) {
        const legacy = open(owner, id);
        await legacy.saveDraft(draft(id));
        legacy.close();
    }
    localStorage.removeItem('assistant-device-id');
    const restored = await open(owner).load();
    expect(restored.drafts.map(item => item.text).sort()).toEqual(['legacy-first-tab', 'legacy-second-tab']);
    expect(restored.drafts.every(item => item.conflict)).toBe(true);
});

it('converges production initializers even when both legacy reads observe no identity', async () => {
    const owner = crypto.randomUUID();
    // Controlled first-use interleaving: both constructors see the pre-write value.
    vi.spyOn(Storage.prototype, 'getItem').mockReturnValue(null);
    const first = open(owner);
    const second = open(owner);
    await Promise.all([first.saveDraft(draft('first tab')), second.saveDraft(draft('second tab'))]);
    expect(first.deviceId).toBe(second.deviceId);
    const results = await Promise.all([first.load(), second.load()]);
    for (const result of results) {
        expect(result.scope).toBe('device');
        expect(result.drafts.map(item => item.text).sort()).toEqual(['first tab', 'second tab']);
    }
});

it('retains legacy branch receipts and enforces their deletion epoch after migration', async () => {
    const owner = crypto.randomUUID();
    const legacy = open(owner, 'lost-legacy-id');
    const command = { clientRequestId: crypto.randomUUID(), action: 'branch', expectedDeletionEpoch: 1,
        requestIssuedAt: '2026-10-15T08:00:00Z', text: '  branch\n', sourceConversationId: 'source',
        payload: { content: '  branch\n', expectedContextVersion: 4 } };
    await legacy.savePending(command);
    await legacy.resolveBranch(command.clientRequestId, 'target');
    legacy.close();
    const current = open(owner);
    const migrated = await current.load();
    expect(migrated.pending).toMatchObject([{ ...command, resolvedConversationId: 'target' }]);
    expect(migrated.drafts).toMatchObject([{ text: command.text, conversationId: 'target' }]);
    await current.savePending(command);
    expect((await current.load()).pending).toHaveLength(1);
    await current.reconcileEpoch(2);
    await expect(current.savePending(command)).rejects.toThrow('Stale local command epoch');
    expect((await open(owner).load()).pending).toEqual([]);
    expect((await open(owner).load()).drafts).toEqual([]);
});

it.each(['closed', 'blocked'])('closes a late IndexedDB connection after %s initialization', async mode => {
    const close = vi.fn();
    const request = { result: { close }, onsuccess: undefined, onblocked: undefined } as unknown as IDBOpenDBRequest;
    vi.spyOn(IDBFactory.prototype, 'open').mockReturnValue(request);
    const store = open(crypto.randomUUID());
    const loading = store.load();
    await vi.waitFor(() => expect(request.onsuccess).toBeTypeOf('function'));
    if (mode === 'closed') {
        store.close();
        const rejected = expect(loading).rejects.toThrow('closed');
        request.onsuccess!.call(request, new Event('success'));
        await rejected;
    } else {
        request.onblocked!.call(request, new Event('blocked') as IDBVersionChangeEvent);
        expect((await loading).scope).toBe('page');
        request.onsuccess!.call(request, new Event('success'));
    }
    expect(close).toHaveBeenCalledOnce();
});

it('notifies open instances once when a read migrates legacy data', async () => {
    const owner = crypto.randomUUID();
    const legacy = open(owner, 'legacy');
    await legacy.saveDraft(draft('legacy text'));
    legacy.close();
    const watcher = open(owner);
    const notices: number[] = [];
    watcher.subscribe(notice => notices.push(notice.version));
    const reader = open(owner);
    const migrated = await reader.load();
    await vi.waitFor(() => expect(notices).toEqual([migrated.version]));
    const unchanged = await reader.load();
    expect(unchanged.version).toBe(migrated.version);
    expect(notices).toEqual([migrated.version]);
});
