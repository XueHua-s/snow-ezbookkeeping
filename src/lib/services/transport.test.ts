import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import axios, { AxiosError, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';
const identity = vi.hoisted(() => ({ token: 'old', username: 'owner', clear: vi.fn(), update: vi.fn() }));
vi.mock('@/lib/userstate.ts', () => ({
    getCurrentToken: () => identity.token,
    getCurrentUserInfo: () => ({ username: identity.username }),
    updateCurrentToken: (token: string) => { identity.update(token); identity.token = token; },
    clearCurrentTokenAndUserInfo: identity.clear
}));
vi.mock('@/lib/web.ts', () => ({ getBasePath: () => '/' }));
vi.mock('@/lib/settings.ts', () => ({ getTimeZone: () => 'UTC' }));
vi.mock('@/lib/logger.ts', () => ({ default: { debug: vi.fn() } }));
vi.mock('@/lib/datetime.ts', () => ({ getBrowserTimezoneName: () => 'UTC', getCurrentUnixTime: () => 1, getTimezoneOffsetMinutes: () => 0 }));
let service: typeof import('./transport.ts')['transportServices'];
const originalAdapter = axios.defaults.adapter;
function response(config: InternalAxiosRequestConfig, newToken?: string): AxiosResponse {
    return { config, data: { success: true, result: { newToken, user: { username: identity.username } } }, status: 200, statusText: 'OK', headers: {} };
}
function deferred<T>() {
    let resolve!: (value: T) => void;
    let reject!: (reason: unknown) => void;
    const promise = new Promise<T>((a, b) => { resolve = a; reject = b; });
    return { promise, resolve, reject };
}
beforeEach(async () => {
    vi.resetModules();
    axios.interceptors.request.clear(); axios.interceptors.response.clear();
    identity.token = 'old'; identity.username = 'owner';
    vi.stubGlobal('location', { reload: vi.fn() });
    service = (await import('./transport.ts')).transportServices;
});
afterEach(() => { axios.defaults.adapter = originalAdapter; vi.unstubAllGlobals(); });

it('rejects the refresh and every blocked request on failure, then permits new requests', async () => {
    const pending = deferred<AxiosResponse>();
    const adapter = vi.fn().mockImplementation((config: InternalAxiosRequestConfig) => config.url?.includes('refresh') ? pending.promise : Promise.resolve(response(config)));
    axios.defaults.adapter = adapter;
    const failed: unknown[] = [];
    const refresh = service.refreshToken().catch(error => { failed.push(error); });
    const queued = axios.get('queued').catch(error => { failed.push(error); });
    await vi.waitFor(() => expect(adapter).toHaveBeenCalledTimes(1));
    pending.reject(new Error('offline'));
    await vi.waitFor(() => expect(failed).toHaveLength(2));
    await Promise.all([refresh, queued]);
    await axios.get('after-failure');
    expect(adapter).toHaveBeenCalledTimes(2);
});

it('coalesces same-tab refresh and releases blocked requests with the committed token', async () => {
    const pending = deferred<AxiosResponse>();
    const adapter = vi.fn().mockImplementation((config: InternalAxiosRequestConfig) => config.url?.includes('refresh') ? pending.promise : Promise.resolve(response(config)));
    axios.defaults.adapter = adapter;
    const first = service.refreshToken(); const second = service.refreshToken();
    const queued = axios.get('queued');
    await vi.waitFor(() => expect(adapter).toHaveBeenCalledTimes(1));
    pending.resolve(response(adapter.mock.calls[0]![0], 'fresh'));
    await Promise.all([first, second, queued]);
    expect(identity.update).toHaveBeenCalledExactlyOnceWith('fresh');
    expect(adapter.mock.calls[1]![0].headers.Authorization).toBe('Bearer fresh');
});

it('retries a stale unauthorized refresh against the current same-owner token without logging out', async () => {
    const adapter = vi.fn().mockImplementation(async (config: InternalAxiosRequestConfig) => {
        if (adapter.mock.calls.length === 1) {
            identity.token = 'other-tab-fresh';
            throw new AxiosError('revoked old token', 'ERR_BAD_REQUEST', config, undefined,
                { ...response(config), status: 401, data: { errorCode: 202002 } });
        }
        return response(config);
    });
    axios.defaults.adapter = adapter;
    await service.refreshToken();
    expect(adapter).toHaveBeenCalledTimes(2);
    expect(adapter.mock.calls[1]![0].headers.Authorization).toBe('Bearer other-tab-fresh');
    expect(identity.clear).not.toHaveBeenCalled();
    expect(location.reload).not.toHaveBeenCalled();
});

it('keeps current-token unauthorized handling and rejects instead of leaving the gate pending', async () => {
    axios.defaults.adapter = async config => { throw new AxiosError('unauthorized', 'ERR_BAD_REQUEST', config, undefined,
        { ...response(config), status: 401, data: { errorCode: 202002 } }); };
    await expect(service.refreshToken()).rejects.toMatchObject({ processed: true });
    expect(identity.clear).toHaveBeenCalledOnce();
    expect(location.reload).toHaveBeenCalledOnce();
});

it.each(['success', 'failure'])('does not overwrite a new owner or replay old queued writes after late %s', async outcome => {
    const pending = deferred<AxiosResponse>();
    const adapter = vi.fn().mockImplementation(() => pending.promise);
    axios.defaults.adapter = adapter;
    const refresh = service.refreshToken();
    const queued = axios.post('old-owner-write', { content: 'old-owner' });
    const rejection = Promise.all([expect(refresh).rejects.toMatchObject({ processed: true }), expect(queued).rejects.toMatchObject({ processed: true })]);
    await vi.waitFor(() => expect(adapter).toHaveBeenCalledTimes(1));
    identity.token = 'new-owner-token'; identity.username = 'new-owner';
    if (outcome === 'success') pending.resolve(response(adapter.mock.calls[0]![0], 'stale-token'));
    else pending.reject(new Error('late-network-error'));
    await rejection;
    expect(identity.token).toBe('new-owner-token');
    expect(identity.update).not.toHaveBeenCalled();
    expect(adapter).toHaveBeenCalledTimes(1);
});

it('discards a late same-owner success and retries without overwriting the other tab token', async () => {
    const adapter = vi.fn().mockImplementation(async (config: InternalAxiosRequestConfig) => {
        if (adapter.mock.calls.length === 1) {
            identity.token = 'other-tab-fresh';
            return response(config, 'obsolete-result');
        }
        return response(config);
    });
    axios.defaults.adapter = adapter;
    await service.refreshToken();
    expect(identity.token).toBe('other-tab-fresh');
    expect(identity.update).not.toHaveBeenCalled();
    expect(adapter).toHaveBeenCalledTimes(2);
});

it('bounds refresh retries when another tab keeps changing the token', async () => {
    const adapter = vi.fn().mockImplementation(async (config: InternalAxiosRequestConfig) => {
        identity.token = `other-tab-${adapter.mock.calls.length}`;
        return response(config, 'obsolete-result');
    });
    axios.defaults.adapter = adapter;
    await expect(service.refreshToken()).rejects.toMatchObject({ processed: true });
    expect(adapter).toHaveBeenCalledTimes(3);
    expect(identity.update).not.toHaveBeenCalled();
    axios.defaults.adapter = async config => response(config);
    await axios.get('still-available');
});

it('does not retry an ordinary accounting POST whose old token was revoked', async () => {
    const adapter = vi.fn().mockImplementation(async (config: InternalAxiosRequestConfig) => {
        identity.token = 'other-tab-fresh';
        throw new AxiosError('revoked old token', 'ERR_BAD_REQUEST', config, undefined,
            { ...response(config), status: 401, data: { errorCode: 202002 } });
    });
    axios.defaults.adapter = adapter;
    await expect(axios.post('transactions/add', { amount: 10 })).rejects.toMatchObject({ response: { status: 401 } });
    expect(adapter).toHaveBeenCalledOnce();
    expect(identity.clear).not.toHaveBeenCalled();
});

it('does not dispatch refresh under an owner changed before Axios interceptors run', async () => {
    const adapter = vi.fn(); axios.defaults.adapter = adapter;
    queueMicrotask(() => { identity.token = 'new-owner'; identity.username = 'other'; });
    await expect(service.refreshToken()).rejects.toMatchObject({ processed: true });
    expect(adapter).not.toHaveBeenCalled();
});

it('binds revocation to its initiating token across the Axios scheduling boundary', async () => {
    const { authorizationServices } = await import('./authorization.ts');
    const adapter = vi.fn(); axios.defaults.adapter = adapter;
    queueMicrotask(() => { identity.token = 'new-owner'; identity.username = 'other'; });
    await expect(authorizationServices.revokeToken({ tokenId: 'old-token-id', ignoreError: true })).rejects.toMatchObject({ processed: true });
    expect(adapter).not.toHaveBeenCalled();
});

it('rejects a late successful refresh after logout without restoring credentials', async () => {
    const pending = deferred<AxiosResponse>();
    const adapter = vi.fn().mockImplementation(() => pending.promise); axios.defaults.adapter = adapter;
    const refresh = service.refreshToken();
    const rejected = expect(refresh).rejects.toMatchObject({ processed: true });
    await vi.waitFor(() => expect(adapter).toHaveBeenCalledOnce());
    identity.token = '';
    pending.resolve(response(adapter.mock.calls[0]![0], 'obsolete'));
    await rejected;
    expect(identity.update).not.toHaveBeenCalled();
});


it.each(['read', 'clear', 'export'])('does not dispatch assistant %s after owner changes before Axios interceptors', async operation => {
    const { assistantConversationServices: assistant } = await import('./assistant_conversations.ts');
    const adapter = vi.fn(async (config: InternalAxiosRequestConfig) => response(config));
    axios.defaults.adapter = adapter;
    queueMicrotask(() => { identity.token = 'other-token'; identity.username = 'other'; });
    const request = operation === 'read' ? assistant.memories()
        : operation === 'clear' ? assistant.clear({} as Parameters<typeof assistant.clear>[0])
            : assistant.export({} as Parameters<typeof assistant.export>[0]);
    await expect(request).rejects.toMatchObject({ processed: true });
    expect(adapter).not.toHaveBeenCalled();
});

it('releases same-owner assistant reads through normal refresh with the new token', async () => {
    const { assistantConversationServices: assistant } = await import('./assistant_conversations.ts');
    const pending = deferred<AxiosResponse>();
    const adapter = vi.fn().mockImplementation((config: InternalAxiosRequestConfig) => config.url?.includes('refresh') ? pending.promise : Promise.resolve(response(config)));
    axios.defaults.adapter = adapter;
    const refresh = service.refreshToken();
    const read = assistant.memories();
    await vi.waitFor(() => expect(adapter).toHaveBeenCalledOnce());
    pending.resolve(response(adapter.mock.calls[0]![0], 'fresh'));
    await Promise.all([refresh, read]);
    expect(adapter.mock.calls[1]![0].headers.Authorization).toBe('Bearer fresh');
});
