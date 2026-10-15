<script setup lang="ts">
import TransactionNavigation from './ListPage/TransactionNavigation.vue';
import TransactionTableHeader from './ListPage/TransactionTableHeader.vue';
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import PaginationButtons from '@/components/desktop/PaginationButtons.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import {
    type WeekDay,

} from '@/core/datetime.ts';

import { TransactionAmountType, TransactionType } from '@/core/transaction.ts';

import {

    getCategoryIconType
} from '@/lib/icon.ts';
import {
    isDataExportingEnabled,
    isDataImportingEnabled,
    isTransactionFromAIImageRecognitionEnabled,
    isTransactionFromAITextRecognitionEnabled
} from '@/lib/server_settings.ts';
import { allTransactionPictures } from '@/lib/transaction.ts';
import { type Transaction} from '@/models/transaction.ts';
import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import AccountFilterSettingsDialog from '@/views/desktop/common/dialogs/AccountFilterSettingsDialog.vue';
import CategoryFilterSettingsDialog from '@/views/desktop/common/dialogs/CategoryFilterSettingsDialog.vue';
import TransactionTagFilterSettingsDialog from '@/views/desktop/common/dialogs/TransactionTagFilterSettingsDialog.vue';
import {
    mdiArrowLeft,
    mdiArrowRight,

    mdiCheck,

    mdiMagicStaff,
    mdiMagnify,

    mdiPencilBoxOutline,
    mdiPound,
    mdiRefresh,
    mdiTextBoxEditOutline,
    mdiTextBoxOutline,

} from '@mdi/js';
import { VMenu } from 'vuetify/components/VMenu';
import ImportDialog from './import/ImportDialog.vue';
import AIImageRecognitionDialog from './list/dialogs/AIImageRecognitionDialog.vue';
import EditDialog from './list/dialogs/EditDialog.vue';
import { type TransactionListProps, useListPage } from './ListPage/useListPage.ts';

const props = defineProps<TransactionListProps>();

const {
    tt, formatRange, getWeekdayLongName, pageType, loading, customMinDatetime, customMaxDatetime, currentCalendarDate,
    userDefaultCurrency, selectedAccountDefaultCurrency, showTotalAmountInTransactionListPage,
    totalAmountTypeInTransactionListPage, showTagInTransactionListPage, allDateRanges, allAccounts, allAccountsMap,
    allAvailableAccountsCount, allCategories, allPrimaryCategories, allAvailableCategoriesCount,
    allTransactionTagGroupsWithDefault, allTransactionTagsByGroup, allTransactionTags, allAvailableTagsCount, query,
    queryMinTime, queryMaxTime, queryMonth, queryAllFilterCategoryIds, queryAllFilterAccountIds, queryAllFilterTagIds,
    queryAllFilterCategoryIdsCount, queryAllFilterAccountIdsCount, queryAllFilterTagIdsCount, queryAccountName,
    queryCategoryName, queryTagName, queryAmount, transactionCalendarMinDate, transactionCalendarMaxDate,
    currentMonthTransactionData, hasSubCategoryInQuery, hasVisibleTagsInTagGroup, isSameAsDefaultTimezoneOffsetMinutes,
    canAddTransaction, getDailyTotalAmounts, getDisplayTime, getDisplayLongDate, getDisplayTimezone,
    getDisplayTimeInDefaultTimezone, getDisplayAmount, getDisplayAmountCurrency, getTransactionTypeName,
    getTransactionPictureUrl, activeTab, searchKeyword, currentAmountFilterType, currentAmountFilterValue1,
    currentAmountFilterValue2, categoryMenuState, amountMenuState, exportingData, showCustomDateRangeDialog,
    showCustomMonthDialog, showFilterAccountDialog, showFilterCategoryDialog, showFilterTagDialog, isDarkMode, allPageCounts,
    recentMonthDateRanges, allTransactionTemplates, allowCategoryTypes, transactions, transactionsByDay, recentDateRangeIndex,
    queryPageType, queryType, queryAllSelectedFilterCategoryIds, queryAllSelectedFilterAccountIds,
    queryAllSelectedFilterTagIds, countPerPage, totalPageCount, paginationCurrentPage, skeletonData, currentMonthTotalAmount,
    getCategoryListItemCheckedClass, reload, changeDateFilter, changeCustomDateFilter, changeCustomMonthDateFilter,
    shiftDateRange, changeCategoryFilter, changeMultipleCategoriesFilter, changeAccountFilter, changeMultipleAccountsFilter,
    changeTagFilter, changeMultipleTagsFilter, changeKeywordFilter, changeAmountFilter, add, addByRecognizingClipboardText,
    addByRecognizingImage, importTransaction, exportTransactions, show, onShowDateRangeError,
} = useListPage(props);

</script>

<template>
    <main-page-layout>
        <template #nav-items>
<TransactionNavigation
    :loading="loading"
    :pageType="pageType"
    :allPageCounts="allPageCounts"
    :recentMonthDateRanges="recentMonthDateRanges"
    :recentDateRangeIndex="recentDateRangeIndex"
    v-model:queryPageType="queryPageType"
    v-model:queryType="queryType"
    v-model:countPerPage="countPerPage"
    @changeDateFilter="changeDateFilter"
/>
</template>

        <template #content>
            <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container" v-model="activeTab">
                <v-window-item value="transactionPage">
                    <v-card min-height="920">
                        <template #title>
                            <div class="title-and-toolbar d-flex align-center text-no-wrap">
                                <span>{{ tt('Transaction List') }}</span>
                                <v-btn class="ms-3" color="default" variant="outlined"
                                       :disabled="loading || !canAddTransaction" @click="add()">
                                    {{ tt('Add') }}
                                    <v-menu activator="parent" max-height="500" :open-on-hover="true" v-if="isTransactionFromAITextRecognitionEnabled() || isTransactionFromAIImageRecognitionEnabled() || (allTransactionTemplates && allTransactionTemplates.length)">
                                        <v-list>
                                            <v-list-item key="AIClipboardTextRecognition"
                                                         :title="tt('AI Clipboard Text Recognition')"
                                                         :prepend-icon="mdiMagicStaff"
                                                         v-if="isTransactionFromAITextRecognitionEnabled()"
                                                         @click="addByRecognizingClipboardText"></v-list-item>
                                            <v-list-item key="AIImageRecognition"
                                                         :title="tt('AI Image Recognition')"
                                                         :prepend-icon="mdiMagicStaff"
                                                         v-if="isTransactionFromAIImageRecognitionEnabled()"
                                                         @click="addByRecognizingImage"></v-list-item>
                                            <v-list-item :key="template.id"
                                                         :title="template.name"
                                                         :prepend-icon="mdiTextBoxOutline"
                                                         v-for="template in allTransactionTemplates"
                                                         @click="add(template)"></v-list-item>
                                        </v-list>
                                    </v-menu>
                                </v-btn>
                                <v-btn class="ms-3" color="default" variant="outlined"
                                       :disabled="loading" @click="importTransaction"
                                       v-if="isDataImportingEnabled()">
                                    {{ tt('Import') }}
                                    <v-menu activator="parent" :open-on-hover="true" v-if="isDataExportingEnabled()">
                                        <v-list>
                                            <v-list-item :disabled="loading || exportingData || !transactions || !transactions.length || transactions.length < 1"
                                                         @click="exportTransactions('csv')">
                                                <v-list-item-title>{{ tt('Export to CSV (Comma-separated values) File') }}</v-list-item-title>
                                            </v-list-item>
                                            <v-list-item :disabled="loading || exportingData || !transactions || !transactions.length || transactions.length < 1"
                                                         @click="exportTransactions('tsv')">
                                                <v-list-item-title>{{ tt('Export to TSV (Tab-separated values) File') }}</v-list-item-title>
                                            </v-list-item>
                                        </v-list>
                                    </v-menu>
                                </v-btn>
                                <v-btn class="ms-3" color="default" variant="outlined"
                                       :disabled="loading || exportingData || !transactions || !transactions.length || transactions.length < 1" v-if="!isDataImportingEnabled() && isDataExportingEnabled()">
                                    {{ tt('Export') }}
                                    <v-menu activator="parent">
                                        <v-list>
                                            <v-list-item :disabled="loading || exportingData || !transactions || !transactions.length || transactions.length < 1"
                                                         @click="exportTransactions('csv')">
                                                <v-list-item-title>{{ tt('Export to CSV (Comma-separated values) File') }}</v-list-item-title>
                                            </v-list-item>
                                            <v-list-item :disabled="loading || exportingData || !transactions || !transactions.length || transactions.length < 1"
                                                         @click="exportTransactions('tsv')">
                                                <v-list-item-title>{{ tt('Export to TSV (Tab-separated values) File') }}</v-list-item-title>
                                            </v-list-item>
                                        </v-list>
                                    </v-menu>
                                </v-btn>
                                <v-btn density="compact" color="default" variant="text" class="ms-2"
                                       :aria-label="tt('Refresh')" :icon="true" :loading="loading" @click="reload(true, false)">
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
                                                  :append-inner-icon="searchKeyword !== query.keyword ? mdiCheck : undefined"
                                                  :placeholder="tt('Search transaction description')"
                                                  v-model="searchKeyword"
                                                  @click:append-inner="changeKeywordFilter(searchKeyword)"
                                                  @keyup.enter="changeKeywordFilter(searchKeyword)"
                                    />
                                </div>
                            </div>
                        </template>

                        <v-card-text class="pt-0">
                            <div class="transaction-list-datetime-range d-flex align-center">
                                <span class="text-body-large">{{ tt('Date Range') }}</span>
                                <div class="d-flex transaction-list-datetime-range-text ms-2"
                                      v-if="!query.minTime && !query.maxTime">
                                    <span class="text-body-medium">{{ tt('All') }}</span>
                                </div>
                                <div class="d-flex transaction-list-datetime-range-text ms-2"
                                      v-else-if="query.minTime || query.maxTime">
                                    <v-btn class="button-icon-with-direction" size="small"
                                           density="compact" color="default" variant="outlined"
                                           :aria-label="tt('Previous Period')" :disabled="loading" :icon="true"
                                           @click="shiftDateRange(query.minTime, query.maxTime, -1)">
                                        <v-icon :icon="mdiArrowLeft" size="14" />
                                    </v-btn>
                                    <span class="text-body-medium mx-1">{{ formatRange(queryMinTime, queryMaxTime) }}</span>
                                    <v-btn class="button-icon-with-direction" size="small"
                                           density="compact" color="default" variant="outlined"
                                           :aria-label="tt('Next Period')" :disabled="loading" :icon="true"
                                           @click="shiftDateRange(query.minTime, query.maxTime, 1)">
                                        <v-icon :icon="mdiArrowRight" size="14" />
                                    </v-btn>
                                </div>
                                <v-spacer/>
                                <div class="skeleton-no-margin d-flex align-center" v-if="showTotalAmountInTransactionListPage && currentMonthTotalAmount">
                                    <span class="ms-2 text-body-large">{{ totalAmountTypeInTransactionListPage === TransactionAmountType.IncomeAndExpense ? tt('Total Income') : tt('Total Inflows') }}</span>
                                    <span class="text-income ms-2" v-if="loading">
                                        <v-skeleton-loader type="text" style="width: 60px" :loading="true"></v-skeleton-loader>
                                    </span>
                                    <span class="text-body-large text-income ms-2" v-else-if="!loading">
                                        {{ currentMonthTotalAmount.income }}
                                        <v-tooltip activator="parent" v-if="!currentMonthTotalAmount.incomeIsZero && currentMonthTotalAmount.incomeInDefaultCurrency !== currentMonthTotalAmount.income">
                                            <span>{{ currentMonthTotalAmount.incomeInDefaultCurrency }}</span>
                                        </v-tooltip>
                                    </span>
                                    <span class="text-body-large ms-3">{{ totalAmountTypeInTransactionListPage === TransactionAmountType.IncomeAndExpense ? tt('Total Expense') : tt('Total Outflows') }}</span>
                                    <span class="text-expense ms-2" v-if="loading">
                                        <v-skeleton-loader type="text" style="width: 60px" :loading="true"></v-skeleton-loader>
                                    </span>
                                    <span class="text-body-large text-expense ms-2" v-else-if="!loading">
                                        {{ currentMonthTotalAmount.expense }}
                                        <v-tooltip activator="parent" v-if="!currentMonthTotalAmount.expenseIsZero && currentMonthTotalAmount.expenseInDefaultCurrency !== currentMonthTotalAmount.expense">
                                            <span>{{ currentMonthTotalAmount.expenseInDefaultCurrency }}</span>
                                        </v-tooltip>
                                    </span>
                                </div>
                            </div>
                        </v-card-text>

                        <v-card-text class="transaction-calendar-container pt-0" v-if="pageType === TransactionListPageType.Calendar.type">
                            <transaction-calendar show-amount show-income-amount show-expense-amount show-alternate-date
                                                  dates-have-any-transaction-clickable
                                                  day-has-transaction-class="font-weight-bold"
                                                  :readonly="loading" :is-dark-mode="isDarkMode"
                                                  :default-currency="selectedAccountDefaultCurrency"
                                                  :min-date="transactionCalendarMinDate"
                                                  :max-date="transactionCalendarMaxDate"
                                                  :daily-total-amounts="getDailyTotalAmounts(currentMonthTransactionData)"
                                                  v-model="currentCalendarDate"></transaction-calendar>
                        </v-card-text>

                        <v-table class="transaction-table" :hover="!loading" v-if="pageType !== TransactionListPageType.Gallery.type">
                            <thead>
                            <TransactionTableHeader
    :loading="loading"
    :query="query"
    :allDateRanges="allDateRanges"
    :queryMinTime="queryMinTime"
    :queryMaxTime="queryMaxTime"
    :queryCategoryName="queryCategoryName"
    :queryAllSelectedFilterCategoryIds="queryAllSelectedFilterCategoryIds"
    :queryAllFilterCategoryIdsCount="queryAllFilterCategoryIdsCount"
    :allAvailableCategoriesCount="allAvailableCategoriesCount"
    :allPrimaryCategories="allPrimaryCategories"
    :getTransactionTypeName="getTransactionTypeName"
    :queryAllFilterCategoryIds="queryAllFilterCategoryIds"
    :allCategories="allCategories"
    :hasSubCategoryInQuery="hasSubCategoryInQuery"
    :getCategoryListItemCheckedClass="getCategoryListItemCheckedClass"
    :queryAmount="queryAmount"
    :selectedAccountDefaultCurrency="selectedAccountDefaultCurrency"
    :queryAccountName="queryAccountName"
    :queryAllSelectedFilterAccountIds="queryAllSelectedFilterAccountIds"
    :queryAllFilterAccountIdsCount="queryAllFilterAccountIdsCount"
    :allAvailableAccountsCount="allAvailableAccountsCount"
    :allAccounts="allAccounts"
    :allAccountsMap="allAccountsMap"
    :queryAllFilterAccountIds="queryAllFilterAccountIds"
    :showTagInTransactionListPage="showTagInTransactionListPage"
    :queryTagName="queryTagName"
    :queryAllSelectedFilterTagIds="queryAllSelectedFilterTagIds"
    :queryAllFilterTagIdsCount="queryAllFilterTagIdsCount"
    :allAvailableTagsCount="allAvailableTagsCount"
    :allTransactionTagGroupsWithDefault="allTransactionTagGroupsWithDefault"
    :allTransactionTagsByGroup="allTransactionTagsByGroup"
    :hasVisibleTagsInTagGroup="hasVisibleTagsInTagGroup"
    :queryAllFilterTagIds="queryAllFilterTagIds"
    v-model:categoryMenuState="categoryMenuState"
    v-model:amountMenuState="amountMenuState"
    v-model:currentAmountFilterType="currentAmountFilterType"
    v-model:currentAmountFilterValue1="currentAmountFilterValue1"
    v-model:currentAmountFilterValue2="currentAmountFilterValue2"
    v-model:showFilterCategoryDialog="showFilterCategoryDialog"
    v-model:showFilterAccountDialog="showFilterAccountDialog"
    v-model:showFilterTagDialog="showFilterTagDialog"
    @changeDateFilter="changeDateFilter"
    @changeCategoryFilter="changeCategoryFilter"
    @changeAmountFilter="changeAmountFilter"
    @changeAccountFilter="changeAccountFilter"
    @changeTagFilter="changeTagFilter"
/>
                            </thead>

                            <tbody v-if="loading && (!transactions || !transactions.length || transactions.length < 1)">
                            <tr :key="itemIdx" v-for="itemIdx in skeletonData">
                                <td class="px-0" :colspan="showTagInTransactionListPage ? 6 : 5">
                                    <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                                </td>
                            </tr>
                            </tbody>

                            <tbody v-if="!loading && (!transactions || !transactions.length || transactions.length < 1)">
                            <tr>
                                <td :colspan="showTagInTransactionListPage ? 6 : 5">{{ tt('No transaction data') }}</td>
                            </tr>
                            </tbody>

                            <tbody :key="transaction.id"
                                   :class="{ 'disabled': loading, 'has-bottom-border': idx < transactions.length - 1 }"
                                   v-for="(transaction, idx) in transactions">
                                <tr class="transaction-list-row-date no-hover text-body-small"
                                    v-if="pageType === TransactionListPageType.List.type && (idx === 0 || (idx > 0 && (transaction.gregorianCalendarYearDashMonthDashDay !== transactions[idx - 1]!.gregorianCalendarYearDashMonthDashDay)))">
                                    <td :colspan="showTagInTransactionListPage ? 6 : 5" class="font-weight-bold">
                                        <div class="d-flex align-center">
                                            <span>{{ getDisplayLongDate(transaction) }}</span>
                                            <v-chip class="ms-1" color="default" size="x-small"
                                                    v-if="transaction.displayDayOfWeek">
                                                {{ getWeekdayLongName(transaction.displayDayOfWeek) }}
                                            </v-chip>
                                        </div>
                                    </td>
                                </tr>
                                <tr class="transaction-table-row-data cursor-pointer"
                                    @click="show(transaction)">
                                    <td class="transaction-table-column-time">
                                        <div class="d-flex flex-column">
                                            <span>{{ getDisplayTime(transaction) }}</span>
                                            <span class="text-body-small text-medium-emphasis" v-if="!isSameAsDefaultTimezoneOffsetMinutes(transaction)">{{ getDisplayTimezone(transaction) }}</span>
                                            <v-tooltip activator="parent" v-if="!isSameAsDefaultTimezoneOffsetMinutes(transaction)">{{ getDisplayTimeInDefaultTimezone(transaction) }}</v-tooltip>
                                        </div>
                                    </td>
                                    <td class="transaction-table-column-category">
                                        <div class="d-flex align-center">
                                            <ItemIcon size="24px" :icon-type="getCategoryIconType(transaction.category.iconType)"
                                                      :icon-id="transaction.category.icon"
                                                      :color="transaction.category.color"
                                                      v-if="transaction.category && transaction.category.color"></ItemIcon>
                                            <v-icon size="24" :icon="mdiPencilBoxOutline" v-else-if="!transaction.category || !transaction.category.color" />
                                            <span class="ms-1" v-if="transaction.type === TransactionType.ModifyBalance">
                                                {{ tt('Modify Balance') }}
                                            </span>
                                            <span class="ms-1" v-else-if="transaction.type !== TransactionType.ModifyBalance && transaction.category">
                                                {{ transaction.category.name }}
                                            </span>
                                            <span class="ms-1" v-else-if="transaction.type !== TransactionType.ModifyBalance && !transaction.category">
                                                {{ getTransactionTypeName(transaction.type, 'Transaction') }}
                                            </span>
                                        </div>
                                    </td>
                                    <td class="transaction-table-column-amount" :class="{ 'text-expense': transaction.type === TransactionType.Expense, 'text-income': transaction.type === TransactionType.Income }">
                                        <div v-if="transaction.sourceAccount">
                                            <span>{{ getDisplayAmount(transaction) }}</span>
                                            <v-tooltip activator="parent" v-if="!transaction.hideAmount && getDisplayAmountCurrency(transaction) !== userDefaultCurrency">
                                                {{ getDisplayAmount(transaction, true) }}
                                            </v-tooltip>
                                        </div>
                                    </td>
                                    <td class="transaction-table-column-account">
                                        <div class="d-flex align-center">
                                            <span v-if="transaction.sourceAccount">{{ transaction.sourceAccount.name }}</span>
                                            <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" v-if="transaction.sourceAccount && transaction.type === TransactionType.Transfer && transaction.destinationAccount && transaction.sourceAccount.id !== transaction.destinationAccount.id"></v-icon>
                                            <span v-if="transaction.sourceAccount && transaction.type === TransactionType.Transfer && transaction.destinationAccount && transaction.sourceAccount.id !== transaction.destinationAccount.id">{{ transaction.destinationAccount.name }}</span>
                                        </div>
                                    </td>
                                    <td class="transaction-table-column-tags" v-if="showTagInTransactionListPage">
                                        <v-chip class="transaction-tag" size="small" :prepend-icon="mdiPound"
                                                :text="allTransactionTags[tagId]?.name"
                                                :key="tagId"
                                                v-for="tagId in transaction.tagIds"/>
                                        <v-chip class="transaction-tag" size="small"
                                                :text="tt('None')"
                                                v-if="!transaction.tagIds || !transaction.tagIds.length"/>
                                    </td>
                                    <td class="transaction-table-column-description text-truncate">
                                        {{ transaction.comment }}
                                    </td>
                                </tr>
                            </tbody>
                        </v-table>

                        <v-card-text class="transaction-gallery-container pt-0" v-if="pageType === TransactionListPageType.Gallery.type">
                            <div v-if="loading && (!transactions || !transactions.length || transactions.length < 1)">
                                <v-skeleton-loader class="skeleton-no-margin" type="text" :loading="true"></v-skeleton-loader>
                            </div>

                            <div class="text-body-medium" v-if="!loading && (!transactions || !transactions.length || transactions.length < 1)">
                                {{ tt('No transaction data') }}
                            </div>

                            <div :key="date" :class="{ 'disabled': loading }"
                                 v-for="(transactions, date) in transactionsByDay">
                                <div class="transaction-date font-weight-bold">
                                    <div class="d-flex align-center">
                                        <span>{{ getDisplayLongDate(transactions[0] as Transaction) }}</span>
                                        <v-chip class="ms-1" color="default" size="x-small"
                                                v-if="(transactions[0] as Transaction).displayDayOfWeek">
                                            {{ getWeekdayLongName((transactions[0] as Transaction).displayDayOfWeek as WeekDay) }}
                                        </v-chip>
                                    </div>
                                </div>
                                <div class="d-flex flex-wrap gap-2 py-2">
                                    <v-avatar rounded="lg" variant="tonal" size="160"
                                              class="cursor-pointer transaction-picture" color="rgba(0,0,0,0)"
                                              :key="pictureInfo.pictureId"
                                              v-for="[transaction, pictureInfo] in allTransactionPictures(transactions)"
                                              @click="show(transaction)">
                                        <v-img :src="getTransactionPictureUrl(pictureInfo)">
                                            <template #placeholder>
                                                <div class="d-flex align-center justify-center bg-light-primary">
                                                    <v-progress-circular color="grey-500" indeterminate size="48"></v-progress-circular>
                                                </div>
                                            </template>
                                            <template #error>
                                                <div class="d-flex align-center justify-center bg-light-primary">
                                                    <span class="text-body-large">{{ tt('Failed to load image, please check whether the config "domain" and "root_url" are set correctly.') }}</span>
                                                </div>
                                            </template>
                                        </v-img>
                                        <div class="picture-control-icon">
                                            <v-icon size="64" :icon="mdiTextBoxEditOutline"/>
                                        </div>
                                    </v-avatar>
                                </div>
                            </div>
                        </v-card-text>

                        <div class="mt-2 mb-4" v-if="pageType === TransactionListPageType.List.type || pageType === TransactionListPageType.Gallery.type">
                            <pagination-buttons density="comfortable"
                                                :totalPageCount="totalPageCount" :disabled="loading"
                                                v-model="paginationCurrentPage"></pagination-buttons>
                        </div>
                    </v-card>
                </v-window-item>
            </v-window>
        </template>
    </main-page-layout>

    <date-range-selection-dialog :title="tt('Custom Date Range')"
                                 :min-time="customMinDatetime"
                                 :max-time="customMaxDatetime"
                                 v-model:show="showCustomDateRangeDialog"
                                 @dateRange:change="changeCustomDateFilter"
                                 @error="onShowDateRangeError" />

    <month-selection-dialog :title="tt('Select Month')"
                            :model-value="queryMonth"
                            v-model:show="showCustomMonthDialog"
                            @update:modelValue="changeCustomMonthDateFilter"
                            @error="onShowDateRangeError" />

    <edit-dialog ref="editDialog" :type="TransactionEditPageType.Transaction" />
    <a-i-image-recognition-dialog ref="aiImageRecognitionDialog" />
    <import-dialog ref="importDialog" :persistent="true" />

    <account-filter-settings-dialog type="transactionListCurrent"
                                    v-model:show="showFilterAccountDialog"
                                    @settings:change="changeMultipleAccountsFilter" />

    <category-filter-settings-dialog type="transactionListCurrent"
                                     :category-types="allowCategoryTypes"
                                     v-model:show="showFilterCategoryDialog"
                                     @settings:change="changeMultipleCategoriesFilter" />

    <transaction-tag-filter-settings-dialog type="transactionListCurrent"
                                            v-model:show="showFilterTagDialog"
                                            @settings:change="changeMultipleTagsFilter" />

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>



<style src="./ListPage/list.css"></style>
