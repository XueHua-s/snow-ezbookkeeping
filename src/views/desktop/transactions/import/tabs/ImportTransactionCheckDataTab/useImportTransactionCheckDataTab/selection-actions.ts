import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import type { useImportTransactionCheckDataTabFiltering } from './filtering.ts';
import type { useImportTransactionCheckDataTabPagination } from './pagination.ts';

interface Context {
    props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>;
    isTransactionDisplayed: ReturnType<typeof useImportTransactionCheckDataTabFiltering>['isTransactionDisplayed'];
    currentPageTransactions: ReturnType<typeof useImportTransactionCheckDataTabPagination>['currentPageTransactions'];
}

export function useImportTransactionCheckDataTabSelectionActions(context: Context) {
    const { props, isTransactionDisplayed, currentPageTransactions } = context;

    function selectAllValid(): void {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return;
        }

        for (const importTransaction of props.importTransactions) {
            if (importTransaction.valid && isTransactionDisplayed(importTransaction)) {
                importTransaction.selected = true;
            }
        }
    }

    function selectAllInvalid(): void {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return;
        }

        for (const importTransaction of props.importTransactions) {
            if (!importTransaction.valid && isTransactionDisplayed(importTransaction)) {
                importTransaction.selected = true;
            }
        }
    }

    function selectAll(): void {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return;
        }

        for (const importTransaction of props.importTransactions) {
            if (isTransactionDisplayed(importTransaction)) {
                importTransaction.selected = true;
            }
        }
    }

    function selectNone(): void {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return;
        }

        for (const importTransaction of props.importTransactions) {
            if (isTransactionDisplayed(importTransaction)) {
                importTransaction.selected = false;
            }
        }
    }

    function selectInvert(): void {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return;
        }

        for (const importTransaction of props.importTransactions) {
            if (isTransactionDisplayed(importTransaction)) {
                importTransaction.selected = !importTransaction.selected;
            }
        }
    }

    function selectAllInThisPage(): void {
        for (const importTransaction of currentPageTransactions.value) {
            importTransaction.selected = true;
        }
    }

    function selectNoneInThisPage(): void {
        for (const importTransaction of currentPageTransactions.value) {
            importTransaction.selected = false;
        }
    }

    function clearSelectedTransactionsNotDisplayed(): void {
        if (!props.importTransactions) {
            return;
        }

        for (const importTransaction of props.importTransactions) {
            if (importTransaction.selected && !isTransactionDisplayed(importTransaction)) {
                importTransaction.selected = false;
            }
        }
    }

    function selectInvertInThisPage(): void {
        for (const importTransaction of currentPageTransactions.value) {
            importTransaction.selected = !importTransaction.selected;
        }
    }

    return {
        selectAllValid, selectAllInvalid, selectAll, selectNone, selectInvert, selectAllInThisPage, selectNoneInThisPage,
        clearSelectedTransactionsNotDisplayed, selectInvertInThisPage,
    };
}
