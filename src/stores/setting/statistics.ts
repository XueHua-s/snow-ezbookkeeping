import {
    updateApplicationSettingsSubValue
} from '@/lib/settings.ts';
import type { useSettingCloud } from './cloud.ts';
import type { useSettingState } from './state.ts';

interface Context {
    appSettings: ReturnType<typeof useSettingState>['appSettings'];
    updateUserApplicationCloudSettingValue: ReturnType<typeof useSettingCloud>['updateUserApplicationCloudSettingValue'];
}

// Owns statistics for the setting store.
export function useSettingStatistics(context: Context) {
    const { appSettings, updateUserApplicationCloudSettingValue } = context;

    // Statistics Settings
    function setStatisticsDefaultChartDataType(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultChartDataType', value);
        appSettings.value.statistics.defaultChartDataType = value;
        updateUserApplicationCloudSettingValue('statistics.defaultChartDataType', value);
    }

    function setStatisticsDefaultTimezoneType(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultTimezoneType', value);
        appSettings.value.statistics.defaultTimezoneType = value;
        updateUserApplicationCloudSettingValue('statistics.defaultTimezoneType', value);
    }

    function setStatisticsDefaultAccountFilter(value: Record<string, boolean>): void {
        updateApplicationSettingsSubValue('statistics', 'defaultAccountFilter', value);
        appSettings.value.statistics.defaultAccountFilter = value;
        updateUserApplicationCloudSettingValue('statistics.defaultAccountFilter', value);
    }

    function setStatisticsDefaultTransactionCategoryFilter(value: Record<string, boolean>): void {
        updateApplicationSettingsSubValue('statistics', 'defaultTransactionCategoryFilter', value);
        appSettings.value.statistics.defaultTransactionCategoryFilter = value;
        updateUserApplicationCloudSettingValue('statistics.defaultTransactionCategoryFilter', value);
    }

    function setStatisticsDefaultKeywordMatchMode(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultKeywordMatchMode', value);
        appSettings.value.statistics.defaultKeywordMatchMode = value;
        updateUserApplicationCloudSettingValue('statistics.defaultKeywordMatchMode', value);
    }

    function setStatisticsSortingType(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultSortingType', value);
        appSettings.value.statistics.defaultSortingType = value;
        updateUserApplicationCloudSettingValue('statistics.defaultSortingType', value);
    }

    function setStatisticsDefaultCategoricalChartType(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultCategoricalChartType', value);
        appSettings.value.statistics.defaultCategoricalChartType = value;
        updateUserApplicationCloudSettingValue('statistics.defaultCategoricalChartType', value);
    }

    function setStatisticsDefaultCategoricalChartDateRange(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultCategoricalChartDataRangeType', value);
        appSettings.value.statistics.defaultCategoricalChartDataRangeType = value;
        updateUserApplicationCloudSettingValue('statistics.defaultCategoricalChartDataRangeType', value);
    }

    function setStatisticsDefaultTrendChartType(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultTrendChartType', value);
        appSettings.value.statistics.defaultTrendChartType = value;
        updateUserApplicationCloudSettingValue('statistics.defaultTrendChartType', value);
    }

    function setStatisticsDefaultTrendChartDateRange(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultTrendChartDataRangeType', value);
        appSettings.value.statistics.defaultTrendChartDataRangeType = value;
        updateUserApplicationCloudSettingValue('statistics.defaultTrendChartDataRangeType', value);
    }

    function setStatisticsDefaultAssetTrendsChartType(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultAssetTrendsChartType', value);
        appSettings.value.statistics.defaultAssetTrendsChartType = value;
        updateUserApplicationCloudSettingValue('statistics.defaultAssetTrendsChartType', value);
    }

    function setStatisticsDefaultAssetTrendsChartDateRange(value: number): void {
        updateApplicationSettingsSubValue('statistics', 'defaultAssetTrendsChartDataRangeType', value);
        appSettings.value.statistics.defaultAssetTrendsChartDataRangeType = value;
        updateUserApplicationCloudSettingValue('statistics.defaultAssetTrendsChartDataRangeType', value);
    }

    return {
        setStatisticsDefaultChartDataType, setStatisticsDefaultTimezoneType, setStatisticsDefaultAccountFilter,
        setStatisticsDefaultTransactionCategoryFilter, setStatisticsDefaultKeywordMatchMode, setStatisticsSortingType,
        setStatisticsDefaultCategoricalChartType, setStatisticsDefaultCategoricalChartDateRange,
        setStatisticsDefaultTrendChartType, setStatisticsDefaultTrendChartDateRange, setStatisticsDefaultAssetTrendsChartType,
        setStatisticsDefaultAssetTrendsChartDateRange,
    };
}
