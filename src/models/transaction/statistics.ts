import type { PartialRecord } from '@/core/base.ts';
import type { StartEndTime, Year1BasedMonth, YearMonthDay } from '@/core/datetime.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import type { Account } from '@/models/account.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionInfoResponse } from './requests.ts';
import type { Transaction } from './transaction.ts';

export interface TransactionStatisticRequest {
    readonly startTime: number;
    readonly endTime: number;
    readonly tagFilter: string;
    readonly keyword: string;
    readonly matchMode: number;
    readonly useTransactionTimezone: boolean;
}

export interface YearMonthRangeRequest {
    readonly startYearMonth: string;
    readonly endYearMonth: string;
}

export interface TransactionStatisticTrendsRequest extends YearMonthRangeRequest {
    readonly tagFilter: string;
    readonly keyword: string;
    readonly matchMode: number;
    readonly useTransactionTimezone: boolean;
}

export interface TransactionStatisticAssetTrendsRequest {
    readonly startTime: number;
    readonly endTime: number;
}

export const ALL_TRANSACTION_AMOUNTS_REQUEST_TYPE = [
    'today',
    'thisWeek',
    'thisMonth',
    'thisYear',
    'lastMonth',
    'monthBeforeLastMonth',
    'monthBeforeLast2Months',
    'monthBeforeLast3Months',
    'monthBeforeLast4Months',
    'monthBeforeLast5Months',
    'monthBeforeLast6Months',
    'monthBeforeLast7Months',
    'monthBeforeLast8Months',
    'monthBeforeLast9Months',
    'monthBeforeLast10Months'
] as const;

export type TransactionAmountsRequestType = typeof ALL_TRANSACTION_AMOUNTS_REQUEST_TYPE[number];

export const LATEST_12MONTHS_TRANSACTION_AMOUNTS_REQUEST_TYPES: TransactionAmountsRequestType[] = [
    'monthBeforeLast10Months',
    'monthBeforeLast9Months',
    'monthBeforeLast8Months',
    'monthBeforeLast7Months',
    'monthBeforeLast6Months',
    'monthBeforeLast5Months',
    'monthBeforeLast4Months',
    'monthBeforeLast3Months',
    'monthBeforeLast2Months',
    'monthBeforeLastMonth',
    'lastMonth',
    'thisMonth'
];

export interface TransactionAmountsRequestParams extends PartialRecord<TransactionAmountsRequestType, StartEndTime> {
    readonly useTransactionTimezone: boolean;
    today?: StartEndTime;
    thisWeek?: StartEndTime;
    thisMonth?: StartEndTime;
    thisYear?: StartEndTime;
    lastMonth?: StartEndTime;
    monthBeforeLastMonth?: StartEndTime;
    monthBeforeLast2Months?: StartEndTime;
    monthBeforeLast3Months?: StartEndTime;
    monthBeforeLast4Months?: StartEndTime;
    monthBeforeLast5Months?: StartEndTime;
    monthBeforeLast6Months?: StartEndTime;
    monthBeforeLast7Months?: StartEndTime;
    monthBeforeLast8Months?: StartEndTime;
    monthBeforeLast9Months?: StartEndTime;
    monthBeforeLast10Months?: StartEndTime;
}

export class TransactionAmountsRequest {
    public readonly useTransactionTimezone: boolean;
    public readonly query: string;

    public constructor(useTransactionTimezone: boolean, query: string) {
        this.useTransactionTimezone = useTransactionTimezone;
        this.query = query;
    }

    public buildQuery(): string {
        return `use_transaction_timezone=${this.useTransactionTimezone}` + (this.query.length ? '&query=' + this.query : '');
    }

    public static of(params: TransactionAmountsRequestParams): TransactionAmountsRequest {
        const queryParams: string[] = [];

        ALL_TRANSACTION_AMOUNTS_REQUEST_TYPE.forEach((type) => {
            if (params[type]) {
                queryParams.push(`${type}_${params[type].startTime}_${params[type].endTime}`);
            }
        });

        return new TransactionAmountsRequest(params.useTransactionTimezone, (queryParams.length ? queryParams.join('|') : ''));
    }
}

export interface TransactionInfoPageWrapperResponse {
    readonly items: TransactionInfoResponse[];
    readonly nextTimeSequenceId?: number;
    readonly totalCount?: number;
}

export interface TransactionInfoPageWrapperResponse2 {
    readonly items: TransactionInfoResponse[];
    readonly totalCount: number;
}

export interface TransactionReconciliationStatementResponseItem extends TransactionInfoResponse {
    readonly accountOpeningBalance: string;
    readonly accountClosingBalance: string;
}

export interface TransactionReconciliationStatementResponse {
    readonly transactions: TransactionReconciliationStatementResponseItem[];
    readonly totalInflows: string;
    readonly totalOutflows: string;
    readonly openingBalance: string;
    readonly closingBalance: string;
}

export interface TransactionUnreconciledCountRequest {
    readonly accountIds: string[];
}

export interface TransactionUnreconciledCountItem {
    readonly accountId: string;
    readonly count: number;
}

export interface TransactionUnreconciledCountResponse {
    readonly items: TransactionUnreconciledCountItem[];
}

export interface TransactionReconciliationStatementResponseItemWithInfo extends TransactionReconciliationStatementResponseItem {
    readonly sourceAccount?: Account;
    readonly sourceAccountName: string;
    readonly destinationAccount?: Account;
    readonly category?: TransactionCategory;
    readonly categoryName: string;
}

export interface TransactionReconciliationStatementResponseWithInfo {
    readonly transactions: TransactionReconciliationStatementResponseItemWithInfo[];
    readonly totalInflows: BigDecimal;
    readonly totalOutflows: BigDecimal;
    readonly openingBalance: BigDecimal;
    readonly closingBalance: BigDecimal;
}

export interface TransactionPageWrapper {
    readonly items: Transaction[];
    readonly totalCount?: number;
}

export interface TransactionStatisticResponse {
    readonly startTime: number;
    readonly endTime: number;
    readonly items: TransactionStatisticResponseItem[];
}

export interface TransactionStatisticResponseItem {
    readonly categoryId: string;
    readonly accountId: string;
    readonly relatedAccountId?: string;
    readonly relatedAccountType?: number;
    readonly amount: string;
}

export interface TransactionStatisticTrendsResponseItem {
    readonly year: number;
    readonly month: number; // 1-based (1 = January, 12 = December)
    readonly items: TransactionStatisticResponseItem[];
}

export interface TransactionStatisticAssetTrendsResponseItem extends YearMonthDay {
    readonly year: number;
    readonly month: number; // 1-based (1 = January, 12 = December)
    readonly day: number;
    readonly items: TransactionStatisticAssetTrendsResponseDataItem[];
}

export interface TransactionStatisticAssetTrendsResponseDataItem {
    readonly accountId: string;
    readonly accountOpeningBalance: string;
    readonly accountClosingBalance: string;
}

export interface YearMonthDataItem extends Year1BasedMonth, Record<string, unknown> {}

export interface YearMonthDayDataItem extends YearMonthDay, Record<string, unknown> {}

export interface YearMonthItems<T extends Year1BasedMonth> extends Record<string, unknown> {
    readonly items: T[];
}

export interface YearMonthDayItems<T extends YearMonthDay> extends Record<string, unknown> {
    readonly items: T[];
}
