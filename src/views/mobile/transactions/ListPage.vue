<script setup lang="ts">
import TransactionMonthList from './ListPage/TransactionMonthList.vue';
import TransactionListSkeleton from './ListPage/TransactionListSkeleton.vue';
import {
    DateRange
} from '@/core/datetime.ts';
import { AmountFilterType } from '@/core/numeral.ts';

import {
    categoryTypeToTransactionType
} from '@/lib/category.ts';
import {
    isDefined
} from '@/lib/common.ts';
import {
    getAccountIconType,
    getCategoryIconType
} from '@/lib/icon.ts';

import { TransactionTagFilter } from '@/models/transaction.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import type { Router } from 'framework7/types';
import { useListPage } from './ListPage/useListPage.ts';

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const {
    tt, pageType, loading, customMinDatetime, customMaxDatetime, currentCalendarDate, showTotalAmountInTransactionListPage,
    showTagInTransactionListPage, allDateRanges, allAccounts, allAccountsMap, allAvailableAccountsCount, allCategories,
    allPrimaryCategories, allAvailableCategoriesCount, allTransactionTagGroupsWithDefault, allTransactionTagsByGroup,
    allTransactionTags, allAvailableTagsCount, displayPageTypeName, query, queryDateRangeName, queryMinTime, queryMaxTime,
    queryMonth, queryAllFilterCategoryIds, queryAllFilterAccountIds, queryAllFilterTagIds, queryAllFilterCategoryIdsCount,
    queryAllFilterAccountIdsCount, queryAllFilterTagIdsCount, queryAccountName, queryCategoryName, queryAmount,
    transactionCalendarMinDate, transactionCalendarMaxDate, currentMonthTransactionData, hasSubCategoryInQuery,
    hasVisibleTagsInTagGroup, isSameAsDefaultTimezoneOffsetMinutes, canAddTransaction, hasMonthTotalAmount,
    getDailyTotalAmounts, getDisplayMonthTotalAmount, getDisplayTime, getDisplayLongYearMonth, getDisplayTimezone,
    getDisplayAmount, getTransactionTypeName, getTransactionPictureUrl, loadingMore, transactionToDelete, showSearchbar,
    showCustomDateRangeSheet, showCustomMonthSheet, showDeleteActionSheet, textDirection, isDarkMode, transactions,
    noTransaction, hasMoreTransaction, getTransactionMonthTitleDomId, getTransactionMonthListDomId, getTransactionDomId,
    isTransactionMonthListInvisible, getTransactionMonthListHeight, getTransactionDateStyle, getCategoryListItemCheckedClass,
    reload, loadMore, changePageType, changeDateFilter, changeCustomDateFilter, changeCustomMonthDateFilter, shiftDateRange,
    changeTypeFilter, changeCategoryFilter, filterMultipleCategories, changeAccountFilter, filterMultipleAccounts,
    changeTagFilter, filterMultipleTags, toggleSearchbar, changeKeywordFilter, changeAmountFilter, add, duplicate, edit,
    remove, collapseTransactionMonthList, onPopoverOpen, onCategoryPopoverOpen, onPageAfterIn,
    onTransactionMonthListCollapseStateChanged,
} = useListPage(props);

</script>

<template>
    <f7-page ptr
             infinite
             :infinite-preloader="loadingMore"
             :infinite-distance="600"
             :with-subnavbar="showSearchbar"
             @ptr:refresh="reload"
             @page:afterin="onPageAfterIn"
             @infinite="loadMore(true)">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title>
                <f7-link popover-open=".chart-data-type-popover-menu" :class="{ 'disabled': loading }">
                    <span style="color: var(--f7-text-color)">{{ displayPageTypeName }}</span>
                    <f7-icon class="page-title-bar-icon" color="gray" style="opacity: 0.5" f7="chevron_down_circle_fill"></f7-icon>
                </f7-link>
            </f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }">
                <f7-link icon-f7="search" :aria-label="tt('Search')" @click="toggleSearchbar"></f7-link>
                <f7-link icon-f7="plus" :class="{ 'disabled': !canAddTransaction }" :aria-label="tt('Add')" @click="add"></f7-link>
            </f7-nav-right>

            <f7-subnavbar :inner="false" v-if="showSearchbar">
                <f7-searchbar
                    custom-searchs
                    :value="query.keyword"
                    :placeholder="tt('Search transaction description')"
                    :disable-button-text="tt('Cancel')"
                    @change="changeKeywordFilter($event.target.value)"
                    @click:clear="changeKeywordFilter(''); showSearchbar = false"
                    @searchbar:disable="changeKeywordFilter(''); showSearchbar = false"
                ></f7-searchbar>
            </f7-subnavbar>
        </f7-navbar>

        <f7-popover class="chart-data-type-popover-menu">
            <f7-list dividers>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt(type.name)"
                              :class="{ 'list-item-selected': pageType === type.type }"
                              :key="type.type"
                              v-for="type in TransactionListPageType.values()"
                              @click="changePageType(type.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="pageType === type.type"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <f7-toolbar tabbar bottom class="compact-tabbar toolbar-item-auto-size transaction-list-toolbar">
            <f7-link :class="{ 'disabled': loading || query.dateType === DateRange.All.type }" :aria-label="tt('Previous Period')" @click="shiftDateRange(query.minTime, query.maxTime, -1)">
                <f7-icon class="icon-with-direction" f7="arrow_left_square"></f7-icon>
            </f7-link>
            <f7-link popover-open=".date-popover-menu" :class="{ 'tabbar-text-with-ellipsis': true, 'disabled': loading }">
                <span :class="{ 'tabbar-item-changed': query.dateType !== DateRange.All.type }">{{ queryDateRangeName }}</span>
            </f7-link>
            <f7-link :class="{ 'disabled': loading || query.dateType === DateRange.All.type }" :aria-label="tt('Next Period')" @click="shiftDateRange(query.minTime, query.maxTime, 1)">
                <f7-icon class="icon-with-direction" f7="arrow_right_square"></f7-icon>
            </f7-link>
            <f7-link popover-open=".category-popover-menu" :class="{ 'tabbar-text-with-ellipsis': true, 'disabled': loading || query.type === 1 }">
                <span :class="{ 'tabbar-item-changed': query.categoryIds }">{{ queryCategoryName }}</span>
            </f7-link>
            <f7-link popover-open=".account-popover-menu" :class="{ 'tabbar-text-with-ellipsis': true, 'disabled': loading }">
                <span :class="{ 'tabbar-item-changed': query.accountIds }">{{ queryAccountName }}</span>
            </f7-link>
            <f7-link popover-open=".more-popover-menu" :class="{ 'disabled': loading }" :aria-label="tt('More')">
                <f7-icon f7="ellipsis_vertical" :class="{ 'tabbar-item-changed': query.type > 0 || query.amountFilter || query.tagFilter }"></f7-icon>
            </f7-link>
        </f7-toolbar>

        <f7-block class="transaction-calendar-container" :class="{ 'margin-vertical': showSearchbar, 'margin-vertical-half': !showSearchbar }"
                  v-if="pageType === TransactionListPageType.Calendar.type">
            <transaction-calendar show-amount show-income-amount show-expense-amount show-alternate-date
                                  dates-have-any-transaction-clickable
                                  calendar-class="justify-content-center" week-day-name-type="short"
                                  :readonly="loading" :is-dark-mode="isDarkMode"
                                  :default-currency="false"
                                  :min-date="transactionCalendarMinDate"
                                  :max-date="transactionCalendarMaxDate"
                                  :daily-total-amounts="getDailyTotalAmounts(currentMonthTransactionData)"
                                  v-model="currentCalendarDate"></transaction-calendar>
        </f7-block>

        <TransactionListSkeleton
    v-if="loading"
    :loading="loading"
    :pageType="pageType"
    :showSearchbar="showSearchbar"
    :showTotalAmountInTransactionListPage="showTotalAmountInTransactionListPage"
/>

        <f7-list strong inset dividers :class="{ 'margin-vertical': pageType === TransactionListPageType.Calendar.type || showSearchbar, 'margin-vertical-half': pageType !== TransactionListPageType.Calendar.type && !showSearchbar }"
                 v-if="!loading && noTransaction">
            <f7-list-item :title="tt('No transaction data')"></f7-list-item>
        </f7-list>

        <TransactionMonthList
    :pageType="pageType"
    :showSearchbar="showSearchbar"
    :transactions="transactions"
    :getTransactionMonthTitleDomId="getTransactionMonthTitleDomId"
    :getDisplayLongYearMonth="getDisplayLongYearMonth"
    :showTotalAmountInTransactionListPage="showTotalAmountInTransactionListPage"
    :hasMonthTotalAmount="hasMonthTotalAmount"
    :getDisplayMonthTotalAmount="getDisplayMonthTotalAmount"
    :getTransactionMonthListHeight="getTransactionMonthListHeight"
    :isTransactionMonthListInvisible="isTransactionMonthListInvisible"
    :getTransactionMonthListDomId="getTransactionMonthListDomId"
    :getTransactionDomId="getTransactionDomId"
    :getTransactionDateStyle="getTransactionDateStyle"
    :getTransactionTypeName="getTransactionTypeName"
    :getDisplayAmount="getDisplayAmount"
    :showTagInTransactionListPage="showTagInTransactionListPage"
    :allTransactionTags="allTransactionTags"
    :getDisplayTime="getDisplayTime"
    :isSameAsDefaultTimezoneOffsetMinutes="isSameAsDefaultTimezoneOffsetMinutes"
    :getDisplayTimezone="getDisplayTimezone"
    :textDirection="textDirection"
    :getTransactionPictureUrl="getTransactionPictureUrl"
    @collapseTransactionMonthList="collapseTransactionMonthList"
    @onTransactionMonthListCollapseStateChanged="onTransactionMonthListCollapseStateChanged"
    @duplicate="duplicate"
    @edit="edit"
    @remove="remove"
/>

        <f7-block class="text-align-center" :class="{ 'disabled': loadingMore }" v-show="!loading && hasMoreTransaction"
                  v-if="pageType === TransactionListPageType.List.type || pageType === TransactionListPageType.Gallery.type">
            <f7-link href="#" @click="loadMore(false)">{{ tt('Load More') }}</f7-link>
        </f7-block>

        <f7-popover class="date-popover-menu" @popover:open="onPopoverOpen">
            <f7-list dividers>
                <f7-list-item link="#" no-chevron popover-close
                              :title="dateRange.displayName"
                              :class="{ 'list-item-selected': query.dateType === dateRange.type }"
                              :key="dateRange.type"
                              v-for="dateRange in allDateRanges"
                              v-show="pageType === TransactionListPageType.List.type || pageType === TransactionListPageType.Gallery.type || dateRange.type === DateRange.ThisMonth.type || dateRange.type === DateRange.LastMonth.type || dateRange.type === DateRange.Custom.type"
                              @click="changeDateFilter(dateRange.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.dateType === dateRange.type"></f7-icon>
                    </template>
                    <template #footer>
                        <div v-if="dateRange.isUserCustomRange && query.dateType === dateRange.type && query.minTime && query.maxTime">
                            <span>{{ queryMinTime }}</span>
                            <span>&nbsp;-&nbsp;</span>
                            <br/>
                            <span>{{ queryMaxTime }}</span>
                        </div>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <date-range-selection-sheet :title="tt('Custom Date Range')"
                                    :min-time="customMinDatetime"
                                    :max-time="customMaxDatetime"
                                    v-model:show="showCustomDateRangeSheet"
                                    @dateRange:change="changeCustomDateFilter">
        </date-range-selection-sheet>

        <month-selection-sheet :title="tt('Select Month')"
                               :model-value="queryMonth"
                               v-model:show="showCustomMonthSheet"
                               @update:modelValue="changeCustomMonthDateFilter">
        </month-selection-sheet>

        <f7-popover class="category-popover-menu" @popover:open="onCategoryPopoverOpen">
            <f7-list dividers accordion-list>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': !query.categoryIds }"
                              :title="tt('All')"
                              @click="changeCategoryFilter('')">
                    <template #media>
                        <f7-icon f7="rectangle_grid_2x2"></f7-icon>
                    </template>
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="!query.categoryIds"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.categoryIds && queryAllFilterCategoryIdsCount > 1 }"
                              :title="tt('Multiple Categories')"
                              @click="filterMultipleCategories()"
                              v-if="allAvailableCategoriesCount > 0">
                    <template #media>
                        <f7-icon f7="rectangle_on_rectangle"></f7-icon>
                    </template>
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.categoryIds && queryAllFilterCategoryIdsCount > 1"></f7-icon>
                    </template>
                </f7-list-item>
                <template :key="categoryType"
                          v-for="(categories, categoryType) in allPrimaryCategories">
                    <template v-if="categories && categories.length">
                        <f7-list-item divider :title="getTransactionTypeName(categoryTypeToTransactionType(parseInt(categoryType)), 'Type')"></f7-list-item>
                        <f7-list-item accordion-item
                                      :title="category.name"
                                      :class="getCategoryListItemCheckedClass(category, queryAllFilterCategoryIds)"
                                      :key="category.id"
                                      v-for="category in categories"
                                      v-show="!category.hidden || queryAllFilterCategoryIds[category.id] || allCategories[query.categoryIds]?.parentId === category.id || hasSubCategoryInQuery(category)"
                        >
                            <template #media>
                                <ItemIcon :icon-type="getCategoryIconType(category.iconType)" :icon-id="category.icon" :color="category.color"></ItemIcon>
                            </template>
                            <f7-accordion-content>
                                <f7-list dividers class="padding-inline-start">
                                    <f7-list-item link="#" no-chevron popover-close
                                                  :class="{ 'list-item-selected': query.categoryIds === category.id, 'item-in-multiple-selection': queryAllFilterCategoryIdsCount > 1 && queryAllFilterCategoryIds[category.id] }"
                                                  :title="tt('All')" @click="changeCategoryFilter(category.id)">
                                        <template #media>
                                            <f7-icon f7="rectangle_grid_2x2"></f7-icon>
                                        </template>
                                        <template #after>
                                            <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.categoryIds === category.id"></f7-icon>
                                        </template>
                                    </f7-list-item>
                                    <f7-list-item link="#" no-chevron popover-close
                                                  :class="{ 'list-item-selected': query.categoryIds === subCategory.id, 'item-in-multiple-selection': queryAllFilterCategoryIdsCount > 1 && queryAllFilterCategoryIds[subCategory.id] }"
                                                  :title="subCategory.name"
                                                  :key="subCategory.id"
                                                  v-for="subCategory in category.subCategories"
                                                  v-show="!subCategory.hidden || queryAllFilterCategoryIds[subCategory.id]"
                                                  @click="changeCategoryFilter(subCategory.id)"
                                    >
                                        <template #media>
                                            <ItemIcon :icon-type="getCategoryIconType(subCategory.iconType)" :icon-id="subCategory.icon" :color="subCategory.color"></ItemIcon>
                                        </template>
                                        <template #after>
                                            <f7-icon class="list-item-checked-icon"
                                                     f7="checkmark_alt"
                                                     v-if="query.categoryIds === subCategory.id">
                                            </f7-icon>
                                        </template>
                                    </f7-list-item>
                                </f7-list>
                            </f7-accordion-content>
                        </f7-list-item>
                    </template>
                </template>
            </f7-list>
        </f7-popover>

        <f7-popover class="account-popover-menu" @popover:open="onPopoverOpen">
            <f7-list dividers>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': !query.accountIds }"
                              :title="tt('All')"
                              @click="changeAccountFilter('')">
                    <template #media>
                        <f7-icon f7="rectangle_grid_2x2"></f7-icon>
                    </template>
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="!query.accountIds"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.accountIds && queryAllFilterAccountIdsCount > 1 }"
                              :title="tt('Multiple Accounts')"
                              @click="filterMultipleAccounts()"
                              v-if="allAvailableAccountsCount > 0">
                    <template #media>
                        <f7-icon f7="rectangle_on_rectangle"></f7-icon>
                    </template>
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.accountIds && queryAllFilterAccountIdsCount > 1"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.accountIds === account.id, 'item-in-multiple-selection': queryAllFilterAccountIdsCount > 1 && queryAllFilterAccountIds[account.id] }"
                              :title="account.name"
                              :key="account.id"
                              v-for="account in allAccounts"
                              v-show="(!account.hidden && (!allAccountsMap[account.parentId] || !allAccountsMap[account.parentId]!.hidden)) || queryAllFilterAccountIds[account.id]"
                              @click="changeAccountFilter(account.id)"
                >
                    <template #media>
                        <ItemIcon :icon-type="getAccountIconType(account.iconType)" :icon-id="account.icon" :color="account.color"></ItemIcon>
                    </template>
                    <template #after>
                        <f7-icon class="list-item-checked-icon"
                                 f7="checkmark_alt"
                                 v-if="query.accountIds === account.id">
                        </f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <f7-popover class="more-popover-menu">
            <f7-list dividers>
                <f7-list-item group-title>
                    <small>{{ tt('Type') }}</small>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.type === 0 }"
                              :title="tt('All')"
                              @click="changeTypeFilter(0)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.type === 0"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.type === 1 }"
                              :title="tt('Modify Balance')"
                              @click="changeTypeFilter(1)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.type === 1"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.type === 2 }"
                              :title="tt('Income')"
                              @click="changeTypeFilter(2)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.type === 2"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.type === 3 }"
                              :title="tt('Expense')"
                              @click="changeTypeFilter(3)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.type === 3"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.type === 4 }"
                              :title="tt('Transfer')"
                              @click="changeTypeFilter(4)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.type === 4"></f7-icon>
                    </template>
                </f7-list-item>

                <f7-list-item group-title>
                    <small>{{ tt('Amount') }}</small>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': !query.amountFilter }"
                              :title="tt('All')"
                              @click="changeAmountFilter('')">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="!query.amountFilter"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :key="filterType.type"
                              :class="{ 'list-item-selected': query.amountFilter && query.amountFilter.startsWith(`${filterType.type}:`) }"
                              :title="tt(filterType.name)"
                              v-for="filterType in AmountFilterType.values()"
                              @click="changeAmountFilter(filterType.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.amountFilter && query.amountFilter.startsWith(`${filterType.type}:`)"></f7-icon>
                    </template>
                    <template #footer>
                        <span class="margin-inline-end-half" v-if="query.amountFilter && query.amountFilter.startsWith(`${filterType.type}:`)">{{ queryAmount }}</span>
                    </template>
                </f7-list-item>

                <f7-list-item group-title>
                    <small>{{ tt('Tags') }}</small>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': !query.tagFilter }"
                              :title="tt('All')"
                              @click="changeTagFilter('')">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="!query.tagFilter"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.tagFilter === TransactionTagFilter.TransactionNoTagFilterValue }"
                              :title="tt('Without Tags')"
                              @click="changeTagFilter(TransactionTagFilter.TransactionNoTagFilterValue)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.tagFilter === TransactionTagFilter.TransactionNoTagFilterValue"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :class="{ 'list-item-selected': query.tagFilter && queryAllFilterTagIdsCount > 1 }"
                              :title="tt('Multiple Tags')" @click="filterMultipleTags()" v-if="allAvailableTagsCount > 0">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="query.tagFilter && queryAllFilterTagIdsCount > 1"></f7-icon>
                    </template>
                </f7-list-item>

                <template :key="transactionTagGroup.id"
                          v-for="transactionTagGroup in allTransactionTagGroupsWithDefault">
                    <f7-list-item group-title class="transaction-tag-group" v-if="allTransactionTagsByGroup[transactionTagGroup.id] && allTransactionTagsByGroup[transactionTagGroup.id]?.length && hasVisibleTagsInTagGroup(transactionTagGroup)">
                        <small>{{ transactionTagGroup.name }}</small>
                    </f7-list-item>

                    <f7-list-item link="#" no-chevron popover-close
                                  :title="transactionTag.name"
                                  :class="{ 'list-item-selected': queryAllFilterTagIdsCount === 1 && isDefined(queryAllFilterTagIds[transactionTag.id]), 'item-in-multiple-selection': queryAllFilterTagIdsCount > 1 && isDefined(queryAllFilterTagIds[transactionTag.id]) }"
                                  :key="transactionTag.id"
                                  v-for="transactionTag in (allTransactionTagsByGroup[transactionTagGroup.id] ?? [])"
                                  v-show="!transactionTag.hidden || isDefined(queryAllFilterTagIds[transactionTag.id])"
                                  @click="changeTagFilter(TransactionTagFilter.of(transactionTag.id).toTextualTagFilter())"
                    >
                        <template #before-title>
                            <f7-icon class="transaction-tag-name transaction-tag-icon" f7="number"></f7-icon>
                        </template>
                        <template #after>
                            <f7-icon class="list-item-checked-icon"
                                     :f7="queryAllFilterTagIds[transactionTag.id] === true ? 'checkmark_alt' : (queryAllFilterTagIds[transactionTag.id] === false ? 'multiply' : undefined)"
                                     v-if="isDefined(queryAllFilterTagIds[transactionTag.id])">
                            </f7-icon>
                        </template>
                    </f7-list-item>
                </template>
            </f7-list>
        </f7-popover>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteActionSheet" @actions:closed="showDeleteActionSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this transaction?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="remove(transactionToDelete, true)">{{ tt('Delete') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>



<style src="./ListPage/list.css"></style>
