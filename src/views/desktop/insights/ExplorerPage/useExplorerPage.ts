import type { SetupProps } from '@/lib/ui/props.ts';
import { useExplorerPageDateFilter } from './useExplorerPage/date-filter.ts';
import { useExplorerPageExport } from './useExplorerPage/export.ts';
import { useExplorerPageImport } from './useExplorerPage/import.ts';
import { useExplorerPageLifecycle } from './useExplorerPage/lifecycle.ts';
import { useExplorerPageLoading } from './useExplorerPage/loading.ts';
import { useExplorerPageManagement } from './useExplorerPage/management.ts';
import { useExplorerPageState } from './useExplorerPage/state.ts';
import { type InsightsExplorerProps } from './useExplorerPage/types.ts';
export * from './useExplorerPage/types.ts';

// Coordinates ExplorerPage state and effects; called once from the component setup.
export function useExplorerPage(props: SetupProps<InsightsExplorerProps>) {
    const {
        router, tt, accountsStore, transactionCategoriesStore, transactionTagsStore, explorersStore, timezoneTypeIconMap,
        renameDialog, confirmDialog, snackbar, dataExportDialog, queryImportDialog, queryExportDialog, explorerDataTableTab,
        explorerChartTab, explorerChangeDisplayOrderDialog, editDialog, loading, initing, updating, clientSessionId,
        initExploration, isCurrentDataTableEditable, activeTab, showCustomDateRangeDialog, firstDayOfWeek, fiscalYearStart,
        isCustomChartInAIMode, isCustomChartAIGenerating, queryExportFileName, sampleQueryJson, allExplorations,
        allVisibleExplorations, currentFilter, currentExploration, filteredTransactionsInDataTable, allDateRanges,
        allTimezoneTypesUsedForDateRange, canShiftDateRange, displayQueryDateRangeName, displayQueryStartTime,
        displayQueryEndTime, isCurrentExplorationModified, allTabs, getFilterLinkUrl,
    } = useExplorerPageState();
    const {
        createNewExploration, loadExploration, showChangeExplorerDisplayOrderDialog, saveExploration, restoreExploration,
        setExplorationName, hideExploration, removeExploration,
    } = useExplorerPageManagement({
        currentExploration, explorersStore, initExploration, router, getFilterLinkUrl, loading, snackbar,
        explorerChangeDisplayOrderDialog, renameDialog, tt, updating, clientSessionId, confirmDialog,
    });
    const { init, reload } = useExplorerPageLoading({
        clientSessionId, currentFilter, activeTab, explorersStore, initExploration, loading, initing, accountsStore,
        transactionCategoriesStore, transactionTagsStore, loadExploration, snackbar,
    });
    const { importQueries, exportQueries, exportResults } = useExplorerPageExport({ activeTab, queryImportDialog, queryExportDialog, explorersStore, filteredTransactionsInDataTable, explorerDataTableTab, dataExportDialog, explorerChartTab });
    const { setDateFilter, setCustomDateFilter, shiftDateRange } = useExplorerPageDateFilter({ showCustomDateRangeDialog, currentFilter, firstDayOfWeek, fiscalYearStart, explorersStore, loading, router, getFilterLinkUrl });
    const { onImportQueries } = useExplorerPageImport({ snackbar, explorersStore, initExploration, currentExploration });
    const { onShowTransaction, onUpdateTransactions, onShowDateRangeError } = useExplorerPageLifecycle({ editDialog, snackbar, reload, init, activeTab, initing, loading, router, getFilterLinkUrl, props });

    return {
        router, tt, timezoneTypeIconMap, renameDialog, confirmDialog, snackbar, dataExportDialog, queryImportDialog,
        queryExportDialog, explorerDataTableTab, explorerChartTab, explorerChangeDisplayOrderDialog, editDialog, loading,
        updating, isCurrentDataTableEditable, activeTab, showCustomDateRangeDialog, isCustomChartInAIMode,
        isCustomChartAIGenerating, queryExportFileName, sampleQueryJson, allExplorations, allVisibleExplorations, currentFilter,
        currentExploration, filteredTransactionsInDataTable, allDateRanges, allTimezoneTypesUsedForDateRange, canShiftDateRange,
        displayQueryDateRangeName, displayQueryStartTime, displayQueryEndTime, isCurrentExplorationModified, allTabs, reload,
        createNewExploration, loadExploration, showChangeExplorerDisplayOrderDialog, saveExploration, restoreExploration,
        setExplorationName, hideExploration, removeExploration, importQueries, exportQueries, exportResults, setDateFilter,
        setCustomDateFilter, shiftDateRange, onImportQueries, onShowTransaction, onUpdateTransactions, onShowDateRangeError,
    };
}
