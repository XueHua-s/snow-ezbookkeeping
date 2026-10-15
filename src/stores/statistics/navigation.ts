import { type TimeRangeAndDateType, DateRange } from '@/core/datetime.ts';
import {
    ChartDataType,
    ChartDateAggregationType,
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import { getFinalAccountIdsByFilteredAccountIds } from '@/lib/account.ts';
import { getFinalCategoryIdsByFilteredCategoryIds } from '@/lib/category.ts';
import {
    isObjectEmpty,
    objectFieldToArrayItem
} from '@/lib/common.ts';
import type { useStatisticsState } from './state.ts';

interface Context {
    transactionStatisticsFilter: ReturnType<typeof useStatisticsState>['transactionStatisticsFilter'];
    accountsStore: ReturnType<typeof useStatisticsState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useStatisticsState>['transactionCategoriesStore'];
}

// Owns navigation for the statistics store.
export function useStatisticsNavigation(context: Context) {
    const { transactionStatisticsFilter, accountsStore, transactionCategoriesStore } = context;

    function getTransactionStatisticsPageParams(analysisType: StatisticsAnalysisType, trendDateAggregationType: number, assetTrendsDateAggregationType: number): string {
        const querys: string[] = [];

        querys.push('analysisType=' + analysisType);
        querys.push('chartDataType=' + transactionStatisticsFilter.value.chartDataType);

        if (analysisType === StatisticsAnalysisType.CategoricalAnalysis) {
            querys.push('chartType=' + transactionStatisticsFilter.value.categoricalChartType);
            querys.push('chartDateType=' + transactionStatisticsFilter.value.categoricalChartDateType);

            if (transactionStatisticsFilter.value.categoricalChartDateType === DateRange.Custom.type) {
                querys.push('startTime=' + transactionStatisticsFilter.value.categoricalChartStartTime);
                querys.push('endTime=' + transactionStatisticsFilter.value.categoricalChartEndTime);
            }
        } else if (analysisType === StatisticsAnalysisType.TrendAnalysis) {
            querys.push('chartType=' + transactionStatisticsFilter.value.trendChartType);
            querys.push('chartDateType=' + transactionStatisticsFilter.value.trendChartDateType);

            if (transactionStatisticsFilter.value.trendChartDateType === DateRange.Custom.type) {
                querys.push('startTime=' + transactionStatisticsFilter.value.trendChartStartYearMonth);
                querys.push('endTime=' + transactionStatisticsFilter.value.trendChartEndYearMonth);
            }

            if (trendDateAggregationType !== ChartDateAggregationType.Default.type) {
                querys.push('trendDateAggregationType=' + trendDateAggregationType);
            }
        } else if (analysisType === StatisticsAnalysisType.AssetTrends) {
            querys.push('chartType=' + transactionStatisticsFilter.value.assetTrendsChartType);
            querys.push('chartDateType=' + transactionStatisticsFilter.value.assetTrendsChartDateType);

            if (transactionStatisticsFilter.value.assetTrendsChartDateType === DateRange.Custom.type) {
                querys.push('startTime=' + transactionStatisticsFilter.value.assetTrendsChartStartTime);
                querys.push('endTime=' + transactionStatisticsFilter.value.assetTrendsChartEndTime);
            }

            if (assetTrendsDateAggregationType !== ChartDateAggregationType.Default.type) {
                querys.push('assetTrendsDateAggregationType=' + assetTrendsDateAggregationType);
            }
        }

        if (transactionStatisticsFilter.value.filterAccountIds) {
            const ids = objectFieldToArrayItem(transactionStatisticsFilter.value.filterAccountIds);

            if (ids && ids.length) {
                querys.push('filterAccountIds=' + ids.join(','));
            }
        }

        if (transactionStatisticsFilter.value.filterCategoryIds) {
            const ids = objectFieldToArrayItem(transactionStatisticsFilter.value.filterCategoryIds);

            if (ids && ids.length) {
                querys.push('filterCategoryIds=' + ids.join(','));
            }
        }

        if (transactionStatisticsFilter.value.tagFilter) {
            querys.push('tagFilter=' + transactionStatisticsFilter.value.tagFilter);
        }

        if (transactionStatisticsFilter.value.keyword) {
            querys.push('keyword=' + encodeURIComponent(transactionStatisticsFilter.value.keyword));
            querys.push('matchMode=' + transactionStatisticsFilter.value.matchMode);
        }

        querys.push('sortingType=' + transactionStatisticsFilter.value.sortingType);

        return querys.join('&');
    }

    function getTransactionListPageParams(analysisType: StatisticsAnalysisType, itemId: string, dateRange?: TimeRangeAndDateType): string {
        const querys: string[] = [];

        if (transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeByAccount.type
            || transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeByPrimaryCategory.type
            || transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeBySecondaryCategory.type
            || transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalIncome.type) {
            querys.push('type=2');
        } else if (transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseByAccount.type
            || transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseByPrimaryCategory.type
            || transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseBySecondaryCategory.type
            || transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalExpense.type) {
            querys.push('type=3');
        }

        if (itemId && transactionStatisticsFilter.value.chartDataType === ChartDataType.Overview.type) {
            const items = itemId.split('-');
            const sourceItems = (items[0] || '').split(':');
            const queryAccountIds: string[] = [];
            const queryCategoryIds: string[] = [];

            if (sourceItems.length === 2) {
                if (sourceItems[0] === 'account') {
                    queryAccountIds.push(sourceItems[1] as string);
                } else if (sourceItems[0] === 'category') {
                    queryCategoryIds.push(sourceItems[1] as string);
                }
            }

            if (items.length === 2) {
                const targetItems = (items[1] || '').split(':');

                if (targetItems.length === 2) {
                    if (targetItems[0] === 'account') {
                        queryAccountIds.push(targetItems[1] as string);
                    } else if (targetItems[0] === 'category') {
                        queryCategoryIds.push(targetItems[1] as string);
                    }
                }
            }

            if (queryAccountIds.length) {
                if (queryAccountIds.length === 2) {
                    querys.push('type=4');
                }

                querys.push('accountIds=' + queryAccountIds.join(','));
            } else {
                querys.push('accountIds=' + getFinalAccountIdsByFilteredAccountIds(accountsStore.allAccountsMap, transactionStatisticsFilter.value.filterAccountIds));
            }

            if (queryCategoryIds.length) {
                querys.push('categoryIds=' + queryCategoryIds.join(','));
            } else {
                querys.push('categoryIds=' + getFinalCategoryIdsByFilteredCategoryIds(transactionCategoriesStore.allTransactionCategoriesMap, transactionStatisticsFilter.value.filterCategoryIds));
            }
        } else if (itemId && (transactionStatisticsFilter.value.chartDataType === ChartDataType.InflowsByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.OutflowsByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalAssets.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalLiabilities.type)
        ) {
            querys.push('accountIds=' + itemId);

            if ((analysisType === StatisticsAnalysisType.CategoricalAnalysis || analysisType === StatisticsAnalysisType.TrendAnalysis) && !isObjectEmpty(transactionStatisticsFilter.value.filterCategoryIds)) {
                querys.push('categoryIds=' + getFinalCategoryIdsByFilteredCategoryIds(transactionCategoriesStore.allTransactionCategoriesMap, transactionStatisticsFilter.value.filterCategoryIds));
            }
        } else if (itemId && (transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalAssetsByCurrency.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalLiabilitiesByCurrency.type)
        ) {
            const accountIds: string[] = [];

            for (const account of accountsStore.allPlainAccounts) {
                if (account.currency !== itemId || (transactionStatisticsFilter.value.filterAccountIds && transactionStatisticsFilter.value.filterAccountIds[account.id])) {
                    continue;
                }

                if (transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalAssetsByCurrency.type && !account.isAsset) {
                    continue;
                } else if (transactionStatisticsFilter.value.chartDataType === ChartDataType.TotalLiabilitiesByCurrency.type && !account.isLiability) {
                    continue;
                }

                accountIds.push(account.id);
            }

            querys.push('accountIds=' + accountIds.join(','));
        } else if (itemId && (transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeByPrimaryCategory.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeBySecondaryCategory.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseByPrimaryCategory.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseBySecondaryCategory.type)
        ) {
            querys.push('categoryIds=' + itemId);

            if (!isObjectEmpty(transactionStatisticsFilter.value.filterAccountIds)) {
                querys.push('accountIds=' + getFinalAccountIdsByFilteredAccountIds(accountsStore.allAccountsMap, transactionStatisticsFilter.value.filterAccountIds));
            }
        } else if (!itemId) {
            if (!isObjectEmpty(transactionStatisticsFilter.value.filterCategoryIds)) {
                querys.push('categoryIds=' + getFinalCategoryIdsByFilteredCategoryIds(transactionCategoriesStore.allTransactionCategoriesMap, transactionStatisticsFilter.value.filterCategoryIds));
            }

            if (!isObjectEmpty(transactionStatisticsFilter.value.filterAccountIds)) {
                querys.push('accountIds=' + getFinalAccountIdsByFilteredAccountIds(accountsStore.allAccountsMap, transactionStatisticsFilter.value.filterAccountIds));
            }
        }

        if (analysisType === StatisticsAnalysisType.CategoricalAnalysis || analysisType === StatisticsAnalysisType.TrendAnalysis) {
            if (transactionStatisticsFilter.value.tagFilter) {
                querys.push('tagFilter=' + transactionStatisticsFilter.value.tagFilter);
            }

            if (transactionStatisticsFilter.value.keyword) {
                querys.push('keyword=' + encodeURIComponent(transactionStatisticsFilter.value.keyword));
                querys.push('matchMode=' + transactionStatisticsFilter.value.matchMode);
            }
        }

        if (analysisType === StatisticsAnalysisType.CategoricalAnalysis
            && transactionStatisticsFilter.value.chartDataType !== ChartDataType.AccountTotalAssets.type
            && transactionStatisticsFilter.value.chartDataType !== ChartDataType.AccountTotalLiabilities.type
            && transactionStatisticsFilter.value.chartDataType !== ChartDataType.TotalAssetsByCurrency.type
            && transactionStatisticsFilter.value.chartDataType !== ChartDataType.TotalLiabilitiesByCurrency.type) {
            querys.push('dateType=' + transactionStatisticsFilter.value.categoricalChartDateType);

            if (transactionStatisticsFilter.value.categoricalChartDateType === DateRange.Custom.type) {
                querys.push('minTime=' + transactionStatisticsFilter.value.categoricalChartStartTime);
                querys.push('maxTime=' + transactionStatisticsFilter.value.categoricalChartEndTime);
            }
        } else if ((analysisType === StatisticsAnalysisType.TrendAnalysis || analysisType === StatisticsAnalysisType.AssetTrends) && dateRange) {
            querys.push('dateType=' + dateRange.dateType);
            querys.push('minTime=' + dateRange.minTime);
            querys.push('maxTime=' + dateRange.maxTime);
        }

        return querys.join('&');
    }

    return { getTransactionStatisticsPageParams, getTransactionListPageParams };
}
