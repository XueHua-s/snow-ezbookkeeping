import logger from '@/lib/logger.ts';
import type { useEditDialogState } from './state.ts';

interface Context {
    recognizing: ReturnType<typeof useEditDialogState>['recognizing'];
    loading: ReturnType<typeof useEditDialogState>['loading'];
    submitting: ReturnType<typeof useEditDialogState>['submitting'];
    transactionsStore: ReturnType<typeof useEditDialogState>['transactionsStore'];
    updateTransactionModelFromRecognizedResponse: ReturnType<typeof useEditDialogState>['updateTransactionModelFromRecognizedResponse'];
    snackbar: ReturnType<typeof useEditDialogState>['snackbar'];
    pastedText: ReturnType<typeof useEditDialogState>['pastedText'];
    settingsStore: ReturnType<typeof useEditDialogState>['settingsStore'];
    showPasteTextDialog: ReturnType<typeof useEditDialogState>['showPasteTextDialog'];
}

export function useEditDialogRecognition(context: Context) {
    const { recognizing, loading, submitting, transactionsStore, updateTransactionModelFromRecognizedResponse, snackbar, pastedText, settingsStore, showPasteTextDialog } = context;

    function recognizeText(text: string): void {
        if (recognizing.value || loading.value || submitting.value) {
            return;
        }

        if (!text || !text.trim()) {
            return;
        }

        recognizing.value = true;

        transactionsStore.recognizeTransactionText({ text }).then(response => {
            updateTransactionModelFromRecognizedResponse(response);
            recognizing.value = false;
        }).catch(error => {
            recognizing.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function recognizeFromClipboard(): void {
        if (recognizing.value || loading.value || submitting.value) {
            return;
        }

        pastedText.value = '';

        navigator.clipboard.readText().then(text => {
            pastedText.value = text && text.trim() ? text.trim() : '';

            if (pastedText.value && !settingsStore.appSettings.alwaysRequireConfirmationOfClipboardContentBeforeSubmission) {
                recognizeText(pastedText.value);
            } else {
                showPasteTextDialog.value = true;
            }
        }).catch(error => {
            logger.error('failed to read clipboard', error);
            showPasteTextDialog.value = true;
        });
    }

    return { recognizeText, recognizeFromClipboard };
}
