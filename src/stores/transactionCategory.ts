import { defineStore } from 'pinia';
import { useTransactionCategoryLoading } from './transactionCategory/loading.ts';
import { useTransactionCategoryManagement } from './transactionCategory/management.ts';
import { useTransactionCategoryMutation } from './transactionCategory/mutation.ts';
import { useTransactionCategoryState } from './transactionCategory/state.ts';





export const useTransactionCategoriesStore = defineStore('transactionCategories', () =>{
    const {
        loadingPromise, allTransactionCategories, allTransactionCategoriesMap, transactionCategoryListStateInvalid,
        allAvailablePrimaryCategoriesCount, allAvailableSecondaryCategoriesCount, hasVisibleExpenseCategories,
        hasVisibleIncomeCategories, hasVisibleTransferCategories,
    } = useTransactionCategoryState();
    const {
        loadTransactionCategoryList, addCategoryToTransactionCategoryList, updateCategoryInTransactionCategoryList,
        updateCategoryDisplayOrderInCategoryList, updateCategoryVisibilityInTransactionCategoryList,
        removeCategoryFromTransactionCategoryList, updateTransactionCategoryListInvalidState, resetTransactionCategories,
    } = useTransactionCategoryMutation({ allTransactionCategories, allTransactionCategoriesMap, transactionCategoryListStateInvalid, loadingPromise });
    const { loadAllCategories, getCategory, saveCategory, addCategories, addPresetCategories } = useTransactionCategoryLoading({
        loadingPromise, transactionCategoryListStateInvalid, allTransactionCategories, updateTransactionCategoryListInvalidState,
        loadTransactionCategoryList, addCategoryToTransactionCategoryList, updateCategoryInTransactionCategoryList,
        allTransactionCategoriesMap,
    });
    const { changeCategoryDisplayOrder, updateCategoryDisplayOrders, hideCategory, deleteCategory } = useTransactionCategoryManagement({
        allTransactionCategoriesMap, allTransactionCategories, transactionCategoryListStateInvalid,
        updateTransactionCategoryListInvalidState, updateCategoryDisplayOrderInCategoryList, loadAllCategories,
        updateCategoryVisibilityInTransactionCategoryList, removeCategoryFromTransactionCategoryList,
    });

    return {
        // states
        allTransactionCategories,
        allTransactionCategoriesMap,
        transactionCategoryListStateInvalid,
        // computed states
        allAvailablePrimaryCategoriesCount,
        allAvailableSecondaryCategoriesCount,
        hasVisibleExpenseCategories,
        hasVisibleIncomeCategories,
        hasVisibleTransferCategories,
        // functions
        updateTransactionCategoryListInvalidState,
        resetTransactionCategories,
        loadAllCategories,
        getCategory,
        saveCategory,
        addCategories,
        addPresetCategories,
        changeCategoryDisplayOrder,
        updateCategoryDisplayOrders,
        hideCategory,
        deleteCategory
    };
});
