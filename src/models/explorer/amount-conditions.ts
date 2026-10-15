import type { PartialRecord } from '@/core/base.ts';
import {
    TransactionExplorerConditionFieldType,
    TransactionExplorerConditionOperatorType
} from '@/core/explorer.ts';
import type { Account } from '@/models/account.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionTag } from '@/models/transaction_tag.ts';
import type { TransactionExplorerCondition } from './condition-types.ts';

export type AmountConditionField = TransactionExplorerConditionFieldType.SourceAmount | TransactionExplorerConditionFieldType.DestinationAmount;

export type AmountConditionOperator = TransactionExplorerConditionOperatorType.Equals |
    TransactionExplorerConditionOperatorType.NotEquals |
    TransactionExplorerConditionOperatorType.GreaterThan |
    TransactionExplorerConditionOperatorType.LessThan |
    TransactionExplorerConditionOperatorType.Between |
    TransactionExplorerConditionOperatorType.NotBetween;

export abstract class AbstractTransactionExplorerAmountCondition<T = AmountConditionField> implements TransactionExplorerCondition<T, [number, number]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.Equals]: true,
        [TransactionExplorerConditionOperatorType.NotEquals]: true,
        [TransactionExplorerConditionOperatorType.GreaterThan]: true,
        [TransactionExplorerConditionOperatorType.LessThan]: true,
        [TransactionExplorerConditionOperatorType.Between]: true,
        [TransactionExplorerConditionOperatorType.NotBetween]: true
    };
    public abstract readonly field: T;
    public readonly operator: AmountConditionOperator = TransactionExplorerConditionOperatorType.Between;
    public value: [number, number];

    protected constructor(operator: AmountConditionOperator, value: [number, number]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): [number, number] {
        if (this.operator === TransactionExplorerConditionOperatorType.Between || this.operator === TransactionExplorerConditionOperatorType.NotBetween) {
            return [this.value[0], this.value[1]];
        } else {
            return [this.value[0], this.value[0]];
        }
    }

    public abstract match(transaction: TransactionInsightDataItem): boolean;
    public abstract toExpression(allCategoriesMap: Record<string, TransactionCategory>, allAccountsMap: Record<string, Account>, allTagsMap: Record<string, TransactionTag>): string;

    protected matchAmount(amount: number): boolean {
        switch (this.operator) {
            case TransactionExplorerConditionOperatorType.GreaterThan:
                return amount > this.value[0];
            case TransactionExplorerConditionOperatorType.LessThan:
                return amount < this.value[0];
            case TransactionExplorerConditionOperatorType.Equals:
                return amount === this.value[0];
            case TransactionExplorerConditionOperatorType.NotEquals:
                return amount !== this.value[0];
            case TransactionExplorerConditionOperatorType.Between:
                return amount >= this.value[0] && amount <= this.value[1];
            case TransactionExplorerConditionOperatorType.NotBetween:
                return amount < this.value[0] || amount > this.value[1];
            default:
                return false;
        }
    }

    protected getExpression(amountFieldName: string): string {
        let expressionAmount1 = this.value[0].toString(10);
        let expressionAmount2 = this.value[1].toString(10);

        if (expressionAmount1.length > 2) {
            expressionAmount1 = `${expressionAmount1.substring(0, expressionAmount1.length - 2)}.${expressionAmount1.substring(expressionAmount1.length - 2)}`;
        }

        if (expressionAmount2.length > 2) {
            expressionAmount2 = `${expressionAmount2.substring(0, expressionAmount2.length - 2)}.${expressionAmount2.substring(expressionAmount2.length - 2)}`;
        }

        switch (this.operator) {
            case TransactionExplorerConditionOperatorType.GreaterThan:
                return `${amountFieldName} > ${expressionAmount1}`;
            case TransactionExplorerConditionOperatorType.LessThan:
                return `${amountFieldName} < ${expressionAmount1}`;
            case TransactionExplorerConditionOperatorType.Equals:
                return `${amountFieldName} = ${expressionAmount1}`;
            case TransactionExplorerConditionOperatorType.NotEquals:
                return `${amountFieldName} <> ${expressionAmount1}`;
            case TransactionExplorerConditionOperatorType.Between:
                return `(${amountFieldName} >= ${expressionAmount1} AND ${amountFieldName} <= ${expressionAmount2})`;
            case TransactionExplorerConditionOperatorType.NotBetween:
                return `(${amountFieldName} < ${expressionAmount1} OR ${amountFieldName} > ${expressionAmount2})`;
            default:
                return '';
        }
    }
}

export class TransactionExplorerSourceAmountCondition extends AbstractTransactionExplorerAmountCondition<TransactionExplorerConditionFieldType.SourceAmount> {
    public readonly field = TransactionExplorerConditionFieldType.SourceAmount;

    constructor(operator: AmountConditionOperator, value: [number, number]) {
        super(operator, value);
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        return super.matchAmount(transaction.sourceAmount);
    }

    public toExpression(): string {
        return this.getExpression('source_amount');
    }
}

export class TransactionExplorerDestinationAmountCondition extends AbstractTransactionExplorerAmountCondition<TransactionExplorerConditionFieldType.DestinationAmount> {
    public readonly field = TransactionExplorerConditionFieldType.DestinationAmount;

    constructor(operator: AmountConditionOperator, value: [number, number]) {
        super(operator, value);
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        return super.matchAmount(transaction.destinationAmount);
    }

    public toExpression(): string {
        return this.getExpression('destination_amount');
    }
}
