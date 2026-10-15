import { BIG_DECIMAL_ZERO } from '@/lib/numeral.ts';
import {
    type TransactionInfoResponse,
    Transaction
} from '@/models/transaction.ts';
import type { useTransactionDraft } from './draft.ts';
import type { useTransactionState } from './state.ts';
import type { useTransactionTotals } from './totals.ts';
import type { TransactionDailyTotalAmounts, TransactionMonthList } from './types.ts';

interface Context {
    fillTransactionObject: ReturnType<typeof useTransactionDraft>['fillTransactionObject'];
    calculateMonthTotalAmount: ReturnType<typeof useTransactionTotals>['calculateMonthTotalAmount'];
    userStore: ReturnType<typeof useTransactionState>['userStore'];
}

// Owns daily amounts for the transaction store.
export function useTransactionDailyAmounts(context: Context) {
    const { fillTransactionObject, calculateMonthTotalAmount, userStore } = context;

    function getCurrentMonthTransactionDailyTotalAmounts(transactions: TransactionInfoResponse[], accountIds: string): TransactionDailyTotalAmounts {
        const monthList: TransactionMonthList = {
            year: 0,
            month: 0,
            yearDashMonth: '0-0',
            opened: true,
            items: transactions.map(transaction => fillTransactionObject(Transaction.of(transaction))),
            inflowOutflowTotalAmount: {
                transactionCount: 0,
                expense: BIG_DECIMAL_ZERO,
                incompleteExpense: true,
                income: BIG_DECIMAL_ZERO,
                incompleteIncome: true
            },
            inflowOutflowDailyTotalAmounts: {},
            incomeExpenseTotalAmount: {
                transactionCount: 0,
                expense: BIG_DECIMAL_ZERO,
                incompleteExpense: true,
                income: BIG_DECIMAL_ZERO,
                incompleteIncome: true
            },
            incomeExpenseDailyTotalAmounts: {}
        };

        calculateMonthTotalAmount(monthList, userStore.currentUserDefaultCurrency, accountIds, false);

        return {
            inflowOutflowDailyTotalAmounts: monthList.inflowOutflowDailyTotalAmounts,
            incomeExpenseDailyTotalAmounts: monthList.incomeExpenseDailyTotalAmounts
        };
    }

    return { getCurrentMonthTransactionDailyTotalAmounts };
}
