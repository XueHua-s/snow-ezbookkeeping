<script setup lang="ts">
import { AccountCategory, AccountType } from '@/core/account.ts';
import { TextDirection } from '@/core/text.ts';
import { getAccountIconType } from '@/lib/icon.ts';
import AssetSummaryWidget from '@/views/mobile/overview/widgets/AssetSummaryWidget.vue';
import type { Router } from 'framework7/types';
import { useListPage } from './ListPage/useListPage.ts';

const props = defineProps<{
    f7router: Router.Router;
}>();

const {
    tt, loading, showHidden, displayOrderModified, showAccountBalance, customAccountCategoryOrder, useLastReconciledTime,
    allCategorizedAccountsMap, allAccountCount, maxCategoryAccountCount, accountCategoryTotalBalance, accountBalance,
    accountBalanceOrAvailableCredit, sortable, accountForMoreActionSheet, accountToDelete, accountToClearTransactions,
    currentPasswordForClearData, clearingData, showAccountMoreActionSheet, showMoreActionSheet, showDeleteActionSheet,
    showInputPasswordSheetForClearAllTransactions, displayOrderSaving, textDirection, showAvailableCreditForCreditCard,
    firstShowingIds, lastShowingIds, hasAnyVisibleAccount, noAvailableAccount, hasAccount, hasVisibleSubAccount,
    getAccountDomId, reload, edit, showMoreActionSheetForAccount, showReconciliationStatement, updateLastReconciledTime,
    moveAllTransactions, showPasswordSheetForClearAllTransaction, clearAllTransactions, hide, remove, setSortable,
    saveSortResult, cancelSort, setAccountsIncludedInTotal, onSort, onPageAfterIn,
} = useListPage(props);

</script>

<template>
    <f7-page :ptr="!sortable" @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')" v-if="!sortable"></f7-nav-left>
            <f7-nav-left v-else-if="sortable">
                <f7-link icon-f7="xmark" :class="{ 'disabled': displayOrderSaving }" :aria-label="tt('Cancel')" @click="cancelSort"></f7-link>
            </f7-nav-left>
            <f7-nav-title :title="tt('Account List')"></f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }">
                <f7-link icon-f7="ellipsis" :class="{ 'disabled': !allAccountCount || sortable }" :aria-label="tt('More')" @click="showMoreActionSheet = true"></f7-link>
                <f7-link icon-f7="plus" href="/account/add" :aria-label="tt('Add')" v-if="!sortable"></f7-link>
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': displayOrderSaving || !displayOrderModified }" :aria-label="tt('Save')" @click="saveSortResult" v-else-if="sortable"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <div class="margin-top-half margin-bottom">
            <asset-summary-widget scene="accountList" :loading="loading" :support-hide-account-balance="true" />
        </div>

        <div class="skeleton-text" v-if="loading">
            <f7-list strong inset dividers sortable class="list-has-group-title account-list margin-vertical"
                     :key="listIdx" v-for="listIdx in [ 1, 2, 3 ]">
                <f7-list-item group-title :sortable="false">
                    <small>
                        <span>Account Category</span>
                        <span style="margin-inline-start: 10px">0.00 USD</span>
                    </small>
                </f7-list-item>
                <f7-list-item class="nested-list-item" after="0.00 USD" link="#"
                              :key="itemIdx" v-for="itemIdx in (listIdx === 1 ? [ 1 ] : [ 1, 2 ])">
                    <template #media>
                        <f7-icon f7="app_fill"></f7-icon>
                    </template>
                    <template #title>
                        <div class="display-flex padding-top-half padding-bottom-half">
                            <div class="nested-list-item-title">
                                <span>Account Name</span>
                            </div>
                        </div>
                    </template>
                </f7-list-item>
            </f7-list>
        </div>

        <f7-list strong inset dividers class="margin-vertical" v-if="!loading && noAvailableAccount">
            <f7-list-item :title="tt('No available account')"></f7-list-item>
        </f7-list>

        <div :key="accountCategory.type"
             v-for="accountCategory in AccountCategory.values(customAccountCategoryOrder)"
             v-show="!loading && ((showHidden && hasAccount(accountCategory, false)) || hasAccount(accountCategory, true))">
            <f7-list strong inset dividers sortable class="list-has-group-title account-list margin-vertical"
                     :sortable-enabled="sortable"
                     v-if="allCategorizedAccountsMap[accountCategory.type]"
                     @sortable:sort="onSort">
                <f7-list-item group-title class="justify-content-space-between" :sortable="false">
                    <small>
                        <span>{{ tt(accountCategory.name) }}</span>
                        <span style="margin-inline-start: 10px">{{ accountCategoryTotalBalance(accountCategory, showAvailableCreditForCreditCard) }}</span>
                    </small>
                    <small class="account-list-credit-card-amount-link" :class="{ 'disabled': sortable }"
                         v-if="accountCategory.type === AccountCategory.CreditCard.type">
                        <f7-link href="#" class="text-color-gray" popover-open=".credit-card-amount-popover">
                            <span class="text-truncate">{{ showAvailableCreditForCreditCard ? tt('pageSpecific.accountListPage.availableBalance') : tt('pageSpecific.accountListPage.outstandingBalance') }}</span>
                            <f7-icon class="picker-chevron bold-chevron" f7="chevron_up_chevron_down"></f7-icon>
                        </f7-link>
                    </small>
                </f7-list-item>
                <f7-list-item swipeout
                              class="nested-list-item"
                              :id="getAccountDomId(account)"
                              :class="{ 'has-child-list-item': account.type === AccountType.MultiSubAccounts.type && hasVisibleSubAccount(account), 'actual-first-child': account.id === firstShowingIds.accounts[accountCategory.type], 'actual-last-child': account.id === lastShowingIds.accounts[accountCategory.type] }"
                              :after="account.type === AccountType.SingleAccount.type ? accountBalanceOrAvailableCredit(account, undefined, showAvailableCreditForCreditCard, showAccountBalance) : ''"
                              :link="!sortable ? '/transaction/list?accountIds=' + account.id : null"
                              :key="account.id"
                              v-for="account in allCategorizedAccountsMap[accountCategory.type]!.accounts"
                              v-show="showHidden || !account.hidden"
                              @taphold="setSortable()"
                >
                    <template #media v-if="account.type !== AccountType.MultiSubAccounts.type || !hasVisibleSubAccount(account)">
                        <ItemIcon :icon-type="getAccountIconType(account.iconType)" :icon-id="account.icon" :color="account.color">
                            <f7-badge color="gray" class="right-bottom-icon" v-if="account.hidden">
                                <f7-icon f7="eye_slash_fill"></f7-icon>
                            </f7-badge>
                        </ItemIcon>
                    </template>

                    <template #title>
                        <div class="nested-list-item-inner display-flex padding-top-half padding-bottom-half">
                            <ItemIcon :icon-type="getAccountIconType(account.iconType)" :icon-id="account.icon" :color="account.color"
                                      v-if="account.type === AccountType.MultiSubAccounts.type && hasVisibleSubAccount(account)">
                                <f7-badge color="gray" class="right-bottom-icon" v-if="account.hidden">
                                    <f7-icon f7="eye_slash_fill"></f7-icon>
                                </f7-badge>
                            </ItemIcon>
                            <div class="nested-list-item-title">
                                <span>{{ account.name }}</span>
                                <div class="item-footer" v-if="account.comment">{{ account.comment }}</div>
                            </div>
                            <div class="nested-list-item-after" v-if="account.type === AccountType.MultiSubAccounts.type">
                                <span>{{ accountBalanceOrAvailableCredit(account, undefined, showAvailableCreditForCreditCard, showAccountBalance) }}</span>
                            </div>
                        </div>
                        <li v-if="account.type === AccountType.MultiSubAccounts.type">
                            <ul class="no-padding">
                                <f7-list-item class="no-sortable nested-list-item-child"
                                              :class="{ 'actual-first-child': subAccount.id === firstShowingIds.subAccounts[account.id], 'actual-last-child': subAccount.id === lastShowingIds.subAccounts[account.id] }"
                                              :id="getAccountDomId(subAccount)"
                                              :title="subAccount.name" :footer="subAccount.comment"
                                              :after="!showAvailableCreditForCreditCard || account.category !== AccountCategory.CreditCard.type ? accountBalance(account, subAccount.id, showAccountBalance) : ''"
                                              :link="!sortable ? '/transaction/list?accountIds=' + subAccount.id : null"
                                              :key="subAccount.id"
                                              v-for="subAccount in account.subAccounts"
                                              v-show="showHidden || !subAccount.hidden"
                                >
                                    <template #media>
                                        <ItemIcon :icon-type="getAccountIconType(subAccount.iconType)" :icon-id="subAccount.icon" :color="subAccount.color">
                                            <f7-badge color="gray" class="right-bottom-icon" v-if="subAccount.hidden">
                                                <f7-icon f7="eye_slash_fill"></f7-icon>
                                            </f7-badge>
                                        </ItemIcon>
                                    </template>
                                </f7-list-item>
                            </ul>
                        </li>
                    </template>
                    <f7-swipeout-actions :left="textDirection === TextDirection.LTR"
                                         :right="textDirection === TextDirection.RTL"
                                         v-if="sortable">
                        <f7-swipeout-button class="padding-horizontal" overswipe close
                                            :aria-label="account.hidden ? tt('Show') : tt('Hide')"
                                            :color="account.hidden ? 'blue' : 'gray'"
                                            @click="hide(account, !account.hidden)">
                            <f7-icon :f7="account.hidden ? 'eye' : 'eye_slash'"></f7-icon>
                        </f7-swipeout-button>
                    </f7-swipeout-actions>
                    <f7-swipeout-actions :left="textDirection === TextDirection.RTL"
                                         :right="textDirection === TextDirection.LTR"
                                         v-if="!sortable">
                        <f7-swipeout-button color="orange" close :text="tt('Edit')" @click="edit(account)"></f7-swipeout-button>
                        <f7-swipeout-button color="primary" close :text="tt('More')" @click="showMoreActionSheetForAccount(account)"></f7-swipeout-button>
                        <f7-swipeout-button color="red" class="padding-horizontal" :aria-label="tt('Delete')" @click="remove(account, false)">
                            <f7-icon f7="trash"></f7-icon>
                        </f7-swipeout-button>
                    </f7-swipeout-actions>
                </f7-list-item>
            </f7-list>
        </div>

        <f7-actions close-by-outside-click close-on-escape :opened="showAccountMoreActionSheet" @actions:closed="showAccountMoreActionSheet = false">
            <f7-actions-group v-if="accountForMoreActionSheet && accountForMoreActionSheet.type === AccountType.SingleAccount.type">
                <f7-actions-button @click="showReconciliationStatement(accountForMoreActionSheet)">{{ tt('Reconciliation Statement') }}</f7-actions-button>
                <f7-actions-button @click="updateLastReconciledTime(accountForMoreActionSheet)" v-if="useLastReconciledTime">{{ tt('Mark as Reconciled') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="accountForMoreActionSheet && accountForMoreActionSheet.type === AccountType.SingleAccount.type">
                <f7-actions-button @click="moveAllTransactions(accountForMoreActionSheet)">{{ tt('Move All Transactions') }}</f7-actions-button>
                <f7-actions-button color="red" @click="showPasswordSheetForClearAllTransaction(accountForMoreActionSheet)">{{ tt('Clear All Transactions') }}</f7-actions-button>
            </f7-actions-group>
            <template v-if="accountForMoreActionSheet && accountForMoreActionSheet.type === AccountType.MultiSubAccounts.type">
                <f7-actions-group :key="subAccount.id"
                                  v-for="subAccount in accountForMoreActionSheet.subAccounts"
                                  v-show="showHidden || !subAccount.hidden">
                    <f7-actions-label>{{ subAccount.name }}</f7-actions-label>
                    <f7-actions-button @click="showReconciliationStatement(subAccount)">{{ tt('Reconciliation Statement') }}</f7-actions-button>
                    <f7-actions-button @click="updateLastReconciledTime(subAccount)" v-if="useLastReconciledTime">{{ tt('Mark as Reconciled') }}</f7-actions-button>
                    <f7-actions-button @click="moveAllTransactions(subAccount)">{{ tt('Move All Transactions') }}</f7-actions-button>
                    <f7-actions-button color="red" @click="showPasswordSheetForClearAllTransaction(subAccount)">{{ tt('Clear All Transactions') }}</f7-actions-button>
                </f7-actions-group>
            </template>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-actions close-by-outside-click close-on-escape :opened="showMoreActionSheet" @actions:closed="showMoreActionSheet = false">
            <f7-actions-group>
                <f7-actions-button :class="{ 'disabled': maxCategoryAccountCount < 2 }" @click="setSortable()">{{ tt('Sort') }}</f7-actions-button>
                <f7-actions-button v-if="!showHidden" @click="showHidden = true">{{ tt('Show Hidden Accounts') }}</f7-actions-button>
                <f7-actions-button v-if="showHidden" @click="showHidden = false">{{ tt('Hide Hidden Accounts') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="hasAnyVisibleAccount">
                <f7-actions-button @click="setAccountsIncludedInTotal()">{{ tt('Set Accounts Included in Total') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteActionSheet" @actions:closed="showDeleteActionSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this account?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="remove(accountToDelete, true)">{{ tt('Delete') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-popover class="credit-card-amount-popover">
            <f7-list dividers>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt('Outstanding Balance')"
                              :class="{ 'list-item-selected': !showAvailableCreditForCreditCard }"
                              @click="showAvailableCreditForCreditCard = false">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="!showAvailableCreditForCreditCard" />
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt('Available Credit')"
                              :class="{ 'list-item-selected': showAvailableCreditForCreditCard }"
                              @click="showAvailableCreditForCreditCard = true">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="showAvailableCreditForCreditCard" />
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <password-input-sheet :title="tt('Are you sure you want to clear all transactions?')"
                              :hint="tt('format.misc.clearTransactionsInAccountTip', { account: accountToClearTransactions?.name ?? 'undefined' })"
                              :confirm-disabled="clearingData"
                              :cancel-disabled="clearingData"
                              color="red"
                              v-model:show="showInputPasswordSheetForClearAllTransactions"
                              v-model="currentPasswordForClearData"
                              @password:confirm="clearAllTransactions">
        </password-input-sheet>
    </f7-page>
</template>



<style>
.account-overview-card {
    background-color: var(--app-tint);
}

.dark .account-overview-card {
    background-color: var(--app-tint);
}

.dark .account-overview-card a {
    color: var(--f7-text-color);
    opacity: 0.6;
}

.net-assets {
    font-size: 1.5em;
}

.account-overview-info {
    opacity: 0.6;
}

.account-overview-info > span {
    margin-inline-end: 4px;
}

.account-overview-info > span:last-child {
    margin-inline-end: 0;
}

.account-list {
    --f7-list-item-footer-font-size: var(--ebk-large-footer-font-size);
}

.account-list .item-footer {
    padding-top: 4px;
}

.account-list-credit-card-amount-link {
    display: flex;
    min-width: 0;
    margin-inline-start: 4px;

    > .link {
        display: flex;
        min-width: 0;
        max-width: 100%;

        > .picker-chevron {
            flex-shrink: 0;
            font-weight: bold;
        }
    }
}
</style>
