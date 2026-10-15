import { type SetTransactionOptions } from '@/lib/transaction.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { useI18n } from '@/locales/helpers.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useUserStore } from '@/stores/user.ts';
import {
    TransactionEditPageMode,
    TransactionEditPageType,
    useTransactionEditPageBase
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { PhotoBrowser, Router } from 'framework7/types';
import { ref, useTemplateRef } from 'vue';


interface Context {
    props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
    autoUploadPicture?: File;
    autoRecognizeClipboardText?: string;
}>;
}

export function useEditPageState(context: Context) {
    const { props } = context;

    const query = props.f7route.query;

    const pageTypeAndMode = getPageTypeNameMode();

    const {
        tt,
        getMultiMonthAndDayLongNames,
        getMultiMonthdayShortNames,
        getMultiWeekdayLongNames,
        formatDateTimeToLongDate,
        formatDateTimeToLongTime,
        formatGregorianTextualYearMonthDayToLongDate,
        parseAmountFromLocalizedNumerals,
        formatCoordinate
    } = useI18n();

    const { showAlert, showConfirm, showCancelableLoading, showToast, routeBackOnError } = useI18nUIComponents();

    const {
        mode,
        isSupportGeoLocation,
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
        numeralSystem,
        currentTimezoneOffsetMinutes,
        defaultCurrency,
        firstDayOfWeek,
        imageUploadQualityType,
        allTimezones,
        allVisibleAccounts,
        allVisibleCategorizedAccounts,
        allCategories,
        allTagsMap,
        firstVisibleAccountId,
        hasVisibleExpenseCategories,
        hasVisibleIncomeCategories,
        hasVisibleTransferCategories,
        canAddTransactionPicture,
        title,
        quickSaveButtonTitle,
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
        setTransactionModel,
        updateTransactionModelFromRecognizedResponse,
        updateTransactionModelByAfterSaveAction,
        updateTransactionTime,
        updateTransactionTimezone,
        swapTransactionData,
        getDisplayAmount,
        getTransactionPictureUrl
    } = useTransactionEditPageBase(pageTypeAndMode?.type || TransactionEditPageType.Transaction, pageTypeAndMode?.mode, query['type'] ? parseInt(query['type']) : undefined);

    const isSupportClipboard = !!navigator.clipboard;

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const transactionsStore = useTransactionsStore();

    const transactionTemplatesStore = useTransactionTemplatesStore();

    const pictureBrowser = useTemplateRef<PhotoBrowser.PhotoBrowser>('pictureBrowser');

    const pictureInput = useTemplateRef<HTMLInputElement>('pictureInput');

    const loadingError = ref<unknown | null>(null);

    const removingPictureId = ref<string | null>(null);

    const pastedText = ref<string>('');

    const transactionDateTimeSheetMode = ref<string>('time');

    const showTimeInDefaultTimezone = ref<boolean>(false);

    const showQuickSavePopover = ref<boolean>(false);

    const showTimezonePopup = ref<boolean>(false);

    const showGeoLocationActionSheet = ref<boolean>(false);

    const showMoreActionSheet = ref<boolean>(false);

    const showSourceAmountSheet = ref<boolean>(false);

    const showDestinationAmountSheet = ref<boolean>(false);

    const showCategorySheet = ref<boolean>(false);

    const showSourceAccountSheet = ref<boolean>(false);

    const showDestinationAccountSheet = ref<boolean>(false);

    const showTransactionDateTimeSheet = ref<boolean>(false);

    const showTransactionScheduledFrequencySheet = ref<boolean>(false);

    const showScheduledStartDateSheet = ref<boolean>(false);

    const showScheduledEndDateSheet = ref<boolean>(false);

    const showGeoLocationMapSheet = ref<boolean>(false);

    const showTransactionTagSheet = ref<boolean>(false);

    const showTransactionPictures = ref<boolean>(pageTypeAndMode?.type === TransactionEditPageType.Transaction
        && (pageTypeAndMode?.mode === TransactionEditPageMode.Add || pageTypeAndMode?.mode === TransactionEditPageMode.Edit)
        && settingsStore.appSettings.alwaysShowTransactionPicturesInMobileTransactionEditPage);

    const showAITextRecognitionSheet = ref<boolean>(false);

    function getPageTypeNameMode(): { type: TransactionEditPageType, mode: TransactionEditPageMode } | null {
        if (props.f7route.path === '/transaction/add') {
            return {
                type: TransactionEditPageType.Transaction,
                mode: TransactionEditPageMode.Add
            };
        } else if (props.f7route.path === '/transaction/edit') {
            return {
                type: TransactionEditPageType.Transaction,
                mode: TransactionEditPageMode.Edit
            };
        } else if (props.f7route.path === '/transaction/detail') {
            return {
                type: TransactionEditPageType.Transaction,
                mode: TransactionEditPageMode.View
            };
        } else if (props.f7route.path === '/template/add') {
            return {
                type: TransactionEditPageType.Template,
                mode: TransactionEditPageMode.Add
            };
        } else if (props.f7route.path === '/template/edit') {
            return {
                type: TransactionEditPageType.Template,
                mode: TransactionEditPageMode.Edit
            };
        } else {
            return null;
        }
    }

    function getQueryTransactionOptions(): SetTransactionOptions {
        return {
            time: query['time'] ? parseInt(query['time']) : undefined,
            type: query['type'] ? parseInt(query['type']) : 0,
            categoryId: query['categoryId'],
            accountId: query['accountId'],
            destinationAccountId: query['destinationAccountId'],
            amount: query['amount'] ? parseInt(query['amount']) : undefined,
            destinationAmount: query['destinationAmount'] ? parseInt(query['destinationAmount']) : undefined,
            tagIds: query['tagIds'],
            comment: query['comment']
        };
    }

    return {
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
    };
}
