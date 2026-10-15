import { useTransactionTagSelectionBase } from '@/components/base/TransactionTagSelectionBase.ts';
import SnackBar from '@/components/desktop/SnackBar.vue';
import type { GenericNameValue, NameNumeralValue, NameValue, TypeAndDisplayName } from '@/core/base.ts';
import { KnownFileType } from '@/core/file.ts';
import {
ImportTransactionReplaceRuleAction,
ImportTransactionReplaceRuleActionType,
type ImportTransactionReplaceRuleApplyScope,
ImportTransactionReplaceRuleConditionFieldType,
ImportTransactionReplaceRuleConditionOperator,
ImportTransactionReplaceRuleConditionOperatorType
} from '@/core/rule.ts';
import type { LocalizedTimezoneInfo } from '@/core/timezone.ts';
import { TransactionType } from '@/core/transaction.ts';
import { getTransactionPrimaryCategoryName, getTransactionSecondaryCategoryName } from '@/lib/category.ts';
import { getCurrentUnixTime } from '@/lib/datetime.ts';
import logger from '@/lib/logger.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import { openTextFileContent, startDownloadFile } from '@/lib/ui/common.ts';
import { useI18n } from '@/locales/helpers.ts';
import { Account, type CategorizedAccountWithDisplayBalance } from '@/models/account.ts';
import { ImportTransactionReplaceRule, ImportTransactionReplaceRules } from '@/models/rule.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useUserStore } from '@/stores/user.ts';
import { computed, nextTick, ref, useTemplateRef } from 'vue';
import { getActionCategoryType, isAmountCondition, isEmptyTextConditionOperator, isItemCondition, isRangeOperator } from './rule-fields.ts';

export type SnackBarType = InstanceType<typeof SnackBar>;

export interface BatchApplyRulesDialogResponse {
    rules: ImportTransactionReplaceRule[];
    scope: ImportTransactionReplaceRuleApplyScope;
}

// Coordinates BatchApplyRulesDialog state and effects; called once from the component setup.
export function useBatchApplyRulesDialog() {
    const {
        tt,
        getAllTimezones,
        getAllImportTransactionReplaceRuleConditionFields,
        formatNumberToLocalizedNumeralsWithoutDigitGrouping,
        getCategorizedAccountsWithDisplayBalance
    } = useI18n();

    const {
        allTagsWithGroupHeader
    } = useTransactionTagSelectionBase({ modelValue: [] }, false);

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const rulesContainer = useTemplateRef<HTMLDivElement>('rulesContainer');

    let resolveFunc: ((response: BatchApplyRulesDialogResponse) => void) | null = null;

    let rejectFunc: ((reason?: unknown) => void) | null = null;

    const showState = ref<boolean>(false);

    const loading = ref<boolean>(false);

    const rules = ref<ImportTransactionReplaceRule[]>([]);

    const editingRule = ref<ImportTransactionReplaceRule | undefined>(undefined);

    const editingRuleName = ref<string>('');

    const selectedTransactionCount = ref<number>(0);

    const sourceExpenseCategoryNames = ref<NameValue[]>([]);

    const sourceIncomeCategoryNames = ref<NameValue[]>([]);

    const sourceTransferCategoryNames = ref<NameValue[]>([]);

    const sourceAccountNames = ref<NameValue[]>([]);

    const destinationAccountNames = ref<NameValue[]>([]);

    const sourceTagNames = ref<NameValue[]>([]);

    const showAccountBalance = computed<boolean>(() => settingsStore.appSettings.showAccountBalance);

    const customAccountCategoryOrder = computed<string>(() => settingsStore.appSettings.accountCategoryOrders);

    const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

    const allAccounts = computed<Account[]>(() => accountsStore.allPlainAccounts);

    const allVisibleAccounts = computed<Account[]>(() => accountsStore.allVisiblePlainAccounts);

    const allVisibleCategorizedAccounts = computed<CategorizedAccountWithDisplayBalance[]>(() => getCategorizedAccountsWithDisplayBalance(allVisibleAccounts.value, showAccountBalance.value, customAccountCategoryOrder.value));

    const allCategories = computed<Record<number, TransactionCategory[]>>(() => transactionCategoriesStore.allTransactionCategories);

    const allTimezones = computed<LocalizedTimezoneInfo[]>(() => getAllTimezones(getCurrentUnixTime(), false));

    const allConditionFields = computed<NameValue[]>(() => getAllImportTransactionReplaceRuleConditionFields());

    const allActions = computed<GenericNameValue<ImportTransactionReplaceRuleActionType>[]>(() => {
        const result = [];

        for (const action of ImportTransactionReplaceRuleAction.values()) {
            result.push({
                name: tt(action.name),
                value: action.value
            })
        }

        return result;
    });

    const transactionTypes = computed<TypeAndDisplayName[]>(() => [
        { displayName: tt('Income'), type: TransactionType.Income },
        { displayName: tt('Expense'), type: TransactionType.Expense },
        { displayName: tt('Transfer'), type: TransactionType.Transfer }
    ]);

    const amountSigns = computed<NameNumeralValue[]>(() => [
        { name: tt('Positive'), value: 1 },
        { name: tt('Negative'), value: -1 }
    ]);

    const canApply = computed<boolean>(() => rules.value.length > 0 && rules.value.every(rule => rule.isValid()));

    const isModified = computed<boolean>(() => rules.value.length > 1 || (rules.value.length === 1 && !rules.value[0]?.equals(createNewRule(), false)));

    function createNewRule(): ImportTransactionReplaceRule {
        return ImportTransactionReplaceRule.of(
            generateRandomUUID(), '',
            ImportTransactionReplaceRuleConditionFieldType.DescriptionNormalized,
            ImportTransactionReplaceRuleConditionOperatorType.Contains, '',
            ImportTransactionReplaceRuleActionType.SetExpenseCategory, '');
    }

    function getDefaultRuleName(index: number): string {
        return tt('format.misc.ruleIndex', { index: formatNumberToLocalizedNumeralsWithoutDigitGrouping(index + 1) });
    }

    function getRuleName(rule: ImportTransactionReplaceRule, index: number): string {
        return rule.name || getDefaultRuleName(index);
    }

    function getConditionOperators(rule: ImportTransactionReplaceRule): NameValue[] {
        let operators: ImportTransactionReplaceRuleConditionOperatorType[];

        if (rule.conditionField === ImportTransactionReplaceRuleConditionFieldType.Tag) {
            operators = [
                ImportTransactionReplaceRuleConditionOperatorType.Contains
            ];
        } else if (isItemCondition(rule.conditionField)) {
            operators = [
                ImportTransactionReplaceRuleConditionOperatorType.Is
            ];
        } else if (isAmountCondition(rule.conditionField)) {
            operators = [
                ImportTransactionReplaceRuleConditionOperatorType.Equals,
                ImportTransactionReplaceRuleConditionOperatorType.NotEquals,
                ImportTransactionReplaceRuleConditionOperatorType.GreaterThan,
                ImportTransactionReplaceRuleConditionOperatorType.LessThan,
                ImportTransactionReplaceRuleConditionOperatorType.Between,
                ImportTransactionReplaceRuleConditionOperatorType.NotBetween
            ];
        } else {
            operators = [
                ImportTransactionReplaceRuleConditionOperatorType.IsEmpty,
                ImportTransactionReplaceRuleConditionOperatorType.IsNotEmpty,
                ImportTransactionReplaceRuleConditionOperatorType.Equals,
                ImportTransactionReplaceRuleConditionOperatorType.NotEquals,
                ImportTransactionReplaceRuleConditionOperatorType.Contains,
                ImportTransactionReplaceRuleConditionOperatorType.NotContains,
                ImportTransactionReplaceRuleConditionOperatorType.StartsWith,
                ImportTransactionReplaceRuleConditionOperatorType.NotStartsWith,
                ImportTransactionReplaceRuleConditionOperatorType.EndsWith,
                ImportTransactionReplaceRuleConditionOperatorType.NotEndsWith,
                ImportTransactionReplaceRuleConditionOperatorType.RegexMatch,
                ImportTransactionReplaceRuleConditionOperatorType.NotRegexMatch
            ];
        }

        return operators.map(operator => ({
            name: tt(ImportTransactionReplaceRuleConditionOperator.valueOf(operator)?.name || ''),
            value: operator
        }));
    }

    function isActionDisabledForRule(type: ImportTransactionReplaceRuleActionType, rule: ImportTransactionReplaceRule): boolean {
        if (type === ImportTransactionReplaceRuleActionType.SetExpenseCategory) {
            return rule.conditionField === ImportTransactionReplaceRuleConditionFieldType.IncomeCategory || rule.conditionField === ImportTransactionReplaceRuleConditionFieldType.TransferCategory;
        } else if (type === ImportTransactionReplaceRuleActionType.SetIncomeCategory) {
            return rule.conditionField === ImportTransactionReplaceRuleConditionFieldType.ExpenseCategory || rule.conditionField === ImportTransactionReplaceRuleConditionFieldType.TransferCategory;
        } else if (type === ImportTransactionReplaceRuleActionType.SetTransferCategory) {
            return rule.conditionField === ImportTransactionReplaceRuleConditionFieldType.ExpenseCategory || rule.conditionField === ImportTransactionReplaceRuleConditionFieldType.IncomeCategory;
        } else if (type === ImportTransactionReplaceRuleActionType.ReplaceTag || type === ImportTransactionReplaceRuleActionType.DeleteTag) {
            return rule.conditionField !== ImportTransactionReplaceRuleConditionFieldType.Tag;
        }

        return false;
    }

    function getDefaultActionTypeForConditionField(field: ImportTransactionReplaceRuleConditionFieldType): ImportTransactionReplaceRuleActionType {
        if (field === ImportTransactionReplaceRuleConditionFieldType.ExpenseCategory) {
            return ImportTransactionReplaceRuleActionType.SetExpenseCategory;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.IncomeCategory) {
            return ImportTransactionReplaceRuleActionType.SetIncomeCategory;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.TransferCategory) {
            return ImportTransactionReplaceRuleActionType.SetTransferCategory;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.SourceAccount) {
            return ImportTransactionReplaceRuleActionType.SetSourceAccount;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.DestinationAccount) {
            return ImportTransactionReplaceRuleActionType.SetDestinationAccount;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.Tag) {
            return ImportTransactionReplaceRuleActionType.ReplaceTag;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.Amount) {
            return ImportTransactionReplaceRuleActionType.SetSourceAmount;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.TransferInAmount) {
            return ImportTransactionReplaceRuleActionType.SetDestinationAmount;
        }

        return ImportTransactionReplaceRuleActionType.SetTransactionType;
    }

    function getSourceItems(field: ImportTransactionReplaceRuleConditionFieldType): NameValue[] {
        if (field === ImportTransactionReplaceRuleConditionFieldType.ExpenseCategory) {
            return sourceExpenseCategoryNames.value;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.IncomeCategory) {
            return sourceIncomeCategoryNames.value;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.TransferCategory) {
            return sourceTransferCategoryNames.value;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.SourceAccount) {
            return sourceAccountNames.value;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.DestinationAccount) {
            return destinationAccountNames.value;
        } else if (field === ImportTransactionReplaceRuleConditionFieldType.Tag) {
            return sourceTagNames.value;
        } else {
            return [];
        }
    }

    function getSelectedCategoryPrimaryName(rule: ImportTransactionReplaceRule): string {
        const targetValue = typeof rule.targetValue === 'string' ? rule.targetValue : '';
        return getTransactionPrimaryCategoryName(targetValue, allCategories.value[getActionCategoryType(rule.actionType)]);
    }

    function getSelectedCategorySecondaryName(rule: ImportTransactionReplaceRule): string {
        const targetValue = typeof rule.targetValue === 'string' ? rule.targetValue : '';
        return getTransactionSecondaryCategoryName(targetValue, allCategories.value[getActionCategoryType(rule.actionType)]);
    }

    function getAccountDisplayName(accountId: string | number): string {
        return typeof accountId === 'string' ? Account.findAccountNameById(allAccounts.value, accountId) || '' : '';
    }

    function open(options: { expenseCategoryNames: NameValue[], incomeCategoryNames: NameValue[], transferCategoryNames: NameValue[], sourceAccountNames: NameValue[], destinationAccountNames: NameValue[], tagNames: NameValue[], selectedTransactionCount: number }): Promise<BatchApplyRulesDialogResponse> {
        rules.value = [];
        editingRule.value = undefined;
        editingRuleName.value = '';
        sourceExpenseCategoryNames.value = options.expenseCategoryNames;
        sourceIncomeCategoryNames.value = options.incomeCategoryNames;
        sourceTransferCategoryNames.value = options.transferCategoryNames;
        sourceAccountNames.value = options.sourceAccountNames;
        destinationAccountNames.value = options.destinationAccountNames;
        sourceTagNames.value = options.tagNames;
        selectedTransactionCount.value = options.selectedTransactionCount;
        showState.value = true;

        addRule();

        return new Promise((resolve, reject) => {
            resolveFunc = resolve;
            rejectFunc = reject;
        });
    }

    function reload(): void {
        loading.value = true;

        Promise.allSettled([
            accountsStore.loadAllAccounts({ force: true }),
            transactionCategoriesStore.loadAllCategories({ force: true }),
            transactionTagsStore.loadAllTags({ force: true })
        ]).then(results => {
            loading.value = false;

            const isAllUpToDate = results.length === 3
                && results[0].status === 'rejected' && results[0].reason?.isUpToDate
                && results[1].status === 'rejected' && results[1].reason?.isUpToDate
                && results[2].status === 'rejected' && results[2].reason?.isUpToDate;

            // show info if all up to date
            if (isAllUpToDate) {
                snackbar.value?.showMessage('Data is up to date');
                return;
            }

            // show error if any
            for (const result of results) {
                if (result.status === 'rejected' && !result.reason?.isUpToDate) {
                    snackbar.value?.showError(result.reason);
                    return;
                }
            }

            // show info if one of them updated
            for (const result of results) {
                if (result.status === 'fulfilled') {
                    snackbar.value?.showMessage('Data has been updated');
                    return;
                }
            }
        });
    }

    function addRule(): void {
        const rule = createNewRule();
        rules.value.push(rule);
        editingRuleName.value = '';
        editingRule.value = undefined;

        nextTick(() => {
            const ruleCards = rulesContainer.value?.querySelectorAll<HTMLElement>('.rule-card');

            if (!ruleCards?.length) {
                return;
            }

            ruleCards[ruleCards.length - 1]?.scrollIntoView({
                behavior: 'smooth',
                block: 'nearest'
            });
        });
    }

    function removeRule(index: number): void {
        rules.value.splice(index, 1);
    }

    function updateRuleName(rule: ImportTransactionReplaceRule): void {
        rule.name = editingRuleName.value;
        editingRule.value = undefined;
        editingRuleName.value = '';
    }

    function cancelUpdateRuleName(): void {
        editingRule.value = undefined;
        editingRuleName.value = '';
    }

    function updateConditionField(rule: ImportTransactionReplaceRule, field: ImportTransactionReplaceRuleConditionFieldType): void {
        rule.conditionField = field;

        if (field === ImportTransactionReplaceRuleConditionFieldType.Tag) {
            rule.conditionOperator = ImportTransactionReplaceRuleConditionOperatorType.Contains;
            rule.conditionValue = '';
        } else if (isItemCondition(field)) {
            rule.conditionOperator = ImportTransactionReplaceRuleConditionOperatorType.Is;
            rule.conditionValue = '';
        } else if (isAmountCondition(field)) {
            rule.conditionOperator = ImportTransactionReplaceRuleConditionOperatorType.Between;
            rule.conditionValue = [0, 0];
        } else {
            rule.conditionOperator = ImportTransactionReplaceRuleConditionOperatorType.Contains;
            rule.conditionValue = '';
        }

        const currentAction = ImportTransactionReplaceRuleAction.valueOf(rule.actionType);

        if (!currentAction || isActionDisabledForRule(currentAction.value, rule)) {
            updateActionType(rule, getDefaultActionTypeForConditionField(field));
        }
    }

    function updateStringConditionValue(rule: ImportTransactionReplaceRule, value: unknown): void {
        rule.conditionValue = typeof value === 'string' ? value : '';
    }

    function updateConditionOperator(rule: ImportTransactionReplaceRule, operator: ImportTransactionReplaceRuleConditionOperatorType): void {
        rule.conditionOperator = operator;

        if (isEmptyTextConditionOperator(operator)) {
            rule.conditionValue = '';
        }
    }

    function getAmountConditionValue(rule: ImportTransactionReplaceRule, index: number): number {
        return Array.isArray(rule.conditionValue) ? rule.conditionValue[index] || 0 : 0;
    }

    function updateAmountConditionValue(rule: ImportTransactionReplaceRule, index: number, value: number): void {
        if (!Array.isArray(rule.conditionValue)) {
            rule.conditionValue = [0, 0];
        }

        rule.conditionValue[index] = value;

        if (!isRangeOperator(rule.conditionOperator)) {
            rule.conditionValue[1] = rule.conditionValue[0];
        }
    }

    function updateActionType(rule: ImportTransactionReplaceRule, actionType: ImportTransactionReplaceRuleActionType): void {
        rule.actionType = actionType;

        if (actionType === ImportTransactionReplaceRuleActionType.SetTransactionType) {
            rule.targetValue = TransactionType.Expense;
        } else if (actionType === ImportTransactionReplaceRuleActionType.SetSourceAmount || actionType === ImportTransactionReplaceRuleActionType.SetDestinationAmount) {
            rule.targetValue = 1;
        } else {
            rule.targetValue = '';
        }
    }

    function loadReplaceRuleFile(): void {
        openTextFileContent({
            allowedExtensions: KnownFileType.JSON.contentType
        }).then(content => {
            const result = ImportTransactionReplaceRules.parseFromJson(content, generateRandomUUID);

            if (result) {
                rules.value = result.getRules();
            } else {
                logger.error('Failed to parse replace rule file');
                snackbar.value?.showError('Replace rule file is invalid');
            }
        }).catch(error => {
            logger.error('Failed to open replace rule file', error);
            snackbar.value?.showError('Replace rule file is invalid');
        });
    }

    function saveReplaceRuleFile(): void {
        const fileName = KnownFileType.JSON.formatFileName(tt('dataExport.defaultImportReplaceRuleFileName'));
        startDownloadFile(fileName, KnownFileType.JSON.createBlob(ImportTransactionReplaceRules.of(rules.value).toJson()));
    }

    function apply(scope: ImportTransactionReplaceRuleApplyScope): void {
        resolveFunc?.({
            rules: rules.value,
            scope: scope
        });
        showState.value = false;
    }

    function cancel(): void {
        rejectFunc?.();
        showState.value = false;
    }

    return {
        tt, allTagsWithGroupHeader, snackbar, rulesContainer, showState, loading, rules, editingRule, editingRuleName,
        defaultCurrency, allVisibleAccounts, allVisibleCategorizedAccounts, allCategories, allTimezones, allConditionFields,
        allActions, transactionTypes, amountSigns, canApply, isModified, getDefaultRuleName, getRuleName, getConditionOperators,
        isActionDisabledForRule, getSourceItems, isItemCondition, isAmountCondition, isRangeOperator,
        isEmptyTextConditionOperator, getActionCategoryType, getSelectedCategoryPrimaryName, getSelectedCategorySecondaryName,
        getAccountDisplayName, open, reload, addRule, removeRule, updateRuleName, cancelUpdateRuleName, updateConditionField,
        updateStringConditionValue, updateConditionOperator, getAmountConditionValue, updateAmountConditionValue,
        updateActionType, loadReplaceRuleFile, saveReplaceRuleFile, apply, cancel,
    };
}
