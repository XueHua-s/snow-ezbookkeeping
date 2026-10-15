import { keysIfValueEquals, values } from '@/core/base.ts';
import { type LocalizedPresetCategory, CategoryType } from '@/core/category.ts';
import { IconType } from '@/core/icon.ts';
import { NormalizedText } from '@/core/text.ts';
import { TransactionType } from '@/core/transaction.ts';
import type { TransactionCategoryCreateRequest, TransactionCategoryCreateWithSubCategories, TransactionCategory } from '@/models/transaction_category.ts';
import { isCategoryOrSubCategoriesAllChecked } from './selection.ts';

export function transactionTypeToCategoryType(transactionType: TransactionType): CategoryType | null {
    if (transactionType === TransactionType.Income) {
        return CategoryType.Income;
    } else if (transactionType === TransactionType.Expense) {
        return CategoryType.Expense;
    } else if (transactionType === TransactionType.Transfer) {
        return CategoryType.Transfer;
    } else {
        return null;
    }
}

export function categoryTypeToTransactionType(categoryType: CategoryType): TransactionType | null {
    if (categoryType === CategoryType.Income) {
        return TransactionType.Income;
    } else if (categoryType === CategoryType.Expense) {
        return TransactionType.Expense;
    } else if (categoryType === CategoryType.Transfer) {
        return TransactionType.Transfer;
    } else {
        return null;
    }
}

export function localizedPresetCategoryToTransactionCategoryCreateWithSubCategorys(presetCategory: LocalizedPresetCategory): TransactionCategoryCreateWithSubCategories {
    const subCategories: TransactionCategoryCreateRequest[] = [];

    for (const subPresetCategory of presetCategory.subCategories) {
        const subCategory: TransactionCategoryCreateRequest = {
            name: subPresetCategory.name,
            type: subPresetCategory.type,
            parentId: '0',
            icon: subPresetCategory.icon,
            iconType: IconType.System,
            color: subPresetCategory.color,
            comment: '',
            clientSessionId: ''
        };

        subCategories.push(subCategory);
    }

    const categoryWithSubCategories: TransactionCategoryCreateWithSubCategories = {
        name: presetCategory.name,
        type: presetCategory.type,
        icon: presetCategory.icon,
        iconType: IconType.System,
        color: presetCategory.color,
        subCategories: subCategories
    };

    return categoryWithSubCategories;
}

export function localizedPresetCategoriesToTransactionCategoryCreateWithSubCategories(presetCategories: LocalizedPresetCategory[]): TransactionCategoryCreateWithSubCategories[] {
    const categories: TransactionCategoryCreateWithSubCategories[] = [];

    for (const presetCategory of presetCategories) {
        const categoryWithSubCategories = localizedPresetCategoryToTransactionCategoryCreateWithSubCategorys(presetCategory);
        categories.push(categoryWithSubCategories);
    }

    return categories;
}

export function getSecondaryTransactionMapByName(allCategories?: TransactionCategory[]): Record<string, TransactionCategory> {
    const ret: Record<string, TransactionCategory> = {};

    if (!allCategories) {
        return ret;
    }

    for (const category of allCategories) {
        if (category.subCategories) {
            for (const subCategory of category.subCategories) {
                ret[subCategory.name] = subCategory;
            }
        }
    }

    return ret;
}

export function getTransactionPrimaryCategoryName(categoryId: string | null | undefined, allCategories?: TransactionCategory[]): string {
    if (!allCategories) {
        return '';
    }

    for (const category of allCategories) {
        const subCategoryList = category.subCategories;

        if (!subCategoryList) {
            continue;
        }

        for (const subCategory of subCategoryList) {
            if (subCategory.id === categoryId) {
                return category.name;
            }
        }
    }

    return '';
}

export function getTransactionSecondaryCategoryName(categoryId: string | null | undefined, allCategories?: TransactionCategory[]): string {
    if (!allCategories) {
        return '';
    }

    for (const category of allCategories) {
        const subCategoryList = category.subCategories;

        if (!subCategoryList) {
            continue;
        }

        for (const subCategory of subCategoryList) {
            if (subCategory.id === categoryId) {
                return subCategory.name;
            }
        }
    }

    return '';
}

export function filterTransactionCategories(allTransactionCategories: Record<number, TransactionCategory[]>, allowCategoryTypes?: Record<number, boolean>, allowCategoryName?: string, showHidden?: boolean): Record<string, TransactionCategory[]> {
    const ret: Record<string, TransactionCategory[]> = {};
    const hasAllowCategoryTypes = allowCategoryTypes
        && (allowCategoryTypes[CategoryType.Income]
            || allowCategoryTypes[CategoryType.Expense]
            || allowCategoryTypes[CategoryType.Transfer]);

    const allCategoryTypes = [ CategoryType.Income, CategoryType.Expense, CategoryType.Transfer ];
    const normalizedFilterContent = allowCategoryName ? NormalizedText.normalizeForSearch(allowCategoryName) : '';

    for (const categoryType of allCategoryTypes) {
        const allCategories = allTransactionCategories[categoryType];

        if (!allCategories || allCategories.length < 1) {
            continue;
        }

        if (hasAllowCategoryTypes && !allowCategoryTypes[categoryType]) {
            continue;
        }

        const allFilteredCategories: TransactionCategory[] = [];

        for (const category of allCategories) {
            if (!showHidden && category.hidden) {
                continue;
            }

            const categoryMatchesName = !normalizedFilterContent || NormalizedText.normalizeForSearch(category.name).includes(normalizedFilterContent);
            const filteredSubCategories: TransactionCategory[] = [];

            if (category.subCategories) {
                for (const subCategory of category.subCategories) {
                    if (!showHidden && subCategory.hidden) {
                        continue;
                    }

                    if (!categoryMatchesName && normalizedFilterContent && !NormalizedText.normalizeForSearch(subCategory.name).includes(normalizedFilterContent)) {
                        continue;
                    }

                    const filteredSubCategory = subCategory.clone();
                    filteredSubCategories.push(filteredSubCategory);
                }
            }

            if (!categoryMatchesName && filteredSubCategories.length < 1) {
                continue;
            }

            const filteredCategory = category.clone();
            filteredCategory.subCategories = filteredSubCategories;
            allFilteredCategories.push(filteredCategory);
        }

        ret[`${categoryType}`] = allFilteredCategories;
    }

    return ret;
}

export function allVisiblePrimaryTransactionCategoriesByType(allTransactionCategories: Record<number, TransactionCategory[]>, categoryType: number): TransactionCategory[] {
    const allCategories = allTransactionCategories[categoryType];
    const visibleCategories: TransactionCategory[] = [];

    if (!allCategories) {
        return visibleCategories;
    }

    for (const category of allCategories) {
        if (category.hidden) {
            continue;
        }

        visibleCategories.push(category);
    }

    return visibleCategories;
}

export function getFinalCategoryIdsByFilteredCategoryIds(allTransactionCategoriesMap: Record<number, TransactionCategory>, filteredCategoryIds: Record<string, boolean>): string {
    let finalCategoryIds = '';

    if (!allTransactionCategoriesMap) {
        return finalCategoryIds;
    }

    for (const category of values(allTransactionCategoriesMap)) {
        if (filteredCategoryIds && !isCategoryOrSubCategoriesAllChecked(category, filteredCategoryIds)) {
            continue;
        }

        if (finalCategoryIds.length > 0) {
            finalCategoryIds += ',';
        }

        finalCategoryIds += category.id;
    }

    return finalCategoryIds;
}

export function getIncludedTransactionCategoriesDisplayContent(excludeTransactionCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): string {
    if (!allTransactionCategoriesMap) {
        return '';
    }

    let hasExcludeTransactionCategory = false;

    for (const transactionCategoryId of keysIfValueEquals(excludeTransactionCategoryIds, true)) {
        if (allTransactionCategoriesMap[transactionCategoryId]) {
            hasExcludeTransactionCategory = true;
            break;
        }
    }

    if (!hasExcludeTransactionCategory) {
        return 'All';
    }

    let allTransactionCategoryExcluded = true;

    for (const transactionCategory of values(allTransactionCategoriesMap)) {
        if (transactionCategory.type !== CategoryType.Income && transactionCategory.type !== CategoryType.Expense) {
            continue;
        }

        if (!excludeTransactionCategoryIds[transactionCategory.id]) {
            allTransactionCategoryExcluded = false;
            break;
        }
    }

    if (allTransactionCategoryExcluded) {
        return 'None';
    }

    return 'Partial';
}
