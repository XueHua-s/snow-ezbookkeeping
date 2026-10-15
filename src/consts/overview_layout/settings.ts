import type { OverviewWidgetColorSettingItem, OverviewWidgetSwitchSettingItem, OverviewWidgetTextboxSettingItem } from '@/core/overview_layout.ts';


export const DESKTOP_OVERVIEW_LAYOUT_COLUMNS: number = 12;

export const DESKTOP_OVERVIEW_LAYOUT_MAX_WIDGETS: number = 100;

export const DESKTOP_OVERVIEW_LAYOUT_MAX_ROWS: number = 1000;

export const MOBILE_OVERVIEW_LAYOUT_MAX_WIDGETS: number = 100;

export const RECENT_TRANSACTIONS_WIDGET_DEFAULT_ITEM_COUNT: number = 3;

export const WIDGET_TITLE_SETTING: OverviewWidgetTextboxSettingItem = {
    settingType: 'textbox',
    settingName: 'title',
    displayName: 'Widget Title',
    placeholder: 'Widget Title'
};

export const WIDGET_SHOW_TITLE_SETTING: OverviewWidgetSwitchSettingItem = {
    settingType: 'switch',
    settingName: 'showTitle',
    displayName: 'Show Title'
};

export const WIDGET_BACKGROUND_COLOR_SETTINGS: OverviewWidgetColorSettingItem[] = [
    {
        settingType: 'color',
        settingName: 'lightBackgroundColor',
        displayName: 'Light Mode Background Color'
    },
    {
        settingType: 'color',
        settingName: 'darkBackgroundColor',
        displayName: 'Dark Mode Background Color'
    }
];
