import SnackBar from '@/components/desktop/SnackBar.vue';
import { DEFAULT_PAGE_COUNTS } from '@/consts/page.ts';
import type { NameNumeralValue } from '@/core/base.ts';
import { type LocalizedDateRange, DateRangeScene } from '@/core/datetime.ts';
import { DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_DESKTOP } from '@/core/statistics.ts';
import { ThemeType } from '@/core/theme.ts';
import { isDefaultDesktopOverviewLayout, parseDesktopOverviewLayout } from '@/lib/overview_layout.ts';
import { getSystemTheme } from '@/lib/ui/common.ts';
import { useI18n } from '@/locales/helpers.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useAppSettingPageBase } from '@/views/base/settings/AppSettingsPageBase.ts';
import AccountCategoryDisplayOrderDialog from '@/views/desktop/settings/dialogs/AccountCategoryDisplayOrderDialog.vue';
import ChartColorSchemeDialog from '@/views/desktop/settings/dialogs/ChartColorSchemeDialog.vue';
import { computed, ref, useTemplateRef } from 'vue';
import { useRouter } from 'vue-router';
import { useTheme } from 'vuetify';

export type SnackBarType = InstanceType<typeof SnackBar>;

export type ChartColorSchemeDialogType = InstanceType<typeof ChartColorSchemeDialog>;

export type AccountCategoryDisplayOrderDialogType = InstanceType<typeof AccountCategoryDisplayOrderDialog>;

// Coordinates PreferencesSettingsPage state and effects; called once from the component setup.
export function usePreferencesSettingsPage() {
    const theme = useTheme();

    const router = useRouter();

    const { tt, getAllDateRanges, getTablePageOptions } = useI18n();

    const {
        loadingAccounts,
        loadingTransactionCategories,
        loadingTags,
        allThemes,
        allTimezones,
        allTimezoneTypesUsedForStatistics,
        allCurrencySortingTypes,
        allKeywordMatchModes,
        allAutoSaveTransactionDraftTypes,
        allImageUploadQualityTypes,
        allReconciliationStatementDateRanges,
        hasAnyAccount,
        hasAnyVisibleAccount,
        hasAnyTransactionCategory,
        hasAnyTransactionTag,
        timeZone,
        isAutoUpdateExchangeRatesData,
        showAccountBalance,
        showAmountInHomePage,
        itemsCountInTransactionListPage,
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

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const chartColorSchemeDialog = useTemplateRef<ChartColorSchemeDialogType>('chartColorSchemeDialog');

    const accountCategorysDisplayOrderDialog = useTemplateRef<AccountCategoryDisplayOrderDialogType>('accountCategorysDisplayOrderDialog');

    const showAccountsIncludedInHomePageOverviewDialog = ref<boolean>(false);

    const showTransactionCategoriesIncludedInHomePageOverviewDialog = ref<boolean>(false);

    const showTransactionTagsIncludedInHomePageOverviewDialog = ref<boolean>(false);

    const showAccountsIncludedInTotalDialog = ref<boolean>(false);

    const allPageCounts = computed<NameNumeralValue[]>(() => getTablePageOptions(DEFAULT_PAGE_COUNTS, undefined, false, true));

    const allInsightsExplorerDefaultDateRanges = computed<LocalizedDateRange[]>(() => getAllDateRanges(DateRangeScene.InsightsExplorer, {}));

    const currentTheme = computed<string>({
        get: () => settingsStore.appSettings.theme,
        set: (value: string) => {
            if (value !== settingsStore.appSettings.theme) {
                settingsStore.setTheme(value);

                if (value === ThemeType.Light || value === ThemeType.Dark) {
                    theme.change(value);
                } else {
                    theme.change(getSystemTheme());
                }
            }
        }
    });

    const desktopOverviewPageLayoutDisplayContent = computed(() => {
        try {
            return tt(isDefaultDesktopOverviewLayout(parseDesktopOverviewLayout(settingsStore.appSettings.desktopOverviewPageLayout)) ? 'Default' : 'Custom');
        } catch {
            return tt('Custom');
        }
    });

    const showAddTransactionButtonInDesktopNavbar = computed<boolean>({
        get: () => settingsStore.appSettings.showAddTransactionButtonInDesktopNavbar,
        set: (value) => settingsStore.setShowAddTransactionButtonInDesktopNavbar(value)
    });

    const rememberLastSelectedFileTypeInImportTransactionDialog = computed<boolean>({
        get: () => settingsStore.appSettings.rememberLastSelectedFileTypeInImportTransactionDialog,
        set: (value) => settingsStore.setRememberLastSelectedFileTypeInImportTransactionDialog(value)
    });

    const insightsExplorerDefaultDateRangeType = computed<number>({
        get: () => settingsStore.appSettings.insightsExplorerDefaultDateRangeType,
        set: (value) => settingsStore.setInsightsExplorerDefaultDateRangeType(value)
    });

    const showTagInInsightsExplorerPage = computed<boolean>({
        get: () => settingsStore.appSettings.showTagInInsightsExplorerPage,
        set: (value) => settingsStore.setShowTagInInsightsExplorerPage(value)
    });

    const hideCategoriesWithoutAccounts = computed<boolean>({
        get: () => settingsStore.appSettings.hideCategoriesWithoutAccounts,
        set: (value) => settingsStore.setHideCategoriesWithoutAccounts(value)
    });

    const reconciliationStatementButtonDefaultDateRangeTypeInDesktop = computed<number>({
        get: () => getValidReconciliationStatementPageDefaultDateRangeType(settingsStore.appSettings.reconciliationStatementButtonDefaultDateRangeTypeInDesktop, DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_DESKTOP.type),
        set: (value: number) => settingsStore.setReconciliationStatementButtonDefaultDateRangeTypeInDesktop(value)
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
                snackbar.value?.showError(error);
            }
        });

        transactionCategoriesStore.loadAllCategories({
            force: false
        }).then(() => {
            loadingTransactionCategories.value = false;
        }).catch(error => {
            loadingTransactionCategories.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });

        transactionTagsStore.loadAllTags({
            force: false
        }).then(() => {
            loadingTags.value = false;
        }).catch(error => {
            loadingTags.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    init();

    return {
        router, tt, loadingAccounts, loadingTransactionCategories, loadingTags, allThemes, allTimezones,
        allTimezoneTypesUsedForStatistics, allCurrencySortingTypes, allKeywordMatchModes, allAutoSaveTransactionDraftTypes,
        allImageUploadQualityTypes, allReconciliationStatementDateRanges, hasAnyAccount, hasAnyVisibleAccount,
        hasAnyTransactionCategory, hasAnyTransactionTag, timeZone, isAutoUpdateExchangeRatesData, showAccountBalance,
        showAmountInHomePage, itemsCountInTransactionListPage, timezoneUsedForStatisticsInHomePage,
        showTotalAmountInTransactionListPage, totalAmountTypeInTransactionListPage, showTagInTransactionListPage,
        defaultKeywordMatchModeInTransactionListPage, autoSaveTransactionDraft, isAutoGetCurrentGeoLocation,
        transactionPictureQuality, isAlwaysRequireConfirmationOfClipboardContentBeforeSubmission,
        isAutoUploadTransactionPictureForAIRecognition, currencySortByInExchangeRatesPage, chartColorSchemeContent,
        accountsIncludedInHomePageOverviewDisplayContent, accountsIncludedInTotalDisplayContent,
        accountCategorysDisplayOrderContent, transactionCategoriesIncludedInHomePageOverviewDisplayContent, settingsStore,
        snackbar, chartColorSchemeDialog, accountCategorysDisplayOrderDialog, showAccountsIncludedInHomePageOverviewDialog,
        showTransactionCategoriesIncludedInHomePageOverviewDialog, showTransactionTagsIncludedInHomePageOverviewDialog,
        showAccountsIncludedInTotalDialog, allPageCounts, allInsightsExplorerDefaultDateRanges, currentTheme,
        desktopOverviewPageLayoutDisplayContent, showAddTransactionButtonInDesktopNavbar,
        rememberLastSelectedFileTypeInImportTransactionDialog, insightsExplorerDefaultDateRangeType,
        showTagInInsightsExplorerPage, hideCategoriesWithoutAccounts, reconciliationStatementButtonDefaultDateRangeTypeInDesktop,
    };
}
