import { CategoryType } from '@/core/category.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TransactionRelatedAccountType } from '@/core/transaction.ts';
import {
    max
} from '@/lib/math.ts';
import {
    BIG_DECIMAL_ZERO,
    isBigDecimal
} from '@/lib/numeral.ts';
import { sortStatisticsItems } from '@/lib/statistics.ts';
import {
    type TransactionCategoricalOverviewAnalysisData,
    type TransactionCategoricalOverviewAnalysisDataItem,
    TransactionCategoricalOverviewAnalysisDataItemType
} from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useStatisticsMetadata } from './metadata.ts';
import type { useStatisticsState } from './state.ts';
import type { TransactionStatisticsFilter } from './types.ts';

interface Context {
    transactionCategoryStatisticsDataWithCategoryAndAccountInfo: ReturnType<typeof useStatisticsMetadata>['transactionCategoryStatisticsDataWithCategoryAndAccountInfo'];
    transactionStatisticsFilter: ReturnType<typeof useStatisticsState>['transactionStatisticsFilter'];
    settingsStore: ReturnType<typeof useStatisticsState>['settingsStore'];
    accountsStore: ReturnType<typeof useStatisticsState>['accountsStore'];
}

// Owns overview for the statistics store.
export function useStatisticsOverview(context: Context) {
    const { transactionCategoryStatisticsDataWithCategoryAndAccountInfo, transactionStatisticsFilter, settingsStore, accountsStore } = context;

    const categoricalOverviewAnalysisData = computed<TransactionCategoricalOverviewAnalysisData | null>(() => {
        if (!transactionCategoryStatisticsDataWithCategoryAndAccountInfo.value || !transactionCategoryStatisticsDataWithCategoryAndAccountInfo.value.items) {
            return null;
        }

        const allDataItemsMap: Record<string, TransactionCategoricalOverviewAnalysisDataItem> = {};
        const allIncomeByPrimaryCategoryDataItems: TransactionCategoricalOverviewAnalysisDataItem[] = [];
        const allIncomeBySecondaryCategoryDataItems: TransactionCategoricalOverviewAnalysisDataItem[] = [];
        const allIncomeByAccountDataItems: TransactionCategoricalOverviewAnalysisDataItem[] = [];
        const allExpenseByAccountDataItems: TransactionCategoricalOverviewAnalysisDataItem[] = [];
        const allExpenseBySecondaryCategoryDataItems: TransactionCategoricalOverviewAnalysisDataItem[] = [];
        const allExpenseByPrimaryCategoryDataItems: TransactionCategoricalOverviewAnalysisDataItem[] = [];
        const allOpeningBalanceDataItems: TransactionCategoricalOverviewAnalysisDataItem[] = [];
        const allNetCashFlowDataItems: TransactionCategoricalOverviewAnalysisDataItem[] = [];

        let totalIncome: BigDecimal = BIG_DECIMAL_ZERO;
        let totalExpense: BigDecimal = BIG_DECIMAL_ZERO;

        for (const item of transactionCategoryStatisticsDataWithCategoryAndAccountInfo.value.items) {
            if (!item.primaryAccount || !item.account || !item.primaryCategory || !item.category) {
                continue;
            }

            if (item.relatedAccount && item.relatedAccountType === TransactionRelatedAccountType.TransferFrom) {
                continue;
            }

            if (!isBigDecimal(item.amountInDefaultCurrency)) {
                continue;
            }

            if (transactionStatisticsFilter.value.filterAccountIds && transactionStatisticsFilter.value.filterAccountIds[item.account.id]) {
                continue;
            }

            if (transactionStatisticsFilter.value.filterCategoryIds && transactionStatisticsFilter.value.filterCategoryIds[item.category.id]) {
                continue;
            }

            if (item.category.type === CategoryType.Income) {
                totalIncome = totalIncome.add(item.amountInDefaultCurrency);
            } else if (item.category.type === CategoryType.Expense) {
                totalExpense = totalExpense.add(item.amountInDefaultCurrency);
            }

            const primaryAccountCategoryDisplayOrder = settingsStore.accountCategoryDisplayOrders[item.primaryAccount.category] || Number.MAX_SAFE_INTEGER;
            const incomeByAccountKey = `${TransactionCategoricalOverviewAnalysisDataItemType.IncomeByAccount}:${item.account.id}`;
            const expenseByAccountKey = `${TransactionCategoricalOverviewAnalysisDataItemType.ExpenseByAccount}:${item.account.id}`;
            let incomeByAccountItem: TransactionCategoricalOverviewAnalysisDataItem | undefined = allDataItemsMap[incomeByAccountKey];
            let expenseByAccountItem: TransactionCategoricalOverviewAnalysisDataItem | undefined = allDataItemsMap[expenseByAccountKey];

            if (!incomeByAccountItem) {
                incomeByAccountItem = createNewTransactionCategoricalOverviewAnalysisDataItem(
                    item.account.id,
                    item.account.name,
                    TransactionCategoricalOverviewAnalysisDataItemType.IncomeByAccount,
                    [primaryAccountCategoryDisplayOrder, item.primaryAccount.displayOrder, item.account.displayOrder],
                    item.primaryAccount.hidden || item.account.hidden);
                allDataItemsMap[incomeByAccountKey] = incomeByAccountItem;
                allIncomeByAccountDataItems.push(incomeByAccountItem);
            }

            if (!expenseByAccountItem) {
                expenseByAccountItem = createNewTransactionCategoricalOverviewAnalysisDataItem(
                    item.account.id,
                    item.account.name,
                    TransactionCategoricalOverviewAnalysisDataItemType.ExpenseByAccount,
                    [primaryAccountCategoryDisplayOrder, item.primaryAccount.displayOrder, item.account.displayOrder],
                    item.primaryAccount.hidden || item.account.hidden);
                allDataItemsMap[expenseByAccountKey] = expenseByAccountItem;
                allExpenseByAccountDataItems.push(expenseByAccountItem);
            }

            if (item.category.type === CategoryType.Income) {
                const primaryCategoryItemKey = `${TransactionCategoricalOverviewAnalysisDataItemType.IncomeByPrimaryCategory}:${item.primaryCategory.id}`;
                const secondaryCategoryItemKey = `${TransactionCategoricalOverviewAnalysisDataItemType.IncomeBySecondaryCategory}:${item.category.id}`;

                let primaryCategoryDataItem: TransactionCategoricalOverviewAnalysisDataItem | undefined = allDataItemsMap[primaryCategoryItemKey];
                let secondaryCategoryDataItem: TransactionCategoricalOverviewAnalysisDataItem | undefined = allDataItemsMap[secondaryCategoryItemKey];

                if (!primaryCategoryDataItem) {
                    primaryCategoryDataItem = createNewTransactionCategoricalOverviewAnalysisDataItem(
                        item.primaryCategory.id,
                        item.primaryCategory.name,
                        TransactionCategoricalOverviewAnalysisDataItemType.IncomeByPrimaryCategory,
                        [item.primaryCategory.displayOrder],
                        item.primaryCategory.hidden);
                    allDataItemsMap[primaryCategoryItemKey] = primaryCategoryDataItem;
                    allIncomeByPrimaryCategoryDataItems.push(primaryCategoryDataItem);
                }

                if (!secondaryCategoryDataItem) {
                    secondaryCategoryDataItem = createNewTransactionCategoricalOverviewAnalysisDataItem(
                        item.category.id,
                        item.category.name,
                        TransactionCategoricalOverviewAnalysisDataItemType.IncomeBySecondaryCategory,
                        [item.primaryCategory.displayOrder, item.category.displayOrder],
                        item.primaryCategory.hidden || item.category.hidden);
                    allDataItemsMap[secondaryCategoryItemKey] = secondaryCategoryDataItem;
                    allIncomeBySecondaryCategoryDataItems.push(secondaryCategoryDataItem);
                }

                primaryCategoryDataItem.value = primaryCategoryDataItem.value.add(item.amountInDefaultCurrency);
                primaryCategoryDataItem.totalNonNegativeAmount = primaryCategoryDataItem.totalNonNegativeAmount.add(max(item.amountInDefaultCurrency, BIG_DECIMAL_ZERO));
                primaryCategoryDataItem.includeInPercent = true;
                primaryCategoryDataItem.outflows.push({ amount: item.amountInDefaultCurrency, relatedItem: secondaryCategoryDataItem });

                secondaryCategoryDataItem.value = secondaryCategoryDataItem.value.add(item.amountInDefaultCurrency);
                secondaryCategoryDataItem.totalNonNegativeAmount = secondaryCategoryDataItem.totalNonNegativeAmount.add(max(item.amountInDefaultCurrency, BIG_DECIMAL_ZERO));
                secondaryCategoryDataItem.includeInPercent = true;
                secondaryCategoryDataItem.inflows.push({ amount: item.amountInDefaultCurrency, relatedItem: primaryCategoryDataItem });
                secondaryCategoryDataItem.outflows.push({ amount: item.amountInDefaultCurrency, relatedItem: incomeByAccountItem });

                incomeByAccountItem.value = incomeByAccountItem.value.add(item.amountInDefaultCurrency);
                incomeByAccountItem.totalNonNegativeAmount = incomeByAccountItem.totalNonNegativeAmount.add(max(item.amountInDefaultCurrency, BIG_DECIMAL_ZERO));
                incomeByAccountItem.includeInPercent = true;
                incomeByAccountItem.inflows.push({ amount: item.amountInDefaultCurrency, relatedItem: secondaryCategoryDataItem });
            } else if (item.category.type === CategoryType.Expense) {
                const primaryCategoryItemKey = `${TransactionCategoricalOverviewAnalysisDataItemType.ExpenseByPrimaryCategory}:${item.primaryCategory.id}`;
                const secondaryCategoryItemKey = `${TransactionCategoricalOverviewAnalysisDataItemType.ExpenseBySecondaryCategory}:${item.category.id}`;

                let primaryCategoryDataItem: TransactionCategoricalOverviewAnalysisDataItem | undefined = allDataItemsMap[primaryCategoryItemKey];
                let secondaryCategoryDataItem: TransactionCategoricalOverviewAnalysisDataItem | undefined = allDataItemsMap[secondaryCategoryItemKey];

                if (!primaryCategoryDataItem) {
                    primaryCategoryDataItem = createNewTransactionCategoricalOverviewAnalysisDataItem(
                        item.primaryCategory.id,
                        item.primaryCategory.name,
                        TransactionCategoricalOverviewAnalysisDataItemType.ExpenseByPrimaryCategory,
                        [item.primaryCategory.displayOrder],
                        item.primaryCategory.hidden);
                    allDataItemsMap[primaryCategoryItemKey] = primaryCategoryDataItem;
                    allExpenseByPrimaryCategoryDataItems.push(primaryCategoryDataItem);
                }

                if (!secondaryCategoryDataItem) {
                    secondaryCategoryDataItem = createNewTransactionCategoricalOverviewAnalysisDataItem(
                        item.category.id,
                        item.category.name,
                        TransactionCategoricalOverviewAnalysisDataItemType.ExpenseBySecondaryCategory,
                        [item.primaryCategory.displayOrder, item.category.displayOrder],
                        item.primaryCategory.hidden || item.category.hidden);
                    allDataItemsMap[secondaryCategoryItemKey] = secondaryCategoryDataItem;
                    allExpenseBySecondaryCategoryDataItems.push(secondaryCategoryDataItem);
                }

                expenseByAccountItem.value = expenseByAccountItem.value.add(item.amountInDefaultCurrency);
                expenseByAccountItem.totalNonNegativeAmount = expenseByAccountItem.totalNonNegativeAmount.add(max(item.amountInDefaultCurrency, BIG_DECIMAL_ZERO));
                expenseByAccountItem.includeInPercent = true;
                expenseByAccountItem.outflows.push({ amount: item.amountInDefaultCurrency, relatedItem: secondaryCategoryDataItem });

                secondaryCategoryDataItem.value = secondaryCategoryDataItem.value.add(item.amountInDefaultCurrency);
                secondaryCategoryDataItem.totalNonNegativeAmount = secondaryCategoryDataItem.totalNonNegativeAmount.add(max(item.amountInDefaultCurrency, BIG_DECIMAL_ZERO));
                secondaryCategoryDataItem.includeInPercent = true;
                secondaryCategoryDataItem.inflows.push({ amount: item.amountInDefaultCurrency, relatedItem: expenseByAccountItem });
                secondaryCategoryDataItem.outflows.push({ amount: item.amountInDefaultCurrency, relatedItem: primaryCategoryDataItem });

                primaryCategoryDataItem.value = primaryCategoryDataItem.value.add(item.amountInDefaultCurrency);
                primaryCategoryDataItem.totalNonNegativeAmount = primaryCategoryDataItem.totalNonNegativeAmount.add(max(item.amountInDefaultCurrency, BIG_DECIMAL_ZERO));
                primaryCategoryDataItem.includeInPercent = true;
                primaryCategoryDataItem.inflows.push({ amount: item.amountInDefaultCurrency, relatedItem: secondaryCategoryDataItem });
            } else if (item.category.type === CategoryType.Transfer && item.relatedPrimaryAccount && item.relatedAccount) {
                const transferToAccountKey = `${TransactionCategoricalOverviewAnalysisDataItemType.ExpenseByAccount}:${item.relatedAccount.id}`;
                let transferToAccountItem: TransactionCategoricalOverviewAnalysisDataItem | undefined = allDataItemsMap[transferToAccountKey];

                if (!transferToAccountItem) {
                    const relatedPrimaryAccountCategoryDisplayOrder = settingsStore.accountCategoryDisplayOrders[item.relatedPrimaryAccount.category] || Number.MAX_SAFE_INTEGER;
                    transferToAccountItem = createNewTransactionCategoricalOverviewAnalysisDataItem(
                        item.relatedAccount.id,
                        item.relatedAccount.name,
                        TransactionCategoricalOverviewAnalysisDataItemType.ExpenseByAccount,
                        [relatedPrimaryAccountCategoryDisplayOrder, item.relatedPrimaryAccount.displayOrder, item.relatedAccount.displayOrder],
                        item.relatedPrimaryAccount.hidden || item.relatedAccount.hidden);
                    allDataItemsMap[transferToAccountKey] = transferToAccountItem;
                    allExpenseByAccountDataItems.push(transferToAccountItem);
                }

                incomeByAccountItem.outflows.push({ amount: item.amountInDefaultCurrency, relatedItem: transferToAccountItem });
                transferToAccountItem.inflows.push({ amount: item.amountInDefaultCurrency, relatedItem: incomeByAccountItem });
            }
        }

        sortCategoricalOverviewAnalysisDataItems(allIncomeByPrimaryCategoryDataItems, transactionStatisticsFilter.value);
        sortCategoricalOverviewAnalysisDataItems(allIncomeBySecondaryCategoryDataItems, transactionStatisticsFilter.value);
        sortCategoricalOverviewAnalysisDataItems(allIncomeByAccountDataItems, transactionStatisticsFilter.value);
        sortCategoricalOverviewAnalysisDataItems(allExpenseByAccountDataItems, transactionStatisticsFilter.value);
        sortCategoricalOverviewAnalysisDataItems(allExpenseBySecondaryCategoryDataItems, transactionStatisticsFilter.value);
        sortCategoricalOverviewAnalysisDataItems(allExpenseByPrimaryCategoryDataItems, transactionStatisticsFilter.value);

        for (const item of allExpenseByAccountDataItems) {
            const incomeByAccountKey = `${TransactionCategoricalOverviewAnalysisDataItemType.IncomeByAccount}:${item.id}`;
            const incomeByAccountItem: TransactionCategoricalOverviewAnalysisDataItem | undefined = allDataItemsMap[incomeByAccountKey];

            let accountTotalInflowsAmount: BigDecimal = BIG_DECIMAL_ZERO;
            let accountTotalIncomeAmount: BigDecimal = BIG_DECIMAL_ZERO;
            let accountTotalTransferAmount: BigDecimal = BIG_DECIMAL_ZERO;
            let accountTotalOutflowsAmount: BigDecimal = BIG_DECIMAL_ZERO;

            if (incomeByAccountItem) {
                for (const inflow of incomeByAccountItem.inflows) {
                    accountTotalInflowsAmount = accountTotalInflowsAmount.add(inflow.amount);
                    accountTotalIncomeAmount = accountTotalIncomeAmount.add(inflow.amount);
                }

                for (const outflow of incomeByAccountItem.outflows) {
                    accountTotalTransferAmount = accountTotalTransferAmount.add(outflow.amount);
                }
            }

            for (const inflow of item.inflows) {
                if (inflow.relatedItem.type === item.type && inflow.relatedItem.id === item.id) {
                    continue;
                }

                accountTotalInflowsAmount = accountTotalInflowsAmount.add(inflow.amount);
            }

            for (const outflow of item.outflows) {
                accountTotalOutflowsAmount = accountTotalOutflowsAmount.add(outflow.amount);
            }

            const accountBalance: BigDecimal = accountTotalIncomeAmount.subtract(accountTotalTransferAmount).subtract(accountTotalOutflowsAmount);
            const accountNetCashFlow: BigDecimal = accountTotalInflowsAmount.subtract(accountTotalTransferAmount).subtract(accountTotalOutflowsAmount);

            if (incomeByAccountItem && accountsStore.allAccountsMap[item.id]?.isAsset) {
                if (accountBalance.isPositive()) { // has positive balance, transfer the amount from income account to expense account
                    incomeByAccountItem.outflows.push({ amount: accountBalance.add(accountTotalOutflowsAmount), relatedItem: item });
                    item.inflows.push({ amount: accountBalance.add(accountTotalOutflowsAmount), relatedItem: incomeByAccountItem });
                } else if (accountNetCashFlow.isNegative()) { // has negative net cash flow, add the difference to income account
                    incomeByAccountItem.value = incomeByAccountItem.value.add(accountNetCashFlow.negate());
                    incomeByAccountItem.totalNonNegativeAmount = incomeByAccountItem.totalNonNegativeAmount.add(max(accountNetCashFlow.negate(), BIG_DECIMAL_ZERO));
                    incomeByAccountItem.outflows.push({ amount: accountNetCashFlow.negate(), relatedItem: item });
                    item.inflows.push({ amount: accountNetCashFlow.negate(), relatedItem: incomeByAccountItem });
                }
            }

            if (accountNetCashFlow.isPositive()) {
                let netCashFlowItem: TransactionCategoricalOverviewAnalysisDataItem | undefined = allDataItemsMap[TransactionCategoricalOverviewAnalysisDataItemType.NetCashFlow];

                if (!netCashFlowItem) {
                    netCashFlowItem = createNewTransactionCategoricalOverviewAnalysisDataItem(
                        TransactionCategoricalOverviewAnalysisDataItemType.NetCashFlow,
                        'Net Cash Flow',
                        TransactionCategoricalOverviewAnalysisDataItemType.NetCashFlow,
                        [Number.MAX_SAFE_INTEGER],
                        false);
                    allDataItemsMap[TransactionCategoricalOverviewAnalysisDataItemType.NetCashFlow] = netCashFlowItem;
                    allNetCashFlowDataItems.push(netCashFlowItem);
                }

                item.outflows.push({ amount: accountNetCashFlow, relatedItem: netCashFlowItem });

                netCashFlowItem.value = netCashFlowItem.value.add(accountNetCashFlow);
                netCashFlowItem.totalNonNegativeAmount = netCashFlowItem.totalNonNegativeAmount.add(max(accountNetCashFlow, BIG_DECIMAL_ZERO));
                netCashFlowItem.inflows.push({ amount: accountNetCashFlow, relatedItem: item });
            }
        }

        const allDataItems: TransactionCategoricalOverviewAnalysisDataItem[] = [
            ...allIncomeByPrimaryCategoryDataItems,
            ...allIncomeBySecondaryCategoryDataItems,
            ...allIncomeByAccountDataItems,
            ...allOpeningBalanceDataItems,
            ...allExpenseByAccountDataItems,
            ...allExpenseBySecondaryCategoryDataItems,
            ...allNetCashFlowDataItems,
            ...allExpenseByPrimaryCategoryDataItems
        ];

        return {
            totalIncome: totalIncome,
            totalExpense: totalExpense,
            items: allDataItems
        };
    });

    function createNewTransactionCategoricalOverviewAnalysisDataItem(id: string, name: string, type: TransactionCategoricalOverviewAnalysisDataItemType, displayOrders: number[], hidden: boolean): TransactionCategoricalOverviewAnalysisDataItem {
        const dataItem: TransactionCategoricalOverviewAnalysisDataItem = {
            id: id,
            name: name,
            type: type,
            displayOrders: displayOrders,
            hidden: hidden,
            inflows: [],
            outflows: [],
            value: BIG_DECIMAL_ZERO,
            totalNonNegativeAmount: BIG_DECIMAL_ZERO
        };

        return dataItem;
    }

    function sortCategoricalOverviewAnalysisDataItems(items: TransactionCategoricalOverviewAnalysisDataItem[], transactionStatisticsFilter: TransactionStatisticsFilter): void {
        let totalNonNegativeAmount: BigDecimal = BIG_DECIMAL_ZERO;

        for (const item of items) {
            totalNonNegativeAmount = totalNonNegativeAmount.add(item.totalNonNegativeAmount);
        }

        if (totalNonNegativeAmount.isPositive()) {
            for (const item of items) {
                if (!item.includeInPercent) {
                    continue;
                }

                item.percent = item.value.divide(totalNonNegativeAmount).multiply(100).toDoubleNumber();
            }
        }

        sortStatisticsItems(items, transactionStatisticsFilter.sortingType);
    }

    return { categoricalOverviewAnalysisData };
}
