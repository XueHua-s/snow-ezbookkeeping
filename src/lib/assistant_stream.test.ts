import { afterEach, describe, expect, it, vi } from 'vitest';
import { readAssistantStream, startAssistantDeadline } from './assistant_stream.ts';

function stream(text: string, oneByte = false): ReadableStream<Uint8Array> {
    const bytes = new TextEncoder().encode(text);
    return new ReadableStream({ start(controller) {
        if (oneByte) {
            for (const byte of bytes) {
                controller.enqueue(new Uint8Array([byte]));
            }
        } else {
            controller.enqueue(bytes);
        }
        controller.close();
    } });
}

const done = 'data: {"type":"done","mode":"chat","reply":"完成"}';

describe('assistant SSE protocol', () => {
    it('incrementally decodes UTF-8 and split CRLF, multiline data, heartbeats, unknown events and a final frame without blank line', async () => {
        const onReplyDelta = vi.fn();
        const onStatus = vi.fn();
        const onDone = vi.fn();
        await readAssistantStream(stream(': heartbeat\r\n\r\ndata: {"type":"future_event"}\r\n\r\ndata: {"type":"status","status":"querying"}\r\n\r\ndata: {"type":"reply_delta",\r\ndata: "delta":"你好😀"}\r\n\r\n' + done, true), { onReplyDelta, onStatus, onDone });
        expect(onReplyDelta).toHaveBeenCalledExactlyOnceWith('你好😀');
        expect(onStatus).toHaveBeenCalledExactlyOnceWith('querying');
        expect(onDone).toHaveBeenCalledExactlyOnceWith({ type: 'done', mode: 'chat', reply: '完成' });
    });

    it.each(['\n', '\r\n', '\r', '\r\n\n', '\n\r\n'])('accepts SSE line endings split at every byte: %j', async ending => {
        const onDone = vi.fn();
        const separator = ending === '\r\n\n' || ending === '\n\r\n' ? ending : ending + ending;
        await readAssistantStream(stream('data: {"type":"reply_delta","delta":"a"}' + separator + done + ending, true), { onDone });
        expect(onDone).toHaveBeenCalledOnce();
    });

    it.each([
        '', ': heartbeat\n\n', 'data: {"type":"reply_delta","delta":"partial"}\n\n',
        'data: {broken}\n\n', 'data: {"type":"done"', 'data: [DONE]\n\n',
        done + '\n\n' + done, done + '\n\ndata: {broken}',
        done + '\n\ndata: {"success":false,"errorMessage":"error"}',
        'data: {"success":false}\n\n' + done,
        done + '\n\ndata: {"type":"reply_delta","delta":"late"}',
        'data: {"type":"reply_delta","delta":42}\n\n' + done
    ])('rejects missing/corrupt/duplicate terminal without a completion callback: %s', async text => {
        const onDone = vi.fn();
        await expect(readAssistantStream(stream(text), { onDone })).rejects.toMatchObject({ incomplete: true });
        expect(onDone).not.toHaveBeenCalled();
    });

    it.each([
        { type: 'done' }, { type: 'done', mode: 'chat' }, { type: 'done', mode: 'other', reply: 'a' },
        { type: 'done', mode: 'chat', reply: 123 }, { type: 'done', mode: 'chat', reply: 'a', thinking: false },
        { type: 'done', mode: 'chat', reply: 'a', queries: false },
        { type: 'done', mode: 'chat', reply: 'a', queries: [{ queryId: 'q1' }] },
        { type: 'done', mode: 'chat', reply: 'a', continuation: { tool: 'query_transactions', cursor: 123 } },
        { type: 'done', mode: 'chat', reply: 'a', continuation: { tool: 'get_transaction', cursor: 'opaque' } },
        { type: 'done', mode: 'chat', reply: 'a', references: {} },
        { type: 'done', mode: 'chat', reply: 'a', references: [{ id: '1', time: 1, type: 1, sourceAmount: '1' }] }
    ])('rejects structurally corrupt terminal projection %j', async value => {
        const onDone = vi.fn();
        await expect(readAssistantStream(stream('data: ' + JSON.stringify(value)), { onDone })).rejects.toMatchObject({ incomplete: true });
        expect(onDone).not.toHaveBeenCalled();
    });

    it('delivers valid reference and query projections intact', async () => {
        const references = [{ id: '123', time: 1, type: 2, sourceAmount: 1200, currency: 'CNY', hideAmount: false }];
        const queries = [{ queryId: 'q1', tool: 'aggregate_transactions', coverage: 'all_matching', queriedAt: '2026-10-14T00:00:00Z', warnings: [], scope: {
            startTime: null, endTime: null, timezone: 'UTC', includesHidden: true, excludesDeleted: true,
            accountIds: ['123'], periods: [{ name: 'current', startTime: '2026-10-01T00:00:00Z', endTime: '2026-10-14T00:00:00Z', matchedCount: '1' }]
        } }];
        const terminal = { type: 'done', mode: 'summary', reply: 'answer', references, queries, continuation: { tool: 'aggregate_transactions', cursor: 'opaque' } };
        const onReferences = vi.fn();
        const onDone = vi.fn();
        await readAssistantStream(stream('data: ' + JSON.stringify({ type: 'references', references }) + '\n\ndata: ' + JSON.stringify(terminal)), { onReferences, onDone });
        expect(onReferences).toHaveBeenCalledExactlyOnceWith(references);
        expect(onDone).toHaveBeenCalledExactlyOnceWith(terminal);
    });

    it('accepts omitted terminal reply only when reply text already arrived', async () => {
        const onDone = vi.fn();
        await readAssistantStream(stream('data: {"type":"reply_delta","delta":"a"}\n\ndata: {"type":"done","mode":"chat"}'), { onDone });
        expect(onDone).toHaveBeenCalledExactlyOnceWith({ type: 'done', mode: 'chat' });
    });

    it('rejects a truncated multibyte codepoint even following a valid terminal', async () => {
        const onDone = vi.fn();
        const bytes = new TextEncoder().encode(done + '\n\n');
        const body = new ReadableStream<Uint8Array>({ start(controller) {
            controller.enqueue(bytes);
            controller.enqueue(new Uint8Array([0xf0, 0x9f]));
            controller.close();
        } });
        await expect(readAssistantStream(body, { onDone })).rejects.toMatchObject({ incomplete: true });
        expect(onDone).not.toHaveBeenCalled();
    });

    it('preserves the entire server error envelope', async () => {
        const envelope = { success: false, errorCode: 1, errorMessage: 'limited', path: '/assistant', context: {
            assistant: { state: 'incomplete', queries: [{ queryId: 'q1', tool: 'query_transactions', coverage: 'page', queriedAt: '2026-10-14T00:00:00Z', warnings: [], scope: { startTime: null, endTime: null, timezone: 'UTC', includesHidden: true, excludesDeleted: true } }], continuation: { tool: 'query_transactions', cursor: 'opaque' } }, extra: 'retained'
        } };
        await expect(readAssistantStream(stream('data: ' + JSON.stringify(envelope)))).rejects.toEqual({ response: { data: envelope } });
    });

    it('cancels the reader and ignores queued events after abort', async () => {
        const controller = new AbortController();
        const cancel = vi.fn();
        const onDone = vi.fn();
        const body = new ReadableStream<Uint8Array>({ cancel });
        const pending = readAssistantStream(body, { onDone }, controller.signal);
        controller.abort();
        await expect(pending).rejects.toEqual({ canceled: true });
        expect(cancel).toHaveBeenCalledOnce();
        expect(onDone).not.toHaveBeenCalled();
    });

    it('bounds unterminated frames and stops the reader on malformed data', async () => {
        const cancel = vi.fn();
        const body = new ReadableStream<Uint8Array>({ start(controller) {
            controller.enqueue(new TextEncoder().encode('data: ' + 'x'.repeat(1024 * 1024)));
        }, cancel });
        await expect(readAssistantStream(body)).rejects.toMatchObject({ incomplete: true });
        expect(cancel).toHaveBeenCalledOnce();
    });
});

describe('assistant overall deadline', () => {
    afterEach(() => vi.useRealTimers());
    it('chunks uint32-second runtimes without signed timer overflow and cleans up', () => {
        vi.useFakeTimers();
        const controller = new AbortController();
        const duration = 0xffffffff * 1000 + 10000;
        const stop = startAssistantDeadline(controller, duration);
        vi.advanceTimersByTime(0x7fffffff);
        expect(controller.signal.aborted).toBe(false);
        expect(vi.getTimerCount()).toBe(1);
        stop();
        expect(vi.getTimerCount()).toBe(0);
    });
    it('aborts at the full deadline after crossing a timer chunk', () => {
        vi.useFakeTimers();
        const controller = new AbortController();
        startAssistantDeadline(controller, 0x7fffffff + 10000);
        vi.advanceTimersByTime(0x7fffffff + 9999);
        expect(controller.signal.aborted).toBe(false);
        vi.advanceTimersByTime(1);
        expect(controller.signal.aborted).toBe(true);
    });
});
