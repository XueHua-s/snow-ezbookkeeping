import { itemAndIndex, values } from '@/core/base.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { useTransactionCategoryState } from './state.ts';

interface Context {
    allTransactionCategories: ReturnType<typeof useTransactionCategoryState>['allTransactionCategories'];
    allTransactionCategoriesMap: ReturnType<typeof useTransactionCategoryState>['allTransactionCategoriesMap'];
    transactionCategoryListStateInvalid: ReturnType<typeof useTransactionCategoryState>['transactionCategoryListStateInvalid'];
    loadingPromise: ReturnType<typeof useTransactionCategoryState>['loadingPromise'];
}

// Owns mutation for the transactionCategory store.
export function useTransactionCategoryMutation(context: Context) {
    const { allTransactionCategories, allTransactionCategoriesMap, transactionCategoryListStateInvalid, loadingPromise } = context;

    function loadTransactionCategoryList(allCategories: Record<number, TransactionCategory[]>): void {
        allTransactionCategories.value = allCategories;
        allTransactionCategoriesMap.value = {};

        for (const categories of values(allCategories)) {
            for (const category of categories) {
                allTransactionCategoriesMap.value[category.id] = category;

                if (!category.subCategories) {
                    continue;
                }

                for (const subCategory of category.subCategories) {
                    allTransactionCategoriesMap.value[subCategory.id] = subCategory;
                }
            }
        }
    }

    function addCategoryToTransactionCategoryList(category: TransactionCategory): void {
        let categoryList: TransactionCategory[] | undefined = undefined;

        if (!category.parentId || category.parentId === '0') {
            categoryList = allTransactionCategories.value[category.type];
        } else if (allTransactionCategoriesMap.value[category.parentId]) {
            categoryList = allTransactionCategoriesMap.value[category.parentId]!.subCategories;
        }

        if (categoryList) {
            categoryList.push(category);
        }

        allTransactionCategoriesMap.value[category.id] = category;
    }

    function updateCategoryInTransactionCategoryList(currentCategory: TransactionCategory, oldCategory?: TransactionCategory): boolean {
        if (oldCategory && currentCategory.parentId !== oldCategory.parentId) {
            return false;
        }

        let categoryList: TransactionCategory[] | undefined = undefined;

        if (!currentCategory.parentId || currentCategory.parentId === '0') {
            categoryList = allTransactionCategories.value[currentCategory.type];
        } else if (allTransactionCategoriesMap.value[currentCategory.parentId]) {
            categoryList = allTransactionCategoriesMap.value[currentCategory.parentId]!.subCategories;
        }

        if (categoryList) {
            for (const [category, index] of itemAndIndex(categoryList)) {
                if (category.id === currentCategory.id) {
                    if (!currentCategory.parentId || currentCategory.parentId === '0') {
                        currentCategory.subCategories = category.subCategories;
                    }

                    categoryList.splice(index, 1, currentCategory);
                    break;
                }
            }
        }

        allTransactionCategoriesMap.value[currentCategory.id] = currentCategory;
        return true;
    }

    function updateCategoryDisplayOrderInCategoryList({ category, from, to }: { category: TransactionCategory, from: number, to: number }): void {
        let categoryList: TransactionCategory[] | undefined = undefined;

        if (!category.parentId || category.parentId === '0') {
            categoryList = allTransactionCategories.value[category.type];
        } else if (allTransactionCategoriesMap.value[category.parentId]) {
            categoryList = allTransactionCategoriesMap.value[category.parentId]!.subCategories;
        }

        if (categoryList) {
            categoryList.splice(to, 0, categoryList.splice(from, 1)[0] as TransactionCategory);
        }
    }

    function updateCategoryVisibilityInTransactionCategoryList({ category, hidden }: { category: TransactionCategory, hidden: boolean }): void {
        if (allTransactionCategoriesMap.value[category.id]) {
            allTransactionCategoriesMap.value[category.id]!.visible = !hidden;
        }
    }

    function removeCategoryFromTransactionCategoryList(currentCategory: TransactionCategory): void {
        let categoryList: TransactionCategory[] | undefined = undefined;

        if (!currentCategory.parentId || currentCategory.parentId === '0') {
            categoryList = allTransactionCategories.value[currentCategory.type];
        } else if (allTransactionCategoriesMap.value[currentCategory.parentId]) {
            categoryList = allTransactionCategoriesMap.value[currentCategory.parentId]!.subCategories;
        }

        if (categoryList) {
            for (const [category, index] of itemAndIndex(categoryList)) {
                if (category.id === currentCategory.id) {
                    categoryList.splice(index, 1);
                    break;
                }
            }
        }

        if (allTransactionCategoriesMap.value[currentCategory.id] && allTransactionCategoriesMap.value[currentCategory.id]!.subCategories) {
            const subCategoryList = allTransactionCategoriesMap.value[currentCategory.id]!.subCategories;

            if (subCategoryList) {
                for (const subCategory of subCategoryList) {
                    if (allTransactionCategoriesMap.value[subCategory.id]) {
                        delete allTransactionCategoriesMap.value[subCategory.id];
                    }
                }
            }
        }

        if (allTransactionCategoriesMap.value[currentCategory.id]) {
            delete allTransactionCategoriesMap.value[currentCategory.id];
        }
    }

    function updateTransactionCategoryListInvalidState(invalidState: boolean): void {
        transactionCategoryListStateInvalid.value = invalidState;
    }

    function resetTransactionCategories(): void {
        loadingPromise.value = null;
        allTransactionCategories.value = {};
        allTransactionCategoriesMap.value = {};
        transactionCategoryListStateInvalid.value = true;
    }

    return {
        loadTransactionCategoryList, addCategoryToTransactionCategoryList, updateCategoryInTransactionCategoryList,
        updateCategoryDisplayOrderInCategoryList, updateCategoryVisibilityInTransactionCategoryList,
        removeCategoryFromTransactionCategoryList, updateTransactionCategoryListInvalidState, resetTransactionCategories,
    };
}
