import { defineStore } from 'pinia';
import { useTransactionTagGroups } from './transactionTag/groups.ts';
import { useTransactionTagLoading } from './transactionTag/loading.ts';
import { useTransactionTagMutation } from './transactionTag/mutation.ts';
import { useTransactionTagState } from './transactionTag/state.ts';
import { useTransactionTagTags } from './transactionTag/tags.ts';






export const useTransactionTagsStore = defineStore('transactionTags', () => {
    const {
        loadingPromise, allTransactionTagGroups, allTransactionTagGroupsMap, allTransactionTags, allTransactionTagsMap,
        allTransactionTagsByGroupMap, transactionTagGroupListStateInvalid, transactionTagListStateInvalid, allAvailableTagsCount,
    } = useTransactionTagState();
    const {
        loadTransactionTagGroupList, loadTransactionTagList, addTagGroupToTransactionTagGroupList, addTagToTransactionTagList,
        updateTagGroupInTransactionTagGroupList, updateTagInTransactionTagList, updateTagGroupDisplayOrderInTransactionTagList,
        sortTagDisplayOrderByTagName, updateTagDisplayOrderInTransactionTagList, updateTagVisibilityInTransactionTagList,
        removeTagGroupFromTransactionTagGroupList, removeTagFromTransactionTagList, updateTransactionTagGroupListInvalidState,
        updateTransactionTagListInvalidState, resetTransactionTags,
    } = useTransactionTagMutation({
        allTransactionTagGroups, allTransactionTagGroupsMap, allTransactionTags, allTransactionTagsMap,
        allTransactionTagsByGroupMap, transactionTagListStateInvalid, transactionTagGroupListStateInvalid, loadingPromise,
    });
    const { loadAllTagGroups, loadAllTags } = useTransactionTagLoading({
        transactionTagGroupListStateInvalid, allTransactionTagGroups, updateTransactionTagGroupListInvalidState,
        loadTransactionTagGroupList, loadingPromise, transactionTagListStateInvalid, allTransactionTags,
        updateTransactionTagListInvalidState, loadTransactionTagList,
    });
    const { saveTagGroup, changeTagGroupDisplayOrder, updateTagGroupDisplayOrders, deleteTagGroup } = useTransactionTagGroups({
        addTagGroupToTransactionTagGroupList, updateTagGroupInTransactionTagGroupList, allTransactionTagGroups,
        transactionTagGroupListStateInvalid, updateTransactionTagGroupListInvalidState,
        updateTagGroupDisplayOrderInTransactionTagList, loadAllTagGroups, removeTagGroupFromTransactionTagGroupList,
    });
    const { saveTag, addTags, changeTagDisplayOrder, updateTagDisplayOrders, hideTag, deleteTag } = useTransactionTagTags({
        allTransactionTagsMap, addTagToTransactionTagList, updateTagInTransactionTagList, transactionTagListStateInvalid,
        updateTransactionTagListInvalidState, allTransactionTagsByGroupMap, updateTagDisplayOrderInTransactionTagList,
        loadAllTags, updateTagVisibilityInTransactionTagList, removeTagFromTransactionTagList,
    });

    return {
        // states
        allTransactionTagGroups,
        allTransactionTagGroupsMap,
        allTransactionTagsMap,
        allTransactionTagsByGroupMap,
        transactionTagGroupListStateInvalid,
        transactionTagListStateInvalid,
        // computed states
        allAvailableTagsCount,
        // functions
        sortTagDisplayOrderByTagName,
        updateTransactionTagListInvalidState,
        resetTransactionTags,
        loadAllTagGroups,
        loadAllTags,
        saveTagGroup,
        changeTagGroupDisplayOrder,
        updateTagGroupDisplayOrders,
        deleteTagGroup,
        saveTag,
        addTags,
        changeTagDisplayOrder,
        updateTagDisplayOrders,
        hideTag,
        deleteTag
    }
});
