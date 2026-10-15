import { useTransactionListPageBasePresentation } from './TransactionListPageBase/presentation.ts';
import { useTransactionListPageBaseState } from './TransactionListPageBase/state.ts';
export * from './TransactionListPageBase/types.ts';








export function useTransactionListPageBase() {
    const {
        tt, formatDateTimeToLongDateTime, formatDateTimeToLongDate, formatDateTimeToShortTime,
        formatDateTimeToGregorianLikeLongYearMonth, formatAmountToLocalizedNumeralsWithCurrency, transactionsStore,
        exchangeRatesStore, pageType, loading, customMinDatetime, customMaxDatetime, currentCalendarDate, numeralSystem,
        firstDayOfWeek, fiscalYearStart, userDefaultCurrency, selectedAccountDefaultCurrency,
        showTotalAmountInTransactionListPage, totalAmountTypeInTransactionListPage, showTagInTransactionListPage, allDateRanges,
        allAccounts, allAccountsMap, allAvailableAccountsCount, allPrimaryCategories, allCategories, allAvailableCategoriesCount,
        allTransactionTagGroupsWithDefault, allTransactionTagsByGroup, allTransactionTags, allAvailableTagsCount,
        displayPageTypeName, query, queryDateRangeName, queryMinTime, queryMaxTime, queryMonthlyData, queryMonth,
        queryAllFilterCategoryIds, queryAllFilterAccountIds, queryAllFilterTagIds, queryAllFilterCategoryIdsCount,
        queryAllFilterAccountIdsCount, queryAllFilterTagIdsCount, queryAccountName, queryCategoryName, queryTagName, queryAmount,
        transactionCalendarMinDate, transactionCalendarMaxDate, currentMonthTransactionData, canAddTransaction,
    } = useTransactionListPageBaseState();
    const {
        hasSubCategoryInQuery, hasVisibleTagsInTagGroup, isSameAsDefaultTimezoneOffsetMinutes, hasMonthTotalAmount,
        getDailyTotalAmounts, getDisplayMonthTotalAmount, getDisplayTime, getDisplayLongDate, getDisplayLongYearMonth,
        getDisplayTimezone, getDisplayTimeInDefaultTimezone, getDisplayAmount, getDisplayAmountCurrency, getTransactionTypeName,
        getTransactionPictureUrl,
    } = useTransactionListPageBasePresentation({
        queryAllFilterCategoryIds, allTransactionTagsByGroup, queryAllFilterTagIds, formatAmountToLocalizedNumeralsWithCurrency,
        userDefaultCurrency, exchangeRatesStore, totalAmountTypeInTransactionListPage, selectedAccountDefaultCurrency,
        formatDateTimeToShortTime, formatDateTimeToLongDate, formatDateTimeToGregorianLikeLongYearMonth, numeralSystem,
        formatDateTimeToLongDateTime, queryAllFilterAccountIdsCount, queryAllFilterAccountIds, tt, transactionsStore,
    });

    return {
        // states
        pageType,
        loading,
        customMinDatetime,
        customMaxDatetime,
        currentCalendarDate,
        // computed states
        firstDayOfWeek,
        fiscalYearStart,
        userDefaultCurrency,
        selectedAccountDefaultCurrency,
        showTotalAmountInTransactionListPage,
        totalAmountTypeInTransactionListPage,
        showTagInTransactionListPage,
        allDateRanges,
        allAccounts,
        allAccountsMap,
        allAvailableAccountsCount,
        allCategories,
        allPrimaryCategories,
        allAvailableCategoriesCount,
        allTransactionTagGroupsWithDefault,
        allTransactionTagsByGroup,
        allTransactionTags,
        allAvailableTagsCount,
        displayPageTypeName,
        query,
        queryDateRangeName,
        queryMinTime,
        queryMaxTime,
        queryMonthlyData,
        queryMonth,
        queryAllFilterCategoryIds,
        queryAllFilterAccountIds,
        queryAllFilterTagIds,
        queryAllFilterCategoryIdsCount,
        queryAllFilterAccountIdsCount,
        queryAllFilterTagIdsCount,
        queryAccountName,
        queryCategoryName,
        queryTagName,
        queryAmount,
        transactionCalendarMinDate,
        transactionCalendarMaxDate,
        currentMonthTransactionData,
        canAddTransaction,
        // functions
        hasSubCategoryInQuery,
        hasVisibleTagsInTagGroup,
        isSameAsDefaultTimezoneOffsetMinutes,
        hasMonthTotalAmount,
        getDailyTotalAmounts,
        getDisplayMonthTotalAmount,
        getDisplayTime,
        getDisplayLongDate,
        getDisplayLongYearMonth,
        getDisplayTimezone,
        getDisplayTimeInDefaultTimezone,
        getDisplayAmount,
        getDisplayAmountCurrency,
        getTransactionTypeName,
        getTransactionPictureUrl
    };
}
