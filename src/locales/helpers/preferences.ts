import { ACCOUNT_CURRENCY_NOT_SET_VALUE } from '@/consts/currency.ts';
import {
    CalendarDisplayType,
    DateDisplayType
} from '@/core/calendar.ts';
import {
    DateFormatOrder,
    LongDateFormat,
    LongTimeFormat,
    ShortDateFormat,
    ShortTimeFormat
} from '@/core/datetime.ts';
import {
    FiscalYearFormat
} from '@/core/fiscalyear.ts';
import {
    DecimalSeparator,
    DigitGroupingSymbol,
    DigitGroupingType,
    NumeralSystem
} from '@/core/numeral.ts';
import type { useHelpersDatePatterns } from './date-patterns.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
    userStore: ReturnType<typeof useHelpersState>['userStore'];
    getLocalizedDateTimeType: ReturnType<typeof useHelpersDatePatterns>['getLocalizedDateTimeType'];
    getLocalizedLongTimeFormat: ReturnType<typeof useHelpersDatePatterns>['getLocalizedLongTimeFormat'];
}

// Owns preferences for the helpers store.
export function useHelpersPreferences(context: Context) {
    const { t, userStore, getLocalizedDateTimeType, getLocalizedLongTimeFormat } = context;

    function getAllLocalizedDigits(): string[] {
        const numeralSystem = getCurrentNumeralSystemType();
        return numeralSystem.getAllDigits();
    }

    function getLocaleDefaultCalendarDisplayType(): CalendarDisplayType {
        const defaultCalendarDisplayTypeName = t('default.calendarDisplayType');
        let calendarDisplayType = CalendarDisplayType.parse(defaultCalendarDisplayTypeName);

        if (!calendarDisplayType) {
            calendarDisplayType = CalendarDisplayType.Default;
        }

        return calendarDisplayType;
    }

    function getLocaleDefaultDateDisplayType(): DateDisplayType {
        const defaultDateDisplayTypeName = t('default.dateDisplayType');
        let dateDisplayType = DateDisplayType.parse(defaultDateDisplayTypeName);

        if (!dateDisplayType) {
            dateDisplayType = DateDisplayType.Default;
        }

        return dateDisplayType;
    }

    function getLocaleDefaultNumeralSystemType(): NumeralSystem {
        const defaultNumeralSystemTypeName = t('default.numeralSystem');
        let numeralSystemType = NumeralSystem.parse(defaultNumeralSystemTypeName);

        if (!numeralSystemType) {
            numeralSystemType = NumeralSystem.Default;
        }

        return numeralSystemType;
    }

    function getLocaleDefaultDecimalSeparator(): DecimalSeparator {
        const defaultDecimalSeparatorTypeName = t('default.decimalSeparator');
        let decimalSeparatorType = DecimalSeparator.parse(defaultDecimalSeparatorTypeName);

        if (!decimalSeparatorType) {
            decimalSeparatorType = DecimalSeparator.Default;
        }

        return decimalSeparatorType;
    }

    function getLocaleDefaultDigitGroupingSymbol(): DigitGroupingSymbol {
        const defaultDigitGroupingSymbolTypeName = t('default.digitGroupingSymbol');
        let digitGroupingSymbolType = DigitGroupingSymbol.parse(defaultDigitGroupingSymbolTypeName);

        if (!digitGroupingSymbolType) {
            digitGroupingSymbolType = DigitGroupingSymbol.Default;
        }

        return digitGroupingSymbolType;
    }

    function getLocaleDefaultDigitGroupingType(): DigitGroupingType {
        const defaultDigitGroupingTypeName = t('default.digitGrouping');
        let digitGroupingType = DigitGroupingType.parse(defaultDigitGroupingTypeName);

        if (!digitGroupingType) {
            digitGroupingType = DigitGroupingType.Default;
        }

        return digitGroupingType;
    }

    function getLocaleFiscalYearFormatType(): FiscalYearFormat {
        const defaultFiscalYearFormatTypeName = t('default.fiscalYearFormat');
        let fiscalYearFormat = FiscalYearFormat.parse(defaultFiscalYearFormatTypeName);

        if (!fiscalYearFormat) {
            fiscalYearFormat = FiscalYearFormat.Default;
        }

        return fiscalYearFormat;
    }

    function getCurrentCalendarDisplayType(): CalendarDisplayType {
        return CalendarDisplayType.valueOf(userStore.currentUserCalendarDisplayType) ?? getLocaleDefaultCalendarDisplayType();
    }

    function getCurrentDateDisplayType(): DateDisplayType {
        return DateDisplayType.valueOf(userStore.currentUserDateDisplayType) ?? getLocaleDefaultDateDisplayType();
    }

    function getCurrentNumeralSystemType(): NumeralSystem {
        return NumeralSystem.valueOf(userStore.currentUserNumeralSystem) ?? getLocaleDefaultNumeralSystemType();
    }

    function getCurrentDecimalSeparator(): string {
        return DecimalSeparator.valueOf(userStore.currentUserDecimalSeparator)?.symbol ?? getLocaleDefaultDecimalSeparator().symbol;
    }

    function getCurrentDigitGroupingSymbol(): string {
        return DigitGroupingSymbol.valueOf(userStore.currentUserDigitGroupingSymbol)?.symbol ?? getLocaleDefaultDigitGroupingSymbol().symbol;
    }

    function getCurrentDigitGroupingType(): DigitGroupingType {
        return DigitGroupingType.valueOf(userStore.currentUserDigitGrouping) ?? getLocaleDefaultDigitGroupingType();
    }

    function getCurrentFiscalYearFormatType(): number {
        return FiscalYearFormat.valueOf(userStore.currentUserFiscalYearFormat)?.type ?? getLocaleFiscalYearFormatType().type;
    }

    function getCurrencyName(currencyCode: string): string {
        if (!currencyCode) {
            return '';
        }

        if (currencyCode === ACCOUNT_CURRENCY_NOT_SET_VALUE) {
            return t('Not set');
        }

        return t(`currency.name.${currencyCode}`);
    }

    function getLongDateFormatOrder(): DateFormatOrder {
        return getLocalizedDateTimeType(LongDateFormat.all(), LongDateFormat.values(), userStore.currentUserLongDateFormat, 'longDateFormat', LongDateFormat.Default).order;
    }

    function getShortDateFormatOrder(): DateFormatOrder {
        return getLocalizedDateTimeType(ShortDateFormat.all(), ShortDateFormat.values(), userStore.currentUserShortDateFormat, 'shortDateFormat', ShortDateFormat.Default).order;
    }

    function isLongDateMonthAfterYear(): boolean {
        const order: DateFormatOrder = getLongDateFormatOrder();
        return order === DateFormatOrder.YMD;
    }

    function isShortDateMonthAfterYear(): boolean {
        const order: DateFormatOrder = getShortDateFormatOrder();
        return order === DateFormatOrder.YMD;
    }

    function isLongTime24HourFormat(): boolean {
        return getLocalizedDateTimeType(LongTimeFormat.all(), LongTimeFormat.values(), userStore.currentUserLongTimeFormat, 'longTimeFormat', LongTimeFormat.Default).is24HourFormat;
    }

    function isLongTimeMeridiemIndicatorFirst(): boolean {
        return getLocalizedDateTimeType(LongTimeFormat.all(), LongTimeFormat.values(), userStore.currentUserLongTimeFormat, 'longTimeFormat', LongTimeFormat.Default).isMeridiemIndicatorFirst || false;
    }

    function isShortTime24HourFormat(): boolean {
        return getLocalizedDateTimeType(ShortTimeFormat.all(), ShortTimeFormat.values(), userStore.currentUserShortTimeFormat, 'shortTimeFormat', ShortTimeFormat.Default).is24HourFormat;
    }

    function isShortTimeMeridiemIndicatorFirst(): boolean {
        return getLocalizedDateTimeType(ShortTimeFormat.all(), ShortTimeFormat.values(), userStore.currentUserShortTimeFormat, 'shortTimeFormat', ShortTimeFormat.Default).isMeridiemIndicatorFirst || false;
    }

    function isLongTimeHourTwoDigits(): boolean {
        const longTimeFormat = getLocalizedLongTimeFormat();
        return longTimeFormat.indexOf('HH') >= 0 || longTimeFormat.indexOf('hh') >= 0;
    }

    function isLongTimeMinuteTwoDigits(): boolean {
        return getLocalizedLongTimeFormat().indexOf('mm') >= 0;
    }

    function isLongTimeSecondTwoDigits(): boolean {
        return getLocalizedLongTimeFormat().indexOf('ss') >= 0;
    }

    return {
        getAllLocalizedDigits, getLocaleDefaultCalendarDisplayType, getLocaleDefaultDateDisplayType,
        getLocaleDefaultNumeralSystemType, getLocaleDefaultDecimalSeparator, getLocaleDefaultDigitGroupingSymbol,
        getLocaleDefaultDigitGroupingType, getLocaleFiscalYearFormatType, getCurrentCalendarDisplayType,
        getCurrentDateDisplayType, getCurrentNumeralSystemType, getCurrentDecimalSeparator, getCurrentDigitGroupingSymbol,
        getCurrentDigitGroupingType, getCurrentFiscalYearFormatType, getCurrencyName, getLongDateFormatOrder,
        getShortDateFormatOrder, isLongDateMonthAfterYear, isShortDateMonthAfterYear, isLongTime24HourFormat,
        isLongTimeMeridiemIndicatorFirst, isShortTime24HourFormat, isShortTimeMeridiemIndicatorFirst, isLongTimeHourTwoDigits,
        isLongTimeMinuteTwoDigits, isLongTimeSecondTwoDigits,
    };
}
