import { type TextualYearMonth, DateRange, DateRangeScene } from '@/core/datetime.ts';
import {
    ChartDataType,
    ChartSortingType,
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import { isNumber, isString } from '@/lib/common.ts';
import {
    getDateRangeByDateType,
    getDateTypeByDateRange,
    getGregorianCalendarYearAndMonthFromUnixTime,
    getShiftedDateRangeAndDateType,
    getYearMonthFirstUnixTime,
    getYearMonthLastUnixTime
} from '@/lib/datetime.ts';
import { computed } from 'vue';
import type { useTransactionPageLoading } from './loading.ts';
import type { useTransactionPageState } from './state.ts';

interface Context {
    analysisType: ReturnType<typeof useTransactionPageState>['analysisType'];
    query: ReturnType<typeof useTransactionPageState>['query'];
    statisticsStore: ReturnType<typeof useTransactionPageState>['statisticsStore'];
    reload: ReturnType<typeof useTransactionPageLoading>['reload'];
    trendDateAggregationType: ReturnType<typeof useTransactionPageState>['trendDateAggregationType'];
    assetTrendsDateAggregationType: ReturnType<typeof useTransactionPageState>['assetTrendsDateAggregationType'];
    showCustomDateRangeSheet: ReturnType<typeof useTransactionPageState>['showCustomDateRangeSheet'];
    showCustomMonthRangeSheet: ReturnType<typeof useTransactionPageState>['showCustomMonthRangeSheet'];
    firstDayOfWeek: ReturnType<typeof useTransactionPageState>['firstDayOfWeek'];
    fiscalYearStart: ReturnType<typeof useTransactionPageState>['fiscalYearStart'];
}

export function useTransactionPageFilters(context: Context) {
    const {
        analysisType, query, statisticsStore, reload, trendDateAggregationType, assetTrendsDateAggregationType,
        showCustomDateRangeSheet, showCustomMonthRangeSheet, firstDayOfWeek, fiscalYearStart,
    } = context;

    const queryChartType = computed<number | undefined>({
        get: () => {
            if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
                return query.value.categoricalChartType;
            } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
                return query.value.trendChartType;
            } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
                return query.value.assetTrendsChartType;
            } else {
                return undefined;
            }
        },
        set: (value: number | undefined) => {
            setChartType(value);
        }
    });

    function setChartType(type?: number): void {
        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            statisticsStore.updateTransactionStatisticsFilter({
                categoricalChartType: type
            });
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
            statisticsStore.updateTransactionStatisticsFilter({
                trendChartType: type
            });
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            statisticsStore.updateTransactionStatisticsFilter({
                assetTrendsChartType: type
            });
        }
    }

    function setChartDataType(type: number, chartDataType: number): void {
        let analysisTypeChanged = false;

        if (analysisType.value !== type) {
            if (!ChartDataType.isAvailableForAnalysisType(query.value.chartDataType, type)) {
                statisticsStore.updateTransactionStatisticsFilter({
                    chartDataType: ChartDataType.Default.type
                });
            }

            analysisType.value = type;
            statisticsStore.updateTransactionStatisticsInvalidState(true);
            analysisTypeChanged = true;
        }

        statisticsStore.updateTransactionStatisticsFilter({
            chartDataType: chartDataType
        });

        if (analysisTypeChanged) {
            reload();
        }
    }

    function setSortingType(type: number): void {
        if (type < ChartSortingType.Amount.type || type > ChartSortingType.Name.type) {
            return;
        }

        statisticsStore.updateTransactionStatisticsFilter({
            sortingType: type
        });
    }

    function setTrendDateAggregationType(type: number): void {
        trendDateAggregationType.value = type;
    }

    function setAssetTrendsDateAggregationType(type: number): void {
        assetTrendsDateAggregationType.value = type;
    }

    function setDateFilter(dateType: number): void {
        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            if (dateType === DateRange.Custom.type) { // Custom
                showCustomDateRangeSheet.value = true;
                return;
            } else if (query.value.categoricalChartDateType === dateType) {
                return;
            }
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
            if (dateType === DateRange.Custom.type) { // Custom
                showCustomMonthRangeSheet.value = true;
                return;
            } else if (query.value.trendChartDateType === dateType) {
                return;
            }
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            if (dateType === DateRange.Custom.type) { // Custom
                showCustomDateRangeSheet.value = true;
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
            reload();
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

            showCustomDateRangeSheet.value = false;
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis && isString(startTime) && isString(endTime)) {
            const chartDateType = getDateTypeByDateRange(getYearMonthFirstUnixTime(startTime), getYearMonthLastUnixTime(endTime), firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.TrendAnalysis);

            changed = statisticsStore.updateTransactionStatisticsFilter({
                trendChartDateType: chartDateType,
                trendChartStartYearMonth: startTime,
                trendChartEndYearMonth: endTime
            });

            showCustomMonthRangeSheet.value = false;
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends && isNumber(startTime) && isNumber(endTime)) {
            const chartDateType = getDateTypeByDateRange(startTime, endTime, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.AssetTrends);

            changed = statisticsStore.updateTransactionStatisticsFilter({
                assetTrendsChartDateType: chartDateType,
                assetTrendsChartStartTime: startTime,
                assetTrendsChartEndTime: endTime
            });

            showCustomDateRangeSheet.value = false;
        }

        if (changed) {
            reload();
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
            reload();
        }
    }

    return {
        queryChartType, setChartType, setChartDataType, setSortingType, setTrendDateAggregationType,
        setAssetTrendsDateAggregationType, setDateFilter, setCustomDateFilter, shiftDateRange,
    };
}
