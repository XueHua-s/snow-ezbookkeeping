import { CategoryType } from '@/core/category.ts';
import { isEquals } from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services,{ type ApiResponsePromise } from '@/lib/services.ts';
import {
    type TransactionCategoryCreateBatchRequest,
    type TransactionCategoryInfoResponse,
    TransactionCategory
} from '@/models/transaction_category.ts';
import type { useTransactionCategoryMutation } from './mutation.ts';
import type { useTransactionCategoryState } from './state.ts';

interface Context {
    loadingPromise: ReturnType<typeof useTransactionCategoryState>['loadingPromise'];
    transactionCategoryListStateInvalid: ReturnType<typeof useTransactionCategoryState>['transactionCategoryListStateInvalid'];
    allTransactionCategories: ReturnType<typeof useTransactionCategoryState>['allTransactionCategories'];
    updateTransactionCategoryListInvalidState: ReturnType<typeof useTransactionCategoryMutation>['updateTransactionCategoryListInvalidState'];
    loadTransactionCategoryList: ReturnType<typeof useTransactionCategoryMutation>['loadTransactionCategoryList'];
    addCategoryToTransactionCategoryList: ReturnType<typeof useTransactionCategoryMutation>['addCategoryToTransactionCategoryList'];
    updateCategoryInTransactionCategoryList: ReturnType<typeof useTransactionCategoryMutation>['updateCategoryInTransactionCategoryList'];
    allTransactionCategoriesMap: ReturnType<typeof useTransactionCategoryState>['allTransactionCategoriesMap'];
}

// Owns loading for the transactionCategory store.
export function useTransactionCategoryLoading(context: Context) {
    const {
        loadingPromise, transactionCategoryListStateInvalid, allTransactionCategories, updateTransactionCategoryListInvalidState,
        loadTransactionCategoryList, addCategoryToTransactionCategoryList, updateCategoryInTransactionCategoryList,
        allTransactionCategoriesMap,
    } = context;

    function loadAllCategories({ force }: { force?: boolean }): Promise<Record<number, TransactionCategory[]>> {
        if (loadingPromise.value) {
            return loadingPromise.value;
        }

        if (!force && !transactionCategoryListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allTransactionCategories.value);
            });
        }

        const currentPromise = loadingPromise.value = new Promise((resolve, reject) => {
            services.getAllTransactionCategories().then(response => {
                if (!loadingPromise.value || loadingPromise.value !== currentPromise) {
                    logger.error('loadingPromise.value is invalid after retrieving category list');
                    reject({ message: 'An error occurred' });
                    return;
                }

                const data = response.data;

                if (!data || !data.success || !data.result) {
                    loadingPromise.value = null;
                    reject({ message: 'Unable to retrieve category list' });
                    return;
                }

                if (!data.result[CategoryType.Income]) {
                    data.result[CategoryType.Income] = [];
                }

                if (!data.result[CategoryType.Expense]) {
                    data.result[CategoryType.Expense] = [];
                }

                if (!data.result[CategoryType.Transfer]) {
                    data.result[CategoryType.Transfer] = [];
                }

                if (transactionCategoryListStateInvalid.value) {
                    updateTransactionCategoryListInvalidState(false);
                }

                const transactionCategories = TransactionCategory.ofMap(data.result);

                if (force && data.result && isEquals(allTransactionCategories.value, transactionCategories)) {
                    loadingPromise.value = null;
                    reject({ message: 'Category list is up to date', isUpToDate: true });
                    return;
                }

                loadTransactionCategoryList(transactionCategories);
                loadingPromise.value = null;

                resolve(transactionCategories);
            }).catch(error => {
                if (loadingPromise.value === currentPromise) {
                    loadingPromise.value = null;
                }

                if (force) {
                    logger.error('failed to force load category list', error);
                } else {
                    logger.error('failed to load category list', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve category list' });
                } else {
                    reject(error);
                }
            });
        });

        return loadingPromise.value;
    }

    function getCategory({ categoryId }: { categoryId: string }): Promise<TransactionCategory> {
        return new Promise((resolve, reject) => {
            services.getTransactionCategory({
                id: categoryId
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve category' });
                    return;
                }

                const transactionCategory = TransactionCategory.of(data.result);

                resolve(transactionCategory);
            }).catch(error => {
                logger.error('failed to load category info', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve category' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveCategory({ category, isEdit, clientSessionId }: { category: TransactionCategory, isEdit: boolean, clientSessionId: string }): Promise<TransactionCategory> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<TransactionCategoryInfoResponse>;

            if (!isEdit) {
                promise = services.addTransactionCategory(category.toCreateRequest(clientSessionId));
            } else {
                promise = services.modifyTransactionCategory(category.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!isEdit) {
                        reject({ message: 'Unable to add category' });
                    } else {
                        reject({ message: 'Unable to save category' });
                    }
                    return;
                }

                const transactionCategory = TransactionCategory.of(data.result);

                if (!isEdit) {
                    addCategoryToTransactionCategoryList(transactionCategory);
                } else {
                    const result = updateCategoryInTransactionCategoryList(transactionCategory, allTransactionCategoriesMap.value[category.id]);

                    if (!result && !transactionCategoryListStateInvalid.value) {
                        updateTransactionCategoryListInvalidState(true);
                    }
                }

                resolve(transactionCategory);
            }).catch(error => {
                logger.error('failed to save category', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!isEdit) {
                        reject({ message: 'Unable to add category' });
                    } else {
                        reject({ message: 'Unable to save category' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function addCategories(req: TransactionCategoryCreateBatchRequest): Promise<Record<number, TransactionCategory[]>> {
        return new Promise((resolve, reject) => {
            services.addTransactionCategoryBatch(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to add category' });
                    return;
                }

                if (!transactionCategoryListStateInvalid.value) {
                    updateTransactionCategoryListInvalidState(true);
                }

                const transactionCategories = TransactionCategory.ofMap(data.result);

                resolve(transactionCategories);
            }).catch(error => {
                logger.error('failed to add categories', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to add category' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function addPresetCategories(req: TransactionCategoryCreateBatchRequest): Promise<Record<number, TransactionCategory[]>> {
        return new Promise((resolve, reject) => {
            services.addTransactionCategoryBatch(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to add preset categories' });
                    return;
                }

                if (!transactionCategoryListStateInvalid.value) {
                    updateTransactionCategoryListInvalidState(true);
                }

                const transactionCategories = TransactionCategory.ofMap(data.result);

                resolve(transactionCategories);
            }).catch(error => {
                logger.error('failed to add preset categories', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to add preset categories' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { loadAllCategories, getCategory, saveCategory, addCategories, addPresetCategories };
}
