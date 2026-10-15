import {
    objectFieldWithValueToArrayItem
} from '@/lib/common.ts';
import {
    getCurrentUnixTime,
    parseDateTimeFromUnixTime
} from '@/lib/datetime.ts';
import type { TransactionTemplate } from '@/models/transaction_template.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import { computed } from 'vue';
import type { useListPageState } from './state.ts';

interface Context {
    temporaryCountPerPage: ReturnType<typeof useListPageState>['temporaryCountPerPage'];
    settingsStore: ReturnType<typeof useListPageState>['settingsStore'];
    totalCount: ReturnType<typeof useListPageState>['totalCount'];
    currentPage: ReturnType<typeof useListPageState>['currentPage'];
    queryMonthlyData: ReturnType<typeof useListPageState>['queryMonthlyData'];
    loading: ReturnType<typeof useListPageState>['loading'];
    pageType: ReturnType<typeof useListPageState>['pageType'];
    accountsStore: ReturnType<typeof useListPageState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useListPageState>['transactionCategoriesStore'];
    transactionTagsStore: ReturnType<typeof useListPageState>['transactionTagsStore'];
    desktopPageStore: ReturnType<typeof useListPageState>['desktopPageStore'];
    query: ReturnType<typeof useListPageState>['query'];
    transactionsStore: ReturnType<typeof useListPageState>['transactionsStore'];
    selectedAccountDefaultCurrency: ReturnType<typeof useListPageState>['selectedAccountDefaultCurrency'];
    currentPageTransactions: ReturnType<typeof useListPageState>['currentPageTransactions'];
    snackbar: ReturnType<typeof useListPageState>['snackbar'];
    editDialog: ReturnType<typeof useListPageState>['editDialog'];
    queryAllFilterCategoryIdsCount: ReturnType<typeof useListPageState>['queryAllFilterCategoryIdsCount'];
    queryAllFilterAccountIdsCount: ReturnType<typeof useListPageState>['queryAllFilterAccountIdsCount'];
    queryAllFilterTagIds: ReturnType<typeof useListPageState>['queryAllFilterTagIds'];
}

export function useListPageLoading(context: Context) {
    const {
        temporaryCountPerPage, settingsStore, totalCount, currentPage, queryMonthlyData, loading, pageType, accountsStore,
        transactionCategoriesStore, transactionTagsStore, desktopPageStore, query, transactionsStore,
        selectedAccountDefaultCurrency, currentPageTransactions, snackbar, editDialog, queryAllFilterCategoryIdsCount,
        queryAllFilterAccountIdsCount, queryAllFilterTagIds,
    } = context;

    const countPerPage = computed<number>({
        get: () => {
            if (temporaryCountPerPage.value) {
                return temporaryCountPerPage.value;
            }

            return settingsStore.appSettings.itemsCountInTransactionListPage;
        },
        set: (value) => {
            const newTotalPageCount = Math.ceil(totalCount.value / value);

            if (currentPage.value > newTotalPageCount) {
                currentPage.value = newTotalPageCount;
            }

            temporaryCountPerPage.value = value;

            if (!queryMonthlyData.value) {
                reload(false, false);
            }
        }
    });

    const paginationCurrentPage = computed<number>({
        get: () => currentPage.value,
        set: (value) => {
            currentPage.value = value;

            if (!queryMonthlyData.value) {
                reload(false, false);
            }
        }
    });

    function reload(force: boolean, init: boolean): void {
        loading.value = true;

        const isGalleryMode = pageType.value === TransactionListPageType.Gallery.type;
        const page = currentPage.value;

        Promise.all([
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false }),
            transactionTagsStore.loadAllTags({ force: false })
        ]).then(() => {
            if (init) {
                if (desktopPageStore.showAddTransactionDialogInTransactionList) {
                    desktopPageStore.resetShowAddTransactionDialogInTransactionList();
                    add();
                }
            }

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
                    count: countPerPage.value,
                    page: page,
                    mustHavePictures: isGalleryMode,
                    withCount: page <= 1,
                    withPictures: isGalleryMode,
                    autoExpand: true,
                    defaultCurrency: selectedAccountDefaultCurrency.value
                });
            }
        }).then(data => {
            loading.value = false;
            currentPageTransactions.value = data && data.items && data.items.length ? data.items : [];

            if (page <= 1) {
                totalCount.value = data && data.totalCount ? data.totalCount : 1;
            }

            if (force) {
                snackbar.value?.showMessage('Data has been updated');
            }
        }).catch(error => {
            loading.value = false;
            currentPageTransactions.value = [];
            totalCount.value = 1;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function add(template?: TransactionTemplate, autoRecognizeClipboardText?: string): void {
        const currentUnixTime = getCurrentUnixTime();

        let newTransactionTime: number | undefined = undefined;

        if (query.value.maxTime && query.value.minTime) {
            if (query.value.maxTime < currentUnixTime) {
                newTransactionTime = query.value.maxTime;
            } else if (currentUnixTime < query.value.minTime) {
                newTransactionTime = query.value.minTime;
            }
        }

        editDialog.value?.open({
            time: newTransactionTime,
            type: query.value.type,
            categoryId: queryAllFilterCategoryIdsCount.value === 1 ? query.value.categoryIds : '',
            accountId: queryAllFilterAccountIdsCount.value === 1 ? query.value.accountIds : '',
            tagIds: objectFieldWithValueToArrayItem(queryAllFilterTagIds.value, true).join(',') || '',
            template: template,
            autoRecognizeClipboardText: autoRecognizeClipboardText
        }).then(result => {
            if (result && result.message) {
                snackbar.value?.showMessage(result.message);
            }

            reload(false, false);
        }).catch(error => {
            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }

    return { countPerPage, paginationCurrentPage, reload, add };
}
