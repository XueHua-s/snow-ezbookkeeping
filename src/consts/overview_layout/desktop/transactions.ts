import { RECENT_TRANSACTIONS_WIDGET_DEFAULT_ITEM_COUNT, WIDGET_TITLE_SETTING } from '@/consts/overview_layout/settings.ts';
import type { PartialRecord } from '@/core/base.ts';
import { TransactionCalendarHeatmapOutlierColorMode } from '@/core/chart.ts';
import {
    type DesktopOverviewWidgetDefinition,
    OverviewWidgetDataRequirement,
    type OverviewWidgetSettingValue,
    OverviewWidgetType
} from '@/core/overview_layout.ts';
import { TransactionAmountType, TransactionType } from '@/core/transaction.ts';

export const DESKTOP_OVERVIEW_WIDGET_DEFINITIONS_TRANSACTIONS: PartialRecord<OverviewWidgetType, DesktopOverviewWidgetDefinition> = {
    [OverviewWidgetType.RecentTransactions]: {
        type: OverviewWidgetType.RecentTransactions,
        name: 'Recent Transactions',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'itemCountSelect',
                settingName: 'itemCount',
                displayName: 'Item Count',
                itemCountValues: [3, 4, 5, 6, 7, 8, 9, 10]
            },
            {

                settingType: 'accountSelect',
                settingName: 'accountIds',
                displayName: 'Account'
            },
            {
                settingType: 'categorySelect',
                settingName: 'categoryIds',
                displayName: 'Category'
            },
            {
                settingType: 'tagSelect',
                settingName: 'tagFilter',
                displayName: 'Tags'
            },
            {
                settingType: 'amountFilter',
                settingName: 'amountFilter',
                displayName: 'Amount'
            },
            {
                settingType: 'textbox',
                settingName: 'keyword',
                displayName: 'Description',
                placeholder: 'Filter Description'
            }
        ],
        defaultSettings: {
            itemCount: RECENT_TRANSACTIONS_WIDGET_DEFAULT_ITEM_COUNT,
            accountIds: [],
            categoryIds: [],
            tagFilter: '',
            amountFilter: '',
            keyword: ''
        },
        defaultWidth: 3,
        defaultHeight: 3,
        minWidth: 2,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.RecentTransactions
        ]
    },
    [OverviewWidgetType.TransactionCalendar]: {
        type: OverviewWidgetType.TransactionCalendar,
        name: 'Transaction Calendar',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'customSelect',
                settingName: 'totalAmountType',
                displayName: 'Total Amount Calculation Method',
                selectValues: [
                    { name: 'Inflows and Outflows', value: TransactionAmountType.InflowsAndOutflows },
                    { name: 'Income and Expense', value: TransactionAmountType.IncomeAndExpense }
                ]
            },
            {
                settingType: 'customSelect',
                settingName: 'transactionTypes',
                displayName: 'Transaction Type',
                condition: (settings?: Record<string, OverviewWidgetSettingValue>) => {
                    return settings?.['totalAmountType'] === TransactionAmountType.InflowsAndOutflows;
                },
                selectValues: [
                    {
                        name: 'Inflow',
                        value: TransactionType.Income
                    },
                    {
                        name: 'Outflow',
                        value: TransactionType.Expense
                    }
                ],
                multiple: true,
                minSelections: 1
            },
            {
                settingType: 'customSelect',
                settingName: 'transactionTypes',
                displayName: 'Transaction Type',
                condition: (settings?: Record<string, OverviewWidgetSettingValue>) => {
                    return settings?.['totalAmountType'] === TransactionAmountType.IncomeAndExpense;
                },
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
                settingType: 'switch',
                settingName: 'showAlternateDate',
                displayName: 'Show Alternate Date'
            },
            {
                settingType: 'switch',
                settingName: 'showAmount',
                displayName: 'Show Amount'
            }
        ],
        defaultSettings: {
            totalAmountType: TransactionAmountType.InflowsAndOutflows,
            transactionTypes: [
                TransactionType.Income,
                TransactionType.Expense
            ],
            showAlternateDate: true,
            showAmount: true
        },
        defaultWidth: 6,
        defaultHeight: 6,
        minWidth: 3,
        minHeight: 6,
        dataRequirements: [
            OverviewWidgetDataRequirement.CurrentMonthTransactions
        ]
    },
    [OverviewWidgetType.TransactionCalendarHeatmap]: {
        type: OverviewWidgetType.TransactionCalendarHeatmap,
        name: 'Transaction Calendar Heatmap',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'customSelect',
                settingName: 'transactionType',
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
                ]
            },
            {
                settingType: 'monthSelect',
                settingName: 'months',
                displayName: 'Date Range',
                monthValues: [6, 12]
            },
            {
                settingType: 'customSelect',
                settingName: 'outlierColorMode',
                displayName: 'Color Outliers Separately',
                selectValues: [
                    {
                        name: 'Disabled',
                        value: TransactionCalendarHeatmapOutlierColorMode.Disabled
                    },
                    {
                        name: 'Days with Highest Amounts',
                        value: TransactionCalendarHeatmapOutlierColorMode.TopCount
                    },
                    {
                        name: 'Amount Threshold',
                        value: TransactionCalendarHeatmapOutlierColorMode.AboveAmount
                    }
                ]
            },
            {
                settingType: 'itemCountSelect',
                settingName: 'outlierTopCount',
                displayName: 'Outlier Day Count',
                itemCountValues: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
                condition: (settings?: Record<string, OverviewWidgetSettingValue>) => {
                    return settings?.['outlierColorMode'] === TransactionCalendarHeatmapOutlierColorMode.TopCount;
                }
            },
            {
                settingType: 'amount',
                settingName: 'outlierAmountThreshold',
                displayName: 'Outlier Amount Threshold',
                condition: (settings?: Record<string, OverviewWidgetSettingValue>) => {
                    return settings?.['outlierColorMode'] === TransactionCalendarHeatmapOutlierColorMode.AboveAmount;
                }
            }
        ],
        defaultSettings: {
            transactionType: TransactionType.Expense,
            months: 12,
            outlierColorMode: TransactionCalendarHeatmapOutlierColorMode.Disabled,
            outlierTopCount: 1,
            outlierAmountThreshold: 0
        },
        defaultWidth: 6,
        defaultHeight: 3,
        minWidth: 3,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.DailyTransactionAmounts
        ]
    },
};
