import { DateRange } from '@/core/datetime.ts';
import {
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import {
    arrayItemToObjectField,
    isDefined
} from '@/lib/common.ts';
import { type TransactionStatisticsPartialFilter } from '@/stores/statistics.ts';
import type { useTransactionPageState } from './state.ts';
import { type TransactionStatisticsProps } from './types.ts';

interface Context {
    filterKeyword: ReturnType<typeof useTransactionPageState>['filterKeyword'];
    query: ReturnType<typeof useTransactionPageState>['query'];
    analysisType: ReturnType<typeof useTransactionPageState>['analysisType'];
    trendDateAggregationType: ReturnType<typeof useTransactionPageState>['trendDateAggregationType'];
    assetTrendsDateAggregationType: ReturnType<typeof useTransactionPageState>['assetTrendsDateAggregationType'];
    statisticsStore: ReturnType<typeof useTransactionPageState>['statisticsStore'];
    loading: ReturnType<typeof useTransactionPageState>['loading'];
    initing: ReturnType<typeof useTransactionPageState>['initing'];
    accountsStore: ReturnType<typeof useTransactionPageState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useTransactionPageState>['transactionCategoriesStore'];
    snackbar: ReturnType<typeof useTransactionPageState>['snackbar'];
}

export function useTransactionPageInitialization(context: Context) {
    const {
        filterKeyword, query, analysisType, trendDateAggregationType, assetTrendsDateAggregationType, statisticsStore, loading,
        initing, accountsStore, transactionCategoriesStore, snackbar,
    } = context;

    function init(initProps: TransactionStatisticsProps): void {
        let needReload = !isDefined(initProps.initAnalysisType);

        const filter: TransactionStatisticsPartialFilter = {
            chartDataType: initProps.initChartDataType ? parseInt(initProps.initChartDataType) : undefined,
            filterAccountIds: initProps.initFilterAccountIds ? arrayItemToObjectField(initProps.initFilterAccountIds.split(','), true) : {},
            filterCategoryIds: initProps.initFilterCategoryIds ? arrayItemToObjectField(initProps.initFilterCategoryIds.split(','), true) : {},
            tagFilter: initProps.initTagFilter,
            keyword: initProps.initKeyword,
            matchMode: initProps.initMatchMode ? parseInt(initProps.initMatchMode) : undefined,
            sortingType: initProps.initSortingType ? parseInt(initProps.initSortingType) : undefined
        };

        filterKeyword.value = filter.keyword || '';

        if (initProps.initAnalysisType === StatisticsAnalysisType.CategoricalAnalysis.toString()) {
            filter.categoricalChartType = initProps.initChartType ? parseInt(initProps.initChartType) : undefined;
            filter.categoricalChartDateType = initProps.initChartDateType ? parseInt(initProps.initChartDateType) : undefined;
            filter.categoricalChartStartTime = initProps.initStartTime ? parseInt(initProps.initStartTime) : undefined;
            filter.categoricalChartEndTime = initProps.initEndTime ? parseInt(initProps.initEndTime) : undefined;

            if (filter.categoricalChartDateType !== query.value.categoricalChartDateType) {
                needReload = true;
            } else if (filter.categoricalChartDateType === DateRange.Custom.type) {
                if (filter.categoricalChartStartTime !== query.value.categoricalChartStartTime
                    || filter.categoricalChartEndTime !== query.value.categoricalChartEndTime) {
                    needReload = true;
                }
            }

            if (initProps.initAnalysisType !== analysisType.value.toString()) {
                analysisType.value = StatisticsAnalysisType.CategoricalAnalysis;
                needReload = true;
            }
        } else if (initProps.initAnalysisType === StatisticsAnalysisType.TrendAnalysis.toString()) {
            filter.trendChartType = initProps.initChartType ? parseInt(initProps.initChartType) : undefined;
            filter.trendChartDateType = initProps.initChartDateType ? parseInt(initProps.initChartDateType) : undefined;
            filter.trendChartStartYearMonth = initProps.initStartTime;
            filter.trendChartEndYearMonth = initProps.initEndTime;

            if (filter.trendChartDateType !== query.value.trendChartDateType) {
                needReload = true;
            } else if (filter.trendChartDateType === DateRange.Custom.type) {
                if (filter.trendChartStartYearMonth !== query.value.trendChartStartYearMonth
                    || filter.trendChartEndYearMonth !== query.value.trendChartEndYearMonth) {
                    needReload = true;
                }
            }

            if (initProps.initAnalysisType !== analysisType.value.toString()) {
                analysisType.value = StatisticsAnalysisType.TrendAnalysis;
                needReload = true;
            }

            if (initProps.initTrendDateAggregationType) {
                trendDateAggregationType.value = parseInt(initProps.initTrendDateAggregationType);
            }
        } else if (initProps.initAnalysisType === StatisticsAnalysisType.AssetTrends.toString()) {
            filter.assetTrendsChartType = initProps.initChartType ? parseInt(initProps.initChartType) : undefined;
            filter.assetTrendsChartDateType = initProps.initChartDateType ? parseInt(initProps.initChartDateType) : undefined;
            filter.assetTrendsChartStartTime = initProps.initStartTime ? parseInt(initProps.initStartTime) : undefined;
            filter.assetTrendsChartEndTime = initProps.initEndTime ? parseInt(initProps.initEndTime) : undefined;

            if (filter.assetTrendsChartDateType !== query.value.assetTrendsChartDateType) {
                needReload = true;
            } else if (filter.assetTrendsChartDateType === DateRange.Custom.type) {
                if (filter.assetTrendsChartStartTime !== query.value.assetTrendsChartStartTime
                    || filter.assetTrendsChartEndTime !== query.value.assetTrendsChartEndTime) {
                    needReload = true;
                }
            }

            if (initProps.initAnalysisType !== analysisType.value.toString()) {
                analysisType.value = StatisticsAnalysisType.AssetTrends;
                needReload = true;
            }

            if (initProps.initAssetTrendsDateAggregationType) {
                assetTrendsDateAggregationType.value = parseInt(initProps.initAssetTrendsDateAggregationType);
            }
        }

        if (!isDefined(initProps.initAnalysisType)) {
            analysisType.value = StatisticsAnalysisType.CategoricalAnalysis;
            statisticsStore.initTransactionStatisticsFilter(analysisType.value);
        } else {
            statisticsStore.initTransactionStatisticsFilter(analysisType.value, filter);
        }

        if (!needReload && !statisticsStore.transactionStatisticsStateInvalid) {
            loading.value = false;
            initing.value = false;
            return;
        }

        Promise.all([
            accountsStore.loadAllAccounts({force: false}),
            transactionCategoriesStore.loadAllCategories({force: false})
        ]).then(() => {
            if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
                return statisticsStore.loadCategoricalAnalysis({
                    force: false
                }) as Promise<unknown>;
            } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
                return statisticsStore.loadTrendAnalysis({
                    force: false
                }) as Promise<unknown>;
            } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
                return statisticsStore.loadAssetTrends({
                    force: false
                }) as Promise<unknown>;
            } else {
                return Promise.reject('An error occurred');
            }
        }).then(() => {
            loading.value = false;
            initing.value = false;
        }).catch(error => {
            loading.value = false;
            initing.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    return { init };
}
