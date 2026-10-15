import {
    type DateTime,
    type WeekDayValue,
    WeekDay
} from '@/core/datetime.ts';
import {
    isNumber
} from '@/lib/common.ts';
import {
    formatDateTime,
    getCurrentDateTime
} from '@/lib/datetime.ts';
import type { useHelpersDateOptions } from './date-options.ts';
import type { useHelpersDatePatterns } from './date-patterns.ts';
import type { useHelpersFormatOptions } from './format-options.ts';
import type { useHelpersMonthNames } from './month-names.ts';
import type { useHelpersState } from './state.ts';
import type { useHelpersTranslation } from './translation.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
    getLocalizedLongMonthDayFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedLongMonthDayFormat'];
    getDateTimeFormatOptions: ReturnType<typeof useHelpersFormatOptions>['getDateTimeFormatOptions'];
    joinMultiText: ReturnType<typeof useHelpersTranslation>['joinMultiText'];
    getMonthdayOrdinal: ReturnType<typeof useHelpersMonthNames>['getMonthdayOrdinal'];
    getAllWeekDays: ReturnType<typeof useHelpersDateOptions>['getAllWeekDays'];
}

// Owns calendar labels for the helpers store.
export function useHelpersCalendarLabels(context: Context) {
    const { t, getLocalizedLongMonthDayFormat, getDateTimeFormatOptions, joinMultiText, getMonthdayOrdinal, getAllWeekDays } = context;

    function getMonthShortName(monthName: string): string {
        return t(`datetime.${monthName}.short`);
    }

    function getMonthLongName(monthName: string): string {
        return t(`datetime.${monthName}.long`);
    }

    function getWeekdayShortName(weekDay: WeekDay): string {
        return t(`datetime.${weekDay.name}.short`);
    }

    function getWeekdayLongName(weekDay: WeekDay): string {
        return t(`datetime.${weekDay.name}.long`);
    }

    function getMultiMonthAndDayLongNames(monthAndDays: number[]): string {
        if (!monthAndDays) {
            return '';
        }

        const now: DateTime = getCurrentDateTime();

        const finalMonthAndDayNames: string[] = monthAndDays.map(monthAndDay => {
            const month = Math.trunc(monthAndDay / 100);
            const day = monthAndDay % 100;
            const dateTime = now.set({
                month: month,
                dayOfMonth: day,
                hour: 0,
                minute: 0,
                second: 0,
                millisecond: 0
            });

            return formatDateTime(dateTime, getLocalizedLongMonthDayFormat(), getDateTimeFormatOptions());
        });

        return joinMultiText(finalMonthAndDayNames);
    }

    function getMultiMonthdayShortNames(monthDays: number[]): string {
        if (!monthDays) {
            return '';
        }

        monthDays.sort(function (d1, d2) {
            if (d1 >= 0 && d2 >= 0) {
                return d1 - d2;
            } else if (d1 < 0 && d2 < 0) {
                return d1 - d2;
            } else if (d1 >= 0 && d2 < 0) {
                return -1; // make positive month day come first
            } else { // if (d1 < 0 && d2 >= 0)
                return 1; // make positive month day come first
            }
        });

        if (monthDays.length === 1) {
            const monthDay = monthDays[0] as number;

            if (monthDay >= 0) {
                return t('format.misc.monthDay', {
                    ordinal: getMonthdayOrdinal(monthDay)
                });
            } else if (monthDay === -1) {
                return t('last day');
            } else {
                return t('format.misc.lastMonthDayInLowercase', {
                    ordinal: getMonthdayOrdinal(-monthDay)
                });
            }
        } else {
            return t('format.misc.monthDays', {
                multiMonthDays: joinMultiText(monthDays.map(monthDay => {
                    if (monthDay >= 0) {
                        return t('format.misc.eachMonthDayInMonthDays', {
                            ordinal: getMonthdayOrdinal(monthDay)
                        });
                    } else {
                        return t('format.misc.eachLastMonthDayInMonthDays', {
                            ordinal: getMonthdayOrdinal(-monthDay)
                        });
                    }
                }))
            });
        }
    }

    function getMultiWeekdayLongNames(weekdayTypes: number[], firstDayOfWeek?: WeekDayValue): string {
        const weekdayTypesMap: Record<number, boolean> = {};

        if (!isNumber(firstDayOfWeek)) {
            firstDayOfWeek = WeekDay.DefaultFirstDay.type;
        }

        for (const weekdayType of weekdayTypes) {
            weekdayTypesMap[weekdayType] = true;
        }

        const allWeekDays = getAllWeekDays(firstDayOfWeek);
        const finalWeekdayNames = [];

        for (const weekDay of allWeekDays) {
            if (weekdayTypesMap[weekDay.type]) {
                finalWeekdayNames.push(weekDay.displayName);
            }
        }

        return joinMultiText(finalWeekdayNames);
    }

    return {
        getMonthShortName, getMonthLongName, getWeekdayShortName, getWeekdayLongName, getMultiMonthAndDayLongNames,
        getMultiMonthdayShortNames, getMultiWeekdayLongNames,
    };
}
