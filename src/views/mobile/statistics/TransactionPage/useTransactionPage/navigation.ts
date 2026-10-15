import { type CategoricalChartSourceDataItem } from '@/core/chart.ts';
import { type TimeRangeAndDateType } from '@/core/datetime.ts';
import {
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import { scrollToSelectedItem } from '@/lib/ui/common.ts';
import { type Framework7Dom } from '@/lib/ui/mobile.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import type { Router } from 'framework7/types';
import type { useTransactionPageLoading } from './loading.ts';
import type { useTransactionPageState } from './state.ts';

interface Context {
    statisticsStore: ReturnType<typeof useTransactionPageState>['statisticsStore'];
    analysisType: ReturnType<typeof useTransactionPageState>['analysisType'];
    props: SetupProps<{
    f7router: Router.Router;
}>;
    showPrompt: ReturnType<typeof useTransactionPageState>['showPrompt'];
    query: ReturnType<typeof useTransactionPageState>['query'];
    reload: ReturnType<typeof useTransactionPageLoading>['reload'];
    loading: ReturnType<typeof useTransactionPageState>['loading'];
    routeBackOnError: ReturnType<typeof useTransactionPageState>['routeBackOnError'];
    loadingError: ReturnType<typeof useTransactionPageState>['loadingError'];
    init: ReturnType<typeof useTransactionPageLoading>['init'];
}

export function useTransactionPageNavigation(context: Context) {
    const { statisticsStore, analysisType, props, showPrompt, query, reload, loading, routeBackOnError, loadingError, init } = context;

    function getTransactionItemLinkUrl(itemId: string, dateRange?: TimeRangeAndDateType): string {
        return `/transaction/list?${statisticsStore.getTransactionListPageParams(analysisType.value, itemId, dateRange)}`;
    }

    function filterAccounts(): void {
        props.f7router.navigate('/settings/filter/account?type=statisticsCurrent');
    }

    function filterCategories(): void {
        props.f7router.navigate('/settings/filter/category?type=statisticsCurrent');
    }

    function filterTags(): void {
        props.f7router.navigate('/settings/filter/tag?type=statisticsCurrent');
    }

    function filterDescription(): void {
        if (analysisType.value === StatisticsAnalysisType.AssetTrends) {
            return;
        }

        showPrompt('Filter transaction description', query.value.keyword, value => {
            if (query.value.keyword === value) {
                return;
            }

            let changed = false;

            if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis) {
                changed = statisticsStore.updateTransactionStatisticsFilter({
                    keyword: value
                });
            } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis) {
                changed = statisticsStore.updateTransactionStatisticsFilter({
                    keyword: value
                });
            }

            if (changed) {
                reload();
            }
        });
    }

    function settings(): void {
        props.f7router.navigate('/statistic/settings');
    }

    function scrollPopoverToSelectedItem(event: { $el: Framework7Dom }): void {
        scrollToSelectedItem(event.$el[0], '.popover-inner', '.popover-inner', 'li.list-item-selected');
    }

    function onClickPieChartItem(item: CategoricalChartSourceDataItem): void {
        if (item.id) {
            props.f7router.navigate(getTransactionItemLinkUrl(item.id));
        }
    }

    function onClickTrendChartItem(item: { itemId: string, dateRange: TimeRangeAndDateType }): void {
        props.f7router.navigate(getTransactionItemLinkUrl(item.itemId, item.dateRange));
    }

    function onPageAfterIn(): void {
        if (statisticsStore.transactionStatisticsStateInvalid && !loading.value) {
            reload();
        }

        routeBackOnError(props.f7router, loadingError);
    }

    init();

    return {
        getTransactionItemLinkUrl, filterAccounts, filterCategories, filterTags, filterDescription, settings,
        scrollPopoverToSelectedItem, onClickPieChartItem, onClickTrendChartItem, onPageAfterIn,
    };
}
