import { type NameValue, keys } from '@/core/base.ts';
import { TransactionType } from '@/core/transaction.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import { computed } from 'vue';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>;
    tt: ReturnType<typeof useImportTransactionCheckDataTabState>['tt'];
}

export function useImportTransactionCheckDataTabOriginalItems(context: Context) {
    const { props, tt } = context;

    const allOriginalExpenseCategoryNames = computed<NameValue[]>(() => getAllOriginalCategoryNames(TransactionType.Expense));

    const allOriginalIncomeCategoryNames = computed<NameValue[]>(() => getAllOriginalCategoryNames(TransactionType.Income));

    const allOriginalTransferCategoryNames = computed<NameValue[]>(() => getAllOriginalCategoryNames(TransactionType.Transfer));

    const allOriginalSourceAccountNames = computed<NameValue[]>(() => getAllOriginalAccountNames(false));

    const allOriginalDestinationAccountNames = computed<NameValue[]>(() => getAllOriginalAccountNames(true));

    const allOriginalTransactionTagNames = computed<NameValue[]>(() => getAllOriginalTagNames());

    function getAllOriginalCategoryNames(transactionType: TransactionType): NameValue[] {
        const categoryNames: Record<string, boolean> = {};
        const categories: NameValue[] = [];

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return categories;
        }

        for (const importTransaction of props.importTransactions) {
            if (importTransaction.type === transactionType) {
                categoryNames[importTransaction.originalCategoryName] = true;
            }
        }

        for (const name of keys(categoryNames)) {
            categories.push({
                name: name || tt('(Empty)'),
                value: name
            });
        }

        return categories;
    }

    function getAllOriginalAccountNames(destination: boolean): NameValue[] {
        const accountNames: Record<string, boolean> = {};
        const accounts: NameValue[] = [];

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return accounts;
        }

        for (const importTransaction of props.importTransactions) {
            if (destination && importTransaction.type === TransactionType.Transfer) {
                accountNames[importTransaction.originalDestinationAccountName || ''] = true;
            } else if (!destination) {
                accountNames[importTransaction.originalSourceAccountName] = true;
            }
        }

        for (const name of keys(accountNames)) {
            accounts.push({
                name: name || tt('(Empty)'),
                value: name
            });
        }

        return accounts;
    }

    function getAllOriginalTagNames(): NameValue[] {
        const allOriginalTagNames: Record<string, boolean> = {};
        const allOriginalTags: NameValue[] = [];

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return allOriginalTags;
        }

        for (const importTransaction of props.importTransactions) {
            if (!importTransaction.originalTagNames) {
                continue;
            }

            for (const tagName of importTransaction.originalTagNames) {
                allOriginalTagNames[tagName] = true;
            }
        }

        for (const name of keys(allOriginalTagNames)) {
            allOriginalTags.push({
                name: name || tt('(Empty)'),
                value: name
            });
        }

        return allOriginalTags;
    }

    return {
        allOriginalExpenseCategoryNames, allOriginalIncomeCategoryNames, allOriginalTransferCategoryNames,
        allOriginalSourceAccountNames, allOriginalDestinationAccountNames, allOriginalTransactionTagNames,
    };
}
