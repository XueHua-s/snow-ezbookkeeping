import {
    ChartDataType,
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import type { useTransactionPageState } from './state.ts';

interface Context {
    statisticsStore: ReturnType<typeof useTransactionPageState>['statisticsStore'];
    analysisType: ReturnType<typeof useTransactionPageState>['analysisType'];
    accountsStore: ReturnType<typeof useTransactionPageState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useTransactionPageState>['transactionCategoriesStore'];
    loading: ReturnType<typeof useTransactionPageState>['loading'];
    loadingError: ReturnType<typeof useTransactionPageState>['loadingError'];
    showToast: ReturnType<typeof useTransactionPageState>['showToast'];
    reloading: ReturnType<typeof useTransactionPageState>['reloading'];
    query: ReturnType<typeof useTransactionPageState>['query'];
}

export function useTransactionPageLoading(context: Context) {
    const { statisticsStore, analysisType, accountsStore, transactionCategoriesStore, loading, loadingError, showToast, reloading, query } = context;

    function init(): void {
        statisticsStore.initTransactionStatisticsFilter(analysisType.value);

        Promise.all([
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false })
        ]).then(() => {
            if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
                return statisticsStore.loadCategoricalAnalysis({
                    force: false
                }) as Promise<unknown>;
            } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
                return statisticsStore.loadTrendAnalysis({
                    force: false
                }) as Promise<unknown>;
            } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
                return statisticsStore.loadAssetTrends({
                    force: false
                }) as Promise<unknown>;
            } else {
                return Promise.reject('An error occurred');
            }
        }).then(() => {
            loading.value = false;
        }).catch(error => {
            if (error.processed) {
                loading.value = false;
            } else {
                loadingError.value = error;
                showToast(error.message || error);
            }
        });
    }

    function reload(done?: () => void): void {
        const force = !!done;
        let dispatchPromise: Promise<unknown> | null = null;

        reloading.value = true;

        if (query.value.chartDataType === ChartDataType.OutflowsByAccount.type ||
            query.value.chartDataType === ChartDataType.ExpenseByAccount.type ||
            query.value.chartDataType === ChartDataType.ExpenseByPrimaryCategory.type ||
            query.value.chartDataType === ChartDataType.ExpenseBySecondaryCategory.type ||
            query.value.chartDataType === ChartDataType.InflowsByAccount.type ||
            query.value.chartDataType === ChartDataType.IncomeByAccount.type ||
            query.value.chartDataType === ChartDataType.IncomeByPrimaryCategory.type ||
            query.value.chartDataType === ChartDataType.IncomeBySecondaryCategory.type ||
            query.value.chartDataType === ChartDataType.TotalOutflows.type ||
            query.value.chartDataType === ChartDataType.TotalExpense.type ||
            query.value.chartDataType === ChartDataType.TotalInflows.type ||
            query.value.chartDataType === ChartDataType.TotalIncome.type ||
            query.value.chartDataType === ChartDataType.NetCashFlow.type ||
            query.value.chartDataType === ChartDataType.NetIncome.type ||
            query.value.chartDataType === ChartDataType.NetWorth.type) {
            if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
                dispatchPromise = statisticsStore.loadCategoricalAnalysis({
                    force: force
                });
            } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
                dispatchPromise = statisticsStore.loadTrendAnalysis({
                    force: force
                });
            } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
                dispatchPromise = statisticsStore.loadAssetTrends({
                    force: force
                });
            }
        } else if (query.value.chartDataType === ChartDataType.AccountTotalAssets.type ||
            query.value.chartDataType === ChartDataType.AccountTotalLiabilities.type ||
            query.value.chartDataType === ChartDataType.TotalAssetsByCurrency.type ||
            query.value.chartDataType === ChartDataType.TotalLiabilitiesByCurrency.type) {
            if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
                dispatchPromise = accountsStore.loadAllAccounts({
                    force: force
                });
            } else if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
                dispatchPromise = statisticsStore.loadAssetTrends({
                    force: force
                });
            }
        }

        if (dispatchPromise) {
            dispatchPromise.then(() => {
                reloading.value = false;

                done?.();

                if (force) {
                    showToast('Data has been updated');
                }
            }).catch(error => {
                reloading.value = false;

                done?.();

                if (!error.processed) {
                    showToast(error.message || error);
                }
            });
        } else {
            reloading.value = false;
        }
    }

    return { init, reload };
}
