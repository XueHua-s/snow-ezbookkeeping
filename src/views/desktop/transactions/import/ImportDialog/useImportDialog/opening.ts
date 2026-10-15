import logger from '@/lib/logger.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import type { useImportDialogFiles } from './files.ts';
import type { useImportDialogState } from './state.ts';
import { ImportCustomFileFormatProcessMethod } from './types.ts';

interface Context {
    settingsStore: ReturnType<typeof useImportDialogState>['settingsStore'];
    allSupportedImportFileTypesMap: ReturnType<typeof useImportDialogState>['allSupportedImportFileTypesMap'];
    fileType: ReturnType<typeof useImportDialogState>['fileType'];
    fileSubType: ReturnType<typeof useImportDialogState>['fileSubType'];
    fileEncoding: ReturnType<typeof useImportDialogState>['fileEncoding'];
    detectingFileEncoding: ReturnType<typeof useImportDialogState>['detectingFileEncoding'];
    autoDetectedFileEncoding: ReturnType<typeof useImportDialogState>['autoDetectedFileEncoding'];
    processCustomFileFormatMethod: ReturnType<typeof useImportDialogState>['processCustomFileFormatMethod'];
    currentStep: ReturnType<typeof useImportDialogState>['currentStep'];
    importProcess: ReturnType<typeof useImportDialogState>['importProcess'];
    importFile: ReturnType<typeof useImportDialogState>['importFile'];
    importData: ReturnType<typeof useImportDialogState>['importData'];
    importAdditionalOptions: ReturnType<typeof useImportDialogState>['importAdditionalOptions'];
    supportedAdditionalOptions: ReturnType<typeof useImportDialogState>['supportedAdditionalOptions'];
    importAIAdditionalPrompt: ReturnType<typeof useImportDialogState>['importAIAdditionalPrompt'];
    parsedFileData: ReturnType<typeof useImportDialogState>['parsedFileData'];
    parsedFileColumnSeparator: ReturnType<typeof useImportDialogState>['parsedFileColumnSeparator'];
    importTransactionDefineColumnTab: ReturnType<typeof useImportDialogState>['importTransactionDefineColumnTab'];
    importTransactionExecuteCustomScriptTab: ReturnType<typeof useImportDialogState>['importTransactionExecuteCustomScriptTab'];
    importTransactions: ReturnType<typeof useImportDialogState>['importTransactions'];
    importTransactionCheckDataTab: ReturnType<typeof useImportDialogState>['importTransactionCheckDataTab'];
    showState: ReturnType<typeof useImportDialogState>['showState'];
    clientSessionId: ReturnType<typeof useImportDialogState>['clientSessionId'];
    clearImportImageFiles: ReturnType<typeof useImportDialogFiles>['clearImportImageFiles'];
    accountsStore: ReturnType<typeof useImportDialogState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useImportDialogState>['transactionCategoriesStore'];
    transactionTagsStore: ReturnType<typeof useImportDialogState>['transactionTagsStore'];
    loading: ReturnType<typeof useImportDialogState>['loading'];
    rejectFunc: ReturnType<typeof useImportDialogState>['rejectFunc'];
    resolveFunc: ReturnType<typeof useImportDialogState>['resolveFunc'];
    submitting: ReturnType<typeof useImportDialogState>['submitting'];
    fileInput: ReturnType<typeof useImportDialogState>['fileInput'];
}

export function useImportDialogOpening(context: Context) {
    const {
        settingsStore, allSupportedImportFileTypesMap, fileType, fileSubType, fileEncoding, detectingFileEncoding,
        autoDetectedFileEncoding, processCustomFileFormatMethod, currentStep, importProcess, importFile, importData,
        importAdditionalOptions, supportedAdditionalOptions, importAIAdditionalPrompt, parsedFileData, parsedFileColumnSeparator,
        importTransactionDefineColumnTab, importTransactionExecuteCustomScriptTab, importTransactions,
        importTransactionCheckDataTab, showState, clientSessionId, clearImportImageFiles, accountsStore,
        transactionCategoriesStore, transactionTagsStore, loading, rejectFunc, resolveFunc, submitting, fileInput,
    } = context;

    function loadInitFileTypeFromSettings(): void {
        if (!settingsStore.appSettings.lastSelectedFileTypeInImportTransactionDialog) {
            return;
        }

        const lastSelectedFileTypes = settingsStore.appSettings.lastSelectedFileTypeInImportTransactionDialog.split('|');
        const lastSelectedFileType = lastSelectedFileTypes[0];

        if (!lastSelectedFileType || !allSupportedImportFileTypesMap.value[lastSelectedFileType]) {
            return;
        }

        fileType.value = lastSelectedFileType;

        const fileSubTypes = allSupportedImportFileTypesMap.value[lastSelectedFileType].subTypes;

        if (!fileSubTypes || fileSubTypes.length < 1) {
            return;
        }

        const lastSelectedFileSubType = lastSelectedFileTypes[1];

        if (lastSelectedFileSubType) {
            for (const subType of fileSubTypes) {
                if (subType.type === lastSelectedFileSubType) {
                    fileSubType.value = lastSelectedFileSubType;
                    return;
                }
            }
        }

        const firstFileSubType = fileSubTypes[0];

        if (firstFileSubType) {
            fileSubType.value = firstFileSubType.type;
        }
    }

    function open(): Promise<void> {
        fileType.value = 'ezbookkeeping';
        fileSubType.value = 'ezbookkeeping_csv';

        if (settingsStore.appSettings.rememberLastSelectedFileTypeInImportTransactionDialog && settingsStore.appSettings.lastSelectedFileTypeInImportTransactionDialog) {
            loadInitFileTypeFromSettings();
        }

        fileEncoding.value = 'auto';
        detectingFileEncoding.value = false;
        autoDetectedFileEncoding.value = undefined;
        processCustomFileFormatMethod.value = ImportCustomFileFormatProcessMethod.ColumnMapping;
        currentStep.value = 'uploadFile';
        importProcess.value = 0;
        importFile.value = null;
        importData.value = '';
        importAdditionalOptions.value = Object.assign({}, supportedAdditionalOptions.value ?? {});
        importAIAdditionalPrompt.value = '';
        parsedFileData.value = undefined;
        parsedFileColumnSeparator.value = undefined;
        importTransactionDefineColumnTab.value?.reset();
        importTransactionExecuteCustomScriptTab.value?.reset();
        importTransactions.value = undefined;
        importTransactionCheckDataTab.value?.reset();
        showState.value = true;
        clientSessionId.value = generateRandomUUID();
        clearImportImageFiles();

        const promises = [
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false }),
            transactionTagsStore.loadAllTags({ force: false })
        ];

        Promise.all(promises).then(() => {
            loading.value = false;
        }).catch(error => {
            logger.error('failed to load essential data for importing transaction', error);

            loading.value = false;
            showState.value = false;

            if (!error.processed) {
                if (rejectFunc.value) {
                    rejectFunc.value(error);
                }
            }
        });

        return new Promise((resolve, reject) => {
            resolveFunc.value = resolve;
            rejectFunc.value = reject;
        });
    }

    function showOpenFileDialog(): void {
        if (submitting.value) {
            return;
        }

        fileInput.value?.click();
    }

    return { open, showOpenFileDialog };
}
