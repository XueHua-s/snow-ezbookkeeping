import type { SetupProps } from '@/lib/ui/props.ts';
import { TRANSACTION_MAX_AMOUNT, TRANSACTION_MIN_AMOUNT } from '@/consts/transaction.ts';
import { AccountType } from '@/core/account.ts';
import { type TimeRangeAndDateType, DateRange, DateRangeScene } from '@/core/datetime.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_MOBILE } from '@/core/statistics.ts';
import { TextDirection } from '@/core/text.ts';
import { TransactionType } from '@/core/transaction.ts';
import { findDisplayNameByType, isDefined } from '@/lib/common.ts';
import {
    getCurrentUnixTime,
    getDateRangeByBillingCycleDateType,
    getDateRangeByDateType,
    getDateRangeByLastReconciledTimeRangeDateType,
    getDateTypeByBillingCycleDateRange,
    getDateTypeByDateRange
} from '@/lib/datetime.ts';
import { BIG_DECIMAL_ZERO, parseBigDecimal } from '@/lib/numeral.ts';
import { hideLoading, onSwipeoutDeleted, showLoading, useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useI18n } from '@/locales/helpers.ts';
import { type TransactionReconciliationStatementResponseItemWithInfo } from '@/models/transaction.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useUserStore } from '@/stores/user.ts';
import { useReconciliationStatementPageBase } from '@/views/base/accounts/ReconciliationStatementPageBase.ts';
import type { Router } from 'framework7/types';
import { computed, nextTick, ref } from 'vue';

export interface ReconciliationStatementVirtualListData {
    items: ReconciliationStatementVirtualListItem[],
    topPosition: number
}

export interface ReconciliationStatementVirtualListItem {
    index: number;
    type: ReconciliationStatementVirtualListItemType;
    displayDate?: string;
    transaction?: TransactionReconciliationStatementResponseItemWithInfo;
}

export type ReconciliationStatementVirtualListItemType = 'transaction' | 'date';

// Coordinates ReconciliationStatementPage state and effects; called once from the component setup.
export function useReconciliationStatementPage(props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>) {
    const {
        tt,
        getCurrentLanguageTextDirection,
        getAllDateRanges,
        formatNumberToLocalizedNumerals
    } = useI18n();

    const { showAlert, showToast, routeBackOnError } = useI18nUIComponents();

    const {
        accountId,
        startTime,
        endTime,
        reconciliationStatements,
        chartDataDateAggregationType,
        timezoneUsedForDateRange,
        firstDayOfWeek,
        fiscalYearStart,
        allDateAggregationTypes,
        allTimezoneTypesUsedForDateRange,
        isCurrentLiabilityAccount,
        canUpdateAccountCloseBalance,
        newLastReconciledTime,
        currentAccount,
        currentAccountCurrency,
        currentAccountStatementDate,
        currentAccountLastReconciledTime,
        displayStartDateTime,
        displayEndDateTime,
        displayTotalInflows,
        displayTotalOutflows,
        displayTotalBalance,
        displayOpeningBalance,
        displayClosingBalance,
        updatePageOpenTime,
        isReconciliationStatementsResponseEquals,
        setReconciliationStatements,
        getDisplayDate,
        getDisplayTime,
        isSameAsDefaultTimezoneOffsetMinutes,
        getDisplayTimezone,
        getDisplaySourceAmount,
        getDisplayDestinationAmount,
        getDisplayAccountBalance
    } = useReconciliationStatementPageBase();

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionsStore = useTransactionsStore();

    const finishQuery = ref<boolean>(false);

    const loading = ref<boolean>(false);

    const loadingError = ref<unknown | null>(null);

    const queryDateRangeType = ref<number>(DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_MOBILE.type);

    const showAccountBalanceTrendsCharts = ref<boolean>(false);

    const updatingLastReconciledTime = ref<boolean>(false);

    const transactionToDelete = ref<TransactionReconciliationStatementResponseItemWithInfo | null>(null);

    const newClosingBalance = ref<number>(0);

    const showCustomDateRangeSheet = ref<boolean>(false);

    const showNewClosingBalanceSheet = ref<boolean>(false);

    const showMoreActionSheet = ref<boolean>(false);

    const showDeleteActionSheet = ref<boolean>(false);

    const virtualDataItems = ref<ReconciliationStatementVirtualListData>({
        items: [],
        topPosition: 0
    });

    const textDirection = computed<TextDirection>(() => getCurrentLanguageTextDirection());

    const validQuery = computed(() => currentAccount.value && currentAccount.value.type === AccountType.SingleAccount.type);

    const allAvailableDateRanges = computed(() => getAllDateRanges(DateRangeScene.Normal, {
        includeCustom: true,
        includeBillingCycle: !!accountsStore.getAccountStatementDate(accountId.value),
        includeLastReconciledTimeRange: userStore.currentUserUseLastReconciledTime && !!currentAccountLastReconciledTime.value
    }));

    const allReconciliationStatementVirtualListItems = computed<ReconciliationStatementVirtualListItem[]>(() => {
        const ret: ReconciliationStatementVirtualListItem[] = [];

        if (!reconciliationStatements.value || !reconciliationStatements.value.transactions || reconciliationStatements.value.transactions.length < 1) {
            return ret;
        }

        let index = 0;
        let lastDisplayDate: string | null = null;

        for (const transaction of reconciliationStatements.value.transactions) {
            const displayDate = getDisplayDate(transaction);

            if (lastDisplayDate !== displayDate) {
                lastDisplayDate = displayDate;
                ret.push({
                    index: index++,
                    type: 'date',
                    displayDate: displayDate
                });
            }

            ret.push({
                index: index++,
                type: 'transaction',
                transaction: transaction
            });
        }

        return ret;
    });

    const chartDataDateAggregationTypeDisplayName = computed<string>(() => {
        return findDisplayNameByType(allDateAggregationTypes.value, chartDataDateAggregationType.value) || tt('Unknown');
    });

    function getTransactionDomId(transaction: TransactionReconciliationStatementResponseItemWithInfo): string {
        return 'transaction_' + transaction.id;
    }

    function init(): void {
        const query = props.f7route.query;
        const initialDateRangeType = query['dateType'] ? parseInt(query['dateType']) : undefined;
        const initialStartTime = query['startTime'] ? parseInt(query['startTime']) : undefined;
        const initialEndTime = query['endTime'] ? parseInt(query['endTime']) : undefined;
        const defaultDateRange = getDateRangeByDateType(queryDateRangeType.value, firstDayOfWeek.value, fiscalYearStart.value);

        if (isDefined(initialDateRangeType) && initialDateRangeType !== DateRange.Custom.type) {
            loading.value = true;
        } else if (isDefined(initialDateRangeType) && initialDateRangeType === DateRange.Custom.type && isDefined(initialStartTime) && isDefined(initialEndTime) && initialStartTime > 0 && initialEndTime >= initialStartTime) {
            loading.value = true;
        } else {
            loading.value = false;
        }

        updatePageOpenTime();
        finishQuery.value = false;

        accountId.value = query['accountId'] || '';
        startTime.value = defaultDateRange?.minTime || 0;
        endTime.value = defaultDateRange?.maxTime || 0;
        reconciliationStatements.value = undefined;

        initDateFilter();

        Promise.all([
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false })
        ]).then(() => {
            let autoReload: boolean = false;

            if (isDefined(initialDateRangeType) && initialDateRangeType !== DateRange.Custom.type) {
                changeDateFilter(initialDateRangeType);
                autoReload = true;
            } else if (isDefined(initialDateRangeType) && initialDateRangeType === DateRange.Custom.type && isDefined(initialStartTime) && isDefined(initialEndTime) && initialStartTime > 0 && initialEndTime >= initialStartTime) {
                changeCustomDateFilter(initialStartTime, initialEndTime);
                autoReload = true;
            }

            if (autoReload) {
                reload(false);
            }
        }).catch(error => {
            loading.value = false;
            loadingError.value = error;
            showToast(error.message || error);
        });
    }

    function initDateFilter(): void {
        let defaultDateRangeType = settingsStore.appSettings.reconciliationStatementPageDefaultDateRangeTypeInMobile;
        const defualtDateRange = DateRange.valueOf(defaultDateRangeType);

        if (!defualtDateRange) {
            defaultDateRangeType = DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_MOBILE.type;
        } else if (defualtDateRange.isBillingCycle && !accountsStore.getAccountStatementDate(accountId.value)) {
            defaultDateRangeType = DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_MOBILE.type;
        } else if (defualtDateRange.isLastReconciledTimeRange && (!userStore.currentUserUseLastReconciledTime || !currentAccount.value || !currentAccount.value.lastReconciledTime)) {
            defaultDateRangeType = DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_MOBILE.type;
        } else {
            defaultDateRangeType = defualtDateRange.type;
        }

        nextTick(() => {
            changeDateFilter(defaultDateRangeType);
        });
    }

    function changeDateFilter(dateRangeType: number): void {
        if (dateRangeType === DateRange.Custom.type) {
            showCustomDateRangeSheet.value = true;
            return;
        }

        let dateRange: TimeRangeAndDateType | null = null;

        if (DateRange.isBillingCycle(dateRangeType)) {
            dateRange = getDateRangeByBillingCycleDateType(dateRangeType, firstDayOfWeek.value, fiscalYearStart.value, accountsStore.getAccountStatementDate(accountId.value));
        } else if (DateRange.isLastReconciledTimeRange(dateRangeType)) {
            dateRange = getDateRangeByLastReconciledTimeRangeDateType(dateRangeType, currentAccountLastReconciledTime.value);
        } else {
            dateRange = getDateRangeByDateType(dateRangeType, firstDayOfWeek.value, fiscalYearStart.value);
        }

        if (!dateRange) {
            return;
        }

        queryDateRangeType.value = dateRange.dateType;
        startTime.value = dateRange.minTime;
        endTime.value = dateRange.maxTime;
    }

    function changeCustomDateFilter(minTime: number, maxTime: number): void {
        if (!minTime || !maxTime) {
            return;
        }

        let dateType: number | null = getDateTypeByBillingCycleDateRange(minTime, maxTime, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal, accountsStore.getAccountStatementDate(accountId.value));

        if (!dateType) {
            dateType = getDateTypeByDateRange(minTime, maxTime, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal);
        }

        queryDateRangeType.value = dateType;
        startTime.value = minTime;
        endTime.value = maxTime;
        showCustomDateRangeSheet.value = false;
    }

    function reload(force: boolean): void {
        finishQuery.value = true;
        loading.value = true;

        transactionsStore.getReconciliationStatements({
            accountId: accountId.value,
            startTime: startTime.value,
            endTime: endTime.value
        }).then(result => {
            if (force) {
                if (isReconciliationStatementsResponseEquals(result)) {
                    showToast('Data is up to date');
                } else {
                    showToast('Data has been updated');
                }
            }

            loading.value = false;
            setReconciliationStatements(result);
        }).catch(error => {
            loading.value = false;

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function addTransaction(): void {
        props.f7router.navigate(`/transaction/add?accountId=${accountId.value}`);
    }

    function duplicateTransaction(transaction: TransactionReconciliationStatementResponseItemWithInfo): void {
        props.f7router.navigate(`/transaction/add?id=${transaction.id}&type=${transaction.type}`);
    }

    function editTransaction(transaction: TransactionReconciliationStatementResponseItemWithInfo): void {
        props.f7router.navigate(`/transaction/edit?id=${transaction.id}&type=${transaction.type}`);
    }

    function updateClosingBalance(balance?: number): void {
        let currentClosingBalance: BigDecimal = reconciliationStatements.value?.closingBalance ?? BIG_DECIMAL_ZERO;

        if (isCurrentLiabilityAccount.value) {
            currentClosingBalance = currentClosingBalance.negate();
        }

        if (!isDefined(balance)) {
            if (!currentClosingBalance.between(TRANSACTION_MIN_AMOUNT, TRANSACTION_MAX_AMOUNT)) {
                showToast('Numeric Overflow');
                return;
            }

            newClosingBalance.value = currentClosingBalance.toSafeIntegerNumber();
            showNewClosingBalanceSheet.value = true;
            return;
        }

        const currentUnixTime = getCurrentUnixTime();
        let setTransactionTime = false;
        let newTransactionTime: number | undefined = undefined;

        if (endTime.value < currentUnixTime) {
            setTransactionTime = true;
            newTransactionTime = endTime.value;
        } else if (currentUnixTime < startTime.value) {
            setTransactionTime = true;
            newTransactionTime = startTime.value;
        }

        let newTransactionType: TransactionType = isCurrentLiabilityAccount.value ? TransactionType.Expense : TransactionType.Income;
        let newTransactionAmount: BigDecimal = parseBigDecimal(balance).subtract(currentClosingBalance);

        if (newTransactionAmount.isNegative()) {
            newTransactionType = isCurrentLiabilityAccount.value ? TransactionType.Income : TransactionType.Expense;
            newTransactionAmount = newTransactionAmount.negate();
        }

        if (!newTransactionAmount.between(TRANSACTION_MIN_AMOUNT, TRANSACTION_MAX_AMOUNT)) {
            showToast('Numeric Overflow');
            return;
        }

        const params: string[] = [];

        if (setTransactionTime) {
            params.push(`time=${newTransactionTime}`);
        }

        params.push(`type=${newTransactionType}`);
        params.push(`amount=${newTransactionAmount.toSafeIntegerNumber()}`);
        params.push(`accountId=${accountId.value}`);
        params.push(`noTransactionDraft=true`);

        props.f7router.navigate(`/transaction/add?${params.join('&')}`);
    }

    function updateLastReconciledTime(): void {
        if (!newLastReconciledTime.value) {
            return;
        }

        updatingLastReconciledTime.value = true;
        showLoading(() => updatingLastReconciledTime.value);

        accountsStore.updateAccountLastReconciledTime(accountId.value, newLastReconciledTime.value).then(() => {
            updatingLastReconciledTime.value = false;
            hideLoading();
            showToast('Last reconciled time have been updated');
        }).catch(error => {
            updatingLastReconciledTime.value = false;
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function removeTransaction(transaction: TransactionReconciliationStatementResponseItemWithInfo | null, confirm: boolean): void {
        if (!transaction) {
            showAlert('An error occurred');
            return;
        }

        if (!confirm) {
            transactionToDelete.value = transaction;
            showDeleteActionSheet.value = true;
            return;
        }

        showDeleteActionSheet.value = false;
        transactionToDelete.value = null;
        showLoading();

        transactionsStore.deleteTransaction({
            transaction: transaction,
            defaultCurrency: currentAccountCurrency.value,
            beforeResolve: (done) => {
                onSwipeoutDeleted(getTransactionDomId(transaction), done);
            }
        }).then(() => {
            hideLoading();
            reload(false);
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function setChartDataDateAggregationType(type: number): void {
        chartDataDateAggregationType.value = type;
    }

    function setTimezoneUsedForDateRange(type: number): void {
        timezoneUsedForDateRange.value = type;
    }

    function renderExternal(vl: unknown, vlData: ReconciliationStatementVirtualListData): void {
        virtualDataItems.value = vlData;
    }

    function onPageAfterIn(): void {
        if (finishQuery.value && transactionsStore.transactionReconciliationStatementStateInvalid) {
            reload(false);
        }

        routeBackOnError(props.f7router, loadingError);
    }

    init();

    return {
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
    };
}
