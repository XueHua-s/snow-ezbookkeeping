import {
    updateApplicationSettingsValue
} from '@/lib/settings.ts';
import type { useSettingCloud } from './cloud.ts';
import type { useSettingState } from './state.ts';

interface Context {
    appSettings: ReturnType<typeof useSettingState>['appSettings'];
    updateUserApplicationCloudSettingValue: ReturnType<typeof useSettingCloud>['updateUserApplicationCloudSettingValue'];
}

// Owns appearance for the setting store.
export function useSettingAppearance(context: Context) {
    const { appSettings, updateUserApplicationCloudSettingValue } = context;

    // General Settings
    function setTheme(value: string): void {
        updateApplicationSettingsValue('theme', value);
        appSettings.value.theme = value;
    }

    function setFontSize(value: number): void {
        updateApplicationSettingsValue('fontSize', value);
        appSettings.value.fontSize = value;
    }

    function setTimeZone(value: string): void {
        updateApplicationSettingsValue('timeZone', value);
        appSettings.value.timeZone = value;
    }

    function setAutoUpdateExchangeRatesData(value: boolean): void {
        updateApplicationSettingsValue('autoUpdateExchangeRatesData', value);
        appSettings.value.autoUpdateExchangeRatesData = value;
        updateUserApplicationCloudSettingValue('autoUpdateExchangeRatesData', value);
    }

    function setShowAccountBalance(value: boolean): void {
        updateApplicationSettingsValue('showAccountBalance', value);
        appSettings.value.showAccountBalance = value;
        updateUserApplicationCloudSettingValue('showAccountBalance', value);
    }

    function setAccountCategoryOrders(value: string): void {
        updateApplicationSettingsValue('accountCategoryOrders', value);
        appSettings.value.accountCategoryOrders = value;
        updateUserApplicationCloudSettingValue('accountCategoryOrders', value);
    }

    function setChartColors(value: string): void {
        updateApplicationSettingsValue('chartColors', value);
        appSettings.value.chartColors = value;
        updateUserApplicationCloudSettingValue('chartColors', value);
    }

    function setEnableSwipeBack(value: boolean): void {
        updateApplicationSettingsValue('swipeBack', value);
        appSettings.value.swipeBack = value;
    }

    function setEnableAnimate(value: boolean): void {
        updateApplicationSettingsValue('animate', value);
        appSettings.value.animate = value;
    }

    // Application Lock
    function setEnableApplicationLock(value: boolean): void {
        updateApplicationSettingsValue('applicationLock', value);
        appSettings.value.applicationLock = value;
    }

    function setEnableApplicationLockWebAuthn(value: boolean): void {
        updateApplicationSettingsValue('applicationLockWebAuthn', value);
        appSettings.value.applicationLockWebAuthn = value;
    }

    // Navigation Bar
    function setShowAddTransactionButtonInDesktopNavbar(value: boolean): void {
        updateApplicationSettingsValue('showAddTransactionButtonInDesktopNavbar', value);
        appSettings.value.showAddTransactionButtonInDesktopNavbar = value;
        updateUserApplicationCloudSettingValue('showAddTransactionButtonInDesktopNavbar', value);
    }

    // Overview Page
    function setDesktopOverviewPageLayout(value: string): void {
        updateApplicationSettingsValue('desktopOverviewPageLayout', value);
        appSettings.value.desktopOverviewPageLayout = value;
        updateUserApplicationCloudSettingValue('desktopOverviewPageLayout', value);
    }

    function setMobileOverviewPageLayout(value: string): void {
        updateApplicationSettingsValue('mobileOverviewPageLayout', value);
        appSettings.value.mobileOverviewPageLayout = value;
        updateUserApplicationCloudSettingValue('mobileOverviewPageLayout', value);
    }

    function setShowAmountInHomePage(value: boolean): void {
        updateApplicationSettingsValue('showAmountInHomePage', value);
        appSettings.value.showAmountInHomePage = value;
        updateUserApplicationCloudSettingValue('showAmountInHomePage', value);
    }

    function setTimezoneUsedForStatisticsInHomePage(value: number): void {
        updateApplicationSettingsValue('timezoneUsedForStatisticsInHomePage', value);
        appSettings.value.timezoneUsedForStatisticsInHomePage = value;
        updateUserApplicationCloudSettingValue('timezoneUsedForStatisticsInHomePage', value);
    }

    function setOverviewAccountFilterInHomePage(value: Record<string, boolean>): void {
        updateApplicationSettingsValue('overviewAccountFilterInHomePage', value);
        appSettings.value.overviewAccountFilterInHomePage = value;
        updateUserApplicationCloudSettingValue('overviewAccountFilterInHomePage', value);
    }

    function setOverviewTransactionCategoryFilterInHomePage(value: Record<string, boolean>): void {
        updateApplicationSettingsValue('overviewTransactionCategoryFilterInHomePage', value);
        appSettings.value.overviewTransactionCategoryFilterInHomePage = value;
        updateUserApplicationCloudSettingValue('overviewTransactionCategoryFilterInHomePage', value);
    }

    function setOverviewTransactionTagFilterInHomePage(value: string): void {
        updateApplicationSettingsValue('overviewTransactionTagFilterInHomePage', value);
        appSettings.value.overviewTransactionTagFilterInHomePage = value;
        updateUserApplicationCloudSettingValue('overviewTransactionTagFilterInHomePage', value);
    }

    return {
        setTheme, setFontSize, setTimeZone, setAutoUpdateExchangeRatesData, setShowAccountBalance, setAccountCategoryOrders,
        setChartColors, setEnableSwipeBack, setEnableAnimate, setEnableApplicationLock, setEnableApplicationLockWebAuthn,
        setShowAddTransactionButtonInDesktopNavbar, setDesktopOverviewPageLayout, setMobileOverviewPageLayout,
        setShowAmountInHomePage, setTimezoneUsedForStatisticsInHomePage, setOverviewAccountFilterInHomePage,
        setOverviewTransactionCategoryFilterInHomePage, setOverviewTransactionTagFilterInHomePage,
    };
}
