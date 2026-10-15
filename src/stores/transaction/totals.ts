import { entries, keys } from '@/core/base.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    isNumber
} from '@/lib/common.ts';
import { BIG_DECIMAL_ZERO, parseBigDecimal } from '@/lib/numeral.ts';
import type { useTransactionState } from './state.ts';
import type { TransactionMonthList, TransactionTotalAmount } from './types.ts';

interface Context {
    exchangeRatesStore: ReturnType<typeof useTransactionState>['exchangeRatesStore'];
}

// Owns totals for the transaction store.
export function useTransactionTotals(context: Context) {
    const { exchangeRatesStore } = context;

    function calculateMonthTotalAmount(transactionMonthList: TransactionMonthList | null, defaultCurrency: string, accountIds: string, incomplete: boolean): void {
        if (!transactionMonthList) {
            return;
        }

        const inflowOutflowDailyTotalAmounts: Record<string, TransactionTotalAmount> = {};
        const incomeExpenseDailyTotalAmounts: Record<string, TransactionTotalAmount> = {};
        let totalOutflow: BigDecimal = BIG_DECIMAL_ZERO;
        let totalInflow: BigDecimal = BIG_DECIMAL_ZERO;
        let totalExpense: BigDecimal = BIG_DECIMAL_ZERO;
        let totalIncome: BigDecimal = BIG_DECIMAL_ZERO;
        let hasUnCalculatedTotalOutflow = false;
        let hasUnCalculatedTotalInflow = false;
        let hasUnCalculatedTotalExpense = false;
        let hasUnCalculatedTotalIncome = false;

        const allAccountIdsMap: Record<string, boolean> = {};
        let totalAccountIdsCount = 0;

        if (accountIds && accountIds !== '0') {
            const allAccountIdsArray = accountIds.split(',');

            for (const accountId of allAccountIdsArray) {
                if (accountId) {
                    allAccountIdsMap[accountId] = true;
                    totalAccountIdsCount++;
                }
            }
        }

        for (const transaction of transactionMonthList.items) {
            const transactionDay = isNumber(transaction.gregorianCalendarDayOfMonth) ? transaction.gregorianCalendarDayOfMonth.toString() : '0';
            let inflowOutflowDailyTotalAmount = inflowOutflowDailyTotalAmounts[transactionDay];
            let incomeExpenseDailyTotalAmount = incomeExpenseDailyTotalAmounts[transactionDay];

            if (!inflowOutflowDailyTotalAmount) {
                inflowOutflowDailyTotalAmount = {
                    transactionCount: 0,
                    expense: BIG_DECIMAL_ZERO,
                    incompleteExpense: false,
                    income: BIG_DECIMAL_ZERO,
                    incompleteIncome: false
                };
                inflowOutflowDailyTotalAmounts[transactionDay] = inflowOutflowDailyTotalAmount;
            }

            if (!incomeExpenseDailyTotalAmount) {
                incomeExpenseDailyTotalAmount = {
                    transactionCount: 0,
                    expense: BIG_DECIMAL_ZERO,
                    incompleteExpense: false,
                    income: BIG_DECIMAL_ZERO,
                    incompleteIncome: false
                };
                incomeExpenseDailyTotalAmounts[transactionDay] = incomeExpenseDailyTotalAmount;
            }

            let amount: BigDecimal = parseBigDecimal(transaction.sourceAmount);
            let account = transaction.sourceAccount;

            if (totalAccountIdsCount > 0 && transaction.destinationAccount
                && (!allAccountIdsMap[transaction.sourceAccount?.id || ''] && !allAccountIdsMap[transaction.sourceAccount?.parentId || ''])
                && (allAccountIdsMap[transaction.destinationAccount.id] || allAccountIdsMap[transaction.destinationAccount.parentId])) {
                amount = parseBigDecimal(transaction.destinationAmount);
                account = transaction.destinationAccount;
            }

            if (!account) {
                continue;
            }

            inflowOutflowDailyTotalAmount.transactionCount++;
            incomeExpenseDailyTotalAmount.transactionCount++;

            if (account.currency !== defaultCurrency) {
                const balance = exchangeRatesStore.getExchangedAmount(amount, account.currency, defaultCurrency);

                if (!balance) {
                    if (transaction.type === TransactionType.Expense) {
                        hasUnCalculatedTotalOutflow = true;
                        hasUnCalculatedTotalExpense = true;
                        inflowOutflowDailyTotalAmount.incompleteExpense = true;
                        incomeExpenseDailyTotalAmount.incompleteExpense = true;
                    } else if (transaction.type === TransactionType.Income) {
                        hasUnCalculatedTotalInflow = true;
                        hasUnCalculatedTotalIncome = true;
                        inflowOutflowDailyTotalAmount.incompleteIncome = true;
                        incomeExpenseDailyTotalAmount.incompleteIncome = true;
                    } else if (transaction.type === TransactionType.Transfer) {
                        if (allAccountIdsMap[transaction.sourceAccountId] && allAccountIdsMap[transaction.destinationAccountId]) {
                            // Do Nothing
                        } else if (transaction.sourceAccount && transaction.destinationAccount && allAccountIdsMap[transaction.sourceAccount.parentId] && allAccountIdsMap[transaction.destinationAccount.parentId]) {
                            // Do Nothing
                        } else if (transaction.sourceAccount && allAccountIdsMap[transaction.sourceAccount.parentId] && allAccountIdsMap[transaction.destinationAccountId]) {
                            // Do Nothing
                        } else if (transaction.destinationAccount && allAccountIdsMap[transaction.sourceAccountId] && allAccountIdsMap[transaction.destinationAccount.parentId]) {
                            // Do Nothing
                        } else if (allAccountIdsMap[transaction.sourceAccountId] || (transaction.sourceAccount && allAccountIdsMap[transaction.sourceAccount.parentId])) {
                            hasUnCalculatedTotalOutflow = true;
                            inflowOutflowDailyTotalAmount.incompleteExpense = true;
                        } else if (allAccountIdsMap[transaction.destinationAccountId] || (transaction.destinationAccount && allAccountIdsMap[transaction.destinationAccount.parentId])) {
                            hasUnCalculatedTotalInflow = true;
                            inflowOutflowDailyTotalAmount.incompleteIncome = true;
                        }
                    }

                    continue;
                }

                amount = balance;
            }

            if (transaction.type === TransactionType.Expense) {
                totalOutflow = totalOutflow.add(amount);
                totalExpense = totalExpense.add(amount);
                inflowOutflowDailyTotalAmount.expense = inflowOutflowDailyTotalAmount.expense.add(amount);
                incomeExpenseDailyTotalAmount.expense = incomeExpenseDailyTotalAmount.expense.add(amount);
            } else if (transaction.type === TransactionType.Income) {
                totalInflow = totalInflow.add(amount);
                totalIncome = totalIncome.add(amount);
                inflowOutflowDailyTotalAmount.income = inflowOutflowDailyTotalAmount.income.add(amount);
                incomeExpenseDailyTotalAmount.income = incomeExpenseDailyTotalAmount.income.add(amount);
            } else if (transaction.type === TransactionType.Transfer && totalAccountIdsCount > 0) {
                if (allAccountIdsMap[transaction.sourceAccountId] && allAccountIdsMap[transaction.destinationAccountId]) {
                    // Do Nothing
                } else if (transaction.sourceAccount && transaction.destinationAccount && allAccountIdsMap[transaction.sourceAccount.parentId] && allAccountIdsMap[transaction.destinationAccount.parentId]) {
                    // Do Nothing
                } else if (transaction.sourceAccount && allAccountIdsMap[transaction.sourceAccount.parentId] && allAccountIdsMap[transaction.destinationAccountId]) {
                    // Do Nothing
                } else if (transaction.destinationAccount && allAccountIdsMap[transaction.sourceAccountId] && allAccountIdsMap[transaction.destinationAccount.parentId]) {
                    // Do Nothing
                } else if (allAccountIdsMap[transaction.sourceAccountId] || (transaction.sourceAccount && allAccountIdsMap[transaction.sourceAccount.parentId])) {
                    totalOutflow = totalOutflow.add(amount);
                    inflowOutflowDailyTotalAmount.expense = inflowOutflowDailyTotalAmount.expense.add(amount);
                } else if (allAccountIdsMap[transaction.destinationAccountId] || (transaction.destinationAccount && allAccountIdsMap[transaction.destinationAccount.parentId])) {
                    totalInflow = totalInflow.add(amount);
                    inflowOutflowDailyTotalAmount.income = inflowOutflowDailyTotalAmount.income.add(amount);
                }
            }
        }

        transactionMonthList.inflowOutflowTotalAmount.transactionCount = transactionMonthList.items.length;
        transactionMonthList.inflowOutflowTotalAmount.expense = totalOutflow.truncate();
        transactionMonthList.inflowOutflowTotalAmount.incompleteExpense = incomplete || hasUnCalculatedTotalOutflow;
        transactionMonthList.inflowOutflowTotalAmount.income = totalInflow.truncate();
        transactionMonthList.inflowOutflowTotalAmount.incompleteIncome = incomplete || hasUnCalculatedTotalInflow;

        transactionMonthList.incomeExpenseTotalAmount.transactionCount = transactionMonthList.items.length;
        transactionMonthList.incomeExpenseTotalAmount.expense = totalExpense.truncate();
        transactionMonthList.incomeExpenseTotalAmount.incompleteExpense = incomplete || hasUnCalculatedTotalExpense;
        transactionMonthList.incomeExpenseTotalAmount.income = totalIncome.truncate();
        transactionMonthList.incomeExpenseTotalAmount.incompleteIncome = incomplete || hasUnCalculatedTotalIncome;

        for (const day of keys(transactionMonthList.inflowOutflowDailyTotalAmounts)) {
            delete transactionMonthList.inflowOutflowDailyTotalAmounts[day];
        }
        for (const day of keys(transactionMonthList.incomeExpenseDailyTotalAmounts)) {
            delete transactionMonthList.incomeExpenseDailyTotalAmounts[day];
        }

        for (const [day, inflowOutflowDailyTotalAmount] of entries(inflowOutflowDailyTotalAmounts)) {
            transactionMonthList.inflowOutflowDailyTotalAmounts[day] = {
                transactionCount: inflowOutflowDailyTotalAmount.transactionCount,
                expense: inflowOutflowDailyTotalAmount.expense.truncate(),
                incompleteExpense: incomplete || inflowOutflowDailyTotalAmount.incompleteExpense,
                income: inflowOutflowDailyTotalAmount.income.truncate(),
                incompleteIncome: incomplete || inflowOutflowDailyTotalAmount.incompleteIncome
            };
        }
        for (const [day, incomeExpenseDailyTotalAmount] of entries(incomeExpenseDailyTotalAmounts)) {
            transactionMonthList.incomeExpenseDailyTotalAmounts[day] = {
                transactionCount: incomeExpenseDailyTotalAmount.transactionCount,
                expense: incomeExpenseDailyTotalAmount.expense.truncate(),
                incompleteExpense: incomplete || incomeExpenseDailyTotalAmount.incompleteExpense,
                income: incomeExpenseDailyTotalAmount.income.truncate(),
                incompleteIncome: incomplete || incomeExpenseDailyTotalAmount.incompleteIncome
            };
        }
    }

    return { calculateMonthTotalAmount };
}
