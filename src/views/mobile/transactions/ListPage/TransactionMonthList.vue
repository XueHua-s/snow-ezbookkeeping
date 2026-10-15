<script setup lang="ts">
import { TextDirection } from '@/core/text.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    getCategoryIconType
} from '@/lib/icon.ts';
import { allTransactionPictures } from '@/lib/transaction.ts';
import { useI18n } from '@/locales/helpers.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import type { useListPage } from '@/views/mobile/transactions/ListPage/useListPage.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useListPage>>;

defineProps<{
    pageType: Scope['pageType'];
    showSearchbar: Scope['showSearchbar'];
    transactions: Scope['transactions'];
    getTransactionMonthTitleDomId: Scope['getTransactionMonthTitleDomId'];
    getDisplayLongYearMonth: Scope['getDisplayLongYearMonth'];
    showTotalAmountInTransactionListPage: Scope['showTotalAmountInTransactionListPage'];
    hasMonthTotalAmount: Scope['hasMonthTotalAmount'];
    getDisplayMonthTotalAmount: Scope['getDisplayMonthTotalAmount'];
    getTransactionMonthListHeight: Scope['getTransactionMonthListHeight'];
    isTransactionMonthListInvisible: Scope['isTransactionMonthListInvisible'];
    getTransactionMonthListDomId: Scope['getTransactionMonthListDomId'];
    getTransactionDomId: Scope['getTransactionDomId'];
    getTransactionDateStyle: Scope['getTransactionDateStyle'];
    getTransactionTypeName: Scope['getTransactionTypeName'];
    getDisplayAmount: Scope['getDisplayAmount'];
    showTagInTransactionListPage: Scope['showTagInTransactionListPage'];
    allTransactionTags: Scope['allTransactionTags'];
    getDisplayTime: Scope['getDisplayTime'];
    isSameAsDefaultTimezoneOffsetMinutes: Scope['isSameAsDefaultTimezoneOffsetMinutes'];
    getDisplayTimezone: Scope['getDisplayTimezone'];
    textDirection: Scope['textDirection'];
    getTransactionPictureUrl: Scope['getTransactionPictureUrl'];
}>();

const emit = defineEmits<{
    collapseTransactionMonthList: Parameters<Scope['collapseTransactionMonthList']>;
    onTransactionMonthListCollapseStateChanged: Parameters<Scope['onTransactionMonthListCollapseStateChanged']>;
    duplicate: Parameters<Scope['duplicate']>;
    edit: Parameters<Scope['edit']>;
    remove: Parameters<Scope['remove']>;
}>();

const collapseTransactionMonthList = (...args: Parameters<Scope['collapseTransactionMonthList']>) => emit('collapseTransactionMonthList', ...args);
const onTransactionMonthListCollapseStateChanged = (...args: Parameters<Scope['onTransactionMonthListCollapseStateChanged']>) => emit('onTransactionMonthListCollapseStateChanged', ...args);
const duplicate = (...args: Parameters<Scope['duplicate']>) => emit('duplicate', ...args);
const edit = (...args: Parameters<Scope['edit']>) => emit('edit', ...args);
const remove = (...args: Parameters<Scope['remove']>) => emit('remove', ...args);

const { formatNumberToLocalizedNumeralsWithoutDigitGrouping, getWeekdayShortName, tt } = useI18n();

</script>

<template>
    <f7-block class="combination-list-wrapper"
              :class="{ 'margin-vertical': index > 0 || pageType === TransactionListPageType.Calendar.type || showSearchbar, 'margin-vertical-half': index === 0 && pageType !== TransactionListPageType.Calendar.type && !showSearchbar, 'no-accordion-toggle': pageType !== TransactionListPageType.List.type && pageType !== TransactionListPageType.Gallery.type }"
              :key="transactionMonthList.yearDashMonth" v-for="(transactionMonthList, index) in transactions">
        <f7-accordion-item :opened="transactionMonthList.opened"
                           @accordion:open="collapseTransactionMonthList(transactionMonthList, false)"
                           @accordion:opened="onTransactionMonthListCollapseStateChanged"
                           @accordion:close="collapseTransactionMonthList(transactionMonthList, true)"
                           @accordion:closed="onTransactionMonthListCollapseStateChanged">
            <f7-block-title :id="getTransactionMonthTitleDomId(transactionMonthList.yearDashMonth)" v-if="pageType === TransactionListPageType.List.type || pageType === TransactionListPageType.Gallery.type">
                <f7-accordion-toggle>
                    <f7-list strong inset dividers media-list
                             class="transaction-amount-list combination-list-header"
                             :class="transactionMonthList.opened ? 'combination-list-opened' : 'combination-list-closed'">
                        <f7-list-item>
                            <template #title>
                                <small>
                                    <span>{{ getDisplayLongYearMonth(transactionMonthList) }}</span>
                                </small>
                                <small class="transaction-amount-statistics" v-if="showTotalAmountInTransactionListPage && hasMonthTotalAmount(transactionMonthList)">
                                    <span class="text-income">
                                        {{ getDisplayMonthTotalAmount(transactionMonthList, '+')?.income }}
                                    </span>
                                    <span class="text-expense">
                                        {{ getDisplayMonthTotalAmount(transactionMonthList, '-')?.expense }}
                                    </span>
                                </small>
                                <f7-icon class="combination-list-chevron-icon" :f7="transactionMonthList.opened ? 'chevron_up' : 'chevron_down'"></f7-icon>
                            </template>
                        </f7-list-item>
                    </f7-list>
                </f7-accordion-toggle>
            </f7-block-title>
            <f7-accordion-content>
                <f7-block :style="{ height: getTransactionMonthListHeight(transactionMonthList) }"
                          v-if="isTransactionMonthListInvisible(transactionMonthList)" />
                <f7-list strong inset dividers media-list accordion-list
                         class="transaction-info-list transaction-month-list combination-list-content"
                         :id="getTransactionMonthListDomId(transactionMonthList.yearDashMonth)"
                         v-if="!isTransactionMonthListInvisible(transactionMonthList) && (pageType === TransactionListPageType.List.type || pageType === TransactionListPageType.Calendar.type)"
                >
                    <f7-list-item swipeout chevron-center accordion-item
                                  class="transaction-info"
                                  :id="getTransactionDomId(transaction)"
                                  :link="`/transaction/detail?id=${transaction.id}&type=${transaction.type}`"
                                  :key="transaction.id"
                                  v-for="(transaction, idx) in transactionMonthList.items"
                    >
                        <template #media>
                            <div class="display-flex flex-direction-column transaction-date" :style="getTransactionDateStyle(transaction, idx > 0 ? transactionMonthList.items[idx - 1] : undefined)">
                                <span class="transaction-day width-100 flex-direction-column">
                                    {{ transaction.gregorianCalendarDayOfMonth ? formatNumberToLocalizedNumeralsWithoutDigitGrouping(transaction.gregorianCalendarDayOfMonth) : '' }}
                                </span>
                                <span class="transaction-day-of-week width-100 flex-direction-column" v-if="transaction.displayDayOfWeek">
                                    {{ getWeekdayShortName(transaction.displayDayOfWeek) }}
                                </span>
                            </div>
                        </template>
                        <template #inner>
                            <div class="display-flex no-padding-horizontal">
                                <div class="item-media">
                                    <div class="transaction-icon display-flex align-items-center">
                                        <ItemIcon :icon-type="getCategoryIconType(transaction.category.iconType)"
                                                  :icon-id="transaction.category.icon"
                                                  :color="transaction.category.color"
                                                  v-if="transaction.category && transaction.category.color"></ItemIcon>
                                        <f7-icon v-else-if="!transaction.category || !transaction.category.color"
                                                 f7="pencil_ellipsis_rectangle">
                                        </f7-icon>
                                    </div>
                                </div>
                                <div class="actual-item-inner">
                                    <div class="item-title-row">
                                        <div class="item-title">
                                            <div class="transaction-category-name no-padding">
                                                <span v-if="transaction.type === TransactionType.ModifyBalance">
                                                    {{ tt('Modify Balance') }}
                                                </span>
                                                    <span v-else-if="transaction.type !== TransactionType.ModifyBalance && transaction.category">
                                                    {{ transaction.category.name }}
                                                </span>
                                                    <span v-else-if="transaction.type !== TransactionType.ModifyBalance && !transaction.category">
                                                    {{ getTransactionTypeName(transaction.type, 'Transaction') }}
                                                </span>
                                            </div>
                                        </div>
                                        <div class="item-after">
                                            <div class="transaction-amount" v-if="transaction.sourceAccount"
                                                 :class="{ 'text-expense': transaction.type === TransactionType.Expense, 'text-income': transaction.type === TransactionType.Income }">
                                                <span>{{ getDisplayAmount(transaction) }}</span>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="item-text">
                                        <div class="transaction-description" v-if="transaction.comment">
                                            <span>{{ transaction.comment }}</span>
                                        </div>
                                    </div>
                                    <div class="item-footer">
                                        <div class="transaction-tags" v-if="showTagInTransactionListPage && transaction.tagIds && transaction.tagIds.length">
                                            <f7-chip media-text-color="var(--f7-chip-text-color)" class="transaction-tag"
                                                     :text="allTransactionTags[tagId]?.name"
                                                     :key="tagId"
                                                     v-for="tagId in transaction.tagIds">
                                                <template #media>
                                                    <f7-icon f7="number"></f7-icon>
                                                </template>
                                            </f7-chip>
                                        </div>
                                        <div class="transaction-footer">
                                            <span>{{ getDisplayTime(transaction) }}</span>
                                            <span v-if="!isSameAsDefaultTimezoneOffsetMinutes(transaction)">{{ `(${getDisplayTimezone(transaction)})` }}</span>
                                            <span v-if="transaction.sourceAccount">·</span>
                                            <span v-if="transaction.sourceAccount">{{ transaction.sourceAccount.name }}</span>
                                            <f7-icon class="transaction-account-arrow icon-with-direction" f7="arrow_right" v-if="transaction.sourceAccount && transaction.type === TransactionType.Transfer && transaction.destinationAccount && transaction.sourceAccount.id !== transaction.destinationAccount.id"></f7-icon>
                                            <span v-if="transaction.sourceAccount && transaction.type === TransactionType.Transfer && transaction.destinationAccount && transaction.sourceAccount.id !== transaction.destinationAccount.id">{{ transaction.destinationAccount.name }}</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </template>
                        <f7-swipeout-actions :left="textDirection === TextDirection.RTL"
                                             :right="textDirection === TextDirection.LTR">
                            <f7-swipeout-button color="primary" close
                                                :text="tt('Duplicate')"
                                                v-if="transaction.type !== TransactionType.ModifyBalance"
                                                @click="duplicate(transaction)"></f7-swipeout-button>
                            <f7-swipeout-button color="orange" close
                                                :text="tt('Edit')"
                                                v-if="transaction.editable"
                                                @click="edit(transaction)"></f7-swipeout-button>
                            <f7-swipeout-button color="red" class="padding-horizontal"
                                                :aria-label="tt('Delete')"
                                                v-if="transaction.editable"
                                                @click="remove(transaction, false)">
                                <f7-icon f7="trash"></f7-icon>
                            </f7-swipeout-button>
                        </f7-swipeout-actions>
                    </f7-list-item>
                </f7-list>
                <f7-list strong inset dividers media-list accordion-list
                         class="transaction-info-list transaction-month-list combination-list-content transaction-gallery-list"
                         :id="getTransactionMonthListDomId(transactionMonthList.yearDashMonth)"
                         v-if="!isTransactionMonthListInvisible(transactionMonthList) && (pageType === TransactionListPageType.Gallery.type)">
                    <f7-list-item class="transaction-gallery-container">
                        <template #default>
                            <div class="transaction-gallery-grid">
                                <div class="transaction-picture-cell" :key="pictureInfo.pictureId"
                                     v-for="[transaction, pictureInfo] in allTransactionPictures(transactionMonthList.items)">
                                    <image-box class="transaction-picture-img" alt="picture"
                                               :link="`/transaction/detail?id=${transaction.id}&type=${transaction.type}`"
                                               :src="getTransactionPictureUrl(pictureInfo)">
                                        <template #error>
                                            {{ tt('Failed to load image, please check whether the config "domain" and "root_url" are set correctly.') }}
                                        </template>
                                    </image-box>
                                </div>
                            </div>
                        </template>
                    </f7-list-item>
                </f7-list>
            </f7-accordion-content>
        </f7-accordion-item>
    </f7-block>
</template>
