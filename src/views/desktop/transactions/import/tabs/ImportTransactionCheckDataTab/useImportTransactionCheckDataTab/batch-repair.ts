import { type NameValue, reversed } from '@/core/base.ts';
import { TransactionType } from '@/core/transaction.ts';
import { applyImportTransactionReplaceRules } from '@/lib/rule.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import { type BatchCreateDialogDataType } from '@/views/desktop/transactions/import/dialogs/BatchCreateDialog.vue';
import { type BatchReplaceDialogDataType } from '@/views/desktop/transactions/import/dialogs/BatchReplaceDialog.vue';
import type { useImportTransactionCheckDataTabEditing } from './editing.ts';
import type { useImportTransactionCheckDataTabOriginalItems } from './original-items.ts';
import type { useImportTransactionCheckDataTabSelection } from './selection.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    isEditing: ReturnType<typeof useImportTransactionCheckDataTabState>['isEditing'];
    batchReplaceDialog: ReturnType<typeof useImportTransactionCheckDataTabState>['batchReplaceDialog'];
    props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>;
    allCategoriesMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allCategoriesMap'];
    allAccountsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allAccountsMap'];
    allTagsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allTagsMap'];
    updateTransactionData: ReturnType<typeof useImportTransactionCheckDataTabEditing>['updateTransactionData'];
    snackbar: ReturnType<typeof useImportTransactionCheckDataTabState>['snackbar'];
    formatNumberToLocalizedNumerals: ReturnType<typeof useImportTransactionCheckDataTabState>['formatNumberToLocalizedNumerals'];
    batchApplyRulesDialog: ReturnType<typeof useImportTransactionCheckDataTabState>['batchApplyRulesDialog'];
    allOriginalExpenseCategoryNames: ReturnType<typeof useImportTransactionCheckDataTabOriginalItems>['allOriginalExpenseCategoryNames'];
    allOriginalIncomeCategoryNames: ReturnType<typeof useImportTransactionCheckDataTabOriginalItems>['allOriginalIncomeCategoryNames'];
    allOriginalTransferCategoryNames: ReturnType<typeof useImportTransactionCheckDataTabOriginalItems>['allOriginalTransferCategoryNames'];
    allOriginalSourceAccountNames: ReturnType<typeof useImportTransactionCheckDataTabOriginalItems>['allOriginalSourceAccountNames'];
    allOriginalDestinationAccountNames: ReturnType<typeof useImportTransactionCheckDataTabOriginalItems>['allOriginalDestinationAccountNames'];
    allOriginalTransactionTagNames: ReturnType<typeof useImportTransactionCheckDataTabOriginalItems>['allOriginalTransactionTagNames'];
    selectedImportTransactionCount: ReturnType<typeof useImportTransactionCheckDataTabSelection>['selectedImportTransactionCount'];
    allSecondaryCategoriesMapByName: ReturnType<typeof useImportTransactionCheckDataTabState>['allSecondaryCategoriesMapByName'];
    allAccountsMapByName: ReturnType<typeof useImportTransactionCheckDataTabState>['allAccountsMapByName'];
    batchCreateDialog: ReturnType<typeof useImportTransactionCheckDataTabState>['batchCreateDialog'];
}

export function useImportTransactionCheckDataTabBatchRepair(context: Context) {
    const {
        isEditing, batchReplaceDialog, props, allCategoriesMap, allAccountsMap, allTagsMap, updateTransactionData, snackbar,
        formatNumberToLocalizedNumerals, batchApplyRulesDialog, allOriginalExpenseCategoryNames, allOriginalIncomeCategoryNames,
        allOriginalTransferCategoryNames, allOriginalSourceAccountNames, allOriginalDestinationAccountNames,
        allOriginalTransactionTagNames, selectedImportTransactionCount, allSecondaryCategoriesMapByName, allAccountsMapByName,
        batchCreateDialog,
    } = context;

    function showReplaceInvalidItemDialog(type: BatchReplaceDialogDataType, invalidItems: NameValue[]): void {
        if (isEditing.value) {
            return;
        }

        batchReplaceDialog.value?.open({
            mode: 'replaceInvalidItems',
            type: type,
            invalidItems: invalidItems
        }).then(result => {
            if (!result || (!result.sourceItem && result.sourceItem !== '')) {
                return;
            }

            if (type !== 'tag') {
                if (!result.targetItem) {
                    return;
                }
            }

            let updatedCount = 0;

            if (props.importTransactions) {
                for (const importTransaction of props.importTransactions) {
                    if (importTransaction.valid) {
                        continue;
                    }

                    let updated = false;

                    if (type === 'expenseCategory' || type === 'incomeCategory' || type === 'transferCategory') {
                        const categoryId = importTransaction.categoryId;
                        const originalCategoryName = importTransaction.originalCategoryName;

                        if (importTransaction.type !== TransactionType.ModifyBalance && originalCategoryName === result.sourceItem && (!categoryId || categoryId === '0' || !allCategoriesMap.value[categoryId])) {
                            if (type === 'expenseCategory' && importTransaction.type === TransactionType.Expense) {
                                importTransaction.categoryId = result.targetItem as string;
                                updated = true;
                            } else if (type === 'incomeCategory' && importTransaction.type === TransactionType.Income) {
                                importTransaction.categoryId = result.targetItem as string;
                                updated = true;
                            } else if (type === 'transferCategory' && importTransaction.type === TransactionType.Transfer) {
                                importTransaction.categoryId = result.targetItem as string;
                                updated = true;
                            }
                        }
                    } else if (type === 'account') {
                        const sourceAccountId = importTransaction.sourceAccountId;
                        const originalSourceAccountName = importTransaction.originalSourceAccountName;
                        const destinationAccountId = importTransaction.destinationAccountId;
                        const originalDestinationAccountName = importTransaction.originalDestinationAccountName;

                        if (originalSourceAccountName === result.sourceItem && (!sourceAccountId || sourceAccountId === '0' || !allAccountsMap.value[sourceAccountId])) {
                            importTransaction.sourceAccountId = result.targetItem as string;
                            updated = true;
                        }

                        if (importTransaction.type === TransactionType.Transfer && originalDestinationAccountName === result.sourceItem && (!destinationAccountId || destinationAccountId === '0' || !allAccountsMap.value[destinationAccountId])) {
                            importTransaction.destinationAccountId = result.targetItem as string;
                            updated = true;
                        }
                    } else if (type === 'tag' && importTransaction.tagIds) {
                        const removeIndex: number[] = [];

                        for (let tagIndex = 0; tagIndex < importTransaction.tagIds.length; tagIndex++) {
                            const tagId = importTransaction.tagIds[tagIndex] as string;
                            const originalTagName = importTransaction.originalTagNames ? (importTransaction.originalTagNames[tagIndex] ?? '') : '';

                            if (originalTagName === result.sourceItem && (!tagId || tagId === '0' || !allTagsMap.value[tagId])) {
                                if (result.targetItem) {
                                    importTransaction.tagIds[tagIndex] = result.targetItem;
                                    importTransaction.originalTagNames[tagIndex] = allTagsMap.value[result.targetItem]?.name || '';
                                } else {
                                    removeIndex.push(tagIndex);
                                }
                                updated = true;
                            }
                        }

                        for (const tagIndex of reversed(removeIndex)) {
                            importTransaction.tagIds.splice(tagIndex, 1);
                            importTransaction.originalTagNames.splice(tagIndex, 1);
                        }
                    }

                    if (updated) {
                        updatedCount++;
                        updateTransactionData(importTransaction);
                    }
                }
            }

            if (updatedCount > 0) {
                snackbar.value?.showMessage('format.misc.youHaveUpdatedTransactions', {
                    count: formatNumberToLocalizedNumerals(updatedCount)
                });
            }
        });
    }

    function showBatchApplyRulesDialog(): void {
        if (isEditing.value) {
            return;
        }

        batchApplyRulesDialog.value?.open({
            expenseCategoryNames: allOriginalExpenseCategoryNames.value,
            incomeCategoryNames: allOriginalIncomeCategoryNames.value,
            transferCategoryNames: allOriginalTransferCategoryNames.value,
            sourceAccountNames: allOriginalSourceAccountNames.value,
            destinationAccountNames: allOriginalDestinationAccountNames.value,
            tagNames: allOriginalTransactionTagNames.value,
            selectedTransactionCount: selectedImportTransactionCount.value
        }).then(result => {
            if (!result || !result.rules || !result.scope) {
                return;
            }

            const applyResult = applyImportTransactionReplaceRules(props.importTransactions || [], result.rules, result.scope, {
                allCategoriesMap: allCategoriesMap.value,
                allSecondaryCategoriesMapByName: allSecondaryCategoriesMapByName.value,
                allAccountsMap: allAccountsMap.value,
                allAccountsMapByName: allAccountsMapByName.value,
                allTagsMap: allTagsMap.value,
                updateTransaction: updateTransactionData
            });

            if (applyResult.updatedTransactionCount > 0) {
                snackbar.value?.showMessage('format.misc.youHaveUpdatedTransactions', {
                    count: formatNumberToLocalizedNumerals(applyResult.updatedTransactionCount)
                });
            }
        });
    }

    function showBatchCreateInvalidItemDialog(type: BatchCreateDialogDataType, invalidItems: NameValue[]): void {
        if (isEditing.value) {
            return;
        }

        batchCreateDialog.value?.open({
            type: type,
            invalidItems: invalidItems
        }).then(result => {
            if (!result || !result.sourceTargetMap) {
                return;
            }

            let updatedCount = 0;

            if (props.importTransactions) {
                const sourceTargetMap: Record<string, string> = result.sourceTargetMap;

                for (const importTransaction of props.importTransactions) {
                    if (importTransaction.valid) {
                        continue;
                    }

                    let updated = false;

                    if (type === 'expenseCategory' || type === 'incomeCategory' || type === 'transferCategory') {
                        const categoryId = importTransaction.categoryId;
                        const originalCategoryName = importTransaction.originalCategoryName;
                        const targetItem = sourceTargetMap[originalCategoryName];

                        if (importTransaction.type !== TransactionType.ModifyBalance && targetItem && (!categoryId || categoryId === '0' || !allCategoriesMap.value[categoryId])) {
                            if (type === 'expenseCategory' && importTransaction.type === TransactionType.Expense) {
                                importTransaction.categoryId = targetItem;
                                updated = true;
                            } else if (type === 'incomeCategory' && importTransaction.type === TransactionType.Income) {
                                importTransaction.categoryId = targetItem;
                                updated = true;
                            } else if (type === 'transferCategory' && importTransaction.type === TransactionType.Transfer) {
                                importTransaction.categoryId = targetItem;
                                updated = true;
                            }
                        }
                    } else if (type === 'tag' && importTransaction.tagIds) {
                        for (let tagIndex = 0; tagIndex < importTransaction.tagIds.length; tagIndex++) {
                            const tagId = importTransaction.tagIds[tagIndex] as string;
                            const originalTagName = importTransaction.originalTagNames ? (importTransaction.originalTagNames[tagIndex] ?? '') : '';
                            const targetItem = sourceTargetMap[originalTagName];

                            if (targetItem && (!tagId || tagId === '0' || !allTagsMap.value[tagId])) {
                                importTransaction.tagIds[tagIndex] = targetItem;
                                updated = true;
                            }
                        }
                    }

                    if (updated) {
                        updatedCount++;
                        updateTransactionData(importTransaction);
                    }
                }
            }

            if (updatedCount > 0) {
                snackbar.value?.showMessage('format.misc.youHaveUpdatedTransactions', {
                    count: formatNumberToLocalizedNumerals(updatedCount)
                });
            }
        });
    }

    return { showReplaceInvalidItemDialog, showBatchApplyRulesDialog, showBatchCreateInvalidItemDialog };
}
