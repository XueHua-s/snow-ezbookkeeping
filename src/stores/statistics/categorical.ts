import { DEFAULT_ACCOUNT_COLOR } from '@/consts/color.ts';
import { DEFAULT_ACCOUNT_ICON } from '@/consts/icon.ts';
import { values } from '@/core/base.ts';
import { IconType } from '@/core/icon.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import {
    ChartDataType
} from '@/core/statistics.ts';
import {
    BIG_DECIMAL_ZERO,
    parseBigDecimal
} from '@/lib/numeral.ts';
import type { TransactionCategoricalAnalysisData, TransactionCategoricalAnalysisDataItem } from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useStatisticsCategoryTotals } from './category-totals.ts';
import type { useStatisticsState } from './state.ts';
import type { WritableTransactionCategoricalAnalysisData, WritableTransactionCategoricalAnalysisDataItem } from './types.ts';

interface Context {
    accountsStore: ReturnType<typeof useStatisticsState>['accountsStore'];
    transactionStatisticsFilter: ReturnType<typeof useStatisticsState>['transactionStatisticsFilter'];
    settingsStore: ReturnType<typeof useStatisticsState>['settingsStore'];
    userStore: ReturnType<typeof useStatisticsState>['userStore'];
    exchangeRatesStore: ReturnType<typeof useStatisticsState>['exchangeRatesStore'];
    transactionCategoryTotalAmountAnalysisData: ReturnType<typeof useStatisticsCategoryTotals>['transactionCategoryTotalAmountAnalysisData'];
    sortCategoryTotalAmountItems: ReturnType<typeof useStatisticsCategoryTotals>['sortCategoryTotalAmountItems'];
}

// Owns categorical for the statistics store.
export function useStatisticsCategorical(context: Context) {
    const {
        accountsStore, transactionStatisticsFilter, settingsStore, userStore, exchangeRatesStore,
        transactionCategoryTotalAmountAnalysisData, sortCategoryTotalAmountItems,
    } = context;

    const accountTotalAmountAnalysisData = computed<WritableTransactionCategoricalAnalysisData | null>(() => {
        if (!accountsStore.allPlainAccounts) {
            return null;
        }

        const allDataItems: Record<string, WritableTransactionCategoricalAnalysisDataItem> = {};
        let totalAmount: BigDecimal = BIG_DECIMAL_ZERO;
        let totalNonNegativeAmount: BigDecimal = BIG_DECIMAL_ZERO;

        for (const account of accountsStore.allPlainAccounts) {
            if (transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalAssets.type) {
                if (!account.isAsset) {
                    continue;
                }
            } else if (transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalLiabilities.type) {
                if (!account.isLiability) {
                    continue;
                }
            }

            if (transactionStatisticsFilter.value.filterAccountIds && transactionStatisticsFilter.value.filterAccountIds[account.id]) {
                continue;
            }

            let primaryAccount = accountsStore.allAccountsMap[account.parentId];

            if (!primaryAccount) {
                primaryAccount = account;
            }

            const primaryAccountCategoryDisplayOrder = settingsStore.accountCategoryDisplayOrders[primaryAccount.category] || Number.MAX_SAFE_INTEGER;

            let amount: BigDecimal = parseBigDecimal(account.balance);

            if (account.currency !== userStore.currentUserDefaultCurrency) {
                const finalAmount = exchangeRatesStore.getExchangedAmount(amount, account.currency, userStore.currentUserDefaultCurrency);

                if (!finalAmount) {
                    continue;
                }

                amount = finalAmount.truncate();
            }

            if (account.isLiability) {
                amount = amount.negate();
            }

            const data: WritableTransactionCategoricalAnalysisDataItem = {
                name: account.name,
                type: 'account',
                id: account.id,
                icon: account.icon || DEFAULT_ACCOUNT_ICON.icon,
                iconType: account.iconType,
                color: account.color || DEFAULT_ACCOUNT_COLOR,
                hidden: primaryAccount.hidden || account.hidden,
                displayOrders: [primaryAccountCategoryDisplayOrder, primaryAccount.displayOrder, account.displayOrder],
                value: amount
            };

            totalAmount = totalAmount.add(amount);

            if (amount.isPositive()) {
                totalNonNegativeAmount = totalNonNegativeAmount.add(amount);
            }

            allDataItems[account.id] = data;
        }

        return {
            value: totalAmount,
            totalNonNegativeAmount: totalNonNegativeAmount,
            items: allDataItems
        };
    });

    const currencyTotalAmountAnalysisData = computed<WritableTransactionCategoricalAnalysisData | null>(() => {
        if (!accountsStore.allPlainAccounts) {
            return null;
        }

        const allDataItems: Record<string, WritableTransactionCategoricalAnalysisDataItem> = {};
        let totalAmount: BigDecimal = BIG_DECIMAL_ZERO;
        let totalNonNegativeAmount: BigDecimal = BIG_DECIMAL_ZERO;

        for (const account of accountsStore.allPlainAccounts) {
            if (transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalAssetsByCurrency.type) {
                if (!account.isAsset) {
                    continue;
                }
            } else if (transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalLiabilitiesByCurrency.type) {
                if (!account.isLiability) {
                    continue;
                }
            }

            if (account.hidden) {
                continue;
            }

            if (transactionStatisticsFilter.value.filterAccountIds && transactionStatisticsFilter.value.filterAccountIds[account.id]) {
                continue;
            }

            let primaryAccount = accountsStore.allAccountsMap[account.parentId];

            if (!primaryAccount) {
                primaryAccount = account;
            }

            let originalAmount: BigDecimal = parseBigDecimal(account.balance);
            let amount: BigDecimal = originalAmount;

            if (account.currency !== userStore.currentUserDefaultCurrency) {
                const finalAmount = exchangeRatesStore.getExchangedAmount(amount, account.currency, userStore.currentUserDefaultCurrency);

                if (!finalAmount) {
                    continue;
                }

                amount = finalAmount.truncate();
            }

            if (account.isLiability) {
                originalAmount = originalAmount.negate();
                amount = amount.negate();
            }

            let data = allDataItems[account.currency];

            if (data) {
                data.value = data.value.add(amount);
                data.originalValue = data.originalValue?.add(originalAmount) || originalAmount;
            } else {
                data = {
                    name: account.currency,
                    type: 'total',
                    id: account.currency,
                    icon: DEFAULT_ACCOUNT_ICON.icon,
                    iconType: IconType.System,
                    color: DEFAULT_ACCOUNT_COLOR,
                    hidden: false,
                    displayOrders: [],
                    value: amount,
                    originalValue: originalAmount,
                    originalCurrency: account.currency
                };
            }

            totalAmount = totalAmount.add(amount);

            allDataItems[account.currency] = data;
        }

        for (const dataItem of values(allDataItems)) {
            if (dataItem.value.isPositive()) {
                totalNonNegativeAmount = totalNonNegativeAmount.add(dataItem.value);
            }
        }

        return {
            value: totalAmount,
            totalNonNegativeAmount: totalNonNegativeAmount,
            items: allDataItems
        };
    });

    const categoricalAnalysisData = computed<TransactionCategoricalAnalysisData>(() => {
        let combinedData: WritableTransactionCategoricalAnalysisData | null = null;

        if (transactionStatisticsFilter.value.chartDataType === ChartDataType.OutflowsByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseByPrimaryCategory.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseBySecondaryCategory.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.InflowsByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeByPrimaryCategory.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeBySecondaryCategory.type) {
            combinedData = transactionCategoryTotalAmountAnalysisData.value;
        } else if (transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalAssets.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalLiabilities.type) {
            combinedData = accountTotalAmountAnalysisData.value;
        } else if (transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalAssetsByCurrency.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalLiabilitiesByCurrency.type) {
            combinedData = currencyTotalAmountAnalysisData.value;
        }

        const allStatisticsItems: TransactionCategoricalAnalysisDataItem[] = [];

        if (combinedData && combinedData.items) {
            let maxTotalAmount: BigDecimal = BIG_DECIMAL_ZERO;

            for (const dataItem of values(combinedData.items)) {
                if (dataItem.value.greaterThan(maxTotalAmount)) {
                    maxTotalAmount = dataItem.value;
                }
            }

            for (const dataItem of values(combinedData.items)) {
                let percent = 0;

                if (transactionStatisticsFilter.value.chartDataType === ChartDataType.OutflowsByAccount.type ||
                    transactionStatisticsFilter.value.chartDataType === ChartDataType.InflowsByAccount.type) {
                    if (maxTotalAmount.isPositive()) {
                        percent = dataItem.value.divide(maxTotalAmount).multiply(100).toDoubleNumber();
                    } else {
                        percent = 0;
                    }

                    if (percent < 0) {
                        percent = 0;
                    }
                } else {
                    if (dataItem.value.isPositive()) {
                        percent = dataItem.value.divide(combinedData.totalNonNegativeAmount).multiply(100).toDoubleNumber();
                    } else {
                        percent = 0;
                    }

                    if (percent < 0) {
                        percent = 0;
                    }
                }

                const statisticDataItem: TransactionCategoricalAnalysisDataItem = {
                    name: dataItem.name,
                    type: dataItem.type,
                    id: dataItem.id,
                    icon: dataItem.icon,
                    iconType: dataItem.iconType,
                    color: dataItem.color,
                    hidden: dataItem.hidden,
                    displayOrders: dataItem.displayOrders,
                    value: dataItem.value,
                    percent: percent,
                    originalValue: dataItem.originalValue,
                    originalCurrency: dataItem.originalCurrency
                };

                allStatisticsItems.push(statisticDataItem);
            }
        }

        sortCategoryTotalAmountItems(allStatisticsItems, transactionStatisticsFilter.value);

        const statisticData: TransactionCategoricalAnalysisData = {
            value: combinedData?.value || BIG_DECIMAL_ZERO,
            items: allStatisticsItems
        };

        return statisticData;
    });

    return { categoricalAnalysisData };
}
