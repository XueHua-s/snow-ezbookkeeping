import {
    type TimeRangeAndDateType,
    type Year0BasedMonth,
    DateRange,
    DateRangeScene
} from '@/core/datetime.ts';
import {
    getCurrentUnixTime,
    getDateRangeByBillingCycleDateType,
    getDateRangeByDateType,
    getDateRangeByLastReconciledTimeRangeDateType,
    getDateTypeByBillingCycleDateRange,
    getDateTypeByDateRange,
    getDayFirstDateTimeBySpecifiedUnixTime,
    getFullMonthDateRange,
    getShiftedDateRangeAndDateType,
    getShiftedDateRangeAndDateTypeForBillingCycle,
    getValidMonthDayOrCurrentDayShortDate,
    getYearMonthFirstUnixTime,
    getYearMonthLastUnixTime
} from '@/lib/datetime.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import type { useListPageLoading } from './loading.ts';
import type { useListPageState } from './state.ts';

interface Context {
    pageType: ReturnType<typeof useListPageState>['pageType'];
    currentCalendarDate: ReturnType<typeof useListPageState>['currentCalendarDate'];
    query: ReturnType<typeof useListPageState>['query'];
    firstDayOfWeek: ReturnType<typeof useListPageState>['firstDayOfWeek'];
    fiscalYearStart: ReturnType<typeof useListPageState>['fiscalYearStart'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    reload: ReturnType<typeof useListPageLoading>['reload'];
    customMaxDatetime: ReturnType<typeof useListPageState>['customMaxDatetime'];
    customMinDatetime: ReturnType<typeof useListPageState>['customMinDatetime'];
    showCustomMonthSheet: ReturnType<typeof useListPageState>['showCustomMonthSheet'];
    showCustomDateRangeSheet: ReturnType<typeof useListPageState>['showCustomDateRangeSheet'];
    accountsStore: ReturnType<typeof useListPageState>['accountsStore'];
    allAccountsMap: ReturnType<typeof useListPageState>['allAccountsMap'];
}

export function useListPageDateFilter(context: Context) {
    const {
        pageType, currentCalendarDate, query, firstDayOfWeek, fiscalYearStart, transactionsStore, reload, customMaxDatetime,
        customMinDatetime, showCustomMonthSheet, showCustomDateRangeSheet, accountsStore, allAccountsMap,
    } = context;

    function changePageType(type: number): void {
        pageType.value = type;
        currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(query.value.minTime, currentCalendarDate.value);

        if (pageType.value === TransactionListPageType.Calendar.type) {
            const dateRange = getFullMonthDateRange(query.value.minTime, query.value.maxTime, firstDayOfWeek.value, fiscalYearStart.value);

            if (dateRange) {
                const changed = transactionsStore.updateTransactionListFilter({
                    dateType: dateRange.dateType,
                    maxTime: dateRange.maxTime,
                    minTime: dateRange.minTime
                });

                if (changed) {
                    currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(query.value.minTime, currentCalendarDate.value);
                    reload();
                }
            }
        } else {
            reload();
        }
    }

    function changeDateFilter(dateType: number): void {
        if (dateType === DateRange.Custom.type) { // Custom
            if (!query.value.minTime || !query.value.maxTime) {
                customMaxDatetime.value = getCurrentUnixTime();
                customMinDatetime.value = getDayFirstDateTimeBySpecifiedUnixTime(customMaxDatetime.value).getUnixTime();
            } else {
                customMaxDatetime.value = query.value.maxTime;
                customMinDatetime.value = query.value.minTime;
            }

            if (pageType.value === TransactionListPageType.Calendar.type) {
                showCustomMonthSheet.value = true;
            } else {
                showCustomDateRangeSheet.value = true;
            }

            return;
        } else if (query.value.dateType === dateType) {
            return;
        }

        let dateRange: TimeRangeAndDateType | null = null;

        if (DateRange.isBillingCycle(dateType)) {
            dateRange = getDateRangeByBillingCycleDateType(dateType, firstDayOfWeek.value, fiscalYearStart.value, accountsStore.getAccountStatementDate(query.value.accountIds));
        } else if (DateRange.isLastReconciledTimeRange(dateType)) {
            dateRange = getDateRangeByLastReconciledTimeRangeDateType(dateType, allAccountsMap.value[query.value.accountIds]?.lastReconciledTime);
        } else {
            dateRange = getDateRangeByDateType(dateType, firstDayOfWeek.value, fiscalYearStart.value);
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

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: dateRange.dateType,
            maxTime: dateRange.maxTime,
            minTime: dateRange.minTime
        });

        if (changed) {
            reload();
        }
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

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: dateType,
            maxTime: maxTime,
            minTime: minTime
        });

        showCustomDateRangeSheet.value = false;

        if (changed) {
            reload();
        }
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

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: dateType,
            maxTime: maxTime,
            minTime: minTime
        });

        showCustomMonthSheet.value = false;

        if (changed) {
            reload();
        }
    }

    function shiftDateRange(minTime: number, maxTime: number, scale: number): void {
        if (query.value.dateType === DateRange.All.type) {
            return;
        }

        let newDateRange: TimeRangeAndDateType | null = null;

        if (DateRange.isBillingCycle(query.value.dateType) || query.value.dateType === DateRange.Custom.type) {
            newDateRange = getShiftedDateRangeAndDateTypeForBillingCycle(minTime, maxTime, scale, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal, accountsStore.getAccountStatementDate(query.value.accountIds));
        }

        if (!newDateRange) {
            newDateRange = getShiftedDateRangeAndDateType(minTime, maxTime, scale, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal);
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

        if (changed) {
            reload();
        }
    }

    return { changePageType, changeDateFilter, changeCustomDateFilter, changeCustomMonthDateFilter, shiftDateRange };
}
