import { type TextualYearMonth, DateRange, DateRangeScene } from '@/core/datetime.ts';
import {
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import {
    isNumber,
    isString
} from '@/lib/common.ts';
import {
    getDateRangeByDateType,
    getDateTypeByDateRange,
    getGregorianCalendarYearAndMonthFromUnixTime,
    getShiftedDateRangeAndDateType,
    getYearMonthFirstUnixTime,
    getYearMonthLastUnixTime
} from '@/lib/datetime.ts';
import type { useTransactionPageState } from './state.ts';

interface Context {
    analysisType: ReturnType<typeof useTransactionPageState>['analysisType'];
    showCustomDateRangeDialog: ReturnType<typeof useTransactionPageState>['showCustomDateRangeDialog'];
    query: ReturnType<typeof useTransactionPageState>['query'];
    showCustomMonthRangeDialog: ReturnType<typeof useTransactionPageState>['showCustomMonthRangeDialog'];
    firstDayOfWeek: ReturnType<typeof useTransactionPageState>['firstDayOfWeek'];
    fiscalYearStart: ReturnType<typeof useTransactionPageState>['fiscalYearStart'];
    statisticsStore: ReturnType<typeof useTransactionPageState>['statisticsStore'];
    loading: ReturnType<typeof useTransactionPageState>['loading'];
    router: ReturnType<typeof useTransactionPageState>['router'];
    getFilterLinkUrl: ReturnType<typeof useTransactionPageState>['getFilterLinkUrl'];
}

export function useTransactionPageDateFilter(context: Context) {
    const {
        analysisType, showCustomDateRangeDialog, query, showCustomMonthRangeDialog, firstDayOfWeek, fiscalYearStart,
        statisticsStore, loading, router, getFilterLinkUrl,
    } = context;

    function setDateFilter(dateType: number): void {
        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            if (dateType === DateRange.Custom.type) { // Custom
                showCustomDateRangeDialog.value = true;
                return;
            } else if (query.value.categoricalChartDateType === dateType) {
                return;
            }
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
            if (dateType === DateRange.Custom.type) { // Custom
                showCustomMonthRangeDialog.value = true;
                return;
            } else if (query.value.trendChartDateType === dateType) {
                return;
            }
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            if (dateType === DateRange.Custom.type) { // Custom
                showCustomDateRangeDialog.value = true;
                return;
            } else if (query.value.assetTrendsChartDateType === dateType) {
                return;
            }
        }

        const dateRange = getDateRangeByDateType(dateType, firstDayOfWeek.value, fiscalYearStart.value);

        if (!dateRange) {
            return;
        }

        let changed = false;

        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            changed = statisticsStore.updateTransactionStatisticsFilter({
                categoricalChartDateType: dateRange.dateType,
                categoricalChartStartTime: dateRange.minTime,
                categoricalChartEndTime: dateRange.maxTime
            });
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
            changed = statisticsStore.updateTransactionStatisticsFilter({
                trendChartDateType: dateRange.dateType,
                trendChartStartYearMonth: getGregorianCalendarYearAndMonthFromUnixTime(dateRange.minTime),
                trendChartEndYearMonth: getGregorianCalendarYearAndMonthFromUnixTime(dateRange.maxTime)
            });
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            changed = statisticsStore.updateTransactionStatisticsFilter({
                assetTrendsChartDateType: dateRange.dateType,
                assetTrendsChartStartTime: dateRange.minTime,
                assetTrendsChartEndTime: dateRange.maxTime
            });
        }

        if (changed) {
            loading.value = true;
            statisticsStore.updateTransactionStatisticsInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    function setCustomDateFilter(startTime: number | TextualYearMonth, endTime: number | TextualYearMonth): void {
        if (!startTime || !endTime) {
            return;
        }

        let changed = false;

        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis && isNumber(startTime) && isNumber(endTime)) {
            const chartDateType = getDateTypeByDateRange(startTime, endTime, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal);

            changed = statisticsStore.updateTransactionStatisticsFilter({
                categoricalChartDateType: chartDateType,
                categoricalChartStartTime: startTime,
                categoricalChartEndTime: endTime
            });

            showCustomDateRangeDialog.value = false;
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis && isString(startTime) && isString(endTime)) {
            const chartDateType = getDateTypeByDateRange(getYearMonthFirstUnixTime(startTime), getYearMonthLastUnixTime(endTime), firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.TrendAnalysis);

            changed = statisticsStore.updateTransactionStatisticsFilter({
                trendChartDateType: chartDateType,
                trendChartStartYearMonth: startTime,
                trendChartEndYearMonth: endTime
            });

            showCustomMonthRangeDialog.value = false;
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends && isNumber(startTime) && isNumber(endTime)) {
            const chartDateType = getDateTypeByDateRange(startTime, endTime, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.AssetTrends);

            changed = statisticsStore.updateTransactionStatisticsFilter({
                assetTrendsChartDateType: chartDateType,
                assetTrendsChartStartTime: startTime,
                assetTrendsChartEndTime: endTime
            });

            showCustomDateRangeDialog.value = false;
        }

        if (changed) {
            loading.value = true;
            statisticsStore.updateTransactionStatisticsInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    function shiftDateRange(scale: number): void {
        let changed = false;

        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            if (query.value.categoricalChartDateType === DateRange.All.type) {
                return;
            }

            const newDateRange = getShiftedDateRangeAndDateType(query.value.categoricalChartStartTime, query.value.categoricalChartEndTime, scale, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal);

            changed = statisticsStore.updateTransactionStatisticsFilter({
                categoricalChartDateType: newDateRange.dateType,
                categoricalChartStartTime: newDateRange.minTime,
                categoricalChartEndTime: newDateRange.maxTime
            });
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
            const newDateRange = getShiftedDateRangeAndDateType(getYearMonthFirstUnixTime(query.value.trendChartStartYearMonth), getYearMonthLastUnixTime(query.value.trendChartEndYearMonth), scale, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.TrendAnalysis);

            changed = statisticsStore.updateTransactionStatisticsFilter({
                trendChartDateType: newDateRange.dateType,
                trendChartStartYearMonth: getGregorianCalendarYearAndMonthFromUnixTime(newDateRange.minTime),
                trendChartEndYearMonth: getGregorianCalendarYearAndMonthFromUnixTime(newDateRange.maxTime)
            });
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            if (query.value.assetTrendsChartDateType === DateRange.All.type) {
                return;
            }

            const newDateRange = getShiftedDateRangeAndDateType(query.value.assetTrendsChartStartTime, query.value.assetTrendsChartEndTime, scale, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.AssetTrends);

            changed = statisticsStore.updateTransactionStatisticsFilter({
                assetTrendsChartDateType: newDateRange.dateType,
                assetTrendsChartStartTime: newDateRange.minTime,
                assetTrendsChartEndTime: newDateRange.maxTime
            });
        }

        if (changed) {
            loading.value = true;
            statisticsStore.updateTransactionStatisticsInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    return { setDateFilter, setCustomDateFilter, shiftDateRange };
}
