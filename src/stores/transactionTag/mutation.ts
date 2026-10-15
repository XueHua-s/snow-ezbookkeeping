import { itemAndIndex, values } from '@/core/base.ts';
import { isEquals } from '@/lib/common.ts';
import type { TransactionTag } from '@/models/transaction_tag.ts';
import type { TransactionTagGroup } from '@/models/transaction_tag_group.ts';
import type { useTransactionTagState } from './state.ts';

interface Context {
    allTransactionTagGroups: ReturnType<typeof useTransactionTagState>['allTransactionTagGroups'];
    allTransactionTagGroupsMap: ReturnType<typeof useTransactionTagState>['allTransactionTagGroupsMap'];
    allTransactionTags: ReturnType<typeof useTransactionTagState>['allTransactionTags'];
    allTransactionTagsMap: ReturnType<typeof useTransactionTagState>['allTransactionTagsMap'];
    allTransactionTagsByGroupMap: ReturnType<typeof useTransactionTagState>['allTransactionTagsByGroupMap'];
    transactionTagListStateInvalid: ReturnType<typeof useTransactionTagState>['transactionTagListStateInvalid'];
    transactionTagGroupListStateInvalid: ReturnType<typeof useTransactionTagState>['transactionTagGroupListStateInvalid'];
    loadingPromise: ReturnType<typeof useTransactionTagState>['loadingPromise'];
}

// Owns mutation for the transactionTag store.
export function useTransactionTagMutation(context: Context) {
    const {
        allTransactionTagGroups, allTransactionTagGroupsMap, allTransactionTags, allTransactionTagsMap,
        allTransactionTagsByGroupMap, transactionTagListStateInvalid, transactionTagGroupListStateInvalid, loadingPromise,
    } = context;

    function loadTransactionTagGroupList(tagGroups: TransactionTagGroup[]): void {
        allTransactionTagGroups.value = tagGroups;
        allTransactionTagGroupsMap.value = {};

        for (const tagGroup of tagGroups) {
            allTransactionTagGroupsMap.value[tagGroup.id] = tagGroup;
        }
    }

    function loadTransactionTagList(tags: TransactionTag[]): void {
        allTransactionTags.value = tags;
        allTransactionTagsMap.value = {};
        allTransactionTagsByGroupMap.value = {};

        for (const tag of tags) {
            allTransactionTagsMap.value[tag.id] = tag;
        }

        for (const tag of tags) {
            let tagsInGroup = allTransactionTagsByGroupMap.value[tag.groupId];

            if (!tagsInGroup) {
                tagsInGroup = [];
                allTransactionTagsByGroupMap.value[tag.groupId] = tagsInGroup;
            }

            tagsInGroup.push(tag);
        }
    }

    function addTagGroupToTransactionTagGroupList(tagGroup: TransactionTagGroup): void {
        allTransactionTagGroups.value.push(tagGroup);
        allTransactionTagGroupsMap.value[tagGroup.id] = tagGroup;
    }

    function addTagToTransactionTagList(tag: TransactionTag): void {
        allTransactionTags.value.push(tag);
        allTransactionTagsMap.value[tag.id] = tag;

        let tagsInGroup = allTransactionTagsByGroupMap.value[tag.groupId];

        if (!tagsInGroup) {
            tagsInGroup = [];
            allTransactionTagsByGroupMap.value[tag.groupId] = tagsInGroup;
        }

        tagsInGroup.push(tag);
    }

    function updateTagGroupInTransactionTagGroupList(currentTagGroup: TransactionTagGroup): void {
        for (const [transactionTagGroup, index] of itemAndIndex(allTransactionTagGroups.value)) {
            if (transactionTagGroup.id === currentTagGroup.id) {
                allTransactionTagGroups.value.splice(index, 1, currentTagGroup);
                break;
            }
        }

        allTransactionTagGroupsMap.value[currentTagGroup.id] = currentTagGroup;
    }

    function updateTagInTransactionTagList(currentTag: TransactionTag, oldTagGroupId?: string): void {
        // update in the main list
        for (const [transactionTag, index] of itemAndIndex(allTransactionTags.value)) {
            if (transactionTag.id === currentTag.id) {
                if (oldTagGroupId && oldTagGroupId !== currentTag.groupId) {
                    allTransactionTags.value.splice(index, 1);
                } else {
                    allTransactionTags.value.splice(index, 1, currentTag);
                }
                break;
            }
        }

        if (oldTagGroupId && oldTagGroupId !== currentTag.groupId) {
            let insertIndex = allTransactionTags.value.length;

            for (const [tag, index] of itemAndIndex(allTransactionTags.value)) {
                if (tag.groupId === currentTag.groupId) {
                    insertIndex = index;
                    break;
                }
            }

            allTransactionTags.value.splice(insertIndex, 0, currentTag);
        }

        // update in the map
        allTransactionTagsMap.value[currentTag.id] = currentTag;

        // update in the group list
        for (const tags of values(allTransactionTagsByGroupMap.value)) {
            for (const [transactionTag, index] of itemAndIndex(tags)) {
                if (transactionTag.id === currentTag.id) {
                    if (oldTagGroupId && oldTagGroupId !== currentTag.groupId) {
                        tags.splice(index, 1);
                    } else {
                        tags.splice(index, 1, currentTag);
                    }
                    break;
                }
            }
        }

        if (oldTagGroupId && oldTagGroupId !== currentTag.groupId) {
            let newGroupTags = allTransactionTagsByGroupMap.value[currentTag.groupId];

            if (!newGroupTags) {
                newGroupTags = [];
                allTransactionTagsByGroupMap.value[currentTag.groupId] = newGroupTags;
            }

            newGroupTags.push(currentTag);
        }
    }

    function updateTagGroupDisplayOrderInTransactionTagList({ from, to }: { from: number, to: number }): void {
        allTransactionTagGroups.value.splice(to, 0, allTransactionTagGroups.value.splice(from, 1)[0] as TransactionTagGroup);
    }

    function sortTagDisplayOrderByTagName(groupId: string, desc: boolean): boolean {
        // update in the group list
        const tagsInGroup = allTransactionTagsByGroupMap.value[groupId];

        if (!tagsInGroup) {
            return false;
        }

        const oldTagsInGroup: TransactionTag[] = [...tagsInGroup];

        if (!desc) {
            tagsInGroup.sort((a, b) => a.name.localeCompare(b.name, undefined, { // asc
                numeric: true,
                sensitivity: 'base'
            }));
        } else {
            tagsInGroup.sort((a, b) => b.name.localeCompare(a.name, undefined, { // desc
                numeric: true,
                sensitivity: 'base'
            }));
        }

        const isOrderChanged = !isEquals(oldTagsInGroup, tagsInGroup);

        if (!isOrderChanged) {
            return false;
        }

        // update in the main list
        for (const [oldTag, oldIndex] of itemAndIndex(oldTagsInGroup)) {
            const newTag = tagsInGroup[oldIndex];

            if (!newTag || newTag.id === oldTag.id) {
                continue;
            }

            for (const [tag, index] of itemAndIndex(allTransactionTags.value)) {
                if (tag.id === oldTag.id) {
                    allTransactionTags.value.splice(index, 1, newTag);
                    break;
                }
            }
        }

        if (!transactionTagListStateInvalid.value) {
            updateTransactionTagListInvalidState(true);
        }

        return true;
    }

    function updateTagDisplayOrderInTransactionTagList({ groupId, from, to }: { groupId: string, from: number, to: number }): void {
        // update in the group list
        const tagsInGroup = allTransactionTagsByGroupMap.value[groupId];

        if (!tagsInGroup) {
            return;
        }

        const fromTag = tagsInGroup[from];

        if (!fromTag) {
            return;
        }

        const toTag = tagsInGroup[to];

        if (!toTag) {
            return;
        }

        tagsInGroup.splice(to, 0, tagsInGroup.splice(from, 1)[0] as TransactionTag);

        // update in the main list
        let mainListFromIndex = -1;
        let mainListToIndex = -1;

        for (const [tag, index] of itemAndIndex(allTransactionTags.value)) {
            if (tag.id === fromTag.id) {
                mainListFromIndex = index;
            }

            if (tag.id === toTag.id) {
                mainListToIndex = index;
            }
        }

        if (mainListFromIndex === -1 || mainListToIndex === -1) {
            return;
        }

        allTransactionTags.value.splice(mainListToIndex, 0, allTransactionTags.value.splice(mainListFromIndex, 1)[0] as TransactionTag);
    }

    function updateTagVisibilityInTransactionTagList({ tag, hidden }: { tag: TransactionTag, hidden: boolean }): void {
        if (allTransactionTagsMap.value[tag.id]) {
            allTransactionTagsMap.value[tag.id]!.hidden = hidden;
        }
    }

    function removeTagGroupFromTransactionTagGroupList(currentTagGroup: TransactionTagGroup): void {
        for (const [transactionTagGroup, index] of itemAndIndex(allTransactionTagGroups.value)) {
            if (transactionTagGroup.id === currentTagGroup.id) {
                allTransactionTagGroups.value.splice(index, 1);
                break;
            }
        }

        if (allTransactionTagGroupsMap.value[currentTagGroup.id]) {
            delete allTransactionTagGroupsMap.value[currentTagGroup.id];
        }
    }

    function removeTagFromTransactionTagList(currentTag: TransactionTag): void {
        for (const [transactionTag, index] of itemAndIndex(allTransactionTags.value)) {
            if (transactionTag.id === currentTag.id) {
                allTransactionTags.value.splice(index, 1);
                break;
            }
        }

        if (allTransactionTagsMap.value[currentTag.id]) {
            delete allTransactionTagsMap.value[currentTag.id];
        }

        for (const tags of values(allTransactionTagsByGroupMap.value)) {
            for (const [transactionTag, index] of itemAndIndex(tags)) {
                if (transactionTag.id === currentTag.id) {
                    tags.splice(index, 1);
                    break;
                }
            }
        }
    }

    function updateTransactionTagGroupListInvalidState(invalidState: boolean): void {
        transactionTagGroupListStateInvalid.value = invalidState;
    }

    function updateTransactionTagListInvalidState(invalidState: boolean): void {
        transactionTagListStateInvalid.value = invalidState;
    }

    function resetTransactionTags(): void {
        loadingPromise.value = null;
        allTransactionTagGroups.value = [];
        allTransactionTagGroupsMap.value = {};
        allTransactionTags.value = [];
        allTransactionTagsMap.value = {};
        allTransactionTagsByGroupMap.value = {};
        transactionTagGroupListStateInvalid.value = true;
        transactionTagListStateInvalid.value = true;
    }

    return {
        loadTransactionTagGroupList, loadTransactionTagList, addTagGroupToTransactionTagGroupList, addTagToTransactionTagList,
        updateTagGroupInTransactionTagGroupList, updateTagInTransactionTagList, updateTagGroupDisplayOrderInTransactionTagList,
        sortTagDisplayOrderByTagName, updateTagDisplayOrderInTransactionTagList, updateTagVisibilityInTransactionTagList,
        removeTagGroupFromTransactionTagGroupList, removeTagFromTransactionTagList, updateTransactionTagGroupListInvalidState,
        updateTransactionTagListInvalidState, resetTransactionTags,
    };
}
