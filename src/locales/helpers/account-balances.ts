import { DISPLAY_HIDDEN_AMOUNT, INCOMPLETE_AMOUNT_SUFFIX } from '@/consts/numeral.ts';
import {
    AccountCategory
} from '@/core/account.ts';
import {
    type BigDecimal
} from '@/core/numeral.ts';
import {
    getAllFilteredAccountsBalance,
    getCategorizedAccountsMap
} from '@/lib/account.ts';
import {
    BIG_DECIMAL_ZERO,
    parseBigDecimal
} from '@/lib/numeral.ts';
import {
    type CategorizedAccount,
    Account,
    AccountWithDisplayBalance,
    CategorizedAccountWithDisplayBalance
} from '@/models/account.ts';
import type { useHelpersAmountFormatting } from './amount-formatting.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    userStore: ReturnType<typeof useHelpersState>['userStore'];
    getFormattedAmountWithCurrency: ReturnType<typeof useHelpersAmountFormatting>['getFormattedAmountWithCurrency'];
    exchangeRatesStore: ReturnType<typeof useHelpersState>['exchangeRatesStore'];
}

// Owns account balances for the helpers store.
export function useHelpersAccountBalances(context: Context) {
    const { userStore, getFormattedAmountWithCurrency, exchangeRatesStore } = context;

    function getCategorizedAccountsWithDisplayBalance(allVisibleAccounts: Account[], showAccountBalance: boolean, customAccountCategoryOrder: string): CategorizedAccountWithDisplayBalance[] {
        const ret: CategorizedAccountWithDisplayBalance[] = [];
        const defaultCurrency = userStore.currentUserDefaultCurrency;
        const allCategories = AccountCategory.values(customAccountCategoryOrder);
        const categorizedAccounts: Record<number, CategorizedAccount> = getCategorizedAccountsMap(Account.cloneAccounts(allVisibleAccounts));

        for (const category of allCategories) {
            const accountCategory = categorizedAccounts[category.type];

            if (!accountCategory) {
                continue;
            }

            const accountsWithDisplayBalance: AccountWithDisplayBalance[] = [];

            if (accountCategory.accounts) {
                for (const account of accountCategory.accounts) {
                    let accountWithDisplaceBalance: AccountWithDisplayBalance;

                    if (showAccountBalance && account.isAsset) {
                        accountWithDisplaceBalance = AccountWithDisplayBalance.fromAccount(account, getFormattedAmountWithCurrency(parseBigDecimal(account.balance), account.currency));
                    } else if (showAccountBalance && account.isLiability) {
                        accountWithDisplaceBalance = AccountWithDisplayBalance.fromAccount(account, getFormattedAmountWithCurrency(parseBigDecimal(account.balance).negate(), account.currency));
                    } else {
                        accountWithDisplaceBalance = AccountWithDisplayBalance.fromAccount(account, DISPLAY_HIDDEN_AMOUNT);
                    }

                    accountsWithDisplayBalance.push(accountWithDisplaceBalance);
                }
            }

            let finalTotalBalance = '';

            if (showAccountBalance) {
                const accountsBalance = getAllFilteredAccountsBalance(categorizedAccounts, customAccountCategoryOrder,
                        account => account.category === accountCategory.category);
                let totalBalance: BigDecimal = BIG_DECIMAL_ZERO;
                let hasUnCalculatedAmount = false;

                for (const accountBalance of accountsBalance) {
                    if (accountBalance.currency === defaultCurrency) {
                        if (accountBalance.isAsset) {
                            totalBalance = totalBalance.add(accountBalance.balance);
                        } else if (accountBalance.isLiability) {
                            totalBalance = totalBalance.subtract(accountBalance.balance);
                        }
                    } else {
                        const balance = exchangeRatesStore.getExchangedAmount(accountBalance.balance, accountBalance.currency, defaultCurrency);

                        if (!balance) {
                            hasUnCalculatedAmount = true;
                            continue;
                        }

                        if (accountBalance.isAsset) {
                            totalBalance = totalBalance.add(balance.truncate());
                        } else if (accountBalance.isLiability) {
                            totalBalance = totalBalance.subtract(balance.truncate());
                        }
                    }
                }

                finalTotalBalance = getFormattedAmountWithCurrency(totalBalance, defaultCurrency);

                if (hasUnCalculatedAmount) {
                    finalTotalBalance = finalTotalBalance + INCOMPLETE_AMOUNT_SUFFIX;
                }
            } else {
                finalTotalBalance = DISPLAY_HIDDEN_AMOUNT;
            }

            const accountCategoryWithDisplayBalance = CategorizedAccountWithDisplayBalance.of(accountCategory, accountsWithDisplayBalance, finalTotalBalance);
            ret.push(accountCategoryWithDisplayBalance);
        }

        return ret;
    }

    return { getCategorizedAccountsWithDisplayBalance };
}
