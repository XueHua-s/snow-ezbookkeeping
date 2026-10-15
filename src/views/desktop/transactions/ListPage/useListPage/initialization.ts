import {
    type TimeRangeAndDateType,
    DateRange
} from '@/core/datetime.ts';
import { TemplateType } from '@/core/template.ts';
import {
    getDateRangeByDateType,
    getFullMonthDateRange,
    getValidMonthDayOrCurrentDayShortDate
} from '@/lib/datetime.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import type { useListPageLoading } from './loading.ts';
import type { useListPageState } from './state.ts';
import { type TransactionListProps } from './types.ts';

interface Context {
    firstDayOfWeek: ReturnType<typeof useListPageState>['firstDayOfWeek'];
    fiscalYearStart: ReturnType<typeof useListPageState>['fiscalYearStart'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    pageType: ReturnType<typeof useListPageState>['pageType'];
    currentCalendarDate: ReturnType<typeof useListPageState>['currentCalendarDate'];
    query: ReturnType<typeof useListPageState>['query'];
    updateUrlWhenChanged: ReturnType<typeof useListPageState>['updateUrlWhenChanged'];
    searchKeyword: ReturnType<typeof useListPageState>['searchKeyword'];
    currentAmountFilterType: ReturnType<typeof useListPageState>['currentAmountFilterType'];
    currentPage: ReturnType<typeof useListPageState>['currentPage'];
    reload: ReturnType<typeof useListPageLoading>['reload'];
    transactionTemplatesStore: ReturnType<typeof useListPageState>['transactionTemplatesStore'];
}

export function useListPageInitialization(context: Context) {
    const {
        firstDayOfWeek, fiscalYearStart, transactionsStore, pageType, currentCalendarDate, query, updateUrlWhenChanged,
        searchKeyword, currentAmountFilterType, currentPage, reload, transactionTemplatesStore,
    } = context;

    function init(initProps: TransactionListProps): void {
        let dateRange: TimeRangeAndDateType | null = getDateRangeByDateType(initProps.initDateType ? parseInt(initProps.initDateType) : undefined, firstDayOfWeek.value, fiscalYearStart.value);

        if (!dateRange && initProps.initDateType && initProps.initMaxTime && initProps.initMinTime &&
            (DateRange.isBillingCycle(parseInt(initProps.initDateType)) || DateRange.isLastReconciledTimeRange(parseInt(initProps.initDateType)) || initProps.initDateType === DateRange.Custom.type.toString()) &&
            parseInt(initProps.initMaxTime) > 0 && parseInt(initProps.initMinTime) > 0) {
            dateRange = {
                dateType: parseInt(initProps.initDateType),
                maxTime: parseInt(initProps.initMaxTime),
                minTime: parseInt(initProps.initMinTime)
            };
        }

        transactionsStore.initTransactionListFilter({
            dateType: dateRange ? dateRange.dateType : undefined,
            maxTime: dateRange ? dateRange.maxTime : undefined,
            minTime: dateRange ? dateRange.minTime : undefined,
            type: initProps.initType && parseInt(initProps.initType) > 0 ? parseInt(initProps.initType) : undefined,
            categoryIds: initProps.initCategoryIds,
            accountIds: initProps.initAccountIds,
            tagFilter: initProps.initTagFilter,
            amountFilter: initProps.initAmountFilter || '',
            keyword: initProps.initKeyword || '',
            matchMode: initProps.initMatchMode && parseInt(initProps.initMatchMode) >= 0 ? parseInt(initProps.initMatchMode) : undefined
        });

        if (initProps.initPageType) {
            const type = TransactionListPageType.valueOf(parseInt(initProps.initPageType));

            if (type) {
                pageType.value = type.type;
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
                            updateUrlWhenChanged(changed);
                            return;
                        }
                    }
                }
            }
        }

        searchKeyword.value = initProps.initKeyword || '';
        currentAmountFilterType.value = '';

        currentPage.value = 1;
        reload(false, true);

        transactionTemplatesStore.loadAllTemplates({
            templateType: TemplateType.Normal.type,
            force: false
        });
    }

    return { init };
}
