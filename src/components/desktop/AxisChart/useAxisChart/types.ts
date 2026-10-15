import type { ColorStyleValue } from '@/core/color.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import type { SortableTransactionStatisticDataItem } from '@/models/transaction.ts';

export type AxisChartDisplayType = 'line' | 'area' | 'column' | 'bubble';

export interface AxisChartData {
    allSeries: AxisChartDataItem[];
    allOriginalData: BigDecimal[][];
}

export interface AxisChartDataItem {
    id: string;
    name: string;
    itemStyle: {
        color: ColorStyleValue;
    };
    selected: boolean;
    type: string;
    areaStyle?: object;
    stack?: string;
    smooth?: boolean;
    showSymbol?: boolean;
    symbolSize?: (data: number) => number;
    animation: boolean;
    data: number[];  // only used for echarts rendering, the actual value is in allOriginalData
}

export interface AxisChartTooltipItem extends SortableTransactionStatisticDataItem {
    readonly id: string;
    readonly name: string;
    readonly color: unknown;
    readonly displayOrders: number[];
    readonly value: BigDecimal;
}
