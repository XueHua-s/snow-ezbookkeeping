import { afterEach, describe, expect, it, vi } from 'vitest';
import { AssistantLocalEpochError, createAssistantLocalStorage } from './assistant_local_storage';

const stores: ReturnType<typeof createAssistantLocalStorage>[] = [];
function storage(ownerScope: string = crypto.randomUUID(), editorId: string = crypto.randomUUID(), deviceId = 'device') {
    const result = createAssistantLocalStorage({ ownerScope, editorId, deviceId });
    stores.push(result);
    return result;
}
const draft = (text: string, conversationId = 'new') => ({ text, conversationId, expectedDeletionEpoch: 1 });
const pending = (clientRequestId = crypto.randomUUID()) => ({ clientRequestId, action: 'send',
    requestIssuedAt: '2026-10-15T08:00:00Z', expectedDeletionEpoch: 1, text: '  原文\n',
    expectedContextVersion: 4, sourceMessageId: 'source', payload: { mode: 'chat', content: '  原文\n' } });

afterEach(() => { vi.restoreAllMocks(); for (const store of stores.splice(0)) store.close(); });

describe('native IndexedDB assistant storage', () => {
    it('restores exact pending commands and drafts after closing and reopening', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        await first.saveDraft(draft('  提问\n'));
        const command = pending();
        await first.savePending(command);
        first.close();
        const second = storage(owner);
        const loaded = await second.load();
        expect(loaded.scope).toBe('device');
        expect(loaded.drafts[0]?.text).toBe('  提问\n');
        expect(loaded.pending[0]).toMatchObject(command);
        await second.removePending(command.clientRequestId);
        expect((await storage(owner).load()).pending).toEqual([]);
    });

    it('isolates owners and devices, and migrates new drafts without losing text', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        await first.saveDraft(draft('draft'));
        expect((await storage().load()).drafts).toEqual([]);
        expect((await storage(owner, 'other', 'other-device').load()).drafts).toEqual([]);
        await first.migrateDraft('new', 'formal');
        expect((await first.load()).drafts[0]).toMatchObject({ conversationId: 'formal', text: 'draft' });
    });

    it('preserves stale-tab conflicts as independent copies', async () => {
        const owner = crypto.randomUUID();
        const foreground = storage(owner, 'foreground');
        const background = storage(owner, 'background');
        const old = (await foreground.saveDraft(draft('initial'))).drafts[0]!;
        await foreground.saveDraft({ ...draft('actively typing'), revision: old.revision });
        const conflict = await background.saveDraft({ ...draft('stale tab edit'), revision: old.revision });
        expect(conflict.drafts.map(d => d.text).sort()).toEqual(['actively typing', 'stale tab edit']);
        expect(conflict.drafts.every(d => d.conflict)).toBe(true);
    });

    it('enforces owner-wide transactional draft limits without evicting old text', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        for (let i = 0; i < 20; i++) await first.saveDraft(draft(`text ${i}`, `${i}`));
        const full = await first.saveDraft(draft('retained on this page', '21'));
        expect(full).toMatchObject({ scope: 'page', reason: 'quota' });
        expect(full.drafts).toHaveLength(21);
        expect((await storage(owner).load()).drafts).toHaveLength(20);
        // Cleanup must still remove durable data after falling back to page memory.
        await first.clearOwner();
        expect((await storage(owner).load()).drafts).toEqual([]);
        expect((await first.load()).drafts).toEqual([]);
    });

    it('serializes concurrent tabs at the quota boundary', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        const second = storage(owner);
        for (let i = 0; i < 19; i++) await first.saveDraft(draft(`text ${i}`, `${i}`));
        const results = await Promise.all([
            first.saveDraft(draft('first concurrent', 'first')),
            second.saveDraft(draft('second concurrent', 'second'))
        ]);
        expect(results.map(r => r.scope).sort()).toEqual(['device', 'page']);
        expect((await storage(owner).load()).drafts).toHaveLength(20);
        expect(results.find(r => r.scope === 'page')?.drafts).toHaveLength(21);
    });

    it('refuses to change the canonical command under an existing request ID', async () => {
        const first = storage();
        const command = pending();
        await first.savePending(command);
        await expect(first.savePending({ ...command, text: 'changed' })).rejects.toBeInstanceOf(AssistantLocalEpochError);
        expect((await first.load()).pending[0]?.text).toBe(command.text);
    });

    it('recovers page drafts after cleanup without overwriting another tab', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner, 'first');
        const second = storage(owner, 'second');
        for (let i = 0; i < 20; i++) await first.saveDraft(draft(`text ${i}`, `${i}`));
        const initial = await first.load();
        await first.saveDraft(draft('page-only new draft', 'new-page'));
        await first.saveDraft({ ...draft('local changed text', '0'), id: initial.drafts[0]!.id });
        await second.saveDraft({ ...draft('other tab changed text', '0'), revision: initial.drafts[0]!.revision });
        // Make room for both the newly created draft and the conflict copy.
        await first.removeDraft(initial.drafts[1]!.id);
        await first.removeDraft(initial.drafts[2]!.id);
        const recovered = await first.retryPersistence();
        expect(recovered.scope).toBe('device');
        const restored = await storage(owner).load();
        expect(restored.drafts.map(d => d.text)).toContain('page-only new draft');
        expect(restored.drafts.filter(d => d.conversationId === '0').map(d => d.text).sort())
            .toEqual(['local changed text', 'other tab changed text']);
    });

    it('persists nested canonical JSON but rejects credential containers', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        const command = { ...pending(), action: 'saveMemory', payload: {
            sourceMessageIds: ['message-a', 'message-b'], details: { selected: true, versions: [1, 2] }
        } };
        await first.savePending(command);
        expect((await storage(owner).load()).pending[0]?.payload).toEqual(command.payload);
        await expect(first.savePending({ ...pending(), payload: { nested: { authorization: 'private' } } }))
            .rejects.toThrow('Credentials cannot be persisted');
        expect((await first.load()).pending).toHaveLength(1);
    });

    it('retains definitive rejection across refresh without allowing a changed command', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        const command = pending();
        await first.savePending(command);
        await first.markPendingRejected(command.clientRequestId);
        first.close();
        const reopened = storage(owner);
        expect((await reopened.load()).pending[0]).toMatchObject({ ...command, rejected: true });
        await reopened.savePending(command);
        expect((await reopened.load()).pending[0]?.rejected).toBe(true);
        await expect(reopened.savePending({ ...command, payload: { content: 'changed' } }))
            .rejects.toBeInstanceOf(AssistantLocalEpochError);
        expect((await reopened.load()).pending[0]?.payload).toEqual(command.payload);
    });

    it('atomically restores accepted branch target and original text after refresh', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        const command = { ...pending(), action: 'branch', sourceConversationId: 'source-conversation' };
        await first.savePending(command);
        await first.resolveBranch(command.clientRequestId, 'target');
        first.close();
        const reopened = storage(owner);
        const restored = await reopened.load();
        expect(restored.pending[0]).toMatchObject({ ...command, resolvedConversationId: 'target' });
        expect(restored.drafts[0]).toMatchObject({ conversationId: 'target', text: command.text });
        await reopened.savePending(command);
        expect((await reopened.load()).pending[0]?.resolvedConversationId).toBe('target');
        await reopened.resolveBranch(command.clientRequestId, 'target');
        expect((await reopened.load()).drafts).toHaveLength(1);
        await expect(reopened.resolveBranch(command.clientRequestId, 'other-target')).rejects.toBeInstanceOf(AssistantLocalEpochError);
    });

    it('does not overwrite an existing branch target draft', async () => {
        const first = storage();
        const command = { ...pending(), action: 'branch' };
        await first.savePending(command);
        await first.saveDraft(draft('newer target text', 'target'));
        const restored = await first.resolveBranch(command.clientRequestId, 'target');
        expect(restored.drafts).toHaveLength(1);
        expect(restored.drafts[0]?.text).toBe('newer target text');
        expect(restored.pending[0]).toMatchObject({ text: command.text, resolvedConversationId: 'target' });
    });

    it('removes a branch receipt when its resolved target is deleted before route acknowledgement', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        const command = { ...pending(), action: 'branch', sourceConversationId: 'source' };
        await first.savePending(command);
        await first.resolveBranch(command.clientRequestId, 'target');
        await first.clearConversation('target');
        const restored = await storage(owner).load();
        expect(restored.pending).toEqual([]);
        expect(restored.drafts).toEqual([]);
    });

    it('limits sends to four and counts UTF-8 metadata bytes', async () => {
        const first = storage();
        for (let i = 0; i < 4; i++) await first.savePending(pending());
        const full = await first.savePending(pending());
        expect(full.scope).toBe('page');
        expect(full.pending).toHaveLength(5);
        const bytes = await storage().saveDraft(draft('中'.repeat(180_000)));
        expect(bytes.reason).toBe('quota');
        expect(bytes.drafts[0]?.text.length).toBe(180_000);
    });

    it('clears all device records on epoch changes and rejects stale commands', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        const stale = storage(owner);
        await first.savePending(pending());
        await first.saveDraft(draft('old'));
        await stale.load();
        await first.reconcileEpoch(2);
        await expect(stale.savePending(pending())).rejects.toBeInstanceOf(AssistantLocalEpochError);
        expect((await first.load()).drafts).toEqual([]);
        expect((await first.load()).pending).toEqual([]);
        await expect(first.reconcileEpoch(1)).rejects.toBeInstanceOf(AssistantLocalEpochError);
    });

    it('retains only a body-free global clear receipt across the deletion fence', async () => {
        const owner = crypto.randomUUID();
        const first = storage(owner);
        const command = { clientRequestId: crypto.randomUUID(), requestIssuedAt: '2026-10-15T08:00:00Z', expectedDeletionEpoch: 1 };
        await first.saveDraft(draft('must clear'));
        await first.savePending(pending());
        await first.savePending({ ...command, action: 'clearAllData', payload: { ...command, unnecessary: 'remove this' } });
        await first.savePending({ ...pending(), action: 'clearAllData' });
        await first.reconcileEpoch(2);
        first.close();
        const reopened = storage(owner);
        const restored = await reopened.load();
        expect(restored.drafts).toEqual([]);
        expect(restored.pending).toHaveLength(1);
        expect(restored.pending[0]).toMatchObject({ ...command, action: 'clearAllData', payload: command });
        expect(restored.pending[0]?.payload).not.toHaveProperty('unnecessary');
        // Receipt retention must not relax epoch fencing for reissuing that command.
        await expect(reopened.savePending({ ...command, action: 'clearAllData', payload: command }))
            .rejects.toBeInstanceOf(AssistantLocalEpochError);
        await reopened.clearOwner();
        expect((await storage(owner).load()).pending).toEqual([]);
    });

    it('clears deleted conversation and related branch commands only', async () => {
        const first = storage();
        await first.saveDraft(draft('delete', 'a'));
        await first.saveDraft(draft('keep', 'b'));
        await first.savePending({ ...pending(), sourceConversationId: 'a' });
        await first.clearConversation('a');
        const loaded = await first.load();
        expect(loaded.pending).toEqual([]);
        expect(loaded.drafts.map(d => d.text)).toEqual(['keep']);
    });

    it('reports page-only retention when IndexedDB is denied', async () => {
        vi.spyOn(IDBFactory.prototype, 'open').mockImplementation(() => { throw new DOMException('denied', 'SecurityError'); });
        const first = storage();
        const saved = await first.saveDraft(draft('never silently discarded'));
        expect(saved).toMatchObject({ scope: 'page', reason: 'unavailable' });
        expect((await first.load()).drafts[0]?.text).toBe('never silently discarded');
    });

    it('broadcasts only owner identity and revision, never body text', async () => {
        const owner = crypto.randomUUID();
        const receiver = storage(owner);
        const notices: unknown[] = [];
        receiver.subscribe(notice => notices.push(notice));
        await storage(owner).saveDraft(draft('private finance data'));
        await vi.waitFor(() => expect(notices.length).toBeGreaterThan(0));
        expect(Object.keys(notices[0] as object).sort()).toEqual(['ownerScope', 'version']);
        expect(JSON.stringify(notices)).not.toContain('private');
    });
});
