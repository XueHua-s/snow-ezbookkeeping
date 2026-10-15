import { DEFAULT_PAGE_COUNTS } from '@/consts/page.ts';
import {
    type NameNumeralValue
} from '@/core/base.ts';
import {
    type LocalizedRecentMonthDateRange
} from '@/core/datetime.ts';
import { TemplateType } from '@/core/template.ts';
import { ThemeType } from '@/core/theme.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    transactionTypeToCategoryType
} from '@/lib/category.ts';
import { useI18n } from '@/locales/helpers.ts';
import { type Transaction } from '@/models/transaction.ts';
import type { TransactionTemplate } from '@/models/transaction_template.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useDesktopPageStore } from '@/stores/desktopPage.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useUserStore } from '@/stores/user.ts';
import { TransactionListPageType, useTransactionListPageBase } from '@/views/base/transactions/TransactionListPageBase.ts';
import { computed, ref, useTemplateRef } from 'vue';
import { useRouter } from 'vue-router';
import { useTheme } from 'vuetify';

import { type AIImageRecognitionDialogType, type ConfirmDialogType, type EditDialogType, type ImportDialogType, type SnackBarType } from './types.ts';


export function useListPageState() {
    const router = useRouter();

    const theme = useTheme();

    const {
        tt,
        formatRange,
        getAllRecentMonthDateRanges,
        getWeekdayLongName,
        getTablePageOptions
    } = useI18n();

    const {
        pageType,
        loading,
        customMinDatetime,
        customMaxDatetime,
        currentCalendarDate,
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
        query,
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
        hasSubCategoryInQuery,
        hasVisibleTagsInTagGroup,
        isSameAsDefaultTimezoneOffsetMinutes,
        canAddTransaction,
        getDailyTotalAmounts,
        getDisplayMonthTotalAmount,
        getDisplayTime,
        getDisplayLongDate,
        getDisplayTimezone,
        getDisplayTimeInDefaultTimezone,
        getDisplayAmount,
        getDisplayAmountCurrency,
        getTransactionTypeName,
        getTransactionPictureUrl
    } = useTransactionListPageBase();

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const transactionsStore = useTransactionsStore();

    const transactionTemplatesStore = useTransactionTemplatesStore();

    const desktopPageStore = useDesktopPageStore();






    const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const editDialog = useTemplateRef<EditDialogType>('editDialog');

    const aiImageRecognitionDialog = useTemplateRef<AIImageRecognitionDialogType>('aiImageRecognitionDialog');

    const importDialog = useTemplateRef<ImportDialogType>('importDialog');

    const activeTab = ref<string>('transactionPage');

    const currentPage = ref<number>(1);

    const temporaryCountPerPage = ref<number | null>(null);

    const totalCount = ref<number>(1);

    const searchKeyword = ref<string>('');

    const currentAmountFilterType = ref<string>('');

    const currentAmountFilterValue1 = ref<number>(0);

    const currentAmountFilterValue2 = ref<number>(0);

    const currentPageTransactions = ref<Transaction[]>([]);

    const categoryMenuState = ref<boolean>(false);

    const amountMenuState = ref<boolean>(false);

    const exportingData = ref<boolean>(false);

    const showCustomDateRangeDialog = ref<boolean>(false);

    const showCustomMonthDialog = ref<boolean>(false);

    const showFilterAccountDialog = ref<boolean>(false);

    const showFilterCategoryDialog = ref<boolean>(false);

    const showFilterTagDialog = ref<boolean>(false);

    const isDarkMode = computed<boolean>(() => theme.global.name.value === ThemeType.Dark);

    const allPageCounts = computed<NameNumeralValue[]>(() => getTablePageOptions(DEFAULT_PAGE_COUNTS, undefined, false, true));

    const recentMonthDateRanges = computed<LocalizedRecentMonthDateRange[]>(() => getAllRecentMonthDateRanges(pageType.value === TransactionListPageType.List.type || pageType.value === TransactionListPageType.Gallery.type, true));

    const allTransactionTemplates = computed<TransactionTemplate[]>(() => {
        const allTemplates = transactionTemplatesStore.allVisibleTemplates;
        return allTemplates[TemplateType.Normal.type] || [];
    });

    const allowCategoryTypes = computed<string>(() => {
        if (TransactionType.Income <= query.value.type && query.value.type <= TransactionType.Transfer) {
            return transactionTypeToCategoryType(query.value.type)?.toString() ?? '';
        }

        return '';
    });

    function updateUrlWhenChanged(changed: boolean): void {
        if (changed) {
            loading.value = true;
            currentPageTransactions.value = [];
            transactionsStore.clearTransactions();
            router.push(`/transaction/list?${transactionsStore.getTransactionListPageParams(pageType.value)}`);
        }
    }

    return {
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
    };
}
