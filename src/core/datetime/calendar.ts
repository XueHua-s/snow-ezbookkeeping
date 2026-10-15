import type { TypeAndName } from '@/core/base.ts';
import type { UnixTimeRange, Year0BasedMonth, YearMonthDay, YearQuarter } from './types.ts';

export class YearUnixTime implements UnixTimeRange {
    public readonly year: number;
    public readonly minUnixTime: number;
    public readonly maxUnixTime: number;

    private constructor(year: number, minUnixTime: number, maxUnixTime: number) {
        this.year = year;
        this.minUnixTime = minUnixTime;
        this.maxUnixTime = maxUnixTime;
    }

    public static of(year: number, minUnixTime: number, maxUnixTime: number): YearUnixTime {
        return new YearUnixTime(year, minUnixTime, maxUnixTime);
    }
}

export class YearQuarterUnixTime implements YearQuarter, UnixTimeRange {
    public readonly year: number;
    public readonly quarter: number;
    public readonly minUnixTime: number;
    public readonly maxUnixTime: number;

    private constructor(year: number, quarter: number, minUnixTime: number, maxUnixTime: number) {
        this.year = year;
        this.quarter = quarter;
        this.minUnixTime = minUnixTime;
        this.maxUnixTime = maxUnixTime;
    }

    public static of(yearQuarter: YearQuarter, minUnixTime: number, maxUnixTime: number): YearQuarterUnixTime {
        return new YearQuarterUnixTime(yearQuarter.year, yearQuarter.quarter, minUnixTime, maxUnixTime);
    }
}

export class YearMonthUnixTime implements Year0BasedMonth, UnixTimeRange {
    public readonly year: number;
    public readonly month0base: number;
    public readonly minUnixTime: number;
    public readonly maxUnixTime: number;

    private constructor(year: number, month0base: number, minUnixTime: number, maxUnixTime: number) {
        this.year = year;
        this.month0base = month0base;
        this.minUnixTime = minUnixTime;
        this.maxUnixTime = maxUnixTime;
    }

    public static of(yearMonth: Year0BasedMonth, minUnixTime: number, maxUnixTime: number): YearMonthUnixTime {
        return new YearMonthUnixTime(yearMonth.year, yearMonth.month0base, minUnixTime, maxUnixTime);
    }
}

export class YearMonthDayUnixTime implements YearMonthDay, UnixTimeRange {
    public readonly year: number;
    public readonly month: number;
    public readonly day: number;
    public readonly minUnixTime: number;
    public readonly maxUnixTime: number;

    private constructor(year: number, month: number, day: number, minUnixTime: number, maxUnixTime: number) {
        this.year = year;
        this.month = month;
        this.day = day
        this.minUnixTime = minUnixTime;
        this.maxUnixTime = maxUnixTime;
    }

    public static of(yearMonthDay: YearMonthDay, minUnixTime: number, maxUnixTime: number): YearMonthDayUnixTime {
        return new YearMonthDayUnixTime(yearMonthDay.year, yearMonthDay.month, yearMonthDay.day, minUnixTime, maxUnixTime);
    }
}

export type MonthValue = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12;

export class Month {
    private static readonly allInstances: Month[] = [];

    public static readonly January = new Month(1, 'January');
    public static readonly February = new Month(2, 'February');
    public static readonly March = new Month(3, 'March');
    public static readonly April = new Month(4, 'April');
    public static readonly May = new Month(5, 'May');
    public static readonly June = new Month(6, 'June');
    public static readonly July = new Month(7, 'July');
    public static readonly August = new Month(8, 'August');
    public static readonly September = new Month(9, 'September');
    public static readonly October = new Month(10, 'October');
    public static readonly November = new Month(11, 'November');
    public static readonly December = new Month(12, 'December');

    public readonly month: MonthValue; // 1-based (1 = January, 12 = December)
    public readonly name: string;

    private constructor(month: MonthValue, name: string) {
        this.month = month;
        this.name = name;

        Month.allInstances.push(this);
    }

    public static values(): Month[] {
        return Month.allInstances;
    }

    public static valueOf(month: number): Month | undefined {
        return Month.allInstances[month - 1];
    }
}

export type WeekDayValue = 0 | 1 | 2 | 3 | 4 | 5 | 6;

export class WeekDay implements TypeAndName {
    private static readonly allInstances: WeekDay[] = [];
    private static readonly allInstancesByName: Record<string, WeekDay> = {};

    public static readonly Sunday = new WeekDay(0, 'Sunday');
    public static readonly Monday = new WeekDay(1, 'Monday');
    public static readonly Tuesday = new WeekDay(2, 'Tuesday');
    public static readonly Wednesday = new WeekDay(3, 'Wednesday');
    public static readonly Thursday = new WeekDay(4, 'Thursday');
    public static readonly Friday = new WeekDay(5, 'Friday');
    public static readonly Saturday = new WeekDay(6, 'Saturday');

    public static readonly DefaultFirstDay = WeekDay.Sunday;

    public readonly type: WeekDayValue;
    public readonly name: string;

    private constructor(type: WeekDayValue, name: string) {
        this.type = type;
        this.name = name;

        WeekDay.allInstances.push(this);
        WeekDay.allInstancesByName[name] = this;
    }

    public getDisplayOrder(firstDayOfWeek: WeekDayValue): number {
        return (this.type - firstDayOfWeek + 7) % 7;
    }

    public static values(): WeekDay[] {
        return WeekDay.allInstances;
    }

    public static valueOf(dayOfWeek: number): WeekDay | undefined {
        return WeekDay.allInstances[dayOfWeek];
    }

    public static parse(typeName: string): WeekDay | undefined {
        return WeekDay.allInstancesByName[typeName];
    }
}

export class MeridiemIndicator {
    private static readonly allInstances: MeridiemIndicator[] = [];

    public static readonly AM = new MeridiemIndicator(0, 'AM');
    public static readonly PM = new MeridiemIndicator(1, 'PM');

    public readonly type: number;
    public readonly name: string;

    private constructor(type: number, name: string) {
        this.type = type;
        this.name = name;

        MeridiemIndicator.allInstances.push(this);
    }

    public static values(): MeridiemIndicator[] {
        return MeridiemIndicator.allInstances;
    }
}
