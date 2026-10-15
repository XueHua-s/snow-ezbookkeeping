import { itemAndIndex } from '@/core/base.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    objectFieldToArrayItem
} from '@/lib/common.ts';
import {
    parseDateTimeFromUnixTime
} from '@/lib/datetime.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import { computed } from 'vue';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>;
    allTagsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allTagsMap'];
    filters: ReturnType<typeof useImportTransactionCheckDataTabState>['filters'];
    formatDateTimeToLongDateTime: ReturnType<typeof useImportTransactionCheckDataTabState>['formatDateTimeToLongDateTime'];
    formatRange: ReturnType<typeof useImportTransactionCheckDataTabState>['formatRange'];
}

export function useImportTransactionCheckDataTabSelection(context: Context) {
    const { props, allTagsMap, filters, formatDateTimeToLongDateTime, formatRange } = context;

    const canImport = computed<boolean>(() => selectedImportTransactionCount.value > 0 && selectedInvalidTransactionCount.value < 1);

    const selectedImportTransactionCount = computed<number>(() => {
        let count = 0;

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return count;
        }

        for (const importTransaction of props.importTransactions) {
            if (importTransaction.selected) {
                count++;
            }
        }

        return count;
    });

    const selectedExpenseTransactionCount = computed<number>(() => {
        let count = 0;

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return count;
        }

        for (const importTransaction of props.importTransactions) {
            if (importTransaction.selected && importTransaction.type === TransactionType.Expense) {
                count++;
            }
        }

        return count;
    });

    const selectedIncomeTransactionCount = computed<number>(() => {
        let count = 0;

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return count;
        }

        for (const importTransaction of props.importTransactions) {
            if (importTransaction.selected && importTransaction.type === TransactionType.Income) {
                count++;
            }
        }

        return count;
    });

    const selectedTransferTransactionCount = computed<number>(() => {
        let count = 0;

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return count;
        }

        for (const importTransaction of props.importTransactions) {
            if (importTransaction.selected && importTransaction.type === TransactionType.Transfer) {
                count++;
            }
        }

        return count;
    });

    const selectedInvalidTransactionCount = computed<number>(() => {
        let count = 0;

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return count;
        }

        for (const importTransaction of props.importTransactions) {
            if (!importTransaction.valid && importTransaction.selected) {
                count++;
            }
        }

        return count;
    });

    const anyButNotAllTransactionSelected = computed<boolean>(() => !!props.importTransactions && selectedImportTransactionCount.value > 0 && selectedImportTransactionCount.value !== props.importTransactions.length);

    const allTransactionSelected = computed<boolean>(() => !!props.importTransactions && selectedImportTransactionCount.value === props.importTransactions.length);

    const allUsedCategoryNames = computed<string[]>(() => {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return [];
        }

        const categoryNames: Record<string, boolean> = {};

        for (const transaction of props.importTransactions) {
            if (transaction.actualCategoryName && transaction.actualCategoryName !== '') {
                categoryNames[transaction.actualCategoryName] = true;
            }
        }

        return objectFieldToArrayItem(categoryNames);
    });

    const allUsedAccountNames = computed<string[]>(() => {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return [];
        }

        const accountNames: Record<string, boolean> = {};

        for (const transaction of props.importTransactions) {
            if (transaction.actualSourceAccountName && transaction.actualSourceAccountName !== '') {
                accountNames[transaction.actualSourceAccountName] = true;
            }

            if (transaction.actualDestinationAccountName && transaction.actualDestinationAccountName !== '') {
                accountNames[transaction.actualDestinationAccountName] = true;
            }
        }

        return objectFieldToArrayItem(accountNames);
    });

    const allUsedTagNames = computed<string[]>(() => {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return [];
        }

        const tagNames: Record<string, boolean> = {};

        for (const transaction of props.importTransactions) {
            if (!transaction.tagIds || !transaction.originalTagNames) {
                continue;
            }

            for (const [tagId, tagIndex] of itemAndIndex(transaction.tagIds)) {
                const originalTagName = transaction.originalTagNames[tagIndex] as string | undefined;

                if (tagId && tagId !== '0' && allTagsMap.value[tagId] && allTagsMap.value[tagId].name) {
                    tagNames[allTagsMap.value[tagId].name] = true;
                } else if (originalTagName) {
                    tagNames[originalTagName] = true;
                }
            }
        }

        return objectFieldToArrayItem(tagNames);
    });

    const displayFilterCustomDateRange = computed<string>(() => {
        if (filters.value.minDatetime === null || filters.value.maxDatetime === null) {
            return '';
        }

        const minDisplayTime = formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(filters.value.minDatetime));
        const maxDisplayTime = formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(filters.value.maxDatetime));

        return formatRange(minDisplayTime, maxDisplayTime);
    });

    return {
        canImport, selectedImportTransactionCount, selectedExpenseTransactionCount, selectedIncomeTransactionCount,
        selectedTransferTransactionCount, selectedInvalidTransactionCount, anyButNotAllTransactionSelected,
        allTransactionSelected, allUsedCategoryNames, allUsedAccountNames, allUsedTagNames, displayFilterCustomDateRange,
    };
}
