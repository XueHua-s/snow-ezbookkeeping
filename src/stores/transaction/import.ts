import type { ImportFileTypeSupportedAdditionalOptions } from '@/core/file.ts';
import type { TransactionType } from '@/core/transaction.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type { ImportTransactionResponsePageWrapper, ImportTransaction } from '@/models/imported_transaction.ts';
import type { TransactionCreateRequest } from '@/models/transaction.ts';


// Owns import for the transaction store.
export function useTransactionImport() {
    function parseImportCustomFile({ fileType, fileEncoding, importFile }: { fileType: string, fileEncoding?: string, importFile: File }): Promise<string[][]> {
        return new Promise((resolve, reject) => {
            services.parseImportCustomFile({ fileType, fileEncoding, importFile }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to parse import file' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('Unable to parse import file', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to parse import file' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function parseImportTransaction({
        fileType, additionalOptions, aiAdditionalPrompt, fileEncoding, importFile, columnMapping, transactionTypeMapping,
        hasHeaderLine, timeFormat, timezoneFormat, amountDecimalSeparator, amountDigitGroupingSymbol, geoSeparator, geoOrder,
        tagSeparator, cancelableUuid,
    }: { fileType: string, additionalOptions?: ImportFileTypeSupportedAdditionalOptions, aiAdditionalPrompt?: string, fileEncoding?: string, importFile: File, columnMapping?: Record<number, number>, transactionTypeMapping?: Record<string, TransactionType>, hasHeaderLine?: boolean, timeFormat?: string, timezoneFormat?: string, amountDecimalSeparator?: string, amountDigitGroupingSymbol?: string, geoSeparator?: string, geoOrder?: string, tagSeparator?: string, cancelableUuid?: string }): Promise<ImportTransactionResponsePageWrapper> {
        return new Promise((resolve, reject) => {
            services.parseImportTransaction({
                fileType, additionalOptions, aiAdditionalPrompt, fileEncoding, importFile, columnMapping, transactionTypeMapping,
                hasHeaderLine, timeFormat, timezoneFormat, amountDecimalSeparator, amountDigitGroupingSymbol, geoSeparator, geoOrder,
                tagSeparator, cancelableUuid,
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to parse import file' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('Unable to parse import file', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to parse import file' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function importTransactions({ transactions, clientSessionId }: { transactions: ImportTransaction[], clientSessionId: string }): Promise<number> {
        const submitTransactions: TransactionCreateRequest[] = [];

        if (transactions) {
            for (const transaction of transactions) {
                const submitTransaction = transaction.toCreateRequest();
                submitTransactions.push(submitTransaction);
            }
        }

        return new Promise((resolve, reject) => {
            services.importTransactions({
                transactions: submitTransactions,
                clientSessionId: clientSessionId
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to import transactions' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('Unable to import transactions', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to import transactions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function getImportTransactionsProcess({ clientSessionId }: { clientSessionId: string }): Promise<number | null> {
        return new Promise((resolve, reject) => {
            services.getImportTransactionsProcess(clientSessionId).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to get transactions import process' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('Unable to get transactions import process', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to get transactions import process' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { parseImportCustomFile, parseImportTransaction, importTransactions, getImportTransactionsProcess };
}
