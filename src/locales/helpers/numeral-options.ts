import {
    type TypeAndDisplayName
} from '@/core/base.ts';
import {
    type LocalizedDigitGroupingType,
    DigitGroupingType,
    NumeralSystem
} from '@/core/numeral.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
}

// Owns numeral options for the helpers store.
export function useHelpersNumeralOptions(context: Context) {
    const { t } = context;

    function getAllNumeralSystemTypes(): TypeAndDisplayName[] {
        const defaultNumeralSystemTypeName = t('default.numeralSystem');
        let defaultNumeralSystemType = NumeralSystem.parse(defaultNumeralSystemTypeName);

        if (!defaultNumeralSystemType) {
            defaultNumeralSystemType = NumeralSystem.Default;
        }

        const ret: TypeAndDisplayName[] = [];

        ret.push({
            type: NumeralSystem.LanguageDefaultType,
            displayName: `${t('Language Default')} (${defaultNumeralSystemType.textualAllDigits})`
        });

        const allNumeralSystemTypes = NumeralSystem.values();

        for (const type of allNumeralSystemTypes) {
            ret.push({
                type: type.type,
                displayName: `${t('numeral.' + type.name)} (${type.textualAllDigits})`
            });
        }

        return ret;
    }

    function getAllDigitGroupingTypes(numeralSystem: NumeralSystem, digitGroupingSymbol: string): LocalizedDigitGroupingType[] {
        const defaultDigitGroupingTypeName = t('default.digitGrouping');
        let defaultDigitGroupingType = DigitGroupingType.parse(defaultDigitGroupingTypeName);

        if (!defaultDigitGroupingType) {
            defaultDigitGroupingType = DigitGroupingType.Default;
        }

        const ret: LocalizedDigitGroupingType[] = [];

        ret.push({
            type: DigitGroupingType.LanguageDefaultType,
            enabled: defaultDigitGroupingType.enabled,
            displayName: `${t('Language Default')} (${t('numeral.' + defaultDigitGroupingType.name)})`
        });

        const allDigitGroupingTypes = DigitGroupingType.values();
        const numberCharacters = numeralSystem.replaceWesternArabicDigitsToLocalizedDigits('123456789').split('');

        for (const type of allDigitGroupingTypes) {
            const sampleValue = type.format(numberCharacters, digitGroupingSymbol);

            ret.push({
                type: type.type,
                enabled: type.enabled,
                displayName: `${t('numeral.' + type.name)} (${sampleValue})`
            });
        }

        return ret;
    }

    return { getAllNumeralSystemTypes, getAllDigitGroupingTypes };
}
