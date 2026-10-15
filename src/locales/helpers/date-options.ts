import { ALL_TIMEZONES, UTC_TIMEZONE } from '@/consts/timezone.ts';
import {
    type TypeAndDisplayName
} from '@/core/base.ts';
import {
    CalendarType
} from '@/core/calendar.ts';
import {
    type DateFormat,
    type DateTime,
    type DateTimeFormatOptions,
    type LocalizedDateRange,
    type LocalizedDateTimeFormat,
    type LocalizedRecentMonthDateRange,
    type TimeFormat,
    type WeekDayValue,
    DateRange,
    DateRangeScene,
    LANGUAGE_DEFAULT_DATE_TIME_FORMAT_VALUE,
    Month,
    WeekDay
} from '@/core/datetime.ts';
import {
    FiscalYearFormat,
    FiscalYearStart,
    LANGUAGE_DEFAULT_FISCAL_YEAR_FORMAT_VALUE
} from '@/core/fiscalyear.ts';
import {
    NumeralSystem
} from '@/core/numeral.ts';
import {
    type LocalizedTimezoneInfo,
    TimezoneTypeForStatistics
} from '@/core/timezone.ts';
import {
    isNumber
} from '@/lib/common.ts';
import {
    formatCurrentTime,
    formatUnixTime,
    getBrowserTimezoneOffset,
    getBrowserTimezoneOffsetMinutes,
    getCurrentDateTime,
    getCurrentUnixTime,
    getFiscalYearTimeRangeFromUnixTime,
    getRecentMonthDateRanges,
    getTimezoneOffset,
    getTimezoneOffsetMinutes
} from '@/lib/datetime.ts';
import type { useHelpersDatePatterns } from './date-patterns.ts';
import type { useHelpersFormatOptions } from './format-options.ts';
import type { useHelpersMonthNames } from './month-names.ts';
import type { useHelpersPreferences } from './preferences.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    getAllMonthNames: ReturnType<typeof useHelpersMonthNames>['getAllMonthNames'];
    getAllWeekdayNames: ReturnType<typeof useHelpersMonthNames>['getAllWeekdayNames'];
    t: ReturnType<typeof useHelpersState>['t'];
    getLocalizedShortTimeFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedShortTimeFormat'];
    getDateTimeFormatOptions: ReturnType<typeof useHelpersFormatOptions>['getDateTimeFormatOptions'];
    getMonthdayShortName: ReturnType<typeof useHelpersMonthNames>['getMonthdayShortName'];
    getMonthLastDayShortName: ReturnType<typeof useHelpersMonthNames>['getMonthLastDayShortName'];
    getLocalizedDateTimeFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedDateTimeFormat'];
    getCurrentCalendarDisplayType: ReturnType<typeof useHelpersPreferences>['getCurrentCalendarDisplayType'];
    getLocalizedLongYearMonthFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedLongYearMonthFormat'];
    getCurrentNumeralSystemType: ReturnType<typeof useHelpersPreferences>['getCurrentNumeralSystemType'];
    userStore: ReturnType<typeof useHelpersState>['userStore'];
    formatTimeRangeToGregorianLikeFiscalYearFormat: ReturnType<typeof useHelpersFormatOptions>['formatTimeRangeToGregorianLikeFiscalYearFormat'];
}

// Owns date options for the helpers store.
export function useHelpersDateOptions(context: Context) {
    const {
        getAllMonthNames, getAllWeekdayNames, t, getLocalizedShortTimeFormat, getDateTimeFormatOptions, getMonthdayShortName,
        getMonthLastDayShortName, getLocalizedDateTimeFormat, getCurrentCalendarDisplayType, getLocalizedLongYearMonthFormat,
        getCurrentNumeralSystemType, userStore, formatTimeRangeToGregorianLikeFiscalYearFormat,
    } = context;

    function getAllLongMonthNames(): string[] {
        return getAllMonthNames('long');
    }

    function getAllShortMonthNames(): string[] {
        return getAllMonthNames('short');
    }

    function getAllLongWeekdayNames(): string[] {
        return getAllWeekdayNames('long');
    }

    function getAllShortWeekdayNames(): string[] {
        return getAllWeekdayNames('short');
    }

    function getAllMinWeekdayNames(): string[] {
        return getAllWeekdayNames('min');
    }

    function getAllMonths(): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];
        const allMonths = Month.values();

        for (const month of allMonths) {
            ret.push({
                type: month.month,
                displayName: t(`datetime.${month.name}.long`)
            });
        }

        return ret;
    }

    function getAllWeekDays(firstDayOfWeek?: WeekDayValue): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];
        const allWeekDays = WeekDay.values();

        if (!isNumber(firstDayOfWeek)) {
            firstDayOfWeek = WeekDay.DefaultFirstDay.type;
        }

        for (let i = firstDayOfWeek; i < allWeekDays.length; i++) {
            const weekDay = allWeekDays[i] as WeekDay;

            ret.push({
                type: weekDay.type,
                displayName: t(`datetime.${weekDay.name}.long`)
            });
        }

        for (let i = 0; i < firstDayOfWeek; i++) {
            const weekDay = allWeekDays[i] as WeekDay;

            ret.push({
                type: weekDay.type,
                displayName: t(`datetime.${weekDay.name}.long`)
            });
        }

        return ret;
    }

    function getAllHours(): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];
        const now: DateTime = getCurrentDateTime();
        const format: string = getLocalizedShortTimeFormat();
        const options: DateTimeFormatOptions = getDateTimeFormatOptions();

        for (let i = 0; i < 24; i++) {
            const dateTime = now.set({
                hour: i,
                minute: 0,
                second: 0,
                millisecond: 0
            });

            ret.push({
                type: i,
                displayName: dateTime.format(format, options)
            });
        }

        return ret;
    }

    function getAvailableMonthDays(daysInMonth: number, lastDaysOfMonth?: number): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];

        for (let i = 1; i <= daysInMonth; i++) {
            ret.push({
                type: i,
                displayName: getMonthdayShortName(i),
            });
        }

        if (isNumber(lastDaysOfMonth) && lastDaysOfMonth > 0) {
            for (let i = -lastDaysOfMonth; i < 0; i++) {
                ret.push({
                    type: i,
                    displayName: (i === -1) ? t('Last day') : getMonthLastDayShortName(-i),
                });
            }
        }

        return ret;
    }

    function getLocalizedDateTimeFormats<T extends DateFormat | TimeFormat>(type: string, allFormatMap: Record<string, T>, allFormatArray: T[], languageDefaultTypeNameKey: string, systemDefaultFormatType: T, numeralSystem: NumeralSystem, calendarType?: CalendarType): LocalizedDateTimeFormat[] {
        const defaultFormat = getLocalizedDateTimeFormat<T>(type, allFormatMap, allFormatArray, LANGUAGE_DEFAULT_DATE_TIME_FORMAT_VALUE, languageDefaultTypeNameKey, systemDefaultFormatType);
        const ret: LocalizedDateTimeFormat[] = [];
        const dateTimeFormatOptions = getDateTimeFormatOptions({ numeralSystem, calendarType });

        ret.push({
            type: LANGUAGE_DEFAULT_DATE_TIME_FORMAT_VALUE,
            format: defaultFormat,
            displayName: `${t('Language Default')} (${formatCurrentTime(defaultFormat, dateTimeFormatOptions)})`
        });

        for (const formatType of allFormatArray) {
            const format = t(`format.${type}.${formatType.typeName}`);

            ret.push({
                type: formatType.type,
                format: format,
                displayName: formatCurrentTime(format, dateTimeFormatOptions)
            });
        }

        return ret;
    }

    function getAllDateRanges(scene: DateRangeScene, { includeCustom, includeBillingCycle, includeLastReconciledTimeRange } : { includeCustom?: boolean, includeBillingCycle?: boolean, includeLastReconciledTimeRange?: boolean }): LocalizedDateRange[] {
        const ret: LocalizedDateRange[] = [];
        const allDateRanges = DateRange.values();

        for (const dateRange of allDateRanges) {
            const shouldSkip: boolean = !dateRange.isAvailableForScene(scene)
                || (dateRange.isBillingCycle && !includeBillingCycle)
                || (dateRange.isLastReconciledTimeRange && !includeLastReconciledTimeRange)
                || (dateRange.type === DateRange.Custom.type && !includeCustom);

            if (shouldSkip) {
                continue;
            }

            ret.push({
                type: dateRange.type,
                displayName: t(dateRange.name),
                isBillingCycle: dateRange.isBillingCycle,
                isUserCustomRange: dateRange.isUserCustomRange
            });
        }

        return ret;
    }

    function getAllRecentMonthDateRanges(includeAll: boolean, includeCustom: boolean): LocalizedRecentMonthDateRange[] {
        const allRecentMonthDateRanges: LocalizedRecentMonthDateRange[] = [];
        const recentDateRanges = getRecentMonthDateRanges(12);
        const currentCalendarDisplayType = getCurrentCalendarDisplayType();
        const dateTimeFormatOptions = getDateTimeFormatOptions({ calendarType: currentCalendarDisplayType.primaryCalendarType });

        if (includeAll) {
            allRecentMonthDateRanges.push({
                dateType: DateRange.All.type,
                minTime: 0,
                maxTime: 0,
                displayName: t('All')
            });
        }

        for (const recentDateRange of recentDateRanges) {
            allRecentMonthDateRanges.push({
                dateType: recentDateRange.dateType,
                minTime: recentDateRange.minTime,
                maxTime: recentDateRange.maxTime,
                year: recentDateRange.year,
                month: recentDateRange.month,
                isPreset: true,
                displayName: formatUnixTime(recentDateRange.minTime, getLocalizedLongYearMonthFormat(), dateTimeFormatOptions)
            });
        }

        if (includeCustom) {
            allRecentMonthDateRanges.push({
                dateType: DateRange.Custom.type,
                minTime: 0,
                maxTime: 0,
                displayName: t('Custom Date')
            });
        }

        return allRecentMonthDateRanges;
    }

    function getAllTimezones(unixTime: number, includeSystemDefault?: boolean): LocalizedTimezoneInfo[] {
        const numeralSystem = getCurrentNumeralSystemType();
        const defaultTimezoneOffset = numeralSystem.replaceWesternArabicDigitsToLocalizedDigits(getBrowserTimezoneOffset(unixTime));
        const defaultTimezoneOffsetMinutes = getBrowserTimezoneOffsetMinutes(unixTime);
        const allTimezoneInfos: LocalizedTimezoneInfo[] = [];

        for (const timezoneInfo of ALL_TIMEZONES) {
            const utcOffset = (timezoneInfo.timezoneName !== UTC_TIMEZONE.timezoneName ? numeralSystem.replaceWesternArabicDigitsToLocalizedDigits(getTimezoneOffset(unixTime, timezoneInfo.timezoneName)) : '');
            const displayName = t(`timezone.${timezoneInfo.displayName}`);

            allTimezoneInfos.push({
                name: timezoneInfo.timezoneName,
                utcOffset: utcOffset,
                utcOffsetMinutes: getTimezoneOffsetMinutes(unixTime, timezoneInfo.timezoneName),
                displayName: displayName,
                displayNameWithUtcOffset: `(UTC${utcOffset}) ${displayName}`
            });
        }

        if (includeSystemDefault) {
            const defaultDisplayName = t('System Default');

            allTimezoneInfos.push({
                name: '',
                utcOffset: defaultTimezoneOffset,
                utcOffsetMinutes: defaultTimezoneOffsetMinutes,
                displayName: defaultDisplayName,
                displayNameWithUtcOffset: `(UTC${defaultTimezoneOffset}) ${defaultDisplayName}`
            });
        }

        allTimezoneInfos.sort(function (c1, c2) {
            const utcOffsetMinutes1 = c1.utcOffsetMinutes;
            const utcOffsetMinutes2 = c2.utcOffsetMinutes;

            if (utcOffsetMinutes1 !== utcOffsetMinutes2) {
                return utcOffsetMinutes1 - utcOffsetMinutes2;
            }

            return c1.displayName.localeCompare(c2.displayName);
        })

        return allTimezoneInfos;
    }

    function getAllTimezoneTypesUsedForStatistics(currentTimezone?: string): TypeAndDisplayName[] {
        const numeralSystem = getCurrentNumeralSystemType();
        const currentTimezoneOffset = numeralSystem.replaceWesternArabicDigitsToLocalizedDigits(getTimezoneOffset(getCurrentUnixTime(), currentTimezone));

        return [
            {
                type: TimezoneTypeForStatistics.ApplicationTimezone.type,
                displayName: t(TimezoneTypeForStatistics.ApplicationTimezone.name) + ` (UTC${currentTimezoneOffset})`
            },
            {
                type: TimezoneTypeForStatistics.TransactionTimezone.type,
                displayName: t(TimezoneTypeForStatistics.TransactionTimezone.name)
            }
        ];
    }

    function getAllFiscalYearFormats(numeralSystem: NumeralSystem, calendarType: CalendarType): TypeAndDisplayName[] {
        const now = getCurrentUnixTime();
        let fiscalYearStart = userStore.currentUserFiscalYearStart;

        if (!fiscalYearStart) {
            fiscalYearStart = FiscalYearStart.Default.value;
        }

        const currentFiscalYearRange = getFiscalYearTimeRangeFromUnixTime(now, fiscalYearStart);
        let defaultFiscalYearFormat = FiscalYearFormat.parse(t('default.fiscalYearFormat'));

        if (!defaultFiscalYearFormat) {
            defaultFiscalYearFormat = FiscalYearFormat.Default;
        }

        const ret: TypeAndDisplayName[] = [];

        ret.push({
            type: LANGUAGE_DEFAULT_FISCAL_YEAR_FORMAT_VALUE,
            displayName: `${t('Language Default')} (${formatTimeRangeToGregorianLikeFiscalYearFormat(defaultFiscalYearFormat, currentFiscalYearRange, numeralSystem, calendarType)})`
        });

        const allFiscalYearFormats = FiscalYearFormat.values();

        for (const fiscalYearFormat of allFiscalYearFormats) {
            ret.push({
                type: fiscalYearFormat.type,
                displayName: formatTimeRangeToGregorianLikeFiscalYearFormat(fiscalYearFormat, currentFiscalYearRange, numeralSystem, calendarType),
            });
        }

        return ret;
    }

    return {
        getAllLongMonthNames, getAllShortMonthNames, getAllLongWeekdayNames, getAllShortWeekdayNames, getAllMinWeekdayNames,
        getAllMonths, getAllWeekDays, getAllHours, getAvailableMonthDays, getLocalizedDateTimeFormats, getAllDateRanges,
        getAllRecentMonthDateRanges, getAllTimezones, getAllTimezoneTypesUsedForStatistics, getAllFiscalYearFormats,
    };
}
