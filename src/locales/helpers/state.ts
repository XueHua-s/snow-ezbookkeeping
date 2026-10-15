import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { useI18n as useVueI18n } from 'vue-i18n';


// Owns state for the helpers store.
export function useHelpersState() {
    const { t, locale, getLocaleMessage, setLocaleMessage } = useVueI18n();

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const exchangeRatesStore = useExchangeRatesStore();

    return { t, locale, getLocaleMessage, setLocaleMessage, settingsStore, userStore, exchangeRatesStore };
}
