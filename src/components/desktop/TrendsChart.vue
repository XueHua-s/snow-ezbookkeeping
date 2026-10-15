<script setup lang="ts">
import {
    type TrendsBarChartClickEvent,
    type TrendsChartDateType
} from '@/components/base/TrendsChartBase.ts';
import AxisChart from './AxisChart.vue';
import { type DesktopTrendsChartProps, useTrendsChart } from './TrendsChart/useTrendsChart.ts';

const props = defineProps<DesktopTrendsChartProps<TrendsChartDateType>>();

const emit = defineEmits<{
    (e: 'click', value: TrendsBarChartClickEvent): void;
}>();

const {
    tt, chartDisplayType, allTooltipExtraColumnNames, allDisplayDateRanges, allSeriesData, getTooltipExtraColumnTotalValues,
    getTooltipExtraColumnValues, clickItem, exportData,
} = useTrendsChart(props, emit);

defineExpose({
    exportData
})

</script>

<template>
    <axis-chart class="trends-chart-container" ref="axisChart" values-field="values"
                :skeleton="skeleton" :no-animation="noAnimation" :type="chartDisplayType" :stacked="stacked" :sorting-type="sortingType"
                :show-value="showValue"
                :show-total-amount-in-tooltip="showTotalAmountInTooltip" :total-name-in-tooltip="tt('Total Amount')"
                :category-type-name="tt('Date')" :all-category-names="allDisplayDateRanges"
                :items="allSeriesData" :value-type="valueType"
                :hide-legend="hideLegend" :legend-position="legendPosition"
                :hide-x-axis-labels="hideXAxisLabels" :hide-x-axis-line="hideXAxisLine"
                :hide-y-axis-labels="hideYAxisLabels"
                :hide-horizontal-grid-lines="hideHorizontalGridLines"
                :hide-line-symbols="hideLineSymbols"
                :no-margin="noMargin"
                :smooth-curve="smoothCurve"
                :translate-name="translateName"
                :default-currency="defaultCurrency" :use-custom-color="useCustomColor"
                :enable-click-item="enableClickItem"
                :tooltip-extra-column-names="allTooltipExtraColumnNames"
                :tooltip-extra-column-total-values="showYearOverYear || showPeriodOverPeriod ? getTooltipExtraColumnTotalValues : undefined"
                :tooltip-extra-column-values="showYearOverYear || showPeriodOverPeriod ? getTooltipExtraColumnValues : undefined"
                @click="clickItem"
                v-if="chartDisplayType"
    />
</template>



<style scoped>
.trends-chart-container {
    width: 100%;
    height: 650px;
}

@media (min-width: 600px) {
    .trends-chart-container {
        height: 700px;
    }
}
</style>
