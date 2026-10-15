import type { PartialRecord } from '@/core/base.ts';
import type { DesktopOverviewWidgetDefinition, OverviewWidgetType } from '@/core/overview_layout.ts';
import { DESKTOP_OVERVIEW_WIDGET_DEFINITIONS_ACCOUNTS } from './desktop/accounts.ts';
import { DESKTOP_OVERVIEW_WIDGET_DEFINITIONS_STATISTICS } from './desktop/statistics.ts';
import { DESKTOP_OVERVIEW_WIDGET_DEFINITIONS_TRANSACTIONS } from './desktop/transactions.ts';

export const DESKTOP_OVERVIEW_WIDGET_DEFINITIONS: PartialRecord<OverviewWidgetType, DesktopOverviewWidgetDefinition> = {
    ...DESKTOP_OVERVIEW_WIDGET_DEFINITIONS_ACCOUNTS,
    ...DESKTOP_OVERVIEW_WIDGET_DEFINITIONS_STATISTICS,
    ...DESKTOP_OVERVIEW_WIDGET_DEFINITIONS_TRANSACTIONS,
};
