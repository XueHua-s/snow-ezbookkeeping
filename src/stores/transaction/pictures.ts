import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type {
TransactionPictureInfoBasicResponse
} from '@/models/transaction_picture_info.ts';


// Owns pictures for the transaction store.
export function useTransactionPictures() {
    function uploadTransactionPicture({ pictureFile, clientSessionId }: { pictureFile: File, clientSessionId?: string }): Promise<TransactionPictureInfoBasicResponse> {
        return new Promise((resolve, reject) => {
            services.uploadTransactionPicture({ pictureFile, clientSessionId }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to upload transaction picture' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('Unable to upload transaction picture', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to upload transaction picture' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function removeUnusedTransactionPicture({ pictureInfo }: { pictureInfo: TransactionPictureInfoBasicResponse }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.removeUnusedTransactionPicture({ id: pictureInfo.pictureId }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to remove transaction picture' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to remove transaction picture', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to remove transaction picture' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function getTransactionPictureUrl(pictureInfo?: TransactionPictureInfoBasicResponse | null, disableBrowserCache?: boolean | string): string | undefined {
        if (!pictureInfo || !pictureInfo.originalUrl) {
            return undefined;
        }

        return services.getTransactionPictureUrlWithToken(pictureInfo.originalUrl, disableBrowserCache);
    }

    return { uploadTransactionPicture, removeUnusedTransactionPicture, getTransactionPictureUrl };
}
