import { itemAndIndex } from '@/core/base.ts';
import { parseDateTimeFromUnixTimeWithTimezoneOffset } from '@/lib/datetime.ts';
import { BIG_DECIMAL_ZERO } from '@/lib/numeral.ts';
import type { TransactionInfoResponse, TransactionPageWrapper, Transaction } from '@/models/transaction.ts';
import type { useTransactionDraft } from './draft.ts';
import type { useTransactionState } from './state.ts';
import type { useTransactionTotals } from './totals.ts';
import type { TransactionMonthList } from './types.ts';

interface Context {
    transactions: ReturnType<typeof useTransactionState>['transactions'];
    fillTransactionObject: ReturnType<typeof useTransactionDraft>['fillTransactionObject'];
    calculateMonthTotalAmount: ReturnType<typeof useTransactionTotals>['calculateMonthTotalAmount'];
    transactionsFilter: ReturnType<typeof useTransactionState>['transactionsFilter'];
    transactionsNextTimeId: ReturnType<typeof useTransactionState>['transactionsNextTimeId'];
    transactionListStateInvalid: ReturnType<typeof useTransactionState>['transactionListStateInvalid'];
    allFilterCategoryIds: ReturnType<typeof useTransactionState>['allFilterCategoryIds'];
    allFilterAccountIds: ReturnType<typeof useTransactionState>['allFilterAccountIds'];
}

// Owns list for the transaction store.
export function useTransactionList(context: Context) {
    const {
        transactions, fillTransactionObject, calculateMonthTotalAmount, transactionsFilter, transactionsNextTimeId,
        transactionListStateInvalid, allFilterCategoryIds, allFilterAccountIds,
    } = context;

    function loadTransactionList({ transactionPageWrapper, reload, autoExpand, defaultCurrency, nextTimeSequenceId }: { transactionPageWrapper: TransactionPageWrapper, reload: boolean, autoExpand: boolean, defaultCurrency: string, nextTimeSequenceId?: number }): void {
        if (reload) {
            transactions.value = [];
        }

        if (transactionPageWrapper.items && transactionPageWrapper.items.length) {
            let currentMonthListIndex = -1;
            let currentMonthList: TransactionMonthList | null = null;

            for (const [item, index] of itemAndIndex(transactionPageWrapper.items)) {
                fillTransactionObject(item);

                const transactionTime = parseDateTimeFromUnixTimeWithTimezoneOffset(item.time, item.utcOffset);
                const transactionYear = transactionTime.getGregorianCalendarYear();
                const transactionMonth = transactionTime.getGregorianCalendarMonth();
                const transactionYearDashMonth = transactionTime.getGregorianCalendarYearDashMonth();

                if (index === 0 && transactions.value.length > 0) {
                    const lastMonthList = transactions.value[transactions.value.length - 1] as TransactionMonthList;

                    if (lastMonthList.inflowOutflowTotalAmount.incompleteExpense || lastMonthList.inflowOutflowTotalAmount.incompleteIncome
                        || lastMonthList.incomeExpenseTotalAmount.incompleteExpense || lastMonthList.incomeExpenseTotalAmount.incompleteIncome) {
                        // calculate the total amount of last month which has incomplete total amount before starting to process a new request
                        calculateMonthTotalAmount(lastMonthList, defaultCurrency, transactionsFilter.value.accountIds, false);
                    }
                }

                if (currentMonthList && currentMonthList.year === transactionYear && currentMonthList.month === transactionMonth) {
                    currentMonthList.items.push(Object.freeze(item));

                    if (index === transactionPageWrapper.items.length - 1) {
                        // calculate the total amount of current month when processing the last transaction item of this request
                        calculateMonthTotalAmount(currentMonthList, defaultCurrency, transactionsFilter.value.accountIds, true);
                    }
                    continue;
                }

                for (let j = currentMonthListIndex + 1; j < transactions.value.length; j++) {
                    if (transactions.value[j]!.year === transactionYear && transactions.value[j]!.month === transactionMonth) {
                        currentMonthListIndex = j;
                        currentMonthList = transactions.value[j] as TransactionMonthList;
                        break;
                    }
                }

                if (!currentMonthList || currentMonthList.year !== transactionYear || currentMonthList.month !== transactionMonth) {
                    // calculate the total amount of current month when processing the first transaction item of the next month
                    calculateMonthTotalAmount(currentMonthList, defaultCurrency, transactionsFilter.value.accountIds, false);

                    const monthList: TransactionMonthList = {
                        year: transactionYear,
                        month: transactionMonth,
                        yearDashMonth: transactionYearDashMonth,
                        opened: autoExpand,
                        items: [],
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

                    transactions.value.push(monthList);

                    currentMonthListIndex = transactions.value.length - 1;
                    currentMonthList = transactions.value[transactions.value.length - 1] as TransactionMonthList;
                }

                currentMonthList.items.push(Object.freeze(item));
                // init the total amount struct of current month when processing the first transaction item of current month
                calculateMonthTotalAmount(currentMonthList, defaultCurrency, transactionsFilter.value.accountIds, true);
            }
        }

        if (nextTimeSequenceId) {
            transactionsNextTimeId.value = nextTimeSequenceId;
        } else {
            calculateMonthTotalAmount(transactions.value[transactions.value.length - 1] as TransactionMonthList, defaultCurrency, transactionsFilter.value.accountIds, false);
            transactionsNextTimeId.value = -1;
        }
    }

    function updateTransactionInTransactionList({ currentTransaction, defaultCurrency }: { currentTransaction: Transaction, defaultCurrency: string }): void {
        const transactionTime = parseDateTimeFromUnixTimeWithTimezoneOffset(currentTransaction.time, currentTransaction.utcOffset);
        const transactionYear = transactionTime.getGregorianCalendarYear();
        const transactionMonth = transactionTime.getGregorianCalendarMonth();

        for (const [transactionMonthList, monthIndex] of itemAndIndex(transactions.value)) {
            if (!transactionMonthList.items) {
                continue;
            }

            for (const [transaction, transactionIndex] of itemAndIndex(transactionMonthList.items)) {
                if (transaction.id === currentTransaction.id) {
                    fillTransactionObject(currentTransaction);

                    if (transactionYear !== transactionMonthList.year ||
                        transactionMonth !== transactionMonthList.month ||
                        currentTransaction.gregorianCalendarDayOfMonth !== transaction.gregorianCalendarDayOfMonth) {
                        transactionListStateInvalid.value = true;
                        return;
                    }

                    if ((transactionsFilter.value.categoryIds && !allFilterCategoryIds.value[currentTransaction.categoryId]) ||
                        (transactionsFilter.value.accountIds && !allFilterAccountIds.value[currentTransaction.sourceAccountId] && !allFilterAccountIds.value[currentTransaction.destinationAccountId] &&
                            (!currentTransaction.sourceAccount || !allFilterAccountIds.value[currentTransaction.sourceAccount.parentId]) &&
                            (!currentTransaction.destinationAccount || !allFilterAccountIds.value[currentTransaction.destinationAccount.parentId])
                        )
                    ) {
                        transactionMonthList.items.splice(transactionIndex, 1);
                    } else {
                        transactionMonthList.items.splice(transactionIndex, 1, currentTransaction);
                    }

                    if (transactionMonthList.items.length < 1) {
                        transactions.value.splice(monthIndex, 1);
                    } else {
                        calculateMonthTotalAmount(transactionMonthList, defaultCurrency, transactionsFilter.value.accountIds, monthIndex >= transactions.value.length - 1 && transactionsNextTimeId.value > 0);
                    }

                    return;
                }
            }
        }
    }

    function removeTransactionFromTransactionList({ currentTransaction, defaultCurrency }: { currentTransaction: TransactionInfoResponse, defaultCurrency: string }): void {
        for (const [transactionMonthList, monthIndex] of itemAndIndex(transactions.value)) {
            if (!transactionMonthList.items ||
                transactionMonthList.items[0]!.time < currentTransaction.time ||
                transactionMonthList.items[transactionMonthList.items.length - 1]!.time > currentTransaction.time) {
                continue;
            }

            for (const [transaction, transactionIndex] of itemAndIndex(transactionMonthList.items)) {
                if (transaction.id === currentTransaction.id) {
                    transactionMonthList.items.splice(transactionIndex, 1);
                }
            }

            if (transactionMonthList.items.length < 1) {
                transactions.value.splice(monthIndex, 1);
            } else {
                calculateMonthTotalAmount(transactionMonthList, defaultCurrency, transactionsFilter.value.accountIds, monthIndex >= transactions.value.length - 1 && transactionsNextTimeId.value > 0);
            }
        }
    }

    function collapseMonthInTransactionList({ monthList, collapse }: { monthList: TransactionMonthList, collapse: boolean }): void {
        if (monthList) {
            monthList.opened = !collapse;
        }
    }

    return { loadTransactionList, updateTransactionInTransactionList, removeTransactionFromTransactionList, collapseMonthInTransactionList };
}
