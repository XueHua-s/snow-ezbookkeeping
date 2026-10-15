<script setup lang="ts">
import { TRANSACTION_MAX_AMOUNT, TRANSACTION_MIN_AMOUNT } from '@/consts/transaction.ts';
import { TextDirection } from '@/core/text.ts';
import { TransactionType } from '@/core/transaction.ts';
import { getCategoryIconType } from '@/lib/icon.ts';
import type { Router } from 'framework7/types';
import { useReconciliationStatementPage } from './ReconciliationStatementPage/useReconciliationStatementPage.ts';

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const {
    tt, formatNumberToLocalizedNumerals, accountId, startTime, endTime, reconciliationStatements,
    chartDataDateAggregationType, timezoneUsedForDateRange, fiscalYearStart, allDateAggregationTypes,
    allTimezoneTypesUsedForDateRange, isCurrentLiabilityAccount, canUpdateAccountCloseBalance, newLastReconciledTime,
    currentAccount, currentAccountCurrency, currentAccountStatementDate, displayStartDateTime, displayEndDateTime,
    displayTotalInflows, displayTotalOutflows, displayTotalBalance, displayOpeningBalance, displayClosingBalance,
    getDisplayTime, isSameAsDefaultTimezoneOffsetMinutes, getDisplayTimezone, getDisplaySourceAmount,
    getDisplayDestinationAmount, getDisplayAccountBalance, finishQuery, loading, queryDateRangeType,
    showAccountBalanceTrendsCharts, updatingLastReconciledTime, transactionToDelete, newClosingBalance,
    showCustomDateRangeSheet, showNewClosingBalanceSheet, showMoreActionSheet, showDeleteActionSheet, virtualDataItems,
    textDirection, validQuery, allAvailableDateRanges, allReconciliationStatementVirtualListItems,
    chartDataDateAggregationTypeDisplayName, getTransactionDomId, changeDateFilter, changeCustomDateFilter, reload,
    addTransaction, duplicateTransaction, editTransaction, updateClosingBalance, updateLastReconciledTime, removeTransaction,
    setChartDataDateAggregationType, setTimezoneUsedForDateRange, renderExternal, onPageAfterIn,
} = useReconciliationStatementPage(props);

</script>

<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading || updatingLastReconciledTime }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title>
                <span style="color: var(--f7-text-color)" v-if="!finishQuery">{{ tt('Reconciliation Statement') }}</span>
                <f7-link popover-open=".display-mode-popover-menu" :class="{ 'disabled': loading || updatingLastReconciledTime }" v-if="finishQuery">
                    <span style="color: var(--f7-text-color)">{{ tt('Reconciliation Statement') }}</span>
                    <f7-icon class="page-title-bar-icon" style="opacity: 0.5"
                             color="gray" f7="chevron_down_circle_fill"></f7-icon>
                </f7-link>
            </f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading || updatingLastReconciledTime }">
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': !validQuery }" :aria-label="tt('Continue')" @click="reload(false)" v-if="!finishQuery"></f7-link>
                <f7-link icon-f7="ellipsis" :class="{ 'disabled': loading || updatingLastReconciledTime }" :aria-label="tt('More')" v-if="finishQuery" @click="showMoreActionSheet = true"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-popover class="display-mode-popover-menu">
            <f7-list dividers>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt('Transaction List')"
                              :class="{ 'list-item-selected': !showAccountBalanceTrendsCharts }"
                              @click="showAccountBalanceTrendsCharts = false">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="!showAccountBalanceTrendsCharts"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt('Account Balance Trends')"
                              :class="{ 'list-item-selected': showAccountBalanceTrendsCharts, 'disabled': !reconciliationStatements?.transactions?.length }"
                              @click="showAccountBalanceTrendsCharts = true">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="showAccountBalanceTrendsCharts"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <f7-list form strong inset dividers class="margin-vertical-half" v-if="!finishQuery">
            <f7-list-item group-title>
                <small>{{ tt('Date Range') }}</small>
            </f7-list-item>
            <f7-list-item :key="dateRange.type"
                          :title="dateRange.displayName"
                          :disabled="!validQuery"
                          v-for="dateRange in allAvailableDateRanges"
                          @click="changeDateFilter(dateRange.type)">
                <template #after>
                    <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="queryDateRangeType === dateRange.type"></f7-icon>
                </template>
                <template #footer>
                    <div v-if="dateRange.isUserCustomRange && queryDateRangeType === dateRange.type && startTime && endTime">
                        <span>{{ displayStartDateTime }}</span>
                        <span>&nbsp;-&nbsp;</span>
                        <br/>
                        <span>{{ displayEndDateTime }}</span>
                    </div>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical-half" v-if="finishQuery && !startTime && !endTime">
            <f7-list-item :title="tt('Date Range')" :after="tt('All')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical-half" v-if="finishQuery && (startTime || endTime)">
            <f7-list-item :title="tt('Start Time')" :after="displayStartDateTime"></f7-list-item>
            <f7-list-item :title="tt('End Time')" :after="displayEndDateTime"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="finishQuery && loading">
            <f7-list-item :title="tt('Total Transactions')" after="Count"></f7-list-item>
            <f7-list-item :title="tt('Total Inflows')" after="Count"></f7-list-item>
            <f7-list-item :title="tt('Total Outflows')" after="Count"></f7-list-item>
            <f7-list-item :title="tt('Net Cash Flow')" after="Count"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical" v-if="finishQuery && !loading">
            <f7-list-item :title="tt('Total Transactions')"
                          :after="formatNumberToLocalizedNumerals(reconciliationStatements.transactions.length)"
                          v-if="reconciliationStatements && reconciliationStatements.transactions"></f7-list-item>
            <f7-list-item :title="tt('Total Inflows')" :after="displayTotalInflows"></f7-list-item>
            <f7-list-item :title="tt('Total Outflows')" :after="displayTotalOutflows"></f7-list-item>
            <f7-list-item :title="tt('Net Cash Flow')" :after="displayTotalBalance"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="finishQuery && loading">
            <f7-list-item :title="tt('Opening Balance')" after="Count"></f7-list-item>
            <f7-list-item :title="tt('Closing Balance')" after="Count"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical" v-if="finishQuery && !loading">
            <f7-list-item :title="tt('Opening Balance')" :after="displayOpeningBalance"></f7-list-item>
            <f7-list-item :title="tt('Closing Balance')" :after="displayClosingBalance"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers media-list
                 class="skeleton-text margin-vertical transaction-info-list reconciliation-statement-list"
                 v-if="finishQuery && !showAccountBalanceTrendsCharts && loading">
            <ul>
                <f7-list-item chevron-center
                    :key="index"
                    :class="{ 'transaction-info': type === 't', 'last-transaction-of-day': index === 2, 'reconciliation-statement-transaction-date': type === 'd' }"
                    :link="type === 't' ? '#' : null"
                    v-for="(type, index) in [ 'd', 't', 't', 'd', 't', 't', 't' ]"
                >
                    <div class="display-flex no-padding-horizontal" v-if="type === 'd'">
                        <div class="actual-item-inner">
                            <div class="item-title-row">
                                <div class="item-title">
                                    <small>yyyy-MM-dd</small>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="display-flex no-padding-horizontal" v-if="type === 't'">
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
                                <div class="transaction-footer display-flex justify-content-space-between">
                                    <div class="flex-shrink-0">
                                        <span>HH:mm</span>
                                    </div>
                                    <div class="account-balance flex-shrink-1">
                                        <span>Balance</span>
                                        <span style="margin-inline-start: 4px">0.00 USD</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </f7-list-item>
            </ul>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical"
                 v-if="finishQuery && !showAccountBalanceTrendsCharts && !loading && (!allReconciliationStatementVirtualListItems || !allReconciliationStatementVirtualListItems.length)">
            <f7-list-item :title="tt('No transaction data')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers media-list virtual-list
                 class="margin-vertical transaction-info-list reconciliation-statement-list"
                 :virtual-list-params="{ items: allReconciliationStatementVirtualListItems, renderExternal, height: 'auto' }"
                 v-if="finishQuery && !showAccountBalanceTrendsCharts && !loading && allReconciliationStatementVirtualListItems && allReconciliationStatementVirtualListItems.length">
            <ul>
                <f7-list-item chevron-center
                              :key="item.index"
                              :id="item.transaction ? getTransactionDomId(item.transaction) : undefined"
                              :class="{ 'transaction-info': item.type == 'transaction', 'last-transaction-of-day': allReconciliationStatementVirtualListItems[item.index + 1] && allReconciliationStatementVirtualListItems[item.index + 1]!.type === 'date', 'reconciliation-statement-transaction-date': item.type == 'date' }"
                              :style="`top: ${virtualDataItems.topPosition}px`"
                              :virtual-list-index="item.index"
                              :swipeout="item.type === 'transaction' && !!item.transaction"
                              :accordion-item="item.type === 'transaction' && !!item.transaction"
                              :link="item.type == 'transaction' && item.transaction && item.transaction.type !== TransactionType.ModifyBalance ? `/transaction/detail?id=${item.transaction?.id}&type=${item.transaction.type}` : null"
                              v-for="item in virtualDataItems.items"
                >
                    <template #inner>
                        <div class="display-flex no-padding-horizontal" v-if="item.type == 'date' && item.displayDate">
                            <div class="actual-item-inner">
                                <div class="item-title-row">
                                    <div class="item-title">
                                        <small>{{ item.displayDate }}</small>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="display-flex no-padding-horizontal" v-if="item.type == 'transaction' && item.transaction">
                            <div class="item-media">
                                <div class="transaction-icon display-flex align-items-center">
                                    <ItemIcon :icon-type="getCategoryIconType(item.transaction.category?.iconType)"
                                              :icon-id="item.transaction.category?.icon"
                                              :color="item.transaction.category?.color"
                                              v-if="item.transaction.category && item.transaction.category?.color"></ItemIcon>
                                    <f7-icon v-else-if="!item.transaction.category || !item.transaction.category?.color"
                                             f7="pencil_ellipsis_rectangle">
                                    </f7-icon>
                                </div>
                            </div>
                            <div class="actual-item-inner">
                                <div class="item-title-row">
                                    <div class="item-title">
                                        <div class="transaction-category-name no-padding">
                                        <span v-if="item.transaction.type === TransactionType.ModifyBalance">
                                            {{ tt('Modify Balance') }}
                                        </span>
                                            <span v-else-if="item.transaction.type !== TransactionType.ModifyBalance && item.transaction.category">
                                            {{ item.transaction.category.name }}
                                        </span>
                                        </div>
                                    </div>
                                    <div class="item-after">
                                        <div class="transaction-amount"
                                             :class="{ 'text-expense': item.transaction.type === TransactionType.Expense, 'text-income': item.transaction.type === TransactionType.Income }">
                                            <span v-if="item.transaction.type === TransactionType.Transfer && item.transaction.destinationAccountId === accountId">{{ getDisplayDestinationAmount(item.transaction) }}</span>
                                            <span v-else-if="item.transaction.type !== TransactionType.Transfer || item.transaction.destinationAccountId !== accountId">{{ getDisplaySourceAmount(item.transaction) }}</span>
                                        </div>
                                    </div>
                                </div>
                                <div class="item-text">
                                    <div class="transaction-description" v-if="item.transaction.comment">
                                        <span>{{ item.transaction.comment }}</span>
                                    </div>
                                </div>
                                <div class="item-footer">
                                    <div class="transaction-footer display-flex justify-content-space-between">
                                        <div class="flex-shrink-0">
                                            <span>{{ getDisplayTime(item.transaction) }}</span>
                                            <span style="margin-inline-start: 4px" v-if="!isSameAsDefaultTimezoneOffsetMinutes(item.transaction)">{{ `(${getDisplayTimezone(item.transaction)})` }}</span>
                                        </div>
                                        <div class="account-balance flex-shrink-1">
                                            <span>{{ isCurrentLiabilityAccount ? tt('Outstanding Balance') : tt('Balance') }}</span>
                                            <span style="margin-inline-start: 4px">{{ getDisplayAccountBalance(item.transaction) }}</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </template>
                    <f7-swipeout-actions :left="textDirection === TextDirection.RTL"
                                         :right="textDirection === TextDirection.LTR"
                                         v-if="item.type == 'transaction' && item.transaction">
                        <f7-swipeout-button color="primary" close
                                            :text="tt('Duplicate')"
                                            v-if="item.transaction.type !== TransactionType.ModifyBalance"
                                            @click="duplicateTransaction(item.transaction)"></f7-swipeout-button>
                        <f7-swipeout-button color="orange" close
                                            :text="tt('Edit')"
                                            v-if="item.transaction.editable && item.transaction.type !== TransactionType.ModifyBalance"
                                            @click="editTransaction(item.transaction)"></f7-swipeout-button>
                        <f7-swipeout-button color="red" class="padding-horizontal" :aria-label="tt('Delete')"
                                            v-if="item.transaction.editable"
                                            @click="removeTransaction(item.transaction, false)">
                            <f7-icon f7="trash"></f7-icon>
                        </f7-swipeout-button>
                    </f7-swipeout-actions>
                </f7-list-item>
            </ul>
        </f7-list>

        <f7-card v-if="finishQuery && showAccountBalanceTrendsCharts">
            <f7-card-header class="no-border display-block">
                <div class="statistics-chart-header display-flex width-100 justify-content-space-between">
                    <div></div>
                    <div class="align-self-flex-end">
                        <span style="margin-inline-end: 4px;">{{ tt('Time Granularity') }}</span>
                        <f7-link :class="{ 'disabled': loading || updatingLastReconciledTime }" href="#" popover-open=".chart-data-date-aggregation-type-popover-menu">{{ chartDataDateAggregationTypeDisplayName }}</f7-link>
                    </div>
                </div>
            </f7-card-header>
            <f7-card-content style="margin-top: -14px" :padding="false">
                <account-balance-trends-bar-chart
                    :loading="loading"
                    :date-aggregation-type="chartDataDateAggregationType"
                    :timezone-used-for-date-range="timezoneUsedForDateRange"
                    :fiscal-year-start="fiscalYearStart"
                    :items="reconciliationStatements?.transactions"
                    :account="currentAccount"
                    :statement-date="currentAccountStatementDate"
                />
            </f7-card-content>
        </f7-card>

        <f7-popover class="chart-data-date-aggregation-type-popover-menu">
            <f7-list dividers>
                <f7-list-item group-title>
                    <small>{{ tt('Time Granularity') }}</small>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :title="dateAggregationType.displayName"
                              :class="{ 'list-item-selected': chartDataDateAggregationType === dateAggregationType.type }"
                              :key="dateAggregationType.type"
                              v-for="dateAggregationType in allDateAggregationTypes"
                              @click="setChartDataDateAggregationType(dateAggregationType.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="chartDataDateAggregationType === dateAggregationType.type"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item group-title>
                    <small>{{ tt('Timezone Used for Date Range') }}</small>
                </f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :title="timezoneType.displayName"
                              :class="{ 'list-item-selected': timezoneUsedForDateRange === timezoneType.type }"
                              :key="timezoneType.type"
                              v-for="timezoneType in allTimezoneTypesUsedForDateRange"
                              @click="setTimezoneUsedForDateRange(timezoneType.type)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="timezoneUsedForDateRange === timezoneType.type"></f7-icon>
                    </template>
                </f7-list-item>

            </f7-list>
        </f7-popover>

        <date-range-selection-sheet :title="tt('Custom Date Range')"
                                    :min-time="startTime"
                                    :max-time="endTime"
                                    v-model:show="showCustomDateRangeSheet"
                                    @dateRange:change="changeCustomDateFilter">
        </date-range-selection-sheet>

        <number-pad-sheet :min-value="TRANSACTION_MIN_AMOUNT"
                          :max-value="TRANSACTION_MAX_AMOUNT"
                          :currency="currentAccountCurrency"
                          :hint="tt('Please enter the new closing balance for this account')"
                          v-model:show="showNewClosingBalanceSheet"
                          v-model="newClosingBalance"
                          @update:model-value="updateClosingBalance"
        ></number-pad-sheet>

        <f7-actions close-by-outside-click close-on-escape :opened="showMoreActionSheet" @actions:closed="showMoreActionSheet = false">
            <f7-actions-group>
                <f7-actions-button :class="{ 'disabled': loading || updatingLastReconciledTime }" @click="addTransaction()">{{ tt('Add Transaction') }}</f7-actions-button>
                <f7-actions-button :class="{ 'disabled': loading || updatingLastReconciledTime }" @click="updateClosingBalance(undefined)" v-if="canUpdateAccountCloseBalance">{{ tt('Update Closing Balance') }}</f7-actions-button>
                <f7-actions-button :class="{ 'disabled': loading || updatingLastReconciledTime }" @click="updateLastReconciledTime()" v-if="newLastReconciledTime">{{ tt('Mark as Reconciled') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button :class="{ 'disabled': loading || updatingLastReconciledTime }" @click="reload(true)">{{ tt('Refresh') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteActionSheet" @actions:closed="showDeleteActionSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this transaction?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="removeTransaction(transactionToDelete, true)">{{ tt('Delete') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>



<style>
.list.reconciliation-statement-list li.reconciliation-statement-transaction-date:first-child {
    border-radius: var(--f7-list-inset-border-radius) var(--f7-list-inset-border-radius) 0 0;
}

.list.reconciliation-statement-list li.reconciliation-statement-transaction-date {
    display: flex;
    align-items: center;
    align-content: center;
    padding-top: 0;
    padding-bottom: 0;
    height: var(--f7-list-group-title-height);
    line-height: var(--f7-list-group-title-height);
    background-color: var(--f7-list-group-title-bg-color);
}

.list.reconciliation-statement-list li.reconciliation-statement-transaction-date > .item-content {
    padding-inline-start: 0 !important;
}

.list.reconciliation-statement-list li.reconciliation-statement-transaction-date > .item-content > .item-inner {
    padding-inline-start: calc(var(--f7-list-item-padding-horizontal) + var(--f7-safe-area-left));
}

.list.reconciliation-statement-list li.reconciliation-statement-transaction-date > .item-content > .item-inner:after {
    background-color: inherit;
}

.list.reconciliation-statement-list li.transaction-info.last-transaction-of-day > .item-link > .item-content > .item-inner:after,
.list.reconciliation-statement-list li.transaction-info.last-transaction-of-day > .swipeout-content > .item-link > .item-content > .item-inner:after {
    background-color: inherit;
}

.list.reconciliation-statement-list li.transaction-info .account-balance {
    margin-inline-start: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
