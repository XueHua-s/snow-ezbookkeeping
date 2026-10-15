import type { SetupProps } from '@/lib/ui/props.ts';
import { useTransactionPageChartSelection } from './useTransactionPage/chart-selection.ts';
import { useTransactionPageDateFilter } from './useTransactionPage/date-filter.ts';
import { useTransactionPageExport } from './useTransactionPage/export.ts';
import { useTransactionPageFilters } from './useTransactionPage/filters.ts';
import { useTransactionPageInitialization } from './useTransactionPage/initialization.ts';
import { useTransactionPageLifecycle } from './useTransactionPage/lifecycle.ts';
import { useTransactionPageLoading } from './useTransactionPage/loading.ts';
import { useTransactionPageState } from './useTransactionPage/state.ts';
import { type TransactionStatisticsProps } from './useTransactionPage/types.ts';
export * from './useTransactionPage/types.ts';

// Coordinates TransactionPage state and effects; called once from the component setup.
export function useTransactionPage(props: SetupProps<TransactionStatisticsProps>) {
    const {
        router, tt, formatAmountToWesternArabicNumeralsWithoutDigitGrouping,
        formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping, formatPercentToLocalizedNumerals, loading, analysisType,
        trendDateAggregationType, assetTrendsDateAggregationType, defaultCurrency, firstDayOfWeek, fiscalYearStart, allDateRanges,
        allSortingTypes, allTrendAnalysisDateAggregationTypes, allAssetTrendsDateAggregationTypes, query, queryChartDataCategory,
        queryDateType, queryStartTime, queryEndTime, queryDateRangeName, queryTrendDateAggregationTypeName,
        queryAssetTrendsDateAggregationTypeName, canChangeDateRange, canShiftDateRange, canUseCategoryFilter, canUseTagFilter,
        canUseKeywordFilter, showAmountInChart, totalAmountName, showPercentInCategoricalChart, showTotalAmountInTrendsChart,
        showStackedInTrendsChart, translateNameInTrendsChart, categoricalOverviewAnalysisData, categoricalAnalysisData,
        trendsAnalysisData, radarChartCategoryNames, radarChartData, assetTrendsData, canShowCustomDateRange,
        getTransactionCategoricalAnalysisDataItemDisplayColor, getDisplayAmount, accountsStore, transactionCategoriesStore,
        statisticsStore, snackbar, monthlyTrendsChart, dailyTrendsChart, dataExportDialog, activeTab, initing, filterKeyword,
        showCustomDateRangeDialog, showCustomMonthRangeDialog, showFilterAccountDialog, showFilterCategoryDialog,
        showFilterTagDialog, isDarkMode, statisticsDataHasData, allChartTypes, statisticsTextColor, getFilterLinkUrl,
        getTransactionItemLinkUrl,
    } = useTransactionPageState();
    const { init } = useTransactionPageInitialization({
        filterKeyword, query, analysisType, trendDateAggregationType, assetTrendsDateAggregationType, statisticsStore, loading,
        initing, accountsStore, transactionCategoriesStore, snackbar,
    });
    const { reload } = useTransactionPageLoading({ loading, query, analysisType, statisticsStore, accountsStore, snackbar });
    const {
        queryAnalysisType, queryChartType, queryChartDataType, querySortingType, isQuerySpecialChartType,
        setTrendDateAggregationType, setAssetTrendsDateAggregationType,
    } = useTransactionPageChartSelection({ analysisType, query, statisticsStore, trendDateAggregationType, assetTrendsDateAggregationType, loading, router, getFilterLinkUrl });
    const { setDateFilter, setCustomDateFilter, shiftDateRange } = useTransactionPageDateFilter({
        analysisType, showCustomDateRangeDialog, query, showCustomMonthRangeDialog, firstDayOfWeek, fiscalYearStart,
        statisticsStore, loading, router, getFilterLinkUrl,
    });
    const { setAccountFilter, setCategoryFilter, setTagFilter, setKeywordFilter } = useTransactionPageFilters({ showFilterAccountDialog, loading, statisticsStore, router, getFilterLinkUrl, showFilterCategoryDialog, showFilterTagDialog, analysisType, query });
    const { exportResults } = useTransactionPageExport({
        analysisType, categoricalAnalysisData, query, dataExportDialog, tt, defaultCurrency,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        trendsAnalysisData, monthlyTrendsChart, assetTrendsData, dailyTrendsChart,
    });
    const { onClickSankeyChartItem, onClickPieChartItem, onClickTrendChartItem, onShowDateRangeError } = useTransactionPageLifecycle({ transactionCategoriesStore, router, getTransactionItemLinkUrl, snackbar, init, props });

    return {
        router, tt, formatPercentToLocalizedNumerals, loading, trendDateAggregationType, assetTrendsDateAggregationType,
        defaultCurrency, fiscalYearStart, allDateRanges, allSortingTypes, allTrendAnalysisDateAggregationTypes,
        allAssetTrendsDateAggregationTypes, query, queryChartDataCategory, queryDateType, queryStartTime, queryEndTime,
        queryDateRangeName, queryTrendDateAggregationTypeName, queryAssetTrendsDateAggregationTypeName, canChangeDateRange,
        canShiftDateRange, canUseCategoryFilter, canUseTagFilter, canUseKeywordFilter, showAmountInChart, totalAmountName,
        showPercentInCategoricalChart, showTotalAmountInTrendsChart, showStackedInTrendsChart, translateNameInTrendsChart,
        categoricalOverviewAnalysisData, categoricalAnalysisData, trendsAnalysisData, radarChartCategoryNames, radarChartData,
        assetTrendsData, canShowCustomDateRange, getTransactionCategoricalAnalysisDataItemDisplayColor, getDisplayAmount,
        snackbar, monthlyTrendsChart, dailyTrendsChart, dataExportDialog, activeTab, initing, filterKeyword,
        showCustomDateRangeDialog, showCustomMonthRangeDialog, showFilterAccountDialog, showFilterCategoryDialog,
        showFilterTagDialog, isDarkMode, statisticsDataHasData, allChartTypes, queryAnalysisType, queryChartType,
        queryChartDataType, querySortingType, isQuerySpecialChartType, statisticsTextColor, getTransactionItemLinkUrl, init,
        reload, setTrendDateAggregationType, setAssetTrendsDateAggregationType, setDateFilter, setCustomDateFilter,
        shiftDateRange, setAccountFilter, setCategoryFilter, setTagFilter, setKeywordFilter, exportResults,
        onClickSankeyChartItem, onClickPieChartItem, onClickTrendChartItem, onShowDateRangeError,
    };
}
