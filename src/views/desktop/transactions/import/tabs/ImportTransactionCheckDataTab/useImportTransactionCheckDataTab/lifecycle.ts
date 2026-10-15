import { focusTableScrollContainer } from '@/lib/ui/desktop.ts';
import { watch } from 'vue';
import type { useImportTransactionCheckDataTabSelectionActions } from './selection-actions.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    isEditing: ReturnType<typeof useImportTransactionCheckDataTabState>['isEditing'];
    snackbar: ReturnType<typeof useImportTransactionCheckDataTabState>['snackbar'];
    editingTransaction: ReturnType<typeof useImportTransactionCheckDataTabState>['editingTransaction'];
    editingTags: ReturnType<typeof useImportTransactionCheckDataTabState>['editingTags'];
    filters: ReturnType<typeof useImportTransactionCheckDataTabState>['filters'];
    currentPage: ReturnType<typeof useImportTransactionCheckDataTabState>['currentPage'];
    countPerPage: ReturnType<typeof useImportTransactionCheckDataTabState>['countPerPage'];
    clearSelectedTransactionsNotDisplayed: ReturnType<typeof useImportTransactionCheckDataTabSelectionActions>['clearSelectedTransactionsNotDisplayed'];
}

export function useImportTransactionCheckDataTabLifecycle(context: Context) {
    const { isEditing, snackbar, editingTransaction, editingTags, filters, currentPage, countPerPage, clearSelectedTransactionsNotDisplayed } = context;

    function focusTableScrollContainerWhenNonEditing(event: MouseEvent): void {
        if (isEditing.value) {
            return;
        }

        focusTableScrollContainer(event);
    }

    function onShowDateRangeError(message: string): void {
        snackbar.value?.showError(message);
    }

    function reset(): void {
        editingTransaction.value = null;
        editingTags.value = [];
        filters.value.minDatetime = null;
        filters.value.maxDatetime = null;
        filters.value.transactionType = null;
        filters.value.category = null;
        filters.value.account = null;
        filters.value.tag = null;
        filters.value.description = null;
        currentPage.value = 1;
        countPerPage.value = 10;
    }

    function setCountPerPage(count: number): void {
        countPerPage.value = count;
    }

    watch(filters, () => {
        clearSelectedTransactionsNotDisplayed();
        currentPage.value = 1;
    }, {
        deep: true
    });

    return { focusTableScrollContainerWhenNonEditing, onShowDateRangeError, reset, setCountPerPage };
}
