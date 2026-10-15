import { useImportDialogFiles } from './useImportDialog/files.ts';
import { useImportDialogLifecycle } from './useImportDialog/lifecycle.ts';
import { useImportDialogOpening } from './useImportDialog/opening.ts';
import { useImportDialogParsing } from './useImportDialog/parsing.ts';
import { useImportDialogRecognition } from './useImportDialog/recognition.ts';
import { useImportDialogState } from './useImportDialog/state.ts';
import { useImportDialogSubmission } from './useImportDialog/submission.ts';
export * from './useImportDialog/types.ts';

// Coordinates ImportDialog state and effects; called once from the component setup.
export function useImportDialog() {
    const {
        tt, te, formatNumberToLocalizedNumerals, settingsStore, accountsStore, transactionCategoriesStore, transactionTagsStore,
        transactionsStore, overviewStore, statisticsStore, confirmDialog, snackbar, importTransactionDefineColumnTab,
        importTransactionExecuteCustomScriptTab, importTransactionRecognizeImagesTab, importTransactionCheckDataTab, fileInput,
        allSupportedAdditionalOptions, resolveFunc, rejectFunc, showState, additionalOptionsMenuState, clientSessionId,
        currentStep, importProcess, fileType, fileSubType, fileEncoding, detectingFileEncoding, autoDetectedFileEncoding,
        processCustomFileFormatMethod, importFile, importImageFiles, importData, importAdditionalOptions,
        importAIAdditionalPrompt, importImageCancelRecognizingUuid, parsedFileData, parsedFileColumnSeparator, importTransactions,
        importedCount, loading, submitting, allSupportedImportFileCategoryAndTypes, allFileSubTypes, allSupportedEncodings,
        isCustomFileFormat, isImportDataFromTextbox, needAITextRecognition, isAIImageImport, needAIImageRecognition,
        supportedAdditionalOptions, supportedAIAdditionalPrompt, allSteps, allSupportedImportFileTypesMap,
        supportedImportFileExtensions, displaySelectedAdditionalOptions, exportFileGuideDocumentUrl,
        exportFileGuideDocumentLanguageName, fileName,
    } = useImportDialogState();
    const { setImportFile, removeImportImageFile, clearImportImageFiles, reloadBasisData } = useImportDialogFiles({
        isAIImageImport, importImageFiles, importFile, detectingFileEncoding, autoDetectedFileEncoding, allSupportedEncodings,
        loading, accountsStore, transactionCategoriesStore, transactionTagsStore, snackbar, importTransactionCheckDataTab,
    });
    const { open, showOpenFileDialog } = useImportDialogOpening({
        settingsStore, allSupportedImportFileTypesMap, fileType, fileSubType, fileEncoding, detectingFileEncoding,
        autoDetectedFileEncoding, processCustomFileFormatMethod, currentStep, importProcess, importFile, importData,
        importAdditionalOptions, supportedAdditionalOptions, importAIAdditionalPrompt, parsedFileData, parsedFileColumnSeparator,
        importTransactionDefineColumnTab, importTransactionExecuteCustomScriptTab, importTransactions,
        importTransactionCheckDataTab, showState, clientSessionId, clearImportImageFiles, accountsStore,
        transactionCategoriesStore, transactionTagsStore, loading, rejectFunc, resolveFunc, submitting, fileInput,
    });
    const { recognizeImage, batchRecognizeImages, cancelBatchRecognizeImages } = useImportDialogRecognition({
        importImageCancelRecognizingUuid, transactionsStore, importImageFiles, snackbar, submitting, supportedAIAdditionalPrompt,
        importAIAdditionalPrompt, importTransactions, te,
    });
    const { parseData } = useImportDialogParsing({
        fileType, allFileSubTypes, fileSubType, isAIImageImport, currentStep, importImageFiles, snackbar, submitting,
        importTransactions, batchRecognizeImages, importTransactionCheckDataTab, allSupportedEncodings, fileEncoding,
        autoDetectedFileEncoding, isImportDataFromTextbox, importFile, importData, isCustomFileFormat, transactionsStore,
        processCustomFileFormatMethod, importTransactionExecuteCustomScriptTab, parsedFileData, parsedFileColumnSeparator,
        importTransactionDefineColumnTab, importAdditionalOptions, supportedAIAdditionalPrompt, importAIAdditionalPrompt,
    });
    const { batchApplyRules, submit } = useImportDialogSubmission({
        importTransactionCheckDataTab, importTransactions, snackbar, confirmDialog, formatNumberToLocalizedNumerals, submitting,
        transactionsStore, clientSessionId, importProcess, importedCount, currentStep, accountsStore, overviewStore,
        statisticsStore,
    });
    const { close } = useImportDialogLifecycle({
        resolveFunc, rejectFunc, showState, fileType, allSupportedImportFileTypesMap, fileSubType, settingsStore, loading,
        importFile, parsedFileData, parsedFileColumnSeparator, importAdditionalOptions, supportedAdditionalOptions,
        importTransactions, clearImportImageFiles, allFileSubTypes,
    });

    return {
        tt, te, formatNumberToLocalizedNumerals, confirmDialog, snackbar, importTransactionDefineColumnTab,
        importTransactionExecuteCustomScriptTab, importTransactionRecognizeImagesTab, importTransactionCheckDataTab, fileInput,
        allSupportedAdditionalOptions, showState, additionalOptionsMenuState, currentStep, importProcess, fileType, fileSubType,
        fileEncoding, autoDetectedFileEncoding, processCustomFileFormatMethod, importFile, importImageFiles, importData,
        importAdditionalOptions, importAIAdditionalPrompt, parsedFileData, parsedFileColumnSeparator, importTransactions,
        importedCount, loading, submitting, allSupportedImportFileCategoryAndTypes, allFileSubTypes, allSupportedEncodings,
        isCustomFileFormat, isImportDataFromTextbox, needAITextRecognition, isAIImageImport, needAIImageRecognition,
        supportedAdditionalOptions, supportedAIAdditionalPrompt, allSteps, supportedImportFileExtensions,
        displaySelectedAdditionalOptions, exportFileGuideDocumentUrl, exportFileGuideDocumentLanguageName, fileName, open,
        showOpenFileDialog, setImportFile, removeImportImageFile, reloadBasisData, recognizeImage, cancelBatchRecognizeImages,
        parseData, batchApplyRules, submit, close,
    };
}
