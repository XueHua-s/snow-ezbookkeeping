import type { BeforeResolveFunction } from '@/core/base.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type { TransactionInfoResponse } from '@/models/transaction.ts';
import type { useTransactionInvalidation } from './invalidation.ts';
import type { useTransactionList } from './list.ts';

interface Context {
    removeTransactionFromTransactionList: ReturnType<typeof useTransactionList>['removeTransactionFromTransactionList'];
    updateStoreInvalidState: ReturnType<typeof useTransactionInvalidation>['updateStoreInvalidState'];
}

// Owns delete for the transaction store.
export function useTransactionDelete(context: Context) {
    const { removeTransactionFromTransactionList, updateStoreInvalidState } = context;

    function deleteTransaction({ transaction, defaultCurrency, beforeResolve }: { transaction: TransactionInfoResponse, defaultCurrency: string, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteTransaction({
                id: transaction.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this transaction' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeTransactionFromTransactionList({
                            currentTransaction: transaction,
                            defaultCurrency: defaultCurrency
                        });
                    });
                } else {
                    removeTransactionFromTransactionList({
                        currentTransaction: transaction,
                        defaultCurrency: defaultCurrency
                    });
                }

                updateStoreInvalidState({
                    reconciliationStatement: true,
                    accountList: true,
                    overview: true,
                    statistics: true,
                    explorer: true
                });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete transaction', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this transaction' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function batchDeleteTransactions({ transactionIds, password }: { transactionIds: string[], password: string }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.batchDeleteTransaction({
                ids: transactionIds,
                password: password
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete these transactions' });
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
                logger.error('failed to delete transactions', error);

                updateStoreInvalidState({
                    transactionList: true,
                    reconciliationStatement: true,
                    accountList: true,
                    overview: true,
                    statistics: true,
                    explorer: true
                });

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete these transactions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { deleteTransaction, batchDeleteTransactions };
}
