import { TRANSACTION_MAX_AMOUNT, TRANSACTION_MIN_AMOUNT } from '@/consts/transaction.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    isNumber
} from '@/lib/common.ts';
import { getCurrencyFraction } from '@/lib/currency.ts';
import { getAmountWithDecimalNumberCount, parseBigDecimal } from '@/lib/numeral.ts';
import type { Transaction } from '@/models/transaction.ts';
import type { useTransactionState } from './state.ts';

interface Context {
    accountsStore: ReturnType<typeof useTransactionState>['accountsStore'];
    exchangeRatesStore: ReturnType<typeof useTransactionState>['exchangeRatesStore'];
}

// Owns exchange for the transaction store.
export function useTransactionExchange(context: Context) {
    const { accountsStore, exchangeRatesStore } = context;

    function setTransactionSuitableDestinationAmount(transaction: Transaction, oldSourceAmount: number, newSourceAmount: number, oldSourceAccountId?: string, oldDestinationAccountId?: string): void {
        if (transaction.type === TransactionType.Expense || transaction.type === TransactionType.Income) {
            transaction.destinationAmount = newSourceAmount;
        } else if (transaction.type === TransactionType.Transfer) {
            const sourceAccount = accountsStore.allAccountsMap[transaction.sourceAccountId];
            const destinationAccount = accountsStore.allAccountsMap[transaction.destinationAccountId];

            if (!sourceAccount || !destinationAccount) {
                return;
            }

            const oldSourceAccount = oldSourceAccountId ? accountsStore.allAccountsMap[oldSourceAccountId] : sourceAccount;
            const oldDestinationAccount = oldDestinationAccountId ? accountsStore.allAccountsMap[oldDestinationAccountId] : destinationAccount;

            let oldValueToCompare: BigDecimal = parseBigDecimal(oldSourceAmount);
            let newValueToSet: BigDecimal = parseBigDecimal(newSourceAmount);

            if (oldSourceAccount && oldDestinationAccount && oldSourceAccount.currency !== oldDestinationAccount.currency) {
                const decimalNumberCount = getCurrencyFraction(oldDestinationAccount.currency);
                const exchangedOldValue = exchangeRatesStore.getExchangedAmount(parseBigDecimal(oldSourceAmount), oldSourceAccount.currency, oldDestinationAccount.currency);

                if (isNumber(decimalNumberCount) && exchangedOldValue) {
                    oldValueToCompare = exchangedOldValue.truncate();
                    oldValueToCompare = getAmountWithDecimalNumberCount(oldValueToCompare, decimalNumberCount);
                }
            }

            if (sourceAccount.currency !== destinationAccount.currency) {
                const decimalNumberCount = getCurrencyFraction(destinationAccount.currency);
                const exchangedNewValue = exchangeRatesStore.getExchangedAmount(parseBigDecimal(newSourceAmount), sourceAccount.currency, destinationAccount.currency);

                if (isNumber(decimalNumberCount) && exchangedNewValue) {
                    newValueToSet = exchangedNewValue.truncate();
                    newValueToSet = getAmountWithDecimalNumberCount(newValueToSet, decimalNumberCount);
                } else {
                    return;
                }
            }

            if ((oldValueToCompare.equals(transaction.destinationAmount) || transaction.destinationAmount === 0) &&
                newValueToSet.between(TRANSACTION_MIN_AMOUNT, TRANSACTION_MAX_AMOUNT)) {
                transaction.destinationAmount = newValueToSet.toSafeIntegerNumber();
            }
        }
    }

    return { setTransactionSuitableDestinationAmount };
}
