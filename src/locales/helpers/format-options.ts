import { ALL_CURRENCIES } from '@/consts/currency.ts';
import {
    CalendarType
} from '@/core/calendar.ts';
import {
    CurrencyDisplayType
} from '@/core/currency.ts';
import {
    type CalendarAlternateDate,
    type DateTime,
    type DateTimeFormatOptions,
    type UnixTimeRange
} from '@/core/datetime.ts';
import {
    FiscalYearFormat,
    FiscalYearUnixTime
} from '@/core/fiscalyear.ts';
import {
    type NumberFormatOptions,
    DigitGroupingType,
    NumeralSystem
} from '@/core/numeral.ts';
import {
    isDefined
} from '@/lib/common.ts';
import {
    getCurrencyFraction
} from '@/lib/currency.ts';
import {
    formatUnixTime
} from '@/lib/datetime.ts';
import type { useHelpersCalendarLocales } from './calendar-locales.ts';
import type { useHelpersMonthNames } from './month-names.ts';
import type { useHelpersPreferences } from './preferences.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    getCurrentNumeralSystemType: ReturnType<typeof useHelpersPreferences>['getCurrentNumeralSystemType'];
    getCurrentDateDisplayType: ReturnType<typeof useHelpersPreferences>['getCurrentDateDisplayType'];
    getDateTimeLocaleData: ReturnType<typeof useHelpersCalendarLocales>['getDateTimeLocaleData'];
    getChineseCalendarLocaleData: ReturnType<typeof useHelpersCalendarLocales>['getChineseCalendarLocaleData'];
    getPersianCalendarLocaleData: ReturnType<typeof useHelpersCalendarLocales>['getPersianCalendarLocaleData'];
    getCurrentDigitGroupingType: ReturnType<typeof useHelpersPreferences>['getCurrentDigitGroupingType'];
    getCurrentDecimalSeparator: ReturnType<typeof useHelpersPreferences>['getCurrentDecimalSeparator'];
    getCurrentDigitGroupingSymbol: ReturnType<typeof useHelpersPreferences>['getCurrentDigitGroupingSymbol'];
    t: ReturnType<typeof useHelpersState>['t'];
    getQuarterName: ReturnType<typeof useHelpersMonthNames>['getQuarterName'];
    userStore: ReturnType<typeof useHelpersState>['userStore'];
}

// Owns format options for the helpers store.
export function useHelpersFormatOptions(context: Context) {
    const {
        getCurrentNumeralSystemType, getCurrentDateDisplayType, getDateTimeLocaleData, getChineseCalendarLocaleData,
        getPersianCalendarLocaleData, getCurrentDigitGroupingType, getCurrentDecimalSeparator, getCurrentDigitGroupingSymbol, t,
        getQuarterName, userStore,
    } = context;

    function getDateTimeFormatOptions(options?: { calendarType?: CalendarType, numeralSystem?: NumeralSystem }): DateTimeFormatOptions {
        let numeralSystem: NumeralSystem | undefined = options?.numeralSystem;
        let calendarType: CalendarType | undefined = options?.calendarType;

        if (!isDefined(numeralSystem)) {
            numeralSystem = getCurrentNumeralSystemType();
        }

        if (!isDefined(calendarType)) {
            calendarType = getCurrentDateDisplayType().calendarType;
        }

        return {
            numeralSystem: numeralSystem,
            calendarType: calendarType,
            localeData: getDateTimeLocaleData(),
            chineseCalendarLocaleData: getChineseCalendarLocaleData(),
            persianCalendarLocaleData: getPersianCalendarLocaleData()
        };
    }

    function getNumberFormatOptions({numeralSystem, digitGrouping, decimalSeparator, currencyCode}: {
        numeralSystem?: NumeralSystem,
        digitGrouping?: DigitGroupingType,
        decimalSeparator?: string,
        currencyCode?: string
    }): NumberFormatOptions {
        if (!isDefined(numeralSystem)) {
            numeralSystem = getCurrentNumeralSystemType();
        }

        if (!isDefined(digitGrouping)) {
            digitGrouping = getCurrentDigitGroupingType();
        }

        if (!isDefined(decimalSeparator)) {
            decimalSeparator = getCurrentDecimalSeparator();
        }

        return {
            numeralSystem: numeralSystem,
            digitGrouping: digitGrouping,
            digitGroupingSymbol: getCurrentDigitGroupingSymbol(),
            decimalSeparator: decimalSeparator,
            decimalNumberCount: getCurrencyFraction(currencyCode),
        };
    }

    function getCurrencyUnitName(currencyCode: string, isPlural: boolean): string {
        const currencyInfo = ALL_CURRENCIES[currencyCode];

        if (currencyInfo && currencyInfo.unit) {
            if (isPlural) {
                return t(`currency.unit.${currencyInfo.unit}.plural`);
            } else {
                return t(`currency.unit.${currencyInfo.unit}.normal`);
            }
        }

        return '';
    }

    function isGregorianLikeCalendarType(calendarType: CalendarType): boolean {
        return calendarType === CalendarType.Gregorian || calendarType === CalendarType.Buddhist;
    }

    function getGregorianLikeCalendarType(): CalendarType {
        const currentDateDisplayType = getCurrentDateDisplayType();

        if (isGregorianLikeCalendarType(currentDateDisplayType.calendarType)) {
            return currentDateDisplayType.calendarType;
        }

        return CalendarType.Gregorian;
    }

    function formatYearQuarter(year: string, quarter: number): string {
        const quarterName = getQuarterName(quarter);

        if (quarterName) {
            return t('format.yearQuarter.content', {
                year: year,
                quarter: getQuarterName(quarter)
            });
        } else {
            return '';
        }
    }

    function formatTimeRangeToGregorianLikeFiscalYearFormat(format: FiscalYearFormat, timeRange: FiscalYearUnixTime | UnixTimeRange, numeralSystem?: NumeralSystem, calendarType?: CalendarType): string {
        if (!format) {
            format = FiscalYearFormat.Default;
        }

        if (!isDefined(calendarType)) {
            calendarType = getGregorianLikeCalendarType();
        } else if (!isGregorianLikeCalendarType(calendarType)) {
            calendarType = CalendarType.Gregorian;
        }

        const dateTimeFormatOptions = getDateTimeFormatOptions({
            calendarType: calendarType,
            numeralSystem: numeralSystem
        });

        return t('format.fiscalYear.' + format.typeName, {
            StartYYYY: formatUnixTime(timeRange.minUnixTime, 'YYYY', dateTimeFormatOptions),
            StartYY: formatUnixTime(timeRange.minUnixTime, 'YY', dateTimeFormatOptions),
            EndYYYY: formatUnixTime(timeRange.maxUnixTime, 'YYYY', dateTimeFormatOptions),
            EndYY: formatUnixTime(timeRange.maxUnixTime, 'YY', dateTimeFormatOptions),
        });
    }

    function getCurrentCurrencyDisplayType(): CurrencyDisplayType {
        let currencyDisplayType = CurrencyDisplayType.valueOf(userStore.currentUserCurrencyDisplayType);

        if (!currencyDisplayType) {
            const defaultCurrencyDisplayTypeName = t('default.currencyDisplayType');
            currencyDisplayType = CurrencyDisplayType.parse(defaultCurrencyDisplayTypeName);
        }

        if (!currencyDisplayType) {
            currencyDisplayType = CurrencyDisplayType.Default;
        }

        return currencyDisplayType;
    }

    function getCalendarAlternateDisplayDate(dateTime: DateTime, dateTimeFormatOptions: DateTimeFormatOptions): CalendarAlternateDate {
        const numeralSystem = getCurrentNumeralSystemType();
        let displayDate = numeralSystem.replaceWesternArabicDigitsToLocalizedDigits(dateTime.getLocalizedCalendarDay(dateTimeFormatOptions));

        if (dateTime.isLocalizedCalendarFirstDayOfMonth(dateTimeFormatOptions)) {
            displayDate = dateTime.getLocalizedCalendarMonthDisplayShortName(dateTimeFormatOptions);
        }

        const alternateDate: CalendarAlternateDate = {
            year: dateTime.getGregorianCalendarYear(),
            month: dateTime.getGregorianCalendarMonth(),
            day: dateTime.getGregorianCalendarDay(),
            displayDate: displayDate
        };

        return alternateDate;
    }

    return {
        getDateTimeFormatOptions, getNumberFormatOptions, getCurrencyUnitName, getGregorianLikeCalendarType, formatYearQuarter,
        formatTimeRangeToGregorianLikeFiscalYearFormat, getCurrentCurrencyDisplayType, getCalendarAlternateDisplayDate,
    };
}
