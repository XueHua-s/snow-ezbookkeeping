import { defineStore } from 'pinia';
import { useStatisticsAssets } from './statistics/assets.ts';
import { useStatisticsCategorical } from './statistics/categorical.ts';
import { useStatisticsCategoryTotals } from './statistics/category-totals.ts';
import { useStatisticsFilter } from './statistics/filter.ts';
import { useStatisticsLoading } from './statistics/loading.ts';
import { useStatisticsMetadata } from './statistics/metadata.ts';
import { useStatisticsNavigation } from './statistics/navigation.ts';
import { useStatisticsOverview } from './statistics/overview.ts';
import { useStatisticsState } from './statistics/state.ts';
import { useStatisticsTrends } from './statistics/trends.ts';
export * from './statistics/types.ts';





export const useStatisticsStore = defineStore('statistics', () => {
    const {
        settingsStore, userStore, accountsStore, transactionCategoriesStore, exchangeRatesStore, transactionStatisticsFilter,
        transactionCategoryStatisticsData, transactionCategoryTrendsData, transactionAssetTrendsData,
        transactionStatisticsStateInvalid, categoricalAnalysisChartDataCategory,
    } = useStatisticsState();
    const { transactionCategoryStatisticsDataWithCategoryAndAccountInfo, assembleAccountAndCategoryInfo } = useStatisticsMetadata({ transactionCategoryStatisticsData, userStore, accountsStore, transactionCategoriesStore, exchangeRatesStore });
    const { transactionCategoryTotalAmountAnalysisData, getCategoryTotalAmountItems, sortCategoryTotalAmountItems } = useStatisticsCategoryTotals({ transactionCategoryStatisticsDataWithCategoryAndAccountInfo, transactionStatisticsFilter, settingsStore });
    const { categoricalOverviewAnalysisData } = useStatisticsOverview({ transactionCategoryStatisticsDataWithCategoryAndAccountInfo, transactionStatisticsFilter, settingsStore, accountsStore });
    const { categoricalAnalysisData } = useStatisticsCategorical({
        accountsStore, transactionStatisticsFilter, settingsStore, userStore, exchangeRatesStore,
        transactionCategoryTotalAmountAnalysisData, sortCategoryTotalAmountItems,
    });
    const { trendsAnalysisData } = useStatisticsTrends({ transactionCategoryTrendsData, assembleAccountAndCategoryInfo, getCategoryTotalAmountItems, transactionStatisticsFilter, sortCategoryTotalAmountItems });
    const { assetTrendsData } = useStatisticsAssets({ transactionAssetTrendsData, assembleAccountAndCategoryInfo, transactionStatisticsFilter, settingsStore, sortCategoryTotalAmountItems });
    const { updateTransactionStatisticsInvalidState, resetTransactionStatistics, initTransactionStatisticsFilter, updateTransactionStatisticsFilter } = useStatisticsFilter({ transactionStatisticsStateInvalid, transactionStatisticsFilter, transactionCategoryStatisticsData, transactionCategoryTrendsData, settingsStore, userStore });
    const { getTransactionStatisticsPageParams, getTransactionListPageParams } = useStatisticsNavigation({ transactionStatisticsFilter, accountsStore, transactionCategoriesStore });
    const { loadCategoricalAnalysis, loadTrendAnalysis, loadAssetTrends } = useStatisticsLoading({
        transactionStatisticsFilter, settingsStore, transactionStatisticsStateInvalid, updateTransactionStatisticsInvalidState,
        transactionCategoryStatisticsData, transactionCategoryTrendsData, transactionAssetTrendsData,
    });

    return {
        // states
        transactionStatisticsFilter,
        transactionCategoryStatisticsData,
        transactionCategoryTrendsData,
        transactionStatisticsStateInvalid,
        // computed states
        categoricalAnalysisChartDataCategory,
        categoricalOverviewAnalysisData,
        categoricalAnalysisData,
        trendsAnalysisData,
        assetTrendsData,
        // functions
        updateTransactionStatisticsInvalidState,
        resetTransactionStatistics,
        initTransactionStatisticsFilter,
        updateTransactionStatisticsFilter,
        getTransactionStatisticsPageParams,
        getTransactionListPageParams,
        loadCategoricalAnalysis,
        loadTrendAnalysis,
        loadAssetTrends
    };
});
