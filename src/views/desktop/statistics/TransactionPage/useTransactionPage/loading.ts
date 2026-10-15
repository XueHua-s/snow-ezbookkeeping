import {
    ChartDataType,
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import type { useTransactionPageState } from './state.ts';

interface Context {
    loading: ReturnType<typeof useTransactionPageState>['loading'];
    query: ReturnType<typeof useTransactionPageState>['query'];
    analysisType: ReturnType<typeof useTransactionPageState>['analysisType'];
    statisticsStore: ReturnType<typeof useTransactionPageState>['statisticsStore'];
    accountsStore: ReturnType<typeof useTransactionPageState>['accountsStore'];
    snackbar: ReturnType<typeof useTransactionPageState>['snackbar'];
}

export function useTransactionPageLoading(context: Context) {
    const { loading, query, analysisType, statisticsStore, accountsStore, snackbar } = context;

    function reload(force: boolean): Promise<unknown> | null {
        let dispatchPromise: Promise<unknown> | null = null;

        loading.value = true;

        if (query.value.chartDataType === ChartDataType.Overview.type ||
            query.value.chartDataType === ChartDataType.OutflowsByAccount.type ||
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
                loading.value = false;

                if (force) {
                    snackbar.value?.showMessage('Data has been updated');
                }
            }).catch(error => {
                loading.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        }

        return dispatchPromise;
    }

    return { reload };
}
