import { defineStore } from 'pinia';
import { useAccountBalances } from './account/balances.ts';
import { useAccountLoading } from './account/loading.ts';
import { useAccountLookup } from './account/lookup.ts';
import { useAccountManagement } from './account/management.ts';
import { useAccountMutation } from './account/mutation.ts';
import { useAccountSelection } from './account/selection.ts';
import { useAccountState } from './account/state.ts';





export const useAccountsStore = defineStore('accounts', () => {
    const {
        settingsStore, userStore, overviewStore, exchangeRatesStore, loadingPromise, allAccounts, allAccountsMap,
        allCategorizedAccountsMap, accountListStateInvalid, allPlainAccounts, allMixedPlainAccounts, allVisiblePlainAccounts,
        allAvailableAccountsCount, allVisibleAccountsCount, maxCategoryAccountCount,
    } = useAccountState();
    const {
        loadAccountList, addAccountToAccountList, updateAccountToAccountList, updateAccountDisplayOrderInAccountList,
        updateAccountVisibilityInAccountList, removeAccountFromAccountList, removeSubAccountFromAccountList,
        updateAccountListInvalidState, resetAccounts,
    } = useAccountMutation({ allAccounts, allAccountsMap, allCategorizedAccountsMap, settingsStore, accountListStateInvalid, loadingPromise });
    const { getFirstShowingIds, getLastShowingIds, getAccountStatementDate } = useAccountLookup({ allCategorizedAccountsMap, allAccountsMap });
    const { getNetAssets, getTotalAssets, getTotalLiabilities, getAccountCategoryTotalBalance, getAccountBalance, getAccountSubAccountBalance } = useAccountBalances({ allCategorizedAccountsMap, settingsStore, userStore, exchangeRatesStore });
    const { getSortedAccounts, hasAccount, hasVisibleSubAccount } = useAccountSelection({ allAccounts, userStore, exchangeRatesStore, settingsStore, allAccountsMap, allCategorizedAccountsMap });
    const { loadAllAccounts, getAccount, saveAccount } = useAccountLoading({
        loadingPromise, accountListStateInvalid, allAccounts, updateAccountListInvalidState, settingsStore, loadAccountList,
        allAccountsMap, addAccountToAccountList, updateAccountToAccountList, overviewStore,
    });
    const { updateAccountLastReconciledTime, changeAccountDisplayOrder, updateAccountDisplayOrders, hideAccount, deleteAccount, deleteSubAccount } = useAccountManagement({
        allAccountsMap, updateAccountListInvalidState, overviewStore, allCategorizedAccountsMap, accountListStateInvalid,
        updateAccountDisplayOrderInAccountList, loadAllAccounts, updateAccountVisibilityInAccountList,
        removeAccountFromAccountList, removeSubAccountFromAccountList,
    });

    return {
        // states
        allAccounts,
        allAccountsMap,
        allCategorizedAccountsMap,
        accountListStateInvalid,
        // computed states
        allPlainAccounts,
        allMixedPlainAccounts,
        allVisiblePlainAccounts,
        allAvailableAccountsCount,
        allVisibleAccountsCount,
        maxCategoryAccountCount,
        // functions
        updateAccountListInvalidState,
        resetAccounts,
        getFirstShowingIds,
        getLastShowingIds,
        getAccountStatementDate,
        getNetAssets,
        getTotalAssets,
        getTotalLiabilities,
        getAccountCategoryTotalBalance,
        getAccountBalance,
        getAccountSubAccountBalance,
        getSortedAccounts,
        hasAccount,
        hasVisibleSubAccount,
        loadAllAccounts,
        getAccount,
        saveAccount,
        updateAccountLastReconciledTime,
        changeAccountDisplayOrder,
        updateAccountDisplayOrders,
        hideAccount,
        deleteAccount,
        deleteSubAccount
    }
});
