import {
    keys
} from '@/core/base.ts';
import {
    transactionTypeToCategoryType
} from '@/lib/category.ts';
import { TransactionTagFilter } from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useListPagePresentation } from './presentation.ts';
import type { useListPageState } from './state.ts';

interface Context {
    query: ReturnType<typeof useListPageState>['query'];
    queryAllFilterCategoryIdsCount: ReturnType<typeof useListPageState>['queryAllFilterCategoryIdsCount'];
    queryAllFilterAccountIdsCount: ReturnType<typeof useListPageState>['queryAllFilterAccountIdsCount'];
    queryAllFilterTagIdsCount: ReturnType<typeof useListPageState>['queryAllFilterTagIdsCount'];
    queryAllFilterTagIds: ReturnType<typeof useListPageState>['queryAllFilterTagIds'];
    queryAllFilterCategoryIds: ReturnType<typeof useListPageState>['queryAllFilterCategoryIds'];
    allCategories: ReturnType<typeof useListPageState>['allCategories'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    updateUrlWhenChanged: ReturnType<typeof useListPageState>['updateUrlWhenChanged'];
    categoryMenuState: ReturnType<typeof useListPageState>['categoryMenuState'];
    showFilterCategoryDialog: ReturnType<typeof useListPageState>['showFilterCategoryDialog'];
    showFilterAccountDialog: ReturnType<typeof useListPageState>['showFilterAccountDialog'];
    showFilterTagDialog: ReturnType<typeof useListPageState>['showFilterTagDialog'];
    currentAmountFilterType: ReturnType<typeof useListPageState>['currentAmountFilterType'];
    amountMenuState: ReturnType<typeof useListPageState>['amountMenuState'];
    getAmountFilterParameterCount: ReturnType<typeof useListPagePresentation>['getAmountFilterParameterCount'];
    currentAmountFilterValue1: ReturnType<typeof useListPageState>['currentAmountFilterValue1'];
    currentAmountFilterValue2: ReturnType<typeof useListPageState>['currentAmountFilterValue2'];
    snackbar: ReturnType<typeof useListPageState>['snackbar'];
}

export function useListPageFilters(context: Context) {
    const {
        query, queryAllFilterCategoryIdsCount, queryAllFilterAccountIdsCount, queryAllFilterTagIdsCount, queryAllFilterTagIds,
        queryAllFilterCategoryIds, allCategories, transactionsStore, updateUrlWhenChanged, categoryMenuState,
        showFilterCategoryDialog, showFilterAccountDialog, showFilterTagDialog, currentAmountFilterType, amountMenuState,
        getAmountFilterParameterCount, currentAmountFilterValue1, currentAmountFilterValue2, snackbar,
    } = context;

    const queryType = computed<number>({
        get: () => query.value.type,
        set: (value) => changeTypeFilter(value)
    });

    const queryAllSelectedFilterCategoryIds = computed<string>(() => {
        if (queryAllFilterCategoryIdsCount.value === 0) {
            return '';
        } else if (queryAllFilterCategoryIdsCount.value === 1) {
            return query.value.categoryIds;
        } else { // queryAllFilterCategoryIdsCount.value > 1
            return 'multiple';
        }
    });

    const queryAllSelectedFilterAccountIds = computed<string>(() => {
        if (queryAllFilterAccountIdsCount.value === 0) {
            return '';
        } else if (queryAllFilterAccountIdsCount.value === 1) {
            return query.value.accountIds;
        } else { // queryAllFilterAccountIdsCount.value > 1
            return 'multiple';
        }
    });

    const queryAllSelectedFilterTagIds = computed<string>(() => {
        if (query.value.tagFilter === TransactionTagFilter.TransactionNoTagFilterValue) {
            return TransactionTagFilter.TransactionNoTagFilterValue;
        } else if (queryAllFilterTagIdsCount.value === 0) {
            return '';
        } else if (queryAllFilterTagIdsCount.value === 1) {
            for (const tagId of keys(queryAllFilterTagIds.value)) {
                return tagId;
            }

            return '';
        } else { // queryAllFilterTagIdsCount.value > 1
            return 'multiple';
        }
    });

    function changeTypeFilter(type: number): void {
        let newCategoryFilter: string | undefined = undefined;

        if (type && query.value.categoryIds) {
            newCategoryFilter = '';

            for (const categoryId of keys(queryAllFilterCategoryIds.value)) {
                const category = allCategories.value[categoryId];

                if (category && category.type === transactionTypeToCategoryType(type)) {
                    if (newCategoryFilter.length > 0) {
                        newCategoryFilter += ',';
                    }

                    newCategoryFilter += categoryId;
                }
            }
        }

        const changed = transactionsStore.updateTransactionListFilter({
            type: type,
            categoryIds: newCategoryFilter
        });

        updateUrlWhenChanged(changed);
    }

    function changeCategoryFilter(categoryIds: string): void {
        categoryMenuState.value = false;

        if (query.value.categoryIds === categoryIds) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            categoryIds: categoryIds
        });

        updateUrlWhenChanged(changed);
    }

    function changeMultipleCategoriesFilter(changed: boolean): void {
        categoryMenuState.value = false;
        showFilterCategoryDialog.value = false;
        updateUrlWhenChanged(changed);
    }

    function changeAccountFilter(accountIds: string): void {
        if (query.value.accountIds === accountIds) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            accountIds: accountIds
        });

        updateUrlWhenChanged(changed);
    }

    function changeMultipleAccountsFilter(changed: boolean): void {
        showFilterAccountDialog.value = false;
        updateUrlWhenChanged(changed);
    }

    function changeTagFilter(tagFilter: string): void {
        if (query.value.tagFilter === tagFilter) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            tagFilter: tagFilter
        });

        updateUrlWhenChanged(changed);
    }

    function changeMultipleTagsFilter(changed: boolean): void {
        showFilterTagDialog.value = false;

        updateUrlWhenChanged(changed);
    }

    function changeKeywordFilter(keyword: string): void {
        if (query.value.keyword === keyword) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            keyword: keyword
        });

        updateUrlWhenChanged(changed);
    }

    function changeAmountFilter(filterType: string): void {
        currentAmountFilterType.value = '';
        amountMenuState.value = false;

        if (query.value.amountFilter === filterType) {
            return;
        }

        let amountFilter = filterType;

        if (filterType) {
            const amountCount = getAmountFilterParameterCount(filterType);

            if (!amountCount) {
                return;
            }

            if (amountCount === 1) {
                amountFilter += ':' + currentAmountFilterValue1.value;
            } else if (amountCount === 2) {
                if (currentAmountFilterValue2.value < currentAmountFilterValue1.value) {
                    snackbar.value?.showMessage('Incorrect amount range');
                    return;
                }

                amountFilter += ':' + currentAmountFilterValue1.value + ':' + currentAmountFilterValue2.value;
            } else {
                return;
            }
        }

        if (query.value.amountFilter === amountFilter) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            amountFilter: amountFilter
        });

        updateUrlWhenChanged(changed);
    }

    return {
        queryType, queryAllSelectedFilterCategoryIds, queryAllSelectedFilterAccountIds, queryAllSelectedFilterTagIds,
        changeCategoryFilter, changeMultipleCategoriesFilter, changeAccountFilter, changeMultipleAccountsFilter, changeTagFilter,
        changeMultipleTagsFilter, changeKeywordFilter, changeAmountFilter,
    };
}
