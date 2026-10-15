import { AccountType } from '@/core/account.ts';
import { values } from '@/core/base.ts';
import {
    type CategorizedAccount,
    Account
} from '@/models/account.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useOverviewStore } from '@/stores/overview.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import type { ComputedRef, Ref, ShallowRef, UnwrapRef } from 'vue';
import { computed, ref, shallowRef } from 'vue';
interface AccountState {
    settingsStore: ReturnType<typeof useSettingsStore>;
    userStore: ReturnType<typeof useUserStore>;
    overviewStore: ReturnType<typeof useOverviewStore>;
    exchangeRatesStore: ReturnType<typeof useExchangeRatesStore>;
    loadingPromise: ShallowRef<Promise<Account[]> | null>;
    allAccounts: Ref<UnwrapRef<Account[]>>;
    allAccountsMap: Ref<UnwrapRef<Record<string, Account>>>;
    allCategorizedAccountsMap: Ref<UnwrapRef<Record<number, CategorizedAccount>>>;
    accountListStateInvalid: Ref<UnwrapRef<boolean>>;
    allPlainAccounts: ComputedRef<Account[]>;
    allMixedPlainAccounts: ComputedRef<Account[]>;
    allVisiblePlainAccounts: ComputedRef<Account[]>;
    allAvailableAccountsCount: ComputedRef<number>;
    allVisibleAccountsCount: ComputedRef<number>;
    maxCategoryAccountCount: ComputedRef<number>;
}



// Owns state for the account store.
export function useAccountState(): AccountState  {
    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const overviewStore = useOverviewStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const loadingPromise = shallowRef<Promise<Account[]> | null>(null);

    const allAccounts = ref<Account[]>([]);

    const allAccountsMap = ref<Record<string, Account>>({});

    const allCategorizedAccountsMap = ref<Record<number, CategorizedAccount>>({});

    const accountListStateInvalid = ref<boolean>(true);

    const allPlainAccounts = computed<Account[]>(() => {
        const allAccountsList: Account[] = [];

        for (const account of allAccounts.value) {
            if (account.type === AccountType.SingleAccount.type) {
                allAccountsList.push(account);
            } else if (account.type === AccountType.MultiSubAccounts.type) {
                if (account.subAccounts) {
                    for (const subAccount of account.subAccounts) {
                        allAccountsList.push(subAccount);
                    }
                }
            }
        }

        return Account.sortAccounts(allAccountsList, settingsStore.accountCategoryDisplayOrders, allAccountsMap.value);
    });

    const allMixedPlainAccounts = computed<Account[]>(() => {
        const allAccountsList: Account[] = [];

        for (const account of allAccounts.value) {
            if (account.type === AccountType.SingleAccount.type) {
                allAccountsList.push(account);
            } else if (account.type === AccountType.MultiSubAccounts.type) {
                allAccountsList.push(account);

                if (account.subAccounts) {
                    for (const subAccount of account.subAccounts) {
                        allAccountsList.push(subAccount);
                    }
                }
            }
        }

        return Account.sortAccounts(allAccountsList, settingsStore.accountCategoryDisplayOrders, allAccountsMap.value);
    });

    const allVisiblePlainAccounts = computed<Account[]>(() => {
        const allVisibleAccounts: Account[] = [];

        for (const account of allAccounts.value) {
            if (account.hidden) {
                continue;
            }

            if (account.type === AccountType.SingleAccount.type) {
                allVisibleAccounts.push(account);
            } else if (account.type === AccountType.MultiSubAccounts.type) {
                if (account.subAccounts) {
                    for (const subAccount of account.subAccounts) {

                        if (subAccount.hidden) {
                            continue;
                        }

                        allVisibleAccounts.push(subAccount);
                    }
                }
            }
        }

        return Account.sortAccounts(allVisibleAccounts, settingsStore.accountCategoryDisplayOrders, allAccountsMap.value);
    });

    const allAvailableAccountsCount = computed<number>(() => {
        let allAccountCount = 0;

        for (const categorizedAccounts of values(allCategorizedAccountsMap.value)) {
            allAccountCount += categorizedAccounts.accounts.length;
        }

        return allAccountCount;
    });

    const allVisibleAccountsCount = computed<number>(() => {
        let shownAccountCount = 0;

        for (const categorizedAccounts of values(allCategorizedAccountsMap.value)) {
            const accountList = categorizedAccounts.accounts;

            for (const account of accountList) {
                if (!account.hidden) {
                    shownAccountCount++;
                }
            }
        }

        return shownAccountCount;
    });

    const maxCategoryAccountCount = computed<number>(() => {
        let maxCount = 0;

        for (const categorizedAccounts of values(allCategorizedAccountsMap.value)) {
            const accountCount = categorizedAccounts.accounts.length;

            if (accountCount > maxCount) {
                maxCount = accountCount;
            }
        }

        return maxCount;
    });

    return {
        settingsStore, userStore, overviewStore, exchangeRatesStore, loadingPromise, allAccounts, allAccountsMap,
        allCategorizedAccountsMap, accountListStateInvalid, allPlainAccounts, allMixedPlainAccounts, allVisiblePlainAccounts,
        allAvailableAccountsCount, allVisibleAccountsCount, maxCategoryAccountCount,
    };
}
