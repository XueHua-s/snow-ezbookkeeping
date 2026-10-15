import {
    CalendarType
} from '@/core/calendar.ts';
import {
    type CalendarAlternateDate,
    type DateTime,
    type TextualMonthDay,
    type Year1BasedMonth,
    type YearMonthDay,
    DateRange
} from '@/core/datetime.ts';
import {
    FiscalYearFormat,
    FiscalYearStart
} from '@/core/fiscalyear.ts';
import {
    NumeralSystem
} from '@/core/numeral.ts';
import {
    type ChineseYearMonthDayInfo,
    getChineseCalendarAlternateDisplayDate,
    getChineseYearMonthAllDayInfos,
    getChineseYearMonthDayInfo
} from '@/lib/calendar/chinese_calendar.ts';
import {
    formatDateTime,
    formatGregorianCalendarMonthDashDay,
    formatUnixTime,
    getFiscalYearTimeRangeFromUnixTime,
    getFiscalYearTimeRangeFromYear,
    getGregorianCalendarYearMonthDays,
    getTimeDifferenceHoursAndMinutes,
    getTimezoneOffsetMinutes,
    getYearMonthDayDateTime,
    isDateRangeMatchFullMonths,
    isDateRangeMatchFullYears,
    parseDateTimeFromUnixTime
} from '@/lib/datetime.ts';
import type { useHelpersCalendarLocales } from './calendar-locales.ts';
import type { useHelpersDatePatterns } from './date-patterns.ts';
import type { useHelpersFormatOptions } from './format-options.ts';
import type { useHelpersPreferences } from './preferences.ts';
import type { useHelpersState } from './state.ts';
import type { useHelpersTranslation } from './translation.ts';

interface Context {
    getGregorianLikeCalendarType: ReturnType<typeof useHelpersFormatOptions>['getGregorianLikeCalendarType'];
    getLocalizedLongMonthDayFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedLongMonthDayFormat'];
    getDateTimeFormatOptions: ReturnType<typeof useHelpersFormatOptions>['getDateTimeFormatOptions'];
    getLocalizedShortYearFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedShortYearFormat'];
    formatYearQuarter: ReturnType<typeof useHelpersFormatOptions>['formatYearQuarter'];
    getCurrentFiscalYearFormatType: ReturnType<typeof useHelpersPreferences>['getCurrentFiscalYearFormatType'];
    userStore: ReturnType<typeof useHelpersState>['userStore'];
    formatTimeRangeToGregorianLikeFiscalYearFormat: ReturnType<typeof useHelpersFormatOptions>['formatTimeRangeToGregorianLikeFiscalYearFormat'];
    t: ReturnType<typeof useHelpersState>['t'];
    formatRange: ReturnType<typeof useHelpersTranslation>['formatRange'];
    getLocalizedShortYearMonthFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedShortYearMonthFormat'];
    getLocalizedShortDateFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedShortDateFormat'];
    getLocalizedShortMonthDayFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedShortMonthDayFormat'];
    getCurrentNumeralSystemType: ReturnType<typeof useHelpersPreferences>['getCurrentNumeralSystemType'];
    getCurrentCalendarDisplayType: ReturnType<typeof useHelpersPreferences>['getCurrentCalendarDisplayType'];
    getChineseCalendarLocaleData: ReturnType<typeof useHelpersCalendarLocales>['getChineseCalendarLocaleData'];
    getCalendarAlternateDisplayDate: ReturnType<typeof useHelpersFormatOptions>['getCalendarAlternateDisplayDate'];
}

// Owns date formatting for the helpers store.
export function useHelpersDateFormatting(context: Context) {
    const {
        getGregorianLikeCalendarType, getLocalizedLongMonthDayFormat, getDateTimeFormatOptions, getLocalizedShortYearFormat,
        formatYearQuarter, getCurrentFiscalYearFormatType, userStore, formatTimeRangeToGregorianLikeFiscalYearFormat, t,
        formatRange, getLocalizedShortYearMonthFormat, getLocalizedShortDateFormat, getLocalizedShortMonthDayFormat,
        getCurrentNumeralSystemType, getCurrentCalendarDisplayType, getChineseCalendarLocaleData, getCalendarAlternateDisplayDate,
    } = context;

    function formatGregorianTextualMonthDayToGregorianLikeLongMonthDay(monthDay: TextualMonthDay, numeralSystem?: NumeralSystem): string {
        const gregorianLikeCalendarType = getGregorianLikeCalendarType();
        return formatGregorianCalendarMonthDashDay(monthDay, getLocalizedLongMonthDayFormat(), getDateTimeFormatOptions({ calendarType: gregorianLikeCalendarType, numeralSystem: numeralSystem }));
    }

    function formatDateTimeToGregorianLikeYearQuarter(dateTime: DateTime): string {
        const gregorianLikeCalendarType = getGregorianLikeCalendarType();
        const dateTimeFormatOptions = getDateTimeFormatOptions({ calendarType: gregorianLikeCalendarType });
        const year = formatDateTime(dateTime, getLocalizedShortYearFormat(), dateTimeFormatOptions);
        const quarter = dateTime.getLocalizedCalendarQuarter(dateTimeFormatOptions);
        return formatYearQuarter(year, quarter);
    }

    function formatYearQuarterToGregorianLikeYearQuarter(year: number, quarter: number): string {
        const gregorianLikeCalendarType = getGregorianLikeCalendarType();
        const dateTimeFormatOptions = getDateTimeFormatOptions({ calendarType: gregorianLikeCalendarType });
        const date = getYearMonthDayDateTime(year, 1, 1);
        const textualYear = date.getLocalizedCalendarYear(dateTimeFormatOptions);
        return formatYearQuarter(textualYear, quarter);
    }

    function formatUnixTimeToGregorianLikeFiscalYear(unixTime: number): string {
        let fiscalYearFormat = FiscalYearFormat.valueOf(getCurrentFiscalYearFormatType());

        if (!fiscalYearFormat) {
            fiscalYearFormat = FiscalYearFormat.Default;
        }

        const timeRange = getFiscalYearTimeRangeFromUnixTime(unixTime, userStore.currentUserFiscalYearStart);
        return formatTimeRangeToGregorianLikeFiscalYearFormat(fiscalYearFormat, timeRange);
    }

    function formatGregorianYearToGregorianLikeFiscalYear(year: number) {
        let fiscalYearFormat = FiscalYearFormat.valueOf(getCurrentFiscalYearFormatType());

        if (!fiscalYearFormat) {
            fiscalYearFormat = FiscalYearFormat.Default;
        }

        const timeRange = getFiscalYearTimeRangeFromYear(year, userStore.currentUserFiscalYearStart);
        return formatTimeRangeToGregorianLikeFiscalYearFormat(fiscalYearFormat, timeRange);
    }

    function formatFiscalYearStartToGregorianLikeLongMonth(fiscalYearStartValue: number, numeralSystem?: NumeralSystem): string {
        let fiscalYearStart = FiscalYearStart.valueOf(fiscalYearStartValue);

        if (!fiscalYearStart) {
            fiscalYearStart = FiscalYearStart.Default;
        }

        return formatGregorianTextualMonthDayToGregorianLikeLongMonthDay(fiscalYearStart.toMonthDashDayString(), numeralSystem);
    }

    function formatDateRange(dateType: number, startTime: number, endTime: number): string {
        if (dateType === DateRange.All.type) {
            return t(DateRange.All.name);
        }

        const allDateRanges = DateRange.values();
        const gregorianLikeCalendarType = getGregorianLikeCalendarType();
        const dateTimeFormatOptions = getDateTimeFormatOptions();
        const gregorianLikeDateTimeFormatOptions = getDateTimeFormatOptions({ calendarType: gregorianLikeCalendarType });

        for (const dateRange of allDateRanges) {
            if (dateRange && dateRange.type !== DateRange.Custom.type && dateRange.type === dateType && dateRange.name) {
                return t(dateRange.name);
            }
        }

        if (isDateRangeMatchFullYears(startTime, endTime)) {
            const format = getLocalizedShortYearFormat();
            const displayStartTime = formatUnixTime(startTime, format, gregorianLikeDateTimeFormatOptions);
            const displayEndTime = formatUnixTime(endTime, format, gregorianLikeDateTimeFormatOptions);

            return displayStartTime !== displayEndTime ? formatRange(displayStartTime, displayEndTime) : displayStartTime;
        }

        if (isDateRangeMatchFullMonths(startTime, endTime)) {
            const format = getLocalizedShortYearMonthFormat();
            const displayStartTime = formatUnixTime(startTime, format, gregorianLikeDateTimeFormatOptions);
            const displayEndTime = formatUnixTime(endTime, format, gregorianLikeDateTimeFormatOptions);

            return displayStartTime !== displayEndTime ? formatRange(displayStartTime, displayEndTime) : displayStartTime;
        }

        const startTimeYear = parseDateTimeFromUnixTime(startTime).getLocalizedCalendarYear(gregorianLikeDateTimeFormatOptions);
        const endTimeYear = parseDateTimeFromUnixTime(endTime).getLocalizedCalendarYear(gregorianLikeDateTimeFormatOptions);

        const format = getLocalizedShortDateFormat();
        const displayStartTime = formatUnixTime(startTime, format, dateTimeFormatOptions);
        const displayEndTime = formatUnixTime(endTime, format, dateTimeFormatOptions);

        if (displayStartTime === displayEndTime) {
            return displayStartTime;
        } else if (startTimeYear === endTimeYear) {
            const displayShortEndTime = formatUnixTime(endTime, getLocalizedShortMonthDayFormat(), gregorianLikeDateTimeFormatOptions);
            return formatRange(displayStartTime, displayShortEndTime);
        }

        return formatRange(displayStartTime, displayEndTime);
    }

    function getTimezoneDifferenceDisplayText(unixTime: number, utcOffset: number): string {
        const numeralSystem = getCurrentNumeralSystemType();
        const defaultTimezoneOffset = getTimezoneOffsetMinutes(unixTime);
        const offsetTime = getTimeDifferenceHoursAndMinutes(utcOffset - defaultTimezoneOffset);

        if (utcOffset > defaultTimezoneOffset) {
            if (offsetTime.offsetMinutes) {
                return t('format.misc.hoursMinutesAheadOfDefaultTimezone', {
                    hours: numeralSystem.formatNumber(offsetTime.offsetHours),
                    minutes: numeralSystem.formatNumber(offsetTime.offsetMinutes)
                });
            } else {
                return t('format.misc.hoursAheadOfDefaultTimezone', {
                    hours: numeralSystem.formatNumber(offsetTime.offsetHours)
                });
            }
        } else if (utcOffset < defaultTimezoneOffset) {
            if (offsetTime.offsetMinutes) {
                return t('format.misc.hoursMinutesBehindDefaultTimezone', {
                    hours: numeralSystem.formatNumber(offsetTime.offsetHours),
                    minutes: numeralSystem.formatNumber(offsetTime.offsetMinutes)
                });
            } else {
                return t('format.misc.hoursBehindDefaultTimezone', {
                    hours: numeralSystem.formatNumber(offsetTime.offsetHours)
                });
            }
        } else {
            return t('Same time as default timezone');
        }
    }

    function getCalendarAlternateDates(yearMonth: Year1BasedMonth): CalendarAlternateDate[] | undefined {
        const calendarDisplayType = getCurrentCalendarDisplayType().secondaryCalendarType;

        if (!calendarDisplayType) {
            return undefined;
        }

        const dateTimeFormatOptions = getDateTimeFormatOptions({ calendarType: calendarDisplayType });

        if (calendarDisplayType === CalendarType.Chinese) {
            const chineseCalendarLocaleData = getChineseCalendarLocaleData();
            const chineseDates: ChineseYearMonthDayInfo[] | undefined = getChineseYearMonthAllDayInfos(yearMonth, chineseCalendarLocaleData);

            if (!chineseDates) {
                return undefined;
            }

            const ret: CalendarAlternateDate[] = [];

            for (const chineseDate of chineseDates) {
                const alternateDate = getChineseCalendarAlternateDisplayDate(chineseDate);
                ret.push(alternateDate);
            }

            return ret;
        } else if (calendarDisplayType === CalendarType.Persian) {
            const monthDays: number = getGregorianCalendarYearMonthDays(yearMonth);
            const ret: CalendarAlternateDate[] = [];

            for (let i = 1; i <= monthDays; i++) {
                const dateTime = getYearMonthDayDateTime(yearMonth.year, yearMonth.month1base, i);
                ret.push(getCalendarAlternateDisplayDate(dateTime, dateTimeFormatOptions));
            }

            return ret;
        }

        return undefined;
    }

    function getCalendarAlternateDate(yearMonthDay: YearMonthDay): CalendarAlternateDate | undefined {
        const calendarDisplayType = getCurrentCalendarDisplayType().secondaryCalendarType;

        if (!calendarDisplayType) {
            return undefined;
        }

        const dateTimeFormatOptions = getDateTimeFormatOptions({ calendarType: calendarDisplayType });

        if (calendarDisplayType === CalendarType.Chinese) {
            const chineseCalendarLocaleData = getChineseCalendarLocaleData();
            const chineseDate: ChineseYearMonthDayInfo | undefined = getChineseYearMonthDayInfo(yearMonthDay, chineseCalendarLocaleData);

            if (!chineseDate) {
                return undefined;
            }

            return getChineseCalendarAlternateDisplayDate(chineseDate);
        } else if (calendarDisplayType === CalendarType.Persian) {
            const dateTime = getYearMonthDayDateTime(yearMonthDay.year, yearMonthDay.month, yearMonthDay.day);
            return getCalendarAlternateDisplayDate(dateTime, dateTimeFormatOptions);
        }

        return undefined;
    }

    return {
        formatGregorianTextualMonthDayToGregorianLikeLongMonthDay, formatDateTimeToGregorianLikeYearQuarter,
        formatYearQuarterToGregorianLikeYearQuarter, formatUnixTimeToGregorianLikeFiscalYear,
        formatGregorianYearToGregorianLikeFiscalYear, formatFiscalYearStartToGregorianLikeLongMonth, formatDateRange,
        getTimezoneDifferenceDisplayText, getCalendarAlternateDates, getCalendarAlternateDate,
    };
}
