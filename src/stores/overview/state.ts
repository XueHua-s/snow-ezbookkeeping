import { keys } from '@/core/base.ts';
import type { OverviewRecentTransactionsQuery } from '@/core/overview_layout.ts';
import {
    getThisMonthFirstUnixTime,
    getThisMonthLastUnixTime,
    getThisWeekFirstUnixTime,
    getThisWeekLastUnixTime,
    getThisYearFirstUnixTime,
    getThisYearLastUnixTime,
    getTodayFirstUnixTime,
    getTodayLastUnixTime,
    getUnixTimeBeforeUnixTime
} from '@/lib/datetime.ts';
import {
    BIG_DECIMAL_ZERO,
    parseBigDecimal
} from '@/lib/numeral.ts';
import type {
TransactionAmountsRequestType,
TransactionAmountsResponse,
TransactionDailyAmountsResponseItem,
TransactionOverviewData,
TransactionStatisticAssetTrendsResponseItem,
TransactionStatisticResponse,
TransactionUnreconciledCountItem
} from '@/models/transaction.ts';
import {
    type TransactionInfoResponse,
    ALL_TRANSACTION_AMOUNTS_REQUEST_TYPE
} from '@/models/transaction.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { type TransactionTotalAmount, useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useUserStore } from '@/stores/user.ts';
import type { ComputedRef, Ref, UnwrapRef } from 'vue';
import { computed, ref } from 'vue';
import type { TransactionDataRange, TransactionOverviewOptions } from './types.ts';
interface OverviewState {
    settingsStore: ReturnType<typeof useSettingsStore>;
    accountsStore: ReturnType<typeof useAccountsStore>;
    transactionCategoriesStore: ReturnType<typeof useTransactionCategoriesStore>;
    transactionsStore: ReturnType<typeof useTransactionsStore>;
    transactionDataRange: Ref<UnwrapRef<TransactionDataRange>>;
    transactionOverviewOptions: Ref<UnwrapRef<TransactionOverviewOptions>>;
    transactionOverviewData: Ref<UnwrapRef<TransactionAmountsResponse>>;
    transactionOverviewStateInvalid: Ref<UnwrapRef<boolean>>;
    transactionUnreconciledCounts: Ref<UnwrapRef<TransactionUnreconciledCountItem[]>>;
    transactionUnreconciledCountsStateInvalid: Ref<UnwrapRef<boolean>>;
    transactionUnreconciledCountsAccountIds: Ref<UnwrapRef<string[]>>;
    transactionCategoryStatisticsData: Ref<UnwrapRef<Record<number, TransactionStatisticResponse>>>;
    transactionCategoryStatisticsStateInvalid: Ref<UnwrapRef<Record<number, boolean>>>;
    transactionAssetTrendsData: Ref<UnwrapRef<TransactionStatisticAssetTrendsResponseItem[]>>;
    transactionAssetTrendsStateInvalid: Ref<UnwrapRef<boolean>>;
    recentTransactions: Ref<UnwrapRef<Record<string, TransactionInfoResponse[]>>>;
    recentTransactionsStateInvalid: Ref<UnwrapRef<Record<string, boolean>>>;
    currentMonthTransactions: Ref<UnwrapRef<TransactionInfoResponse[]>>;
    currentMonthTransactionInflowOutflowDailyTotalAmounts: Ref<UnwrapRef<Record<string, TransactionTotalAmount>>>;
    currentMonthTransactionIncomeExpenseDailyTotalAmounts: Ref<UnwrapRef<Record<string, TransactionTotalAmount>>>;
    currentMonthTransactionsStateInvalid: Ref<UnwrapRef<boolean>>;
    transactionDailyAmountsData: Ref<UnwrapRef<TransactionDailyAmountsResponseItem[]>>;
    transactionDailyAmountsStateInvalid: Ref<UnwrapRef<boolean>>;
    transactionOverview: ComputedRef<TransactionOverviewData>;
    getRecentTransactionsQueryCacheKey: (query: OverviewRecentTransactionsQuery) => string;
    updateTransactionDateRange: () => void;
    updateTransactionOverviewInvalidState: (invalidState: boolean) => void;
    resetTransactionOverview: () => void;
}



// Owns state for the overview store.
export function useOverviewState(): OverviewState  {
    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionsStore = useTransactionsStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const transactionDataRange = ref<TransactionDataRange>(getTransactionDateRange());

    const transactionOverviewOptions = ref<TransactionOverviewOptions>({
        loadedMonths: 1
    });

    const transactionOverviewData = ref<TransactionAmountsResponse>({});

    const transactionOverviewStateInvalid = ref<boolean>(true);

    const transactionUnreconciledCounts = ref<TransactionUnreconciledCountItem[]>([]);

    const transactionUnreconciledCountsStateInvalid = ref<boolean>(true);

    const transactionUnreconciledCountsAccountIds = ref<string[]>([]);

    const transactionCategoryStatisticsData = ref<Record<number, TransactionStatisticResponse>>({});

    const transactionCategoryStatisticsStateInvalid = ref<Record<number, boolean>>({});

    const transactionAssetTrendsData = ref<TransactionStatisticAssetTrendsResponseItem[]>([]);

    const transactionAssetTrendsStateInvalid = ref<boolean>(true);

    const recentTransactions = ref<Record<string, TransactionInfoResponse[]>>({});

    const recentTransactionsStateInvalid = ref<Record<string, boolean>>({});

    const currentMonthTransactions = ref<TransactionInfoResponse[]>([]);

    const currentMonthTransactionInflowOutflowDailyTotalAmounts = ref<Record<string, TransactionTotalAmount>>({});

    const currentMonthTransactionIncomeExpenseDailyTotalAmounts = ref<Record<string, TransactionTotalAmount>>({});

    const currentMonthTransactionsStateInvalid = ref<boolean>(true);

    const transactionDailyAmountsData = ref<TransactionDailyAmountsResponseItem[]>([]);

    const transactionDailyAmountsStateInvalid = ref<boolean>(true);

    const transactionOverview = computed<TransactionOverviewData>(() => {
        const overviewData = transactionOverviewData.value;

        if (!overviewData || !overviewData.thisMonth) {
            return {
                thisMonth: {
                    valid: false,
                    incomeAmount: BIG_DECIMAL_ZERO,
                    expenseAmount: BIG_DECIMAL_ZERO,
                    incompleteIncomeAmount: false,
                    incompleteExpenseAmount: false
                }
            } as TransactionOverviewData;
        }

        const finalOverviewData: TransactionOverviewData = {};
        const defaultCurrency = userStore.currentUserDefaultCurrency;

        ALL_TRANSACTION_AMOUNTS_REQUEST_TYPE.forEach(field => {
            const item = overviewData[field];

            if (!item) {
                return;
            }

            let totalIncomeAmount = BIG_DECIMAL_ZERO;
            let totalExpenseAmount = BIG_DECIMAL_ZERO;
            let hasUnCalculatedTotalIncome = false;
            let hasUnCalculatedTotalExpense = false;

            if (item.amounts) {
                for (const amount of item.amounts) {
                    if (amount.currency !== defaultCurrency) {
                        const incomeAmount = exchangeRatesStore.getExchangedAmount(parseBigDecimal(amount.incomeAmount), amount.currency, defaultCurrency);
                        const expenseAmount = exchangeRatesStore.getExchangedAmount(parseBigDecimal(amount.expenseAmount), amount.currency, defaultCurrency);

                        if (incomeAmount) {
                            totalIncomeAmount = totalIncomeAmount.add(incomeAmount.truncate());
                        } else {
                            hasUnCalculatedTotalIncome = true;
                        }

                        if (expenseAmount) {
                            totalExpenseAmount = totalExpenseAmount.add(expenseAmount.truncate());
                        } else {
                            hasUnCalculatedTotalExpense = true;
                        }
                    } else {
                        totalIncomeAmount = totalIncomeAmount.add(parseBigDecimal(amount.incomeAmount));
                        totalExpenseAmount = totalExpenseAmount.add(parseBigDecimal(amount.expenseAmount));
                    }
                }
            }

            finalOverviewData[field] = {
                valid: true,
                incomeAmount: totalIncomeAmount,
                expenseAmount: totalExpenseAmount,
                incompleteIncomeAmount: hasUnCalculatedTotalIncome,
                incompleteExpenseAmount: hasUnCalculatedTotalExpense,
                amounts: item.amounts || []
            };
        });

        return finalOverviewData;
    });

    function getTransactionDateRange(): TransactionDataRange {
        const dateRange: TransactionDataRange = {
            today: { startTime: 0, endTime: 0 },
            thisWeek: { startTime: 0, endTime: 0 },
            thisMonth: { startTime: 0, endTime: 0 },
            thisYear: { startTime: 0, endTime: 0 },
            lastMonth: { startTime: 0, endTime: 0 },
            monthBeforeLastMonth: { startTime: 0, endTime: 0 },
            monthBeforeLast2Months: { startTime: 0, endTime: 0 },
            monthBeforeLast3Months: { startTime: 0, endTime: 0 },
            monthBeforeLast4Months: { startTime: 0, endTime: 0 },
            monthBeforeLast5Months: { startTime: 0, endTime: 0 },
            monthBeforeLast6Months: { startTime: 0, endTime: 0 },
            monthBeforeLast7Months: { startTime: 0, endTime: 0 },
            monthBeforeLast8Months: { startTime: 0, endTime: 0 },
            monthBeforeLast9Months: { startTime: 0, endTime: 0 },
            monthBeforeLast10Months: { startTime: 0, endTime: 0 }
        };

        initTransactionDateRange(dateRange);
        return dateRange;
    }

    function getRecentTransactionsQueryCacheKey(query: OverviewRecentTransactionsQuery): string {
        const normalizedQuery = {...query, count: 0 };
        return JSON.stringify(normalizedQuery);
    }

    function initTransactionDateRange(dateRange: TransactionDataRange): void {
        dateRange.today.startTime = getTodayFirstUnixTime();
        dateRange.today.endTime = getTodayLastUnixTime();

        dateRange.thisWeek.startTime = getThisWeekFirstUnixTime(userStore.currentUserFirstDayOfWeek);
        dateRange.thisWeek.endTime = getThisWeekLastUnixTime(userStore.currentUserFirstDayOfWeek);

        dateRange.thisMonth.startTime = getThisMonthFirstUnixTime();
        dateRange.thisMonth.endTime = getThisMonthLastUnixTime();

        dateRange.thisYear.startTime = getThisYearFirstUnixTime();
        dateRange.thisYear.endTime = getThisYearLastUnixTime();

        dateRange.lastMonth.startTime = getUnixTimeBeforeUnixTime(getThisMonthFirstUnixTime(), 1, 'months');
        dateRange.lastMonth.endTime = getUnixTimeBeforeUnixTime(getThisMonthFirstUnixTime(), 1, 'seconds');

        dateRange.monthBeforeLastMonth.startTime = getUnixTimeBeforeUnixTime(dateRange.lastMonth.startTime, 1, 'months');
        dateRange.monthBeforeLastMonth.endTime = getUnixTimeBeforeUnixTime(dateRange.lastMonth.startTime, 1, 'seconds');

        dateRange.monthBeforeLast2Months.startTime = getUnixTimeBeforeUnixTime(dateRange.monthBeforeLastMonth.startTime, 1, 'months');
        dateRange.monthBeforeLast2Months.endTime = getUnixTimeBeforeUnixTime(dateRange.monthBeforeLastMonth.startTime, 1, 'seconds');

        for (let i = 3; i <= 10; i++) {
            dateRange[`monthBeforeLast${i}Months` as TransactionAmountsRequestType].startTime = getUnixTimeBeforeUnixTime(dateRange[`monthBeforeLast${i - 1}Months` as TransactionAmountsRequestType].startTime, 1, 'months');
            dateRange[`monthBeforeLast${i}Months` as TransactionAmountsRequestType].endTime = getUnixTimeBeforeUnixTime(dateRange[`monthBeforeLast${i - 1}Months` as TransactionAmountsRequestType].startTime, 1, 'seconds');
        }
    }

    function updateTransactionDateRange(): void {
        initTransactionDateRange(transactionDataRange.value);
    }

    function updateTransactionOverviewInvalidState(invalidState: boolean): void {
        transactionOverviewStateInvalid.value = invalidState;
        transactionUnreconciledCountsStateInvalid.value = invalidState;

        for (const dateType of keys(transactionCategoryStatisticsData.value)) {
            transactionCategoryStatisticsStateInvalid.value[parseInt(dateType)] = invalidState;
        }

        transactionAssetTrendsStateInvalid.value = invalidState;

        for (const query of keys(recentTransactions.value)) {
            recentTransactionsStateInvalid.value[query] = invalidState;
        }

        currentMonthTransactionsStateInvalid.value = invalidState;
        transactionDailyAmountsStateInvalid.value = invalidState;
    }

    function resetTransactionOverview(): void {
        updateTransactionDateRange();
        transactionOverviewOptions.value.loadedMonths = 1;
        transactionOverviewData.value = {};
        transactionOverviewStateInvalid.value = true;
        transactionUnreconciledCounts.value = [];
        transactionUnreconciledCountsStateInvalid.value = true;
        transactionUnreconciledCountsAccountIds.value = [];
        transactionCategoryStatisticsData.value = {};
        transactionCategoryStatisticsStateInvalid.value = {};
        transactionAssetTrendsData.value = [];
        transactionAssetTrendsStateInvalid.value = true;
        recentTransactions.value = {};
        recentTransactionsStateInvalid.value = {};
        currentMonthTransactions.value = [];
        currentMonthTransactionInflowOutflowDailyTotalAmounts.value = {};
        currentMonthTransactionIncomeExpenseDailyTotalAmounts.value = {};
        currentMonthTransactionsStateInvalid.value = true;
        transactionDailyAmountsData.value = [];
        transactionDailyAmountsStateInvalid.value = true;
    }

    return {
        settingsStore, accountsStore, transactionCategoriesStore, transactionsStore, transactionDataRange,
        transactionOverviewOptions, transactionOverviewData, transactionOverviewStateInvalid, transactionUnreconciledCounts,
        transactionUnreconciledCountsStateInvalid, transactionUnreconciledCountsAccountIds, transactionCategoryStatisticsData,
        transactionCategoryStatisticsStateInvalid, transactionAssetTrendsData, transactionAssetTrendsStateInvalid,
        recentTransactions, recentTransactionsStateInvalid, currentMonthTransactions,
        currentMonthTransactionInflowOutflowDailyTotalAmounts, currentMonthTransactionIncomeExpenseDailyTotalAmounts,
        currentMonthTransactionsStateInvalid, transactionDailyAmountsData, transactionDailyAmountsStateInvalid,
        transactionOverview, getRecentTransactionsQueryCacheKey, updateTransactionDateRange,
        updateTransactionOverviewInvalidState, resetTransactionOverview,
    };
}
