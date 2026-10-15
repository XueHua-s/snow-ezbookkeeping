import { KnownErrorCode } from '@/consts/api.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import {
    isEquals,
    normalizeInteger,
    objectFieldWithValueToArrayItem
} from '@/lib/common.ts';
import {
    getTodayFirstUnixTime
} from '@/lib/datetime.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type {
TransactionAmountsRequestParams,
TransactionAmountsRequestType,
TransactionAmountsResponse,
TransactionUnreconciledCountItem
} from '@/models/transaction.ts';
import {
    LATEST_12MONTHS_TRANSACTION_AMOUNTS_REQUEST_TYPES
} from '@/models/transaction.ts';
import type { useOverviewState } from './state.ts';

interface Context {
    transactionDataRange: ReturnType<typeof useOverviewState>['transactionDataRange'];
    updateTransactionDateRange: ReturnType<typeof useOverviewState>['updateTransactionDateRange'];
    transactionOverviewOptions: ReturnType<typeof useOverviewState>['transactionOverviewOptions'];
    transactionOverviewStateInvalid: ReturnType<typeof useOverviewState>['transactionOverviewStateInvalid'];
    transactionOverviewData: ReturnType<typeof useOverviewState>['transactionOverviewData'];
    settingsStore: ReturnType<typeof useOverviewState>['settingsStore'];
    transactionUnreconciledCountsAccountIds: ReturnType<typeof useOverviewState>['transactionUnreconciledCountsAccountIds'];
    transactionUnreconciledCountsStateInvalid: ReturnType<typeof useOverviewState>['transactionUnreconciledCountsStateInvalid'];
    transactionUnreconciledCounts: ReturnType<typeof useOverviewState>['transactionUnreconciledCounts'];
}

// Owns loading for the overview store.
export function useOverviewLoading(context: Context) {
    const {
        transactionDataRange, updateTransactionDateRange, transactionOverviewOptions, transactionOverviewStateInvalid,
        transactionOverviewData, settingsStore, transactionUnreconciledCountsAccountIds,
        transactionUnreconciledCountsStateInvalid, transactionUnreconciledCounts,
    } = context;

    function loadTransactionOverview({ force, months }: { force: boolean, months?: number }): Promise<TransactionAmountsResponse> {
        const requestedMonths: number = normalizeInteger(months, 1, 1, 12);
        let dateChanged: boolean = false;
        let rangeChanged: boolean = false;

        if (transactionDataRange.value.today.startTime !== getTodayFirstUnixTime()) {
            dateChanged = true;
            updateTransactionDateRange();
        }

        if (requestedMonths > transactionOverviewOptions.value.loadedMonths) {
            rangeChanged = true;
        }

        if (!dateChanged && !rangeChanged && !force && !transactionOverviewStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(transactionOverviewData.value);
            });
        }

        const requestParams: TransactionAmountsRequestParams = {
            useTransactionTimezone: settingsStore.appSettings.timezoneUsedForStatisticsInHomePage === TimezoneTypeForStatistics.TransactionTimezone.type,
            today: transactionDataRange.value.today,
            thisWeek: transactionDataRange.value.thisWeek,
            thisMonth: transactionDataRange.value.thisMonth,
            thisYear: transactionDataRange.value.thisYear
        };

        const requestedMonthTypes: TransactionAmountsRequestType[] = LATEST_12MONTHS_TRANSACTION_AMOUNTS_REQUEST_TYPES.slice(-requestedMonths, -1);

        for (const requestType of requestedMonthTypes) {
            requestParams[requestType] = transactionDataRange.value[requestType];
        }

        const excludeAccountIds: string[] = objectFieldWithValueToArrayItem(settingsStore.appSettings.overviewAccountFilterInHomePage, true);
        const excludeCategoryIds: string[] = objectFieldWithValueToArrayItem(settingsStore.appSettings.overviewTransactionCategoryFilterInHomePage, true);
        const tagFilter: string = settingsStore.appSettings.overviewTransactionTagFilterInHomePage;

        return new Promise((resolve, reject) => {
            services.getTransactionAmounts(requestParams, excludeAccountIds, excludeCategoryIds, tagFilter).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve transaction overview' });
                    return;
                }

                if (transactionOverviewStateInvalid.value) {
                    transactionOverviewStateInvalid.value = false;
                }

                if (force && data.result && isEquals(transactionOverviewData.value, data.result)) {
                    reject({ message: 'Data is up to date', isUpToDate: true });
                    return;
                }

                transactionOverviewData.value = data.result;
                transactionOverviewOptions.value.loadedMonths = requestedMonths;

                resolve(data.result);
            }).catch(error => {
                if (force) {
                    logger.error('failed to force load transaction overview', error);
                } else {
                    logger.error('failed to load transaction overview', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve transaction overview' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function loadTransactionUnreconciledCounts({ force, accountIds }: { force: boolean, accountIds: string[] }): Promise<TransactionUnreconciledCountItem[]> {
        const requestedAccountIds = [...accountIds].sort();
        const accountIdsChanged = !isEquals(transactionUnreconciledCountsAccountIds.value, requestedAccountIds);

        if (!accountIdsChanged && !force && !transactionUnreconciledCountsStateInvalid.value) {
            return Promise.resolve(transactionUnreconciledCounts.value);
        }

        return new Promise((resolve, reject) => {
            services.getUnreconciledTransactionCounts({
                accountIds: requestedAccountIds
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result || !data.result.items) {
                    reject({ message: 'Unable to retrieve unreconciled transaction counts' });
                    return;
                }

                if (transactionUnreconciledCountsStateInvalid.value) {
                    transactionUnreconciledCountsStateInvalid.value = false;
                }

                transactionUnreconciledCountsAccountIds.value = requestedAccountIds;

                if (!accountIdsChanged && force && data.result && isEquals(transactionUnreconciledCounts.value, data.result.items)) {
                    reject({ message: 'Data is up to date', isUpToDate: true });
                    return;
                }

                transactionUnreconciledCounts.value = data.result.items;

                resolve(data.result.items);
            }).catch(error => {
                logger.error('failed to retrieve unreconciled transaction counts', error);

                if (error.response && error.response.data && error.response.data.errorCode === KnownErrorCode.UserLastReconciledTimeNotEnabled) {
                    transactionUnreconciledCounts.value = [];
                    transactionUnreconciledCountsAccountIds.value = requestedAccountIds;
                    transactionUnreconciledCountsStateInvalid.value = false;
                    resolve([]);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve unreconciled transaction counts' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { loadTransactionOverview, loadTransactionUnreconciledCounts };
}
