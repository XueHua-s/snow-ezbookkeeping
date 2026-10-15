import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import {
    type TransactionPageWrapper,
    EMPTY_TRANSACTION_RESULT,
    Transaction
} from '@/models/transaction.ts';
import type { useTransactionInvalidation } from './invalidation.ts';
import type { useTransactionList } from './list.ts';
import type { useTransactionState } from './state.ts';

interface Context {
    transactionsNextTimeId: ReturnType<typeof useTransactionState>['transactionsNextTimeId'];
    transactionsFilter: ReturnType<typeof useTransactionState>['transactionsFilter'];
    loadTransactionList: ReturnType<typeof useTransactionList>['loadTransactionList'];
    transactionListStateInvalid: ReturnType<typeof useTransactionState>['transactionListStateInvalid'];
    updateTransactionListInvalidState: ReturnType<typeof useTransactionInvalidation>['updateTransactionListInvalidState'];
}

// Owns loading for the transaction store.
export function useTransactionLoading(context: Context) {
    const { transactionsNextTimeId, transactionsFilter, loadTransactionList, transactionListStateInvalid, updateTransactionListInvalidState } = context;

    function loadTransactions({ reload, count, page, mustHavePictures, withCount, withPictures, autoExpand, defaultCurrency }: { reload?: boolean, count?: number, page?: number, mustHavePictures?: boolean, withCount?: boolean, withPictures?: boolean, autoExpand: boolean, defaultCurrency: string }): Promise<TransactionPageWrapper> {
        let actualMaxTime = transactionsNextTimeId.value;

        if (reload && transactionsFilter.value.maxTime > 0) {
            actualMaxTime = transactionsFilter.value.maxTime * 1000 + 999;
        } else if (reload && transactionsFilter.value.maxTime <= 0) {
            actualMaxTime = 0;
        }

        return new Promise((resolve, reject) => {
            services.getTransactions({
                maxTime: actualMaxTime,
                minTime: transactionsFilter.value.minTime * 1000,
                count: count || 50,
                page: page || 1,
                withCount: !!withCount,
                withPictures: !!withPictures,
                mustHavePictures: !!mustHavePictures,
                type: transactionsFilter.value.type,
                categoryIds: transactionsFilter.value.categoryIds,
                accountIds: transactionsFilter.value.accountIds,
                tagFilter: transactionsFilter.value.tagFilter,
                amountFilter: transactionsFilter.value.amountFilter,
                keyword: transactionsFilter.value.keyword,
                matchMode: transactionsFilter.value.matchMode
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (reload) {
                        loadTransactionList({
                            transactionPageWrapper: EMPTY_TRANSACTION_RESULT,
                            reload: reload,
                            autoExpand: autoExpand,
                            defaultCurrency: defaultCurrency
                        });

                        if (!transactionListStateInvalid.value) {
                            updateTransactionListInvalidState(true);
                        }
                    }

                    reject({ message: 'Unable to retrieve transaction list' });
                    return;
                }

                const transactionPageWrapper: TransactionPageWrapper = {
                    items: Transaction.ofMulti(data.result.items),
                    totalCount: data.result.totalCount
                };

                loadTransactionList({
                    transactionPageWrapper: transactionPageWrapper,
                    reload: !!reload,
                    autoExpand: autoExpand,
                    defaultCurrency: defaultCurrency,
                    nextTimeSequenceId: data.result.nextTimeSequenceId
                });

                if (reload) {
                    if (transactionListStateInvalid.value) {
                        updateTransactionListInvalidState(false);
                    }
                }

                resolve(transactionPageWrapper);
            }).catch(error => {
                logger.error('failed to load transaction list', error);

                if (reload) {
                    loadTransactionList({
                        transactionPageWrapper: EMPTY_TRANSACTION_RESULT,
                        reload: reload,
                        autoExpand: autoExpand,
                        defaultCurrency: defaultCurrency
                    });

                    if (!transactionListStateInvalid.value) {
                        updateTransactionListInvalidState(true);
                    }
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve transaction list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function loadMonthlyAllTransactions({ year, month, mustHavePictures, withPictures, autoExpand, defaultCurrency }: { year: number, month: number, mustHavePictures?: boolean, withPictures?: boolean, autoExpand: boolean, defaultCurrency: string }): Promise<TransactionPageWrapper> {
        return new Promise((resolve, reject) => {
            services.getAllTransactionsByMonth({
                year: year,
                month: month,
                type: transactionsFilter.value.type,
                categoryIds: transactionsFilter.value.categoryIds,
                accountIds: transactionsFilter.value.accountIds,
                tagFilter: transactionsFilter.value.tagFilter,
                amountFilter: transactionsFilter.value.amountFilter,
                keyword: transactionsFilter.value.keyword,
                matchMode: transactionsFilter.value.matchMode,
                mustHavePictures: !!mustHavePictures,
                withPictures: !!withPictures
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    loadTransactionList({
                        transactionPageWrapper: EMPTY_TRANSACTION_RESULT,
                        reload: true,
                        autoExpand: autoExpand,
                        defaultCurrency: defaultCurrency
                    });

                    if (!transactionListStateInvalid.value) {
                        updateTransactionListInvalidState(true);
                    }

                    reject({ message: 'Unable to retrieve transaction list' });
                    return;
                }

                const transactionPageWrapper: TransactionPageWrapper = {
                    items: Transaction.ofMulti(data.result.items),
                    totalCount: data.result.totalCount
                };

                loadTransactionList({
                    transactionPageWrapper: transactionPageWrapper,
                    reload: true,
                    autoExpand: autoExpand,
                    defaultCurrency: defaultCurrency
                });

                if (transactionListStateInvalid.value) {
                    updateTransactionListInvalidState(false);
                }

                resolve(transactionPageWrapper);
            }).catch(error => {
                logger.error('failed to load monthly all transaction list', error);

                loadTransactionList({
                    transactionPageWrapper: EMPTY_TRANSACTION_RESULT,
                    reload: true,
                    autoExpand: autoExpand,
                    defaultCurrency: defaultCurrency
                });

                if (!transactionListStateInvalid.value) {
                    updateTransactionListInvalidState(true);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve transaction list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { loadTransactions, loadMonthlyAllTransactions };
}
