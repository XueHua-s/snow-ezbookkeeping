import type { BeforeResolveFunction } from '@/core/base.ts';
import { itemAndIndex } from '@/core/base.ts';
import type { CategoryType } from '@/core/category.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type { TransactionCategoryNewDisplayOrderRequest, TransactionCategory } from '@/models/transaction_category.ts';
import type { useTransactionCategoryLoading } from './loading.ts';
import type { useTransactionCategoryMutation } from './mutation.ts';
import type { useTransactionCategoryState } from './state.ts';

interface Context {
    allTransactionCategoriesMap: ReturnType<typeof useTransactionCategoryState>['allTransactionCategoriesMap'];
    allTransactionCategories: ReturnType<typeof useTransactionCategoryState>['allTransactionCategories'];
    transactionCategoryListStateInvalid: ReturnType<typeof useTransactionCategoryState>['transactionCategoryListStateInvalid'];
    updateTransactionCategoryListInvalidState: ReturnType<typeof useTransactionCategoryMutation>['updateTransactionCategoryListInvalidState'];
    updateCategoryDisplayOrderInCategoryList: ReturnType<typeof useTransactionCategoryMutation>['updateCategoryDisplayOrderInCategoryList'];
    loadAllCategories: ReturnType<typeof useTransactionCategoryLoading>['loadAllCategories'];
    updateCategoryVisibilityInTransactionCategoryList: ReturnType<typeof useTransactionCategoryMutation>['updateCategoryVisibilityInTransactionCategoryList'];
    removeCategoryFromTransactionCategoryList: ReturnType<typeof useTransactionCategoryMutation>['removeCategoryFromTransactionCategoryList'];
}

// Owns management for the transactionCategory store.
export function useTransactionCategoryManagement(context: Context) {
    const {
        allTransactionCategoriesMap, allTransactionCategories, transactionCategoryListStateInvalid,
        updateTransactionCategoryListInvalidState, updateCategoryDisplayOrderInCategoryList, loadAllCategories,
        updateCategoryVisibilityInTransactionCategoryList, removeCategoryFromTransactionCategoryList,
    } = context;

    function changeCategoryDisplayOrder({ categoryId, from, to }: { categoryId: string, from: number, to: number }): Promise<void> {
        const category = allTransactionCategoriesMap.value[categoryId];

        return new Promise((resolve, reject) => {
            if (!category) {
                reject({ message: 'Unable to move category' });
                return;
            }

            if (!category.parentId || category.parentId === '0') {
                if (!allTransactionCategories.value[category.type] ||
                    !allTransactionCategories.value[category.type]![to]) {
                    reject({ message: 'Unable to move category' });
                    return;
                }
            } else {
                const subCategoryList = allTransactionCategoriesMap.value[category.parentId]?.subCategories;

                if (!subCategoryList || !subCategoryList[to]) {
                    reject({ message: 'Unable to move category' });
                    return;
                }
            }

            if (!transactionCategoryListStateInvalid.value) {
                updateTransactionCategoryListInvalidState(true);
            }

            updateCategoryDisplayOrderInCategoryList({ category, from, to });

            resolve();
        });
    }

    function updateCategoryDisplayOrders({ type, parentId }: { type: CategoryType, parentId: string }): Promise<boolean> {
        const newDisplayOrders: TransactionCategoryNewDisplayOrderRequest[] = [];

        let categoryList: TransactionCategory[] | undefined = undefined;

        if (!parentId || parentId === '0') {
            categoryList = allTransactionCategories.value[type];
        } else if (allTransactionCategoriesMap.value[parentId]) {
            categoryList = allTransactionCategoriesMap.value[parentId].subCategories;
        }

        if (categoryList) {
            for (const [category, index] of itemAndIndex(categoryList)) {
                newDisplayOrders.push({
                    id: category.id,
                    displayOrder: index + 1
                });
            }
        }

        return new Promise((resolve, reject) => {
            services.moveTransactionCategory({
                newDisplayOrders: newDisplayOrders
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move category' });
                    return;
                }

                loadAllCategories({ force: false }).finally(() => {
                    if (transactionCategoryListStateInvalid.value) {
                        updateTransactionCategoryListInvalidState(false);
                    }

                    resolve(data.result);
                });
            }).catch(error => {
                logger.error('failed to save categories display order', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move category' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function hideCategory({ category, hidden }: { category: TransactionCategory, hidden: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.hideTransactionCategory({
                id: category.id,
                hidden: hidden
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this category' });
                    } else {
                        reject({ message: 'Unable to unhide this category' });
                    }

                    return;
                }

                updateCategoryVisibilityInTransactionCategoryList({ category, hidden });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to change category visibility', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this category' });
                    } else {
                        reject({ message: 'Unable to unhide this category' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteCategory({ category, beforeResolve }: { category: TransactionCategory, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteTransactionCategory({
                id: category.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this category' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeCategoryFromTransactionCategoryList(category);
                    });
                } else {
                    removeCategoryFromTransactionCategoryList(category);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete category', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this category' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { changeCategoryDisplayOrder, updateCategoryDisplayOrders, hideCategory, deleteCategory };
}
