<script setup lang="ts">
import type { InsightsExplorerDataTableTabProps } from './ExplorerChartTab/useExplorerChartTab.ts';
import AxisChart from '@/components/desktop/AxisChart.vue';
import CustomChart from '@/components/desktop/CustomChart.vue';
import HeatMapChart from '@/components/desktop/HeatMapChart.vue';
import HierarchyChart from '@/components/desktop/HierarchyChart.vue';
import RadarChart from '@/components/desktop/RadarChart.vue';
import { ChartValueType } from '@/core/chart.ts';
import {
    TransactionExplorerChartType,
    TransactionExplorerDataDimension,
    TransactionExplorerDataDimensionType,
    TransactionExplorerValueMetric
} from '@/core/explorer.ts';
import { isDefined } from '@/lib/common.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import TransactionListDialog from '@/views/desktop/insights/dialogs/TransactionListDialog.vue';
import { useExplorerChartTab } from './ExplorerChartTab/useExplorerChartTab.ts';

const emit = defineEmits<{
    (e: 'click:transaction', value: TransactionInsightDataItem): void;
}>();

const {
    tt, explorersStore, defaultCurrency, allTransactionExplorerDataDimensions, allTransactionExplorerValueMetrics,
    allTransactionExplorerChartTypes, allTransactionExplorerChartSortingTypes,
    allTransactionExplorerCustomChartDisplayLayouts, currentTransactionExplorerCategoryDimensionName, currentExploration,
    isUsingAmountRange, allAmountRangeCounts, categoryDimensionTransactionExplorerData, categoriedNamesSortedByDisplayOrder,
    seriesDimensionTransactionExplorerData, axisChartDisplayType, hierarchyChartDisplayType, axisChartSmoothCurve,
    axisChartStacked, axisChart100PercentStacked, axisChartShowYearOverYear, axisChartShowPeriodOverPeriod,
    axisChartTooltipExtraColumnNames, isInAIMode, isAIGenerating, getAxisChartTooltipExtraColumnTotalValues,
    getAxisChartTooltipExtraColumnValues, updateChartType, updateCategoryDimensionType, switchToAIChartMode,
    onClickPieChartItem, onClickTrendChartItem, onClickHierarchyChartItem, onClickHeatmapChartItem,
    onClickCalendarHeatmapChartItem, onClickTransaction, buildExportResults,
} = useExplorerChartTab(emit);

defineExpose({
    isInAIMode,
    isAIGenerating,
    switchToAIChartMode,
    buildExportResults
});

defineProps<InsightsExplorerDataTableTabProps>();
</script>

<template>
    <v-card-text class="px-4 py-0">
        <v-row>
            <v-col cols="12">
                <div class="d-flex overflow-x-auto align-center gap-2 pt-2">
                    <v-select
                        class="flex-0-0"
                        min-width="150"
                        item-title="name"
                        item-value="value"
                        density="compact"
                        :disabled="loading || disabled || isInAIMode()"
                        :label="tt('Chart Type')"
                        :items="allTransactionExplorerChartTypes"
                        :model-value="currentExploration.chartType"
                        @update:model-value="updateChartType"
                    />
                    <v-select
                        class="flex-0-0"
                        min-width="150"
                        item-title="name"
                        item-value="value"
                        density="compact"
                        :disabled="loading || disabled || isDefined(TransactionExplorerChartType.valueOf(currentExploration.chartType)?.fixedCategoryDimension)"
                        :label="tt('Axis / Category')"
                        :items="allTransactionExplorerDataDimensions"
                        :model-value="TransactionExplorerChartType.valueOf(currentExploration.chartType)?.fixedCategoryDimension ?? currentExploration.categoryDimension"
                        @update:model-value="updateCategoryDimensionType"
                        v-if="currentExploration.chartType !== TransactionExplorerChartType.Custom.value"
                    />
                    <v-select
                        class="flex-0-0"
                        min-width="150"
                        item-title="name"
                        item-value="value"
                        density="compact"
                        :disabled="loading || disabled || !TransactionExplorerChartType.valueOf(currentExploration.chartType)?.seriesDimensionRequired"
                        :label="tt('Series')"
                        :items="allTransactionExplorerDataDimensions"
                        :model-value="TransactionExplorerChartType.valueOf(currentExploration.chartType)?.seriesDimensionRequired ? currentExploration.seriesDimension : TransactionExplorerDataDimension.None.value"
                        @update:model-value="currentExploration.seriesDimension = $event as TransactionExplorerDataDimensionType"
                        v-if="currentExploration.chartType !== TransactionExplorerChartType.Custom.value"
                    >
                        <template #item="{ props, internalItem }">
                            <v-list-item :disabled="internalItem.value === currentExploration.categoryDimension && internalItem.value !== TransactionExplorerDataDimension.SeriesDimensionDefault.value" v-bind="props">
                                <template #title>
                                    <div class="text-truncate">{{ internalItem.raw.name }}</div>
                                </template>
                            </v-list-item>
                        </template>
                    </v-select>
                    <v-select
                        class="flex-0-0"
                        min-width="220"
                        item-title="name"
                        item-value="value"
                        density="compact"
                        :disabled="loading || disabled"
                        :label="tt('Number of Amount Ranges')"
                        :items="allAmountRangeCounts"
                        v-model="currentExploration.amountRangeCount"
                        v-if="currentExploration.chartType !== TransactionExplorerChartType.Custom.value && isUsingAmountRange"
                    />
                    <v-select
                        class="flex-0-0"
                        min-width="150"
                        item-title="name"
                        item-value="value"
                        density="compact"
                        :disabled="loading || disabled"
                        :label="tt('Value Metric')"
                        :items="allTransactionExplorerValueMetrics"
                        v-model="currentExploration.valueMetric"
                        v-if="currentExploration.chartType !== TransactionExplorerChartType.Custom.value"
                    />
                    <v-select
                        class="flex-0-0"
                        min-width="150"
                        item-title="displayName"
                        item-value="type"
                        density="compact"
                        :disabled="loading || disabled || isDefined(TransactionExplorerChartType.valueOf(currentExploration.chartType)?.fixedSortingType)"
                        :label="tt('Sort Order')"
                        :items="allTransactionExplorerChartSortingTypes"
                        v-model="currentExploration.chartSortingType"
                        v-if="currentExploration.chartType !== TransactionExplorerChartType.Custom.value"
                    />
                    <v-select
                        class="flex-0-0"
                        min-width="150"
                        item-title="displayName"
                        item-value="type"
                        density="compact"
                        :disabled="loading || disabled || isInAIMode()"
                        :label="tt('Display Layout')"
                        :items="allTransactionExplorerCustomChartDisplayLayouts"
                        v-model="currentExploration.customChartDisplayLayout"
                        v-if="currentExploration.chartType === TransactionExplorerChartType.Custom.value"
                    />
                    <v-spacer class="flex-1-1"/>
                </div>
            </v-col>
        </v-row>
    </v-card-text>
    <v-card-text :class="{ 'readonly': loading }" v-if="currentExploration.chartType === TransactionExplorerChartType.Pie.value || currentExploration.chartType === TransactionExplorerChartType.Donut.value || currentExploration.chartType === TransactionExplorerChartType.NightingaleRose.value">
        <pie-chart
            :items="[
                { id: '1', name: '---', value: parseBigDecimal(60), color: '7c7c7f' },
                { id: '2', name: '---', value: parseBigDecimal(20), color: 'a5a5aa' },
                { id: '3', name: '---', value: parseBigDecimal(20), color: 'c5c5c9' }
            ]"
            :value-type="ChartValueType.Amount"
            :style-type="currentExploration.chartType"
            :skeleton="true"
            :use-custom-color="true"
            v-if="loading"
        />
        <pie-chart
            :items="categoryDimensionTransactionExplorerData && categoryDimensionTransactionExplorerData.length ? categoryDimensionTransactionExplorerData : []"
            :value-type="TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.valueType ?? ChartValueType.Number"
            :style-type="currentExploration.chartType"
            :show-value="true"
            :show-percent="true"
            :enable-click-item="true"
            :default-currency="defaultCurrency"
            v-else-if="!loading"
            @click="onClickPieChartItem"
        />
    </v-card-text>
    <v-card-text :class="{ 'readonly': loading }" v-else-if="currentExploration.chartType === TransactionExplorerChartType.Radar.value">
        <radar-chart
            :items="[
                {
                    name: '---',
                    values: Array.from({ length: 6 }, () => parseBigDecimal(10)),
                    displayOrders: [ 0, 1, 2, 3, 4, 5 ]
                }
            ]"
            :all-category-names="[ '---', '---', '---', '---', '---', '---' ]"
            :value-type="ChartValueType.Amount"
            :skeleton="true"
            :hide-legend="true"
            category-type-name=""
            v-if="loading"
        />
        <radar-chart
            ref="radarChart"
            :category-type-name="currentExploration.seriesDimension === TransactionExplorerDataDimension.None.value ? tt('Name') : currentTransactionExplorerCategoryDimensionName"
            :items="seriesDimensionTransactionExplorerData"
            :all-category-names="categoriedNamesSortedByDisplayOrder"
            :hide-legend="currentExploration.seriesDimension === TransactionExplorerDataDimension.None.value"
            :value-type="TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.valueType ?? ChartValueType.Number"
            :show-value="true"
            :show-percent="true"
            :default-currency="defaultCurrency"
            v-else-if="!loading"
        />
    </v-card-text>
    <v-card-text :class="{ 'readonly': loading }" v-else-if="TransactionExplorerChartType.valueOf(currentExploration.chartType)?.seriesDimensionRequired && axisChartDisplayType">
        <axis-chart
            :skeleton="true"
            :type="axisChartDisplayType"
            :sorting-type="currentExploration.chartSortingType"
            :all-category-names="[]"
            :items="[]"
            :value-type="ChartValueType.Amount"
            category-type-name=""
            v-if="loading"
        />
        <axis-chart
            ref="axisChart"
            :type="axisChartDisplayType"
            :smooth-curve="axisChartSmoothCurve"
            :stacked="axisChartStacked"
            :one-hundred-percent-stacked="axisChart100PercentStacked"
            :sorting-type="currentExploration.chartSortingType"
            :show-value="true"
            :show-total-amount-in-tooltip="TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.supportSum"
            :total-name-in-tooltip="currentExploration.valueMetric === TransactionExplorerValueMetric.TransactionCount.value ? tt('Total Transactions') : tt('Total Amount')"
            :category-type-name="currentTransactionExplorerCategoryDimensionName"
            :all-category-names="categoriedNamesSortedByDisplayOrder"
            :items="seriesDimensionTransactionExplorerData"
            :value-type="TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.valueType ?? ChartValueType.Number"
            :default-currency="defaultCurrency"
            :enable-click-item="true"
            :tooltip-extra-column-names="axisChartTooltipExtraColumnNames"
            :tooltip-extra-column-total-values="axisChartShowYearOverYear || axisChartShowPeriodOverPeriod ? getAxisChartTooltipExtraColumnTotalValues : undefined"
            :tooltip-extra-column-values="axisChartShowYearOverYear || axisChartShowPeriodOverPeriod ? getAxisChartTooltipExtraColumnValues : undefined"
            @click="onClickTrendChartItem"
            v-else-if="!loading"
        />
    </v-card-text>
    <v-card-text :class="{ 'readonly': loading }" v-else-if="TransactionExplorerChartType.valueOf(currentExploration.chartType)?.seriesDimensionRequired && hierarchyChartDisplayType">
        <hierarchy-chart
            :skeleton="true"
            :type="hierarchyChartDisplayType"
            :all-category-names="[]"
            :items="[]"
            :value-type="ChartValueType.Amount"
            category-type-name=""
            v-if="loading"
        />
        <hierarchy-chart
            ref="hierarchyChart"
            :type="hierarchyChartDisplayType"
            :show-value="true"
            :enable-click-item="true"
            :category-type-name="currentTransactionExplorerCategoryDimensionName"
            :all-category-names="categoriedNamesSortedByDisplayOrder"
            :items="seriesDimensionTransactionExplorerData"
            :value-type="TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.valueType ?? ChartValueType.Number"
            :default-currency="defaultCurrency"
            v-else-if="!loading"
            @click="onClickHierarchyChartItem"
        />
    </v-card-text>
    <v-card-text :class="{ 'readonly': loading }" v-else-if="currentExploration.chartType === TransactionExplorerChartType.Heatmap.value">
        <heat-map-chart
            :skeleton="true"
            :all-category-names="[]"
            :items="[]"
            :value-type="ChartValueType.Amount"
            :value-type-name="tt(TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.name ?? 'Value')"
            category-type-name=""
            v-if="loading"
        />
        <heat-map-chart
            ref="heatmapChart"
            :show-value="true"
            :enable-click-item="true"
            :category-type-name="currentTransactionExplorerCategoryDimensionName"
            :all-category-names="categoriedNamesSortedByDisplayOrder"
            :items="seriesDimensionTransactionExplorerData"
            :value-type="TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.valueType ?? ChartValueType.Number"
            :value-type-name="tt(TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.name ?? 'Value')"
            :default-currency="defaultCurrency"
            v-else-if="!loading"
            @click="onClickHeatmapChartItem"
        />
    </v-card-text>
    <v-card-text :class="{ 'readonly': loading }" v-else-if="currentExploration.chartType === TransactionExplorerChartType.CalendarHeatmap.value">
        <calendar-heat-map-chart
            :skeleton="true"
            :items="[]"
            :value-type="ChartValueType.Amount"
            :value-type-name="tt(TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.name ?? 'Value')"
            v-if="loading"
        />
        <calendar-heat-map-chart
            :show-value="true"
            :enable-click-item="true"
            :items="categoryDimensionTransactionExplorerData && categoryDimensionTransactionExplorerData.length ? categoryDimensionTransactionExplorerData : []"
            :value-type="TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.valueType ?? ChartValueType.Number"
            :value-type-name="tt(TransactionExplorerValueMetric.valueOf(currentExploration.valueMetric)?.name ?? 'Value')"
            :default-currency="defaultCurrency"
            v-else-if="!loading"
            @click="onClickCalendarHeatmapChartItem"
        />
    </v-card-text>
    <v-card-text :class="{ 'readonly': loading }" v-else-if="currentExploration.chartType === TransactionExplorerChartType.Custom.value">
        <custom-chart
            ref="customChart"
            :disabled="loading || disabled"
            :display-layout="currentExploration.customChartDisplayLayout"
            :transactions="explorersStore.filteredTransactionsForCustomChart"
            :queries="currentExploration.queries"
            v-model="currentExploration.customChartScript"
        />
    </v-card-text>

    <transaction-list-dialog ref="transactionListDialog" @click:transaction="onClickTransaction" />
</template>
