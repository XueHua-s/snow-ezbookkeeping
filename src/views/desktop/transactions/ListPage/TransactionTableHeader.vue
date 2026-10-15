<script setup lang="ts">
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
import { useI18n } from '@/locales/helpers.ts';
import { TransactionTagFilter } from '@/models/transaction.ts';
import type { useListPage } from '@/views/desktop/transactions/ListPage/useListPage.ts';
import {
    mdiBorderNoneVariant,
    mdiCheck,
    mdiClose,
    mdiMenuDown,
    mdiPound,
    mdiVectorArrangeBelow,
    mdiViewGridOutline
} from '@mdi/js';
import type { ShallowUnwrapRef } from 'vue';
import { useTransactionFilterMenus } from './useTransactionFilterMenus.ts';

type Scope = ShallowUnwrapRef<ReturnType<typeof useListPage>>;

const props = defineProps<{
    loading: Scope['loading'];
    query: Scope['query'];
    allDateRanges: Scope['allDateRanges'];
    queryMinTime: Scope['queryMinTime'];
    queryMaxTime: Scope['queryMaxTime'];
    queryCategoryName: Scope['queryCategoryName'];
    queryAllSelectedFilterCategoryIds: Scope['queryAllSelectedFilterCategoryIds'];
    queryAllFilterCategoryIdsCount: Scope['queryAllFilterCategoryIdsCount'];
    allAvailableCategoriesCount: Scope['allAvailableCategoriesCount'];
    allPrimaryCategories: Scope['allPrimaryCategories'];
    getTransactionTypeName: Scope['getTransactionTypeName'];
    queryAllFilterCategoryIds: Scope['queryAllFilterCategoryIds'];
    allCategories: Scope['allCategories'];
    hasSubCategoryInQuery: Scope['hasSubCategoryInQuery'];
    getCategoryListItemCheckedClass: Scope['getCategoryListItemCheckedClass'];
    queryAmount: Scope['queryAmount'];
    selectedAccountDefaultCurrency: Scope['selectedAccountDefaultCurrency'];
    queryAccountName: Scope['queryAccountName'];
    queryAllSelectedFilterAccountIds: Scope['queryAllSelectedFilterAccountIds'];
    queryAllFilterAccountIdsCount: Scope['queryAllFilterAccountIdsCount'];
    allAvailableAccountsCount: Scope['allAvailableAccountsCount'];
    allAccounts: Scope['allAccounts'];
    allAccountsMap: Scope['allAccountsMap'];
    queryAllFilterAccountIds: Scope['queryAllFilterAccountIds'];
    showTagInTransactionListPage: Scope['showTagInTransactionListPage'];
    queryTagName: Scope['queryTagName'];
    queryAllSelectedFilterTagIds: Scope['queryAllSelectedFilterTagIds'];
    queryAllFilterTagIdsCount: Scope['queryAllFilterTagIdsCount'];
    allAvailableTagsCount: Scope['allAvailableTagsCount'];
    allTransactionTagGroupsWithDefault: Scope['allTransactionTagGroupsWithDefault'];
    allTransactionTagsByGroup: Scope['allTransactionTagsByGroup'];
    hasVisibleTagsInTagGroup: Scope['hasVisibleTagsInTagGroup'];
    queryAllFilterTagIds: Scope['queryAllFilterTagIds'];
}>();

const categoryMenuState = defineModel<Scope['categoryMenuState']>('categoryMenuState', { required: true });
const amountMenuState = defineModel<Scope['amountMenuState']>('amountMenuState', { required: true });
const currentAmountFilterType = defineModel<Scope['currentAmountFilterType']>('currentAmountFilterType', { required: true });
const currentAmountFilterValue1 = defineModel<Scope['currentAmountFilterValue1']>('currentAmountFilterValue1', { required: true });
const currentAmountFilterValue2 = defineModel<Scope['currentAmountFilterValue2']>('currentAmountFilterValue2', { required: true });
const showFilterCategoryDialog = defineModel<Scope['showFilterCategoryDialog']>('showFilterCategoryDialog', { required: true });
const showFilterAccountDialog = defineModel<Scope['showFilterAccountDialog']>('showFilterAccountDialog', { required: true });
const showFilterTagDialog = defineModel<Scope['showFilterTagDialog']>('showFilterTagDialog', { required: true });

const emit = defineEmits<{
    changeDateFilter: Parameters<Scope['changeDateFilter']>;
    changeCategoryFilter: Parameters<Scope['changeCategoryFilter']>;
    changeAmountFilter: Parameters<Scope['changeAmountFilter']>;
    changeAccountFilter: Parameters<Scope['changeAccountFilter']>;
    changeTagFilter: Parameters<Scope['changeTagFilter']>;
}>();

const changeDateFilter = (...args: Parameters<Scope['changeDateFilter']>) => emit('changeDateFilter', ...args);
const changeCategoryFilter = (...args: Parameters<Scope['changeCategoryFilter']>) => emit('changeCategoryFilter', ...args);
const changeAmountFilter = (...args: Parameters<Scope['changeAmountFilter']>) => emit('changeAmountFilter', ...args);
const changeAccountFilter = (...args: Parameters<Scope['changeAccountFilter']>) => emit('changeAccountFilter', ...args);
const changeTagFilter = (...args: Parameters<Scope['changeTagFilter']>) => emit('changeTagFilter', ...args);

const { tt } = useI18n();

const {
    scrollTimeMenuToSelectedItem, scrollCategoryMenuToSelectedItem, scrollAmountMenuToSelectedItem,
    scrollAccountMenuToSelectedItem, scrollTagMenuToSelectedItem
} = useTransactionFilterMenus(
    () => props.query.amountFilter, currentAmountFilterType, currentAmountFilterValue1, currentAmountFilterValue2
);

</script>

<template>
    <tr>
        <th class="transaction-table-column-time text-no-wrap">
            <v-menu ref="timeFilterMenu" class="transaction-time-menu"
                    eager location="bottom" max-height="500"
                    @update:model-value="scrollTimeMenuToSelectedItem">
                <template #activator="{ props }">
                    <div class="d-flex align-center cursor-pointer"
                         :class="{ 'readonly': loading, 'text-primary': query.dateType !== DateRange.ThisMonth.type }" v-bind="props">
                        <span>{{ tt('Time') }}</span>
                        <v-icon :icon="mdiMenuDown" />
                    </div>
                </template>
                <v-list :selected="[query.dateType]">
                    <v-list-item class="text-body-medium" density="compact"
                                 :key="dateRange.type" :value="dateRange.type"
                                 :class="{ 'list-item-selected': query.dateType === dateRange.type }"
                                 :append-icon="(query.dateType === dateRange.type ? mdiCheck : undefined)"
                                 v-for="dateRange in allDateRanges">
                        <v-list-item-title class="cursor-pointer"
                                           @click="changeDateFilter(dateRange.type)">
                            <div class="d-flex align-center">
                                <span class="text-body-medium ms-3">{{ dateRange.displayName }}</span>
                            </div>
                            <div class="transaction-list-custom-datetime-range ms-3 smaller" v-if="dateRange.isUserCustomRange && query.dateType === dateRange.type && query.minTime && query.maxTime">
                                <span>{{ queryMinTime }}</span>
                                <span>&nbsp;-&nbsp;</span>
                                <br/>
                                <span>{{ queryMaxTime }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>
                </v-list>
            </v-menu>
        </th>
        <th class="transaction-table-column-category text-no-wrap">
            <v-menu ref="categoryFilterMenu" class="transaction-category-menu"
                    eager location="bottom" max-height="500"
                    :disabled="query.type === 1"
                    :close-on-content-click="false"
                    v-model="categoryMenuState"
                    @update:model-value="scrollCategoryMenuToSelectedItem">
                <template #activator="{ props }">
                    <div class="d-flex align-center"
                        :class="{ 'readonly': loading, 'cursor-pointer': query.type !== 1, 'text-primary': query.categoryIds }" v-bind="props">
                        <span>{{ queryCategoryName }}</span>
                        <v-icon :icon="mdiMenuDown" v-show="query.type !== 1" />
                    </div>
                </template>
                <v-list :selected="[queryAllSelectedFilterCategoryIds]">
                    <v-list-item key="" value="" class="text-body-medium" density="compact"
                                 :class="{ 'list-item-selected': !query.categoryIds }"
                                 :append-icon="(!query.categoryIds ? mdiCheck : undefined)">
                        <v-list-item-title class="cursor-pointer"
                                           @click="changeCategoryFilter('')">
                            <div class="d-flex align-center">
                                <v-icon :icon="mdiViewGridOutline" />
                                <span class="text-body-medium ms-3">{{ tt('All') }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>
                    <v-list-item key="multiple" value="multiple" class="text-body-medium" density="compact"
                                 :class="{ 'list-item-selected': query.categoryIds && queryAllFilterCategoryIdsCount > 1 }"
                                 :append-icon="(query.categoryIds && queryAllFilterCategoryIdsCount > 1 ? mdiCheck : undefined)"
                                 v-if="allAvailableCategoriesCount > 0">
                        <v-list-item-title class="cursor-pointer"
                                           @click="showFilterCategoryDialog = true">
                            <div class="d-flex align-center">
                                <v-icon :icon="mdiVectorArrangeBelow" />
                                <span class="text-body-medium ms-3">{{ tt('Multiple Categories') }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>

                    <template :key="categoryType"
                              v-for="(categories, categoryType) in allPrimaryCategories">
                        <v-divider />

                        <v-list-item density="compact" v-show="categories && categories.length">
                            <v-list-item-title>
                                <span class="text-body-small">{{ getTransactionTypeName(categoryTypeToTransactionType(parseInt(categoryType)), 'Type') }}</span>
                            </v-list-item-title>
                        </v-list-item>

                        <v-list-group :key="category.id" v-for="(category, index) in categories">
                            <template #activator="{ props }" v-if="!category.hidden || queryAllFilterCategoryIds[category.id] || allCategories[query.categoryIds]?.parentId === category.id || hasSubCategoryInQuery(category)">
                                <v-divider v-if="index > 0" />
                                <v-list-item class="text-body-medium" density="compact"
                                             :class="getCategoryListItemCheckedClass(category, queryAllFilterCategoryIds)"
                                             v-bind="props">
                                    <v-list-item-title>
                                        <div class="d-flex align-center">
                                            <ItemIcon :icon-type="getCategoryIconType(category.iconType)" size="24px" :icon-id="category.icon" :color="category.color"></ItemIcon>
                                            <span class="text-body-medium ms-2">{{ category.name }}</span>
                                        </div>
                                    </v-list-item-title>
                                </v-list-item>
                            </template>

                            <v-divider />
                            <v-list-item class="text-body-medium" density="compact"
                                         :class="{ 'item-in-multiple-selection': queryAllFilterCategoryIdsCount > 1 && queryAllFilterCategoryIds[category.id] }"
                                         :value="category.id"
                                         :append-icon="(query.categoryIds === category.id ? mdiCheck : undefined)">
                                <v-list-item-title class="cursor-pointer"
                                                   @click="changeCategoryFilter(category.id)">
                                    <div class="d-flex align-center">
                                        <v-icon :icon="mdiViewGridOutline" />
                                        <span class="text-body-medium ms-3">{{ tt('All') }}</span>
                                    </div>
                                </v-list-item-title>
                            </v-list-item>

                            <template :key="subCategory.id"
                                      v-for="subCategory in category.subCategories">
                                <v-divider v-if="!subCategory.hidden || queryAllFilterCategoryIds[subCategory.id]" />
                                <v-list-item class="text-body-medium" density="compact"
                                             :value="subCategory.id"
                                             :class="{ 'list-item-selected': query.categoryIds === subCategory.id, 'item-in-multiple-selection': queryAllFilterCategoryIdsCount > 1 && queryAllFilterCategoryIds[subCategory.id] }"
                                             :append-icon="(query.categoryIds === subCategory.id ? mdiCheck : undefined)"
                                             v-if="!subCategory.hidden || queryAllFilterCategoryIds[subCategory.id]">
                                    <v-list-item-title class="cursor-pointer"
                                                       @click="changeCategoryFilter(subCategory.id)">
                                        <div class="d-flex align-center">
                                            <ItemIcon :icon-type="getCategoryIconType(subCategory.iconType)" size="24px" :icon-id="subCategory.icon" :color="subCategory.color"></ItemIcon>
                                            <span class="text-body-medium ms-2">{{ subCategory.name }}</span>
                                        </div>
                                    </v-list-item-title>
                                </v-list-item>
                            </template>
                        </v-list-group>
                    </template>
                </v-list>
            </v-menu>
        </th>
        <th class="transaction-table-column-amount text-no-wrap">
            <v-menu ref="amountFilterMenu" class="transaction-amount-menu"
                    eager location="bottom" max-height="500"
                    :close-on-content-click="false"
                    v-model="amountMenuState"
                    @update:model-value="scrollAmountMenuToSelectedItem">
                <template #activator="{ props }">
                    <div class="d-flex align-center cursor-pointer"
                         :class="{ 'readonly': loading, 'text-primary': query.amountFilter }" v-bind="props">
                        <span>{{ tt('Amount') }}</span>
                        <v-icon :icon="mdiMenuDown" />
                    </div>
                </template>
                <v-list :selected="[query.amountFilter.split(':')[0]]">
                    <v-list-item key="" value="" class="text-body-medium" density="compact"
                                 :class="{ 'list-item-selected': !query.amountFilter }"
                                 :append-icon="(!query.amountFilter && !currentAmountFilterType ? mdiCheck : undefined)">
                        <v-list-item-title class="cursor-pointer"
                                           @click="changeAmountFilter('')">
                            <div class="d-flex align-center">
                                <span class="text-body-medium ms-3">{{ tt('All') }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>
                    <template :key="filterType.type"
                              v-for="filterType in AmountFilterType.values()">
                        <v-list-item class="text-body-medium" density="compact"
                                     :value="filterType.type"
                                     :class="{ 'list-item-selected': query.amountFilter && query.amountFilter.startsWith(`${filterType.type}:`) }"
                                     :append-icon="(query.amountFilter && query.amountFilter.startsWith(`${filterType.type}:`) && currentAmountFilterType !== filterType.type ? mdiCheck : undefined)">
                            <v-list-item-title class="cursor-pointer"
                                               @click="currentAmountFilterType = filterType.type">
                                <div class="d-flex align-center">
                                    <span class="text-body-medium ms-3">{{ tt(filterType.name) }}</span>
                                    <span class="text-body-medium ms-3" v-if="query.amountFilter && query.amountFilter.startsWith(`${filterType.type}:`) && currentAmountFilterType !== filterType.type">{{ queryAmount }}</span>
                                    <amount-input class="transaction-amount-filter-value ms-4" density="compact"
                                                  :currency="selectedAccountDefaultCurrency"
                                                  v-model="currentAmountFilterValue1"
                                                  v-if="currentAmountFilterType === filterType.type"/>
                                    <span class="ms-2 me-2" v-if="currentAmountFilterType === filterType.type && filterType.paramCount === 2">{{ tt('format.misc.rangeSeparator') }}</span>
                                    <amount-input class="transaction-amount-filter-value" density="compact"
                                                  :currency="selectedAccountDefaultCurrency"
                                                  v-model="currentAmountFilterValue2"
                                                  v-if="currentAmountFilterType === filterType.type && filterType.paramCount === 2"/>
                                    <v-btn class="ms-2" density="compact" color="primary" variant="tonal"
                                           @click="changeAmountFilter(filterType.type)"
                                           v-if="currentAmountFilterType === filterType.type">{{ tt('Apply') }}</v-btn>
                                </div>
                            </v-list-item-title>
                        </v-list-item>
                    </template>
                </v-list>
            </v-menu>
        </th>
        <th class="transaction-table-column-account text-no-wrap">
            <v-menu ref="accountFilterMenu" class="transaction-account-menu"
                    eager location="bottom" max-height="500"
                    @update:model-value="scrollAccountMenuToSelectedItem">
                <template #activator="{ props }">
                    <div class="d-flex align-center cursor-pointer"
                         :class="{ 'readonly': loading, 'text-primary': query.accountIds }" v-bind="props">
                        <span>{{ queryAccountName }}</span>
                        <v-icon :icon="mdiMenuDown" />
                    </div>
                </template>
                <v-list :selected="[queryAllSelectedFilterAccountIds]">
                    <v-list-item key="" value="" class="text-body-medium" density="compact"
                                 :class="{ 'list-item-selected': !query.accountIds }"
                                 :append-icon="(!query.accountIds ? mdiCheck : undefined)">
                        <v-list-item-title class="cursor-pointer"
                                           @click="changeAccountFilter('')">
                            <div class="d-flex align-center">
                                <v-icon :icon="mdiViewGridOutline" />
                                <span class="text-body-medium ms-3">{{ tt('All') }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>
                    <v-list-item key="multiple" value="multiple" class="text-body-medium" density="compact"
                                 :class="{ 'list-item-selected': query.accountIds && queryAllFilterAccountIdsCount > 1 }"
                                 :append-icon="(query.accountIds && queryAllFilterAccountIdsCount > 1 ? mdiCheck : undefined)"
                                 v-if="allAvailableAccountsCount > 0">
                        <v-list-item-title class="cursor-pointer"
                                           @click="showFilterAccountDialog = true">
                            <div class="d-flex align-center">
                                <v-icon :icon="mdiVectorArrangeBelow" />
                                <span class="text-body-medium ms-3">{{ tt('Multiple Accounts') }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>
                    <template :key="account.id"
                              v-for="account in allAccounts">
                        <v-divider v-if="(!account.hidden && (!allAccountsMap[account.parentId] || !allAccountsMap[account.parentId]!.hidden)) || queryAllFilterAccountIds[account.id]" />
                        <v-list-item class="text-body-medium" density="compact"
                                     :value="account.id"
                                     :class="{ 'list-item-selected': query.accountIds === account.id, 'item-in-multiple-selection': queryAllFilterAccountIdsCount > 1 && queryAllFilterAccountIds[account.id] }"
                                     :append-icon="(query.accountIds === account.id ? mdiCheck : undefined)"
                                     v-if="(!account.hidden && (!allAccountsMap[account.parentId] || !allAccountsMap[account.parentId]!.hidden)) || queryAllFilterAccountIds[account.id]">
                            <v-list-item-title class="cursor-pointer"
                                               @click="changeAccountFilter(account.id)">
                                <div class="d-flex align-center">
                                    <ItemIcon :icon-type="getAccountIconType(account.iconType)" size="24px" :icon-id="account.icon" :color="account.color"></ItemIcon>
                                    <span class="text-body-medium ms-2">{{ account.name }}</span>
                                </div>
                            </v-list-item-title>
                        </v-list-item>
                    </template>
                </v-list>
            </v-menu>
        </th>
        <th class="transaction-table-column-tags text-no-wrap" v-if="showTagInTransactionListPage">
            <v-menu ref="tagFilterMenu" class="transaction-tag-menu"
                    eager location="bottom" max-height="500"
                    @update:model-value="scrollTagMenuToSelectedItem">
                <template #activator="{ props }">
                    <div class="d-flex align-center cursor-pointer"
                         :class="{ 'readonly': loading, 'text-primary': query.tagFilter }" v-bind="props">
                        <span>{{ queryTagName }}</span>
                        <v-icon :icon="mdiMenuDown" />
                    </div>
                </template>
                <v-list :selected="[queryAllSelectedFilterTagIds]">
                    <v-list-item key="" value="" class="text-body-medium" density="compact"
                                 :class="{ 'list-item-selected': !query.tagFilter }"
                                 :append-icon="(!query.tagFilter ? mdiCheck : undefined)">
                        <v-list-item-title class="cursor-pointer"
                                           @click="changeTagFilter('')">
                            <div class="d-flex align-center">
                                <v-icon :icon="mdiViewGridOutline" />
                                <span class="text-body-medium ms-2">{{ tt('All') }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>
                    <v-list-item class="text-body-medium" density="compact"
                                 :key="TransactionTagFilter.TransactionNoTagFilterValue"
                                 :value="TransactionTagFilter.TransactionNoTagFilterValue"
                                 :class="{ 'list-item-selected': query.tagFilter === TransactionTagFilter.TransactionNoTagFilterValue }"
                                 :append-icon="(query.tagFilter === TransactionTagFilter.TransactionNoTagFilterValue ? mdiCheck : undefined)">
                        <v-list-item-title class="cursor-pointer"
                                           @click="changeTagFilter(TransactionTagFilter.TransactionNoTagFilterValue)">
                            <div class="d-flex align-center">
                                <v-icon :icon="mdiBorderNoneVariant" />
                                <span class="text-body-medium ms-2">{{ tt('Without Tags') }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>
                    <v-list-item key="multiple" value="multiple" class="text-body-medium" density="compact"
                                 :class="{ 'list-item-selected': query.tagFilter && queryAllFilterTagIdsCount > 1 }"
                                 :append-icon="(query.tagFilter && queryAllFilterTagIdsCount > 1 ? mdiCheck : undefined)"
                                 v-if="allAvailableTagsCount > 0">
                        <v-list-item-title class="cursor-pointer"
                                           @click="showFilterTagDialog = true">
                            <div class="d-flex align-center">
                                <v-icon :icon="mdiVectorArrangeBelow" />
                                <span class="text-body-medium ms-2">{{ tt('Multiple Tags') }}</span>
                            </div>
                        </v-list-item-title>
                    </v-list-item>

                    <template :key="transactionTagGroup.id"
                              v-for="transactionTagGroup in allTransactionTagGroupsWithDefault">
                        <v-divider v-if="allTransactionTagsByGroup[transactionTagGroup.id] && allTransactionTagsByGroup[transactionTagGroup.id]?.length && hasVisibleTagsInTagGroup(transactionTagGroup)" />

                        <v-list-item density="compact" v-if="allTransactionTagsByGroup[transactionTagGroup.id] && allTransactionTagsByGroup[transactionTagGroup.id]?.length && hasVisibleTagsInTagGroup(transactionTagGroup)">
                            <v-list-item-title>
                                <span class="text-body-small">{{ transactionTagGroup.name }}</span>
                            </v-list-item-title>
                        </v-list-item>

                        <template :key="transactionTag.id"
                                  v-for="(transactionTag, index) in (allTransactionTagsByGroup[transactionTagGroup.id] ?? [])">
                            <v-divider v-if="index > 0 && (!transactionTag.hidden || isDefined(queryAllFilterTagIds[transactionTag.id]))" />
                            <v-list-item class="text-body-medium" density="compact"
                                         :value="transactionTag.id"
                                         :class="{ 'list-item-selected': queryAllFilterTagIdsCount === 1 && isDefined(queryAllFilterTagIds[transactionTag.id]), 'item-in-multiple-selection': queryAllFilterTagIdsCount > 1 && isDefined(queryAllFilterTagIds[transactionTag.id]) }"
                                         :append-icon="(queryAllFilterTagIds[transactionTag.id] === true ? mdiCheck : (queryAllFilterTagIds[transactionTag.id] === false ? mdiClose : undefined))"
                                         v-if="!transactionTag.hidden || isDefined(queryAllFilterTagIds[transactionTag.id])">
                                <v-list-item-title class="cursor-pointer"
                                                   @click="changeTagFilter(TransactionTagFilter.of(transactionTag.id).toTextualTagFilter())">
                                    <div class="d-flex align-center">
                                        <v-icon size="20" :icon="mdiPound"/>
                                        <span class="text-body-medium ms-2">{{ transactionTag.name }}</span>
                                    </div>
                                </v-list-item-title>
                            </v-list-item>
                        </template>
                    </template>
                </v-list>
            </v-menu>
        </th>
        <th class="transaction-table-column-description text-no-wrap">{{ tt('Description') }}</th>
    </tr>
</template>
