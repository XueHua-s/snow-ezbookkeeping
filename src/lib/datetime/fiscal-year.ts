import type { DateTime } from '@/core/datetime.ts';
import {
    type FiscalYearUnixTime,
    FiscalYearStart
} from '@/core/fiscalyear.ts';
import {
    isNumber
} from '@/lib/common.ts';
import moment from 'moment-timezone';
import 'moment-timezone/moment-timezone-utils';
import { getFixedTimezoneName, MomentDateTime } from './moment.ts';

export function getFiscalYearFromUnixTime(unixTime: number, fiscalYearStartValue: number, utcOffset?: number): number {
    let date = moment.unix(unixTime);

    if (isNumber(utcOffset)) {
        date = date.tz(getFixedTimezoneName(utcOffset));
    }

    // For January 1 fiscal year start, fiscal year matches calendar year
    if (fiscalYearStartValue === FiscalYearStart.JanuaryFirstDay.value) {
        return date.year();
    }

    // Get date components
    const month = date.month() + 1; // 1-index
    const day = date.date();
    const year = date.year();

    let fiscalYearStart = FiscalYearStart.valueOf(fiscalYearStartValue);

    if (!fiscalYearStart) {
        fiscalYearStart = FiscalYearStart.Default;
    }

    // For other fiscal year starts:
    // If input time comes before the fiscal year start day in the calendar year,
    // it belongs to the fiscal year that ends in the current calendar year
    if (month < fiscalYearStart.month || (month === fiscalYearStart.month && day < fiscalYearStart.day)) {
        return year;
    }

    // If input time is on or after the fiscal year start day in the calendar year,
    // it belongs to the fiscal year that ends in the next calendar year
    return year + 1;
}

export function getFiscalYearStartDateTime(unixTime: number, fiscalYearStartValue: number, utcOffset?: number): DateTime {
    let date = moment.unix(unixTime);

    if (isNumber(utcOffset)) {
        date = date.tz(getFixedTimezoneName(utcOffset));
    }

    // For January 1 fiscal year start, fiscal year start time is always January 1 in the input calendar year
    if (fiscalYearStartValue === FiscalYearStart.JanuaryFirstDay.value) {
        let finalDate = moment();

        if (isNumber(utcOffset)) {
            finalDate = finalDate.tz(getFixedTimezoneName(utcOffset));
        }

        return MomentDateTime.of(finalDate.year(date.year()).month(0).date(1).hour(0).minute(0).second(0).millisecond(0));
    }

    let fiscalYearStart = FiscalYearStart.valueOf(fiscalYearStartValue);

    if (!fiscalYearStart) {
        fiscalYearStart = FiscalYearStart.Default;
    }

    const month = date.month() + 1; // 1-index
    const day = date.date();
    const year = date.year();

    // For other fiscal year starts:
    // If input time comes before the fiscal year start day in the calendar year,
    // the relevant fiscal year has a start date in Calendar Year = Input Year, and end date in Calendar Year = Input Year + 1.
    // If input time comes on or after the fiscal year start day in the calendar year,
    // the relevant fiscal year has a start date in Calendar Year = Input Year - 1, and end date in Calendar Year = Input Year.
    let startYear = year - 1;
    if (month > fiscalYearStart.month || (month === fiscalYearStart.month && day >= fiscalYearStart.day)) {
        startYear = year;
    }

    let finalDate = moment();

    if (isNumber(utcOffset)) {
        finalDate = finalDate.tz(getFixedTimezoneName(utcOffset));
    }

    return MomentDateTime.of(finalDate.set({
        year: startYear,
        month: fiscalYearStart.month - 1, // 0-index
        date: fiscalYearStart.day,
        hour: 0,
        minute: 0,
        second: 0,
        millisecond: 0,
    }));
}

export function getFiscalYearEndDateTime(unixTime: number, fiscalYearStart: number, utcOffset?: number): DateTime {
    return getFiscalYearStartDateTime(unixTime, fiscalYearStart, utcOffset).add(1, 'years').subtract(1, 'seconds');
}

export function getFiscalYearStartUnixTime(unixTime: number, fiscalYearStart: number, utcOffset?: number): number {
    return getFiscalYearStartDateTime(unixTime, fiscalYearStart, utcOffset).getUnixTime();
}

export function getFiscalYearEndUnixTime(unixTime: number, fiscalYearStart: number, utcOffset?: number): number {
    return getFiscalYearEndDateTime(unixTime, fiscalYearStart, utcOffset).getUnixTime();
}

export function getCurrentFiscalYear(fiscalYearStart: number, utcOffset?: number): number {
    const date = moment();
    return getFiscalYearFromUnixTime(date.unix(), fiscalYearStart, utcOffset);
}

export function getFiscalYearTimeRangeFromUnixTime(unixTime: number, fiscalYearStart: number, utcOffset?: number): FiscalYearUnixTime {
    const start = getFiscalYearStartUnixTime(unixTime, fiscalYearStart, utcOffset);
    const end = getFiscalYearEndUnixTime(unixTime, fiscalYearStart, utcOffset);
    return {
        year: getFiscalYearFromUnixTime(unixTime, fiscalYearStart, utcOffset),
        minUnixTime: start,
        maxUnixTime: end,
    };
}

export function getFiscalYearTimeRangeFromYear(year: number, fiscalYearStartValue: number): FiscalYearUnixTime {
    const fiscalYear = year;
    let fiscalYearStart = FiscalYearStart.valueOf(fiscalYearStartValue);

    if (!fiscalYearStart) {
        fiscalYearStart = FiscalYearStart.Default;
    }

    // For a specified fiscal year (e.g., 2023), the start date is in the previous calendar year
    // unless fiscal year starts on January 1
    const calendarStartYear = fiscalYearStartValue === FiscalYearStart.JanuaryFirstDay.value ? fiscalYear : fiscalYear - 1;

    // Create the timestamp for the start of the fiscal year
    const fiscalYearStartUnixTime = moment().set({
        year: calendarStartYear,
        month: fiscalYearStart.month - 1, // 0-index
        date: fiscalYearStart.day,
        hour: 0,
        minute: 0,
        second: 0,
        millisecond: 0,
    }).unix();

    // Fiscal year end is one year after start minus 1 second
    const fiscalYearEndUnixTime = moment.unix(fiscalYearStartUnixTime).add(1, 'years').subtract(1, 'seconds').unix();

    return {
        year: fiscalYear,
        minUnixTime: fiscalYearStartUnixTime,
        maxUnixTime: fiscalYearEndUnixTime,
    };
}
