import {
    ChartDataType,
    ChartDateAggregationType,
    ChartSortingType,
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import { computed } from 'vue';
import type { useTransactionPageState } from './state.ts';

interface Context {
    analysisType: ReturnType<typeof useTransactionPageState>['analysisType'];
    query: ReturnType<typeof useTransactionPageState>['query'];
    statisticsStore: ReturnType<typeof useTransactionPageState>['statisticsStore'];
    trendDateAggregationType: ReturnType<typeof useTransactionPageState>['trendDateAggregationType'];
    assetTrendsDateAggregationType: ReturnType<typeof useTransactionPageState>['assetTrendsDateAggregationType'];
    loading: ReturnType<typeof useTransactionPageState>['loading'];
    router: ReturnType<typeof useTransactionPageState>['router'];
    getFilterLinkUrl: ReturnType<typeof useTransactionPageState>['getFilterLinkUrl'];
}

export function useTransactionPageChartSelection(context: Context) {
    const { analysisType, query, statisticsStore, trendDateAggregationType, assetTrendsDateAggregationType, loading, router, getFilterLinkUrl } = context;

    const queryAnalysisType = computed<StatisticsAnalysisType>({
        get: () => analysisType.value,
        set: (value: number) => {
            setAnalysisType(value);
        }
    });

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

    const queryChartDataType = computed<number>({
        get: () => query.value.chartDataType,
        set: (value: number) => {
            setChartDataType(value);
        }
    });

    const querySortingType = computed<number>({
        get: () => query.value.sortingType,
        set: (value: number) => {
            setSortingType(value);
        }
    });

    const isQuerySpecialChartType = computed<boolean>(() => {
        return ChartDataType.valueOf(queryChartDataType.value)?.specialChart ?? false;
    });

    function setAnalysisType(type: StatisticsAnalysisType): void {
        if (analysisType.value === type) {
            return;
        }

        if (!ChartDataType.isAvailableForAnalysisType(query.value.chartDataType, type)) {
            let defaultChartDataType: ChartDataType = ChartDataType.Default;

            if (type === StatisticsAnalysisType.AssetTrends) {
                defaultChartDataType = ChartDataType.DefaultForAssetTrends;
            }

            statisticsStore.updateTransactionStatisticsFilter({
                chartDataType: defaultChartDataType.type
            });
        }

        if (analysisType.value !== StatisticsAnalysisType.TrendAnalysis && type === StatisticsAnalysisType.TrendAnalysis) {
            trendDateAggregationType.value = ChartDateAggregationType.Default.type;
        } else if (analysisType.value !== StatisticsAnalysisType.AssetTrends && type === StatisticsAnalysisType.AssetTrends) {
            assetTrendsDateAggregationType.value = ChartDateAggregationType.Default.type;
        }

        analysisType.value = type;
        loading.value = true;
        statisticsStore.updateTransactionStatisticsInvalidState(true);
        router.push(getFilterLinkUrl());
    }

    function setChartType(type?: number): void {
        let changed = false;

        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            changed = statisticsStore.updateTransactionStatisticsFilter({
                categoricalChartType: type
            });
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
            changed = statisticsStore.updateTransactionStatisticsFilter({
                trendChartType: type
            });
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            changed = statisticsStore.updateTransactionStatisticsFilter({
                assetTrendsChartType: type
            });
        }

        if (changed) {
            router.push(getFilterLinkUrl());
        }
    }

    function setChartDataType(type: number): void {
        const changed = statisticsStore.updateTransactionStatisticsFilter({
            chartDataType: type
        });

        if (changed) {
            router.push(getFilterLinkUrl());
        }
    }

    function setSortingType(type: number): void {
        if (type < ChartSortingType.Amount.type || type > ChartSortingType.Name.type) {
            return;
        }

        const changed = statisticsStore.updateTransactionStatisticsFilter({
            sortingType: type
        });

        if (changed) {
            router.push(getFilterLinkUrl());
        }
    }

    function setTrendDateAggregationType(type: number): void {
        const changed = trendDateAggregationType.value !== type;
        trendDateAggregationType.value = type;

        if (changed) {
            router.push(getFilterLinkUrl());
        }
    }

    function setAssetTrendsDateAggregationType(type: number): void {
        const changed = assetTrendsDateAggregationType.value !== type;
        assetTrendsDateAggregationType.value = type;

        if (changed) {
            router.push(getFilterLinkUrl());
        }
    }

    return {
        queryAnalysisType, queryChartType, queryChartDataType, querySortingType, isQuerySpecialChartType,
        setTrendDateAggregationType, setAssetTrendsDateAggregationType,
    };
}
