import type { TypeAndDisplayName } from '@/core/base.ts';
import {
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import { TextDirection } from '@/core/text.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useI18n } from '@/locales/helpers.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useStatisticsStore } from '@/stores/statistics.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useStatisticsTransactionPageBase } from '@/views/base/statistics/StatisticsTransactionPageBase.ts';
import { computed, ref } from 'vue';


export function useTransactionPageState() {
    const {
        tt,
        getCurrentLanguageTextDirection,
        getAllCategoricalChartTypes,
        formatPercentToLocalizedNumerals
    } = useI18n();

    const { showPrompt, showToast, routeBackOnError } = useI18nUIComponents();

    const {
        loading,
        analysisType,
        trendDateAggregationType,
        assetTrendsDateAggregationType,
        defaultCurrency,
        firstDayOfWeek,
        fiscalYearStart,
        allDateRanges,
        allSortingTypes,
        allTrendAnalysisDateAggregationTypes,
        allAssetTrendsDateAggregationTypes,
        query,
        queryChartDataCategory,
        queryDateType,
        queryStartTime,
        queryEndTime,
        queryDateRangeName,
        queryChartDataTypeName,
        querySortingTypeName,
        queryTrendDateAggregationTypeName,
        queryAssetTrendsDateAggregationTypeName,
        isQueryDateRangeChanged,
        canChangeDateRange,
        canShiftDateRange,
        canUseCategoryFilter,
        canUseTagFilter,
        canUseKeywordFilter,
        showAmountInChart,
        totalAmountName,
        showPercentInCategoricalChart,
        showStackedInTrendsChart,
        translateNameInTrendsChart,
        categoricalAnalysisData,
        trendsAnalysisData,
        assetTrendsData,
        canShowCustomDateRange,
        getTransactionCategoricalAnalysisDataItemDisplayColor,
        getDisplayAmount
    } = useStatisticsTransactionPageBase();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const statisticsStore = useStatisticsStore();

    const loadingError = ref<unknown | null>(null);

    const reloading = ref<boolean>(false);

    const showCustomDateRangeSheet = ref<boolean>(false);

    const showCustomMonthRangeSheet = ref<boolean>(false);

    const showMoreActionSheet = ref<boolean>(false);

    const textDirection = computed<TextDirection>(() => getCurrentLanguageTextDirection());

    const allChartTypes = computed<TypeAndDisplayName[]>(() => {
        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            return getAllCategoricalChartTypes();
        } else {
            return [];
        }
    });

    return {
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
    };
}
