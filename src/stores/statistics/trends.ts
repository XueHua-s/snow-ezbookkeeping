import { entries, values } from '@/core/base.ts';
import {
    BIG_DECIMAL_ZERO
} from '@/lib/numeral.ts';
import type { TransactionStatisticTrendsResponseItemWithInfo, TransactionTrendsAnalysisData, TransactionTrendsAnalysisDataItem } from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useStatisticsCategoryTotals } from './category-totals.ts';
import type { useStatisticsMetadata } from './metadata.ts';
import type { useStatisticsState } from './state.ts';
import type { WritableTransactionTrendsAnalysisDataItem } from './types.ts';

interface Context {
    transactionCategoryTrendsData: ReturnType<typeof useStatisticsState>['transactionCategoryTrendsData'];
    assembleAccountAndCategoryInfo: ReturnType<typeof useStatisticsMetadata>['assembleAccountAndCategoryInfo'];
    getCategoryTotalAmountItems: ReturnType<typeof useStatisticsCategoryTotals>['getCategoryTotalAmountItems'];
    transactionStatisticsFilter: ReturnType<typeof useStatisticsState>['transactionStatisticsFilter'];
    sortCategoryTotalAmountItems: ReturnType<typeof useStatisticsCategoryTotals>['sortCategoryTotalAmountItems'];
}

// Owns trends for the statistics store.
export function useStatisticsTrends(context: Context) {
    const { transactionCategoryTrendsData, assembleAccountAndCategoryInfo, getCategoryTotalAmountItems, transactionStatisticsFilter, sortCategoryTotalAmountItems } = context;

    const transactionCategoryTrendsDataWithCategoryAndAccountInfo = computed<TransactionStatisticTrendsResponseItemWithInfo[]>(() => {
        const trendsData = transactionCategoryTrendsData.value;
        const finalTrendsData: TransactionStatisticTrendsResponseItemWithInfo[] = [];

        if (trendsData && trendsData.length) {
            for (const trendItem of trendsData) {
                const finalTrendItem: TransactionStatisticTrendsResponseItemWithInfo = {
                    year: trendItem.year,
                    month: trendItem.month,
                    items: []
                };

                if (trendItem && trendItem.items && trendItem.items.length) {
                    finalTrendItem.items.push(...assembleAccountAndCategoryInfo(trendItem.items));
                }

                finalTrendsData.push(finalTrendItem);
            }
        }

        return finalTrendsData;
    });

    const trendsAnalysisData = computed<TransactionTrendsAnalysisData | null>(() => {
        if (!transactionCategoryTrendsDataWithCategoryAndAccountInfo.value || !transactionCategoryTrendsDataWithCategoryAndAccountInfo.value.length) {
            return null;
        }

        const combinedDataMap: Record<string, WritableTransactionTrendsAnalysisDataItem> = {};

        for (const trendItem of transactionCategoryTrendsDataWithCategoryAndAccountInfo.value) {
            const totalAmountItems = getCategoryTotalAmountItems(trendItem.items, transactionStatisticsFilter.value);

            for (const [id, item] of entries(totalAmountItems.items)) {
                let combinedData = combinedDataMap[id];

                if (!combinedData) {
                    combinedData = {
                        name: item.name,
                        type: item.type,
                        id: item.id,
                        icon: item.icon,
                        iconType: item.iconType,
                        color: item.color,
                        hidden: item.hidden,
                        displayOrders: item.displayOrders,
                        value: BIG_DECIMAL_ZERO,
                        items: []
                    };
                }

                combinedData.items.push({
                    year: trendItem.year,
                    month1base: trendItem.month,
                    value: item.value
                });

                combinedData.value = combinedData.value.add(item.value);
                combinedDataMap[id] = combinedData;
            }
        }

        const totalAmountsTrends: TransactionTrendsAnalysisDataItem[] = [];

        for (const trendData of values(combinedDataMap)) {
            totalAmountsTrends.push(trendData);
        }

        sortCategoryTotalAmountItems(totalAmountsTrends, transactionStatisticsFilter.value);

        const trendsData: TransactionTrendsAnalysisData = {
            items: totalAmountsTrends
        };

        return trendsData;
    });

    return { trendsAnalysisData };
}
