import { KnownErrorCode } from '@/consts/api.ts';
import { KnownFileType } from '@/core/file.ts';
import { compressJpgImageByQuality } from '@/lib/ui/common.ts';
import type { TransactionPictureInfoBasicResponse } from '@/models/transaction_picture_info.ts';
import {
    TransactionEditPageMode
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import { nextTick, watch } from 'vue';
import type { useEditDialogState } from './state.ts';

interface Context {
    canAddTransactionPicture: ReturnType<typeof useEditDialogState>['canAddTransactionPicture'];
    submitting: ReturnType<typeof useEditDialogState>['submitting'];
    pictureInput: ReturnType<typeof useEditDialogState>['pictureInput'];
    uploadingPicture: ReturnType<typeof useEditDialogState>['uploadingPicture'];
    imageUploadQualityType: ReturnType<typeof useEditDialogState>['imageUploadQualityType'];
    transactionsStore: ReturnType<typeof useEditDialogState>['transactionsStore'];
    transaction: ReturnType<typeof useEditDialogState>['transaction'];
    snackbar: ReturnType<typeof useEditDialogState>['snackbar'];
    mode: ReturnType<typeof useEditDialogState>['mode'];
    getTransactionPictureUrl: ReturnType<typeof useEditDialogState>['getTransactionPictureUrl'];
    confirmDialog: ReturnType<typeof useEditDialogState>['confirmDialog'];
    removingPictureId: ReturnType<typeof useEditDialogState>['removingPictureId'];
    activeTab: ReturnType<typeof useEditDialogState>['activeTab'];
    map: ReturnType<typeof useEditDialogState>['map'];
}

export function useEditDialogPictures(context: Context) {
    const {
        canAddTransactionPicture, submitting, pictureInput, uploadingPicture, imageUploadQualityType, transactionsStore,
        transaction, snackbar, mode, getTransactionPictureUrl, confirmDialog, removingPictureId, activeTab, map,
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
                snackbar.value?.showError(error);
            }
        });
    }

    function viewOrRemovePicture(pictureInfo: TransactionPictureInfoBasicResponse): void {
        if (mode.value !== TransactionEditPageMode.Add && mode.value !== TransactionEditPageMode.Edit) {
            window.open(getTransactionPictureUrl(pictureInfo), '_blank');
            return;
        }

        confirmDialog.value?.open('Are you sure you want to remove this transaction picture?').then(() => {
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
                    snackbar.value?.showError(error);
                }

                removingPictureId.value = '';
                submitting.value = false;
            });
        });
    }

    function onSavingTag(state: boolean): void {
        submitting.value = state;
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

    function onShowDateTimeError(error: string): void {
        snackbar.value?.showError(error);
    }

    watch(activeTab, (newValue) => {
        if (newValue === 'map') {
            nextTick(() => {
                map.value?.initMapView();
            });
        }
    });

    return { showOpenPictureDialog, uploadPicture, viewOrRemovePicture, onSavingTag, onUploadPicture, onShowDateTimeError };
}
