<script setup lang="ts">
import { type AxisChartSourceDataItem, ChartValueType } from '@/core/chart.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { type AxisChartDisplayType, useAxisChart } from './AxisChart/useAxisChart.ts';
export type { AxisChartDisplayType } from './AxisChart/useAxisChart.ts';

const props = defineProps<{
    class?: string;
    skeleton?: boolean;
    noAnimation?: boolean;
    type: AxisChartDisplayType;
    stacked?: boolean;
    hideLegend?: boolean;
    legendPosition?: 'top' | 'bottom';
    hideXAxisLabels?: boolean;
    hideXAxisLine?: boolean;
    hideYAxisLabels?: boolean;
    hideHorizontalGridLines?: boolean;
    hideLineSymbols?: boolean;
    noMargin?: boolean;
    smoothCurve?: boolean;
    oneHundredPercentStacked?: boolean;
    sortingType: number;
    showValue?: boolean;
    showTotalAmountInTooltip?: boolean;
    totalNameInTooltip?: string;
    categoryTypeName: string;
    allCategoryNames: string[];
    items: AxisChartSourceDataItem[];
    valueType: ChartValueType;
    translateName?: boolean;
    defaultCurrency?: string;
    useCustomColor?: boolean;
    enableClickItem?: boolean;
    tooltipExtraColumnNames?: string[];
    tooltipExtraColumnTotalValues?: (categoryIndex: number, totalValue: BigDecimal, visibleSeriesIds: string[]) => string[];
    tooltipExtraColumnValues?: (seriesId: string, categoryIndex: number, currentValue: BigDecimal) => string[];
}>();

const emit = defineEmits<{
    (e: 'click', itemId: string, categoryIndex: number, item: AxisChartSourceDataItem): void;
}>();

const { finalClass, chartOptions, clickItem, exportData, onLegendSelectChanged } = useAxisChart(props, emit);

defineExpose({
    exportData
});

</script>

<template>
    <v-chart autoresize :class="finalClass"
             :option="chartOptions" :update-options="{ notMerge: true }"
             @click="clickItem" @legendselectchanged="onLegendSelectChanged" />
</template>



<style scoped>
.axis-chart-container {
    width: 100%;
    height: 560px;
    margin-top: 10px;
}

@media (min-width: 600px) {
    .axis-chart-container {
        height: 650px;
    }
}
</style>
