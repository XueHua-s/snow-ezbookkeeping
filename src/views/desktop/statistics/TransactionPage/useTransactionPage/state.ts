import type { TypeAndDisplayName } from '@/core/base.ts';
import { type TimeRangeAndDateType } from '@/core/datetime.ts';
import {
    ChartDataType,
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import { ThemeType } from '@/core/theme.ts';
import { useI18n } from '@/locales/helpers.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useStatisticsStore } from '@/stores/statistics.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useStatisticsTransactionPageBase } from '@/views/base/statistics/StatisticsTransactionPageBase.ts';
import { computed, ref, useTemplateRef } from 'vue';
import { useRouter } from 'vue-router';
import { useTheme } from 'vuetify';
import { type DataExportDialogType, type SnackBarType, type TrendsChartType } from './types.ts';


export function useTransactionPageState() {
    const router = useRouter();

    const theme = useTheme();

    const {
        tt,
        getAllCategoricalChartTypes,
        getAllTrendChartTypes,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping,
        formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        formatPercentToLocalizedNumerals
    } = useI18n();

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
        queryTrendDateAggregationTypeName,
        queryAssetTrendsDateAggregationTypeName,
        canChangeDateRange,
        canShiftDateRange,
        canUseCategoryFilter,
        canUseTagFilter,
        canUseKeywordFilter,
        showAmountInChart,
        totalAmountName,
        showPercentInCategoricalChart,
        showTotalAmountInTrendsChart,
        showStackedInTrendsChart,
        translateNameInTrendsChart,
        categoricalOverviewAnalysisData,
        categoricalAnalysisData,
        trendsAnalysisData,
        radarChartCategoryNames,
        radarChartData,
        assetTrendsData,
        canShowCustomDateRange,
        getTransactionCategoricalAnalysisDataItemDisplayColor,
        getDisplayAmount
    } = useStatisticsTransactionPageBase();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const statisticsStore = useStatisticsStore();

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const monthlyTrendsChart = useTemplateRef<TrendsChartType>('monthlyTrendsChart');

    const dailyTrendsChart = useTemplateRef<TrendsChartType>('dailyTrendsChart');

    const dataExportDialog = useTemplateRef<DataExportDialogType>('dataExportDialog');

    const activeTab = ref<string>('statisticsPage');

    const initing = ref<boolean>(true);

    const filterKeyword = ref<string>('');

    const showCustomDateRangeDialog = ref<boolean>(false);

    const showCustomMonthRangeDialog = ref<boolean>(false);

    const showFilterAccountDialog = ref<boolean>(false);

    const showFilterCategoryDialog = ref<boolean>(false);

    const showFilterTagDialog = ref<boolean>(false);

    const isDarkMode = computed<boolean>(() => theme.global.name.value === ThemeType.Dark);

    const statisticsDataHasData = computed<boolean>(() => {
        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            return !!categoricalAnalysisData.value && !!categoricalAnalysisData.value.items && categoricalAnalysisData.value.items.length > 0;
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
            return !!trendsAnalysisData.value && !!trendsAnalysisData.value.items && trendsAnalysisData.value.items.length > 0 && !!monthlyTrendsChart.value;
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            return !!assetTrendsData.value && !!assetTrendsData.value.items && assetTrendsData.value.items.length > 0 && !!dailyTrendsChart.value;
        }

        return false;
    });

    const allChartTypes = computed<TypeAndDisplayName[]>(() => {
        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            return getAllCategoricalChartTypes(true);
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
            return getAllTrendChartTypes();
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            return getAllTrendChartTypes();
        } else {
            return [];
        }
    });

    const statisticsTextColor = computed<string>(() => {
        if (query.value.chartDataType === ChartDataType.OutflowsByAccount.type ||
            query.value.chartDataType === ChartDataType.ExpenseByAccount.type ||
            query.value.chartDataType === ChartDataType.ExpenseByPrimaryCategory.type ||
            query.value.chartDataType === ChartDataType.ExpenseBySecondaryCategory.type) {
            return 'text-expense';
        } else if (query.value.chartDataType === ChartDataType.InflowsByAccount.type ||
            query.value.chartDataType === ChartDataType.IncomeByAccount.type ||
            query.value.chartDataType === ChartDataType.IncomeByPrimaryCategory.type ||
            query.value.chartDataType === ChartDataType.IncomeBySecondaryCategory.type) {
            return 'text-income';
        } else {
            return 'text-default';
        }
    });

    function getFilterLinkUrl(): string {
        return `/statistics/transaction?${statisticsStore.getTransactionStatisticsPageParams(analysisType.value, trendDateAggregationType.value, assetTrendsDateAggregationType.value)}`;
    }

    function getTransactionItemLinkUrl(itemId: string, dateRange?: TimeRangeAndDateType): string {
        return `/transaction/list?${statisticsStore.getTransactionListPageParams(analysisType.value, itemId, dateRange)}`;
    }

    return {
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
    };
}
