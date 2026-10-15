import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import axios, { type InternalAxiosRequestConfig } from 'axios';

vi.mock('@/core/transaction.ts', () => ({ TransactionType: {} }));
vi.mock('@/models/transaction.ts', () => ({ TransactionAmountsRequest: {} }));
vi.mock('./userstate.ts', () => ({ getCurrentToken: () => 'test-token', clearCurrentTokenAndUserInfo: vi.fn() }));
vi.mock('./settings.ts', () => ({ getTimeZone: () => 'Asia/Shanghai' }));
vi.mock('./server_settings.ts', () => ({ getAIAssistantRunTimeout: vi.fn(), getGoogleMapAPIKey: vi.fn(), getBaiduMapAK: vi.fn(), getAmapApplicationKey: vi.fn(), getExchangeRatesRequestTimeout: vi.fn() }));
vi.mock('./datetime.ts', () => ({ getTimezoneOffsetMinutes: () => 480, getBrowserTimezoneName: () => 'UTC', getCurrentUnixTime: () => 1 }));
vi.mock('./misc.ts', () => ({ generateRandomUUID: () => 'uuid' }));
vi.mock('./web.ts', () => ({ getBasePath: () => '/book' }));
vi.mock('./logger.ts', () => ({ default: { debug: vi.fn() } }));
import { getAIAssistantRunTimeout } from './server_settings.ts';
import services from './services.ts';

const req = { mode: 'chat', message: 'next page', continuation: { tool: 'query_transactions', cursor: 'opaque' } } as const;
const envelope = { success: false, errorCode: 42, errorMessage: 'budget', path: '/assistant', context: { assistant: { state: 'incomplete', queries: [{ queryId: 'q1', tool: 'query_transactions', coverage: 'page', queriedAt: '2026-10-14T00:00:00Z', warnings: [], scope: { startTime: null, endTime: null, timezone: 'UTC', includesHidden: true, excludesDeleted: true } }], continuation: req.continuation } } };
const fetchMock = vi.fn<typeof fetch>();
const originalAdapter = axios.defaults.adapter;

beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock);
    vi.mocked(getAIAssistantRunTimeout).mockReturnValue(180000);
});
afterEach(() => {
    axios.defaults.adapter = originalAdapter;
    vi.unstubAllGlobals();
    vi.useRealTimers();
});

function pendingJSON(): Promise<InternalAxiosRequestConfig> {
    return new Promise(resolve => {
        axios.defaults.adapter = config => new Promise((_resolve, reject) => {
            resolve(config);
            config.signal?.addEventListener?.('abort', () => reject(new axios.CanceledError('aborted', config)));
        });
    });
}

function pendingSSE(): void {
    fetchMock.mockImplementation((_url, init) => new Promise((_resolve, reject) => {
        init?.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')));
    }));
}

describe('assistant service transport', () => {
    it('sends authenticated body-only continuation and consumes the actual SSE parser', async () => {
        fetchMock.mockResolvedValue(new Response('data: {"type":"status","status":"answering"}\n\ndata: {"type":"done","mode":"chat","reply":"answer"}'));
        services.setLocale('zh-CN');
        const onDone = vi.fn();
        await services.chatWithAIAssistantStream({ req, callbacks: { onDone } });
        const [url, init] = fetchMock.mock.calls[0]!;
        expect(String(url)).not.toContain('opaque');
        expect(init?.body).toBe(JSON.stringify(req));
        expect(init?.headers).toMatchObject({ Authorization: 'Bearer test-token', 'X-Timezone-Name': 'Asia/Shanghai', 'Accept-Language': 'zh-CN' });
        expect(onDone).toHaveBeenCalledExactlyOnceWith({ type: 'done', mode: 'chat', reply: 'answer' });
    });

    it.each(['http', 'sse'] as const)('preserves the %s error envelope and its successful evidence', async transport => {
        fetchMock.mockResolvedValue(new Response(transport === 'http' ? JSON.stringify(envelope) : 'data: ' + JSON.stringify(envelope), { status: transport === 'http' ? 503 : 200 }));
        await expect(services.chatWithAIAssistantStream({ req })).rejects.toEqual({ response: { data: envelope } });
    });

    it('does not turn an ordinary SSE EOF into success', async () => {
        fetchMock.mockResolvedValue(new Response('data: {"type":"reply_delta","delta":"partial"}\n\n'));
        await expect(services.chatWithAIAssistantStream({ req })).rejects.toMatchObject({ incomplete: true });
    });

    it('uses runtime plus ten seconds for SSE and releases its timer', async () => {
        vi.useFakeTimers();
        pendingSSE();
        const result = services.chatWithAIAssistantStream({ req, cancelableUuid: 'sse-timeout' });
        const assertion = expect(result).rejects.toMatchObject({ timedOut: true, incomplete: true });
        await vi.advanceTimersByTimeAsync(189999);
        expect(fetchMock.mock.calls[0]![1]!.signal!.aborted).toBe(false);
        await vi.advanceTimersByTimeAsync(1);
        await assertion;
        expect(vi.getTimerCount()).toBe(0);
    });

    it('enforces the overall SSE deadline while an established response stalls, retaining partial text without done', async () => {
        vi.useFakeTimers();
        const cancel = vi.fn();
        const onReplyDelta = vi.fn();
        const onDone = vi.fn();
        const body = new ReadableStream<Uint8Array>({ start(controller) {
            controller.enqueue(new TextEncoder().encode('data: {"type":"reply_delta","delta":"partial"}\n\n'));
        }, cancel });
        fetchMock.mockResolvedValue(new Response(body));
        const pending = services.chatWithAIAssistantStream({ req, callbacks: { onReplyDelta, onDone } });
        const assertion = expect(pending).rejects.toMatchObject({ timedOut: true });
        await vi.advanceTimersByTimeAsync(190000);
        await assertion;
        expect(onReplyDelta).toHaveBeenCalledExactlyOnceWith('partial');
        expect(onDone).not.toHaveBeenCalled();
        expect(cancel).toHaveBeenCalledOnce();
        expect(vi.getTimerCount()).toBe(0);
    });

    it('uses the same deadline for JSON and preserves authentication', async () => {
        vi.useFakeTimers();
        const ready = pendingJSON();
        const result = services.chatWithAIAssistant({ req, cancelableUuid: 'json-timeout' });
        const assertion = expect(result).rejects.toMatchObject({ timedOut: true, incomplete: true });
        const config = await ready;
        expect(config.timeout).toBe(0);
        expect(config.headers.Authorization).toBe('Bearer test-token');
        await vi.advanceTimersByTimeAsync(189999);
        expect(config.signal!.aborted).toBe(false);
        await vi.advanceTimersByTimeAsync(1);
        await assertion;
        expect(vi.getTimerCount()).toBe(0);
    });

    it.each(['json', 'sse'] as const)('actively aborts %s and differentiates user cancellation from deadline', async transport => {
        vi.useFakeTimers();
        const ready = transport === 'json' ? pendingJSON() : undefined;
        if (transport === 'sse') pendingSSE();
        const result = transport === 'json'
            ? services.chatWithAIAssistant({ req, cancelableUuid: 'cancel' })
            : services.chatWithAIAssistantStream({ req, cancelableUuid: 'cancel' });
        const assertion = expect(result).rejects.toEqual({ canceled: true });
        if (ready) await ready;
        services.cancelRequest('cancel');
        await assertion;
        expect(vi.getTimerCount()).toBe(0);
    });

    it('keeps JSON error context intact through axios interceptors', async () => {
        axios.defaults.adapter = config => Promise.reject({ response: { config, data: envelope, status: 503 } });
        await expect(services.chatWithAIAssistant({ req })).rejects.toMatchObject({ response: { data: envelope } });
    });

    it.each(['json', 'sse'] as const)('does not overflow %s deadlines for uint32-second runtimes', async transport => {
        vi.useFakeTimers();
        vi.mocked(getAIAssistantRunTimeout).mockReturnValue(0xffffffff * 1000);
        const ready = transport === 'json' ? pendingJSON() : undefined;
        if (transport === 'sse') pendingSSE();
        const result = transport === 'json'
            ? services.chatWithAIAssistant({ req, cancelableUuid: 'long' })
            : services.chatWithAIAssistantStream({ req, cancelableUuid: 'long' });
        const assertion = expect(result).rejects.toEqual({ canceled: true });
        const config = ready ? await ready : undefined;
        await vi.advanceTimersByTimeAsync(0x7fffffff);
        const signal = config?.signal ?? fetchMock.mock.calls[0]![1]!.signal;
        expect(signal!.aborted).toBe(false);
        services.cancelRequest('long');
        await assertion;
    });
});
