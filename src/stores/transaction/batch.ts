import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type { useTransactionInvalidation } from './invalidation.ts';

interface Context {
    updateStoreInvalidState: ReturnType<typeof useTransactionInvalidation>['updateStoreInvalidState'];
}

// Owns batch for the transaction store.
export function useTransactionBatch(context: Context) {
    const { updateStoreInvalidState } = context;

    function batchUpdateTransactionCategories({ transactionIds, categoryId }: { transactionIds: string[], categoryId: string }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.batchUpdateTransactionCategories({ transactionIds, categoryId }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to update categories for transactions' });
                    return;
                }

                updateStoreInvalidState({
                    transactionList: true,
                    reconciliationStatement: true,
                    overview: true,
                    statistics: true,
                    explorer: true
                });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to update categories for transactions', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to update categories for transactions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function batchUpdateTransactionAccounts({ transactionIds, accountId, isDestinationAccount }: { transactionIds: string[], accountId: string, isDestinationAccount: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.batchUpdateTransactionAccounts({ transactionIds, accountId, isDestinationAccount }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to update accounts for transactions' });
                    return;
                }

                updateStoreInvalidState({
                    transactionList: true,
                    reconciliationStatement: true,
                    accountList: true,
                    overview: true,
                    statistics: true,
                    explorer: true
                });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to update accounts for transactions', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to update accounts for transactions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function batchAddTagsToTransaction({ transactionIds, tagIds }: { transactionIds: string[], tagIds: string[] }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.batchAddTagsToTransaction({ transactionIds, tagIds }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to update tags for transactions' });
                    return;
                }

                updateStoreInvalidState({
                    transactionList: true,
                    reconciliationStatement: true,
                    explorer: true
                });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to update tags for transactions', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to update tags for transactions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function batchRemoveTagsFromTransaction({ transactionIds, tagIds }: { transactionIds: string[], tagIds: string[] }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.batchRemoveTagsFromTransaction({ transactionIds, tagIds }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to update tags for transactions' });
                    return;
                }

                updateStoreInvalidState({
                    transactionList: true,
                    reconciliationStatement: true,
                    explorer: true
                });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to update tags for transactions', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to update tags for transactions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function batchClearAllTagsFromTransaction({ transactionIds }: { transactionIds: string[] }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.batchClearAllTagsFromTransaction({ transactionIds }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to update tags for transactions' });
                    return;
                }

                updateStoreInvalidState({
                    transactionList: true,
                    reconciliationStatement: true,
                    explorer: true
                });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to update tags for transactions', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to update tags for transactions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function moveAllTransactionsBetweenAccounts({ fromAccountId, toAccountId }: { fromAccountId: string, toAccountId: string }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.moveAllTransactionsBetweenAccounts({ fromAccountId, toAccountId }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move transactions' });
                    return;
                }

                updateStoreInvalidState({
                    transactionList: true,
                    reconciliationStatement: true,
                    accountList: true,
                    overview: true,
                    statistics: true,
                    explorer: true
                });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to move transactions', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move transactions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return {
        batchUpdateTransactionCategories, batchUpdateTransactionAccounts, batchAddTagsToTransaction,
        batchRemoveTagsFromTransaction, batchClearAllTagsFromTransaction, moveAllTransactionsBetweenAccounts,
    };
}
