import type { TypeAndDisplayName } from '@/core/base.ts';
import { DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_MOBILE } from '@/core/statistics.ts';
import { isDefaultMobileOverviewLayout, parseMobileOverviewLayout } from '@/lib/overview_layout.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useI18n } from '@/locales/helpers.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useAppSettingPageBase } from '@/views/base/settings/AppSettingsPageBase.ts';
import { computed, ref } from 'vue';



// Coordinates PreferencesSettingsPage state and effects; called once from the component setup.
export function usePreferencesSettingsPage() {
    const {
        tt,
        getAllCreditCardAmountDisplayTypes,
        getAllTransactionQuickSaveButtonStyles,
        getAllTransactionQuickAddButtonActionTypes
    } = useI18n();

    const { showToast } = useI18nUIComponents();

    const {
        loadingAccounts,
        loadingTransactionCategories,
        loadingTags,
        hasAnyAccount,
        hasAnyVisibleAccount,
        hasAnyTransactionCategory,
        hasAnyTransactionTag,
        allTimezoneTypesUsedForStatistics,
        allCurrencySortingTypes,
        allKeywordMatchModes,
        allAutoSaveTransactionDraftTypes,
        allImageUploadQualityTypes,
        allReconciliationStatementDateRanges,
        isAutoUpdateExchangeRatesData,
        showAccountBalance,
        showAmountInHomePage,
        timezoneUsedForStatisticsInHomePage,
        showTotalAmountInTransactionListPage,
        totalAmountTypeInTransactionListPage,
        showTagInTransactionListPage,
        defaultKeywordMatchModeInTransactionListPage,
        autoSaveTransactionDraft,
        isAutoGetCurrentGeoLocation,
        transactionPictureQuality,
        isAlwaysRequireConfirmationOfClipboardContentBeforeSubmission,
        isAutoUploadTransactionPictureForAIRecognition,
        currencySortByInExchangeRatesPage,
        chartColorSchemeContent,
        accountsIncludedInHomePageOverviewDisplayContent,
        accountsIncludedInTotalDisplayContent,
        accountCategorysDisplayOrderContent,
        transactionCategoriesIncludedInHomePageOverviewDisplayContent,
        getValidReconciliationStatementPageDefaultDateRangeType
    } = useAppSettingPageBase();

    const settingsStore = useSettingsStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const showTimezoneUsedForStatisticsInHomePagePopup = ref<boolean>(false);

    const showTransactionPictureQualityPopup = ref<boolean>(false);

    const showReconciliationStatementDefaultDateRangePopup = ref<boolean>(false);

    const allTransactionQuickSaveButtonStyles = computed<TypeAndDisplayName[]>(() => getAllTransactionQuickSaveButtonStyles());

    const allTransactionQuickAddButtonActionTypes = computed<TypeAndDisplayName[]>(() => getAllTransactionQuickAddButtonActionTypes());

    const allCreditCardAmountDisplayTypes = computed<TypeAndDisplayName[]>(() => getAllCreditCardAmountDisplayTypes());

    const overviewPageLayoutDisplayContent = computed<string>(() => {
        try {
            return tt(isDefaultMobileOverviewLayout(parseMobileOverviewLayout(settingsStore.appSettings.mobileOverviewPageLayout)) ? 'Default' : 'Custom');
        } catch {
            return tt('Custom');
        }
    });

    const quickSaveButtonStyleInMobileTransactionListPage = computed<number>({
        get: () => settingsStore.appSettings.quickSaveButtonStyleInMobileTransactionListPage,
        set: (value) => settingsStore.setQuickSaveButtonStyleInMobileTransactionListPage(value)
    });

    const quickAddButtonActionInMobileTransactionEditPage = computed<number>({
        get: () => settingsStore.appSettings.quickAddButtonActionInMobileTransactionEditPage,
        set: (value) => settingsStore.setQuickAddButtonActionInMobileTransactionEditPage(value)
    });

    const alwaysShowTransactionPicturesInMobileTransactionEditPage = computed<boolean>({
        get: () => settingsStore.appSettings.alwaysShowTransactionPicturesInMobileTransactionEditPage,
        set: (value) => settingsStore.setAlwaysShowTransactionPicturesInMobileTransactionEditPage(value)
    });

    const defaultCreditCardAmountDisplayTypeInMobile = computed<number>({
        get: () => settingsStore.appSettings.defaultCreditCardAmountDisplayTypeInMobile,
        set: (value: number) => settingsStore.setDefaultCreditCardAmountDisplayTypeInMobile(value)
    });

    const reconciliationStatementPageDefaultDateRangeTypeInMobile = computed<number>({
        get: () => getValidReconciliationStatementPageDefaultDateRangeType(settingsStore.appSettings.reconciliationStatementPageDefaultDateRangeTypeInMobile, DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_MOBILE.type),
        set: (value: number) => settingsStore.setReconciliationStatementPageDefaultDateRangeTypeInMobile(value)
    });

    function init(): void {
        loadingAccounts.value = true;
        loadingTransactionCategories.value = true;
        loadingTags.value = true;

        accountsStore.loadAllAccounts({
            force: false
        }).then(() => {
            loadingAccounts.value = false;
        }).catch(error => {
            loadingAccounts.value = false;

            if (!error.processed) {
                showToast(error.message || error);
            }
        });

        transactionCategoriesStore.loadAllCategories({
            force: false
        }).then(() => {
            loadingTransactionCategories.value = false;
        }).catch(error => {
            loadingTransactionCategories.value = false;

            if (!error.processed) {
                showToast(error.message || error);
            }
        });

        transactionTagsStore.loadAllTags({
            force: false
        }).then(() => {
            loadingTags.value = false;
        }).catch(error => {
            loadingTags.value = false;

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    init();

    return {
        tt, loadingAccounts, loadingTransactionCategories, loadingTags, hasAnyAccount, hasAnyVisibleAccount,
        hasAnyTransactionCategory, hasAnyTransactionTag, allTimezoneTypesUsedForStatistics, allCurrencySortingTypes,
        allKeywordMatchModes, allAutoSaveTransactionDraftTypes, allImageUploadQualityTypes, allReconciliationStatementDateRanges,
        isAutoUpdateExchangeRatesData, showAccountBalance, showAmountInHomePage, timezoneUsedForStatisticsInHomePage,
        showTotalAmountInTransactionListPage, totalAmountTypeInTransactionListPage, showTagInTransactionListPage,
        defaultKeywordMatchModeInTransactionListPage, autoSaveTransactionDraft, isAutoGetCurrentGeoLocation,
        transactionPictureQuality, isAlwaysRequireConfirmationOfClipboardContentBeforeSubmission,
        isAutoUploadTransactionPictureForAIRecognition, currencySortByInExchangeRatesPage, chartColorSchemeContent,
        accountsIncludedInHomePageOverviewDisplayContent, accountsIncludedInTotalDisplayContent,
        accountCategorysDisplayOrderContent, transactionCategoriesIncludedInHomePageOverviewDisplayContent, settingsStore,
        showTimezoneUsedForStatisticsInHomePagePopup, showTransactionPictureQualityPopup,
        showReconciliationStatementDefaultDateRangePopup, allTransactionQuickSaveButtonStyles,
        allTransactionQuickAddButtonActionTypes, allCreditCardAmountDisplayTypes, overviewPageLayoutDisplayContent,
        quickSaveButtonStyleInMobileTransactionListPage, quickAddButtonActionInMobileTransactionEditPage,
        alwaysShowTransactionPicturesInMobileTransactionEditPage, defaultCreditCardAmountDisplayTypeInMobile,
        reconciliationStatementPageDefaultDateRangeTypeInMobile,
    };
}
