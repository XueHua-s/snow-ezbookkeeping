import type { TransactionExplorerDataDimension } from '@/core/explorer.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    parseBigDecimal
} from '@/lib/numeral.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { useExplorerDimensions } from './dimensions.ts';
import type { useExplorerState } from './state.ts';
import type { AmountRanges, CategoriedTransactions, TransactionInsightDataItemInQuery } from './types.ts';

interface Context {
    exchangeRatesStore: ReturnType<typeof useExplorerState>['exchangeRatesStore'];
    getDataCategoryInfo: ReturnType<typeof useExplorerDimensions>['getDataCategoryInfo'];
}

// Owns grouping for the explorer store.
export function useExplorerGrouping(context: Context) {
    const { exchangeRatesStore, getDataCategoryInfo } = context;

    function addTransactionToFilteredList(filteredTransactions: TransactionInsightDataItemInQuery[], filteredTransactionSourceAmountsInDefaultCurrency: BigDecimal[], filteredTransactionDestinationAmountsInDefaultCurrency: BigDecimal[], defaultCurrency: string, queryName: string, queryIndex: number, transaction: TransactionInsightDataItem): void {
        filteredTransactions.push({
            queryIndex: queryIndex,
            queryName: queryName,
            transaction: transaction
        });

        let sourceAmountInDefaultCurrency: BigDecimal | undefined = parseBigDecimal(transaction.sourceAmount);
        let destinationAmountInDefaultCurrency: BigDecimal | undefined = transaction.type === TransactionType.Transfer && transaction.destinationAccount ? parseBigDecimal(transaction.destinationAmount) : undefined;

        if (transaction.sourceAccount.currency !== defaultCurrency) {
            const amount = exchangeRatesStore.getExchangedAmount(parseBigDecimal(transaction.sourceAmount), transaction.sourceAccount.currency, defaultCurrency);
            sourceAmountInDefaultCurrency = amount?.truncate();
        }

        if (transaction.type === TransactionType.Transfer && transaction.destinationAccount && transaction.destinationAccount.currency !== defaultCurrency) {
            const amount = exchangeRatesStore.getExchangedAmount(parseBigDecimal(transaction.destinationAmount), transaction.destinationAccount.currency, defaultCurrency);
            destinationAmountInDefaultCurrency = amount?.truncate();
        }

        if (sourceAmountInDefaultCurrency) {
            filteredTransactionSourceAmountsInDefaultCurrency.push(sourceAmountInDefaultCurrency);
        }

        if (destinationAmountInDefaultCurrency) {
            filteredTransactionDestinationAmountsInDefaultCurrency.push(destinationAmountInDefaultCurrency);
        }
    }

    function addTransactionToCategoriedDataMap(timezoneUsedForDateRange: number, categoriedDataMap: Record<string, CategoriedTransactions>, categoryDimension: TransactionExplorerDataDimension, seriesDemension: TransactionExplorerDataDimension, allAmountRanges: AmountRanges, queryName: string, queryIndex: number, transaction: TransactionInsightDataItem): void {
        const categoriedInfo = getDataCategoryInfo(timezoneUsedForDateRange, categoryDimension, allAmountRanges.categorySourceAmountRanges, allAmountRanges.categoryDestinationAmountRanges, queryName, queryIndex, transaction);
        let categoriedData = categoriedDataMap[categoriedInfo.categoryId];

        if (!categoriedData) {
            categoriedData = {
                categoryName: categoriedInfo.categoryName,
                categoryNameNeedI18n: categoriedInfo.categoryNameNeedI18n,
                categoryNameI18nParameters: categoriedInfo.categoryNameI18nParameters,
                categoryId: categoriedInfo.categoryId,
                categoryIdType: categoriedInfo.categoryIdType,
                categoryDisplayOrders: categoriedInfo.categoryDisplayOrders,
                trasactions: {}
            };
            categoriedDataMap[categoriedInfo.categoryId] = categoriedData;
        }

        const seriesInfo = getDataCategoryInfo(timezoneUsedForDateRange, seriesDemension, allAmountRanges.seriesSourceAmountRanges, allAmountRanges.seriesDestinationAmountRanges, queryName, queryIndex, transaction);
        let seriesData = categoriedData.trasactions[seriesInfo.categoryId];

        if (!seriesData) {
            seriesData = {
                seriesName: seriesInfo.categoryName,
                seriesNameNeedI18n: seriesInfo.categoryNameNeedI18n,
                seriesNameI18nParameters: seriesInfo.categoryNameI18nParameters,
                seriesId: seriesInfo.categoryId,
                seriesIdType: seriesInfo.categoryIdType,
                seriesDisplayOrders: seriesInfo.categoryDisplayOrders,
                trasactions: []
            };
            categoriedData.trasactions[seriesInfo.categoryId] = seriesData;
        }

        seriesData.trasactions.push(transaction);
    }

    return { addTransactionToFilteredList, addTransactionToCategoriedDataMap };
}
