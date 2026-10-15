import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type {
RecognizedTransactionResponse
} from '@/models/large_language_model.ts';


// Owns recognition for the transaction store.
export function useTransactionRecognition() {
    function recognizeTransactionText({ text }: { text: string }): Promise<RecognizedTransactionResponse> {
        return new Promise((resolve, reject) => {
            services.recognizeTransactionText({ text }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to recognize text' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to recognize text', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to recognize text' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function recognizeReceiptImage({ imageFile, cancelableUuid }: { imageFile: File, cancelableUuid?: string }): Promise<RecognizedTransactionResponse> {
        return new Promise((resolve, reject) => {
            services.recognizeReceiptImage({ imageFile, cancelableUuid }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to recognize image' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                if (error.canceled) {
                    reject(error);
                }

                logger.error('failed to recognize image', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to recognize image' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function cancelRecognizeReceiptImage(cancelableUuid: string): void {
        services.cancelRequest(cancelableUuid);
    }

    return { recognizeTransactionText, recognizeReceiptImage, cancelRecognizeReceiptImage };
}
