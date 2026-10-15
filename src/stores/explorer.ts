import { defineStore } from 'pinia';
import { useExplorerAmountRanges } from './explorer/amount-ranges.ts';
import { useExplorerCategorization } from './explorer/categorization.ts';
import { useExplorerChart } from './explorer/chart.ts';
import { useExplorerCoding } from './explorer/coding.ts';
import { useExplorerDimensions } from './explorer/dimensions.ts';
import { useExplorerFilter } from './explorer/filter.ts';
import { useExplorerGrouping } from './explorer/grouping.ts';
import { useExplorerLoading } from './explorer/loading.ts';
import { useExplorerManagement } from './explorer/management.ts';
import { useExplorerMutation } from './explorer/mutation.ts';
import { useExplorerState } from './explorer/state.ts';
import { useExplorerTable } from './explorer/table.ts';
import { useExplorerTransactions } from './explorer/transactions.ts';
export * from './explorer/types.ts';





export const useExplorersStore = defineStore('explorers', () => {
    const {
        settingsStore, userStore, accountsStore, transactionCategoriesStore, transactionTagsStore, exchangeRatesStore,
        currencyDisplayOrders, buildInsightsExplorerMatchContext, transactionExplorerFilter, transactionExplorerAllData,
        transactionExplorerStateInvalid, allExplorationBasicInfos, allExplorationBasicInfosMap, currentExploration,
        insightsExplorerListStateInvalid,
    } = useExplorerState();
    const { buildAllAmountRanges } = useExplorerAmountRanges();
    const { getDataCategoryInfo } = useExplorerDimensions({ userStore, currentExploration, accountsStore, settingsStore, currencyDisplayOrders, exchangeRatesStore });
    const { addTransactionToFilteredList, addTransactionToCategoriedDataMap } = useExplorerGrouping({ exchangeRatesStore, getDataCategoryInfo });
    const {
        loadInsightsExplorerList, addExplorationToInsightsExplorerList, updateExplorationInInsightsExplorerList,
        updateExplorerDisplayOrderInInsightsExplorerList, updateExplorationVisibilityInInsightsExplorerList,
        removeExplorationFromInsightsExplorerList,
    } = useExplorerMutation({ allExplorationBasicInfos, allExplorationBasicInfosMap });
    const { allTransactions, isUsingAmountRange } = useExplorerTransactions({ transactionExplorerAllData, accountsStore, transactionCategoriesStore, transactionTagsStore, currentExploration });
    const { filteredTransactionsInDataTable, filteredTransactionsForCustomChart, filteredTransactionsInDataTableStatistic } = useExplorerTable({ allTransactions, currentExploration, buildInsightsExplorerMatchContext, userStore, exchangeRatesStore });
    const { categoriedTransactions } = useExplorerCategorization({
        allTransactions, currentExploration, userStore, addTransactionToFilteredList, buildInsightsExplorerMatchContext,
        buildAllAmountRanges, addTransactionToCategoriedDataMap,
    });
    const { categoriedTransactionExplorerData } = useExplorerChart({ allTransactions, currentExploration, userStore, categoriedTransactions, exchangeRatesStore });
    const {
        updateTransactionExplorerInvalidState, updateInsightsExplorerListInvalidState, updateCurrentExploration,
        resetTransactionExplorers, initTransactionExplorerFilter, updateTransactionExplorerFilter,
        getTransactionExplorerPageParams,
    } = useExplorerFilter({
        transactionExplorerStateInvalid, insightsExplorerListStateInvalid, currentExploration, transactionExplorerFilter,
        transactionExplorerAllData, allExplorationBasicInfos, allExplorationBasicInfosMap, settingsStore, userStore,
    });
    const { loadAllTransactions, loadAllExplorationBasicInfos, getExploration } = useExplorerLoading({
        transactionExplorerFilter, transactionExplorerStateInvalid, updateTransactionExplorerInvalidState,
        transactionExplorerAllData, insightsExplorerListStateInvalid, allExplorationBasicInfos,
        updateInsightsExplorerListInvalidState, loadInsightsExplorerList,
    });
    const { saveExploration, changeExplorationDisplayOrder, updateExplorationDisplayOrders, hideExploration, deleteExploration } = useExplorerManagement({
        addExplorationToInsightsExplorerList, updateExplorationInInsightsExplorerList, allExplorationBasicInfos,
        insightsExplorerListStateInvalid, updateInsightsExplorerListInvalidState,
        updateExplorerDisplayOrderInInsightsExplorerList, loadAllExplorationBasicInfos,
        updateExplorationVisibilityInInsightsExplorerList, removeExplorationFromInsightsExplorerList,
    });
    const { generateCustomChartCode, cancelGenerateCustomChartCode } = useExplorerCoding();

    return {
        // states
        transactionExplorerFilter,
        transactionExplorerStateInvalid,
        allExplorationBasicInfos,
        allExplorationBasicInfosMap,
        currentExploration,
        insightsExplorerListStateInvalid,
        // computed
        isUsingAmountRange,
        filteredTransactionsInDataTable,
        filteredTransactionsForCustomChart,
        filteredTransactionsInDataTableStatistic,
        categoriedTransactionExplorerData,
        categoriedTransactions,
        // functions
        updateTransactionExplorerInvalidState,
        updateInsightsExplorerListInvalidState,
        updateCurrentExploration,
        resetTransactionExplorers,
        initTransactionExplorerFilter,
        updateTransactionExplorerFilter,
        getTransactionExplorerPageParams,
        loadAllTransactions,
        loadAllExplorationBasicInfos,
        getExploration,
        saveExploration,
        changeExplorationDisplayOrder,
        updateExplorationDisplayOrders,
        hideExploration,
        deleteExploration,
        generateCustomChartCode,
        cancelGenerateCustomChartCode
    };
});
