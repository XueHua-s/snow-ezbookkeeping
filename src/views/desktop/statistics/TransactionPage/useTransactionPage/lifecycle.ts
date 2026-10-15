import { type TextualYearMonth, type TimeRangeAndDateType } from '@/core/datetime.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { onBeforeRouteUpdate } from 'vue-router';
import type { useTransactionPageInitialization } from './initialization.ts';
import type { useTransactionPageState } from './state.ts';
import { type TransactionStatisticsProps } from './types.ts';

interface Context {
    transactionCategoriesStore: ReturnType<typeof useTransactionPageState>['transactionCategoriesStore'];
    router: ReturnType<typeof useTransactionPageState>['router'];
    getTransactionItemLinkUrl: ReturnType<typeof useTransactionPageState>['getTransactionItemLinkUrl'];
    snackbar: ReturnType<typeof useTransactionPageState>['snackbar'];
    init: ReturnType<typeof useTransactionPageInitialization>['init'];
    props: SetupProps<TransactionStatisticsProps>;
}

export function useTransactionPageLifecycle(context: Context) {
    const { transactionCategoriesStore, router, getTransactionItemLinkUrl, snackbar, init, props } = context;

    function onClickSankeyChartItem(sourceItemType: 'account' | 'category', sourceItemId: string, targetItemType?: 'account' | 'category', targetItemId?: string): void {
        if (sourceItemType === 'category' && targetItemType === 'category' && sourceItemId && targetItemId) {
            const sourceCategory = transactionCategoriesStore.allTransactionCategoriesMap[sourceItemId];
            const targetCategory = transactionCategoriesStore.allTransactionCategoriesMap[targetItemId];

            if (sourceCategory?.parentId === targetCategory?.id) {
                router.push(getTransactionItemLinkUrl(`${sourceItemType}:${sourceItemId}`));
                return;
            } else if (targetCategory?.parentId === sourceCategory?.id) {
                router.push(getTransactionItemLinkUrl(`${targetItemType}:${targetItemId}`));
                return;
            }
        }

        router.push(getTransactionItemLinkUrl(`${sourceItemType}:${sourceItemId}` + (targetItemType && targetItemId ? `-${targetItemType}:${targetItemId}` : '')));
    }

    function onClickPieChartItem(item: Record<string, unknown>): void {
        router.push(getTransactionItemLinkUrl(item['id'] as string));
    }

    function onClickTrendChartItem(item: { itemId: string, dateRange: TimeRangeAndDateType }): void {
        router.push(getTransactionItemLinkUrl(item.itemId, item.dateRange));
    }

    function onShowDateRangeError(message: string): void {
        snackbar.value?.showError(message);
    }

    onBeforeRouteUpdate((to) => {
        if (to.query) {
            init({
                initAnalysisType: (to.query['analysisType'] as string | null) || undefined,
                initChartDataType: (to.query['chartDataType'] as string | null) || undefined,
                initChartType: (to.query['chartType'] as string | null) || undefined,
                initChartDateType: (to.query['chartDateType'] as string | null) || undefined,
                initStartTime: (to.query['startTime'] as TextualYearMonth | null) || undefined,
                initEndTime: (to.query['endTime'] as TextualYearMonth | null) || undefined,
                initFilterAccountIds: (to.query['filterAccountIds'] as string | null) || undefined,
                initFilterCategoryIds: (to.query['filterCategoryIds'] as string | null) || undefined,
                initTagFilter: (to.query['tagFilter'] as string | null) || undefined,
                initKeyword: (to.query['keyword'] as string | null) || undefined,
                initMatchMode: (to.query['matchMode'] as string | null) || undefined,
                initSortingType: (to.query['sortingType'] as string | null) || undefined,
                initTrendDateAggregationType: (to.query['trendDateAggregationType'] as string | null) || undefined,
                initAssetTrendsDateAggregationType: (to.query['assetTrendsDateAggregationType'] as string | null) || undefined
            });
        } else {
            init({});
        }
    });

    init(props);

    return { onClickSankeyChartItem, onClickPieChartItem, onClickTrendChartItem, onShowDateRangeError };
}
