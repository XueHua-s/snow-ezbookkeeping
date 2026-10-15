import {
    parseBigDecimal
} from '@/lib/numeral.ts';
import type { TransactionStatisticResponseItem, TransactionStatisticResponseItemWithInfo, TransactionStatisticResponseWithInfo } from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useStatisticsState } from './state.ts';

interface Context {
    transactionCategoryStatisticsData: ReturnType<typeof useStatisticsState>['transactionCategoryStatisticsData'];
    userStore: ReturnType<typeof useStatisticsState>['userStore'];
    accountsStore: ReturnType<typeof useStatisticsState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useStatisticsState>['transactionCategoriesStore'];
    exchangeRatesStore: ReturnType<typeof useStatisticsState>['exchangeRatesStore'];
}

// Owns metadata for the statistics store.
export function useStatisticsMetadata(context: Context) {
    const { transactionCategoryStatisticsData, userStore, accountsStore, transactionCategoriesStore, exchangeRatesStore } = context;

    const transactionCategoryStatisticsDataWithCategoryAndAccountInfo = computed<TransactionStatisticResponseWithInfo | null>(() => {
        const statistics = transactionCategoryStatisticsData.value;

        if (!statistics) {
            return null;
        }

        const finalStatistics: TransactionStatisticResponseWithInfo = {
            startTime: statistics.startTime,
            endTime: statistics.endTime,
            items: []
        };

        if (statistics && statistics.items && statistics.items.length) {
            finalStatistics.items.push(...assembleAccountAndCategoryInfo(statistics.items));
        }

        return finalStatistics;
    });

    function assembleAccountAndCategoryInfo(items: TransactionStatisticResponseItem[]): TransactionStatisticResponseItemWithInfo[] {
        const finalItems: TransactionStatisticResponseItemWithInfo[] = [];
        const defaultCurrency = userStore.currentUserDefaultCurrency;

        for (const dataItem of items) {
            const item: TransactionStatisticResponseItemWithInfo = {
                categoryId: dataItem.categoryId,
                accountId: dataItem.accountId,
                relatedAccountId: dataItem.relatedAccountId,
                relatedAccountType: dataItem.relatedAccountType,
                amount: dataItem.amount,
                amountInDefaultCurrency: null
            };

            if (item.accountId) {
                item.account = accountsStore.allAccountsMap[item.accountId];
            }

            if (item.account && item.account.parentId !== '0') {
                item.primaryAccount = accountsStore.allAccountsMap[item.account.parentId];
            } else {
                item.primaryAccount = item.account;
            }

            if (item.relatedAccountId) {
                item.relatedAccount = accountsStore.allAccountsMap[item.relatedAccountId];
            }

            if (item.relatedAccount && item.relatedAccount.parentId !== '0') {
                item.relatedPrimaryAccount = accountsStore.allAccountsMap[item.relatedAccount.parentId];
            } else {
                item.relatedPrimaryAccount = item.relatedAccount;
            }

            if (item.categoryId) {
                item.category = transactionCategoriesStore.allTransactionCategoriesMap[item.categoryId];
            }

            if (item.category && item.category.parentId !== '0') {
                item.primaryCategory = transactionCategoriesStore.allTransactionCategoriesMap[item.category.parentId];
            } else {
                item.primaryCategory = item.category;
            }

            if (item.account && item.account.currency !== defaultCurrency) {
                const amount = exchangeRatesStore.getExchangedAmount(parseBigDecimal(item.amount), item.account.currency, defaultCurrency);

                if (amount) {
                    item.amountInDefaultCurrency = amount.truncate();
                }
            } else if (item.account && item.account.currency === defaultCurrency) {
                item.amountInDefaultCurrency = parseBigDecimal(item.amount);
            } else {
                item.amountInDefaultCurrency = null;
            }

            finalItems.push(item);
        }

        return finalItems;
    }

    return { transactionCategoryStatisticsDataWithCategoryAndAccountInfo, assembleAccountAndCategoryInfo };
}
