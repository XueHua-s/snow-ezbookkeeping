import { KnownErrorCode, PARAMETERIZED_ERRORS, PARSE_TRANSACTION_API_URL_PATH, SPECIFIED_API_NOT_FOUND_ERRORS, SUPPORTED_PARSE_TRANSACTION_RECORD_ID_TYPES } from '@/consts/api.ts';
import type { ErrorResponse } from '@/core/api.ts';
import {
    isDefined,
    isObject,
    isString
} from '@/lib/common.ts';
import type { useHelpersState } from './state.ts';
import { type LocalizedError, type LocalizedErrorParameter } from './types.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
}

// Owns errors for the helpers store.
export function useHelpersErrors(context: Context) {
    const { t } = context;

    function getLocalizedError(error: ErrorResponse): LocalizedError {
        if (error.errorCode === KnownErrorCode.ApiNotFound && SPECIFIED_API_NOT_FOUND_ERRORS[error.path]) {
            return {
                message: `${SPECIFIED_API_NOT_FOUND_ERRORS[error.path]!.message}`
            };
        }

        if (error.errorCode !== KnownErrorCode.ValidatorError) {
            return {
                message: `error.${error.errorMessage}`
            };
        }

        for (const errorInfo of PARAMETERIZED_ERRORS) {
            const matches = error.errorMessage.match(errorInfo.regex);

            if (matches && matches.length === errorInfo.parameters.length + 1) {
                return {
                    message: `parameterizedError.${errorInfo.localeKey}`,
                    parameters: errorInfo.parameters.map((param, index) => {
                        return {
                            key: param.field,
                            localized: param.localized,
                            value: matches[index + 1] as string
                        }
                    })
                };
            }
        }

        return {
            message: `error.${error.errorMessage}`
        };
    }

    function getLocalizedErrorParameters(parameters?: LocalizedErrorParameter[]): Record<string, string> {
        const localizedParameters: Record<string, string> = {};

        if (parameters) {
            for (const parameter of parameters) {
                if (parameter.localized) {
                    localizedParameters[parameter.key] = t(`parameter.${parameter.value}`);
                } else {
                    localizedParameters[parameter.key] = parameter.value;
                }
            }
        }

        return localizedParameters;
    }

    function getLocalizedErrorForParseImportTransactionsAPI(translatedMessage: string, recordId: string | undefined, originalData: unknown | undefined): string {
        if (!isDefined(recordId) && !isDefined(originalData)) {
            return translatedMessage;
        }

        const localizedRecordIds: string[] = [];

        if (isString(recordId)) {
            for (const recordIdItem of recordId.split('-')) {
                const parts = recordIdItem.split(':');

                if (parts.length !== 2) {
                    localizedRecordIds.length = 0;
                    break;
                }

                const recordIdType = parts[0]!;
                const recordIdIndex = parts[1]!;
                const recordIdTypeI18nKey = SUPPORTED_PARSE_TRANSACTION_RECORD_ID_TYPES[recordIdType];

                if (!recordIdTypeI18nKey || !recordIdIndex) {
                    localizedRecordIds.length = 0;
                    break;
                }

                localizedRecordIds.push(t(recordIdTypeI18nKey, {
                    index: recordIdIndex
                }));
            }
        }

        const textualOriginalData: string | undefined = isString(originalData) ? (originalData.length > 0 ? originalData : undefined) : (isDefined(originalData) ? JSON.stringify(originalData) : undefined);
        const textualRecordId: string | undefined = localizedRecordIds.length > 0 ? localizedRecordIds.join(t('format.misc.multiIndexJoinSeparator')) : undefined;

        if (isDefined(textualRecordId) && isDefined(textualOriginalData)) {
            return t('format.misc.importTransactionParseErrorWithRecordIdAndData', {
                error: translatedMessage,
                recordId: textualRecordId,
                data: textualOriginalData
            });
        } else if (isDefined(textualRecordId)) {
            return t('format.misc.importTransactionParseErrorWithRecordId', {
                error: translatedMessage,
                recordId: textualRecordId
            });
        } else if (isDefined(textualOriginalData)) {
            return t('format.misc.importTransactionParseErrorWithData', {
                error: translatedMessage,
                data: textualOriginalData
            });
        } else {
            return translatedMessage;
        }
    }

    function translateError(message: string | { error: ErrorResponse }): string {
        let finalMessage = '';
        let parameters = {};

        if (isObject(message) && isObject(message.error)) {
            const localizedError = getLocalizedError(message.error);
            finalMessage = localizedError.message;
            parameters = getLocalizedErrorParameters(localizedError.parameters);
        } else if (isString(message)) {
            finalMessage = message;
        } else {
            return '';
        }

        const translatedMessage = t(finalMessage, parameters);

        if (isObject(message) && isObject(message.error) && isString(message.error.path) && isObject(message.error.context)) {
            if (message.error.path.endsWith(PARSE_TRANSACTION_API_URL_PATH)) {
                return getLocalizedErrorForParseImportTransactionsAPI(translatedMessage, message.error.context.recordId, message.error.context.originalData);
            }
        }

        return translatedMessage;
    }

    return { translateError };
}
