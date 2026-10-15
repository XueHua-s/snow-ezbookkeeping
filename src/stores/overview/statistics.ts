import { type StartEndTime, DateRange } from '@/core/datetime.ts';
import { KeywordMatchMode } from '@/core/text.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import {
    isEquals
} from '@/lib/common.ts';
import {
    getTodayFirstUnixTime,
    getUnixTimeBeforeUnixTime
} from '@/lib/datetime.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type {
TransactionStatisticAssetTrendsResponseItem,
TransactionStatisticResponse
} from '@/models/transaction.ts';
import type { useOverviewState } from './state.ts';

interface Context {
    transactionDataRange: ReturnType<typeof useOverviewState>['transactionDataRange'];
    updateTransactionDateRange: ReturnType<typeof useOverviewState>['updateTransactionDateRange'];
    transactionCategoryStatisticsData: ReturnType<typeof useOverviewState>['transactionCategoryStatisticsData'];
    transactionCategoryStatisticsStateInvalid: ReturnType<typeof useOverviewState>['transactionCategoryStatisticsStateInvalid'];
    settingsStore: ReturnType<typeof useOverviewState>['settingsStore'];
    transactionAssetTrendsStateInvalid: ReturnType<typeof useOverviewState>['transactionAssetTrendsStateInvalid'];
    transactionAssetTrendsData: ReturnType<typeof useOverviewState>['transactionAssetTrendsData'];
}

// Owns statistics for the overview store.
export function useOverviewStatistics(context: Context) {
    const {
        transactionDataRange, updateTransactionDateRange, transactionCategoryStatisticsData,
        transactionCategoryStatisticsStateInvalid, settingsStore, transactionAssetTrendsStateInvalid, transactionAssetTrendsData,
    } = context;

    function loadTransactionCategoryStatistics({ force, dateType }: { force: boolean, dateType: number }): Promise<TransactionStatisticResponse> {
        if (transactionDataRange.value.today.startTime !== getTodayFirstUnixTime()) {
            updateTransactionDateRange();
            transactionCategoryStatisticsData.value = {};
            transactionCategoryStatisticsStateInvalid.value = {};
        }

        if (!force && !transactionCategoryStatisticsStateInvalid.value[dateType] && transactionCategoryStatisticsData.value[dateType]) {
            return Promise.resolve(transactionCategoryStatisticsData.value[dateType]);
        }

        let requestDateRange: StartEndTime | null = null;

        if (dateType === DateRange.ThisMonth.type) {
            requestDateRange = transactionDataRange.value.thisMonth;
        } else if (dateType === DateRange.ThisYear.type) {
            requestDateRange = transactionDataRange.value.thisYear;
        }

        if (!requestDateRange) {
            return Promise.reject({ message: 'Invalid date range' });
        }

        return new Promise((resolve, reject) => {
            services.getTransactionStatistics({
                startTime: requestDateRange.startTime,
                endTime: requestDateRange.endTime,
                tagFilter: settingsStore.appSettings.overviewTransactionTagFilterInHomePage,
                keyword: '',
                matchMode: KeywordMatchMode.Default.type,
                useTransactionTimezone: settingsStore.appSettings.timezoneUsedForStatisticsInHomePage === TimezoneTypeForStatistics.TransactionTimezone.type
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve transaction statistics' });
                    return;
                }

                if (transactionCategoryStatisticsStateInvalid.value[dateType]) {
                    transactionCategoryStatisticsStateInvalid.value[dateType] = false;
                }

                if (force && data.result && isEquals(transactionCategoryStatisticsData.value[dateType], data.result)) {
                    reject({ message: 'Data is up to date', isUpToDate: true });
                    return;
                }

                transactionCategoryStatisticsData.value[dateType] = data.result;

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

    function loadTransactionAssetTrends({ force, months }: { force: boolean, months: number }): Promise<TransactionStatisticAssetTrendsResponseItem[]> {
        if (!force && !transactionAssetTrendsStateInvalid.value) {
            return Promise.resolve(transactionAssetTrendsData.value);
        }

        const endTime = transactionDataRange.value.thisMonth.endTime;
        const startTime = getUnixTimeBeforeUnixTime(transactionDataRange.value.thisMonth.startTime, Math.max(1, months) - 1, 'months');

        return new Promise((resolve, reject) => {
            services.getTransactionStatisticsAssetTrends({
                startTime: startTime,
                endTime: endTime
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve transaction statistics' });
                    return;
                }

                if (transactionAssetTrendsStateInvalid.value) {
                    transactionAssetTrendsStateInvalid.value = false;
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

    return { loadTransactionCategoryStatistics, loadTransactionAssetTrends };
}
