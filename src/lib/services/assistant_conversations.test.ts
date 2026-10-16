import { beforeEach, describe, expect, it, vi } from 'vitest';
import axios from 'axios';
import { assistantConversationServices as service } from './assistant_conversations.ts';

vi.mock('./transport.ts', () => ({}));
vi.mock('@/lib/userstate.ts', () => ({ getCurrentToken: () => 'secret-token', getCurrentUserInfo: () => ({ username: 'owner' }) }));
vi.mock('@/lib/web.ts', () => ({ getBasePath: () => '/' }));
vi.mock('@/lib/settings.ts', () => ({ getTimeZone: () => 'Asia/Shanghai' }));
vi.mock('@/lib/datetime.ts', () => ({ getBrowserTimezoneName: () => 'UTC', getCurrentUnixTime: () => 1, getTimezoneOffsetMinutes: () => 480 }));

beforeEach(() => { vi.unstubAllGlobals(); });
describe('conversation service transport', () => {
    it('sends an ambiguous command once and preserves its ID, timestamp, epoch and whitespace', async () => {
        const post = vi.spyOn(axios, 'post').mockRejectedValue(new Error('lost response'));
        const request = { clientRequestId: 'stable-id', requestIssuedAt: '2026-10-15T08:00:00Z', expectedDeletionEpoch: 3,
            expectedContextVersion: 1, mode: 'chat', content: '  提问\n' };
        await expect(service.send(request)).rejects.toThrow('lost response');
        expect(post).toHaveBeenCalledExactlyOnceWith('v1/llm/assistant/messages/send.json', request, { signal: undefined, ignoreError: true, expectedOwner: 'owner' });
    });
    it('cancels reads and export, and keeps search text outside the URL', async () => {
        const get = vi.spyOn(axios, 'get').mockResolvedValue({ data: {} });
        const post = vi.spyOn(axios, 'post').mockResolvedValue({ data: {} });
        const signal = new AbortController().signal;
        await service.get({ id: '1' }, signal);
        await service.search({ query: 'private search' }, signal);
        await service.export({ all: true, format: 'json' }, signal);
        expect(get).toHaveBeenCalledWith('v1/llm/assistant/conversations/get.json', { params: { id: '1' }, signal, ignoreError: true, expectedOwner: 'owner' });
        expect(post).toHaveBeenCalledWith('v1/llm/assistant/conversations/search.json', { query: 'private search' }, { signal, ignoreError: true, expectedOwner: 'owner' });
        expect(post).toHaveBeenCalledWith('v1/llm/assistant/conversations/export.json', { all: true, format: 'json' }, { responseType: 'blob', signal, timeout: 0, ignoreError: true, expectedOwner: 'owner' });
    });
    it.each([401, 429, 503])('reports HTTP %i to the owner without retry or a new generation', async status => {
        const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ success: false }), { status, headers: { 'Retry-After': '3' } }));
        vi.stubGlobal('fetch', fetcher);
        await expect(service.subscribe('123', 9, { onSnapshot: vi.fn() })).rejects.toMatchObject({
            kind: status === 401 ? 'unauthorized' : status === 429 ? 'limited' : 'http', status, retryAfterMs: 3000
        });
        expect(fetcher).toHaveBeenCalledTimes(1);
        const [url, options] = fetcher.mock.calls[0]!;
        expect(url).toContain('runId=123&afterRevision=9');
        expect(url).not.toContain('secret-token');
        expect(options.headers.Authorization).toBe('Bearer secret-token');
        expect(options.headers['X-Timezone-Name']).toBe('Asia/Shanghai');
        expect(options.cache).toBe('no-store');
    });
});
