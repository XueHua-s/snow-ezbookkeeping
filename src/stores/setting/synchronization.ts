import {
    values
} from '@/core/base.ts';
import {
    ALL_ALLOWED_CLOUD_SYNC_APP_SETTING_KEY_TYPES,
    type ApplicationCloudSetting,
    type LocaleDefaultSettings,
    UserApplicationCloudSettingType
} from '@/core/setting.ts';
import {
    arrayItemToObjectField,
    isBoolean,
    isObject
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import {
    clearSettings,
    getApplicationSettings
} from '@/lib/settings.ts';
import type { useSettingCloud } from './cloud.ts';
import type { useSettingState } from './state.ts';

interface Context {
    appSettings: ReturnType<typeof useSettingState>['appSettings'];
    createUserApplicationCloudSetting: ReturnType<typeof useSettingCloud>['createUserApplicationCloudSetting'];
    syncedAppSettings: ReturnType<typeof useSettingState>['syncedAppSettings'];
    updateApplicationSettingsValueAndAppSettingsFromCloudSetting: ReturnType<typeof useSettingCloud>['updateApplicationSettingsValueAndAppSettingsFromCloudSetting'];
    localeDefaultSettings: ReturnType<typeof useSettingState>['localeDefaultSettings'];
}

// Owns synchronization for the setting store.
export function useSettingSynchronization(context: Context) {
    const { appSettings, createUserApplicationCloudSetting, syncedAppSettings, updateApplicationSettingsValueAndAppSettingsFromCloudSetting, localeDefaultSettings } = context;

    function clearAppSettings(): void {
        clearSettings();
        appSettings.value = getApplicationSettings();
    }

    function createApplicationCloudSettings(applicationSettingKeys: string[]): ApplicationCloudSetting[] {
        if (!applicationSettingKeys || applicationSettingKeys.length < 1) {
            return [];
        }

        const settings: ApplicationCloudSetting[] = [];

        for (const settingKey of applicationSettingKeys) {
            const cloudSetting = createUserApplicationCloudSetting(settingKey);

            if (cloudSetting) {
                settings.push(cloudSetting);
            }
        }

        return settings;
    }

    function setApplicationSettingsFromCloudSettings(cloudSettings?: ApplicationCloudSetting[]): void {
        if (!cloudSettings || cloudSettings.length < 1) {
            syncedAppSettings.value = {};
            return;
        }

        syncedAppSettings.value = arrayItemToObjectField(cloudSettings.map(item => item.settingKey), true);

        for (const setting of cloudSettings) {
            if (!setting || !setting.settingKey) {
                continue;
            }

            const settingType = ALL_ALLOWED_CLOUD_SYNC_APP_SETTING_KEY_TYPES[setting.settingKey];

            if (!settingType) {
                logger.warn(`cannot load application cloud setting "${setting.settingKey}", because it is not supported to sync`);
                continue;
            }

            if (settingType === UserApplicationCloudSettingType.String) {
                updateApplicationSettingsValueAndAppSettingsFromCloudSetting(setting.settingKey, setting.settingValue);
            } else if (settingType === UserApplicationCloudSettingType.Number) {
                const value = parseFloat(setting.settingValue);

                if (isNaN(value)) {
                    logger.warn(`cannot load application cloud setting "${setting.settingKey}", because it has invalid number value`);
                    continue;
                }

                updateApplicationSettingsValueAndAppSettingsFromCloudSetting(setting.settingKey, value);
            } else if (settingType === UserApplicationCloudSettingType.Boolean) {
                if (setting.settingValue !== 'true' && setting.settingValue !== 'false') {
                    logger.warn(`cannot load application cloud setting "${setting.settingKey}", because it has invalid boolean value`);
                    continue;
                }

                updateApplicationSettingsValueAndAppSettingsFromCloudSetting(setting.settingKey, setting.settingValue === 'true');
            } else if (settingType === UserApplicationCloudSettingType.StringBooleanMap) {
                try {
                    const map = JSON.parse(setting.settingValue);
                    let isValid = isObject(map);

                    if (isValid) {
                        for (const value of values(map)) {
                            if (!isBoolean(value)) {
                                isValid = false;
                                break;
                            }
                        }
                    }

                    if (!isValid) {
                        logger.warn(`cannot load application cloud setting "${setting.settingKey}", because it has invalid map value`);
                        continue;
                    }

                    updateApplicationSettingsValueAndAppSettingsFromCloudSetting(setting.settingKey, map as Record<string, boolean>);
                } catch (error) {
                    logger.warn(`cannot load application cloud setting "${setting.settingKey}", because cannot parse JSON (${error})`);
                }
            } else {
                logger.warn(`cannot load application cloud setting "${setting.settingKey}", because it has unknown type "${settingType}"`);
            }
        }
    }

    function updateApplicationSyncSettingKeys(settingKeys?: string[]): void {
        if (!settingKeys || settingKeys.length < 1) {
            syncedAppSettings.value = {};
        } else {
            syncedAppSettings.value = arrayItemToObjectField(settingKeys, true);
        }
    }

    function updateLocalizedDefaultSettings(newLocaleDefaultSettings: LocaleDefaultSettings | null) {
        if (!newLocaleDefaultSettings) {
            return;
        }

        localeDefaultSettings.value.currency = newLocaleDefaultSettings.currency;
        localeDefaultSettings.value.firstDayOfWeek = newLocaleDefaultSettings.firstDayOfWeek;
    }

    return { clearAppSettings, createApplicationCloudSettings, setApplicationSettingsFromCloudSettings, updateApplicationSyncSettingKeys, updateLocalizedDefaultSettings };
}
