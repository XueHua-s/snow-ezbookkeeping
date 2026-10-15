import { AccountCategory, AccountType } from '@/core/account.ts';
import { itemAndIndex } from '@/core/base.ts';
import { getCategorizedAccountsMap } from '@/lib/account.ts';
import type { Account } from '@/models/account.ts';
import type { useAccountState } from './state.ts';

interface Context {
    allAccounts: ReturnType<typeof useAccountState>['allAccounts'];
    allAccountsMap: ReturnType<typeof useAccountState>['allAccountsMap'];
    allCategorizedAccountsMap: ReturnType<typeof useAccountState>['allCategorizedAccountsMap'];
    settingsStore: ReturnType<typeof useAccountState>['settingsStore'];
    accountListStateInvalid: ReturnType<typeof useAccountState>['accountListStateInvalid'];
    loadingPromise: ReturnType<typeof useAccountState>['loadingPromise'];
}

// Owns mutation for the account store.
export function useAccountMutation(context: Context) {
    const { allAccounts, allAccountsMap, allCategorizedAccountsMap, settingsStore, accountListStateInvalid, loadingPromise } = context;

    function loadAccountList(accounts: Account[]): void {
        allAccounts.value = accounts;
        allAccountsMap.value = {};

        for (const account of accounts) {
            allAccountsMap.value[account.id] = account;

            if (account.subAccounts) {
                for (const subAccount of account.subAccounts) {
                    allAccountsMap.value[subAccount.id] = subAccount;
                }
            }
        }

        allCategorizedAccountsMap.value = getCategorizedAccountsMap(accounts);
    }

    function addAccountToAccountList(currentAccount: Account): void {
        const newAccountCategory = AccountCategory.valueOf(currentAccount.category);
        let insertIndexToAllList = allAccounts.value.length;

        if (newAccountCategory) {
            for (const [account, index] of itemAndIndex(allAccounts.value)) {
                const accountCategory = AccountCategory.valueOf(account.category);
                const accountCategoryDisplayOrder = settingsStore.accountCategoryDisplayOrders[accountCategory?.type ?? 0] || Number.MAX_SAFE_INTEGER;
                const newAccountCategoryDisplayOrder = settingsStore.accountCategoryDisplayOrders[newAccountCategory.type] || Number.MAX_SAFE_INTEGER;

                if (accountCategory && accountCategoryDisplayOrder > newAccountCategoryDisplayOrder) {
                    insertIndexToAllList = index;
                    break;
                }
            }
        }

        allAccounts.value.splice(insertIndexToAllList, 0, currentAccount);

        allAccountsMap.value[currentAccount.id] = currentAccount;

        if (currentAccount.subAccounts) {
            for (const subAccount of currentAccount.subAccounts) {
                allAccountsMap.value[subAccount.id] = subAccount;
            }
        }

        if (allCategorizedAccountsMap.value[currentAccount.category]) {
            const accountList = allCategorizedAccountsMap.value[currentAccount.category]!.accounts;
            accountList.push(currentAccount);
        } else {
            allCategorizedAccountsMap.value = getCategorizedAccountsMap(allAccounts.value);
        }
    }

    function updateAccountToAccountList(oldAccount: Account, newAccount: Account): void {
        for (const [account, index] of itemAndIndex(allAccounts.value)) {
            if (account.id === newAccount.id) {
                allAccounts.value.splice(index, 1, newAccount);
                break;
            }
        }

        if (oldAccount.subAccounts) {
            for (const subAccount of oldAccount.subAccounts) {
                if (allAccountsMap.value[subAccount.id]) {
                    delete allAccountsMap.value[subAccount.id];
                }
            }
        }

        allAccountsMap.value[newAccount.id] = newAccount;

        if (newAccount.subAccounts) {
            for (const subAccount of newAccount.subAccounts) {
                allAccountsMap.value[subAccount.id] = subAccount;
            }
        }

        if (allCategorizedAccountsMap.value[newAccount.category]) {
            const accountList = allCategorizedAccountsMap.value[newAccount.category]!.accounts;

            for (const [account, index] of itemAndIndex(accountList)) {
                if (account.id === newAccount.id) {
                    accountList.splice(index, 1, newAccount);
                    break;
                }
            }
        }
    }

    function updateAccountDisplayOrderInAccountList({ account, from, to, updateListOrder, updateGlobalListOrder }: { account: Account, from: number, to: number, updateListOrder: boolean, updateGlobalListOrder: boolean }): void {
        let fromAccount = null;
        let toAccount = null;

        if (allCategorizedAccountsMap.value[account.category]) {
            const accountList = allCategorizedAccountsMap.value[account.category]!.accounts;

            if (updateListOrder) {
                fromAccount = accountList[from];
                toAccount = accountList[to];
                accountList.splice(to, 0, accountList.splice(from, 1)[0] as Account);
            } else {
                fromAccount = accountList[to];

                if (from < to) {
                    toAccount = accountList[to - 1];
                } else if (from > to) {
                    toAccount = accountList[to + 1];
                }
            }
        }

        if (updateGlobalListOrder && fromAccount && toAccount) {
            let globalFromIndex = -1;
            let globalToIndex = -1;

            for (const [account, index] of itemAndIndex(allAccounts.value)) {
                if (account.id === fromAccount.id) {
                    globalFromIndex = index;
                } else if (account.id === toAccount.id) {
                    globalToIndex = index;
                }
            }

            if (globalFromIndex >= 0 && globalToIndex >= 0) {
                allAccounts.value.splice(globalToIndex, 0, allAccounts.value.splice(globalFromIndex, 1)[0] as Account);
            }
        }
    }

    function updateAccountVisibilityInAccountList({ account, hidden }: { account: Account, hidden: boolean }): void {
        if (allAccountsMap.value[account.id]) {
            allAccountsMap.value[account.id]!.visible = !hidden;
        }
    }

    function removeAccountFromAccountList(currentAccount: Account): void {
        for (const [account, index] of itemAndIndex(allAccounts.value)) {
            if (account.id === currentAccount.id) {
                allAccounts.value.splice(index, 1);
                break;
            }
        }

        if (allAccountsMap.value[currentAccount.id] && allAccountsMap.value[currentAccount.id]!.subAccounts) {
            const subAccounts = allAccountsMap.value[currentAccount.id]!.subAccounts as Account[];

            for (const subAccount of subAccounts) {
                if (allAccountsMap.value[subAccount.id]) {
                    delete allAccountsMap.value[subAccount.id];
                }
            }
        }

        if (allAccountsMap.value[currentAccount.id]) {
            delete allAccountsMap.value[currentAccount.id];
        }

        if (allCategorizedAccountsMap.value[currentAccount.category]) {
            const accountList = allCategorizedAccountsMap.value[currentAccount.category]!.accounts;

            for (const [account, index] of itemAndIndex(accountList)) {
                if (account.id === currentAccount.id) {
                    accountList.splice(index, 1);
                    break;
                }
            }
        }
    }

    function removeSubAccountFromAccountList(currentSubAccount: Account): void {
        for (const account of allAccounts.value) {
            if (account.type !== AccountType.MultiSubAccounts.type || !account.subAccounts) {
                continue;
            }

            const subAccounts = account.subAccounts as Account[];

            for (const [subAccount, index] of itemAndIndex(subAccounts)) {
                if (subAccount.id === currentSubAccount.id) {
                    subAccounts.splice(index, 1);
                    break;
                }
            }
        }

        if (allAccountsMap.value[currentSubAccount.id]) {
            delete allAccountsMap.value[currentSubAccount.id];
        }

        if (allCategorizedAccountsMap.value[currentSubAccount.category]) {
            const accountList = allCategorizedAccountsMap.value[currentSubAccount.category]!.accounts;

            for (const account of accountList) {
                if (account.type !== AccountType.MultiSubAccounts.type || !account.subAccounts) {
                    continue;
                }

                const subAccounts = account.subAccounts as Account[];

                for (const [subAccount, index] of itemAndIndex(subAccounts)) {
                    if (subAccount.id === currentSubAccount.id) {
                        subAccounts.splice(index, 1);
                        break;
                    }
                }
            }
        }
    }

    function updateAccountListInvalidState(invalidState: boolean): void {
        accountListStateInvalid.value = invalidState;
    }

    function resetAccounts(): void {
        loadingPromise.value = null;
        allAccounts.value = [];
        allAccountsMap.value = {};
        allCategorizedAccountsMap.value = {};
        accountListStateInvalid.value = true;
    }

    return {
        loadAccountList, addAccountToAccountList, updateAccountToAccountList, updateAccountDisplayOrderInAccountList,
        updateAccountVisibilityInAccountList, removeAccountFromAccountList, removeSubAccountFromAccountList,
        updateAccountListInvalidState, resetAccounts,
    };
}
