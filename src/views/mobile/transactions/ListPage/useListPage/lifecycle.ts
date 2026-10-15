import { scrollToSelectedItem } from '@/lib/ui/common.ts';
import {
    type Framework7Dom,
    onInfiniteScrolling
} from '@/lib/ui/mobile.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import type { Router } from 'framework7/types';
import { onMounted, onUnmounted } from 'vue';
import type { useListPageLoading } from './loading.ts';
import type { useListPageState } from './state.ts';
import type { useListPageVirtualList } from './virtual-list.ts';

interface Context {
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    loading: ReturnType<typeof useListPageState>['loading'];
    reload: ReturnType<typeof useListPageLoading>['reload'];
    routeBackOnError: ReturnType<typeof useListPageState>['routeBackOnError'];
    props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>;
    loadingError: ReturnType<typeof useListPageState>['loadingError'];
    onResize: ReturnType<typeof useListPageVirtualList>['onResize'];
    onScroll: ReturnType<typeof useListPageVirtualList>['onScroll'];
    init: ReturnType<typeof useListPageLoading>['init'];
}

export function useListPageLifecycle(context: Context) {
    const { transactionsStore, loading, reload, routeBackOnError, props, loadingError, onResize, onScroll, init } = context;

    function onPopoverOpen(event: { $el: Framework7Dom }): void {
        scrollToSelectedItem(event.$el[0], '.popover-inner', '.popover-inner', 'li.list-item-selected');
    }

    function onCategoryPopoverOpen(event: { $el: Framework7Dom }): void {
        scrollToSelectedItem(event.$el[0], '.popover-inner', '.popover-inner', 'li.list-item-checked');
    }

    function onPageAfterIn(): void {
        if (transactionsStore.transactionListStateInvalid && !loading.value) {
            reload();
        }

        routeBackOnError(props.f7router, loadingError);
    }

    onMounted(() => {
        window.addEventListener('resize', onResize);
        onInfiniteScrolling(onScroll);
    });

    onUnmounted(() => {
        window.removeEventListener('resize', onResize);
    });

    init();

    return { onPopoverOpen, onCategoryPopoverOpen, onPageAfterIn };
}
