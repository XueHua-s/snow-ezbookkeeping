import { itemAndIndex } from '@/core/base.ts';
import { AmountFilterType } from '@/core/numeral.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    isNumber,
    isString
} from '@/lib/common.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    filters: ReturnType<typeof useImportTransactionCheckDataTabState>['filters'];
    allTagsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allTagsMap'];
}

export function useImportTransactionCheckDataTabFiltering(context: Context) {
    const { filters, allTagsMap } = context;

    function isTransactionDisplayed(transaction: ImportTransaction): boolean {
        if (isNumber(filters.value.minDatetime) && isNumber(filters.value.maxDatetime) && (transaction.time < filters.value.minDatetime || transaction.time > filters.value.maxDatetime)) {
            return false;
        }

        if (isNumber(filters.value.transactionType) && transaction.type !== filters.value.transactionType) {
            return false;
        }

        if (isString(filters.value.category)) {
            if (filters.value.category === '' && transaction.actualCategoryName !== '') {
                return false;
            } else if (filters.value.category !== '' && transaction.actualCategoryName !== filters.value.category) {
                return false;
            }
        } else if (filters.value.category === undefined) {
            if (transaction.type !== TransactionType.ModifyBalance && transaction.categoryId && transaction.categoryId !== '0') {
                return false;
            }
        }

        if (isString(filters.value.amount)) {
            const match: boolean = AmountFilterType.match(filters.value.amount, transaction.sourceAmount);

            if (!match) {
                return false;
            }
        }

        if (isString(filters.value.account)) {
            if (filters.value.account === '' && (transaction.actualSourceAccountName !== '' || transaction.actualDestinationAccountName !== '')) {
                return false;
            } else if (filters.value.account !== '' && transaction.actualSourceAccountName !== filters.value.account && transaction.actualDestinationAccountName !== filters.value.account) {
                return false;
            }
        } else if (filters.value.account === undefined) {
            if (transaction.type !== TransactionType.Transfer && transaction.sourceAccountId && transaction.sourceAccountId !== '0') {
                return false;
            } else if (transaction.type === TransactionType.Transfer && transaction.sourceAccountId && transaction.sourceAccountId !== '0' && transaction.destinationAccountId && transaction.destinationAccountId !== '0') {
                return false;
            }
        }

        if (isString(filters.value.tag)) {
            if (filters.value.tag === '' && transaction.tagIds && transaction.tagIds.length) {
                return false;
            } else if (filters.value.tag !== '') {
                let hasTagName = false;

                if (transaction.tagIds && transaction.tagIds.length) {
                    for (const [tagId, tagIndex] of itemAndIndex(transaction.tagIds)) {
                        let tagName: string = transaction.originalTagNames ? (transaction.originalTagNames[tagIndex] ?? '') : '';

                        if (tagId && tagId !== '0' && allTagsMap.value[tagId] && allTagsMap.value[tagId].name) {
                            tagName = allTagsMap.value[tagId].name;
                        }

                        if (tagName === filters.value.tag) {
                            hasTagName = true;
                            break;
                        }
                    }
                }

                if (!hasTagName) {
                    return false;
                }
            }
        } else if (filters.value.tag === undefined) {
            if (transaction.tagIds && transaction.tagIds.length) {
                let hasInvalidTag = false;

                for (const tagId of transaction.tagIds) {
                    if (!tagId || tagId === '0') {
                        hasInvalidTag = true;
                        break;
                    }
                }

                if (!hasInvalidTag) {
                    return false;
                }
            } else {
                return false;
            }
        }

        if (isString(filters.value.description)) {
            if (filters.value.description === '' && transaction.comment !== '') {
                return false;
            } else if (filters.value.description !== '' && transaction.comment.indexOf(filters.value.description) < 0) {
                return false;
            }
        }

        return true;
    }

    function importTransactionsFilter(value: string, query: string, item?: { value: unknown, raw: ImportTransaction }): boolean {
        if (!item || !item.raw) {
            return false;
        }

        return isTransactionDisplayed(item.raw);
    }

    return { isTransactionDisplayed, importTransactionsFilter };
}
