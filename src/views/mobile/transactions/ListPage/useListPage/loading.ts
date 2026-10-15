import {
    type TimeRangeAndDateType,
    DateRange
} from '@/core/datetime.ts';
import {
    getDateRangeByDateType,
    parseDateTimeFromUnixTime
} from '@/lib/datetime.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import type { Router } from 'framework7/types';
import type { useListPageState } from './state.ts';
import type { useListPageVirtualList } from './virtual-list.ts';

interface Context {
    props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>;
    firstDayOfWeek: ReturnType<typeof useListPageState>['firstDayOfWeek'];
    fiscalYearStart: ReturnType<typeof useListPageState>['fiscalYearStart'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    loading: ReturnType<typeof useListPageState>['loading'];
    pageType: ReturnType<typeof useListPageState>['pageType'];
    transactionInvisibleYearMonths: ReturnType<typeof useListPageState>['transactionInvisibleYearMonths'];
    transactionYearMonthListHeights: ReturnType<typeof useListPageState>['transactionYearMonthListHeights'];
    accountsStore: ReturnType<typeof useListPageState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useListPageState>['transactionCategoriesStore'];
    transactionTagsStore: ReturnType<typeof useListPageState>['transactionTagsStore'];
    queryMonthlyData: ReturnType<typeof useListPageState>['queryMonthlyData'];
    query: ReturnType<typeof useListPageState>['query'];
    selectedAccountDefaultCurrency: ReturnType<typeof useListPageState>['selectedAccountDefaultCurrency'];
    showToast: ReturnType<typeof useListPageState>['showToast'];
    setTransactionMonthListHeights: ReturnType<typeof useListPageVirtualList>['setTransactionMonthListHeights'];
    loadingError: ReturnType<typeof useListPageState>['loadingError'];
    hasMoreTransaction: ReturnType<typeof useListPageVirtualList>['hasMoreTransaction'];
    loadingMore: ReturnType<typeof useListPageState>['loadingMore'];
}

export function useListPageLoading(context: Context) {
    const {
        props, firstDayOfWeek, fiscalYearStart, transactionsStore, loading, pageType, transactionInvisibleYearMonths,
        transactionYearMonthListHeights, accountsStore, transactionCategoriesStore, transactionTagsStore, queryMonthlyData, query,
        selectedAccountDefaultCurrency, showToast, setTransactionMonthListHeights, loadingError, hasMoreTransaction, loadingMore,
    } = context;

    function init(): void {
        const initQuery = props.f7route.query;

        let dateRange: TimeRangeAndDateType | null = getDateRangeByDateType(initQuery['dateType'] ? parseInt(initQuery['dateType']) : undefined, firstDayOfWeek.value, fiscalYearStart.value);

        if (!dateRange && initQuery['dateType'] && initQuery['maxTime'] && initQuery['minTime'] &&
            (DateRange.isBillingCycle(parseInt(initQuery['dateType'])) || DateRange.isLastReconciledTimeRange(parseInt(initQuery['dateType'])) || initQuery['dateType'] === DateRange.Custom.type.toString()) &&
            parseInt(initQuery['maxTime']) > 0 && parseInt(initQuery['minTime']) > 0) {
            dateRange = {
                dateType: parseInt(initQuery['dateType']),
                maxTime: parseInt(initQuery['maxTime']),
                minTime: parseInt(initQuery['minTime'])
            };
        }

        transactionsStore.initTransactionListFilter({
            dateType: dateRange ? dateRange.dateType : undefined,
            maxTime: dateRange ? dateRange.maxTime : undefined,
            minTime: dateRange ? dateRange.minTime : undefined,
            type: initQuery['type'] && parseInt(initQuery['type']) > 0 ? parseInt(initQuery['type']) : undefined,
            categoryIds: initQuery['categoryIds'],
            accountIds: initQuery['accountIds'],
            tagFilter: initQuery['tagFilter'],
            keyword: initQuery['keyword'],
            matchMode: initQuery['matchMode'] && parseInt(initQuery['matchMode']) >= 0 ? parseInt(initQuery['matchMode']) : undefined
        });

        reload();
    }

    function reload(done?: () => void): void {
        const force = !!done;

        if (!done) {
            loading.value = true;
        }

        const isGalleryMode = pageType.value === TransactionListPageType.Gallery.type;

        transactionInvisibleYearMonths.value = {};
        transactionYearMonthListHeights.value = {};

        Promise.all([
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false }),
            transactionTagsStore.loadAllTags({ force: false })
        ]).then(() => {
            if (queryMonthlyData.value) {
                const currentMonthMinDate = parseDateTimeFromUnixTime(query.value.minTime);
                const currentYear = currentMonthMinDate.getGregorianCalendarYear();
                const currentMonth = currentMonthMinDate.getGregorianCalendarMonth();

                return transactionsStore.loadMonthlyAllTransactions({
                    year: currentYear,
                    month: currentMonth,
                    mustHavePictures: isGalleryMode,
                    withPictures: isGalleryMode,
                    autoExpand: true,
                    defaultCurrency: selectedAccountDefaultCurrency.value
                });
            } else {
                return transactionsStore.loadTransactions({
                    reload: true,
                    mustHavePictures: isGalleryMode,
                    withPictures: isGalleryMode,
                    autoExpand: true,
                    defaultCurrency: selectedAccountDefaultCurrency.value
                });
            }
        }).then(() => {
            done?.();

            if (force) {
                showToast('Data has been updated');
            }

            loading.value = false;
            setTransactionMonthListHeights(true);
        }).catch(error => {
            if (error.processed || done) {
                loading.value = false;
            }

            done?.();

            if (!error.processed) {
                if (!done) {
                    loadingError.value = error;
                }

                showToast(error.message || error);
            }
        });
    }

    function loadMore(autoExpand: boolean): void {
        if (!hasMoreTransaction.value) {
            return;
        }

        if (loadingMore.value || loading.value) {
            return;
        }

        const isGalleryMode = pageType.value === TransactionListPageType.Gallery.type;

        loadingMore.value = true;

        transactionsStore.loadTransactions({
            reload: false,
            mustHavePictures: isGalleryMode,
            withPictures: isGalleryMode,
            autoExpand: autoExpand,
            defaultCurrency: selectedAccountDefaultCurrency.value
        }).then(() => {
            loadingMore.value = false;
            setTransactionMonthListHeights(false);
        }).catch(error => {
            loadingMore.value = false;

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    return { init, reload, loadMore };
}
