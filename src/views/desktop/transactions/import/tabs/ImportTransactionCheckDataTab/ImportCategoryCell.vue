<script setup lang="ts">
import { CategoryType } from '@/core/category.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    getTransactionPrimaryCategoryName,
    getTransactionSecondaryCategoryName
} from '@/lib/category.ts';
import { getCategoryIconType } from '@/lib/icon.ts';
import { useI18n } from '@/locales/helpers.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import type { useImportTransactionCheckDataTab } from '@/views/desktop/transactions/import/tabs/ImportTransactionCheckDataTab/useImportTransactionCheckDataTab.ts';
import {
    mdiAlertOutline
} from '@mdi/js';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useImportTransactionCheckDataTab>>;

type ParentProps = {
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
};

defineProps<{
    item: ImportTransaction;
    editingTransaction: Scope['editingTransaction'];
    allCategoriesMap: Scope['allCategoriesMap'];
    disabled: ParentProps['disabled'];
    hasVisibleExpenseCategories: Scope['hasVisibleExpenseCategories'];
    allCategories: Scope['allCategories'];
    hasVisibleIncomeCategories: Scope['hasVisibleIncomeCategories'];
    hasVisibleTransferCategories: Scope['hasVisibleTransferCategories'];
}>();

const categoryId = defineModel<string>('categoryId', { required: true });

const { tt } = useI18n();

</script>

<template>
    <div class="d-flex align-center" v-if="editingTransaction !== item || item.type === TransactionType.ModifyBalance">
        <span v-if="item.type === TransactionType.ModifyBalance">-</span>
        <ItemIcon size="24px" :icon-type="getCategoryIconType(allCategoriesMap[item.categoryId]?.iconType)"
                  :icon-id="allCategoriesMap[item.categoryId]?.icon ?? ''"
                  :color="allCategoriesMap[item.categoryId]?.color ?? ''"
                  v-if="item.type !== TransactionType.ModifyBalance && item.categoryId && item.categoryId !== '0' && allCategoriesMap[item.categoryId]"></ItemIcon>
        <span class="ms-1" v-if="item.type !== TransactionType.ModifyBalance && item.categoryId && item.categoryId !== '0' && allCategoriesMap[item.categoryId]">
                            {{ allCategoriesMap[item.categoryId]?.name }}
                        </span>
        <div class="text-error font-italic" v-else-if="item.type !== TransactionType.ModifyBalance && (!item.categoryId || item.categoryId === '0' || !allCategoriesMap[item.categoryId])">
            <v-icon class="me-1" :icon="mdiAlertOutline"/>
            <span>{{ item.originalCategoryName }}</span>
        </div>
    </div><div style="width: 260px" v-if="editingTransaction === item && item.type === TransactionType.Expense">
        <two-column-select density="compact" variant="plain"
                           primary-key-field="id" primary-value-field="id" primary-title-field="name"
                           primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                           primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                           secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                           secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                           secondary-hidden-field="hidden"
                           :disabled="!!disabled || !hasVisibleExpenseCategories"
                           :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                           :show-selection-primary-text="true"
                           :custom-selection-primary-text="getTransactionPrimaryCategoryName(item.categoryId, allCategories[CategoryType.Expense])"
                           :custom-selection-secondary-text="getTransactionSecondaryCategoryName(item.categoryId, allCategories[CategoryType.Expense])"
                           :placeholder="tt('Category')"
                           :items="allCategories[CategoryType.Expense]"
                           v-model="categoryId">
        </two-column-select>
    </div><div style="width: 260px" v-if="editingTransaction === item && item.type === TransactionType.Income">
        <two-column-select density="compact" variant="plain"
                           primary-key-field="id" primary-value-field="id" primary-title-field="name"
                           primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                           primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                           secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                           secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                           secondary-hidden-field="hidden"
                           :disabled="!!disabled || !hasVisibleIncomeCategories"
                           :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                           :show-selection-primary-text="true"
                           :custom-selection-primary-text="getTransactionPrimaryCategoryName(item.categoryId, allCategories[CategoryType.Income])"
                           :custom-selection-secondary-text="getTransactionSecondaryCategoryName(item.categoryId, allCategories[CategoryType.Income])"
                           :placeholder="tt('Category')"
                           :items="allCategories[CategoryType.Income]"
                           v-model="categoryId">
        </two-column-select>
    </div><div style="width: 260px" v-if="editingTransaction === item && item.type === TransactionType.Transfer">
        <two-column-select density="compact" variant="plain"
                           primary-key-field="id" primary-value-field="id" primary-title-field="name"
                           primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                           primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                           secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                           secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                           secondary-hidden-field="hidden"
                           :disabled="!!disabled || !hasVisibleTransferCategories"
                           :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                           :show-selection-primary-text="true"
                           :custom-selection-primary-text="getTransactionPrimaryCategoryName(item.categoryId, allCategories[CategoryType.Transfer])"
                           :custom-selection-secondary-text="getTransactionSecondaryCategoryName(item.categoryId, allCategories[CategoryType.Transfer])"
                           :placeholder="tt('Category')"
                           :items="allCategories[CategoryType.Transfer]"
                           v-model="categoryId">
        </two-column-select>
    </div>
</template>
