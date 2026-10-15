import { values } from '@/core/base.ts';
import { CategoryType } from '@/core/category.ts';
import { getFirstVisibleCategoryId } from '@/lib/category.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { ComputedRef, Ref, ShallowRef, UnwrapRef } from 'vue';
import { computed, ref, shallowRef } from 'vue';
interface TransactionCategoryState {
    loadingPromise: ShallowRef<Promise<Record<number, TransactionCategory[]>> | null>;
    allTransactionCategories: Ref<UnwrapRef<Record<number, TransactionCategory[]>>>;
    allTransactionCategoriesMap: Ref<UnwrapRef<Record<string, TransactionCategory>>>;
    transactionCategoryListStateInvalid: Ref<UnwrapRef<boolean>>;
    allAvailablePrimaryCategoriesCount: ComputedRef<number>;
    allAvailableSecondaryCategoriesCount: ComputedRef<number>;
    hasVisibleExpenseCategories: ComputedRef<boolean>;
    hasVisibleIncomeCategories: ComputedRef<boolean>;
    hasVisibleTransferCategories: ComputedRef<boolean>;
}



// Owns state for the transactionCategory store.
export function useTransactionCategoryState(): TransactionCategoryState  {
    const loadingPromise = shallowRef<Promise<Record<number, TransactionCategory[]>> | null>(null);

    const allTransactionCategories = ref<Record<number, TransactionCategory[]>>({});

    const allTransactionCategoriesMap = ref<Record<string, TransactionCategory>>({});

    const transactionCategoryListStateInvalid = ref<boolean>(true);

    const allAvailablePrimaryCategoriesCount = computed<number>(() => {
        let count = 0;

        for (const categories of values(allTransactionCategories.value)) {
            count += categories.length;
        }

        return count;
    });

    const allAvailableSecondaryCategoriesCount = computed<number>(() => {
        let count = 0;

        for (const categories of values(allTransactionCategories.value)) {
            for (const category of categories) {
                if (category.subCategories) {
                    count += category.subCategories.length;
                }
            }
        }

        return count;
    });

    const hasVisibleExpenseCategories = computed<boolean>(() => {
        if (!allTransactionCategories.value || !allTransactionCategories.value[CategoryType.Expense] || !allTransactionCategories.value[CategoryType.Expense].length) {
            return false;
        }

        const firstVisibleCategoryId = getFirstVisibleCategoryId(allTransactionCategories.value[CategoryType.Expense]);
        return firstVisibleCategoryId !== '';
    });

    const hasVisibleIncomeCategories = computed<boolean>(() => {
        if (!allTransactionCategories.value || !allTransactionCategories.value[CategoryType.Income] || !allTransactionCategories.value[CategoryType.Income].length) {
            return false;
        }

        const firstVisibleCategoryId = getFirstVisibleCategoryId(allTransactionCategories.value[CategoryType.Income]);
        return firstVisibleCategoryId !== '';
    });

    const hasVisibleTransferCategories = computed<boolean>(() => {
        if (!allTransactionCategories.value || !allTransactionCategories.value[CategoryType.Transfer] || !allTransactionCategories.value[CategoryType.Transfer].length) {
            return false;
        }

        const firstVisibleCategoryId = getFirstVisibleCategoryId(allTransactionCategories.value[CategoryType.Transfer]);
        return firstVisibleCategoryId !== '';
    });

    return {
        loadingPromise, allTransactionCategories, allTransactionCategoriesMap, transactionCategoryListStateInvalid,
        allAvailablePrimaryCategoriesCount, allAvailableSecondaryCategoriesCount, hasVisibleExpenseCategories,
        hasVisibleIncomeCategories, hasVisibleTransferCategories,
    };
}
