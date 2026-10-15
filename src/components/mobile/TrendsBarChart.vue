<script setup lang="ts">
import {
    type TrendsBarChartClickEvent,
    type TrendsChartDateType
} from '@/components/base/TrendsChartBase.ts';
import { type MobileTrendsChartProps, useTrendsBarChart } from './TrendsBarChart/useTrendsBarChart.ts';

const props = defineProps<MobileTrendsChartProps<TrendsChartDateType>>();

const emit = defineEmits<{
    (e: 'click', value: TrendsBarChartClickEvent): void;
}>();

const {
    tt, formatAmountToLocalizedNumeralsWithCurrency, allDisplayDataItemsVersion, unselectedLegends, virtualDataItems,
    useVirtualList, allDisplayDataItems, clickItem, toggleLegend, renderExternal,
} = useTrendsBarChart(props, emit);

</script>

<template>
    <f7-list strong inset dividers class="skeleton-text" v-if="loading">
        <f7-list-item class="statistics-list-item item-no-divider" link="#" :key="itemIdx" v-for="itemIdx in [ 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12 ]">
            <template #media>
                <div class="display-flex no-padding-horizontal">
                    <div class="display-flex align-items-center statistics-icon">
                        <f7-icon f7="app_fill"></f7-icon>
                    </div>
                </div>
            </template>
            <template #title>
                <div class="statistics-list-item-text">
                    <span>Date Range</span>
                </div>
            </template>
            <template #after>
                <span>0.00 USD</span>
            </template>
            <template #inner-end>
                <div class="statistics-item-end">
                    <div class="statistics-percent-line">
                        <f7-progressbar></f7-progressbar>
                    </div>
                </div>
            </template>
        </f7-list-item>
    </f7-list>

    <f7-list strong inset dividers v-else-if="!loading && (!allDisplayDataItems || !allDisplayDataItems.data || !allDisplayDataItems.data.length)">
        <f7-list-item :title="tt('No transaction data')"></f7-list-item>
    </f7-list>

    <f7-list strong inset dividers v-if="!loading && allDisplayDataItems && allDisplayDataItems.data && allDisplayDataItems.data.length">
        <f7-list-item class="item-no-divider display-none"></f7-list-item>
        <f7-list-item class="item-no-divider" v-if="allDisplayDataItems.legends && allDisplayDataItems.legends.length > 1">
            <div class="display-flex" style="flex-wrap: wrap">
                <div class="trends-bar-chart-legend display-flex align-items-center"
                     :class="{ 'trends-bar-chart-legend-unselected': !!unselectedLegends[legend.id] }"
                     :key="idx"
                     v-for="(legend, idx) in allDisplayDataItems.legends"
                     @click="toggleLegend(legend)">
                    <f7-icon f7="app_fill" class="trends-bar-chart-legend-icon" :style="{ 'color': unselectedLegends[legend.id] ? '' : legend.color }"></f7-icon>
                    <span class="trends-bar-chart-legend-text">{{ legend.name }}</span>
                </div>
            </div>
        </f7-list-item>
    </f7-list>

    <f7-list strong inset dividers :key="`trends-bar-chart-${allDisplayDataItemsVersion}`"
             :virtual-list="useVirtualList"
             :virtual-list-params="useVirtualList ? { items: allDisplayDataItems.data, renderExternal, height: 'auto' } : undefined"
             v-if="!loading && allDisplayDataItems && allDisplayDataItems.data && allDisplayDataItems.data.length">
        <f7-list-item class="item-no-divider display-none"></f7-list-item>
        <f7-list-item class="item-no-divider" link="#"
                      :key="item.index"
                      :class="{ 'statistics-list-item': true, 'statistics-list-item-stacked': stacked, 'statistics-list-item-non-stacked': !stacked }"
                      :style="useVirtualList ? `top: ${virtualDataItems.topPosition}px` : undefined"
                      :virtual-list-index="item.index"
                      v-for="item in (useVirtualList ? virtualDataItems.items : allDisplayDataItems.data)"
                      @click="clickItem(item)"
        >
            <template #media>
                <div class="display-flex no-padding-horizontal">
                    <div class="display-flex align-items-center statistics-icon">
                        <f7-icon f7="calendar"></f7-icon>
                    </div>
                </div>
            </template>

            <template #title>
                <div class="statistics-list-item-text">
                    <span>{{ item.displayDateRange }}</span>
                </div>
                <div class="width-100 statistics-percent-line statistics-multi-percent-line display-flex flex-direction-column" v-if="!stacked && item.items.length > 1">
                    <div class="display-flex flex-direction-column"
                         style="margin-top: 4px"
                         :key="dataIdx"
                         v-for="(data, dataIdx) in item.items"
                         v-show="data.value.isPositive()">
                        <div class="width-100 display-flex flex-direction-row">
                            <div class="display-inline-flex" :style="{ 'width': data.value.divide(item.maxAmount).multiply(item.percent).toDoubleNumber() + '%' }">
                                <f7-progressbar :progress="100" :style="{ '--f7-progressbar-progress-color': (data.color ? data.color : '') } "></f7-progressbar>
                            </div>
                            <div class="display-inline-flex" :style="{ 'width': (100.0 - data.value.divide(item.maxAmount).multiply(item.percent).toDoubleNumber()) + '%' }"
                                 v-if="data.value.divide(item.maxAmount).multiply(item.percent).lessThan(100)">
                                <f7-progressbar :progress="0"></f7-progressbar>
                            </div>
                        </div>
                    </div>
                </div>
            </template>

            <template #after>
                <span v-if="stacked || item.items.length <= 1">{{ formatAmountToLocalizedNumeralsWithCurrency(item.totalAmount, defaultCurrency) }}</span>
            </template>

            <template #inner-end>
                <div class="statistics-item-end" v-if="stacked || item.items.length <= 1">
                    <div class="statistics-percent-line statistics-multi-percent-line display-flex">
                        <div class="display-inline-flex" :style="{ 'width': (data.value.divide(item.totalPositiveAmount).multiply(item.percent).toDoubleNumber()) + '%' }"
                             :key="dataIdx"
                             v-for="(data, dataIdx) in item.items"
                             v-show="data.value.isPositive()">
                            <f7-progressbar :progress="100" :style="{ '--f7-progressbar-progress-color': (data.color ? data.color : '') } "></f7-progressbar>
                        </div>
                        <div class="display-inline-flex" :style="{ 'width': (100.0 - item.percent) + '%' }"
                             v-if="item.percent < 100.0">
                            <f7-progressbar :progress="0"></f7-progressbar>
                        </div>
                    </div>
                </div>
            </template>
        </f7-list-item>
    </f7-list>
</template>



<style>
.trends-bar-chart-legend {
    margin-inline-end: 4px;
    cursor: pointer;
}

.trends-bar-chart-legend-icon.f7-icons {
    font-size: var(--ebk-trends-bar-chart-legend-icon-font-size);
    margin-inline-end: 2px;
}

.trends-bar-chart-legend-unselected .trends-bar-chart-legend-icon.f7-icons {
    color: #cccccc;
}

.trends-bar-chart-legend-text {
    font-size: var(--ebk-trends-bar-chart-legend-text-font-size);
}

.trends-bar-chart-legend-unselected .trends-bar-chart-legend-text {
    color: #cccccc;
}
</style>
