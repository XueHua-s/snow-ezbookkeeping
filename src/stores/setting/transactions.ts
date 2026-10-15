import {
    updateApplicationSettingsValue
} from '@/lib/settings.ts';
import type { useSettingCloud } from './cloud.ts';
import type { useSettingState } from './state.ts';

interface Context {
    appSettings: ReturnType<typeof useSettingState>['appSettings'];
    updateUserApplicationCloudSettingValue: ReturnType<typeof useSettingCloud>['updateUserApplicationCloudSettingValue'];
}

// Owns transactions for the setting store.
export function useSettingTransactions(context: Context) {
    const { appSettings, updateUserApplicationCloudSettingValue } = context;

    // Transaction List Page
    function setItemsCountInTransactionListPage(value: number): void {
        updateApplicationSettingsValue('itemsCountInTransactionListPage', value);
        appSettings.value.itemsCountInTransactionListPage = value;
        updateUserApplicationCloudSettingValue('itemsCountInTransactionListPage', value);
    }

    function setShowTotalAmountInTransactionListPage(value: boolean): void {
        updateApplicationSettingsValue('showTotalAmountInTransactionListPage', value);
        appSettings.value.showTotalAmountInTransactionListPage = value;
        updateUserApplicationCloudSettingValue('showTotalAmountInTransactionListPage', value);
    }

    function setTotalAmountTypeInTransactionListPage(value: number): void {
        updateApplicationSettingsValue('totalAmountTypeInTransactionListPage', value);
        appSettings.value.totalAmountTypeInTransactionListPage = value;
        updateUserApplicationCloudSettingValue('totalAmountTypeInTransactionListPage', value);
    }

    function setShowTagInTransactionListPage(value: boolean): void {
        updateApplicationSettingsValue('showTagInTransactionListPage', value);
        appSettings.value.showTagInTransactionListPage = value;
        updateUserApplicationCloudSettingValue('showTagInTransactionListPage', value);
    }

    function setDefaultKeywordMatchModeInTransactionListPage(value: number): void {
        updateApplicationSettingsValue('defaultKeywordMatchModeInTransactionListPage', value);
        appSettings.value.defaultKeywordMatchModeInTransactionListPage = value;
        updateUserApplicationCloudSettingValue('defaultKeywordMatchModeInTransactionListPage', value);
    }

    // Transaction Edit Page
    function setQuickSaveButtonStyleInMobileTransactionListPage(value: number): void {
        updateApplicationSettingsValue('quickSaveButtonStyleInMobileTransactionListPage', value);
        appSettings.value.quickSaveButtonStyleInMobileTransactionListPage = value;
        updateUserApplicationCloudSettingValue('quickSaveButtonStyleInMobileTransactionListPage', value);
    }

    function setQuickAddButtonActionInMobileTransactionEditPage(value: number): void {
        updateApplicationSettingsValue('quickAddButtonActionInMobileTransactionEditPage', value);
        appSettings.value.quickAddButtonActionInMobileTransactionEditPage = value;
        updateUserApplicationCloudSettingValue('quickAddButtonActionInMobileTransactionEditPage', value);
    }

    function setAutoSaveTransactionDraft(value: string): void {
        updateApplicationSettingsValue('autoSaveTransactionDraft', value);
        appSettings.value.autoSaveTransactionDraft = value;
        updateUserApplicationCloudSettingValue('autoSaveTransactionDraft', value);
    }

    function setAutoGetCurrentGeoLocation(value: boolean): void {
        updateApplicationSettingsValue('autoGetCurrentGeoLocation', value);
        appSettings.value.autoGetCurrentGeoLocation = value;
        updateUserApplicationCloudSettingValue('autoGetCurrentGeoLocation', value);
    }

    function setAlwaysShowTransactionPicturesInMobileTransactionEditPage(value: boolean): void {
        updateApplicationSettingsValue('alwaysShowTransactionPicturesInMobileTransactionEditPage', value);
        appSettings.value.alwaysShowTransactionPicturesInMobileTransactionEditPage = value;
        updateUserApplicationCloudSettingValue('alwaysShowTransactionPicturesInMobileTransactionEditPage', value);
    }

    function setTransactionPictureQuality(value: number): void {
        updateApplicationSettingsValue('transactionPictureQuality', value);
        appSettings.value.transactionPictureQuality = value;
        updateUserApplicationCloudSettingValue('transactionPictureQuality', value);
    }

    // AI Clipboard Text Recognition
    function setAlwaysRequireConfirmationOfClipboardContentBeforeSubmission(value: boolean): void {
        updateApplicationSettingsValue('alwaysRequireConfirmationOfClipboardContentBeforeSubmission', value);
        appSettings.value.alwaysRequireConfirmationOfClipboardContentBeforeSubmission = value;
        updateUserApplicationCloudSettingValue('alwaysRequireConfirmationOfClipboardContentBeforeSubmission', value);
    }

    // AI Image Recognition
    function setAutoUploadTransactionPictureForAIRecognition(value: boolean): void {
        updateApplicationSettingsValue('autoUploadTransactionPictureForAIRecognition', value);
        appSettings.value.autoUploadTransactionPictureForAIRecognition = value;
        updateUserApplicationCloudSettingValue('autoUploadTransactionPictureForAIRecognition', value);
    }

    // Import Transaction Dialog
    function setRememberLastSelectedFileTypeInImportTransactionDialog(value: boolean): void {
        updateApplicationSettingsValue('rememberLastSelectedFileTypeInImportTransactionDialog', value);
        appSettings.value.rememberLastSelectedFileTypeInImportTransactionDialog = value;
        updateUserApplicationCloudSettingValue('rememberLastSelectedFileTypeInImportTransactionDialog', value);

        if (!value) {
            setLastSelectedFileTypeInImportTransactionDialog('');
        }
    }

    function setLastSelectedFileTypeInImportTransactionDialog(value: string): void {
        if (!appSettings.value.rememberLastSelectedFileTypeInImportTransactionDialog) {
            value = '';
        }

        updateApplicationSettingsValue('lastSelectedFileTypeInImportTransactionDialog', value);
        appSettings.value.lastSelectedFileTypeInImportTransactionDialog = value;
        updateUserApplicationCloudSettingValue('lastSelectedFileTypeInImportTransactionDialog', value);
    }

    // Insights Explorer Page
    function setInsightsExplorerDefaultDateRangeType(value: number): void {
        updateApplicationSettingsValue('insightsExplorerDefaultDateRangeType', value);
        appSettings.value.insightsExplorerDefaultDateRangeType = value;
        updateUserApplicationCloudSettingValue('insightsExplorerDefaultDateRangeType', value);
    }

    function setShowTagInInsightsExplorerPage(value: boolean): void {
        updateApplicationSettingsValue('showTagInInsightsExplorerPage', value);
        appSettings.value.showTagInInsightsExplorerPage = value;
        updateUserApplicationCloudSettingValue('showTagInInsightsExplorerPage', value);
    }

    // Account List Page
    function setTotalAmountExcludeAccountIds(value: Record<string, boolean>): void {
        updateApplicationSettingsValue('totalAmountExcludeAccountIds', value);
        appSettings.value.totalAmountExcludeAccountIds = value;
        updateUserApplicationCloudSettingValue('totalAmountExcludeAccountIds', value);
    }

    function setHideCategoriesWithoutAccounts(value: boolean): void {
        updateApplicationSettingsValue('hideCategoriesWithoutAccounts', value);
        appSettings.value.hideCategoriesWithoutAccounts = value;
        updateUserApplicationCloudSettingValue('hideCategoriesWithoutAccounts', value);
    }

    function setDefaultCreditCardAmountDisplayTypeInMobile(value: number): void {
        updateApplicationSettingsValue('defaultCreditCardAmountDisplayTypeInMobile', value);
        appSettings.value.defaultCreditCardAmountDisplayTypeInMobile = value;
        updateUserApplicationCloudSettingValue('defaultCreditCardAmountDisplayTypeInMobile', value);
    }

    function setReconciliationStatementButtonDefaultDateRangeTypeInDesktop(value: number): void {
        updateApplicationSettingsValue('reconciliationStatementButtonDefaultDateRangeTypeInDesktop', value);
        appSettings.value.reconciliationStatementButtonDefaultDateRangeTypeInDesktop = value;
        updateUserApplicationCloudSettingValue('reconciliationStatementButtonDefaultDateRangeTypeInDesktop', value);
    }

    function setReconciliationStatementPageDefaultDateRangeTypeInMobile(value: number): void {
        updateApplicationSettingsValue('reconciliationStatementPageDefaultDateRangeTypeInMobile', value);
        appSettings.value.reconciliationStatementPageDefaultDateRangeTypeInMobile = value;
        updateUserApplicationCloudSettingValue('reconciliationStatementPageDefaultDateRangeTypeInMobile', value);
    }

    // Exchange Rates Data Page
    function setCurrencySortByInExchangeRatesPage(value: number): void {
        updateApplicationSettingsValue('currencySortByInExchangeRatesPage', value);
        appSettings.value.currencySortByInExchangeRatesPage = value;
        updateUserApplicationCloudSettingValue('currencySortByInExchangeRatesPage', value);
    }

    // Browser Cache Management
    function setMapCacheExpiration(value: number): void {
        updateApplicationSettingsValue('mapCacheExpiration', value);
        appSettings.value.mapCacheExpiration = value;
        updateUserApplicationCloudSettingValue('mapCacheExpiration', value);
    }

    function setExchangeRatesDataCacheExpiration(value: number): void {
        updateApplicationSettingsValue('exchangeRatesDataCacheExpiration', value);
        appSettings.value.exchangeRatesDataCacheExpiration = value;
        updateUserApplicationCloudSettingValue('exchangeRatesDataCacheExpiration', value);
    }

    return {
        setItemsCountInTransactionListPage, setShowTotalAmountInTransactionListPage, setTotalAmountTypeInTransactionListPage,
        setShowTagInTransactionListPage, setDefaultKeywordMatchModeInTransactionListPage,
        setQuickSaveButtonStyleInMobileTransactionListPage, setQuickAddButtonActionInMobileTransactionEditPage,
        setAutoSaveTransactionDraft, setAutoGetCurrentGeoLocation, setAlwaysShowTransactionPicturesInMobileTransactionEditPage,
        setTransactionPictureQuality, setAlwaysRequireConfirmationOfClipboardContentBeforeSubmission,
        setAutoUploadTransactionPictureForAIRecognition, setRememberLastSelectedFileTypeInImportTransactionDialog,
        setLastSelectedFileTypeInImportTransactionDialog, setInsightsExplorerDefaultDateRangeType,
        setShowTagInInsightsExplorerPage, setTotalAmountExcludeAccountIds, setHideCategoriesWithoutAccounts,
        setDefaultCreditCardAmountDisplayTypeInMobile, setReconciliationStatementButtonDefaultDateRangeTypeInDesktop,
        setReconciliationStatementPageDefaultDateRangeTypeInMobile, setCurrencySortByInExchangeRatesPage, setMapCacheExpiration,
        setExchangeRatesDataCacheExpiration,
    };
}
