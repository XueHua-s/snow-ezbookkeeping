import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import DataExportDialog from '@/components/desktop/DataExportDialog.vue';
import JsonExportDialog from '@/components/desktop/JsonExportDialog.vue';
import JsonImportDialog from '@/components/desktop/JsonImportDialog.vue';
import RenameDialog from '@/components/desktop/RenameDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import ExplorerChangeDisplayOrderDialog from '@/views/desktop/insights/dialogs/ExplorerChangeDisplayOrderDialog.vue';
import ExplorerChartTab from '@/views/desktop/insights/tabs/ExplorerChartTab.vue';
import ExplorerDataTableTab from '@/views/desktop/insights/tabs/ExplorerDataTableTab.vue';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';

export interface InsightsExplorerProps {
    initId?: string;
    initActiveTab?: string,
    initDateRangeType?: string,
    initStartTime?: string,
    initEndTime?: string,
}

export type ExplorerPageTabType = 'query' | 'table' | 'chart';

export type RenameDialogType = InstanceType<typeof RenameDialog>;

export type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;

export type SnackBarType = InstanceType<typeof SnackBar>;

export type DataExportDialogType = InstanceType<typeof DataExportDialog>;

export type JsonImportDialogType = InstanceType<typeof JsonImportDialog>;

export type JsonExportDialogType = InstanceType<typeof JsonExportDialog>;

export type ExplorerDataTableTabType = InstanceType<typeof ExplorerDataTableTab>;

export type ExplorerChartTabType = InstanceType<typeof ExplorerChartTab>;

export type ExplorerChangeDisplayOrderDialogType = InstanceType<typeof ExplorerChangeDisplayOrderDialog>;

export type EditDialogType = InstanceType<typeof EditDialog>;
