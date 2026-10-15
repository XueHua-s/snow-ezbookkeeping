import { isEquals } from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import {
    TransactionTag
} from '@/models/transaction_tag.ts';
import {
    TransactionTagGroup
} from '@/models/transaction_tag_group.ts';
import type { useTransactionTagMutation } from './mutation.ts';
import type { useTransactionTagState } from './state.ts';

interface Context {
    transactionTagGroupListStateInvalid: ReturnType<typeof useTransactionTagState>['transactionTagGroupListStateInvalid'];
    allTransactionTagGroups: ReturnType<typeof useTransactionTagState>['allTransactionTagGroups'];
    updateTransactionTagGroupListInvalidState: ReturnType<typeof useTransactionTagMutation>['updateTransactionTagGroupListInvalidState'];
    loadTransactionTagGroupList: ReturnType<typeof useTransactionTagMutation>['loadTransactionTagGroupList'];
    loadingPromise: ReturnType<typeof useTransactionTagState>['loadingPromise'];
    transactionTagListStateInvalid: ReturnType<typeof useTransactionTagState>['transactionTagListStateInvalid'];
    allTransactionTags: ReturnType<typeof useTransactionTagState>['allTransactionTags'];
    updateTransactionTagListInvalidState: ReturnType<typeof useTransactionTagMutation>['updateTransactionTagListInvalidState'];
    loadTransactionTagList: ReturnType<typeof useTransactionTagMutation>['loadTransactionTagList'];
}

// Owns loading for the transactionTag store.
export function useTransactionTagLoading(context: Context) {
    const {
        transactionTagGroupListStateInvalid, allTransactionTagGroups, updateTransactionTagGroupListInvalidState,
        loadTransactionTagGroupList, loadingPromise, transactionTagListStateInvalid, allTransactionTags,
        updateTransactionTagListInvalidState, loadTransactionTagList,
    } = context;

    function loadAllTagGroups({ force }: { force?: boolean }): Promise<TransactionTagGroup[]> {
        if (!force && !transactionTagGroupListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allTransactionTagGroups.value);
            });
        }

        return new Promise((resolve, reject) => {
            services.getAllTransactionTagGroups().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve tag group list' });
                    return;
                }

                if (transactionTagGroupListStateInvalid.value) {
                    updateTransactionTagGroupListInvalidState(false);
                }

                const transactionTagGroups = TransactionTagGroup.ofMulti(data.result);

                loadTransactionTagGroupList(transactionTagGroups);

                resolve(transactionTagGroups);
            }).catch(error => {
                logger.error('failed to load tag group list', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve tag group list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function loadAllTags({ force }: { force?: boolean }): Promise<TransactionTag[]> {
        if (loadingPromise.value) {
            return loadingPromise.value;
        }

        if (!force && !transactionTagGroupListStateInvalid.value && !transactionTagListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allTransactionTags.value);
            });
        }

        const currentPromise = loadingPromise.value = new Promise((resolve, reject) => {
            services.getAllTransactionTagGroups().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve tag list' });
                    return;
                }

                if (transactionTagGroupListStateInvalid.value) {
                    updateTransactionTagGroupListInvalidState(false);
                }

                const transactionTagGroups = TransactionTagGroup.ofMulti(data.result);

                loadTransactionTagGroupList(transactionTagGroups);

                return services.getAllTransactionTags();
            }).then(response => {
                if (!loadingPromise.value || loadingPromise.value !== currentPromise) {
                    logger.error('loadingPromise.value is invalid after retrieving tag list');
                    reject({ message: 'An error occurred' });
                    return;
                }

                if (!response) {
                    loadingPromise.value = null;
                    reject({ message: 'Unable to retrieve tag list' });
                    return;
                }

                const data = response.data;

                if (!data || !data.success || !data.result) {
                    loadingPromise.value = null;
                    reject({ message: 'Unable to retrieve tag list' });
                    return;
                }

                if (transactionTagListStateInvalid.value) {
                    updateTransactionTagListInvalidState(false);
                }

                const transactionTags = TransactionTag.ofMulti(data.result);

                if (force && data.result && isEquals(allTransactionTags.value, transactionTags)) {
                    loadingPromise.value = null;
                    reject({ message: 'Tag list is up to date', isUpToDate: true });
                    return;
                }

                loadTransactionTagList(transactionTags);
                loadingPromise.value = null;

                resolve(transactionTags);
            }).catch(error => {
                if (loadingPromise.value === currentPromise) {
                    loadingPromise.value = null;
                }

                if (force) {
                    logger.error('failed to force load tag list', error);
                } else {
                    logger.error('failed to load tag list', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve tag list' });
                } else {
                    reject(error);
                }
            });
        });

        return loadingPromise.value;
    }

    return { loadAllTagGroups, loadAllTags };
}
