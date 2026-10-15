<script setup lang="ts">
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import type { useListPage } from '@/views/mobile/transactions/ListPage/useListPage.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useListPage>>;

defineProps<{
    loading: Scope['loading'];
    pageType: Scope['pageType'];
    showSearchbar: Scope['showSearchbar'];
    showTotalAmountInTransactionListPage: Scope['showTotalAmountInTransactionListPage'];
}>();

</script>

<template>
    <div class="skeleton-text" >
        <f7-block class="combination-list-wrapper"
                  :class="{ 'margin-vertical': blockIdx > 1 || pageType === TransactionListPageType.Calendar.type || showSearchbar, 'margin-vertical-half': blockIdx === 1 && pageType !== TransactionListPageType.Calendar.type && !showSearchbar, 'no-accordion-toggle': pageType !== TransactionListPageType.List.type && pageType !== TransactionListPageType.Gallery.type }"
                  :key="blockIdx" v-for="blockIdx in (pageType === TransactionListPageType.List.type ? [ 1, 2 ] : [ 1 ])">
            <f7-accordion-item>
                <f7-block-title v-if="pageType === TransactionListPageType.List.type || pageType === TransactionListPageType.Gallery.type">
                    <f7-accordion-toggle>
                        <f7-list strong inset dividers media-list
                                 class="transaction-amount-list combination-list-header combination-list-opened">
                            <f7-list-item>
                                <template #title>
                                    <small>YYYY-MM</small>
                                    <small class="transaction-amount-statistics" v-if="showTotalAmountInTransactionListPage">
                                        <span>0.00 USD</span>
                                        <span>0.00 USD</span>
                                    </small>
                                    <f7-icon class="combination-list-chevron-icon" f7="chevron_up"></f7-icon>
                                </template>
                            </f7-list-item>
                        </f7-list>
                    </f7-accordion-toggle>
                </f7-block-title>
                <f7-accordion-content style="height: auto">
                    <f7-list strong inset dividers media-list accordion-list class="transaction-info-list combination-list-content"
                             v-if="pageType === TransactionListPageType.List.type || pageType === TransactionListPageType.Calendar.type">
                        <f7-list-item link="#" chevron-center class="transaction-info"
                                      :key="itemIdx" v-for="itemIdx in (pageType === TransactionListPageType.List.type && blockIdx === 1 ? [ 1, 2, 3, 4, 5, 6, 7 ] : [ 1, 2, 3 ])">
                            <template #media>
                                <div class="display-flex flex-direction-column transaction-date">
                                    <span class="transaction-day width-100 flex-direction-column">DD</span>
                                    <span class="transaction-day-of-week width-100 flex-direction-column">Sun</span>
                                </div>
                            </template>
                            <template #inner>
                                <div class="display-flex no-padding-horizontal">
                                    <div class="item-media">
                                        <div class="transaction-icon display-flex align-items-center">
                                            <f7-icon f7="app_fill"></f7-icon>
                                        </div>
                                    </div>
                                    <div class="actual-item-inner">
                                        <div class="item-title-row">
                                            <div class="item-title">
                                                <div class="transaction-category-name no-padding">
                                                    <span>Category</span>
                                                </div>
                                            </div>
                                            <div class="item-after">
                                                <div class="transaction-amount">
                                                    <span>0.00 USD</span>
                                                </div>
                                            </div>
                                        </div>
                                        <div class="item-text">
                                            <div class="transaction-description">
                                                <span>Transaction Description</span>
                                            </div>
                                        </div>
                                        <div class="item-footer">
                                            <div class="transaction-footer">
                                                <span>HH:mm</span>
                                                <span>·</span>
                                                <span>Source Account</span>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </template>
                        </f7-list-item>
                    </f7-list>
                    <f7-list strong inset dividers media-list accordion-list class="transaction-info-list combination-list-content transaction-gallery-list"
                             v-if="pageType === TransactionListPageType.Gallery.type">
                        <f7-list-item class="transaction-gallery-container">
                            <template #default>
                                <div class="transaction-gallery-grid">
                                    <div class="transaction-picture-cell"
                                         :key="itemIdx" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]">
                                        <div class="display-flex justify-content-center align-items-center" style="height: 100%">
                                            <f7-skeleton-block class="transaction-picture-img" style="width: 40px; height: 40px; border-radius: 50%" />
                                        </div>
                                    </div>
                                </div>
                            </template>
                        </f7-list-item>
                    </f7-list>
                </f7-accordion-content>
            </f7-accordion-item>
        </f7-block>
    </div>
</template>
