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

export type TransactionTagConditionOperator = TransactionExplorerConditionOperatorType.IsEmpty |
    TransactionExplorerConditionOperatorType.IsNotEmpty |
    TransactionExplorerConditionOperatorType.Equals |
    TransactionExplorerConditionOperatorType.NotEquals |
    TransactionExplorerConditionOperatorType.HasAny |
    TransactionExplorerConditionOperatorType.HasAll |
    TransactionExplorerConditionOperatorType.NotHasAny |
    TransactionExplorerConditionOperatorType.NotHasAll;

export class TransactionExplorerTransactionTagCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.TransactionTag, string[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.IsEmpty]: true,
        [TransactionExplorerConditionOperatorType.IsNotEmpty]: true,
        [TransactionExplorerConditionOperatorType.Equals]: true,
        [TransactionExplorerConditionOperatorType.NotEquals]: true,
        [TransactionExplorerConditionOperatorType.HasAny]: true,
        [TransactionExplorerConditionOperatorType.HasAll]: true,
        [TransactionExplorerConditionOperatorType.NotHasAny]: true,
        [TransactionExplorerConditionOperatorType.NotHasAll]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.TransactionTag;
    public readonly operator: TransactionTagConditionOperator = TransactionExplorerConditionOperatorType.HasAny;
    public value: string[];

    constructor(operator: TransactionTagConditionOperator, value: string[]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): string[] {
        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty || this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return [];
        }

        return this.value;
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        const transactionTags: Record<string, boolean> = {};

        for (const tagId of transaction.tagIds) {
            transactionTags[tagId] = true;
        }

        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty || this.value.length < 1) {
            return transaction.tagIds.length < 1;
        } else if (this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return transaction.tagIds.length > 0;
        } else if (this.operator === TransactionExplorerConditionOperatorType.Equals || this.operator === TransactionExplorerConditionOperatorType.NotEquals) {
            let hasAll = true;

            for (const tagId of this.value) {
                if (!transactionTags[tagId]) {
                    hasAll = false;
                    break;
                }
            }

            const hasSameCount = transaction.tagIds.length === this.value.length;

            if (this.operator === TransactionExplorerConditionOperatorType.Equals && hasAll && hasSameCount) {
                return true;
            } else if (this.operator === TransactionExplorerConditionOperatorType.NotEquals && (!hasAll || !hasSameCount)) {
                return true;
            }
        } else if (this.operator === TransactionExplorerConditionOperatorType.HasAny || this.operator === TransactionExplorerConditionOperatorType.NotHasAny) {
            let hasAny = false;

            for (const tagId of this.value) {
                if (transactionTags[tagId]) {
                    hasAny = true;
                    break;
                }
            }

            if (this.operator === TransactionExplorerConditionOperatorType.HasAny && hasAny) {
                return true;
            } else if (this.operator === TransactionExplorerConditionOperatorType.NotHasAny && !hasAny) {
                return true;
            }
        } else if (this.operator === TransactionExplorerConditionOperatorType.HasAll || this.operator === TransactionExplorerConditionOperatorType.NotHasAll) {
            let hasAll = true;

            for (const tagId of this.value) {
                if (!transactionTags[tagId]) {
                    hasAll = false;
                    break;
                }
            }

            if (this.operator === TransactionExplorerConditionOperatorType.HasAll && hasAll) {
                return true;
            } else if (this.operator === TransactionExplorerConditionOperatorType.NotHasAll && !hasAll) {
                return true;
            }
        }

        return false;
    }

    public toExpression(allCategoriesMap: Record<string, TransactionCategory>, allAccountsMap: Record<string, Account>, allTagsMap: Record<string, TransactionTag>): string {
        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty) {
            return `tags IS EMPTY`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return `tags IS NOT EMPTY`;
        }

        const textualTags = this.value.map(id => {
            const tag = allTagsMap[id];

            if (tag) {
                return `'${tag.name}'`;
            } else {
                return `'${id}'`;
            }
        }).join(', ');

        if (this.operator === TransactionExplorerConditionOperatorType.Equals) {
            return `tags FULL MATCHES (${textualTags})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotEquals) {
            return `tags NOT FULL MATCHES (${textualTags})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.HasAny) {
            return `tags HAS ANY (${textualTags})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.HasAll) {
            return `tags HAS ALL (${textualTags})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotHasAny) {
            return `tags NOT HAS ANY (${textualTags})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotHasAll) {
            return `tags NOT HAS ALL (${textualTags})`;
        } else {
            return '';
        }
    }
}
