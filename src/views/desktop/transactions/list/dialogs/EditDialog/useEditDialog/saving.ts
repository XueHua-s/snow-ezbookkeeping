import { KnownErrorCode } from '@/consts/api.ts';
import { TransactionEditScopeType } from '@/core/transaction.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { Transaction } from '@/models/transaction.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import {
    AfterSaveAction,
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { useEditDialogState } from './state.ts';

interface Context {
    inputEmptyProblemMessage: ReturnType<typeof useEditDialogState>['inputEmptyProblemMessage'];
    snackbar: ReturnType<typeof useEditDialogState>['snackbar'];
    props: SetupProps<{
    type: TransactionEditPageType;
    persistent?: boolean;
    show?: boolean;
}>;
    mode: ReturnType<typeof useEditDialogState>['mode'];
    submitting: ReturnType<typeof useEditDialogState>['submitting'];
    transactionsStore: ReturnType<typeof useEditDialogState>['transactionsStore'];
    transaction: ReturnType<typeof useEditDialogState>['transaction'];
    defaultCurrency: ReturnType<typeof useEditDialogState>['defaultCurrency'];
    clientSessionId: ReturnType<typeof useEditDialogState>['clientSessionId'];
    submitted: ReturnType<typeof useEditDialogState>['submitted'];
    noTransactionDraft: ReturnType<typeof useEditDialogState>['noTransactionDraft'];
    addByTemplateId: ReturnType<typeof useEditDialogState>['addByTemplateId'];
    duplicateFromId: ReturnType<typeof useEditDialogState>['duplicateFromId'];
    updateTransactionModelByAfterSaveAction: ReturnType<typeof useEditDialogState>['updateTransactionModelByAfterSaveAction'];
    initOptions: ReturnType<typeof useEditDialogState>['initOptions'];
    resolveFunc: ReturnType<typeof useEditDialogState>['resolveFunc'];
    showState: ReturnType<typeof useEditDialogState>['showState'];
    confirmDialog: ReturnType<typeof useEditDialogState>['confirmDialog'];
    userStore: ReturnType<typeof useEditDialogState>['userStore'];
    transactionTemplatesStore: ReturnType<typeof useEditDialogState>['transactionTemplatesStore'];
}

export function useEditDialogSaving(context: Context) {
    const {
        inputEmptyProblemMessage, snackbar, props, mode, submitting, transactionsStore, transaction, defaultCurrency,
        clientSessionId, submitted, noTransactionDraft, addByTemplateId, duplicateFromId, updateTransactionModelByAfterSaveAction,
        initOptions, resolveFunc, showState, confirmDialog, userStore, transactionTemplatesStore,
    } = context;

    function save(afterAction: AfterSaveAction): void {
        const problemMessage = inputEmptyProblemMessage.value;

        if (problemMessage) {
            snackbar.value?.showMessage(problemMessage);
            return;
        }

        if (props.type === TransactionEditPageType.Transaction && (mode.value === TransactionEditPageMode.Add || mode.value === TransactionEditPageMode.Edit)) {
            const doSubmit = function () {
                submitting.value = true;

                transactionsStore.saveTransaction({
                    transaction: transaction.value as Transaction,
                    defaultCurrency: defaultCurrency.value,
                    isEdit: mode.value === TransactionEditPageMode.Edit,
                    clientSessionId: clientSessionId.value
                }).then(() => {
                    submitting.value = false;
                    submitted.value = true;

                    if (mode.value === TransactionEditPageMode.Add && !noTransactionDraft.value && !addByTemplateId.value && !duplicateFromId.value) {
                        transactionsStore.clearTransactionDraft();
                    }

                    if (mode.value === TransactionEditPageMode.Add && (afterAction === AfterSaveAction.StayWithNewTransaction || afterAction === AfterSaveAction.StayWithCurrentTransaction)) {
                        snackbar.value?.showMessage('You have added a new transaction');
                        updateTransactionModelByAfterSaveAction(afterAction, initOptions.value);
                        clientSessionId.value = generateRandomUUID();
                    } else {
                        if (resolveFunc.value) {
                            if (mode.value === TransactionEditPageMode.Add) {
                                resolveFunc.value({
                                    message: 'You have added a new transaction'
                                });
                            } else if (mode.value === TransactionEditPageMode.Edit) {
                                resolveFunc.value({
                                    message: 'You have saved this transaction'
                                });
                            }
                        }

                        showState.value = false;
                    }
                }).catch(error => {
                    submitting.value = false;

                    if (error.error && (error.error.errorCode === KnownErrorCode.TransactionCannotCreateInThisTime || error.error.errorCode === KnownErrorCode.TransactionCannotModifyInThisTime)) {
                        confirmDialog.value?.open('You have set this time range to prevent editing transactions. Would you like to change the editable transaction range to All?').then(() => {
                            submitting.value = true;

                            userStore.updateUserTransactionEditScope({
                                transactionEditScope: TransactionEditScopeType.All.type
                            }).then(() => {
                                submitting.value = false;

                                snackbar.value?.showMessage('Your editable transaction range has been set to All');
                            }).catch(error => {
                                submitting.value = false;

                                if (!error.processed) {
                                    snackbar.value?.showError(error);
                                }
                            });
                        });
                    } else if (!error.processed) {
                        snackbar.value?.showError(error);
                    }
                });
            };

            if (transaction.value.sourceAmount === 0) {
                confirmDialog.value?.open('Are you sure you want to save this transaction with a zero amount?').then(() => {
                    doSubmit();
                });
            } else {
                doSubmit();
            }
        } else if (props.type === TransactionEditPageType.Template && (mode.value === TransactionEditPageMode.Add || mode.value === TransactionEditPageMode.Edit)) {
            submitting.value = true;

            transactionTemplatesStore.saveTemplateContent({
                template: transaction.value as TransactionTemplate,
                isEdit: mode.value === TransactionEditPageMode.Edit,
                clientSessionId: clientSessionId.value
            }).then(() => {
                submitting.value = false;

                if (resolveFunc.value) {
                    if (mode.value === TransactionEditPageMode.Add) {
                        resolveFunc.value({
                            message: 'You have added a new template'
                        });
                    } else if (mode.value === TransactionEditPageMode.Edit) {
                        resolveFunc.value({
                            message: 'You have saved this template'
                        });
                    }
                }

                showState.value = false;
            }).catch(error => {
                submitting.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        }
    }

    return { save };
}
