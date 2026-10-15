import {
    itemAndIndex
} from '@/core/base.ts';
import {
    type DateFormat,
    type LocalizedRecentMonthDateRange,
    type RecentMonthDateRange,
    type TimeFormat,
    type TimeRange,
    type TimeRangeAndDateType,
    type WeekDayValue,
    DateRange,
    DateRangeScene,
    LANGUAGE_DEFAULT_DATE_TIME_FORMAT_VALUE
} from '@/core/datetime.ts';
import moment from 'moment-timezone';
import 'moment-timezone/moment-timezone-utils';
import {
    getMonthFirstDateTimeBySpecifiedUnixTime, getMonthLastDateTimeBySpecifiedUnixTime, getThisMonthFirstUnixTime,
    getThisMonthLastUnixTime, getThisMonthSpecifiedDayFirstUnixTime, getThisMonthSpecifiedDayLastUnixTime,
    getThisWeekFirstUnixTime, getThisWeekLastUnixTime, getThisYearFirstUnixTime, getThisYearLastUnixTime,
    getTodayFirstUnixTime, getTodayLastUnixTime,
} from './boundaries.ts';
import { isDateRangeMatchOneMonth } from './calendar-input.ts';
import { getCurrentDateTime, getUnixTimeAfterUnixTime, getUnixTimeBeforeUnixTime, parseDateTimeFromUnixTime } from './conversion.ts';
import { getFiscalYearEndUnixTime, getFiscalYearStartUnixTime } from './fiscal-year.ts';

export function getDateTimeFormatType<T extends DateFormat | TimeFormat>(allFormatMap: Record<string, T>, allFormatArray: T[], formatTypeValue: number, languageDefaultTypeName: string, systemDefaultFormatType: T): T {
    if (formatTypeValue > LANGUAGE_DEFAULT_DATE_TIME_FORMAT_VALUE && allFormatArray[formatTypeValue - 1] && allFormatArray[formatTypeValue - 1]!.typeName) {
        return allFormatArray[formatTypeValue - 1] as T;
    } else if (formatTypeValue === LANGUAGE_DEFAULT_DATE_TIME_FORMAT_VALUE && allFormatMap[languageDefaultTypeName] && allFormatMap[languageDefaultTypeName].typeName) {
        return allFormatMap[languageDefaultTypeName];
    } else {
        return systemDefaultFormatType;
    }
}

export function getShiftedDateRange(minTime: number, maxTime: number, scale: number): TimeRange {
    const minDateTime = moment.unix(parseDateTimeFromUnixTime(minTime).getUnixTime()).set({ millisecond: 0 });
    const maxDateTime = moment.unix(parseDateTimeFromUnixTime(maxTime).getUnixTime()).set({ millisecond: 999 });

    const firstDayOfMonth = minDateTime.clone().startOf('month');
    const lastDayOfMonth = maxDateTime.clone().endOf('month');

    // check whether the date range matches full months
    if (firstDayOfMonth.unix() === minDateTime.unix() && lastDayOfMonth.unix() === maxDateTime.unix()) {
        const months = maxDateTime.year() * 12 + (maxDateTime.month() + 1) - minDateTime.year() * 12 - (minDateTime.month() + 1) + 1;
        const newMinDateTime = minDateTime.add(months * scale, 'months');
        const newMaxDateTime = newMinDateTime.clone().add(months, 'months').subtract(1, 'seconds');

        return {
            minTime: newMinDateTime.unix(),
            maxTime: newMaxDateTime.unix()
        };
    }

    // check whether the date range matches one full year
    if (minDateTime.clone().add(1, 'years').subtract(1, 'seconds').unix() === maxDateTime.unix() ||
        maxDateTime.clone().subtract(1, 'years').add(1, 'seconds').unix() === minDateTime.unix()) {
        const newMinDateTime = minDateTime.add(1 * scale, 'years');
        const newMaxDateTime = maxDateTime.add(1 * scale, 'years');

        return {
            minTime: newMinDateTime.unix(),
            maxTime: newMaxDateTime.unix()
        };
    }

    // check whether the date range matches one full month
    if (minDateTime.clone().add(1, 'months').subtract(1, 'seconds').unix() === maxDateTime.unix() ||
        maxDateTime.clone().subtract(1, 'months').add(1, 'seconds').unix() === minDateTime.unix()) {
        const newMinDateTime = minDateTime.add(1 * scale, 'months');
        const newMaxDateTime = maxDateTime.add(1 * scale, 'months');

        return {
            minTime: newMinDateTime.unix(),
            maxTime: newMaxDateTime.unix()
        };
    }

    const range = (maxTime - minTime + 1) * scale;

    return {
        minTime: minTime + range,
        maxTime: maxTime + range
    };
}

export function getShiftedDateRangeAndDateType(minTime: number, maxTime: number, scale: number, firstDayOfWeek: WeekDayValue, fiscalYearStart: number, scene: DateRangeScene): TimeRangeAndDateType {
    const newDateRange = getShiftedDateRange(minTime, maxTime, scale);
    const newDateType = getDateTypeByDateRange(newDateRange.minTime, newDateRange.maxTime, firstDayOfWeek, fiscalYearStart, scene);

    return {
        dateType: newDateType,
        minTime: newDateRange.minTime,
        maxTime: newDateRange.maxTime
    };
}

export function getShiftedDateRangeAndDateTypeForBillingCycle(minTime: number, maxTime: number, scale: number, firstDayOfWeek: WeekDayValue, fiscalYearStart: number, scene: number, statementDate: number | undefined | null): TimeRangeAndDateType | null {
    if (!statementDate || !DateRange.PreviousBillingCycle.isAvailableForScene(scene) || !DateRange.CurrentBillingCycle.isAvailableForScene(scene)) {
        return null;
    }

    const previousBillingCycleRange = getDateRangeByBillingCycleDateType(DateRange.PreviousBillingCycle.type, firstDayOfWeek, fiscalYearStart, statementDate);
    const currentBillingCycleRange = getDateRangeByBillingCycleDateType(DateRange.CurrentBillingCycle.type, firstDayOfWeek, fiscalYearStart, statementDate);

    if (previousBillingCycleRange && getUnixTimeBeforeUnixTime(previousBillingCycleRange.maxTime, 1, 'months') === maxTime && getUnixTimeBeforeUnixTime(previousBillingCycleRange.minTime, 1, 'months') === minTime && scale === 1) {
        return previousBillingCycleRange;
    } else if (previousBillingCycleRange && previousBillingCycleRange.maxTime === maxTime && previousBillingCycleRange.minTime === minTime && scale === 1) {
        return currentBillingCycleRange;
    } else if (currentBillingCycleRange && currentBillingCycleRange.maxTime === maxTime && currentBillingCycleRange.minTime === minTime && scale === -1) {
        return previousBillingCycleRange;
    } else if (currentBillingCycleRange && getUnixTimeAfterUnixTime(currentBillingCycleRange.maxTime, 1, 'months') === maxTime && getUnixTimeAfterUnixTime(currentBillingCycleRange.minTime, 1, 'months') === minTime && scale === -1) {
        return currentBillingCycleRange;
    }

    return null;
}

export function getDateTypeByDateRange(minTime: number, maxTime: number, firstDayOfWeek: WeekDayValue, fiscalYearStart: number, scene: DateRangeScene): number {
    const allDateRanges = DateRange.values();
    let newDateType = DateRange.Custom.type;

    for (const dateRange of allDateRanges) {
        if (!dateRange.isAvailableForScene(scene)) {
            continue;
        }

        const range = getDateRangeByDateType(dateRange.type, firstDayOfWeek, fiscalYearStart);

        if (range && range.minTime === minTime && range.maxTime === maxTime) {
            newDateType = dateRange.type;
            break;
        }
    }

    return newDateType;
}

export function getDateTypeByBillingCycleDateRange(minTime: number, maxTime: number, firstDayOfWeek: WeekDayValue, fiscalYearStart: number, scene: DateRangeScene, statementDate: number | undefined | null): number | null {
    if (!statementDate || !DateRange.PreviousBillingCycle.isAvailableForScene(scene) || !DateRange.CurrentBillingCycle.isAvailableForScene(scene)) {
        return null;
    }

    const previousBillingCycleRange = getDateRangeByBillingCycleDateType(DateRange.PreviousBillingCycle.type, firstDayOfWeek, fiscalYearStart, statementDate);
    const currentBillingCycleRange = getDateRangeByBillingCycleDateType(DateRange.CurrentBillingCycle.type, firstDayOfWeek, fiscalYearStart, statementDate);

    if (previousBillingCycleRange && previousBillingCycleRange.maxTime === maxTime && previousBillingCycleRange.minTime === minTime) {
        return previousBillingCycleRange.dateType;
    } else if (currentBillingCycleRange && currentBillingCycleRange.maxTime === maxTime && currentBillingCycleRange.minTime === minTime) {
        return currentBillingCycleRange.dateType;
    }

    return null;
}

export function getDateRangeByDateType(dateType: number | undefined, firstDayOfWeek: WeekDayValue, fiscalYearStart: number): TimeRangeAndDateType | null {
    let maxTime = 0;
    let minTime = 0;

    if (dateType === DateRange.All.type) { // All
        maxTime = 0;
        minTime = 0;
    } else if (dateType === DateRange.Today.type) { // Today
        maxTime = getTodayLastUnixTime();
        minTime = getTodayFirstUnixTime();
    } else if (dateType === DateRange.Yesterday.type) { // Yesterday
        maxTime = getUnixTimeBeforeUnixTime(getTodayLastUnixTime(), 1, 'days');
        minTime = getUnixTimeBeforeUnixTime(getTodayFirstUnixTime(), 1, 'days');
    } else if (dateType === DateRange.LastSevenDays.type) { // Last 7 days
        maxTime = getTodayLastUnixTime();
        minTime = getUnixTimeBeforeUnixTime(getTodayFirstUnixTime(), 6, 'days');
    } else if (dateType === DateRange.LastThirtyDays.type) { // Last 30 days
        maxTime = getTodayLastUnixTime();
        minTime = getUnixTimeBeforeUnixTime(getTodayFirstUnixTime(), 29, 'days');
    } else if (dateType === DateRange.ThisWeek.type) { // This week
        maxTime = getThisWeekLastUnixTime(firstDayOfWeek);
        minTime = getThisWeekFirstUnixTime(firstDayOfWeek);
    } else if (dateType === DateRange.LastWeek.type) { // Last week
        maxTime = getUnixTimeBeforeUnixTime(getThisWeekLastUnixTime(firstDayOfWeek), 7, 'days');
        minTime = getUnixTimeBeforeUnixTime(getThisWeekFirstUnixTime(firstDayOfWeek), 7, 'days');
    } else if (dateType === DateRange.ThisMonth.type) { // This month
        maxTime = getThisMonthLastUnixTime();
        minTime = getThisMonthFirstUnixTime();
    } else if (dateType === DateRange.LastMonth.type) { // Last month
        maxTime = getUnixTimeBeforeUnixTime(getThisMonthFirstUnixTime(), 1, 'seconds');
        minTime = getUnixTimeBeforeUnixTime(getThisMonthFirstUnixTime(), 1, 'months');
    } else if (dateType === DateRange.ThisYear.type) { // This year
        maxTime = getThisYearLastUnixTime();
        minTime = getThisYearFirstUnixTime();
    } else if (dateType === DateRange.LastYear.type) { // Last year
        maxTime = getUnixTimeBeforeUnixTime(getThisYearLastUnixTime(), 1, 'years');
        minTime = getUnixTimeBeforeUnixTime(getThisYearFirstUnixTime(), 1, 'years');
    } else if (dateType === DateRange.ThisFiscalYear.type) { // This fiscal year
        maxTime = getFiscalYearEndUnixTime(getTodayFirstUnixTime(), fiscalYearStart);
        minTime = getFiscalYearStartUnixTime(getTodayFirstUnixTime(), fiscalYearStart);
    } else if (dateType === DateRange.LastFiscalYear.type) { // Last fiscal year
        maxTime = getUnixTimeBeforeUnixTime(getFiscalYearEndUnixTime(getTodayFirstUnixTime(), fiscalYearStart), 1, 'years');
        minTime = getUnixTimeBeforeUnixTime(getFiscalYearStartUnixTime(getTodayFirstUnixTime(), fiscalYearStart), 1, 'years');
    } else if (dateType === DateRange.RecentTwelveMonths.type) { // Recent 12 months
        maxTime = getThisMonthLastUnixTime();
        minTime = getUnixTimeBeforeUnixTime(getThisMonthFirstUnixTime(), 11, 'months');
    } else if (dateType === DateRange.RecentTwentyFourMonths.type) { // Recent 24 months
        maxTime = getThisMonthLastUnixTime();
        minTime = getUnixTimeBeforeUnixTime(getThisMonthFirstUnixTime(), 23, 'months');
    } else if (dateType === DateRange.RecentThirtySixMonths.type) { // Recent 36 months
        maxTime = getThisMonthLastUnixTime();
        minTime = getUnixTimeBeforeUnixTime(getThisMonthFirstUnixTime(), 35, 'months');
    } else if (dateType === DateRange.RecentTwoYears.type) { // Recent 2 years
        maxTime = getThisYearLastUnixTime();
        minTime = getUnixTimeBeforeUnixTime(getThisYearFirstUnixTime(), 1, 'years');
    } else if (dateType === DateRange.RecentThreeYears.type) { // Recent 3 years
        maxTime = getThisYearLastUnixTime();
        minTime = getUnixTimeBeforeUnixTime(getThisYearFirstUnixTime(), 2, 'years');
    } else if (dateType === DateRange.RecentFiveYears.type) { // Recent 5 years
        maxTime = getThisYearLastUnixTime();
        minTime = getUnixTimeBeforeUnixTime(getThisYearFirstUnixTime(), 4, 'years');
    } else {
        return null;
    }

    return {
        dateType: dateType,
        maxTime: maxTime,
        minTime: minTime
    };
}

export function getDateRangeByBillingCycleDateType(dateType: number, firstDayOfWeek: WeekDayValue, fiscalYearStart: number, statementDate: number | undefined | null): TimeRangeAndDateType | null {
    let maxTime = 0;
    let minTime = 0;

    if (dateType === DateRange.PreviousBillingCycle.type || dateType === DateRange.CurrentBillingCycle.type) { // Previous Billing Cycle | Current Billing Cycle
        if (statementDate) {
            if (getCurrentDateTime().getGregorianCalendarDay() <= statementDate) {
                maxTime = getThisMonthSpecifiedDayLastUnixTime(statementDate);
                minTime = getUnixTimeBeforeUnixTime(getUnixTimeAfterUnixTime(getThisMonthSpecifiedDayFirstUnixTime(statementDate), 1, 'days'), 1, 'months');
            } else {
                maxTime = getUnixTimeAfterUnixTime(getThisMonthSpecifiedDayLastUnixTime(statementDate), 1, 'months');
                minTime = getUnixTimeAfterUnixTime(getThisMonthSpecifiedDayFirstUnixTime(statementDate), 1, 'days');
            }

            if (dateType === DateRange.PreviousBillingCycle.type) {
                maxTime = getUnixTimeBeforeUnixTime(maxTime, 1, 'months');
                minTime = getUnixTimeBeforeUnixTime(minTime, 1, 'months');
            }
        } else {
            let fallbackDateRange = null;

            if (dateType === DateRange.CurrentBillingCycle.type) { // same as This Month
                fallbackDateRange = getDateRangeByDateType(DateRange.ThisMonth.type, firstDayOfWeek, fiscalYearStart);
            } else if (dateType === DateRange.PreviousBillingCycle.type) { // same as Last Month
                fallbackDateRange = getDateRangeByDateType(DateRange.LastMonth.type, firstDayOfWeek, fiscalYearStart);
            }

            if (fallbackDateRange) {
                maxTime = fallbackDateRange.maxTime;
                minTime = fallbackDateRange.minTime;
            }
        }
    } else {
        return null;
    }

    return {
        dateType: dateType,
        maxTime: maxTime,
        minTime: minTime
    };
}

export function getDateRangeByLastReconciledTimeRangeDateType(dateType: number, lastReconciledTime: number | undefined | null): TimeRangeAndDateType | null {
    let maxTime = 0;
    let minTime = 0;

    if (dateType === DateRange.SinceLastReconciledTime.type) { // Since Last Reconciled Time
        if (lastReconciledTime) {
            maxTime = getTodayLastUnixTime();
            minTime = lastReconciledTime;
        } else {
            return null;
        }
    } else {
        return null;
    }

    return {
        dateType: dateType,
        maxTime: maxTime,
        minTime: minTime
    };
}

export function getRecentMonthDateRanges(monthCount: number): RecentMonthDateRange[] {
    const recentDateRanges: RecentMonthDateRange[] = [];
    const thisMonthFirstUnixTime = getThisMonthFirstUnixTime();

    for (let i = 0; i < monthCount; i++) {
        let minTime = thisMonthFirstUnixTime;

        if (i > 0) {
            minTime = getUnixTimeBeforeUnixTime(thisMonthFirstUnixTime, i, 'months');
        }

        const maxTime = getUnixTimeBeforeUnixTime(getUnixTimeAfterUnixTime(minTime, 1, 'months'), 1, 'seconds');
        let dateType = DateRange.Custom.type;
        const year = parseDateTimeFromUnixTime(minTime).getGregorianCalendarYear();
        const month = parseDateTimeFromUnixTime(minTime).getGregorianCalendarMonth();

        if (i === 0) {
            dateType = DateRange.ThisMonth.type;
        } else if (i === 1) {
            dateType = DateRange.LastMonth.type;
        }

        recentDateRanges.push({
            dateType: dateType,
            minTime: minTime,
            maxTime: maxTime,
            year: year,
            month: month
        });
    }

    return recentDateRanges;
}

export function getRecentDateRangeIndexByDateType(allRecentMonthDateRanges: LocalizedRecentMonthDateRange[], dateType: number): number {
    for (const [recentDateRange, index] of itemAndIndex(allRecentMonthDateRanges)) {
        if (!recentDateRange.isPreset && recentDateRange.dateType === dateType) {
            return index;
        }
    }

    return -1;
}

export function getRecentDateRangeIndex(allRecentMonthDateRanges: LocalizedRecentMonthDateRange[], dateType: number, minTime: number, maxTime: number, firstDayOfWeek: WeekDayValue, fiscalYearStart: number): number {
    let dateRange = getDateRangeByDateType(dateType, firstDayOfWeek, fiscalYearStart);

    if (dateRange && dateRange.dateType === DateRange.All.type) {
        return getRecentDateRangeIndexByDateType(allRecentMonthDateRanges, DateRange.All.type);
    }

    if (!dateRange && (!maxTime || !minTime)) {
        return getRecentDateRangeIndexByDateType(allRecentMonthDateRanges, DateRange.Custom.type);
    }

    if (!dateRange) {
        dateRange = {
            dateType: DateRange.Custom.type,
            maxTime: maxTime,
            minTime: minTime
        };
    }

    for (const [recentDateRange, index] of itemAndIndex(allRecentMonthDateRanges)) {
        if (recentDateRange.isPreset && recentDateRange.minTime === dateRange.minTime && recentDateRange.maxTime === dateRange.maxTime) {
            return index;
        }
    }

    return getRecentDateRangeIndexByDateType(allRecentMonthDateRanges, DateRange.Custom.type);
}

export function getFullMonthDateRange(minTime: number, maxTime: number, firstDayOfWeek: WeekDayValue, fiscalYearStart: number): TimeRangeAndDateType | null {
    if (isDateRangeMatchOneMonth(minTime, maxTime)) {
        return null;
    }

    if (!minTime) {
        return getDateRangeByDateType(DateRange.ThisMonth.type, firstDayOfWeek, fiscalYearStart);
    }

    const monthFirstDateTime = getMonthFirstDateTimeBySpecifiedUnixTime(minTime);
    const monthLastDateTime = getMonthLastDateTimeBySpecifiedUnixTime(minTime);
    const dateType = getDateTypeByDateRange(monthFirstDateTime.getUnixTime(), monthLastDateTime.getUnixTime(), firstDayOfWeek, fiscalYearStart, DateRangeScene.Normal);

    const dateRange: TimeRangeAndDateType = {
        dateType: dateType,
        maxTime: monthLastDateTime.getUnixTime(),
        minTime: monthFirstDateTime.getUnixTime()
    };

    return dateRange;
}
