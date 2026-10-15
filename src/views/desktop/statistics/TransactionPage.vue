<script setup lang="ts">
import StatisticsToolbar from './TransactionPage/StatisticsToolbar.vue';
import DataExportDialog from '@/components/desktop/DataExportDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import TrendsChart from '@/components/desktop/TrendsChart.vue';
import { ChartValueType } from '@/core/chart.ts';
import {
    CategoricalChartType,
    ChartDataAggregationType,
    ChartDataType,
    ChartDateAggregationType,
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import {
    getIconType
} from '@/lib/icon.ts';
import {
    parseBigDecimal
} from '@/lib/numeral.ts';
import AccountFilterSettingsDialog from '@/views/desktop/common/dialogs/AccountFilterSettingsDialog.vue';
import CategoryFilterSettingsDialog from '@/views/desktop/common/dialogs/CategoryFilterSettingsDialog.vue';
import TransactionTagFilterSettingsDialog from '@/views/desktop/common/dialogs/TransactionTagFilterSettingsDialog.vue';
import {

    mdiSquareRounded
} from '@mdi/js';
import { type TransactionStatisticsProps, useTransactionPage } from './TransactionPage/useTransactionPage.ts';

const props = defineProps<TransactionStatisticsProps>();

const {
    tt, formatPercentToLocalizedNumerals, loading, trendDateAggregationType, assetTrendsDateAggregationType, defaultCurrency,
    fiscalYearStart, allDateRanges, allSortingTypes, allTrendAnalysisDateAggregationTypes, allAssetTrendsDateAggregationTypes,
    query, queryChartDataCategory, queryDateType, queryStartTime, queryEndTime, queryDateRangeName,
    queryTrendDateAggregationTypeName, queryAssetTrendsDateAggregationTypeName, canChangeDateRange, canShiftDateRange,
    canUseCategoryFilter, canUseTagFilter, canUseKeywordFilter, showAmountInChart, totalAmountName,
    showPercentInCategoricalChart, showTotalAmountInTrendsChart, showStackedInTrendsChart, translateNameInTrendsChart,
    categoricalOverviewAnalysisData, categoricalAnalysisData, trendsAnalysisData, radarChartCategoryNames, radarChartData,
    assetTrendsData, canShowCustomDateRange, getTransactionCategoricalAnalysisDataItemDisplayColor, getDisplayAmount,
    activeTab, initing, filterKeyword, showCustomDateRangeDialog, showCustomMonthRangeDialog, showFilterAccountDialog,
    showFilterCategoryDialog, showFilterTagDialog, isDarkMode, statisticsDataHasData, allChartTypes, queryAnalysisType,
    queryChartType, queryChartDataType, querySortingType, isQuerySpecialChartType, statisticsTextColor,
    getTransactionItemLinkUrl, reload, setTrendDateAggregationType, setAssetTrendsDateAggregationType, setDateFilter,
    setCustomDateFilter, shiftDateRange, setAccountFilter, setCategoryFilter, setTagFilter, setKeywordFilter, exportResults,
    onClickSankeyChartItem, onClickPieChartItem, onClickTrendChartItem, onShowDateRangeError,
} = useTransactionPage(props);

</script>

<template>
    <main-page-layout>
        <template #nav-items>
            <div class="mb-2">
                <btn-vertical-group :disabled="loading" :buttons="[
                    { name: tt('Categorical Analysis'), value: StatisticsAnalysisType.CategoricalAnalysis },
                    { name: tt('Trend Analysis'), value: StatisticsAnalysisType.TrendAnalysis },
                    { name: tt('Asset Trends'), value: StatisticsAnalysisType.AssetTrends }
                ]" v-model="queryAnalysisType" />
            </div>
            <v-divider class="my-2" />
            <div class="my-2">
                <span class="mx-3 text-body-medium">{{ tt('Chart Type') }}</span>
                <v-select
                    item-title="displayName"
                    item-value="type"
                    class="mt-1"
                    density="compact"
                    :disabled="loading"
                    :items="allChartTypes"
                    v-model="queryChartType"
                    v-show="!isQuerySpecialChartType"
                />
                <v-select
                    item-title="displayName"
                    item-value="type"
                    class="mt-1"
                    density="compact"
                    :disabled="true"
                    :items="[{ displayName: tt('Sankey Chart'), type: 0 }]"
                    :model-value="0"
                    v-show="isQuerySpecialChartType && queryChartDataType === ChartDataType.Overview.type"
                />
            </div>
            <div class="my-2">
                <span class="mx-3 text-body-medium">{{ tt('Sort Order') }}</span>
                <v-select
                    item-title="displayName"
                    item-value="type"
                    class="mt-1"
                    density="compact"
                    :disabled="loading"
                    :items="allSortingTypes"
                    v-model="querySortingType"
                />
            </div>
            <li class="nav-link" :key="dataType.type"
                v-for="dataType in ChartDataType.values(undefined, true)"
                v-show="dataType.isAvailableAnalysisType(queryAnalysisType)">
                <a class="d-flex align-center cursor-pointer my-1"
                   :class="{ 'router-link-active router-link-exact-active': queryChartDataType === dataType.type, 'disabled': loading }"
                   @click="queryChartDataType = dataType.type">
                    <span class="nav-item-title text-truncate">{{ tt(dataType.name) }}</span>
                    <v-tooltip activator="parent" location="right">{{ tt(dataType.name) }}</v-tooltip>
                </a>
            </li>
        </template>

        <template #content>
            <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container" v-model="activeTab">
                <v-window-item value="statisticsPage">
                    <v-card min-height="780">
                        <template #title>
                            <StatisticsToolbar
    :loading="loading"
    :canShiftDateRange="canShiftDateRange"
    :canChangeDateRange="canChangeDateRange"
    :queryDateRangeName="queryDateRangeName"
    :queryDateType="queryDateType"
    :allDateRanges="allDateRanges"
    :canShowCustomDateRange="canShowCustomDateRange"
    :queryStartTime="queryStartTime"
    :queryEndTime="queryEndTime"
    :queryAnalysisType="queryAnalysisType"
    :queryTrendDateAggregationTypeName="queryTrendDateAggregationTypeName"
    :trendDateAggregationType="trendDateAggregationType"
    :allTrendAnalysisDateAggregationTypes="allTrendAnalysisDateAggregationTypes"
    :queryAssetTrendsDateAggregationTypeName="queryAssetTrendsDateAggregationTypeName"
    :assetTrendsDateAggregationType="assetTrendsDateAggregationType"
    :allAssetTrendsDateAggregationTypes="allAssetTrendsDateAggregationTypes"
    :query="query"
    :canUseKeywordFilter="canUseKeywordFilter"
    :canUseCategoryFilter="canUseCategoryFilter"
    :canUseTagFilter="canUseTagFilter"
    :isQuerySpecialChartType="isQuerySpecialChartType"
    :statisticsDataHasData="statisticsDataHasData"
    v-model:filterKeyword="filterKeyword"
    v-model:showFilterAccountDialog="showFilterAccountDialog"
    v-model:showFilterCategoryDialog="showFilterCategoryDialog"
    v-model:showFilterTagDialog="showFilterTagDialog"
    @shiftDateRange="shiftDateRange"
    @setDateFilter="setDateFilter"
    @setTrendDateAggregationType="setTrendDateAggregationType"
    @setAssetTrendsDateAggregationType="setAssetTrendsDateAggregationType"
    @reload="reload"
    @setKeywordFilter="setKeywordFilter"
    @exportResults="exportResults"
/>
                        </template>

                        <v-card-text class="statistics-overview-title pt-0" :class="{ 'disabled': loading }"
                                     v-if="queryAnalysisType === StatisticsAnalysisType.CategoricalAnalysis && isQuerySpecialChartType && queryChartDataType === ChartDataType.Overview.type && (initing || categoricalOverviewAnalysisData && categoricalOverviewAnalysisData.items && categoricalOverviewAnalysisData.items.length)">
                            <span class="statistics-subtitle">{{ tt('Total Income') }}</span>
                            <span class="statistics-overview-amount ms-3 text-income"
                                  v-if="!initing && categoricalOverviewAnalysisData && categoricalOverviewAnalysisData.items && categoricalOverviewAnalysisData.items.length">
                                {{ getDisplayAmount(categoricalOverviewAnalysisData.totalIncome, defaultCurrency) }}
                            </span>
                            <v-skeleton-loader class="skeleton-no-margin ms-3 mb-2"
                                               width="120px" type="text" :loading="true"
                                               v-else-if="initing"></v-skeleton-loader>
                            <span class="statistics-subtitle ms-3">{{ tt('Total Expense') }}</span>
                            <span class="statistics-overview-amount ms-3 text-expense"
                                  v-if="!initing && categoricalOverviewAnalysisData && categoricalOverviewAnalysisData.items && categoricalOverviewAnalysisData.items.length">
                                {{ getDisplayAmount(categoricalOverviewAnalysisData.totalExpense, defaultCurrency) }}
                            </span>
                            <v-skeleton-loader class="skeleton-no-margin ms-3 mb-2"
                                               width="120px" type="text" :loading="true"
                                               v-else-if="initing"></v-skeleton-loader>
                        </v-card-text>

                        <v-card-text class="statistics-overview-title pt-0" :class="{ 'disabled': loading }"
                                     v-else-if="queryAnalysisType === StatisticsAnalysisType.CategoricalAnalysis && !isQuerySpecialChartType && (initing || (categoricalAnalysisData && categoricalAnalysisData.items && categoricalAnalysisData.items.length))">
                            <span class="statistics-subtitle">{{ totalAmountName }}</span>
                            <span class="statistics-overview-amount ms-3"
                                  :class="statisticsTextColor"
                                  v-if="!initing && categoricalAnalysisData && categoricalAnalysisData.items && categoricalAnalysisData.items.length">
                                {{ getDisplayAmount(categoricalAnalysisData.value, defaultCurrency) }}
                            </span>
                            <v-skeleton-loader class="skeleton-no-margin ms-3 mb-2"
                                               width="120px" type="text" :loading="true"
                                               v-else-if="initing"></v-skeleton-loader>
                        </v-card-text>

                        <v-card-text class="statistics-overview-title pt-0"
                                     v-else-if="!loading && (
                                         (queryAnalysisType === StatisticsAnalysisType.CategoricalAnalysis && isQuerySpecialChartType && queryChartDataType === ChartDataType.Overview.type && (!categoricalOverviewAnalysisData || !categoricalOverviewAnalysisData.items || !categoricalOverviewAnalysisData.items.length))
                                      || (queryAnalysisType === StatisticsAnalysisType.CategoricalAnalysis && !isQuerySpecialChartType && (!categoricalAnalysisData || !categoricalAnalysisData.items || !categoricalAnalysisData.items.length))
                                      || (queryAnalysisType === StatisticsAnalysisType.TrendAnalysis && (!trendsAnalysisData || !trendsAnalysisData.items || !trendsAnalysisData.items.length))
                                      || (queryAnalysisType === StatisticsAnalysisType.AssetTrends && (!assetTrendsData || !assetTrendsData.items || !assetTrendsData.items.length))
                                      )">
                            <span class="statistics-subtitle statistics-overview-empty-tip">{{ tt('No transaction data') }}</span>
                        </v-card-text>

                        <v-card-text class="py-0" :class="{ 'readonly': loading }" v-if="queryAnalysisType === StatisticsAnalysisType.CategoricalAnalysis && queryChartDataType === ChartDataType.Overview.type">
                            <account-and-category-sankey-chart
                                :items="[]"
                                :sorting-type="querySortingType"
                                :skeleton="true"
                                v-if="initing"
                            />
                            <account-and-category-sankey-chart
                                :items="categoricalOverviewAnalysisData && categoricalOverviewAnalysisData.items && categoricalOverviewAnalysisData.items.length ? categoricalOverviewAnalysisData.items : []"
                                :enable-click-item="true"
                                :default-currency="defaultCurrency"
                                v-else-if="!initing"
                                @click="onClickSankeyChartItem"
                            />
                        </v-card-text>

                        <v-card-text class="py-0" :class="{ 'readonly': loading }" v-if="queryAnalysisType === StatisticsAnalysisType.CategoricalAnalysis && !isQuerySpecialChartType && query.categoricalChartType === CategoricalChartType.Pie.type">
                            <pie-chart
                                :items="[
                                    { id: '1', name: '---', value: parseBigDecimal(60), color: '7c7c7f' },
                                    { id: '2', name: '---', value: parseBigDecimal(20), color: 'a5a5aa' },
                                    { id: '3', name: '---', value: parseBigDecimal(20), color: 'c5c5c9' }
                                ]"
                                :value-type="ChartValueType.Amount"
                                :skeleton="true"
                                :use-custom-color="true"
                                v-if="initing"
                            />
                            <pie-chart
                                :items="categoricalAnalysisData && categoricalAnalysisData.items && categoricalAnalysisData.items.length ? categoricalAnalysisData.items : []"
                                :value-type="ChartValueType.Amount"
                                :show-value="showAmountInChart"
                                :show-percent="showPercentInCategoricalChart"
                                :enable-click-item="true"
                                :default-currency="defaultCurrency"
                                v-else-if="!initing"
                                @click="onClickPieChartItem"
                            />
                        </v-card-text>

                        <v-card-text class="py-0" :class="{ 'readonly': loading }" v-if="queryAnalysisType === StatisticsAnalysisType.CategoricalAnalysis && !isQuerySpecialChartType && query.categoricalChartType === CategoricalChartType.Bar.type">
                            <v-list rounded lines="two" v-if="initing">
                                <template :key="itemIdx" v-for="itemIdx in [ 1, 2, 3 ]">
                                    <v-list-item class="ps-0">
                                        <template #prepend>
                                            <div>
                                                <v-icon class="disabled me-0" size="34" :icon="mdiSquareRounded" />
                                            </div>
                                        </template>
                                        <div class="d-flex flex-column ms-2">
                                            <div class="d-flex">
                                                <v-skeleton-loader class="skeleton-no-margin my-2"
                                                                   width="120px" type="text" :loading="true"></v-skeleton-loader>
                                            </div>
                                            <div>
                                                <v-progress-linear :model-value="0" :height="4"></v-progress-linear>
                                            </div>
                                        </div>
                                    </v-list-item>
                                    <v-divider v-if="itemIdx < 3"/>
                                </template>
                            </v-list>
                            <v-list class="py-0" rounded lines="two" v-else-if="!initing && categoricalAnalysisData && categoricalAnalysisData.items && categoricalAnalysisData.items.length">
                                <template :key="idx"
                                          v-for="(item, idx) in categoricalAnalysisData.items">
                                    <v-list-item class="ps-0" density="comfortable" v-if="!item.hidden">
                                        <template #prepend>
                                            <router-link class="statistics-list-item" :to="getTransactionItemLinkUrl(item.id)">
                                                <ItemIcon size="34px" :icon-type="getIconType(queryChartDataCategory, item.iconType)"
                                                          :icon-id="item.icon" :color="item.color"></ItemIcon>
                                            </router-link>
                                        </template>
                                        <router-link class="statistics-list-item link-no-color" :to="getTransactionItemLinkUrl(item.id)">
                                            <div class="d-flex flex-column ms-2">
                                                <div class="d-flex">
                                                    <span>{{ item.name }}</span>
                                                    <small class="statistics-percent" v-if="showPercentInCategoricalChart && item.percent >= 0 && item.value.isPositiveOrZero()">{{ formatPercentToLocalizedNumerals(item.percent, 2, '<0.01') }}</small>
                                                    <v-spacer/>
                                                    <span class="statistics-amount">{{ getDisplayAmount(item.value, defaultCurrency) }}</span>
                                                </div>
                                                <div class="mt-1">
                                                    <v-progress-linear :color="item.color ? getTransactionCategoricalAnalysisDataItemDisplayColor(item) : 'primary'"
                                                                       :bg-color="isDarkMode ? '#444444' : '#f8f8f8'" :bg-opacity="1"
                                                                       :model-value="item.percent >= 0 ? item.percent : 0"
                                                                       :height="4"></v-progress-linear>
                                                </div>
                                            </div>
                                        </router-link>
                                    </v-list-item>
                                    <v-divider v-if="!item.hidden && idx !== categoricalAnalysisData.items.length - 1"/>
                                </template>
                            </v-list>
                        </v-card-text>

                        <v-card-text class="py-0" :class="{ 'readonly': loading }" v-if="queryAnalysisType === StatisticsAnalysisType.CategoricalAnalysis && !isQuerySpecialChartType && query.categoricalChartType === CategoricalChartType.Radar.type">
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
                                v-if="initing"
                            />
                            <radar-chart
                                :items="radarChartData"
                                :all-category-names="radarChartCategoryNames"
                                :value-type="ChartValueType.Amount"
                                :show-value="showAmountInChart"
                                :show-percent="showPercentInCategoricalChart"
                                :default-currency="defaultCurrency"
                                :category-type-name="tt('Name')"
                                :hide-legend="true"
                                v-else-if="!initing"
                            />
                        </v-card-text>

                        <v-card-text class="py-0 mt-2" :class="{ 'readonly': loading }" v-if="queryAnalysisType === StatisticsAnalysisType.TrendAnalysis">
                            <trends-chart
                                chart-mode="monthly"
                                :type="queryChartType"
                                :start-time="undefined"
                                :end-time="undefined"
                                :start-year-month="query.trendChartStartYearMonth"
                                :end-year-month="query.trendChartEndYearMonth"
                                :sorting-type="querySortingType"
                                :data-aggregation-type="ChartDataAggregationType.Sum"
                                :date-aggregation-type="trendDateAggregationType"
                                :fiscal-year-start="fiscalYearStart"
                                :items="[]"
                                :value-type="ChartValueType.Amount"
                                :skeleton="true"
                                v-if="initing"
                            />
                            <trends-chart
                                chart-mode="monthly"
                                :type="queryChartType"
                                :start-time="undefined"
                                :end-time="undefined"
                                :start-year-month="query.trendChartStartYearMonth"
                                :end-year-month="query.trendChartEndYearMonth"
                                :sorting-type="querySortingType"
                                :data-aggregation-type="ChartDataAggregationType.Sum"
                                :date-aggregation-type="trendDateAggregationType"
                                :fiscal-year-start="fiscalYearStart"
                                :items="trendsAnalysisData && trendsAnalysisData.items && trendsAnalysisData.items.length ? trendsAnalysisData.items : []"
                                :value-type="ChartValueType.Amount"
                                :translate-name="translateNameInTrendsChart"
                                :show-value="showAmountInChart"
                                :enable-click-item="true"
                                :default-currency="defaultCurrency"
                                :stacked="showStackedInTrendsChart"
                                :show-total-amount-in-tooltip="showTotalAmountInTrendsChart"
                                :show-year-over-year="true"
                                :show-period-over-period="trendDateAggregationType === ChartDateAggregationType.Month.type || trendDateAggregationType === ChartDateAggregationType.Quarter.type"
                                ref="monthlyTrendsChart"
                                v-else-if="!initing && trendsAnalysisData && trendsAnalysisData.items && trendsAnalysisData.items.length"
                                @click="onClickTrendChartItem"
                            />
                        </v-card-text>

                        <v-card-text class="py-0 mt-2" :class="{ 'readonly': loading }" v-if="queryAnalysisType === StatisticsAnalysisType.AssetTrends">
                            <trends-chart
                                chart-mode="daily"
                                :type="queryChartType"
                                :start-time="query.assetTrendsChartStartTime"
                                :end-time="query.assetTrendsChartEndTime"
                                :start-year-month="undefined"
                                :end-year-month="undefined"
                                :sorting-type="querySortingType"
                                :data-aggregation-type="ChartDataAggregationType.Last"
                                :date-aggregation-type="assetTrendsDateAggregationType"
                                :fiscal-year-start="fiscalYearStart"
                                :items="[]"
                                :value-type="ChartValueType.Amount"
                                :skeleton="true"
                                v-if="initing"
                            />
                            <trends-chart
                                chart-mode="daily"
                                :type="queryChartType"
                                :start-time="query.assetTrendsChartStartTime"
                                :end-time="query.assetTrendsChartEndTime"
                                :start-year-month="undefined"
                                :end-year-month="undefined"
                                :sorting-type="querySortingType"
                                :data-aggregation-type="ChartDataAggregationType.Last"
                                :date-aggregation-type="assetTrendsDateAggregationType"
                                :fiscal-year-start="fiscalYearStart"
                                :items="assetTrendsData && assetTrendsData.items && assetTrendsData.items.length ? assetTrendsData.items : []"
                                :value-type="ChartValueType.Amount"
                                :translate-name="translateNameInTrendsChart"
                                :show-value="showAmountInChart"
                                :enable-click-item="true"
                                :default-currency="defaultCurrency"
                                :stacked="showStackedInTrendsChart"
                                :show-total-amount-in-tooltip="showTotalAmountInTrendsChart"
                                :show-year-over-year="true"
                                :show-period-over-period="assetTrendsDateAggregationType === ChartDateAggregationType.Day.type || assetTrendsDateAggregationType === ChartDateAggregationType.Month.type || assetTrendsDateAggregationType === ChartDateAggregationType.Quarter.type"
                                ref="dailyTrendsChart"
                                v-else-if="!initing && assetTrendsData && assetTrendsData.items && assetTrendsData.items.length"
                                @click="onClickTrendChartItem"
                            />
                        </v-card-text>
                    </v-card>
                </v-window-item>
            </v-window>
        </template>
    </main-page-layout>

    <date-range-selection-dialog :title="tt('Custom Date Range')"
                                  :min-time="query.categoricalChartStartTime"
                                  :max-time="query.categoricalChartEndTime"
                                  v-model:show="showCustomDateRangeDialog"
                                  @dateRange:change="setCustomDateFilter"
                                  @error="onShowDateRangeError" />

    <month-range-selection-dialog :title="tt('Custom Date Range')"
                                  :min-time="query.trendChartStartYearMonth"
                                  :max-time="query.trendChartEndYearMonth"
                                  v-model:show="showCustomMonthRangeDialog"
                                  @dateRange:change="setCustomDateFilter"
                                  @error="onShowDateRangeError" />

    <account-filter-settings-dialog type="statisticsCurrent"
                                    v-model:show="showFilterAccountDialog"
                                    @settings:change="setAccountFilter" />


    <category-filter-settings-dialog type="statisticsCurrent"
                                     v-model:show="showFilterCategoryDialog"
                                     @settings:change="setCategoryFilter" />

    <transaction-tag-filter-settings-dialog type="statisticsCurrent"
                                            v-model:show="showFilterTagDialog"
                                            @settings:change="setTagFilter" />

    <data-export-dialog ref="dataExportDialog" />

    <snack-bar ref="snackbar" />
</template>



<style>
.statistics-custom-datetime-range {
    line-height: 1rem;
}

.statistics-overview-title {
    line-height: 2rem !important;
    height: 46px;
    display: flex;
    align-items: flex-end;
}

.statistics-overview-amount {
    font-size: 1.5rem;
    overflow: hidden;
    text-overflow: ellipsis;
}

.statistics-subtitle {
    font-size: 1rem;
    line-height: 1.75rem
}

.statistics-overview-empty-tip {
    color: rgba(var(--v-theme-on-background), var(--v-medium-emphasis-opacity)) !important;
}

.statistics-list-item {
    color: var(--v-theme-on-default);
    font-size: 1rem !important;
    line-height: 1.75rem;
    overflow: hidden;
    text-overflow: ellipsis;
}

.statistics-list-item .statistics-percent {
    font-size: 0.75rem;
    opacity: 0.7;
    margin-inline-start: 6px;
}

.statistics-list-item .statistics-amount {
    opacity: 0.8;
}
</style>
