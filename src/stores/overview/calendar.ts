import { KeywordMatchMode } from '@/core/text.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import { getFinalAccountIdsByFilteredAccountIds } from '@/lib/account.ts';
import { getFinalCategoryIdsByFilteredCategoryIds } from '@/lib/category.ts';
import {
    isEquals,
    isObjectEmpty,
    objectFieldWithValueToArrayItem
} from '@/lib/common.ts';
import {
    getTodayFirstUnixTime,
    getUnixTimeAfterUnixTime,
    getUnixTimeBeforeUnixTime,
    parseDateTimeFromUnixTime
} from '@/lib/datetime.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type {
TransactionDailyAmountsResponseItem
} from '@/models/transaction.ts';
import type { TransactionDailyTotalAmounts } from '@/stores/transaction.ts';
import type { useOverviewState } from './state.ts';

interface Context {
    transactionDataRange: ReturnType<typeof useOverviewState>['transactionDataRange'];
    updateTransactionDateRange: ReturnType<typeof useOverviewState>['updateTransactionDateRange'];
    transactionDailyAmountsData: ReturnType<typeof useOverviewState>['transactionDailyAmountsData'];
    transactionDailyAmountsStateInvalid: ReturnType<typeof useOverviewState>['transactionDailyAmountsStateInvalid'];
    settingsStore: ReturnType<typeof useOverviewState>['settingsStore'];
    currentMonthTransactions: ReturnType<typeof useOverviewState>['currentMonthTransactions'];
    currentMonthTransactionInflowOutflowDailyTotalAmounts: ReturnType<typeof useOverviewState>['currentMonthTransactionInflowOutflowDailyTotalAmounts'];
    currentMonthTransactionIncomeExpenseDailyTotalAmounts: ReturnType<typeof useOverviewState>['currentMonthTransactionIncomeExpenseDailyTotalAmounts'];
    currentMonthTransactionsStateInvalid: ReturnType<typeof useOverviewState>['currentMonthTransactionsStateInvalid'];
    accountsStore: ReturnType<typeof useOverviewState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useOverviewState>['transactionCategoriesStore'];
    transactionsStore: ReturnType<typeof useOverviewState>['transactionsStore'];
}

// Owns calendar for the overview store.
export function useOverviewCalendar(context: Context) {
    const {
        transactionDataRange, updateTransactionDateRange, transactionDailyAmountsData, transactionDailyAmountsStateInvalid,
        settingsStore, currentMonthTransactions, currentMonthTransactionInflowOutflowDailyTotalAmounts,
        currentMonthTransactionIncomeExpenseDailyTotalAmounts, currentMonthTransactionsStateInvalid, accountsStore,
        transactionCategoriesStore, transactionsStore,
    } = context;

    function loadTransactionDailyAmounts({ force, months }: { force: boolean, months: number }): Promise<TransactionDailyAmountsResponseItem[]> {
        if (transactionDataRange.value.today.startTime !== getTodayFirstUnixTime()) {
            updateTransactionDateRange();
            transactionDailyAmountsData.value = [];
            transactionDailyAmountsStateInvalid.value = true;
        }

        if (!force && !transactionDailyAmountsStateInvalid.value) {
            return Promise.resolve(transactionDailyAmountsData.value);
        }

        const excludeAccountIds = objectFieldWithValueToArrayItem(settingsStore.appSettings.overviewAccountFilterInHomePage, true);
        const excludeCategoryIds = objectFieldWithValueToArrayItem(settingsStore.appSettings.overviewTransactionCategoryFilterInHomePage, true);
        const endTime = transactionDataRange.value.today.endTime;
        const startTime = getUnixTimeAfterUnixTime(getUnixTimeBeforeUnixTime(transactionDataRange.value.today.startTime, Math.max(1, months), 'months'), 1, 'days');

        return new Promise((resolve, reject) => {
            services.getTransactionDailyAmounts({
                startTime: startTime,
                endTime: endTime,
                useTransactionTimezone: settingsStore.appSettings.timezoneUsedForStatisticsInHomePage === TimezoneTypeForStatistics.TransactionTimezone.type,
                excludeAccountIds: excludeAccountIds,
                excludeCategoryIds: excludeCategoryIds,
                tagFilter: settingsStore.appSettings.overviewTransactionTagFilterInHomePage
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve transaction overview' });
                    return;
                }

                if (transactionDailyAmountsStateInvalid.value) {
                    transactionDailyAmountsStateInvalid.value = false;
                }

                if (force && data.result && isEquals(transactionDailyAmountsData.value, data.result)) {
                    reject({ message: 'Data is up to date', isUpToDate: true });
                    return;
                }

                transactionDailyAmountsData.value = data.result;

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to retrieve transaction overview', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve transaction overview' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function loadCurrentMonthTransactions({ force }: { force: boolean }): Promise<TransactionDailyTotalAmounts> {
        if (transactionDataRange.value.today.startTime !== getTodayFirstUnixTime()) {
            updateTransactionDateRange();
            currentMonthTransactions.value = [];
            currentMonthTransactionInflowOutflowDailyTotalAmounts.value = {};
            currentMonthTransactionIncomeExpenseDailyTotalAmounts.value = {};
            currentMonthTransactionsStateInvalid.value = true;
        }

        if (!force && !currentMonthTransactionsStateInvalid.value) {
            return Promise.resolve({
                inflowOutflowDailyTotalAmounts: currentMonthTransactionInflowOutflowDailyTotalAmounts.value,
                incomeExpenseDailyTotalAmounts: currentMonthTransactionIncomeExpenseDailyTotalAmounts.value
            });
        }

        return new Promise((resolve, reject) => {
            Promise.all([
                accountsStore.loadAllAccounts({ force: false }),
                transactionCategoriesStore.loadAllCategories({ force: false })
            ]).then(() => {
                const currentMonth = parseDateTimeFromUnixTime(transactionDataRange.value.thisMonth.startTime);
                let accountIds = '';
                let categoryIds = '';

                if (!isObjectEmpty(settingsStore.appSettings.overviewAccountFilterInHomePage)) {
                    accountIds = getFinalAccountIdsByFilteredAccountIds(accountsStore.allAccountsMap, settingsStore.appSettings.overviewAccountFilterInHomePage);
                }

                if (!isObjectEmpty(settingsStore.appSettings.overviewTransactionCategoryFilterInHomePage)) {
                    categoryIds = getFinalCategoryIdsByFilteredCategoryIds(transactionCategoriesStore.allTransactionCategoriesMap, settingsStore.appSettings.overviewTransactionCategoryFilterInHomePage);
                }

                services.getAllTransactionsByMonth({
                    year: currentMonth.getGregorianCalendarYear(),
                    month: currentMonth.getGregorianCalendarMonth(),
                    type: 0,
                    categoryIds: categoryIds,
                    accountIds: accountIds,
                    tagFilter: settingsStore.appSettings.overviewTransactionTagFilterInHomePage,
                    amountFilter: '',
                    keyword: '',
                    matchMode: KeywordMatchMode.Default.type,
                    mustHavePictures: false,
                    withPictures: false
                }).then(response => {
                    const data = response.data;

                    if (!data || !data.success || !data.result) {
                        reject({ message: 'Unable to retrieve transaction list' });
                        return;
                    }

                    if (currentMonthTransactionsStateInvalid.value) {
                        currentMonthTransactionsStateInvalid.value = false;
                    }

                    if (force && data.result && isEquals(currentMonthTransactions.value, data.result.items)) {
                        reject({ message: 'Data is up to date', isUpToDate: true });
                        return;
                    }

                    currentMonthTransactions.value = data.result.items;

                    const dailyTotalAmounts = transactionsStore.getCurrentMonthTransactionDailyTotalAmounts(currentMonthTransactions.value, accountIds);
                    currentMonthTransactionInflowOutflowDailyTotalAmounts.value = dailyTotalAmounts.inflowOutflowDailyTotalAmounts;
                    currentMonthTransactionIncomeExpenseDailyTotalAmounts.value = dailyTotalAmounts.incomeExpenseDailyTotalAmounts;

                    resolve(dailyTotalAmounts);
                }).catch(error => {
                    logger.error('failed to retrieve transaction list', error);

                    if (error.response && error.response.data && error.response.data.errorMessage) {
                        reject({ error: error.response.data });
                    } else if (!error.processed) {
                        reject({ message: 'Unable to retrieve transaction list' });
                    } else {
                        reject(error);
                    }
                });
            }).catch(error => {
                logger.error('failed to load accounts or categories', error);
                reject(error);
            });
        });
    }

    return { loadTransactionDailyAmounts, loadCurrentMonthTransactions };
}
