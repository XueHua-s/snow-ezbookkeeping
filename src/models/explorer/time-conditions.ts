import type { PartialRecord } from '@/core/base.ts';
import {
    TransactionExplorerConditionFieldType,
    TransactionExplorerConditionOperatorType
} from '@/core/explorer.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { InsightsExplorerMatchContext, TransactionExplorerCondition } from './condition-types.ts';

export type TransactionTimeDayOfWeekConditionOperator = TransactionExplorerConditionOperatorType.In |
    TransactionExplorerConditionOperatorType.NotIn;

export class TransactionExplorerTransactionTimeDayOfWeekCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.TransactionTimeDayOfWeek, number[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.In]: true,
        [TransactionExplorerConditionOperatorType.NotIn]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.TransactionTimeDayOfWeek;
    public readonly operator: TransactionTimeDayOfWeekConditionOperator = TransactionExplorerConditionOperatorType.In;
    public value: number[];

    constructor(operator: TransactionTimeDayOfWeekConditionOperator, value: number[]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): number[] {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem, context: InsightsExplorerMatchContext): boolean {
        const transactionDateTime = context.getTransactionDateTime();

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return this.value.includes(transactionDateTime.getWeekDay().type);
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return !this.value.includes(transactionDateTime.getWeekDay().type);
        }

        return false;
    }

    public toExpression(): string {
        const textualDayOfWeeks = this.value.join(', ');

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return `DAY_OF_WEEK(transaction_time) IN (${textualDayOfWeeks})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return `DAY_OF_WEEK(transaction_time) NOT IN (${textualDayOfWeeks})`;
        } else {
            return '';
        }
    }
}

export type TransactionTimeDayOfMonthConditionOperator = TransactionExplorerConditionOperatorType.In |
    TransactionExplorerConditionOperatorType.NotIn;

export class TransactionExplorerTransactionTimeDayOfMonthCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.TransactionTimeDayOfMonth, number[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.In]: true,
        [TransactionExplorerConditionOperatorType.NotIn]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.TransactionTimeDayOfMonth;
    public readonly operator: TransactionTimeDayOfMonthConditionOperator = TransactionExplorerConditionOperatorType.In;
    public value: number[];

    constructor(operator: TransactionTimeDayOfMonthConditionOperator, value: number[]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): number[] {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem, context: InsightsExplorerMatchContext): boolean {
        const transactionDateTime = context.getTransactionDateTime();
        const normalizedValue: number[] = this.value.map(day => day >= 0 ? day : transactionDateTime.getMaxDayOfGregorianCalendarMonth() + day + 1);

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return normalizedValue.includes(transactionDateTime.getGregorianCalendarDay());
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return !normalizedValue.includes(transactionDateTime.getGregorianCalendarDay());
        }

        return false;
    }

    public toExpression(): string {
        const textualDayOfMonths = this.value.join(', ');

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return `DAY(transaction_time) IN (${textualDayOfMonths})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return `DAY(transaction_time) NOT IN (${textualDayOfMonths})`;
        } else {
            return '';
        }
    }
}

export type TransactionTimeMonthOfYearConditionOperator = TransactionExplorerConditionOperatorType.In |
    TransactionExplorerConditionOperatorType.NotIn;

export class TransactionExplorerTransactionTimeMonthOfYearCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.TransactionTimeMonthOfYear, number[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.In]: true,
        [TransactionExplorerConditionOperatorType.NotIn]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.TransactionTimeMonthOfYear;
    public readonly operator: TransactionTimeMonthOfYearConditionOperator = TransactionExplorerConditionOperatorType.In;
    public value: number[];

    constructor(operator: TransactionTimeMonthOfYearConditionOperator, value: number[]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): number[] {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem, context: InsightsExplorerMatchContext): boolean {
        const transactionDateTime = context.getTransactionDateTime();

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return this.value.includes(transactionDateTime.getGregorianCalendarMonth());
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return !this.value.includes(transactionDateTime.getGregorianCalendarMonth());
        }

        return false;
    }

    public toExpression(): string {
        const textualMonthOfYears = this.value.join(', ');

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return `MONTH(transaction_time) IN (${textualMonthOfYears})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return `MONTH(transaction_time) NOT IN (${textualMonthOfYears})`;
        } else {
            return '';
        }
    }
}

export type TransactionTimeHourOfDayConditionOperator = TransactionExplorerConditionOperatorType.In |
    TransactionExplorerConditionOperatorType.NotIn;

export class TransactionExplorerTransactionTimeHourOfDayCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.TransactionTimeHourOfDay, number[]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.In]: true,
        [TransactionExplorerConditionOperatorType.NotIn]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.TransactionTimeHourOfDay;
    public readonly operator: TransactionTimeHourOfDayConditionOperator = TransactionExplorerConditionOperatorType.In;
    public value: number[];

    constructor(operator: TransactionTimeHourOfDayConditionOperator, value: number[]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): number[] {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem, context: InsightsExplorerMatchContext): boolean {
        const transactionDateTime = context.getTransactionDateTime();

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return this.value.includes(transactionDateTime.getHour());
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return !this.value.includes(transactionDateTime.getHour());
        }

        return false;
    }

    public toExpression(): string {
        const textualHourOfDays = this.value.join(', ');

        if (this.operator === TransactionExplorerConditionOperatorType.In) {
            return `HOUR(transaction_time) IN (${textualHourOfDays})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.NotIn) {
            return `HOUR(transaction_time) NOT IN (${textualHourOfDays})`;
        } else {
            return '';
        }
    }
}

export type TransactionTimezoneConditionOperator = TransactionExplorerConditionOperatorType.MinuteOffsetBetween |
    TransactionExplorerConditionOperatorType.MinuteOffsetNotBetween;

export class TransactionExplorerTransactionTimezoneCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.TransactionTimezone, [number, number]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.MinuteOffsetBetween]: true,
        [TransactionExplorerConditionOperatorType.MinuteOffsetNotBetween]: true
    };
    public readonly field = TransactionExplorerConditionFieldType.TransactionTimezone;
    public readonly operator: TransactionTimezoneConditionOperator = TransactionExplorerConditionOperatorType.MinuteOffsetBetween;
    public value: [number, number];

    constructor(operator: TransactionTimezoneConditionOperator, value: [number, number]) {
        this.operator = operator;
        this.value = value;
    }

    public getValueForStore(): [number, number] {
        return this.value;
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        if (this.operator === TransactionExplorerConditionOperatorType.MinuteOffsetBetween) {
            return transaction.utcOffset >= this.value[0] && transaction.utcOffset <= this.value[1];
        } else if (this.operator === TransactionExplorerConditionOperatorType.MinuteOffsetNotBetween) {
            return transaction.utcOffset < this.value[0] || transaction.utcOffset > this.value[1];
        }

        return false;
    }

    public toExpression(): string {
        if (this.operator === TransactionExplorerConditionOperatorType.MinuteOffsetBetween) {
            return `(UTC_OFFSET(timezone) >= ${this.value[0]} AND UTC_OFFSET(timezone) <= ${this.value[1]})`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.MinuteOffsetNotBetween) {
            return `(UTC_OFFSET(timezone) < ${this.value[0]} OR UTC_OFFSET(timezone) > ${this.value[1]})`;
        }

        return '';
    }
}
