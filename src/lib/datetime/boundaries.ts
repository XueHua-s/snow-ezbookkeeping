import type { DateTime, TextualYearMonth, WeekDayValue, Year0BasedMonth, Year1BasedMonth, YearQuarter } from '@/core/datetime.ts';
import {
    isNumber,
    isObject,
    isString
} from '@/lib/common.ts';
import moment from 'moment-timezone';
import 'moment-timezone/moment-timezone-utils';
import { getYear0BasedMonthObjectFromString, isYear0BasedMonthValid } from './conversion.ts';
import { getFixedTimezoneName, MomentDateTime } from './moment.ts';

export function getTodayFirstUnixTime(): number {
    return moment().set({ hour: 0, minute: 0, second: 0, millisecond: 0 }).unix();
}

export function getTodayLastUnixTime(): number {
    return moment.unix(getTodayFirstUnixTime()).add(1, 'days').subtract(1, 'seconds').unix();
}

export function getThisWeekFirstUnixTime(firstDayOfWeek: WeekDayValue): number {
    const today = moment.unix(getTodayFirstUnixTime());

    if (!isNumber(firstDayOfWeek)) {
        firstDayOfWeek = 0;
    }

    let dayOfWeek = today.day() - firstDayOfWeek;

    if (dayOfWeek < 0) {
        dayOfWeek += 7;
    }

    return today.subtract(dayOfWeek, 'days').unix();
}

export function getThisWeekLastUnixTime(firstDayOfWeek: WeekDayValue): number {
    return moment.unix(getThisWeekFirstUnixTime(firstDayOfWeek)).add(7, 'days').subtract(1, 'seconds').unix();
}

export function getThisMonthFirstUnixTime(): number {
    const today = moment.unix(getTodayFirstUnixTime());
    return today.subtract(today.date() - 1, 'days').unix();
}

export function getThisMonthLastUnixTime(): number {
    return moment.unix(getThisMonthFirstUnixTime()).add(1, 'months').subtract(1, 'seconds').unix();
}

export function getThisMonthSpecifiedDayFirstUnixTime(date: number): number {
    return moment().set({ date: date, hour: 0, minute: 0, second: 0, millisecond: 0 }).unix();
}

export function getThisMonthSpecifiedDayLastUnixTime(date: number): number {
    return moment.unix(getThisMonthSpecifiedDayFirstUnixTime(date)).add(1, 'days').subtract(1, 'seconds').unix();
}

export function getThisYearFirstUnixTime(): number {
    const today = moment.unix(getTodayFirstUnixTime());
    return today.subtract(today.dayOfYear() - 1, 'days').unix();
}

export function getThisYearLastUnixTime(): number {
    return moment.unix(getThisYearFirstUnixTime()).add(1, 'years').subtract(1, 'seconds').unix();
}

export function getYearFirstDateTimeBySpecifiedDateTime(unixTime: number, utcOffset?: number): DateTime {
    let date = moment.unix(unixTime);

    if (isNumber(utcOffset)) {
        date = date.tz(getFixedTimezoneName(utcOffset));
    }

    return MomentDateTime.of(date.set({ hour: 0, minute: 0, second: 0, millisecond: 0 }).subtract(date.dayOfYear() - 1, 'days'));
}

export function getYearLastDateTimeBySpecifiedUnixTime(unixTime: number, utcOffset?: number): DateTime {
    return getYearFirstDateTimeBySpecifiedDateTime(unixTime, utcOffset).add(1, 'years').subtract(1, 'seconds');
}

export function getQuarterFirstTimeTimeBySpecifiedUnixTime(unixTime: number, utcOffset?: number): DateTime {
    let date = moment.unix(unixTime);

    if (isNumber(utcOffset)) {
        date = date.tz(getFixedTimezoneName(utcOffset));
    }

    date = date.set({ hour: 0, minute: 0, second: 0, millisecond: 0 });
    const month = date.month();
    const quarterStartMonth = Math.floor(month / 3) * 3;
    return MomentDateTime.of(date.set({ month: quarterStartMonth, date: 1 }));
}

export function getQuarterLastTimeTimeBySpecifiedUnixTime(unixTime: number, utcOffset?: number): DateTime {
    return getQuarterFirstTimeTimeBySpecifiedUnixTime(unixTime, utcOffset).add(3, 'months').subtract(1, 'seconds');
}

export function getMonthFirstDateTimeBySpecifiedUnixTime(unixTime: number, utcOffset?: number): DateTime {
    let date = moment.unix(unixTime);

    if (isNumber(utcOffset)) {
        date = date.tz(getFixedTimezoneName(utcOffset));
    }

    return MomentDateTime.of(date.set({ hour: 0, minute: 0, second: 0, millisecond: 0 }).subtract(date.date() - 1, 'days'));
}

export function getMonthLastDateTimeBySpecifiedUnixTime(unixTime: number, utcOffset?: number): DateTime {
    return getMonthFirstDateTimeBySpecifiedUnixTime(unixTime, utcOffset).add(1, 'months').subtract(1, 'seconds');
}

export function getDayFirstDateTimeBySpecifiedUnixTime(unixTime: number, utcOffset?: number): DateTime {
    let date = moment.unix(unixTime);

    if (isNumber(utcOffset)) {
        date = date.tz(getFixedTimezoneName(utcOffset));
    }

    return MomentDateTime.of(date.set({ hour: 0, minute: 0, second: 0, millisecond: 0 }));
}

export function getDayLastDateTimeBySpecifiedUnixTime(unixTime: number, utcOffset?: number): DateTime {
    return getDayFirstDateTimeBySpecifiedUnixTime(unixTime, utcOffset).add(1, 'days').subtract(1, 'seconds');
}

export function getBillingCycleFirstUnixTimeBySpecifiedUnixTime(unixTime: number, statementDate: number, utcOffset?: number): DateTime {
    let date = moment.unix(unixTime);

    if (isNumber(utcOffset)) {
        date = date.tz(getFixedTimezoneName(utcOffset));
    }

    if (date.date() > statementDate) {
        date = date.set({ date: statementDate + 1, hour: 0, minute: 0, second: 0, millisecond: 0 });
    } else {
        date = date.set({ date: statementDate, hour: 0, minute: 0, second: 0, millisecond: 0 }).add(-1, 'months').add(1, 'days');
    }

    return MomentDateTime.of(date);
}

export function getBillingCycleLastUnixTimeBySpecifiedUnixTime(unixTime: number, statementDate: number, utcOffset?: number): DateTime {
    return getBillingCycleFirstUnixTimeBySpecifiedUnixTime(unixTime, statementDate, utcOffset).add(1, 'months').subtract(1, 'seconds');
}

export function getYearFirstUnixTime(year: number): number {
    return moment().set({ year: year, month: 0, date: 1, hour: 0, minute: 0, second: 0, millisecond: 0 }).unix();
}

export function getYearLastUnixTime(year: number): number {
    return moment.unix(getYearFirstUnixTime(year)).add(1, 'years').subtract(1, 'seconds').unix();
}

export function getQuarterFirstUnixTime(yearQuarter: YearQuarter): number {
    return moment().set({ year: yearQuarter.year, month: (yearQuarter.quarter - 1) * 3, date: 1, hour: 0, minute: 0, second: 0, millisecond: 0 }).unix();
}

export function getQuarterLastUnixTime(yearQuarter: YearQuarter): number {
    return moment.unix(getQuarterFirstUnixTime(yearQuarter)).add(3, 'months').subtract(1, 'seconds').unix();
}

export function getYearMonthFirstUnixTime(yearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | ''): number {
    let yearMonthObj: Year0BasedMonth | null = null;

    if (isString(yearMonth)) {
        yearMonthObj = getYear0BasedMonthObjectFromString(yearMonth);
    } else if (isObject(yearMonth) && ('month0base' in yearMonth) && isYear0BasedMonthValid(yearMonth.year, yearMonth.month0base)) {
        yearMonthObj = yearMonth;
    } else if (isObject(yearMonth) && ('month1base' in yearMonth) && isYear0BasedMonthValid(yearMonth.year, yearMonth.month1base - 1)) {
        yearMonthObj = {
            year: yearMonth.year,
            month0base: yearMonth.month1base - 1
        };
    }

    if (!yearMonthObj) {
        return 0;
    }

    return moment().set({ year: yearMonthObj.year, month: yearMonthObj.month0base, date: 1, hour: 0, minute: 0, second: 0, millisecond: 0 }).unix();
}

export function getYearMonthLastUnixTime(yearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | ''): number {
    return moment.unix(getYearMonthFirstUnixTime(yearMonth)).add(1, 'months').subtract(1, 'seconds').unix();
}
