import { AccountType } from '@/core/account.ts';
import type { PartialRecord } from '@/core/base.ts';
import {
    TransactionExplorerConditionFieldType,
    TransactionExplorerConditionOperatorType
} from '@/core/explorer.ts';
import { TransactionType } from '@/core/transaction.ts';
import type { Account } from '@/models/account.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionExplorerCondition } from './condition-types.ts';

export type TransactionTypeConditionOperator = TransactionExplorerConditionOperatorType.In |
    TransactionExplorerConditionOperatorType.NotIn;

export class TransactionExplorerTransactionTypeCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.TransactionType, number[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.In]: true,
        [TransactionExplorerConditionOperatorType.NotIn]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.TransactionType;
    public readonly operator: TransactionTypeConditionOperator = TransactionExplorerConditionOperatorType.In;
    public value: number[];

    constructor(operator: TransactionTypeConditionOperator, value: number[]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): number[] {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return this.value.includes(transaction.type);
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return !this.value.includes(transaction.type);
        }

        return false;
    }

    public toExpression(): string {
        const textualTypes = this.value.map(type => {
            if (type === TransactionType.Income) {
                return 'Income';
            } else if (type === TransactionType.Expense) {
                return 'Expense';
            } else if (type === TransactionType.Transfer) {
                return 'Transfer';
            } else {
                return type.toString();
            }
        }).join(', ');

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return `type IN (${textualTypes})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return `type NOT IN (${textualTypes})`;
        } else {
            return '';
        }
    }
}

export type TransactionCategoryConditionOperator = TransactionExplorerConditionOperatorType.In |
    TransactionExplorerConditionOperatorType.NotIn;

export class TransactionExplorerTransactionCategoryCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.TransactionCategory, string[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.In]: true,
        [TransactionExplorerConditionOperatorType.NotIn]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.TransactionCategory;
    public readonly operator: TransactionCategoryConditionOperator = TransactionExplorerConditionOperatorType.In;
    public value: string[];

    constructor(operator: TransactionCategoryConditionOperator, value: string[]) {
        this.operator = operator
        this.value = value;
    }

    public getValueForStore(): string[] {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return this.value.includes(transaction.primaryCategory?.id ?? '') || this.value.includes(transaction.secondaryCategory?.id ?? transaction.categoryId);
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return !this.value.includes(transaction.primaryCategory?.id ?? '') && !this.value.includes(transaction.secondaryCategory?.id ?? transaction.categoryId);
        }

        return false;
    }

    public toExpression(allCategoriesMap: Record<string, TransactionCategory>): string {
        const textualCategories = this.value.map(id => {
            const category = allCategoriesMap[id];

            if (category) {
                if (!category.parentId || category.parentId === '0') {
                    return '';
                }

                return `'${category.name}'`;
            } else {
                return `'${id}'`;
            }
        }).filter(item => !!item).join(', ');

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return `category IN (${textualCategories})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return `category NOT IN (${textualCategories})`;
        } else {
            return '';
        }
    }
}

export type AccountConditionOperator = TransactionExplorerConditionOperatorType.In |
    TransactionExplorerConditionOperatorType.NotIn;

export class TransactionExplorerSourceAccountCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.SourceAccount, string[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.In]: true,
        [TransactionExplorerConditionOperatorType.NotIn]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.SourceAccount;
    public readonly operator: AccountConditionOperator = TransactionExplorerConditionOperatorType.In;
    public value: string[];

    constructor(operator: AccountConditionOperator, value: string[]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): string[] {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return this.value.includes(transaction.sourceAccountId);
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return !this.value.includes(transaction.sourceAccountId);
        }

        return false;
    }

    public toExpression(allCategoriesMap: Record<string, TransactionCategory>, allAccountsMap: Record<string, Account>): string {
        const textualAccounts = this.value.map(id => {
            const account = allAccountsMap[id];

            if (account) {
                if (account.type === AccountType.MultiSubAccounts.type) {
                    return '';
                }

                return `'${account.name}'`;
            } else {
                return `'${id}'`;
            }
        }).filter(item => !!item).join(', ');

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return `source_account IN (${textualAccounts})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return `source_account NOT IN (${textualAccounts})`;
        } else {
            return '';
        }
    }
}

export class TransactionExplorerDestinationAccountCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.DestinationAccount, string[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.In]: true,
        [TransactionExplorerConditionOperatorType.NotIn]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.DestinationAccount;
    public readonly operator: AccountConditionOperator = TransactionExplorerConditionOperatorType.In;
    public value: string[];

    constructor(operator: AccountConditionOperator, value: string[]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): string[] {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return this.value.includes(transaction.destinationAccountId);
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return !this.value.includes(transaction.destinationAccountId);
        }

        return false;
    }

    public toExpression(allCategoriesMap: Record<string, TransactionCategory>, allAccountsMap: Record<string, Account>): string {
        const textualAccounts = this.value.map(id => {
            const account = allAccountsMap[id];

            if (account) {
                if (account.type === AccountType.MultiSubAccounts.type) {
                    return '';
                }

                return `'${account.name}'`;
            } else {
                return `'${id}'`;
            }
        }).filter(item => !!item).join(', ');

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return `destination_account IN (${textualAccounts})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return `destination_account NOT IN (${textualAccounts})`;
        } else {
            return '';
        }
    }
}
