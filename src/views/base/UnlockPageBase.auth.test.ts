import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { createPinia, disposePinia, setActivePinia, type Pinia } from 'pinia';
const state = vi.hoisted(() => ({ token: 'fresh', owner: 'owner', refresh: vi.fn(), revoke: vi.fn(), user: vi.fn(), settings: vi.fn(), language: vi.fn(), notification: vi.fn(), rates: vi.fn() }));
vi.mock('@/lib/services.ts', () => ({ default: { refreshToken: state.refresh, revokeToken: state.revoke } }));
vi.mock('@/lib/userstate.ts', () => ({ getCurrentToken: () => state.token, getCurrentUserInfo: () => ({ username: state.owner }), hasWebAuthnConfig: () => false }));
vi.mock('@/lib/logger.ts', () => ({ default: { error: vi.fn(), warn: vi.fn() } }));
vi.mock('@/stores/setting.ts', () => ({ useSettingsStore: () => ({ appSettings: {}, setApplicationSettingsFromCloudSettings: state.settings, updateLocalizedDefaultSettings: vi.fn() }) }));
vi.mock('@/stores/user.ts', () => ({ useUserStore: () => ({ storeUserBasicInfo: state.user }) }));
vi.mock('@/stores/index.ts', () => ({ useRootStore: () => ({ setNotificationContent: state.notification }) }));
vi.mock('@/stores/transaction.ts', () => ({ useTransactionsStore: () => ({ initTransactionDraft: vi.fn() }) }));
vi.mock('@/stores/exchangeRates.ts', () => ({ useExchangeRatesStore: () => ({ removeExpiredExchangeRates: state.rates, autoUpdateExchangeRatesData: vi.fn() }) }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ setLanguage: state.language, initLocale: vi.fn() }) }));
vi.mock('@/lib/webauthn.ts', () => ({ isWebAuthnSupported: () => false }));
vi.mock('@/lib/cache.ts', () => ({ updateMapCacheExpiration: vi.fn() }));
vi.mock('@/lib/version.ts', () => ({ getClientDisplayVersion: () => 'test' }));
vi.mock('@/lib/ui/common.ts', () => ({ setExpenseAndIncomeAmountColor: vi.fn() }));
import { useUnlockPageBase } from './UnlockPageBase.ts';
import { useTokensStore } from '@/stores/token.ts';
let pinia: Pinia;
beforeEach(() => {
    pinia = createPinia(); setActivePinia(pinia);
    state.token = 'fresh'; state.owner = 'owner';
    state.refresh.mockResolvedValue({ config: { headers: { Authorization: 'Bearer old' } }, data: { success: true,
        result: { newToken: 'fresh', user: { username: 'owner', language: 'old-language' }, applicationCloudSettings: [], notificationContent: 'old-notification' } } });
});
afterEach(() => disposePinia(pinia));
it.each(['owner', 'token'])('does not apply bootstrap UI after %s changes between store and caller continuations', async boundary => {
    state.user.mockImplementationOnce(() => {
        queueMicrotask(() => { if (boundary === 'owner') state.owner = 'other'; else state.token = 'newer'; });
    });
    useUnlockPageBase().doAfterUnlocked();
    await useTokensStore().refreshTokenAndRevokeOldToken();
    expect(state.language).not.toHaveBeenCalled();
    expect(state.notification).not.toHaveBeenCalled();
    expect(state.rates).not.toHaveBeenCalled();
});

it('still applies bootstrap UI for the current session', async () => {
    useUnlockPageBase().doAfterUnlocked();
    const response = await useTokensStore().refreshTokenAndRevokeOldToken();
    expect(response.isCurrent()).toBe(true);
    expect(state.language).toHaveBeenCalledExactlyOnceWith('old-language');
    expect(state.notification).toHaveBeenCalledExactlyOnceWith('old-notification');
    expect(state.rates).toHaveBeenCalledExactlyOnceWith(true);
});
