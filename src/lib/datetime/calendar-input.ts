import {
    type TextualYearMonthDay,
    MeridiemIndicator
} from '@/core/datetime.ts';
import type { NumeralSystem } from '@/core/numeral.ts';
import 'moment-timezone/moment-timezone-utils';
import { getMonthLastDateTimeBySpecifiedUnixTime } from './boundaries.ts';
import { getCurrentDateTime, parseDateTimeFromUnixTime } from './conversion.ts';
import { MomentDateTime } from './moment.ts';

export function getCombinedDateAndTimeValues(date: Date, numeralSystem: NumeralSystem, hour: string, minute: string, second: string, meridiemIndicator: string, is24Hour: boolean): Date {
    const newDateTime = new Date(date.valueOf());
    let hours = numeralSystem.parseInt(hour);
    const minutes = numeralSystem.parseInt(minute);
    const seconds = numeralSystem.parseInt(second);

    if (!is24Hour) {
        if (hours === 12) {
            hours = 0;
        }

        if (meridiemIndicator === MeridiemIndicator.PM.name) {
            hours += 12;
        }
    }

    newDateTime.setHours(hours);
    newDateTime.setMinutes(minutes);
    newDateTime.setSeconds(seconds);

    return newDateTime;
}

export function getValidMonthDayOrCurrentDayShortDate(unixTime: number, currentShortDate: string): TextualYearMonthDay {
    const currentTime = getCurrentDateTime();
    const monthLastTime = getMonthLastDateTimeBySpecifiedUnixTime(unixTime);

    if (currentShortDate) {
        const yearMonthDay = currentShortDate.split('-');

        if (yearMonthDay.length === 3) {
            const currentDay = parseInt(yearMonthDay[2] as string);

            if (currentDay < monthLastTime.getGregorianCalendarDay()) {
                return monthLastTime.set({ dayOfMonth: currentDay }).getGregorianCalendarYearDashMonthDashDay();
            }
        }
    }

    if (monthLastTime.getGregorianCalendarYear() === currentTime.getGregorianCalendarYear() && monthLastTime.getGregorianCalendarMonth() === currentTime.getGregorianCalendarMonth()) {
        return currentTime.getGregorianCalendarYearDashMonthDashDay();
    }

    return monthLastTime.getGregorianCalendarYearDashMonthDashDay();
}

export function isDateRangeMatchFullYears(minTime: number, maxTime: number): boolean {
    const minDateTime = parseDateTimeFromUnixTime(minTime);
    const maxDateTime = parseDateTimeFromUnixTime(maxTime);
    return MomentDateTime.isGregorianCalendarYearFirstTime(minDateTime as MomentDateTime) && MomentDateTime.isGregorianCalendarYearLastTime(maxDateTime as MomentDateTime);
}

export function isDateRangeMatchFullMonths(minTime: number, maxTime: number): boolean {
    const minDateTime = parseDateTimeFromUnixTime(minTime);
    const maxDateTime = parseDateTimeFromUnixTime(maxTime);
    return MomentDateTime.isGregorianCalendarMonthFirstTime(minDateTime as MomentDateTime) && MomentDateTime.isGregorianCalendarMonthLastTime(maxDateTime as MomentDateTime);
}

export function isDateRangeMatchOneMonth(minTime: number, maxTime: number): boolean {
    const minDateTime = parseDateTimeFromUnixTime(minTime);
    const maxDateTime = parseDateTimeFromUnixTime(maxTime);

    if (minDateTime.getGregorianCalendarYear() !== maxDateTime.getGregorianCalendarYear() || minDateTime.getGregorianCalendarMonth() !== maxDateTime.getGregorianCalendarMonth()) {
        return false;
    }

    return isDateRangeMatchFullMonths(minTime, maxTime);
}
