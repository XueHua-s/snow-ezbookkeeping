import {
    type NameValue,
    type TypeAndDisplayName,
    type TypeAndName,
    type TypeAndNameWithAlternativeName
} from '@/core/base.ts';
import {
    CalendarDisplayType,
    DateDisplayType
} from '@/core/calendar.ts';
import {
    CurrencyDisplayType
} from '@/core/currency.ts';
import {
    ImageUploadQualityType
} from '@/core/image.ts';
import {
    type BigDecimal,
    type LocalizedNumeralSymbolType,
    type NumeralSymbolType,
    NumeralSystem
} from '@/core/numeral.ts';
import {
    ChartDateAggregationType,
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import {
    isNumber
} from '@/lib/common.ts';
import {
    parseBigDecimal
} from '@/lib/numeral.ts';
import type { useHelpersAmountFormatting } from './amount-formatting.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
    userStore: ReturnType<typeof useHelpersState>['userStore'];
    getFormattedAmountWithCurrency: ReturnType<typeof useHelpersAmountFormatting>['getFormattedAmountWithCurrency'];
    getFormattedNumber: ReturnType<typeof useHelpersAmountFormatting>['getFormattedNumber'];
}

// Owns options for the helpers store.
export function useHelpersOptions(context: Context) {
    const { t, userStore, getFormattedAmountWithCurrency, getFormattedNumber } = context;

    function getAllCurrencyDisplayTypes(numeralSystem: NumeralSystem, decimalSeparator: string): TypeAndDisplayName[] {
        const defaultCurrencyDisplayTypeName = t('default.currencyDisplayType');
        let defaultCurrencyDisplayType = CurrencyDisplayType.parse(defaultCurrencyDisplayTypeName);

        if (!defaultCurrencyDisplayType) {
            defaultCurrencyDisplayType = CurrencyDisplayType.Default;
        }

        const defaultCurrency = userStore.currentUserDefaultCurrency;

        const ret = [];
        const sampleBigDecimalValue: BigDecimal = parseBigDecimal(12345);
        const defaultSampleValue = getFormattedAmountWithCurrency(sampleBigDecimalValue, defaultCurrency, defaultCurrencyDisplayType, numeralSystem, decimalSeparator);

        ret.push({
            type: CurrencyDisplayType.LanguageDefaultType,
            displayName: `${t('Language Default')} (${defaultSampleValue})`
        });

        const allCurrencyDisplayTypes = CurrencyDisplayType.values();

        for (const type of allCurrencyDisplayTypes) {
            const sampleValue = getFormattedAmountWithCurrency(sampleBigDecimalValue, defaultCurrency, type, numeralSystem, decimalSeparator);
            const displayName = `${t(type.name)} (${sampleValue})`

            ret.push({
                type: type.type,
                displayName: displayName
            });
        }

        return ret;
    }

    function getAllLocalizedCalendarTypes(allCalendarDisplayTypeArray: CalendarDisplayType[] | DateDisplayType[], localeDefaultType: CalendarDisplayType | DateDisplayType | undefined, systemDefaultType: CalendarDisplayType | DateDisplayType, languageDefaultValue: number): TypeAndDisplayName[] {
        let defaultType: TypeAndName | undefined = localeDefaultType;

        if (!defaultType) {
            defaultType = systemDefaultType;
        }

        const ret: TypeAndDisplayName[] = [];

        ret.push({
            type: languageDefaultValue,
            displayName: `${t('Language Default')} (${t('calendar.' + defaultType.name)})`
        });

        for (const calendarDisplayType of allCalendarDisplayTypeArray) {
            ret.push({
                type: calendarDisplayType.type,
                displayName: t('calendar.' + calendarDisplayType.name)
            });
        }

        return ret;
    }

    function getAllImageUploadQualityTypes(): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];

        for (const qualityType of ImageUploadQualityType.values()) {
            if (isNumber(qualityType.maxLongSidePixels)) {
                ret.push({
                    type: qualityType.type,
                    displayName: t(`format.volume.${qualityType.name}`, {
                        size: qualityType.estimatedKiB ? getFormattedNumber(qualityType.estimatedKiB) : '-',
                    })
                });
            } else {
                ret.push({
                    type: qualityType.type,
                    displayName: t(qualityType.name)
                })
            }
        }

        return ret;
    }

    function getLocalizedDisplayNameAndType(typeAndNames: TypeAndName[] | TypeAndNameWithAlternativeName[], useAlternativeName?: boolean): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];

        for (const typeAndName of typeAndNames) {
            let name: string = typeAndName.name;

            if (useAlternativeName && 'alternativeName' in typeAndName && typeAndName.alternativeName) {
                name = typeAndName.alternativeName;
            }

            ret.push({
                type: typeAndName.type,
                displayName: t(name)
            });
        }

        return ret;
    }

    function getLocalizedNameValue(nameValues: NameValue[]): NameValue[] {
        const ret: NameValue[] = [];

        for (const nameValue of nameValues) {
            ret.push({
                name: t(nameValue.name),
                value: nameValue.value
            });
        }

        return ret;
    }

    function getLocalizedDisplayNameAndTypeWithSystemDefault(typeAndNames: TypeAndName[], defaultValue: number, defaultType: TypeAndName): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];

        ret.push({
            type: defaultValue,
            displayName: t('System Default') + (defaultType.name ? ` (${t(defaultType.name)})` : '')
        });

        for (const typeAndName of typeAndNames) {
            ret.push({
                type: typeAndName.type,
                displayName: t(typeAndName.name)
            });
        }

        return ret;
    }

    function getLocalizedNumeralSeparatorFormats<T extends NumeralSymbolType>(allSeparatorArray: T[], localeDefaultType: T | undefined, systemDefaultType: T, languageDefaultValue: number): LocalizedNumeralSymbolType[] {
        let defaultSeparatorType: T | undefined = localeDefaultType;

        if (!defaultSeparatorType) {
            defaultSeparatorType = systemDefaultType;
        }

        const ret: LocalizedNumeralSymbolType[] = [];

        ret.push({
            type: languageDefaultValue,
            symbol: defaultSeparatorType.symbol,
            displayName: `${t('Language Default')} (${defaultSeparatorType.symbol})`
        });

        for (const separator of allSeparatorArray) {
            ret.push({
                type: separator.type,
                symbol: separator.symbol,
                displayName: `${t('numeral.' + separator.name)} (${separator.symbol})`
            });
        }

        return ret;
    }

    function getLocalizedChartDateAggregationTypeAndDisplayName(analysisType: StatisticsAnalysisType, fullName: boolean, includeBillingCycle: boolean): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];
        const allTypes: ChartDateAggregationType[] = ChartDateAggregationType.values(analysisType);

        for (const type of allTypes) {
            ret.push({
                type: type.type,
                displayName: t(fullName ? type.fullName : `granularity.${type.shortName}`)
            });
        }

        if (includeBillingCycle) {
            ret.push({
                type: ChartDateAggregationType.BillingCycle.type,
                displayName: t(fullName ? ChartDateAggregationType.BillingCycle.fullName : `granularity.${ChartDateAggregationType.BillingCycle.shortName}`)
            });
        }

        return ret;
    }

    return {
        getAllCurrencyDisplayTypes, getAllLocalizedCalendarTypes, getAllImageUploadQualityTypes, getLocalizedDisplayNameAndType,
        getLocalizedNameValue, getLocalizedDisplayNameAndTypeWithSystemDefault, getLocalizedNumeralSeparatorFormats,
        getLocalizedChartDateAggregationTypeAndDisplayName,
    };
}
