import type { SetupProps } from '@/lib/ui/props.ts';
import { useListPageActions } from './useListPage/actions.ts';
import { useListPageDateFilter } from './useListPage/date-filter.ts';
import { useListPageFilters } from './useListPage/filters.ts';
import { useListPageInitialization } from './useListPage/initialization.ts';
import { useListPageLifecycle } from './useListPage/lifecycle.ts';
import { useListPageLoading } from './useListPage/loading.ts';
import { useListPagePresentation } from './useListPage/presentation.ts';
import { useListPageState } from './useListPage/state.ts';
import { type TransactionListProps } from './useListPage/types.ts';
export * from './useListPage/types.ts';

// Coordinates ListPage state and effects; called once from the component setup.
export function useListPage(props: SetupProps<TransactionListProps>) {
    const {
        router, tt, formatRange, getWeekdayLongName, pageType, loading, customMinDatetime, customMaxDatetime, currentCalendarDate,
        firstDayOfWeek, fiscalYearStart, userDefaultCurrency, selectedAccountDefaultCurrency,
        showTotalAmountInTransactionListPage, totalAmountTypeInTransactionListPage, showTagInTransactionListPage, allDateRanges,
        allAccounts, allAccountsMap, allAvailableAccountsCount, allCategories, allPrimaryCategories, allAvailableCategoriesCount,
        allTransactionTagGroupsWithDefault, allTransactionTagsByGroup, allTransactionTags, allAvailableTagsCount, query,
        queryMinTime, queryMaxTime, queryMonthlyData, queryMonth, queryAllFilterCategoryIds, queryAllFilterAccountIds,
        queryAllFilterTagIds, queryAllFilterCategoryIdsCount, queryAllFilterAccountIdsCount, queryAllFilterTagIdsCount,
        queryAccountName, queryCategoryName, queryTagName, queryAmount, transactionCalendarMinDate, transactionCalendarMaxDate,
        currentMonthTransactionData, hasSubCategoryInQuery, hasVisibleTagsInTagGroup, isSameAsDefaultTimezoneOffsetMinutes,
        canAddTransaction, getDailyTotalAmounts, getDisplayMonthTotalAmount, getDisplayTime, getDisplayLongDate,
        getDisplayTimezone, getDisplayTimeInDefaultTimezone, getDisplayAmount, getDisplayAmountCurrency, getTransactionTypeName,
        getTransactionPictureUrl, settingsStore, userStore, accountsStore, transactionCategoriesStore, transactionTagsStore,
        transactionsStore, transactionTemplatesStore, desktopPageStore, confirmDialog, snackbar, editDialog,
        aiImageRecognitionDialog, importDialog, activeTab, currentPage, temporaryCountPerPage, totalCount, searchKeyword,
        currentAmountFilterType, currentAmountFilterValue1, currentAmountFilterValue2, currentPageTransactions, categoryMenuState,
        amountMenuState, exportingData, showCustomDateRangeDialog, showCustomMonthDialog, showFilterAccountDialog,
        showFilterCategoryDialog, showFilterTagDialog, isDarkMode, allPageCounts, recentMonthDateRanges, allTransactionTemplates,
        allowCategoryTypes, updateUrlWhenChanged,
    } = useListPageState();
    const { countPerPage, paginationCurrentPage, reload, add } = useListPageLoading({
        temporaryCountPerPage, settingsStore, totalCount, currentPage, queryMonthlyData, loading, pageType, accountsStore,
        transactionCategoriesStore, transactionTagsStore, desktopPageStore, query, transactionsStore,
        selectedAccountDefaultCurrency, currentPageTransactions, snackbar, editDialog, queryAllFilterCategoryIdsCount,
        queryAllFilterAccountIdsCount, queryAllFilterTagIds,
    });
    const { transactions, transactionsByDay, totalPageCount, skeletonData, currentMonthTotalAmount, getCategoryListItemCheckedClass, getAmountFilterParameterCount } = useListPagePresentation({
        pageType, queryMonthlyData, currentMonthTransactionData, currentPage, countPerPage, currentPageTransactions,
        currentCalendarDate, totalCount, getDisplayMonthTotalAmount,
    });
    const { init } = useListPageInitialization({
        firstDayOfWeek, fiscalYearStart, transactionsStore, pageType, currentCalendarDate, query, updateUrlWhenChanged,
        searchKeyword, currentAmountFilterType, currentPage, reload, transactionTemplatesStore,
    });
    const { recentDateRangeIndex, queryPageType, changeDateFilter, changeCustomDateFilter, changeCustomMonthDateFilter, shiftDateRange } = useListPageDateFilter({
        recentMonthDateRanges, query, firstDayOfWeek, fiscalYearStart, pageType, currentCalendarDate, transactionsStore,
        updateUrlWhenChanged, customMaxDatetime, customMinDatetime, showCustomMonthDialog, showCustomDateRangeDialog,
        accountsStore, allAccountsMap,
    });
    const {
        queryType, queryAllSelectedFilterCategoryIds, queryAllSelectedFilterAccountIds, queryAllSelectedFilterTagIds,
        changeCategoryFilter, changeMultipleCategoriesFilter, changeAccountFilter, changeMultipleAccountsFilter, changeTagFilter,
        changeMultipleTagsFilter, changeKeywordFilter, changeAmountFilter,
    } = useListPageFilters({
        query, queryAllFilterCategoryIdsCount, queryAllFilterAccountIdsCount, queryAllFilterTagIdsCount, queryAllFilterTagIds,
        queryAllFilterCategoryIds, allCategories, transactionsStore, updateUrlWhenChanged, categoryMenuState,
        showFilterCategoryDialog, showFilterAccountDialog, showFilterTagDialog, currentAmountFilterType, amountMenuState,
        getAmountFilterParameterCount, currentAmountFilterValue1, currentAmountFilterValue2, snackbar,
    });
    const { addByRecognizingClipboardText, addByRecognizingImage, importTransaction, exportTransactions, show } = useListPageActions({ add, aiImageRecognitionDialog, settingsStore, editDialog, snackbar, reload, importDialog, exportingData, userStore, tt, transactionsStore });
    const { onShowDateRangeError } = useListPageLifecycle({ snackbar, init, desktopPageStore, add, props });

    return {
        router, tt, formatRange, getWeekdayLongName, pageType, loading, customMinDatetime, customMaxDatetime, currentCalendarDate,
        userDefaultCurrency, selectedAccountDefaultCurrency, showTotalAmountInTransactionListPage,
        totalAmountTypeInTransactionListPage, showTagInTransactionListPage, allDateRanges, allAccounts, allAccountsMap,
        allAvailableAccountsCount, allCategories, allPrimaryCategories, allAvailableCategoriesCount,
        allTransactionTagGroupsWithDefault, allTransactionTagsByGroup, allTransactionTags, allAvailableTagsCount, query,
        queryMinTime, queryMaxTime, queryMonth, queryAllFilterCategoryIds, queryAllFilterAccountIds, queryAllFilterTagIds,
        queryAllFilterCategoryIdsCount, queryAllFilterAccountIdsCount, queryAllFilterTagIdsCount, queryAccountName,
        queryCategoryName, queryTagName, queryAmount, transactionCalendarMinDate, transactionCalendarMaxDate,
        currentMonthTransactionData, hasSubCategoryInQuery, hasVisibleTagsInTagGroup, isSameAsDefaultTimezoneOffsetMinutes,
        canAddTransaction, getDailyTotalAmounts, getDisplayTime, getDisplayLongDate, getDisplayTimezone,
        getDisplayTimeInDefaultTimezone, getDisplayAmount, getDisplayAmountCurrency, getTransactionTypeName,
        getTransactionPictureUrl, confirmDialog, snackbar, editDialog, aiImageRecognitionDialog, importDialog, activeTab,
        searchKeyword, currentAmountFilterType, currentAmountFilterValue1, currentAmountFilterValue2, categoryMenuState,
        amountMenuState, exportingData, showCustomDateRangeDialog, showCustomMonthDialog, showFilterAccountDialog,
        showFilterCategoryDialog, showFilterTagDialog, isDarkMode, allPageCounts, recentMonthDateRanges, allTransactionTemplates,
        allowCategoryTypes, transactions, transactionsByDay, recentDateRangeIndex, queryPageType, queryType,
        queryAllSelectedFilterCategoryIds, queryAllSelectedFilterAccountIds, queryAllSelectedFilterTagIds, countPerPage,
        totalPageCount, paginationCurrentPage, skeletonData, currentMonthTotalAmount, getCategoryListItemCheckedClass, reload,
        changeDateFilter, changeCustomDateFilter, changeCustomMonthDateFilter, shiftDateRange, changeCategoryFilter,
        changeMultipleCategoriesFilter, changeAccountFilter, changeMultipleAccountsFilter, changeTagFilter,
        changeMultipleTagsFilter, changeKeywordFilter, changeAmountFilter, add, addByRecognizingClipboardText,
        addByRecognizingImage, importTransaction, exportTransactions, show, onShowDateRangeError,
    };
}
