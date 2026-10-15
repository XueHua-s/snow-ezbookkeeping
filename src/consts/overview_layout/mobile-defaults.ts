import { DateRange } from '@/core/datetime.ts';
import {
    type MobileOverviewLayout,
    OverviewWidgetType
} from '@/core/overview_layout.ts';


export const DEFAULT_MOBILE_OVERVIEW_LAYOUT: MobileOverviewLayout = {
    widgets: [
        {
            id: 'default-current-month-overview',
            type: OverviewWidgetType.CurrentMonthOverview,
            settings: {
                height: 3
            }
        },
        {
            id: 'default-period-income-expense',
            type: OverviewWidgetType.PeriodIncomeExpense,
            settings: {
                dateRanges: [
                    DateRange.Today.type,
                    DateRange.ThisWeek.type,
                    DateRange.ThisMonth.type,
                    DateRange.ThisYear.type
                ]
            }
        }
    ]
};
