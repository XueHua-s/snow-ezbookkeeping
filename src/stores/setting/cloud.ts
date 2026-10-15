import {
    ALL_ALLOWED_CLOUD_SYNC_APP_SETTING_KEY_TYPES,
    type ApplicationCloudSetting,
    type ApplicationSettingSubValue,
    type ApplicationSettingValue,
    UserApplicationCloudSettingType
} from '@/core/setting.ts';
import {
    isString
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import {
    updateApplicationSettingsSubValue,
    updateApplicationSettingsValue
} from '@/lib/settings.ts';
import type { useSettingState } from './state.ts';

interface Context {
    appSettings: ReturnType<typeof useSettingState>['appSettings'];
    syncedAppSettings: ReturnType<typeof useSettingState>['syncedAppSettings'];
}

// Owns cloud for the setting store.
export function useSettingCloud(context: Context) {
    const { appSettings, syncedAppSettings } = context;

    function updateApplicationSettingsValueAndAppSettingsFromCloudSetting(key: string, value: string | number | boolean | Record<string, boolean>): void {
        const keyItems = key.split('.');
        const keyFirstPart = keyItems[0] as string;

        if (keyItems.length === 1) {
            updateApplicationSettingsValue(keyFirstPart, value);
            appSettings.value[keyFirstPart] = value;
        } else if (keyItems.length === 2) {
            const subKey = keyItems[1] as string;
            updateApplicationSettingsSubValue(keyFirstPart, subKey, value);
            (appSettings.value[keyFirstPart] as Record<string, ApplicationSettingSubValue>)[subKey] = value;
        } else {
            logger.warn(`cannot load application cloud setting "${key}", because it has invalid key format`);
        }
    }

    function createUserApplicationCloudSetting(key: string): ApplicationCloudSetting | null {
        const settingType = ALL_ALLOWED_CLOUD_SYNC_APP_SETTING_KEY_TYPES[key];

        if (!settingType) {
            logger.warn(`cannot get application cloud setting "${key}", because it is not supported to sync`);
            return null;
        }

        const keyItems = key.split('.');
        let value: ApplicationSettingValue | ApplicationSettingSubValue = appSettings.value[key] as (ApplicationSettingValue | ApplicationSettingSubValue);

        if (keyItems.length === 2) {
            const primaryKey = keyItems[0] as string;
            const subKey = keyItems[1] as string;
            value = (appSettings.value[primaryKey] as Record<string, ApplicationSettingSubValue>)[subKey] as ApplicationSettingSubValue;
        } else if (keyItems.length > 2) {
            logger.warn(`cannot get application cloud setting "${key}", because it has invalid key format`);
            return null;
        }

        let settingValue = '';

        if (settingType === UserApplicationCloudSettingType.String) {
            if (!value) {
                settingValue = '';
            } else {
                settingValue = value.toString();
            }
        } else {
            settingValue = JSON.stringify(value);
        }

        return {
            settingKey: key,
            settingValue: settingValue
        };
    }

    function updateUserApplicationCloudSettingValue(key: string, value: string | number | boolean | Record<string, boolean>): void {
        if (!syncedAppSettings.value || !syncedAppSettings.value[key]) {
            return;
        }

        const settingType = ALL_ALLOWED_CLOUD_SYNC_APP_SETTING_KEY_TYPES[key];

        if (!settingType) {
            return;
        }

        const settingValue = isString(value) ? value : JSON.stringify(value);

        services.updateUserApplicationCloudSettings({
            settings: [{
                settingKey: key,
                settingValue: settingValue
            }],
            fullUpdate: false
        }).then(response => {
            const data = response.data;

            if (!data || !data.success || !data.result) {
                logger.debug(`failed to update user application cloud setting "${key}" with value "${settingValue}"`);
                return;
            }

            logger.debug(`update user application cloud setting "${key}" with value "${settingValue}" successfully`);
        }).catch(error => {
            logger.debug(`failed to update user application cloud setting "${key}" with value "${settingValue}"`, error);
        });
    }

    return { updateApplicationSettingsValueAndAppSettingsFromCloudSetting, createUserApplicationCloudSetting, updateUserApplicationCloudSettingValue };
}
