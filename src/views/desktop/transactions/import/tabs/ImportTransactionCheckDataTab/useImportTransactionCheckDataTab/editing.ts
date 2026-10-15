import { TransactionType } from '@/core/transaction.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    editingTransaction: ReturnType<typeof useImportTransactionCheckDataTabState>['editingTransaction'];
    editingTags: ReturnType<typeof useImportTransactionCheckDataTabState>['editingTags'];
    allCategoriesMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allCategoriesMap'];
    allAccountsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allAccountsMap'];
    allTagsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allTagsMap'];
    props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>;
}

export function useImportTransactionCheckDataTabEditing(context: Context) {
    const { editingTransaction, editingTags, allCategoriesMap, allAccountsMap, allTagsMap, props } = context;

    function editTransaction(transaction: ImportTransaction): void {
        if (editingTransaction.value) {
            editingTransaction.value.tagIds = editingTags.value;
            updateTransactionData(editingTransaction.value);
        }

        if (editingTransaction.value === transaction) {
            editingTags.value = [];
            editingTransaction.value = null;
        } else {
            editingTransaction.value = transaction;
            editingTags.value = editingTransaction.value.tagIds;
        }
    }

    function updateTransactionData(transaction: ImportTransaction): void {
        transaction.valid = transaction.isTransactionValid();

        if (transaction.categoryId && allCategoriesMap.value[transaction.categoryId]) {
            transaction.actualCategoryName = allCategoriesMap.value[transaction.categoryId]!.name;
        } else {
            if (transaction.type !== TransactionType.ModifyBalance) {
                transaction.valid = false;
            }
        }

        if (transaction.sourceAccountId && allAccountsMap.value[transaction.sourceAccountId]) {
            transaction.actualSourceAccountName = allAccountsMap.value[transaction.sourceAccountId]!.name;
        } else {
            transaction.valid = false;
        }

        if (transaction.destinationAccountId && allAccountsMap.value[transaction.destinationAccountId]) {
            transaction.actualDestinationAccountName = allAccountsMap.value[transaction.destinationAccountId]!.name;
        } else {
            if (transaction.type === TransactionType.Transfer) {
                transaction.valid = false;
            }
        }

        if (transaction.tagIds && transaction.tagIds.length) {
            for (const tagId of transaction.tagIds) {
                if (!tagId || !allTagsMap.value[tagId]) {
                    transaction.valid = false;
                    break;
                }
            }
        }
    }

    function updateAllTransactionsIsValid(): void {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return;
        }

        for (const importTransaction of props.importTransactions) {
            updateTransactionData(importTransaction);
        }
    }

    return { editTransaction, updateTransactionData, updateAllTransactionsIsValid };
}
