import type { TransactionTag } from '@/models/transaction_tag.ts';
import type { TransactionTagGroup } from '@/models/transaction_tag_group.ts';
import type { ComputedRef, Ref, ShallowRef, UnwrapRef } from 'vue';
import { computed, ref, shallowRef } from 'vue';
interface TransactionTagState {
    loadingPromise: ShallowRef<Promise<TransactionTag[]> | null>;
    allTransactionTagGroups: Ref<UnwrapRef<TransactionTagGroup[]>>;
    allTransactionTagGroupsMap: Ref<UnwrapRef<Record<string, TransactionTagGroup>>>;
    allTransactionTags: Ref<UnwrapRef<TransactionTag[]>>;
    allTransactionTagsMap: Ref<UnwrapRef<Record<string, TransactionTag>>>;
    allTransactionTagsByGroupMap: Ref<UnwrapRef<Record<string, TransactionTag[]>>>;
    transactionTagGroupListStateInvalid: Ref<UnwrapRef<boolean>>;
    transactionTagListStateInvalid: Ref<UnwrapRef<boolean>>;
    allAvailableTagsCount: ComputedRef<number>;
}



// Owns state for the transactionTag store.
export function useTransactionTagState(): TransactionTagState  {
    const loadingPromise = shallowRef<Promise<TransactionTag[]> | null>(null);

    const allTransactionTagGroups = ref<TransactionTagGroup[]>([]);

    const allTransactionTagGroupsMap = ref<Record<string, TransactionTagGroup>>({});

    const allTransactionTags = ref<TransactionTag[]>([]);

    const allTransactionTagsMap = ref<Record<string, TransactionTag>>({});

    const allTransactionTagsByGroupMap = ref<Record<string, TransactionTag[]>>({});

    const transactionTagGroupListStateInvalid = ref<boolean>(true);

    const transactionTagListStateInvalid = ref<boolean>(true);

    const allAvailableTagsCount = computed<number>(() => allTransactionTags.value.length);

    return {
        loadingPromise, allTransactionTagGroups, allTransactionTagGroupsMap, allTransactionTags, allTransactionTagsMap,
        allTransactionTagsByGroupMap, transactionTagGroupListStateInvalid, transactionTagListStateInvalid, allAvailableTagsCount,
    };
}
