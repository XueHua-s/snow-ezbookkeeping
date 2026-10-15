import { keys, reversed, values } from '@/core/base.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';


export function isSubCategoryIdAvailable(categories: TransactionCategory[], categoryId: string): boolean {
    if (!categories || !categories.length) {
        return false;
    }

    for (const primaryCategory of categories) {
        if (primaryCategory.hidden) {
            continue;
        }

        const subCategoryList = primaryCategory.subCategories;

        if (!subCategoryList) {
            continue;
        }

        for (const secondaryCategory of subCategoryList) {
            if (secondaryCategory.hidden) {
                continue;
            }

            if (secondaryCategory.id === categoryId) {
                return true;
            }
        }
    }

    return false;
}

export function getFirstVisibleCategoryId(categories?: TransactionCategory[]): string {
    if (!categories || !categories.length) {
        return '';
    }

    for (const primaryCategory of categories) {
        if (primaryCategory.hidden) {
            continue;
        }

        const subCategoryList = primaryCategory.subCategories;

        if (!subCategoryList) {
            continue;
        }

        for (const secondaryCategory of subCategoryList) {
            if (secondaryCategory.hidden) {
                continue;
            }

            return secondaryCategory.id;
        }
    }

    return '';
}

export function getFirstAvailableSubCategoryId(categories: TransactionCategory[], categoryId: string): string {
    if (!categories || !categories.length) {
        return '';
    }

    for (const primaryCategory of categories) {
        if (primaryCategory.hidden || primaryCategory.id !== categoryId) {
            continue;
        }

        const subCategoryList = primaryCategory.subCategories;

        if (!subCategoryList) {
            return '';
        }

        for (const secondaryCategory of subCategoryList) {
            if (secondaryCategory.hidden) {
                continue;
            }

            return secondaryCategory.id;
        }

        return '';
    }

    return '';
}

export function isNoAvailableCategory(categories: TransactionCategory[], showHidden: boolean): boolean {
    for (const category of categories) {
        if (showHidden || !category.hidden) {
            return false;
        }
    }

    return true;
}

export function getAvailableCategoryCount(categories: TransactionCategory[], showHidden: boolean): number {
    let count = 0;

    for (const category of categories) {
        if (showHidden || !category.hidden) {
            count++;
        }
    }

    return count;
}

export function getFirstShowingId(categories: TransactionCategory[], showHidden: boolean): string | null {
    for (const category of categories) {
        if (showHidden || !category.hidden) {
            return category.id;
        }
    }

    return null;
}

export function getLastShowingId(categories: TransactionCategory[], showHidden: boolean): string | null {
    for (const category of reversed(categories)) {
        if (showHidden || !category.hidden) {
            return category.id;
        }
    }

    return null;
}

export function selectAllSubCategories(filterCategoryIds: Record<string, boolean>, value: boolean, category?: TransactionCategory): void {
    if (!category || !category.subCategories || !category.subCategories.length) {
        return;
    }

    for (const subCategory of category.subCategories) {
        filterCategoryIds[subCategory.id] = value;
    }
}

export function selectAll(filterCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): void {
    for (const categoryId of keys(filterCategoryIds)) {
        const category = allTransactionCategoriesMap[categoryId];

        if (category) {
            filterCategoryIds[category.id] = false;
        }
    }
}

export function selectNone(filterCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): void {
    for (const categoryId of keys(filterCategoryIds)) {
        const category = allTransactionCategoriesMap[categoryId];

        if (category) {
            filterCategoryIds[category.id] = true;
        }
    }
}

export function selectInvert(filterCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): void {
    for (const categoryId of keys(filterCategoryIds)) {
        const category = allTransactionCategoriesMap[categoryId];

        if (category) {
            filterCategoryIds[category.id] = !filterCategoryIds[category.id];
        }
    }
}

export function isCategoryOrSubCategoriesAllChecked(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): boolean {
    if (!category.subCategories || category.subCategories.length < 1) {
        return !filterCategoryIds[category.id];
    }

    for (const subCategory of category.subCategories) {
        if (filterCategoryIds[subCategory.id]) {
            return false;
        }
    }

    return true;
}

export function isSubCategoriesAllChecked(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): boolean {
    if (!category.subCategories || category.subCategories.length < 1) {
        return false;
    }

    for (const subCategory of category.subCategories) {
        if (filterCategoryIds[subCategory.id]) {
            return false;
        }
    }

    return true;
}

export function isSubCategoriesHasButNotAllChecked(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): boolean {
    let checkedCount = 0;

    if (!category.subCategories || category.subCategories.length < 1) {
        return false;
    }

    for (const subCategory of category.subCategories) {
        if (!filterCategoryIds[subCategory.id]) {
            checkedCount++;
        }
    }

    return checkedCount > 0 && checkedCount < category.subCategories.length;
}

export function isAllCategoriesChecked(allCategories: Record<number, TransactionCategory[]>, includeCategoryIds: Record<string, boolean>): boolean {
    if (!allCategories) {
        return true;
    }

    for (const categories of values(allCategories)) {
        for (const category of categories) {
            if (!category.subCategories || category.subCategories.length < 1) {
                continue;
            }

            for (const subCategory of category.subCategories) {
                if (!includeCategoryIds[subCategory.id]) {
                    return false;
                }
            }
        }
    }

    return true;
}
