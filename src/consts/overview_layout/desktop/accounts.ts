import { WIDGET_TITLE_SETTING } from '@/consts/overview_layout/settings.ts';
import type { PartialRecord } from '@/core/base.ts';
import {
    type DesktopOverviewWidgetDefinition,
    OverviewWidgetDataRequirement,
    OverviewWidgetType
} from '@/core/overview_layout.ts';

export const DESKTOP_OVERVIEW_WIDGET_DEFINITIONS_ACCOUNTS: PartialRecord<OverviewWidgetType, DesktopOverviewWidgetDefinition> = {
    [OverviewWidgetType.AssetSummary]: {
        type: OverviewWidgetType.AssetSummary,
        name: 'Asset Summary',
        supportsSettings: [
            WIDGET_TITLE_SETTING
        ],
        defaultSettings: {},
        defaultWidth: 8,
        defaultHeight: 3,
        minWidth: 3,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.Accounts
        ]
    },
    [OverviewWidgetType.AccountBalanceList]: {
        type: OverviewWidgetType.AccountBalanceList,
        name: 'Account Balance List',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'accountSelect',
                settingName: 'accountIds',
                displayName: 'Account',
                disableHiddenAccounts: true
            },
            {
                settingType: 'itemCountSelect',
                settingName: 'itemCount',
                displayName: 'Item Count',
                itemCountValues: [3, 4, 5, 6, 7, 8, 9, 10]
            },
            {
                settingType: 'customSelect',
                settingName: 'sortBy',
                displayName: 'Sort By',
                selectValues: [
                    {
                        name: 'Display Order',
                        value: 'displayOrder'
                    },
                    {
                        name: 'Balance',
                        value: 'balance'
                    }
                ]
            },
            {
                settingType: 'switch',
                settingName: 'alwaysShowAmount',
                displayName: 'Always Show Amount'
            },
            {
                settingType: 'switch',
                settingName: 'showAvailableCreditForCreditCard',
                displayName: 'Show Available Credit for Credit Cards'
            }
        ],
        defaultSettings: {
            accountIds: [],
            itemCount: 4,
            sortBy: 'displayOrder',
            alwaysShowAmount: false,
            showAvailableCreditForCreditCard: false
        },
        defaultWidth: 3,
        defaultHeight: 3,
        minWidth: 2,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.Accounts
        ]
    },
    [OverviewWidgetType.AccountUnreconciledTransactions]: {
        type: OverviewWidgetType.AccountUnreconciledTransactions,
        name: 'Account Unreconciled Transactions',
        supportsSettings: [
            WIDGET_TITLE_SETTING,
            {
                settingType: 'accountSelect',
                settingName: 'accountIds',
                displayName: 'Account',
                disableHiddenAccounts: true
            },
            {
                settingType: 'itemCountSelect',
                settingName: 'itemCount',
                displayName: 'Item Count',
                itemCountValues: [3, 4, 5, 6, 7, 8, 9, 10]
            },
            {
                settingType: 'customSelect',
                settingName: 'sortBy',
                displayName: 'Sort By',
                selectValues: [
                    {
                        name: 'Display Order',
                        value: 'displayOrder'
                    },
                    {
                        name: 'Transaction Count',
                        value: 'transactionCount'
                    }
                ]
            }
        ],
        defaultSettings: {
            accountIds: [],
            itemCount: 4,
            sortBy: 'displayOrder'
        },
        defaultWidth: 3,
        defaultHeight: 3,
        minWidth: 2,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.Accounts,
            OverviewWidgetDataRequirement.AccountUnreconciledTransactionCounts
        ]
    },
    [OverviewWidgetType.CurrentMonthOverview]: {
        type: OverviewWidgetType.CurrentMonthOverview,
        name: 'This Month\'s Income and Expense Overview',
        supportsSettings: [],
        defaultSettings: {},
        defaultWidth: 4,
        defaultHeight: 3,
        minWidth: 3,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.TransactionOverview
        ]
    },
    [OverviewWidgetType.CurrentMonthExpenseProgress]: {
        type: OverviewWidgetType.CurrentMonthExpenseProgress,
        name: 'This Month\'s Expense Progress',
        supportsSettings: [
            WIDGET_TITLE_SETTING
        ],
        defaultSettings: {},
        defaultWidth: 3,
        defaultHeight: 3,
        minWidth: 2,
        minHeight: 3,
        dataRequirements: [
            OverviewWidgetDataRequirement.TransactionOverviewLast2Months
        ]
    },
};
