import { KnownFileType } from '@/core/file.ts';
import {
    extendMdiSemicolon
} from '@/exticons/desktop/extend_mdi_icons.ts';
import {
    mdiComma,
    mdiKeyboardTab,
    mdiPlus,
    mdiTextBoxEditOutline
} from '@mdi/js';
import { computed } from 'vue';
import type { useImportTransactionCheckDataTabBatchRepair } from './batch-repair.ts';
import type { useImportTransactionCheckDataTabBatchReplace } from './batch-replace.ts';
import type { useImportTransactionCheckDataTabExport } from './export.ts';
import type { useImportTransactionCheckDataTabInvalidItems } from './invalid-items.ts';
import type { useImportTransactionCheckDataTabOriginalItems } from './original-items.ts';
import type { useImportTransactionCheckDataTabSelection } from './selection.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';
import { type ImportTransactionCheckDataMenu } from './types.ts';

interface Context {
    tt: ReturnType<typeof useImportTransactionCheckDataTabState>['tt'];
    isEditing: ReturnType<typeof useImportTransactionCheckDataTabState>['isEditing'];
    selectedExpenseTransactionCount: ReturnType<typeof useImportTransactionCheckDataTabSelection>['selectedExpenseTransactionCount'];
    showBatchReplaceDialog: ReturnType<typeof useImportTransactionCheckDataTabBatchReplace>['showBatchReplaceDialog'];
    selectedIncomeTransactionCount: ReturnType<typeof useImportTransactionCheckDataTabSelection>['selectedIncomeTransactionCount'];
    selectedTransferTransactionCount: ReturnType<typeof useImportTransactionCheckDataTabSelection>['selectedTransferTransactionCount'];
    selectedImportTransactionCount: ReturnType<typeof useImportTransactionCheckDataTabSelection>['selectedImportTransactionCount'];
    allOriginalTransactionTagNames: ReturnType<typeof useImportTransactionCheckDataTabOriginalItems>['allOriginalTransactionTagNames'];
    showBatchAddDialog: ReturnType<typeof useImportTransactionCheckDataTabBatchReplace>['showBatchAddDialog'];
    allInvalidExpenseCategoryNames: ReturnType<typeof useImportTransactionCheckDataTabInvalidItems>['allInvalidExpenseCategoryNames'];
    showReplaceInvalidItemDialog: ReturnType<typeof useImportTransactionCheckDataTabBatchRepair>['showReplaceInvalidItemDialog'];
    allInvalidIncomeCategoryNames: ReturnType<typeof useImportTransactionCheckDataTabInvalidItems>['allInvalidIncomeCategoryNames'];
    allInvalidTransferCategoryNames: ReturnType<typeof useImportTransactionCheckDataTabInvalidItems>['allInvalidTransferCategoryNames'];
    allInvalidAccountNames: ReturnType<typeof useImportTransactionCheckDataTabInvalidItems>['allInvalidAccountNames'];
    allInvalidTransactionTagNames: ReturnType<typeof useImportTransactionCheckDataTabInvalidItems>['allInvalidTransactionTagNames'];
    showBatchCreateInvalidItemDialog: ReturnType<typeof useImportTransactionCheckDataTabBatchRepair>['showBatchCreateInvalidItemDialog'];
    useCommaDecimalSeparator: ReturnType<typeof useImportTransactionCheckDataTabState>['useCommaDecimalSeparator'];
    exportData: ReturnType<typeof useImportTransactionCheckDataTabExport>['exportData'];
}

export function useImportTransactionCheckDataTabToolMenu(context: Context) {
    const {
        tt, isEditing, selectedExpenseTransactionCount, showBatchReplaceDialog, selectedIncomeTransactionCount,
        selectedTransferTransactionCount, selectedImportTransactionCount, allOriginalTransactionTagNames, showBatchAddDialog,
        allInvalidExpenseCategoryNames, showReplaceInvalidItemDialog, allInvalidIncomeCategoryNames,
        allInvalidTransferCategoryNames, allInvalidAccountNames, allInvalidTransactionTagNames, showBatchCreateInvalidItemDialog,
        useCommaDecimalSeparator, exportData,
    } = context;

    const toolMenus = computed<ImportTransactionCheckDataMenu[]>(() => [
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Batch Replace Selected Expense Categories'),
            disabled: isEditing.value || selectedExpenseTransactionCount.value < 1,
            onClick: () => showBatchReplaceDialog('expenseCategory')
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Batch Replace Selected Income Categories'),
            disabled: isEditing.value || selectedIncomeTransactionCount.value < 1,
            onClick: () => showBatchReplaceDialog('incomeCategory')
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Batch Replace Selected Transfer Categories'),
            disabled: isEditing.value || selectedTransferTransactionCount.value < 1,
            onClick: () => showBatchReplaceDialog('transferCategory')
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Batch Replace Selected Accounts'),
            disabled: isEditing.value || selectedImportTransactionCount.value < 1,
            onClick: () => showBatchReplaceDialog('account')
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Batch Replace Selected Destination Accounts'),
            disabled: isEditing.value || selectedTransferTransactionCount.value < 1,
            onClick: () => showBatchReplaceDialog('destinationAccount')
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Batch Replace Selected Transaction Timezones'),
            disabled: isEditing.value || selectedImportTransactionCount.value < 1,
            onClick: () => showBatchReplaceDialog('timezone')
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Batch Replace Selected Transaction Tags'),
            disabled: isEditing.value || selectedImportTransactionCount.value < 1,
            onClick: () => showBatchReplaceDialog('tag', allOriginalTransactionTagNames.value)
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Batch Add Transaction Tags'),
            disabled: isEditing.value || selectedImportTransactionCount.value < 1,
            onClick: () => showBatchAddDialog('tag')
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Replace Invalid Expense Categories'),
            disabled: isEditing.value || !allInvalidExpenseCategoryNames.value || allInvalidExpenseCategoryNames.value.length < 1,
            divider: true,
            onClick: () => showReplaceInvalidItemDialog('expenseCategory', allInvalidExpenseCategoryNames.value)
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Replace Invalid Income Categories'),
            disabled: isEditing.value || !allInvalidIncomeCategoryNames.value || allInvalidIncomeCategoryNames.value.length < 1,
            onClick: () => showReplaceInvalidItemDialog('incomeCategory', allInvalidIncomeCategoryNames.value)
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Replace Invalid Transfer Categories'),
            disabled: isEditing.value || !allInvalidTransferCategoryNames.value || allInvalidTransferCategoryNames.value.length < 1,
            onClick: () => showReplaceInvalidItemDialog('transferCategory', allInvalidTransferCategoryNames.value)
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Replace Invalid Accounts'),
            disabled: isEditing.value || !allInvalidAccountNames.value || allInvalidAccountNames.value.length < 1,
            onClick: () => showReplaceInvalidItemDialog('account', allInvalidAccountNames.value)
        },
        {
            prependIcon: mdiTextBoxEditOutline,
            title: tt('Replace Invalid Transaction Tags'),
            disabled: isEditing.value || !allInvalidTransactionTagNames.value || allInvalidTransactionTagNames.value.length < 1,
            onClick: () => showReplaceInvalidItemDialog('tag', allInvalidTransactionTagNames.value)
        },
        {
            prependIcon: mdiPlus,
            title: tt('Create Nonexistent Expense Categories'),
            disabled: isEditing.value || !allInvalidExpenseCategoryNames.value || allInvalidExpenseCategoryNames.value.length < 1,
            divider: true,
            onClick: () => showBatchCreateInvalidItemDialog('expenseCategory', allInvalidExpenseCategoryNames.value)
        },
        {
            prependIcon: mdiPlus,
            title: tt('Create Nonexistent Income Categories'),
            disabled: isEditing.value || !allInvalidIncomeCategoryNames.value || allInvalidIncomeCategoryNames.value.length < 1,
            onClick: () => showBatchCreateInvalidItemDialog('incomeCategory', allInvalidIncomeCategoryNames.value)
        },
        {
            prependIcon: mdiPlus,
            title: tt('Create Nonexistent Transfer Categories'),
            disabled: isEditing.value || !allInvalidTransferCategoryNames.value || allInvalidTransferCategoryNames.value.length < 1,
            onClick: () => showBatchCreateInvalidItemDialog('transferCategory', allInvalidTransferCategoryNames.value)
        },
        {
            prependIcon: mdiPlus,
            title: tt('Create Nonexistent Transaction Tags'),
            disabled: isEditing.value || !allInvalidTransactionTagNames.value || allInvalidTransactionTagNames.value.length < 1,
            onClick: () => showBatchCreateInvalidItemDialog('tag', allInvalidTransactionTagNames.value)
        },
        {
            prependIcon: mdiComma,
            title: tt('Save as CSV (Comma-separated values) File'),
            disabled: isEditing.value || selectedImportTransactionCount.value < 1 || useCommaDecimalSeparator.value,
            divider: true,
            onClick: () => exportData(KnownFileType.CSV)
        },
        {
            prependIcon: mdiKeyboardTab,
            title: tt('Save as TSV (Tab-separated values) File'),
            disabled: isEditing.value || selectedImportTransactionCount.value < 1,
            onClick: () => exportData(KnownFileType.TSV)
        },
        {
            prependIcon: extendMdiSemicolon,
            title: tt('Save as SSV (Semicolon-separated values) File'),
            disabled: isEditing.value || selectedImportTransactionCount.value < 1,
            onClick: () => exportData(KnownFileType.SSV)
        }
    ]);

    return { toolMenus };
}
