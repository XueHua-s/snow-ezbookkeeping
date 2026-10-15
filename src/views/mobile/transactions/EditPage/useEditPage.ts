import type { SetupProps } from '@/lib/ui/props.ts';
import type { Router } from 'framework7/types';
import { useEditPageInitialization } from './useEditPage/initialization.ts';
import { useEditPageLifecycle } from './useEditPage/lifecycle.ts';
import { useEditPageLocation } from './useEditPage/location.ts';
import { useEditPagePictures } from './useEditPage/pictures.ts';
import { useEditPagePresentation } from './useEditPage/presentation.ts';
import { useEditPageRecognition } from './useEditPage/recognition.ts';
import { useEditPageSaving } from './useEditPage/saving.ts';
import { useEditPageState } from './useEditPage/state.ts';



// Coordinates EditPage state and effects; called once from the component setup.
export function useEditPage(props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
    autoUploadPicture?: File;
    autoRecognizeClipboardText?: string;
}>) {
    const {
        query, pageTypeAndMode, tt, getMultiMonthAndDayLongNames, getMultiMonthdayShortNames, getMultiWeekdayLongNames,
        formatDateTimeToLongDate, formatDateTimeToLongTime, formatGregorianTextualYearMonthDayToLongDate,
        parseAmountFromLocalizedNumerals, formatCoordinate, showAlert, showConfirm, showCancelableLoading, showToast,
        routeBackOnError, mode, isSupportGeoLocation, editId, addByTemplateId, duplicateFromId, clientSessionId, loading,
        recognizing, submitting, submitted, uploadingPicture, geoLocationStatus, setGeoLocationByClickMap, transaction,
        numeralSystem, currentTimezoneOffsetMinutes, defaultCurrency, firstDayOfWeek, imageUploadQualityType, allTimezones,
        allVisibleAccounts, allVisibleCategorizedAccounts, allCategories, allTagsMap, firstVisibleAccountId,
        hasVisibleExpenseCategories, hasVisibleIncomeCategories, hasVisibleTransferCategories, canAddTransactionPicture, title,
        quickSaveButtonTitle, sourceAmountTitle, sourceAccountTitle, transferInAmountTitle, sourceAccountName,
        destinationAccountName, sourceAccountCurrency, destinationAccountCurrency, transactionDisplayTimezone,
        transactionTimezoneTimeDifference, geoLocationStatusInfo, transactionDescriptionTitle, inputEmptyProblemMessage,
        inputIsEmpty, setTransactionModel, updateTransactionModelFromRecognizedResponse, updateTransactionModelByAfterSaveAction,
        updateTransactionTime, updateTransactionTimezone, swapTransactionData, getDisplayAmount, getTransactionPictureUrl,
        isSupportClipboard, settingsStore, userStore, accountsStore, transactionCategoriesStore, transactionTagsStore,
        transactionsStore, transactionTemplatesStore, pictureBrowser, pictureInput, loadingError, removingPictureId, pastedText,
        transactionDateTimeSheetMode, showTimeInDefaultTimezone, showQuickSavePopover, showTimezonePopup,
        showGeoLocationActionSheet, showMoreActionSheet, showSourceAmountSheet, showDestinationAmountSheet, showCategorySheet,
        showSourceAccountSheet, showDestinationAccountSheet, showTransactionDateTimeSheet, showTransactionScheduledFrequencySheet,
        showScheduledStartDateSheet, showScheduledEndDateSheet, showGeoLocationMapSheet, showTransactionTagSheet,
        showTransactionPictures, showAITextRecognitionSheet, getQueryTransactionOptions,
    } = useEditPageState({ props });
    const {
        quickSaveButtonStyleType, quickSaveButtonFloatingPosition, sourceAmountClass, destinationAmountClass,
        transactionDisplayDate, transactionDisplayTime, transactionDisplayTimezoneName, transactionPictures, transactionThumbs,
        transactionDisplayScheduledFrequency, transactionDisplayScheduledStartDate, transactionDisplayScheduledEndDate,
    } = useEditPagePresentation({
        settingsStore, mode, transaction, showTimeInDefaultTimezone, formatDateTimeToLongDate, formatDateTimeToLongTime,
        numeralSystem, allTimezones, getTransactionPictureUrl, pageTypeAndMode, tt, getMultiWeekdayLongNames, firstDayOfWeek,
        getMultiMonthdayShortNames, getMultiMonthAndDayLongNames, formatGregorianTextualYearMonthDayToLongDate,
    });
    const { showOpenPictureDialog, uploadPicture, viewOrRemovePicture, onUploadPicture } = useEditPagePictures({
        canAddTransactionPicture, submitting, pictureInput, uploadingPicture, imageUploadQualityType, transactionsStore,
        transaction, showToast, mode, pictureBrowser, showConfirm, removingPictureId,
    });
    const { recognizeText, recognizeFromClipboard, pasteAmount } = useEditPageRecognition({
        recognizing, loading, submitting, showCancelableLoading, transactionsStore, updateTransactionModelFromRecognizedResponse,
        showToast, pastedText, isSupportClipboard, settingsStore, showAITextRecognitionSheet, mode,
        parseAmountFromLocalizedNumerals, transaction,
    });
    const { init } = useEditPageInitialization({
        pageTypeAndMode, showToast, loadingError, loading, accountsStore, transactionCategoriesStore, transactionTagsStore,
        transactionTemplatesStore, query, mode, editId, duplicateFromId, transactionsStore, transaction,
        getQueryTransactionOptions, clientSessionId, addByTemplateId, settingsStore, setTransactionModel, props,
        showTransactionPictures, uploadPicture, pastedText, recognizeText, showAITextRecognitionSheet,
    });
    const { save, quickSave, duplicate } = useEditPageSaving({
        props, mode, inputEmptyProblemMessage, showAlert, pageTypeAndMode, submitting, transactionsStore, transaction,
        defaultCurrency, clientSessionId, submitted, query, addByTemplateId, duplicateFromId, showToast,
        updateTransactionModelByAfterSaveAction, getQueryTransactionOptions, showConfirm, userStore, transactionTemplatesStore,
        settingsStore, showQuickSavePopover,
    });
    const { updateGeoLocation, clearGeoLocation, showDateTimeDialog } = useEditPageLocation({ isSupportGeoLocation, showToast, geoLocationStatus, transaction, mode, showTimeInDefaultTimezone, transactionDateTimeSheetMode, showTransactionDateTimeSheet });
    const { onPageAfterIn, onPageBeforeOut } = useEditPageLifecycle({
        routeBackOnError, props, loadingError, settingsStore, mode, geoLocationStatus, transaction, updateGeoLocation, submitted,
        pageTypeAndMode, query, addByTemplateId, duplicateFromId, transactionsStore, firstVisibleAccountId, showConfirm, init,
    });

    return {
        pageTypeAndMode, tt, formatCoordinate, mode, loading, recognizing, submitting, uploadingPicture, setGeoLocationByClickMap,
        transaction, currentTimezoneOffsetMinutes, allTimezones, allVisibleAccounts, allVisibleCategorizedAccounts, allCategories,
        allTagsMap, hasVisibleExpenseCategories, hasVisibleIncomeCategories, hasVisibleTransferCategories,
        canAddTransactionPicture, title, quickSaveButtonTitle, sourceAmountTitle, sourceAccountTitle, transferInAmountTitle,
        sourceAccountName, destinationAccountName, sourceAccountCurrency, destinationAccountCurrency, transactionDisplayTimezone,
        transactionTimezoneTimeDifference, geoLocationStatusInfo, transactionDescriptionTitle, inputIsEmpty,
        updateTransactionTime, updateTransactionTimezone, swapTransactionData, getDisplayAmount, getTransactionPictureUrl,
        isSupportClipboard, pictureBrowser, pictureInput, removingPictureId, pastedText, transactionDateTimeSheetMode,
        showQuickSavePopover, showTimezonePopup, showGeoLocationActionSheet, showMoreActionSheet, showSourceAmountSheet,
        showDestinationAmountSheet, showCategorySheet, showSourceAccountSheet, showDestinationAccountSheet,
        showTransactionDateTimeSheet, showTransactionScheduledFrequencySheet, showScheduledStartDateSheet,
        showScheduledEndDateSheet, showGeoLocationMapSheet, showTransactionTagSheet, showTransactionPictures,
        showAITextRecognitionSheet, quickSaveButtonStyleType, quickSaveButtonFloatingPosition, sourceAmountClass,
        destinationAmountClass, transactionDisplayDate, transactionDisplayTime, transactionDisplayTimezoneName,
        transactionPictures, transactionThumbs, transactionDisplayScheduledFrequency, transactionDisplayScheduledStartDate,
        transactionDisplayScheduledEndDate, init, save, quickSave, recognizeText, recognizeFromClipboard, pasteAmount,
        updateGeoLocation, clearGeoLocation, showDateTimeDialog, showOpenPictureDialog, viewOrRemovePicture, duplicate,
        onUploadPicture, onPageAfterIn, onPageBeforeOut,
    };
}
