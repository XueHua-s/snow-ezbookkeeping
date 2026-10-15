import { UTF_8 } from '@/consts/file.ts';
import { itemAndIndex } from '@/core/base.ts';
import {
    KnownFileType
} from '@/core/file.ts';
import { TransactionType } from '@/core/transaction.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import type { useImportDialogRecognition } from './recognition.ts';
import type { useImportDialogState } from './state.ts';
import { ImportCustomFileFormatProcessMethod } from './types.ts';

interface Context {
    fileType: ReturnType<typeof useImportDialogState>['fileType'];
    allFileSubTypes: ReturnType<typeof useImportDialogState>['allFileSubTypes'];
    fileSubType: ReturnType<typeof useImportDialogState>['fileSubType'];
    isAIImageImport: ReturnType<typeof useImportDialogState>['isAIImageImport'];
    currentStep: ReturnType<typeof useImportDialogState>['currentStep'];
    importImageFiles: ReturnType<typeof useImportDialogState>['importImageFiles'];
    snackbar: ReturnType<typeof useImportDialogState>['snackbar'];
    submitting: ReturnType<typeof useImportDialogState>['submitting'];
    importTransactions: ReturnType<typeof useImportDialogState>['importTransactions'];
    batchRecognizeImages: ReturnType<typeof useImportDialogRecognition>['batchRecognizeImages'];
    importTransactionCheckDataTab: ReturnType<typeof useImportDialogState>['importTransactionCheckDataTab'];
    allSupportedEncodings: ReturnType<typeof useImportDialogState>['allSupportedEncodings'];
    fileEncoding: ReturnType<typeof useImportDialogState>['fileEncoding'];
    autoDetectedFileEncoding: ReturnType<typeof useImportDialogState>['autoDetectedFileEncoding'];
    isImportDataFromTextbox: ReturnType<typeof useImportDialogState>['isImportDataFromTextbox'];
    importFile: ReturnType<typeof useImportDialogState>['importFile'];
    importData: ReturnType<typeof useImportDialogState>['importData'];
    isCustomFileFormat: ReturnType<typeof useImportDialogState>['isCustomFileFormat'];
    transactionsStore: ReturnType<typeof useImportDialogState>['transactionsStore'];
    processCustomFileFormatMethod: ReturnType<typeof useImportDialogState>['processCustomFileFormatMethod'];
    importTransactionExecuteCustomScriptTab: ReturnType<typeof useImportDialogState>['importTransactionExecuteCustomScriptTab'];
    parsedFileData: ReturnType<typeof useImportDialogState>['parsedFileData'];
    parsedFileColumnSeparator: ReturnType<typeof useImportDialogState>['parsedFileColumnSeparator'];
    importTransactionDefineColumnTab: ReturnType<typeof useImportDialogState>['importTransactionDefineColumnTab'];
    importAdditionalOptions: ReturnType<typeof useImportDialogState>['importAdditionalOptions'];
    supportedAIAdditionalPrompt: ReturnType<typeof useImportDialogState>['supportedAIAdditionalPrompt'];
    importAIAdditionalPrompt: ReturnType<typeof useImportDialogState>['importAIAdditionalPrompt'];
}

export function useImportDialogParsing(context: Context) {
    const {
        fileType, allFileSubTypes, fileSubType, isAIImageImport, currentStep, importImageFiles, snackbar, submitting,
        importTransactions, batchRecognizeImages, importTransactionCheckDataTab, allSupportedEncodings, fileEncoding,
        autoDetectedFileEncoding, isImportDataFromTextbox, importFile, importData, isCustomFileFormat, transactionsStore,
        processCustomFileFormatMethod, importTransactionExecuteCustomScriptTab, parsedFileData, parsedFileColumnSeparator,
        importTransactionDefineColumnTab, importAdditionalOptions, supportedAIAdditionalPrompt, importAIAdditionalPrompt,
    } = context;

    function parseData(): void {
        let uploadFile: File;
        let type: string = fileType.value;
        let encoding: string | undefined = undefined;

        if (allFileSubTypes.value) {
            type = fileSubType.value;
        }

        if (isAIImageImport.value && currentStep.value === 'uploadFile') {
            if (importImageFiles.value.length < 1) {
                snackbar.value?.showError('Please select a file to import');
                return;
            }

            currentStep.value = 'recognizeImages';
            submitting.value = true;
            importTransactions.value = undefined;

            batchRecognizeImages().then(() => {
                if (!importTransactions.value || importTransactions.value.length < 1) {
                    if (submitting.value) {
                        currentStep.value = 'uploadFile';
                        snackbar.value?.showMessage('No data to import');
                    }
                    submitting.value = false;
                    return;
                }

                importTransactionCheckDataTab.value?.reset();

                if (importTransactions.value.length > 0 && importTransactions.value.length < 10) {
                    importTransactionCheckDataTab.value?.setCountPerPage(-1);
                } else {
                    importTransactionCheckDataTab.value?.setCountPerPage(10);
                }

                const anyFailed = importImageFiles.value.some(item => item.status === 'failed');

                if (submitting.value && !anyFailed) {
                    currentStep.value = 'checkData';
                }

                submitting.value = false;
            }).catch(() => {
                submitting.value = false;
            });

            return;
        } else if (isAIImageImport.value && currentStep.value === 'recognizeImages') {
            if (!importTransactions.value || importTransactions.value.length < 1) {
                snackbar.value?.showError('No data to import');
                return;
            }

            currentStep.value = 'checkData';
            return;
        }

        if (allSupportedEncodings.value) {
            if (fileEncoding.value === 'auto') {
                encoding = autoDetectedFileEncoding.value;
            } else {
                encoding = fileEncoding.value;
            }
        }

        if (!isImportDataFromTextbox.value) {
            if (!importFile.value) {
                snackbar.value?.showError('Please select a file to import');
                return;
            }

            if (allSupportedEncodings.value) {
                if (fileEncoding.value === 'auto' && !autoDetectedFileEncoding.value) {
                    snackbar.value?.showError('Unable to detect the file encoding automatically. Please select the actual encoding.');
                    return;
                }
            }

            uploadFile = importFile.value;
        } else if (isImportDataFromTextbox.value) {
            if (!importData.value) {
                snackbar.value?.showError('No data to import');
                return;
            }

            if (type === 'custom_csv') {
                uploadFile = KnownFileType.CSV.createFile(importData.value, 'import');
            } else if (type === 'custom_tsv') {
                uploadFile = KnownFileType.TSV.createFile(importData.value, 'import');
            } else if (type === 'custom_ssv') {
                uploadFile = KnownFileType.TXT.createFile(importData.value, 'import');
            } else if (type === 'ai_txt') {
                uploadFile = KnownFileType.TXT.createFile(importData.value, 'import');
            } else {
                snackbar.value?.showError('Parameter Invalid');
                return;
            }

            encoding = UTF_8;
        } else { // should not happen, but ts would check whether uploadFile has been assigned a value
            snackbar.value?.showMessage('An error occurred');
            return;
        }

        if (isCustomFileFormat.value && currentStep.value === 'uploadFile') {
            submitting.value = true;
            let columnSeparator: string | undefined = undefined;

            if (type === 'custom_csv') {
                columnSeparator = ',';
            } else if (type === 'custom_tsv') {
                columnSeparator = '\t';
            } else if (type === 'custom_ssv') {
                columnSeparator = ';';
            }

            transactionsStore.parseImportCustomFile({
                fileType: type,
                fileEncoding: encoding,
                importFile: uploadFile
            }).then(response => {
                if (response && response.length) {
                    if (processCustomFileFormatMethod.value === ImportCustomFileFormatProcessMethod.CustomScript) {
                        importTransactionExecuteCustomScriptTab.value?.reset();
                        parsedFileData.value = response;
                        parsedFileColumnSeparator.value = columnSeparator;
                        currentStep.value = 'executeCustomScript';
                    } else {
                        importTransactionDefineColumnTab.value?.reset();
                        parsedFileData.value = response;
                        parsedFileColumnSeparator.value = columnSeparator;
                        currentStep.value = 'defineColumn';
                    }
                } else {
                    parsedFileData.value = undefined;
                    parsedFileColumnSeparator.value = undefined;
                    snackbar.value?.showError('No data to import');
                }

                submitting.value = false;
            }).catch(error => {
                submitting.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        } else {
            let columnMapping: Record<number, number> | undefined = undefined;
            let transactionTypeMapping: Record<string, TransactionType> | undefined = undefined;
            let hasHeaderLine: boolean | undefined = undefined;
            let timeFormat: string | undefined = undefined;
            let timezoneFormat: string | undefined = undefined;
            let amountDecimalSeparator: string | undefined = undefined;
            let amountDigitGroupingSymbol: string | undefined = undefined;
            let geoLocationSeparator: string | undefined = undefined;
            let geoLocationOrder: string | undefined = undefined;
            let tagSeparator: string | undefined = undefined;

            if (isCustomFileFormat.value && processCustomFileFormatMethod.value === ImportCustomFileFormatProcessMethod.ColumnMapping) {
                const defineColumnResult = importTransactionDefineColumnTab.value?.generateResult();

                if (!defineColumnResult) {
                    return;
                }

                columnMapping = defineColumnResult.columnMapping;
                transactionTypeMapping = defineColumnResult.transactionTypeMapping;
                hasHeaderLine = defineColumnResult.includeHeader;
                timeFormat = defineColumnResult.timeFormat;
                timezoneFormat = defineColumnResult.timezoneFormat;
                amountDecimalSeparator = defineColumnResult.amountDecimalSeparator;
                amountDigitGroupingSymbol = defineColumnResult.amountDigitGroupingSymbol;
                geoLocationSeparator = defineColumnResult.geoLocationSeparator;
                geoLocationOrder = defineColumnResult.geoLocationOrder;
                tagSeparator = defineColumnResult.tagSeparator;
            } else if (isCustomFileFormat.value && processCustomFileFormatMethod.value === ImportCustomFileFormatProcessMethod.CustomScript) {
                const executeCustomScriptResult = importTransactionExecuteCustomScriptTab.value?.generateResult();

                if (!executeCustomScriptResult) {
                    return;
                }

                type = 'ezbookkeeping_json';
                encoding = undefined;
                uploadFile = KnownFileType.JSON.createFile(executeCustomScriptResult, 'import');
            }

            submitting.value = true;

            transactionsStore.parseImportTransaction({
                fileType: type,
                additionalOptions: importAdditionalOptions.value,
                aiAdditionalPrompt: supportedAIAdditionalPrompt.value ? importAIAdditionalPrompt.value : undefined,
                fileEncoding: encoding,
                importFile: uploadFile,
                columnMapping: columnMapping,
                transactionTypeMapping: transactionTypeMapping,
                hasHeaderLine: hasHeaderLine,
                timeFormat: timeFormat,
                timezoneFormat: timezoneFormat,
                amountDecimalSeparator: amountDecimalSeparator,
                amountDigitGroupingSymbol: amountDigitGroupingSymbol,
                geoSeparator: geoLocationSeparator,
                geoOrder: geoLocationOrder,
                tagSeparator: tagSeparator
            }).then(response => {
                const parsedTransactions: ImportTransaction[] = [];

                if (response.items) {
                    for (const [importTransaction, index] of itemAndIndex(response.items)) {
                        const parsedTransaction = ImportTransaction.of(importTransaction, index);
                        parsedTransactions.push(parsedTransaction);
                    }
                }

                importTransactionCheckDataTab.value?.reset();

                if (parsedTransactions && parsedTransactions.length >= 0 && parsedTransactions.length < 10) {
                    importTransactionCheckDataTab.value?.setCountPerPage(-1);
                } else {
                    importTransactionCheckDataTab.value?.setCountPerPage(10);
                }

                importTransactions.value = parsedTransactions;
                currentStep.value = 'checkData';
                submitting.value = false;
            }).catch(error => {
                submitting.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        }
    }

    return { parseData };
}
