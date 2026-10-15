import type { SetupProps } from '@/lib/ui/props.ts';
import { ALL_CURRENCIES } from '@/consts/currency.ts';
import { AMOUNT_FACTOR } from '@/consts/numeral.ts';
import { type BigDecimal, NumeralSystem } from '@/core/numeral.ts';
import { isNumber } from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import { BIG_DECIMAL_ZERO, isBigDecimal, parseBigDecimal } from '@/lib/numeral.ts';
import { isiOS, useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useI18n } from '@/locales/helpers.ts';
import { computed, ref, watch } from 'vue';



// Coordinates NumberPadSheet state and effects; called once from the component setup.
export function useNumberPadSheet(props: SetupProps<{
    modelValue: number;
    minValue?: number;
    maxValue?: number;
    currency?: string;
    flipNegative?: boolean;
    hint?: string;
    show: boolean;
}>, emit: {
    (e: 'update:modelValue', value: number): void;
    (e: 'update:show', value: boolean): void;
}) {
    const {
        tt,
        getAllLocalizedDigits,
        getCurrentNumeralSystemType,
        getCurrentDecimalSeparator,
        parseAmountFromLocalizedNumerals,
        parseAmountFromWesternArabicNumerals,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping,
        appendDigitGroupingSymbolAndDecimalSeparator
    } = useI18n();

    const { showToast } = useI18nUIComponents();

    const isSupportClipboard = !!navigator.clipboard;

    const previousValue = ref<string>('');

    const currentSymbol = ref<string>('');

    const currentValue = ref<string>(getInitedStringValue(props.modelValue, props.flipNegative));

    const pastingAmount = ref<boolean>(false);

    const showPastePopover = ref<boolean>(false);

    const numeralSystem = computed<NumeralSystem>(() => getCurrentNumeralSystemType());

    const digits = computed<string[]>(() => getAllLocalizedDigits());

    const decimalSeparator = computed<string>(() => getCurrentDecimalSeparator());

    const supportDecimalSeparator = computed<boolean>(() => {
        if (!props.currency || !ALL_CURRENCIES[props.currency] || !isNumber(ALL_CURRENCIES[props.currency]!.fraction)) {
            return true;
        }

        return (ALL_CURRENCIES[props.currency]!.fraction as number) > 0;
    });

    const currentDisplay = computed<string>(() => {
        const finalPreviousValue = appendDigitGroupingSymbolAndDecimalSeparator(numeralSystem.value.replaceWesternArabicDigitsToLocalizedDigits(previousValue.value));
        let finalCurrentValue = currentValue.value;
        let currentValueSuffix = '';

        if (finalCurrentValue && finalCurrentValue[finalCurrentValue.length - 1] === decimalSeparator.value) {
            finalCurrentValue = finalCurrentValue.substring(0, finalCurrentValue.length - 1);
            currentValueSuffix = decimalSeparator.value;
        }

        finalCurrentValue = appendDigitGroupingSymbolAndDecimalSeparator(numeralSystem.value.replaceWesternArabicDigitsToLocalizedDigits(finalCurrentValue));

        if (currentValueSuffix) {
            finalCurrentValue += currentValueSuffix;
        }

        if (currentSymbol.value) {
            return `${finalPreviousValue} ${currentSymbol.value} ${finalCurrentValue}`;
        } else {
            return finalCurrentValue;
        }
    });

    const currentDisplayNumClass = computed<string>(() => {
        if (currentDisplay.value && currentDisplay.value.length >= 28) {
            return 'numpad-value-extra-small';
        } else if (currentDisplay.value && currentDisplay.value.length >= 22) {
            return 'numpad-value-small';
        } else if (currentDisplay.value && currentDisplay.value.length >= 16) {
            return 'numpad-value-normal';
        } else {
            return 'numpad-value-large';
        }
    });

    const confirmText = computed<string>(() => {
        if (currentSymbol.value) {
            return '=';
        } else {
            return tt('OK');
        }
    });

    function getInitedStringValue(value: number, flipNegative?: boolean): string {
        if (flipNegative) {
            value = -value;
        }

        return getStringValue(parseBigDecimal(value), true);
    }

    function getStringValue(value: BigDecimal, hideZero: boolean): string {
        if (!isBigDecimal(value)) {
            return '';
        }

        const textualNumber = formatAmountToWesternArabicNumeralsWithoutDigitGrouping(value, props.currency);

        const decimalSeparator = getCurrentDecimalSeparator();
        const decimalSeparatorPos = textualNumber.indexOf(decimalSeparator);

        if (decimalSeparatorPos < 0) {
            if (hideZero && textualNumber === NumeralSystem.WesternArabicNumerals.digitZero) {
                return '';
            }

            return textualNumber;
        }

        const integer = textualNumber.substring(0, decimalSeparatorPos);
        const decimals = textualNumber.substring(decimalSeparatorPos + 1, textualNumber.length);
        let newDecimals = '';

        for (let i = decimals.length - 1; i >= 0; i--) {
            if (decimals[i] !== NumeralSystem.WesternArabicNumerals.digitZero || newDecimals.length > 0) {
                newDecimals = decimals[i] + newDecimals;
            }
        }

        if (newDecimals.length < 1) {
            if (hideZero && integer === NumeralSystem.WesternArabicNumerals.digitZero) {
                return '';
            }

            return integer;
        }

        return `${integer}${decimalSeparator}${newDecimals}`;
    }

    function inputNum(num: number): void {
        if (!previousValue.value && currentSymbol.value === '−') {
            if (isNumber(props.minValue) && props.minValue >= 0) {
                return;
            }

            currentValue.value = '-' + currentValue.value;
            currentSymbol.value = '';
        }

        if (currentValue.value === NumeralSystem.WesternArabicNumerals.digitZero) {
            currentValue.value = num.toString();
            return;
        } else if (currentValue.value === `-${NumeralSystem.WesternArabicNumerals.digitZero}`) {
            currentValue.value = '-' + num.toString();
            return;
        }

        const decimalSeparatorPos = currentValue.value.indexOf(decimalSeparator.value);

        if (decimalSeparatorPos >= 0 && currentValue.value.substring(decimalSeparatorPos + 1, currentValue.value.length).length >= 2) {
            return;
        }

        const newValue = currentValue.value + num.toString();

        if (isNumber(props.minValue)) {
            const current = parseAmountFromWesternArabicNumerals(newValue);

            if (current < (props.minValue)) {
                return;
            }
        }

        if (isNumber(props.maxValue)) {
            const current = parseAmountFromWesternArabicNumerals(newValue);

            if (current > (props.maxValue)) {
                return;
            }
        }

        currentValue.value = newValue;
    }

    function inputDoubleNum(num: number): void {
        inputNum(num);
        inputNum(num);
    }

    function inputDecimalSeparator(): void {
        if (currentValue.value.indexOf(decimalSeparator.value) >= 0) {
            return;
        }

        if (!previousValue.value && currentSymbol.value === '−') {
            currentValue.value = '-' + currentValue.value;
            currentSymbol.value = '';
        }

        if (currentValue.value.length < 1) {
            currentValue.value = NumeralSystem.WesternArabicNumerals.digitZero;
        } else if (currentValue.value === '-') {
            currentValue.value = '-' + NumeralSystem.WesternArabicNumerals.digitZero;
        }

        currentValue.value = currentValue.value + decimalSeparator.value;
    }

    function setSymbol(symbol: string): void {
        if (currentValue.value) {
            if (currentSymbol.value) {
                const lastFormulaCalcResult = confirm();

                if (!lastFormulaCalcResult) {
                    return;
                }
            }

            previousValue.value = currentValue.value;
            currentValue.value = '';
        }

        currentSymbol.value = symbol;
    }

    function backspace(): void {
        if (!currentValue.value || currentValue.value.length < 1) {
            if (currentSymbol.value) {
                currentValue.value = previousValue.value;
                previousValue.value = '';
                currentSymbol.value = '';
            }

            return;
        }

        currentValue.value = currentValue.value.substring(0, currentValue.value.length - 1);
    }

    function clear(): void {
        currentValue.value = '';
        previousValue.value = '';
        currentSymbol.value = '';
    }

    function paste(): void {
        if (pastingAmount.value) {
            pastingAmount.value = false;
            return;
        }

        pastingAmount.value = true;

        navigator.clipboard.readText().then(text => {
            pastingAmount.value = false;

            if (!text) {
                return;
            }

            const parsedAmount = parseAmountFromLocalizedNumerals(text);

            if (Number.isNaN(parsedAmount) || !Number.isFinite(parsedAmount)) {
                showToast('Cannot parse amount from clipboard');
                return;
            }

            if (isNumber(props.minValue)) {
                if (parsedAmount < (props.minValue)) {
                    showToast('Numeric Overflow');
                    return;
                }
            }

            if (isNumber(props.maxValue)) {
                if (parsedAmount > (props.maxValue)) {
                    showToast('Numeric Overflow');
                    return;
                }
            }

            currentValue.value = getStringValue(parseBigDecimal(parsedAmount), false);
        }).catch(error => {
            // Do not set pastingAmount to false here
            // In iOS, system will show the paste context menu, if user click outside, the paste action should not be triggered again
            logger.error('failed to read clipboard text', error);
        });
    }

    function confirm(): boolean {
        if (currentSymbol.value && currentValue.value.length >= 1) {
            const previous: BigDecimal = parseBigDecimal(parseAmountFromWesternArabicNumerals(previousValue.value));
            const current: BigDecimal = parseBigDecimal(parseAmountFromWesternArabicNumerals(currentValue.value));
            let finalValue: BigDecimal = BIG_DECIMAL_ZERO;

            switch (currentSymbol.value) {
                case '+':
                    finalValue = previous.add(current);
                    break;
                case '−':
                    finalValue = previous.subtract(current);
                    break;
                case '×':
                    finalValue = previous.multiply(current).divide(AMOUNT_FACTOR).truncate();
                    break;
                default:
                    finalValue = previous;
            }

            if (isNumber(props.minValue)) {
                if (finalValue.lessThan(props.minValue)) {
                    showToast('Numeric Overflow');
                    return false;
                }
            }

            if (isNumber(props.maxValue)) {
                if (finalValue.greaterThan(props.maxValue)) {
                    showToast('Numeric Overflow');
                    return false;
                }
            }

            currentValue.value = getStringValue(finalValue, false);
            previousValue.value = '';
            currentSymbol.value = '';

            return true;
        } else if (currentSymbol.value && currentValue.value.length < 1) {
            currentValue.value = previousValue.value;
            previousValue.value = '';
            currentSymbol.value = '';

            return true;
        } else {
            let value: number = parseAmountFromWesternArabicNumerals(currentValue.value);

            if (props.flipNegative) {
                value = -value;
            }

            emit('update:modelValue', value);
            close();

            return true;
        }
    }

    function close(): void {
        emit('update:show', false);
    }

    function onSheetOpen(): void {
        currentValue.value = getInitedStringValue(props.modelValue, props.flipNegative);
        previousValue.value = '';
        currentSymbol.value = '';
    }

    function onSheetClosed(): void {
        close();
    }

    function onDisplayValueClick(): void {
        if (!isSupportClipboard) {
            return;
        }

        if (isiOS()) {
            paste();
        } else {
            showPastePopover.value = true;
        }
    }

    watch(() => props.flipNegative, (newValue) => {
        currentValue.value = getInitedStringValue(props.modelValue, newValue);
    });

    return {
        tt, currentSymbol, showPastePopover, digits, decimalSeparator, supportDecimalSeparator, currentDisplay,
        currentDisplayNumClass, confirmText, inputNum, inputDoubleNum, inputDecimalSeparator, setSymbol, backspace, clear, paste,
        confirm, close, onSheetOpen, onSheetClosed, onDisplayValueClick,
    };
}
