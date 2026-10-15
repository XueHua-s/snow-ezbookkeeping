import {
    type TextualYearMonth
} from '@/core/datetime.ts';
import { TextDirection } from '@/core/text.ts';
import {
    useI18nUIComponents
} from '@/lib/ui/mobile.ts';
import { useI18n } from '@/locales/helpers.ts';
import { type Transaction } from '@/models/transaction.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useEnvironmentsStore } from '@/stores/environment.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionListPageBase } from '@/views/base/transactions/TransactionListPageBase.ts';
import { computed, ref } from 'vue';


export function useListPageState() {
    const {
        tt,
        getCurrentLanguageTextDirection,
        getWeekdayShortName,
        formatNumberToLocalizedNumeralsWithoutDigitGrouping
    } = useI18n();

    const { showAlert, showToast, routeBackOnError } = useI18nUIComponents();

    const {
        pageType,
        loading,
        customMinDatetime,
        customMaxDatetime,
        currentCalendarDate,
        firstDayOfWeek,
        fiscalYearStart,
        selectedAccountDefaultCurrency,
        showTotalAmountInTransactionListPage,
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
        queryAmount,
        transactionCalendarMinDate,
        transactionCalendarMaxDate,
        currentMonthTransactionData,
        hasSubCategoryInQuery,
        hasVisibleTagsInTagGroup,
        isSameAsDefaultTimezoneOffsetMinutes,
        canAddTransaction,
        hasMonthTotalAmount,
        getDailyTotalAmounts,
        getDisplayMonthTotalAmount,
        getDisplayTime,
        getDisplayLongYearMonth,
        getDisplayTimezone,
        getDisplayAmount,
        getTransactionTypeName,
        getTransactionPictureUrl
    } = useTransactionListPageBase();

    const environmentsStore = useEnvironmentsStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const transactionsStore = useTransactionsStore();

    const loadingError = ref<unknown | null>(null);

    const loadingMore = ref<boolean>(false);

    const transactionToDelete = ref<Transaction | null>(null);

    const transactionInvisibleYearMonths = ref<Record<TextualYearMonth, boolean>>({});

    const transactionYearMonthListHeights = ref<Record<TextualYearMonth, number>>({});

    const showSearchbar = ref<boolean>(false);

    const showCustomDateRangeSheet = ref<boolean>(false);

    const showCustomMonthSheet = ref<boolean>(false);

    const showDeleteActionSheet = ref<boolean>(false);

    const textDirection = computed<TextDirection>(() => getCurrentLanguageTextDirection());

    const isDarkMode = computed<boolean>(() => environmentsStore.framework7DarkMode || false);

    return {
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
    };
}
