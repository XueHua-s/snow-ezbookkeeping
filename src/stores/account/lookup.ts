import { AccountCategory, AccountType } from '@/core/account.ts';
import { entries, reversed } from '@/core/base.ts';
import type { AccountShowingIds } from '@/models/account.ts';
import type { useAccountState } from './state.ts';

interface Context {
    allCategorizedAccountsMap: ReturnType<typeof useAccountState>['allCategorizedAccountsMap'];
    allAccountsMap: ReturnType<typeof useAccountState>['allAccountsMap'];
}

// Owns lookup for the account store.
export function useAccountLookup(context: Context) {
    const { allCategorizedAccountsMap, allAccountsMap } = context;

    function getFirstShowingIds(showHidden: boolean): AccountShowingIds {
        const ret: AccountShowingIds = {
            accounts: {},
            subAccounts: {}
        };

        for (const [category, categorizedAccounts] of entries(allCategorizedAccountsMap.value)) {
            if (!categorizedAccounts || !categorizedAccounts.accounts) {
                continue;
            }

            const accounts = categorizedAccounts.accounts;

            for (const account of accounts) {
                if (account.type === AccountType.MultiSubAccounts.type && account.subAccounts) {
                    for (const subAccount of account.subAccounts) {
                        if (showHidden || !subAccount.hidden) {
                            ret.subAccounts[account.id] = subAccount.id;
                            break;
                        }
                    }
                }

                if (showHidden || !account.hidden) {
                    ret.accounts[parseInt(category)] = account.id;
                    break;
                }
            }
        }

        return ret;
    }

    function getLastShowingIds(showHidden: boolean): AccountShowingIds {
        const ret: AccountShowingIds = {
            accounts: {},
            subAccounts: {}
        };

        for (const [category, categorizedAccounts] of entries(allCategorizedAccountsMap.value)) {
            if (!categorizedAccounts || !categorizedAccounts.accounts) {
                continue;
            }

            const accounts = categorizedAccounts.accounts;

            for (const account of reversed(accounts)) {
                if (account.type === AccountType.MultiSubAccounts.type && account.subAccounts) {
                    for (const subAccount of reversed(account.subAccounts)) {
                        if (showHidden || !subAccount.hidden) {
                            ret.subAccounts[account.id] = subAccount.id;
                            break;
                        }
                    }
                }

                if (showHidden || !account.hidden) {
                    ret.accounts[parseInt(category)] = account.id;
                    break;
                }
            }
        }

        return ret;
    }

    function getAccountStatementDate(accountId?: string): number | undefined | null {
        if (!accountId) {
            return null;
        }

        const accountIds = accountId.split(',');
        let mainAccount = null;

        for (const accountId of accountIds) {
            let account = allAccountsMap.value[accountId];

            if (!account) {
                return null;
            }

            if (account.parentId !== '0') {
                account = allAccountsMap.value[account.parentId];
            }

            if (!account) {
                return null;
            }

            if (mainAccount) {
                if (mainAccount.id !== account.id) {
                    return null;
                } else {
                    continue;
                }
            }

            mainAccount = account;
        }

        if (!mainAccount) {
            return null;
        }

        if (mainAccount.category === AccountCategory.CreditCard.type) {
            return mainAccount.creditCardStatementDate;
        }

        return null;
    }

    return { getFirstShowingIds, getLastShowingIds, getAccountStatementDate };
}
