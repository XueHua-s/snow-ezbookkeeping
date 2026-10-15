import { ACCOUNT_CURRENCY_NOT_SET_VALUE, ALL_CURRENCIES } from '@/consts/currency.ts';
import {
    type LocalizedSwitchOption,
    type NameValue,
    keys
} from '@/core/base.ts';
import {
    type LocalizedCurrencyInfo
} from '@/core/currency.ts';
import {
    MeridiemIndicator
} from '@/core/datetime.ts';
import type { useHelpersPreferences } from './preferences.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
    getCurrencyName: ReturnType<typeof useHelpersPreferences>['getCurrencyName'];
}

// Owns currency options for the helpers store.
export function useHelpersCurrencyOptions(context: Context) {
    const { t, getCurrencyName } = context;

    function getAllEnableDisableOptions(): LocalizedSwitchOption[] {
        return [{
            value: true,
            displayName: t('Enable')
        }, {
            value: false,
            displayName: t('Disable')
        }];
    }

    function getAllCurrencies(withNotSet?: boolean): LocalizedCurrencyInfo[] {
        const allCurrencies: LocalizedCurrencyInfo[] = [];

        for (const currencyCode of keys(ALL_CURRENCIES)) {
            const localizedCurrencyInfo: LocalizedCurrencyInfo = {
                currencyCode: currencyCode,
                displayName: getCurrencyName(currencyCode)
            };

            allCurrencies.push(localizedCurrencyInfo);
        }

        allCurrencies.sort(function (c1, c2) {
            return c1.displayName.localeCompare(c2.displayName);
        })

        if (withNotSet) {
            allCurrencies.splice(0, 0, {
                currencyCode: ACCOUNT_CURRENCY_NOT_SET_VALUE,
                displayName: t('Not set')
            });
        }

        return allCurrencies;
    }

    function getAllMeridiemIndicators(): NameValue[] {
        const allMeridiemIndicators = MeridiemIndicator.values();
        const localizedMeridiemIndicatorNames = [];

        for (const indicator of allMeridiemIndicators) {
            localizedMeridiemIndicatorNames.push({
                name: t(`datetime.${indicator.name}.content`),
                value: indicator.name
            });
        }

        return localizedMeridiemIndicatorNames;
    }

    return { getAllEnableDisableOptions, getAllCurrencies, getAllMeridiemIndicators };
}
