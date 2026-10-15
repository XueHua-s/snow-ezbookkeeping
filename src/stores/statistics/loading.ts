import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import {
    isEquals
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type { TransactionStatisticAssetTrendsResponseItem, TransactionStatisticResponse, TransactionStatisticTrendsResponseItem } from '@/models/transaction.ts';
import type { useStatisticsFilter } from './filter.ts';
import type { useStatisticsState } from './state.ts';

interface Context {
    transactionStatisticsFilter: ReturnType<typeof useStatisticsState>['transactionStatisticsFilter'];
    settingsStore: ReturnType<typeof useStatisticsState>['settingsStore'];
    transactionStatisticsStateInvalid: ReturnType<typeof useStatisticsState>['transactionStatisticsStateInvalid'];
    updateTransactionStatisticsInvalidState: ReturnType<typeof useStatisticsFilter>['updateTransactionStatisticsInvalidState'];
    transactionCategoryStatisticsData: ReturnType<typeof useStatisticsState>['transactionCategoryStatisticsData'];
    transactionCategoryTrendsData: ReturnType<typeof useStatisticsState>['transactionCategoryTrendsData'];
    transactionAssetTrendsData: ReturnType<typeof useStatisticsState>['transactionAssetTrendsData'];
}

// Owns loading for the statistics store.
export function useStatisticsLoading(context: Context) {
    const {
        transactionStatisticsFilter, settingsStore, transactionStatisticsStateInvalid, updateTransactionStatisticsInvalidState,
        transactionCategoryStatisticsData, transactionCategoryTrendsData, transactionAssetTrendsData,
    } = context;

    function loadCategoricalAnalysis({ force }: { force: boolean }): Promise<TransactionStatisticResponse> {
        return new Promise((resolve, reject) => {
            services.getTransactionStatistics({
                startTime: transactionStatisticsFilter.value.categoricalChartStartTime,
                endTime: transactionStatisticsFilter.value.categoricalChartEndTime,
                tagFilter: transactionStatisticsFilter.value.tagFilter,
                keyword: transactionStatisticsFilter.value.keyword,
                matchMode: transactionStatisticsFilter.value.matchMode,
                useTransactionTimezone: settingsStore.appSettings.statistics.defaultTimezoneType === TimezoneTypeForStatistics.TransactionTimezone.type
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve transaction statistics' });
                    return;
                }

                if (transactionStatisticsStateInvalid.value) {
                    updateTransactionStatisticsInvalidState(false);
                }

                if (force && data.result && isEquals(transactionCategoryStatisticsData.value, data.result)) {
                    reject({ message: 'Data is up to date', isUpToDate: true });
                    return;
                }

                transactionCategoryStatisticsData.value = data.result;

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to retrieve transaction statistics', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve transaction statistics' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function loadTrendAnalysis({ force }: { force: boolean }): Promise<TransactionStatisticTrendsResponseItem[]> {
        return new Promise((resolve, reject) => {
            services.getTransactionStatisticsTrends({
                startYearMonth: transactionStatisticsFilter.value.trendChartStartYearMonth,
                endYearMonth: transactionStatisticsFilter.value.trendChartEndYearMonth,
                tagFilter: transactionStatisticsFilter.value.tagFilter,
                keyword: transactionStatisticsFilter.value.keyword,
                matchMode: transactionStatisticsFilter.value.matchMode,
                useTransactionTimezone: settingsStore.appSettings.statistics.defaultTimezoneType === TimezoneTypeForStatistics.TransactionTimezone.type
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve transaction statistics' });
                    return;
                }

                if (transactionStatisticsStateInvalid.value) {
                    updateTransactionStatisticsInvalidState(false);
                }

                if (force && data.result && isEquals(transactionCategoryTrendsData.value, data.result)) {
                    reject({ message: 'Data is up to date', isUpToDate: true });
                    return;
                }

                transactionCategoryTrendsData.value = data.result;

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to retrieve transaction statistics', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve transaction statistics' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function loadAssetTrends({ force }: { force: boolean }): Promise<TransactionStatisticAssetTrendsResponseItem[]> {
        return new Promise((resolve, reject) => {
            services.getTransactionStatisticsAssetTrends({
                startTime: transactionStatisticsFilter.value.assetTrendsChartStartTime,
                endTime: transactionStatisticsFilter.value.assetTrendsChartEndTime
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve transaction statistics' });
                    return;
                }

                if (transactionStatisticsStateInvalid.value) {
                    updateTransactionStatisticsInvalidState(false);
                }

                if (force && data.result && isEquals(transactionAssetTrendsData.value, data.result)) {
                    reject({ message: 'Data is up to date', isUpToDate: true });
                    return;
                }

                transactionAssetTrendsData.value = data.result;

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to retrieve transaction statistics', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve transaction statistics' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { loadCategoricalAnalysis, loadTrendAnalysis, loadAssetTrends };
}
