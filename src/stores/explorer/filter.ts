import { DateRange, DateRangeScene } from '@/core/datetime.ts';
import {
    DEFAULT_TRANSACTION_EXPLORER_DATE_RANGE
} from '@/core/explorer.ts';
import {
    isInteger
} from '@/lib/common.ts';
import {
    getDateRangeByDateType
} from '@/lib/datetime.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import {
    InsightsExplorer
} from '@/models/explorer.ts';
import type { useExplorerState } from './state.ts';
import type { TransactionExplorerPartialFilter } from './types.ts';

interface Context {
    transactionExplorerStateInvalid: ReturnType<typeof useExplorerState>['transactionExplorerStateInvalid'];
    insightsExplorerListStateInvalid: ReturnType<typeof useExplorerState>['insightsExplorerListStateInvalid'];
    currentExploration: ReturnType<typeof useExplorerState>['currentExploration'];
    transactionExplorerFilter: ReturnType<typeof useExplorerState>['transactionExplorerFilter'];
    transactionExplorerAllData: ReturnType<typeof useExplorerState>['transactionExplorerAllData'];
    allExplorationBasicInfos: ReturnType<typeof useExplorerState>['allExplorationBasicInfos'];
    allExplorationBasicInfosMap: ReturnType<typeof useExplorerState>['allExplorationBasicInfosMap'];
    settingsStore: ReturnType<typeof useExplorerState>['settingsStore'];
    userStore: ReturnType<typeof useExplorerState>['userStore'];
}

// Owns filter for the explorer store.
export function useExplorerFilter(context: Context) {
    const {
        transactionExplorerStateInvalid, insightsExplorerListStateInvalid, currentExploration, transactionExplorerFilter,
        transactionExplorerAllData, allExplorationBasicInfos, allExplorationBasicInfosMap, settingsStore, userStore,
    } = context;

    function updateTransactionExplorerInvalidState(invalidState: boolean): void {
        transactionExplorerStateInvalid.value = invalidState;
    }

    function updateInsightsExplorerListInvalidState(invalidState: boolean): void {
        insightsExplorerListStateInvalid.value = invalidState;
    }

    function updateCurrentExploration(exploration: InsightsExplorer): void {
        currentExploration.value = exploration;
    }

    function resetTransactionExplorers(): void {
        transactionExplorerFilter.value.dateRangeType = DEFAULT_TRANSACTION_EXPLORER_DATE_RANGE.type;
        transactionExplorerFilter.value.startTime = 0;
        transactionExplorerFilter.value.endTime = 0;
        transactionExplorerAllData.value = [];
        allExplorationBasicInfos.value = [];
        allExplorationBasicInfosMap.value = {};
        currentExploration.value = InsightsExplorer.createNewExplorer(generateRandomUUID());
        transactionExplorerStateInvalid.value = true;
        insightsExplorerListStateInvalid.value = true;
    }

    function initTransactionExplorerFilter(filter?: TransactionExplorerPartialFilter, resetQuery?: boolean): void {
        if (filter && isInteger(filter.dateRangeType)) {
            transactionExplorerFilter.value.dateRangeType = filter.dateRangeType;
        } else {
            transactionExplorerFilter.value.dateRangeType = settingsStore.appSettings.insightsExplorerDefaultDateRangeType;
        }

        let dateRangeTypeValid = true;

        if (!DateRange.isAvailableForScene(transactionExplorerFilter.value.dateRangeType, DateRangeScene.InsightsExplorer)) {
            transactionExplorerFilter.value.dateRangeType = DEFAULT_TRANSACTION_EXPLORER_DATE_RANGE.type;
            dateRangeTypeValid = false;
        }

        if (dateRangeTypeValid && transactionExplorerFilter.value.dateRangeType === DateRange.Custom.type) {
            if (filter && isInteger(filter.startTime)) {
                transactionExplorerFilter.value.startTime = filter.startTime;
            } else {
                transactionExplorerFilter.value.startTime = 0;
            }

            if (filter && isInteger(filter.endTime)) {
                transactionExplorerFilter.value.endTime = filter.endTime;
            } else {
                transactionExplorerFilter.value.endTime = 0;
            }
        } else {
            const dateRange = getDateRangeByDateType(transactionExplorerFilter.value.dateRangeType, userStore.currentUserFirstDayOfWeek, userStore.currentUserFiscalYearStart);

            if (dateRange) {
                transactionExplorerFilter.value.dateRangeType = dateRange.dateType;
                transactionExplorerFilter.value.startTime = dateRange.minTime;
                transactionExplorerFilter.value.endTime = dateRange.maxTime;
            }
        }

        if (resetQuery) {
            currentExploration.value = InsightsExplorer.createNewExplorer(generateRandomUUID());
        }
    }

    function updateTransactionExplorerFilter(filter: TransactionExplorerPartialFilter): boolean {
        let changed = false;

        if (filter && isInteger(filter.dateRangeType) && transactionExplorerFilter.value.dateRangeType !== filter.dateRangeType) {
            transactionExplorerFilter.value.dateRangeType = filter.dateRangeType;
            changed = true;
        }

        if (filter && isInteger(filter.startTime) && transactionExplorerFilter.value.startTime !== filter.startTime) {
            transactionExplorerFilter.value.startTime = filter.startTime;
            changed = true;
        }

        if (filter && isInteger(filter.endTime) && transactionExplorerFilter.value.endTime !== filter.endTime) {
            transactionExplorerFilter.value.endTime = filter.endTime;
            changed = true;
        }

        return changed;
    }

    function getTransactionExplorerPageParams(currentExplorationId: string, activeTab: string): string {
        const querys: string[] = [];

        if (currentExplorationId) {
            querys.push('id=' + currentExplorationId);
        }

        if (activeTab) {
            querys.push('activeTab=' + activeTab);
        }

        querys.push('dateRangeType=' + transactionExplorerFilter.value.dateRangeType);
        querys.push('startTime=' + transactionExplorerFilter.value.startTime);
        querys.push('endTime=' + transactionExplorerFilter.value.endTime);

        return querys.join('&');
    }

    return {
        updateTransactionExplorerInvalidState, updateInsightsExplorerListInvalidState, updateCurrentExploration,
        resetTransactionExplorers, initTransactionExplorerFilter, updateTransactionExplorerFilter,
        getTransactionExplorerPageParams,
    };
}
