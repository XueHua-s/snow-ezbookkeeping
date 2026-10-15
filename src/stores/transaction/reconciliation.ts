import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type { TransactionReconciliationStatementResponse } from '@/models/transaction.ts';
import type { useTransactionInvalidation } from './invalidation.ts';
import type { useTransactionState } from './state.ts';

interface Context {
    transactionReconciliationStatementStateInvalid: ReturnType<typeof useTransactionState>['transactionReconciliationStatementStateInvalid'];
    updateTransactionReconciliationStatementInvalidState: ReturnType<typeof useTransactionInvalidation>['updateTransactionReconciliationStatementInvalidState'];
}

// Owns reconciliation for the transaction store.
export function useTransactionReconciliation(context: Context) {
    const { transactionReconciliationStatementStateInvalid, updateTransactionReconciliationStatementInvalidState } = context;

    function getReconciliationStatements({ accountId, startTime, endTime }: { accountId: string, startTime: number, endTime: number }): Promise<TransactionReconciliationStatementResponse> {
        return new Promise((resolve, reject) => {
            services.getReconciliationStatements({
                accountId: accountId,
                startTime: startTime,
                endTime: endTime
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!transactionReconciliationStatementStateInvalid.value) {
                        updateTransactionReconciliationStatementInvalidState(true);
                    }

                    reject({ message: 'Unable to retrieve reconciliation statements' });
                    return;
                }

                if (transactionReconciliationStatementStateInvalid.value) {
                    updateTransactionReconciliationStatementInvalidState(false);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to load reconciliation statements', error);

                if (!transactionReconciliationStatementStateInvalid.value) {
                    updateTransactionReconciliationStatementInvalidState(true);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve reconciliation statements' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { getReconciliationStatements };
}
