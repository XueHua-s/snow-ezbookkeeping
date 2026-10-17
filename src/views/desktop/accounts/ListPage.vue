<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { AccountCategory, AccountType } from '@/core/account.ts';
import { getAccountIconType } from '@/lib/icon.ts';
import ClearAllTransactionsDialog from '@/views/desktop/accounts/list/dialogs/ClearAllTransactionsDialog.vue';
import MoveAllTransactionsDialog from '@/views/desktop/accounts/list/dialogs/MoveAllTransactionsDialog.vue';
import AccountFilterSettingsDialog from '@/views/desktop/common/dialogs/AccountFilterSettingsDialog.vue';
import {
    mdiCalculatorVariantOutline,
    mdiDeleteOutline,
    mdiDotsHorizontalCircleOutline,
    mdiDotsVertical,
    mdiDrag,
    mdiEraser,
    mdiEyeOffOutline,
    mdiEyeOutline,
    mdiInvoiceListOutline,
    mdiListBoxOutline,
    mdiPencilOutline,
    mdiReceiptTextCheckOutline,
    mdiRefresh,
    mdiSquareRounded,
    mdiSwapHorizontal
} from '@mdi/js';
import EditDialog from './list/dialogs/EditDialog.vue';
import ReconciliationStatementDialog from './list/dialogs/ReconciliationStatementDialog.vue';
import { useListPage } from './ListPage/useListPage.ts';

const {
    tt, loading, showHidden, displayOrderModified, showAccountBalance, customAccountCategoryOrder, useLastReconciledTime,
    allCategorizedAccountsMap, allAccountCount, totalAssets, canShowAvailableCredit,
    accountCategoryTotalBalance, accountBalance, accountAvailableCredit, activeAccountCategoryType, activeTab,
    activeSubAccount, showAccountsIncludedInTotalDialog, showCustomDateRangeDialog, hideAccountCategoriesWithoutAccounts,
    hasAnyVisibleAccount, activeAccountCategory, activeAccountCategoryTotalBalance, activeCreditCardTotalAvailabeCredit,
    activeAccountCategoryVisibleAccountCount, reload, hasAccount, accountCurrency, accountReconciliationStatementDateRanges,
    add, edit, showReconciliationStatementDialog, updateLastReconciledTime, moveAllTransactions, clearAllTransactions, hide,
    remove, saveSortResult, onMove, onCustomDateRangeChanged, onShowDateRangeError,
} = useListPage();

</script>

<template>
    <main-page-layout nav-items-class="account-category-tabs my-3">
        <template #nav-items>
            <div class="mx-2 mb-2">
                <span class="text-body-medium">{{ tt('Total assets') }}</span>
                <div class="text-body-large text-primary text-truncate mt-1">
                    <span v-if="!loading || allAccountCount > 0">{{ totalAssets }}</span>
                    <span v-else-if="loading && allAccountCount <= 0">
                        <v-skeleton-loader class="skeleton-no-margin mt-3 mb-2 pb-1" type="text" :loading="true"></v-skeleton-loader>
                    </span>
                </div>
            </div>

            <v-divider class="my-2" />

            <li class="nav-link"
                :key="accountCategory.type"
                v-for="accountCategory in AccountCategory.values(customAccountCategoryOrder)"
                v-show="!hideAccountCategoriesWithoutAccounts || (allCategorizedAccountsMap[accountCategory.type] && allCategorizedAccountsMap[accountCategory.type]!.accounts.length > 0)">
                <a class="d-flex align-center cursor-pointer"
                   :class="{ 'router-link-active router-link-exact-active': activeAccountCategoryType === accountCategory.type, 'disabled': loading }"
                   @click="activeAccountCategoryType = accountCategory.type">
                    <ItemIcon icon-type="account" :icon-id="accountCategory.defaultAccountIconId" />
                    <div class="nav-item-title d-flex flex-column text-truncate">
                        <small class="text-truncate text-start smaller" v-if="!loading || allAccountCount > 0">{{ accountCategoryTotalBalance(accountCategory, false) }}</small>
                        <small class="text-truncate text-start smaller mb-1" v-else-if="loading && allAccountCount <= 0">
                            <v-skeleton-loader class="skeleton-no-margin"
                                               width="100px" height="16" type="text" :loading="true"></v-skeleton-loader>
                        </small>
                        <span class="text-body-medium text-truncate text-start">{{ tt(accountCategory.name) }}</span>
                    </div>
                </a>
            </li>
        </template>

        <template #content>
            <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container" v-model="activeTab">
                <v-window-item value="accountPage">
                    <v-card min-height="780">
                        <template #title>
                            <div class="title-and-toolbar d-flex align-center">
                                <span>{{ tt('Account List') }}</span>
                                <v-btn class="ms-3" color="default" variant="outlined"
                                       :disabled="loading" @click="add">{{ tt('Add') }}</v-btn>
                                <v-btn class="ms-3" color="primary" variant="tonal"
                                       :disabled="loading" @click="saveSortResult"
                                       v-if="displayOrderModified">{{ tt('Save Display Order') }}</v-btn>
                                <v-btn density="compact" color="default" variant="text" class="ms-2"
                                       :aria-label="tt('Refresh')" :icon="true" :loading="loading" @click="reload(true)">
                                    <template #loader>
                                        <v-progress-circular indeterminate size="20"/>
                                    </template>
                                    <v-icon :icon="mdiRefresh" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                                </v-btn>
                                <v-spacer/>
                                <v-btn density="comfortable" color="default" variant="text" class="ms-2"
                                       :aria-label="tt('More')" :disabled="loading" :icon="true">
                                    <v-icon :icon="mdiDotsVertical" />
                                    <v-menu activator="parent">
                                        <v-list>
                                            <v-list-item :prepend-icon="mdiEyeOutline"
                                                         :title="tt('Show Hidden Accounts')"
                                                         v-if="!showHidden" @click="showHidden = true"></v-list-item>
                                            <v-list-item :prepend-icon="mdiEyeOffOutline"
                                                         :title="tt('Hide Hidden Accounts')"
                                                         v-if="showHidden" @click="showHidden = false"></v-list-item>
                                            <v-divider class="my-2" v-if="hasAnyVisibleAccount"/>
                                            <v-list-item :prepend-icon="mdiCalculatorVariantOutline"
                                                         :title="tt('Set Accounts Included in Total')"
                                                         v-if="hasAnyVisibleAccount" @click="showAccountsIncludedInTotalDialog = true"></v-list-item>
                                        </v-list>
                                    </v-menu>
                                </v-btn>
                            </div>
                        </template>

                        <v-card-text class="accounts-overview-title text-truncate pt-0">
                            <div class="d-flex align-center">
                                <span class="accounts-overview-subtitle">{{ activeAccountCategory?.isLiability ? tt('Outstanding Balance') : tt('Balance') }}</span>
                                <v-skeleton-loader class="skeleton-no-margin ms-3" width="120px" type="text" :loading="true" v-if="loading && activeAccountCategory && !hasAccount(activeAccountCategory)"></v-skeleton-loader>
                                <template v-else-if="!loading || !activeAccountCategory || hasAccount(activeAccountCategory)">
                                    <span class="accounts-overview-amount ms-3">{{ activeAccountCategoryTotalBalance }}</span>
                                    <template v-if="activeAccountCategory?.type === AccountCategory.CreditCard.type && activeCreditCardTotalAvailabeCredit">
                                        <v-divider vertical class="mx-3 my-1" />
                                        <span class="accounts-overview-subtitle">{{ tt('Available Credit') }}</span>
                                        <span class="accounts-overview-amount ms-3">{{ activeCreditCardTotalAvailabeCredit }}</span>
                                    </template>
                                </template>
                                <v-btn class="ms-2" density="compact" color="default" variant="text"
                                       :aria-label="showAccountBalance ? tt('Hide Account Balance') : tt('Show Account Balance')"
                                       :icon="true" :disabled="loading"
                                       @click="showAccountBalance = !showAccountBalance">
                                    <v-icon :icon="showAccountBalance ? mdiEyeOffOutline : mdiEyeOutline" size="20" />
                                    <v-tooltip activator="parent">{{ showAccountBalance ? tt('Hide Account Balance') : tt('Show Account Balance') }}</v-tooltip>
                                </v-btn>
                            </div>
                        </v-card-text>

                        <v-row class="ps-4 pe-4" v-if="loading && activeAccountCategory && !hasAccount(activeAccountCategory)">
                            <v-col cols="12">
                                <v-card border class="card-title-with-bg account-card mb-4 h-auto">
                                    <template #title>
                                        <div class="account-title d-flex align-center">
                                            <v-icon class="disabled me-0" size="28px" :icon="mdiSquareRounded" />
                                            <span class="account-name text-truncate ms-2">
                                                <v-skeleton-loader class="skeleton-no-margin my-1"
                                                                   width="120px" type="text" :loading="true"></v-skeleton-loader>
                                            </span>
                                            <v-spacer/>
                                            <span class="align-self-center">
                                                <v-icon class="disabled" :icon="mdiDrag"/>
                                            </span>
                                        </div>
                                    </template>
                                    <v-divider/>
                                    <v-card-text>
                                        <div class="d-flex account-toolbar align-center">
                                            <v-btn class="px-2" density="comfortable" color="default" variant="text"
                                                   :disabled="true" :prepend-icon="mdiListBoxOutline">
                                                {{ tt('Transaction List') }}
                                            </v-btn>
                                            <v-spacer/>
                                            <span class="account-balance ms-2">
                                                <v-skeleton-loader class="skeleton-no-margin"
                                                                   width="100px" type="text" :loading="true"></v-skeleton-loader>
                                            </span>
                                        </div>
                                    </v-card-text>
                                </v-card>
                            </v-col>
                        </v-row>

                        <v-row class="ps-4 pe-2 pe-md-4" v-if="!loading && activeAccountCategory && !hasAccount(activeAccountCategory)">
                            <v-col cols="12">
                                <span class="text-body-medium">{{ tt('No available account') }}</span>
                            </v-col>
                        </v-row>

                        <v-row class="ps-4 pe-4">
                            <v-col cols="12">
                                <draggable-list
                                    class="list-group"
                                    item-key="id"
                                    handle=".drag-handle"
                                    ghost-class="dragging-item"
                                    :disabled="activeAccountCategoryVisibleAccountCount <= 1"
                                    :list="allCategorizedAccountsMap[activeAccountCategory.type]!.accounts"
                                    v-if="activeAccountCategory && allCategorizedAccountsMap[activeAccountCategory.type] && allCategorizedAccountsMap[activeAccountCategory.type]!.accounts && allCategorizedAccountsMap[activeAccountCategory.type]!.accounts.length"
                                    @change="onMove"
                                >
                                    <template #item="{ element }">
                                        <div class="list-group-item">
                                            <v-card border class="card-title-with-bg account-card mb-4 h-auto" v-if="showHidden || !element.hidden">
                                                <template #title>
                                                    <div class="account-title d-flex align-center">
                                                        <ItemIcon size="1.5rem" :icon-type="getAccountIconType(element.iconType)" :icon-id="element.icon"
                                                                  :color="element.color" :hidden-status="element.hidden" />
                                                        <span class="account-name text-truncate ms-2">{{ element.name }}</span>
                                                        <small class="account-currency text-truncate ms-2 align-self-end">
                                                            {{ accountCurrency(element) }}
                                                        </small>
                                                        <v-spacer/>
                                                        <span class="align-self-center">
                                                            <v-icon :class="!loading && activeAccountCategoryVisibleAccountCount > 1 ? 'drag-handle' : 'disabled'"
                                                                    :aria-label="tt('Drag to Reorder')" :icon="mdiDrag"/>
                                                            <v-tooltip activator="parent" v-if="!loading && activeAccountCategoryVisibleAccountCount > 1">{{ tt('Drag to Reorder') }}</v-tooltip>
                                                        </span>
                                                    </div>

                                                    <div class="mt-4" v-if="element.type === AccountType.MultiSubAccounts.type">
                                                        <v-btn-toggle
                                                            class="account-subaccounts"
                                                            variant="outlined"
                                                            color="primary"
                                                            density="compact"
                                                            mandatory="force"
                                                            rounded="xl"
                                                            :disabled="loading"
                                                            v-model="activeSubAccount[element.id]"
                                                        >
                                                            <v-btn :value="''">
                                                                <span>{{ tt('All') }}</span>
                                                            </v-btn>
                                                            <v-btn :key="subAccount.id" :value="subAccount.id"
                                                                   v-for="subAccount in element.subAccounts"
                                                                   v-show="showHidden || !subAccount.hidden">
                                                                <ItemIcon size="1.5rem" :icon-type="getAccountIconType(subAccount.iconType)" :icon-id="subAccount.icon"
                                                                          :color="subAccount.color" :hidden-status="subAccount.hidden" />
                                                                <span class="ms-2">{{ subAccount.name }}</span>
                                                            </v-btn>
                                                        </v-btn-toggle>
                                                    </div>
                                                </template>

                                                <v-divider/>

                                                <v-card-text v-if="element.getAccountOrSubAccountComment(activeSubAccount[element.id])">
                                                    {{ element.getAccountOrSubAccountComment(activeSubAccount[element.id]) }}
                                                </v-card-text>

                                                <v-card-text>
                                                    <div class="d-flex account-toolbar align-center">
                                                        <v-btn class="px-2" density="comfortable" color="default" variant="text"
                                                               :disabled="loading" :prepend-icon="mdiListBoxOutline"
                                                               :to="`/transaction/list?accountIds=${element.getAccountOrSubAccountId(activeSubAccount[element.id])}`">
                                                            {{ tt('Transaction List') }}
                                                        </v-btn>
                                                        <v-btn class="ms-1" density="comfortable" color="default" variant="text"
                                                               :disabled="loading" :prepend-icon="mdiInvoiceListOutline"
                                                               @click="showReconciliationStatementDialog(element.getAccountOrSubAccount(activeSubAccount[element.id]))"
                                                               v-if="element.type === AccountType.SingleAccount.type || element.getSubAccount(activeSubAccount[element.id])">
                                                            {{ tt('Reconciliation Statement') }}
                                                            <v-menu activator="parent" :open-on-hover="true">
                                                                <v-list>
                                                                    <template :key="dateRange.type"
                                                                              v-for="dateRange in accountReconciliationStatementDateRanges(element.getAccountOrSubAccount(activeSubAccount[element.id]))">
                                                                        <v-list-item class="text-body-medium" density="compact"
                                                                                     :value="dateRange.type">
                                                                            <v-list-item-title class="cursor-pointer"
                                                                                               @click="showReconciliationStatementDialog(element.getAccountOrSubAccount(activeSubAccount[element.id]), dateRange.type)">
                                                                                <div class="d-flex align-center">
                                                                                    <span class="text-body-medium ms-3">{{ dateRange.displayName }}</span>
                                                                                </div>
                                                                            </v-list-item-title>
                                                                        </v-list-item>
                                                                    </template>
                                                                </v-list>
                                                            </v-menu>
                                                        </v-btn>
                                                        <v-btn class="ms-1" density="comfortable" color="default" variant="text"
                                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                                               :disabled="loading"
                                                               :prepend-icon="element.isAccountOrSubAccountHidden(activeSubAccount[element.id]) ? mdiEyeOutline : mdiEyeOffOutline"
                                                               v-if="!activeSubAccount[element.id] || element.getSubAccount(activeSubAccount[element.id])"
                                                               @click="hide(element, element.getAccountOrSubAccount(activeSubAccount[element.id]), !element.isAccountOrSubAccountHidden(activeSubAccount[element.id]))">
                                                            {{ element.isAccountOrSubAccountHidden(activeSubAccount[element.id]) ? tt('Show') : tt('Hide') }}
                                                        </v-btn>
                                                        <v-btn class="ms-1" density="comfortable" color="default" variant="text"
                                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                                               :disabled="loading" :prepend-icon="mdiPencilOutline"
                                                               v-if="!activeSubAccount[element.id] || element.getSubAccount(activeSubAccount[element.id])"
                                                               @click="edit(element)">
                                                            {{ tt('Edit') }}
                                                        </v-btn>
                                                        <v-btn class="ms-1" density="comfortable" color="default" variant="text"
                                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                                               :disabled="loading" :prepend-icon="mdiDotsHorizontalCircleOutline"
                                                               v-if="element.type === AccountType.SingleAccount.type || element.getSubAccount(activeSubAccount[element.id])">
                                                            {{ tt('More') }}
                                                            <v-menu activator="parent" :open-on-hover="true">
                                                                <v-list>
                                                                    <v-list-item class="text-body-medium" density="compact"
                                                                                 :title="tt('Mark as Reconciled')"
                                                                                 :prepend-icon="mdiReceiptTextCheckOutline"
                                                                                 @click="updateLastReconciledTime(element.getAccountOrSubAccount(activeSubAccount[element.id]))"
                                                                                 v-if="useLastReconciledTime"></v-list-item>
                                                                    <v-divider class="my-2" v-if="useLastReconciledTime" />
                                                                    <v-list-item class="text-body-medium" density="compact"
                                                                                 :title="tt('Move All Transactions')"
                                                                                 :prepend-icon="mdiSwapHorizontal"
                                                                                 @click="moveAllTransactions(element.getAccountOrSubAccount(activeSubAccount[element.id]))"></v-list-item>
                                                                    <v-list-item class="text-body-medium" density="compact"
                                                                                 :title="tt('Clear All Transactions')"
                                                                                 :prepend-icon="mdiEraser"
                                                                                 @click="clearAllTransactions(element.getAccountOrSubAccount(activeSubAccount[element.id]))"></v-list-item>
                                                                </v-list>
                                                            </v-menu>
                                                        </v-btn>
                                                        <v-btn class="ms-1" density="comfortable" color="default" variant="text"
                                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                                               :disabled="loading" :prepend-icon="mdiDeleteOutline"
                                                               v-if="!activeSubAccount[element.id] || element.getSubAccount(activeSubAccount[element.id])"
                                                               @click="remove(element)">
                                                            {{ tt('Delete') }}
                                                        </v-btn>
                                                        <v-spacer/>
                                                        <div class="d-flex align-center ms-2" v-if="element.category === AccountCategory.CreditCard.type">
                                                            <div class="d-flex align-center">
                                                                <span class="text-medium-emphasis me-3 pt-h1">{{ tt('pageSpecific.accountListPage.outstandingBalance') }}</span>
                                                                <span class="account-balance">{{ accountBalance(element, activeSubAccount[element.id], showAccountBalance) }}</span>
                                                            </div>
                                                            <template v-if="!activeSubAccount[element.id] && canShowAvailableCredit(element) && accountAvailableCredit(element, showAccountBalance)">
                                                                <v-divider vertical class="mx-3 my-1"/>
                                                                <div class="d-flex align-center">
                                                                    <span class="text-medium-emphasis me-3 pt-h1">{{ tt('pageSpecific.accountListPage.availableBalance') }}</span>
                                                                    <span class="account-balance">{{ accountAvailableCredit(element, showAccountBalance) }}</span>
                                                                </div>
                                                            </template>
                                                        </div>
                                                        <span class="account-balance ms-2" v-else-if="element.category !== AccountCategory.CreditCard.type">{{ accountBalance(element, activeSubAccount[element.id], showAccountBalance) }}</span>
                                                    </div>
                                                </v-card-text>
                                            </v-card>
                                        </div>
                                    </template>
                                </draggable-list>
                            </v-col>
                        </v-row>
                    </v-card>
                </v-window-item>
            </v-window>
        </template>
    </main-page-layout>

    <account-filter-settings-dialog type="accountListTotalAmount"
                                    v-model:show="showAccountsIncludedInTotalDialog"
                                    @settings:change="showAccountsIncludedInTotalDialog = false" />

    <edit-dialog ref="editDialog" />
    <reconciliation-statement-dialog ref="reconciliationStatementDialog"
                                     @error="onShowDateRangeError" />
    <move-all-transactions-dialog ref="moveAllTransactionsDialog" />
    <clear-all-transactions-dialog ref="clearAllTransactionsDialog" />

    <date-range-selection-dialog :title="tt('Custom Date Range')"
                                 v-model:show="showCustomDateRangeDialog"
                                 @dateRange:change="onCustomDateRangeChanged"
                                 @error="onShowDateRangeError" />

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>



<style>
.account-category-tabs .nav-link {
    height: calc(var(--v-tabs-height) * 1.5);
}

.accounts-overview-title {
    line-height: 2rem !important;
    min-height: 52px;
    display: flex;
    align-items: flex-end;
}

.accounts-overview-amount {
    font-size: 1.5rem;
    color: rgba(var(--v-theme-on-background), var(--v-high-emphasis-opacity));
    overflow: hidden;
    text-overflow: ellipsis;
}

.accounts-overview-subtitle {
    font-size: 1rem;
    line-height: 1.75rem;
}

.account-card {
    > .v-card-item {
        padding-top: 0.875rem;
        padding-bottom: 0.875rem;
    }

    .account-title {
        font-size: 1rem;
        line-height: 1.5rem !important;
    }

    .account-title .account-name {
        color: rgba(var(--v-theme-on-background), var(--v-high-emphasis-opacity));
    }

    .account-currency {
        font-size: 0.8rem;
        color: rgba(var(--v-theme-on-background), var(--v-medium-emphasis-opacity));
    }

    .account-subaccounts {
        overflow-x: auto;
        white-space: nowrap;
        min-height: 34px;

        &.v-btn-toggle {
            height: auto !important;
            padding: 0;
            border: none;

            > .v-btn {
                width: auto !important;
                border: thin solid rgba(var(--v-border-color), var(--v-border-opacity));

                &:not(:first-child) {
                    border-top-left-radius: 0;
                    border-bottom-left-radius: 0;
                    border-left: none;
                }

                &:not(:last-child) {
                    border-top-right-radius: 0;
                    border-bottom-right-radius: 0;
                }
            }
        }
    }

    .account-toolbar {
        overflow-x: auto;
        white-space: nowrap;

        .v-btn {
            padding: 0 12px;
            min-height: 32px;
        }

        .hover-display {
            display: none;
        }

        &:hover .hover-display {
            display: inline-flex;
        }
    }

    .account-balance {
        font-size: 1.25rem;
        color: rgba(var(--v-theme-on-background), var(--v-high-emphasis-opacity));
    }
}
</style>
