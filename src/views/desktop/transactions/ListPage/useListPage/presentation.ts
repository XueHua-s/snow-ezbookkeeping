import { AmountFilterType } from '@/core/numeral.ts';
import { type Transaction } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import { type TransactionListDisplayTotalAmount, TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import { computed } from 'vue';
import type { useListPageLoading } from './loading.ts';
import type { useListPageState } from './state.ts';

interface Context {
    pageType: ReturnType<typeof useListPageState>['pageType'];
    queryMonthlyData: ReturnType<typeof useListPageState>['queryMonthlyData'];
    currentMonthTransactionData: ReturnType<typeof useListPageState>['currentMonthTransactionData'];
    currentPage: ReturnType<typeof useListPageState>['currentPage'];
    countPerPage: ReturnType<typeof useListPageLoading>['countPerPage'];
    currentPageTransactions: ReturnType<typeof useListPageState>['currentPageTransactions'];
    currentCalendarDate: ReturnType<typeof useListPageState>['currentCalendarDate'];
    totalCount: ReturnType<typeof useListPageState>['totalCount'];
    getDisplayMonthTotalAmount: ReturnType<typeof useListPageState>['getDisplayMonthTotalAmount'];
}

export function useListPagePresentation(context: Context) {
    const {
        pageType, queryMonthlyData, currentMonthTransactionData, currentPage, countPerPage, currentPageTransactions,
        currentCalendarDate, totalCount, getDisplayMonthTotalAmount,
    } = context;

    const transactions = computed<Transaction[]>(() => {
        if (pageType.value === TransactionListPageType.List.type || pageType.value === TransactionListPageType.Gallery.type) {
            if (queryMonthlyData.value) {
                const transactionData = currentMonthTransactionData.value;

                if (!transactionData || !transactionData.items) {
                    return [];
                }

                const firstIndex = (currentPage.value - 1) * countPerPage.value;
                const lastIndex = currentPage.value * countPerPage.value;

                return transactionData.items.slice(firstIndex, lastIndex);
            } else {
                return currentPageTransactions.value;
            }
        } else if (pageType.value === TransactionListPageType.Calendar.type) {
            if (queryMonthlyData.value) {
                const transactionData = currentMonthTransactionData.value;

                if (!transactionData || !transactionData.items) {
                    return [];
                }

                const transactions :Transaction[] = [];

                for (const transaction of transactionData.items) {
                    if (transaction.gregorianCalendarYearDashMonthDashDay === currentCalendarDate.value) {
                        transactions.push(transaction);
                    }
                }

                return transactions;
            } else {
                return [];
            }
        } else {
            return [];
        }
    });

    const transactionsByDay = computed<Record<string, Transaction[]>>(() => {
        const transactionsByDay: Record<string, Transaction[]> = {};

        for (const transaction of transactions.value) {
            if (!transaction.gregorianCalendarYearDashMonthDashDay) {
                continue;
            }

            const transactions: Transaction[] = transactionsByDay[transaction.gregorianCalendarYearDashMonthDashDay] ?? [];
            transactions.push(transaction);
            transactionsByDay[transaction.gregorianCalendarYearDashMonthDashDay] = transactions;
        }

        return transactionsByDay;
    });

    const totalPageCount = computed<number>(() => Math.ceil(totalCount.value / countPerPage.value));

    const skeletonData = computed<number[]>(() => {
        const data: number[] = [];
        const totalCount = (pageType.value === TransactionListPageType.List.type ? countPerPage.value : 3);

        for (let i = 0; i < totalCount; i++) {
            data.push(i);
        }

        return data;
    });

    const currentMonthTotalAmount = computed<TransactionListDisplayTotalAmount | null>(() => {
        if (queryMonthlyData.value) {
            return getDisplayMonthTotalAmount(currentMonthTransactionData.value, '');
        } else {
            return null;
        }
    });

    function getCategoryListItemCheckedClass(category: TransactionCategory, queryCategoryIds: Record<string, boolean>): Record<string, boolean> {
        if (queryCategoryIds && queryCategoryIds[category.id]) {
            return {
                'list-item-selected': true,
                'has-children-item-selected': true
            };
        }

        if (category.subCategories) {
            for (const subCategory of category.subCategories) {
                if (queryCategoryIds && queryCategoryIds[subCategory.id]) {
                    return {
                        'list-item-selected': true,
                        'has-children-item-selected': true
                    };
                }
            }
        }

        return {};
    }

    function getAmountFilterParameterCount(filterType: string): number {
        const amountFilterType = AmountFilterType.valueOf(filterType);
        return amountFilterType ? amountFilterType.paramCount : 0;
    }

    return { transactions, transactionsByDay, totalPageCount, skeletonData, currentMonthTotalAmount, getCategoryListItemCheckedClass, getAmountFilterParameterCount };
}
