import { DateRange } from '@/core/datetime.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import { InsightsExplorer } from '@/models/explorer.ts';
import { type TransactionExplorerPartialFilter } from '@/stores/explorer.ts';
import type { useExplorerPageManagement } from './management.ts';
import type { useExplorerPageState } from './state.ts';
import { type InsightsExplorerProps } from './types.ts';

interface Context {
    clientSessionId: ReturnType<typeof useExplorerPageState>['clientSessionId'];
    currentFilter: ReturnType<typeof useExplorerPageState>['currentFilter'];
    activeTab: ReturnType<typeof useExplorerPageState>['activeTab'];
    explorersStore: ReturnType<typeof useExplorerPageState>['explorersStore'];
    initExploration: ReturnType<typeof useExplorerPageState>['initExploration'];
    loading: ReturnType<typeof useExplorerPageState>['loading'];
    initing: ReturnType<typeof useExplorerPageState>['initing'];
    accountsStore: ReturnType<typeof useExplorerPageState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useExplorerPageState>['transactionCategoriesStore'];
    transactionTagsStore: ReturnType<typeof useExplorerPageState>['transactionTagsStore'];
    loadExploration: ReturnType<typeof useExplorerPageManagement>['loadExploration'];
    snackbar: ReturnType<typeof useExplorerPageState>['snackbar'];
}

export function useExplorerPageLoading(context: Context) {
    const {
        clientSessionId, currentFilter, activeTab, explorersStore, initExploration, loading, initing, accountsStore,
        transactionCategoriesStore, transactionTagsStore, loadExploration, snackbar,
    } = context;

    function init(initProps: InsightsExplorerProps): void {
        clientSessionId.value = generateRandomUUID();

        const filter: TransactionExplorerPartialFilter = {
            dateRangeType: initProps.initDateRangeType ? parseInt(initProps.initDateRangeType) : undefined,
            startTime: initProps.initStartTime ? parseInt(initProps.initStartTime) : undefined,
            endTime: initProps.initEndTime ? parseInt(initProps.initEndTime) : undefined
        };

        let needReload = false;

        if (filter.dateRangeType !== currentFilter.value.dateRangeType) {
            needReload = true;
        } else if (filter.dateRangeType === DateRange.Custom.type) {
            if (filter.startTime !== currentFilter.value.startTime
                || filter.endTime !== currentFilter.value.endTime) {
                needReload = true;
            }
        }

        if (initProps.initActiveTab === 'query' || initProps.initActiveTab === 'table' || initProps.initActiveTab === 'chart') {
            if (initProps.initActiveTab !== activeTab.value) {
                activeTab.value = initProps.initActiveTab;
            }
        } else {
            activeTab.value = 'query';
        }

        explorersStore.initTransactionExplorerFilter(filter);

        if (initProps.initId) {
            if (explorersStore.currentExploration.id !== initProps.initId) {
                needReload = true;
            }
        } else if (!initProps.initId && !initProps.initActiveTab && !initProps.initDateRangeType && !initProps.initStartTime && !initProps.initEndTime) { // first time open the page
            explorersStore.updateCurrentExploration(InsightsExplorer.createNewExplorer(generateRandomUUID()));
            initExploration.value = InsightsExplorer.of(explorersStore.currentExploration);
        }

        if (!needReload && !explorersStore.transactionExplorerStateInvalid && !explorersStore.insightsExplorerListStateInvalid) {
            loading.value = false;
            initing.value = false;
            return;
        }

        Promise.all([
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false }),
            transactionTagsStore.loadAllTags({ force: false }),
            explorersStore.loadAllExplorationBasicInfos({ force: false })
        ]).then(() => {
            const promises: Promise<unknown>[] = [
                explorersStore.loadAllTransactions({ force: false })
            ];

            if (initProps.initId && explorersStore.currentExploration.id !== initProps.initId) {
                const loadExplorerPromise = loadExploration(initProps.initId, false, true);

                if (loadExplorerPromise) {
                    promises.push(loadExplorerPromise);
                }
            }

            return Promise.all(promises);
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

    function reload(force: boolean): Promise<unknown> | null {
        loading.value = true;

        return explorersStore.loadAllTransactions({
            force: force
        }).then(() => {
            loading.value = false;

            if (force) {
                snackbar.value?.showMessage('Data has been updated');
            }
        }).catch(error => {
            loading.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    return { init, reload };
}
