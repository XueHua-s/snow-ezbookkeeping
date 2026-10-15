import AmountInputDialog from '@/components/desktop/AmountInputDialog.vue';
import OneColumnDialogLayout from '@/components/desktop/OneColumnDialogLayout.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { DEFAULT_PAGE_COUNTS } from '@/consts/page.ts';
import { TRANSACTION_MAX_AMOUNT, TRANSACTION_MIN_AMOUNT } from '@/consts/transaction.ts';
import type { NameNumeralValue } from '@/core/base.ts';
import { KnownFileType } from '@/core/file.ts';
import { type BigDecimal, DecimalSeparator } from '@/core/numeral.ts';
import { AccountBalanceTrendChartType, ChartDateAggregationType } from '@/core/statistics.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import { TransactionType } from '@/core/transaction.ts';
import { getCurrentUnixTime } from '@/lib/datetime.ts';
import { BIG_DECIMAL_ZERO, parseBigDecimal } from '@/lib/numeral.ts';
import { startDownloadFile } from '@/lib/ui/common.ts';
import { useI18n } from '@/locales/helpers.ts';
import { Transaction, type TransactionReconciliationStatementResponseItem } from '@/models/transaction.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useReconciliationStatementPageBase } from '@/views/base/accounts/ReconciliationStatementPageBase.ts';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';
import {
    mdiCalendarMonthOutline,
    mdiCalendarTextOutline,
    mdiCalendarTodayOutline,
    mdiChartAreasplineVariant,
    mdiChartBar,
    mdiChartWaterfall,
    mdiHomeClockOutline,
    mdiInvoiceTextClockOutline,
    mdiLayersTripleOutline
} from '@mdi/js';
import { computed, ref, useTemplateRef, watch } from 'vue';

export type OneColumnDialogLayoutType = InstanceType<typeof OneColumnDialogLayout>;

export type SnackBarType = InstanceType<typeof SnackBar>;

export type AmountInputDialogType = InstanceType<typeof AmountInputDialog>;

export type EditDialogType = InstanceType<typeof EditDialog>;

// Coordinates ReconciliationStatementDialog state and effects; called once from the component setup.
export function useReconciliationStatementDialog(emit: {
    (e: 'update:last-reconciled-time', newLastReconciledTime: number): void;
    (e: 'error', message: string): void;
}) {
    const {
        tt,
        formatRange,
        getCurrentDecimalSeparator,
        formatNumberToLocalizedNumerals,
        getTablePageOptions
    } = useI18n();

    const {
        accountId,
        startTime,
        endTime,
        reconciliationStatements,
        chartDataDateAggregationType,
        timezoneUsedForDateRange,
        fiscalYearStart,
        defaultCurrency,
        allChartTypes,
        allDateAggregationTypes,
        allTimezoneTypesUsedForDateRange,
        currentAccount,
        currentAccountCurrency,
        currentAccountStatementDate,
        isCurrentLiabilityAccount,
        canUpdateAccountCloseBalance,
        newLastReconciledTime,
        exportFileName,
        displayStartDateTime,
        displayEndDateTime,
        displayTotalInflows,
        displayTotalInflowsInDefaultCurrency,
        displayTotalOutflows,
        displayTotalOutflowsInDefaultCurrency,
        displayTotalBalance,
        displayTotalBalanceInDefaultCurrency,
        displayOpeningBalance,
        displayOpeningBalanceInDefaultCurrency,
        displayClosingBalance,
        displayClosingBalanceInDefaultCurrency,
        updatePageOpenTime,
        isReconciliationStatementsResponseEquals,
        setReconciliationStatements,
        getDisplayTransactionType,
        getDisplayDateTime,
        isSameAsDefaultTimezoneOffsetMinutes,
        getDisplayTimezone,
        getDisplayTimeInDefaultTimezone,
        getDisplaySourceAmount,
        getDisplayDestinationAmount,
        getDisplayAccountBalance,
        getExportedData
    } = useReconciliationStatementPageBase();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionsStore = useTransactionsStore();

    const chartTypeIconMap = {
        [AccountBalanceTrendChartType.Column.type]: mdiChartBar,
        [AccountBalanceTrendChartType.Area.type]: mdiChartAreasplineVariant,
        [AccountBalanceTrendChartType.SmoothArea.type]: mdiChartAreasplineVariant,
        [AccountBalanceTrendChartType.Boxplot.type]: mdiChartWaterfall,
        [AccountBalanceTrendChartType.Candlestick.type]: mdiChartWaterfall,
    };

    const chartDataDateAggregationTypeIconMap = {
        [ChartDateAggregationType.Day.type]: mdiCalendarTodayOutline,
        [ChartDateAggregationType.Month.type]: mdiCalendarMonthOutline,
        [ChartDateAggregationType.Quarter.type]: mdiLayersTripleOutline,
        [ChartDateAggregationType.Year.type]: mdiLayersTripleOutline,
        [ChartDateAggregationType.FiscalYear.type]: mdiLayersTripleOutline,
        [ChartDateAggregationType.BillingCycle.type]: mdiCalendarTextOutline,
    };

    const timezoneTypeIconMap = {
        [TimezoneTypeForStatistics.ApplicationTimezone.type]: mdiHomeClockOutline,
        [TimezoneTypeForStatistics.TransactionTimezone.type]: mdiInvoiceTextClockOutline
    };

    let rejectFunc: ((reason?: unknown) => void) | null = null;

    const dialogLayout = useTemplateRef<OneColumnDialogLayoutType>('dialogLayout');

    const amountInputDialog = useTemplateRef<AmountInputDialogType>('amountInputDialog');

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const editDialog = useTemplateRef<EditDialogType>('editDialog');

    const showState = ref<boolean>(false);

    const loading = ref<boolean>(false);

    const updatingLastReconciledTime = ref<boolean>(false);

    const currentPage = ref<number>(1);

    const countPerPage = ref<number>(10);

    const showAccountBalanceTrendsCharts = ref<boolean>(false);

    const chartType = ref<number>(AccountBalanceTrendChartType.Default.type);

    const transactionListDialogHeight = ref<number>(0);

    const useCommaDecimalSeparator = computed<boolean>(() => getCurrentDecimalSeparator() === DecimalSeparator.Comma.symbol);

    const reconciliationStatementsTablePageOptions = computed<NameNumeralValue[]>(() => getTablePageOptions(DEFAULT_PAGE_COUNTS, reconciliationStatements.value?.transactions.length, true, false));

    const preserveDialogHeight = computed<boolean>(() => showAccountBalanceTrendsCharts.value && transactionListDialogHeight.value > 0);

    const totalPageCount = computed<number>(() => {
        if (!reconciliationStatements.value || !reconciliationStatements.value.transactions || reconciliationStatements.value.transactions.length < 1) {
            return 1;
        }

        const count = reconciliationStatements.value.transactions.length;
        return Math.ceil(count / countPerPage.value);
    });

    const dataTableHeaders = computed<object[]>(() => {
        const headers: object[] = [];
        const accountBalanceName = isCurrentLiabilityAccount.value ? 'Account Outstanding Balance' : 'Account Balance';

        headers.push({ key: 'time', value: 'time', title: tt('Transaction Time'), sortable: true, nowrap: true });
        headers.push({ key: 'type', value: 'type', title: tt('Type'), sortable: true, nowrap: true });
        headers.push({ key: 'categoryName', value: 'categoryName', title: tt('Category'), sortable: true, nowrap: true });
        headers.push({ key: 'sourceAmount', value: 'sourceAmount', title: tt('Amount'), sortable: true, nowrap: true });
        headers.push({ key: 'sourceAccountName', value: 'sourceAccountName', title: tt('Account'), sortable: true, nowrap: true });
        headers.push({ key: 'accountBalance', value: 'accountBalance', title: tt(accountBalanceName), sortable: true, nowrap: true });
        headers.push({ key: 'comment', value: 'comment', title: tt('Description'), sortable: true, nowrap: true });
        headers.push({ key: 'operation', title: tt('Operation'), sortable: false, nowrap: true, align: 'center' });
        return headers;
    });

    function getTransactionTypeColor(transaction: TransactionReconciliationStatementResponseItem): string | undefined {
        if (transaction.type === TransactionType.ModifyBalance) {
            return 'secondary';
        } else if (transaction.type === TransactionType.Income) {
            return undefined;
        } else if (transaction.type === TransactionType.Expense) {
            return undefined;
        } else if (transaction.type === TransactionType.Transfer) {
            return 'primary';
        } else {
            return 'default';
        }
    }

    function open(options: { accountId: string, startTime: number, endTime: number }): Promise<void> {
        updatePageOpenTime();
        accountId.value = options.accountId;
        startTime.value = options.startTime;
        endTime.value = options.endTime;
        reconciliationStatements.value = undefined;
        currentPage.value = 1;
        countPerPage.value = 10;
        showAccountBalanceTrendsCharts.value = false;
        chartType.value = AccountBalanceTrendChartType.Default.type;
        chartDataDateAggregationType.value = ChartDateAggregationType.Day.type;
        timezoneUsedForDateRange.value = TimezoneTypeForStatistics.ApplicationTimezone.type;
        showState.value = true;
        loading.value = true;

        Promise.all([
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false })
        ]).then(() => {
            return transactionsStore.getReconciliationStatements({
                accountId: options.accountId,
                startTime: options.startTime,
                endTime: options.endTime
            });
        }).then(result => {
            setReconciliationStatements(result);
            loading.value = false;
        }).catch(error => {
            loading.value = false;

            if (!error.processed) {
                emit('error', error);
                showState.value = false;
            }
        });

        return new Promise<void>((resolve, reject) => {
            rejectFunc = reject;
        });
    }

    function reload(force: boolean): void {
        loading.value = true;

        transactionsStore.getReconciliationStatements({
            accountId: accountId.value,
            startTime: startTime.value,
            endTime: endTime.value
        }).then(result => {
            if (force) {
                if (isReconciliationStatementsResponseEquals(result)) {
                    snackbar.value?.showMessage('Data is up to date');
                } else {
                    snackbar.value?.showMessage('Data has been updated');
                }
            }

            setReconciliationStatements(result);
            loading.value = false;
        }).catch(error => {
            loading.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        })
    }

    function addTransaction(): void {
        editDialog.value?.open({
            accountId: accountId.value
        }).then(result => {
            if (result && result.message) {
                snackbar.value?.showMessage(result.message);
            }

            reload(false);
        }).catch(error => {
            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }

    function updateClosingBalance(): void {
        let currentClosingBalance: BigDecimal = reconciliationStatements.value?.closingBalance ?? BIG_DECIMAL_ZERO;

        if (isCurrentLiabilityAccount.value) {
            currentClosingBalance = currentClosingBalance.negate();
        }

        if (!currentClosingBalance.between(TRANSACTION_MIN_AMOUNT, TRANSACTION_MAX_AMOUNT)) {
            snackbar.value?.showError('Numeric Overflow');
            return;
        }

        amountInputDialog.value?.open({
            title: 'Update Closing Balance',
            text: 'Please enter the new closing balance for this account',
            inputLabel: 'Closing Balance',
            inputPlaceholder: 'Closing Balance',
            currency: currentAccountCurrency.value,
            initAmount: currentClosingBalance.toSafeIntegerNumber()
        }).then(newClosingBalance => {
            if (!newClosingBalance) {
                return;
            }

            const currentUnixTime = getCurrentUnixTime();
            let newTransactionTime: number | undefined = undefined;

            if (endTime.value > 0 && endTime.value < currentUnixTime) {
                newTransactionTime = endTime.value;
            } else if (currentUnixTime < startTime.value) {
                newTransactionTime = startTime.value;
            }

            let newTransactionType: TransactionType = isCurrentLiabilityAccount.value ? TransactionType.Expense : TransactionType.Income;
            let newTransactionAmount: BigDecimal = parseBigDecimal(newClosingBalance).subtract(currentClosingBalance);

            if (newTransactionAmount.isNegative()) {
                newTransactionType = isCurrentLiabilityAccount.value ? TransactionType.Income : TransactionType.Expense;
                newTransactionAmount = newTransactionAmount.negate();
            }

            if (!newTransactionAmount.between(TRANSACTION_MIN_AMOUNT, TRANSACTION_MAX_AMOUNT)) {
                snackbar.value?.showError('Numeric Overflow');
                return;
            }

            editDialog.value?.open({
                time: newTransactionTime,
                type: newTransactionType,
                amount: newTransactionAmount.toSafeIntegerNumber(),
                accountId: accountId.value,
                noTransactionDraft: true
            }).then(result => {
                if (result && result.message) {
                    snackbar.value?.showMessage(result.message);
                }

                reload(false);
            }).catch(error => {
                if (error) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    function exportReconciliationStatements(fileType: KnownFileType): void {
        if (!reconciliationStatements.value || !reconciliationStatements.value.transactions || reconciliationStatements.value.transactions.length < 1) {
            return;
        }

        const exportedData = getExportedData(fileType);
        startDownloadFile(fileType.formatFileName(exportFileName.value), fileType.createBlob(exportedData));
    }

    function showTransaction(transaction: TransactionReconciliationStatementResponseItem): void {
        if (transaction.type === TransactionType.ModifyBalance) {
            return;
        }

        editDialog.value?.open({
            id: transaction.id,
            currentTransaction: Transaction.of(transaction)
        }).then(result => {
            if (result && result.message) {
                snackbar.value?.showMessage(result.message);
            }

            reload(false);
        }).catch(error => {
            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }

    function updateLastReconciledTime(): void {
        if (!newLastReconciledTime.value) {
            return;
        }

        const reconciledTime: number = newLastReconciledTime.value;
        updatingLastReconciledTime.value = true;

        accountsStore.updateAccountLastReconciledTime(accountId.value, reconciledTime).then(() => {
            updatingLastReconciledTime.value = false;
            snackbar.value?.showMessage('Last reconciled time have been updated');
            emit('update:last-reconciled-time', reconciledTime);
        }).catch(error => {
            updatingLastReconciledTime.value = false;

            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }

    function close(): void {
        rejectFunc?.();
        showState.value = false;
    }

    watch(showAccountBalanceTrendsCharts, (showCharts) => {
        if (!showCharts) {
            return;
        }

        const dialogCard = dialogLayout.value?.$el;

        if (dialogCard instanceof HTMLElement) {
            transactionListDialogHeight.value = dialogCard.getBoundingClientRect().height;
        }
    }, { flush: 'sync' });

    return {
        tt, formatRange, formatNumberToLocalizedNumerals, startTime, endTime, reconciliationStatements,
        chartDataDateAggregationType, timezoneUsedForDateRange, fiscalYearStart, defaultCurrency, allChartTypes,
        allDateAggregationTypes, allTimezoneTypesUsedForDateRange, currentAccount, currentAccountCurrency,
        currentAccountStatementDate, isCurrentLiabilityAccount, canUpdateAccountCloseBalance, newLastReconciledTime,
        displayStartDateTime, displayEndDateTime, displayTotalInflows, displayTotalInflowsInDefaultCurrency, displayTotalOutflows,
        displayTotalOutflowsInDefaultCurrency, displayTotalBalance, displayTotalBalanceInDefaultCurrency, displayOpeningBalance,
        displayOpeningBalanceInDefaultCurrency, displayClosingBalance, displayClosingBalanceInDefaultCurrency,
        getDisplayTransactionType, getDisplayDateTime, isSameAsDefaultTimezoneOffsetMinutes, getDisplayTimezone,
        getDisplayTimeInDefaultTimezone, getDisplaySourceAmount, getDisplayDestinationAmount, getDisplayAccountBalance,
        chartTypeIconMap, chartDataDateAggregationTypeIconMap, timezoneTypeIconMap, dialogLayout, amountInputDialog, snackbar,
        editDialog, showState, loading, updatingLastReconciledTime, currentPage, countPerPage, showAccountBalanceTrendsCharts,
        chartType, transactionListDialogHeight, useCommaDecimalSeparator, reconciliationStatementsTablePageOptions,
        preserveDialogHeight, totalPageCount, dataTableHeaders, getTransactionTypeColor, open, reload, addTransaction,
        updateClosingBalance, exportReconciliationStatements, showTransaction, updateLastReconciledTime, close,
    };
}
