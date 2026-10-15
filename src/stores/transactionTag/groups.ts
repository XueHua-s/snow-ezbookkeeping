import { type BeforeResolveFunction, itemAndIndex } from '@/core/base.ts';
import logger from '@/lib/logger.ts';
import services,{ type ApiResponsePromise } from '@/lib/services.ts';
import {
    type TransactionTagGroupInfoResponse,
    type TransactionTagGroupNewDisplayOrderRequest,
    TransactionTagGroup
} from '@/models/transaction_tag_group.ts';
import type { useTransactionTagLoading } from './loading.ts';
import type { useTransactionTagMutation } from './mutation.ts';
import type { useTransactionTagState } from './state.ts';

interface Context {
    addTagGroupToTransactionTagGroupList: ReturnType<typeof useTransactionTagMutation>['addTagGroupToTransactionTagGroupList'];
    updateTagGroupInTransactionTagGroupList: ReturnType<typeof useTransactionTagMutation>['updateTagGroupInTransactionTagGroupList'];
    allTransactionTagGroups: ReturnType<typeof useTransactionTagState>['allTransactionTagGroups'];
    transactionTagGroupListStateInvalid: ReturnType<typeof useTransactionTagState>['transactionTagGroupListStateInvalid'];
    updateTransactionTagGroupListInvalidState: ReturnType<typeof useTransactionTagMutation>['updateTransactionTagGroupListInvalidState'];
    updateTagGroupDisplayOrderInTransactionTagList: ReturnType<typeof useTransactionTagMutation>['updateTagGroupDisplayOrderInTransactionTagList'];
    loadAllTagGroups: ReturnType<typeof useTransactionTagLoading>['loadAllTagGroups'];
    removeTagGroupFromTransactionTagGroupList: ReturnType<typeof useTransactionTagMutation>['removeTagGroupFromTransactionTagGroupList'];
}

// Owns groups for the transactionTag store.
export function useTransactionTagGroups(context: Context) {
    const {
        addTagGroupToTransactionTagGroupList, updateTagGroupInTransactionTagGroupList, allTransactionTagGroups,
        transactionTagGroupListStateInvalid, updateTransactionTagGroupListInvalidState,
        updateTagGroupDisplayOrderInTransactionTagList, loadAllTagGroups, removeTagGroupFromTransactionTagGroupList,
    } = context;

    function saveTagGroup({ tagGroup }: { tagGroup: TransactionTagGroup }): Promise<TransactionTagGroup> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<TransactionTagGroupInfoResponse>;

            if (!tagGroup.id) {
                promise = services.addTransactionTagGroup(tagGroup.toCreateRequest());
            } else {
                promise = services.modifyTransactionTagGroup(tagGroup.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!tagGroup.id) {
                        reject({ message: 'Unable to add tag group' });
                    } else {
                        reject({ message: 'Unable to save tag group' });
                    }
                    return;
                }

                const transactionTagGroup = TransactionTagGroup.of(data.result);

                if (!tagGroup.id) {
                    addTagGroupToTransactionTagGroupList(transactionTagGroup);
                } else {
                    updateTagGroupInTransactionTagGroupList(transactionTagGroup);
                }

                resolve(transactionTagGroup);
            }).catch(error => {
                logger.error('failed to save tag group', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!tagGroup.id) {
                        reject({ message: 'Unable to add tag group' });
                    } else {
                        reject({ message: 'Unable to save tag group' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeTagGroupDisplayOrder({ tagGroupId, from, to }: { tagGroupId: string, from: number, to: number }): Promise<void> {
        return new Promise((resolve, reject) => {
            let currentTagGroup: TransactionTagGroup | null = null;

            for (const transactionTagGroup of allTransactionTagGroups.value) {
                if (transactionTagGroup.id === tagGroupId) {
                    currentTagGroup = transactionTagGroup;
                    break;
                }
            }

            if (!currentTagGroup || !allTransactionTagGroups.value[to]) {
                reject({ message: 'Unable to move tag group' });
                return;
            }

            if (!transactionTagGroupListStateInvalid.value) {
                updateTransactionTagGroupListInvalidState(true);
            }

            updateTagGroupDisplayOrderInTransactionTagList({ from, to });

            resolve();
        });
    }

    function updateTagGroupDisplayOrders(): Promise<boolean> {
        const newDisplayOrders: TransactionTagGroupNewDisplayOrderRequest[] = [];

        for (const [transactionTagGroup, index] of itemAndIndex(allTransactionTagGroups.value)) {
            newDisplayOrders.push({
                id: transactionTagGroup.id,
                displayOrder: index + 1
            });
        }

        return new Promise((resolve, reject) => {
            services.moveTransactionTagGroup({
                newDisplayOrders: newDisplayOrders
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move tag group' });
                    return;
                }

                loadAllTagGroups({ force: false }).finally(() => {
                    if (transactionTagGroupListStateInvalid.value) {
                        updateTransactionTagGroupListInvalidState(false);
                    }

                    resolve(data.result);
                });
            }).catch(error => {
                logger.error('failed to save tag groups display order', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move tag group' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteTagGroup({ tagGroup, beforeResolve }: { tagGroup: TransactionTagGroup, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteTransactionTagGroup({
                id: tagGroup.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this tag group' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeTagGroupFromTransactionTagGroupList(tagGroup);
                    });
                } else {
                    removeTagGroupFromTransactionTagGroupList(tagGroup);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete tag group', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this tag group' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { saveTagGroup, changeTagGroupDisplayOrder, updateTagGroupDisplayOrders, deleteTagGroup };
}
