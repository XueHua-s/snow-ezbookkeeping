<script setup lang="ts">
import {
    StatisticsAnalysisType
} from '@/core/statistics.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { useTransactionPage } from '@/views/desktop/statistics/TransactionPage/useTransactionPage.ts';
import {
    mdiArrowLeft,
    mdiArrowRight,
    mdiCalendarRangeOutline,
    mdiCheck,
    mdiDotsVertical,
    mdiExport,
    mdiFilterCogOutline,
    mdiFilterOutline,
    mdiMagnify,
    mdiRefresh
} from '@mdi/js';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useTransactionPage>>;

defineProps<{
    loading: Scope['loading'];
    canShiftDateRange: Scope['canShiftDateRange'];
    canChangeDateRange: Scope['canChangeDateRange'];
    queryDateRangeName: Scope['queryDateRangeName'];
    queryDateType: Scope['queryDateType'];
    allDateRanges: Scope['allDateRanges'];
    canShowCustomDateRange: Scope['canShowCustomDateRange'];
    queryStartTime: Scope['queryStartTime'];
    queryEndTime: Scope['queryEndTime'];
    queryAnalysisType: Scope['queryAnalysisType'];
    queryTrendDateAggregationTypeName: Scope['queryTrendDateAggregationTypeName'];
    trendDateAggregationType: Scope['trendDateAggregationType'];
    allTrendAnalysisDateAggregationTypes: Scope['allTrendAnalysisDateAggregationTypes'];
    queryAssetTrendsDateAggregationTypeName: Scope['queryAssetTrendsDateAggregationTypeName'];
    assetTrendsDateAggregationType: Scope['assetTrendsDateAggregationType'];
    allAssetTrendsDateAggregationTypes: Scope['allAssetTrendsDateAggregationTypes'];
    query: Scope['query'];
    canUseKeywordFilter: Scope['canUseKeywordFilter'];
    canUseCategoryFilter: Scope['canUseCategoryFilter'];
    canUseTagFilter: Scope['canUseTagFilter'];
    isQuerySpecialChartType: Scope['isQuerySpecialChartType'];
    statisticsDataHasData: Scope['statisticsDataHasData'];
}>();

const filterKeyword = defineModel<Scope['filterKeyword']>('filterKeyword', { required: true });
const showFilterAccountDialog = defineModel<Scope['showFilterAccountDialog']>('showFilterAccountDialog', { required: true });
const showFilterCategoryDialog = defineModel<Scope['showFilterCategoryDialog']>('showFilterCategoryDialog', { required: true });
const showFilterTagDialog = defineModel<Scope['showFilterTagDialog']>('showFilterTagDialog', { required: true });

const emit = defineEmits<{
    shiftDateRange: Parameters<Scope['shiftDateRange']>;
    setDateFilter: Parameters<Scope['setDateFilter']>;
    setTrendDateAggregationType: Parameters<Scope['setTrendDateAggregationType']>;
    setAssetTrendsDateAggregationType: Parameters<Scope['setAssetTrendsDateAggregationType']>;
    reload: Parameters<Scope['reload']>;
    setKeywordFilter: Parameters<Scope['setKeywordFilter']>;
    exportResults: Parameters<Scope['exportResults']>;
}>();

const shiftDateRange = (...args: Parameters<Scope['shiftDateRange']>) => emit('shiftDateRange', ...args);
const setDateFilter = (...args: Parameters<Scope['setDateFilter']>) => emit('setDateFilter', ...args);
const setTrendDateAggregationType = (...args: Parameters<Scope['setTrendDateAggregationType']>) => emit('setTrendDateAggregationType', ...args);
const setAssetTrendsDateAggregationType = (...args: Parameters<Scope['setAssetTrendsDateAggregationType']>) => emit('setAssetTrendsDateAggregationType', ...args);
const reload = (...args: Parameters<Scope['reload']>) => emit('reload', ...args);
const setKeywordFilter = (...args: Parameters<Scope['setKeywordFilter']>) => emit('setKeywordFilter', ...args);
const exportResults = (...args: Parameters<Scope['exportResults']>) => emit('exportResults', ...args);

const { tt } = useI18n();

</script>

<template>
    <div class="title-and-toolbar d-flex align-center">
        <span>{{ tt('Statistics & Analysis') }}</span>
        <v-btn-group class="ms-4" color="default" density="comfortable" variant="outlined" divided>
            <v-btn class="button-icon-with-direction" :icon="mdiArrowLeft"
                   :aria-label="tt('Previous Period')" :disabled="loading || !canShiftDateRange"
                   @click="shiftDateRange(-1)"/>
            <v-menu location="bottom" max-height="500">
                <template #activator="{ props }">
                    <v-btn :disabled="loading || !canChangeDateRange"
                           v-bind="props">{{ queryDateRangeName }}</v-btn>
                </template>
                <v-list :selected="[queryDateType]">
                    <v-list-item :key="dateRange.type" :value="dateRange.type"
                                 :append-icon="(queryDateType === dateRange.type ? mdiCheck : undefined)"
                                 v-for="dateRange in allDateRanges">
                        <v-list-item-title class="cursor-pointer"
                                           @click="setDateFilter(dateRange.type)">
                            <div class="d-flex align-center">
                                <span>{{ dateRange.displayName }}</span>
                            </div>
                            <div class="statistics-custom-datetime-range smaller" v-if="dateRange.isUserCustomRange && canShowCustomDateRange(dateRange.type)">
                                <span>{{ queryStartTime }}</span>
                                <span>&nbsp;-&nbsp;</span>
                                <br/>
                                <span>{{ queryEndTime }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>
                </v-list>
            </v-menu>
            <v-btn class="button-icon-with-direction" :icon="mdiArrowRight"
                   :aria-label="tt('Next Period')" :disabled="loading || !canShiftDateRange"
                   @click="shiftDateRange(1)"/>
        </v-btn-group>

        <v-menu location="bottom" max-height="500" v-if="queryAnalysisType === StatisticsAnalysisType.TrendAnalysis">
            <template #activator="{ props }">
                <v-btn class="ms-3" color="default" variant="outlined"
                       :prepend-icon="mdiCalendarRangeOutline" :disabled="loading"
                       v-bind="props">{{ queryTrendDateAggregationTypeName }}</v-btn>
            </template>
            <v-list>
                <v-list-item class="cursor-pointer" :key="aggregationType.type" :value="aggregationType.type"
                             :append-icon="(trendDateAggregationType === aggregationType.type ? mdiCheck : undefined)"
                             :title="aggregationType.displayName"
                             v-for="aggregationType in allTrendAnalysisDateAggregationTypes"
                             @click="setTrendDateAggregationType(aggregationType.type)">
                </v-list-item>
            </v-list>
        </v-menu>

        <v-menu location="bottom" max-height="500" v-if="queryAnalysisType === StatisticsAnalysisType.AssetTrends">
            <template #activator="{ props }">
                <v-btn class="ms-3" color="default" variant="outlined"
                       :prepend-icon="mdiCalendarRangeOutline" :disabled="loading"
                       v-bind="props">{{ queryAssetTrendsDateAggregationTypeName }}</v-btn>
            </template>
            <v-list>
                <v-list-item class="cursor-pointer" :key="aggregationType.type" :value="aggregationType.type"
                             :append-icon="(assetTrendsDateAggregationType === aggregationType.type ? mdiCheck : undefined)"
                             :title="aggregationType.displayName"
                             v-for="aggregationType in allAssetTrendsDateAggregationTypes"
                             @click="setAssetTrendsDateAggregationType(aggregationType.type)">
                </v-list-item>
            </v-list>
        </v-menu>

        <v-btn density="compact" color="default" variant="text" class="ms-2"
               :aria-label="tt('Refresh')" :icon="true" :loading="loading" @click="reload(true)">
            <template #loader>
                <v-progress-circular indeterminate size="20"/>
            </template>
            <v-icon :icon="mdiRefresh" size="24" />
            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
        </v-btn>
        <v-spacer/>
        <div class="transaction-keyword-filter ms-2">
            <v-text-field autocomplete="off" density="compact" :disabled="loading"
                          :prepend-inner-icon="mdiMagnify"
                          :append-inner-icon="filterKeyword !== query.keyword ? mdiCheck : undefined"
                          :placeholder="tt('Filter transaction description')"
                          v-model="filterKeyword"
                          v-if="canUseKeywordFilter"
                          @click:append-inner="setKeywordFilter(filterKeyword)"
                          @keyup.enter="setKeywordFilter(filterKeyword)"
            />
        </div>
        <v-btn density="comfortable" color="default" variant="text" class="ms-2"
               :aria-label="tt('More')" :disabled="loading" :icon="true">
            <v-icon :icon="mdiDotsVertical" />
            <v-menu activator="parent">
                <v-list>
                    <v-list-item :disabled="loading"
                                 :prepend-icon="mdiFilterOutline"
                                 :title="tt('Filter Accounts')"
                                 @click="showFilterAccountDialog = true"></v-list-item>
                    <v-list-item :disabled="loading"
                                 :prepend-icon="mdiFilterOutline"
                                 :title="tt('Filter Transaction Categories')"
                                 @click="showFilterCategoryDialog = true"
                                 v-if="canUseCategoryFilter"></v-list-item>
                    <v-list-item :disabled="loading"
                                 :prepend-icon="mdiFilterOutline"
                                 :title="tt('Filter Transaction Tags')"
                                 @click="showFilterTagDialog = true"
                                 v-if="canUseTagFilter"></v-list-item>
                    <v-divider class="my-2" v-if="!isQuerySpecialChartType" />
                    <v-list-item :prepend-icon="mdiExport"
                                 :title="tt('Export Results')"
                                 :disabled="!statisticsDataHasData"
                                 @click="exportResults"
                                 v-if="!isQuerySpecialChartType"></v-list-item>
                    <v-divider class="my-2"/>
                    <v-list-item to="/settings/statistics"
                                 :prepend-icon="mdiFilterCogOutline"
                                 :title="tt('Settings')"></v-list-item>
                </v-list>
            </v-menu>
        </v-btn>
    </div>
</template>
