import { AMOUNT_FACTOR } from '@/consts/numeral.ts';
import {
    ChartValueType
} from '@/core/chart.ts';
import {
    type CurrencyPrependAndAppendText,
    CurrencyDisplayType
} from '@/core/currency.ts';
import {
    type BigDecimal,
    type BigDecimalWithSuffix,
    type HiddenAmount,
    DigitGroupingType,
    NumeralSystem
} from '@/core/numeral.ts';
import {
    isBoolean,
    isObject,
    isString
} from '@/lib/common.ts';
import {
    appendCurrencySymbol,
    getAmountPrependAndAppendCurrencySymbol
} from '@/lib/currency.ts';
import {
    formatAmount,
    formatBigDecimal,
    formatExchangeRateAmount,
    formatHiddenAmount,
    formatNumber,
    formatPercent,
    getAdaptiveDisplayAmountRate,
    isBigDecimal,
    parseAmount
} from '@/lib/numeral.ts';
import type { useHelpersFormatOptions } from './format-options.ts';
import type { useHelpersPreferences } from './preferences.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    getNumberFormatOptions: ReturnType<typeof useHelpersFormatOptions>['getNumberFormatOptions'];
    userStore: ReturnType<typeof useHelpersState>['userStore'];
    getCurrentCurrencyDisplayType: ReturnType<typeof useHelpersFormatOptions>['getCurrentCurrencyDisplayType'];
    getCurrentNumeralSystemType: ReturnType<typeof useHelpersPreferences>['getCurrentNumeralSystemType'];
    getCurrencyName: ReturnType<typeof useHelpersPreferences>['getCurrencyName'];
    getCurrencyUnitName: ReturnType<typeof useHelpersFormatOptions>['getCurrencyUnitName'];
}

// Owns amount formatting for the helpers store.
export function useHelpersAmountFormatting(context: Context) {
    const { getNumberFormatOptions, userStore, getCurrentCurrencyDisplayType, getCurrentNumeralSystemType, getCurrencyName, getCurrencyUnitName } = context;

    function getParsedAmountNumber(value: string, numeralSystem?: NumeralSystem): number {
        const numberFormatOptions = getNumberFormatOptions({ numeralSystem });
        return parseAmount(value, numberFormatOptions);
    }

    function getFormattedAmount(value: BigDecimal, numeralSystem?: NumeralSystem, digitGrouping?: DigitGroupingType, currencyCode?: string): string {
        const numberFormatOptions = getNumberFormatOptions({ numeralSystem, digitGrouping, currencyCode });
        return formatAmount(value, numberFormatOptions);
    }

    function getFormattedAmountWithCurrency(value: BigDecimal | HiddenAmount | BigDecimalWithSuffix, currencyCode?: string | false, currencyDisplayType?: CurrencyDisplayType, numeralSystem?: NumeralSystem, decimalSeparator?: string): string {
        let finalCurrencyCode = '';

        if (!isBoolean(currencyCode) && !currencyCode) {
            finalCurrencyCode = userStore.currentUserDefaultCurrency;
        } else if (isBoolean(currencyCode) && !currencyCode) {
            finalCurrencyCode = '';
        } else {
            finalCurrencyCode = currencyCode;
        }

        if (!currencyDisplayType) {
            currencyDisplayType = getCurrentCurrencyDisplayType();
        }

        if (!numeralSystem) {
            numeralSystem = getCurrentNumeralSystemType();
        }

        let suffix = '';

        if (isObject(value) && 'suffix' in value && value.value && isString(value.suffix)) {
            suffix = value.suffix;
            value = value.value;
        }

        const numberFormatOptions = getNumberFormatOptions({ numeralSystem, decimalSeparator, currencyCode: finalCurrencyCode });
        const currencyName = getCurrencyName(finalCurrencyCode);

        if (isBigDecimal(value)) {
            const isPlural: boolean = value.notEquals(AMOUNT_FACTOR) && value.notEquals(-AMOUNT_FACTOR);
            const textualValue = formatAmount(value, numberFormatOptions);

            if (!finalCurrencyCode) {
                return textualValue;
            }

            const currencyUnit = getCurrencyUnitName(finalCurrencyCode, isPlural);
            const ret = appendCurrencySymbol(textualValue, currencyDisplayType, finalCurrencyCode, currencyUnit, currencyName, isPlural);

            if (suffix) {
                return ret + suffix;
            } else {
                return ret;
            }
        } else if (isString(value)) {
            const isPlural: boolean = true;
            const textualValue = formatHiddenAmount(value, numberFormatOptions);

            if (!finalCurrencyCode) {
                return textualValue;
            }

            const currencyUnit = getCurrencyUnitName(finalCurrencyCode, isPlural);
            return appendCurrencySymbol(textualValue, currencyDisplayType, finalCurrencyCode, currencyUnit, currencyName, isPlural);
        } else {
            return '';
        }
    }

    function getFormattedNumber(value: number, numeralSystem?: NumeralSystem, digitGrouping?: DigitGroupingType, precision?: number): string {
        const numberFormatOptions = getNumberFormatOptions({ numeralSystem, digitGrouping: digitGrouping });
        return formatNumber(value, numberFormatOptions, precision);
    }

    function getFormattedBigDecimal(value: BigDecimal, numeralSystem?: NumeralSystem, digitGrouping?: DigitGroupingType, precision?: number): string {
        const numberFormatOptions = getNumberFormatOptions({ numeralSystem, digitGrouping: digitGrouping });
        return formatBigDecimal(value, numberFormatOptions, precision);
    }

    function getFormattedPercentValue(value: number, precision: number, lowPrecisionValue: string, numeralSystem?: NumeralSystem): string {
        const numberFormatOptions = getNumberFormatOptions({ numeralSystem });
        return formatPercent(value, precision, lowPrecisionValue, numberFormatOptions);
    }

    function getFormattedChartValue(value: BigDecimal, valueType: ChartValueType, currencyCode?: string) {
        if (valueType === ChartValueType.Amount) {
            return getFormattedAmountWithCurrency(value, currencyCode);
        } else if (valueType === ChartValueType.Percent) {
            return getFormattedPercentValue(value.toDoubleNumber(), 2, '<0.01');
        } else {
            return getFormattedBigDecimal(value, undefined, undefined, 4);
        }
    }

    function getFormattedVolume(value: number, precision?: number, unit?: 'KiB' | 'MiB'): string {
        const numberFormatOptions = getNumberFormatOptions({});
        let displayUnit = unit || 'B';

        if (unit === 'KiB') {
            value = value / 1024.0;
        } else if (unit === 'MiB') {
            value = value / 1024.0 / 1024.0;
        } else {
            displayUnit = 'B';

            if (value >= 1024.0) {
                value = value / 1024.0;
                displayUnit = 'KiB';
            }

            if (value >= 1024.0) {
                value = value / 1024.0;
                displayUnit = 'MiB';
            }
        }

        return formatNumber(value, numberFormatOptions, precision) + ' ' + displayUnit;
    }

    function getFormattedExchangeRateAmount(value: BigDecimal, numeralSystem?: NumeralSystem): string {
        const numberFormatOptions = getNumberFormatOptions({ numeralSystem });
        return formatExchangeRateAmount(value, numberFormatOptions);
    }

    function getAdaptiveAmountRate(amount1: number, amount2: number, fromExchangeRate?: { rate: string }, toExchangeRate?: { rate: string }): string | null {
        const numberFormatOptions = getNumberFormatOptions({});
        return getAdaptiveDisplayAmountRate(amount1, amount2, numberFormatOptions, fromExchangeRate, toExchangeRate);
    }

    function getAmountPrependAndAppendText(currencyCode: string, isPlural: boolean): CurrencyPrependAndAppendText | null {
        const currencyDisplayType = getCurrentCurrencyDisplayType();
        const currencyUnit = getCurrencyUnitName(currencyCode, isPlural);
        const currencyName = getCurrencyName(currencyCode);
        return getAmountPrependAndAppendCurrencySymbol(currencyDisplayType, currencyCode, currencyUnit, currencyName, isPlural);
    }

    return {
        getParsedAmountNumber, getFormattedAmount, getFormattedAmountWithCurrency, getFormattedNumber, getFormattedBigDecimal,
        getFormattedPercentValue, getFormattedChartValue, getFormattedVolume, getFormattedExchangeRateAmount,
        getAdaptiveAmountRate, getAmountPrependAndAppendText,
    };
}
