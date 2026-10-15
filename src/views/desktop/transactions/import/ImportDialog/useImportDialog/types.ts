import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import ImportTransactionCheckDataTab from '@/views/desktop/transactions/import/tabs/ImportTransactionCheckDataTab.vue';
import ImportTransactionDefineColumnTab from '@/views/desktop/transactions/import/tabs/ImportTransactionDefineColumnTab.vue';
import ImportTransactionExecuteCustomScriptTab from '@/views/desktop/transactions/import/tabs/ImportTransactionExecuteCustomScriptTab.vue';
import ImportTransactionRecognizeImagesTab from '@/views/desktop/transactions/import/tabs/ImportTransactionRecognizeImagesTab.vue';

export type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;

export type SnackBarType = InstanceType<typeof SnackBar>;

export type ImportTransactionDefineColumnTabType = InstanceType<typeof ImportTransactionDefineColumnTab>;

export type ImportTransactionExecuteCustomScriptTabType = InstanceType<typeof ImportTransactionExecuteCustomScriptTab>;

export type ImportTransactionRecognizeImagesTabType = InstanceType<typeof ImportTransactionRecognizeImagesTab>;

export type ImportTransactionCheckDataTabType = InstanceType<typeof ImportTransactionCheckDataTab>;

export type ImportTransactionDialogStep = 'uploadFile' | 'defineColumn' | 'executeCustomScript' | 'recognizeImages' | 'checkData' | 'finalResult';

export enum ImportCustomFileFormatProcessMethod {
    ColumnMapping,
    CustomScript
}
