import type { useExplorerPageState } from './state.ts';

interface Context {
    activeTab: ReturnType<typeof useExplorerPageState>['activeTab'];
    queryImportDialog: ReturnType<typeof useExplorerPageState>['queryImportDialog'];
    queryExportDialog: ReturnType<typeof useExplorerPageState>['queryExportDialog'];
    explorersStore: ReturnType<typeof useExplorerPageState>['explorersStore'];
    filteredTransactionsInDataTable: ReturnType<typeof useExplorerPageState>['filteredTransactionsInDataTable'];
    explorerDataTableTab: ReturnType<typeof useExplorerPageState>['explorerDataTableTab'];
    dataExportDialog: ReturnType<typeof useExplorerPageState>['dataExportDialog'];
    explorerChartTab: ReturnType<typeof useExplorerPageState>['explorerChartTab'];
}

export function useExplorerPageExport(context: Context) {
    const { activeTab, queryImportDialog, queryExportDialog, explorersStore, filteredTransactionsInDataTable, explorerDataTableTab, dataExportDialog, explorerChartTab } = context;

    function importQueries(): void {
        if (activeTab.value === 'query') {
            queryImportDialog.value?.open();
        }
    }

    function exportQueries(): void {
        if (activeTab.value === 'query') {
            queryExportDialog.value?.open({
                json: explorersStore.currentExploration.getQueryiesPrettyJson()
            });
        }
    }

    function exportResults(): void {
        if (activeTab.value === 'table' && filteredTransactionsInDataTable.value) {
            const results = explorerDataTableTab.value?.buildExportResults();

            if (results) {
                dataExportDialog.value?.open(results);
            }
        } else if (activeTab.value === 'chart') {
            const results = explorerChartTab.value?.buildExportResults();

            if (results) {
                dataExportDialog.value?.open(results);
            }
        }
    }

    return { importQueries, exportQueries, exportResults };
}
