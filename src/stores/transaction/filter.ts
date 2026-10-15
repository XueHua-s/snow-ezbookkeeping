import { DateRange } from '@/core/datetime.ts';
import {
    isNumber,
    isString
} from '@/lib/common.ts';
import type { ExportTransactionDataRequest } from '@/models/data_management.ts';
import type { useTransactionState } from './state.ts';
import type { TransactionListPartialFilter } from './types.ts';

interface Context {
    transactionsFilter: ReturnType<typeof useTransactionState>['transactionsFilter'];
    settingsStore: ReturnType<typeof useTransactionState>['settingsStore'];
    accountsStore: ReturnType<typeof useTransactionState>['accountsStore'];
}

// Owns filter for the transaction store.
export function useTransactionFilter(context: Context) {
    const { transactionsFilter, settingsStore, accountsStore } = context;

    function initTransactionListFilter(filter: TransactionListPartialFilter): void {
        if (filter && isNumber(filter.dateType)) {
            transactionsFilter.value.dateType = filter.dateType;
        } else {
            transactionsFilter.value.dateType = DateRange.All.type;
        }

        if (filter && isNumber(filter.maxTime)) {
            transactionsFilter.value.maxTime = filter.maxTime;
        } else {
            transactionsFilter.value.maxTime = 0;
        }

        if (filter && isNumber(filter.minTime)) {
            transactionsFilter.value.minTime = filter.minTime;
        } else {
            transactionsFilter.value.minTime = 0;
        }

        if (filter && isNumber(filter.type)) {
            transactionsFilter.value.type = filter.type;
        } else {
            transactionsFilter.value.type = 0;
        }

        if (filter && isString(filter.categoryIds)) {
            transactionsFilter.value.categoryIds = filter.categoryIds;
        } else {
            transactionsFilter.value.categoryIds = '';
        }

        if (filter && isString(filter.accountIds)) {
            transactionsFilter.value.accountIds = filter.accountIds;
        } else {
            transactionsFilter.value.accountIds = '';
        }

        if (filter && isString(filter.tagFilter)) {
            transactionsFilter.value.tagFilter = filter.tagFilter;
        } else {
            transactionsFilter.value.tagFilter = '';
        }

        if (filter && isString(filter.amountFilter)) {
            transactionsFilter.value.amountFilter = filter.amountFilter;
        } else {
            transactionsFilter.value.amountFilter = '';
        }

        if (filter && isString(filter.keyword)) {
            transactionsFilter.value.keyword = filter.keyword;
        } else {
            transactionsFilter.value.keyword = '';
        }

        if (filter && isNumber(filter.matchMode)) {
            transactionsFilter.value.matchMode = filter.matchMode;
        } else {
            transactionsFilter.value.matchMode = settingsStore.appSettings.defaultKeywordMatchModeInTransactionListPage;
        }
    }

    function updateTransactionListFilter(filter: TransactionListPartialFilter): boolean {
        let changed = false;

        if (filter && isNumber(filter.dateType) && transactionsFilter.value.dateType !== filter.dateType) {
            transactionsFilter.value.dateType = filter.dateType;
            changed = true;
        }

        if (filter && isNumber(filter.maxTime) && transactionsFilter.value.maxTime !== filter.maxTime) {
            transactionsFilter.value.maxTime = filter.maxTime;
            changed = true;
        }

        if (filter && isNumber(filter.minTime) && transactionsFilter.value.minTime !== filter.minTime) {
            transactionsFilter.value.minTime = filter.minTime;
            changed = true;
        }

        if (filter && isNumber(filter.type) && transactionsFilter.value.type !== filter.type) {
            transactionsFilter.value.type = filter.type;
            changed = true;
        }

        if (filter && isString(filter.categoryIds) && transactionsFilter.value.categoryIds !== filter.categoryIds) {
            transactionsFilter.value.categoryIds = filter.categoryIds;
            changed = true;
        }

        if (filter && isString(filter.accountIds) && transactionsFilter.value.accountIds !== filter.accountIds) {
            if (DateRange.isBillingCycle(transactionsFilter.value.dateType) &&
                (!accountsStore.getAccountStatementDate(filter.accountIds) || accountsStore.getAccountStatementDate(filter.accountIds) !== accountsStore.getAccountStatementDate(transactionsFilter.value.accountIds))) {
                transactionsFilter.value.dateType = DateRange.Custom.type;
            } else if (DateRange.isLastReconciledTimeRange(transactionsFilter.value.dateType) &&
                (!accountsStore.allAccountsMap[filter.accountIds] || accountsStore.allAccountsMap[filter.accountIds]?.lastReconciledTime !== accountsStore.allAccountsMap[transactionsFilter.value.accountIds]?.lastReconciledTime)) {
                transactionsFilter.value.dateType = DateRange.Custom.type;
            }

            transactionsFilter.value.accountIds = filter.accountIds;
            changed = true;
        }

        if (filter && isString(filter.tagFilter) && transactionsFilter.value.tagFilter !== filter.tagFilter) {
            transactionsFilter.value.tagFilter = filter.tagFilter;
            changed = true;
        }

        if (filter && isString(filter.amountFilter) && transactionsFilter.value.amountFilter !== filter.amountFilter) {
            transactionsFilter.value.amountFilter = filter.amountFilter;
            changed = true;
        }

        if (filter && isString(filter.keyword) && transactionsFilter.value.keyword !== filter.keyword) {
            transactionsFilter.value.keyword = filter.keyword;
            changed = true;
        }

        if (filter && isNumber(filter.matchMode) && transactionsFilter.value.matchMode !== filter.matchMode) {
            transactionsFilter.value.matchMode = filter.matchMode;
            changed = true;
        }

        return changed;
    }

    function getTransactionListPageParams(pageType: number): string {
        const querys: string[] = [];

        querys.push('pageType=' + pageType);

        if (transactionsFilter.value.type) {
            querys.push('type=' + transactionsFilter.value.type);
        }

        if (transactionsFilter.value.accountIds) {
            querys.push('accountIds=' + transactionsFilter.value.accountIds);
        }

        if (transactionsFilter.value.categoryIds) {
            querys.push('categoryIds=' + transactionsFilter.value.categoryIds);
        }

        if (transactionsFilter.value.tagFilter) {
            querys.push('tagFilter=' + transactionsFilter.value.tagFilter);
        }

        querys.push('dateType=' + transactionsFilter.value.dateType);

        if (DateRange.isBillingCycle(transactionsFilter.value.dateType)
            || DateRange.isLastReconciledTimeRange(transactionsFilter.value.dateType)
            || transactionsFilter.value.dateType === DateRange.Custom.type) {
            querys.push('maxTime=' + transactionsFilter.value.maxTime);
            querys.push('minTime=' + transactionsFilter.value.minTime);
        }

        if (transactionsFilter.value.amountFilter) {
            querys.push('amountFilter=' + encodeURIComponent(transactionsFilter.value.amountFilter));
        }

        if (transactionsFilter.value.keyword) {
            querys.push('keyword=' + encodeURIComponent(transactionsFilter.value.keyword));
            querys.push('matchMode=' + transactionsFilter.value.matchMode);
        }

        return querys.join('&');
    }

    function getExportTransactionDataRequestByTransactionFilter(): ExportTransactionDataRequest {
        return {
            maxTime: transactionsFilter.value.maxTime,
            minTime: transactionsFilter.value.minTime,
            type: transactionsFilter.value.type,
            categoryIds: transactionsFilter.value.categoryIds,
            accountIds: transactionsFilter.value.accountIds,
            tagFilter: transactionsFilter.value.tagFilter,
            amountFilter: transactionsFilter.value.amountFilter,
            keyword: transactionsFilter.value.keyword,
            matchMode: transactionsFilter.value.matchMode
        };
    }

    return { initTransactionListFilter, updateTransactionListFilter, getTransactionListPageParams, getExportTransactionDataRequestByTransactionFilter };
}
