import type { SetupProps } from '@/lib/ui/props.ts';
import type { Router } from 'framework7/types';
import { useListPageActions } from './useListPage/actions.ts';
import { useListPageDateFilter } from './useListPage/date-filter.ts';
import { useListPageFilters } from './useListPage/filters.ts';
import { useListPageLifecycle } from './useListPage/lifecycle.ts';
import { useListPageLoading } from './useListPage/loading.ts';
import { useListPageState } from './useListPage/state.ts';
import { useListPageVirtualList } from './useListPage/virtual-list.ts';



// Coordinates ListPage state and effects; called once from the component setup.
export function useListPage(props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>) {
    const {
        tt, getWeekdayShortName, formatNumberToLocalizedNumeralsWithoutDigitGrouping, showAlert, showToast, routeBackOnError,
        pageType, loading, customMinDatetime, customMaxDatetime, currentCalendarDate, firstDayOfWeek, fiscalYearStart,
        selectedAccountDefaultCurrency, showTotalAmountInTransactionListPage, showTagInTransactionListPage, allDateRanges,
        allAccounts, allAccountsMap, allAvailableAccountsCount, allCategories, allPrimaryCategories, allAvailableCategoriesCount,
        allTransactionTagGroupsWithDefault, allTransactionTagsByGroup, allTransactionTags, allAvailableTagsCount,
        displayPageTypeName, query, queryDateRangeName, queryMinTime, queryMaxTime, queryMonthlyData, queryMonth,
        queryAllFilterCategoryIds, queryAllFilterAccountIds, queryAllFilterTagIds, queryAllFilterCategoryIdsCount,
        queryAllFilterAccountIdsCount, queryAllFilterTagIdsCount, queryAccountName, queryCategoryName, queryAmount,
        transactionCalendarMinDate, transactionCalendarMaxDate, currentMonthTransactionData, hasSubCategoryInQuery,
        hasVisibleTagsInTagGroup, isSameAsDefaultTimezoneOffsetMinutes, canAddTransaction, hasMonthTotalAmount,
        getDailyTotalAmounts, getDisplayMonthTotalAmount, getDisplayTime, getDisplayLongYearMonth, getDisplayTimezone,
        getDisplayAmount, getTransactionTypeName, getTransactionPictureUrl, accountsStore, transactionCategoriesStore,
        transactionTagsStore, transactionsStore, loadingError, loadingMore, transactionToDelete, transactionInvisibleYearMonths,
        transactionYearMonthListHeights, showSearchbar, showCustomDateRangeSheet, showCustomMonthSheet, showDeleteActionSheet,
        textDirection, isDarkMode,
    } = useListPageState();
    const {
        transactions, noTransaction, hasMoreTransaction, getTransactionMonthTitleDomId, getTransactionMonthListDomId,
        getTransactionDomId, isTransactionMonthListInvisible, getTransactionMonthListHeight, setTransactionMonthListHeights,
        getTransactionDateStyle, getCategoryListItemCheckedClass, collapseTransactionMonthList, onResize, onScroll,
        onTransactionMonthListCollapseStateChanged,
    } = useListPageVirtualList({
        loading, pageType, transactionsStore, queryMonthlyData, currentMonthTransactionData, currentCalendarDate,
        transactionYearMonthListHeights, transactionInvisibleYearMonths,
    });
    const { init, reload, loadMore } = useListPageLoading({
        props, firstDayOfWeek, fiscalYearStart, transactionsStore, loading, pageType, transactionInvisibleYearMonths,
        transactionYearMonthListHeights, accountsStore, transactionCategoriesStore, transactionTagsStore, queryMonthlyData, query,
        selectedAccountDefaultCurrency, showToast, setTransactionMonthListHeights, loadingError, hasMoreTransaction, loadingMore,
    });
    const { changePageType, changeDateFilter, changeCustomDateFilter, changeCustomMonthDateFilter, shiftDateRange } = useListPageDateFilter({
        pageType, currentCalendarDate, query, firstDayOfWeek, fiscalYearStart, transactionsStore, reload, customMaxDatetime,
        customMinDatetime, showCustomMonthSheet, showCustomDateRangeSheet, accountsStore, allAccountsMap,
    });
    const {
        changeTypeFilter, changeCategoryFilter, filterMultipleCategories, changeAccountFilter, filterMultipleAccounts,
        changeTagFilter, filterMultipleTags, toggleSearchbar, changeKeywordFilter, changeAmountFilter,
    } = useListPageFilters({ query, queryAllFilterCategoryIds, allCategories, transactionsStore, reload, props, showSearchbar });
    const { add, duplicate, edit, remove } = useListPageActions({
        query, queryAllFilterCategoryIdsCount, queryAllFilterAccountIdsCount, queryAllFilterTagIds, props, showAlert,
        transactionToDelete, showDeleteActionSheet, transactionsStore, selectedAccountDefaultCurrency, getTransactionDomId,
        showToast,
    });
    const { onPopoverOpen, onCategoryPopoverOpen, onPageAfterIn } = useListPageLifecycle({ transactionsStore, loading, reload, routeBackOnError, props, loadingError, onResize, onScroll, init });

    return {
        tt, getWeekdayShortName, formatNumberToLocalizedNumeralsWithoutDigitGrouping, pageType, loading, customMinDatetime,
        customMaxDatetime, currentCalendarDate, showTotalAmountInTransactionListPage, showTagInTransactionListPage, allDateRanges,
        allAccounts, allAccountsMap, allAvailableAccountsCount, allCategories, allPrimaryCategories, allAvailableCategoriesCount,
        allTransactionTagGroupsWithDefault, allTransactionTagsByGroup, allTransactionTags, allAvailableTagsCount,
        displayPageTypeName, query, queryDateRangeName, queryMinTime, queryMaxTime, queryMonth, queryAllFilterCategoryIds,
        queryAllFilterAccountIds, queryAllFilterTagIds, queryAllFilterCategoryIdsCount, queryAllFilterAccountIdsCount,
        queryAllFilterTagIdsCount, queryAccountName, queryCategoryName, queryAmount, transactionCalendarMinDate,
        transactionCalendarMaxDate, currentMonthTransactionData, hasSubCategoryInQuery, hasVisibleTagsInTagGroup,
        isSameAsDefaultTimezoneOffsetMinutes, canAddTransaction, hasMonthTotalAmount, getDailyTotalAmounts,
        getDisplayMonthTotalAmount, getDisplayTime, getDisplayLongYearMonth, getDisplayTimezone, getDisplayAmount,
        getTransactionTypeName, getTransactionPictureUrl, loadingMore, transactionToDelete, showSearchbar,
        showCustomDateRangeSheet, showCustomMonthSheet, showDeleteActionSheet, textDirection, isDarkMode, transactions,
        noTransaction, hasMoreTransaction, getTransactionMonthTitleDomId, getTransactionMonthListDomId, getTransactionDomId,
        isTransactionMonthListInvisible, getTransactionMonthListHeight, getTransactionDateStyle, getCategoryListItemCheckedClass,
        init, reload, loadMore, changePageType, changeDateFilter, changeCustomDateFilter, changeCustomMonthDateFilter,
        shiftDateRange, changeTypeFilter, changeCategoryFilter, filterMultipleCategories, changeAccountFilter,
        filterMultipleAccounts, changeTagFilter, filterMultipleTags, toggleSearchbar, changeKeywordFilter, changeAmountFilter,
        add, duplicate, edit, remove, collapseTransactionMonthList, onPopoverOpen, onCategoryPopoverOpen, onPageAfterIn,
        onTransactionMonthListCollapseStateChanged,
    };
}
