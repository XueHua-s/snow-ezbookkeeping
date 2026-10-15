import { type PartialRecord, keysIfValueEquals } from '@/core/base.ts';
import {
    TransactionExplorerConditionField,
    TransactionExplorerConditionOperator,
    type TransactionExplorerConditionOperatorType,
    TransactionExplorerConditionRelation
} from '@/core/explorer.ts';
import {
    type AccountConditionOperator, type TransactionCategoryConditionOperator, type TransactionTypeConditionOperator,
    TransactionExplorerDestinationAccountCondition, TransactionExplorerSourceAccountCondition,
    TransactionExplorerTransactionCategoryCondition, TransactionExplorerTransactionTypeCondition,
} from './account-conditions.ts';
import { type AmountConditionOperator, TransactionExplorerDestinationAmountCondition, TransactionExplorerSourceAmountCondition } from './amount-conditions.ts';
import { type TransactionExplorerCondition, TransactionExplorerUndefinedCondition } from './condition-types.ts';
import {
    type DescriptionConditionOperator, type PicturesConditionOperator, TransactionExplorerDescriptionCaseInsensitiveCondition,
    TransactionExplorerDescriptionCondition, TransactionExplorerDescriptionNormalizedCondition,
    TransactionExplorerPicturesCondition,
} from './description-conditions.ts';
import { type GeoLocationConditionOperator, TransactionExplorerGeoLocationCondition } from './location-condition.ts';
import { type TransactionTagConditionOperator, TransactionExplorerTransactionTagCondition } from './tag-condition.ts';
import {
    type TransactionTimeDayOfMonthConditionOperator, type TransactionTimeDayOfWeekConditionOperator,
    type TransactionTimeHourOfDayConditionOperator, type TransactionTimeMonthOfYearConditionOperator,
    type TransactionTimezoneConditionOperator, TransactionExplorerTransactionTimeDayOfMonthCondition,
    TransactionExplorerTransactionTimeDayOfWeekCondition, TransactionExplorerTransactionTimeHourOfDayCondition,
    TransactionExplorerTransactionTimeMonthOfYearCondition, TransactionExplorerTransactionTimezoneCondition,
} from './time-conditions.ts';

export class TransactionExplorerConditionWithRelation {
    public condition: TransactionExplorerCondition;
    public relation: TransactionExplorerConditionRelation;

    constructor(condition: TransactionExplorerCondition, relation: TransactionExplorerConditionRelation) {
        this.condition = condition;
        this.relation = relation;
    }

    public getSupportedOperators(): TransactionExplorerConditionOperator[] {
        let operatorTypes: PartialRecord<TransactionExplorerConditionOperatorType, true> = {};

        switch (this.condition.field) {
            case TransactionExplorerConditionField.TransactionTimeDayOfWeek.value:
                operatorTypes = TransactionExplorerTransactionTimeDayOfWeekCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.TransactionTimeDayOfMonth.value:
                operatorTypes = TransactionExplorerTransactionTimeDayOfMonthCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.TransactionTimeMonthOfYear.value:
                operatorTypes = TransactionExplorerTransactionTimeMonthOfYearCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.TransactionTimeHourOfDay.value:
                operatorTypes = TransactionExplorerTransactionTimeHourOfDayCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.TransactionTimezone.value:
                operatorTypes = TransactionExplorerTransactionTimezoneCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.TransactionType.value:
                operatorTypes = TransactionExplorerTransactionTypeCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.TransactionCategory.value:
                operatorTypes = TransactionExplorerTransactionCategoryCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.SourceAccount.value:
                operatorTypes = TransactionExplorerSourceAccountCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.DestinationAccount.value:
                operatorTypes = TransactionExplorerDestinationAccountCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.SourceAmount.value:
                operatorTypes = TransactionExplorerSourceAmountCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.DestinationAmount.value:
                operatorTypes = TransactionExplorerDestinationAmountCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.GeoLocation.value:
                operatorTypes = TransactionExplorerGeoLocationCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.TransactionTag.value:
                operatorTypes = TransactionExplorerTransactionTagCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.Pictures.value:
                operatorTypes = TransactionExplorerPicturesCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.Description.value:
                operatorTypes = TransactionExplorerDescriptionCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.DescriptionCaseInsensitive.value:
                operatorTypes = TransactionExplorerDescriptionCaseInsensitiveCondition.supportedOperators;
                break;
            case TransactionExplorerConditionField.DescriptionNormalized.value:
                operatorTypes = TransactionExplorerDescriptionNormalizedCondition.supportedOperators;
                break;
            default:
                return [];
        }

        const result: TransactionExplorerConditionOperator[] = [];

        for (const key of keysIfValueEquals(operatorTypes, true)) {
            const operator = TransactionExplorerConditionOperator.valueOf(key);

            if (operator) {
                result.push(operator);
            }
        }

        return result;
    }

    public toJsonObject(): unknown {
        if (this.relation === TransactionExplorerConditionRelation.SubEnd) {
            return {
                relation: this.relation
            };
        }

        return {
            condition: {
                field: this.condition.field,
                operator: this.condition.operator,
                value: this.condition.getValueForStore()
            },
            relation: this.relation
        };
    }

    public static parse(data: unknown): TransactionExplorerConditionWithRelation | null {
        if (typeof data !== 'object' || !data || !('relation' in data)) {
            return null;
        }

        const relation = data['relation'];
        let condition: TransactionExplorerCondition | null = null;

        if (relation === TransactionExplorerConditionRelation.First ||
            relation === TransactionExplorerConditionRelation.And || relation === TransactionExplorerConditionRelation.Or ||
            relation === TransactionExplorerConditionRelation.AndSub || relation === TransactionExplorerConditionRelation.OrSub) {
            if (!('condition' in data)) {
                return null;
            }

            const conditionObject = data['condition'];

            if (typeof conditionObject !== 'object' || !conditionObject || !('field' in conditionObject) || !('operator' in conditionObject) || !('value' in conditionObject) || typeof relation !== 'string') {
                return null;
            }

            const conditionField = conditionObject['field'];
            const conditionOperator = conditionObject['operator'] as TransactionExplorerConditionOperatorType;
            const conditionValue = conditionObject['value'];

            switch (conditionField) {
                case TransactionExplorerConditionField.TransactionTimeDayOfWeek.value:
                    if (TransactionExplorerTransactionTimeDayOfWeekCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerTransactionTimeDayOfWeekCondition(conditionOperator as TransactionTimeDayOfWeekConditionOperator, conditionValue as number[]);
                    }
                    break;
                case TransactionExplorerConditionField.TransactionTimeDayOfMonth.value:
                    if (TransactionExplorerTransactionTimeDayOfMonthCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerTransactionTimeDayOfMonthCondition(conditionOperator as TransactionTimeDayOfMonthConditionOperator, conditionValue as number[]);
                    }
                    break;
                case TransactionExplorerConditionField.TransactionTimeMonthOfYear.value:
                    if (TransactionExplorerTransactionTimeMonthOfYearCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerTransactionTimeMonthOfYearCondition(conditionOperator as TransactionTimeMonthOfYearConditionOperator, conditionValue as number[]);
                    }
                    break;
                case TransactionExplorerConditionField.TransactionTimeHourOfDay.value:
                    if (TransactionExplorerTransactionTimeHourOfDayCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerTransactionTimeHourOfDayCondition(conditionOperator as TransactionTimeHourOfDayConditionOperator, conditionValue as number[]);
                    }
                    break;
                case TransactionExplorerConditionField.TransactionTimezone.value:
                    if (TransactionExplorerTransactionTimezoneCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue) && conditionValue.length === 2) {
                        condition = new TransactionExplorerTransactionTimezoneCondition(conditionOperator as TransactionTimezoneConditionOperator, conditionValue as [number, number]);
                    }
                    break;
                case TransactionExplorerConditionField.TransactionType.value:
                    if (TransactionExplorerTransactionTypeCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerTransactionTypeCondition(conditionOperator as TransactionTypeConditionOperator, conditionValue as number[]);
                    }
                    break;
                case TransactionExplorerConditionField.TransactionCategory.value:
                    if (TransactionExplorerTransactionCategoryCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerTransactionCategoryCondition(conditionOperator as TransactionCategoryConditionOperator, conditionValue as string[]);
                    }
                    break;
                case TransactionExplorerConditionField.SourceAccount.value:
                    if (TransactionExplorerSourceAccountCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerSourceAccountCondition(conditionOperator as AccountConditionOperator, conditionValue as string[]);
                    }
                    break;
                case TransactionExplorerConditionField.DestinationAccount.value:
                    if (TransactionExplorerDestinationAccountCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerDestinationAccountCondition(conditionOperator as AccountConditionOperator, conditionValue as string[]);
                    }
                    break;
                case TransactionExplorerConditionField.SourceAmount.value:
                    if (TransactionExplorerSourceAmountCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue) && conditionValue.length === 2) {
                        condition = new TransactionExplorerSourceAmountCondition(conditionOperator as AmountConditionOperator, conditionValue as [number, number]);
                    }
                    break;
                case TransactionExplorerConditionField.DestinationAmount.value:
                    if (TransactionExplorerDestinationAmountCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue) && conditionValue.length === 2) {
                        condition = new TransactionExplorerDestinationAmountCondition(conditionOperator as AmountConditionOperator, conditionValue as [number, number]);
                    }
                    break;
                case TransactionExplorerConditionField.GeoLocation.value:
                    if (TransactionExplorerGeoLocationCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerGeoLocationCondition(conditionOperator as GeoLocationConditionOperator, conditionValue as [number, number, number, number]);
                    }
                    break;
                case TransactionExplorerConditionField.TransactionTag.value:
                    if (TransactionExplorerTransactionTagCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerTransactionTagCondition(conditionOperator as TransactionTagConditionOperator, conditionValue as string[]);
                    }
                    break;
                case TransactionExplorerConditionField.Pictures.value:
                    if (TransactionExplorerPicturesCondition.supportedOperators[conditionOperator] && Array.isArray(conditionValue)) {
                        condition = new TransactionExplorerPicturesCondition(conditionOperator as PicturesConditionOperator, conditionValue as string[]);
                    }
                    break;
                case TransactionExplorerConditionField.Description.value:
                    if (TransactionExplorerDescriptionCondition.supportedOperators[conditionOperator] && typeof conditionValue === 'string') {
                        condition = new TransactionExplorerDescriptionCondition(conditionOperator as DescriptionConditionOperator, conditionValue);
                    }
                    break;
                case TransactionExplorerConditionField.DescriptionCaseInsensitive.value:
                    if (TransactionExplorerDescriptionCaseInsensitiveCondition.supportedOperators[conditionOperator] && typeof conditionValue === 'string') {
                        condition = new TransactionExplorerDescriptionCaseInsensitiveCondition(conditionOperator as DescriptionConditionOperator, conditionValue);
                    }
                    break;
                case TransactionExplorerConditionField.DescriptionNormalized.value:
                    if (TransactionExplorerDescriptionNormalizedCondition.supportedOperators[conditionOperator] && typeof conditionValue === 'string') {
                        condition = new TransactionExplorerDescriptionNormalizedCondition(conditionOperator as DescriptionConditionOperator, conditionValue);
                    }
                    break;
                default:
                    break;
            }
        } else if (relation === TransactionExplorerConditionRelation.SubEnd) {
            condition = new TransactionExplorerUndefinedCondition();
        } else {
            return null;
        }

        if (condition === null) {
            return null;
        }

        return new TransactionExplorerConditionWithRelation(condition, relation);
    }
}
