import { useTransactionTagSelectionBase } from '@/components/base/TransactionTagSelectionBase.ts';
import { CategoryType } from '@/core/category.ts';
import { AmountFilterType, DecimalSeparator } from '@/core/numeral.ts';
import { getAccountMapByName } from '@/lib/account.ts';
import {
    getSecondaryTransactionMapByName
} from '@/lib/category.ts';
import { useI18n } from '@/locales/helpers.ts';
import { Account, type CategorizedAccountWithDisplayBalance } from '@/models/account.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import { TransactionTag } from '@/models/transaction_tag.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useUserStore } from '@/stores/user.ts';
import { computed, ref, useTemplateRef } from 'vue';
import { type BatchApplyRulesDialogType, type BatchCreateDialogType, type BatchReplaceDialogType, type ImportTransactionCheckDataFilter, type SnackBarType } from './types.ts';


export function useImportTransactionCheckDataTabState() {
    const {
        tt,
        formatRange,
        getCurrentDecimalSeparator,
        formatDateTimeToLongDateTime,
        formatDateTimeToGregorianDefaultDateTime,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping,
        formatAmountToLocalizedNumeralsWithCurrency,
        formatNumberToLocalizedNumerals,
        formatNumberToWesternArabicNumeralsWithoutDigitGrouping,
        getCategorizedAccountsWithDisplayBalance,
        getTablePageOptions,
        formatCoordinate
    } = useI18n();

    const { allTagsWithGroupHeader } = useTransactionTagSelectionBase({ modelValue: [] }, false);

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const batchReplaceDialog = useTemplateRef<BatchReplaceDialogType>('batchReplaceDialog');

    const batchApplyRulesDialog = useTemplateRef<BatchApplyRulesDialogType>('batchApplyRulesDialog');

    const batchCreateDialog = useTemplateRef<BatchCreateDialogType>('batchCreateDialog');

    const editingTransaction = ref<ImportTransaction | null>(null);

    const editingTags = ref<string[]>([]);

    const filters = ref<ImportTransactionCheckDataFilter>({
        minDatetime: null,
        maxDatetime: null,
        transactionType: null,
        category: null,
        amount: null,
        account: null,
        tag: null,
        description: null
    });

    const currentPage = ref<number>(1);

    const countPerPage = ref<number>(10);

    const showCustomDateRangeDialog = ref<boolean>(false);

    const showCustomAmountFilterDialog = ref<boolean>(false);

    const showCustomDescriptionDialog = ref<boolean>(false);

    const currentAmountFilterType = ref<AmountFilterType | null>(null);

    const currentAmountFilterValue1 = ref<number>(0);

    const currentAmountFilterValue2 = ref<number>(0);

    const currentDescriptionFilterValue = ref<string | null>(null);

    const showAccountBalance = computed<boolean>(() => settingsStore.appSettings.showAccountBalance);

    const customAccountCategoryOrder = computed<string>(() => settingsStore.appSettings.accountCategoryOrders);

    const useCommaDecimalSeparator = computed<boolean>(() => getCurrentDecimalSeparator() === DecimalSeparator.Comma.symbol);

    const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

    const allAccounts = computed<Account[]>(() => accountsStore.allPlainAccounts);

    const allVisibleAccounts = computed<Account[]>(() => accountsStore.allVisiblePlainAccounts);

    const allVisibleCategorizedAccounts = computed<CategorizedAccountWithDisplayBalance[]>(() => getCategorizedAccountsWithDisplayBalance(allVisibleAccounts.value, showAccountBalance.value, customAccountCategoryOrder.value));

    const allAccountsMap = computed<Record<string, Account>>(() => accountsStore.allAccountsMap);

    const allAccountsMapByName = computed<Record<string, Account>>(() => getAccountMapByName(accountsStore.allAccounts));

    const allCategories = computed<Record<number, TransactionCategory[]>>(() => transactionCategoriesStore.allTransactionCategories);

    const allCategoriesMap = computed<Record<string, TransactionCategory>>(() => transactionCategoriesStore.allTransactionCategoriesMap);

    const allSecondaryCategoriesMapByName = computed<Record<number, Record<string, TransactionCategory>>>(() => ({
        [CategoryType.Expense]: getSecondaryTransactionMapByName(allCategories.value[CategoryType.Expense]),
        [CategoryType.Income]: getSecondaryTransactionMapByName(allCategories.value[CategoryType.Income]),
        [CategoryType.Transfer]: getSecondaryTransactionMapByName(allCategories.value[CategoryType.Transfer])
    }));

    const allTagsMap = computed<Record<string, TransactionTag>>(() => transactionTagsStore.allTransactionTagsMap);

    const hasVisibleExpenseCategories = computed<boolean>(() => transactionCategoriesStore.hasVisibleExpenseCategories);

    const hasVisibleIncomeCategories = computed<boolean>(() => transactionCategoriesStore.hasVisibleIncomeCategories);

    const hasVisibleTransferCategories = computed<boolean>(() => transactionCategoriesStore.hasVisibleTransferCategories);

    const isEditing = computed<boolean>(() => !!editingTransaction.value);

    return {
        tt, formatRange, formatDateTimeToLongDateTime, formatDateTimeToGregorianDefaultDateTime,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, formatAmountToLocalizedNumeralsWithCurrency,
        formatNumberToLocalizedNumerals, formatNumberToWesternArabicNumeralsWithoutDigitGrouping, getTablePageOptions,
        formatCoordinate, allTagsWithGroupHeader, exchangeRatesStore, snackbar, batchReplaceDialog, batchApplyRulesDialog,
        batchCreateDialog, editingTransaction, editingTags, filters, currentPage, countPerPage, showCustomDateRangeDialog,
        showCustomAmountFilterDialog, showCustomDescriptionDialog, currentAmountFilterType, currentAmountFilterValue1,
        currentAmountFilterValue2, currentDescriptionFilterValue, useCommaDecimalSeparator, defaultCurrency, allAccounts,
        allVisibleAccounts, allVisibleCategorizedAccounts, allAccountsMap, allAccountsMapByName, allCategories, allCategoriesMap,
        allSecondaryCategoriesMapByName, allTagsMap, hasVisibleExpenseCategories, hasVisibleIncomeCategories,
        hasVisibleTransferCategories, isEditing,
    };
}
