import SnackBar from '@/components/desktop/SnackBar.vue';
import { TransactionType } from '@/core/transaction.ts';
import BatchApplyRulesDialog from '@/views/desktop/transactions/import/dialogs/BatchApplyRulesDialog.vue';
import BatchCreateDialog from '@/views/desktop/transactions/import/dialogs/BatchCreateDialog.vue';
import BatchReplaceDialog from '@/views/desktop/transactions/import/dialogs/BatchReplaceDialog.vue';

export type SnackBarType = InstanceType<typeof SnackBar>;

export type BatchReplaceDialogType = InstanceType<typeof BatchReplaceDialog>;

export type BatchApplyRulesDialogType = InstanceType<typeof BatchApplyRulesDialog>;

export type BatchCreateDialogType = InstanceType<typeof BatchCreateDialog>;

export interface ImportTransactionCheckDataFilter {
    minDatetime: number | null; // minDatetime or maxDatetime is null for 'All Date Range', all are not null for 'Custom Date Range'
    maxDatetime: number | null;
    transactionType: TransactionType | null; // null for 'All Transaction Type'
    category: string | null | undefined; // null for 'All Category', undefined for 'Invalid Category'
    amount: string | null; // null for 'All Amount'
    account: string | null | undefined; // null for 'All Account', undefined for 'Invalid Account'
    tag: string | null | undefined; // null for 'All Tag', undefined for 'Invalid Tag'
    description: string | null; // null for 'All Description'
}

export interface ImportTransactionCheckDataMenuGroup {
    title?: string;
    items: ImportTransactionCheckDataMenu[];
}

export interface ImportTransactionCheckDataMenu {
    prependIcon?: string;
    title: string;
    subTitle?: string;
    appendIcon?: string;
    disabled?: boolean;
    divider?: boolean;
    onClick: () => void;
}
