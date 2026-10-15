import type { SetupProps } from '@/lib/ui/props.ts';
import { type TransactionInsightDataItem, Transaction } from '@/models/transaction.ts';
import { watch } from 'vue';
import { onBeforeRouteUpdate } from 'vue-router';
import type { useExplorerPageLoading } from './loading.ts';
import type { useExplorerPageState } from './state.ts';
import { type InsightsExplorerProps } from './types.ts';

interface Context {
    editDialog: ReturnType<typeof useExplorerPageState>['editDialog'];
    snackbar: ReturnType<typeof useExplorerPageState>['snackbar'];
    reload: ReturnType<typeof useExplorerPageLoading>['reload'];
    init: ReturnType<typeof useExplorerPageLoading>['init'];
    activeTab: ReturnType<typeof useExplorerPageState>['activeTab'];
    initing: ReturnType<typeof useExplorerPageState>['initing'];
    loading: ReturnType<typeof useExplorerPageState>['loading'];
    router: ReturnType<typeof useExplorerPageState>['router'];
    getFilterLinkUrl: ReturnType<typeof useExplorerPageState>['getFilterLinkUrl'];
    props: SetupProps<InsightsExplorerProps>;
}

export function useExplorerPageLifecycle(context: Context) {
    const { editDialog, snackbar, reload, init, activeTab, initing, loading, router, getFilterLinkUrl, props } = context;

    function onShowTransaction(transaction: TransactionInsightDataItem): void {
        editDialog.value?.open({
            id: transaction.id,
            currentTransaction: Transaction.of(transaction)
        }).then(result => {
            if (result && result.message) {
                snackbar.value?.showMessage(result.message);
            }

            reload(false);
        }).catch(error => {
            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }

    function onUpdateTransactions(): void {
        reload(false);
    }

    function onShowDateRangeError(message: string): void {
        snackbar.value?.showError(message);
    }

    onBeforeRouteUpdate((to) => {
        if (to.query) {
            init({
                initId: (to.query['id'] as string | null) || undefined,
                initActiveTab: (to.query['activeTab'] as string | null) || undefined,
                initDateRangeType: (to.query['dateRangeType'] as string | null) || undefined,
                initStartTime: (to.query['startTime'] as string | null) || undefined,
                initEndTime: (to.query['endTime'] as string | null) || undefined,
            });
        } else {
            init({});
        }
    });

    watch(activeTab, () => {
        if (initing.value || loading.value) {
            return;
        }

        router.push(getFilterLinkUrl());
    });

    init(props);

    return { onShowTransaction, onUpdateTransactions, onShowDateRangeError };
}
