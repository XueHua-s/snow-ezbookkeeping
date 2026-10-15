import { CategoryType } from '@/core/category.ts';
import {
ImportTransactionReplaceRuleActionType,
ImportTransactionReplaceRuleConditionFieldType,
ImportTransactionReplaceRuleConditionOperatorType
} from '@/core/rule.ts';

export function isItemCondition(field: ImportTransactionReplaceRuleConditionFieldType): boolean {
    return field === ImportTransactionReplaceRuleConditionFieldType.ExpenseCategory
        || field === ImportTransactionReplaceRuleConditionFieldType.IncomeCategory
        || field === ImportTransactionReplaceRuleConditionFieldType.TransferCategory
        || field === ImportTransactionReplaceRuleConditionFieldType.SourceAccount
        || field === ImportTransactionReplaceRuleConditionFieldType.DestinationAccount
        || field === ImportTransactionReplaceRuleConditionFieldType.Tag;
}

export function isAmountCondition(field: ImportTransactionReplaceRuleConditionFieldType): boolean {
    return field === ImportTransactionReplaceRuleConditionFieldType.Amount || field === ImportTransactionReplaceRuleConditionFieldType.TransferInAmount;
}

export function isRangeOperator(operator: ImportTransactionReplaceRuleConditionOperatorType): boolean {
    return operator === ImportTransactionReplaceRuleConditionOperatorType.Between || operator === ImportTransactionReplaceRuleConditionOperatorType.NotBetween;
}

export function isEmptyTextConditionOperator(operator: ImportTransactionReplaceRuleConditionOperatorType): boolean {
    return operator === ImportTransactionReplaceRuleConditionOperatorType.IsEmpty || operator === ImportTransactionReplaceRuleConditionOperatorType.IsNotEmpty;
}

export function getActionCategoryType(action: ImportTransactionReplaceRuleActionType): CategoryType {
    if (action === ImportTransactionReplaceRuleActionType.SetIncomeCategory) {
        return CategoryType.Income;
    } else if (action === ImportTransactionReplaceRuleActionType.SetTransferCategory) {
        return CategoryType.Transfer;
    } else {
        return CategoryType.Expense;
    }
}
