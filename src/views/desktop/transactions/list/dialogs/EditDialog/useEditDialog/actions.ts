import {
    getCurrentUnixTime,
    getTimezoneOffsetMinutes
} from '@/lib/datetime.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { Transaction } from '@/models/transaction.ts';
import {
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { useEditDialogState } from './state.ts';

interface Context {
    props: SetupProps<{
    type: TransactionEditPageType;
    persistent?: boolean;
    show?: boolean;
}>;
    mode: ReturnType<typeof useEditDialogState>['mode'];
    editId: ReturnType<typeof useEditDialogState>['editId'];
    duplicateFromId: ReturnType<typeof useEditDialogState>['duplicateFromId'];
    transaction: ReturnType<typeof useEditDialogState>['transaction'];
    clientSessionId: ReturnType<typeof useEditDialogState>['clientSessionId'];
    submitted: ReturnType<typeof useEditDialogState>['submitted'];
    activeTab: ReturnType<typeof useEditDialogState>['activeTab'];
    settingsStore: ReturnType<typeof useEditDialogState>['settingsStore'];
    confirmDialog: ReturnType<typeof useEditDialogState>['confirmDialog'];
    submitting: ReturnType<typeof useEditDialogState>['submitting'];
    transactionsStore: ReturnType<typeof useEditDialogState>['transactionsStore'];
    defaultCurrency: ReturnType<typeof useEditDialogState>['defaultCurrency'];
    resolveFunc: ReturnType<typeof useEditDialogState>['resolveFunc'];
    showState: ReturnType<typeof useEditDialogState>['showState'];
    snackbar: ReturnType<typeof useEditDialogState>['snackbar'];
    rejectFunc: ReturnType<typeof useEditDialogState>['rejectFunc'];
    noTransactionDraft: ReturnType<typeof useEditDialogState>['noTransactionDraft'];
    addByTemplateId: ReturnType<typeof useEditDialogState>['addByTemplateId'];
    initOptions: ReturnType<typeof useEditDialogState>['initOptions'];
    firstVisibleAccountId: ReturnType<typeof useEditDialogState>['firstVisibleAccountId'];
}

export function useEditDialogActions(context: Context) {
    const {
        props, mode, editId, duplicateFromId, transaction, clientSessionId, submitted, activeTab, settingsStore, confirmDialog,
        submitting, transactionsStore, defaultCurrency, resolveFunc, showState, snackbar, rejectFunc, noTransactionDraft,
        addByTemplateId, initOptions, firstVisibleAccountId,
    } = context;

    function duplicate(withTime?: boolean, withGeoLocation?: boolean): void {
        if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.View) {
            return;
        }

        editId.value = null;
        duplicateFromId.value = transaction.value.id;
        clientSessionId.value = generateRandomUUID();
        submitted.value = false;
        activeTab.value = 'basicInfo';
        transaction.value.id = '';

        if (!withTime) {
            transaction.value.time = getCurrentUnixTime();
            transaction.value.timeZone = settingsStore.appSettings.timeZone;
            transaction.value.utcOffset = getTimezoneOffsetMinutes(transaction.value.time, transaction.value.timeZone);
        }

        if (!withGeoLocation) {
            transaction.value.removeGeoLocation();
        }

        transaction.value.clearPictures();
        mode.value = TransactionEditPageMode.Add;
    }

    function edit(): void {
        if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.View) {
            return;
        }

        mode.value = TransactionEditPageMode.Edit;
    }

    function remove(): void {
        if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.View) {
            return;
        }

        confirmDialog.value?.open('Are you sure you want to delete this transaction?').then(() => {
            submitting.value = true;

            transactionsStore.deleteTransaction({
                transaction: transaction.value as Transaction,
                defaultCurrency: defaultCurrency.value
            }).then(() => {
                if (resolveFunc.value) {
                    resolveFunc.value();
                }

                submitting.value = false;
                showState.value = false;
            }).catch(error => {
                submitting.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    function cancel(): void {
        const doClose = function () {
            if (props.type === TransactionEditPageType.Transaction && mode.value === TransactionEditPageMode.Add && submitted.value && resolveFunc.value) {
                resolveFunc.value({
                    message: 'You have added a new transaction'
                });
            } else if (rejectFunc.value) {
                rejectFunc.value();
            }

            showState.value = false;
        };

        if (props.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.Add || noTransactionDraft.value || addByTemplateId.value || duplicateFromId.value) {
            doClose();
            return;
        }

        if (settingsStore.appSettings.autoSaveTransactionDraft === 'confirmation') {
            if (transactionsStore.isTransactionDraftModified(transaction.value, initOptions.value?.amount, initOptions.value?.categoryId, initOptions.value?.accountId, initOptions.value?.tagIds, firstVisibleAccountId.value)) {
                confirmDialog.value?.open('Do you want to save this transaction draft?').then(() => {
                    transactionsStore.saveTransactionDraft(transaction.value, initOptions.value?.amount, initOptions.value?.categoryId, initOptions.value?.accountId, initOptions.value?.tagIds, firstVisibleAccountId.value);
                    doClose();
                }).catch(() => {
                    transactionsStore.clearTransactionDraft();
                    doClose();
                });
            } else {
                transactionsStore.clearTransactionDraft();
                doClose();
            }
        } else if (settingsStore.appSettings.autoSaveTransactionDraft === 'enabled') {
            transactionsStore.saveTransactionDraft(transaction.value, initOptions.value?.amount, initOptions.value?.categoryId, initOptions.value?.accountId, initOptions.value?.tagIds, firstVisibleAccountId.value);
            doClose();
        } else {
            doClose();
        }
    }

    return { duplicate, edit, remove, cancel };
}
