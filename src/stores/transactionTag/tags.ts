import { type BeforeResolveFunction, itemAndIndex } from '@/core/base.ts';
import logger from '@/lib/logger.ts';
import services,{ type ApiResponsePromise } from '@/lib/services.ts';
import {
    type TransactionTagCreateBatchRequest,
    type TransactionTagInfoResponse,
    type TransactionTagNewDisplayOrderRequest,
    TransactionTag
} from '@/models/transaction_tag.ts';
import type { useTransactionTagLoading } from './loading.ts';
import type { useTransactionTagMutation } from './mutation.ts';
import type { useTransactionTagState } from './state.ts';

interface Context {
    allTransactionTagsMap: ReturnType<typeof useTransactionTagState>['allTransactionTagsMap'];
    addTagToTransactionTagList: ReturnType<typeof useTransactionTagMutation>['addTagToTransactionTagList'];
    updateTagInTransactionTagList: ReturnType<typeof useTransactionTagMutation>['updateTagInTransactionTagList'];
    transactionTagListStateInvalid: ReturnType<typeof useTransactionTagState>['transactionTagListStateInvalid'];
    updateTransactionTagListInvalidState: ReturnType<typeof useTransactionTagMutation>['updateTransactionTagListInvalidState'];
    allTransactionTagsByGroupMap: ReturnType<typeof useTransactionTagState>['allTransactionTagsByGroupMap'];
    updateTagDisplayOrderInTransactionTagList: ReturnType<typeof useTransactionTagMutation>['updateTagDisplayOrderInTransactionTagList'];
    loadAllTags: ReturnType<typeof useTransactionTagLoading>['loadAllTags'];
    updateTagVisibilityInTransactionTagList: ReturnType<typeof useTransactionTagMutation>['updateTagVisibilityInTransactionTagList'];
    removeTagFromTransactionTagList: ReturnType<typeof useTransactionTagMutation>['removeTagFromTransactionTagList'];
}

// Owns tags for the transactionTag store.
export function useTransactionTagTags(context: Context) {
    const {
        allTransactionTagsMap, addTagToTransactionTagList, updateTagInTransactionTagList, transactionTagListStateInvalid,
        updateTransactionTagListInvalidState, allTransactionTagsByGroupMap, updateTagDisplayOrderInTransactionTagList,
        loadAllTags, updateTagVisibilityInTransactionTagList, removeTagFromTransactionTagList,
    } = context;

    function saveTag({ tag, beforeResolve }: { tag: TransactionTag, beforeResolve?: BeforeResolveFunction }): Promise<TransactionTag> {
        const oldTagGroupId = allTransactionTagsMap.value[tag.id]?.groupId;

        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<TransactionTagInfoResponse>;

            if (!tag.id) {
                promise = services.addTransactionTag(tag.toCreateRequest());
            } else {
                promise = services.modifyTransactionTag(tag.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!tag.id) {
                        reject({ message: 'Unable to add tag' });
                    } else {
                        reject({ message: 'Unable to save tag' });
                    }
                    return;
                }

                const transactionTag = TransactionTag.of(data.result);

                if (beforeResolve) {
                    beforeResolve(() => {
                        if (!tag.id) {
                            addTagToTransactionTagList(transactionTag);
                        } else {
                            updateTagInTransactionTagList(transactionTag, oldTagGroupId);
                        }
                    });
                } else {
                    if (!tag.id) {
                        addTagToTransactionTagList(transactionTag);
                    } else {
                        updateTagInTransactionTagList(transactionTag, oldTagGroupId);
                    }
                }

                resolve(transactionTag);
            }).catch(error => {
                logger.error('failed to save tag', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!tag.id) {
                        reject({ message: 'Unable to add tag' });
                    } else {
                        reject({ message: 'Unable to save tag' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function addTags(req: TransactionTagCreateBatchRequest): Promise<TransactionTag[]> {
        return new Promise((resolve, reject) => {
            services.addTransactionTagBatch(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to add tag' });
                    return;
                }

                if (!transactionTagListStateInvalid.value) {
                    updateTransactionTagListInvalidState(true);
                }

                const transactionTags = TransactionTag.ofMulti(data.result);

                resolve(transactionTags);
            }).catch(error => {
                logger.error('failed to add tags', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to add tag' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeTagDisplayOrder({ tagId, from, to }: { tagId: string, from: number, to: number }): Promise<void> {
        return new Promise((resolve, reject) => {
            const currentTag = allTransactionTagsMap.value[tagId];

            if (!currentTag || !allTransactionTagsByGroupMap.value[currentTag.groupId] || !allTransactionTagsByGroupMap.value[currentTag.groupId]![to]) {
                reject({ message: 'Unable to move tag' });
                return;
            }

            if (!transactionTagListStateInvalid.value) {
                updateTransactionTagListInvalidState(true);
            }

            updateTagDisplayOrderInTransactionTagList({
                groupId: currentTag.groupId,
                from: from,
                to: to
            });

            resolve();
        });
    }

    function updateTagDisplayOrders(groupId: string): Promise<boolean> {
        const tagsInGroup = allTransactionTagsByGroupMap.value[groupId];

        if (!tagsInGroup) {
            return Promise.reject('Unable to move tag');
        }

        const newDisplayOrders: TransactionTagNewDisplayOrderRequest[] = [];

        for (const [transactionTag, index] of itemAndIndex(tagsInGroup)) {
            newDisplayOrders.push({
                id: transactionTag.id,
                displayOrder: index + 1
            });
        }

        return new Promise((resolve, reject) => {
            services.moveTransactionTag({
                newDisplayOrders: newDisplayOrders
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move tag' });
                    return;
                }

                loadAllTags({ force: false }).finally(() => {
                    if (transactionTagListStateInvalid.value) {
                        updateTransactionTagListInvalidState(false);
                    }

                    resolve(data.result);
                });
            }).catch(error => {
                logger.error('failed to save tags display order', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move tag' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function hideTag({ tag, hidden }: { tag: TransactionTag, hidden: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.hideTransactionTag({
                id: tag.id,
                hidden: hidden
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this tag' });
                    } else {
                        reject({ message: 'Unable to unhide this tag' });
                    }
                    return;
                }

                updateTagVisibilityInTransactionTagList({ tag, hidden });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to change tag visibility', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this tag' });
                    } else {
                        reject({ message: 'Unable to unhide this tag' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteTag({ tag, beforeResolve }: { tag: TransactionTag, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteTransactionTag({
                id: tag.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this tag' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeTagFromTransactionTagList(tag);
                    });
                } else {
                    removeTagFromTransactionTagList(tag);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete tag', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this tag' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { saveTag, addTags, changeTagDisplayOrder, updateTagDisplayOrders, hideTag, deleteTag };
}
