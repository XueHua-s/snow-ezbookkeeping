import { defineStore } from 'pinia';
import { useSettingAppearance } from './setting/appearance.ts';
import { useSettingCloud } from './setting/cloud.ts';
import { useSettingState } from './setting/state.ts';
import { useSettingStatistics } from './setting/statistics.ts';
import { useSettingSynchronization } from './setting/synchronization.ts';
import { useSettingTransactions } from './setting/transactions.ts';








export const useSettingsStore = defineStore('settings', () => {
    const { appSettings, syncedAppSettings, localeDefaultSettings, enableApplicationCloudSync, chartColorList, accountCategoryDisplayOrders } = useSettingState();
    const { updateApplicationSettingsValueAndAppSettingsFromCloudSetting, createUserApplicationCloudSetting, updateUserApplicationCloudSettingValue } = useSettingCloud({ appSettings, syncedAppSettings });
    const {
        setTheme, setFontSize, setTimeZone, setAutoUpdateExchangeRatesData, setShowAccountBalance, setAccountCategoryOrders,
        setChartColors, setEnableSwipeBack, setEnableAnimate, setEnableApplicationLock, setEnableApplicationLockWebAuthn,
        setShowAddTransactionButtonInDesktopNavbar, setDesktopOverviewPageLayout, setMobileOverviewPageLayout,
        setShowAmountInHomePage, setTimezoneUsedForStatisticsInHomePage, setOverviewAccountFilterInHomePage,
        setOverviewTransactionCategoryFilterInHomePage, setOverviewTransactionTagFilterInHomePage,
    } = useSettingAppearance({ appSettings, updateUserApplicationCloudSettingValue });
    const {
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
    } = useSettingTransactions({ appSettings, updateUserApplicationCloudSettingValue });
    const {
        setStatisticsDefaultChartDataType, setStatisticsDefaultTimezoneType, setStatisticsDefaultAccountFilter,
        setStatisticsDefaultTransactionCategoryFilter, setStatisticsDefaultKeywordMatchMode, setStatisticsSortingType,
        setStatisticsDefaultCategoricalChartType, setStatisticsDefaultCategoricalChartDateRange,
        setStatisticsDefaultTrendChartType, setStatisticsDefaultTrendChartDateRange, setStatisticsDefaultAssetTrendsChartType,
        setStatisticsDefaultAssetTrendsChartDateRange,
    } = useSettingStatistics({ appSettings, updateUserApplicationCloudSettingValue });
    const { clearAppSettings, createApplicationCloudSettings, setApplicationSettingsFromCloudSettings, updateApplicationSyncSettingKeys, updateLocalizedDefaultSettings } = useSettingSynchronization({ appSettings, createUserApplicationCloudSetting, syncedAppSettings, updateApplicationSettingsValueAndAppSettingsFromCloudSetting, localeDefaultSettings });

    return {
        // states
        appSettings,
        syncedAppSettings,
        localeDefaultSettings,
        // computed states
        enableApplicationCloudSync,
        accountCategoryDisplayOrders,
        chartColorList,
        // functions
        // -- General Settings
        setTheme,
        setFontSize,
        setTimeZone,
        setAutoUpdateExchangeRatesData,
        setShowAccountBalance,
        setAccountCategoryOrders,
        setChartColors,
        setEnableSwipeBack,
        setEnableAnimate,
        // -- Application Lock
        setEnableApplicationLock,
        setEnableApplicationLockWebAuthn,
        // -- Navigation Bar
        setShowAddTransactionButtonInDesktopNavbar,
        // -- Overview Page
        setDesktopOverviewPageLayout,
        setMobileOverviewPageLayout,
        setShowAmountInHomePage,
        setTimezoneUsedForStatisticsInHomePage,
        setOverviewAccountFilterInHomePage,
        setOverviewTransactionCategoryFilterInHomePage,
        setOverviewTransactionTagFilterInHomePage,
        // -- Transaction List Page
        setItemsCountInTransactionListPage,
        setShowTotalAmountInTransactionListPage,
        setTotalAmountTypeInTransactionListPage,
        setShowTagInTransactionListPage,
        setDefaultKeywordMatchModeInTransactionListPage,
        // -- Transaction Edit Page
        setQuickSaveButtonStyleInMobileTransactionListPage,
        setQuickAddButtonActionInMobileTransactionEditPage,
        setAutoSaveTransactionDraft,
        setAutoGetCurrentGeoLocation,
        setAlwaysShowTransactionPicturesInMobileTransactionEditPage,
        setTransactionPictureQuality,
        // -- AI Clipboard Text Recognition
        setAlwaysRequireConfirmationOfClipboardContentBeforeSubmission,
        // -- AI Image Recognition
        setAutoUploadTransactionPictureForAIRecognition,
        // -- Import Transaction Dialog
        setRememberLastSelectedFileTypeInImportTransactionDialog,
        setLastSelectedFileTypeInImportTransactionDialog,
        // -- Insights Explorer Page
        setInsightsExplorerDefaultDateRangeType,
        setShowTagInInsightsExplorerPage,
        // -- Account List Page
        setTotalAmountExcludeAccountIds,
        setHideCategoriesWithoutAccounts,
        setDefaultCreditCardAmountDisplayTypeInMobile,
        setReconciliationStatementButtonDefaultDateRangeTypeInDesktop,
        setReconciliationStatementPageDefaultDateRangeTypeInMobile,
        // -- Exchange Rates Data Page
        setCurrencySortByInExchangeRatesPage,
        // -- Browser Cache Management
        setMapCacheExpiration,
        setExchangeRatesDataCacheExpiration,
        // -- Statistics Settings
        setStatisticsDefaultChartDataType,
        setStatisticsDefaultTimezoneType,
        setStatisticsDefaultAccountFilter,
        setStatisticsDefaultTransactionCategoryFilter,
        setStatisticsDefaultKeywordMatchMode,
        setStatisticsSortingType,
        setStatisticsDefaultCategoricalChartType,
        setStatisticsDefaultCategoricalChartDateRange,
        setStatisticsDefaultTrendChartType,
        setStatisticsDefaultTrendChartDateRange,
        setStatisticsDefaultAssetTrendsChartType,
        setStatisticsDefaultAssetTrendsChartDateRange,
        clearAppSettings,
        createApplicationCloudSettings,
        setApplicationSettingsFromCloudSettings,
        updateApplicationSyncSettingKeys,
        updateLocalizedDefaultSettings
    };
});
