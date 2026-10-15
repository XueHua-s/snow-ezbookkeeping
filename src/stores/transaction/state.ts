import { DateRange } from '@/core/datetime.ts';
import { KeywordMatchMode } from '@/core/text.ts';
import { TransactionTagFilterType } from '@/core/transaction.ts';
import {
    countSplitItems,
    getObjectOwnFieldCount,
    splitItemsToMap
} from '@/lib/common.ts';
import {
    getUserTransactionDraft
} from '@/lib/userstate.ts';
import {
    type TransactionDraft,
    TransactionTagFilter
} from '@/models/transaction.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useExplorersStore } from '@/stores/explorer.ts';
import { useOverviewStore } from '@/stores/overview.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useStatisticsStore } from '@/stores/statistics.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useUserStore } from '@/stores/user.ts';
import type { ComputedRef, Ref, UnwrapRef } from 'vue';
import { computed, ref } from 'vue';
import type { TransactionListFilter, TransactionMonthList } from './types.ts';
interface TransactionState {
    settingsStore: ReturnType<typeof useSettingsStore>;
    userStore: ReturnType<typeof useUserStore>;
    accountsStore: ReturnType<typeof useAccountsStore>;
    transactionCategoriesStore: ReturnType<typeof useTransactionCategoriesStore>;
    overviewStore: ReturnType<typeof useOverviewStore>;
    statisticsStore: ReturnType<typeof useStatisticsStore>;
    explorersStore: ReturnType<typeof useExplorersStore>;
    exchangeRatesStore: ReturnType<typeof useExchangeRatesStore>;
    transactionDraft: Ref<UnwrapRef<TransactionDraft | null>>;
    transactionsFilter: Ref<UnwrapRef<TransactionListFilter>>;
    transactions: Ref<UnwrapRef<TransactionMonthList[]>>;
    transactionsNextTimeId: Ref<UnwrapRef<number>>;
    transactionListStateInvalid: Ref<UnwrapRef<boolean>>;
    transactionReconciliationStatementStateInvalid: Ref<UnwrapRef<boolean>>;
    allFilterCategoryIds: ComputedRef<Record<string, boolean>>;
    allFilterAccountIds: ComputedRef<Record<string, boolean>>;
    allFilterTagIds: ComputedRef<Record<string, boolean>>;
    allFilterCategoryIdsCount: ComputedRef<number>;
    allFilterAccountIdsCount: ComputedRef<number>;
    allFilterTagIdsCount: ComputedRef<number>;
    noTransaction: ComputedRef<boolean>;
    hasMoreTransaction: ComputedRef<boolean>;
}



// Owns state for the transaction store.
export function useTransactionState(): TransactionState  {
    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const overviewStore = useOverviewStore();

    const statisticsStore = useStatisticsStore();

    const explorersStore = useExplorersStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const transactionDraft = ref<TransactionDraft | null>(getUserTransactionDraft());

    const transactionsFilter = ref<TransactionListFilter>({
        dateType: DateRange.All.type,
        maxTime: 0,
        minTime: 0,
        type: 0,
        categoryIds: '',
        accountIds: '',
        tagFilter: '',
        amountFilter: '',
        keyword: '',
        matchMode: KeywordMatchMode.Default.type
    });

    const transactions = ref<TransactionMonthList[]>([]);

    const transactionsNextTimeId = ref<number>(0);

    const transactionListStateInvalid = ref<boolean>(true);

    const transactionReconciliationStatementStateInvalid = ref<boolean>(true);

    const allFilterCategoryIds = computed<Record<string, boolean>>(() => splitItemsToMap(transactionsFilter.value.categoryIds, ','));

    const allFilterAccountIds = computed<Record<string, boolean>>(() => splitItemsToMap(transactionsFilter.value.accountIds, ','));

    const allFilterTagIds = computed<Record<string, boolean>>(() => {
        const tagFilters: TransactionTagFilter[] = TransactionTagFilter.parse(transactionsFilter.value.tagFilter);
        const allTagIdsMap: Record<string, boolean> = {};

        for (const tagFilter of tagFilters) {
            let state: boolean = true;

            if (tagFilter.type === TransactionTagFilterType.HasAny || tagFilter.type === TransactionTagFilterType.HasAll) {
                state = true;
            } else if (tagFilter.type === TransactionTagFilterType.NotHasAny || tagFilter.type === TransactionTagFilterType.NotHasAll) {
                state = false;
            } else {
                continue;
            }

            for (const tagId of tagFilter.tagIds) {
                allTagIdsMap[tagId] = state;
            }
        }

        return allTagIdsMap;
    });

    const allFilterCategoryIdsCount = computed<number>(() => countSplitItems(transactionsFilter.value.categoryIds, ','));

    const allFilterAccountIdsCount = computed<number>(() => countSplitItems(transactionsFilter.value.accountIds, ','));

    const allFilterTagIdsCount = computed<number>(() => getObjectOwnFieldCount(allFilterTagIds.value));

    const noTransaction = computed<boolean>(() => {
        for (const transactionMonthList of transactions.value) {
            for (const transaction of transactionMonthList.items) {
                if (transaction) {
                    return false;
                }
            }
        }

        return true;
    });

    const hasMoreTransaction = computed<boolean>(() => {
        return transactionsNextTimeId.value > 0;
    });

    return {
        settingsStore, userStore, accountsStore, transactionCategoriesStore, overviewStore, statisticsStore, explorersStore,
        exchangeRatesStore, transactionDraft, transactionsFilter, transactions, transactionsNextTimeId,
        transactionListStateInvalid, transactionReconciliationStatementStateInvalid, allFilterCategoryIds, allFilterAccountIds,
        allFilterTagIds, allFilterCategoryIdsCount, allFilterAccountIdsCount, allFilterTagIdsCount, noTransaction,
        hasMoreTransaction,
    };
}
