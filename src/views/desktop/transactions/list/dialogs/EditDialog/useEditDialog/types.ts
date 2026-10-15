import MapView from '@/components/common/MapView.vue';
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { type SetTransactionOptions } from '@/lib/transaction.ts';
import { Transaction } from '@/models/transaction.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';

export interface TransactionEditOptions extends SetTransactionOptions {
    id?: string;
    templateType?: number;
    template?: TransactionTemplate;
    currentTransaction?: Transaction;
    currentTemplate?: TransactionTemplate;
    autoUploadPicture?: File;
    autoRecognizeClipboardText?: string;
    noTransactionDraft?: boolean;
}

export interface TransactionEditResponse {
    message: string;
}

export type MapViewType = InstanceType<typeof MapView>;

export type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;

export type SnackBarType = InstanceType<typeof SnackBar>;
