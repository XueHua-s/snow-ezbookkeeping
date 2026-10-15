import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import { useImportTransactionCheckDataTabBatchRepair } from './useImportTransactionCheckDataTab/batch-repair.ts';
import { useImportTransactionCheckDataTabBatchReplace } from './useImportTransactionCheckDataTab/batch-replace.ts';
import { useImportTransactionCheckDataTabEditing } from './useImportTransactionCheckDataTab/editing.ts';
import { useImportTransactionCheckDataTabExport } from './useImportTransactionCheckDataTab/export.ts';
import { useImportTransactionCheckDataTabFilterMenu } from './useImportTransactionCheckDataTab/filter-menu.ts';
import { useImportTransactionCheckDataTabFiltering } from './useImportTransactionCheckDataTab/filtering.ts';
import { useImportTransactionCheckDataTabFormatting } from './useImportTransactionCheckDataTab/formatting.ts';
import { useImportTransactionCheckDataTabInvalidItems } from './useImportTransactionCheckDataTab/invalid-items.ts';
import { useImportTransactionCheckDataTabLifecycle } from './useImportTransactionCheckDataTab/lifecycle.ts';
import { useImportTransactionCheckDataTabOriginalItems } from './useImportTransactionCheckDataTab/original-items.ts';
import { useImportTransactionCheckDataTabPagination } from './useImportTransactionCheckDataTab/pagination.ts';
import { useImportTransactionCheckDataTabSelectionActions } from './useImportTransactionCheckDataTab/selection-actions.ts';
import { useImportTransactionCheckDataTabSelection } from './useImportTransactionCheckDataTab/selection.ts';
import { useImportTransactionCheckDataTabState } from './useImportTransactionCheckDataTab/state.ts';
import { useImportTransactionCheckDataTabToolMenu } from './useImportTransactionCheckDataTab/tool-menu.ts';
export * from './useImportTransactionCheckDataTab/types.ts';

// Coordinates ImportTransactionCheckDataTab state and effects; called once from the component setup.
export function useImportTransactionCheckDataTab(props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>) {
    const {
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
    } = useImportTransactionCheckDataTabState();
    const {
        canImport, selectedImportTransactionCount, selectedExpenseTransactionCount, selectedIncomeTransactionCount,
        selectedTransferTransactionCount, selectedInvalidTransactionCount, anyButNotAllTransactionSelected,
        allTransactionSelected, allUsedCategoryNames, allUsedAccountNames, allUsedTagNames, displayFilterCustomDateRange,
    } = useImportTransactionCheckDataTabSelection({ props, allTagsMap, filters, formatDateTimeToLongDateTime, formatRange });
    const { filterMenus } = useImportTransactionCheckDataTabFilterMenu({
        tt, filters, displayFilterCustomDateRange, showCustomDateRangeDialog, allUsedCategoryNames, currentAmountFilterType,
        currentAmountFilterValue1, currentAmountFilterValue2, showCustomAmountFilterDialog, allUsedAccountNames, allUsedTagNames,
        currentDescriptionFilterValue, showCustomDescriptionDialog,
    });
    const { editTransaction, updateTransactionData, updateAllTransactionsIsValid } = useImportTransactionCheckDataTabEditing({ editingTransaction, editingTags, allCategoriesMap, allAccountsMap, allTagsMap, props });
    const { showBatchReplaceDialog, showBatchAddDialog } = useImportTransactionCheckDataTabBatchReplace({ isEditing, batchReplaceDialog, props, allTagsMap, updateTransactionData, snackbar, formatNumberToLocalizedNumerals });
    const {
        allOriginalExpenseCategoryNames, allOriginalIncomeCategoryNames, allOriginalTransferCategoryNames,
        allOriginalSourceAccountNames, allOriginalDestinationAccountNames, allOriginalTransactionTagNames,
    } = useImportTransactionCheckDataTabOriginalItems({ props, tt });
    const { allInvalidExpenseCategoryNames, allInvalidIncomeCategoryNames, allInvalidTransferCategoryNames, allInvalidAccountNames, allInvalidTransactionTagNames } = useImportTransactionCheckDataTabInvalidItems({ props, allCategoriesMap, tt, allAccountsMap, allTagsMap });
    const { showReplaceInvalidItemDialog, showBatchApplyRulesDialog, showBatchCreateInvalidItemDialog } = useImportTransactionCheckDataTabBatchRepair({
        isEditing, batchReplaceDialog, props, allCategoriesMap, allAccountsMap, allTagsMap, updateTransactionData, snackbar,
        formatNumberToLocalizedNumerals, batchApplyRulesDialog, allOriginalExpenseCategoryNames, allOriginalIncomeCategoryNames,
        allOriginalTransferCategoryNames, allOriginalSourceAccountNames, allOriginalDestinationAccountNames,
        allOriginalTransactionTagNames, selectedImportTransactionCount, allSecondaryCategoriesMapByName, allAccountsMapByName,
        batchCreateDialog,
    });
    const {
        isTagValid, getDisplayDateTime, isSameAsDefaultTimezoneOffsetMinutes, getDisplayTimezone, getDisplayTransactionType,
        getTransactionSourceAccountCurrency, getTransactionDestinationAccountCurrency, getTransactionDisplayAmount,
        getTransactionDisplayDestinationAmount, getTransactionDisplaySourceAmountInDefaultCurrency,
        getTransactionDisplayDestinationAmountInDefaultCurrency, getSourceAccountTitle, getSourceAccountDisplayName,
        getDestinationAccountDisplayName, getTransactionDescriptionTooltip,
    } = useImportTransactionCheckDataTabFormatting({
        allTagsMap, formatDateTimeToLongDateTime, tt, defaultCurrency, allAccountsMap,
        formatAmountToLocalizedNumeralsWithCurrency, exchangeRatesStore, allAccounts, formatNumberToLocalizedNumerals,
    });
    const { changeCustomDateFilter, exportData } = useImportTransactionCheckDataTabExport({
        filters, showCustomDateRangeDialog, props, selectedImportTransactionCount, tt, getDisplayTransactionType, allAccountsMap,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, formatNumberToWesternArabicNumeralsWithoutDigitGrouping,
        allCategoriesMap, allTagsMap, formatDateTimeToGregorianDefaultDateTime,
    });
    const { toolMenus } = useImportTransactionCheckDataTabToolMenu({
        tt, isEditing, selectedExpenseTransactionCount, showBatchReplaceDialog, selectedIncomeTransactionCount,
        selectedTransferTransactionCount, selectedImportTransactionCount, allOriginalTransactionTagNames, showBatchAddDialog,
        allInvalidExpenseCategoryNames, showReplaceInvalidItemDialog, allInvalidIncomeCategoryNames,
        allInvalidTransferCategoryNames, allInvalidAccountNames, allInvalidTransactionTagNames, showBatchCreateInvalidItemDialog,
        useCommaDecimalSeparator, exportData,
    });
    const { isTransactionDisplayed, importTransactionsFilter } = useImportTransactionCheckDataTabFiltering({ filters, allTagsMap });
    const { importTransactionHeaders, importTransactionsTablePageOptions, totalPageCount, currentPageTransactions } = useImportTransactionCheckDataTabPagination({ tt, getTablePageOptions, props, isTransactionDisplayed, countPerPage, currentPage });
    const {
        selectAllValid, selectAllInvalid, selectAll, selectNone, selectInvert, selectAllInThisPage, selectNoneInThisPage,
        clearSelectedTransactionsNotDisplayed, selectInvertInThisPage,
    } = useImportTransactionCheckDataTabSelectionActions({ props, isTransactionDisplayed, currentPageTransactions });
    const { focusTableScrollContainerWhenNonEditing, onShowDateRangeError, reset, setCountPerPage } = useImportTransactionCheckDataTabLifecycle({ isEditing, snackbar, editingTransaction, editingTags, filters, currentPage, countPerPage, clearSelectedTransactionsNotDisplayed });

    return {
        tt, formatNumberToLocalizedNumerals, formatCoordinate, allTagsWithGroupHeader, snackbar, batchReplaceDialog,
        batchApplyRulesDialog, batchCreateDialog, editingTransaction, editingTags, filters, currentPage, countPerPage,
        showCustomDateRangeDialog, showCustomAmountFilterDialog, showCustomDescriptionDialog, currentAmountFilterType,
        currentAmountFilterValue1, currentAmountFilterValue2, currentDescriptionFilterValue, defaultCurrency, allVisibleAccounts,
        allVisibleCategorizedAccounts, allAccountsMap, allCategories, allCategoriesMap, allTagsMap, hasVisibleExpenseCategories,
        hasVisibleIncomeCategories, hasVisibleTransferCategories, isEditing, canImport, filterMenus, toolMenus,
        importTransactionHeaders, importTransactionsTablePageOptions, totalPageCount, selectedImportTransactionCount,
        selectedInvalidTransactionCount, anyButNotAllTransactionSelected, allTransactionSelected, isTagValid, getDisplayDateTime,
        isSameAsDefaultTimezoneOffsetMinutes, getDisplayTimezone, getTransactionSourceAccountCurrency,
        getTransactionDestinationAccountCurrency, getTransactionDisplayAmount, getTransactionDisplayDestinationAmount,
        getTransactionDisplaySourceAmountInDefaultCurrency, getTransactionDisplayDestinationAmountInDefaultCurrency,
        getSourceAccountTitle, getSourceAccountDisplayName, getDestinationAccountDisplayName, getTransactionDescriptionTooltip,
        importTransactionsFilter, selectAllValid, selectAllInvalid, selectAll, selectNone, selectInvert, selectAllInThisPage,
        selectNoneInThisPage, selectInvertInThisPage, editTransaction, updateAllTransactionsIsValid, showBatchApplyRulesDialog,
        changeCustomDateFilter, focusTableScrollContainerWhenNonEditing, onShowDateRangeError, reset, setCountPerPage,
    };
}
