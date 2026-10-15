import { DEFAULT_TAG_GROUP_ID } from '@/consts/tag.ts';
import { AccountType } from '@/core/account.ts';
import { entries, keys } from '@/core/base.ts';
import { type LocalizedDateRange, type TextualYearMonthDay, type WeekDayValue, type Year0BasedMonth, DateRange, DateRangeScene } from '@/core/datetime.ts';
import type { NumeralSystem } from '@/core/numeral.ts';
import {
    getUnifiedSelectedAccountsCurrencyOrDefaultCurrency
} from '@/lib/account.ts';
import {
    categoryTypeToTransactionType
} from '@/lib/category.ts';
import {
    getCurrentDateTime,
    getLocalDatetimeFromUnixTime,
    getSameDateTimeWithBrowserTimezone,
    isDateRangeMatchOneMonth,
    parseDateTimeFromUnixTime
} from '@/lib/datetime.ts';
import {
    parseBigDecimal
} from '@/lib/numeral.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { Account } from '@/models/account.ts';
import { TransactionTagFilter } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionTag } from '@/models/transaction_tag.ts';
import { TransactionTagGroup } from '@/models/transaction_tag_group.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { type TransactionListFilter, type TransactionMonthList, useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useUserStore } from '@/stores/user.ts';
import { computed, ref } from 'vue';
import { TransactionListPageType } from './types.ts';


// Owns state for the TransactionListPageBase store.
export function useTransactionListPageBaseState() {
    const {
        tt,
        formatRange,
        getAllDateRanges,
        getCurrentNumeralSystemType,
        formatDateTimeToLongDateTime,
        formatDateTimeToLongDate,
        formatDateTimeToShortTime,
        formatDateTimeToGregorianLikeLongYearMonth,
        formatDateRange,
        formatAmountToLocalizedNumeralsWithCurrency
    } = useI18n();

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const transactionsStore = useTransactionsStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const pageType = ref<number>(TransactionListPageType.List.type);

    const loading = ref<boolean>(true);

    const customMinDatetime = ref<number>(0);

    const customMaxDatetime = ref<number>(0);

    const currentCalendarDate = ref<TextualYearMonthDay | ''>('');

    const numeralSystem = computed<NumeralSystem>(() => getCurrentNumeralSystemType());

    const firstDayOfWeek = computed<WeekDayValue>(() => userStore.currentUserFirstDayOfWeek);

    const fiscalYearStart = computed<number>(() => userStore.currentUserFiscalYearStart);

    const userDefaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

    const selectedAccountDefaultCurrency = computed<string>(() => getUnifiedSelectedAccountsCurrencyOrDefaultCurrency(allAccountsMap.value, queryAllFilterAccountIds.value, userStore.currentUserDefaultCurrency));

    const showTotalAmountInTransactionListPage = computed<boolean>(() => settingsStore.appSettings.showTotalAmountInTransactionListPage);

    const totalAmountTypeInTransactionListPage = computed<number>(() => settingsStore.appSettings.totalAmountTypeInTransactionListPage);

    const showTagInTransactionListPage = computed<boolean>(() => settingsStore.appSettings.showTagInTransactionListPage);

    const allDateRanges = computed<LocalizedDateRange[]>(() => getAllDateRanges(DateRangeScene.Normal, {
        includeCustom: true,
        includeBillingCycle: !!accountsStore.getAccountStatementDate(query.value.accountIds),
        includeLastReconciledTimeRange: userStore.currentUserUseLastReconciledTime && !!allAccountsMap.value[query.value.accountIds]?.lastReconciledTime
    }));

    const allAccounts = computed<Account[]>(() => accountsStore.allMixedPlainAccounts);

    const allAccountsMap = computed<Record<string, Account>>(() => accountsStore.allAccountsMap);

    const allAvailableAccountsCount = computed<number>(() => accountsStore.allAvailableAccountsCount);

    const allPrimaryCategories = computed<Record<string, TransactionCategory[]>>(() => {
        const primaryCategories: Record<string, TransactionCategory[]> = {};

        for (const [categoryType, categories] of entries(transactionCategoriesStore.allTransactionCategories)) {
            if (query.value.type && categoryTypeToTransactionType(parseInt(categoryType)) !== query.value.type) {
                continue;
            }

            primaryCategories[categoryType] = categories;
        }

        return primaryCategories;
    });

    const allCategories = computed<Record<string, TransactionCategory>>(() => transactionCategoriesStore.allTransactionCategoriesMap);

    const allAvailableCategoriesCount = computed<number>(() => {
        let totalCount = 0;

        for (const [categoryType, categories] of entries(transactionCategoriesStore.allTransactionCategories)) {
            if (query.value.type && categoryTypeToTransactionType(parseInt(categoryType)) !== query.value.type) {
                continue;
            }

            if (categories) {
                totalCount += categories.length;
            }
        }

        return totalCount;
    });

    const allTransactionTagGroupsWithDefault = computed<TransactionTagGroup[]>(() => {
        const allGroups: TransactionTagGroup[] = [];
        const defaultGroup = TransactionTagGroup.createNewTagGroup(tt('Default Group'));
        defaultGroup.id = DEFAULT_TAG_GROUP_ID;
        allGroups.push(defaultGroup);
        allGroups.push(...transactionTagsStore.allTransactionTagGroups);
        return allGroups;
    });

    const allTransactionTagsByGroup = computed<Record<string, TransactionTag[]>>(() => transactionTagsStore.allTransactionTagsByGroupMap);

    const allTransactionTags = computed<Record<string, TransactionTag>>(() => transactionTagsStore.allTransactionTagsMap);

    const allAvailableTagsCount = computed<number>(() => transactionTagsStore.allAvailableTagsCount);

    const displayPageTypeName = computed<string>(() => {
        const type = TransactionListPageType.valueOf(pageType.value);

        if (type) {
            return tt(type.name);
        }

        return tt(TransactionListPageType.List.name);
    });

    const query = computed<TransactionListFilter>(() => transactionsStore.transactionsFilter);

    const queryDateRangeName = computed<string>(() => {
        if (query.value.dateType === DateRange.All.type) {
            return tt('Date');
        }

        return formatDateRange(query.value.dateType, query.value.minTime, query.value.maxTime);
    });

    const queryMinTime = computed<string>(() => formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(query.value.minTime)));

    const queryMaxTime = computed<string>(() => formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(query.value.maxTime)));

    const queryMonthlyData = computed<boolean>(() => isDateRangeMatchOneMonth(query.value.minTime, query.value.maxTime));

    const queryMonth = computed<Year0BasedMonth>(() => {
        if (!query.value.minTime || !query.value.maxTime) {
            return getCurrentDateTime().toGregorianCalendarYear0BasedMonth();
        }

        return parseDateTimeFromUnixTime(query.value.minTime).toGregorianCalendarYear0BasedMonth();
    });

    const queryAllFilterCategoryIds = computed<Record<string, boolean>>(() => transactionsStore.allFilterCategoryIds);

    const queryAllFilterAccountIds = computed<Record<string, boolean>>(() => transactionsStore.allFilterAccountIds);

    const queryAllFilterTagIds = computed<Record<string, boolean>>(() => transactionsStore.allFilterTagIds);

    const queryAllFilterCategoryIdsCount = computed<number>(() => transactionsStore.allFilterCategoryIdsCount);

    const queryAllFilterAccountIdsCount = computed<number>(() => transactionsStore.allFilterAccountIdsCount);

    const queryAllFilterTagIdsCount = computed<number>(() => transactionsStore.allFilterTagIdsCount);

    const queryAccountName = computed<string>(() => {
        if (queryAllFilterAccountIdsCount.value > 1) {
            return tt('Multiple Accounts');
        }

        return allAccountsMap.value[query.value.accountIds]?.name || tt('Account');
    });

    const queryCategoryName = computed<string>(() => {
        if (queryAllFilterCategoryIdsCount.value > 1) {
            return tt('Multiple Categories');
        }

        return allCategories.value[query.value.categoryIds]?.name || tt('Category');
    });

    const queryTagName = computed<string>(() => {
        if (query.value.tagFilter === TransactionTagFilter.TransactionNoTagFilterValue) {
            return tt('Without Tags');
        }

        if (queryAllFilterTagIdsCount.value > 1) {
            return tt('Multiple Tags');
        }

        for (const tagId of keys(queryAllFilterTagIds.value)) {
            const tagName = allTransactionTags.value[tagId]?.name;

            if (tagName) {
                return tagName;
            }
        }

        return tt('Tags');
    });

    const queryAmount = computed<string>(() => {
        if (!query.value.amountFilter) {
            return '';
        }

        const amountFilterItems = query.value.amountFilter.split(':');

        if (amountFilterItems.length < 2) {
            return '';
        }

        const displayAmount: string[] = [];

        for (let i = 1; i < amountFilterItems.length; i++) {
            displayAmount.push(formatAmountToLocalizedNumeralsWithCurrency(parseBigDecimal(amountFilterItems[i] as string), false));
        }

        if (displayAmount.length === 1) {
            return displayAmount[0]!;
        } else if (displayAmount.length === 2) {
            return formatRange(displayAmount[0]!, displayAmount[1]!);
        } else {
            return '';
        }
    });

    const transactionCalendarMinDate = computed<Date>(() => getLocalDatetimeFromUnixTime(getSameDateTimeWithBrowserTimezone(parseDateTimeFromUnixTime(query.value.minTime)).getUnixTime()));

    const transactionCalendarMaxDate = computed<Date>(() => getLocalDatetimeFromUnixTime(getSameDateTimeWithBrowserTimezone(parseDateTimeFromUnixTime(query.value.maxTime)).getUnixTime()));

    const currentMonthTransactionData = computed<TransactionMonthList | null>(() => {
        const allTransactions = transactionsStore.transactions;

        if (!allTransactions || !allTransactions.length) {
            return null;
        }

        const currentMonthMinDate = parseDateTimeFromUnixTime(query.value.minTime);
        const currentYear = currentMonthMinDate.getGregorianCalendarYear();
        const currentMonth = currentMonthMinDate.getGregorianCalendarMonth();

        for (const transactionMonth of allTransactions) {
            if (transactionMonth.year === currentYear && transactionMonth.month === currentMonth) {
                return transactionMonth;
            }
        }

        return null;
    });

    const canAddTransaction = computed<boolean>(() => {
        if (query.value.accountIds && queryAllFilterAccountIdsCount.value === 1) {
            const account = allAccountsMap.value[query.value.accountIds];

            if (account && account.type === AccountType.MultiSubAccounts.type) {
                return false;
            }
        }

        return true;
    });

    return {
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
    };
}
