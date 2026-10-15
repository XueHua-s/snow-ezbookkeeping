import { isEquals } from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import {
    Account
} from '@/models/account.ts';
import type { useAccountMutation } from './mutation.ts';
import type { useAccountState } from './state.ts';

interface Context {
    loadingPromise: ReturnType<typeof useAccountState>['loadingPromise'];
    accountListStateInvalid: ReturnType<typeof useAccountState>['accountListStateInvalid'];
    allAccounts: ReturnType<typeof useAccountState>['allAccounts'];
    updateAccountListInvalidState: ReturnType<typeof useAccountMutation>['updateAccountListInvalidState'];
    settingsStore: ReturnType<typeof useAccountState>['settingsStore'];
    loadAccountList: ReturnType<typeof useAccountMutation>['loadAccountList'];
    allAccountsMap: ReturnType<typeof useAccountState>['allAccountsMap'];
    addAccountToAccountList: ReturnType<typeof useAccountMutation>['addAccountToAccountList'];
    updateAccountToAccountList: ReturnType<typeof useAccountMutation>['updateAccountToAccountList'];
    overviewStore: ReturnType<typeof useAccountState>['overviewStore'];
}

// Owns loading for the account store.
export function useAccountLoading(context: Context) {
    const {
        loadingPromise, accountListStateInvalid, allAccounts, updateAccountListInvalidState, settingsStore, loadAccountList,
        allAccountsMap, addAccountToAccountList, updateAccountToAccountList, overviewStore,
    } = context;

    function loadAllAccounts({ force }: { force: boolean }): Promise<Account[]> {
        if (loadingPromise.value) {
            return loadingPromise.value;
        }

        if (!force && !accountListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allAccounts.value);
            });
        }

        const currentPromise = loadingPromise.value = new Promise((resolve, reject) => {
            services.getAllAccounts({
                visibleOnly: false
            }).then(response => {
                if (!loadingPromise.value || loadingPromise.value !== currentPromise) {
                    logger.error('loadingPromise.value is invalid after retrieving account list');
                    reject({ message: 'An error occurred' });
                    return;
                }

                const data = response.data;

                if (!data || !data.success || !data.result) {
                    loadingPromise.value = null;
                    reject({ message: 'Unable to retrieve account list' });
                    return;
                }

                if (accountListStateInvalid.value) {
                    updateAccountListInvalidState(false);
                }

                const accounts = Account.sortAccounts(Account.ofMulti(data.result), settingsStore.accountCategoryDisplayOrders);

                if (force && data.result && isEquals(allAccounts.value, accounts)) {
                    loadingPromise.value = null;
                    reject({ message: 'Account list is up to date', isUpToDate: true });
                    return;
                }

                loadAccountList(accounts);
                loadingPromise.value = null;

                resolve(accounts);
            }).catch(error => {
                if (loadingPromise.value === currentPromise) {
                    loadingPromise.value = null;
                }

                if (force) {
                    logger.error('failed to force load account list', error);
                } else {
                    logger.error('failed to load account list', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve account list' });
                } else {
                    reject(error);
                }
            });
        });

        return loadingPromise.value;
    }

    function getAccount({ accountId }: { accountId: string }): Promise<Account> {
        return new Promise((resolve, reject) => {
            services.getAccount({
                id: accountId
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve account' });
                    return;
                }

                const account = Account.of(data.result);

                resolve(account);
            }).catch(error => {
                logger.error('failed to load account info', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve account' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveAccount({ account, subAccounts, isEdit, clientSessionId }: { account: Account, subAccounts: Account[], isEdit: boolean, clientSessionId: string }): Promise<Account> {
        return new Promise((resolve, reject) => {
            const oldAccount = isEdit ? allAccountsMap.value[account.id] : null;
            let promise = null;

            if (!isEdit) {
                promise = services.addAccount(account.toCreateRequest(clientSessionId, subAccounts));
            } else {
                promise = services.modifyAccount(account.toModifyRequest(clientSessionId, subAccounts));
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!isEdit) {
                        reject({ message: 'Unable to add account' });
                    } else {
                        reject({ message: 'Unable to save account' });
                    }
                    return;
                }

                const newAccount = Account.of(data.result);

                if (!isEdit) {
                    addAccountToAccountList(newAccount);
                } else {
                    if (oldAccount && oldAccount.category === newAccount.category) {
                        updateAccountToAccountList(oldAccount, newAccount);
                    } else {
                        updateAccountListInvalidState(true);
                    }
                }

                overviewStore.updateTransactionOverviewInvalidState(true);

                resolve(newAccount);
            }).catch(error => {
                logger.error('failed to save account', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!isEdit) {
                        reject({ message: 'Unable to add account' });
                    } else {
                        reject({ message: 'Unable to save account' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    return { loadAllAccounts, getAccount, saveAccount };
}
