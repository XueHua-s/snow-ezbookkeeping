import { ACCOUNT_CURRENCY_NOT_SET_VALUE } from '@/consts/currency.ts';
import { DISPLAY_HIDDEN_AMOUNT, INCOMPLETE_AMOUNT_SUFFIX } from '@/consts/numeral.ts';
import { AccountCategory, AccountType } from '@/core/account.ts';
import type { BigDecimal, BigDecimalWithSuffix, HiddenAmount } from '@/core/numeral.ts';
import { getAllFilteredAccountsBalance } from '@/lib/account.ts';
import { arrayItemToObjectField, isArray, isDefined } from '@/lib/common.ts';
import { BIG_DECIMAL_ZERO, isBigDecimal, parseBigDecimal } from '@/lib/numeral.ts';
import type { AccountDisplayBalance, Account } from '@/models/account.ts';
import type { useAccountState } from './state.ts';

interface Context {
    allCategorizedAccountsMap: ReturnType<typeof useAccountState>['allCategorizedAccountsMap'];
    settingsStore: ReturnType<typeof useAccountState>['settingsStore'];
    userStore: ReturnType<typeof useAccountState>['userStore'];
    exchangeRatesStore: ReturnType<typeof useAccountState>['exchangeRatesStore'];
}

// Owns balances for the account store.
export function useAccountBalances(context: Context) {
    const { allCategorizedAccountsMap, settingsStore, userStore, exchangeRatesStore } = context;

    function getNetAssets(showAccountBalance: boolean, excludedAccountIds: Record<string, boolean>): BigDecimal | HiddenAmount | BigDecimalWithSuffix {
        if (!showAccountBalance) {
            return DISPLAY_HIDDEN_AMOUNT;
        }

        const accountsBalance = getAllFilteredAccountsBalance(allCategorizedAccountsMap.value, settingsStore.appSettings.accountCategoryOrders,
            account => !(account.type === AccountType.SingleAccount.type && excludedAccountIds[account.id])
        );
        let netAssets: BigDecimal = BIG_DECIMAL_ZERO;
        let hasUnCalculatedAmount = false;

        for (const accountBalance of accountsBalance) {
            if (accountBalance.currency === userStore.currentUserDefaultCurrency) {
                netAssets = netAssets.add(accountBalance.balance);
            } else {
                const balance = exchangeRatesStore.getExchangedAmount(accountBalance.balance, accountBalance.currency, userStore.currentUserDefaultCurrency);

                if (!balance) {
                    hasUnCalculatedAmount = true;
                    continue;
                }

                netAssets = netAssets.add(balance.truncate());
            }
        }

        if (hasUnCalculatedAmount) {
            return {
                value: netAssets,
                suffix: INCOMPLETE_AMOUNT_SUFFIX
            };
        } else {
            return netAssets;
        }
    }

    function getTotalAssets(showAccountBalance: boolean, excludedAccountIds: Record<string, boolean>): BigDecimal | HiddenAmount | BigDecimalWithSuffix {
        if (!showAccountBalance) {
            return DISPLAY_HIDDEN_AMOUNT;
        }

        const accountsBalance = getAllFilteredAccountsBalance(allCategorizedAccountsMap.value, settingsStore.appSettings.accountCategoryOrders,
            account => (account.isAsset || false) && !(account.type === AccountType.SingleAccount.type && excludedAccountIds[account.id])
        );
        let totalAssets: BigDecimal = BIG_DECIMAL_ZERO;
        let hasUnCalculatedAmount = false;

        for (const accountBalance of accountsBalance) {
            if (accountBalance.currency === userStore.currentUserDefaultCurrency) {
                totalAssets = totalAssets.add(accountBalance.balance);
            } else {
                const balance = exchangeRatesStore.getExchangedAmount(accountBalance.balance, accountBalance.currency, userStore.currentUserDefaultCurrency);

                if (!balance) {
                    hasUnCalculatedAmount = true;
                    continue;
                }

                totalAssets = totalAssets.add(balance.truncate());
            }
        }

        if (hasUnCalculatedAmount) {
            return {
                value: totalAssets,
                suffix: INCOMPLETE_AMOUNT_SUFFIX
            };
        } else {
            return totalAssets;
        }
    }

    function getTotalLiabilities(showAccountBalance: boolean, excludedAccountIds: Record<string, boolean>): BigDecimal | HiddenAmount | BigDecimalWithSuffix {
        if (!showAccountBalance) {
            return DISPLAY_HIDDEN_AMOUNT;
        }

        const accountsBalance = getAllFilteredAccountsBalance(allCategorizedAccountsMap.value, settingsStore.appSettings.accountCategoryOrders,
            account => (account.isLiability || false) && !(account.type === AccountType.SingleAccount.type && excludedAccountIds[account.id])
        );
        let totalLiabilities: BigDecimal = BIG_DECIMAL_ZERO;
        let hasUnCalculatedAmount = false;

        for (const accountBalance of accountsBalance) {
            if (accountBalance.currency === userStore.currentUserDefaultCurrency) {
                totalLiabilities = totalLiabilities.subtract(accountBalance.balance);
            } else {
                const balance = exchangeRatesStore.getExchangedAmount(accountBalance.balance, accountBalance.currency, userStore.currentUserDefaultCurrency);

                if (!balance) {
                    hasUnCalculatedAmount = true;
                    continue;
                }

                totalLiabilities = totalLiabilities.subtract(balance.truncate());
            }
        }

        if (hasUnCalculatedAmount) {
            return {
                value: totalLiabilities,
                suffix: INCOMPLETE_AMOUNT_SUFFIX
            };
        } else {
            return totalLiabilities;
        }
    }

    function getAccountCategoryTotalBalance(showAccountBalance: boolean, accountCategory: AccountCategory, showAvailableCreditForCreditCard: boolean): BigDecimal | HiddenAmount | BigDecimalWithSuffix | undefined {
        if (!showAccountBalance) {
            return DISPLAY_HIDDEN_AMOUNT;
        }

        const accountsBalance = getAllFilteredAccountsBalance(allCategorizedAccountsMap.value, settingsStore.appSettings.accountCategoryOrders,
            account => account.category === accountCategory.type);
        let totalBalance: BigDecimal = BIG_DECIMAL_ZERO;
        let totalCreditCardAvailableCredit: BigDecimal = BIG_DECIMAL_ZERO;
        let hasUnCalculatedAmount = false;
        let hasUnCalculatedAvailableCredit = false;
        let hasCalculatedAvailableCredit = false;

        for (const accountBalance of accountsBalance) {
            let balance: BigDecimal | null = null;

            if (accountBalance.currency === userStore.currentUserDefaultCurrency) {
                balance = accountBalance.balance;
            } else {
                balance = exchangeRatesStore.getExchangedAmount(accountBalance.balance, accountBalance.currency, userStore.currentUserDefaultCurrency);

                if (!balance) {
                    hasUnCalculatedAmount = true;
                    continue;
                }
            }

            if (accountBalance.isAsset) {
                totalBalance = totalBalance.add(balance);
            } else if (accountBalance.isLiability) {
                totalBalance = totalBalance.subtract(balance);
            } else {
                totalBalance = totalBalance.add(balance);
            }

            if (accountBalance.category === AccountCategory.CreditCard.type && showAvailableCreditForCreditCard) {
                if (isDefined(accountBalance.creditCardLimit) && isBigDecimal(accountBalance.creditCardLimit.amount) && accountBalance.creditCardLimit.shareByCount > 0) {
                    const amount = accountBalance.creditCardLimit.amount.divide(accountBalance.creditCardLimit.shareByCount);

                    if (accountBalance.creditCardLimit.currency === userStore.currentUserDefaultCurrency) {
                        totalCreditCardAvailableCredit = totalCreditCardAvailableCredit.add(amount).add(balance);
                        hasCalculatedAvailableCredit = true;
                    } else {
                        const limit = exchangeRatesStore.getExchangedAmount(amount, accountBalance.creditCardLimit.currency, userStore.currentUserDefaultCurrency);

                        if (limit) {
                            totalCreditCardAvailableCredit = totalCreditCardAvailableCredit.add(limit).add(balance);
                            hasCalculatedAvailableCredit = true;
                        } else {
                            hasUnCalculatedAvailableCredit = true;
                        }
                    }
                } else {
                    hasUnCalculatedAvailableCredit = true;
                }
            }
        }

        if (accountCategory.type === AccountCategory.CreditCard.type && showAvailableCreditForCreditCard) {
            if (hasUnCalculatedAmount || !hasCalculatedAvailableCredit) {
                return undefined;
            } else if (hasUnCalculatedAvailableCredit) {
                return {
                    value: totalCreditCardAvailableCredit,
                    suffix: INCOMPLETE_AMOUNT_SUFFIX
                };
            } else {
                return totalCreditCardAvailableCredit;
            }
        } else {
            if (hasUnCalculatedAmount) {
                return {
                    value: totalBalance,
                    suffix: INCOMPLETE_AMOUNT_SUFFIX
                };
            } else {
                return totalBalance;
            }
        }
    }

    function getAccountBalance(showAccountBalance: boolean, account: Account): BigDecimal | HiddenAmount | null {
        if (account.type !== AccountType.SingleAccount.type) {
            return null;
        }

        if (showAccountBalance) {
            if (account.isAsset) {
                return parseBigDecimal(account.balance);
            } else if (account.isLiability) {
                return parseBigDecimal(account.balance).negate();
            } else {
                return parseBigDecimal(account.balance);
            }
        } else {
            return DISPLAY_HIDDEN_AMOUNT;
        }
    }

    function getAccountSubAccountBalance(showAccountBalance: boolean, showHidden: boolean, account: Account, subAccountId?: string, onlyShowSelectedAccountIds?: string[]): AccountDisplayBalance | null {
        if (account.type !== AccountType.MultiSubAccounts.type) {
            return null;
        }

        let resultCurrency = userStore.currentUserDefaultCurrency;
        let parentHasSetCurrency = false;

        if (account.category === AccountCategory.CreditCard.type && account.currency && account.currency !== ACCOUNT_CURRENCY_NOT_SET_VALUE) {
            resultCurrency = account.currency;
            parentHasSetCurrency = true;
        }

        if (!account.subAccounts || !account.subAccounts.length) {
            return {
                balance: showAccountBalance ? BIG_DECIMAL_ZERO : DISPLAY_HIDDEN_AMOUNT,
                currency: resultCurrency
            };
        }

        const selectedAccountIds: Record<string, boolean> = isArray(onlyShowSelectedAccountIds) ? arrayItemToObjectField(onlyShowSelectedAccountIds, true) : {};
        const allSubAccountCurrenciesMap: Record<string, boolean> = {};
        const allSubAccountCurrencies: string[] = [];
        let totalBalance: BigDecimal = BIG_DECIMAL_ZERO;

        for (const subAccount of account.subAccounts) {
            if (!showHidden && subAccount.hidden) {
                continue;
            }

            if (!allSubAccountCurrenciesMap[subAccount.currency]) {
                allSubAccountCurrenciesMap[subAccount.currency] = true;
                allSubAccountCurrencies.push(subAccount.currency);
            }
        }

        if (allSubAccountCurrencies.length === 0) {
            return {
                balance: showAccountBalance ? BIG_DECIMAL_ZERO : DISPLAY_HIDDEN_AMOUNT,
                currency: resultCurrency
            };
        }

        if (allSubAccountCurrencies.length === 1 && !parentHasSetCurrency) {
            resultCurrency = allSubAccountCurrencies[0] as string;
        }

        let hasUnCalculatedAmount = false;

        for (const subAccount of account.subAccounts) {
            if (!showHidden && subAccount.hidden) {
                continue;
            }

            if (subAccountId) {
                if (subAccountId === subAccount.id) {
                    return {
                        balance: getAccountBalance(showAccountBalance, subAccount)!,
                        currency: subAccount.currency
                    };
                }
            }

            if (onlyShowSelectedAccountIds && onlyShowSelectedAccountIds.length > 0 && !selectedAccountIds[subAccount.id]) {
                continue;
            }

            if (subAccount.currency === resultCurrency) {
                if (subAccount.isAsset) {
                    totalBalance = totalBalance.add(parseBigDecimal(subAccount.balance));
                } else if (subAccount.isLiability) {
                    totalBalance = totalBalance.subtract(parseBigDecimal(subAccount.balance));
                } else {
                    totalBalance = totalBalance.add(parseBigDecimal(subAccount.balance));
                }
            } else {
                const balance = exchangeRatesStore.getExchangedAmount(parseBigDecimal(subAccount.balance), subAccount.currency, resultCurrency);

                if (!balance) {
                    hasUnCalculatedAmount = true;
                    continue;
                }

                if (subAccount.isAsset) {
                    totalBalance = totalBalance.add(balance);
                } else if (subAccount.isLiability) {
                    totalBalance = totalBalance.subtract(balance);
                } else {
                    totalBalance = totalBalance.add(balance);
                }
            }
        }

        if (subAccountId) { // not found specified id in sub accounts
            return null;
        }

        const displayTotalBalance: BigDecimalWithSuffix = {
            value: totalBalance,
            suffix: hasUnCalculatedAmount ? INCOMPLETE_AMOUNT_SUFFIX : ''
        };

        return {
            balance: showAccountBalance ? displayTotalBalance : DISPLAY_HIDDEN_AMOUNT,
            currency: resultCurrency
        };
    }

    return { getNetAssets, getTotalAssets, getTotalLiabilities, getAccountCategoryTotalBalance, getAccountBalance, getAccountSubAccountBalance };
}
