import {
    type LocalizedRecentMonthDateRange,
    type TimeRangeAndDateType,
    type Year0BasedMonth,
    DateRange,
    DateRangeScene
} from '@/core/datetime.ts';
import {
    isNumber,
    isObject
} from '@/lib/common.ts';
import {
    getCurrentUnixTime,
    getDateRangeByBillingCycleDateType,
    getDateRangeByDateType,
    getDateRangeByLastReconciledTimeRangeDateType,
    getDateTypeByBillingCycleDateRange,
    getDateTypeByDateRange,
    getDayFirstDateTimeBySpecifiedUnixTime,
    getFullMonthDateRange,
    getRecentDateRangeIndex,
    getShiftedDateRangeAndDateType,
    getShiftedDateRangeAndDateTypeForBillingCycle,
    getValidMonthDayOrCurrentDayShortDate,
    getYearMonthFirstUnixTime,
    getYearMonthLastUnixTime
} from '@/lib/datetime.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import { computed } from 'vue';
import type { useListPageState } from './state.ts';

interface Context {
    recentMonthDateRanges: ReturnType<typeof useListPageState>['recentMonthDateRanges'];
    query: ReturnType<typeof useListPageState>['query'];
    firstDayOfWeek: ReturnType<typeof useListPageState>['firstDayOfWeek'];
    fiscalYearStart: ReturnType<typeof useListPageState>['fiscalYearStart'];
    pageType: ReturnType<typeof useListPageState>['pageType'];
    currentCalendarDate: ReturnType<typeof useListPageState>['currentCalendarDate'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    updateUrlWhenChanged: ReturnType<typeof useListPageState>['updateUrlWhenChanged'];
    customMaxDatetime: ReturnType<typeof useListPageState>['customMaxDatetime'];
    customMinDatetime: ReturnType<typeof useListPageState>['customMinDatetime'];
    showCustomMonthDialog: ReturnType<typeof useListPageState>['showCustomMonthDialog'];
    showCustomDateRangeDialog: ReturnType<typeof useListPageState>['showCustomDateRangeDialog'];
    accountsStore: ReturnType<typeof useListPageState>['accountsStore'];
    allAccountsMap: ReturnType<typeof useListPageState>['allAccountsMap'];
}

export function useListPageDateFilter(context: Context) {
    const {
        recentMonthDateRanges, query, firstDayOfWeek, fiscalYearStart, pageType, currentCalendarDate, transactionsStore,
        updateUrlWhenChanged, customMaxDatetime, customMinDatetime, showCustomMonthDialog, showCustomDateRangeDialog,
        accountsStore, allAccountsMap,
    } = context;

    const recentDateRangeIndex = computed<number>({
        get: () => getRecentDateRangeIndex(recentMonthDateRanges.value, query.value.dateType, query.value.minTime, query.value.maxTime, firstDayOfWeek.value, fiscalYearStart.value),
        set: (value) => {
            if (value < 0 || value >= recentMonthDateRanges.value.length) {
                value = 0;
            }

            changeDateFilter(recentMonthDateRanges.value[value] as LocalizedRecentMonthDateRange);
        }
    });

    const queryPageType = computed<number>({
        get: () => pageType.value,
        set: (value) => changePageType(value)
    });

    function changePageType(type: number): void {
        pageType.value = type;
        currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(query.value.minTime, currentCalendarDate.value);

        if (pageType.value === TransactionListPageType.Calendar.type) {
            const dateRange = getFullMonthDateRange(query.value.minTime, query.value.maxTime, firstDayOfWeek.value, fiscalYearStart.value);

            if (dateRange) {
                transactionsStore.updateTransactionListFilter({
                    dateType: dateRange.dateType,
                    maxTime: dateRange.maxTime,
                    minTime: dateRange.minTime
                });
                currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(query.value.minTime, currentCalendarDate.value);
            }
        }

        updateUrlWhenChanged(true);
    }

    function changeDateFilter(dateRange: TimeRangeAndDateType | number | null): void {
        if (dateRange === DateRange.Custom.type || (isObject(dateRange) && dateRange.dateType === DateRange.Custom.type && !dateRange.minTime && !dateRange.maxTime)) { // Custom
            if (!query.value.minTime || !query.value.maxTime) {
                customMaxDatetime.value = getCurrentUnixTime();
                customMinDatetime.value = getDayFirstDateTimeBySpecifiedUnixTime(customMaxDatetime.value).getUnixTime();
            } else {
                customMaxDatetime.value = query.value.maxTime;
                customMinDatetime.value = query.value.minTime;
            }

            if (pageType.value === TransactionListPageType.Calendar.type) {
                showCustomMonthDialog.value = true;
            } else {
                showCustomDateRangeDialog.value = true;
            }

            return;
        }

        if (isNumber(dateRange)) {
            if (DateRange.isBillingCycle(dateRange)) {
                dateRange = getDateRangeByBillingCycleDateType(dateRange, firstDayOfWeek.value, fiscalYearStart.value, accountsStore.getAccountStatementDate(query.value.accountIds));
            } else if (DateRange.isLastReconciledTimeRange(dateRange)) {
                dateRange = getDateRangeByLastReconciledTimeRangeDateType(dateRange, allAccountsMap.value[query.value.accountIds]?.lastReconciledTime);
            } else {
                dateRange = getDateRangeByDateType(dateRange, firstDayOfWeek.value, fiscalYearStart.value);
            }
        }

        if (!dateRange) {
            return;
        }

        if (pageType.value === TransactionListPageType.Calendar.type) {
            currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(dateRange.minTime, currentCalendarDate.value);
            const fullMonthDateRange = getFullMonthDateRange(dateRange.minTime, dateRange.maxTime, firstDayOfWeek.value, fiscalYearStart.value);

            if (fullMonthDateRange) {
                dateRange = fullMonthDateRange;
                currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(dateRange.minTime, currentCalendarDate.value);
            }
        }

        if (query.value.dateType === dateRange.dateType && query.value.maxTime === dateRange.maxTime && query.value.minTime === dateRange.minTime) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: dateRange.dateType,
            maxTime: dateRange.maxTime,
            minTime: dateRange.minTime
        });

        updateUrlWhenChanged(changed);
    }

    function changeCustomDateFilter(minTime: number, maxTime: number): void {
        if (!minTime || !maxTime) {
            return;
        }

        let dateType: number | null = getDateTypeByBillingCycleDateRange(minTime, maxTime, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal, accountsStore.getAccountStatementDate(query.value.accountIds));

        if (!dateType) {
            dateType = getDateTypeByDateRange(minTime, maxTime, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal);
        }

        if (pageType.value === TransactionListPageType.Calendar.type) {
            currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(minTime, currentCalendarDate.value);
            const dateRange = getFullMonthDateRange(minTime, maxTime, firstDayOfWeek.value, fiscalYearStart.value);

            if (dateRange) {
                minTime = dateRange.minTime;
                maxTime = dateRange.maxTime;
                dateType = dateRange.dateType;
                currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(minTime, currentCalendarDate.value);
            }
        }

        if (query.value.dateType === dateType && query.value.maxTime === maxTime && query.value.minTime === minTime) {
            showCustomDateRangeDialog.value = false;
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: dateType,
            maxTime: maxTime,
            minTime: minTime
        });

        showCustomDateRangeDialog.value = false;
        updateUrlWhenChanged(changed);
    }

    function changeCustomMonthDateFilter(yearMonth: Year0BasedMonth): void {
        if (!yearMonth) {
            return;
        }

        const minTime = getYearMonthFirstUnixTime(yearMonth);
        const maxTime = getYearMonthLastUnixTime(yearMonth);
        const dateType = getDateTypeByDateRange(minTime, maxTime, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal);

        if (pageType.value === TransactionListPageType.Calendar.type) {
            currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(minTime, currentCalendarDate.value);
        }

        if (query.value.dateType === dateType && query.value.maxTime === maxTime && query.value.minTime === minTime) {
            showCustomMonthDialog.value = false;
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: dateType,
            maxTime: maxTime,
            minTime: minTime
        });

        showCustomMonthDialog.value = false;
        updateUrlWhenChanged(changed);
    }

    function shiftDateRange(startTime: number, endTime: number, scale: number): void {
        if (recentMonthDateRanges.value[recentDateRangeIndex.value]?.dateType === DateRange.All.type) {
            return;
        }

        let newDateRange: TimeRangeAndDateType | null = null;

        if (DateRange.isBillingCycle(query.value.dateType) || query.value.dateType === DateRange.Custom.type) {
            newDateRange = getShiftedDateRangeAndDateTypeForBillingCycle(startTime, endTime, scale, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal, accountsStore.getAccountStatementDate(query.value.accountIds));
        }

        if (!newDateRange) {
            newDateRange = getShiftedDateRangeAndDateType(startTime, endTime, scale, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal);
        }

        if (pageType.value === TransactionListPageType.Calendar.type) {
            currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(newDateRange.minTime, currentCalendarDate.value);
            const fullMonthDateRange = getFullMonthDateRange(newDateRange.minTime, newDateRange.maxTime, firstDayOfWeek.value, fiscalYearStart.value);

            if (fullMonthDateRange) {
                newDateRange = fullMonthDateRange;
                currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(newDateRange.minTime, currentCalendarDate.value);
            }
        }

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: newDateRange.dateType,
            maxTime: newDateRange.maxTime,
            minTime: newDateRange.minTime
        });

        updateUrlWhenChanged(changed);
    }

    return { recentDateRangeIndex, queryPageType, changeDateFilter, changeCustomDateFilter, changeCustomMonthDateFilter, shiftDateRange };
}
