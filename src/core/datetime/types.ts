import type { TypeAndDisplayName } from '@/core/base.ts';
import type { CalendarType, ChineseCalendarLocaleData, PersianCalendarLocaleData } from '@/core/calendar.ts';
import type { NumeralSystem } from '@/core/numeral.ts';
import type { WeekDay } from './calendar.ts';

export type DateTimeUnit = 'years' | 'months' | 'days' | 'hours' | 'minutes' | 'seconds';

export interface DateTimeSetObject {
    year?: number;
    month?: number;
    dayOfMonth?: number;
    hour?: number;
    minute?: number;
    second?: number;
    millisecond?: number;
}

export interface DateTime {
    getUnixTime(): number;
    getLocalizedCalendarYear(options: DateTimeFormatOptions): string;
    getGregorianCalendarYear(): number;
    getGregorianCalendarQuarter(): number;
    getLocalizedCalendarQuarter(options: DateTimeFormatOptions): number;
    getGregorianCalendarMonth(): number;
    getGregorianCalendarMonthDisplayName(options: DateTimeFormatOptions): string;
    getGregorianCalendarMonthDisplayShortName(options: DateTimeFormatOptions): string;
    getLocalizedCalendarMonth(options: DateTimeFormatOptions): string;
    getLocalizedCalendarMonthDisplayName(options: DateTimeFormatOptions): string;
    getLocalizedCalendarMonthDisplayShortName(options: DateTimeFormatOptions): string;
    getGregorianCalendarDay(): number;
    getLocalizedCalendarDay(options: DateTimeFormatOptions): string;
    isLocalizedCalendarFirstDayOfMonth(options: DateTimeFormatOptions): boolean;
    getGregorianCalendarYearDashMonthDashDay(): TextualYearMonthDay;
    getGregorianCalendarYearDashMonth(): TextualYearMonth;
    getMaxDayOfGregorianCalendarMonth(): number;
    getWeekDay(): WeekDay;
    getWeekDayDisplayName(options: DateTimeFormatOptions): string
    getWeekDayDisplayShortName(options: DateTimeFormatOptions): string;
    getWeekDayDisplayMinName(options: DateTimeFormatOptions): string;
    getHour(): number;
    getMinute(): number;
    getSecond(): number;
    getDisplayAMPM(options: DateTimeFormatOptions): string;
    getTimezoneUtcOffsetMinutes(): number;
    setTimezoneByUtcOffsetMinutes(offsetMinutes: number): DateTime;
    setTimezoneByIANATimeZoneName(zoneName: string): DateTime;
    add(amount: number, unit: DateTimeUnit): DateTime;
    subtract(amount: number, unit: DateTimeUnit): DateTime;
    set(value: DateTimeSetObject): DateTime;
    toGregorianCalendarYearMonthDay(): YearMonthDay;
    toGregorianCalendarYear0BasedMonth(): Year0BasedMonth;
    format(format: string, options: DateTimeFormatOptions): string;
}

export interface DateTimeFormatOptions {
    numeralSystem: NumeralSystem;
    calendarType: CalendarType;
    localeData: DateTimeLocaleData;
    chineseCalendarLocaleData: ChineseCalendarLocaleData;
    persianCalendarLocaleData: PersianCalendarLocaleData;
}

export interface DateTimeLocaleData {
    months: () => string[];
    monthsShort: () => string[];
    weekdays: () => string[];
    weekdaysShort: () => string[];
    weekdaysMin: () => string[];
    meridiem: (hour: number, minute: number, isLower: boolean) => string;
}

export type TextualYearMonth = `${number}-${number}`;

export type TextualMonthDay = `${number}-${number}`;

export type TextualYearMonthDay = `${number}-${number}-${number}`;

export interface YearQuarter {
    readonly year: number;
    readonly quarter: number;
}

export interface Year0BasedMonth {
    readonly year: number;
    readonly month0base: number;
}

export interface Year1BasedMonth {
    readonly year: number;
    readonly month1base: number;
}

export interface YearMonthRange {
    readonly startYearMonth: Year0BasedMonth;
    readonly endYearMonth: Year0BasedMonth;
}

export interface YearMonthDay extends MonthDay {
    readonly year: number;
    readonly month: number; // 1-based (1 = January, 12 = December)
    readonly day: number;
}

export interface MonthDay {
    readonly month: number; // 1-based (1 = January, 12 = December
    readonly day: number;
}

export interface CalendarAlternateDate extends YearMonthDay {
    readonly displayDate: string;
}

export interface TimeRange {
    readonly minTime: number;
    readonly maxTime: number;
}

export interface StartEndTime {
    readonly startTime: number;
    readonly endTime: number;
}

export interface WritableStartEndTime extends StartEndTime {
    startTime: number;
    endTime: number;
}

export interface UnixTimeRange {
    readonly minUnixTime: number;
    readonly maxUnixTime: number;
}

export interface TimeRangeAndDateType extends TimeRange {
    readonly dateType: number;
}

export interface TimeDifference {
    readonly offsetHours: number;
    readonly offsetMinutes: number;
}

export interface RecentMonthDateRange {
    readonly dateType: number;
    readonly minTime: number;
    readonly maxTime: number;
    readonly year: number;
    readonly month: number; // 1-based (1 = January, 12 = December)
}

export interface PresetDateRange {
    readonly label: string;
    readonly value: Date[];
}

export interface LocalizedDateTimeFormat extends TypeAndDisplayName {
    readonly type: number;
    readonly format: string;
    readonly displayName: string;
}

export interface LocalizedDateRange extends TypeAndDisplayName {
    readonly type: number;
    readonly displayName: string;
    readonly isBillingCycle: boolean;
    readonly isUserCustomRange: boolean;
}

export interface LocalizedRecentMonthDateRange extends TimeRangeAndDateType {
    readonly dateType: number;
    readonly minTime: number;
    readonly maxTime: number;
    readonly year?: number;
    readonly month?: number;
    readonly isPreset?: boolean;
    readonly displayName: string;
}
