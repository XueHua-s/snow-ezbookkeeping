import { TransactionType } from '@/core/transaction.ts';
import {
    objectFieldWithValueToArrayItem
} from '@/lib/common.ts';
import {
    getCurrentUnixTime
} from '@/lib/datetime.ts';
import {
    hideLoading,
    onSwipeoutDeleted,
    showLoading
} from '@/lib/ui/mobile.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { type Transaction } from '@/models/transaction.ts';
import type { Router } from 'framework7/types';
import type { useListPageState } from './state.ts';
import type { useListPageVirtualList } from './virtual-list.ts';

interface Context {
    query: ReturnType<typeof useListPageState>['query'];
    queryAllFilterCategoryIdsCount: ReturnType<typeof useListPageState>['queryAllFilterCategoryIdsCount'];
    queryAllFilterAccountIdsCount: ReturnType<typeof useListPageState>['queryAllFilterAccountIdsCount'];
    queryAllFilterTagIds: ReturnType<typeof useListPageState>['queryAllFilterTagIds'];
    props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>;
    showAlert: ReturnType<typeof useListPageState>['showAlert'];
    transactionToDelete: ReturnType<typeof useListPageState>['transactionToDelete'];
    showDeleteActionSheet: ReturnType<typeof useListPageState>['showDeleteActionSheet'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    selectedAccountDefaultCurrency: ReturnType<typeof useListPageState>['selectedAccountDefaultCurrency'];
    getTransactionDomId: ReturnType<typeof useListPageVirtualList>['getTransactionDomId'];
    showToast: ReturnType<typeof useListPageState>['showToast'];
}

export function useListPageActions(context: Context) {
    const {
        query, queryAllFilterCategoryIdsCount, queryAllFilterAccountIdsCount, queryAllFilterTagIds, props, showAlert,
        transactionToDelete, showDeleteActionSheet, transactionsStore, selectedAccountDefaultCurrency, getTransactionDomId,
        showToast,
    } = context;

    function add(): void {
        const currentUnixTime = getCurrentUnixTime();

        let setTransactionTime = false;
        let newTransactionTime: number | undefined = undefined;

        if (query.value.maxTime && query.value.minTime) {
            if (query.value.maxTime < currentUnixTime) {
                setTransactionTime = true;
                newTransactionTime = query.value.maxTime;
            } else if (currentUnixTime < query.value.minTime) {
                setTransactionTime = true;
                newTransactionTime = query.value.minTime;
            }
        }

        const params: string[] = [];

        if (setTransactionTime) {
            params.push(`time=${newTransactionTime}`);
        }

        if (query.value.type !== TransactionType.ModifyBalance) {
            params.push(`type=${query.value.type}`);
        }

        if (queryAllFilterCategoryIdsCount.value === 1) {
            params.push(`categoryId=${query.value.categoryIds}`);
        }

        if (queryAllFilterAccountIdsCount.value === 1) {
            params.push(`accountId=${query.value.accountIds}`);
        }

        if (query.value.tagFilter) {
            params.push(`tagIds=${objectFieldWithValueToArrayItem(queryAllFilterTagIds.value, true).join(',') || ''}`);
        }

        props.f7router.navigate(`/transaction/add?${params.join('&')}`);
    }

    function duplicate(transaction: Transaction): void {
        props.f7router.navigate(`/transaction/add?id=${transaction.id}&type=${transaction.type}`);
    }

    function edit(transaction: Transaction): void {
        props.f7router.navigate(`/transaction/edit?id=${transaction.id}&type=${transaction.type}`);
    }

    function remove(transaction: Transaction | null, confirm: boolean): void {
        if (!transaction) {
            showAlert('An error occurred');
            return;
        }

        if (!confirm) {
            transactionToDelete.value = transaction;
            showDeleteActionSheet.value = true;
            return;
        }

        showDeleteActionSheet.value = false;
        transactionToDelete.value = null;
        showLoading();

        transactionsStore.deleteTransaction({
            transaction: transaction,
            defaultCurrency: selectedAccountDefaultCurrency.value,
            beforeResolve: (done) => {
                onSwipeoutDeleted(getTransactionDomId(transaction), done);
            }
        }).then(() => {
            hideLoading();
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    return { add, duplicate, edit, remove };
}
