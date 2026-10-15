import { entries, itemAndIndex } from '@/core/base.ts';
import type { OverviewRecentTransactionsQuery } from '@/core/overview_layout.ts';
import { KeywordMatchMode } from '@/core/text.ts';
import {
    getObjectOwnFieldCount,
    objectValueToArrayItem
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services,{ type ApiResponsePromise } from '@/lib/services.ts';
import type {
TransactionInfoPageWrapperResponse
} from '@/models/transaction.ts';
import type { TransactionInfoResponse } from '@/models/transaction.ts';
import type { useOverviewState } from './state.ts';

interface Context {
    getRecentTransactionsQueryCacheKey: ReturnType<typeof useOverviewState>['getRecentTransactionsQueryCacheKey'];
    recentTransactions: ReturnType<typeof useOverviewState>['recentTransactions'];
    recentTransactionsStateInvalid: ReturnType<typeof useOverviewState>['recentTransactionsStateInvalid'];
}

// Owns recent transactions for the overview store.
export function useOverviewRecentTransactions(context: Context) {
    const { getRecentTransactionsQueryCacheKey, recentTransactions, recentTransactionsStateInvalid } = context;

    function loadRecentTransactions({ force, queries }: { force: boolean, queries: Record<string, OverviewRecentTransactionsQuery> }): Promise<Record<string, TransactionInfoResponse[]>> {
        const widetIdCacheKeyMap: Record<string, string> = {};
        const cacheKeyQueryMap: Record<string, OverviewRecentTransactionsQuery> = {};

        for (const [widgetId, query] of entries(queries)) {
            const cacheKey = getRecentTransactionsQueryCacheKey(query);
            widetIdCacheKeyMap[widgetId] = cacheKey;

            const existsQuery = cacheKeyQueryMap[cacheKey];

            if (!existsQuery || existsQuery.count < query.count) {
                cacheKeyQueryMap[cacheKey] = query;
            }
        }

        const result: Record<string, TransactionInfoResponse[]> = {};
        let missedCacheKeyQueryMap: Record<string, OverviewRecentTransactionsQuery> = {};

        if (!force) {
            for (const [widgetId, query] of entries(queries)) {
                const cacheKey = widetIdCacheKeyMap[widgetId];

                if (!cacheKey) {
                    continue;
                }

                const cachedTransactions = recentTransactions.value[cacheKey];

                if (cachedTransactions && cachedTransactions.length >= query.count && !recentTransactionsStateInvalid.value[cacheKey]) {
                    result[widgetId] = cachedTransactions;
                } else {
                    missedCacheKeyQueryMap[cacheKey] = cacheKeyQueryMap[cacheKey] ?? query;
                }
            }
        } else {
            missedCacheKeyQueryMap = cacheKeyQueryMap;
        }

        if (getObjectOwnFieldCount(missedCacheKeyQueryMap) < 1) {
            return Promise.resolve(result);
        }

        const promises: ApiResponsePromise<TransactionInfoPageWrapperResponse>[] = [];
        const missedQueries: OverviewRecentTransactionsQuery[] = objectValueToArrayItem(missedCacheKeyQueryMap);

        return new Promise((resolve, reject) => {
            for (const query of missedQueries) {
                promises.push(services.getTransactions({
                    maxTime: 0,
                    minTime: 0,
                    count: query.count,
                    page: 1,
                    withCount: false,
                    withPictures: false,
                    mustHavePictures: false,
                    type: 0,
                    categoryIds: query.categoryIds.join(','),
                    accountIds: query.accountIds.join(','),
                    tagFilter: query.tagFilter,
                    amountFilter: query.amountFilter,
                    keyword: query.keyword,
                    matchMode: KeywordMatchMode.Default.type
                }));
            }

            Promise.all(promises).then(responses => {
                let hasError = false;

                for (const [response, index] of itemAndIndex(responses)) {
                    const query = missedQueries[index];

                    if (!query) {
                        continue;
                    }

                    const cacheKey = getRecentTransactionsQueryCacheKey(query);
                    const data = response.data;

                    if (!data || !data.success || !data.result) {
                        hasError = true;
                        continue;
                    }

                    if (recentTransactionsStateInvalid.value[cacheKey]) {
                        recentTransactionsStateInvalid.value[cacheKey] = false;
                    }

                    recentTransactions.value[cacheKey] = data.result.items;
                }

                if (hasError) {
                    reject({ message: 'Unable to retrieve transaction list' });
                    return;
                }

                for (const [widgetId, query] of entries(queries)) {
                    const cacheKey = widetIdCacheKeyMap[widgetId];

                    if (!cacheKey) {
                        continue;
                    }

                    const cachedTransactions = recentTransactions.value[cacheKey];

                    if (cachedTransactions && cachedTransactions.length >= query.count) {
                        result[widgetId] = cachedTransactions;
                    }
                }

                resolve(result);
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
        });
    }

    return { loadRecentTransactions };
}
