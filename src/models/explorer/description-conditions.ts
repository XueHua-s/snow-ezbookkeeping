import type { PartialRecord } from '@/core/base.ts';
import {
    TransactionExplorerConditionFieldType,
    TransactionExplorerConditionOperatorType
} from '@/core/explorer.ts';
import { NormalizedText } from '@/core/text.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { InsightsExplorerMatchContext, TransactionExplorerCondition } from './condition-types.ts';

export type PicturesConditionOperator = TransactionExplorerConditionOperatorType.IsEmpty |
    TransactionExplorerConditionOperatorType.IsNotEmpty;

export class TransactionExplorerPicturesCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.Pictures, string[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.IsEmpty]: true,
        [TransactionExplorerConditionOperatorType.IsNotEmpty]: true
    };

    public readonly field = TransactionExplorerConditionFieldType.Pictures;
    public readonly operator: PicturesConditionOperator = TransactionExplorerConditionOperatorType.IsNotEmpty;
    public value: string[];

    constructor(operator: PicturesConditionOperator, value: string[]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): string[] {
        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty || this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return [];
        }

        return [];
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty) {
            return !transaction.pictures || transaction.pictures.length < 1;
        } else if (this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return !!transaction.pictures && transaction.pictures.length > 0;
        }

        return false;
    }

    public toExpression(): string {
        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty) {
            return `pictures IS EMPTY`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return `pictures IS NOT EMPTY`;
        }

        return '';
    }
}

export type DescriptionConditionField = TransactionExplorerConditionFieldType.Description | TransactionExplorerConditionFieldType.DescriptionCaseInsensitive | TransactionExplorerConditionFieldType.DescriptionNormalized;

export type DescriptionConditionOperator = TransactionExplorerConditionOperatorType.IsEmpty |
    TransactionExplorerConditionOperatorType.IsNotEmpty |
    TransactionExplorerConditionOperatorType.Equals |
    TransactionExplorerConditionOperatorType.NotEquals |
    TransactionExplorerConditionOperatorType.Contains |
    TransactionExplorerConditionOperatorType.NotContains |
    TransactionExplorerConditionOperatorType.StartsWith |
    TransactionExplorerConditionOperatorType.NotStartsWith |
    TransactionExplorerConditionOperatorType.EndsWith |
    TransactionExplorerConditionOperatorType.NotEndsWith |
    TransactionExplorerConditionOperatorType.RegexMatch |
    TransactionExplorerConditionOperatorType.NotRegexMatch;

export enum DescriptionMatchMode {
    Exact,
    CaseInsensitive,
    Normalized
}

export abstract class AbstractTransactionExplorerDescriptionCondition<T = DescriptionConditionField> implements TransactionExplorerCondition<T, string> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.IsEmpty]: true,
        [TransactionExplorerConditionOperatorType.IsNotEmpty]: true,
        [TransactionExplorerConditionOperatorType.Equals]: true,
        [TransactionExplorerConditionOperatorType.NotEquals]: true,
        [TransactionExplorerConditionOperatorType.Contains]: true,
        [TransactionExplorerConditionOperatorType.NotContains]: true,
        [TransactionExplorerConditionOperatorType.StartsWith]: true,
        [TransactionExplorerConditionOperatorType.NotStartsWith]: true,
        [TransactionExplorerConditionOperatorType.EndsWith]: true,
        [TransactionExplorerConditionOperatorType.NotEndsWith]: true,
        [TransactionExplorerConditionOperatorType.RegexMatch]: true,
        [TransactionExplorerConditionOperatorType.NotRegexMatch]: true
    };

    public abstract readonly field: T;
    public readonly operator: DescriptionConditionOperator = TransactionExplorerConditionOperatorType.Contains;
    public value: string;
    protected abstract readonly matchMode: DescriptionMatchMode;
    private cachedNormalizedValue: NormalizedText | undefined = undefined;
    private cachedRegex: RegExp | undefined = undefined;

    constructor(operator: DescriptionConditionOperator, value: string) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): string {
        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty || this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return '';
        }

        return this.value;
    }

    public match(transaction: TransactionInsightDataItem, context: InsightsExplorerMatchContext): boolean {
        let description: string;

        if (!transaction.comment) {
            description = '';
        } else if (this.matchMode === DescriptionMatchMode.Normalized) {
            description = context.getNormalizedDescription().normalizedText;
        } else if (this.matchMode === DescriptionMatchMode.CaseInsensitive) {
            description = context.getNormalizedDescription().lowerCaseText;
        } else {
            description = transaction.comment;
        }

        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty) {
            return description.length === 0;
        } else if (this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return description.length > 0;
        } else if (this.operator === TransactionExplorerConditionOperatorType.Equals) {
            return description === this.getCachedNormalizedValue();
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotEquals) {
            return description !== this.getCachedNormalizedValue();
        } else if (this.operator === TransactionExplorerConditionOperatorType.Contains) {
            return description.includes(this.getCachedNormalizedValue());
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotContains) {
            return !description.includes(this.getCachedNormalizedValue());
        } else if (this.operator === TransactionExplorerConditionOperatorType.StartsWith) {
            return description.startsWith(this.getCachedNormalizedValue());
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotStartsWith) {
            return !description.startsWith(this.getCachedNormalizedValue());
        } else if (this.operator === TransactionExplorerConditionOperatorType.EndsWith) {
            return description.endsWith(this.getCachedNormalizedValue());
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotEndsWith) {
            return !description.endsWith(this.getCachedNormalizedValue());
        } else if (this.operator === TransactionExplorerConditionOperatorType.RegexMatch) {
            return this.getCachedRegex()?.test(description) ?? false;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotRegexMatch) {
            return !(this.getCachedRegex()?.test(description) ?? false);
        }

        return false;
    }

    public toExpression(): string {
        const expression: string = this.getExpression();

        if (expression && this.matchMode === DescriptionMatchMode.Normalized) {
            return `${expression} NORMALIZED`;
        } else if (expression && this.matchMode === DescriptionMatchMode.CaseInsensitive) {
            return `${expression} CASE INSENSITIVE`;
        } else {
            return expression;
        }
    }

    private getExpression(): string {
        const escapedValue = this.value.replace(/'/g, "''");

        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty) {
            return `description IS EMPTY`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return `description IS NOT EMPTY`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.Equals) {
            return `description = '${escapedValue}'`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotEquals) {
            return `description <> '${escapedValue}'`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.Contains) {
            return `description CONTAINS '${escapedValue}'`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotContains) {
            return `description NOT CONTAINS '${escapedValue}'`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.StartsWith) {
            return `description STARTS WITH '${escapedValue}'`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotStartsWith) {
            return `description NOT STARTS WITH '${escapedValue}'`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.EndsWith) {
            return `description ENDS WITH '${escapedValue}'`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotEndsWith) {
            return `description NOT ENDS WITH '${escapedValue}'`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.RegexMatch) {
            return `description REGEXP LIKE '${escapedValue}'`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotRegexMatch) {
            return `description NOT REGEXP LIKE '${escapedValue}'`;
        }

        return '';
    }

    private getCachedNormalizedValue(): string {
        let normalizedValue = this.cachedNormalizedValue;

        if (!normalizedValue || normalizedValue.originalText !== this.value) {
            normalizedValue = NormalizedText.of(this.value);
            this.cachedNormalizedValue = normalizedValue;
            this.cachedRegex = undefined;
        }

        if (this.matchMode === DescriptionMatchMode.Normalized) {
            return normalizedValue.normalizedText;
        } else if (this.matchMode === DescriptionMatchMode.CaseInsensitive) {
            return normalizedValue.lowerCaseText;
        } else {
            return normalizedValue.originalText;
        }
    }

    private getCachedRegex(): RegExp | undefined {
        if (this.operator !== TransactionExplorerConditionOperatorType.RegexMatch && this.operator !== TransactionExplorerConditionOperatorType.NotRegexMatch) {
            return undefined;
        }

        if (!this.cachedNormalizedValue || this.cachedNormalizedValue.originalText !== this.value || !this.cachedRegex) {
            try {
                let regex: RegExp;

                if (this.matchMode === DescriptionMatchMode.CaseInsensitive || this.matchMode === DescriptionMatchMode.Normalized) {
                    regex = new RegExp(this.getCachedNormalizedValue(), 'i');
                } else {
                    regex = new RegExp(this.getCachedNormalizedValue());
                }

                this.cachedRegex = regex;
            } catch {
                this.cachedRegex = undefined;
            }
        }

        return this.cachedRegex;
    }
}

export class TransactionExplorerDescriptionCondition extends AbstractTransactionExplorerDescriptionCondition<TransactionExplorerConditionFieldType.Description> {
    public readonly field = TransactionExplorerConditionFieldType.Description;
    protected readonly matchMode = DescriptionMatchMode.Exact;

    constructor(operator: DescriptionConditionOperator, value: string) {
        super(operator, value);
    }
}

export class TransactionExplorerDescriptionCaseInsensitiveCondition extends AbstractTransactionExplorerDescriptionCondition<TransactionExplorerConditionFieldType.DescriptionCaseInsensitive> {
    public readonly field = TransactionExplorerConditionFieldType.DescriptionCaseInsensitive;
    protected readonly matchMode = DescriptionMatchMode.CaseInsensitive;

    constructor(operator: DescriptionConditionOperator, value: string) {
        super(operator, value);
    }
}

export class TransactionExplorerDescriptionNormalizedCondition extends AbstractTransactionExplorerDescriptionCondition<TransactionExplorerConditionFieldType.DescriptionNormalized> {
    public readonly field = TransactionExplorerConditionFieldType.DescriptionNormalized;
    protected readonly matchMode = DescriptionMatchMode.Normalized;

    constructor(operator: DescriptionConditionOperator, value: string) {
        super(operator, value);
    }
}
