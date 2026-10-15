import type { StepBarItem } from '@/components/desktop/StepsBar.vue';
import { type KeyAndName } from '@/core/base.ts';
import {
    type ImportFileTypeSupportedAdditionalOptions,
    type LocalizedImportFileCategoryAndTypes,
    type LocalizedImportFileType,
    type LocalizedImportFileTypeSubType,
    type LocalizedImportFileTypeSupportedEncodings
} from '@/core/file.ts';
import { isDefined } from '@/lib/common.ts';
import { findExtensionByType } from '@/lib/file.ts';
import { isTransactionFromAIImageRecognitionEnabled, isTransactionFromAITextRecognitionEnabled } from '@/lib/server_settings.ts';
import { useI18n } from '@/locales/helpers.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useOverviewStore } from '@/stores/overview.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useStatisticsStore } from '@/stores/statistics.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { type BatchImportImageItem } from '@/views/desktop/transactions/import/tabs/ImportTransactionRecognizeImagesTab.vue';
import { computed, ref, shallowRef, useTemplateRef } from 'vue';
import {
    type ConfirmDialogType, type ImportTransactionCheckDataTabType, type ImportTransactionDefineColumnTabType,
    type ImportTransactionDialogStep, type ImportTransactionExecuteCustomScriptTabType,
    type ImportTransactionRecognizeImagesTabType, type SnackBarType, ImportCustomFileFormatProcessMethod,
} from './types.ts';


export function useImportDialogState() {
    ;

    const {
        tt,
        te,
        joinMultiText,
        getAllSupportedImportFileCagtegoryAndTypes,
        formatNumberToLocalizedNumerals,
        getLocalizedFileEncodingName
    } = useI18n();

    const settingsStore = useSettingsStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const transactionsStore = useTransactionsStore();

    const overviewStore = useOverviewStore();

    const statisticsStore = useStatisticsStore();

    const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const importTransactionDefineColumnTab = useTemplateRef<ImportTransactionDefineColumnTabType>('importTransactionDefineColumnTab');

    const importTransactionExecuteCustomScriptTab = useTemplateRef<ImportTransactionExecuteCustomScriptTabType>('importTransactionExecuteCustomScriptTab');

    const importTransactionRecognizeImagesTab = useTemplateRef<ImportTransactionRecognizeImagesTabType>('importTransactionRecognizeImagesTab');

    const importTransactionCheckDataTab = useTemplateRef<ImportTransactionCheckDataTabType>('importTransactionCheckDataTab');

    const fileInput = useTemplateRef<HTMLInputElement>('fileInput');

    const allSupportedAdditionalOptions: KeyAndName[] = [
        {
            key: 'payeeAsTag',
            name: 'Parse Payee as Tag'
        },
        {
            key: 'payeeAsDescription',
            name: 'Parse Payee as Description'
        },
        {
            key: 'memberAsTag',
            name: 'Parse Member as Tag'
        },
        {
            key: 'projectAsTag',
            name: 'Parse Project as Tag'
        },
        {
            key: 'merchantAsTag',
            name: 'Parse Merchant as Tag'
        }
    ];

    const resolveFunc = shallowRef<(() => void) | null>(null);

    const rejectFunc = shallowRef<((reason?: unknown) => void) | null>(null);

    const showState = ref<boolean>(false);

    const additionalOptionsMenuState = ref<boolean>(false);

    const clientSessionId = ref<string>('');

    const currentStep = ref<ImportTransactionDialogStep>('uploadFile');

    const importProcess = ref<number>(0);

    const fileType = ref<string>('ezbookkeeping');

    const fileSubType = ref<string>('ezbookkeeping_csv');

    const fileEncoding = ref<string>('auto');

    const detectingFileEncoding = ref<boolean>(false);

    const autoDetectedFileEncoding = ref<string | undefined>(undefined);

    const processCustomFileFormatMethod = ref<ImportCustomFileFormatProcessMethod>(ImportCustomFileFormatProcessMethod.ColumnMapping);

    const importFile = ref<File | null>(null);

    const importImageFiles = ref<BatchImportImageItem[]>([]);

    const importData = ref<string>('');

    const importAdditionalOptions = ref<ImportFileTypeSupportedAdditionalOptions>({});

    const importAIAdditionalPrompt = ref<string>('');

    const importImageCancelRecognizingUuid = ref<string | undefined>(undefined);

    const parsedFileData = ref<string[][] | undefined>(undefined);

    const parsedFileColumnSeparator = ref<string | undefined>(undefined);

    const importTransactions = ref<ImportTransaction[] | undefined>(undefined);

    const importedCount = ref<number | null>(null);

    const loading = ref<boolean>(true);

    const submitting = ref<boolean>(false);

    const allSupportedImportFileCategoryAndTypes = computed<LocalizedImportFileCategoryAndTypes[]>(() => getAllSupportedImportFileCagtegoryAndTypes(isTransactionFromAITextRecognitionEnabled(), isTransactionFromAIImageRecognitionEnabled()));

    const allFileSubTypes = computed<LocalizedImportFileTypeSubType[] | undefined>(() => allSupportedImportFileTypesMap.value[fileType.value]?.subTypes);

    const allSupportedEncodings = computed<LocalizedImportFileTypeSupportedEncodings[] | undefined>(() => {
        const supportedEncodings = allSupportedImportFileTypesMap.value[fileType.value]?.supportedEncodings;

        if (!supportedEncodings) {
            return undefined;
        }

        const ret: LocalizedImportFileTypeSupportedEncodings[] = [];
        let autoDetectDisplayName = tt('Auto detect');

        if (importFile.value) {
            if (detectingFileEncoding.value) {
                autoDetectDisplayName += ` [${tt('Detecting...')}]`;
            } else if (autoDetectedFileEncoding.value) {
                autoDetectDisplayName += ` [${getLocalizedFileEncodingName(autoDetectedFileEncoding.value)}]`;
            } else {
                autoDetectDisplayName += ` [${tt('Unknown')}]`;
            }
        }

        const autoDetectEncoding: LocalizedImportFileTypeSupportedEncodings = {
            displayName: autoDetectDisplayName,
            encoding: 'auto'
        };

        ret.push(autoDetectEncoding);

        if (supportedEncodings && supportedEncodings.length) {
            ret.push(...supportedEncodings);
        }

        return ret;
    });

    const isCustomFileFormat = computed<boolean>(() => fileType.value === 'dsv' || fileType.value === 'dsv_data' || fileType.value === 'excel');

    const isImportDataFromTextbox = computed<boolean>(() => allSupportedImportFileTypesMap.value[fileType.value]?.dataFromTextbox ?? false);

    const needAITextRecognition = computed<boolean>(() => allSupportedImportFileTypesMap.value[fileType.value]?.needAITextRecognition ?? false);

    const isAIImageImport = computed<boolean>(() => fileType.value === 'ai_image');

    const needAIImageRecognition = computed<boolean>(() => allSupportedImportFileTypesMap.value[fileType.value]?.needAIImageRecognition ?? false);

    const supportedAdditionalOptions = computed<ImportFileTypeSupportedAdditionalOptions | undefined>(() => allSupportedImportFileTypesMap.value[fileType.value]?.supportedAdditionalOptions);

    const supportedAIAdditionalPrompt = computed<boolean>(() => !!allSupportedImportFileTypesMap.value[fileType.value]?.supportedAIAdditionalPrompt);

    const allSteps = computed<StepBarItem[]>(() => {
        const steps: StepBarItem[] = [
            {
                name: 'uploadFile',
                title: tt('Upload File'),
                subTitle: tt('Upload Transaction Data File')
            }
        ];

        if (isCustomFileFormat.value) {
            if (processCustomFileFormatMethod.value === ImportCustomFileFormatProcessMethod.CustomScript) {
                steps.push({
                    name: 'executeCustomScript',
                    title: tt('Execute Custom Script'),
                    subTitle: tt('Execute Custom Script to Parse Data')
                });
            } else {
                steps.push({
                    name: 'defineColumn',
                    title: tt('Define Column'),
                    subTitle: tt('Define and Check Column Mapping')
                });
            }
        }

        if (isAIImageImport.value) {
            steps.push({
                name: 'recognizeImages',
                title: tt('Recognize Images'),
                subTitle: tt('Recognize Transactions from Images')
            });
        }

        steps.push(...[
            {
                name: 'checkData',
                title: tt('Check & Modify'),
                subTitle: tt('Check and Modify Your Data')
            },
            {
                name: 'finalResult',
                title: tt('Complete'),
                subTitle: tt('Data Import Completed')
            }
        ]);

        return steps;
    });

    const allSupportedImportFileTypesMap = computed<Record<string, LocalizedImportFileType>>(() => {
        const ret: Record<string, LocalizedImportFileType> = {};

        for (const importFileCategoryAndTypes of allSupportedImportFileCategoryAndTypes.value) {
            for (const importFileType of importFileCategoryAndTypes.fileTypes) {
                ret[importFileType.type] = importFileType;
            }
        }

        return ret;
    });

    const supportedImportFileExtensions = computed<string | undefined>(() => {
        if (allFileSubTypes.value && allFileSubTypes.value.length) {
            const subTypeExtensions = findExtensionByType(allFileSubTypes.value, fileSubType.value);

            if (subTypeExtensions) {
                return subTypeExtensions;
            }
        }

        return allSupportedImportFileTypesMap.value[fileType.value]?.extensions;
    });

    const displaySelectedAdditionalOptions = computed<string>(() => {
        if (!supportedAdditionalOptions.value) {
            return tt('None');
        }

        const selectedOptions: string[] = [];

        for (const option of allSupportedAdditionalOptions) {
            if (isDefined(supportedAdditionalOptions.value[option.key]) && importAdditionalOptions.value[option.key]) {
                selectedOptions.push(tt(option.name));
            }
        }

        if (selectedOptions.length < 1) {
            return tt('None');
        }

        return joinMultiText(selectedOptions);
    });

    const exportFileGuideDocumentUrl = computed<string | undefined>(() => {
        const document = allSupportedImportFileTypesMap.value[fileType.value]?.document;

        if (!document) {
            return undefined;
        }

        const language = document.language ? document.language + '/' : '';
        const anchor = document.anchor ? '#' + document.anchor : '';
        return `https://ezbookkeeping.mayswind.net/${language}export_and_import${anchor}`;
    });

    const exportFileGuideDocumentLanguageName = computed<string | undefined>(() => allSupportedImportFileTypesMap.value[fileType.value]?.document?.displayLanguageName);

    const fileName = computed<string>(() => importFile.value?.name || '');

    return {
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
    };
}
