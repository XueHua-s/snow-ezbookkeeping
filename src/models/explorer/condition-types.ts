import type { DateTime } from '@/core/datetime.ts';
import {
    TransactionExplorerConditionFieldType,
    TransactionExplorerConditionOperatorType
} from '@/core/explorer.ts';
import type { NormalizedText } from '@/core/text.ts';
import type { Account } from '@/models/account.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionTag } from '@/models/transaction_tag.ts';


export interface InsightsExplorerMatchContext {
    getTransactionDateTime(): DateTime;
    getNormalizedDescription(): NormalizedText;
}

export interface TransactionExplorerCondition<T = TransactionExplorerConditionFieldType, V = string | string[] | number[]> {
    readonly field: T;
    readonly operator: TransactionExplorerConditionOperatorType;
    value: V;

    getValueForStore(): V;
    match(transaction: TransactionInsightDataItem, context: InsightsExplorerMatchContext): boolean;
    toExpression(allCategoriesMap: Record<string, TransactionCategory>, allAccountsMap: Record<string, Account>, allTagsMap: Record<string, TransactionTag>): string;
}

export class TransactionExplorerUndefinedCondition implements TransactionExplorerCondition {
    public readonly field = TransactionExplorerConditionFieldType.Undefined;
    public readonly operator = TransactionExplorerConditionOperatorType.Equals;
    public value = '';

    public getValueForStore(): string {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        return !!transaction;
    }

    public toExpression(): string {
        return '';
    }
}
