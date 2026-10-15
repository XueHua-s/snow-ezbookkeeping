import {
    DEFAULT_IMPORT_API_TIMEOUT,
    DEFAULT_LLM_API_TIMEOUT,
    DEFAULT_UPLOAD_API_TIMEOUT
} from '@/consts/api.ts';
import type { ApiResponse } from '@/core/api.ts';
import type {
ImportFileTypeSupportedAdditionalOptions
} from '@/core/file.ts';
import {
    TransactionType
} from '@/core/transaction.ts';
import {
    objectFieldWithValueToArrayItem
} from '@/lib/common.ts';
import type {
ImportTransactionResponsePageWrapper
} from '@/models/imported_transaction.ts';
import type {
TransactionImportRequest
} from '@/models/transaction.ts';
import type {
TransactionPictureInfoBasicResponse,
TransactionPictureUnusedDeleteRequest
} from '@/models/transaction_picture_info.ts';
import axios from 'axios';

import { type ApiRequestConfig, type ApiResponsePromise } from './transport.ts';

export const importServices = {
    parseImportCustomFile: ({ fileType, fileEncoding, importFile }: { fileType: string, fileEncoding?: string, importFile: File }): ApiResponsePromise<string[][]> => {
        return axios.postForm<ApiResponse<string[][]>>('v1/transactions/parse_custom_file.json', {
            fileType: fileType,
            fileEncoding: fileEncoding,
            file: importFile
        }, {
            timeout: DEFAULT_UPLOAD_API_TIMEOUT
        } as ApiRequestConfig);
    },
    parseImportTransaction: ({
        fileType, additionalOptions, aiAdditionalPrompt, fileEncoding, importFile, columnMapping, transactionTypeMapping,
        hasHeaderLine, timeFormat, timezoneFormat, amountDecimalSeparator, amountDigitGroupingSymbol, geoSeparator, geoOrder,
        tagSeparator, cancelableUuid,
    }: { fileType: string, additionalOptions?: ImportFileTypeSupportedAdditionalOptions, aiAdditionalPrompt?: string, fileEncoding?: string, importFile: File, columnMapping?: Record<number, number>, transactionTypeMapping?: Record<string, TransactionType>, hasHeaderLine?: boolean, timeFormat?: string, timezoneFormat?: string, amountDecimalSeparator?: string, amountDigitGroupingSymbol?: string, geoSeparator?: string, geoOrder?: string, tagSeparator?: string, cancelableUuid?: string }): ApiResponsePromise<ImportTransactionResponsePageWrapper> => {
        let textualAdditionalOptions: string | undefined = undefined;
        let textualColumnMapping: string | undefined = undefined;
        let textualTransactionTypeMapping: string | undefined = undefined;
        let textualHasHeaderLine: string | undefined = undefined;

        if (additionalOptions) {
            textualAdditionalOptions = objectFieldWithValueToArrayItem(additionalOptions, true).join(',');
        }

        if (columnMapping) {
            textualColumnMapping = JSON.stringify(columnMapping);
        }

        if (transactionTypeMapping) {
            textualTransactionTypeMapping = JSON.stringify(transactionTypeMapping);
        }

        if (hasHeaderLine) {
            textualHasHeaderLine = 'true';
        }

        let timeout: number = DEFAULT_UPLOAD_API_TIMEOUT;

        if (fileType === 'ai_txt' || fileType === 'ai_image') {
            timeout = DEFAULT_LLM_API_TIMEOUT;
        }

        return axios.postForm<ApiResponse<ImportTransactionResponsePageWrapper>>('v1/transactions/parse_import.json', {
            fileType: fileType,
            options: textualAdditionalOptions,
            aiPrompt: aiAdditionalPrompt,
            fileEncoding: fileEncoding,
            file: importFile,
            columnMapping: textualColumnMapping,
            transactionTypeMapping: textualTransactionTypeMapping,
            hasHeaderLine: textualHasHeaderLine,
            timeFormat: timeFormat,
            timezoneFormat: timezoneFormat,
            amountDecimalSeparator: amountDecimalSeparator,
            amountDigitGroupingSymbol: amountDigitGroupingSymbol,
            geoSeparator: geoSeparator,
            geoOrder: geoOrder,
            tagSeparator: tagSeparator
        }, {
            timeout: timeout,
            cancelableUuid: cancelableUuid
        } as ApiRequestConfig);
    },
    importTransactions: (req: TransactionImportRequest): ApiResponsePromise<number> => {
        return axios.post<ApiResponse<number>>('v1/transactions/import.json', req, {
            timeout: DEFAULT_IMPORT_API_TIMEOUT
        } as ApiRequestConfig);
    },
    getImportTransactionsProcess: (clientSessionId: string): ApiResponsePromise<number | null> => {
        return axios.get<ApiResponse<number | null>>('v1/transactions/import/process.json?client_session_id=' + clientSessionId, {
            ignoreError: true
        } as ApiRequestConfig);
    },
    uploadTransactionPicture: ({ pictureFile, clientSessionId }: { pictureFile: File, clientSessionId?: string }): ApiResponsePromise<TransactionPictureInfoBasicResponse> => {
        return axios.postForm<ApiResponse<TransactionPictureInfoBasicResponse>>('v1/transaction/pictures/upload.json', {
            picture: pictureFile,
            clientSessionId: clientSessionId
        }, {
            timeout: DEFAULT_UPLOAD_API_TIMEOUT
        } as ApiRequestConfig);
    },
    removeUnusedTransactionPicture: (req: TransactionPictureUnusedDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/pictures/remove_unused.json', req);
    }
};
