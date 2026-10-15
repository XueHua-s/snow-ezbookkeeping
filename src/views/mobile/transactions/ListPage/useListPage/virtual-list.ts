import {
    type TextualYearMonth
} from '@/core/datetime.ts';
import {
    isNumber
} from '@/lib/common.ts';
import { BIG_DECIMAL_ZERO } from '@/lib/numeral.ts';
import {
    getElementActualHeights,
    getElementBoundingRect
} from '@/lib/ui/mobile.ts';
import { type Transaction } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import { type TransactionMonthList } from '@/stores/transaction.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import { computed, nextTick } from 'vue';
import type { useListPageState } from './state.ts';

interface Context {
    loading: ReturnType<typeof useListPageState>['loading'];
    pageType: ReturnType<typeof useListPageState>['pageType'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    queryMonthlyData: ReturnType<typeof useListPageState>['queryMonthlyData'];
    currentMonthTransactionData: ReturnType<typeof useListPageState>['currentMonthTransactionData'];
    currentCalendarDate: ReturnType<typeof useListPageState>['currentCalendarDate'];
    transactionYearMonthListHeights: ReturnType<typeof useListPageState>['transactionYearMonthListHeights'];
    transactionInvisibleYearMonths: ReturnType<typeof useListPageState>['transactionInvisibleYearMonths'];
}

export function useListPageVirtualList(context: Context) {
    const {
        loading, pageType, transactionsStore, queryMonthlyData, currentMonthTransactionData, currentCalendarDate,
        transactionYearMonthListHeights, transactionInvisibleYearMonths,
    } = context;

    const transactions = computed<TransactionMonthList[]>(() => {
        if (loading.value) {
            return [];
        }

        if (pageType.value === TransactionListPageType.List.type || pageType.value === TransactionListPageType.Gallery.type) {
            return transactionsStore.transactions;
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

                const dailyTransactionList: TransactionMonthList = {
                    year: currentMonthTransactionData.value.year,
                    month: currentMonthTransactionData.value.month,
                    yearDashMonth: currentMonthTransactionData.value.yearDashMonth,
                    opened: true,
                    items: transactions,
                    inflowOutflowTotalAmount: {
                        transactionCount: transactions.length,
                        income: BIG_DECIMAL_ZERO,
                        expense: BIG_DECIMAL_ZERO,
                        incompleteIncome: false,
                        incompleteExpense: false
                    },
                    inflowOutflowDailyTotalAmounts: {},
                    incomeExpenseTotalAmount: {
                        transactionCount: transactions.length,
                        income: BIG_DECIMAL_ZERO,
                        expense: BIG_DECIMAL_ZERO,
                        incompleteIncome: false,
                        incompleteExpense: false
                    },
                    incomeExpenseDailyTotalAmounts: {}
                };

                return [dailyTransactionList];
            } else {
                return [];
            }
        } else {
            return [];
        }
    });

    const noTransaction = computed<boolean>(() => {
        if (pageType.value === TransactionListPageType.List.type || pageType.value === TransactionListPageType.Gallery.type) {
            return transactionsStore.noTransaction;
        } else if (pageType.value === TransactionListPageType.Calendar.type) {
            return !transactions.value || !transactions.value.length || !transactions.value[0]!.items || !transactions.value[0]!.items.length;
        } else {
            return true;
        }
    });

    const hasMoreTransaction = computed<boolean>(() => transactionsStore.hasMoreTransaction);

    function getTransactionMonthTitleDomId(yearMonth: TextualYearMonth): string {
        return 'transaction_month_title_' + yearMonth;
    }

    function getTransactionMonthListDomId(yearMonth: TextualYearMonth): string {
        return 'transaction_month_list_' + yearMonth;
    }

    function getTransactionDomId(transaction: Transaction): string {
        return 'transaction_' + transaction.id;
    }

    function isTransactionMonthListInvisible(transactionMonthList: TransactionMonthList): boolean {
        if (!transactionYearMonthListHeights.value[transactionMonthList.yearDashMonth]) {
            return false;
        }

        if (!transactionMonthList.opened) {
            return true;
        }

        if (transactionInvisibleYearMonths.value[transactionMonthList.yearDashMonth]) {
            return true;
        }

        return false;
    }

    function getTransactionMonthListHeight(transactionMonthList: TransactionMonthList): string {
        if (isTransactionMonthListInvisible(transactionMonthList)) {
            return transactionYearMonthListHeights.value[transactionMonthList.yearDashMonth] + 'px';
        }

        return 'auto';
    }

    function setTransactionMonthListHeights(reset: boolean): Promise<unknown> {
        return nextTick(() => {
            if (reset) {
                transactionInvisibleYearMonths.value = {};
                transactionYearMonthListHeights.value = {};
            }

            if (transactions.value && transactions.value.length) {
                const heights: Record<string, number> = getElementActualHeights('.transaction-month-list');

                for (let i = 0; i < transactions.value.length - 1; i++) {
                    const transactionMonthList = transactions.value[i] as TransactionMonthList;
                    const yearDashMonth = transactionMonthList.yearDashMonth;
                    const domId = getTransactionMonthListDomId(yearDashMonth);
                    const height = heights[domId];

                    if (!transactionYearMonthListHeights.value[yearDashMonth] && isNumber(height)) {
                        transactionYearMonthListHeights.value[yearDashMonth] = height;
                    }
                }
            }
        });
    }

    function setTransactionInvisibleYearMonthList(): void {
        if (!transactions.value || !transactions.value.length) {
            return;
        }

        for (let i = 0; i < transactions.value.length - 1; i++) {
            const transactionMonthList = transactions.value[i] as TransactionMonthList;
            const yearDashMonth = transactionMonthList.yearDashMonth;

            const titleDomId = getTransactionMonthTitleDomId(yearDashMonth);
            const titleRect = getElementBoundingRect(`#${titleDomId}`);

            if (!titleRect) {
                continue;
            }

            const listHeight = transactionYearMonthListHeights.value[yearDashMonth] || 0;
            const listRectTop = titleRect.top + titleRect.height;
            const listRectBottom = listRectTop + listHeight;
            const invisible = listRectTop > 2 * window.innerHeight || listRectBottom < -2 * window.innerHeight;

            if (invisible) {
                transactionInvisibleYearMonths.value[yearDashMonth] = true;
            } else {
                delete transactionInvisibleYearMonths.value[yearDashMonth];
            }
        }
    }

    function getTransactionDateStyle(transaction: Transaction, previousTransaction: Transaction | undefined): Record<string, string> {
        if (!previousTransaction || transaction.gregorianCalendarDayOfMonth !== previousTransaction.gregorianCalendarDayOfMonth) {
            return {};
        }

        return {
            color: 'transparent'
        };
    }

    function getCategoryListItemCheckedClass(category: TransactionCategory, queryCategoryIds: Record<string, boolean>): Record<string, boolean> {
        if (queryCategoryIds && queryCategoryIds[category.id]) {
            return {
                'list-item-checked': true
            };
        }

        if (category.subCategories) {
            for (const subCategory of category.subCategories) {
                if (queryCategoryIds && queryCategoryIds[subCategory.id]) {
                    return {
                        'list-item-checked': true
                    };
                }
            }
        }

        return {};
    }

    function collapseTransactionMonthList(monthList: TransactionMonthList, collapse: boolean): void {
        transactionsStore.collapseMonthInTransactionList({
            monthList: monthList,
            collapse: collapse
        });

        if (!collapse && transactionInvisibleYearMonths.value[monthList.yearDashMonth]) {
            delete transactionInvisibleYearMonths.value[monthList.yearDashMonth];
        }
    }

    function onResize(): void {
        setTransactionMonthListHeights(true)
            .then(() => {
                setTransactionMonthListHeights(false);
            });
    }

    function onScroll(): void {
        setTransactionInvisibleYearMonthList();
    }

    function onTransactionMonthListCollapseStateChanged(): void {
        setTransactionMonthListHeights(false)
            .then(() => {
                setTransactionInvisibleYearMonthList();
            });
    }

    return {
        transactions, noTransaction, hasMoreTransaction, getTransactionMonthTitleDomId, getTransactionMonthListDomId,
        getTransactionDomId, isTransactionMonthListInvisible, getTransactionMonthListHeight, setTransactionMonthListHeights,
        getTransactionDateStyle, getCategoryListItemCheckedClass, collapseTransactionMonthList, onResize, onScroll,
        onTransactionMonthListCollapseStateChanged,
    };
}
