import { defineStore } from 'pinia';
import { useOverviewCalendar } from './overview/calendar.ts';
import { useOverviewLoading } from './overview/loading.ts';
import { useOverviewNavigation } from './overview/navigation.ts';
import { useOverviewRecentTransactions } from './overview/recent-transactions.ts';
import { useOverviewState } from './overview/state.ts';
import { useOverviewStatistics } from './overview/statistics.ts';
export * from './overview/types.ts';







export const useOverviewStore = defineStore('overview', () => {
    const {
        settingsStore, accountsStore, transactionCategoriesStore, transactionsStore, transactionDataRange,
        transactionOverviewOptions, transactionOverviewData, transactionOverviewStateInvalid, transactionUnreconciledCounts,
        transactionUnreconciledCountsStateInvalid, transactionUnreconciledCountsAccountIds, transactionCategoryStatisticsData,
        transactionCategoryStatisticsStateInvalid, transactionAssetTrendsData, transactionAssetTrendsStateInvalid,
        recentTransactions, recentTransactionsStateInvalid, currentMonthTransactions,
        currentMonthTransactionInflowOutflowDailyTotalAmounts, currentMonthTransactionIncomeExpenseDailyTotalAmounts,
        currentMonthTransactionsStateInvalid, transactionDailyAmountsData, transactionDailyAmountsStateInvalid,
        transactionOverview, getRecentTransactionsQueryCacheKey, updateTransactionDateRange,
        updateTransactionOverviewInvalidState, resetTransactionOverview,
    } = useOverviewState();
    const { loadTransactionOverview, loadTransactionUnreconciledCounts } = useOverviewLoading({
        transactionDataRange, updateTransactionDateRange, transactionOverviewOptions, transactionOverviewStateInvalid,
        transactionOverviewData, settingsStore, transactionUnreconciledCountsAccountIds,
        transactionUnreconciledCountsStateInvalid, transactionUnreconciledCounts,
    });
    const { loadTransactionCategoryStatistics, loadTransactionAssetTrends } = useOverviewStatistics({
        transactionDataRange, updateTransactionDateRange, transactionCategoryStatisticsData,
        transactionCategoryStatisticsStateInvalid, settingsStore, transactionAssetTrendsStateInvalid, transactionAssetTrendsData,
    });
    const { loadRecentTransactions } = useOverviewRecentTransactions({ getRecentTransactionsQueryCacheKey, recentTransactions, recentTransactionsStateInvalid });
    const { loadTransactionDailyAmounts, loadCurrentMonthTransactions } = useOverviewCalendar({
        transactionDataRange, updateTransactionDateRange, transactionDailyAmountsData, transactionDailyAmountsStateInvalid,
        settingsStore, currentMonthTransactions, currentMonthTransactionInflowOutflowDailyTotalAmounts,
        currentMonthTransactionIncomeExpenseDailyTotalAmounts, currentMonthTransactionsStateInvalid, accountsStore,
        transactionCategoriesStore, transactionsStore,
    });
    const { getRecentTransactions, getTransactionListPageParams } = useOverviewNavigation({ getRecentTransactionsQueryCacheKey, recentTransactions, settingsStore, transactionCategoriesStore, accountsStore });

    return {
        // states
        transactionDataRange,
        transactionOverviewOptions,
        transactionOverviewData,
        transactionOverviewStateInvalid,
        transactionUnreconciledCounts,
        transactionCategoryStatisticsData,
        transactionAssetTrendsData,
        recentTransactions,
        currentMonthTransactionInflowOutflowDailyTotalAmounts,
        currentMonthTransactionIncomeExpenseDailyTotalAmounts,
        currentMonthTransactionsStateInvalid,
        transactionDailyAmountsData,
        // computed states,
        transactionOverview,
        // functions
        updateTransactionOverviewInvalidState,
        resetTransactionOverview,
        loadTransactionOverview,
        loadTransactionUnreconciledCounts,
        loadTransactionCategoryStatistics,
        loadTransactionAssetTrends,
        loadRecentTransactions,
        loadCurrentMonthTransactions,
        loadTransactionDailyAmounts,
        getRecentTransactions,
        getTransactionListPageParams
    };
});
