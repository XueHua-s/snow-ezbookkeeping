import type { AIAssistantChatStreamChunk } from '@/models/large_language_model.ts';
import type { AIAssistantErrorEnvelope, AIAssistantRequestError, AIAssistantChatStreamCallbacks } from './types/assistant.ts';
import { ASSISTANT_PAGEABLE_TOOLS, ASSISTANT_MAX_FRAME_LENGTH } from '@/consts/assistant.ts';

function protocolError(): AIAssistantRequestError {
    return { message: 'Unable to get AI assistant response', incomplete: true };
}

function record(value: unknown): value is Record<string, unknown> {
    return !!value && typeof value === 'object' && !Array.isArray(value);
}

function strings(value: unknown): boolean {
    return Array.isArray(value) && value['every'](item => typeof item === 'string');
}

function optionalStrings(value: Record<string, unknown>, keys: string[]): boolean {
    return keys.every(key => value[key] === undefined || typeof value[key] === 'string');
}

function validContinuation(value: unknown): boolean {
    return record(value) && ASSISTANT_PAGEABLE_TOOLS.includes(value['tool'] as string) && typeof value['cursor'] === 'string' && value['cursor'].length > 0;
}

function validScope(value: unknown): boolean {
    if (!record(value) || !(value['startTime'] === null || typeof value['startTime'] === 'string') ||
        !(value['endTime'] === null || typeof value['endTime'] === 'string') || typeof value['timezone'] !== 'string' ||
        typeof value['includesHidden'] !== 'boolean' || typeof value['excludesDeleted'] !== 'boolean') {
        return false;
    }
    if (!optionalStrings(value, ['currency', 'amountMin', 'amountMax', 'keyword']) ||
        !['types', 'accountIds', 'categoryIds', 'tagIds'].every(key => value[key] === undefined || strings(value[key])) ||
        (value['noTags'] !== undefined && typeof value['noTags'] !== 'boolean') ||
        (value['tagMatch'] !== undefined && !['any', 'all'].includes(value['tagMatch'] as string)) ||
        (value['transferSide'] !== undefined && !['source', 'destination', 'either'].includes(value['transferSide'] as string))) {
        return false;
    }
    return value['periods'] === undefined || (Array.isArray(value['periods']) && value['periods'].every(period =>
        record(period) && typeof period['name'] === 'string' && typeof period['startTime'] === 'string' &&
        typeof period['endTime'] === 'string' && optionalStrings(period, ['matchedCount'])));
}

function validQueries(value: unknown): boolean {
    return Array.isArray(value) && value['every'](query => record(query) && typeof query['queryId'] === 'string' &&
        [...ASSISTANT_PAGEABLE_TOOLS, 'get_transaction'].includes(query['tool'] as string) && validScope(query['scope']) &&
        ['page', 'all_matching'].includes(query['coverage'] as string) && typeof query['queriedAt'] === 'string' && strings(query['warnings']));
}

function validReferences(value: unknown): boolean {
    return Array.isArray(value) && value['every'](reference => record(reference) && typeof reference['id'] === 'string' &&
        /^[1-9][0-9]*$/.test(reference['id']) && typeof reference['time'] === 'number' && Number.isFinite(reference['time']) &&
        typeof reference['type'] === 'number' && Number.isFinite(reference['type']) &&
        typeof reference['sourceAmount'] === 'number' && Number.isFinite(reference['sourceAmount']) &&
        (reference['destinationAmount'] === undefined || (typeof reference['destinationAmount'] === 'number' && Number.isFinite(reference['destinationAmount']))) &&
        (reference['hideAmount'] === undefined || typeof reference['hideAmount'] === 'boolean') &&
        optionalStrings(reference, ['timeText', 'categoryName', 'sourceAccountName', 'destinationAccountName', 'currency', 'destinationCurrency', 'comment']));
}

function validProjection(value: Record<string, unknown>): boolean {
    return (value['queries'] === undefined || validQueries(value['queries'])) &&
        (value['continuation'] === undefined || validContinuation(value['continuation'])) &&
        (value['references'] === undefined || validReferences(value['references']));
}

/** One bounded frame at a time; completion is committed only after a valid EOF. */
export async function readAssistantStream(body: ReadableStream<Uint8Array>, callbacks?: AIAssistantChatStreamCallbacks, signal?: AbortSignal): Promise<void> {
    const reader = body.getReader();
    const decoder = new TextDecoder('utf-8', { fatal: true });
    let buffer = '';
    let frameLines: string[] = [];
    let frameLength = 0;
    let terminalSeen = false;
    let hasReply = false;
    let terminal: AIAssistantChatStreamChunk | AIAssistantErrorEnvelope | undefined;
    const abort = () => { void reader.cancel().catch(() => undefined); };
    signal?.addEventListener('abort', abort, { once: true });

    function checkCanceled(): void {
        if (signal?.aborted) {
            throw { canceled: true };
        }
    }

    function frame(raw: string): void {
        checkCanceled();
        const data = raw.split(/\r\n|\r|\n/).filter(line => line.startsWith('data:'))
            .map(line => line.slice(5).replace(/^ /, '')).join('\n');
        if (!data) {
            return;
        }
        let value: unknown;
        try {
            value = JSON.parse(data);
        } catch {
            throw protocolError();
        }
        if (!value || typeof value !== 'object' || Array.isArray(value)) {
            throw protocolError();
        }
        const chunk = value as AIAssistantChatStreamChunk;
        const envelope = value as AIAssistantErrorEnvelope;
        if (envelope.success === false || chunk.type === 'done') {
            if (envelope.success === false) {
                if (!optionalStrings(value as Record<string, unknown>, ['errorMessage', 'path']) ||
                    (envelope.errorCode !== undefined && (typeof envelope.errorCode !== 'number' || !Number.isInteger(envelope.errorCode)))) {
                    throw protocolError();
                }
                const context = (value as Record<string, unknown>)['context'];
                if (context !== undefined && (!record(context) || (context['assistant'] !== undefined &&
                    (!record(context['assistant']) || !['failed', 'incomplete'].includes(context['assistant']['state'] as string) ||
                    !validQueries(context['assistant']['queries']) || !validProjection(context['assistant']))))) {
                    throw protocolError();
                }
            } else if (!['chat', 'summary'].includes(chunk.mode as string) ||
                !optionalStrings(value as Record<string, unknown>, ['reply', 'thinking']) ||
                !(chunk.reply === undefined ? hasReply : chunk.reply) || !validProjection(value as Record<string, unknown>)) {
                throw protocolError();
            }
            if (terminalSeen) {
                throw protocolError();
            }
            terminalSeen = true;
            terminal = envelope.success === false ? envelope : chunk;
            return;
        }
        if (!['status', 'thinking_delta', 'reply_delta', 'references'].includes(chunk.type)) {
            return; // Future nonterminal events do not change completion semantics.
        }
        if (terminalSeen) {
            throw protocolError();
        }
        if (chunk.type === 'status') {
            if (!chunk.status || !['querying', 'aggregating', 'answering'].includes(chunk.status)) {
                throw protocolError();
            }
            callbacks?.onStatus?.(chunk.status);
        } else if (chunk.type === 'references') {
            if (!validReferences(chunk.references)) {
                throw protocolError();
            }
            callbacks?.onReferences?.(chunk.references);
        } else {
            if (typeof chunk.delta !== 'string') {
                throw protocolError();
            }
            if (chunk.type === 'thinking_delta') {
                callbacks?.onThinkingDelta?.(chunk.delta);
            } else {
                hasReply ||= !!chunk.delta;
                callbacks?.onReplyDelta?.(chunk.delta);
            }
        }
    }

    function line(value: string): void {
        if (!value) {
            frame(frameLines.join('\n'));
            frameLines = [];
            frameLength = 0;
        } else {
            frameLines.push(value);
            frameLength += value['length'] + 1;
            if (frameLength > ASSISTANT_MAX_FRAME_LENGTH) {
                throw protocolError();
            }
        }
    }

    function drain(final: boolean): void {
        let match: RegExpExecArray | null;
        while ((match = /[\r\n]/.exec(buffer))) {
            const index = match.index;
            // A trailing CR may be the first half of CRLF in the next read.
            if (!final && buffer[index] === '\r' && index === buffer.length - 1) {
                break;
            }
            const endingLength = buffer[index] === '\r' && buffer[index + 1] === '\n' ? 2 : 1;
            line(buffer.slice(0, index));
            buffer = buffer.slice(index + endingLength);
        }
        if (frameLength + buffer.length > ASSISTANT_MAX_FRAME_LENGTH) {
            throw protocolError();
        }
        if (final) {
            if (buffer) {
                line(buffer);
            }
            if (frameLines.length) {
                frame(frameLines.join('\n'));
            }
        }
    }

    try {
        checkCanceled();
        while (true) {
            const { done, value } = await reader.read();
            checkCanceled();
            try {
                buffer += done ? decoder.decode() : decoder.decode(value, { stream: true });
            } catch {
                throw protocolError();
            }
            drain(done);
            if (done) {
                break;
            }
        }
        if (!terminalSeen || !terminal) {
            throw protocolError();
        }
        if ('success' in terminal) {
            throw { response: { data: terminal } };
        }
        callbacks?.onDone?.(terminal);
    } finally {
        signal?.removeEventListener('abort', abort);
        // Also stop a transport that is still sending after a malformed frame.
        void reader.cancel().catch(() => undefined);
        reader.releaseLock();
    }
}

/** Browser timers overflow above signed int32; the server accepts uint32 seconds. */
export function startAssistantDeadline(controller: AbortController, timeout: number): () => void {
    const deadline = Date.now() + timeout;
    let timer: ReturnType<typeof setTimeout>;
    function schedule(): void {
        const remaining = deadline - Date.now();
        if (remaining <= 0) {
            controller.abort(new DOMException('Assistant request timed out', 'TimeoutError'));
            return;
        }
        timer = setTimeout(schedule, Math.min(remaining, 0x7fffffff));
    }
    schedule();
    return () => clearTimeout(timer);
}
