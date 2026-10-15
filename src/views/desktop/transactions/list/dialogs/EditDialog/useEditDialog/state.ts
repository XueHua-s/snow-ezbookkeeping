import { TransactionType } from '@/core/transaction.ts';
import { isEquals } from '@/lib/common.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { useI18n } from '@/locales/helpers.ts';
import { Transaction } from '@/models/transaction.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
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
import { computed, ref, shallowRef, useTemplateRef } from 'vue';
import { type ConfirmDialogType, type MapViewType, type SnackBarType, type TransactionEditOptions, type TransactionEditResponse } from './types.ts';


interface Context {
    props: SetupProps<{
    type: TransactionEditPageType;
    persistent?: boolean;
    show?: boolean;
}>;
}

export function useEditDialogState(context: Context) {
    const { props } = context;

    const { tt, formatCoordinate } = useI18n();

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
        defaultCurrency,
        imageUploadQualityType,
        allTimezones,
        allVisibleAccounts,
        allVisibleCategorizedAccounts,
        allCategories,
        firstVisibleAccountId,
        hasVisibleExpenseCategories,
        hasVisibleIncomeCategories,
        hasVisibleTransferCategories,
        canAddTransactionPicture,
        title,
        saveButtonTitle,
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
        createNewTransactionModel,
        setTransactionModel,
        updateTransactionModelFromRecognizedResponse,
        updateTransactionModelByAfterSaveAction,
        updateTransactionTime,
        updateTransactionTimezone,
        swapTransactionData,
        getTransactionPictureUrl
    } = useTransactionEditPageBase(props.type);

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const transactionsStore = useTransactionsStore();

    const transactionTemplatesStore = useTransactionTemplatesStore();

    const map = useTemplateRef<MapViewType>('map');

    const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const pictureInput = useTemplateRef<HTMLInputElement>('pictureInput');

    const resolveFunc = shallowRef<((response?: TransactionEditResponse) => void) | null>(null);

    const rejectFunc = shallowRef<((reason?: unknown) => void) | null>(null);

    const showState = ref<boolean>(false);

    const showPasteTextDialog = ref<boolean>(false);

    const activeTab = ref<string>('basicInfo');

    const initTransaction = ref<Transaction | null>(null);

    const initTemplate = ref<TransactionTemplate | null>(null);

    const originalTransactionEditable = ref<boolean>(false);

    const noTransactionDraft = ref<boolean>(false);

    const geoMenuState = ref<boolean>(false);

    const removingPictureId = ref<string>('');

    const pastedText = ref<string>('');

    const initOptions = ref<TransactionEditOptions | undefined>(undefined);

    const sourceAmountColor = computed<string | undefined>(() => {
        if (transaction.value.type === TransactionType.Expense) {
            return 'expense';
        } else if (transaction.value.type === TransactionType.Income) {
            return 'income';
        } else if (transaction.value.type === TransactionType.Transfer) {
            return 'primary';
        }

        return undefined;
    });

    const isTransactionModified = computed<boolean>(() => {
        if (mode.value === TransactionEditPageMode.Add) {
            if (props.type === TransactionEditPageType.Transaction) {
                return transactionsStore.isTransactionDraftModified(transaction.value, initOptions.value?.amount, initOptions.value?.categoryId, initOptions.value?.accountId, initOptions.value?.tagIds, firstVisibleAccountId.value);
            } else if (props.type === TransactionEditPageType.Template && transaction.value instanceof TransactionTemplate) {
                const template = transaction.value as TransactionTemplate;
                return !!initTemplate.value && !isEquals(template.toTemplateCreateRequest(clientSessionId.value), initTemplate.value.toTemplateCreateRequest(clientSessionId.value));
            } else {
                return true;
            }
        } else if (mode.value === TransactionEditPageMode.Edit) {
            if (props.type === TransactionEditPageType.Transaction) {
                return !!initTransaction.value && !isEquals(transaction.value.toModifyRequest(), initTransaction.value.toModifyRequest());
            } else if (props.type === TransactionEditPageType.Template && transaction.value instanceof TransactionTemplate) {
                const template = transaction.value as TransactionTemplate;
                return !!initTemplate.value && !isEquals(template.toTemplateModifyRequest(), initTemplate.value.toTemplateModifyRequest());
            } else {
                return true;
            }
        } else {
            return false;
        }
    });

    return {
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
    };
}
