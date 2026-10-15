import {
    isEquals
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import {
    InsightsExplorer,
    InsightsExplorerBasicInfo
} from '@/models/explorer.ts';
import type { TransactionInfoResponse } from '@/models/transaction.ts';
import type { useExplorerFilter } from './filter.ts';
import type { useExplorerMutation } from './mutation.ts';
import type { useExplorerState } from './state.ts';

interface Context {
    transactionExplorerFilter: ReturnType<typeof useExplorerState>['transactionExplorerFilter'];
    transactionExplorerStateInvalid: ReturnType<typeof useExplorerState>['transactionExplorerStateInvalid'];
    updateTransactionExplorerInvalidState: ReturnType<typeof useExplorerFilter>['updateTransactionExplorerInvalidState'];
    transactionExplorerAllData: ReturnType<typeof useExplorerState>['transactionExplorerAllData'];
    insightsExplorerListStateInvalid: ReturnType<typeof useExplorerState>['insightsExplorerListStateInvalid'];
    allExplorationBasicInfos: ReturnType<typeof useExplorerState>['allExplorationBasicInfos'];
    updateInsightsExplorerListInvalidState: ReturnType<typeof useExplorerFilter>['updateInsightsExplorerListInvalidState'];
    loadInsightsExplorerList: ReturnType<typeof useExplorerMutation>['loadInsightsExplorerList'];
}

// Owns loading for the explorer store.
export function useExplorerLoading(context: Context) {
    const {
        transactionExplorerFilter, transactionExplorerStateInvalid, updateTransactionExplorerInvalidState,
        transactionExplorerAllData, insightsExplorerListStateInvalid, allExplorationBasicInfos,
        updateInsightsExplorerListInvalidState, loadInsightsExplorerList,
    } = context;

    function loadAllTransactions({ force }: { force: boolean }): Promise<TransactionInfoResponse[]> {
        return new Promise((resolve, reject) => {
            services.getAllTransactions({
                startTime: transactionExplorerFilter.value.startTime,
                endTime: transactionExplorerFilter.value.endTime,
                withPictures: true
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve all transactions' });
                    return;
                }

                if (transactionExplorerStateInvalid.value) {
                    updateTransactionExplorerInvalidState(false);
                }

                if (force && data.result && isEquals(transactionExplorerAllData.value, data.result)) {
                    reject({ message: 'Data is up to date', isUpToDate: true });
                    return;
                }

                transactionExplorerAllData.value = data.result;

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to load all transactions', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve all transactions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function loadAllExplorationBasicInfos({ force }: { force?: boolean }): Promise<InsightsExplorerBasicInfo[]> {
        if (!force && !insightsExplorerListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allExplorationBasicInfos.value);
            });
        }

        return new Promise((resolve, reject) => {
            services.getAllExplorations().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve exploration list' });
                    return;
                }

                if (insightsExplorerListStateInvalid.value) {
                    updateInsightsExplorerListInvalidState(false);
                }

                const explorationBasicInfos = InsightsExplorerBasicInfo.ofMulti(data.result);

                if (force && data.result && isEquals(allExplorationBasicInfos.value, explorationBasicInfos)) {
                    reject({ message: 'Exploration list is up to date', isUpToDate: true });
                    return;
                }

                loadInsightsExplorerList(explorationBasicInfos);

                resolve(explorationBasicInfos);
            }).catch(error => {
                if (force) {
                    logger.error('failed to force load exploration list', error);
                } else {
                    logger.error('failed to load exploration list', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve exploration list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function getExploration({ explorationId }: { explorationId: string }): Promise<InsightsExplorer> {
        return new Promise((resolve, reject) => {
            services.getExploration({
                id: explorationId
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve exploration' });
                    return;
                }

                const transactionCategory = InsightsExplorer.of(data.result);

                resolve(transactionCategory);
            }).catch(error => {
                logger.error('failed to load exploration info', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve exploration' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { loadAllTransactions, loadAllExplorationBasicInfos, getExploration };
}
