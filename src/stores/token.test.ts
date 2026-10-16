import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { createPinia, disposePinia, setActivePinia, type Pinia } from 'pinia';
const state = vi.hoisted(() => ({ token: 'fresh', owner: 'owner', refresh: vi.fn(), revoke: vi.fn(), user: vi.fn(), settings: vi.fn() }));
vi.mock('@/lib/services.ts', () => ({ default: { refreshToken: state.refresh, revokeToken: state.revoke } }));
vi.mock('@/lib/userstate.ts', () => ({ getCurrentToken: () => state.token, getCurrentUserInfo: () => ({ username: state.owner }) }));
vi.mock('@/lib/logger.ts', () => ({ default: { error: vi.fn() } }));
vi.mock('./setting.ts', () => ({ useSettingsStore: () => ({ setApplicationSettingsFromCloudSettings: state.settings }) }));
vi.mock('./user.ts', () => ({ useUserStore: () => ({ storeUserBasicInfo: state.user }) }));
import { useTokensStore } from './token.ts';
let pinia: Pinia;
const refreshed = () => ({ config: { headers: { Authorization: 'Bearer old' } }, data: { success: true,
    result: { newToken: 'fresh', oldTokenId: 'old-id', user: { username: 'owner' }, applicationCloudSettings: [] } } });
beforeEach(() => {
    pinia = createPinia(); setActivePinia(pinia);
    state.token = 'fresh'; state.owner = 'owner';
    state.refresh.mockResolvedValue(refreshed());
    state.revoke.mockResolvedValue({ data: { success: true, result: true } });
});
afterEach(() => disposePinia(pinia));

it('propagates failed refresh instead of leaving bootstrap pending', async () => {
    state.refresh.mockRejectedValue(new Error('offline'));
    await expect(useTokensStore().refreshTokenAndRevokeOldToken()).rejects.toThrow('offline');
    expect(state.settings).not.toHaveBeenCalled();
});
it.each(['owner', 'token'])('fences the transport-to-store continuation on changed %s', async boundary => {
    state.refresh.mockImplementation(async () => {
        queueMicrotask(() => { if (boundary === 'owner') state.owner = 'other'; else state.token = 'newer-same-owner'; });
        return refreshed();
    });
    await expect(useTokensStore().refreshTokenAndRevokeOldToken()).rejects.toMatchObject({ processed: true });
    expect(state.settings).not.toHaveBeenCalled();
    expect(state.user).not.toHaveBeenCalled();
    expect(state.revoke).not.toHaveBeenCalled();
});
it('applies current settings and handles failed best-effort old-token revocation', async () => {
    state.revoke.mockRejectedValue(new Error('already revoked'));
    await expect(useTokensStore().refreshTokenAndRevokeOldToken()).resolves.toMatchObject({ newToken: 'fresh' });
    expect(state.settings).toHaveBeenCalledOnce();
    expect(state.user).toHaveBeenCalledExactlyOnceWith({ username: 'owner' });
    await vi.waitFor(() => expect(state.revoke).toHaveBeenCalledExactlyOnceWith({ tokenId: 'old-id', ignoreError: true }));
});
it('coalesces concurrent bootstrap callers including old-token cleanup', async () => {
    const store = useTokensStore();
    await Promise.all([store.refreshTokenAndRevokeOldToken(), store.refreshTokenAndRevokeOldToken()]);
    expect(state.refresh).toHaveBeenCalledOnce();
    expect(state.user).toHaveBeenCalledOnce();
    expect(state.revoke).toHaveBeenCalledOnce();
});
