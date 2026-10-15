import { DEFAULT_CHART_COLORS } from '@/consts/color.ts';
import { AccountCategory } from '@/core/account.ts';
import type { ColorValue } from '@/core/color.ts';
import type { ApplicationSettings, LocaleDefaultSettings } from '@/core/setting.ts';
import {
    getObjectOwnFieldCount,
    isHextualColor
} from '@/lib/common.ts';
import {
    getApplicationSettings,
    getLocaleDefaultSettings
} from '@/lib/settings.ts';
import type { ComputedRef, Ref, UnwrapRef } from 'vue';
import { computed, ref } from 'vue';
interface SettingState {
    appSettings: Ref<UnwrapRef<ApplicationSettings>>;
    syncedAppSettings: Ref<UnwrapRef<Record<string, boolean>>>;
    localeDefaultSettings: Ref<UnwrapRef<LocaleDefaultSettings>>;
    enableApplicationCloudSync: ComputedRef<boolean>;
    chartColorList: ComputedRef<ColorValue[]>;
    accountCategoryDisplayOrders: ComputedRef<Record<number, number>>;
}



// Owns state for the setting store.
export function useSettingState(): SettingState  {
    const appSettings = ref<ApplicationSettings>(getApplicationSettings());

    const syncedAppSettings = ref<Record<string, boolean>>({});

    const localeDefaultSettings = ref<LocaleDefaultSettings>(getLocaleDefaultSettings());

    const enableApplicationCloudSync = computed<boolean>(() => getObjectOwnFieldCount(syncedAppSettings.value) > 0);

    const chartColorList = computed<ColorValue[]>(() => {
        const chartColors = appSettings.value.chartColors;

        if (!chartColors) {
            return [...DEFAULT_CHART_COLORS];
        }

        const colors = chartColors.split(',').map(c => c.trim().toLowerCase()).filter(c => isHextualColor(c));
        return colors.length > 0 ? colors : [...DEFAULT_CHART_COLORS];
    });

    const accountCategoryDisplayOrders = computed<Record<number, number>>(() => AccountCategory.allDisplayOrders(appSettings.value.accountCategoryOrders));

    return { appSettings, syncedAppSettings, localeDefaultSettings, enableApplicationCloudSync, chartColorList, accountCategoryDisplayOrders };
}
