import { keys } from '@/core/base.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    transactionTypeToCategoryType
} from '@/lib/category.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import type { Router } from 'framework7/types';
import type { useListPageLoading } from './loading.ts';
import type { useListPageState } from './state.ts';

interface Context {
    query: ReturnType<typeof useListPageState>['query'];
    queryAllFilterCategoryIds: ReturnType<typeof useListPageState>['queryAllFilterCategoryIds'];
    allCategories: ReturnType<typeof useListPageState>['allCategories'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    reload: ReturnType<typeof useListPageLoading>['reload'];
    props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>;
    showSearchbar: ReturnType<typeof useListPageState>['showSearchbar'];
}

export function useListPageFilters(context: Context) {
    const { query, queryAllFilterCategoryIds, allCategories, transactionsStore, reload, props, showSearchbar } = context;

    function changeTypeFilter(type: number): void {
        if (query.value.type === type) {
            return;
        }

        let newCategoryFilter = undefined;

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

        if (changed) {
            reload();
        }
    }

    function changeCategoryFilter(categoryIds: string): void {
        if (query.value.categoryIds === categoryIds) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            categoryIds: categoryIds
        });

        if (changed) {
            reload();
        }
    }

    function filterMultipleCategories(): void {
        let navigateUrl = '/settings/filter/category?type=transactionListCurrent';

        if (TransactionType.Income <= query.value.type && query.value.type <= TransactionType.Transfer) {
            navigateUrl += '&allowCategoryTypes=' + transactionTypeToCategoryType(query.value.type);
        }

        props.f7router.navigate(navigateUrl);
    }

    function changeAccountFilter(accountIds: string): void {
        if (query.value.accountIds === accountIds) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            accountIds: accountIds
        });

        if (changed) {
            reload();
        }
    }

    function filterMultipleAccounts(): void {
        props.f7router.navigate('/settings/filter/account?type=transactionListCurrent');
    }

    function changeTagFilter(tagFilter: string): void {
        if (query.value.tagFilter === tagFilter) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            tagFilter: tagFilter
        });

        if (changed) {
            reload();
        }
    }

    function filterMultipleTags(): void {
        props.f7router.navigate('/settings/filter/tag?type=transactionListCurrent');
    }

    function toggleSearchbar(): void {
        if (!showSearchbar.value) {
            showSearchbar.value = true;
        } else {
            showSearchbar.value = false;

            if (query.value.keyword) {
                changeKeywordFilter('');
            }
        }
    }

    function changeKeywordFilter(keyword: string): void {
        if (query.value.keyword === keyword) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            keyword: keyword
        });

        if (changed) {
            reload();
        }
    }

    function changeAmountFilter(filterType: string): void {
        if (query.value.amountFilter === filterType) {
            return;
        }

        if (filterType) {
            props.f7router.navigate(`/transaction/filter/amount?type=${filterType}&value=${query.value.amountFilter}`);
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            amountFilter: filterType
        });

        if (changed) {
            reload();
        }
    }

    return {
        changeTypeFilter, changeCategoryFilter, filterMultipleCategories, changeAccountFilter, filterMultipleAccounts,
        changeTagFilter, filterMultipleTags, toggleSearchbar, changeKeywordFilter, changeAmountFilter,
    };
}
