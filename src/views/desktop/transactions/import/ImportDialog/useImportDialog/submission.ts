import { isNumber } from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import type { useImportDialogState } from './state.ts';

interface Context {
    importTransactionCheckDataTab: ReturnType<typeof useImportDialogState>['importTransactionCheckDataTab'];
    importTransactions: ReturnType<typeof useImportDialogState>['importTransactions'];
    snackbar: ReturnType<typeof useImportDialogState>['snackbar'];
    confirmDialog: ReturnType<typeof useImportDialogState>['confirmDialog'];
    formatNumberToLocalizedNumerals: ReturnType<typeof useImportDialogState>['formatNumberToLocalizedNumerals'];
    submitting: ReturnType<typeof useImportDialogState>['submitting'];
    transactionsStore: ReturnType<typeof useImportDialogState>['transactionsStore'];
    clientSessionId: ReturnType<typeof useImportDialogState>['clientSessionId'];
    importProcess: ReturnType<typeof useImportDialogState>['importProcess'];
    importedCount: ReturnType<typeof useImportDialogState>['importedCount'];
    currentStep: ReturnType<typeof useImportDialogState>['currentStep'];
    accountsStore: ReturnType<typeof useImportDialogState>['accountsStore'];
    overviewStore: ReturnType<typeof useImportDialogState>['overviewStore'];
    statisticsStore: ReturnType<typeof useImportDialogState>['statisticsStore'];
}

export function useImportDialogSubmission(context: Context) {
    const {
        importTransactionCheckDataTab, importTransactions, snackbar, confirmDialog, formatNumberToLocalizedNumerals, submitting,
        transactionsStore, clientSessionId, importProcess, importedCount, currentStep, accountsStore, overviewStore,
        statisticsStore,
    } = context;

    function batchApplyRules(): void {
        if (importTransactionCheckDataTab.value?.isEditing) {
            return;
        }

        importTransactionCheckDataTab.value?.showBatchApplyRulesDialog();
    }

    function submit(): void {
        if (importTransactionCheckDataTab.value?.isEditing) {
            return;
        }

        const transactions: ImportTransaction[] = [];

        if (importTransactions.value) {
            for (const importTransaction of importTransactions.value) {
                if (importTransaction.valid && importTransaction.selected) {
                    transactions.push(importTransaction);
                } else if (!importTransaction.valid && importTransaction.selected) {
                    snackbar.value?.showError('Cannot import invalid transactions');
                    return;
                }
            }
        }

        if (transactions.length < 1) {
            snackbar.value?.showError('No data to import');
            return;
        }

        confirmDialog.value?.open('format.misc.confirmImportTransactions', {
            count: formatNumberToLocalizedNumerals(transactions.length)
        }).then(() => {
            submitting.value = true;

            let showProcessTimer : number | undefined = undefined;

            if (transactions.length > 100) {
                setTimeout(() => {
                    if (!submitting.value) {
                        logger.warn('transaction import is not submitting');
                        return;
                    }

                    // @ts-expect-error the return value of setInterval is number, but lint shows it as NodeJS.Timer
                    showProcessTimer = setInterval(() => {
                        if (submitting.value) {
                            transactionsStore.getImportTransactionsProcess({
                                clientSessionId: clientSessionId.value
                            }).then(response => {
                                if (isNumber(response) && 0 <= response && response < 100) {
                                    importProcess.value = response;
                                } else {
                                    importProcess.value = 0;
                                    clearInterval(showProcessTimer);
                                    showProcessTimer = undefined;
                                }
                            }).catch(() => {
                                importProcess.value = 0;
                                clearInterval(showProcessTimer);
                                showProcessTimer = undefined;
                            });
                        }
                    }, 2000);
                }, 2000);
            }

            transactionsStore.importTransactions({
                transactions: transactions,
                clientSessionId: clientSessionId.value
            }).then(response => {
                if (showProcessTimer) {
                    importProcess.value = 0;
                    clearInterval(showProcessTimer);
                    showProcessTimer = undefined;
                }

                importedCount.value = response;
                currentStep.value = 'finalResult';

                accountsStore.updateAccountListInvalidState(true);
                transactionsStore.updateTransactionListInvalidState(true);
                overviewStore.updateTransactionOverviewInvalidState(true);
                statisticsStore.updateTransactionStatisticsInvalidState(true);

                submitting.value = false;
            }).catch(error => {
                if (showProcessTimer) {
                    importProcess.value = 0;
                    clearInterval(showProcessTimer);
                    showProcessTimer = undefined;
                }

                submitting.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    return { batchApplyRules, submit };
}
