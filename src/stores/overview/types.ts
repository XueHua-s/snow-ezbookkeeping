import type { WritableStartEndTime } from '@/core/datetime.ts';
import type {
TransactionAmountsRequestType
} from '@/models/transaction.ts';

export interface TransactionDataRange extends Record<TransactionAmountsRequestType, WritableStartEndTime> {
    today: {
        startTime: number;
        endTime: number;
    };
    thisWeek: {
        startTime: number;
        endTime: number;
    };
    thisMonth: {
        startTime: number;
        endTime: number;
    };
    thisYear: {
        startTime: number;
        endTime: number;
    };
    lastMonth: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLastMonth: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast2Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast3Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast4Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast5Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast6Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast7Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast8Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast9Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast10Months: {
        startTime: number;
        endTime: number;
    };
}

export interface TransactionOverviewOptions {
    loadedMonths: number;
}
