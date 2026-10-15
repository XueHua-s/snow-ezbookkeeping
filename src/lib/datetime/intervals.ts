import {
    type DateTime,
    type TextualYearMonth,
    type Year0BasedMonth,
    type Year1BasedMonth,
    type YearMonthRange,
    type YearQuarter,
    type YearUnixTime,
    YearMonthDayUnixTime,
    YearMonthUnixTime,
    YearQuarterUnixTime
} from '@/core/datetime.ts';
import {
    type FiscalYearUnixTime,
    FiscalYearStart
} from '@/core/fiscalyear.ts';
import {
    isObject,
    isString
} from '@/lib/common.ts';
import 'moment-timezone/moment-timezone-utils';
import {
    getDayFirstDateTimeBySpecifiedUnixTime, getDayLastDateTimeBySpecifiedUnixTime, getMonthFirstDateTimeBySpecifiedUnixTime,
    getQuarterFirstUnixTime, getQuarterLastUnixTime, getYearFirstUnixTime, getYearLastUnixTime, getYearMonthFirstUnixTime,
    getYearMonthLastUnixTime,
} from './boundaries.ts';
import { getYear0BasedMonthObjectFromString, parseDateTimeFromUnixTime } from './conversion.ts';
import { getFiscalYearEndUnixTime, getFiscalYearFromUnixTime, getFiscalYearStartUnixTime } from './fiscal-year.ts';

export function getStartEndYearMonthRange(startYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | '', endYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | ''): YearMonthRange | null {
    let startYearMonthObj: Year0BasedMonth | null = null;
    let endYearMonthObj: Year0BasedMonth | null = null;

    if (isString(startYearMonth)) {
        startYearMonthObj = getYear0BasedMonthObjectFromString(startYearMonth);
    } else if (isObject(startYearMonth) && ('month0base' in startYearMonth)) {
        startYearMonthObj = startYearMonth;
    } else if (isObject(startYearMonth) && ('month1base' in startYearMonth)) {
        startYearMonthObj = {
            year: startYearMonth.year,
            month0base: startYearMonth.month1base - 1
        };
    }

    if (isString(endYearMonth)) {
        endYearMonthObj = getYear0BasedMonthObjectFromString(endYearMonth);
    } else if (isObject(endYearMonth) && ('month0base' in endYearMonth)) {
        endYearMonthObj = endYearMonth;
    } else if (isObject(endYearMonth) && ('month1base' in endYearMonth)) {
        endYearMonthObj = {
            year: endYearMonth.year,
            month0base: endYearMonth.month1base - 1
        };
    }

    if (!startYearMonthObj || !endYearMonthObj) {
        return null;
    }

    return {
        startYearMonth: startYearMonthObj,
        endYearMonth: endYearMonthObj
    };
}

export function getAllYearsStartAndEndUnixTimes(startYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | '', endYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | ''): YearUnixTime[] {
    const allYearTimes: YearUnixTime[] = [];
    const range = getStartEndYearMonthRange(startYearMonth, endYearMonth);

    if (!range) {
        return allYearTimes;
    }

    for (let year = range.startYearMonth.year; year <= range.endYearMonth.year; year++) {
        const yearTime: YearUnixTime = {
            year: year,
            minUnixTime: getYearFirstUnixTime(year),
            maxUnixTime: getYearLastUnixTime(year),
        };

        allYearTimes.push(yearTime);
    }

    return allYearTimes;
}

export function getAllFiscalYearsStartAndEndUnixTimes(startYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | '', endYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | '', fiscalYearStartValue: number): FiscalYearUnixTime[] {
    // user selects date range: start=2024-01 and end=2026-12
    // result should be 4x FiscalYearUnixTime made up of:
    // - 2024-01->2024-06 (FY 24) - input start year-month->end of fiscal year in which the input start year-month falls
    // - 2024-07->2025-06 (FY 25) - complete fiscal year
    // - 2025-07->2026-06 (FY 26) - complete fiscal year
    // - 2026-07->2026-12 (FY 27) - start of fiscal year->end of fiscal year in which the input end year-month falls

    const allFiscalYearTimes: FiscalYearUnixTime[] = [];
    const range = getStartEndYearMonthRange(startYearMonth, endYearMonth);

    if (!range) {
        return allFiscalYearTimes;
    }

    const inputStartUnixTime = getYearMonthFirstUnixTime(range.startYearMonth);
    const inputEndUnixTime = getYearMonthLastUnixTime(range.endYearMonth);
    let fiscalYearStart = FiscalYearStart.valueOf(fiscalYearStartValue);

    if (!fiscalYearStart) {
        fiscalYearStart = FiscalYearStart.Default;
    }

    // Loop over 1 year before and 1 year after the input date range
    // to include fiscal years that start in the previous calendar year.
    for (let year = range.startYearMonth.year - 1; year <= range.endYearMonth.year + 1; year++) {
        const thisYearMonthUnixTime = getYearMonthFirstUnixTime({ year: year, month1base: fiscalYearStart.month });
        const fiscalStartTime = getFiscalYearStartUnixTime(thisYearMonthUnixTime, fiscalYearStart.value);
        const fiscalEndTime = getFiscalYearEndUnixTime(thisYearMonthUnixTime, fiscalYearStart.value);

        const fiscalYear = getFiscalYearFromUnixTime(fiscalStartTime, fiscalYearStart.value);

        if (fiscalStartTime <= inputEndUnixTime && fiscalEndTime >= inputStartUnixTime) {
            const fiscalYearTime: FiscalYearUnixTime = {
                year: fiscalYear,
                minUnixTime: fiscalStartTime,
                maxUnixTime: fiscalEndTime,
            };

            allFiscalYearTimes.push(fiscalYearTime);
        }

        if (fiscalStartTime > inputEndUnixTime) {
            break;
        }
    }

    return allFiscalYearTimes;
}

export function getAllQuartersStartAndEndUnixTimes(startYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | '', endYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | ''): YearQuarterUnixTime[] {
    const allYearQuarterTimes: YearQuarterUnixTime[] = [];
    const range = getStartEndYearMonthRange(startYearMonth, endYearMonth);

    if (!range) {
        return allYearQuarterTimes;
    }

    for (let year = range.startYearMonth.year, month0base = range.startYearMonth.month0base; year < range.endYearMonth.year || (year === range.endYearMonth.year && (Math.floor(month0base / 3) <= Math.floor(range.endYearMonth.month0base / 3))); ) {
        const yearQuarter: YearQuarter = {
            year: year,
            quarter: Math.floor((month0base / 3)) + 1
        };

        const minUnixTime = getQuarterFirstUnixTime(yearQuarter);
        const maxUnixTime = getQuarterLastUnixTime(yearQuarter);

        allYearQuarterTimes.push(YearQuarterUnixTime.of(yearQuarter, minUnixTime, maxUnixTime));

        if (year === range.endYearMonth.year && month0base >= range.endYearMonth.month0base) {
            break;
        }

        if (month0base >= 9) {
            year++;
            month0base = 0;
        } else {
            month0base += 3;
        }
    }

    return allYearQuarterTimes;
}

export function getAllMonthsStartAndEndUnixTimes(startYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | '', endYearMonth: Year0BasedMonth | Year1BasedMonth | TextualYearMonth | ''): YearMonthUnixTime[] {
    const allYearMonthTimes: YearMonthUnixTime[] = [];
    const range = getStartEndYearMonthRange(startYearMonth, endYearMonth);

    if (!range) {
        return allYearMonthTimes;
    }

    for (let year = range.startYearMonth.year, month0base = range.startYearMonth.month0base; year <= range.endYearMonth.year || month0base <= range.endYearMonth.month0base; ) {
        const yearMonth: Year0BasedMonth = {
            year: year,
            month0base: month0base
        };

        const minUnixTime = getYearMonthFirstUnixTime(yearMonth);
        const maxUnixTime = getYearMonthLastUnixTime(yearMonth);

        allYearMonthTimes.push(YearMonthUnixTime.of(yearMonth, minUnixTime, maxUnixTime));

        if (year === range.endYearMonth.year && month0base === range.endYearMonth.month0base) {
            break;
        }

        if (month0base >= 11) {
            year++;
            month0base = 0;
        } else {
            month0base++;
        }
    }

    return allYearMonthTimes;
}

export function getAllBillingCyclesStartAndEndUnixTimes(startUnixTime: number, endUnixTime: number, statementDate: number): YearMonthUnixTime[] {
    const allYearMonthTimes: YearMonthUnixTime[] = [];

    if (!startUnixTime || !endUnixTime) {
        return allYearMonthTimes;
    }

    let unixTime: number = startUnixTime;

    while (unixTime <= endUnixTime) {
        const currentDateTime = parseDateTimeFromUnixTime(unixTime);
        let currentBillingCycleMinDateTime: DateTime;

        if (currentDateTime.getGregorianCalendarDay() > statementDate) {
            const currentMonthMinDateTime = getMonthFirstDateTimeBySpecifiedUnixTime(unixTime);
            currentBillingCycleMinDateTime = currentMonthMinDateTime.add(statementDate, 'days');
        } else {
            const currentMonthMinDateTime = getMonthFirstDateTimeBySpecifiedUnixTime(unixTime);
            const previousMonthMinDateTime = currentMonthMinDateTime.add(-1, 'months');
            currentBillingCycleMinDateTime = previousMonthMinDateTime.add(statementDate, 'days');
        }

        const currentBillingCycleMaxDateTime = currentBillingCycleMinDateTime.add(1, 'months').subtract(1, 'seconds');
        const yearMonth: Year0BasedMonth = {
            year: currentBillingCycleMaxDateTime.getGregorianCalendarYear(),
            month0base: currentBillingCycleMaxDateTime.getGregorianCalendarMonth() - 1
        };

        allYearMonthTimes.push(YearMonthUnixTime.of(yearMonth, currentBillingCycleMinDateTime.getUnixTime(), currentBillingCycleMaxDateTime.getUnixTime()));
        unixTime = currentBillingCycleMaxDateTime.getUnixTime() + 1;
    }

    return allYearMonthTimes;
}

export function getAllDaysStartAndEndUnixTimes(startUnixTime: number, endUnixTime: number): YearMonthDayUnixTime[] {
    const allYearMonthDayTimes: YearMonthDayUnixTime[] = [];

    if (!startUnixTime || !endUnixTime) {
        return allYearMonthDayTimes;
    }

    let unixTime: number = startUnixTime;

    while (unixTime <= endUnixTime) {
        const currentDateTime = parseDateTimeFromUnixTime(unixTime);
        const currentDayMinDateTime = getDayFirstDateTimeBySpecifiedUnixTime(unixTime);
        const currentDayMaxDateTime = getDayLastDateTimeBySpecifiedUnixTime(unixTime);

        allYearMonthDayTimes.push(YearMonthDayUnixTime.of(currentDateTime.toGregorianCalendarYearMonthDay(), currentDayMinDateTime.getUnixTime(), currentDayMaxDateTime.getUnixTime()));
        unixTime = currentDayMaxDateTime.getUnixTime() + 1;
    }

    return allYearMonthDayTimes;
}
