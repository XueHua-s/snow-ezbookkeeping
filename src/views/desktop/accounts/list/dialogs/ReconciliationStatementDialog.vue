<script setup lang="ts">
import AmountInputDialog from '@/components/desktop/AmountInputDialog.vue';
import OneColumnDialogLayout from '@/components/desktop/OneColumnDialogLayout.vue';
import PaginationButtons from '@/components/desktop/PaginationButtons.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { KnownFileType } from '@/core/file.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    extendMdiSemicolon
} from '@/exticons/desktop/extend_mdi_icons.ts';
import { getCategoryIconType } from '@/lib/icon.ts';
import { focusTableScrollContainer } from '@/lib/ui/desktop.ts';
import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';
import {
    mdiArrowRight,
    mdiCheck,
    mdiComma,
    mdiDotsVertical,
    mdiInvoiceTextEditOutline,
    mdiInvoiceTextPlusOutline,
    mdiKeyboardTab,
    mdiPencilBoxOutline,
    mdiRefresh,
    mdiTuneVertical
} from '@mdi/js';
import { useReconciliationStatementDialog } from './ReconciliationStatementDialog/useReconciliationStatementDialog.ts';

const emit = defineEmits<{
    (e: 'update:last-reconciled-time', newLastReconciledTime: number): void;
    (e: 'error', message: string): void;
}>();

const {
    tt, formatRange, formatNumberToLocalizedNumerals, startTime, endTime, reconciliationStatements,
    chartDataDateAggregationType, timezoneUsedForDateRange, fiscalYearStart, defaultCurrency, allChartTypes,
    allDateAggregationTypes, allTimezoneTypesUsedForDateRange, currentAccount, currentAccountCurrency,
    currentAccountStatementDate, isCurrentLiabilityAccount, canUpdateAccountCloseBalance, newLastReconciledTime,
    displayStartDateTime, displayEndDateTime, displayTotalInflows, displayTotalInflowsInDefaultCurrency, displayTotalOutflows,
    displayTotalOutflowsInDefaultCurrency, displayTotalBalance, displayTotalBalanceInDefaultCurrency, displayOpeningBalance,
    displayOpeningBalanceInDefaultCurrency, displayClosingBalance, displayClosingBalanceInDefaultCurrency,
    getDisplayTransactionType, getDisplayDateTime, isSameAsDefaultTimezoneOffsetMinutes, getDisplayTimezone,
    getDisplayTimeInDefaultTimezone, getDisplaySourceAmount, getDisplayDestinationAmount, getDisplayAccountBalance,
    chartTypeIconMap, chartDataDateAggregationTypeIconMap, timezoneTypeIconMap, showState, loading,
    updatingLastReconciledTime, currentPage, countPerPage, showAccountBalanceTrendsCharts, chartType,
    transactionListDialogHeight, useCommaDecimalSeparator, reconciliationStatementsTablePageOptions, preserveDialogHeight,
    totalPageCount, dataTableHeaders, getTransactionTypeColor, open, reload, addTransaction, updateClosingBalance,
    exportReconciliationStatements, showTransaction, updateLastReconciledTime, close,
} = useReconciliationStatementDialog(emit);

defineExpose({
    open
});

</script>

<template>
    <v-dialog :persistent="loading || updatingLastReconciledTime" v-model="showState">
        <one-column-dialog-layout ref="dialogLayout" class="reconciliation-statement-dialog-layout d-flex flex-column"
                                  content-class="reconciliation-statement-dialog-content pa-0 d-flex flex-column" footer-class="py-1"
                                  :style="{ '--reconciliation-statement-dialog-height': preserveDialogHeight ? transactionListDialogHeight + 'px' : undefined }"
                                  :disabled="loading || updatingLastReconciledTime"
                                  :title="tt('Reconciliation Statement')" :cancel-button-title="tt('Close')"
                                  @cancel="close">
            <template #after-title>
                <v-btn density="compact" color="default" variant="text" class="ms-2"
                       :aria-label="tt('Refresh')" :icon="true" :disabled="updatingLastReconciledTime"
                       :loading="loading" @click="reload(true)">
                    <template #loader>
                        <v-progress-circular indeterminate size="20"/>
                    </template>
                    <v-icon :icon="mdiRefresh" size="22" />
                    <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                </v-btn>

                <v-btn density="compact" color="primary" variant="outlined" class="ms-2"
                       :disabled="loading || updatingLastReconciledTime" @click="updateLastReconciledTime"
                       v-if="newLastReconciledTime">
                    {{ tt('Mark as Reconciled') }}
                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="updatingLastReconciledTime"></v-progress-circular>
                </v-btn>
            </template>

            <template #toolbar>
                <toggle-button class="ms-2" :disabled="loading || updatingLastReconciledTime || !reconciliationStatements?.transactions?.length"
                               :false-name="tt('Transaction List')" :true-name="tt('Account Balance Trends')"
                               v-model="showAccountBalanceTrendsCharts"/>

                <v-btn density="compact" color="default" variant="text" class="ms-2" :aria-label="tt('Settings')"
                       :icon="true" :disabled="loading || updatingLastReconciledTime || !showAccountBalanceTrendsCharts">
                    <v-icon :icon="mdiTuneVertical" size="22" />
                    <v-menu activator="parent">
                        <v-list>
                            <v-list-subheader class="text-body-small" :title="tt('Chart Type')"/>
                            <v-list-item :key="type.type"
                                         :prepend-icon="chartTypeIconMap[type.type]"
                                         :append-icon="chartType === type.type ? mdiCheck : undefined"
                                         :title="type.displayName"
                                         @click="chartType = type.type"
                                         v-for="type in allChartTypes"></v-list-item>
                            <v-divider class="my-2"/>
                            <v-list-subheader class="text-body-small" :title="tt('Time Granularity')"/>
                            <v-list-item :key="dateAggregationType.type"
                                         :prepend-icon="chartDataDateAggregationTypeIconMap[dateAggregationType.type]"
                                         :append-icon="chartDataDateAggregationType === dateAggregationType.type ? mdiCheck : undefined"
                                         :title="dateAggregationType.displayName"
                                         @click="chartDataDateAggregationType = dateAggregationType.type"
                                         v-for="dateAggregationType in allDateAggregationTypes"></v-list-item>
                            <v-divider class="my-2"/>
                            <v-list-subheader class="text-body-small" :title="tt('Timezone Used for Date Range')"/>
                            <v-list-item :key="timezoneType.type" :value="timezoneType.type"
                                         :prepend-icon="timezoneTypeIconMap[timezoneType.type]"
                                         :append-icon="timezoneUsedForDateRange === timezoneType.type ? mdiCheck : undefined"
                                         :title="timezoneType.displayName"
                                         v-for="timezoneType in allTimezoneTypesUsedForDateRange"
                                         @click="timezoneUsedForDateRange = timezoneType.type"></v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-btn density="compact" color="default" variant="text" class="ms-2" :aria-label="tt('More')"
                       :icon="true" :disabled="loading || updatingLastReconciledTime">
                    <v-icon :icon="mdiDotsVertical" size="22" />
                    <v-menu activator="parent">
                        <v-list>
                            <v-list-item :prepend-icon="mdiInvoiceTextPlusOutline"
                                         :title="tt('Add Transaction')"
                                         @click="addTransaction()"></v-list-item>
                            <v-list-item :prepend-icon="mdiInvoiceTextEditOutline"
                                         :title="tt('Update Closing Balance')"
                                         @click="updateClosingBalance()"
                                         v-if="canUpdateAccountCloseBalance"></v-list-item>
                            <v-divider class="my-2"/>
                            <v-list-item :prepend-icon="mdiComma"
                                         :disabled="!reconciliationStatements || !reconciliationStatements.transactions || reconciliationStatements.transactions.length < 1 || useCommaDecimalSeparator"
                                         @click="exportReconciliationStatements(KnownFileType.CSV)">
                                <v-list-item-title>{{ tt('Save as CSV (Comma-separated values) File') }}</v-list-item-title>
                            </v-list-item>
                            <v-list-item :prepend-icon="mdiKeyboardTab"
                                         :disabled="!reconciliationStatements || !reconciliationStatements.transactions || reconciliationStatements.transactions.length < 1"
                                         @click="exportReconciliationStatements(KnownFileType.TSV)">
                                <v-list-item-title>{{ tt('Save as TSV (Tab-separated values) File') }}</v-list-item-title>
                            </v-list-item>
                            <v-list-item :prepend-icon="extendMdiSemicolon"
                                         :disabled="!reconciliationStatements || !reconciliationStatements.transactions || reconciliationStatements.transactions.length < 1"
                                         @click="exportReconciliationStatements(KnownFileType.SSV)">
                                <v-list-item-title>{{ tt('Save as SSV (Semicolon-separated values) File') }}</v-list-item-title>
                            </v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
            </template>

            <template #subtitle>
                <v-divider class="mt-2"/>
                <div class="mt-2 mx-4">
                    <div class="text-body-large text-wrap" v-if="!startTime && !endTime">
                        <span>{{ tt('All') }}</span>
                    </div>
                    <div class="text-body-large text-wrap" v-if="startTime || endTime">
                        <span>{{ formatRange(displayStartDateTime, displayEndDateTime) }}</span>
                    </div>

                    <div class="d-flex align-center mt-1 overflow-x-auto">
                        <div class="d-flex align-center text-body-large">
                            <span>{{ tt('Opening Balance') }}</span>
                            <span class="text-primary" v-if="loading">
                            <v-skeleton-loader class="skeleton-no-margin ms-2" type="text" style="width: 80px" :loading="true"></v-skeleton-loader>
                        </span>
                            <span class="text-primary ms-2" v-else-if="!loading">
                            {{ displayOpeningBalance }}
                            <v-tooltip activator="parent" v-if="currentAccountCurrency !== defaultCurrency">
                                <span>{{ displayOpeningBalanceInDefaultCurrency }}</span>
                            </v-tooltip>
                        </span>
                            <span class="ms-3">{{ tt('Closing Balance') }}</span>
                            <span class="text-primary" v-if="loading">
                            <v-skeleton-loader class="skeleton-no-margin ms-2" type="text" style="width: 80px" :loading="true"></v-skeleton-loader>
                        </span>
                            <span class="text-primary ms-2" v-else-if="!loading">
                            {{ displayClosingBalance }}
                            <v-tooltip activator="parent" v-if="currentAccountCurrency !== defaultCurrency">
                                <span>{{ displayClosingBalanceInDefaultCurrency }}</span>
                            </v-tooltip>
                        </span>
                        </div>
                        <v-spacer/>
                        <div class="d-flex align-center text-body-large">
                            <span class="ms-2">{{ tt('Total Inflows') }}</span>
                            <span class="text-income" v-if="loading">
                            <v-skeleton-loader class="skeleton-no-margin ms-2" type="text" style="width: 80px" :loading="true"></v-skeleton-loader>
                        </span>
                            <span class="text-income ms-2" v-else-if="!loading">
                            {{ displayTotalInflows }}
                            <v-tooltip activator="parent" v-if="currentAccountCurrency !== defaultCurrency">
                                <span>{{ displayTotalInflowsInDefaultCurrency }}</span>
                            </v-tooltip>
                        </span>
                            <span class="ms-3">{{ tt('Total Outflows') }}</span>
                            <span class="text-expense" v-if="loading">
                            <v-skeleton-loader class="skeleton-no-margin ms-2" type="text" style="width: 80px" :loading="true"></v-skeleton-loader>
                        </span>
                            <span class="text-expense ms-2" v-else-if="!loading">
                            {{ displayTotalOutflows }}
                            <v-tooltip activator="parent" v-if="currentAccountCurrency !== defaultCurrency">
                                <span>{{ displayTotalOutflowsInDefaultCurrency }}</span>
                            </v-tooltip>
                        </span>
                            <span class="ms-3">{{ tt('Net Cash Flow') }}</span>
                            <span class="text-primary" v-if="loading">
                            <v-skeleton-loader class="skeleton-no-margin ms-2" type="text" style="width: 80px" :loading="true"></v-skeleton-loader>
                        </span>
                            <span class="text-primary ms-2" v-else-if="!loading">
                            {{ displayTotalBalance }}
                            <v-tooltip activator="parent" v-if="currentAccountCurrency !== defaultCurrency">
                                <span>{{ displayTotalBalanceInDefaultCurrency }}</span>
                            </v-tooltip>
                        </span>
                        </div>
                    </div>
                </div>
            </template>

            <template #content>
                <v-data-table
                    fixed-header
                    fixed-footer
                    height="100%"
                    multi-sort
                    density="compact"
                    item-value="index"
                    :class="{ 'reconciliation-statement-table': true, 'disabled': loading }"
                    :headers="dataTableHeaders"
                    :items="reconciliationStatements?.transactions ?? []"
                    :hover="true"
                    :no-data-text="loading ? '' : tt('No transaction data')"
                    v-model:items-per-page="countPerPage"
                    v-model:page="currentPage"
                    @click="focusTableScrollContainer"
                    v-if="!showAccountBalanceTrendsCharts"
                >
                    <template #item.time="{ item }">
                        <div class="d-flex align-center">
                            <span>{{ getDisplayDateTime(item) }}</span>
                            <v-chip class="ms-1" variant="flat" color="grey" size="x-small"
                                    v-if="!isSameAsDefaultTimezoneOffsetMinutes(item)">{{ getDisplayTimezone(item) }}</v-chip>
                            <v-tooltip activator="parent" v-if="!isSameAsDefaultTimezoneOffsetMinutes(item)">{{ getDisplayTimeInDefaultTimezone(item) }}</v-tooltip>
                        </div>
                    </template>
                    <template #item.type="{ item }">
                        <v-chip label variant="outlined" size="x-small"
                                :class="{ 'text-income' : item.type === TransactionType.Income, 'text-expense': item.type === TransactionType.Expense }"
                                :color="getTransactionTypeColor(item)">{{ getDisplayTransactionType(item) }}</v-chip>
                    </template>
                    <template #item.categoryName="{ item }">
                        <div class="d-flex align-center">
                            <ItemIcon size="24px" :icon-type="getCategoryIconType(item.category?.iconType)"
                                      :icon-id="item.category?.icon ?? ''"
                                      :color="item.category?.color ?? ''"
                                      v-if="item.category && item.category?.color"></ItemIcon>
                            <v-icon size="24" :icon="mdiPencilBoxOutline" v-else-if="!item.category || !item.category?.color" />
                            <span class="ms-1" v-if="item.type === TransactionType.ModifyBalance">
                                {{ tt('Modify Balance') }}
                            </span>
                            <span class="ms-1" v-else-if="item.type !== TransactionType.ModifyBalance && item.category">
                                {{ item.category?.name }}
                            </span>
                        </div>
                    </template>
                    <template #item.sourceAmount="{ item }">
                        <span :class="{ 'text-expense': item.type === TransactionType.Expense, 'text-income': item.type === TransactionType.Income }">{{ getDisplaySourceAmount(item) }}</span>
                        <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" v-if="item.type === TransactionType.Transfer && item.sourceAccountId !== item.destinationAccountId && getDisplaySourceAmount(item) !== getDisplayDestinationAmount(item)"></v-icon>
                        <span v-if="item.type === TransactionType.Transfer && item.sourceAccountId !== item.destinationAccountId && getDisplaySourceAmount(item) !== getDisplayDestinationAmount(item)">{{ getDisplayDestinationAmount(item) }}</span>
                        <v-tooltip activator="parent" v-if="!item.hideAmount && ((item.type !== TransactionType.Transfer && item.sourceAccount?.currency !== defaultCurrency) || (item.type === TransactionType.Transfer && item.sourceAccount?.currency !== defaultCurrency && item.destinationAccount?.currency !== defaultCurrency))">
                            <span>{{ getDisplaySourceAmount(item, true) }}</span>
                            <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" v-if="item.type === TransactionType.Transfer && item.sourceAccount?.id !== item.destinationAccount?.id && item.sourceAccount?.currency !== item.destinationAccount?.currency && item.sourceAmount !== item.destinationAmount"></v-icon>
                            <span v-if="item.type === TransactionType.Transfer && item.sourceAccount?.id !== item.destinationAccount?.id && item.sourceAccount?.currency !== item.destinationAccount?.currency && item.sourceAmount !== item.destinationAmount">{{ getDisplayDestinationAmount(item, true) }}</span>
                        </v-tooltip>
                    </template>
                    <template #item.sourceAccountName="{ item }">
                        <div class="d-flex align-center">
                            <span v-if="item.sourceAccount">{{ item.sourceAccount?.name }}</span>
                            <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" v-if="item.type === TransactionType.Transfer"></v-icon>
                            <span v-if="item.type === TransactionType.Transfer && item.destinationAccount">{{ item.destinationAccount?.name }}</span>
                        </div>
                    </template>
                    <template #item.accountBalance="{ item }">
                        <span>{{ getDisplayAccountBalance(item) }}</span>
                        <v-tooltip activator="parent" v-if="currentAccountCurrency !== defaultCurrency">
                            <span>{{ getDisplayAccountBalance(item, true) }}</span>
                        </v-tooltip>
                    </template>
                    <template #item.operation="{ item }">
                        <v-btn density="compact" variant="text" color="default" :disabled="loading || updatingLastReconciledTime || item.type === TransactionType.ModifyBalance"
                               @click="showTransaction(item)">
                            {{ tt('View') }}
                        </v-btn>
                    </template>
                    <template #no-data>
                        <div class="d-flex flex-column gap-1 mt-2 mb-4" v-if="loading && (!reconciliationStatements || !reconciliationStatements.transactions || !reconciliationStatements.transactions.length)">
                            <v-skeleton-loader class="skeleton-no-margin mt-3 mb-2" type="text" :loading="true"
                                               :key="idx" v-for="idx in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]"></v-skeleton-loader>
                        </div>
                        <div class="my-5" v-else>
                            {{ tt('No transaction data') }}
                        </div>
                    </template>
                    <template #bottom></template>
                </v-data-table>

                <div class="reconciliation-statement-chart-container d-flex flex-grow-1 pa-4" v-if="showAccountBalanceTrendsCharts">
                    <account-balance-trends-chart
                        :type="chartType"
                        :date-aggregation-type="chartDataDateAggregationType"
                        :timezone-used-for-date-range="timezoneUsedForDateRange"
                        :fiscal-year-start="fiscalYearStart"
                        :items="[]"
                        :legend-name="isCurrentLiabilityAccount ? tt('Account Outstanding Balance') : tt('Account Balance')"
                        :account="currentAccount"
                        :statement-date="currentAccountStatementDate"
                        :skeleton="true"
                        v-if="showAccountBalanceTrendsCharts && loading"
                    />

                    <account-balance-trends-chart
                        :type="chartType"
                        :date-aggregation-type="chartDataDateAggregationType"
                        :timezone-used-for-date-range="timezoneUsedForDateRange"
                        :fiscal-year-start="fiscalYearStart"
                        :items="reconciliationStatements?.transactions"
                        :legend-name="isCurrentLiabilityAccount ? tt('Account Outstanding Balance') : tt('Account Balance')"
                        :account="currentAccount"
                        :statement-date="currentAccountStatementDate"
                        v-if="showAccountBalanceTrendsCharts && !loading"
                    />
                </div>
            </template>

            <template #footer v-if="!showAccountBalanceTrendsCharts && (loading || (reconciliationStatements && reconciliationStatements.transactions && reconciliationStatements.transactions.length))">
                <div class="title-and-toolbar d-flex w-100 text-body-large align-center text-no-wrap">
                    <span>{{ tt('Total Transactions') }}</span>
                    <span v-if="loading">
                        <v-skeleton-loader class="skeleton-no-margin ms-3" type="text" style="width: 80px" :loading="true"></v-skeleton-loader>
                    </span>
                    <span class="ms-2" v-else-if="!loading">
                        {{ formatNumberToLocalizedNumerals(reconciliationStatements?.transactions.length ?? 0) }}
                    </span>
                    <v-spacer/>
                    <span v-if="reconciliationStatements && reconciliationStatements.transactions && reconciliationStatements.transactions.length > 10">
                        {{ tt('Transactions Per Page') }}
                    </span>
                    <v-select class="ms-2" density="compact" max-width="100"
                              item-title="name"
                              item-value="value"
                              :disabled="loading"
                              :items="reconciliationStatementsTablePageOptions"
                              v-model="countPerPage"
                              v-if="reconciliationStatements && reconciliationStatements.transactions && reconciliationStatements.transactions.length > 10"
                    />
                    <pagination-buttons density="comfortable"
                                        :disabled="loading"
                                        :totalPageCount="totalPageCount"
                                        v-model="currentPage"
                                        v-if="reconciliationStatements && reconciliationStatements.transactions && reconciliationStatements.transactions.length > 10">
                    </pagination-buttons>
                </div>
            </template>
        </one-column-dialog-layout>
    </v-dialog>

    <amount-input-dialog ref="amountInputDialog" />
    <edit-dialog ref="editDialog" :type="TransactionEditPageType.Transaction" />

    <snack-bar ref="snackbar" />
</template>



<style>
.reconciliation-statement-dialog-layout {
    min-height: min(max(var(--reconciliation-statement-dialog-height), 581px), calc(100dvh - 48px));

    .reconciliation-statement-dialog-content {
        min-height: 0;
        overflow-y: hidden !important;

        .v-table.reconciliation-statement-table {
            min-height: 0;
            flex: 1 1 auto;

            > .v-table__wrapper > table {
                th:not(:nth-last-child(2)),
                td:not(:nth-last-child(2)) {
                    width: auto !important;
                    white-space: nowrap;
                }

                th:nth-last-child(2),
                td:nth-last-child(2) {
                    width: 100% !important;
                }
            }
        }

        .reconciliation-statement-chart-container > .account-balance-trends-chart-container {
            flex: 1 1 auto;
            height: auto !important;
            min-height: 422px;
        }
    }
}
</style>
