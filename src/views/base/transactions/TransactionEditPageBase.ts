import { useTransactionEditPageBaseEditing } from './TransactionEditPageBase/editing.ts';
import { useTransactionEditPageBasePresentation } from './TransactionEditPageBase/presentation.ts';
import { useTransactionEditPageBaseState } from './TransactionEditPageBase/state.ts';
import { TransactionEditPageMode, TransactionEditPageType } from './TransactionEditPageBase/types.ts';
export * from './TransactionEditPageBase/types.ts';









export function useTransactionEditPageBase(type: TransactionEditPageType, initMode?: TransactionEditPageMode, transactionDefaultType?: number) {
    const {
        tt, getTimezoneDifferenceDisplayText, formatNumberToLocalizedNumerals, getAdaptiveAmountRate, settingsStore,
        transactionsStore, exchangeRatesStore, isSupportGeoLocation, mode, editId, addByTemplateId, duplicateFromId,
        clientSessionId, loading, recognizing, submitting, submitted, uploadingPicture, geoLocationStatus,
        setGeoLocationByClickMap, transaction, numeralSystem, currentTimezoneOffsetMinutes, showAccountBalance, defaultCurrency,
        defaultAccountId, firstDayOfWeek, imageUploadQualityType, allTimezones, allAccounts, allVisibleAccounts, allAccountsMap,
        allVisibleCategorizedAccounts, allCategories, allCategoriesMap, allTagsMap, firstVisibleAccountId,
        hasVisibleExpenseCategories, hasVisibleIncomeCategories, hasVisibleTransferCategories, canAddTransactionPicture,
        isNormalTransactionTemplate, isScheduledTransactionTemplate, createNewTransactionModel, getDisplayAmount,
    } = useTransactionEditPageBaseState({ initMode, transactionDefaultType, type });
    const {
        title, saveButtonTitle, quickSaveButtonTitle, cancelButtonTitle, sourceAmountName, sourceAmountTitle, sourceAccountTitle,
        transferInAmountTitle, sourceAccountName, destinationAccountName, sourceAccountCurrency, destinationAccountCurrency,
        transactionDisplayTimezone, transactionTimezoneTimeDifference, geoLocationStatusInfo, transactionDescriptionTitle,
        inputEmptyProblemMessage, inputIsEmpty,
    } = useTransactionEditPageBasePresentation({
        type, mode, isNormalTransactionTemplate, transaction, isScheduledTransactionTemplate, settingsStore, allAccountsMap, tt,
        defaultCurrency, exchangeRatesStore, getDisplayAmount, getAdaptiveAmountRate, allAccounts, numeralSystem,
        getTimezoneDifferenceDisplayText, geoLocationStatus, formatNumberToLocalizedNumerals,
    });
    const {
        setTransactionModel, updateTransactionModelFromRecognizedResponse, updateTransactionModelByAfterSaveAction,
        updateTransactionTime, updateTransactionTimezone, swapTransactionData, getTransactionPictureUrl,
    } = useTransactionEditPageBaseEditing({
        transaction, allCategories, allCategoriesMap, allVisibleAccounts, allAccountsMap, allTagsMap, defaultAccountId,
        createNewTransactionModel, transactionDefaultType, geoLocationStatus, allTimezones, isScheduledTransactionTemplate,
        transactionsStore, mode, loading,
    });

    return {
        // constants
        isSupportGeoLocation,
        // states
        mode,
        editId,
        addByTemplateId,
        duplicateFromId,
        clientSessionId,
        loading,
        recognizing,
        submitting,
        submitted,
        uploadingPicture,
        geoLocationStatus,
        setGeoLocationByClickMap,
        transaction,
        // computed states
        numeralSystem,
        currentTimezoneOffsetMinutes,
        showAccountBalance,
        defaultCurrency,
        defaultAccountId,
        firstDayOfWeek,
        imageUploadQualityType,
        allTimezones,
        allAccounts,
        allVisibleAccounts,
        allAccountsMap,
        allVisibleCategorizedAccounts,
        allCategories,
        allCategoriesMap,
        allTagsMap,
        firstVisibleAccountId,
        hasVisibleExpenseCategories,
        hasVisibleIncomeCategories,
        hasVisibleTransferCategories,
        canAddTransactionPicture,
        title,
        saveButtonTitle,
        quickSaveButtonTitle,
        cancelButtonTitle,
        sourceAmountName,
        sourceAmountTitle,
        sourceAccountTitle,
        transferInAmountTitle,
        sourceAccountName,
        destinationAccountName,
        sourceAccountCurrency,
        destinationAccountCurrency,
        transactionDisplayTimezone,
        transactionTimezoneTimeDifference,
        geoLocationStatusInfo,
        transactionDescriptionTitle,
        inputEmptyProblemMessage,
        inputIsEmpty,
        // functions
        createNewTransactionModel,
        setTransactionModel,
        updateTransactionModelFromRecognizedResponse,
        updateTransactionModelByAfterSaveAction,
        updateTransactionTime,
        updateTransactionTimezone,
        swapTransactionData,
        getDisplayAmount,
        getTransactionPictureUrl
    }
}
