import { DEFAULT_ACCOUNT_COLOR, DEFAULT_CATEGORY_COLOR } from '@/consts/color.ts';
import { DEFAULT_ACCOUNT_ICON, DEFAULT_CATEGORY_ICON } from '@/consts/icon.ts';
import { CategoryType } from '@/core/category.ts';
import { IconType } from '@/core/icon.ts';
import {
    ChartDataType
} from '@/core/statistics.ts';
import { TransactionRelatedAccountType } from '@/core/transaction.ts';
import {
    BIG_DECIMAL_ZERO,
    isBigDecimal
} from '@/lib/numeral.ts';
import { sortStatisticsItems } from '@/lib/statistics.ts';
import type { TransactionStatisticDataItemBase, TransactionStatisticResponseItemWithInfo } from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useStatisticsMetadata } from './metadata.ts';
import type { useStatisticsState } from './state.ts';
import type { TransactionStatisticsFilter, WritableTransactionCategoricalAnalysisData, WritableTransactionCategoricalAnalysisDataItem } from './types.ts';

interface Context {
    transactionCategoryStatisticsDataWithCategoryAndAccountInfo: ReturnType<typeof useStatisticsMetadata>['transactionCategoryStatisticsDataWithCategoryAndAccountInfo'];
    transactionStatisticsFilter: ReturnType<typeof useStatisticsState>['transactionStatisticsFilter'];
    settingsStore: ReturnType<typeof useStatisticsState>['settingsStore'];
}

// Owns category totals for the statistics store.
export function useStatisticsCategoryTotals(context: Context) {
    const { transactionCategoryStatisticsDataWithCategoryAndAccountInfo, transactionStatisticsFilter, settingsStore } = context;

    const transactionCategoryTotalAmountAnalysisData = computed<WritableTransactionCategoricalAnalysisData | null>(() => {
        if (!transactionCategoryStatisticsDataWithCategoryAndAccountInfo.value || !transactionCategoryStatisticsDataWithCategoryAndAccountInfo.value.items) {
            return null;
        }

        return getCategoryTotalAmountItems(transactionCategoryStatisticsDataWithCategoryAndAccountInfo.value.items, transactionStatisticsFilter.value);
    });

    function getCategoryTotalAmountItems(items: TransactionStatisticResponseItemWithInfo[], transactionStatisticsFilter: TransactionStatisticsFilter): WritableTransactionCategoricalAnalysisData {
        const allDataItems: Record<string, WritableTransactionCategoricalAnalysisDataItem> = {};
        let totalAmount = BIG_DECIMAL_ZERO;
        let totalNonNegativeAmount = BIG_DECIMAL_ZERO;

        for (const item of items) {
            if (!item.primaryAccount || !item.account || !item.primaryCategory || !item.category) {
                continue;
            }

            if (transactionStatisticsFilter.chartDataType === ChartDataType.OutflowsByAccount.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.TotalOutflows.type) {
                if (item.category.type === CategoryType.Transfer) {
                    if (item.relatedAccountType !== TransactionRelatedAccountType.TransferTo) {
                        continue;
                    }
                } else if (item.category.type !== CategoryType.Expense) {
                    continue;
                }
            } else if (transactionStatisticsFilter.chartDataType === ChartDataType.ExpenseByAccount.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.ExpenseByPrimaryCategory.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.ExpenseBySecondaryCategory.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.TotalExpense.type) {
                if (item.category.type !== CategoryType.Expense) {
                    continue;
                }
            } else if (transactionStatisticsFilter.chartDataType === ChartDataType.InflowsByAccount.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.TotalInflows.type) {
                if (item.category.type === CategoryType.Transfer) {
                    if (item.relatedAccountType !== TransactionRelatedAccountType.TransferFrom) {
                        continue;
                    }
                } else if (item.category.type !== CategoryType.Income) {
                    continue;
                }
            } else if (transactionStatisticsFilter.chartDataType === ChartDataType.IncomeByAccount.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.IncomeByPrimaryCategory.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.IncomeBySecondaryCategory.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.TotalIncome.type) {
                if (item.category.type !== CategoryType.Income) {
                    continue;
                }
            } else if (transactionStatisticsFilter.chartDataType === ChartDataType.NetCashFlow.type) {
                // Do Nothing
            } else if (transactionStatisticsFilter.chartDataType === ChartDataType.NetIncome.type) {
                if (item.category.type === CategoryType.Transfer) {
                    continue;
                }
            } else {
                continue;
            }

            if (transactionStatisticsFilter.filterAccountIds && transactionStatisticsFilter.filterAccountIds[item.account.id]) {
                continue;
            }

            if (transactionStatisticsFilter.filterCategoryIds && transactionStatisticsFilter.filterCategoryIds[item.category.id]) {
                continue;
            }

            if (transactionStatisticsFilter.chartDataType === ChartDataType.OutflowsByAccount.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.ExpenseByAccount.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.InflowsByAccount.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.IncomeByAccount.type) {
                if (isBigDecimal(item.amountInDefaultCurrency)) {
                    let data = allDataItems[item.account.id];

                    if (data) {
                        data.value = data.value.add(item.amountInDefaultCurrency);
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
                            value: item.amountInDefaultCurrency
                        };
                    }

                    let includeInTotal: boolean = true;

                    // total outflows / inflows do not include transfer transactions between unfiltered accounts
                    if (transactionStatisticsFilter.chartDataType === ChartDataType.OutflowsByAccount.type ||
                        transactionStatisticsFilter.chartDataType === ChartDataType.InflowsByAccount.type) {
                        if (item.relatedAccount && (!transactionStatisticsFilter.filterAccountIds || !transactionStatisticsFilter.filterAccountIds[item.relatedAccount.id])) {
                            includeInTotal = false;
                        }
                    }

                    if (includeInTotal) {
                        totalAmount = totalAmount.add(item.amountInDefaultCurrency);

                        if (item.amountInDefaultCurrency.isPositive()) {
                            totalNonNegativeAmount = totalNonNegativeAmount.add(item.amountInDefaultCurrency);
                        }
                    }

                    allDataItems[item.account.id] = data;
                }
            } else if (transactionStatisticsFilter.chartDataType === ChartDataType.ExpenseByPrimaryCategory.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.IncomeByPrimaryCategory.type) {
                if (isBigDecimal(item.amountInDefaultCurrency)) {
                    let data = allDataItems[item.primaryCategory.id];

                    if (data) {
                        data.value = data.value.add(item.amountInDefaultCurrency);
                    } else {
                        data = {
                            name: item.primaryCategory.name,
                            type: 'category',
                            id: item.primaryCategory.id,
                            icon: item.primaryCategory.icon || DEFAULT_CATEGORY_ICON.icon,
                            iconType: item.primaryCategory.iconType,
                            color: item.primaryCategory.color || DEFAULT_CATEGORY_COLOR,
                            hidden: item.primaryCategory.hidden,
                            displayOrders: [item.primaryCategory.type, item.primaryCategory.displayOrder],
                            value: item.amountInDefaultCurrency
                        };
                    }

                    totalAmount = totalAmount.add(item.amountInDefaultCurrency);

                    if (item.amountInDefaultCurrency.isPositive()) {
                        totalNonNegativeAmount = totalNonNegativeAmount.add(item.amountInDefaultCurrency);
                    }

                    allDataItems[item.primaryCategory.id] = data;
                }
            } else if (transactionStatisticsFilter.chartDataType === ChartDataType.ExpenseBySecondaryCategory.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.IncomeBySecondaryCategory.type) {
                if (isBigDecimal(item.amountInDefaultCurrency)) {
                    let data = allDataItems[item.category.id];

                    if (data) {
                        data.value = data.value.add(item.amountInDefaultCurrency);
                    } else {
                        data = {
                            name: item.category.name,
                            type: 'category',
                            id: item.category.id,
                            icon: item.category.icon || DEFAULT_CATEGORY_ICON.icon,
                            iconType: item.category.iconType,
                            color: item.category.color || DEFAULT_CATEGORY_COLOR,
                            hidden: item.primaryCategory.hidden || item.category.hidden,
                            displayOrders: [item.primaryCategory.type, item.primaryCategory.displayOrder, item.category.displayOrder],
                            value: item.amountInDefaultCurrency
                        };
                    }

                    totalAmount = totalAmount.add(item.amountInDefaultCurrency);

                    if (item.amountInDefaultCurrency.isPositive()) {
                        totalNonNegativeAmount = totalNonNegativeAmount.add(item.amountInDefaultCurrency);
                    }

                    allDataItems[item.category.id] = data;
                }
            } else if (transactionStatisticsFilter.chartDataType === ChartDataType.TotalOutflows.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.TotalExpense.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.TotalInflows.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.TotalIncome.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.NetCashFlow.type ||
                transactionStatisticsFilter.chartDataType === ChartDataType.NetIncome.type) {
                if (isBigDecimal(item.amountInDefaultCurrency)) {
                    let data = allDataItems['total'];
                    let amount = item.amountInDefaultCurrency;
                    let includeInTotal: boolean = true;

                    if (transactionStatisticsFilter.chartDataType === ChartDataType.NetCashFlow.type &&
                        (item.category.type === CategoryType.Expense || (item.category.type === CategoryType.Transfer && item.relatedAccountType === TransactionRelatedAccountType.TransferTo))) {
                        amount = amount.negate();
                    } else if (transactionStatisticsFilter.chartDataType === ChartDataType.NetIncome.type &&
                        item.category.type === CategoryType.Expense) {
                        amount = amount.negate();
                    }

                    // total outflows / inflows do not include transfer transactions between unfiltered accounts
                    if (transactionStatisticsFilter.chartDataType === ChartDataType.TotalOutflows.type ||
                        transactionStatisticsFilter.chartDataType === ChartDataType.TotalInflows.type ||
                        transactionStatisticsFilter.chartDataType === ChartDataType.NetCashFlow.type) {
                        if (item.relatedAccount && (!transactionStatisticsFilter.filterAccountIds || !transactionStatisticsFilter.filterAccountIds[item.relatedAccount.id])) {
                            includeInTotal = false;
                        }
                    }

                    if (!data) {
                        let name = '';

                        if (transactionStatisticsFilter.chartDataType === ChartDataType.TotalOutflows.type) {
                            name = ChartDataType.TotalOutflows.name;
                        } else if (transactionStatisticsFilter.chartDataType === ChartDataType.TotalExpense.type) {
                            name = ChartDataType.TotalExpense.name;
                        } else if (transactionStatisticsFilter.chartDataType === ChartDataType.TotalInflows.type) {
                            name = ChartDataType.TotalInflows.name;
                        } else if (transactionStatisticsFilter.chartDataType === ChartDataType.TotalIncome.type) {
                            name = ChartDataType.TotalIncome.name;
                        } else if (transactionStatisticsFilter.chartDataType === ChartDataType.NetCashFlow.type) {
                            name = ChartDataType.NetCashFlow.name;
                        } else if (transactionStatisticsFilter.chartDataType === ChartDataType.NetIncome.type) {
                            name = ChartDataType.NetIncome.name;
                        }

                        data = {
                            name: name,
                            type: 'total',
                            id: 'total',
                            icon: '',
                            iconType: IconType.System,
                            color: '',
                            hidden: false,
                            displayOrders: [1],
                            value: BIG_DECIMAL_ZERO
                        };
                    }

                    if (includeInTotal) {
                        data.value = data.value.add(amount);

                        totalAmount = totalAmount.add(amount);

                        if (item.amountInDefaultCurrency.isPositive()) {
                            totalNonNegativeAmount = totalNonNegativeAmount.add(amount);
                        }
                    }

                    allDataItems['total'] = data;
                }
            }
        }

        return {
            value: totalAmount,
            totalNonNegativeAmount: totalNonNegativeAmount,
            items: allDataItems
        };
    }

    function sortCategoryTotalAmountItems(items: TransactionStatisticDataItemBase[], transactionStatisticsFilter: TransactionStatisticsFilter): void {
        sortStatisticsItems(items, transactionStatisticsFilter.sortingType);
    }

    return { transactionCategoryTotalAmountAnalysisData, getCategoryTotalAmountItems, sortCategoryTotalAmountItems };
}
