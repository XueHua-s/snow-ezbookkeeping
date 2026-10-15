import {
    type DateFormat,
    type TimeFormat,
    LongDateFormat,
    LongTimeFormat,
    ShortDateFormat,
    ShortTimeFormat
} from '@/core/datetime.ts';
import {
    getDateTimeFormatType
} from '@/lib/datetime.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
    userStore: ReturnType<typeof useHelpersState>['userStore'];
}

// Owns date patterns for the helpers store.
export function useHelpersDatePatterns(context: Context) {
    const { t, userStore } = context;

    function getLocalizedDateTimeType<T extends DateFormat | TimeFormat>(allFormatMap: Record<string, T>, allFormatArray: T[], formatTypeValue: number, languageDefaultTypeNameKey: string, systemDefaultFormatType: T): T {
        return getDateTimeFormatType(allFormatMap, allFormatArray, formatTypeValue, t(`default.${languageDefaultTypeNameKey}`), systemDefaultFormatType);
    }

    function getLocalizedDateTimeFormat<T extends DateFormat | TimeFormat>(type: string, allFormatMap: Record<string, T>, allFormatArray: T[], formatTypeValue: number, languageDefaultTypeNameKey: string, systemDefaultFormatType: T): string {
        const formatType = getLocalizedDateTimeType(allFormatMap, allFormatArray, formatTypeValue, languageDefaultTypeNameKey, systemDefaultFormatType);
        return t(`format.${type}.${formatType.typeName}`);
    }

    function getLocalizedLongDateFormat(): string {
        return getLocalizedDateTimeFormat<LongDateFormat>('longDate', LongDateFormat.all(), LongDateFormat.values(), userStore.currentUserLongDateFormat, 'longDateFormat', LongDateFormat.Default);
    }

    function getLocalizedShortDateFormat(): string {
        return getLocalizedDateTimeFormat<ShortDateFormat>('shortDate', ShortDateFormat.all(), ShortDateFormat.values(), userStore.currentUserShortDateFormat, 'shortDateFormat', ShortDateFormat.Default);
    }

    function getLocalizedLongYearFormat(): string {
        return getLocalizedDateTimeFormat<LongDateFormat>('longYear', LongDateFormat.all(), LongDateFormat.values(), userStore.currentUserLongDateFormat, 'longDateFormat', LongDateFormat.Default);
    }

    function getLocalizedShortYearFormat(): string {
        return getLocalizedDateTimeFormat<ShortDateFormat>('shortYear', ShortDateFormat.all(), ShortDateFormat.values(), userStore.currentUserShortDateFormat, 'shortDateFormat', ShortDateFormat.Default);
    }

    function getLocalizedLongYearMonthFormat(): string {
        return getLocalizedDateTimeFormat<LongDateFormat>('longYearMonth', LongDateFormat.all(), LongDateFormat.values(), userStore.currentUserLongDateFormat, 'longDateFormat', LongDateFormat.Default);
    }

    function getLocalizedShortYearMonthFormat(): string {
        return getLocalizedDateTimeFormat<ShortDateFormat>('shortYearMonth', ShortDateFormat.all(), ShortDateFormat.values(), userStore.currentUserShortDateFormat, 'shortDateFormat', ShortDateFormat.Default);
    }

    function getLocalizedLongMonthDayFormat(): string {
        return getLocalizedDateTimeFormat<LongDateFormat>('longMonthDay', LongDateFormat.all(), LongDateFormat.values(), userStore.currentUserLongDateFormat, 'longDateFormat', LongDateFormat.Default);
    }

    function getLocalizedShortMonthDayFormat(): string {
        return getLocalizedDateTimeFormat<ShortDateFormat>('shortMonthDay', ShortDateFormat.all(), ShortDateFormat.values(), userStore.currentUserShortDateFormat, 'shortDateFormat', ShortDateFormat.Default);
    }

    function getLocalizedShortDayFormat(): string {
        return getLocalizedDateTimeFormat<ShortDateFormat>('shortDay', ShortDateFormat.all(), ShortDateFormat.values(), userStore.currentUserShortDateFormat, 'shortDateFormat', ShortDateFormat.Default);
    }

    function getLocalizedLongTimeFormat(): string {
        return getLocalizedDateTimeFormat<LongTimeFormat>('longTime', LongTimeFormat.all(), LongTimeFormat.values(), userStore.currentUserLongTimeFormat, 'longTimeFormat', LongTimeFormat.Default);
    }

    function getLocalizedShortTimeFormat(): string {
        return getLocalizedDateTimeFormat<ShortTimeFormat>('shortTime', ShortTimeFormat.all(), ShortTimeFormat.values(), userStore.currentUserShortTimeFormat, 'shortTimeFormat', ShortTimeFormat.Default);
    }

    return {
        getLocalizedDateTimeType, getLocalizedDateTimeFormat, getLocalizedLongDateFormat, getLocalizedShortDateFormat,
        getLocalizedLongYearFormat, getLocalizedShortYearFormat, getLocalizedLongYearMonthFormat,
        getLocalizedShortYearMonthFormat, getLocalizedLongMonthDayFormat, getLocalizedShortMonthDayFormat,
        getLocalizedShortDayFormat, getLocalizedLongTimeFormat, getLocalizedShortTimeFormat,
    };
}
