import {
    isFunction
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import { startDownloadFile } from '@/lib/ui/common.ts';
import { type Transaction } from '@/models/transaction.ts';
import type { useListPageLoading } from './loading.ts';
import type { useListPageState } from './state.ts';

interface Context {
    add: ReturnType<typeof useListPageLoading>['add'];
    aiImageRecognitionDialog: ReturnType<typeof useListPageState>['aiImageRecognitionDialog'];
    settingsStore: ReturnType<typeof useListPageState>['settingsStore'];
    editDialog: ReturnType<typeof useListPageState>['editDialog'];
    snackbar: ReturnType<typeof useListPageState>['snackbar'];
    reload: ReturnType<typeof useListPageLoading>['reload'];
    importDialog: ReturnType<typeof useListPageState>['importDialog'];
    exportingData: ReturnType<typeof useListPageState>['exportingData'];
    userStore: ReturnType<typeof useListPageState>['userStore'];
    tt: ReturnType<typeof useListPageState>['tt'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
}

export function useListPageActions(context: Context) {
    const { add, aiImageRecognitionDialog, settingsStore, editDialog, snackbar, reload, importDialog, exportingData, userStore, tt, transactionsStore } = context;

    function addByRecognizingClipboardText(): void {
        if (navigator.clipboard && isFunction(navigator.clipboard.readText)) {
            navigator.clipboard.readText().then(text => {
                const clipboardText = text && text.trim() ? text.trim() : '';
                add(undefined, clipboardText);
            }).catch(error => {
                logger.error('failed to read clipboard', error);
                add(undefined, '');
            });
        } else {
            add(undefined, '');
        }
    }

    function addByRecognizingImage(): void {
        aiImageRecognitionDialog.value?.open().then(result => {
            const recognizedResponse = result.response;
            const autoUploadRecognizedImage = settingsStore.appSettings.autoUploadTransactionPictureForAIRecognition;

            editDialog.value?.open({
                time: recognizedResponse.time,
                type: recognizedResponse.type,
                categoryId: recognizedResponse.categoryId,
                accountId: recognizedResponse.sourceAccountId,
                destinationAccountId: recognizedResponse.destinationAccountId,
                amount: recognizedResponse.sourceAmount,
                destinationAmount: recognizedResponse.destinationAmount,
                tagIds: recognizedResponse.tagIds ? recognizedResponse.tagIds.join(',') : undefined,
                comment: recognizedResponse.comment,
                autoUploadPicture: autoUploadRecognizedImage ? result.imageFile : undefined,
                noTransactionDraft: true
            }).then(result => {
                if (result && result.message) {
                    snackbar.value?.showMessage(result.message);
                }

                reload(false, false);
            }).catch(error => {
                if (error) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    function importTransaction(): void {
        importDialog.value?.open().then(() => {
            reload(false, false);
        }).catch(error => {
            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }

    function exportTransactions(fileExtension: string): void {
        if (exportingData.value) {
            return;
        }

        const nickname = userStore.currentUserNickname;
        let exportFileName = '';

        if (nickname) {
            exportFileName = tt('dataExport.exportFilename', {
                nickname: nickname
            }) + '.' + fileExtension;
        } else {
            exportFileName = tt('dataExport.defaultExportFilename') + '.' + fileExtension;
        }

        const exportTransactionReq = transactionsStore.getExportTransactionDataRequestByTransactionFilter();

        exportingData.value = true;

        userStore.getExportedUserData(fileExtension, exportTransactionReq).then(data => {
            startDownloadFile(exportFileName, data);
            exportingData.value = false;
        }).catch(error => {
            exportingData.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function show(transaction: Transaction): void {
        editDialog.value?.open({
            id: transaction.id,
            currentTransaction: transaction
        }).then(result => {
            if (result && result.message) {
                snackbar.value?.showMessage(result.message);
            }

            reload(false, false);
        }).catch(error => {
            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }

    return { addByRecognizingClipboardText, addByRecognizingImage, importTransaction, exportTransactions, show };
}
