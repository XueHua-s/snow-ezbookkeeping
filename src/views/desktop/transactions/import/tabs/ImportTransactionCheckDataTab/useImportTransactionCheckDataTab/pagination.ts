import { DEFAULT_PAGE_COUNTS } from '@/consts/page.ts';
import { type NameNumeralValue } from '@/core/base.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import { computed } from 'vue';
import type { useImportTransactionCheckDataTabFiltering } from './filtering.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    tt: ReturnType<typeof useImportTransactionCheckDataTabState>['tt'];
    getTablePageOptions: ReturnType<typeof useImportTransactionCheckDataTabState>['getTablePageOptions'];
    props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>;
    isTransactionDisplayed: ReturnType<typeof useImportTransactionCheckDataTabFiltering>['isTransactionDisplayed'];
    countPerPage: ReturnType<typeof useImportTransactionCheckDataTabState>['countPerPage'];
    currentPage: ReturnType<typeof useImportTransactionCheckDataTabState>['currentPage'];
}

export function useImportTransactionCheckDataTabPagination(context: Context) {
    const { tt, getTablePageOptions, props, isTransactionDisplayed, countPerPage, currentPage } = context;

    const importTransactionHeaders = computed<object[]>(() => {
        return [
            { key: 'data-table-select', fixed: true },
            { value: 'valid', sortable: true, nowrap: true, width: 35, fixed: true },
            { value: 'time', title: tt('Transaction Time'), sortable: true, nowrap: true },
            { value: 'type', title: tt('Type'), sortable: true, nowrap: true },
            { value: 'actualCategoryName', title: tt('Category'), sortable: true, nowrap: true },
            { value: 'sourceAmount', title: tt('Amount'), sortable: true, nowrap: true },
            { value: 'actualSourceAccountName', title: tt('Account'), sortable: true, nowrap: true },
            { value: 'geoLocation', title: tt('Geographic Location'), sortable: true, nowrap: true },
            { value: 'tagIds', title: tt('Tags'), sortable: true, nowrap: true },
            { value: 'comment', title: tt('Description'), sortable: true, nowrap: true },
        ];
    });

    const importTransactionsTablePageOptions = computed<NameNumeralValue[]>(() => getTablePageOptions(DEFAULT_PAGE_COUNTS, props.importTransactions?.length, true, false));

    const totalPageCount = computed<number>(() => {
        if (!props.importTransactions || props.importTransactions.length < 1) {
            return 1;
        }

        let count = 0;

        for (const importTransaction of props.importTransactions) {
            if (isTransactionDisplayed(importTransaction)) {
                count++;
            }
        }

        return Math.ceil(count / countPerPage.value);
    });

    const currentPageTransactions = computed<ImportTransaction[]>(() => {
        const ret: ImportTransaction[] = [];

        if (!props.importTransactions || props.importTransactions.length < 1) {
            return ret;
        }

        const previousCount = Math.max(0, (currentPage.value - 1) * countPerPage.value);
        let count = 0;

        for (const importTransaction of props.importTransactions) {
            if (ret.length >= countPerPage.value) {
                break;
            }

            if (isTransactionDisplayed(importTransaction)) {
                if (count >= previousCount) {
                    ret.push(importTransaction);
                }

                count++;
            }
        }

        return ret;
    });

    return { importTransactionHeaders, importTransactionsTablePageOptions, totalPageCount, currentPageTransactions };
}
