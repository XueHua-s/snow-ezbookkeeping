/** Local drafts are never an automatic send queue. Callers must reconcile pending IDs first. */
export interface AssistantLocalDraft {
    id: string;
    conversationId: string;
    deviceId: string;
    editorId: string;
    revision: number;
    text: string;
    expectedDeletionEpoch: number;
    conflict: boolean;
}

export type AssistantLocalJSON = string | number | boolean | null | AssistantLocalJSON[] | { [key: string]: AssistantLocalJSON | undefined };

export interface AssistantLocalPending {
    clientRequestId: string;
    requestIssuedAt: string;
    expectedDeletionEpoch: number;
    action: string;
    /** Server definitively rejected this command; never inferred from a transport failure. */
    rejected?: boolean;
    /** Accepted branch target retained until its route has actually opened. */
    resolvedConversationId?: string;
    text?: string;
    conversationId?: string;
    sourceConversationId?: string;
    sourceMessageId?: string;
    expectedContextVersion?: number;
    /** Exact request metadata, without tokens or credentials. */
    target?: Record<string, AssistantLocalJSON | undefined>;
    source?: Record<string, AssistantLocalJSON | undefined>;
    /** Complete canonical request, including action-specific versions and necessary text only. */
    payload?: Record<string, AssistantLocalJSON | undefined>;
}

interface OwnerData {
    ownerScope: string;
    epoch?: number;
    version: number;
    drafts: AssistantLocalDraft[];
    pending: (AssistantLocalPending & { deviceId: string })[];
}

export interface AssistantLocalSnapshot {
    /** Monotonic completion order within one storage instance, including fallback/recovery. */
    sequence?: number;
    drafts: AssistantLocalDraft[];
    pending: AssistantLocalPending[];
    scope: 'device' | 'page';
    reason?: 'unavailable' | 'quota';
    version: number;
}

export interface AssistantLocalNotice {
    ownerScope: string;
    version: number;
}

export interface AssistantDraftInput {
    conversationId: string;
    text: string;
    expectedDeletionEpoch: number;
    revision?: number;
    id?: string;
}

const DATABASE = 'ezbookkeeping-assistant-local-v1';
const STORE = 'owners';
const empty = (ownerScope: string): OwnerData => ({ ownerScope, version: 0, drafts: [], pending: [] });
const copy = <T>(value: T): T => JSON.parse(JSON.stringify(value)) as T;

/** The origin's IndexedDB is the device boundary; localStorage can disappear independently.
 * Normalize legacy tab IDs inside the owner transaction, retaining every record.
 */
function migrateDevice(data: OwnerData, deviceId: string): boolean {
    const changed = [...data.drafts, ...data.pending].some(item => item.deviceId !== deviceId);
    if (!changed) return false;
    for (const item of [...data.drafts, ...data.pending]) item.deviceId = deviceId;
    for (const draft of data.drafts) {
        if (data.drafts.some(other => other !== draft && other.conversationId === draft.conversationId)) {
            draft.conflict = true;
        }
    }
    return true;
}

function openDatabase(): Promise<IDBDatabase> {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open(DATABASE, 1);
        let blocked = false;
        request.onupgradeneeded = () => request.result.createObjectStore(STORE, { keyPath: 'ownerScope' });
        request.onsuccess = () => {
            if (blocked) request.result.close();
            else resolve(request.result);
        };
        request.onerror = () => reject(request.error);
        request.onblocked = () => {
            blocked = true;
            reject(new Error('Local storage upgrade blocked'));
        };
    });
}

class LocalQuotaError extends Error {}
export class AssistantLocalEpochError extends Error {}

function checkQuota(data: OwnerData): void {
    if (data.drafts.length > 20 || data.pending.filter(item => !item.rejected && (item.text !== undefined || ['send', 'retry', 'branch'].includes(item.action))).length > 4 ||
        new TextEncoder().encode(JSON.stringify(data)).byteLength > 512 * 1024) {
        throw new LocalQuotaError('Local assistant storage full');
    }
}

function requireEpoch(data: OwnerData, epoch: number): void {
    if (data.epoch !== undefined && data.epoch !== epoch) throw new AssistantLocalEpochError('Stale local command epoch');
    data.epoch = epoch;
}

/** Reject credential containers before copying: callers cannot accidentally persist an HTTP config. */
function validatePayload(value: unknown, depth = 0): void {
    if (depth > 12) throw new TypeError('Assistant command payload nesting limit');
    if (value === undefined || value === null || typeof value === 'string' || typeof value === 'boolean') return;
    if (typeof value === 'number' && Number.isFinite(value)) return;
    if (Array.isArray(value)) { for (const child of value) validatePayload(child, depth + 1); return; }
    if (typeof value !== 'object' || Object.getPrototypeOf(value) !== Object.prototype) throw new TypeError('Assistant command payload must be JSON');
    for (const [key, child] of Object.entries(value)) {
        if (/token|password|secret|authorization|cookie|credential|headers|apikey|api_key/i.test(key)) {
            throw new TypeError('Credentials cannot be persisted with assistant commands');
        }
        validatePayload(child, depth + 1);
    }
}

function mergeRecovery(current: OwnerData, local: OwnerData, base: OwnerData): void {
    if (current.epoch !== undefined && current.epoch !== local.epoch) throw new AssistantLocalEpochError('Reconcile epoch before restoring local content');
    current.epoch = local.epoch;
    const equal = (a: unknown, b: unknown) => JSON.stringify(a) === JSON.stringify(b);
    for (const draft of local.drafts) {
        const previous = base.drafts.find(d => d.id === draft.id);
        if (equal(previous, draft)) continue;
        const remote = current.drafts.find(d => d.id === draft.id);
        if (equal(remote, draft)) continue;
        const restored = copy(draft);
        if (remote && !equal(remote, previous)) {
            remote.conflict = true;
            restored.conflict = true;
            restored.id = crypto.randomUUID();
        } else if (remote) current.drafts = current.drafts.filter(d => d.id !== remote.id);
        for (const other of current.drafts) {
            if (other.conversationId === restored.conversationId && other.deviceId === restored.deviceId) {
                other.conflict = true;
                restored.conflict = true;
            }
        }
        current.drafts.push(restored);
    }
    for (const previous of base.drafts) {
        if (!local.drafts.some(d => d.id === previous.id)) {
            current.drafts = current.drafts.filter(d => d.id !== previous.id || !equal(d, previous));
        }
    }
    for (const pending of local.pending) {
        const previous = base.pending.find(p => p.clientRequestId === pending.clientRequestId);
        if (equal(previous, pending)) continue;
        const remote = current.pending.find(p => p.clientRequestId === pending.clientRequestId);
        if (remote && !equal({ ...remote, rejected: undefined, resolvedConversationId: undefined }, { ...pending, rejected: undefined, resolvedConversationId: undefined })) {
            throw new AssistantLocalEpochError('Conflicting pending command');
        }
        if (remote && pending.rejected) remote.rejected = true;
        if (remote && pending.resolvedConversationId) {
            if (remote.resolvedConversationId && remote.resolvedConversationId !== pending.resolvedConversationId) {
                throw new AssistantLocalEpochError('Conflicting branch target');
            }
            remote.resolvedConversationId = pending.resolvedConversationId;
        }
        if (!remote && !previous) current.pending.push(copy(pending));
    }
    for (const previous of base.pending) {
        if (!local.pending.some(p => p.clientRequestId === previous.clientRequestId)) {
            current.pending = current.pending.filter(p => p.clientRequestId !== previous.clientRequestId);
        }
    }
}

/** One owner row makes checks and writes atomic across tabs, including the total byte quota. */
export function createAssistantLocalStorage(options: { ownerScope: string; deviceId?: string; editorId?: string }) {
    const { ownerScope } = options;
    const deviceId = options.deviceId ?? 'browser';
    const editorId = options.editorId ?? crypto.randomUUID();
    let memory = empty(ownerScope);
    let baseline = copy(memory);
    let reason: AssistantLocalSnapshot['reason'];
    let database: IDBDatabase | undefined;
    let closed = false;
    let sequence = 0;
    let queue: Promise<unknown> = Promise.resolve();
    let channel: BroadcastChannel | undefined;
    const listeners = new Set<(notice: AssistantLocalNotice) => void>();
    try {
        channel = new BroadcastChannel(DATABASE);
        channel.onmessage = (event: MessageEvent<AssistantLocalNotice>) => {
            if (event.data?.ownerScope === ownerScope && typeof event.data.version === 'number') {
                for (const listener of listeners) listener({ ownerScope, version: event.data.version });
            }
        };
    } catch { /* Polling remains available when BroadcastChannel is disabled. */ }

    function snapshot(): AssistantLocalSnapshot {
        return { sequence: ++sequence, drafts: copy(memory.drafts.filter(d => d.deviceId === deviceId)),
            pending: copy(memory.pending.filter(p => p.deviceId === deviceId)),
            scope: reason ? 'page' : 'device', reason, version: memory.version };
    }

    function transaction(mutate?: (data: OwnerData) => void, cleanup = false): Promise<AssistantLocalSnapshot> {
        const task = async () => {
            if (closed) throw new Error('Assistant local storage closed');
            const retained = reason && cleanup ? copy(memory) : undefined;
            const retainedReason = reason;
            let migrated = false;
            if (!reason || cleanup) {
                let candidate: OwnerData | undefined;
                try {
                    database ??= await openDatabase();
                    if (closed) {
                        database.close();
                        throw new Error('Assistant local storage closed');
                    }
                    await new Promise<void>((resolve, reject) => {
                        const tx = database!.transaction(STORE, mutate || options.deviceId === undefined ? 'readwrite' : 'readonly');
                        const store = tx.objectStore(STORE);
                        const request = store.get(ownerScope);
                        let failure: unknown;
                        request.onsuccess = () => {
                            candidate = request.result as OwnerData | undefined ?? empty(ownerScope);
                            migrated = options.deviceId === undefined && migrateDevice(candidate, deviceId);
                            // Keep the latest persisted text if quota validation fails below.
                            memory = copy(candidate);
                            if (!retained) baseline = copy(candidate);
                            try {
                                if (mutate || migrated) {
                                    mutate?.(candidate);
                                    candidate.version++;
                                    checkQuota(candidate);
                                    store.put(candidate);
                                }
                            } catch (error) { failure = error; tx.abort(); }
                        };
                        tx.oncomplete = () => resolve();
                        tx.onabort = tx.onerror = () => reject(failure ?? tx.error ?? request.error);
                    });
                    memory = candidate!;
                    if (retained && mutate) {
                        mutate(retained);
                        retained.version = memory.version;
                        memory = retained;
                        reason = retainedReason;
                        mutate(baseline);
                    } else baseline = copy(memory);
                } catch (error) {
                    if (closed || error instanceof AssistantLocalEpochError || error instanceof TypeError) throw error;
                    reason = error instanceof LocalQuotaError ? 'quota' : 'unavailable';
                    if (mutate) {
                        const fallback = copy(retained ?? memory);
                        mutate(fallback);
                        fallback.version++;
                        memory = fallback;
                    }
                }
            } else if (mutate) {
                const fallback = copy(memory);
                mutate(fallback);
                fallback.version++;
                memory = fallback;
            }
            if ((mutate || migrated) && !reason) channel?.postMessage({ ownerScope, version: memory.version });
            return snapshot();
        };
        const result = queue.then(task);
        queue = result.catch(() => undefined);
        return result;
    }

    function retryPersistence(): Promise<AssistantLocalSnapshot> {
        const task = async () => {
            if (closed) throw new Error('Assistant local storage closed');
            if (!reason) return snapshot();
            try {
                database ??= await openDatabase();
                if (closed) {
                    database.close();
                    throw new Error('Assistant local storage closed');
                }
                let recovered: OwnerData | undefined;
                await new Promise<void>((resolve, reject) => {
                    const tx = database!.transaction(STORE, 'readwrite');
                    const store = tx.objectStore(STORE);
                    const request = store.get(ownerScope);
                    let failure: unknown;
                    request.onsuccess = () => {
                        try {
                            recovered = request.result as OwnerData | undefined ?? empty(ownerScope);
                            if (options.deviceId === undefined) migrateDevice(recovered, deviceId);
                            mergeRecovery(recovered, memory, baseline);
                            recovered.version++;
                            checkQuota(recovered);
                            store.put(recovered);
                        } catch (error) { failure = error; tx.abort(); }
                    };
                    tx.oncomplete = () => resolve();
                    tx.onabort = tx.onerror = () => reject(failure ?? tx.error ?? request.error);
                });
                memory = recovered!;
                baseline = copy(memory);
                reason = undefined;
                channel?.postMessage({ ownerScope, version: memory.version });
            } catch (error) {
                if (closed || error instanceof AssistantLocalEpochError || error instanceof TypeError) throw error;
                reason = error instanceof LocalQuotaError ? 'quota' : 'unavailable';
            }
            return snapshot();
        };
        const result = queue.then(task);
        queue = result.catch(() => undefined);
        return result;
    }

    return {
        deviceId, editorId,
        load: () => transaction(),
        retryPersistence,
        saveDraft: (input: AssistantDraftInput) => transaction(data => {
            requireEpoch(data, input.expectedDeletionEpoch);
            const matches = data.drafts.filter(d => d.deviceId === deviceId && d.conversationId === input.conversationId);
            const current = matches.find(d => d.id === input.id) ?? matches.find(d => d.editorId === editorId) ?? matches[0];
            const conflict = !!current && current.editorId !== editorId && current.revision !== input.revision;
            if (current && !conflict) data.drafts = data.drafts.filter(d => d.id !== current.id);
            if (!input.text.trim()) return;
            if (conflict) current.conflict = true;
            data.drafts.push({ id: conflict || !current ? crypto.randomUUID() : current.id,
                conversationId: input.conversationId, deviceId, editorId,
                revision: (current?.revision ?? 0) + 1, text: input.text,
                expectedDeletionEpoch: input.expectedDeletionEpoch, conflict });
        }),
        savePending: (input: AssistantLocalPending) => transaction(data => {
            validatePayload(input);
            requireEpoch(data, input.expectedDeletionEpoch);
            const prior = data.pending.find(p => p.clientRequestId === input.clientRequestId);
            if (prior) {
                if (JSON.stringify({ ...input, deviceId, rejected: undefined, resolvedConversationId: undefined }) !== JSON.stringify({ ...prior, rejected: undefined, resolvedConversationId: undefined })) {
                    throw new AssistantLocalEpochError('Pending command ID cannot be reused with different content');
                }
                return;
            }
            data.pending.push({ ...copy(input), deviceId });
        }),
        markPendingRejected: (id: string) => transaction(data => {
            const pending = data.pending.find(p => p.clientRequestId === id);
            if (pending) pending.rejected = true;
        }, true),
        resolveBranch: (id: string, targetId: string) => transaction(data => {
            const pending = data.pending.find(p => p.clientRequestId === id && p.deviceId === deviceId);
            if (!pending || pending.action !== 'branch' || !targetId) throw new TypeError('Expected a pending branch and target');
            requireEpoch(data, pending.expectedDeletionEpoch);
            if (pending.resolvedConversationId && pending.resolvedConversationId !== targetId) {
                throw new AssistantLocalEpochError('Branch command target cannot change');
            }
            const existing = data.drafts.some(d => d.deviceId === deviceId && d.conversationId === targetId && d.text.trim());
            if (pending.text?.trim() && !existing) {
                data.drafts.push({ id: crypto.randomUUID(), conversationId: targetId, deviceId, editorId,
                    revision: 1, text: pending.text, expectedDeletionEpoch: pending.expectedDeletionEpoch, conflict: false });
            }
            pending.resolvedConversationId = targetId;
        }),
        removePending: (id: string) => transaction(data => {
            data.pending = data.pending.filter(p => p.clientRequestId !== id);
        }, true),
        removeDraft: (id: string) => transaction(data => {
            data.drafts = data.drafts.filter(d => d.id !== id);
        }, true),
        migrateDraft: (from: string, to: string) => transaction(data => {
            for (const draft of data.drafts) {
                if (draft.deviceId === deviceId && draft.conversationId === from) {
                    draft.conversationId = to;
                    draft.revision++;
                    for (const other of data.drafts) {
                        if (other !== draft && other.deviceId === deviceId && other.conversationId === to) {
                            other.conflict = true;
                            draft.conflict = true;
                        }
                    }
                }
            }
        }),
        clearConversation: (id: string) => transaction(data => {
            data.drafts = data.drafts.filter(d => d.conversationId !== id);
            data.pending = data.pending.filter(p => p.conversationId !== id && p.sourceConversationId !== id && p.resolvedConversationId !== id);
        }, true),
        clearOwner: () => transaction(data => {
            data.drafts = [];
            data.pending = [];
        }, true),
        reconcileEpoch: (epoch: number) => transaction(data => {
            if (data.epoch !== undefined && epoch < data.epoch) throw new AssistantLocalEpochError('Stale identity reconciliation');
            if (data.epoch !== undefined && data.epoch !== epoch) {
                data.drafts = [];
                // A global clear fences assistant writes before ledger deletion finishes.
                // Retain only its body-free receipt for status lookup, never for replay.
                data.pending = data.pending.filter(item => item.action === 'clearAllData' && item.text === undefined)
                    .map(item => ({
                        clientRequestId: item.clientRequestId, requestIssuedAt: item.requestIssuedAt,
                        expectedDeletionEpoch: item.expectedDeletionEpoch, deviceId: item.deviceId,
                        action: item.action, rejected: item.rejected,
                        payload: { clientRequestId: item.clientRequestId, requestIssuedAt: item.requestIssuedAt,
                            expectedDeletionEpoch: item.expectedDeletionEpoch }
                    }));
            }
            data.epoch = epoch;
        }, true),
        subscribe(listener: (notice: AssistantLocalNotice) => void) {
            listeners.add(listener);
            return () => { listeners.delete(listener); };
        },
        close() { closed = true; database?.close(); channel?.close(); listeners.clear(); memory = empty(ownerScope); }
    };
}
