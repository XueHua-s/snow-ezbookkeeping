import { type AccountCategory, AccountType } from '@/core/account.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { arrayItemToObjectField, isArray } from '@/lib/common.ts';
import { BIG_DECIMAL_ZERO, parseBigDecimal } from '@/lib/numeral.ts';
import type { Account } from '@/models/account.ts';
import type { useAccountState } from './state.ts';

interface Context {
    allAccounts: ReturnType<typeof useAccountState>['allAccounts'];
    userStore: ReturnType<typeof useAccountState>['userStore'];
    exchangeRatesStore: ReturnType<typeof useAccountState>['exchangeRatesStore'];
    settingsStore: ReturnType<typeof useAccountState>['settingsStore'];
    allAccountsMap: ReturnType<typeof useAccountState>['allAccountsMap'];
    allCategorizedAccountsMap: ReturnType<typeof useAccountState>['allCategorizedAccountsMap'];
}

// Owns selection for the account store.
export function useAccountSelection(context: Context) {
    const { allAccounts, userStore, exchangeRatesStore, settingsStore, allAccountsMap, allCategorizedAccountsMap } = context;

    function getSortedAccounts(accountIds: string[] | undefined, sortBy: 'displayOrder' | 'balance' | string, count: number): Account[] {
        const selectedAccountIds: Record<string, boolean> = isArray(accountIds) ? arrayItemToObjectField(accountIds, true) : {};
        const accounts: Account[] = allAccounts.value.filter(account => {
            if (account.hidden) {
                return false;
            }

            if (account.type === AccountType.MultiSubAccounts.type && account.subAccounts) {
                for (const subAccount of account.subAccounts) {
                    if (!subAccount.hidden && selectedAccountIds[subAccount.id]) {
                        return true;
                    }
                }
            }

            return !accountIds || accountIds.length < 1 || !!selectedAccountIds[account.id];
        });

        if (sortBy === 'balance') {
            const accountBalances: Record<string, BigDecimal> = {};

            for (const account of accounts) {
                let totalBalance: BigDecimal = BIG_DECIMAL_ZERO;

                if (account.type === AccountType.SingleAccount.type) {
                    totalBalance = parseBigDecimal(account.balance);

                    if (account.currency !== userStore.currentUserDefaultCurrency) {
                        const exchangedBalance = exchangeRatesStore.getExchangedAmount(totalBalance, account.currency, userStore.currentUserDefaultCurrency);

                        if (exchangedBalance) {
                            totalBalance = exchangedBalance.truncate();
                        }
                    }
                } else if (account.type === AccountType.MultiSubAccounts.type && account.subAccounts) {
                    for (const subAccount of account.subAccounts) {
                        if (!selectedAccountIds[subAccount.id]) {
                            continue;
                        }

                        let subAccountBalance = parseBigDecimal(subAccount.balance);

                        if (subAccount.currency !== userStore.currentUserDefaultCurrency) {
                            const exchangedBalance = exchangeRatesStore.getExchangedAmount(subAccountBalance, subAccount.currency, userStore.currentUserDefaultCurrency);

                            if (exchangedBalance) {
                                subAccountBalance = exchangedBalance.truncate();
                            }
                        }

                        totalBalance = totalBalance.add(subAccountBalance);
                    }
                }

                accountBalances[account.id] = totalBalance;
            }

            accounts.sort((a, b) => {
                const balanceA = accountBalances[a.id] ?? BIG_DECIMAL_ZERO;
                const balanceB = accountBalances[b.id] ?? BIG_DECIMAL_ZERO;

                if (balanceA === balanceB) {
                    return a.compareTo(b, settingsStore.accountCategoryDisplayOrders, allAccountsMap.value);
                }

                return balanceB.compareTo(balanceA);
            });
        }

        return accounts.slice(0, count);
    }

    function hasAccount(accountCategory: AccountCategory, visibleOnly: boolean): boolean {
        const categorizedAccounts = allCategorizedAccountsMap.value[accountCategory.type];

        if (!categorizedAccounts || !categorizedAccounts.accounts || !categorizedAccounts.accounts.length) {
            return false;
        }

        let shownCount = 0;

        for (const account of categorizedAccounts.accounts) {
            if (!visibleOnly || !account.hidden) {
                shownCount++;
            }
        }

        return shownCount > 0;
    }

    function hasVisibleSubAccount(showHidden: boolean, account: Account): boolean {
        if (!account || account.type !== AccountType.MultiSubAccounts.type || !account.subAccounts) {
            return false;
        }

        for (const subAccount of account.subAccounts) {
            if (showHidden || !subAccount.hidden) {
                return true;
            }
        }

        return false;
    }

    return { getSortedAccounts, hasAccount, hasVisibleSubAccount };
}
