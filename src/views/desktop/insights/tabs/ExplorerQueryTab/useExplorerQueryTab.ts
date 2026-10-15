import type { SetupProps } from '@/lib/ui/props.ts';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { AccountType } from '@/core/account.ts';
import { type NameValue, type TypeAndDisplayName, values } from '@/core/base.ts';
import {
    TransactionExplorerConditionField,
    TransactionExplorerConditionRelation
} from '@/core/explorer.ts';
import {
    isObjectEmpty
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import { useI18n } from '@/locales/helpers.ts';
import {
    type TransactionExplorerCondition,
    TransactionExplorerQuery
} from '@/models/explorer.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useExplorersStore } from '@/stores/explorer.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useUserStore } from '@/stores/user.ts';
import { computed, ref, useTemplateRef, watch } from 'vue';

export interface ExplorerQueryTabProps {
    loading?: boolean;
    disabled?: boolean;
}

export type SnackBarType = InstanceType<typeof SnackBar>;

// Coordinates ExplorerQueryTab state and effects; called once from the component setup.
export function useExplorerQueryTab(props: SetupProps<ExplorerQueryTabProps>) {
    const {
        tt,
        joinMultiText,
        getAllMonths,
        getAllWeekDays,
        getAllHours,
        getAvailableMonthDays,
        getAllTransactionExplorerConditionFields,
        getAllTransactionExplorerConditionOperators,
        formatNumberToLocalizedNumeralsWithoutDigitGrouping
    } = useI18n();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const explorersStore = useExplorersStore();

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const currentCondition = ref<TransactionExplorerCondition | undefined>(undefined);

    const showExpression = ref<Record<string, boolean>>({});

    const showFilterSourceAccountsDialog = ref<boolean>(false);

    const showFilterDestinationAccountsDialog = ref<boolean>(false);

    const showFilterTransactionCategoriesDialog = ref<boolean>(false);

    const editingQuery = ref<TransactionExplorerQuery | undefined>(undefined);

    const editingQueryName = ref<string>('');

    const queries = computed<TransactionExplorerQuery[]>({
        get: () => explorersStore.currentExploration.queries,
        set: (value: TransactionExplorerQuery[]) => {
            explorersStore.currentExploration.queries = value;
        }
    });

    const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

    const hasAnyAccount = computed<boolean>(() => accountsStore.allPlainAccounts.length > 0);

    const hasAnyTransactionCategory = computed<boolean>(() => !isObjectEmpty(transactionCategoriesStore.allTransactionCategoriesMap));

    const allTransactionExplorerConditionFields = computed<NameValue[]>(() => getAllTransactionExplorerConditionFields());

    const allAvailableDayOfWeekOptions = computed<TypeAndDisplayName[]>(() => getAllWeekDays(userStore.currentUserFirstDayOfWeek));

    const allAvailableDayOfMonthOptions = computed<TypeAndDisplayName[]>(() => getAvailableMonthDays(31, 3));

    const allAvailableMonthOfYearOptions = computed<TypeAndDisplayName[]>(() => getAllMonths());

    const allAvailableHourOfDayOptions = computed<TypeAndDisplayName[]>(() => getAllHours());

    function getFilteredAccountsDisplayContent(filterAccountIds?: Record<string, boolean>): string {
        if ((props.loading && !hasAnyAccount.value) || !accountsStore.allVisiblePlainAccounts || !accountsStore.allVisiblePlainAccounts.length) {
            return '';
        }

        if (!filterAccountIds) {
            return tt('All');
        }

        let allAccountSelected = true;
        const selectedAccountNames: string[] = [];

        for (const account of accountsStore.allPlainAccounts) {
            if (account.type === AccountType.MultiSubAccounts.type) {
                continue;
            }

            if (!filterAccountIds[account.id]) {
                allAccountSelected = false;
            } else {
                selectedAccountNames.push(account.name);
            }
        }

        if (allAccountSelected) {
            return tt('All');
        } else if (selectedAccountNames.length < 1) {
            return '';
        }

        return joinMultiText(selectedAccountNames);
    }

    function getFilteredTransactionCategoriesDisplayContent(filterTransactionCategoryIds?: Record<string, boolean>): string {
        if ((props.loading && !hasAnyTransactionCategory.value) || !transactionCategoriesStore.allTransactionCategoriesMap) {
            return '';
        }

        if (!filterTransactionCategoryIds) {
            return tt('All');
        }

        let allCategorySelected = true;
        const selectedCategoryNames: string[] = [];

        for (const transactionCategory of values(transactionCategoriesStore.allTransactionCategoriesMap)) {
            if (!transactionCategory.parentId || transactionCategory.parentId === '0') {
                continue;
            }

            if (!filterTransactionCategoryIds[transactionCategory.id]) {
                allCategorySelected = false;
            } else {
                selectedCategoryNames.push(transactionCategory.name);
            }
        }

        if (allCategorySelected) {
            return tt('All');
        } else if (selectedCategoryNames.length < 1) {
            return '';
        }

        return joinMultiText(selectedCategoryNames);
    }

    function getConditionStyle(query: TransactionExplorerQuery, conditionIndex: number): Record<string, string> {
        const style: Record<string, string> = {};
        const depths = query.getConditionNestingDepths();
        const item = query.conditions[conditionIndex];
        let depth = 0;

        if (item && item.relation === TransactionExplorerConditionRelation.SubEnd) {
            depth = depths[conditionIndex - 1] ?? 0;
        } else {
            depth = depths[conditionIndex] ?? 0;
        }

        if (depth > 0) {
            style['margin-inline-start'] = (depth * 1.5) + 'rem';
        }

        return style;
    }

    function addQuery(): void {
        queries.value.push(TransactionExplorerQuery.create(generateRandomUUID()));
    }

    function updateQueryName(query: TransactionExplorerQuery): void {
        query.name = editingQueryName.value;
        editingQuery.value = undefined;
        editingQueryName.value = '';
    }

    function cancelUpdateQueryName(): void {
        editingQuery.value = undefined;
        editingQueryName.value = '';
    }

    function duplicateQuery(query: TransactionExplorerQuery): void {
        queries.value.push(query.clone(generateRandomUUID()));
    }

    function removeQuery(query: TransactionExplorerQuery, queryIndex: number): void {
        if (queries.value.length > 0) {
            queries.value.splice(queryIndex, 1);
        }

        if (explorersStore.currentExploration.datatableQuerySource === query.id) {
            explorersStore.currentExploration.datatableQuerySource = '';
        }

        if (queries.value.length < 1) {
            queries.value.push(TransactionExplorerQuery.create(generateRandomUUID()));
        }
    }

    function clearAllQueries(): void {
        queries.value.length = 0;
        queries.value.push(TransactionExplorerQuery.create(generateRandomUUID()));
    }

    function addCondition(query: TransactionExplorerQuery): void {
        const newCondition = query.addNewCondition(TransactionExplorerConditionField.TransactionType, query.conditions.length < 1);
        query.conditions.push(newCondition);
    }

    function addSubCondition(query: TransactionExplorerQuery, subEndIndex: number): void {
        const newCondition = query.addNewCondition(TransactionExplorerConditionField.TransactionType, false);
        query.conditions.splice(subEndIndex, 0, newCondition);
    }

    function removeCondition(query: TransactionExplorerQuery, conditionIndex: number): void {
        const item = query.conditions[conditionIndex];

        if (!item || item.relation === TransactionExplorerConditionRelation.SubEnd) {
            return;
        }

        query.conditions.splice(conditionIndex, 1);

        if (conditionIndex === 0 && query.conditions.length > 0) {
            const newFirstCondition = query.conditions[0];

            if (newFirstCondition) {
                if (newFirstCondition.relation === TransactionExplorerConditionRelation.AndSub || newFirstCondition.relation === TransactionExplorerConditionRelation.OrSub) {
                    removeSubCondition(query, conditionIndex + 1);
                }

                newFirstCondition.relation = TransactionExplorerConditionRelation.First;
            }
        }

        const oldStartSubCondition = item.relation === TransactionExplorerConditionRelation.AndSub || item.relation === TransactionExplorerConditionRelation.OrSub;

        if (oldStartSubCondition && conditionIndex < query.conditions.length && query.conditions[conditionIndex]) {
            const nextItem = query.conditions[conditionIndex];

            if (nextItem.relation === TransactionExplorerConditionRelation.SubEnd) {
                query.conditions.splice(conditionIndex, 1);
            } else if (nextItem.relation === TransactionExplorerConditionRelation.AndSub || nextItem.relation === TransactionExplorerConditionRelation.OrSub) {
                nextItem.relation = item.relation;
                removeSubCondition(query, conditionIndex);
            } else {
                nextItem.relation = item.relation;
            }
        }
    }

    function removeSubCondition(query: TransactionExplorerQuery, conditionIndex: number): void {
        let depth = 1;

        for (let i = conditionIndex + 1; i < query.conditions.length; i++) {
            const currentCondition = query.conditions[i];

            if (!currentCondition) {
                continue;
            }

            if (currentCondition.relation === TransactionExplorerConditionRelation.AndSub || currentCondition.relation === TransactionExplorerConditionRelation.OrSub) {
                depth++;
            } else if (currentCondition.relation === TransactionExplorerConditionRelation.SubEnd) {
                depth--;

                if (depth === 0) {
                    query.conditions.splice(i, 1);
                    break;
                }
            }
        }
    }

    function updateItemRelation(query: TransactionExplorerQuery, conditionIndex: number, value: TransactionExplorerConditionRelation): void {
        const item = query.conditions[conditionIndex];

        if (!item || item.relation === TransactionExplorerConditionRelation.SubEnd) {
            return;
        }

        const oldStartSubCondition = item.relation === TransactionExplorerConditionRelation.AndSub || item.relation === TransactionExplorerConditionRelation.OrSub;
        const newStartSubCondition = value === TransactionExplorerConditionRelation.AndSub || value === TransactionExplorerConditionRelation.OrSub;

        item.relation = value;

        if (!oldStartSubCondition && newStartSubCondition) {
            query.conditions.splice(conditionIndex + 1, 0, query.addSubConditionEnd());
        } else if (oldStartSubCondition && !newStartSubCondition) {
            removeSubCondition(query, conditionIndex);
        }
    }

    function updateConditionField(query: TransactionExplorerQuery, conditionIndex: number, newField: TransactionExplorerConditionField | undefined): void {
        if (!newField) {
            return;
        }

        const oldConditionWithRelation = query.conditions[conditionIndex];

        if (!oldConditionWithRelation) {
            return;
        }

        const newConditionWithRelation = query.addNewCondition(newField, conditionIndex < 1);
        oldConditionWithRelation.condition = newConditionWithRelation.condition;
    }

    function updateSourceAccount(changed: boolean, selectedAccountIds?: string[]): void {
        if (!changed || !currentCondition.value || currentCondition.value.field !== TransactionExplorerConditionField.SourceAccount.value) {
            showFilterSourceAccountsDialog.value = false;
            return;
        }

        currentCondition.value.value = selectedAccountIds || [];
        currentCondition.value = undefined;
        showFilterSourceAccountsDialog.value = false;
    }

    function updateDestinationAccount(changed: boolean, selectedAccountIds?: string[]): void {
        if (!changed || !currentCondition.value || currentCondition.value.field !== TransactionExplorerConditionField.DestinationAccount.value) {
            showFilterDestinationAccountsDialog.value = false;
            return;
        }

        currentCondition.value.value = selectedAccountIds || [];
        currentCondition.value = undefined;
        showFilterDestinationAccountsDialog.value = false;
    }

    function updateTransactionCategories(changed: boolean, selectedCategoryIds?: string[]): void {
        if (!changed || !currentCondition.value || currentCondition.value.field !== TransactionExplorerConditionField.TransactionCategory.value) {
            showFilterTransactionCategoriesDialog.value = false;
            return;
        }

        currentCondition.value.value = selectedCategoryIds || [];
        currentCondition.value = undefined;
        showFilterTransactionCategoriesDialog.value = false;
    }

    function getExpression(query: TransactionExplorerQuery, queryIndex: number): string {
        if (!query.conditions || query.conditions.length < 1) {
            return '';
        }

        try {
            return query.toExpression(transactionCategoriesStore.allTransactionCategoriesMap, accountsStore.allAccountsMap, transactionTagsStore.allTransactionTagsMap);
        } catch (ex) {
            logger.error('failed to generate expression for exploration query#' + queryIndex, ex);
            snackbar.value?.showError(tt('Failed to generate expression'));
            return tt('Failed to generate expression');
        }
    }

    if (queries.value.length === 0) {
        queries.value.push(TransactionExplorerQuery.create(generateRandomUUID()));
    }

    watch(() => queries.value, () => {
        editingQuery.value = undefined;
        editingQueryName.value = '';
    });

    return {
        tt, getAllTransactionExplorerConditionOperators, formatNumberToLocalizedNumeralsWithoutDigitGrouping, snackbar,
        currentCondition, showExpression, showFilterSourceAccountsDialog, showFilterDestinationAccountsDialog,
        showFilterTransactionCategoriesDialog, editingQuery, editingQueryName, queries, defaultCurrency, hasAnyAccount,
        hasAnyTransactionCategory, allTransactionExplorerConditionFields, allAvailableDayOfWeekOptions,
        allAvailableDayOfMonthOptions, allAvailableMonthOfYearOptions, allAvailableHourOfDayOptions,
        getFilteredAccountsDisplayContent, getFilteredTransactionCategoriesDisplayContent, getConditionStyle, addQuery,
        updateQueryName, cancelUpdateQueryName, duplicateQuery, removeQuery, clearAllQueries, addCondition, addSubCondition,
        removeCondition, updateItemRelation, updateConditionField, updateSourceAccount, updateDestinationAccount,
        updateTransactionCategories, getExpression,
    };
}
