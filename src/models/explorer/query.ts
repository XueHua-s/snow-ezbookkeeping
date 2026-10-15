import { itemAndIndex } from '@/core/base.ts';
import {
    type TransactionExplorerSubConditionStartRelation,
    TransactionExplorerConditionField,
    TransactionExplorerConditionOperatorType,
    TransactionExplorerConditionRelation,
    TransactionExplorerConditionRelationPriority,
    TransactionExplorerSubConditionStartRelationPlaceholder
} from '@/core/explorer.ts';
import { TransactionType } from '@/core/transaction.ts';
import type { Account } from '@/models/account.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionTag } from '@/models/transaction_tag.ts';
import {
    TransactionExplorerDestinationAccountCondition, TransactionExplorerSourceAccountCondition,
    TransactionExplorerTransactionCategoryCondition, TransactionExplorerTransactionTypeCondition,
} from './account-conditions.ts';
import { TransactionExplorerDestinationAmountCondition, TransactionExplorerSourceAmountCondition } from './amount-conditions.ts';
import { type InsightsExplorerMatchContext, type TransactionExplorerCondition, TransactionExplorerUndefinedCondition } from './condition-types.ts';
import { TransactionExplorerConditionWithRelation } from './condition.ts';
import {
    TransactionExplorerDescriptionCaseInsensitiveCondition, TransactionExplorerDescriptionCondition,
    TransactionExplorerDescriptionNormalizedCondition, TransactionExplorerPicturesCondition,
} from './description-conditions.ts';
import type { ExpressionNode } from './explorer.ts';
import { TransactionExplorerGeoLocationCondition } from './location-condition.ts';
import { TransactionExplorerTransactionTagCondition } from './tag-condition.ts';
import {
    TransactionExplorerTransactionTimeDayOfMonthCondition, TransactionExplorerTransactionTimeDayOfWeekCondition,
    TransactionExplorerTransactionTimeHourOfDayCondition, TransactionExplorerTransactionTimeMonthOfYearCondition,
    TransactionExplorerTransactionTimezoneCondition,
} from './time-conditions.ts';

export class TransactionExplorerQuery {
    public id: string;
    public name: string;
    public conditions: TransactionExplorerConditionWithRelation[];

    private constructor(id: string, name: string, conditions: TransactionExplorerConditionWithRelation[]) {
        this.id = id;
        this.name = name;
        this.conditions = conditions;
    }

    public addNewCondition(field: TransactionExplorerConditionField, isFirst: boolean): TransactionExplorerConditionWithRelation {
        let condition: TransactionExplorerCondition;

        switch (field) {
            case TransactionExplorerConditionField.TransactionTimeDayOfWeek:
                condition = new TransactionExplorerTransactionTimeDayOfWeekCondition(TransactionExplorerConditionOperatorType.In, []);
                break;
            case TransactionExplorerConditionField.TransactionTimeDayOfMonth:
                condition = new TransactionExplorerTransactionTimeDayOfMonthCondition(TransactionExplorerConditionOperatorType.In, []);
                break;
            case TransactionExplorerConditionField.TransactionTimeMonthOfYear:
                condition = new TransactionExplorerTransactionTimeMonthOfYearCondition(TransactionExplorerConditionOperatorType.In, []);
                break;
            case TransactionExplorerConditionField.TransactionTimeHourOfDay:
                condition = new TransactionExplorerTransactionTimeHourOfDayCondition(TransactionExplorerConditionOperatorType.In, []);
                break;
            case TransactionExplorerConditionField.TransactionTimezone:
                condition = new TransactionExplorerTransactionTimezoneCondition(TransactionExplorerConditionOperatorType.MinuteOffsetBetween, [ 0, 0 ]);
                break;
            case TransactionExplorerConditionField.TransactionType:
                condition = new TransactionExplorerTransactionTypeCondition(TransactionExplorerConditionOperatorType.In, [ TransactionType.Expense, TransactionType.Income, TransactionType.Transfer ]);
                break;
            case TransactionExplorerConditionField.TransactionCategory:
                condition = new TransactionExplorerTransactionCategoryCondition(TransactionExplorerConditionOperatorType.In, []);
                break;
            case TransactionExplorerConditionField.SourceAccount:
                condition = new TransactionExplorerSourceAccountCondition(TransactionExplorerConditionOperatorType.In, []);
                break;
            case TransactionExplorerConditionField.DestinationAccount:
                condition = new TransactionExplorerDestinationAccountCondition(TransactionExplorerConditionOperatorType.In, []);
                break;
            case TransactionExplorerConditionField.SourceAmount:
                condition = new TransactionExplorerSourceAmountCondition(TransactionExplorerConditionOperatorType.Between, [0, 0]);
                break;
            case TransactionExplorerConditionField.DestinationAmount:
                condition = new TransactionExplorerDestinationAmountCondition(TransactionExplorerConditionOperatorType.Between, [0, 0]);
                break;
            case TransactionExplorerConditionField.GeoLocation:
                condition = new TransactionExplorerGeoLocationCondition(TransactionExplorerConditionOperatorType.IsNotEmpty, [TransactionExplorerGeoLocationCondition.MIN_LATITUDE, TransactionExplorerGeoLocationCondition.MAX_LATITUDE, TransactionExplorerGeoLocationCondition.MIN_LONGITUDE, TransactionExplorerGeoLocationCondition.MAX_LONGITUDE]);
                break;
            case TransactionExplorerConditionField.TransactionTag:
                condition = new TransactionExplorerTransactionTagCondition(TransactionExplorerConditionOperatorType.HasAny, []);
                break;
            case TransactionExplorerConditionField.Pictures:
                condition = new TransactionExplorerPicturesCondition(TransactionExplorerConditionOperatorType.IsNotEmpty, []);
                break;
            case TransactionExplorerConditionField.Description:
                condition = new TransactionExplorerDescriptionCondition(TransactionExplorerConditionOperatorType.Contains, '');
                break;
            case TransactionExplorerConditionField.DescriptionCaseInsensitive:
                condition = new TransactionExplorerDescriptionCaseInsensitiveCondition(TransactionExplorerConditionOperatorType.Contains, '');
                break;
            case TransactionExplorerConditionField.DescriptionNormalized:
                condition = new TransactionExplorerDescriptionNormalizedCondition(TransactionExplorerConditionOperatorType.Contains, '');
                break;
            default:
                condition = new TransactionExplorerTransactionTypeCondition(TransactionExplorerConditionOperatorType.In, [ TransactionType.Expense, TransactionType.Income, TransactionType.Transfer ]);
                break;
        }

        return new TransactionExplorerConditionWithRelation(
            condition,
            isFirst ? TransactionExplorerConditionRelation.First : TransactionExplorerConditionRelation.And
        );
    }

    public addSubConditionEnd(): TransactionExplorerConditionWithRelation {
        return new TransactionExplorerConditionWithRelation(new TransactionExplorerUndefinedCondition(), TransactionExplorerConditionRelation.SubEnd);
    }

    public match(transaction: TransactionInsightDataItem, context: InsightsExplorerMatchContext): boolean {
        if (!this.conditions || this.conditions.length < 1) {
            return true;
        }

        const postfixExprTokens = this.getPostfixExprTokens();
        const stack: boolean[] = [];

        for (const token of postfixExprTokens) {
            if (token === TransactionExplorerConditionRelation.And || token === TransactionExplorerConditionRelation.Or) {
                const right = stack.pop();
                const left = stack.pop();

                if (left === undefined || right === undefined) {
                    throw new Error('invalid postfix expression');
                }

                if (token === TransactionExplorerConditionRelation.And) {
                    stack.push(left && right);
                } else if (token === TransactionExplorerConditionRelation.Or) {
                    stack.push(left || right);
                } else {
                    throw new Error('invalid postfix expression');
                }
            } else {
                stack.push(token.match(transaction, context));
            }
        }

        if (stack.length !== 1) {
            throw new Error('invalid postfix evaluation result');
        }

        return stack[0] as boolean;
    }

    public toExpression(allCategoriesMap: Record<string, TransactionCategory>, allAccountsMap: Record<string, Account>, allTagsMap: Record<string, TransactionTag>): string {
        if (!this.conditions || this.conditions.length < 1) {
            return '';
        }

        const postfixExprTokens = this.getPostfixExprTokens();
        const stack: ExpressionNode[] = [];

        for (const token of postfixExprTokens) {
            if (token === TransactionExplorerConditionRelation.And || token === TransactionExplorerConditionRelation.Or) {
                const right = stack.pop();
                const left = stack.pop();

                if (left === undefined || right === undefined) {
                    throw new Error('invalid postfix expression');
                }

                let leftExpression = left.textualExpression;
                let rightExpression = right.textualExpression;

                if (left.operator && left.operator !== token) {
                    leftExpression = `(${leftExpression})`;
                }

                if (right.operator && right.operator !== token) {
                    rightExpression = `(${rightExpression})`;
                }

                stack.push({
                    textualExpression: `${leftExpression} ${token.toUpperCase()} ${rightExpression}`,
                    operator: token
                });
            } else {
                stack.push({
                    textualExpression: token.toExpression(allCategoriesMap, allAccountsMap, allTagsMap)
                });
            }
        }

        if (stack.length !== 1) {
            throw new Error('invalid postfix evaluation result');
        }

        const finalNode = stack[0];

        if (!finalNode) {
            throw new Error('invalid postfix evaluation result');
        }

        return finalNode.textualExpression;
    }

    public getPostfixExprTokens(): (TransactionExplorerCondition | TransactionExplorerConditionRelation.And | TransactionExplorerConditionRelation.Or)[] {
        const finalTokens: (TransactionExplorerCondition | TransactionExplorerConditionRelation.And | TransactionExplorerConditionRelation.Or)[] = [];

        if (this.conditions.length < 1) {
            return finalTokens;
        }

        const operatorStack: (TransactionExplorerConditionRelation | TransactionExplorerSubConditionStartRelation)[] = [];
        const firstCondition = this.conditions[0] as TransactionExplorerConditionWithRelation;

        if (firstCondition.relation !== TransactionExplorerConditionRelation.First) {
            throw new Error('the first condition must have relation "first"');
        }

        finalTokens.push(firstCondition.condition);

        for (const [item, index] of itemAndIndex(this.conditions)) {
            if (index < 1) {
                continue;
            }

            if (item.relation === TransactionExplorerConditionRelation.First) {
                throw new Error('only the first condition can have relation "first"');
            }

            if (item.relation === TransactionExplorerConditionRelation.SubEnd) {
                while (operatorStack.length > 0) {
                    const topOperator = operatorStack.pop();

                    if (topOperator === TransactionExplorerSubConditionStartRelationPlaceholder) {
                        break;
                    }

                    const isAndOrOperator = topOperator === TransactionExplorerConditionRelation.And || topOperator === TransactionExplorerConditionRelation.Or;

                    if (isAndOrOperator) {
                        finalTokens.push(topOperator);
                    } else {
                        throw new Error('invalid operator in stack');
                    }
                }
            } else { // And, Or, AndSub, OrSub
                let currentOperator: TransactionExplorerConditionRelation.And | TransactionExplorerConditionRelation.Or;
                let startNewSubCondition = false;

                switch (item.relation) {
                    case TransactionExplorerConditionRelation.AndSub:
                        currentOperator = TransactionExplorerConditionRelation.And;
                        startNewSubCondition = true;
                        break;
                    case TransactionExplorerConditionRelation.OrSub:
                        currentOperator = TransactionExplorerConditionRelation.Or;
                        startNewSubCondition = true;
                        break;
                    case TransactionExplorerConditionRelation.And:
                        currentOperator = item.relation;
                        break;
                    case TransactionExplorerConditionRelation.Or:
                        currentOperator = item.relation;
                        break;
                    default:
                        throw new Error('invalid operator in stack');
                }

                while (operatorStack.length > 0) {
                    const topOperator = operatorStack[operatorStack.length - 1];

                    if (topOperator === TransactionExplorerSubConditionStartRelationPlaceholder) {
                        break;
                    }

                    const isAndOrOperator = topOperator === TransactionExplorerConditionRelation.And || topOperator === TransactionExplorerConditionRelation.Or;

                    if (isAndOrOperator && TransactionExplorerConditionRelationPriority[topOperator] >= TransactionExplorerConditionRelationPriority[currentOperator]) {
                        finalTokens.push(topOperator);
                        operatorStack.pop();
                    } else {
                        break;
                    }
                }

                operatorStack.push(currentOperator);

                if (startNewSubCondition) {
                    operatorStack.push(TransactionExplorerSubConditionStartRelationPlaceholder);
                }

                finalTokens.push(item.condition);
            }
        }

        while (operatorStack.length > 0) {
            const topOperator = operatorStack.pop();

            if (topOperator !== TransactionExplorerConditionRelation.And && topOperator !== TransactionExplorerConditionRelation.Or) {
                throw new Error('invalid operator in stack');
            }

            finalTokens.push(topOperator);
        }

        return finalTokens;
    }

    public getConditionNestingDepths(): number[] {
        const depths: number[] = [];
        let depth = 0;

        for (const item of this.conditions) {
            if (item.relation === TransactionExplorerConditionRelation.SubEnd) {
                depth--;
                depths.push(depth);
            } else if (item.relation === TransactionExplorerConditionRelation.AndSub || item.relation === TransactionExplorerConditionRelation.OrSub) {
                depth++;
                depths.push(depth);
            } else {
                depths.push(depth);
            }
        }

        return depths;
    }

    public clone(newId: string): TransactionExplorerQuery {
        const clonedConditions: TransactionExplorerConditionWithRelation[] = [];

        for (const condition of this.conditions) {
            const clonedCondition = TransactionExplorerConditionWithRelation.parse(condition.toJsonObject());

            if (!clonedCondition) {
                continue;
            }

            clonedConditions.push(clonedCondition);
        }

        return new TransactionExplorerQuery(newId, this.name, clonedConditions);
    }

    public toJsonObject(): object {
        return {
            id: this.id,
            name: this.name,
            conditions: this.conditions.map(condition => condition.toJsonObject())
        };
    }

    public static create(id: string): TransactionExplorerQuery {
        return new TransactionExplorerQuery(id, '', []);
    }

    public static parse(obj: object): TransactionExplorerQuery | null {
        if (!('id' in obj) || !('name' in obj) || !('conditions' in obj)) {
            return null;
        }

        const idFieldValue = obj['id'] as unknown;
        const nameFieldValue = obj['name'] as unknown;
        const conditionsFieldValue = obj['conditions'] as unknown;

        if (typeof idFieldValue !== 'string' || typeof nameFieldValue !== 'string' || !Array.isArray(conditionsFieldValue)) {
            return null;
        }

        const id: string = idFieldValue;
        const name: string = nameFieldValue;
        const conditions: TransactionExplorerConditionWithRelation[] = [];
        let conditionDepth = 0;

        for (const [item, index] of itemAndIndex(conditionsFieldValue)) {
            const condition = TransactionExplorerConditionWithRelation.parse(item);

            if (condition === null) {
                return null;
            }

            if (index === 0 && condition.relation !== TransactionExplorerConditionRelation.First) {
                return null;
            } else if (index > 0 && condition.relation === TransactionExplorerConditionRelation.First) {
                return null;
            }

            if (condition.relation === TransactionExplorerConditionRelation.AndSub ||
                condition.relation === TransactionExplorerConditionRelation.OrSub) {
                conditionDepth++;
            } else if (condition.relation === TransactionExplorerConditionRelation.SubEnd) {
                conditionDepth--;
            }

            conditions.push(condition);
        }

        if (conditionDepth !== 0) {
            return null; // unbalanced parentheses
        }

        return new TransactionExplorerQuery(id, name, conditions);
    }
}
