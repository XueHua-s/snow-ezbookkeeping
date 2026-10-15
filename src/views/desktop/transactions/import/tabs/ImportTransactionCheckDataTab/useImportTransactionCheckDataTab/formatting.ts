import { TRANSACTION_MAX_COMMENT_LENGTH } from '@/consts/transaction.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    getTimezoneOffsetMinutes,
    getUtcOffsetByUtcOffsetMinutes,
    parseDateTimeFromUnixTimeWithTimezoneOffset
} from '@/lib/datetime.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { Account } from '@/models/account.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    allTagsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allTagsMap'];
    formatDateTimeToLongDateTime: ReturnType<typeof useImportTransactionCheckDataTabState>['formatDateTimeToLongDateTime'];
    tt: ReturnType<typeof useImportTransactionCheckDataTabState>['tt'];
    defaultCurrency: ReturnType<typeof useImportTransactionCheckDataTabState>['defaultCurrency'];
    allAccountsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allAccountsMap'];
    formatAmountToLocalizedNumeralsWithCurrency: ReturnType<typeof useImportTransactionCheckDataTabState>['formatAmountToLocalizedNumeralsWithCurrency'];
    exchangeRatesStore: ReturnType<typeof useImportTransactionCheckDataTabState>['exchangeRatesStore'];
    allAccounts: ReturnType<typeof useImportTransactionCheckDataTabState>['allAccounts'];
    formatNumberToLocalizedNumerals: ReturnType<typeof useImportTransactionCheckDataTabState>['formatNumberToLocalizedNumerals'];
}

export function useImportTransactionCheckDataTabFormatting(context: Context) {
    const {
        allTagsMap, formatDateTimeToLongDateTime, tt, defaultCurrency, allAccountsMap,
        formatAmountToLocalizedNumeralsWithCurrency, exchangeRatesStore, allAccounts, formatNumberToLocalizedNumerals,
    } = context;

    function isTagValid(tagIds: string[], tagIndex: number): boolean {
        if (!tagIds || !tagIds[tagIndex]) {
            return false;
        }

        if (tagIds[tagIndex] === '0') {
            return false;
        }

        const tagId = tagIds[tagIndex];
        return !!allTagsMap.value[tagId];
    }

    function getDisplayDateTime(transaction: ImportTransaction): string {
        const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transaction.utcOffset)
        return formatDateTimeToLongDateTime(dateTime);
    }

    function isSameAsDefaultTimezoneOffsetMinutes(transaction: ImportTransaction): boolean {
        return transaction.utcOffset === getTimezoneOffsetMinutes(transaction.time);
    }

    function getDisplayTimezone(transaction: ImportTransaction): string {
        return `UTC${getUtcOffsetByUtcOffsetMinutes(transaction.utcOffset)}`;
    }

    function getDisplayTransactionType(transaction: ImportTransaction): string {
        if (transaction.type === TransactionType.ModifyBalance) {
            return tt('Modify Balance');
        } else if (transaction.type === TransactionType.Income) {
            return tt('Income');
        } else if (transaction.type === TransactionType.Expense) {
            return tt('Expense');
        } else if (transaction.type === TransactionType.Transfer) {
            return tt('Transfer');
        } else {
            return tt('Unknown');
        }
    }

    function getTransactionSourceAccountCurrency(transaction: ImportTransaction): string {
        let currency = transaction.originalSourceAccountCurrency || defaultCurrency.value;

        if (transaction.sourceAccountId && transaction.sourceAccountId !== '0' && allAccountsMap.value[transaction.sourceAccountId]) {
            currency = allAccountsMap.value[transaction.sourceAccountId]!.currency;
        }

        return currency;
    }

    function getTransactionDestinationAccountCurrency(transaction: ImportTransaction): string {
        let currency = transaction.originalDestinationAccountCurrency || defaultCurrency.value;

        if (transaction.destinationAccountId && transaction.destinationAccountId !== '0' && allAccountsMap.value[transaction.destinationAccountId]) {
            currency = allAccountsMap.value[transaction.destinationAccountId]!.currency;
        }

        return currency;
    }

    function getTransactionDisplayAmount(transaction: ImportTransaction): string {
        const currency = getTransactionSourceAccountCurrency(transaction);
        return formatAmountToLocalizedNumeralsWithCurrency(parseBigDecimal(transaction.sourceAmount), currency);
    }

    function getTransactionDisplayDestinationAmount(transaction: ImportTransaction): string {
        if (transaction.type !== TransactionType.Transfer) {
            return '-';
        }

        const currency = getTransactionDestinationAccountCurrency(transaction);
        return formatAmountToLocalizedNumeralsWithCurrency(parseBigDecimal(transaction.destinationAmount), currency);
    }

    function getTransactionDisplaySourceAmountInDefaultCurrency(transaction: ImportTransaction): string {
        const currency = getTransactionSourceAccountCurrency(transaction);

        if (!currency || currency === defaultCurrency.value) {
            return getTransactionDisplayAmount(transaction);
        }

        const amount = exchangeRatesStore.getExchangedAmount(parseBigDecimal(transaction.sourceAmount), currency, defaultCurrency.value);
        return amount ? formatAmountToLocalizedNumeralsWithCurrency(amount.truncate(), defaultCurrency.value) : getTransactionDisplayAmount(transaction);
    }

    function getTransactionDisplayDestinationAmountInDefaultCurrency(transaction: ImportTransaction): string {
        const currency = getTransactionDestinationAccountCurrency(transaction);

        if (!currency || currency === defaultCurrency.value) {
            return getTransactionDisplayDestinationAmount(transaction);
        }

        const amount = exchangeRatesStore.getExchangedAmount(parseBigDecimal(transaction.destinationAmount), currency, defaultCurrency.value);
        return amount ? formatAmountToLocalizedNumeralsWithCurrency(amount.truncate(), defaultCurrency.value) : getTransactionDisplayDestinationAmount(transaction);
    }

    function getSourceAccountTitle(transaction: ImportTransaction): string {
        if (transaction.type === TransactionType.Expense || transaction.type === TransactionType.Income) {
            return tt('Account');
        } else if (transaction.type === TransactionType.Transfer) {
            return tt('Source Account');
        } else {
            return tt('Account');
        }
    }

    function getSourceAccountDisplayName(transaction: ImportTransaction): string {
        if (transaction.sourceAccountId) {
            return Account.findAccountNameById(allAccounts.value, transaction.sourceAccountId) || '';
        } else {
            return tt('None');
        }
    }

    function getDestinationAccountDisplayName(transaction: ImportTransaction): string {
        if (transaction.destinationAccountId) {
            return Account.findAccountNameById(allAccounts.value, transaction.destinationAccountId) || '';
        } else {
            return tt('None');
        }
    }

    function getTransactionDescriptionTooltip(transaction: ImportTransaction): string {
        if (transaction.comment && transaction.comment.length > TRANSACTION_MAX_COMMENT_LENGTH) {
            return tt('format.misc.charactersOverLimit', {
                count: formatNumberToLocalizedNumerals(transaction.comment.length - TRANSACTION_MAX_COMMENT_LENGTH)
            });
        } else {
            return '';
        }
    }

    return {
        isTagValid, getDisplayDateTime, isSameAsDefaultTimezoneOffsetMinutes, getDisplayTimezone, getDisplayTransactionType,
        getTransactionSourceAccountCurrency, getTransactionDestinationAccountCurrency, getTransactionDisplayAmount,
        getTransactionDisplayDestinationAmount, getTransactionDisplaySourceAmountInDefaultCurrency,
        getTransactionDisplayDestinationAmountInDefaultCurrency, getSourceAccountTitle, getSourceAccountDisplayName,
        getDestinationAccountDisplayName, getTransactionDescriptionTooltip,
    };
}
