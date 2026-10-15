<script setup lang="ts">
import { ChartValueType } from '@/core/chart.ts';
import {
    CategoricalChartType,
    ChartDataAggregationType,
    ChartDataType,
    ChartDateAggregationType,
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import { TextDirection } from '@/core/text.ts';
import { getIconType } from '@/lib/icon.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import type { Router } from 'framework7/types';
import { useTransactionPage } from './TransactionPage/useTransactionPage.ts';

const props = defineProps<{
    f7router: Router.Router;
}>();

const {
    tt, formatPercentToLocalizedNumerals, loading, analysisType, trendDateAggregationType, assetTrendsDateAggregationType,
    defaultCurrency, fiscalYearStart, allDateRanges, allSortingTypes, allTrendAnalysisDateAggregationTypes,
    allAssetTrendsDateAggregationTypes, query, queryChartDataCategory, queryDateType, queryStartTime, queryEndTime,
    queryDateRangeName, queryChartDataTypeName, querySortingTypeName, queryTrendDateAggregationTypeName,
    queryAssetTrendsDateAggregationTypeName, isQueryDateRangeChanged, canChangeDateRange, canShiftDateRange,
    canUseCategoryFilter, canUseTagFilter, canUseKeywordFilter, showAmountInChart, totalAmountName,
    showPercentInCategoricalChart, showStackedInTrendsChart, translateNameInTrendsChart, categoricalAnalysisData,
    trendsAnalysisData, assetTrendsData, canShowCustomDateRange, getTransactionCategoricalAnalysisDataItemDisplayColor,
    getDisplayAmount, reloading, showCustomDateRangeSheet, showCustomMonthRangeSheet, showMoreActionSheet, textDirection,
    allChartTypes, queryChartType, getTransactionItemLinkUrl, reload, setChartType, setChartDataType, setSortingType,
    setTrendDateAggregationType, setAssetTrendsDateAggregationType, setDateFilter, setCustomDateFilter, shiftDateRange,
    filterAccounts, filterCategories, filterTags, filterDescription, settings, scrollPopoverToSelectedItem,
    onClickPieChartItem, onClickTrendChartItem, onPageAfterIn,
} = useTransactionPage(props);

</script>

<template>
    <f7-page ptr @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title>
                <f7-link popover-open=".chart-data-type-popover-menu" :class="{ 'disabled': loading }">
                    <span class="statistics-page-title">{{ queryChartDataTypeName }}</span>
                    <f7-icon class="page-title-bar-icon" color="gray" style="opacity: 0.5" f7="chevron_down_circle_fill"></f7-icon>
                </f7-link>
            </f7-nav-title>
            <f7-nav-right :class="{ 'disabled': loading }">
                <f7-link icon-f7="ellipsis" :aria-label="tt('More')" @click="showMoreActionSheet = true"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-popover class="chart-data-type-popover-menu"
                    @popover:open="scrollPopoverToSelectedItem">
            <f7-list dividers>
                <f7-list-item group-title>
                    <small>{{ tt('Categorical Analysis') }}</small>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt(dataType.name)"
                              :class="{ 'list-item-selected': analysisType === StatisticsAnalysisType.CategoricalAnalysis && query.chartDataType === dataType.type }"
                              :key="dataType.type"
                              v-for="dataType in ChartDataType.values(StatisticsAnalysisType.CategoricalAnalysis)"
                              @click="setChartDataType(StatisticsAnalysisType.CategoricalAnalysis, dataType.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="analysisType === StatisticsAnalysisType.CategoricalAnalysis && query.chartDataType === dataType.type"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item group-title>
                    <small>{{ tt('Trend Analysis') }}</small>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt(dataType.name)"
                              :class="{ 'list-item-selected': analysisType === StatisticsAnalysisType.TrendAnalysis && query.chartDataType === dataType.type }"
                              :key="dataType.type"
                              v-for="dataType in ChartDataType.values(StatisticsAnalysisType.TrendAnalysis)"
                              @click="setChartDataType(StatisticsAnalysisType.TrendAnalysis, dataType.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="analysisType === StatisticsAnalysisType.TrendAnalysis && query.chartDataType === dataType.type"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item group-title>
                    <small>{{ tt('Asset Trends') }}</small>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt(dataType.name)"
                              :class="{ 'list-item-selected': analysisType === StatisticsAnalysisType.AssetTrends && query.chartDataType === dataType.type }"
                              :key="dataType.type"
                              v-for="dataType in ChartDataType.values(StatisticsAnalysisType.AssetTrends)"
                              @click="setChartDataType(StatisticsAnalysisType.AssetTrends, dataType.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="analysisType === StatisticsAnalysisType.AssetTrends && query.chartDataType === dataType.type"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <f7-card class="margin-vertical-half" v-if="analysisType === StatisticsAnalysisType.CategoricalAnalysis && query.categoricalChartType === CategoricalChartType.Pie.type">
            <f7-card-header class="no-border display-block">
                <div :class="{ 'statistics-chart-header': true, 'width-100': true, 'text-align-right': textDirection === TextDirection.LTR, 'text-align-left': textDirection === TextDirection.RTL}">
                    <span style="margin-inline-end: 4px;">{{ tt('Sort by') }}</span>
                    <f7-link href="#" popover-open=".sorting-type-popover-menu" :class="{ 'disabled': loading }">{{ querySortingTypeName }}</f7-link>
                </div>
            </f7-card-header>
            <f7-card-content class="pie-chart-container" style="margin-top: -6px" :padding="false">
                <pie-chart
                    :items="[
                        { name: '---', value: parseBigDecimal(60), color: '7c7c7f' },
                        { name: '---', value: parseBigDecimal(20), color: 'a5a5aa' },
                        { name: '---', value: parseBigDecimal(20), color: 'c5c5c9' }
                    ]"
                    :value-type="ChartValueType.Amount"
                    :skeleton="true"
                    :show-center-text="true"
                    :show-selected-item-info="true"
                    :use-custom-color="true"
                    class="statistics-pie-chart"
                    center-text-background="#cccccc"
                    v-if="loading"
                ></pie-chart>
                <pie-chart
                    :items="categoricalAnalysisData.items"
                    :value-type="ChartValueType.Amount"
                    :show-value="showAmountInChart"
                    :show-percent="showPercentInCategoricalChart"
                    :show-center-text="true"
                    :show-selected-item-info="true"
                    :enable-click-item="true"
                    :default-currency="defaultCurrency"
                    class="statistics-pie-chart"
                    v-else-if="!loading"
                    @click="onClickPieChartItem"
                >
                    <text class="statistics-pie-chart-total-amount-title" v-if="categoricalAnalysisData.items && categoricalAnalysisData.items.length">
                        {{ totalAmountName }}
                    </text>
                    <text class="statistics-pie-chart-total-amount-value" v-if="categoricalAnalysisData.items && categoricalAnalysisData.items.length">
                        {{ getDisplayAmount(categoricalAnalysisData.value, defaultCurrency, 16) }}
                    </text>
                    <text class="statistics-pie-chart-total-no-data" cy="50%" v-if="!categoricalAnalysisData.items || !categoricalAnalysisData.items.length">
                        {{ tt('No data') }}
                    </text>
                </pie-chart>
            </f7-card-content>
        </f7-card>

        <f7-card class="margin-vertical-half" v-else-if="analysisType === StatisticsAnalysisType.CategoricalAnalysis && query.categoricalChartType === CategoricalChartType.Bar.type">
            <f7-card-header class="no-border display-block">
                <div class="statistics-chart-header display-flex width-100 justify-content-space-between">
                    <div>
                        {{ totalAmountName }}
                    </div>
                    <div class="align-self-flex-end">
                        <span style="margin-inline-end: 4px;">{{ tt('Sort by') }}</span>
                        <f7-link href="#" popover-open=".sorting-type-popover-menu">{{ querySortingTypeName }}</f7-link>
                    </div>
                </div>
                <div class="display-flex width-100">
                    <div :class="{ 'statistics-list-item-overview-amount': true, 'text-expense': query.chartDataType === ChartDataType.OutflowsByAccount.type || query.chartDataType === ChartDataType.ExpenseByAccount.type || query.chartDataType === ChartDataType.ExpenseByPrimaryCategory.type || query.chartDataType === ChartDataType.ExpenseBySecondaryCategory.type, 'text-income': query.chartDataType === ChartDataType.InflowsByAccount.type || query.chartDataType === ChartDataType.IncomeByAccount.type || query.chartDataType === ChartDataType.IncomeByPrimaryCategory.type || query.chartDataType === ChartDataType.IncomeBySecondaryCategory.type }">
                        <span v-if="!loading && categoricalAnalysisData && categoricalAnalysisData.items && categoricalAnalysisData.items.length">
                            {{ getDisplayAmount(categoricalAnalysisData.value, defaultCurrency) }}
                        </span>
                        <span :class="{ 'skeleton-text': loading }" v-else-if="loading || !categoricalAnalysisData || !categoricalAnalysisData.items || !categoricalAnalysisData.items.length">
                            {{ loading ? '***.**' : '---' }}
                        </span>
                    </div>
                </div>
            </f7-card-header>
            <f7-card-content style="margin-top: -14px" :padding="false">
                <f7-list strong inset dividers class="statistics-list-item skeleton-text" v-if="loading">
                    <f7-list-item class="item-no-divider" link="#" :key="itemIdx" v-for="itemIdx in [ 1, 2, 3 ]">
                        <template #media>
                            <div class="display-flex no-padding-horizontal">
                                <div class="display-flex align-items-center statistics-icon">
                                    <f7-icon f7="app_fill"></f7-icon>
                                </div>
                            </div>
                        </template>
                        <template #title>
                            <div class="statistics-list-item-text">
                                <span>Category Name</span>
                                <small class="statistics-percent">33.33</small>
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

                <f7-list strong inset dividers v-else-if="!loading && (!categoricalAnalysisData || !categoricalAnalysisData.items || !categoricalAnalysisData.items.length)">
                    <f7-list-item :title="tt('No transaction data')"></f7-list-item>
                </f7-list>

                <f7-list strong inset dividers v-else-if="!loading && categoricalAnalysisData && categoricalAnalysisData.items && categoricalAnalysisData.items.length">
                    <f7-list-item class="item-no-divider display-none"></f7-list-item>
                    <template v-for="(item, idx) in categoricalAnalysisData.items">
                        <f7-list-item class="statistics-list-item item-no-divider"
                                      :link="getTransactionItemLinkUrl(item.id)"
                                      :key="idx"
                                      v-if="!item.hidden">
                            <template #media>
                                <div class="display-flex no-padding-horizontal">
                                    <div class="display-flex align-items-center statistics-icon">
                                        <ItemIcon :icon-type="getIconType(queryChartDataCategory, item.iconType)" :icon-id="item.icon" :color="item.color" v-if="item.icon"></ItemIcon>
                                        <f7-icon f7="pencil_ellipsis_rectangle" v-else-if="!item.icon"></f7-icon>
                                    </div>
                                </div>
                            </template>

                            <template #title>
                                <div class="statistics-list-item-text">
                                    <span>{{ item.name }}</span>
                                    <small class="statistics-percent" v-if="showPercentInCategoricalChart && item.percent >= 0 && item.value.isPositiveOrZero()">{{ formatPercentToLocalizedNumerals(item.percent, 2, '<0.01') }}</small>
                                </div>
                            </template>

                            <template #after>
                                <span>{{ getDisplayAmount(item.value, defaultCurrency) }}</span>
                            </template>

                            <template #inner-end>
                                <div class="statistics-item-end">
                                    <div class="statistics-percent-line">
                                        <f7-progressbar :progress="item.percent >= 0 ? item.percent : 0" :style="{ '--f7-progressbar-progress-color': (item.color ? getTransactionCategoricalAnalysisDataItemDisplayColor(item) : '') } "></f7-progressbar>
                                    </div>
                                </div>
                            </template>
                        </f7-list-item>
                    </template>
                </f7-list>
            </f7-card-content>
        </f7-card>

        <f7-card class="margin-vertical-half" v-else-if="analysisType === StatisticsAnalysisType.TrendAnalysis">
            <f7-card-header class="no-border display-block">
                <div class="statistics-chart-header display-flex width-100 justify-content-space-between">
                    <div></div>
                    <div class="align-self-flex-end">
                        <span style="margin-inline-end: 4px;">{{ tt('Sort by') }}</span>
                        <f7-link href="#" popover-open=".sorting-type-popover-menu">{{ querySortingTypeName }}</f7-link>
                    </div>
                </div>
            </f7-card-header>
            <f7-card-content style="margin-top: -14px" :padding="false">
                <trends-bar-chart
                    chart-mode="monthly"
                    :loading="loading || reloading"
                    :start-time="undefined"
                    :end-time="undefined"
                    :start-year-month="query.trendChartStartYearMonth"
                    :end-year-month="query.trendChartEndYearMonth"
                    :sorting-type="query.sortingType"
                    :data-aggregation-type="ChartDataAggregationType.Sum"
                    :date-aggregation-type="trendDateAggregationType"
                    :fiscal-year-start="fiscalYearStart"
                    :items="trendsAnalysisData && trendsAnalysisData.items && trendsAnalysisData.items.length ? trendsAnalysisData.items : []"
                    :value-type="ChartValueType.Amount"
                    :stacked="showStackedInTrendsChart"
                    :translate-name="translateNameInTrendsChart"
                    :default-currency="defaultCurrency"
                    @click="onClickTrendChartItem"
                />
            </f7-card-content>
        </f7-card>

        <f7-card class="margin-vertical-half" v-else-if="analysisType === StatisticsAnalysisType.AssetTrends">
            <f7-card-header class="no-border display-block">
                <div class="statistics-chart-header display-flex width-100 justify-content-space-between">
                    <div></div>
                    <div class="align-self-flex-end">
                        <span style="margin-inline-end: 4px;">{{ tt('Sort by') }}</span>
                        <f7-link href="#" popover-open=".sorting-type-popover-menu">{{ querySortingTypeName }}</f7-link>
                    </div>
                </div>
            </f7-card-header>
            <f7-card-content style="margin-top: -14px" :padding="false">
                <trends-bar-chart
                    chart-mode="daily"
                    :loading="loading || reloading"
                    :start-time="query.assetTrendsChartStartTime"
                    :end-time="query.assetTrendsChartEndTime"
                    :start-year-month="undefined"
                    :end-year-month="undefined"
                    :sorting-type="query.sortingType"
                    :data-aggregation-type="ChartDataAggregationType.Last"
                    :date-aggregation-type="assetTrendsDateAggregationType"
                    :fiscal-year-start="fiscalYearStart"
                    :items="assetTrendsData && assetTrendsData.items && assetTrendsData.items.length ? assetTrendsData.items : []"
                    :value-type="ChartValueType.Amount"
                    :stacked="showStackedInTrendsChart"
                    :translate-name="translateNameInTrendsChart"
                    :default-currency="defaultCurrency"
                    @click="onClickTrendChartItem"
                />
            </f7-card-content>
        </f7-card>

        <f7-popover class="sorting-type-popover-menu">
            <f7-list dividers>
                <f7-list-item link="#" no-chevron popover-close
                              :title="sortingType.displayName"
                              :class="{ 'list-item-selected': query.sortingType === sortingType.type }"
                              :key="sortingType.type"
                              v-for="sortingType in allSortingTypes"
                              @click="setSortingType(sortingType.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.sortingType === sortingType.type"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <f7-toolbar tabbar bottom :class="{ 'compact-tabbar': true, 'toolbar-item-auto-size': true, 'disabled': loading }">
            <f7-link :class="{ 'disabled': reloading || !canShiftDateRange }" :aria-label="tt('Previous Period')" @click="shiftDateRange(-1)">
                <f7-icon class="icon-with-direction" f7="arrow_left_square"></f7-icon>
            </f7-link>
            <f7-link :class="{ 'tabbar-text-with-ellipsis': true, 'disabled': reloading || !canChangeDateRange }" popover-open=".date-popover-menu">
                <span :class="{ 'tabbar-item-changed': isQueryDateRangeChanged }">{{ queryDateRangeName }}</span>
            </f7-link>
            <f7-link :class="{ 'disabled': reloading || !canShiftDateRange }" :aria-label="tt('Next Period')" @click="shiftDateRange(1)">
                <f7-icon class="icon-with-direction" f7="arrow_right_square"></f7-icon>
            </f7-link>
            <f7-link :class="{ 'tabbar-text-with-ellipsis': true, 'disabled': reloading }" popover-open=".date-aggregation-popover-menu"
                     v-if="analysisType === StatisticsAnalysisType.TrendAnalysis">
                <span :class="{ 'tabbar-item-changed': trendDateAggregationType !== ChartDateAggregationType.Default.type }">{{ queryTrendDateAggregationTypeName }}</span>
            </f7-link>
            <f7-link :class="{ 'tabbar-text-with-ellipsis': true, 'disabled': reloading }" popover-open=".date-aggregation-popover-menu"
                     v-if="analysisType === StatisticsAnalysisType.AssetTrends">
                <span :class="{ 'tabbar-item-changed': assetTrendsDateAggregationType !== ChartDateAggregationType.Default.type }">{{ queryAssetTrendsDateAggregationTypeName }}</span>
            </f7-link>
            <f7-link class="tabbar-text-with-ellipsis" :key="chartType.type"
                     v-for="chartType in allChartTypes" @click="setChartType(chartType.type)">
                <span :class="{ 'tabbar-item-changed': queryChartType === chartType.type }">{{ chartType.displayName }}</span>
            </f7-link>
        </f7-toolbar>

        <f7-popover class="date-popover-menu"
                    @popover:open="scrollPopoverToSelectedItem">
            <f7-list dividers>
                <f7-list-item link="#" no-chevron popover-close
                              :title="dateRange.displayName"
                              :class="{ 'list-item-selected': queryDateType === dateRange.type }"
                              :key="dateRange.type"
                              v-for="dateRange in allDateRanges"
                              @click="setDateFilter(dateRange.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="queryDateType === dateRange.type"></f7-icon>
                    </template>
                    <template #footer>
                        <div v-if="dateRange.isUserCustomRange && canShowCustomDateRange(dateRange.type)">
                            <span>{{ queryStartTime }}</span>
                            <span>&nbsp;-&nbsp;</span>
                            <br/>
                            <span>{{ queryEndTime }}</span>
                        </div>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <f7-popover class="date-aggregation-popover-menu"
                    @popover:open="scrollPopoverToSelectedItem">
            <f7-list dividers v-if="analysisType === StatisticsAnalysisType.TrendAnalysis">
                <f7-list-item link="#" no-chevron popover-close
                              :title="aggregationType.displayName"
                              :class="{ 'list-item-selected': trendDateAggregationType === aggregationType.type }"
                              :key="aggregationType.type"
                              v-for="aggregationType in allTrendAnalysisDateAggregationTypes"
                              @click="setTrendDateAggregationType(aggregationType.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="trendDateAggregationType === aggregationType.type"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
            <f7-list dividers v-else-if="analysisType === StatisticsAnalysisType.AssetTrends">
                <f7-list-item link="#" no-chevron popover-close
                              :title="aggregationType.displayName"
                              :class="{ 'list-item-selected': assetTrendsDateAggregationType === aggregationType.type }"
                              :key="aggregationType.type"
                              v-for="aggregationType in allAssetTrendsDateAggregationTypes"
                              @click="setAssetTrendsDateAggregationType(aggregationType.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="assetTrendsDateAggregationType === aggregationType.type"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <date-range-selection-sheet :title="tt('Custom Date Range')"
                                    :min-time="query.categoricalChartStartTime"
                                    :max-time="query.categoricalChartEndTime"
                                    v-model:show="showCustomDateRangeSheet"
                                    @dateRange:change="setCustomDateFilter">
        </date-range-selection-sheet>

        <month-range-selection-sheet :title="tt('Custom Date Range')"
                                     :min-time="query.trendChartStartYearMonth"
                                     :max-time="query.trendChartEndYearMonth"
                                     v-model:show="showCustomMonthRangeSheet"
                                     @dateRange:change="setCustomDateFilter">
        </month-range-selection-sheet>

        <f7-actions close-by-outside-click close-on-escape :opened="showMoreActionSheet" @actions:closed="showMoreActionSheet = false">
            <f7-actions-group>
                <f7-actions-button :class="{ 'disabled': reloading }" @click="filterAccounts">{{ tt('Filter Accounts') }}</f7-actions-button>
                <f7-actions-button :class="{ 'disabled': reloading }" @click="filterCategories" v-if="canUseCategoryFilter">{{ tt('Filter Transaction Categories') }}</f7-actions-button>
                <f7-actions-button :class="{ 'disabled': reloading }" @click="filterTags" v-if="canUseTagFilter">{{ tt('Filter Transaction Tags') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="canUseKeywordFilter">
                <f7-actions-label v-if="query.keyword">{{ query.keyword }}</f7-actions-label>
                <f7-actions-button :class="{ 'disabled': reloading }" @click="filterDescription">{{ tt('Filter transaction description') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button @click="settings">{{ tt('Settings') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>



<style>
.statistics-page-title {
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--f7-text-color);
}

.card-header.no-border:after {
    display: none;
}

.statistics-chart-header {
    font-size: var(--f7-list-item-header-font-size);
}

.statistics-pie-chart .pie-chart-text-group {
    fill: #fff;
    text-anchor: middle;
}

.statistics-pie-chart-total-amount-title {
    -moz-transform: translateY(0.5em);
    -ms-transform: translateY(0.5em);
    -webkit-transform: translateY(0.5em);
    transform: translateY(0.5em);
}

.statistics-pie-chart-total-amount-value {
    -moz-transform: translateY(2em);
    -ms-transform: translateY(2em);
    -webkit-transform: translateY(2em);
    transform: translateY(2em);
}

.statistics-pie-chart-total-no-data {
    -moz-transform: translateY(1.5em);
    -ms-transform: translateY(1.5em);
    -webkit-transform: translateY(1.5em);
    transform: translateY(1.5em);
}

.chart-data-type-popover-menu .popover-inner {
    max-height: 440px;
    overflow-y: auto;
}
</style>
