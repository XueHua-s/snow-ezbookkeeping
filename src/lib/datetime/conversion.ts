import { getUtcOffsetByUtcOffsetMinutes } from './clock.ts';
import {
    type DateTime,
    type DateTimeFormatOptions,
    type TextualMonthDay,
    type TextualYearMonth,
    type TextualYearMonthDay,
    type TimeDifference,
    type Year0BasedMonth,
    type Year1BasedMonth,
    type YearMonthDay,
    type KnownDateTimeFormat,
    MeridiemIndicator
} from '@/core/datetime.ts';
import {
    NumeralSystem
} from '@/core/numeral.ts';
import {
    isNumber,
    isString
} from '@/lib/common.ts';
import moment from 'moment-timezone';
import 'moment-timezone/moment-timezone-utils';
import type { unitOfTime } from 'moment/moment';
import { MomentDateTime } from './moment.ts';

export function getAllowedYearRange(): [number, number] {
    return [2000, moment().year() + 1];
}

export function isYear0BasedMonthValid(year: number, month0base: number): boolean {
    if (!isNumber(year) || !isNumber(month0base)) {
        return false;
    }

    return year > 0 && month0base >= 0 && month0base <= 11;
}

export function getYear0BasedMonthObjectFromUnixTime(unixTime: number): Year0BasedMonth {
    const datetime = moment.unix(unixTime);

    return {
        year: datetime.year(),
        month0base: datetime.month()
    };
}

export function getYear0BasedMonthObjectFromString(yearMonth: TextualYearMonth | ''): Year0BasedMonth | null {
    if (!isString(yearMonth)) {
        return null;
    }

    const items = yearMonth.split('-');

    if (items.length !== 2) {
        return null;
    }

    const year = parseInt(items[0] as string);
    const month0base = parseInt(items[1] as string) - 1;

    if (!isYear0BasedMonthValid(year, month0base)) {
        return null;
    }

    return {
        year: year,
        month0base: month0base
    };
}

export function getYearMonthStringFromYear0BasedMonthObject(yearMonth: Year0BasedMonth | null): TextualYearMonth | '' {
    if (!yearMonth || !isYear0BasedMonthValid(yearMonth.year, yearMonth.month0base)) {
        return '';
    }

    return (`${yearMonth.year}-${yearMonth.month0base + 1}`) as TextualYearMonth;
}

export function isPM(hour: number): boolean {
    if (hour > 11) {
        return true;
    } else {
        return false;
    }
}

export function isUnixTimeYearMonthDayEquals(unixTime1: number, unixTime2: number): boolean {
    const date1 = moment.unix(unixTime1);
    const date2 = moment.unix(unixTime2);

    return date1.year() === date2.year() && date1.month() === date2.month() && date1.date() === date2.date();
}

export function isUnixTimeYearMonthDayHourEquals(unixTime1: number, unixTime2: number): boolean {
    const date1 = moment.unix(unixTime1);
    const date2 = moment.unix(unixTime2);

    return date1.year() === date2.year() && date1.month() === date2.month() && date1.date() === date2.date() && date1.hour() === date2.hour();
}

export function getTimezoneOffset(unixTime: number, timezone?: string): string {
    return getUtcOffsetByUtcOffsetMinutes(getTimezoneOffsetMinutes(unixTime, timezone));
}

export function getTimezoneOffsetMinutes(unixTime: number, timezone?: string): number {
    if (timezone) {
        return moment.unix(unixTime).tz(timezone).utcOffset();
    } else {
        return moment.unix(unixTime).utcOffset();
    }
}

export function getBrowserTimezoneOffset(unixTime: number): string {
    return getUtcOffsetByUtcOffsetMinutes(getBrowserTimezoneOffsetMinutes(unixTime));
}

export function getBrowserTimezoneOffsetMinutes(unixTime: number): number {
    const date = getLocalDatetimeFromUnixTime(getSameDateTimeWithBrowserTimezone(parseDateTimeFromUnixTime(unixTime)).getUnixTime());
    return -date.getTimezoneOffset();
}

export function getBrowserTimezoneName(): string {
    return new Intl.DateTimeFormat().resolvedOptions().timeZone;
}

export function getLocalDatetimeFromUnixTime(unixTime: number): Date {
    return new Date(unixTime * 1000);
}

export function getUnixTimeFromLocalDatetime(datetime: Date): number {
    return Math.floor(datetime.getTime() / 1000);
}

export function getSameDateTimeWithCurrentTimezone(dateTime: DateTime): DateTime {
    return MomentDateTime.now().set({
        year: dateTime.getGregorianCalendarYear(),
        month: dateTime.getGregorianCalendarMonth(),
        dayOfMonth: dateTime.getGregorianCalendarDay(),
        hour: dateTime.getHour(),
        minute: dateTime.getMinute(),
        second: dateTime.getSecond(),
        millisecond: 0
    });
}

export function getSameDateTimeWithBrowserTimezone(dateTime: DateTime): DateTime {
    return MomentDateTime.now().setTimezoneByIANATimeZoneName(getBrowserTimezoneName()).set({
        year: dateTime.getGregorianCalendarYear(),
        month: dateTime.getGregorianCalendarMonth(),
        dayOfMonth: dateTime.getGregorianCalendarDay(),
        hour: dateTime.getHour(),
        minute: dateTime.getMinute(),
        second: dateTime.getSecond(),
        millisecond: 0
    });
}

export function getSameDateTimeWithTimezoneOffset(dateTime: DateTime, utcOffset: number): DateTime {
    return MomentDateTime.now().setTimezoneByUtcOffsetMinutes(utcOffset).set({
        year: dateTime.getGregorianCalendarYear(),
        month: dateTime.getGregorianCalendarMonth(),
        dayOfMonth: dateTime.getGregorianCalendarDay(),
        hour: dateTime.getHour(),
        minute: dateTime.getMinute(),
        second: dateTime.getSecond(),
        millisecond: 0
    });
}

export function getCurrentDateTime(): DateTime {
    return MomentDateTime.now();
}

export function getCurrentUnixTime(): number {
    return moment().unix();
}

export function getYearMonthDayDateTime(year: number, month: number, day: number): DateTime {
    return MomentDateTime.ofFullDateTime(year, month, day, 0, 0, 0, 0);
}

export function parseDateTimeFromUnixTime(unixTime: number): DateTime {
    return MomentDateTime.ofUnixTime(unixTime);
}

export function parseDateTimeFromUnixTimeWithBrowserTimezone(unixTime: number): DateTime {
    return MomentDateTime.ofUnixTime(unixTime).setTimezoneByIANATimeZoneName(getBrowserTimezoneName());
}

export function parseDateTimeFromUnixTimeWithTimezoneOffset(unixTime: number, utcOffset: number): DateTime {
    return MomentDateTime.ofUnixTime(unixTime).setTimezoneByUtcOffsetMinutes(utcOffset);
}

export function parseDateTimeFromKnownDateTimeFormat(dateTime: string, format: KnownDateTimeFormat): DateTime | undefined {
    const m = moment(dateTime, format.format, true);

    if (!m.isValid()) {
        return undefined;
    }

    return MomentDateTime.of(m);
}

export function parseDateTimeFromString(dateTime: string, format: string): DateTime | undefined {
    const m = moment(dateTime, format, true);

    if (!m.isValid()) {
        return undefined;
    }

    return MomentDateTime.of(m);
}

export function formatDateTime(dateTime: DateTime, format: string, options: DateTimeFormatOptions): string {
    return dateTime.format(format, options);
}

export function formatUnixTime(unixTime: number, format: string, options: DateTimeFormatOptions): string {
    return parseDateTimeFromUnixTime(unixTime).format(format, options);
}

export function formatCurrentTime(format: string, options: DateTimeFormatOptions): string {
    return MomentDateTime.now().format(format, options);
}

export function formatGregorianCalendarYearDashMonthDashDay(date: TextualYearMonthDay, format: string, options: DateTimeFormatOptions): string {
    return MomentDateTime.of(moment(date, 'YYYY-MM-DD')).format(format, options);
}

export function formatGregorianCalendarMonthDashDay(monthDay: TextualMonthDay, format: string, options: DateTimeFormatOptions): string {
    return MomentDateTime.of(moment(monthDay, 'MM-DD')).format(format, options);
}

export function getLocalDateFromYearDashMonthDashDay(date: TextualYearMonthDay): Date | null {
    if (!isString(date)) {
        return null;
    }

    const items = date.split('-');

    if (items.length !== 3) {
        return null;
    }

    const year = parseInt(items[0] as string);
    const month = parseInt(items[1] as string);
    const day = parseInt(items[2] as string);

    if (!isNumber(year) || !isNumber(month) || !isNumber(day)) {
        return null;
    }

    if (year < 1000 || year > 9999 || month < 1 || month > 12 || day < 1 || day > 31) {
        return null;
    }

    const dateObj = new Date(year, month - 1, day);

    if (dateObj.getFullYear() !== year || dateObj.getMonth() !== (month - 1) || dateObj.getDate() !== day) {
        return null;
    }

    return dateObj;
}

export function getGregorianCalendarYearAndMonthFromLocalDate(date: Date): TextualYearMonthDay | '' {
    if (!date) {
        return '';
    }

    const year = date.getFullYear().toString().padStart(4, NumeralSystem.WesternArabicNumerals.digitZero);
    const month = (date.getMonth() + 1).toString().padStart(2, NumeralSystem.WesternArabicNumerals.digitZero);
    const day = (date.getDate()).toString().padStart(2, NumeralSystem.WesternArabicNumerals.digitZero);

    return (`${year}-${month}-${day}`) as TextualYearMonthDay;
}

export function getGregorianCalendarYearAndMonthFromUnixTime(unixTime: number): TextualYearMonth | '' {
    if (!unixTime) {
        return '';
    }

    return parseDateTimeFromUnixTime(unixTime).getGregorianCalendarYearDashMonth();
}

export function getGregorianCalendarYearMonthDays(yearMonth: Year1BasedMonth): number {
    return moment().set({ year: yearMonth.year, month: yearMonth.month1base - 1 }).daysInMonth();
}

export function getAMOrPM(hour: number): string {
    return isPM(hour) ? MeridiemIndicator.PM.name : MeridiemIndicator.AM.name;
}

export function getUnixTimeBeforeUnixTime(unixTime: number, amount: number, unit: unitOfTime.DurationConstructor): number {
    return moment.unix(unixTime).subtract(amount, unit).unix();
}

export function getUnixTimeAfterUnixTime(unixTime: number, amount: number, unit: unitOfTime.DurationConstructor): number {
    return moment.unix(unixTime).add(amount, unit).unix();
}

export function getDayDifference(yearMonthDay1: YearMonthDay, yearMonthDay2: YearMonthDay): number {
    const date1 = moment().set({ year: yearMonthDay1.year, month: yearMonthDay1.month - 1, date: yearMonthDay1.day, hour: 0, minute: 0, second: 0, millisecond: 0 });
    const date2 = moment().set({ year: yearMonthDay2.year, month: yearMonthDay2.month - 1, date: yearMonthDay2.day, hour: 0, minute: 0, second: 0, millisecond: 0 });
    return date2.diff(date1, 'days');
}

export function getTimeDifferenceHoursAndMinutes(timeDifferenceInMinutes: number): TimeDifference {
    const offsetHours = Math.trunc(Math.abs(timeDifferenceInMinutes) / 60);
    const offsetMinutes = Math.abs(timeDifferenceInMinutes) - offsetHours * 60;

    return {
        offsetHours: offsetHours,
        offsetMinutes: offsetMinutes,
    };
}
