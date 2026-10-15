import { KnownErrorCode } from '@/consts/api.ts';
import { KnownFileType } from '@/core/file.ts';
import { compressJpgImageByQuality } from '@/lib/ui/common.ts';
import type { TransactionPictureInfoBasicResponse } from '@/models/transaction_picture_info.ts';
import {
    TransactionEditPageMode
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { useEditPageState } from './state.ts';

interface Context {
    canAddTransactionPicture: ReturnType<typeof useEditPageState>['canAddTransactionPicture'];
    submitting: ReturnType<typeof useEditPageState>['submitting'];
    pictureInput: ReturnType<typeof useEditPageState>['pictureInput'];
    uploadingPicture: ReturnType<typeof useEditPageState>['uploadingPicture'];
    imageUploadQualityType: ReturnType<typeof useEditPageState>['imageUploadQualityType'];
    transactionsStore: ReturnType<typeof useEditPageState>['transactionsStore'];
    transaction: ReturnType<typeof useEditPageState>['transaction'];
    showToast: ReturnType<typeof useEditPageState>['showToast'];
    mode: ReturnType<typeof useEditPageState>['mode'];
    pictureBrowser: ReturnType<typeof useEditPageState>['pictureBrowser'];
    showConfirm: ReturnType<typeof useEditPageState>['showConfirm'];
    removingPictureId: ReturnType<typeof useEditPageState>['removingPictureId'];
}

export function useEditPagePictures(context: Context) {
    const {
        canAddTransactionPicture, submitting, pictureInput, uploadingPicture, imageUploadQualityType, transactionsStore,
        transaction, showToast, mode, pictureBrowser, showConfirm, removingPictureId,
    } = context;

    function showOpenPictureDialog(): void {
        if (!canAddTransactionPicture.value || submitting.value) {
            return;
        }

        pictureInput.value?.click();
    }

    function uploadPicture(file: File): void {
        if (!file) {
            return;
        }

        uploadingPicture.value = true;
        submitting.value = true;

        compressJpgImageByQuality(file, imageUploadQualityType.value).then(blob => {
            return transactionsStore.uploadTransactionPicture({
                pictureFile: KnownFileType.JPG.createFileFromBlob(blob, "image")
            });
        }).then(response => {
            transaction.value.addPicture(response);
            uploadingPicture.value = false;
            submitting.value = false;
        }).catch(error => {
            uploadingPicture.value = false;
            submitting.value = false;

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function viewOrRemovePicture(pictureInfo: TransactionPictureInfoBasicResponse): void {
        if (mode.value !== TransactionEditPageMode.Add && mode.value !== TransactionEditPageMode.Edit && transaction.value.pictures && transaction.value.pictures.length) {
            pictureBrowser.value?.open();
            return;
        }

        showConfirm('Are you sure you want to remove this transaction picture?', () => {
            removingPictureId.value = pictureInfo.pictureId;
            submitting.value = true;

            transactionsStore.removeUnusedTransactionPicture({ pictureInfo }).then(response => {
                if (response) {
                    transaction.value.removePicture(pictureInfo);
                }

                removingPictureId.value = '';
                submitting.value = false;
            }).catch(error => {
                if (error.error && error.error.errorCode === KnownErrorCode.TransactionPictureNotFound) {
                    transaction.value.removePicture(pictureInfo);
                } else if (!error.processed) {
                    showToast(error.message || error);
                }

                removingPictureId.value = '';
                submitting.value = false;
            });
        });
    }

    function onUploadPicture(event: Event): void {
        if (!event || !event.target) {
            return;
        }

        const el = event.target as HTMLInputElement;

        if (!el.files || !el.files.length || !el.files[0]) {
            return;
        }

        const pictureFile = el.files[0] as File;

        el.value = '';
        uploadPicture(pictureFile);
    }

    return { showOpenPictureDialog, uploadPicture, viewOrRemovePicture, onUploadPicture };
}
