import { clearCustomIconCache } from '@/lib/cache.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type { useIndexSession } from './session.ts';

interface Context {
    accountsStore: ReturnType<typeof useIndexSession>['accountsStore'];
    overviewStore: ReturnType<typeof useIndexSession>['overviewStore'];
    statisticsStore: ReturnType<typeof useIndexSession>['statisticsStore'];
    transactionCategoriesStore: ReturnType<typeof useIndexSession>['transactionCategoriesStore'];
    transactionTagsStore: ReturnType<typeof useIndexSession>['transactionTagsStore'];
    userCustomIconsStore: ReturnType<typeof useIndexSession>['userCustomIconsStore'];
}

// Owns data management for the index store.
export function useIndexDataManagement(context: Context) {
    const { accountsStore, overviewStore, statisticsStore, transactionCategoriesStore, transactionTagsStore, userCustomIconsStore } = context;

    function clearAllUserTransactionsOfAccount({ accountId, password }: { accountId: string, password: string }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.clearAllTransactionsOfAccount({
                accountId: accountId,
                password: password
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to clear user data' });
                    return;
                }

                if (!accountsStore.accountListStateInvalid) {
                    accountsStore.updateAccountListInvalidState(true);
                }

                if (!overviewStore.transactionOverviewStateInvalid) {
                    overviewStore.updateTransactionOverviewInvalidState(true);
                }

                if (!statisticsStore.transactionStatisticsStateInvalid) {
                    statisticsStore.updateTransactionStatisticsInvalidState(true);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to clear user data', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to clear user data' });
                }
            });
        });
    }

    function clearAllUserTransactions({ password }: { password: string }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.clearAllTransactions({
                password: password
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to clear user data' });
                    return;
                }

                if (!accountsStore.accountListStateInvalid) {
                    accountsStore.updateAccountListInvalidState(true);
                }

                if (!overviewStore.transactionOverviewStateInvalid) {
                    overviewStore.updateTransactionOverviewInvalidState(true);
                }

                if (!statisticsStore.transactionStatisticsStateInvalid) {
                    statisticsStore.updateTransactionStatisticsInvalidState(true);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to clear user data', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to clear user data' });
                }
            });
        });
    }

    function clearAllUserData({ password }: { password: string }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.clearAllData({
                password: password
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to clear user data' });
                    return;
                }

                if (!accountsStore.accountListStateInvalid) {
                    accountsStore.updateAccountListInvalidState(true);
                }

                if (!transactionCategoriesStore.transactionCategoryListStateInvalid) {
                    transactionCategoriesStore.updateTransactionCategoryListInvalidState(true);
                }

                if (!transactionTagsStore.transactionTagListStateInvalid) {
                    transactionTagsStore.updateTransactionTagListInvalidState(true);
                }

                if (!overviewStore.transactionOverviewStateInvalid) {
                    overviewStore.updateTransactionOverviewInvalidState(true);
                }

                if (!statisticsStore.transactionStatisticsStateInvalid) {
                    statisticsStore.updateTransactionStatisticsInvalidState(true);
                }

                userCustomIconsStore.resetCustomIcons();
                clearCustomIconCache();

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to clear user data', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to clear user data' });
                }
            });
        });
    }

    return { clearAllUserTransactionsOfAccount, clearAllUserTransactions, clearAllUserData };
}
