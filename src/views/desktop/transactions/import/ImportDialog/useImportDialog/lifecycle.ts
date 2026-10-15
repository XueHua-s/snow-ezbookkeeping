import { findExtensionByType, isFileExtensionSupported } from '@/lib/file.ts';
import { watch } from 'vue';
import type { useImportDialogFiles } from './files.ts';
import type { useImportDialogState } from './state.ts';

interface Context {
    resolveFunc: ReturnType<typeof useImportDialogState>['resolveFunc'];
    rejectFunc: ReturnType<typeof useImportDialogState>['rejectFunc'];
    showState: ReturnType<typeof useImportDialogState>['showState'];
    fileType: ReturnType<typeof useImportDialogState>['fileType'];
    allSupportedImportFileTypesMap: ReturnType<typeof useImportDialogState>['allSupportedImportFileTypesMap'];
    fileSubType: ReturnType<typeof useImportDialogState>['fileSubType'];
    settingsStore: ReturnType<typeof useImportDialogState>['settingsStore'];
    loading: ReturnType<typeof useImportDialogState>['loading'];
    importFile: ReturnType<typeof useImportDialogState>['importFile'];
    parsedFileData: ReturnType<typeof useImportDialogState>['parsedFileData'];
    parsedFileColumnSeparator: ReturnType<typeof useImportDialogState>['parsedFileColumnSeparator'];
    importAdditionalOptions: ReturnType<typeof useImportDialogState>['importAdditionalOptions'];
    supportedAdditionalOptions: ReturnType<typeof useImportDialogState>['supportedAdditionalOptions'];
    importTransactions: ReturnType<typeof useImportDialogState>['importTransactions'];
    clearImportImageFiles: ReturnType<typeof useImportDialogFiles>['clearImportImageFiles'];
    allFileSubTypes: ReturnType<typeof useImportDialogState>['allFileSubTypes'];
}

export function useImportDialogLifecycle(context: Context) {
    const {
        resolveFunc, rejectFunc, showState, fileType, allSupportedImportFileTypesMap, fileSubType, settingsStore, loading,
        importFile, parsedFileData, parsedFileColumnSeparator, importAdditionalOptions, supportedAdditionalOptions,
        importTransactions, clearImportImageFiles, allFileSubTypes,
    } = context;

    function close(completed: boolean): void {
        if (completed) {
            if (resolveFunc.value) {
                resolveFunc.value();
            }
        } else {
            if (rejectFunc.value) {
                rejectFunc.value();
            }
        }

        showState.value = false;
    }

    watch(fileType, (newValue) => {
        const subFileTypes = allSupportedImportFileTypesMap.value[newValue]?.subTypes;

        if (subFileTypes && subFileTypes.length) {
            if (fileSubType.value !== subFileTypes[0]!.type) {
                fileSubType.value = subFileTypes[0]!.type;
            } else if (settingsStore.appSettings.rememberLastSelectedFileTypeInImportTransactionDialog && !loading.value) {
                settingsStore.setLastSelectedFileTypeInImportTransactionDialog(`${newValue}|${fileSubType.value}`);
            }
        } else {
            if (settingsStore.appSettings.rememberLastSelectedFileTypeInImportTransactionDialog && !loading.value) {
                settingsStore.setLastSelectedFileTypeInImportTransactionDialog(`${newValue}|`);
            }
        }

        importFile.value = null;
        parsedFileData.value = undefined;
        parsedFileColumnSeparator.value = undefined;
        importAdditionalOptions.value = Object.assign({}, supportedAdditionalOptions.value ?? {});
        importTransactions.value = undefined;
        clearImportImageFiles();
    });

    watch(fileSubType, (newValue) => {
        if (settingsStore.appSettings.rememberLastSelectedFileTypeInImportTransactionDialog && !loading.value) {
            settingsStore.setLastSelectedFileTypeInImportTransactionDialog(`${fileType.value}|${newValue}`);
        }

        let supportedExtensions: string | undefined = findExtensionByType(allFileSubTypes.value, newValue);

        if (!supportedExtensions) {
            supportedExtensions = allSupportedImportFileTypesMap.value[fileType.value]?.extensions;
        }

        if (importFile.value && importFile.value.name && !isFileExtensionSupported(importFile.value.name, supportedExtensions || '')) {
            importFile.value = null;
        }
    });

    return { close };
}
