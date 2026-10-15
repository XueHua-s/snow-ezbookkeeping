import { DateRange } from '@/core/datetime.ts';
import {
    type DesktopOverviewLayout,
    OverviewWidgetType
} from '@/core/overview_layout.ts';
import { TrendChartType } from '@/core/statistics.ts';
import { TransactionAmountType, TransactionType } from '@/core/transaction.ts';


export const DEFAULT_DESKTOP_OVERVIEW_LAYOUT: DesktopOverviewLayout = {
    widgets: [
        {
            id: 'default-current-month-overview',
            type: OverviewWidgetType.CurrentMonthOverview,
            x: 0,
            y: 0,
            w: 4,
            h: 3,
            settings: {}
        },
        {
            id: 'default-asset-summary',
            type: OverviewWidgetType.AssetSummary,
            x: 4,
            y: 0,
            w: 8,
            h: 3,
            settings: {}
        },
        {
            id: 'default-transaction-calendar-heatmap',
            type: OverviewWidgetType.TransactionCalendarHeatmap,
            x: 0,
            y: 3,
            w: 8,
            h: 3,
            settings: {
                transactionType: TransactionType.Expense,
                months: 12
            }
        },
        {
            id: 'default-transaction-calendar',
            type: OverviewWidgetType.TransactionCalendar,
            x: 8,
            y: 3,
            w: 4,
            h: 6,
            settings: {
                totalAmountType: TransactionAmountType.InflowsAndOutflows,
                transactionTypes: [
                    TransactionType.Income,
                    TransactionType.Expense
                ],
                showAlternateDate: false,
                showAmount: true
            }
        },
        {
            id: 'default-expense-category-ranking',
            type: OverviewWidgetType.ExpenseCategoryRanking,
            x: 0,
            y: 6,
            w: 4,
            h: 3,
            settings: {
                dateRange: DateRange.ThisMonth.type,
                categoryLevel: 'primary',
                itemCount: 3
            }
        },
        {
            id: 'default-expense-trend',
            type: OverviewWidgetType.IncomeExpenseTrend,
            x: 4,
            y: 6,
            w: 4,
            h: 3,
            settings: {
                chartType: TrendChartType.Area.type,
                transactionTypes: [
                    TransactionType.Expense
                ],
                months: 12,
                smoothCurve: true,
                showXAxisLabels: false,
                showLegend: false
            }
        },
        {
            id: 'default-today-income-expense',
            type: OverviewWidgetType.PeriodIncomeExpense,
            x: 0,
            y: 9,
            w: 3,
            h: 3,
            settings: {
                dateRange: DateRange.Today.type
            }
        },
        {
            id: 'default-week-income-expense',
            type: OverviewWidgetType.PeriodIncomeExpense,
            x: 3,
            y: 9,
            w: 3,
            h: 3,
            settings: {
                dateRange: DateRange.ThisWeek.type
            }
        },
        {
            id: 'default-month-income-expense',
            type: OverviewWidgetType.PeriodIncomeExpense,
            x: 6,
            y: 9,
            w: 3,
            h: 3,
            settings: {
                dateRange: DateRange.ThisMonth.type
            }
        },
        {
            id: 'default-year-income-expense',
            type: OverviewWidgetType.PeriodIncomeExpense,
            x: 9,
            y: 9,
            w: 3,
            h: 3,
            settings: {
                dateRange: DateRange.ThisYear.type
            }
        }
    ]
};
