import { DEFAULT_ACCOUNT_COLOR } from '@/consts/color.ts';
import { DEFAULT_ACCOUNT_ICON } from '@/consts/icon.ts';
import { values } from '@/core/base.ts';
import type { DateTime } from '@/core/datetime.ts';
import { IconType } from '@/core/icon.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import {
    ChartDataType
} from '@/core/statistics.ts';
import {
    getDayDifference,
    getYearMonthDayDateTime
} from '@/lib/datetime.ts';
import {
    BIG_DECIMAL_ZERO,
    isBigDecimal
} from '@/lib/numeral.ts';
import type {
    TransactionAssetTrendsAnalysisData, TransactionAssetTrendsAnalysisDataAmount, TransactionAssetTrendsAnalysisDataItem,
    TransactionStatisticAssetTrendsResponseDataItem, TransactionStatisticAssetTrendsResponseItem,
    TransactionStatisticAssetTrendsResponseItemWithInfo, TransactionStatisticResponseItem,
} from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useStatisticsCategoryTotals } from './category-totals.ts';
import type { useStatisticsMetadata } from './metadata.ts';
import type { useStatisticsState } from './state.ts';
import type { WritableTransactionAssetTrendsAnalysisDataItem } from './types.ts';

interface Context {
    transactionAssetTrendsData: ReturnType<typeof useStatisticsState>['transactionAssetTrendsData'];
    assembleAccountAndCategoryInfo: ReturnType<typeof useStatisticsMetadata>['assembleAccountAndCategoryInfo'];
    transactionStatisticsFilter: ReturnType<typeof useStatisticsState>['transactionStatisticsFilter'];
    settingsStore: ReturnType<typeof useStatisticsState>['settingsStore'];
    sortCategoryTotalAmountItems: ReturnType<typeof useStatisticsCategoryTotals>['sortCategoryTotalAmountItems'];
}

// Owns assets for the statistics store.
export function useStatisticsAssets(context: Context) {
    const { transactionAssetTrendsData, assembleAccountAndCategoryInfo, transactionStatisticsFilter, settingsStore, sortCategoryTotalAmountItems } = context;

    const assetTrendsDataWithAccountInfo = computed<TransactionStatisticAssetTrendsResponseItemWithInfo[]>(() => {
        const assetTrendsData = transactionAssetTrendsData.value;
        const finalAssetTrendsData: TransactionStatisticAssetTrendsResponseItemWithInfo[] = [];

        if (!assetTrendsData || !assetTrendsData.length) {
            return finalAssetTrendsData;
        }

        const firstAssetTrendItem: TransactionStatisticAssetTrendsResponseItem | undefined = assetTrendsData[0];

        if (!firstAssetTrendItem) {
            return finalAssetTrendsData;
        }

        const lastAssetTrendItemMap: Record<string, TransactionStatisticAssetTrendsResponseDataItem> = {};
        let lastAssetTrendItem: TransactionStatisticAssetTrendsResponseItem = firstAssetTrendItem;

        for (const item of firstAssetTrendItem.items) {
            lastAssetTrendItemMap[item.accountId] = item;
        }

        for (const assetTrendItem of assetTrendsData) {
            const statisticResponseItems: TransactionStatisticResponseItem[] = [];
            const existedAccountIds: Record<string, boolean> = {};
            const missingDays: number = getDayDifference(lastAssetTrendItem, assetTrendItem) - 1;
            const lastAssetTrendItemDate: DateTime = getYearMonthDayDateTime(lastAssetTrendItem.year, lastAssetTrendItem.month, lastAssetTrendItem.day);

            // fill in missing days with last known balance
            for (let i = 1; i <= missingDays; i++) {
                const missingStatisticResponseItems: TransactionStatisticResponseItem[] = [];
                const dateTime: DateTime = lastAssetTrendItemDate.add(i, 'days');

                for (const item of values(lastAssetTrendItemMap)) {
                    const statisticResponseItem: TransactionStatisticResponseItem = {
                        categoryId: '',
                        accountId: item.accountId,
                        amount: item.accountClosingBalance
                    };

                    missingStatisticResponseItems.push(statisticResponseItem);
                }

                const finalAssetTrendItem: TransactionStatisticAssetTrendsResponseItemWithInfo = {
                    year: dateTime.getGregorianCalendarYear(),
                    month: dateTime.getGregorianCalendarMonth(),
                    day: dateTime.getGregorianCalendarDay(),
                    items: assembleAccountAndCategoryInfo(missingStatisticResponseItems)
                };

                lastAssetTrendItem = assetTrendItem;
                finalAssetTrendsData.push(finalAssetTrendItem);
            }

            // fill in current day data
            for (const item of assetTrendItem.items) {
                const statisticResponseItem: TransactionStatisticResponseItem = {
                    categoryId: '',
                    accountId: item.accountId,
                    amount: item.accountClosingBalance
                };

                lastAssetTrendItemMap[item.accountId] = item;
                existedAccountIds[item.accountId] = true;
                statisticResponseItems.push(statisticResponseItem);
            }

            // fill in missing accounts with last known balance
            for (const item of values(lastAssetTrendItemMap)) {
                if (existedAccountIds[item.accountId]) {
                    continue;
                }

                const statisticResponseItem: TransactionStatisticResponseItem = {
                    categoryId: '',
                    accountId: item.accountId,
                    amount: item.accountClosingBalance
                };

                existedAccountIds[item.accountId] = true;
                statisticResponseItems.push(statisticResponseItem);
            }

            const finalAssetTrendItem: TransactionStatisticAssetTrendsResponseItemWithInfo = {
                year: assetTrendItem.year,
                month: assetTrendItem.month,
                day: assetTrendItem.day,
                items: assembleAccountAndCategoryInfo(statisticResponseItems)
            };

            lastAssetTrendItem = assetTrendItem;
            finalAssetTrendsData.push(finalAssetTrendItem);
        }

        return finalAssetTrendsData;
    });

    const assetTrendsData = computed<TransactionAssetTrendsAnalysisData | null>(() => {
        if (!assetTrendsDataWithAccountInfo.value || !assetTrendsDataWithAccountInfo.value.length) {
            return null;
        }

        const combinedDataMap: Record<string, WritableTransactionAssetTrendsAnalysisDataItem> = {};

        for (const dailyData of assetTrendsDataWithAccountInfo.value) {
            let dailyTotalAmount: BigDecimal = BIG_DECIMAL_ZERO;

            for (const item of dailyData.items) {
                if (!item.primaryAccount || !item.account) {
                    continue;
                }

                if (transactionStatisticsFilter.value.filterAccountIds && transactionStatisticsFilter.value.filterAccountIds[item.account.id]) {
                    continue;
                }

                if (!isBigDecimal(item.amountInDefaultCurrency)) {
                    continue;
                }

                if (transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalAssets.type) {
                    if (!item.account.isAsset) {
                        continue;
                    }
                } else if (transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalLiabilities.type) {
                    if (!item.account.isLiability) {
                        continue;
                    }
                } else if (transactionStatisticsFilter.value.chartDataType === ChartDataType.NetWorth.type) {
                    // Do Nothing
                } else {
                    continue;
                }

                let amount = item.amountInDefaultCurrency;

                if (item.account.isLiability) {
                    amount = amount.negate();
                }

                if (transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalAssets.type ||
                    transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalLiabilities.type) {
                    let data = combinedDataMap[item.account.id];

                    if (data) {
                        data.value = data.value.add(amount);
                    } else {
                        const primaryAccountCategoryDisplayOrder = settingsStore.accountCategoryDisplayOrders[item.primaryAccount.category] || Number.MAX_SAFE_INTEGER;

                        data = {
                            name: item.account.name,
                            type: 'account',
                            id: item.account.id,
                            icon: item.account.icon || DEFAULT_ACCOUNT_ICON.icon,
                            iconType: item.account.iconType,
                            color: item.account.color || DEFAULT_ACCOUNT_COLOR,
                            hidden: item.primaryAccount.hidden || item.account.hidden,
                            displayOrders: [primaryAccountCategoryDisplayOrder, item.primaryAccount.displayOrder, item.account.displayOrder],
                            value: amount,
                            items: []
                        };
                    }

                    const amountItem: TransactionAssetTrendsAnalysisDataAmount = {
                        year: dailyData.year,
                        month: dailyData.month,
                        day: dailyData.day,
                        value: amount
                    };
                    data.items.push(amountItem);
                    combinedDataMap[item.account.id] = data;
                }

                if (item.account.isAsset) {
                    dailyTotalAmount = dailyTotalAmount.add(amount);
                } else if (item.account.isLiability) {
                    dailyTotalAmount = dailyTotalAmount.subtract(amount);
                }
            }

            if (transactionStatisticsFilter.value.chartDataType === ChartDataType.NetWorth.type) {
                let data = combinedDataMap['total'];

                if (data) {
                    data.value = data.value.add(dailyTotalAmount);
                } else {
                    data = {
                        name: ChartDataType.NetWorth.name,
                        type: 'total',
                        id: 'total',
                        icon: '',
                        iconType: IconType.System,
                        color: '',
                        hidden: false,
                        displayOrders: [1],
                        value: dailyTotalAmount,
                        items: []
                    };
                }

                const amountItem: TransactionAssetTrendsAnalysisDataAmount = {
                    year: dailyData.year,
                    month: dailyData.month,
                    day: dailyData.day,
                    value: dailyTotalAmount
                };
                data.items.push(amountItem);
                combinedDataMap['total'] = data;
            }
        }

        const allAssetTrendsDataItems: TransactionAssetTrendsAnalysisDataItem[] = [];

        for (const assetTrendsDataItem of values(combinedDataMap)) {
            allAssetTrendsDataItems.push(assetTrendsDataItem);
        }

        sortCategoryTotalAmountItems(allAssetTrendsDataItems, transactionStatisticsFilter.value);

        const assetTrendsData: TransactionAssetTrendsAnalysisData = {
            items: allAssetTrendsDataItems
        };

        return assetTrendsData;
    });

    return { assetTrendsData };
}
