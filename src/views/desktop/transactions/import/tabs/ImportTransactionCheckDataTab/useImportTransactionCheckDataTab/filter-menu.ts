import { AmountFilterType } from '@/core/numeral.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    mdiCheck,
    mdiFilterOffOutline
} from '@mdi/js';
import { computed } from 'vue';
import type { useImportTransactionCheckDataTabSelection } from './selection.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';
import { type ImportTransactionCheckDataMenuGroup } from './types.ts';

interface Context {
    tt: ReturnType<typeof useImportTransactionCheckDataTabState>['tt'];
    filters: ReturnType<typeof useImportTransactionCheckDataTabState>['filters'];
    displayFilterCustomDateRange: ReturnType<typeof useImportTransactionCheckDataTabSelection>['displayFilterCustomDateRange'];
    showCustomDateRangeDialog: ReturnType<typeof useImportTransactionCheckDataTabState>['showCustomDateRangeDialog'];
    allUsedCategoryNames: ReturnType<typeof useImportTransactionCheckDataTabSelection>['allUsedCategoryNames'];
    currentAmountFilterType: ReturnType<typeof useImportTransactionCheckDataTabState>['currentAmountFilterType'];
    currentAmountFilterValue1: ReturnType<typeof useImportTransactionCheckDataTabState>['currentAmountFilterValue1'];
    currentAmountFilterValue2: ReturnType<typeof useImportTransactionCheckDataTabState>['currentAmountFilterValue2'];
    showCustomAmountFilterDialog: ReturnType<typeof useImportTransactionCheckDataTabState>['showCustomAmountFilterDialog'];
    allUsedAccountNames: ReturnType<typeof useImportTransactionCheckDataTabSelection>['allUsedAccountNames'];
    allUsedTagNames: ReturnType<typeof useImportTransactionCheckDataTabSelection>['allUsedTagNames'];
    currentDescriptionFilterValue: ReturnType<typeof useImportTransactionCheckDataTabState>['currentDescriptionFilterValue'];
    showCustomDescriptionDialog: ReturnType<typeof useImportTransactionCheckDataTabState>['showCustomDescriptionDialog'];
}

export function useImportTransactionCheckDataTabFilterMenu(context: Context) {
    const {
        tt, filters, displayFilterCustomDateRange, showCustomDateRangeDialog, allUsedCategoryNames, currentAmountFilterType,
        currentAmountFilterValue1, currentAmountFilterValue2, showCustomAmountFilterDialog, allUsedAccountNames, allUsedTagNames,
        currentDescriptionFilterValue, showCustomDescriptionDialog,
    } = context;

    const filterMenus = computed<ImportTransactionCheckDataMenuGroup[]>(() => [
        {
            items: [
                {
                    title: tt('Clear All Filters'),
                    prependIcon: mdiFilterOffOutline,
                    disabled: filters.value.minDatetime === null
                        && filters.value.maxDatetime === null
                        && filters.value.transactionType === null
                        && filters.value.category === null
                        && filters.value.amount === null
                        && filters.value.account === null
                        && filters.value.tag === null
                        && filters.value.description === null,
                    onClick: () => {
                        filters.value.minDatetime = null;
                        filters.value.maxDatetime = null;
                        filters.value.transactionType = null;
                        filters.value.category = null;
                        filters.value.amount = null;
                        filters.value.account = null;
                        filters.value.tag = null;
                        filters.value.description = null;
                    }
                }
            ]
        },
        {
            title: tt('Date Range'),
            items: [
                {
                    title: tt('All'),
                    appendIcon: filters.value.minDatetime === null || filters.value.maxDatetime === null ? mdiCheck : undefined,
                    onClick: () => {
                        filters.value.minDatetime = null;
                        filters.value.maxDatetime = null;
                    }
                },
                {
                    title: tt('Custom'),
                    subTitle: displayFilterCustomDateRange.value,
                    appendIcon: filters.value.minDatetime !== null && filters.value.maxDatetime !== null ? mdiCheck : undefined,
                    onClick: () => showCustomDateRangeDialog.value = true
                }
            ]
        },
        {
            title: tt('Type'),
            items: [
                {
                    title: tt('All'),
                    appendIcon: filters.value.transactionType === null ? mdiCheck : undefined,
                    onClick: () => filters.value.transactionType = null
                },
                {
                    title: tt('Income'),
                    appendIcon: filters.value.transactionType === TransactionType.Income ? mdiCheck : undefined,
                    onClick: () => filters.value.transactionType = TransactionType.Income
                },
                {
                    title: tt('Expense'),
                    appendIcon: filters.value.transactionType === TransactionType.Expense ? mdiCheck : undefined,
                    onClick: () => filters.value.transactionType = TransactionType.Expense
                },
                {
                    title: tt('Transfer'),
                    appendIcon: filters.value.transactionType === TransactionType.Transfer ? mdiCheck : undefined,
                    onClick: () => filters.value.transactionType = TransactionType.Transfer
                }
            ]
        },
        {
            title: tt('Category'),
            items: [
                {
                    title: tt('All'),
                    appendIcon: filters.value.category === null ? mdiCheck : undefined,
                    onClick: () => filters.value.category = null
                },
                {
                    title: tt('Invalid Category'),
                    appendIcon: filters.value.category === undefined ? mdiCheck : undefined,
                    onClick: () => filters.value.category = undefined
                },
                {
                    title: tt('None'),
                    appendIcon: filters.value.category === '' ? mdiCheck : undefined,
                    onClick: () => filters.value.category = ''
                },
                ...allUsedCategoryNames.value.map(name => ({
                    title: name,
                    appendIcon: filters.value.category === name ? mdiCheck : undefined,
                    onClick: () => filters.value.category = name
                }))
            ]
        },
        {
            title: tt('Amount'),
            items: [
                {
                    title: tt('All'),
                    appendIcon: !filters.value.amount ? mdiCheck : undefined,
                    onClick: () => filters.value.amount = null
                },
                ...AmountFilterType.values().map(filterType => ({
                    title: tt(filterType.name),
                    appendIcon: filters.value.amount && filters.value.amount.startsWith(`${filterType.type}:`) ? mdiCheck : undefined,
                    onClick: () => {
                        let filterValue1: number = 0;
                        let filterValue2: number = 0;

                        if (filters.value.amount) {
                            const parts = filters.value.amount.split(':');

                            if (parts.length >= 2) {
                                filterValue1 = parseInt(parts[1] as string);
                            }

                            if (parts.length >= 3) {
                                filterValue2 = parseInt(parts[2] as string);
                            }
                        }

                        if (Number.isNaN(filterValue1) || !Number.isFinite(filterValue1)) {
                            filterValue1 = 0;
                        }

                        if (Number.isNaN(filterValue2) || !Number.isFinite(filterValue2)) {
                            filterValue2 = 0;
                        }

                        currentAmountFilterType.value = filterType;
                        currentAmountFilterValue1.value = filterValue1;
                        currentAmountFilterValue2.value = filterValue2;
                        showCustomAmountFilterDialog.value = true;
                    }
                }))
            ]
        },
        {
            title: tt('Account'),
            items: [
                {
                    title: tt('All'),
                    appendIcon: filters.value.account === null ? mdiCheck : undefined,
                    onClick: () => filters.value.account = null
                },
                {
                    title: tt('Invalid Account'),
                    appendIcon: filters.value.account === undefined ? mdiCheck : undefined,
                    onClick: () => filters.value.account = undefined
                },
                {
                    title: tt('None'),
                    appendIcon: filters.value.account === '' ? mdiCheck : undefined,
                    onClick: () => filters.value.account = ''
                },
                ...allUsedAccountNames.value.map(name => ({
                    title: name,
                    appendIcon: filters.value.account === name ? mdiCheck : undefined,
                    onClick: () => filters.value.account = name
                }))
            ]
        },
        {
            title: tt('Tags'),
            items: [
                {
                    title: tt('All'),
                    appendIcon: filters.value.tag === null ? mdiCheck : undefined,
                    onClick: () => filters.value.tag = null
                },
                {
                    title: tt('Invalid Tag'),
                    appendIcon: filters.value.tag === undefined ? mdiCheck : undefined,
                    onClick: () => filters.value.tag = undefined
                },
                {
                    title: tt('None'),
                    appendIcon: filters.value.tag === '' ? mdiCheck : undefined,
                    onClick: () => filters.value.tag = ''
                },
                ...allUsedTagNames.value.map(name => ({
                    title: name,
                    appendIcon: filters.value.tag === name ? mdiCheck : undefined,
                    onClick: () => filters.value.tag = name
                }))
            ]
        },
        {
            title: tt('Description'),
            items: [
                {
                    title: tt('All'),
                    appendIcon: filters.value.description === null ? mdiCheck : undefined,
                    onClick: () => filters.value.description = null
                },
                {
                    title: tt('None'),
                    appendIcon: filters.value.description === '' ? mdiCheck : undefined,
                    onClick: () => filters.value.description = ''
                },
                {
                    title: tt('Custom'),
                    subTitle: filters.value.description !== null ? filters.value.description : undefined,
                    appendIcon: filters.value.description !== null && filters.value.description !== '' ? mdiCheck : undefined,
                    onClick: () => {
                        currentDescriptionFilterValue.value = filters.value.description || '';
                        showCustomDescriptionDialog.value = true;
                    }
                }
            ]
        }
    ]);

    return { filterMenus };
}
