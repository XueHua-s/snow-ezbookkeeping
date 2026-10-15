import { TransactionType } from '@/core/transaction.ts';
import {
    isDefined
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services,{ type ApiResponsePromise } from '@/lib/services.ts';
import {
    type TransactionInfoResponse,
    Transaction
} from '@/models/transaction.ts';
import type { useTransactionInvalidation } from './invalidation.ts';
import type { useTransactionList } from './list.ts';
import type { useTransactionState } from './state.ts';

interface Context {
    transactionListStateInvalid: ReturnType<typeof useTransactionState>['transactionListStateInvalid'];
    updateTransactionListInvalidState: ReturnType<typeof useTransactionInvalidation>['updateTransactionListInvalidState'];
    updateTransactionInTransactionList: ReturnType<typeof useTransactionList>['updateTransactionInTransactionList'];
    updateStoreInvalidState: ReturnType<typeof useTransactionInvalidation>['updateStoreInvalidState'];
}

// Owns save for the transaction store.
export function useTransactionSave(context: Context) {
    const { transactionListStateInvalid, updateTransactionListInvalidState, updateTransactionInTransactionList, updateStoreInvalidState } = context;

    function getTransaction({ transactionId, withPictures }: { transactionId: string, withPictures?: boolean }): Promise<Transaction> {
        return new Promise((resolve, reject) => {
            if (!isDefined(withPictures)) {
                withPictures = true;
            }

            services.getTransaction({
                id: transactionId,
                withPictures: withPictures
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve transaction' });
                    return;
                }

                const transaction = Transaction.of(data.result);

                resolve(transaction);
            }).catch(error => {
                logger.error('failed to load transaction info', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve transaction' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveTransaction({ transaction, defaultCurrency, isEdit, clientSessionId }: { transaction: Transaction, defaultCurrency: string, isEdit: boolean, clientSessionId: string }): Promise<Transaction> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<TransactionInfoResponse>;

            if (transaction.type !== TransactionType.Expense &&
                transaction.type !== TransactionType.Income &&
                transaction.type !== TransactionType.Transfer &&
                transaction.type !== TransactionType.ModifyBalance) {
                reject({ message: 'An error occurred' });
                return;
            } else if (!isEdit && transaction.type === TransactionType.ModifyBalance) {
                reject({ message: 'An error occurred' });
                return;
            }

            if (!isEdit) {
                promise = services.addTransaction(transaction.toCreateRequest(clientSessionId));
            } else {
                promise = services.modifyTransaction(transaction.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!isEdit) {
                        reject({ message: 'Unable to add transaction' });
                    } else {
                        reject({ message: 'Unable to save transaction' });
                    }
                }

                const transaction = Transaction.of(data.result);

                if (!isEdit) {
                    if (!transactionListStateInvalid.value) {
                        updateTransactionListInvalidState(true);
                    }
                } else {
                    updateTransactionInTransactionList({
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

                resolve(transaction);
            }).catch(error => {
                logger.error('failed to save transaction', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!isEdit) {
                        reject({ message: 'Unable to add transaction' });
                    } else {
                        reject({ message: 'Unable to save transaction' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    return { getTransaction, saveTransaction };
}
