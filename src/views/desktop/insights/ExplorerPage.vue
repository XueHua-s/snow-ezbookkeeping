<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import DataExportDialog from '@/components/desktop/DataExportDialog.vue';
import JsonExportDialog from '@/components/desktop/JsonExportDialog.vue';
import JsonImportDialog from '@/components/desktop/JsonImportDialog.vue';
import RenameDialog from '@/components/desktop/RenameDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { TransactionExplorerChartTypeValue } from '@/core/explorer.ts';
import { isInsightsExplorerCodingAssistantEnabled } from '@/lib/server_settings.ts';
import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';
import ExplorerChangeDisplayOrderDialog from '@/views/desktop/insights/dialogs/ExplorerChangeDisplayOrderDialog.vue';
import ExplorerChartTab from '@/views/desktop/insights/tabs/ExplorerChartTab.vue';
import ExplorerDataTableTab from '@/views/desktop/insights/tabs/ExplorerDataTableTab.vue';
import ExplorerEditableDataTableTab from '@/views/desktop/insights/tabs/ExplorerEditableDataTableTab.vue';
import ExplorerQueryTab from '@/views/desktop/insights/tabs/ExplorerQueryTab.vue';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';
import {
    mdiApplicationExport,
    mdiApplicationImport,
    mdiArrowLeft,
    mdiArrowRight,
    mdiCheck,
    mdiDeleteOutline,
    mdiDotsVertical,
    mdiExport,
    mdiEyeOffOutline,
    mdiEyeOutline,
    mdiMagicStaff,
    mdiPencilOutline,
    mdiRefresh,
    mdiSort,
    mdiTableCheck,
    mdiTableEdit
} from '@mdi/js';
import { type InsightsExplorerProps, useExplorerPage } from './ExplorerPage/useExplorerPage.ts';

const props = defineProps<InsightsExplorerProps>();

const {
    tt, timezoneTypeIconMap, explorerChartTab: explorerChartTabRef, loading, updating, isCurrentDataTableEditable, activeTab,
    showCustomDateRangeDialog, isCustomChartInAIMode, isCustomChartAIGenerating, queryExportFileName, sampleQueryJson,
    allExplorations, allVisibleExplorations, currentFilter, currentExploration, filteredTransactionsInDataTable,
    allDateRanges, allTimezoneTypesUsedForDateRange, canShiftDateRange, displayQueryDateRangeName, displayQueryStartTime,
    displayQueryEndTime, isCurrentExplorationModified, allTabs, reload, createNewExploration, loadExploration,
    showChangeExplorerDisplayOrderDialog, saveExploration, restoreExploration, setExplorationName, hideExploration,
    removeExploration, importQueries, exportQueries, exportResults, setDateFilter, setCustomDateFilter, shiftDateRange,
    onImportQueries, onShowTransaction, onUpdateTransactions, onShowDateRangeError,
} = useExplorerPage(props);

</script>

<template>
    <main-page-layout>
        <template #nav-items>
            <div class="mb-2">
                <btn-vertical-group :disabled="loading || updating || isCustomChartAIGenerating || isCurrentDataTableEditable" :buttons="allTabs" v-model="activeTab" />
            </div>
            <v-divider class="my-2" />
            <div class="insights-explorer-list">
                <li class="nav-link" key="new">
                    <a class="d-flex align-center cursor-pointer"
                       :class="{ 'router-link-active router-link-exact-active': !currentExploration.id, 'disabled': loading || updating || isCustomChartAIGenerating || isCurrentDataTableEditable }"
                       @click="createNewExploration">
                        <span class="nav-item-title text-truncate">{{ tt('New Exploration') }}</span>
                    </a>
                </li>
                <li class="nav-link" :key="exploration.id" v-for="exploration in allVisibleExplorations">
                    <a class="d-flex align-center cursor-pointer"
                       :class="{ 'router-link-active router-link-exact-active': currentExploration.id === exploration.id, 'disabled': loading || updating || isCustomChartAIGenerating || isCurrentDataTableEditable }"
                       @click="loadExploration(exploration.id)">
                        <span class="nav-item-title text-truncate">{{ exploration.name || tt('Untitled Exploration') }}</span>
                    </a>
                </li>
                <template v-if="loading && (!allVisibleExplorations || allVisibleExplorations.length < 1)">
                    <v-skeleton-loader class="skeleton-no-margin mx-3 mt-h1 py-3" type="text"
                                       :key="itemIdx" :loading="true" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]"></v-skeleton-loader>
                </template>
            </div>
        </template>

        <template #content>
            <v-card min-height="800">
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Insights Explorer') }}</span>
                        <v-btn-group class="ms-4" color="default" density="comfortable" variant="outlined" divided>
                            <v-btn class="button-icon-with-direction" :icon="mdiArrowLeft"
                                   :aria-label="tt('Previous Period')" :disabled="loading || updating || isCustomChartAIGenerating || !canShiftDateRange || isCurrentDataTableEditable"
                                   @click="shiftDateRange(-1)"/>
                            <v-menu location="bottom" max-height="500">
                                <template #activator="{ props }">
                                    <v-btn :disabled="loading || updating || isCustomChartAIGenerating || isCurrentDataTableEditable"
                                           v-bind="props">{{ displayQueryDateRangeName }}</v-btn>
                                </template>
                                <v-list :selected="[currentFilter.dateRangeType]">
                                    <v-list-item :key="dateRange.type" :value="dateRange.type"
                                                 :append-icon="(currentFilter.dateRangeType === dateRange.type ? mdiCheck : undefined)"
                                                 v-for="dateRange in allDateRanges">
                                        <v-list-item-title class="cursor-pointer"
                                                           @click="setDateFilter(dateRange.type)">
                                            <div class="d-flex align-center">
                                                <span>{{ dateRange.displayName }}</span>
                                            </div>
                                            <div class="statistics-custom-datetime-range smaller" v-if="dateRange.isUserCustomRange && currentFilter.dateRangeType === dateRange.type && !!currentFilter.startTime && !!currentFilter.endTime">
                                                <span>{{ displayQueryStartTime }}</span>
                                                <span>&nbsp;-&nbsp;</span>
                                                <br/>
                                                <span>{{ displayQueryEndTime }}</span>
                                            </div>
                                        </v-list-item-title>
                                    </v-list-item>
                                </v-list>
                            </v-menu>
                            <v-btn class="button-icon-with-direction" :icon="mdiArrowRight"
                                   :aria-label="tt('Next Period')" :disabled="loading || updating || isCustomChartAIGenerating || !canShiftDateRange || isCurrentDataTableEditable"
                                   @click="shiftDateRange(1)"/>
                        </v-btn-group>

                        <v-btn density="compact" color="default" variant="text" class="ms-2"
                               :aria-label="tt('Refresh')" :icon="true" :loading="loading" :disabled="updating || isCustomChartAIGenerating" @click="reload(true)">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                        <v-btn density="compact" color="default" variant="text" class="ms-2"
                               :aria-label="tt('AI Generate Chart Code')" :icon="true" :disabled="loading || updating || isCustomChartInAIMode"
                               v-if="activeTab === 'chart' && currentExploration.chartType === TransactionExplorerChartTypeValue.Custom && isInsightsExplorerCodingAssistantEnabled()"
                               @click="explorerChartTabRef?.switchToAIChartMode()">
                            <v-icon :icon="mdiMagicStaff" size="20" />
                            <v-tooltip activator="parent">{{ tt('AI Generate Chart Code') }}</v-tooltip>
                        </v-btn>
                        <v-spacer/>
                        <v-btn class="ms-3"
                               :color="isCurrentExplorationModified ? 'primary' : 'default'"
                               :variant="isCurrentExplorationModified ? 'flat' : 'outlined'"
                               :disabled="loading || updating || isCustomChartAIGenerating || isCurrentDataTableEditable" @click="saveExploration(false)">
                            {{ tt('Save Exploration') }}
                            <v-progress-circular indeterminate size="22" class="ms-2" v-if="updating"></v-progress-circular>
                            <v-menu activator="parent" :open-on-hover="true">
                                <v-list>
                                    <v-list-item @click="saveExploration(true)">
                                        <v-list-item-title>{{ tt('Save As New Exploration') }}</v-list-item-title>
                                    </v-list-item>
                                    <v-list-item @click="restoreExploration()" v-if="currentExploration.id">
                                        <v-list-item-title>{{ tt('Restore to Last Saved') }}</v-list-item-title>
                                    </v-list-item>
                                </v-list>
                            </v-menu>
                        </v-btn>
                        <v-btn density="comfortable" color="default" variant="text" class="ms-2"
                               :aria-label="tt('More')" :disabled="loading || updating || isCustomChartAIGenerating" :icon="true">
                            <v-icon :icon="mdiDotsVertical" />
                            <v-menu activator="parent">
                                <v-list>
                                    <template v-if="activeTab === 'query'">
                                        <v-list-subheader class="text-body-small"
                                                          :title="tt('Timezone Used for Date Range')"/>
                                        <v-list-item :key="timezoneType.type" :value="timezoneType.type"
                                                     :prepend-icon="timezoneTypeIconMap[timezoneType.type]"
                                                     :append-icon="(currentExploration.timezoneUsedForDateRange === timezoneType.type ? mdiCheck : undefined)"
                                                     :title="timezoneType.displayName"
                                                     v-for="timezoneType in allTimezoneTypesUsedForDateRange"
                                                     @click="currentExploration.timezoneUsedForDateRange = timezoneType.type"></v-list-item>
                                        <v-divider class="my-2"/>
                                        <v-list-item :prepend-icon="mdiApplicationImport"
                                                     :title="tt('Import Queries')"
                                                     :disabled="loading || updating"
                                                     @click="importQueries"></v-list-item>
                                        <v-list-item :prepend-icon="mdiApplicationExport"
                                                     :title="tt('Export Queries')"
                                                     :disabled="loading || updating"
                                                     @click="exportQueries"></v-list-item>
                                        <v-divider class="my-2"/>
                                    </template>

                                    <template v-if="activeTab === 'table'">
                                        <v-list-item :prepend-icon="mdiTableEdit"
                                                     :title="tt('Enter Edit Mode')"
                                                     :disabled="loading || updating || filteredTransactionsInDataTable.length < 1"
                                                     @click="isCurrentDataTableEditable = true"
                                                     v-if="!isCurrentDataTableEditable"></v-list-item>
                                        <v-list-item :prepend-icon="mdiTableCheck"
                                                     :title="tt('Exit Edit Mode')"
                                                     :disabled="loading || updating"
                                                     @click="isCurrentDataTableEditable = false"
                                                     v-if="isCurrentDataTableEditable"></v-list-item>
                                        <v-divider class="my-2" v-if="!isCurrentDataTableEditable"/>
                                    </template>

                                    <template v-if="(activeTab === 'table' && !isCurrentDataTableEditable) || (activeTab === 'chart' && currentExploration.chartType !== TransactionExplorerChartTypeValue.Custom)">
                                        <v-list-item :prepend-icon="mdiExport"
                                                     :title="tt('Export Results')"
                                                     :disabled="loading || updating || (activeTab === 'table' && (!filteredTransactionsInDataTable || filteredTransactionsInDataTable.length < 1))"
                                                     @click="exportResults"></v-list-item>
                                        <v-divider class="my-2" />
                                    </template>

                                    <template v-if="currentExploration.id && !isCurrentDataTableEditable">
                                        <v-list-item :prepend-icon="mdiPencilOutline" @click="setExplorationName">
                                            <v-list-item-title>{{ tt('Rename Exploration') }}</v-list-item-title>
                                        </v-list-item>
                                        <v-list-item :prepend-icon="mdiEyeOffOutline" @click="hideExploration(true)" v-if="!currentExploration.hidden">
                                            <v-list-item-title>{{ tt('Hide Exploration') }}</v-list-item-title>
                                        </v-list-item>
                                        <v-list-item :prepend-icon="mdiEyeOutline" @click="hideExploration(false)" v-if="currentExploration.hidden">
                                            <v-list-item-title>{{ tt('Unhide Exploration') }}</v-list-item-title>
                                        </v-list-item>
                                        <v-list-item :prepend-icon="mdiDeleteOutline" @click="removeExploration">
                                            <v-list-item-title>{{ tt('Delete Exploration') }}</v-list-item-title>
                                        </v-list-item>
                                        <v-divider class="my-2"/>
                                    </template>

                                    <v-list-item :prepend-icon="mdiSort"
                                                 :disabled="!allExplorations || allExplorations.length < 2"
                                                 :title="tt('Change Exploration Display Order')"
                                                 @click="showChangeExplorerDisplayOrderDialog"
                                                 v-if="!isCurrentDataTableEditable"></v-list-item>
                                </v-list>
                            </v-menu>
                        </v-btn>
                    </div>
                </template>

                <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container" v-model="activeTab">
                    <v-window-item value="query">
                        <explorer-query-tab :loading="loading" :disabled="loading || updating || isCustomChartAIGenerating" />
                    </v-window-item>
                    <v-window-item value="table">
                        <explorer-data-table-tab ref="explorerDataTableTab"
                                                 :loading="loading" :disabled="loading || updating || isCustomChartAIGenerating"
                                                 @click:transaction="onShowTransaction"
                                                 v-if="!isCurrentDataTableEditable" />
                        <explorer-editable-data-table-tab ref="explorerEditableDataTableTab"
                                                          :loading="loading" :disabled="loading || updating || isCustomChartAIGenerating"
                                                          @click:transaction="onShowTransaction"
                                                          @update:transactions="onUpdateTransactions"
                                                          v-if="isCurrentDataTableEditable" />
                    </v-window-item>
                    <v-window-item value="chart">
                        <explorer-chart-tab ref="explorerChartTab"
                                            :loading="loading" :disabled="loading || updating"
                                            @click:transaction="onShowTransaction" />
                    </v-window-item>
                </v-window>
            </v-card>
        </template>
    </main-page-layout>

    <date-range-selection-dialog :title="tt('Custom Date Range')"
                                 :min-time="currentFilter.startTime"
                                 :max-time="currentFilter.endTime"
                                 v-model:show="showCustomDateRangeDialog"
                                 @dateRange:change="setCustomDateFilter"
                                 @error="onShowDateRangeError" />

    <explorer-change-display-order-dialog ref="explorerChangeDisplayOrderDialog" />
    <edit-dialog ref="editDialog" :type="TransactionEditPageType.Transaction" />
    <json-import-dialog ref="queryImportDialog" :title="tt('Import Queries')" :sample-json="sampleQueryJson" :on-import="onImportQueries" />
    <json-export-dialog ref="queryExportDialog" :title="tt('Export Queries')" :file-name="queryExportFileName" />
    <data-export-dialog ref="dataExportDialog" />

    <rename-dialog ref="renameDialog"
                   :default-title="tt('Rename Exploration')"
                   :label="tt('Exploration Name')" :placeholder="tt('Exploration Name')" />
    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>
