import {
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import type { useTransactionPageState } from './state.ts';

interface Context {
    showFilterAccountDialog: ReturnType<typeof useTransactionPageState>['showFilterAccountDialog'];
    loading: ReturnType<typeof useTransactionPageState>['loading'];
    statisticsStore: ReturnType<typeof useTransactionPageState>['statisticsStore'];
    router: ReturnType<typeof useTransactionPageState>['router'];
    getFilterLinkUrl: ReturnType<typeof useTransactionPageState>['getFilterLinkUrl'];
    showFilterCategoryDialog: ReturnType<typeof useTransactionPageState>['showFilterCategoryDialog'];
    showFilterTagDialog: ReturnType<typeof useTransactionPageState>['showFilterTagDialog'];
    analysisType: ReturnType<typeof useTransactionPageState>['analysisType'];
    query: ReturnType<typeof useTransactionPageState>['query'];
}

export function useTransactionPageFilters(context: Context) {
    const { showFilterAccountDialog, loading, statisticsStore, router, getFilterLinkUrl, showFilterCategoryDialog, showFilterTagDialog, analysisType, query } = context;

    function setAccountFilter(changed: boolean): void {
        showFilterAccountDialog.value = false;

        if (changed) {
            loading.value = true;
            statisticsStore.updateTransactionStatisticsInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    function setCategoryFilter(changed: boolean): void {
        showFilterCategoryDialog.value = false;

        if (changed) {
            loading.value = true;
            statisticsStore.updateTransactionStatisticsInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    function setTagFilter(changed: boolean): void {
        showFilterTagDialog.value = false;

        if (changed) {
            loading.value = true;
            statisticsStore.updateTransactionStatisticsInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    function setKeywordFilter(keyword: string): void {
        if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            return;
        }

        if (query.value.keyword === keyword) {
            return;
        }

        let changed = false;

        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
            changed = statisticsStore.updateTransactionStatisticsFilter({
                keyword: keyword
            });
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
            changed = statisticsStore.updateTransactionStatisticsFilter({
                keyword: keyword
            });
        }

        if (changed) {
            loading.value = true;
            statisticsStore.updateTransactionStatisticsInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    return { setAccountFilter, setCategoryFilter, setTagFilter, setKeywordFilter };
}
