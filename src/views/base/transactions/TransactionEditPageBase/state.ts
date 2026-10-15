import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';
import { TRANSACTION_MAX_PICTURE_COUNT } from '@/consts/transaction.ts';
import type { WeekDayValue } from '@/core/datetime.ts';
import { ImageUploadQualityType } from '@/core/image.ts';
import type { BigDecimal, NumeralSystem } from '@/core/numeral.ts';
import { TemplateType } from '@/core/template.ts';
import type { LocalizedTimezoneInfo } from '@/core/timezone.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    isArray
} from '@/lib/common.ts';
import {
    getBrowserTimezoneName,
    getCurrentUnixTime,
    getSameDateTimeWithCurrentTimezone,
    getTimezoneOffsetMinutes,
    parseDateTimeFromUnixTimeWithBrowserTimezone
} from '@/lib/datetime.ts';
import { useI18n } from '@/locales/helpers.ts';
import { Account, type CategorizedAccountWithDisplayBalance } from '@/models/account.ts';
import { Transaction } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionTag } from '@/models/transaction_tag.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useUserStore } from '@/stores/user.ts';
import { computed, ref } from 'vue';
import { GeoLocationStatus, TransactionEditPageMode, TransactionEditPageType } from './types.ts';


interface Context {
    initMode: TransactionEditPageMode | undefined;
    transactionDefaultType: number | undefined;
    type: TransactionEditPageType;
}

// Owns state for the TransactionEditPageBase store.
export function useTransactionEditPageBaseState(context: Context) {
    const { initMode, transactionDefaultType, type } = context;

    const {
        tt,
        getAllTimezones,
        getCurrentNumeralSystemType,
        getTimezoneDifferenceDisplayText,
        formatAmountToLocalizedNumeralsWithCurrency,
        formatNumberToLocalizedNumerals,
        getAdaptiveAmountRate,
        getCategorizedAccountsWithDisplayBalance
    } = useI18n();

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const transactionsStore = useTransactionsStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const isSupportGeoLocation: boolean = !!navigator.geolocation;

    const mode = ref<TransactionEditPageMode>(initMode ?? TransactionEditPageMode.Add);

    const editId = ref<string | null>(null);

    const addByTemplateId = ref<string | null>(null);

    const duplicateFromId = ref<string | null>(null);

    const clientSessionId = ref<string>('');

    const loading = ref<boolean>(true);

    const recognizing = ref<boolean>(false);

    const submitting = ref<boolean>(false);

    const submitted = ref<boolean>(false);

    const uploadingPicture = ref<boolean>(false);

    const geoLocationStatus = ref<GeoLocationStatus | null>(null);

    const setGeoLocationByClickMap = ref<boolean>(false);

    const transaction = ref<Transaction | TransactionTemplate>(createNewTransactionModel(transactionDefaultType));

    const numeralSystem = computed<NumeralSystem>(() => getCurrentNumeralSystemType());

    const currentTimezoneOffsetMinutes = computed<number>(() => getTimezoneOffsetMinutes(transaction.value.time));

    const showAccountBalance = computed<boolean>(() => settingsStore.appSettings.showAccountBalance);

    const customAccountCategoryOrder = computed<string>(() => settingsStore.appSettings.accountCategoryOrders);

    const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

    const defaultAccountId = computed<string>(() => userStore.currentUserDefaultAccountId);

    const firstDayOfWeek = computed<WeekDayValue>(() => userStore.currentUserFirstDayOfWeek);

    const imageUploadQualityType = computed<ImageUploadQualityType>(() => ImageUploadQualityType.valueOf(settingsStore.appSettings.transactionPictureQuality) ?? ImageUploadQualityType.Default);

    const allTimezones = computed<LocalizedTimezoneInfo[]>(() => {
        if (type === TransactionEditPageType.Template && transaction.value instanceof TransactionTemplate) {
            return getAllTimezones(getCurrentUnixTime(), false);
        } else {
            return getAllTimezones(transaction.value.time, true)
        }
    });

    const allAccounts = computed<Account[]>(() => accountsStore.allPlainAccounts);

    const allVisibleAccounts = computed<Account[]>(() => accountsStore.allVisiblePlainAccounts);

    const allAccountsMap = computed<Record<string, Account>>(() => accountsStore.allAccountsMap);

    const allVisibleCategorizedAccounts = computed<CategorizedAccountWithDisplayBalance[]>(() => getCategorizedAccountsWithDisplayBalance(allVisibleAccounts.value, showAccountBalance.value, customAccountCategoryOrder.value));

    const allCategories = computed<Record<number, TransactionCategory[]>>(() => transactionCategoriesStore.allTransactionCategories);

    const allCategoriesMap = computed<Record<string, TransactionCategory>>(() => transactionCategoriesStore.allTransactionCategoriesMap);

    const allTagsMap = computed<Record<string, TransactionTag>>(() => transactionTagsStore.allTransactionTagsMap);

    const firstVisibleAccountId = computed<string | undefined>(() => allVisibleAccounts.value && allVisibleAccounts.value[0] ? allVisibleAccounts.value[0].id : undefined);

    const hasVisibleExpenseCategories = computed<boolean>(() => transactionCategoriesStore.hasVisibleExpenseCategories);

    const hasVisibleIncomeCategories = computed<boolean>(() => transactionCategoriesStore.hasVisibleIncomeCategories);

    const hasVisibleTransferCategories = computed<boolean>(() => transactionCategoriesStore.hasVisibleTransferCategories);

    const canAddTransactionPicture = computed<boolean>(() => {
        if (type !== TransactionEditPageType.Transaction || (mode.value !== TransactionEditPageMode.Add && mode.value !== TransactionEditPageMode.Edit)) {
            return false;
        }

        return !isArray(transaction.value.pictures) || transaction.value.pictures.length < TRANSACTION_MAX_PICTURE_COUNT;
    });

    function getCurrentUnixTimeForNewTransaction(): number {
        return getSameDateTimeWithCurrentTimezone(parseDateTimeFromUnixTimeWithBrowserTimezone(getCurrentUnixTime())).getUnixTime();
    }

    function isNormalTransactionTemplate(item: Transaction | TransactionTemplate): boolean {
        return type === TransactionEditPageType.Template && item instanceof TransactionTemplate && item.templateType === TemplateType.Normal.type;
    }

    function isScheduledTransactionTemplate(item: Transaction | TransactionTemplate): boolean {
        return type === TransactionEditPageType.Template && item instanceof TransactionTemplate && item.templateType === TemplateType.Schedule.type;
    }

    function createNewTransactionModel(transactionType?: number): Transaction | TransactionTemplate {
        const now: number = getCurrentUnixTimeForNewTransaction();
        const currentTimezone: string = settingsStore.appSettings.timeZone;

        let defaultType: TransactionType = TransactionType.Expense;

        if (transactionType === TransactionType.Income) {
            defaultType = TransactionType.Income;
        } else if (transactionType === TransactionType.Transfer) {
            defaultType = TransactionType.Transfer;
        }

        let newTransaction: Transaction | TransactionTemplate = Transaction.createNewTransaction(defaultType, now, currentTimezone, getTimezoneOffsetMinutes(now, currentTimezone));

        if (type === TransactionEditPageType.Template) {
            newTransaction = TransactionTemplate.createNewTransactionTemplate(newTransaction);
            newTransaction.timeZone = currentTimezone || getBrowserTimezoneName();
        }

        return newTransaction;
    }

    function getDisplayAmount(amount: BigDecimal, hideAmount: boolean, currencyCode: string): string {
        if (hideAmount) {
            return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, currencyCode);
        }

        return formatAmountToLocalizedNumeralsWithCurrency(amount, currencyCode);
    }

    return {
        tt, getTimezoneDifferenceDisplayText, formatNumberToLocalizedNumerals, getAdaptiveAmountRate, settingsStore,
        transactionsStore, exchangeRatesStore, isSupportGeoLocation, mode, editId, addByTemplateId, duplicateFromId,
        clientSessionId, loading, recognizing, submitting, submitted, uploadingPicture, geoLocationStatus,
        setGeoLocationByClickMap, transaction, numeralSystem, currentTimezoneOffsetMinutes, showAccountBalance, defaultCurrency,
        defaultAccountId, firstDayOfWeek, imageUploadQualityType, allTimezones, allAccounts, allVisibleAccounts, allAccountsMap,
        allVisibleCategorizedAccounts, allCategories, allCategoriesMap, allTagsMap, firstVisibleAccountId,
        hasVisibleExpenseCategories, hasVisibleIncomeCategories, hasVisibleTransferCategories, canAddTransactionPicture,
        isNormalTransactionTemplate, isScheduledTransactionTemplate, createNewTransactionModel, getDisplayAmount,
    };
}
