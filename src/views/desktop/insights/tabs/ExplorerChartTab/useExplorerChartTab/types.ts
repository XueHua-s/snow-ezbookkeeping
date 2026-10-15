import AxisChart from '@/components/desktop/AxisChart.vue';
import CustomChart from '@/components/desktop/CustomChart.vue';
import HeatMapChart from '@/components/desktop/HeatMapChart.vue';
import HierarchyChart from '@/components/desktop/HierarchyChart.vue';
import RadarChart from '@/components/desktop/RadarChart.vue';
import { type AxisChartSourceDataItem } from '@/core/chart.ts';
import { type BigDecimal } from '@/core/numeral.ts';
import type { SortableTransactionStatisticDataItem } from '@/models/transaction.ts';
import {
    type CategoriedTransactionExplorerData,
    TransactionExplorerDimensionType
} from '@/stores/explorer.ts';
import TransactionListDialog from '@/views/desktop/insights/dialogs/TransactionListDialog.vue';

export type AxisChartType = InstanceType<typeof AxisChart>;

export type RadarChartType = InstanceType<typeof RadarChart>;

export type HierarchyChartType = InstanceType<typeof HierarchyChart>;

export type HeatMapChartType = InstanceType<typeof HeatMapChart>;

export type CustomChartType = InstanceType<typeof CustomChart>;

export type TransactionListDialogType = InstanceType<typeof TransactionListDialog>;

export interface InsightsExplorerDataTableTabProps {
    loading?: boolean;
    disabled?: boolean;
}

export interface CategoryDimensionData extends SortableTransactionStatisticDataItem {
    id: string;
    dimension: TransactionExplorerDimensionType;
    name: string;
    displayOrders: number[];
    value: BigDecimal;
}

export interface SortableCategoriedTransactionExplorerDataItem extends SortableTransactionStatisticDataItem {
    name: string;
    displayOrders: number[];
    value: BigDecimal;
    originalItem: CategoriedTransactionExplorerData;
}

export interface SeriesDimensionData extends SortableTransactionStatisticDataItem, AxisChartSourceDataItem, Record<string, unknown> {
    id: string;
    dimension: TransactionExplorerDimensionType;
    name: string;
    displayOrders: number[];
    values: BigDecimal[];
    value: BigDecimal;
}
