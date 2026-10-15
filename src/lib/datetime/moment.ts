import {
    EASTERNMOST_TIMEZONE_UTC_OFFSET,
    WESTERNMOST_TIMEZONE_UTC_OFFSET,
} from '@/consts/timezone.ts';
import {
    type ChineseCalendarLocaleData,
    CalendarType
} from '@/core/calendar.ts';
import {
    type DateTime,
    type DateTimeFormatOptions,
    type DateTimeSetObject,
    type DateTimeUnit,
    type TextualYearMonth,
    type TextualYearMonthDay,
    type Year0BasedMonth,
    type YearMonthDay,
    WeekDay
} from '@/core/datetime.ts';
import {
    NumeralSystem
} from '@/core/numeral.ts';
import {
    type ChineseYearMonthDayInfo,
    getChineseYearMonthDayInfo
} from '@/lib/calendar/chinese_calendar.ts';
import {
    isDefined,
    isFunction,
    isNumber,
    isString,
    ofObject
} from '@/lib/common.ts';
import { type JalaaliDate, toJalaali } from 'jalaali-js';
import moment from 'moment-timezone';
import 'moment-timezone/moment-timezone-utils';
import { getHourIn12HourFormat, getUtcOffsetByUtcOffsetMinutes } from './clock.ts';

export interface DateTimeFormatResult {
    value: number | string;
    minNumeralLength?: number;
    maxLength?: number;
    hasNumeral?: boolean;
}

export type DateTimeTokenFormatFunction = (d: MomentDateTime, options: DateTimeFormatOptions) => DateTimeFormatResult;

export function getFixedTimezoneName(utcOffset: number): string {
    return `Fixed/Timezone${utcOffset}`;
}

(function initFixedTimezone(): void {
    for (let utcOffset = WESTERNMOST_TIMEZONE_UTC_OFFSET; utcOffset <= EASTERNMOST_TIMEZONE_UTC_OFFSET; utcOffset += 15) {
        const timezoneName = getFixedTimezoneName(utcOffset);

        if (moment.tz.zone(timezoneName)) {
            continue;
        }

        moment.tz.add(moment.tz.pack({
            name: timezoneName,
            abbrs: [`FIX${utcOffset}`],
            offsets: [-utcOffset],
            untils: [0]
        }));
    }
})();

export class MomentDateTime implements DateTime {
    private static readonly tokenFormatFuncs: Record<string, DateTimeTokenFormatFunction> = {
        'YY': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getLocalizedCalendarYear(options), hasNumeral: true, minNumeralLength: 2, maxLength: 2 }),
        'YYYY': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getLocalizedCalendarYear(options), hasNumeral: true, minNumeralLength: 4 }),
        'M': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getLocalizedCalendarMonth(options), hasNumeral: true }),
        'MM': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getLocalizedCalendarMonth(options), hasNumeral: true, minNumeralLength: 2 }),
        'MMM': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getLocalizedCalendarMonthDisplayShortName(options) }),
        'MMMM': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getLocalizedCalendarMonthDisplayName(options) }),
        'D': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getLocalizedCalendarDay(options), hasNumeral: true }),
        'DD': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getLocalizedCalendarDay(options), hasNumeral: true, minNumeralLength: 2 }),
        'dd': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getWeekDayDisplayMinName(options) }),
        'ddd': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getWeekDayDisplayShortName(options) }),
        'dddd': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getWeekDayDisplayName(options) }),
        'H': (d: MomentDateTime) => ofObject<DateTimeFormatResult>({ value: d.getHour() }),
        'HH': (d: MomentDateTime) => ofObject<DateTimeFormatResult>({ value: d.getHour(), minNumeralLength: 2 }),
        'h': (d: MomentDateTime) => ofObject<DateTimeFormatResult>({ value: getHourIn12HourFormat(d.getHour()) }),
        'hh': (d: MomentDateTime) => ofObject<DateTimeFormatResult>({ value: getHourIn12HourFormat(d.getHour()), minNumeralLength: 2 }),
        'm': (d: MomentDateTime) => ofObject<DateTimeFormatResult>({ value: d.getMinute() }),
        'mm': (d: MomentDateTime) => ofObject<DateTimeFormatResult>({ value: d.getMinute(), minNumeralLength: 2 }),
        's': (d: MomentDateTime) => ofObject<DateTimeFormatResult>({ value: d.getSecond() }),
        'ss': (d: MomentDateTime) => ofObject<DateTimeFormatResult>({ value: d.getSecond(), minNumeralLength: 2 }),
        'A': (d: MomentDateTime, options: DateTimeFormatOptions) => ofObject<DateTimeFormatResult>({ value: d.getDisplayAMPM(options) }),
        'Z': (d: MomentDateTime) => ofObject<DateTimeFormatResult>({ value: getUtcOffsetByUtcOffsetMinutes(d.getTimezoneUtcOffsetMinutes()), hasNumeral: true }),
    };

    private readonly instance: moment.Moment;
    private chineseDateInfo?: ChineseYearMonthDayInfo | undefined = undefined;
    private persianDateInfo?: JalaaliDate | undefined = undefined;

    private constructor(instance: moment.Moment) {
        this.instance = instance;
    }

    public getUnixTime(): number {
        return this.instance.unix();
    }

    public getLocalizedCalendarYear(options: DateTimeFormatOptions): string {
        if (options && options.calendarType === CalendarType.Buddhist) {
            return (this.instance.year() + 543).toString();
        } else if (options && options.calendarType === CalendarType.Chinese) {
            return this.getChineseDateInfo(options.chineseCalendarLocaleData)?.displayYear ?? '';
        } else if (options && options.calendarType === CalendarType.Persian) {
            return this.getPersianDateInfo().jy.toString();
        }

        return this.instance.year().toString();
    }

    public getGregorianCalendarYear(): number {
        return this.instance.year();
    }

    public getGregorianCalendarQuarter(): number {
        return this.instance.quarter();
    }

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    public getLocalizedCalendarQuarter(options: DateTimeFormatOptions): number {
        return this.instance.quarter();
    }

    public getGregorianCalendarMonth(): number {
        return this.instance.month() + 1;
    }

    public getGregorianCalendarMonthDisplayName(options: DateTimeFormatOptions): string {
        if (!options || !options.localeData) {
            return '';
        }

        const names = options.localeData.months();
        return names[this.getGregorianCalendarMonth() - 1] || '';
    }

    public getGregorianCalendarMonthDisplayShortName(options: DateTimeFormatOptions): string {
        if (!options || !options.localeData) {
            return '';
        }

        const names = options.localeData.monthsShort();
        return names[this.getGregorianCalendarMonth() - 1] || '';
    }

    public getLocalizedCalendarMonth(options: DateTimeFormatOptions): string {
        if (options && options.calendarType === CalendarType.Chinese) {
            return this.getChineseDateInfo(options.chineseCalendarLocaleData)?.displayMonth ?? '';
        } else if (options && options.calendarType === CalendarType.Persian) {
            return this.getPersianDateInfo().jm.toString();
        }

        return (this.instance.month() + 1).toString();
    }

    public getLocalizedCalendarMonthDisplayName(options: DateTimeFormatOptions): string {
        if (!options || !options.localeData) {
            return '';
        }

        if (options && options.calendarType === CalendarType.Chinese) {
            return this.getChineseDateInfo(options.chineseCalendarLocaleData)?.displayMonth ?? '';
        } else if (options && options.calendarType === CalendarType.Persian) {
            return options.persianCalendarLocaleData.monthNames[this.getPersianDateInfo().jm - 1] ?? '';
        }

        const names = options.localeData.months();
        return names[this.instance.month()] || '';
    }

    public getLocalizedCalendarMonthDisplayShortName(options: DateTimeFormatOptions): string {
        if (!options || !options.localeData) {
            return '';
        }

        if (options && options.calendarType === CalendarType.Chinese) {
            return this.getChineseDateInfo(options.chineseCalendarLocaleData)?.displayMonth ?? '';
        } else if (options && options.calendarType === CalendarType.Persian) {
            return options.persianCalendarLocaleData.monthShortNames[this.getPersianDateInfo().jm - 1] ?? '';
        }

        const names = options.localeData.monthsShort();
        return names[this.instance.month()] || '';
    }

    public getGregorianCalendarDay(): number {
        return this.instance.date();
    }

    public getLocalizedCalendarDay(options: DateTimeFormatOptions): string {
        if (options && options.calendarType === CalendarType.Chinese) {
            return this.getChineseDateInfo(options.chineseCalendarLocaleData)?.displayDay ?? '';
        } else if (options && options.calendarType === CalendarType.Persian) {
            return this.getPersianDateInfo().jd.toString();
        }

        return this.instance.date().toString();
    }

    public isLocalizedCalendarFirstDayOfMonth(options: DateTimeFormatOptions): boolean {
        if (options && options.calendarType === CalendarType.Chinese) {
            return this.getChineseDateInfo(options.chineseCalendarLocaleData)?.day === 1;
        } else if (options && options.calendarType === CalendarType.Persian) {
            return this.getPersianDateInfo().jd === 1;
        }

        return this.instance.date() === 1;
    }

    public getGregorianCalendarYearDashMonthDashDay(): TextualYearMonthDay {
        return (this.instance.year() + '-' + (this.instance.month() + 1).toString().padStart(2, NumeralSystem.WesternArabicNumerals.digitZero) + '-' + this.instance.date().toString().padStart(2, NumeralSystem.WesternArabicNumerals.digitZero)) as TextualYearMonthDay;
    }

    public getGregorianCalendarYearDashMonth(): TextualYearMonth {
        return (this.instance.year() + '-' + (this.instance.month() + 1).toString().padStart(2, NumeralSystem.WesternArabicNumerals.digitZero)) as TextualYearMonth;
    }

    public getMaxDayOfGregorianCalendarMonth(): number {
        return this.instance.clone().endOf('month').date();
    }

    public getWeekDay(): WeekDay {
        return WeekDay.valueOf(this.instance.day()) as WeekDay;
    }

    public getWeekDayDisplayName(options: DateTimeFormatOptions): string {
        if (!options || !options.localeData) {
            return '';
        }

        const names = options.localeData.weekdays();
        return names[this.instance.day()] || '';
    }

    public getWeekDayDisplayShortName(options: DateTimeFormatOptions): string {
        if (!options || !options.localeData) {
            return '';
        }

        const names = options.localeData.weekdaysShort();
        return names[this.instance.day()] || '';
    }

    public getWeekDayDisplayMinName(options: DateTimeFormatOptions): string {
        if (!options || !options.localeData) {
            return '';
        }

        const names = options.localeData.weekdaysMin();
        return names[this.instance.day()] || '';
    }

    public getHour(): number {
        return this.instance.hour();
    }

    public getMinute(): number {
        return this.instance.minute();
    }

    public getSecond(): number {
        return this.instance.second();
    }

    public getDisplayAMPM(options: DateTimeFormatOptions): string {
        if (!options || !options.localeData) {
            return '';
        }

        return options.localeData.meridiem(this.getHour(), this.getMinute(), false);
    }

    public getTimezoneUtcOffsetMinutes(): number {
        return this.instance.utcOffset();
    }

    public setTimezoneByUtcOffsetMinutes(ufcOffset: number): DateTime {
        return MomentDateTime.of(this.instance.clone().tz(getFixedTimezoneName(ufcOffset)));
    }

    public setTimezoneByIANATimeZoneName(timezoneName: string): DateTime {
        return MomentDateTime.of(this.instance.clone().tz(timezoneName));
    }

    public add(amount: number, unit: DateTimeUnit): DateTime {
        return MomentDateTime.of(this.instance.clone().add(amount, unit));
    }

    public subtract(amount: number, unit: DateTimeUnit): DateTime {
        return MomentDateTime.of(this.instance.clone().subtract(amount, unit));
    }

    public set(value: DateTimeSetObject): DateTime {
        return MomentDateTime.of(this.instance.clone().set({
            year: value.year,
            month: isDefined(value.month) ? value.month - 1 : undefined,
            date: value.dayOfMonth,
            hour: value.hour,
            minute: value.minute,
            second: value.second,
            millisecond: value.millisecond
        }));
    }

    public toGregorianCalendarYearMonthDay(): YearMonthDay {
        return {
            year: this.instance.year(),
            month: this.instance.month() + 1,
            day: this.instance.date()
        };
    }

    public toGregorianCalendarYear0BasedMonth(): Year0BasedMonth {
        return {
            year: this.instance.year(),
            month0base: this.instance.month()
        };
    }

    public format(format: string, options: DateTimeFormatOptions): string {
        let result = '';
        let i = 0;

        while (i < format.length) {
            let matched = false;
            for (let len = 4; len > 0; len--) {
                const token = format.substring(i, i + len);
                const formatFunc = MomentDateTime.tokenFormatFuncs[token];

                if (isFunction(formatFunc)) {
                    const formattedResult: DateTimeFormatResult = formatFunc(this, options);
                    let formattedValue: string = formattedResult.value.toString();

                    if (isDefined(formattedResult.minNumeralLength)) {
                        formattedValue = formattedValue.padStart(formattedResult.minNumeralLength, NumeralSystem.WesternArabicNumerals.digitZero);
                    }

                    if (isDefined(formattedResult.maxLength) && formattedValue.length > formattedResult.maxLength) {
                        formattedValue = formattedValue.substring(formattedValue.length - formattedResult.maxLength);
                    }

                    if (isNumber(formattedResult.value)) {
                        if (options && options.numeralSystem) {
                            formattedValue = options.numeralSystem.replaceWesternArabicDigitsToLocalizedDigits(formattedValue);
                        }
                    } else if (isString(formattedValue)) {
                        if (formattedResult.hasNumeral && options && options.numeralSystem) {
                            formattedValue = options.numeralSystem.replaceWesternArabicDigitsToLocalizedDigits(formattedValue);
                        }
                    }

                    result += formattedValue;
                    i += len;
                    matched = true;
                    break;
                }
            }

            if (!matched) {
                result += format[i];
                i++;
            }
        }

        return result;
    }

    public static of(instance: moment.Moment): DateTime {
        return new MomentDateTime(instance);
    }

    public static ofUnixTime(unixTime: number): DateTime {
        return new MomentDateTime(moment.unix(unixTime));
    }

    public static ofFullDateTime(year: number, month: number, day: number, hour: number, minute: number, second: number, millisecond: number): DateTime {
        return new MomentDateTime(moment().set({ year: year, month: month - 1, date: day, hour: hour, minute: minute, second: second, millisecond: millisecond }));
    }

    public static now(): DateTime {
        return new MomentDateTime(moment());
    }

    private getChineseDateInfo(localeData: ChineseCalendarLocaleData): ChineseYearMonthDayInfo | undefined {
        if (!this.chineseDateInfo) {
            this.chineseDateInfo = getChineseYearMonthDayInfo({
                year: this.instance.year(),
                month: this.instance.month() + 1,
                day: this.instance.date()
            }, localeData);
        }

        return this.chineseDateInfo;
    }

    private getPersianDateInfo(): JalaaliDate {
        if (!this.persianDateInfo) {
            this.persianDateInfo = toJalaali(this.instance.year(), this.instance.month() + 1, this.instance.date());
        }

        return this.persianDateInfo;
    }

    static isGregorianCalendarYearFirstTime(dateTime: MomentDateTime): boolean {
        const currentUnixTime = dateTime.instance.clone().set({ millisecond: 0 }).unix();
        const expectedUnxTime = dateTime.instance.clone().set({ millisecond: 0 }).startOf('year').unix();
        return currentUnixTime === expectedUnxTime;
    }

    static isGregorianCalendarYearLastTime(dateTime: MomentDateTime): boolean {
        const currentUnixTime = dateTime.instance.clone().set({ millisecond: 999 }).unix();
        const expectedUnxTime = dateTime.instance.clone().set({ millisecond: 999 }).endOf('year').unix();
        return currentUnixTime === expectedUnxTime;
    }

    static isGregorianCalendarMonthFirstTime(dateTime: MomentDateTime): boolean {
        const currentUnixTime = dateTime.instance.clone().set({ millisecond: 0 }).unix();
        const expectedUnxTime = dateTime.instance.clone().set({ millisecond: 0 }).startOf('month').unix();
        return currentUnixTime === expectedUnxTime;
    }

    static isGregorianCalendarMonthLastTime(dateTime: MomentDateTime): boolean {
        const currentUnixTime = dateTime.instance.clone().set({ millisecond: 999 }).unix();
        const expectedUnxTime = dateTime.instance.clone().set({ millisecond: 999 }).endOf('month').unix();
        return currentUnixTime === expectedUnxTime;
    }
}
