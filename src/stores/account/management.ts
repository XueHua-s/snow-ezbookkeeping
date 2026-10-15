import { type BeforeResolveFunction, itemAndIndex, values } from '@/core/base.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import type { AccountNewDisplayOrderRequest, Account } from '@/models/account.ts';
import type { useAccountLoading } from './loading.ts';
import type { useAccountMutation } from './mutation.ts';
import type { useAccountState } from './state.ts';

interface Context {
    allAccountsMap: ReturnType<typeof useAccountState>['allAccountsMap'];
    updateAccountListInvalidState: ReturnType<typeof useAccountMutation>['updateAccountListInvalidState'];
    overviewStore: ReturnType<typeof useAccountState>['overviewStore'];
    allCategorizedAccountsMap: ReturnType<typeof useAccountState>['allCategorizedAccountsMap'];
    accountListStateInvalid: ReturnType<typeof useAccountState>['accountListStateInvalid'];
    updateAccountDisplayOrderInAccountList: ReturnType<typeof useAccountMutation>['updateAccountDisplayOrderInAccountList'];
    loadAllAccounts: ReturnType<typeof useAccountLoading>['loadAllAccounts'];
    updateAccountVisibilityInAccountList: ReturnType<typeof useAccountMutation>['updateAccountVisibilityInAccountList'];
    removeAccountFromAccountList: ReturnType<typeof useAccountMutation>['removeAccountFromAccountList'];
    removeSubAccountFromAccountList: ReturnType<typeof useAccountMutation>['removeSubAccountFromAccountList'];
}

// Owns management for the account store.
export function useAccountManagement(context: Context) {
    const {
        allAccountsMap, updateAccountListInvalidState, overviewStore, allCategorizedAccountsMap, accountListStateInvalid,
        updateAccountDisplayOrderInAccountList, loadAllAccounts, updateAccountVisibilityInAccountList,
        removeAccountFromAccountList, removeSubAccountFromAccountList,
    } = context;

    function updateAccountLastReconciledTime(accountId: string, lastReconciledTime: number): Promise<boolean> {
        return new Promise((resolve, reject) => {
            const account = allAccountsMap.value[accountId];

            services.updateAccountLastReconciledTime({
                id: accountId,
                lastReconciledTime: lastReconciledTime
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to last reconciled time' });
                    return;
                }

                if (account) {
                    account.lastReconciledTime = lastReconciledTime;
                } else {
                    updateAccountListInvalidState(true);
                }

                overviewStore.updateTransactionOverviewInvalidState(true);

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to update last reconciled time', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to update last reconciled time' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeAccountDisplayOrder({ accountId, from, to, updateListOrder, updateGlobalListOrder }: { accountId: string, from: number, to: number, updateListOrder: boolean, updateGlobalListOrder: boolean }): Promise<void> {
        const account = allAccountsMap.value[accountId];

        return new Promise((resolve, reject) => {
            if (!account ||
                !allCategorizedAccountsMap.value[account.category] ||
                !allCategorizedAccountsMap.value[account.category]!.accounts ||
                !allCategorizedAccountsMap.value[account.category]!.accounts[to]) {
                reject({ message: 'Unable to move account' });
                return;
            }

            if (!accountListStateInvalid.value) {
                updateAccountListInvalidState(true);
            }

            updateAccountDisplayOrderInAccountList({ account, from, to, updateListOrder, updateGlobalListOrder });

            resolve();
        });
    }

    function updateAccountDisplayOrders(): Promise<boolean> {
        const newDisplayOrders: AccountNewDisplayOrderRequest[] = [];

        for (const categorizedAccounts of values(allCategorizedAccountsMap.value)) {
            for (const [account, index] of itemAndIndex(categorizedAccounts.accounts)) {
                newDisplayOrders.push({
                    id: account.id,
                    displayOrder: index + 1
                });
            }
        }

        return new Promise((resolve, reject) => {
            services.moveAccount({
                newDisplayOrders: newDisplayOrders
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move account' });
                    return;
                }

                loadAllAccounts({ force: false }).finally(() => {
                    if (accountListStateInvalid.value) {
                        updateAccountListInvalidState(false);
                    }

                    resolve(data.result);
                });
            }).catch(error => {
                logger.error('failed to save accounts display order', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move account' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function hideAccount({ account, hidden }: { account: Account, hidden: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.hideAccount({
                id: account.id,
                hidden: hidden
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this account' });
                    } else {
                        reject({ message: 'Unable to unhide this account' });
                    }

                    return;
                }

                updateAccountVisibilityInAccountList({ account, hidden });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to change account visibility', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this account' });
                    } else {
                        reject({ message: 'Unable to unhide this account' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteAccount({ account, beforeResolve }: { account: Account, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteAccount({
                id: account.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this account' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeAccountFromAccountList(account);
                    });
                } else {
                    removeAccountFromAccountList(account);
                }

                overviewStore.updateTransactionOverviewInvalidState(true);

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete account', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this account' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteSubAccount({ subAccount, beforeResolve }: { subAccount: Account, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteSubAccount({
                id: subAccount.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this sub-account' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeSubAccountFromAccountList(subAccount);
                    });
                } else {
                    removeSubAccountFromAccountList(subAccount);
                }

                overviewStore.updateTransactionOverviewInvalidState(true);

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete sub-account', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this sub-account' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { updateAccountLastReconciledTime, changeAccountDisplayOrder, updateAccountDisplayOrders, hideAccount, deleteAccount, deleteSubAccount };
}
