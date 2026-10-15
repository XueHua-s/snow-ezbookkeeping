import {
    TransactionExplorerChartType,
    TransactionExplorerDataDimension
} from '@/core/explorer.ts';
import type { Account } from '@/models/account.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionTag } from '@/models/transaction_tag.ts';
import { computed } from 'vue';
import type { useExplorerState } from './state.ts';

interface Context {
    transactionExplorerAllData: ReturnType<typeof useExplorerState>['transactionExplorerAllData'];
    accountsStore: ReturnType<typeof useExplorerState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useExplorerState>['transactionCategoriesStore'];
    transactionTagsStore: ReturnType<typeof useExplorerState>['transactionTagsStore'];
    currentExploration: ReturnType<typeof useExplorerState>['currentExploration'];
}

// Owns transactions for the explorer store.
export function useExplorerTransactions(context: Context) {
    const { transactionExplorerAllData, accountsStore, transactionCategoriesStore, transactionTagsStore, currentExploration } = context;

    const allTransactions = computed<TransactionInsightDataItem[]>(() => {
        if (!transactionExplorerAllData.value || transactionExplorerAllData.value.length < 1) {
            return [];
        }

        const result: TransactionInsightDataItem[] = [];

        for (const transaction of transactionExplorerAllData.value) {
            const sourceAccount: Account | undefined = accountsStore.allAccountsMap[transaction.sourceAccountId];

            if (!sourceAccount) {
                continue;
            }

            let destinationAccount: Account | undefined = undefined

            if (transaction.destinationAccountId && transaction.destinationAccountId !== '0') {
                destinationAccount = accountsStore.allAccountsMap[transaction.destinationAccountId];

                if (!destinationAccount) {
                    continue;
                }
            }

            const secondaryCategory: TransactionCategory | undefined = transactionCategoriesStore.allTransactionCategoriesMap[transaction.categoryId];

            if (!secondaryCategory) {
                continue;
            }

            const primaryCategory: TransactionCategory | undefined = transactionCategoriesStore.allTransactionCategoriesMap[secondaryCategory.parentId];

            if (!primaryCategory) {
                continue;
            }

            const tags: TransactionTag[] = [];

            for (const tagId of transaction.tagIds) {
                const tag: TransactionTag | undefined = transactionTagsStore.allTransactionTagsMap[tagId];

                if (tag) {
                    tags.push(tag);
                }
            }

            const item: TransactionInsightDataItem = {
                ...transaction,
                id: transaction.id,
                time: transaction.time,
                utcOffset: transaction.utcOffset,
                type: transaction.type,
                primaryCategory: primaryCategory,
                primaryCategoryName: primaryCategory.name,
                secondaryCategory: secondaryCategory,
                secondaryCategoryName: secondaryCategory.name,
                sourceAccount: sourceAccount,
                sourceAccountName: sourceAccount.name,
                destinationAccount: destinationAccount,
                destinationAccountName: destinationAccount?.name,
                sourceAmount: transaction.sourceAmount,
                destinationAmount: transaction.destinationAmount,
                hideAmount: transaction.hideAmount,
                tags: tags,
                comment: transaction.comment,
                geoLocation: transaction.geoLocation
            };

            result.push(item);
        }

        return result;
    });

    const isUsingAmountRange = computed<boolean>(() => {
        const chartType = TransactionExplorerChartType.valueOf(currentExploration.value.chartType);
        const categoryDimension = TransactionExplorerDataDimension.valueOf(currentExploration.value.categoryDimension);
        const seriesDimension = chartType?.seriesDimensionRequired ? TransactionExplorerDataDimension.valueOf(currentExploration.value.seriesDimension) : TransactionExplorerDataDimension.SeriesDimensionDefault;
        return categoryDimension?.isSourceAmountRange || seriesDimension?.isSourceAmountRange
            || categoryDimension?.isDestinationAmountRange || seriesDimension?.isDestinationAmountRange
            || false;
    });

    return { allTransactions, isUsingAmountRange };
}
