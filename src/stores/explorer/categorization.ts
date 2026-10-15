import { itemAndIndex } from '@/core/base.ts';
import {
    TransactionExplorerChartType,
    TransactionExplorerDataDimension
} from '@/core/explorer.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import type { InsightsExplorerMatchContext } from '@/models/explorer.ts';
import { computed } from 'vue';
import type { useExplorerAmountRanges } from './amount-ranges.ts';
import type { useExplorerGrouping } from './grouping.ts';
import type { useExplorerState } from './state.ts';
import type { useExplorerTransactions } from './transactions.ts';
import type { AmountRanges, CategoriedTransactions, TransactionInsightDataItemInQuery } from './types.ts';

interface Context {
    allTransactions: ReturnType<typeof useExplorerTransactions>['allTransactions'];
    currentExploration: ReturnType<typeof useExplorerState>['currentExploration'];
    userStore: ReturnType<typeof useExplorerState>['userStore'];
    addTransactionToFilteredList: ReturnType<typeof useExplorerGrouping>['addTransactionToFilteredList'];
    buildInsightsExplorerMatchContext: ReturnType<typeof useExplorerState>['buildInsightsExplorerMatchContext'];
    buildAllAmountRanges: ReturnType<typeof useExplorerAmountRanges>['buildAllAmountRanges'];
    addTransactionToCategoriedDataMap: ReturnType<typeof useExplorerGrouping>['addTransactionToCategoriedDataMap'];
}

// Owns categorization for the explorer store.
export function useExplorerCategorization(context: Context) {
    const {
        allTransactions, currentExploration, userStore, addTransactionToFilteredList, buildInsightsExplorerMatchContext,
        buildAllAmountRanges, addTransactionToCategoriedDataMap,
    } = context;

    const categoriedTransactions = computed<Record<string, CategoriedTransactions>>(() => {
        if (!allTransactions.value || allTransactions.value.length < 1) {
            return {};
        }

        const chartType = TransactionExplorerChartType.valueOf(currentExploration.value.chartType);
        const categoryDimension = TransactionExplorerDataDimension.valueOf(currentExploration.value.categoryDimension);
        const seriesDimension = chartType?.seriesDimensionRequired ? TransactionExplorerDataDimension.valueOf(currentExploration.value.seriesDimension) : TransactionExplorerDataDimension.SeriesDimensionDefault;

        if (!chartType || !categoryDimension || !seriesDimension) {
            return {};
        }

        const defaultCurrency = userStore.currentUserDefaultCurrency;
        const filteredTransactions: TransactionInsightDataItemInQuery[] = [];
        const filteredTransactionSourceAmountsInDefaultCurrency: BigDecimal[] = [];
        const filteredTransactionDestinationAmountsInDefaultCurrency: BigDecimal[] = [];

        for (const transaction of allTransactions.value) {
            if (!currentExploration.value.queries || currentExploration.value.queries.length < 1) {
                addTransactionToFilteredList(filteredTransactions, filteredTransactionSourceAmountsInDefaultCurrency, filteredTransactionDestinationAmountsInDefaultCurrency, defaultCurrency, '', 0, transaction);
                continue;
            }

            const matchContext: InsightsExplorerMatchContext = buildInsightsExplorerMatchContext(currentExploration.value, transaction);

            for (const [query, index] of itemAndIndex(currentExploration.value.queries)) {
                if (query.match(transaction, matchContext)) {
                    addTransactionToFilteredList(filteredTransactions, filteredTransactionSourceAmountsInDefaultCurrency, filteredTransactionDestinationAmountsInDefaultCurrency, defaultCurrency, query.name, index, transaction);

                    if (categoryDimension !== TransactionExplorerDataDimension.Query) {
                        break;
                    }
                }
            }
        }

        const categoriedDataMap: Record<string, CategoriedTransactions> = {};
        const allAmountRanges: AmountRanges = buildAllAmountRanges(categoryDimension, seriesDimension, filteredTransactionSourceAmountsInDefaultCurrency, filteredTransactionDestinationAmountsInDefaultCurrency, currentExploration.value.amountRangeCount);

        for (const item of filteredTransactions) {
            addTransactionToCategoriedDataMap(currentExploration.value.timezoneUsedForDateRange, categoriedDataMap, categoryDimension, seriesDimension, allAmountRanges, item.queryName, item.queryIndex, item.transaction);
        }

        return categoriedDataMap;
    });

    return { categoriedTransactions };
}
