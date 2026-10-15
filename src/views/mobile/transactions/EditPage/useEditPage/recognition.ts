import { TRANSACTION_MAX_AMOUNT, TRANSACTION_MIN_AMOUNT } from '@/consts/transaction.ts';
import logger from '@/lib/logger.ts';
import { closeAllDialog, isiOS } from '@/lib/ui/mobile.ts';
import {
    TransactionEditPageMode
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { useEditPageState } from './state.ts';

interface Context {
    recognizing: ReturnType<typeof useEditPageState>['recognizing'];
    loading: ReturnType<typeof useEditPageState>['loading'];
    submitting: ReturnType<typeof useEditPageState>['submitting'];
    showCancelableLoading: ReturnType<typeof useEditPageState>['showCancelableLoading'];
    transactionsStore: ReturnType<typeof useEditPageState>['transactionsStore'];
    updateTransactionModelFromRecognizedResponse: ReturnType<typeof useEditPageState>['updateTransactionModelFromRecognizedResponse'];
    showToast: ReturnType<typeof useEditPageState>['showToast'];
    pastedText: ReturnType<typeof useEditPageState>['pastedText'];
    isSupportClipboard: ReturnType<typeof useEditPageState>['isSupportClipboard'];
    settingsStore: ReturnType<typeof useEditPageState>['settingsStore'];
    showAITextRecognitionSheet: ReturnType<typeof useEditPageState>['showAITextRecognitionSheet'];
    mode: ReturnType<typeof useEditPageState>['mode'];
    parseAmountFromLocalizedNumerals: ReturnType<typeof useEditPageState>['parseAmountFromLocalizedNumerals'];
    transaction: ReturnType<typeof useEditPageState>['transaction'];
}

export function useEditPageRecognition(context: Context) {
    const {
        recognizing, loading, submitting, showCancelableLoading, transactionsStore, updateTransactionModelFromRecognizedResponse,
        showToast, pastedText, isSupportClipboard, settingsStore, showAITextRecognitionSheet, mode,
        parseAmountFromLocalizedNumerals, transaction,
    } = context;

    function recognizeText(text: string): void {
        if (recognizing.value || loading.value || submitting.value) {
            return;
        }

        if (!text || !text.trim()) {
            return;
        }

        recognizing.value = true;
        showCancelableLoading('Recognizing', 'AI can make mistakes. Check important info.');

        transactionsStore.recognizeTransactionText({ text }).then(response => {
            updateTransactionModelFromRecognizedResponse(response);
            closeAllDialog();
            recognizing.value = false;
        }).catch(error => {
            closeAllDialog();
            recognizing.value = false;

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function recognizeFromClipboard(): void {
        if (recognizing.value || loading.value || submitting.value) {
            return;
        }

        pastedText.value = '';

        if (isSupportClipboard && !isiOS()) {
            navigator.clipboard.readText().then(text => {
                pastedText.value = text && text.trim() ? text.trim() : '';

                if (pastedText.value && !settingsStore.appSettings.alwaysRequireConfirmationOfClipboardContentBeforeSubmission) {
                    recognizeText(pastedText.value);
                } else {
                    showAITextRecognitionSheet.value = true;
                }
            }).catch(error => {
                logger.error('failed to read clipboard', error);
                showAITextRecognitionSheet.value = true;
            });
        } else {
            showAITextRecognitionSheet.value = true;
        }
    }

    function pasteAmount(type: 'sourceAmount' | 'destinationAmount'): void {
        if (mode.value === TransactionEditPageMode.View || !isSupportClipboard) {
            return;
        }

        navigator.clipboard.readText().then(text => {
            if (!text) {
                return;
            }

            const parsedAmount = parseAmountFromLocalizedNumerals(text);

            if (Number.isNaN(parsedAmount) || !Number.isFinite(parsedAmount)) {
                showToast('Cannot parse amount from clipboard');
                return;
            }

            if (parsedAmount < TRANSACTION_MIN_AMOUNT || parsedAmount > TRANSACTION_MAX_AMOUNT) {
                showToast('Numeric Overflow');
                return;
            }

            if (type === 'sourceAmount') {
                transaction.value.sourceAmount = parsedAmount;
            } else if (type === 'destinationAmount') {
                transaction.value.destinationAmount = parsedAmount;
            }
        }).catch(error => {
            logger.error('failed to read clipboard text', error);
            showToast('Unable to read clipboard text');
        });
    }

    return { recognizeText, recognizeFromClipboard, pasteAmount };
}
