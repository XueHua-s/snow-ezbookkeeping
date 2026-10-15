import { ALL_CURRENCIES_A_C } from './currency/a-c.ts';
import { ALL_CURRENCIES_D_I } from './currency/d-i.ts';
import { ALL_CURRENCIES_J_O } from './currency/j-o.ts';
import { ALL_CURRENCIES_P_S } from './currency/p-s.ts';
import { ALL_CURRENCIES_T_Z } from './currency/t-z.ts';
import type { CurrencyInfo } from '@/core/currency.ts';

// ISO 4217
// Reference: https://www.six-group.com/dam/download/financial-information/data-center/iso-currrency/lists/list-one.xml
// Reference: https://github.com/unicode-org/cldr/blob/main/common/supplemental/supplementalData.xml
export const ALL_CURRENCIES: Record<string, CurrencyInfo> = {
    ...ALL_CURRENCIES_A_C,
    ...ALL_CURRENCIES_D_I,
    ...ALL_CURRENCIES_J_O,
    ...ALL_CURRENCIES_P_S,
    ...ALL_CURRENCIES_T_Z,
};

export const DEFAULT_CURRENCY_SYMBOL: string = '¤';
export const DEFAULT_CURRENCY_CODE: string = (ALL_CURRENCIES['USD'] as CurrencyInfo).code;
export const ACCOUNT_CURRENCY_NOT_SET_VALUE: string = '---';
