import {
    Month,
    WeekDay
} from '@/core/datetime.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
}

// Owns month names for the helpers store.
export function useHelpersMonthNames(context: Context) {
    const { t } = context;

    function getAllMonthNames(type: string): string[] {
        const ret = [];
        const allMonths = Month.values();

        for (const month of allMonths) {
            ret.push(t(`datetime.${month.name}.${type}`));
        }

        return ret;
    }

    function getAllWeekdayNames(type: string): string[] {
        const ret = [];
        const allWeekDays = WeekDay.values();

        for (const weekDay of allWeekDays) {
            ret.push(t(`datetime.${weekDay.name}.${type}`));
        }

        return ret;
    }

    function getMonthdayOrdinal(monthDay: number): string {
        return t(`datetime.monthDayOrdinal.${monthDay}`);
    }

    function getMonthdayShortName(monthDay: number): string {
        return t('format.misc.monthDay', {
            ordinal: getMonthdayOrdinal(monthDay)
        });
    }

    function getMonthLastDayShortName(ordinal: number): string {
        return t('format.misc.lastMonthDay', {
            ordinal: getMonthdayOrdinal(ordinal)
        });
    }

    function getQuarterName(quarter: number): string {
        if (1 <= quarter && quarter <= 4) {
            return t('datetime.quarter.q' + quarter);
        } else {
            return '';
        }
    }

    return { getAllMonthNames, getAllWeekdayNames, getMonthdayOrdinal, getMonthdayShortName, getMonthLastDayShortName, getQuarterName };
}
