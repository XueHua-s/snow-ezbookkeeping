<script setup lang="ts">
import ParentAccountForm from './EditPage/ParentAccountForm.vue';
import SingleAccountForm from './EditPage/SingleAccountForm.vue';
import type { AccountContext } from './EditPage/useEditPage.ts';
import { ALL_ACCOUNT_COLORS } from '@/consts/color.ts';

import { ALL_ACCOUNT_ICONS } from '@/consts/icon.ts';
import { TRANSACTION_MAX_AMOUNT, TRANSACTION_MIN_AMOUNT } from '@/consts/transaction.ts';
import { AccountType } from '@/core/account.ts';
import { findDisplayNameByType } from '@/lib/common.ts';
import {
    getCurrentUnixTime
} from '@/lib/datetime.ts';
import { getAccountIconType } from '@/lib/icon.ts';
import type { Router } from 'framework7/types';
import { useEditPage } from './EditPage/useEditPage.ts';

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const {
    tt, getCurrencyName, editAccountId, loading, submitting, account, subAccounts, useLastReconciledTime, title, inputIsEmpty,
    allAccountCategories, allAccountTypes, allAvailableMonthDays, getDefaultTimezoneOffsetMinutes,
    getAccountCreditCardStatementDate, getAccountCreditCardCreditLimitDisplayValue, updateAccountBalanceTime,
    updateAccountLastReconciledTime, isNewAccount, accountContext, subAccountContexts, subAccountToDelete,
    showAccountCategorySheet, showAccountTypeSheet, showMoreActionSheet, showDeleteActionSheet, showCreditCardLimitSheet,
    allCurrencies, allCurrenciesWithNotSet, formatAccountDisplayBalance, formatDate, formatTime, save,
    addSubAccountAndContext, removeSubAccount, showBalanceDateTimeDialog, showLastReconciledDateTimeDialog, onPageAfterIn,
} = useEditPage(props);

</script>

<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt(title)"></f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }">
                <f7-link icon-f7="ellipsis" :class="{ 'disabled': account.type !== AccountType.MultiSubAccounts.type }" :aria-label="tt('More')" @click="showMoreActionSheet = true"></f7-link>
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': inputIsEmpty || submitting }" :aria-label="tt('Save')" @click="save"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="margin-vertical-half skeleton-text" v-if="loading">
            <f7-list-item class="list-item-with-header-and-title" header="Account Category" title="Category"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title" header="Account Type" title="Account Type"></f7-list-item>
        </f7-list>

        <f7-list form strong inset dividers class="margin-vertical-half" v-else-if="!loading">
            <f7-list-item
                link="#" no-chevron
                class="list-item-with-header-and-title"
                :header="tt('Account Category')"
                :title="findDisplayNameByType(allAccountCategories, account.category)"
                @click="showAccountCategorySheet = true"
            >
                <list-item-selection-sheet value-type="item"
                                           key-field="type" value-field="type" title-field="displayName"
                                           icon-field="defaultAccountIconId" icon-type="account"
                                           :items="allAccountCategories"
                                           v-model:show="showAccountCategorySheet"
                                           v-model="account.category">
                </list-item-selection-sheet>
            </f7-list-item>

            <f7-list-item
                link="#" no-chevron
                class="list-item-with-header-and-title"
                :class="{ 'disabled': editAccountId }"
                :header="tt('Account Type')"
                :title="findDisplayNameByType(allAccountTypes, account.type)"
                @click="showAccountTypeSheet = true"
            >
                <list-item-selection-sheet value-type="item"
                                           key-field="type" value-field="type" title-field="displayName"
                                           :items="allAccountTypes"
                                           v-model:show="showAccountTypeSheet"
                                           v-model="account.type">
                </list-item-selection-sheet>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="loading">
            <f7-list-input label="Account Name" placeholder="Your account name"></f7-list-input>
            <f7-list-item class="list-item-with-header-and-title list-item-with-multi-item">
                <template #default>
                    <div class="grid grid-cols-2">
                        <div class="list-item-subitem no-chevron">
                            <a class="item-link" href="#">
                                <div class="item-content">
                                    <div class="item-inner">
                                        <div class="item-header">
                                            <span>Account Icon</span>
                                        </div>
                                        <div class="item-title">
                                            <div class="list-item-custom-title no-padding">
                                                <f7-icon f7="app_fill"></f7-icon>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </a>
                        </div>
                        <div class="list-item-subitem no-chevron">
                            <a class="item-link" href="#">
                                <div class="item-content">
                                    <div class="item-inner">
                                        <div class="item-header">
                                            <span>Account Color</span>
                                        </div>
                                        <div class="item-title">
                                            <div class="list-item-custom-title no-padding">
                                                <f7-icon f7="app_fill"></f7-icon>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </a>
                        </div>
                    </div>
                </template>
            </f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Currency" title="Currency" :link="editAccountId ? null : '#'"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title" header="Account Balance" title="Balance"></f7-list-item>
            <f7-list-item class="list-item-toggle" header="Visible" after="True"></f7-list-item>
            <f7-list-input label="Description" type="textarea" placeholder="Your account description (optional)"></f7-list-input>
        </f7-list>

        <SingleAccountForm
    v-else-if="!loading && account.type === AccountType.SingleAccount.type"
    :loading="loading"
    :editAccountId="editAccountId"
    :allCurrencies="allCurrencies"
    :getAccountCreditCardCreditLimitDisplayValue="getAccountCreditCardCreditLimitDisplayValue"
    :getAccountCreditCardStatementDate="getAccountCreditCardStatementDate"
    :allAvailableMonthDays="allAvailableMonthDays"
    :formatAccountDisplayBalance="formatAccountDisplayBalance"
    :formatDate="formatDate"
    :formatTime="formatTime"
    :getDefaultTimezoneOffsetMinutes="getDefaultTimezoneOffsetMinutes"
    :useLastReconciledTime="useLastReconciledTime"
    v-model:account="account"
    v-model:accountContext="accountContext"
    v-model:showCreditCardLimitSheet="showCreditCardLimitSheet"
    @showBalanceDateTimeDialog="showBalanceDateTimeDialog"
    @updateAccountBalanceTime="updateAccountBalanceTime"
    @showLastReconciledDateTimeDialog="showLastReconciledDateTimeDialog"
    @updateAccountLastReconciledTime="updateAccountLastReconciledTime"
/>

        <ParentAccountForm
    v-else-if="!loading && account.type === AccountType.MultiSubAccounts.type"
    :loading="loading"
    :allCurrenciesWithNotSet="allCurrenciesWithNotSet"
    :getAccountCreditCardCreditLimitDisplayValue="getAccountCreditCardCreditLimitDisplayValue"
    :getAccountCreditCardStatementDate="getAccountCreditCardStatementDate"
    :allAvailableMonthDays="allAvailableMonthDays"
    :editAccountId="editAccountId"
    v-model:account="account"
    v-model:accountContext="accountContext"
    v-model:showCreditCardLimitSheet="showCreditCardLimitSheet"
/>

        <f7-block class="no-padding no-margin" v-if="!loading && account.type === AccountType.MultiSubAccounts.type">
            <f7-list strong inset dividers class="subaccount-edit-list margin-vertical"
                     :key="idx"
                     v-for="(subAccount, idx) in subAccounts">
                <f7-list-item group-title>
                    <small>{{ tt('Sub Account') + ' #' + (idx + 1) }}</small>
                    <f7-button rasied fill class="subaccount-delete-button" color="red" icon-f7="trash" icon-size="16px" :aria-label="tt('Remove')"
                               :tooltip="tt('Remove Sub-account')"
                               @click="removeSubAccount(subAccount, false)">
                    </f7-button>
                </f7-list-item>

                <f7-list-input
                    type="text"
                    autocomplete="off"
                    clear-button
                    :label="tt('Sub-account Name')"
                    :placeholder="tt('Your sub-account name')"
                    v-model:value="subAccount.name"
                ></f7-list-input>

                <f7-list-item class="list-item-with-header-and-title list-item-with-multi-item">
                    <template #default>
                        <div class="grid grid-cols-2">
                            <div class="list-item-subitem no-chevron">
                                <a class="item-link" href="#" @click="subAccountContexts[idx]!.showIconSelectionSheet = true">
                                    <div class="item-content">
                                        <div class="item-inner">
                                            <div class="item-header">
                                                <span>{{ tt('Sub-account Icon') }}</span>
                                            </div>
                                            <div class="item-title">
                                                <div class="list-item-custom-title no-padding">
                                                    <ItemIcon :icon-type="getAccountIconType(subAccount.iconType)" :icon-id="subAccount.icon" :color="subAccount.color"></ItemIcon>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </a>

                                <icon-selection-sheet :all-system-icon-infos="ALL_ACCOUNT_ICONS"
                                                      :color="subAccount.color"
                                                      v-model:show="subAccountContexts[idx]!.showIconSelectionSheet"
                                                      v-model:icon-type="subAccount.iconType"
                                                      v-model="subAccount.icon"
                                ></icon-selection-sheet>
                            </div>
                            <div class="list-item-subitem no-chevron">
                                <a class="item-link" href="#" @click="subAccountContexts[idx]!.showColorSelectionSheet = true">
                                    <div class="item-content">
                                        <div class="item-inner">
                                            <div class="item-header">
                                                <span>{{ tt('Sub-account Color') }}</span>
                                            </div>
                                            <div class="item-title">
                                                <div class="list-item-custom-title no-padding">
                                                    <ItemIcon icon-type="fixed-f7" icon-id="app_fill" :color="subAccount.color"></ItemIcon>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </a>

                                <color-selection-sheet :all-system-color-infos="ALL_ACCOUNT_COLORS"
                                                       v-model:show="subAccountContexts[idx]!.showColorSelectionSheet"
                                                       v-model="subAccount.color"
                                ></color-selection-sheet>
                            </div>
                        </div>
                    </template>
                </f7-list-item>

                <f7-list-item
                    class="list-item-with-header-and-title list-item-no-item-after"
                    link="#"
                    :class="{ 'disabled': editAccountId && !isNewAccount(subAccount) }"
                    :header="tt('Currency')"
                    :no-chevron="!!editAccountId && !isNewAccount(subAccount)"
                    @click="subAccountContexts[idx]!.showCurrencyPopup = true"
                >
                    <template #title>
                        <div class="no-padding no-margin">
                            <span>{{ getCurrencyName(subAccount.currency) }}&nbsp;</span>
                            <small class="smaller">{{ subAccount.currency }}</small>
                        </div>
                    </template>
                    <list-item-selection-popup value-type="item"
                                               key-field="currencyCode" value-field="currencyCode"
                                               title-field="displayName" after-field="currencyCode"
                                               :title="tt('Currency Name')"
                                               :enable-filter="true"
                                               :filter-placeholder="tt('Currency')"
                                               :filter-no-items-text="tt('No results')"
                                               :items="allCurrencies"
                                               v-model:show="subAccountContexts[idx]!.showCurrencyPopup"
                                               v-model="subAccount.currency">
                    </list-item-selection-popup>
                </f7-list-item>

                <f7-list-item
                    link="#" no-chevron
                    class="list-item-with-header-and-title"
                    :class="{ 'disabled': editAccountId && !isNewAccount(subAccount) }"
                    :header="account.isLiability ? tt('Sub-account Outstanding Balance') : tt('Sub-account Balance')"
                    :title="formatAccountDisplayBalance(subAccount)"
                    @click="subAccountContexts[idx]!.showBalanceSheet = true"
                >
                    <number-pad-sheet :min-value="TRANSACTION_MIN_AMOUNT"
                                      :max-value="TRANSACTION_MAX_AMOUNT"
                                      :currency="subAccount.currency"
                                      :flip-negative="account.isLiability"
                                      v-model:show="subAccountContexts[idx]!.showBalanceSheet"
                                      v-model="subAccount.numericBalance"
                    ></number-pad-sheet>
                </f7-list-item>

                <f7-list-item
                    class="account-edit-datetime list-item-with-header-and-title"
                    link="#" no-chevron
                    v-show="subAccount.numericBalance"
                    v-if="!editAccountId || isNewAccount(subAccount)"
                >
                    <template #header>
                        <div class="account-edit-datetime-header" @click="showBalanceDateTimeDialog(subAccountContexts[idx] as AccountContext, 'time')">{{ tt('Sub-account Balance Time') }}</div>
                    </template>
                    <template #title>
                        <div class="account-edit-datetime-title">
                            <div @click="showBalanceDateTimeDialog(subAccountContexts[idx] as AccountContext, 'date')">{{ formatDate(subAccount.balanceTime) }}</div>&nbsp;<div class="account-edit-datetime-time" @click="showBalanceDateTimeDialog(subAccountContexts[idx] as AccountContext, 'time')">{{ formatTime(subAccount.balanceTime) }}</div>
                        </div>
                    </template>
                    <date-time-selection-sheet :init-mode="subAccountContexts[idx]!.balanceDateTimeSheetMode"
                                               :timezone-utc-offset="getDefaultTimezoneOffsetMinutes(subAccount.balanceTime)"
                                               :model-value="subAccount.balanceTime"
                                               v-model:show="subAccountContexts[idx]!.showBalanceDateTimeSheet"
                                               @update:model-value="updateAccountBalanceTime(subAccount, $event)">
                    </date-time-selection-sheet>
                </f7-list-item>

                <f7-list-item
                    class="account-edit-datetime list-item-with-header-and-title"
                    link="#" no-chevron
                    v-if="editAccountId && !isNewAccount(subAccount) && useLastReconciledTime"
                >
                    <template #header>
                        <div class="account-edit-datetime-header" @click="showLastReconciledDateTimeDialog(subAccountContexts[idx] as AccountContext, 'time')">{{ tt('Sub-account Last Reconciled Time') }}</div>
                    </template>
                    <template #title>
                        <div class="account-edit-datetime-title" v-if="subAccount.lastReconciledTime">
                            <div @click="showLastReconciledDateTimeDialog(subAccountContexts[idx] as AccountContext, 'date')">{{ formatDate(subAccount.lastReconciledTime) }}</div>&nbsp;<div class="account-edit-datetime-time" @click="showLastReconciledDateTimeDialog(subAccountContexts[idx] as AccountContext, 'time')">{{ formatTime(subAccount.lastReconciledTime) }}</div>
                        </div>
                        <div class="account-edit-datetime-title" v-else>
                            <div class="account-edit-datetime-time" @click="showLastReconciledDateTimeDialog(subAccountContexts[idx] as AccountContext, 'date')">{{ tt('None') }}</div>
                        </div>
                    </template>
                    <date-time-selection-sheet :init-mode="subAccountContexts[idx]!.lastReconciledDateTimeSheetMode"
                                               :clearable="true"
                                               :timezone-utc-offset="getDefaultTimezoneOffsetMinutes(subAccount.lastReconciledTime)"
                                               :model-value="subAccount.lastReconciledTime ?? getCurrentUnixTime()"
                                               v-model:show="subAccountContexts[idx]!.showLastReconciledTimeSheet"
                                               @update:model-value="updateAccountLastReconciledTime(subAccount, $event)"
                                               @clear:model-value="subAccount.lastReconciledTime = undefined">
                    </date-time-selection-sheet>
                </f7-list-item>

                <f7-list-item :title="tt('Visible')" v-if="editAccountId && !isNewAccount(subAccount)">
                    <template #after>
                        <f7-toggle :checked="subAccount.visible" @toggle:change="subAccount.visible = $event"></f7-toggle>
                    </template>
                </f7-list-item>

                <f7-list-input
                    type="textarea"
                    style="height: auto"
                    autocomplete="off"
                    :label="tt('Description')"
                    :placeholder="tt('Your sub-account description (optional)')"
                    v-textarea-auto-size
                    v-model:value="subAccount.comment"
                ></f7-list-input>
            </f7-list>
        </f7-block>

        <f7-actions close-by-outside-click close-on-escape :opened="showMoreActionSheet" @actions:closed="showMoreActionSheet = false">
            <f7-actions-group>
                <f7-actions-button @click="addSubAccountAndContext">{{ tt('Add Sub-account') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteActionSheet" @actions:closed="showDeleteActionSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to remove this sub-account?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="removeSubAccount(subAccountToDelete, true)">{{ tt('Remove') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>



<style>
.account-edit-datetime .item-title {
    width: 100%;
}

.account-edit-datetime .item-title > .item-header > .account-edit-datetime-header {
    display: block;
    width: 100%;
}

.account-edit-datetime .item-title > .account-edit-datetime-title {
    display: flex;
    width: 100%;
}

.account-edit-datetime .item-title > .account-edit-datetime-title > .account-edit-datetime-time {
    flex-grow: 1;
    overflow: hidden;
    text-overflow: ellipsis;
}

.subaccount-delete-button {
    margin-inline-start: auto;
}
</style>
