import type { SetupProps } from '@/lib/ui/props.ts';
import {
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import { useEditDialogActions } from './useEditDialog/actions.ts';
import { useEditDialogLocation } from './useEditDialog/location.ts';
import { useEditDialogOpening } from './useEditDialog/opening.ts';
import { useEditDialogPictures } from './useEditDialog/pictures.ts';
import { useEditDialogRecognition } from './useEditDialog/recognition.ts';
import { useEditDialogSaving } from './useEditDialog/saving.ts';
import { useEditDialogState } from './useEditDialog/state.ts';
export * from './useEditDialog/types.ts';

// Coordinates EditDialog state and effects; called once from the component setup.
export function useEditDialog(props: SetupProps<{
    type: TransactionEditPageType;
    persistent?: boolean;
    show?: boolean;
}>) {
    const {
        tt, formatCoordinate, mode, isSupportGeoLocation, editId, addByTemplateId, duplicateFromId, clientSessionId, loading,
        recognizing, submitting, submitted, uploadingPicture, geoLocationStatus, setGeoLocationByClickMap, transaction,
        defaultCurrency, imageUploadQualityType, allTimezones, allVisibleAccounts, allVisibleCategorizedAccounts, allCategories,
        firstVisibleAccountId, hasVisibleExpenseCategories, hasVisibleIncomeCategories, hasVisibleTransferCategories,
        canAddTransactionPicture, title, saveButtonTitle, cancelButtonTitle, sourceAmountName, sourceAmountTitle,
        sourceAccountTitle, transferInAmountTitle, sourceAccountName, destinationAccountName, sourceAccountCurrency,
        destinationAccountCurrency, transactionDisplayTimezone, transactionTimezoneTimeDifference, geoLocationStatusInfo,
        transactionDescriptionTitle, inputEmptyProblemMessage, inputIsEmpty, createNewTransactionModel, setTransactionModel,
        updateTransactionModelFromRecognizedResponse, updateTransactionModelByAfterSaveAction, updateTransactionTime,
        updateTransactionTimezone, swapTransactionData, getTransactionPictureUrl, settingsStore, userStore, accountsStore,
        transactionCategoriesStore, transactionTagsStore, transactionsStore, transactionTemplatesStore, map, confirmDialog,
        snackbar, pictureInput, resolveFunc, rejectFunc, showState, showPasteTextDialog, activeTab, initTransaction, initTemplate,
        originalTransactionEditable, noTransactionDraft, geoMenuState, removingPictureId, pastedText, initOptions,
        sourceAmountColor, isTransactionModified,
    } = useEditDialogState({ props });
    const { updateGeoLocation, updateSpecifiedGeoLocation, clearGeoLocation } = useEditDialogLocation({ geoMenuState, isSupportGeoLocation, snackbar, geoLocationStatus, transaction, setGeoLocationByClickMap, map });
    const { showOpenPictureDialog, uploadPicture, viewOrRemovePicture, onSavingTag, onUploadPicture, onShowDateTimeError } = useEditDialogPictures({
        canAddTransactionPicture, submitting, pictureInput, uploadingPicture, imageUploadQualityType, transactionsStore,
        transaction, snackbar, mode, getTransactionPictureUrl, confirmDialog, removingPictureId, activeTab, map,
    });
    const { recognizeText, recognizeFromClipboard } = useEditDialogRecognition({ recognizing, loading, submitting, transactionsStore, updateTransactionModelFromRecognizedResponse, snackbar, pastedText, settingsStore, showPasteTextDialog });
    const { open } = useEditDialogOpening({
        addByTemplateId, duplicateFromId, showState, activeTab, loading, submitting, submitted, geoLocationStatus,
        setGeoLocationByClickMap, originalTransactionEditable, noTransactionDraft, initOptions, createNewTransactionModel,
        setTransactionModel, initTransaction, transaction, accountsStore, transactionCategoriesStore, transactionTagsStore, props,
        mode, editId, transactionsStore, settingsStore, updateGeoLocation, initTemplate, transactionTemplatesStore,
        clientSessionId, rejectFunc, uploadPicture, pastedText, recognizeText, showPasteTextDialog, resolveFunc,
    });
    const { save } = useEditDialogSaving({
        inputEmptyProblemMessage, snackbar, props, mode, submitting, transactionsStore, transaction, defaultCurrency,
        clientSessionId, submitted, noTransactionDraft, addByTemplateId, duplicateFromId, updateTransactionModelByAfterSaveAction,
        initOptions, resolveFunc, showState, confirmDialog, userStore, transactionTemplatesStore,
    });
    const { duplicate, edit, remove, cancel } = useEditDialogActions({
        props, mode, editId, duplicateFromId, transaction, clientSessionId, submitted, activeTab, settingsStore, confirmDialog,
        submitting, transactionsStore, defaultCurrency, resolveFunc, showState, snackbar, rejectFunc, noTransactionDraft,
        addByTemplateId, initOptions, firstVisibleAccountId,
    });

    return {
        tt, formatCoordinate, mode, loading, recognizing, submitting, uploadingPicture, setGeoLocationByClickMap, transaction,
        allTimezones, allVisibleAccounts, allVisibleCategorizedAccounts, allCategories, hasVisibleExpenseCategories,
        hasVisibleIncomeCategories, hasVisibleTransferCategories, canAddTransactionPicture, title, saveButtonTitle,
        cancelButtonTitle, sourceAmountName, sourceAmountTitle, sourceAccountTitle, transferInAmountTitle, sourceAccountName,
        destinationAccountName, sourceAccountCurrency, destinationAccountCurrency, transactionDisplayTimezone,
        transactionTimezoneTimeDifference, geoLocationStatusInfo, transactionDescriptionTitle, inputEmptyProblemMessage,
        inputIsEmpty, updateTransactionTime, updateTransactionTimezone, swapTransactionData, getTransactionPictureUrl, map,
        confirmDialog, snackbar, pictureInput, showState, showPasteTextDialog, activeTab, originalTransactionEditable,
        geoMenuState, removingPictureId, pastedText, sourceAmountColor, isTransactionModified, open, save, recognizeText,
        recognizeFromClipboard, duplicate, edit, remove, cancel, updateGeoLocation, updateSpecifiedGeoLocation, clearGeoLocation,
        showOpenPictureDialog, viewOrRemovePicture, onSavingTag, onUploadPicture, onShowDateTimeError,
    };
}
