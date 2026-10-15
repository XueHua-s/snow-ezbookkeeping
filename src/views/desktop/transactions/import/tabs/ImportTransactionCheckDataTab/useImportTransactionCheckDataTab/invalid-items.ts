import { type NameValue, itemAndIndex, keys } from '@/core/base.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    isString
} from '@/lib/common.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import { computed } from 'vue';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>;
    allCategoriesMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allCategoriesMap'];
    tt: ReturnType<typeof useImportTransactionCheckDataTabState>['tt'];
    allAccountsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allAccountsMap'];
    allTagsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allTagsMap'];
}

export function useImportTransactionCheckDataTabInvalidItems(context: Context) {
    const { props, allCategoriesMap, tt, allAccountsMap, allTagsMap } = context;

    const allInvalidExpenseCategoryNames = computed<NameValue[]>(() => getCurrentInvalidCategoryNames(TransactionType.Expense));

    const allInvalidIncomeCategoryNames = computed<NameValue[]>(() => getCurrentInvalidCategoryNames(TransactionType.Income));

    const allInvalidTransferCategoryNames = computed<NameValue[]>(() => getCurrentInvalidCategoryNames(TransactionType.Transfer));

    const allInvalidAccountNames = computed<NameValue[]>(() => getCurrentInvalidAccountNames());

    const allInvalidTransactionTagNames = computed<NameValue[]>(() => getCurrentInvalidTagNames());

    function getCurrentInvalidCategoryNames(transactionType: TransactionType): NameValue[] {
        const invalidCategoryNames: Record<string, boolean> = {};
        const invalidCategories: NameValue[] = [];

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return invalidCategories;
        }

        for (const importTransaction of props.importTransactions) {
            const categoryId = importTransaction.categoryId;

            if (importTransaction.type === transactionType && (!categoryId || categoryId === '0' || !allCategoriesMap.value[categoryId])) {
                invalidCategoryNames[importTransaction.originalCategoryName] = true;
            }
        }

        for (const name of keys(invalidCategoryNames)) {
            invalidCategories.push({
                name: name || tt('(Empty)'),
                value: name
            });
        }

        return invalidCategories;
    }

    function getCurrentInvalidAccountNames(): NameValue[] {
        const invalidAccountNames: Record<string, boolean> = {};
        const invalidAccounts: NameValue[] = [];

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return invalidAccounts;
        }

        for (const importTransaction of props.importTransactions) {
            const sourceAccountId = importTransaction.sourceAccountId;
            const destinationAccountId = importTransaction.destinationAccountId;

            if (!sourceAccountId || sourceAccountId === '0' || !allAccountsMap.value[sourceAccountId]) {
                invalidAccountNames[importTransaction.originalSourceAccountName] = true;
            }

            if (importTransaction.type === TransactionType.Transfer && isString(importTransaction.originalDestinationAccountName) && (!destinationAccountId || destinationAccountId === '0' || !allAccountsMap.value[destinationAccountId])) {
                invalidAccountNames[importTransaction.originalDestinationAccountName] = true;
            }
        }

        for (const name of keys(invalidAccountNames)) {
            invalidAccounts.push({
                name: name || tt('(Empty)'),
                value: name
            });
        }

        return invalidAccounts;
    }

    function getCurrentInvalidTagNames(): NameValue[] {
        const invalidTagNames: Record<string, boolean> = {};
        const invalidTags: NameValue[] = [];

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return invalidTags;
        }

        for (const importTransaction of props.importTransactions) {
            if (!importTransaction.tagIds || !importTransaction.originalTagNames) {
                continue;
            }

            for (const [tagId, tagIndex] of itemAndIndex(importTransaction.tagIds)) {
                const originalTagName = importTransaction.originalTagNames[tagIndex] as string | undefined;

                if (!originalTagName) {
                    continue;
                }

                if (!tagId || tagId === '0' || !allTagsMap.value[tagId]) {
                    invalidTagNames[originalTagName] = true;
                }
            }
        }

        for (const name of keys(invalidTagNames)) {
            invalidTags.push({
                name: name || tt('(Empty)'),
                value: name
            });
        }

        return invalidTags;
    }

    return { allInvalidExpenseCategoryNames, allInvalidIncomeCategoryNames, allInvalidTransferCategoryNames, allInvalidAccountNames, allInvalidTransactionTagNames };
}
