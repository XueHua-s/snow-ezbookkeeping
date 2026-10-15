import type { SetupProps } from '@/lib/ui/props.ts';
import type { Router } from 'framework7/types';
import { useTransactionPageFilters } from './useTransactionPage/filters.ts';
import { useTransactionPageLoading } from './useTransactionPage/loading.ts';
import { useTransactionPageNavigation } from './useTransactionPage/navigation.ts';
import { useTransactionPageState } from './useTransactionPage/state.ts';



// Coordinates TransactionPage state and effects; called once from the component setup.
export function useTransactionPage(props: SetupProps<{
    f7router: Router.Router;
}>) {
    const {
        tt, formatPercentToLocalizedNumerals, showPrompt, showToast, routeBackOnError, loading, analysisType,
        trendDateAggregationType, assetTrendsDateAggregationType, defaultCurrency, firstDayOfWeek, fiscalYearStart, allDateRanges,
        allSortingTypes, allTrendAnalysisDateAggregationTypes, allAssetTrendsDateAggregationTypes, query, queryChartDataCategory,
        queryDateType, queryStartTime, queryEndTime, queryDateRangeName, queryChartDataTypeName, querySortingTypeName,
        queryTrendDateAggregationTypeName, queryAssetTrendsDateAggregationTypeName, isQueryDateRangeChanged, canChangeDateRange,
        canShiftDateRange, canUseCategoryFilter, canUseTagFilter, canUseKeywordFilter, showAmountInChart, totalAmountName,
        showPercentInCategoricalChart, showStackedInTrendsChart, translateNameInTrendsChart, categoricalAnalysisData,
        trendsAnalysisData, assetTrendsData, canShowCustomDateRange, getTransactionCategoricalAnalysisDataItemDisplayColor,
        getDisplayAmount, accountsStore, transactionCategoriesStore, statisticsStore, loadingError, reloading,
        showCustomDateRangeSheet, showCustomMonthRangeSheet, showMoreActionSheet, textDirection, allChartTypes,
    } = useTransactionPageState();
    const { init, reload } = useTransactionPageLoading({ statisticsStore, analysisType, accountsStore, transactionCategoriesStore, loading, loadingError, showToast, reloading, query });
    const {
        queryChartType, setChartType, setChartDataType, setSortingType, setTrendDateAggregationType,
        setAssetTrendsDateAggregationType, setDateFilter, setCustomDateFilter, shiftDateRange,
    } = useTransactionPageFilters({
        analysisType, query, statisticsStore, reload, trendDateAggregationType, assetTrendsDateAggregationType,
        showCustomDateRangeSheet, showCustomMonthRangeSheet, firstDayOfWeek, fiscalYearStart,
    });
    const {
        getTransactionItemLinkUrl, filterAccounts, filterCategories, filterTags, filterDescription, settings,
        scrollPopoverToSelectedItem, onClickPieChartItem, onClickTrendChartItem, onPageAfterIn,
    } = useTransactionPageNavigation({ statisticsStore, analysisType, props, showPrompt, query, reload, loading, routeBackOnError, loadingError, init });

    return {
        tt, formatPercentToLocalizedNumerals, loading, analysisType, trendDateAggregationType, assetTrendsDateAggregationType,
        defaultCurrency, fiscalYearStart, allDateRanges, allSortingTypes, allTrendAnalysisDateAggregationTypes,
        allAssetTrendsDateAggregationTypes, query, queryChartDataCategory, queryDateType, queryStartTime, queryEndTime,
        queryDateRangeName, queryChartDataTypeName, querySortingTypeName, queryTrendDateAggregationTypeName,
        queryAssetTrendsDateAggregationTypeName, isQueryDateRangeChanged, canChangeDateRange, canShiftDateRange,
        canUseCategoryFilter, canUseTagFilter, canUseKeywordFilter, showAmountInChart, totalAmountName,
        showPercentInCategoricalChart, showStackedInTrendsChart, translateNameInTrendsChart, categoricalAnalysisData,
        trendsAnalysisData, assetTrendsData, canShowCustomDateRange, getTransactionCategoricalAnalysisDataItemDisplayColor,
        getDisplayAmount, reloading, showCustomDateRangeSheet, showCustomMonthRangeSheet, showMoreActionSheet, textDirection,
        allChartTypes, queryChartType, getTransactionItemLinkUrl, reload, setChartType, setChartDataType, setSortingType,
        setTrendDateAggregationType, setAssetTrendsDateAggregationType, setDateFilter, setCustomDateFilter, shiftDateRange,
        filterAccounts, filterCategories, filterTags, filterDescription, settings, scrollPopoverToSelectedItem,
        onClickPieChartItem, onClickTrendChartItem, onPageAfterIn,
    };
}
