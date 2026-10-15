import { beforeEach, expect, it, vi } from 'vitest';
import { useIndexDataManagement } from '../index/data-management.ts';
import { AssistantDataClearIncompleteError } from './types.ts';

const mocks = vi.hoisted(() => ({ request: vi.fn(), prepare: vi.fn(), synchronize: vi.fn(), reset: vi.fn(), log: vi.fn() }));
vi.mock('@/lib/services.ts', () => ({ default: { clearAllData: mocks.request } }));
vi.mock('@/lib/cache.ts', () => ({ clearCustomIconCache: vi.fn() }));
vi.mock('@/lib/logger.ts', () => ({ default: { error: mocks.log } }));
vi.mock('@/stores/assistant.ts', () => ({ useAssistantStore: () => ({ prepareDataClear: mocks.prepare, synchronize: mocks.synchronize, reset: mocks.reset }) }));
const command = { clientRequestId: 'original', requestIssuedAt: '2026-10-15T00:00:00Z', expectedDeletionEpoch: 1 };
function context() {
    return {
        accountsStore: { accountListStateInvalid: true }, overviewStore: { transactionOverviewStateInvalid: true },
        statisticsStore: { transactionStatisticsStateInvalid: true }, transactionCategoriesStore: { transactionCategoryListStateInvalid: true },
        transactionTagsStore: { transactionTagListStateInvalid: true }, userCustomIconsStore: { resetCustomIcons: vi.fn() }
    } as unknown as Parameters<typeof useIndexDataManagement>[0];
}
beforeEach(() => { mocks.prepare.mockResolvedValue(command); mocks.synchronize.mockResolvedValue(undefined); });
it('does not report clear failure when the successful response is followed by a failed refresh', async () => {
    mocks.request.mockResolvedValue({ data: { success: true, result: true } });
    mocks.synchronize.mockRejectedValue(new Error('refresh offline'));
    const store = useIndexDataManagement(context());
    await expect(store.clearAllUserData({ password: 'private-password' })).resolves.toBe(true);
    expect(mocks.prepare).toHaveBeenCalledWith(false);
    expect(mocks.request).toHaveBeenCalledWith({ password: 'private-password', assistantCommand: command });
});
it('keeps incomplete all-data clear distinct and never logs the password-bearing HTTP error', async () => {
    mocks.request.mockRejectedValue({ response: { status: 409, data: { errorCode: 215051 }, config: { data: 'private-password' } } });
    const store = useIndexDataManagement(context());
    await expect(store.clearAllUserData({ password: 'private-password' })).rejects.toBeInstanceOf(AssistantDataClearIncompleteError);
    expect(mocks.log).toHaveBeenCalledExactlyOnceWith('failed to clear user data');
    expect(mocks.request).toHaveBeenCalledTimes(1);
});
