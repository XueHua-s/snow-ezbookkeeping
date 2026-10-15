import type { AssistantRunSnapshot, AssistantRunState } from '@/models/ai_assistant_conversation.ts';

export const ASSISTANT_SNAPSHOT_FRAME_BYTES = 256 * 1024;
const terminalStates: AssistantRunState[] = ['completed', 'incomplete', 'failed', 'canceled', 'interrupted'];
const reasons = ['queue_timeout', 'run_timeout', 'execution_limit', 'generation_unavailable', 'input_too_long', 'conversation_conflict', 'data_epoch_conflict', 'memory_conflict', 'context_changed', 'already_running', 'quota_exceeded', 'queue_full', 'subscription_limit', 'cursor_expired', 'context_limit', 'output_limit', 'empty_response', 'storage_unavailable', 'provider_unavailable', 'interrupted', 'not_accessible'];
const states: AssistantRunState[] = ['queued', 'preparing', 'running', 'cancel_requested', ...terminalStates];

export class AssistantSubscriptionError extends Error {
    constructor(public readonly kind: 'protocol' | 'disconnected' | 'unauthorized' | 'limited' | 'http' | 'canceled',
        public readonly status?: number, public readonly retryAfterMs?: number, public readonly envelope?: unknown) {
        super(`Assistant subscription ${kind}`);
    }
}

export interface AssistantSnapshotCallbacks {
    readonly onSnapshot: (snapshot: AssistantRunSnapshot) => void;
    readonly onTerminal?: (snapshot: AssistantRunSnapshot) => void;
}

function object(value: unknown): value is Record<string, unknown> {
    return value !== null && typeof value === 'object' && !Array.isArray(value);
}
function id(value: unknown): boolean {
    return typeof value === 'string' && /^[1-9][0-9]*$/.test(value) && BigInt(value) <= 9223372036854775807n;
}
function amount(value: unknown): boolean {
    return typeof value === 'string' && /^(0|-?[1-9][0-9]*)$/.test(value) &&
        BigInt(value) >= -9223372036854775808n && BigInt(value) <= 9223372036854775807n;
}
function integer(value: unknown, min: number): boolean {
    return typeof value === 'number' && Number.isSafeInteger(value) && value >= min;
}

/** Validate the persisted wire projection without ever coercing decimal IDs or money to number. */
export function isAssistantRunSnapshot(value: unknown): value is AssistantRunSnapshot {
    if (!object(value) || value['schemaVersion'] !== 1 || !id(value['runId']) ||
        value['runId'] !== value['assistantMessageId'] || !id(value['conversationId']) || !id(value['userMessageId']) ||
        !integer(value['attemptNumber'], 1) || !integer(value['revision'], 1) ||
        !states.includes(value['state'] as AssistantRunState) || typeof value['content'] !== 'string' ||
        typeof value['savedAt'] !== 'string' || !Number.isFinite(Date.parse(value['savedAt'])) ||
        !(value['reasonCode'] === null || reasons.includes(value['reasonCode'] as string)) ||
        !Array.isArray(value['queries']) || !Array.isArray(value['references'])) {
        return false;
    }
    if (!['progress', 'processSummary', 'pageableQueryId'].every(key => value[key] === undefined || typeof value[key] === 'string')) { return false; }
    if (value['state'] === 'completed' && !value['content'].trim()) {
        return false;
    }
    return value['references'].every(reference => object(reference) && id(reference['id']) &&
        integer(reference['time'], Number.MIN_SAFE_INTEGER) && integer(reference['type'], 0) && amount(reference['sourceAmountMinor']) &&
        (reference['destinationAmountMinor'] === undefined || amount(reference['destinationAmountMinor'])) &&
        (reference['hideAmount'] === undefined || typeof reference['hideAmount'] === 'boolean') &&
        ['timeText', 'categoryName', 'sourceAccountName', 'destinationAccountName', 'currency', 'destinationCurrency', 'comment'].every(key =>
            reference[key] === undefined || typeof reference[key] === 'string')) &&
        value['queries'].every(query => object(query) && typeof query['queryId'] === 'string' &&
            typeof query['tool'] === 'string' && object(query['scope']) &&
            typeof query['scope']['timezone'] === 'string' && typeof query['scope']['includesHidden'] === 'boolean' &&
            typeof query['scope']['excludesDeleted'] === 'boolean' &&
            (query['scope']['startTime'] === null || typeof query['scope']['startTime'] === 'string') &&
            (query['scope']['endTime'] === null || typeof query['scope']['endTime'] === 'string') &&
            Array.isArray(query['warnings']) && query['warnings'].every(warning => typeof warning === 'string') &&
            ['page', 'all_matching'].includes(query['coverage'] as string) && typeof query['queriedAt'] === 'string');
}

/** The byte buffer includes headers and delimiters, and never retains more than one frame. */
export async function readAssistantConversationStream(body: ReadableStream<Uint8Array>, runId: string,
    callbacks: AssistantSnapshotCallbacks, signal?: AbortSignal, afterRevision = 0): Promise<void> {
    const reader = body.getReader();
    const bytes = new Uint8Array(ASSISTANT_SNAPSHOT_FRAME_BYTES);
    const decoder = new TextDecoder('utf-8', { fatal: true });
    let size = 0;
    let lineStart = 0;
    let pendingCR = false;
    let revision = afterRevision;
    let identity: string | undefined;
    let terminal: AssistantRunSnapshot | undefined;
    const abort = () => { void reader.cancel().catch(() => undefined); };
    signal?.addEventListener('abort', abort, { once: true });
    function frame(): void {
        let raw: string;
        try { raw = decoder.decode(bytes.subarray(0, size)); } catch { throw new AssistantSubscriptionError('protocol'); }
        const lines = raw.split(/\r\n|\r|\n/);
        const field = (name: string) => lines.filter(line => line.startsWith(`${name}:`)).map(line => line.slice(name.length + 1).replace(/^ /, ''));
        const data = field('data').join('\n');
        if (!data) { return; }
        const events = field('event');
        const ids = field('id');
        if (events.length !== 1 || ids.length !== 1 || !['message.snapshot', 'run.terminal'].includes(events[0]!)) {
            throw new AssistantSubscriptionError('protocol');
        }
        let value: unknown;
        try { value = JSON.parse(data); } catch { throw new AssistantSubscriptionError('protocol'); }
        if (!isAssistantRunSnapshot(value) || value.runId !== runId || ids[0] !== `${runId}:${value.revision}` ||
            terminalStates.includes(value.state) !== (events[0] === 'run.terminal')) {
            throw new AssistantSubscriptionError('protocol');
        }
        const nextIdentity = `${value.conversationId}:${value.userMessageId}:${value.attemptNumber}`;
        if ((identity && identity !== nextIdentity) || terminal) { throw new AssistantSubscriptionError('protocol'); }
        identity = nextIdentity;
        if (value.revision < revision) { return; }
        if (events[0] === 'run.terminal') { terminal = value; }
        if (value.revision <= revision) { return; }
        revision = value.revision;
        callbacks.onSnapshot(value); // Replace the attempt; never append snapshot content.
    }
    function newline(endingBytes: number): void {
        const lineLength = size - lineStart - endingBytes;
        if (lineLength === 0) { frame(); size = 0; lineStart = 0; } else { lineStart = size; }
    }
    try {
        while (true) {
            if (signal?.aborted) { throw new AssistantSubscriptionError('canceled'); }
            const result = await reader.read();
            if (signal?.aborted) { throw new AssistantSubscriptionError('canceled'); }
            if (result.done) { break; }
            for (const byte of result.value) {
                if (pendingCR) {
                    pendingCR = false;
                    if (byte === 10) {
                        if (size >= bytes.length) { throw new AssistantSubscriptionError('protocol'); }
                        bytes[size++] = byte;
                        newline(2);
                        continue;
                    }
                    newline(1);
                }
                if (size >= bytes.length) { throw new AssistantSubscriptionError('protocol'); }
                bytes[size++] = byte;
                if (byte === 13) { pendingCR = true; }
                if (byte === 10) { newline(1); }
            }
        }
        if (pendingCR) { newline(1); }
        if (size !== 0 || !terminal) { throw new AssistantSubscriptionError('disconnected'); }
        callbacks.onTerminal?.(terminal);
    } finally {
        signal?.removeEventListener('abort', abort);
        void reader.cancel().catch(() => undefined);
        reader.releaseLock();
    }
}
