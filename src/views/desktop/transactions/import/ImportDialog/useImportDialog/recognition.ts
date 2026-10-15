import {
    KnownFileType
} from '@/core/file.ts';
import { ImageUploadQualityType } from '@/core/image.ts';
import logger from '@/lib/logger.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import { compressJpgImageByQuality } from '@/lib/ui/common.ts';
import { type ImportTransactionResponse, ImportTransaction } from '@/models/imported_transaction.ts';
import { type BatchImportImageItem } from '@/views/desktop/transactions/import/tabs/ImportTransactionRecognizeImagesTab.vue';
import type { useImportDialogState } from './state.ts';

interface Context {
    importImageCancelRecognizingUuid: ReturnType<typeof useImportDialogState>['importImageCancelRecognizingUuid'];
    transactionsStore: ReturnType<typeof useImportDialogState>['transactionsStore'];
    importImageFiles: ReturnType<typeof useImportDialogState>['importImageFiles'];
    snackbar: ReturnType<typeof useImportDialogState>['snackbar'];
    submitting: ReturnType<typeof useImportDialogState>['submitting'];
    supportedAIAdditionalPrompt: ReturnType<typeof useImportDialogState>['supportedAIAdditionalPrompt'];
    importAIAdditionalPrompt: ReturnType<typeof useImportDialogState>['importAIAdditionalPrompt'];
    importTransactions: ReturnType<typeof useImportDialogState>['importTransactions'];
    te: ReturnType<typeof useImportDialogState>['te'];
}

export function useImportDialogRecognition(context: Context) {
    const {
        importImageCancelRecognizingUuid, transactionsStore, importImageFiles, snackbar, submitting, supportedAIAdditionalPrompt,
        importAIAdditionalPrompt, importTransactions, te,
    } = context;

    function recognizeImage(item: BatchImportImageItem, additionalPrompt?: string): Promise<ImportTransactionResponse[]> {
        return new Promise<ImportTransactionResponse[]>((resolve, reject) => {
            compressJpgImageByQuality(item.file, ImageUploadQualityType.HD720P).then(blob => {
                const compressedFile = KnownFileType.JPG.createFileFromBlob(blob, "image");
                importImageCancelRecognizingUuid.value = generateRandomUUID();

                transactionsStore.parseImportTransaction({
                    fileType: 'ai_image',
                    aiAdditionalPrompt: additionalPrompt,
                    importFile: compressedFile,
                    cancelableUuid: importImageCancelRecognizingUuid.value
                }).then(response => {
                    resolve(response.items || []);
                }).catch(error => {
                    reject(error);
                });
            }).catch(error => {
                logger.error('failed to compress image', error);
                reject('Unable to load image');
            });
        });
    }

    function batchRecognizeImages(): Promise<void> {
        for (const item of importImageFiles.value) {
            item.status = 'pending';
            item.failureReason = undefined;
        }

        return new Promise<void>((resolve, reject) => {
            const finish = function (): void {
                const allFailed = importImageFiles.value.every(item => item.status === 'failed');
                importImageCancelRecognizingUuid.value = '';

                if (allFailed) {
                    snackbar.value?.showError('Unable to recognize image');
                    reject();
                } else {
                    resolve();
                }
            }

            const recurseRecognizeImage = function (index: number): void {
                if (!submitting.value) {
                    finish();
                    return;
                }

                if (index >= importImageFiles.value.length) {
                    finish();
                    return;
                }

                const item = importImageFiles.value[index];

                if (!item) {
                    finish();
                    return;
                }

                item.status = 'recognizing';

                recognizeImage(item, supportedAIAdditionalPrompt.value ? importAIAdditionalPrompt.value : undefined).then(results => {
                    if (submitting.value) {
                        if (!importTransactions.value) {
                            importTransactions.value = [];
                        }

                        for (const importTransactionResp of results) {
                            const index = importTransactions.value.length;
                            importTransactions.value.push(ImportTransaction.of(importTransactionResp, index));
                        }

                        item.status = 'success';
                    }

                    recurseRecognizeImage(index + 1);
                }).catch(error => {
                    if (submitting.value) {
                        item.status = 'failed';
                        item.failureReason = te(error.message || error || 'An error occurred');
                    }

                    recurseRecognizeImage(index + 1);
                });
            }

            recurseRecognizeImage(0);
        });
    }

    function cancelBatchRecognizeImages(): void {
        submitting.value = false;

        if (importImageCancelRecognizingUuid.value) {
            transactionsStore.cancelRecognizeReceiptImage(importImageCancelRecognizingUuid.value);
        }

        snackbar.value?.showMessage('User Canceled');
    }

    return { recognizeImage, batchRecognizeImages, cancelBatchRecognizeImages };
}
