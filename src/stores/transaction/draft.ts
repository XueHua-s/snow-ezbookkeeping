import { CategoryType } from '@/core/category.ts';
import { TransactionType } from '@/core/transaction.ts';
import { getFirstVisibleCategoryId } from '@/lib/category.ts';
import {
    isArray1SubsetOfArray2
} from '@/lib/common.ts';
import { parseDateTimeFromUnixTimeWithTimezoneOffset } from '@/lib/datetime.ts';
import {
    clearUserTransactionDraft,
    getUserTransactionDraft,
    updateUserTransactionDraft
} from '@/lib/userstate.ts';
import type { Transaction } from '@/models/transaction.ts';
import type { useTransactionState } from './state.ts';

interface Context {
    accountsStore: ReturnType<typeof useTransactionState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useTransactionState>['transactionCategoriesStore'];
    settingsStore: ReturnType<typeof useTransactionState>['settingsStore'];
    transactionDraft: ReturnType<typeof useTransactionState>['transactionDraft'];
    userStore: ReturnType<typeof useTransactionState>['userStore'];
}

// Owns draft for the transaction store.
export function useTransactionDraft(context: Context) {
    const { accountsStore, transactionCategoriesStore, settingsStore, transactionDraft, userStore } = context;

    function fillTransactionObject(transaction: Transaction): Transaction {
        if (!transaction) {
            return transaction;
        }

        const transactionTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transaction.utcOffset);
        transaction.setDisplayDate(transactionTime.getGregorianCalendarYearDashMonthDashDay(), transactionTime.getGregorianCalendarDay(), transactionTime.getWeekDay());

        if (transaction.sourceAccountId) {
            transaction.setSourceAccount(accountsStore.allAccountsMap[transaction.sourceAccountId]);
        }

        if (transaction.destinationAccountId) {
            transaction.setDestinationAccount(accountsStore.allAccountsMap[transaction.destinationAccountId]);
        }

        if (transaction.categoryId) {
            transaction.setCategory(transactionCategoriesStore.allTransactionCategoriesMap[transaction.categoryId]);
        }

        return transaction;
    }

    function initTransactionDraft(): void {
        if (settingsStore.appSettings.autoSaveTransactionDraft === 'enabled' || settingsStore.appSettings.autoSaveTransactionDraft === 'confirmation') {
            transactionDraft.value = getUserTransactionDraft();
        } else {
            transactionDraft.value = null;
        }
    }

    function isTransactionDraftModified(transaction?: Transaction, initAmount?: number, initCategoryId?: string, initAccountId?: string, initTagIds?: string, firstVisibleAccountId?: string): boolean {
        if (!transaction) {
            return false;
        }

        if (transaction.sourceAmount !== 0 && transaction.sourceAmount !== initAmount) {
            return true;
        }

        if (transaction.type === TransactionType.Transfer && transaction.destinationAmount !== 0) {
            return true;
        }

        if (transaction.sourceAccountId && transaction.sourceAccountId !== '0' && transaction.sourceAccountId !== userStore.currentUserDefaultAccountId && ((userStore.currentUserDefaultAccountId !== '' && userStore.currentUserDefaultAccountId !== '0') || transaction.sourceAccountId !== firstVisibleAccountId) && transaction.sourceAccountId !== initAccountId) {
            return true;
        }

        if (transaction.type === TransactionType.Transfer && transaction.destinationAccountId && transaction.destinationAccountId !== '0' && transaction.destinationAccountId !== userStore.currentUserDefaultAccountId && transaction.destinationAccountId !== initAccountId) {
            return true;
        }

        const allCategories = transactionCategoriesStore.allTransactionCategories;

        if (allCategories) {
            if (transaction.type === TransactionType.Expense) {
                const defaultCategoryId = getFirstVisibleCategoryId(allCategories[CategoryType.Expense]);

                if (transaction.expenseCategoryId && transaction.expenseCategoryId !== '0' && transaction.expenseCategoryId !== defaultCategoryId && transaction.expenseCategoryId !== initCategoryId) {
                    return true;
                }
            } else if (transaction.type === TransactionType.Income) {
                const defaultCategoryId = getFirstVisibleCategoryId(allCategories[CategoryType.Income]);

                if (transaction.incomeCategoryId && transaction.incomeCategoryId !== '0' && transaction.incomeCategoryId !== defaultCategoryId && transaction.incomeCategoryId !== initCategoryId) {
                    return true;
                }
            } else if (transaction.type === TransactionType.Transfer) {
                const defaultCategoryId = getFirstVisibleCategoryId(allCategories[CategoryType.Transfer]);

                if (transaction.transferCategoryId && transaction.transferCategoryId !== '0' && transaction.transferCategoryId !== defaultCategoryId && transaction.transferCategoryId !== initCategoryId) {
                    return true;
                }
            }
        }

        if (transaction.hideAmount) {
            return true;
        }

        if (transaction.tagIds && transaction.tagIds.length > 0) {
            return !initTagIds || !isArray1SubsetOfArray2(transaction.tagIds, initTagIds.split(','));
        }

        if (transaction.pictures && transaction.pictures.length > 0) {
            return true;
        }

        if (transaction.comment && transaction.comment.trim()) {
            return true;
        }

        return false;
    }

    function saveTransactionDraft(transaction?: Transaction, initAmount?: number, initCategoryId?: string, initAccountId?: string, initTagIds?: string, firstVisibleAccountId?: string): void {
        if (settingsStore.appSettings.autoSaveTransactionDraft !== 'enabled' && settingsStore.appSettings.autoSaveTransactionDraft !== 'confirmation') {
            clearTransactionDraft();
            return;
        }

        if (transaction) {
            if (!isTransactionDraftModified(transaction, initAmount, initCategoryId, initAccountId, initTagIds, firstVisibleAccountId)) {
                clearTransactionDraft();
                return;
            }

            transactionDraft.value = transaction.toTransactionDraft();
        }

        updateUserTransactionDraft(transactionDraft.value);
    }

    function clearTransactionDraft(): void {
        transactionDraft.value = null;
        clearUserTransactionDraft();
    }

    return { fillTransactionObject, initTransactionDraft, isTransactionDraftModified, saveTransactionDraft, clearTransactionDraft };
}
