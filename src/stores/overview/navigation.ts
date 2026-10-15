import { DateRange } from '@/core/datetime.ts';
import type { OverviewRecentTransactionsQuery } from '@/core/overview_layout.ts';
import type { TransactionType } from '@/core/transaction.ts';
import { getFinalAccountIdsByFilteredAccountIds } from '@/lib/account.ts';
import { getFinalCategoryIdsByFilteredCategoryIds } from '@/lib/category.ts';
import {
    isDefined,
    isNumber,
    isObjectEmpty
} from '@/lib/common.ts';
import type { TransactionInfoResponse } from '@/models/transaction.ts';
import type { useOverviewState } from './state.ts';

interface Context {
    getRecentTransactionsQueryCacheKey: ReturnType<typeof useOverviewState>['getRecentTransactionsQueryCacheKey'];
    recentTransactions: ReturnType<typeof useOverviewState>['recentTransactions'];
    settingsStore: ReturnType<typeof useOverviewState>['settingsStore'];
    transactionCategoriesStore: ReturnType<typeof useOverviewState>['transactionCategoriesStore'];
    accountsStore: ReturnType<typeof useOverviewState>['accountsStore'];
}

// Owns navigation for the overview store.
export function useOverviewNavigation(context: Context) {
    const { getRecentTransactionsQueryCacheKey, recentTransactions, settingsStore, transactionCategoriesStore, accountsStore } = context;

    function getRecentTransactions(query: OverviewRecentTransactionsQuery): TransactionInfoResponse[] {
        const cacheKey = getRecentTransactionsQueryCacheKey(query);
        return recentTransactions.value[cacheKey] ?? [];
    }

    function getTransactionListPageParams({ type, dateType, minTime, maxTime, categoryIds }: { type?: TransactionType, dateType?: number, minTime?: number, maxTime?: number, categoryIds?: string[] }): string {
        const querys: string[] = [];

        if (isDefined(type)) {
            querys.push('type=' + type);
        }

        if (isDefined(dateType)) {
            querys.push('dateType=' + dateType);

            if (dateType === DateRange.Custom.type) {
                if (isNumber(minTime) && minTime > 0) {
                    querys.push('minTime=' + minTime);
                }

                if (isNumber(maxTime) && maxTime > 0) {
                    querys.push('maxTime=' + maxTime);
                }
            }
        }

        if (categoryIds && categoryIds.length > 0) {
            querys.push('categoryIds=' + categoryIds.join(','));
        } else if (!isObjectEmpty(settingsStore.appSettings.overviewTransactionCategoryFilterInHomePage)) {
            querys.push('categoryIds=' + getFinalCategoryIdsByFilteredCategoryIds(transactionCategoriesStore.allTransactionCategoriesMap, settingsStore.appSettings.overviewTransactionCategoryFilterInHomePage));
        }

        if (!isObjectEmpty(settingsStore.appSettings.overviewAccountFilterInHomePage)) {
            querys.push('accountIds=' + getFinalAccountIdsByFilteredAccountIds(accountsStore.allAccountsMap, settingsStore.appSettings.overviewAccountFilterInHomePage));
        }

        if (settingsStore.appSettings.overviewTransactionTagFilterInHomePage) {
            querys.push('tagFilter=' + encodeURIComponent(settingsStore.appSettings.overviewTransactionTagFilterInHomePage));
        }

        return querys.join('&');
    }

    return { getRecentTransactions, getTransactionListPageParams };
}
