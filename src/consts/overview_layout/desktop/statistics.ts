import { WIDGET_TITLE_SETTING } from '@/consts/overview_layout/settings.ts';
import type { PartialRecord } from '@/core/base.ts';
import { DateRange } from '@/core/datetime.ts';
import {
    type DesktopOverviewWidgetDefinition,
    OverviewWidgetDataRequirement,
    type OverviewWidgetSettingValue,
    OverviewWidgetType
} from '@/core/overview_layout.ts';
import { TrendChartType } from '@/core/statistics.ts';
import { TransactionType } from '@/core/transaction.ts';

export const DESKTOP_OVERVIEW_WIDGET_DEFINITIONS_STATISTICS: PartialRecord<OverviewWidgetType, DesktopOverviewWidgetDefinition> = {
    [OverviewWidgetType.PeriodIncomeExpense]: {
        type: OverviewWidgetType.PeriodIncomeExpense,
        name: 'Period Income and Expense',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'customSelect',
                settingName: 'dateRange',
                displayName: 'Date Range',
                selectValues: [
                    DateRange.Today,
                    DateRange.ThisWeek,
                    DateRange.ThisMonth,
                    DateRange.ThisYear
                ].map(dateRange => ({
                    name: dateRange.name,
                    value: dateRange.type
                }))
            }
        ],
        defaultSettings: {
            dateRange: DateRange.Today.type
        },
        defaultWidth: 3,
        defaultHeight: 3,
        minWidth: 2,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.TransactionOverview
        ]
    },
    [OverviewWidgetType.PeriodNetIncomeAndSavingsRate]: {
        type: OverviewWidgetType.PeriodNetIncomeAndSavingsRate,
        name: 'Period Net Income and Savings Rate',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'customSelect',
                settingName: 'dateRange',
                displayName: 'Date Range',
                selectValues: [
                    DateRange.Today,
                    DateRange.ThisWeek,
                    DateRange.ThisMonth,
                    DateRange.ThisYear
                ].map(dateRange => ({
                    name: dateRange.name,
                    value: dateRange.type
                }))
            }
        ],
        defaultSettings: {
            dateRange: DateRange.ThisMonth.type
        },
        defaultWidth: 3,
        defaultHeight: 3,
        minWidth: 2,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.TransactionOverview
        ]
    },
    [OverviewWidgetType.IncomeExpenseTrend]: {
        type: OverviewWidgetType.IncomeExpenseTrend,
        name: 'Income and Expense Trends',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'customSelect',
                settingName: 'transactionTypes',
                displayName: 'Transaction Type',
                selectValues: [
                    {
                        name: 'Income',
                        value: TransactionType.Income
                    },
                    {
                        name: 'Expense',
                        value: TransactionType.Expense
                    }
                ],
                multiple: true,
                minSelections: 1
            },
            {
                settingType: 'customSelect',
                settingName: 'chartType',
                displayName: 'Chart Type',
                selectValues: [
                    {
                        name: TrendChartType.Column.name,
                        value: TrendChartType.Column.type
                    },
                    {
                        name: TrendChartType.Area.name,
                        value: TrendChartType.Area.type
                    }
                ]
            },
            {
                settingType: 'monthSelect',
                settingName: 'months',
                displayName: 'Date Range',
                monthValues: [6, 12]
            },
            {
                settingType: 'switch',
                settingName: 'smoothCurve',
                displayName: 'Smooth Curve',
                condition: (settings?: Record<string, OverviewWidgetSettingValue>) => {
                    return settings?.['chartType'] === TrendChartType.Area.type;
                }
            },
            {
                settingType: 'switch',
                settingName: 'showXAxisLabels',
                displayName: 'Show Horizontal Axis Labels'
            },
            {
                settingType: 'switch',
                settingName: 'showLegend',
                displayName: 'Show Legend'
            }
        ],
        defaultSettings: {
            chartType: TrendChartType.Column.type,
            transactionTypes: [
                TransactionType.Income,
                TransactionType.Expense
            ],
            months: 12,
            smoothCurve: false,
            showXAxisLabels: true,
            showLegend: true
        },
        defaultWidth: 6,
        defaultHeight: 6,
        minWidth: 3,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.TransactionOverviewLast12Months
        ]
    },
    [OverviewWidgetType.NetAssetsTrend]: {
        type: OverviewWidgetType.NetAssetsTrend,
        name: 'Net Assets Trends',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'monthSelect',
                settingName: 'months',
                displayName: 'Date Range',
                monthValues: [6, 12]
            },
            {
                settingType: 'switch',
                settingName: 'smoothCurve',
                displayName: 'Smooth Curve'
            },
            {
                settingType: 'switch',
                settingName: 'showXAxisLabels',
                displayName: 'Show Horizontal Axis Labels'
            },
            {
                settingType: 'switch',
                settingName: 'showLegend',
                displayName: 'Show Legend'
            }
        ],
        defaultSettings: {
            months: 12,
            smoothCurve: false,
            showXAxisLabels: true,
            showLegend: true
        },
        defaultWidth: 6,
        defaultHeight: 6,
        minWidth: 3,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.Accounts,
            OverviewWidgetDataRequirement.AssetTrends
        ]
    },
    [OverviewWidgetType.ExpenseCategoryRanking]: {
        type: OverviewWidgetType.ExpenseCategoryRanking,
        name: 'Expense Category Ranking',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'customSelect',
                settingName: 'dateRange',
                displayName: 'Date Range',
                selectValues: [
                    DateRange.ThisMonth,
                    DateRange.ThisYear
                ].map(dateRange => ({
                    name: dateRange.name,
                    value: dateRange.type
                }))
            },
            {
                settingType: 'customSelect',
                settingName: 'categoryLevel',
                displayName: 'Category Level',
                selectValues: [
                    {
                        name: 'Primary Category',
                        value: 'primary'
                    },
                    {
                        name: 'Secondary Category',
                        value: 'secondary'
                    }
                ]
            },
            {
                settingType: 'itemCountSelect',
                settingName: 'itemCount',
                displayName: 'Item Count',
                itemCountValues: [3, 4, 5, 6, 7, 8, 9, 10]
            }
        ],
        defaultSettings: {
            dateRange: DateRange.ThisMonth.type,
            categoryLevel: 'primary',
            itemCount: 3
        },
        defaultWidth: 3,
        defaultHeight: 3,
        minWidth: 2,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.Accounts,
            OverviewWidgetDataRequirement.TransactionCategories,
            OverviewWidgetDataRequirement.TransactionCategoryStatistics
        ]
    },
};
