
export const LANGUAGE_DEFAULT_DATE_TIME_FORMAT_VALUE: number = 0;

export enum DateFormatOrder {
    YMD = 1,
    MDY = 2,
    DMY = 3
}

export interface DateFormat {
    readonly type: number;
    readonly typeName: string;
    readonly order: DateFormatOrder;
}

export type DateFormatTypeName = 'YearMonthDay' | 'MonthDayYear' | 'DayMonthYear';

export class LongDateFormat implements DateFormat {
    private static readonly allInstances: LongDateFormat[] = [];
    private static readonly allInstancesByType: Record<number, LongDateFormat> = {};
    private static readonly allInstancesByTypeName: Record<string, LongDateFormat> = {};

    public static readonly YearMonthDay = new LongDateFormat(1, 'YearMonthDay', DateFormatOrder.YMD);
    public static readonly MonthDayYear = new LongDateFormat(2, 'MonthDayYear', DateFormatOrder.MDY);
    public static readonly DayMonthYear = new LongDateFormat(3, 'DayMonthYear', DateFormatOrder.DMY);

    public static readonly Default = LongDateFormat.YearMonthDay;

    public readonly type: number;
    public readonly typeName: string;
    public readonly order: DateFormatOrder;

    private constructor(type: number, typeName: DateFormatTypeName, order: DateFormatOrder) {
        this.type = type;
        this.typeName = typeName;
        this.order = order;

        LongDateFormat.allInstances.push(this);
        LongDateFormat.allInstancesByType[type] = this;
        LongDateFormat.allInstancesByTypeName[typeName] = this;
    }

    public static values(): LongDateFormat[] {
        return LongDateFormat.allInstances;
    }

    public static all(): Record<DateFormatTypeName, LongDateFormat> {
        return LongDateFormat.allInstancesByTypeName;
    }

    public static valueOf(type: number): LongDateFormat | undefined {
        return LongDateFormat.allInstancesByType[type];
    }
}

export class ShortDateFormat implements DateFormat {
    private static readonly allInstances: ShortDateFormat[] = [];
    private static readonly allInstancesByType: Record<number, ShortDateFormat> = {};
    private static readonly allInstancesByTypeName: Record<string, ShortDateFormat> = {};

    public static readonly YearMonthDay = new ShortDateFormat(1, 'YearMonthDay', DateFormatOrder.YMD);
    public static readonly MonthDayYear = new ShortDateFormat(2, 'MonthDayYear', DateFormatOrder.MDY);
    public static readonly DayMonthYear = new ShortDateFormat(3, 'DayMonthYear', DateFormatOrder.DMY);

    public static readonly Default = ShortDateFormat.YearMonthDay;

    public readonly type: number;
    public readonly typeName: string;
    public readonly order: DateFormatOrder;

    private constructor(type: number, typeName: DateFormatTypeName, order: DateFormatOrder) {
        this.type = type;
        this.typeName = typeName;
        this.order = order;

        ShortDateFormat.allInstances.push(this);
        ShortDateFormat.allInstancesByType[type] = this;
        ShortDateFormat.allInstancesByTypeName[typeName] = this;
    }

    public static values(): ShortDateFormat[] {
        return ShortDateFormat.allInstances;
    }

    public static all(): Record<DateFormatTypeName, ShortDateFormat> {
        return ShortDateFormat.allInstancesByTypeName;
    }

    public static valueOf(type: number): ShortDateFormat | undefined {
        return ShortDateFormat.allInstancesByType[type];
    }
}

export interface TimeFormat {
    readonly type: number;
    readonly typeName: string;
    readonly is24HourFormat: boolean;
    readonly isMeridiemIndicatorFirst: boolean | null;
}

export type LongTimeFormatTypeName = 'HourMinuteSecond' | 'MeridiemIndicatorHourMinuteSecond' | 'HourMinuteSecondMeridiemIndicator';

export class LongTimeFormat implements TimeFormat {
    private static readonly allInstances: LongTimeFormat[] = [];
    private static readonly allInstancesByType: Record<number, LongTimeFormat> = {};
    private static readonly allInstancesByTypeName: Record<string, LongTimeFormat> = {};

    public static readonly HourMinuteSecond = new LongTimeFormat(1, 'HourMinuteSecond', true, null);
    public static readonly MeridiemIndicatorHourMinuteSecond = new LongTimeFormat(2, 'MeridiemIndicatorHourMinuteSecond', false, true);
    public static readonly HourMinuteSecondMeridiemIndicator = new LongTimeFormat(3, 'HourMinuteSecondMeridiemIndicator', false, false);

    public static readonly Default = LongTimeFormat.HourMinuteSecond;

    public readonly type: number;
    public readonly typeName: string;
    public readonly is24HourFormat: boolean;
    public readonly isMeridiemIndicatorFirst: boolean | null;

    private constructor(type: number, typeName: LongTimeFormatTypeName, is24HourFormat: boolean, isMeridiemIndicatorFirst: boolean | null) {
        this.type = type;
        this.typeName = typeName;
        this.is24HourFormat = is24HourFormat;
        this.isMeridiemIndicatorFirst = isMeridiemIndicatorFirst;

        LongTimeFormat.allInstances.push(this);
        LongTimeFormat.allInstancesByType[type] = this;
        LongTimeFormat.allInstancesByTypeName[typeName] = this;
    }

    public static values(): LongTimeFormat[] {
        return LongTimeFormat.allInstances;
    }

    public static all(): Record<LongTimeFormatTypeName, LongTimeFormat> {
        return LongTimeFormat.allInstancesByTypeName;
    }

    public static valueOf(type: number): LongTimeFormat | undefined {
        return LongTimeFormat.allInstancesByType[type];
    }
}

export type ShortTimeFormatTypeName = 'HourMinute' | 'MeridiemIndicatorHourMinute' | 'HourMinuteMeridiemIndicator';

export class ShortTimeFormat implements TimeFormat {
    private static readonly allInstances: ShortTimeFormat[] = [];
    private static readonly allInstancesByType: Record<number, ShortTimeFormat> = {};
    private static readonly allInstancesByTypeName: Record<string, ShortTimeFormat> = {};

    public static readonly HourMinute = new ShortTimeFormat(1, 'HourMinute', true, null);
    public static readonly MeridiemIndicatorHourMinute = new ShortTimeFormat(2, 'MeridiemIndicatorHourMinute', false, true);
    public static readonly HourMinuteMeridiemIndicator = new ShortTimeFormat(3, 'HourMinuteMeridiemIndicator', false, false);

    public static readonly Default = ShortTimeFormat.HourMinute;

    public readonly type: number;
    public readonly typeName: string;
    public readonly is24HourFormat: boolean;
    public readonly isMeridiemIndicatorFirst: boolean | null;

    private constructor(type: number, typeName: ShortTimeFormatTypeName, is24HourFormat: boolean, isMeridiemIndicatorFirst: boolean | null) {
        this.type = type;
        this.typeName = typeName;
        this.is24HourFormat = is24HourFormat;
        this.isMeridiemIndicatorFirst = isMeridiemIndicatorFirst;

        ShortTimeFormat.allInstances.push(this);
        ShortTimeFormat.allInstancesByType[type] = this;
        ShortTimeFormat.allInstancesByTypeName[typeName] = this;
    }

    public static values(): ShortTimeFormat[] {
        return ShortTimeFormat.allInstances;
    }

    public static all(): Record<ShortTimeFormatTypeName, ShortTimeFormat> {
        return ShortTimeFormat.allInstancesByTypeName;
    }

    public static valueOf(type: number): ShortTimeFormat | undefined {
        return ShortTimeFormat.allInstancesByType[type];
    }
}

export class KnownDateTimeFormat {
    private static readonly allInstances: KnownDateTimeFormat[] = [];
    private static readonly allYMDInstances: KnownDateTimeFormat[] = [];
    private static readonly allMDYInstances: KnownDateTimeFormat[] = [];
    private static readonly allDMYInstances: KnownDateTimeFormat[] = [];

    public static readonly DefaultDateTime = new KnownDateTimeFormat('YYYY-MM-DD HH:mm:ss', DateFormatOrder.YMD, /^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1]) ([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/);
    public static readonly DefaultDateTimeWithTimezone = new KnownDateTimeFormat('YYYY-MM-DD HH:mm:ssZ', DateFormatOrder.YMD, /^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1]) ([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](Z|[+-](0[0-9]|1[0-4]):[0-5][0-9])$/);
    public static readonly DefaultDateTimeWithoutSecond = new KnownDateTimeFormat('YYYY-MM-DD HH:mm', DateFormatOrder.YMD, /^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1]) ([0-1][0-9]|2[0-3]):[0-5][0-9]$/);
    public static readonly DefaultDate = new KnownDateTimeFormat('YYYY-MM-DD', DateFormatOrder.YMD, /^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])$/);

    public static readonly RFC3339 = new KnownDateTimeFormat('YYYY-MM-DDTHH:mm:ssZ', DateFormatOrder.YMD, /^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])T([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](Z|[+-](0[0-9]|1[0-4]):[0-5][0-9])$/);

    public static readonly YYYYMMDDSlashWithTime = new KnownDateTimeFormat('YYYY/MM/DD HH:mm:ss', DateFormatOrder.YMD, /^\d{4}\/(0[1-9]|1[0-2])\/(0[1-9]|[1-2][0-9]|3[0-1]) ([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/);
    public static readonly MMDDYYYYSlashWithTime = new KnownDateTimeFormat('MM/DD/YYYY HH:mm:ss', DateFormatOrder.MDY, /^(0[1-9]|1[0-2])\/(0[1-9]|[1-2][0-9]|3[0-1])\/\d{4} ([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/);
    public static readonly DDMMYYYYSlashWithTime = new KnownDateTimeFormat('DD/MM/YYYY HH:mm:ss', DateFormatOrder.DMY, /^(0[1-9]|[1-2][0-9]|3[0-1])\/(0[1-9]|1[0-2])\/\d{4} ([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/);

    public static readonly YYYYMMDDDotWithTime = new KnownDateTimeFormat('YYYY.MM.DD HH:mm:ss', DateFormatOrder.YMD, /^\d{4}\.(0[1-9]|1[0-2])\.(0[1-9]|[1-2][0-9]|3[0-1]) ([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/);
    public static readonly MMDDYYYYDotWithTime = new KnownDateTimeFormat('MM.DD.YYYY HH:mm:ss', DateFormatOrder.MDY, /^(0[1-9]|1[0-2])\.(0[1-9]|[1-2][0-9]|3[0-1])\.\d{4} ([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/);
    public static readonly DDMMYYYYDotWithTime = new KnownDateTimeFormat('DD.MM.YYYY HH:mm:ss', DateFormatOrder.DMY, /^(0[1-9]|[1-2][0-9]|3[0-1])\.(0[1-9]|1[0-2])\.\d{4} ([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/);

    public static readonly MMDDYYYYDash = new KnownDateTimeFormat('MM-DD-YYYY', DateFormatOrder.MDY, /^(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])-\d{4}$/);
    public static readonly DDMMYYYYDash = new KnownDateTimeFormat('DD-MM-YYYY', DateFormatOrder.DMY, /^(0[1-9]|[1-2][0-9]|3[0-1])-(0[1-9]|1[0-2])-\d{4}$/);

    public static readonly YYYYMMDDSlash = new KnownDateTimeFormat('YYYY/MM/DD', DateFormatOrder.YMD, /^\d{4}\/(0[1-9]|1[0-2])\/(0[1-9]|[1-2][0-9]|3[0-1])$/);
    public static readonly MMDDYYYYSlash = new KnownDateTimeFormat('MM/DD/YYYY', DateFormatOrder.MDY, /^(0[1-9]|1[0-2])\/(0[1-9]|[1-2][0-9]|3[0-1])\/\d{4}$/);
    public static readonly DDMMYYYYSlash = new KnownDateTimeFormat('DD/MM/YYYY', DateFormatOrder.DMY, /^(0[1-9]|[1-2][0-9]|3[0-1])\/(0[1-9]|1[0-2])\/\d{4}$/);

    public static readonly YYYYMDSlash = new KnownDateTimeFormat('YYYY/M/D', DateFormatOrder.YMD, /^\d{4}\/([1-9]|1[0-2])\/([1-9]|[1-2][0-9]|3[0-1])$/);
    public static readonly MDYYYYSlash = new KnownDateTimeFormat('M/D/YYYY', DateFormatOrder.MDY, /^([1-9]|1[0-2])\/([1-9]|[1-2][0-9]|3[0-1])\/\d{4}$/);
    public static readonly DMYYYYSlash = new KnownDateTimeFormat('D/M/YYYY', DateFormatOrder.DMY, /^([1-9]|[1-2][0-9]|3[0-1])\/([1-9]|1[0-2])\/\d{4}$/);

    public static readonly YYYYMMDDDot = new KnownDateTimeFormat('YYYY.MM.DD', DateFormatOrder.YMD, /^\d{4}\.(0[1-9]|1[0-2])\.(0[1-9]|[1-2][0-9]|3[0-1])$/);
    public static readonly MMDDYYYYDot = new KnownDateTimeFormat('MM.DD.YYYY', DateFormatOrder.MDY, /^(0[1-9]|1[0-2])\.(0[1-9]|[1-2][0-9]|3[0-1])\.\d{4}$/);
    public static readonly DDMMYYYYDot = new KnownDateTimeFormat('DD.MM.YYYY', DateFormatOrder.DMY, /^(0[1-9]|[1-2][0-9]|3[0-1])\.(0[1-9]|1[0-2])\.\d{4}$/);

    public static readonly YYYYMDDot = new KnownDateTimeFormat('YYYY.M.D', DateFormatOrder.YMD, /^\d{4}\.([1-9]|1[0-2])\.([1-9]|[1-2][0-9]|3[0-1])$/);
    public static readonly MDYYYYDot = new KnownDateTimeFormat('M.D.YYYY', DateFormatOrder.MDY, /^([1-9]|1[0-2])\.([1-9]|[1-2][0-9]|3[0-1])\.\d{4}$/);
    public static readonly DMYYYYDot = new KnownDateTimeFormat('D.M.YYYY', DateFormatOrder.DMY, /^([1-9]|[1-2][0-9]|3[0-1])\.([1-9]|1[0-2])\.\d{4}$/);

    public static readonly YYYYMMDD = new KnownDateTimeFormat('YYYYMMDD', DateFormatOrder.YMD, /^\d{4}(0[1-9]|1[0-2])(0[1-9]|[1-2][0-9]|3[0-1])$/);

    public static readonly MMDDYYDash = new KnownDateTimeFormat('MM-DD-YY', DateFormatOrder.MDY, /^(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])-\d{2}$/);
    public static readonly MMDDYYSlash = new KnownDateTimeFormat('MM/DD/YY', DateFormatOrder.MDY, /^(0[1-9]|1[0-2])\/(0[1-9]|[1-2][0-9]|3[0-1])\/\d{2}$/);

    public readonly format: string;
    public readonly type: DateFormatOrder;
    private readonly regex: RegExp;

    private constructor(format: string, type: DateFormatOrder, regex: RegExp) {
        this.format = format;
        this.type = type;
        this.regex = regex;

        if (type === DateFormatOrder.YMD) {
            KnownDateTimeFormat.allYMDInstances.push(this);
        } else if (type === DateFormatOrder.MDY) {
            KnownDateTimeFormat.allMDYInstances.push(this);
        } else if (type === DateFormatOrder.DMY) {
            KnownDateTimeFormat.allDMYInstances.push(this);
        }

        KnownDateTimeFormat.allInstances.push(this);
    }

    public isValid(dateTime: string): boolean {
        return this.regex.test(dateTime);
    }

    public static values(): KnownDateTimeFormat[] {
        return KnownDateTimeFormat.allInstances;
    }

    public static detect(dateTime: string, longDateTimeFormatOrder: DateFormatOrder, shortDateTimeFormatOrder: DateFormatOrder): KnownDateTimeFormat[] | undefined {
        const allFormats: KnownDateTimeFormat[] = KnownDateTimeFormat.getAllFormatsByOrder(longDateTimeFormatOrder, shortDateTimeFormatOrder);
        return KnownDateTimeFormat.detectSingle(dateTime, allFormats);
    }

    public static detectMulti(dateTimes: string[], longDateTimeFormatOrder: DateFormatOrder, shortDateTimeFormatOrder: DateFormatOrder): KnownDateTimeFormat[] | undefined {
        const detectedCounts: Record<string, number> = {};
        const allFormats: KnownDateTimeFormat[] = KnownDateTimeFormat.getAllFormatsByOrder(longDateTimeFormatOrder, shortDateTimeFormatOrder);

        for (const dateTime of dateTimes) {
            const detectedFormats = KnownDateTimeFormat.detectSingle(dateTime, allFormats);

            if (detectedFormats) {
                for (const format of detectedFormats) {
                    detectedCounts[format.format] = (detectedCounts[format.format] || 0) + 1;
                }
            } else {
                return undefined;
            }
        }

        const result: KnownDateTimeFormat[] = [];

        for (const format of KnownDateTimeFormat.allInstances) {
            if (detectedCounts[format.format] === dateTimes.length) {
                result.push(format);
            }
        }

        return result.length > 0 ? result : undefined;
    }

    private static detectSingle(dateTime: string, allFormats: KnownDateTimeFormat[]): KnownDateTimeFormat[] | undefined {
        const result: KnownDateTimeFormat[] = [];

        for (const format of allFormats) {
            if (format.isValid(dateTime)) {
                result.push(format);
            }
        }

        return result.length > 0 ? result : undefined;
    }

    private static getAllFormatsByOrder(longDateTimeFormatOrder: DateFormatOrder, shortDateTimeFormatOrder: DateFormatOrder): KnownDateTimeFormat[] {
        if (longDateTimeFormatOrder === DateFormatOrder.YMD && (shortDateTimeFormatOrder === DateFormatOrder.YMD || shortDateTimeFormatOrder === DateFormatOrder.MDY)) {
            return [
                ...KnownDateTimeFormat.allYMDInstances,
                ...KnownDateTimeFormat.allMDYInstances,
                ...KnownDateTimeFormat.allDMYInstances
            ];
        } else if (longDateTimeFormatOrder === DateFormatOrder.YMD && shortDateTimeFormatOrder === DateFormatOrder.DMY) {
            return [
                ...KnownDateTimeFormat.allYMDInstances,
                ...KnownDateTimeFormat.allDMYInstances,
                ...KnownDateTimeFormat.allMDYInstances
            ];
        } else if (longDateTimeFormatOrder === DateFormatOrder.MDY && (shortDateTimeFormatOrder === DateFormatOrder.MDY || shortDateTimeFormatOrder === DateFormatOrder.YMD)) {
            return [
                ...KnownDateTimeFormat.allMDYInstances,
                ...KnownDateTimeFormat.allYMDInstances,
                ...KnownDateTimeFormat.allDMYInstances
            ];
        } else if (longDateTimeFormatOrder === DateFormatOrder.MDY && shortDateTimeFormatOrder === DateFormatOrder.DMY) {
            return [
                ...KnownDateTimeFormat.allMDYInstances,
                ...KnownDateTimeFormat.allDMYInstances,
                ...KnownDateTimeFormat.allYMDInstances
            ];
        } else if (longDateTimeFormatOrder === DateFormatOrder.DMY && (shortDateTimeFormatOrder === DateFormatOrder.DMY || shortDateTimeFormatOrder === DateFormatOrder.YMD)) {
            return [
                ...KnownDateTimeFormat.allDMYInstances,
                ...KnownDateTimeFormat.allYMDInstances,
                ...KnownDateTimeFormat.allMDYInstances
            ];
        } else if (longDateTimeFormatOrder === DateFormatOrder.DMY && shortDateTimeFormatOrder === DateFormatOrder.MDY) {
            return [
                ...KnownDateTimeFormat.allDMYInstances,
                ...KnownDateTimeFormat.allMDYInstances,
                ...KnownDateTimeFormat.allYMDInstances
            ];
        } else {
            return KnownDateTimeFormat.allInstances;
        }
    }
}
