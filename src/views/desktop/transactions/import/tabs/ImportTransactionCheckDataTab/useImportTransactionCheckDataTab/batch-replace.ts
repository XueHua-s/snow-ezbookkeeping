import { type NameValue, reversed } from '@/core/base.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    getTimezoneOffsetMinutes
} from '@/lib/datetime.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import { type BatchReplaceDialogDataType } from '@/views/desktop/transactions/import/dialogs/BatchReplaceDialog.vue';
import type { useImportTransactionCheckDataTabEditing } from './editing.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    isEditing: ReturnType<typeof useImportTransactionCheckDataTabState>['isEditing'];
    batchReplaceDialog: ReturnType<typeof useImportTransactionCheckDataTabState>['batchReplaceDialog'];
    props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>;
    allTagsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allTagsMap'];
    updateTransactionData: ReturnType<typeof useImportTransactionCheckDataTabEditing>['updateTransactionData'];
    snackbar: ReturnType<typeof useImportTransactionCheckDataTabState>['snackbar'];
    formatNumberToLocalizedNumerals: ReturnType<typeof useImportTransactionCheckDataTabState>['formatNumberToLocalizedNumerals'];
}

export function useImportTransactionCheckDataTabBatchReplace(context: Context) {
    const { isEditing, batchReplaceDialog, props, allTagsMap, updateTransactionData, snackbar, formatNumberToLocalizedNumerals } = context;

    function showBatchReplaceDialog(type: BatchReplaceDialogDataType, allSourceTagItems?: NameValue[]): void {
        if (isEditing.value) {
            return;
        }

        batchReplaceDialog.value?.open({
            mode: 'batchReplace',
            type: type,
            allSourceTagItems: allSourceTagItems
        }).then(result => {
            if (!result) {
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
                    if (!importTransaction.selected) {
                        continue;
                    }

                    let updated = false;

                    if (type === 'expenseCategory') {
                        if (importTransaction.type === TransactionType.Expense) {
                            importTransaction.categoryId = result.targetItem as string;
                            updated = true;
                        }
                    } else if (type === 'incomeCategory') {
                        if (importTransaction.type === TransactionType.Income) {
                            importTransaction.categoryId = result.targetItem as string;
                            updated = true;
                        }
                    } else if (type === 'transferCategory') {
                        if (importTransaction.type === TransactionType.Transfer) {
                            importTransaction.categoryId = result.targetItem as string;
                            updated = true;
                        }
                    } else if (type === 'account') {
                        importTransaction.sourceAccountId = result.targetItem as string;
                        updated = true;
                    } else if (type === 'destinationAccount') {
                        if (importTransaction.type === TransactionType.Transfer) {
                            importTransaction.destinationAccountId = result.targetItem as string;
                            updated = true;
                        }
                    } else if (type === 'timezone') {
                        const oldUtcOffset = importTransaction.utcOffset;
                        importTransaction.utcOffset = getTimezoneOffsetMinutes(importTransaction.time, result.targetItem as string);
                        importTransaction.time = importTransaction.time - (importTransaction.utcOffset - oldUtcOffset) * 60;
                        updated = true;
                    } else if (type === 'tag') {
                        const removeIndex: number[] = [];

                        for (let tagIndex = 0; tagIndex < importTransaction.originalTagNames.length; tagIndex++) {
                            const originalTagName = importTransaction.originalTagNames ? (importTransaction.originalTagNames[tagIndex] ?? '') : '';

                            if (originalTagName === result.sourceItem) {
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

    function showBatchAddDialog(type: BatchReplaceDialogDataType): void {
        if (isEditing.value) {
            return;
        }

        batchReplaceDialog.value?.open({
            mode: 'batchAdd',
            type: type
        }).then(result => {
            if (!result || !result.targetItem) {
                return;
            }

            let updatedCount = 0;

            if (props.importTransactions) {
                for (const importTransaction of props.importTransactions) {
                    if (!importTransaction.selected) {
                        continue;
                    }

                    let updated = false;

                    if (type === 'tag') {
                        let containsTag = false;

                        for (const tagName of importTransaction.originalTagNames) {
                            if (tagName === result.targetItem) {
                                containsTag = true;
                                break;
                            }
                        }

                        if (!containsTag) {
                            if (!importTransaction.tagIds) {
                                importTransaction.tagIds = [];
                            }

                            if (!importTransaction.originalTagNames) {
                                importTransaction.originalTagNames = [];
                            }

                            importTransaction.tagIds.push(result.targetItem);
                            importTransaction.originalTagNames.push(allTagsMap.value[result.targetItem]?.name ?? '');
                            updated = true;
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

    return { showBatchReplaceDialog, showBatchAddDialog };
}
