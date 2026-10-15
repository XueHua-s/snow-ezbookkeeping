import { DISPLAY_HIDDEN_AMOUNT, INCOMPLETE_AMOUNT_SUFFIX } from '@/consts/numeral.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TransactionAmountType, TransactionType } from '@/core/transaction.ts';
import {
    getTimezoneOffsetMinutes,
    getUtcOffsetByUtcOffsetMinutes,
    getYearMonthFirstUnixTime,
    parseDateTimeFromUnixTime,
    parseDateTimeFromUnixTimeWithTimezoneOffset
} from '@/lib/datetime.ts';
import {
    parseBigDecimal
} from '@/lib/numeral.ts';
import { type Transaction } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionPictureInfoBasicResponse } from '@/models/transaction_picture_info.ts';
import { TransactionTagGroup } from '@/models/transaction_tag_group.ts';
import { type TransactionMonthList, type TransactionTotalAmount } from '@/stores/transaction.ts';
import type { useTransactionListPageBaseState } from './state.ts';
import { type TransactionListDisplayTotalAmount } from './types.ts';

interface Context {
    queryAllFilterCategoryIds: ReturnType<typeof useTransactionListPageBaseState>['queryAllFilterCategoryIds'];
    allTransactionTagsByGroup: ReturnType<typeof useTransactionListPageBaseState>['allTransactionTagsByGroup'];
    queryAllFilterTagIds: ReturnType<typeof useTransactionListPageBaseState>['queryAllFilterTagIds'];
    formatAmountToLocalizedNumeralsWithCurrency: ReturnType<typeof useTransactionListPageBaseState>['formatAmountToLocalizedNumeralsWithCurrency'];
    userDefaultCurrency: ReturnType<typeof useTransactionListPageBaseState>['userDefaultCurrency'];
    exchangeRatesStore: ReturnType<typeof useTransactionListPageBaseState>['exchangeRatesStore'];
    totalAmountTypeInTransactionListPage: ReturnType<typeof useTransactionListPageBaseState>['totalAmountTypeInTransactionListPage'];
    selectedAccountDefaultCurrency: ReturnType<typeof useTransactionListPageBaseState>['selectedAccountDefaultCurrency'];
    formatDateTimeToShortTime: ReturnType<typeof useTransactionListPageBaseState>['formatDateTimeToShortTime'];
    formatDateTimeToLongDate: ReturnType<typeof useTransactionListPageBaseState>['formatDateTimeToLongDate'];
    formatDateTimeToGregorianLikeLongYearMonth: ReturnType<typeof useTransactionListPageBaseState>['formatDateTimeToGregorianLikeLongYearMonth'];
    numeralSystem: ReturnType<typeof useTransactionListPageBaseState>['numeralSystem'];
    formatDateTimeToLongDateTime: ReturnType<typeof useTransactionListPageBaseState>['formatDateTimeToLongDateTime'];
    queryAllFilterAccountIdsCount: ReturnType<typeof useTransactionListPageBaseState>['queryAllFilterAccountIdsCount'];
    queryAllFilterAccountIds: ReturnType<typeof useTransactionListPageBaseState>['queryAllFilterAccountIds'];
    tt: ReturnType<typeof useTransactionListPageBaseState>['tt'];
    transactionsStore: ReturnType<typeof useTransactionListPageBaseState>['transactionsStore'];
}

// Owns presentation for the TransactionListPageBase store.
export function useTransactionListPageBasePresentation(context: Context) {
    const {
        queryAllFilterCategoryIds, allTransactionTagsByGroup, queryAllFilterTagIds, formatAmountToLocalizedNumeralsWithCurrency,
        userDefaultCurrency, exchangeRatesStore, totalAmountTypeInTransactionListPage, selectedAccountDefaultCurrency,
        formatDateTimeToShortTime, formatDateTimeToLongDate, formatDateTimeToGregorianLikeLongYearMonth, numeralSystem,
        formatDateTimeToLongDateTime, queryAllFilterAccountIdsCount, queryAllFilterAccountIds, tt, transactionsStore,
    } = context;

    function getDisplayTotalAmount(amount: BigDecimal, currency: string, prefix: string, incomplete: boolean, inDefaultCurrency?: boolean) {
        const displayAmount = formatAmount(amount, false, currency, inDefaultCurrency);
        return prefix + displayAmount + (incomplete ? INCOMPLETE_AMOUNT_SUFFIX : '');
    }

    function hasSubCategoryInQuery(category: TransactionCategory): boolean {
        if (!category.subCategories || !category.subCategories.length) {
            return false;
        }

        for (const subCategory of category.subCategories) {
            if (queryAllFilterCategoryIds.value[subCategory.id]) {
                return true;
            }
        }

        return false;
    }

    function hasVisibleTagsInTagGroup(tagGroup: TransactionTagGroup): boolean {
        const tagsInGroup = allTransactionTagsByGroup.value[tagGroup.id];

        if (!tagsInGroup || !tagsInGroup.length) {
            return false;
        }

        for (const tag of tagsInGroup) {
            if (!tag.hidden || queryAllFilterTagIds.value[tag.id]) {
                return true;
            }
        }

        return false;
    }

    function isSameAsDefaultTimezoneOffsetMinutes(transaction: Transaction): boolean {
        return transaction.utcOffset === getTimezoneOffsetMinutes(transaction.time);
    }

    function formatAmount(amount: BigDecimal, hideAmount: boolean, currencyCode: string, inUserDefaultCurrency?: boolean): string {
        if (hideAmount) {
            return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, currencyCode);
        }

        if (!inUserDefaultCurrency || currencyCode === userDefaultCurrency.value) {
            return formatAmountToLocalizedNumeralsWithCurrency(amount, currencyCode);
        } else {
            const exchangedAmount = exchangeRatesStore.getExchangedAmount(amount, currencyCode, userDefaultCurrency.value);
            return exchangedAmount ? formatAmountToLocalizedNumeralsWithCurrency(exchangedAmount.truncate(), userDefaultCurrency.value) : formatAmountToLocalizedNumeralsWithCurrency(amount, currencyCode);
        }
    }

    function hasMonthTotalAmount(transactionMonth?: TransactionMonthList | null): boolean {
        if (!transactionMonth) {
            return false;
        }

        if (totalAmountTypeInTransactionListPage.value === TransactionAmountType.InflowsAndOutflows) {
            return !!transactionMonth.inflowOutflowTotalAmount;
        } else if (totalAmountTypeInTransactionListPage.value === TransactionAmountType.IncomeAndExpense) {
            return !!transactionMonth.incomeExpenseTotalAmount;
        } else {
            return false;
        }
    }

    function getDailyTotalAmounts(transactionMonth?: TransactionMonthList | null): Record<string, TransactionTotalAmount> | undefined {
        if (!transactionMonth) {
            return undefined;
        }

        if (totalAmountTypeInTransactionListPage.value === TransactionAmountType.InflowsAndOutflows) {
            return transactionMonth.inflowOutflowDailyTotalAmounts;
        } else if (totalAmountTypeInTransactionListPage.value === TransactionAmountType.IncomeAndExpense) {
            return transactionMonth.incomeExpenseDailyTotalAmounts;
        } else {
            return undefined;
        }
    }

    function getDisplayMonthTotalAmount(transactionMonth: TransactionMonthList | null, prefix: string): TransactionListDisplayTotalAmount | null {
        if (!transactionMonth) {
            return null;
        }

        let totalAmount: TransactionTotalAmount;

        if (totalAmountTypeInTransactionListPage.value === TransactionAmountType.InflowsAndOutflows) {
            totalAmount = transactionMonth.inflowOutflowTotalAmount;
        } else if (totalAmountTypeInTransactionListPage.value === TransactionAmountType.IncomeAndExpense) {
            totalAmount = transactionMonth.incomeExpenseTotalAmount;
        } else {
            return null;
        }

        if (!totalAmount) {
            return null;
        }

        const displayMonthlyTotalAmount: TransactionListDisplayTotalAmount = {
            incomeIsZero: totalAmount.income.isZero(),
            expenseIsZero: totalAmount.expense.isZero(),
            income: getDisplayTotalAmount(totalAmount.income, selectedAccountDefaultCurrency.value, prefix, totalAmount.incompleteIncome),
            expense: getDisplayTotalAmount(totalAmount.expense, selectedAccountDefaultCurrency.value, prefix, totalAmount.incompleteExpense),
            incomeInDefaultCurrency: getDisplayTotalAmount(totalAmount.income, selectedAccountDefaultCurrency.value, prefix, totalAmount.incompleteIncome, true),
            expenseInDefaultCurrency: getDisplayTotalAmount(totalAmount.expense, selectedAccountDefaultCurrency.value, prefix, totalAmount.incompleteExpense, true)
        };

        return displayMonthlyTotalAmount;
    }

    function getDisplayTime(transaction: Transaction): string {
        const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transaction.utcOffset);
        return formatDateTimeToShortTime(dateTime);
    }

    function getDisplayLongDate(transaction: Transaction): string {
        const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transaction.utcOffset);
        return formatDateTimeToLongDate(dateTime);
    }

    function getDisplayLongYearMonth(transactionMonthList: TransactionMonthList): string {
        const yearMonthDateTime = parseDateTimeFromUnixTime(getYearMonthFirstUnixTime(transactionMonthList.yearDashMonth));
        return formatDateTimeToGregorianLikeLongYearMonth(yearMonthDateTime);
    }

    function getDisplayTimezone(transaction: Transaction): string {
        const utcOffset = numeralSystem.value.replaceWesternArabicDigitsToLocalizedDigits(getUtcOffsetByUtcOffsetMinutes(transaction.utcOffset));
        return `UTC${utcOffset}`;
    }

    function getDisplayTimeInDefaultTimezone(transaction: Transaction): string {
        const timezoneOffsetMinutes = getTimezoneOffsetMinutes(transaction.time);
        const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, timezoneOffsetMinutes);
        const utcOffset = numeralSystem.value.replaceWesternArabicDigitsToLocalizedDigits(getUtcOffsetByUtcOffsetMinutes(timezoneOffsetMinutes));
        return `${formatDateTimeToLongDateTime(dateTime)} (UTC${utcOffset})`;
    }

    function getDisplayAmount(transaction: Transaction, inUserDefaultCurrency?: boolean): string {
        if (queryAllFilterAccountIdsCount.value < 1) {
            if (transaction.sourceAccount) {
                return formatAmount(parseBigDecimal(transaction.sourceAmount), transaction.hideAmount, transaction.sourceAccount.currency, inUserDefaultCurrency);
            }
        } else if (queryAllFilterAccountIdsCount.value === 1) {
            if (transaction.sourceAccount && (queryAllFilterAccountIds.value[transaction.sourceAccount.id] || queryAllFilterAccountIds.value[transaction.sourceAccount.parentId])) {
                return formatAmount(parseBigDecimal(transaction.sourceAmount), transaction.hideAmount, transaction.sourceAccount.currency, inUserDefaultCurrency);
            } else if (transaction.destinationAccount && (queryAllFilterAccountIds.value[transaction.destinationAccount.id] || queryAllFilterAccountIds.value[transaction.destinationAccount.parentId])) {
                return formatAmount(parseBigDecimal(transaction.destinationAmount), transaction.hideAmount, transaction.destinationAccount.currency, inUserDefaultCurrency);
            }
        } else { // queryAllFilterAccountIdsCount.value > 1
            if (transaction.sourceAccount && transaction.destinationAccount) {
                if ((queryAllFilterAccountIds.value[transaction.sourceAccount.id] || queryAllFilterAccountIds.value[transaction.sourceAccount.parentId])
                    && !queryAllFilterAccountIds.value[transaction.destinationAccount.id] && !queryAllFilterAccountIds.value[transaction.destinationAccount.parentId]) {
                    return formatAmount(parseBigDecimal(transaction.sourceAmount), transaction.hideAmount, transaction.sourceAccount.currency, inUserDefaultCurrency);
                } else if ((queryAllFilterAccountIds.value[transaction.destinationAccount.id] || queryAllFilterAccountIds.value[transaction.destinationAccount.parentId])
                    && !queryAllFilterAccountIds.value[transaction.sourceAccount.id] && !queryAllFilterAccountIds.value[transaction.sourceAccount.parentId]) {
                    return formatAmount(parseBigDecimal(transaction.destinationAmount), transaction.hideAmount, transaction.destinationAccount.currency, inUserDefaultCurrency);
                }
            }
        }

        if (transaction.sourceAccount) {
            return formatAmount(parseBigDecimal(transaction.sourceAmount), transaction.hideAmount, transaction.sourceAccount.currency, inUserDefaultCurrency);
        }

        return '';
    }

    function getDisplayAmountCurrency(transaction: Transaction): string {
        if (queryAllFilterAccountIdsCount.value < 1) {
            if (transaction.sourceAccount) {
                return transaction.sourceAccount.currency;
            }
        } else if (queryAllFilterAccountIdsCount.value === 1) {
            if (transaction.sourceAccount && (queryAllFilterAccountIds.value[transaction.sourceAccount.id] || queryAllFilterAccountIds.value[transaction.sourceAccount.parentId])) {
                return transaction.sourceAccount.currency;
            } else if (transaction.destinationAccount && (queryAllFilterAccountIds.value[transaction.destinationAccount.id] || queryAllFilterAccountIds.value[transaction.destinationAccount.parentId])) {
                return transaction.destinationAccount.currency;
            }
        } else { // queryAllFilterAccountIdsCount.value > 1
            if (transaction.sourceAccount && transaction.destinationAccount) {
                if ((queryAllFilterAccountIds.value[transaction.sourceAccount.id] || queryAllFilterAccountIds.value[transaction.sourceAccount.parentId])
                    && !queryAllFilterAccountIds.value[transaction.destinationAccount.id] && !queryAllFilterAccountIds.value[transaction.destinationAccount.parentId]) {
                    return transaction.sourceAccount.currency;
                } else if ((queryAllFilterAccountIds.value[transaction.destinationAccount.id] || queryAllFilterAccountIds.value[transaction.destinationAccount.parentId])
                    && !queryAllFilterAccountIds.value[transaction.sourceAccount.id] && !queryAllFilterAccountIds.value[transaction.sourceAccount.parentId]) {
                    return transaction.destinationAccount.currency;
                }
            }
        }

        if (transaction.sourceAccount) {
            return transaction.sourceAccount.currency;
        }

        return userDefaultCurrency.value;
    }

    function getTransactionTypeName(type: number | null, defaultName: string): string {
        switch (type){
            case TransactionType.ModifyBalance:
                return tt('Modify Balance');
            case TransactionType.Income:
                return tt('Income');
            case TransactionType.Expense:
                return tt('Expense');
            case TransactionType.Transfer:
                return tt('Transfer');
            default:
                return tt(defaultName);
        }
    }

    function getTransactionPictureUrl(pictureInfo?: TransactionPictureInfoBasicResponse | null): string | undefined {
        return transactionsStore.getTransactionPictureUrl(pictureInfo);
    }

    return {
        hasSubCategoryInQuery, hasVisibleTagsInTagGroup, isSameAsDefaultTimezoneOffsetMinutes, hasMonthTotalAmount,
        getDailyTotalAmounts, getDisplayMonthTotalAmount, getDisplayTime, getDisplayLongDate, getDisplayLongYearMonth,
        getDisplayTimezone, getDisplayTimeInDefaultTimezone, getDisplayAmount, getDisplayAmountCurrency, getTransactionTypeName,
        getTransactionPictureUrl,
    };
}
