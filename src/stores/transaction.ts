import { defineStore } from 'pinia';
import { useTransactionBatch } from './transaction/batch.ts';
import { useTransactionDailyAmounts } from './transaction/daily-amounts.ts';
import { useTransactionDelete } from './transaction/delete.ts';
import { useTransactionDraft } from './transaction/draft.ts';
import { useTransactionExchange } from './transaction/exchange.ts';
import { useTransactionFilter } from './transaction/filter.ts';
import { useTransactionImport } from './transaction/import.ts';
import { useTransactionInvalidation } from './transaction/invalidation.ts';
import { useTransactionList } from './transaction/list.ts';
import { useTransactionLoading } from './transaction/loading.ts';
import { useTransactionPictures } from './transaction/pictures.ts';
import { useTransactionRecognition } from './transaction/recognition.ts';
import { useTransactionReconciliation } from './transaction/reconciliation.ts';
import { useTransactionSave } from './transaction/save.ts';
import { useTransactionState } from './transaction/state.ts';
import { useTransactionTotals } from './transaction/totals.ts';
export * from './transaction/types.ts';




export const useTransactionsStore = defineStore('transactions', () => {
    const {
        settingsStore, userStore, accountsStore, transactionCategoriesStore, overviewStore, statisticsStore, explorersStore,
        exchangeRatesStore, transactionDraft, transactionsFilter, transactions, transactionsNextTimeId,
        transactionListStateInvalid, transactionReconciliationStatementStateInvalid, allFilterCategoryIds, allFilterAccountIds,
        allFilterTagIds, allFilterCategoryIdsCount, allFilterAccountIdsCount, allFilterTagIdsCount, noTransaction,
        hasMoreTransaction,
    } = useTransactionState();
    const { fillTransactionObject, initTransactionDraft, isTransactionDraftModified, saveTransactionDraft, clearTransactionDraft } = useTransactionDraft({ accountsStore, transactionCategoriesStore, settingsStore, transactionDraft, userStore });
    const { calculateMonthTotalAmount } = useTransactionTotals({ exchangeRatesStore });
    const { loadTransactionList, updateTransactionInTransactionList, removeTransactionFromTransactionList, collapseMonthInTransactionList } = useTransactionList({
        transactions, fillTransactionObject, calculateMonthTotalAmount, transactionsFilter, transactionsNextTimeId,
        transactionListStateInvalid, allFilterCategoryIds, allFilterAccountIds,
    });
    const { getCurrentMonthTransactionDailyTotalAmounts } = useTransactionDailyAmounts({ fillTransactionObject, calculateMonthTotalAmount, userStore });
    const { setTransactionSuitableDestinationAmount } = useTransactionExchange({ accountsStore, exchangeRatesStore });
    const { updateTransactionListInvalidState, updateTransactionReconciliationStatementInvalidState, updateStoreInvalidState, resetTransactions, clearTransactions } = useTransactionInvalidation({
        transactionListStateInvalid, transactionReconciliationStatementStateInvalid, accountsStore, overviewStore,
        statisticsStore, explorersStore, transactionsFilter, transactions, transactionsNextTimeId,
    });
    const { initTransactionListFilter, updateTransactionListFilter, getTransactionListPageParams, getExportTransactionDataRequestByTransactionFilter } = useTransactionFilter({ transactionsFilter, settingsStore, accountsStore });
    const { loadTransactions, loadMonthlyAllTransactions } = useTransactionLoading({ transactionsNextTimeId, transactionsFilter, loadTransactionList, transactionListStateInvalid, updateTransactionListInvalidState });
    const { getReconciliationStatements } = useTransactionReconciliation({ transactionReconciliationStatementStateInvalid, updateTransactionReconciliationStatementInvalidState });
    const { getTransaction, saveTransaction } = useTransactionSave({ transactionListStateInvalid, updateTransactionListInvalidState, updateTransactionInTransactionList, updateStoreInvalidState });
    const {
        batchUpdateTransactionCategories, batchUpdateTransactionAccounts, batchAddTagsToTransaction,
        batchRemoveTagsFromTransaction, batchClearAllTagsFromTransaction, moveAllTransactionsBetweenAccounts,
    } = useTransactionBatch({ updateStoreInvalidState });
    const { deleteTransaction, batchDeleteTransactions } = useTransactionDelete({ removeTransactionFromTransactionList, updateStoreInvalidState });
    const { recognizeTransactionText, recognizeReceiptImage, cancelRecognizeReceiptImage } = useTransactionRecognition();
    const { parseImportCustomFile, parseImportTransaction, importTransactions, getImportTransactionsProcess } = useTransactionImport();
    const { uploadTransactionPicture, removeUnusedTransactionPicture, getTransactionPictureUrl } = useTransactionPictures();

    return {
        // states
        transactionDraft,
        transactionsFilter,
        transactions,
        transactionsNextTimeId,
        transactionListStateInvalid,
        transactionReconciliationStatementStateInvalid,
        // computed states
        allFilterCategoryIds,
        allFilterAccountIds,
        allFilterTagIds,
        allFilterCategoryIdsCount,
        allFilterAccountIdsCount,
        allFilterTagIdsCount,
        noTransaction,
        hasMoreTransaction,
        // functions
        initTransactionDraft,
        isTransactionDraftModified,
        saveTransactionDraft,
        clearTransactionDraft,
        getCurrentMonthTransactionDailyTotalAmounts,
        setTransactionSuitableDestinationAmount,
        updateTransactionListInvalidState,
        updateTransactionReconciliationStatementInvalidState,
        resetTransactions,
        clearTransactions,
        initTransactionListFilter,
        updateTransactionListFilter,
        getTransactionListPageParams,
        getExportTransactionDataRequestByTransactionFilter,
        loadTransactions,
        loadMonthlyAllTransactions,
        getReconciliationStatements,
        getTransaction,
        saveTransaction,
        batchUpdateTransactionCategories,
        batchUpdateTransactionAccounts,
        batchAddTagsToTransaction,
        batchRemoveTagsFromTransaction,
        batchClearAllTagsFromTransaction,
        moveAllTransactionsBetweenAccounts,
        deleteTransaction,
        batchDeleteTransactions,
        recognizeTransactionText,
        recognizeReceiptImage,
        cancelRecognizeReceiptImage,
        parseImportCustomFile,
        parseImportTransaction,
        importTransactions,
        getImportTransactionsProcess,
        uploadTransactionPicture,
        removeUnusedTransactionPicture,
        getTransactionPictureUrl,
        collapseMonthInTransactionList
    };
});
