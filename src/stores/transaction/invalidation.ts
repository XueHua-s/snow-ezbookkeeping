import { DateRange } from '@/core/datetime.ts';
import { KeywordMatchMode } from '@/core/text.ts';
import type { useTransactionState } from './state.ts';

interface Context {
    transactionListStateInvalid: ReturnType<typeof useTransactionState>['transactionListStateInvalid'];
    transactionReconciliationStatementStateInvalid: ReturnType<typeof useTransactionState>['transactionReconciliationStatementStateInvalid'];
    accountsStore: ReturnType<typeof useTransactionState>['accountsStore'];
    overviewStore: ReturnType<typeof useTransactionState>['overviewStore'];
    statisticsStore: ReturnType<typeof useTransactionState>['statisticsStore'];
    explorersStore: ReturnType<typeof useTransactionState>['explorersStore'];
    transactionsFilter: ReturnType<typeof useTransactionState>['transactionsFilter'];
    transactions: ReturnType<typeof useTransactionState>['transactions'];
    transactionsNextTimeId: ReturnType<typeof useTransactionState>['transactionsNextTimeId'];
}

// Owns invalidation for the transaction store.
export function useTransactionInvalidation(context: Context) {
    const {
        transactionListStateInvalid, transactionReconciliationStatementStateInvalid, accountsStore, overviewStore,
        statisticsStore, explorersStore, transactionsFilter, transactions, transactionsNextTimeId,
    } = context;

    function updateTransactionListInvalidState(invalidState: boolean): void {
        transactionListStateInvalid.value = invalidState;
    }

    function updateTransactionReconciliationStatementInvalidState(invalidState: boolean): void {
        transactionReconciliationStatementStateInvalid.value = invalidState;
    }

    function updateStoreInvalidState(options: { transactionList?: boolean, reconciliationStatement?: boolean, accountList?: boolean, overview?: boolean, statistics?: boolean, explorer?: boolean }): void {
        if (options.transactionList && !transactionListStateInvalid.value) {
            updateTransactionListInvalidState(true);
        }

        if (options.reconciliationStatement && !transactionReconciliationStatementStateInvalid.value) {
            updateTransactionReconciliationStatementInvalidState(true);
        }

        if (options.accountList && !accountsStore.accountListStateInvalid) {
            accountsStore.updateAccountListInvalidState(true);
        }

        if (options.overview && !overviewStore.transactionOverviewStateInvalid) {
            overviewStore.updateTransactionOverviewInvalidState(true);
        }

        if (options.statistics && !statisticsStore.transactionStatisticsStateInvalid) {
            statisticsStore.updateTransactionStatisticsInvalidState(true);
        }

        if (options.explorer && !explorersStore.transactionExplorerStateInvalid) {
            explorersStore.updateTransactionExplorerInvalidState(true);
        }
    }

    function resetTransactions(): void {
        transactionsFilter.value.dateType = DateRange.All.type;
        transactionsFilter.value.maxTime = 0;
        transactionsFilter.value.minTime = 0;
        transactionsFilter.value.type = 0;
        transactionsFilter.value.categoryIds = '';
        transactionsFilter.value.accountIds = '';
        transactionsFilter.value.tagFilter = '';
        transactionsFilter.value.amountFilter = '';
        transactionsFilter.value.keyword = '';
        transactionsFilter.value.matchMode = KeywordMatchMode.Default.type;
        transactions.value = [];
        transactionsNextTimeId.value = 0;
        transactionListStateInvalid.value = true;
        transactionReconciliationStatementStateInvalid.value = true;
    }

    function clearTransactions(): void {
        transactions.value = [];
        transactionsNextTimeId.value = 0;
        transactionListStateInvalid.value = true;
    }

    return { updateTransactionListInvalidState, updateTransactionReconciliationStatementInvalidState, updateStoreInvalidState, resetTransactions, clearTransactions };
}
