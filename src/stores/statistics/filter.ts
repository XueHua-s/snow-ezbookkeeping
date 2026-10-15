import { DateRange, DateRangeScene } from '@/core/datetime.ts';
import {
    CategoricalChartType,
    ChartDataType,
    ChartSortingType,
    DEFAULT_ASSET_TRENDS_CHART_DATA_RANGE,
    DEFAULT_CATEGORICAL_CHART_DATA_RANGE,
    DEFAULT_TREND_CHART_DATA_RANGE,
    StatisticsAnalysisType,
    TrendChartType
} from '@/core/statistics.ts';
import { KeywordMatchMode } from '@/core/text.ts';
import {
    isEquals,
    isInteger,
    isNumber,
    isObject,
    isString,
    isYearMonth,
    isYearMonthEquals
} from '@/lib/common.ts';
import {
    getDateRangeByDateType,
    getGregorianCalendarYearAndMonthFromUnixTime
} from '@/lib/datetime.ts';
import type { useStatisticsState } from './state.ts';
import type { TransactionStatisticsPartialFilter } from './types.ts';

interface Context {
    transactionStatisticsStateInvalid: ReturnType<typeof useStatisticsState>['transactionStatisticsStateInvalid'];
    transactionStatisticsFilter: ReturnType<typeof useStatisticsState>['transactionStatisticsFilter'];
    transactionCategoryStatisticsData: ReturnType<typeof useStatisticsState>['transactionCategoryStatisticsData'];
    transactionCategoryTrendsData: ReturnType<typeof useStatisticsState>['transactionCategoryTrendsData'];
    settingsStore: ReturnType<typeof useStatisticsState>['settingsStore'];
    userStore: ReturnType<typeof useStatisticsState>['userStore'];
}

// Owns filter for the statistics store.
export function useStatisticsFilter(context: Context) {
    const { transactionStatisticsStateInvalid, transactionStatisticsFilter, transactionCategoryStatisticsData, transactionCategoryTrendsData, settingsStore, userStore } = context;

    function updateTransactionStatisticsInvalidState(invalidState: boolean): void {
        transactionStatisticsStateInvalid.value = invalidState;
    }

    function resetTransactionStatistics(): void {
        transactionStatisticsFilter.value.chartDataType = ChartDataType.Default.type;
        transactionStatisticsFilter.value.categoricalChartType = CategoricalChartType.Default.type;
        transactionStatisticsFilter.value.categoricalChartDateType = DEFAULT_CATEGORICAL_CHART_DATA_RANGE.type;
        transactionStatisticsFilter.value.categoricalChartStartTime = 0;
        transactionStatisticsFilter.value.categoricalChartEndTime = 0;
        transactionStatisticsFilter.value.trendChartType = TrendChartType.Default.type;
        transactionStatisticsFilter.value.trendChartDateType = DEFAULT_TREND_CHART_DATA_RANGE.type;
        transactionStatisticsFilter.value.trendChartStartYearMonth = '';
        transactionStatisticsFilter.value.trendChartEndYearMonth = '';
        transactionStatisticsFilter.value.assetTrendsChartType = TrendChartType.Default.type;
        transactionStatisticsFilter.value.assetTrendsChartDateType = DEFAULT_ASSET_TRENDS_CHART_DATA_RANGE.type;
        transactionStatisticsFilter.value.assetTrendsChartStartTime = 0;
        transactionStatisticsFilter.value.assetTrendsChartEndTime = 0;
        transactionStatisticsFilter.value.filterAccountIds = {};
        transactionStatisticsFilter.value.filterCategoryIds = {};
        transactionStatisticsFilter.value.tagFilter = '';
        transactionStatisticsFilter.value.keyword = '';
        transactionStatisticsFilter.value.matchMode = KeywordMatchMode.Default.type;
        transactionCategoryStatisticsData.value = null;
        transactionCategoryTrendsData.value = [];
        transactionStatisticsStateInvalid.value = true;
    }

    function initTransactionStatisticsFilter(analysisType: StatisticsAnalysisType, filter?: TransactionStatisticsPartialFilter): void {
        if (filter && isInteger(filter.chartDataType)) {
            transactionStatisticsFilter.value.chartDataType = filter.chartDataType;
        } else {
            transactionStatisticsFilter.value.chartDataType = settingsStore.appSettings.statistics.defaultChartDataType;
        }

        if (analysisType === StatisticsAnalysisType.CategoricalAnalysis || analysisType === StatisticsAnalysisType.TrendAnalysis) {
            if (!ChartDataType.isAvailableForAnalysisType(transactionStatisticsFilter.value.chartDataType, analysisType)) {
                transactionStatisticsFilter.value.chartDataType = ChartDataType.Default.type;
            }
        } else if (analysisType === StatisticsAnalysisType.AssetTrends) {
            if (!ChartDataType.isAvailableForAnalysisType(transactionStatisticsFilter.value.chartDataType, analysisType)) {
                transactionStatisticsFilter.value.chartDataType = ChartDataType.DefaultForAssetTrends.type;
            }
        }

        // Categorical Analysis filter initialization
        if (filter && isInteger(filter.categoricalChartType)) {
            transactionStatisticsFilter.value.categoricalChartType = filter.categoricalChartType;
        } else {
            transactionStatisticsFilter.value.categoricalChartType = settingsStore.appSettings.statistics.defaultCategoricalChartType;
        }

        if (!CategoricalChartType.isValidType(transactionStatisticsFilter.value.categoricalChartType)) {
            transactionStatisticsFilter.value.categoricalChartType = CategoricalChartType.Default.type;
        }

        if (filter && isInteger(filter.categoricalChartDateType)) {
            transactionStatisticsFilter.value.categoricalChartDateType = filter.categoricalChartDateType;
        } else {
            transactionStatisticsFilter.value.categoricalChartDateType = settingsStore.appSettings.statistics.defaultCategoricalChartDataRangeType;
        }

        let categoricalChartDateTypeValid = true;

        if (!DateRange.isAvailableForScene(transactionStatisticsFilter.value.categoricalChartDateType, DateRangeScene.Normal)) {
            transactionStatisticsFilter.value.categoricalChartDateType = DEFAULT_CATEGORICAL_CHART_DATA_RANGE.type;
            categoricalChartDateTypeValid = false;
        }

        if (categoricalChartDateTypeValid && transactionStatisticsFilter.value.categoricalChartDateType === DateRange.Custom.type) {
            if (filter && isInteger(filter.categoricalChartStartTime)) {
                transactionStatisticsFilter.value.categoricalChartStartTime = filter.categoricalChartStartTime;
            } else {
                transactionStatisticsFilter.value.categoricalChartStartTime = 0;
            }

            if (filter && isInteger(filter.categoricalChartEndTime)) {
                transactionStatisticsFilter.value.categoricalChartEndTime = filter.categoricalChartEndTime;
            } else {
                transactionStatisticsFilter.value.categoricalChartEndTime = 0;
            }
        } else {
            const categoricalChartDateRange = getDateRangeByDateType(transactionStatisticsFilter.value.categoricalChartDateType, userStore.currentUserFirstDayOfWeek, userStore.currentUserFiscalYearStart);

            if (categoricalChartDateRange) {
                transactionStatisticsFilter.value.categoricalChartDateType = categoricalChartDateRange.dateType;
                transactionStatisticsFilter.value.categoricalChartStartTime = categoricalChartDateRange.minTime;
                transactionStatisticsFilter.value.categoricalChartEndTime = categoricalChartDateRange.maxTime;
            }
        }

        // Trend Analysis filter initialization
        if (filter && isInteger(filter.trendChartType)) {
            transactionStatisticsFilter.value.trendChartType = filter.trendChartType;
        } else {
            transactionStatisticsFilter.value.trendChartType = settingsStore.appSettings.statistics.defaultTrendChartType;
        }

        if (!TrendChartType.isValidType(transactionStatisticsFilter.value.trendChartType)) {
            transactionStatisticsFilter.value.trendChartType = TrendChartType.Default.type;
        }

        if (filter && isInteger(filter.trendChartDateType)) {
            transactionStatisticsFilter.value.trendChartDateType = filter.trendChartDateType;
        } else {
            transactionStatisticsFilter.value.trendChartDateType = settingsStore.appSettings.statistics.defaultTrendChartDataRangeType;
        }

        let trendChartDateTypeValid = true;

        if (!DateRange.isAvailableForScene(transactionStatisticsFilter.value.trendChartDateType, DateRangeScene.TrendAnalysis)) {
            transactionStatisticsFilter.value.trendChartDateType = DEFAULT_TREND_CHART_DATA_RANGE.type;
            trendChartDateTypeValid = false;
        }

        if (trendChartDateTypeValid && transactionStatisticsFilter.value.trendChartDateType === DateRange.Custom.type) {
            if (filter && isYearMonth(filter.trendChartStartYearMonth)) {
                transactionStatisticsFilter.value.trendChartStartYearMonth = filter.trendChartStartYearMonth;
            } else {
                transactionStatisticsFilter.value.trendChartStartYearMonth = '';
            }

            if (filter && isYearMonth(filter.trendChartEndYearMonth)) {
                transactionStatisticsFilter.value.trendChartEndYearMonth = filter.trendChartEndYearMonth;
            } else {
                transactionStatisticsFilter.value.trendChartEndYearMonth = '';
            }
        } else {
            const trendChartDateRange = getDateRangeByDateType(transactionStatisticsFilter.value.trendChartDateType, userStore.currentUserFirstDayOfWeek, userStore.currentUserFiscalYearStart);

            if (trendChartDateRange) {
                transactionStatisticsFilter.value.trendChartDateType = trendChartDateRange.dateType;
                transactionStatisticsFilter.value.trendChartStartYearMonth = getGregorianCalendarYearAndMonthFromUnixTime(trendChartDateRange.minTime);
                transactionStatisticsFilter.value.trendChartEndYearMonth = getGregorianCalendarYearAndMonthFromUnixTime(trendChartDateRange.maxTime);
            }
        }

        // Asset Trends filter initialization
        if (filter && isInteger(filter.assetTrendsChartType)) {
            transactionStatisticsFilter.value.assetTrendsChartType = filter.assetTrendsChartType;
        } else {
            transactionStatisticsFilter.value.assetTrendsChartType = settingsStore.appSettings.statistics.defaultAssetTrendsChartType;
        }

        if (!TrendChartType.isValidType(transactionStatisticsFilter.value.assetTrendsChartType)) {
            transactionStatisticsFilter.value.assetTrendsChartType = TrendChartType.Default.type;
        }

        if (filter && isInteger(filter.assetTrendsChartDateType)) {
            transactionStatisticsFilter.value.assetTrendsChartDateType = filter.assetTrendsChartDateType;
        } else {
            transactionStatisticsFilter.value.assetTrendsChartDateType = settingsStore.appSettings.statistics.defaultAssetTrendsChartDataRangeType;
        }

        let assetTrendsChartDateTypeValid = true;

        if (!DateRange.isAvailableForScene(transactionStatisticsFilter.value.assetTrendsChartDateType, DateRangeScene.AssetTrends)) {
            transactionStatisticsFilter.value.assetTrendsChartDateType = DEFAULT_ASSET_TRENDS_CHART_DATA_RANGE.type;
            assetTrendsChartDateTypeValid = false;
        }

        if (assetTrendsChartDateTypeValid && transactionStatisticsFilter.value.assetTrendsChartDateType === DateRange.Custom.type) {
            if (filter && isInteger(filter.assetTrendsChartStartTime)) {
                transactionStatisticsFilter.value.assetTrendsChartStartTime = filter.assetTrendsChartStartTime;
            } else {
                transactionStatisticsFilter.value.assetTrendsChartStartTime = 0;
            }

            if (filter && isInteger(filter.assetTrendsChartEndTime)) {
                transactionStatisticsFilter.value.assetTrendsChartEndTime = filter.assetTrendsChartEndTime;
            } else {
                transactionStatisticsFilter.value.assetTrendsChartEndTime = 0;
            }
        } else {
            const assetTrendsChartDateRange = getDateRangeByDateType(transactionStatisticsFilter.value.assetTrendsChartDateType, userStore.currentUserFirstDayOfWeek, userStore.currentUserFiscalYearStart);

            if (assetTrendsChartDateRange) {
                transactionStatisticsFilter.value.assetTrendsChartDateType = assetTrendsChartDateRange.dateType;
                transactionStatisticsFilter.value.assetTrendsChartStartTime = assetTrendsChartDateRange.minTime;
                transactionStatisticsFilter.value.assetTrendsChartEndTime = assetTrendsChartDateRange.maxTime;
            }
        }

        // Other filter initialization
        if (filter && isObject(filter.filterAccountIds)) {
            transactionStatisticsFilter.value.filterAccountIds = filter.filterAccountIds;
        } else {
            transactionStatisticsFilter.value.filterAccountIds = settingsStore.appSettings.statistics.defaultAccountFilter || {};
        }

        if (filter && isObject(filter.filterCategoryIds)) {
            transactionStatisticsFilter.value.filterCategoryIds = filter.filterCategoryIds;
        } else {
            transactionStatisticsFilter.value.filterCategoryIds = settingsStore.appSettings.statistics.defaultTransactionCategoryFilter || {};
        }

        if (filter && isString(filter.tagFilter)) {
            transactionStatisticsFilter.value.tagFilter = filter.tagFilter;
        } else {
            transactionStatisticsFilter.value.tagFilter = '';
        }

        if (filter && isString(filter.keyword)) {
            transactionStatisticsFilter.value.keyword = filter.keyword;
        } else {
            transactionStatisticsFilter.value.keyword = '';
        }

        if (filter && isNumber(filter.matchMode)) {
            transactionStatisticsFilter.value.matchMode = filter.matchMode;
        } else {
            transactionStatisticsFilter.value.matchMode = settingsStore.appSettings.statistics.defaultKeywordMatchMode;
        }

        if (filter && isInteger(filter.sortingType)) {
            transactionStatisticsFilter.value.sortingType = filter.sortingType;
        } else {
            transactionStatisticsFilter.value.sortingType = settingsStore.appSettings.statistics.defaultSortingType;
        }

        if (transactionStatisticsFilter.value.sortingType < ChartSortingType.Amount.type || transactionStatisticsFilter.value.sortingType > ChartSortingType.Name.type) {
            transactionStatisticsFilter.value.sortingType = ChartSortingType.Default.type;
        }
    }

    function updateTransactionStatisticsFilter(filter: TransactionStatisticsPartialFilter): boolean {
        let changed = false;

        if (filter && isInteger(filter.chartDataType) && transactionStatisticsFilter.value.chartDataType !== filter.chartDataType) {
            transactionStatisticsFilter.value.chartDataType = filter.chartDataType;
            changed = true;
        }

        // Categorical Analysis filter update
        if (filter && isInteger(filter.categoricalChartType) && transactionStatisticsFilter.value.categoricalChartType !== filter.categoricalChartType) {
            transactionStatisticsFilter.value.categoricalChartType = filter.categoricalChartType;
            changed = true;
        }

        if (filter && isInteger(filter.categoricalChartDateType) && transactionStatisticsFilter.value.categoricalChartDateType !== filter.categoricalChartDateType) {
            transactionStatisticsFilter.value.categoricalChartDateType = filter.categoricalChartDateType;
            changed = true;
        }

        if (filter && isInteger(filter.categoricalChartStartTime) && transactionStatisticsFilter.value.categoricalChartStartTime !== filter.categoricalChartStartTime) {
            transactionStatisticsFilter.value.categoricalChartStartTime = filter.categoricalChartStartTime;
            changed = true;
        }

        if (filter && isInteger(filter.categoricalChartEndTime) && transactionStatisticsFilter.value.categoricalChartEndTime !== filter.categoricalChartEndTime) {
            transactionStatisticsFilter.value.categoricalChartEndTime = filter.categoricalChartEndTime;
            changed = true;
        }

        // Trend Analysis filter update
        if (filter && isInteger(filter.trendChartType) && transactionStatisticsFilter.value.trendChartType !== filter.trendChartType) {
            transactionStatisticsFilter.value.trendChartType = filter.trendChartType;
            changed = true;
        }

        if (filter && isInteger(filter.trendChartDateType) && transactionStatisticsFilter.value.trendChartDateType !== filter.trendChartDateType) {
            transactionStatisticsFilter.value.trendChartDateType = filter.trendChartDateType;
            changed = true;
        }

        if (filter && (isYearMonth(filter.trendChartStartYearMonth) || filter.trendChartStartYearMonth === '') && !isYearMonthEquals(transactionStatisticsFilter.value.trendChartStartYearMonth, filter.trendChartStartYearMonth)) {
            transactionStatisticsFilter.value.trendChartStartYearMonth = filter.trendChartStartYearMonth;
            changed = true;
        }

        if (filter && (isYearMonth(filter.trendChartEndYearMonth) || filter.trendChartEndYearMonth === '') && !isYearMonthEquals(transactionStatisticsFilter.value.trendChartEndYearMonth, filter.trendChartEndYearMonth)) {
            transactionStatisticsFilter.value.trendChartEndYearMonth = filter.trendChartEndYearMonth;
            changed = true;
        }

        // Asset Trends filter update
        if (filter && isInteger(filter.assetTrendsChartType) && transactionStatisticsFilter.value.assetTrendsChartType !== filter.assetTrendsChartType) {
            transactionStatisticsFilter.value.assetTrendsChartType = filter.assetTrendsChartType;
            changed = true;
        }

        if (filter && isInteger(filter.assetTrendsChartDateType) && transactionStatisticsFilter.value.assetTrendsChartDateType !== filter.assetTrendsChartDateType) {
            transactionStatisticsFilter.value.assetTrendsChartDateType = filter.assetTrendsChartDateType;
            changed = true;
        }

        if (filter && isInteger(filter.assetTrendsChartStartTime) && transactionStatisticsFilter.value.assetTrendsChartStartTime !== filter.assetTrendsChartStartTime) {
            transactionStatisticsFilter.value.assetTrendsChartStartTime = filter.assetTrendsChartStartTime;
            changed = true;
        }

        if (filter && isInteger(filter.assetTrendsChartEndTime) && transactionStatisticsFilter.value.assetTrendsChartEndTime !== filter.assetTrendsChartEndTime) {
            transactionStatisticsFilter.value.assetTrendsChartEndTime = filter.assetTrendsChartEndTime;
            changed = true;
        }

        // Other filter update
        if (filter && isObject(filter.filterAccountIds) && !isEquals(transactionStatisticsFilter.value.filterAccountIds, filter.filterAccountIds)) {
            transactionStatisticsFilter.value.filterAccountIds = filter.filterAccountIds;
            changed = true;
        }

        if (filter && isObject(filter.filterCategoryIds) && !isEquals(transactionStatisticsFilter.value.filterCategoryIds, filter.filterCategoryIds)) {
            transactionStatisticsFilter.value.filterCategoryIds = filter.filterCategoryIds;
            changed = true;
        }

        if (filter && isString(filter.tagFilter) && transactionStatisticsFilter.value.tagFilter !== filter.tagFilter) {
            transactionStatisticsFilter.value.tagFilter = filter.tagFilter;
            changed = true;
        }

        if (filter && isString(filter.keyword) && transactionStatisticsFilter.value.keyword !== filter.keyword) {
            transactionStatisticsFilter.value.keyword = filter.keyword;
            changed = true;
        }

        if (filter && isNumber(filter.matchMode) && transactionStatisticsFilter.value.matchMode !== filter.matchMode) {
            transactionStatisticsFilter.value.matchMode = filter.matchMode;
            changed = true;
        }

        if (filter && isInteger(filter.sortingType) && transactionStatisticsFilter.value.sortingType !== filter.sortingType) {
            transactionStatisticsFilter.value.sortingType = filter.sortingType;
            changed = true;
        }

        return changed;
    }

    return { updateTransactionStatisticsInvalidState, resetTransactionStatistics, initTransactionStatisticsFilter, updateTransactionStatisticsFilter };
}
