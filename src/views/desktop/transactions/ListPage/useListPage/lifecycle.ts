import type { SetupProps } from '@/lib/ui/props.ts';
import { watch } from 'vue';
import { onBeforeRouteUpdate } from 'vue-router';
import type { useListPageInitialization } from './initialization.ts';
import type { useListPageLoading } from './loading.ts';
import type { useListPageState } from './state.ts';
import { type TransactionListProps } from './types.ts';

interface Context {
    snackbar: ReturnType<typeof useListPageState>['snackbar'];
    init: ReturnType<typeof useListPageInitialization>['init'];
    desktopPageStore: ReturnType<typeof useListPageState>['desktopPageStore'];
    add: ReturnType<typeof useListPageLoading>['add'];
    props: SetupProps<TransactionListProps>;
}

export function useListPageLifecycle(context: Context) {
    const { snackbar, init, desktopPageStore, add, props } = context;

    function onShowDateRangeError(message: string): void {
        snackbar.value?.showError(message);
    }

    onBeforeRouteUpdate((to) => {
        if (to.query) {
            init({
                initDateType: (to.query['dateType'] as string | null) || undefined,
                initMinTime: (to.query['minTime'] as string | null) || undefined,
                initMaxTime: (to.query['maxTime'] as string | null) || undefined,
                initType: (to.query['type'] as string | null) || undefined,
                initCategoryIds: (to.query['categoryIds'] as string | null) || undefined,
                initAccountIds: (to.query['accountIds'] as string | null) || undefined,
                initTagFilter: (to.query['tagFilter'] as string | null) || undefined,
                initAmountFilter: (to.query['amountFilter'] as string | null) || undefined,
                initKeyword: (to.query['keyword'] as string | null) || undefined,
                initMatchMode: (to.query['matchMode'] as string | null) || undefined
            });
        } else {
            init({});
        }
    });

    watch(() => desktopPageStore.showAddTransactionDialogInTransactionList, (newValue) => {
        if (newValue) {
            desktopPageStore.resetShowAddTransactionDialogInTransactionList();
            add();
        }
    });

    init(props);

    return { onShowDateRangeError };
}
