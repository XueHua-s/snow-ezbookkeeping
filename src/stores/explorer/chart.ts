import { AMOUNT_FACTOR } from '@/consts/numeral.ts';
import { values } from '@/core/base.ts';
import type { DateTime } from '@/core/datetime.ts';
import {
    TransactionExplorerChartType,
    TransactionExplorerDataDimension,
    TransactionExplorerValueMetric
} from '@/core/explorer.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    getObjectOwnFieldCount,
    isDefined
} from '@/lib/common.ts';
import {
    parseDateTimeFromUnixTime,
    parseDateTimeFromUnixTimeWithTimezoneOffset
} from '@/lib/datetime.ts';
import {
    coefficientOfVariation,
    cumulativePercentage,
    giniCoefficient,
    herfindahlHirschmanIndex,
    kurtosis,
    meanAbsoluteDeviation,
    median,
    medianAbsoluteDeviation,
    percentile,
    skewness,
    sumMaxN,
    varianceAndStandardDeviation
} from '@/lib/math.ts';
import {
    BIG_DECIMAL_NEGATIVE_INFINITY,
    BIG_DECIMAL_POSITIVE_INFINITY,
    BIG_DECIMAL_ZERO,
    parseBigDecimal
} from '@/lib/numeral.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useExplorerCategorization } from './categorization.ts';
import type { useExplorerState } from './state.ts';
import type { useExplorerTransactions } from './transactions.ts';
import { type CategoriedTransactionExplorerData, type CategoriedTransactionExplorerDataItem, type SeriesTransactions, TransactionExplorerDimensionType } from './types.ts';

interface Context {
    allTransactions: ReturnType<typeof useExplorerTransactions>['allTransactions'];
    currentExploration: ReturnType<typeof useExplorerState>['currentExploration'];
    userStore: ReturnType<typeof useExplorerState>['userStore'];
    categoriedTransactions: ReturnType<typeof useExplorerCategorization>['categoriedTransactions'];
    exchangeRatesStore: ReturnType<typeof useExplorerState>['exchangeRatesStore'];
}

// Owns chart for the explorer store.
export function useExplorerChart(context: Context) {
    const { allTransactions, currentExploration, userStore, categoriedTransactions, exchangeRatesStore } = context;

    const categoriedTransactionExplorerData = computed<CategoriedTransactionExplorerData[]>(() => {
        if (!allTransactions.value || allTransactions.value.length < 1) {
            return [];
        }

        const chartType = TransactionExplorerChartType.valueOf(currentExploration.value.chartType);
        const categoryDimension = TransactionExplorerDataDimension.valueOf(currentExploration.value.categoryDimension);
        const seriesDimension = chartType?.seriesDimensionRequired ? TransactionExplorerDataDimension.valueOf(currentExploration.value.seriesDimension) : TransactionExplorerDataDimension.SeriesDimensionDefault;
        const valueMetric = TransactionExplorerValueMetric.valueOf(currentExploration.value.valueMetric);

        if (!chartType || !categoryDimension || !seriesDimension || !valueMetric) {
            return [];
        }

        const defaultCurrency = userStore.currentUserDefaultCurrency;
        const result: CategoriedTransactionExplorerData[] = [];
        const categoriedDataMap = categoriedTransactions.value;
        let needCalculateDailyTransactionCount: boolean = false;

        if (valueMetric === TransactionExplorerValueMetric.ActiveTransactionDays
            || valueMetric === TransactionExplorerValueMetric.TransactionsPerDay
            || valueMetric === TransactionExplorerValueMetric.AverageAmountPerActiveDay) {
            needCalculateDailyTransactionCount = true;
        }

        for (const categoriedTransactions of values(categoriedDataMap)) {
            const dataItems: CategoriedTransactionExplorerDataItem[] = [];
            let allSeriesTransactions: Record<string, SeriesTransactions> = categoriedTransactions.trasactions;

            if (!chartType.seriesDimensionRequired) {
                const transactions: TransactionInsightDataItem[] = [];

                for (const seriesTransactions of values(categoriedTransactions.trasactions)) {
                    transactions.push(...seriesTransactions.trasactions);
                }

                allSeriesTransactions = {};
                allSeriesTransactions['none'] = {
                    seriesName: valueMetric?.name ?? 'Unknown',
                    seriesNameNeedI18n: true,
                    seriesId: 'none',
                    seriesIdType: TransactionExplorerDimensionType.Other,
                    seriesDisplayOrders: [0],
                    trasactions: transactions
                };
            }

            for (const seriesTransactions of values(allSeriesTransactions)) {
                const transactionDateMapCount: Record<string, number> = {};
                const allSourceAmountsInDefaultCurrency: BigDecimal[] = [];
                let totalSourceAmountSumInDefaultCurrency: BigDecimal = BIG_DECIMAL_ZERO;
                let totalSourceIncomeAmountSumInDefaultCurrency: BigDecimal = BIG_DECIMAL_ZERO;
                let totalSourceExpenseAmountSumInDefaultCurrency: BigDecimal = BIG_DECIMAL_ZERO;
                let minimumSourceAmountInDefaultCurrency: BigDecimal = BIG_DECIMAL_POSITIVE_INFINITY;
                let maximumSourceAmountInDefaultCurrency: BigDecimal = BIG_DECIMAL_NEGATIVE_INFINITY;

                for (const transaction of seriesTransactions.trasactions) {
                    let amountInDefaultCurrency: BigDecimal = parseBigDecimal(transaction.sourceAmount);

                    if (transaction.sourceAccount.currency !== defaultCurrency) {
                        const amount = exchangeRatesStore.getExchangedAmount(parseBigDecimal(transaction.sourceAmount), transaction.sourceAccount.currency, defaultCurrency);

                        if (amount) {
                            amountInDefaultCurrency = amount.truncate();
                        } else {
                            continue;
                        }
                    }

                    if (needCalculateDailyTransactionCount) {
                        let transactionTimeUtfOffset: number | undefined = undefined;

                        if (currentExploration.value.timezoneUsedForDateRange === TimezoneTypeForStatistics.TransactionTimezone.type) {
                            transactionTimeUtfOffset = transaction.utcOffset;
                        }

                        const transactionDateTime: DateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);
                        const transactionYearMonthDay: string = transactionDateTime.getGregorianCalendarYearDashMonthDashDay();

                        if (transactionDateMapCount[transactionYearMonthDay]) {
                            transactionDateMapCount[transactionYearMonthDay]++;
                        } else {
                            transactionDateMapCount[transactionYearMonthDay] = 1;
                        }
                    }

                    allSourceAmountsInDefaultCurrency.push(amountInDefaultCurrency);
                    totalSourceAmountSumInDefaultCurrency = totalSourceAmountSumInDefaultCurrency.add(amountInDefaultCurrency);

                    if (transaction.type === TransactionType.Income) {
                        totalSourceIncomeAmountSumInDefaultCurrency = totalSourceIncomeAmountSumInDefaultCurrency.add(amountInDefaultCurrency);
                    } else if (transaction.type === TransactionType.Expense) {
                        totalSourceExpenseAmountSumInDefaultCurrency = totalSourceExpenseAmountSumInDefaultCurrency.add(amountInDefaultCurrency);
                    }

                    if (amountInDefaultCurrency.isPositiveOrZero() && amountInDefaultCurrency.lessThan(minimumSourceAmountInDefaultCurrency)) {
                        minimumSourceAmountInDefaultCurrency = amountInDefaultCurrency;
                    }

                    if (amountInDefaultCurrency.greaterThan(maximumSourceAmountInDefaultCurrency)) {
                        maximumSourceAmountInDefaultCurrency = amountInDefaultCurrency;
                    }
                }

                let value: BigDecimal = BIG_DECIMAL_ZERO;

                if (valueMetric === TransactionExplorerValueMetric.TransactionCount) {
                    value = parseBigDecimal(allSourceAmountsInDefaultCurrency.length);
                } else if (valueMetric === TransactionExplorerValueMetric.ActiveTransactionDays) {
                    value = parseBigDecimal(getObjectOwnFieldCount(transactionDateMapCount));
                } else if (valueMetric === TransactionExplorerValueMetric.TransactionsPerDay) {
                    const activeDays = getObjectOwnFieldCount(transactionDateMapCount);
                    value = activeDays > 0 ? parseBigDecimal(allSourceAmountsInDefaultCurrency.length).divide(activeDays) : BIG_DECIMAL_ZERO;
                } else if (valueMetric === TransactionExplorerValueMetric.AverageAmountPerActiveDay) {
                    const activeDays = getObjectOwnFieldCount(transactionDateMapCount);
                    value = activeDays > 0 ? totalSourceAmountSumInDefaultCurrency.divide(activeDays).truncate() : BIG_DECIMAL_ZERO;
                } else if (valueMetric === TransactionExplorerValueMetric.SourceIncomeAmountSum) {
                    value = totalSourceIncomeAmountSumInDefaultCurrency;
                } else if (valueMetric === TransactionExplorerValueMetric.SourceExpenseAmountSum) {
                    value = totalSourceExpenseAmountSumInDefaultCurrency;
                } else if (valueMetric === TransactionExplorerValueMetric.SourceNetIncomeAmountSum) {
                    value = totalSourceIncomeAmountSumInDefaultCurrency.subtract(totalSourceExpenseAmountSumInDefaultCurrency);
                } else if (valueMetric === TransactionExplorerValueMetric.SrouceAmountExpenseIncomeRatio) {
                    value = !totalSourceIncomeAmountSumInDefaultCurrency.isZero() ? totalSourceExpenseAmountSumInDefaultCurrency.divide(totalSourceIncomeAmountSumInDefaultCurrency).multiply(100) : BIG_DECIMAL_ZERO;
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountSavingsRate) {
                    value = !totalSourceIncomeAmountSumInDefaultCurrency.isZero() ? totalSourceIncomeAmountSumInDefaultCurrency.subtract(totalSourceExpenseAmountSumInDefaultCurrency).divide(totalSourceIncomeAmountSumInDefaultCurrency).multiply(100) : BIG_DECIMAL_ZERO;
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountSum) {
                    value = totalSourceAmountSumInDefaultCurrency;
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountAverage) {
                    value = allSourceAmountsInDefaultCurrency.length > 0 ? totalSourceAmountSumInDefaultCurrency.divide(allSourceAmountsInDefaultCurrency.length).truncate() : BIG_DECIMAL_ZERO;
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountMedian) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
                        value = median(allSourceAmountsInDefaultCurrency, item => item).truncate();
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountMinimum) {
                    value = minimumSourceAmountInDefaultCurrency.isPositiveInfinity() ? BIG_DECIMAL_ZERO : minimumSourceAmountInDefaultCurrency;
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountMaximum) {
                    value = maximumSourceAmountInDefaultCurrency.isNegativeInfinity() ? BIG_DECIMAL_ZERO : maximumSourceAmountInDefaultCurrency;
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountQ1Amount
                    || valueMetric === TransactionExplorerValueMetric.SourceAmountQ3Amount
                    || valueMetric === TransactionExplorerValueMetric.SourceAmount10thPercentile
                    || valueMetric === TransactionExplorerValueMetric.SourceAmount90thPercentile
                    || valueMetric === TransactionExplorerValueMetric.SourceAmount95thPercentile
                    || valueMetric === TransactionExplorerValueMetric.SourceAmount99thPercentile) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));

                        if (valueMetric === TransactionExplorerValueMetric.SourceAmountQ1Amount) {
                            value = percentile(allSourceAmountsInDefaultCurrency, 0.25, item => item).truncate();
                        } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountQ3Amount) {
                            value = percentile(allSourceAmountsInDefaultCurrency, 0.75, item => item).truncate();
                        } else if (valueMetric === TransactionExplorerValueMetric.SourceAmount10thPercentile) {
                            value = percentile(allSourceAmountsInDefaultCurrency, 0.1, item => item).truncate();
                        } else if (valueMetric === TransactionExplorerValueMetric.SourceAmount90thPercentile) {
                            value = percentile(allSourceAmountsInDefaultCurrency, 0.9, item => item).truncate();
                        } else if (valueMetric === TransactionExplorerValueMetric.SourceAmount95thPercentile) {
                            value = percentile(allSourceAmountsInDefaultCurrency, 0.95, item => item).truncate();
                        } else if (valueMetric === TransactionExplorerValueMetric.SourceAmount99thPercentile) {
                            value = percentile(allSourceAmountsInDefaultCurrency, 0.99, item => item).truncate();
                        }
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountRange) {
                    const finalMinimumSourceAmountInDefaultCurrency = minimumSourceAmountInDefaultCurrency.isPositiveInfinity() ? BIG_DECIMAL_ZERO : minimumSourceAmountInDefaultCurrency;
                    const finalMaximumSourceAmountInDefaultCurrency = maximumSourceAmountInDefaultCurrency.isNegativeInfinity() ? BIG_DECIMAL_ZERO : maximumSourceAmountInDefaultCurrency;
                    value = finalMaximumSourceAmountInDefaultCurrency.subtract(finalMinimumSourceAmountInDefaultCurrency);
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountInterquartileRange) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
                        const q1 = percentile(allSourceAmountsInDefaultCurrency, 0.25, item => item);
                        const q3 = percentile(allSourceAmountsInDefaultCurrency, 0.75, item => item);
                        value = q3.subtract(q1).truncate();
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountMeanAbsoluteDeviation) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        const averageSourceAmountInDefaultCurrency = totalSourceAmountSumInDefaultCurrency.divide(allSourceAmountsInDefaultCurrency.length);
                        value = meanAbsoluteDeviation(allSourceAmountsInDefaultCurrency, averageSourceAmountInDefaultCurrency, item => item).truncate();
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountMedianAbsoluteDeviation) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
                        const medianSourceAmountInDefaultCurrency = median(allSourceAmountsInDefaultCurrency, item => item);
                        value = medianAbsoluteDeviation(allSourceAmountsInDefaultCurrency, medianSourceAmountInDefaultCurrency, item => item).truncate();
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountMedianToMeanRatio) {
                    if (allSourceAmountsInDefaultCurrency.length > 0 && !totalSourceAmountSumInDefaultCurrency.isZero()) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
                        const medianSourceAmountInDefaultCurrency = median(allSourceAmountsInDefaultCurrency, item => item);
                        const averageSourceAmountInDefaultCurrency = totalSourceAmountSumInDefaultCurrency.divide(allSourceAmountsInDefaultCurrency.length);
                        value = !averageSourceAmountInDefaultCurrency.isZero() ? medianSourceAmountInDefaultCurrency.divide(averageSourceAmountInDefaultCurrency) : BIG_DECIMAL_ZERO;
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceMaximumAmountShare) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        value = !maximumSourceAmountInDefaultCurrency.isNegativeInfinity() ? maximumSourceAmountInDefaultCurrency.divide(totalSourceAmountSumInDefaultCurrency).multiply(100) : BIG_DECIMAL_ZERO;
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceTop5AmountSum) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
                        value = sumMaxN(allSourceAmountsInDefaultCurrency, 5, item => item);
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceTop5AmountShare) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
                        const top5AmountSum = sumMaxN(allSourceAmountsInDefaultCurrency, 5, item => item);
                        value = totalSourceAmountSumInDefaultCurrency.isPositive() ? top5AmountSum.divide(totalSourceAmountSumInDefaultCurrency).multiply(100) : BIG_DECIMAL_ZERO;
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.TransactionsForEightyPercentOfSourceAmount) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
                        value = cumulativePercentage(allSourceAmountsInDefaultCurrency, 0.8, totalSourceAmountSumInDefaultCurrency, item => item);
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountOutlierCount
                    || valueMetric === TransactionExplorerValueMetric.SourceAmountOutlierRatio) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
                        const q1 = percentile(allSourceAmountsInDefaultCurrency, 0.25, item => item);
                        const q3 = percentile(allSourceAmountsInDefaultCurrency, 0.75, item => item);
                        const iqr = q3.subtract(q1);
                        const lowerBound = q1.subtract(iqr.multiply(1.5));
                        const upperBound = q3.add(iqr.multiply(1.5));

                        let outlierCount = 0;
                        for (const amount of allSourceAmountsInDefaultCurrency) {
                            if (amount.lessThan(lowerBound) || amount.greaterThan(upperBound)) {
                                outlierCount++;
                            }
                        }

                        if (valueMetric === TransactionExplorerValueMetric.SourceAmountOutlierCount) {
                            value = parseBigDecimal(outlierCount);
                        } else {
                            value = allSourceAmountsInDefaultCurrency.length > 0 ? parseBigDecimal(outlierCount).divide(allSourceAmountsInDefaultCurrency.length).multiply(100) : BIG_DECIMAL_ZERO;
                        }
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountVariance
                    || valueMetric === TransactionExplorerValueMetric.SourceAmountStandardDeviation
                    || valueMetric === TransactionExplorerValueMetric.SourceAmountCoefficientOfVariation
                    || valueMetric === TransactionExplorerValueMetric.SourceAmountSkewness
                    || valueMetric === TransactionExplorerValueMetric.SourceAmountKurtosis) {
                    if (allSourceAmountsInDefaultCurrency.length > 0) {
                        const averageSourceAmountInDefaultCurrency = totalSourceAmountSumInDefaultCurrency.divide(allSourceAmountsInDefaultCurrency.length).divide(AMOUNT_FACTOR);
                        const { variance, standardDeviation } = varianceAndStandardDeviation(allSourceAmountsInDefaultCurrency, averageSourceAmountInDefaultCurrency, item => item.divide(AMOUNT_FACTOR));

                        if (valueMetric === TransactionExplorerValueMetric.SourceAmountVariance) {
                            value = variance;
                        } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountStandardDeviation) {
                            value = standardDeviation;
                        } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountCoefficientOfVariation) {
                            value = coefficientOfVariation(standardDeviation, averageSourceAmountInDefaultCurrency) ?? BIG_DECIMAL_ZERO;
                        } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountSkewness) {
                            value = skewness(allSourceAmountsInDefaultCurrency, averageSourceAmountInDefaultCurrency, standardDeviation, item => item.divide(AMOUNT_FACTOR));
                        } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountKurtosis) {
                            value = kurtosis(allSourceAmountsInDefaultCurrency, averageSourceAmountInDefaultCurrency, variance, item => item.divide(AMOUNT_FACTOR));
                        }
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountGiniCoefficient) {
                    if (allSourceAmountsInDefaultCurrency.length > 0 && !totalSourceAmountSumInDefaultCurrency.isZero()) {
                        allSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
                        value = giniCoefficient(allSourceAmountsInDefaultCurrency, totalSourceAmountSumInDefaultCurrency, item => item);
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                } else if (valueMetric === TransactionExplorerValueMetric.SourceAmountHerfindahlHirschmanIndex) {
                    if (allSourceAmountsInDefaultCurrency.length > 0 && !totalSourceAmountSumInDefaultCurrency.isZero()) {
                        value = herfindahlHirschmanIndex(allSourceAmountsInDefaultCurrency, totalSourceAmountSumInDefaultCurrency, item => item);
                    } else {
                        value = BIG_DECIMAL_ZERO;
                    }
                }

                dataItems.push({
                    seriesName: seriesTransactions.seriesName,
                    seriesNameNeedI18n: seriesTransactions.seriesNameNeedI18n,
                    seriesNameI18nParameters: seriesTransactions.seriesNameI18nParameters,
                    seriesId: seriesTransactions.seriesId,
                    seriesIdType: seriesTransactions.seriesIdType,
                    seriesDisplayOrders: seriesTransactions.seriesDisplayOrders,
                    value: value
                });
            }

            result.push({
                categoryName: categoriedTransactions.categoryName,
                categoryNameNeedI18n: categoriedTransactions.categoryNameNeedI18n,
                categoryNameI18nParameters: categoriedTransactions.categoryNameI18nParameters,
                categoryId: categoriedTransactions.categoryId,
                categoryIdType: categoriedTransactions.categoryIdType,
                categoryDisplayOrders: categoriedTransactions.categoryDisplayOrders,
                data: dataItems
            });
        }

        return result;
    });

    return { categoriedTransactionExplorerData };
}
