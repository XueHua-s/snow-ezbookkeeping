import { AMOUNT_FACTOR } from '@/consts/numeral.ts';
import { itemAndIndex } from '@/core/base.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    coefficientOfVariation,
    cumulativePercentage,
    median,
    percentile,
    sumMaxN,
    varianceAndStandardDeviation
} from '@/lib/math.ts';
import {
    BIG_DECIMAL_NEGATIVE_INFINITY,
    BIG_DECIMAL_POSITIVE_INFINITY,
    BIG_DECIMAL_ZERO,
    parseBigDecimal
} from '@/lib/numeral.ts';
import type { InsightsExplorerMatchContext } from '@/models/explorer.ts';
import type { TransactionInsightDataItem, TransactionInsightDataItemWithQueryIndexes } from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useExplorerState } from './state.ts';
import type { useExplorerTransactions } from './transactions.ts';
import type { InsightsExplorerTransactionStatisticData } from './types.ts';

interface Context {
    allTransactions: ReturnType<typeof useExplorerTransactions>['allTransactions'];
    currentExploration: ReturnType<typeof useExplorerState>['currentExploration'];
    buildInsightsExplorerMatchContext: ReturnType<typeof useExplorerState>['buildInsightsExplorerMatchContext'];
    userStore: ReturnType<typeof useExplorerState>['userStore'];
    exchangeRatesStore: ReturnType<typeof useExplorerState>['exchangeRatesStore'];
}

// Owns table for the explorer store.
export function useExplorerTable(context: Context) {
    const { allTransactions, currentExploration, buildInsightsExplorerMatchContext, userStore, exchangeRatesStore } = context;

    const filteredTransactionsInDataTable = computed<TransactionInsightDataItem[]>(() => {
        if (!allTransactions.value || allTransactions.value.length < 1) {
            return [];
        }

        if (!currentExploration.value.queries || currentExploration.value.queries.length < 1) {
            return allTransactions.value;
        }

        const result: TransactionInsightDataItem[] = [];

        for (const transaction of allTransactions.value) {
            const matchOptions: InsightsExplorerMatchContext = buildInsightsExplorerMatchContext(currentExploration.value, transaction);

            for (const query of currentExploration.value.queries) {
                if (currentExploration.value.datatableQuerySource && currentExploration.value.datatableQuerySource !== query.id) {
                    continue;
                }

                if (query.match(transaction, matchOptions)) {
                    result.push(transaction);
                    break;
                }
            }
        }

        return result;
    });

    const filteredTransactionsForCustomChart = computed<TransactionInsightDataItemWithQueryIndexes[]>(() => {
        if (!allTransactions.value || allTransactions.value.length < 1) {
            return [];
        }

        if (!currentExploration.value.queries || currentExploration.value.queries.length < 1) {
            return allTransactions.value.map(transaction => ({
                ...transaction,
                queryIndexes: []
            }));
        }

        const result: TransactionInsightDataItemWithQueryIndexes[] = [];

        for (const transaction of allTransactions.value) {
            const matchOptions: InsightsExplorerMatchContext = buildInsightsExplorerMatchContext(currentExploration.value, transaction);
            const queryIndexes: number[] = [];

            for (const [query, index] of itemAndIndex(currentExploration.value.queries)) {
                if (query.match(transaction, matchOptions)) {
                    queryIndexes.push(index);
                }
            }

            if (queryIndexes.length > 0) {
                result.push({
                    ...transaction,
                    queryIndexes: queryIndexes
                });
            }
        }

        return result;
    });

    const filteredTransactionsInDataTableStatistic = computed<InsightsExplorerTransactionStatisticData>(() => {
        const defaultCurrency = userStore.currentUserDefaultCurrency;
        const statisticData: InsightsExplorerTransactionStatisticData = {
            totalCount: 0,
            totalAmount: BIG_DECIMAL_ZERO,
            totalIncome: BIG_DECIMAL_ZERO,
            totalExpense: BIG_DECIMAL_ZERO,
            netIncome: BIG_DECIMAL_ZERO,
            averageAmount: BIG_DECIMAL_ZERO,
            medianAmount: BIG_DECIMAL_ZERO,
            minimumAmount: BIG_DECIMAL_POSITIVE_INFINITY,
            maximumAmount: BIG_DECIMAL_NEGATIVE_INFINITY,
            p90Amount: BIG_DECIMAL_ZERO,
            range: BIG_DECIMAL_ZERO,
            interquartileRange: BIG_DECIMAL_ZERO,
            medianToMeanRatio: undefined,
            top5AmountShare: undefined,
            transactionsFor80PercentAmount: undefined,
            variance: undefined,
            standardDeviation: undefined,
            coefficientOfVariation: undefined
        };

        const sourceAmounts: BigDecimal[] = [];

        for (const transaction of filteredTransactionsInDataTable.value) {
            statisticData.totalCount++;

            let amountInDefaultCurrency: BigDecimal = parseBigDecimal(transaction.sourceAmount);

            if (transaction.sourceAccount.currency !== defaultCurrency) {
                const amount = exchangeRatesStore.getExchangedAmount(parseBigDecimal(transaction.sourceAmount), transaction.sourceAccount.currency, defaultCurrency);

                if (amount) {
                    amountInDefaultCurrency = amount.truncate();
                } else {
                    continue;
                }
            }

            statisticData.totalAmount = statisticData.totalAmount.add(amountInDefaultCurrency);
            sourceAmounts.push(amountInDefaultCurrency);

            if (transaction.type === TransactionType.Income) {
                statisticData.totalIncome = statisticData.totalIncome.add(amountInDefaultCurrency);
            } else if (transaction.type === TransactionType.Expense) {
                statisticData.totalExpense = statisticData.totalExpense.add(amountInDefaultCurrency);
            }

            if (amountInDefaultCurrency.isPositiveOrZero() && amountInDefaultCurrency.lessThan(statisticData.minimumAmount)) {
                statisticData.minimumAmount = amountInDefaultCurrency;
            }

            if (amountInDefaultCurrency.greaterThan(statisticData.maximumAmount)) {
                statisticData.maximumAmount = amountInDefaultCurrency;
            }
        }

        statisticData.netIncome = statisticData.totalIncome.subtract(statisticData.totalExpense);

        if (statisticData.minimumAmount.isPositiveInfinity()) {
            statisticData.minimumAmount = BIG_DECIMAL_ZERO;
        }

        if (statisticData.maximumAmount.isNegativeInfinity()) {
            statisticData.maximumAmount = BIG_DECIMAL_ZERO;
        }

        if (sourceAmounts.length > 0) {
            statisticData.averageAmount = statisticData.totalAmount.divide(sourceAmounts.length).truncate();
        }

        statisticData.range = statisticData.maximumAmount.subtract(statisticData.minimumAmount);

        if (sourceAmounts.length > 0) {
            sourceAmounts.sort((a, b) => a.compareTo(b));
            statisticData.medianAmount = median(sourceAmounts, item => item).truncate();
            statisticData.p90Amount = percentile(sourceAmounts, 0.9, item => item).truncate();

            const q1 = percentile(sourceAmounts, 0.25, item => item);
            const q3 = percentile(sourceAmounts, 0.75, item => item);
            statisticData.interquartileRange = q3.subtract(q1).truncate();
            statisticData.medianToMeanRatio = !statisticData.averageAmount.isZero() ? statisticData.medianAmount.divide(statisticData.averageAmount) : undefined;
        }

        if (sourceAmounts.length > 5) {
            const top5AmountSum = sumMaxN(sourceAmounts, 5, item => item);
            statisticData.top5AmountShare = statisticData.totalAmount.isPositive() ? top5AmountSum.divide(statisticData.totalAmount).multiply(100) : BIG_DECIMAL_ZERO;
        }

        if (sourceAmounts.length > 0) {
            statisticData.transactionsFor80PercentAmount = cumulativePercentage(sourceAmounts, 0.8, statisticData.totalAmount, item => item);
        }

        if (sourceAmounts.length > 0) {
            const averageAmountForVarianceCalculation: BigDecimal = statisticData.totalAmount.divide(sourceAmounts.length).divide(AMOUNT_FACTOR);
            const { variance, standardDeviation } = varianceAndStandardDeviation(sourceAmounts, averageAmountForVarianceCalculation, item => item.divide(AMOUNT_FACTOR));
            statisticData.variance = variance;
            statisticData.standardDeviation = standardDeviation;
            statisticData.coefficientOfVariation = coefficientOfVariation(standardDeviation, averageAmountForVarianceCalculation);
        }

        return statisticData;
    });

    return { filteredTransactionsInDataTable, filteredTransactionsForCustomChart, filteredTransactionsInDataTableStatistic };
}
