import {
    type ChineseCalendarLocaleData,
    type PersianCalendarLocaleData
} from '@/core/calendar.ts';
import {
    type DateTimeLocaleData
} from '@/core/datetime.ts';
import {
    ALL_LANGUAGES as CHINESE_CALENDAR_ALL_LANGUAGES,
    DEFAULT_CONTENT as CHINESE_CALENDAR_DEFAULT_CONTENT
} from '@/locales/calendar/chinese/index.ts';
import {
    ALL_LANGUAGES as PERSIAN_CALENDAR_ALL_LANGUAGES,
    DEFAULT_CONTENT as PERSIAN_CALENDAR_DEFAULT_CONTENT
} from '@/locales/calendar/persian/index.ts';
import moment from 'moment-timezone';
import type { useHelpersState } from './state.ts';

interface Context {
    locale: ReturnType<typeof useHelpersState>['locale'];
}

// Owns calendar locales for the helpers store.
export function useHelpersCalendarLocales(context: Context) {
    const { locale } = context;

    function getDateTimeLocaleData(): DateTimeLocaleData {
        return moment.localeData();
    }

    function getChineseCalendarLocaleData(): ChineseCalendarLocaleData {
        const localeData = CHINESE_CALENDAR_ALL_LANGUAGES[locale.value] ?? CHINESE_CALENDAR_DEFAULT_CONTENT;
        const chineseCalendarLocaleData: ChineseCalendarLocaleData = {
            numerals: localeData['numerals'],
            monthNames: localeData['monthNames'],
            dayNames: localeData['dayNames'],
            leapMonthPrefix: localeData['leapMonthPrefix'],
            solarTermNames: localeData['solarTermNames']
        };

        return chineseCalendarLocaleData;
    }

    function getPersianCalendarLocaleData(): PersianCalendarLocaleData {
        const localeData = PERSIAN_CALENDAR_ALL_LANGUAGES[locale.value] ?? PERSIAN_CALENDAR_DEFAULT_CONTENT;
        const persianCalendarLocaleData: PersianCalendarLocaleData = {
            monthNames: localeData['monthNames'],
            monthShortNames: localeData['monthShortNames']
        };

        return persianCalendarLocaleData;
    }

    return { getDateTimeLocaleData, getChineseCalendarLocaleData, getPersianCalendarLocaleData };
}
