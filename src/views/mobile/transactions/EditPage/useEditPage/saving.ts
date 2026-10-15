import { KnownErrorCode } from '@/consts/api.ts';
import {
    TransactionEditScopeType,
    TransactionQuickAddButtonActionType
} from '@/core/transaction.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import { hideLoading, showLoading } from '@/lib/ui/mobile.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { Transaction } from '@/models/transaction.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import {
    AfterSaveAction,
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { Router } from 'framework7/types';
import type { useEditPageState } from './state.ts';

interface Context {
    props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
    autoUploadPicture?: File;
    autoRecognizeClipboardText?: string;
}>;
    mode: ReturnType<typeof useEditPageState>['mode'];
    inputEmptyProblemMessage: ReturnType<typeof useEditPageState>['inputEmptyProblemMessage'];
    showAlert: ReturnType<typeof useEditPageState>['showAlert'];
    pageTypeAndMode: ReturnType<typeof useEditPageState>['pageTypeAndMode'];
    submitting: ReturnType<typeof useEditPageState>['submitting'];
    transactionsStore: ReturnType<typeof useEditPageState>['transactionsStore'];
    transaction: ReturnType<typeof useEditPageState>['transaction'];
    defaultCurrency: ReturnType<typeof useEditPageState>['defaultCurrency'];
    clientSessionId: ReturnType<typeof useEditPageState>['clientSessionId'];
    submitted: ReturnType<typeof useEditPageState>['submitted'];
    query: ReturnType<typeof useEditPageState>['query'];
    addByTemplateId: ReturnType<typeof useEditPageState>['addByTemplateId'];
    duplicateFromId: ReturnType<typeof useEditPageState>['duplicateFromId'];
    showToast: ReturnType<typeof useEditPageState>['showToast'];
    updateTransactionModelByAfterSaveAction: ReturnType<typeof useEditPageState>['updateTransactionModelByAfterSaveAction'];
    getQueryTransactionOptions: ReturnType<typeof useEditPageState>['getQueryTransactionOptions'];
    showConfirm: ReturnType<typeof useEditPageState>['showConfirm'];
    userStore: ReturnType<typeof useEditPageState>['userStore'];
    transactionTemplatesStore: ReturnType<typeof useEditPageState>['transactionTemplatesStore'];
    settingsStore: ReturnType<typeof useEditPageState>['settingsStore'];
    showQuickSavePopover: ReturnType<typeof useEditPageState>['showQuickSavePopover'];
}

export function useEditPageSaving(context: Context) {
    const {
        props, mode, inputEmptyProblemMessage, showAlert, pageTypeAndMode, submitting, transactionsStore, transaction,
        defaultCurrency, clientSessionId, submitted, query, addByTemplateId, duplicateFromId, showToast,
        updateTransactionModelByAfterSaveAction, getQueryTransactionOptions, showConfirm, userStore, transactionTemplatesStore,
        settingsStore, showQuickSavePopover,
    } = context;

    function save(afterAction: AfterSaveAction): void {
        const router = props.f7router;

        if (mode.value === TransactionEditPageMode.View) {
            return;
        }

        const problemMessage = inputEmptyProblemMessage.value;

        if (problemMessage) {
            showAlert(problemMessage);
            return;
        }

        if (pageTypeAndMode?.type === TransactionEditPageType.Transaction && (mode.value === TransactionEditPageMode.Add || mode.value === TransactionEditPageMode.Edit)) {
            const doSubmit = function () {
                submitting.value = true;
                showLoading(() => submitting.value);

                transactionsStore.saveTransaction({
                    transaction: transaction.value as Transaction,
                    defaultCurrency: defaultCurrency.value,
                    isEdit: mode.value === TransactionEditPageMode.Edit,
                    clientSessionId: clientSessionId.value
                }).then(() => {
                    submitting.value = false;
                    submitted.value = true;
                    hideLoading();

                    if (mode.value === TransactionEditPageMode.Add && query['noTransactionDraft'] !== 'true' && !addByTemplateId.value && !duplicateFromId.value) {
                        transactionsStore.clearTransactionDraft();
                    }

                    if (mode.value === TransactionEditPageMode.Add && (afterAction === AfterSaveAction.StayWithNewTransaction || afterAction === AfterSaveAction.StayWithCurrentTransaction)) {
                        showToast('You have added a new transaction');
                        updateTransactionModelByAfterSaveAction(afterAction, getQueryTransactionOptions());
                        clientSessionId.value = generateRandomUUID();
                    } else {
                        if (mode.value === TransactionEditPageMode.Add) {
                            showToast('You have added a new transaction');
                        } else if (mode.value === TransactionEditPageMode.Edit) {
                            showToast('You have saved this transaction');
                        }

                        router.back();
                    }
                }).catch(error => {
                    submitting.value = false;
                    hideLoading();

                    if (error.error && (error.error.errorCode === KnownErrorCode.TransactionCannotCreateInThisTime || error.error.errorCode === KnownErrorCode.TransactionCannotModifyInThisTime)) {
                        showConfirm('You have set this time range to prevent editing transactions. Would you like to change the editable transaction range to All?', () => {
                            submitting.value = true;
                            showLoading(() => submitting.value);

                            userStore.updateUserTransactionEditScope({
                                transactionEditScope: TransactionEditScopeType.All.type
                            }).then(() => {
                                submitting.value = false;
                                hideLoading();

                                showToast('Your editable transaction range has been set to All');
                            }).catch(error => {
                                submitting.value = false;
                                hideLoading();

                                if (!error.processed) {
                                    showToast(error.message || error);
                                }
                            });
                        });
                    } else if (!error.processed) {
                        showToast(error.message || error);
                    }
                });
            };

            if (transaction.value.sourceAmount === 0) {
                showConfirm('Are you sure you want to save this transaction with a zero amount?', () => {
                    doSubmit();
                });
            } else {
                doSubmit();
            }
        } else if (pageTypeAndMode?.type === TransactionEditPageType.Template && (mode.value === TransactionEditPageMode.Add || mode.value === TransactionEditPageMode.Edit)) {
            submitting.value = true;
            showLoading(() => submitting.value);

            transactionTemplatesStore.saveTemplateContent({
                template: transaction.value as TransactionTemplate,
                isEdit: mode.value === TransactionEditPageMode.Edit,
                clientSessionId: clientSessionId.value
            }).then(() => {
                submitting.value = false;
                submitted.value = true;
                hideLoading();

                if (mode.value === TransactionEditPageMode.Add) {
                    showToast('You have added a new template');
                } else if (mode.value === TransactionEditPageMode.Edit) {
                    showToast('You have saved this template');
                }

                router.back();
            }).catch(error => {
                submitting.value = false;
                hideLoading();

                if (!error.processed) {
                    showToast(error.message || error);
                }
            });
        }
    }

    function quickSave(): void {
        if (mode.value === TransactionEditPageMode.View) {
            return;
        }

        if (pageTypeAndMode?.type === TransactionEditPageType.Transaction && mode.value === TransactionEditPageMode.Add) {
            const quickAddActionType = settingsStore.appSettings.quickAddButtonActionInMobileTransactionEditPage;

            if (quickAddActionType === TransactionQuickAddButtonActionType.OpenMenu.type) {
                showQuickSavePopover.value = true;
                return;
            } else if (quickAddActionType === TransactionQuickAddButtonActionType.SaveAndAddNewTransaction.type) {
                save(AfterSaveAction.StayWithNewTransaction);
                return;
            } else if (quickAddActionType === TransactionQuickAddButtonActionType.SaveAndKeepCurrentData.type) {
                save(AfterSaveAction.StayWithCurrentTransaction);
                return;
            }
        }

        save(AfterSaveAction.GoBack);
    }

    function duplicate(withTime?: boolean, withGeoLocation?: boolean): void {
        props.f7router.navigate(`/transaction/add?id=${transaction.value.id}&type=${transaction.value.type}&withTime=${withTime ?? false}&withGeoLocation=${withGeoLocation ?? false}`);
    }

    return { save, quickSave, duplicate };
}
