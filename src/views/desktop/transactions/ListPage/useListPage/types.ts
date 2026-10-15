import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import ImportDialog from '@/views/desktop/transactions/import/ImportDialog.vue';
import AIImageRecognitionDialog from '@/views/desktop/transactions/list/dialogs/AIImageRecognitionDialog.vue';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';

export interface TransactionListProps {
    initPageType?: string;
    initDateType?: string,
    initMaxTime?: string,
    initMinTime?: string,
    initType?: string,
    initCategoryIds?: string,
    initAccountIds?: string,
    initTagFilter?: string,
    initAmountFilter?: string,
    initKeyword?: string,
    initMatchMode?: string
}

export type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;

export type SnackBarType = InstanceType<typeof SnackBar>;

export type EditDialogType = InstanceType<typeof EditDialog>;

export type AIImageRecognitionDialogType = InstanceType<typeof AIImageRecognitionDialog>;

export type ImportDialogType = InstanceType<typeof ImportDialog>;
