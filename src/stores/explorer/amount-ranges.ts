import { AMOUNT_FACTOR } from '@/consts/numeral.ts';
import {
    TransactionExplorerDataDimension
} from '@/core/explorer.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import {
    max,
    mean,
    min,
    varianceAndStandardDeviation
} from '@/lib/math.ts';
import {
    BIG_DECIMAL_POSITIVE_INFINITY,
    BIG_DECIMAL_ZERO
} from '@/lib/numeral.ts';
import type { AmountRanges } from './types.ts';


// Owns amount ranges for the explorer store.
export function useExplorerAmountRanges() {
    function calculateAmountRanges(sortedAmounts: BigDecimal[], dimension: TransactionExplorerDataDimension, rangeCount: number): BigDecimal[] {
        const result: BigDecimal[] = [];

        if (sortedAmounts.length < 1 || rangeCount <= 0) {
            return result;
        }

        const minAmount = sortedAmounts[0] as BigDecimal;
        const maxAmount = sortedAmounts[sortedAmounts.length - 1] as BigDecimal;
        rangeCount = Math.min(rangeCount, sortedAmounts.length);

        // [min1, max1), [min2, max2), ..., [minN, maxN]
        if (dimension === TransactionExplorerDataDimension.SourceAmountRangeEqualFrequency
            || dimension === TransactionExplorerDataDimension.DestinationAmountRangeEqualFrequency) {
            for (let i = 0; i < rangeCount; i++) {
                result.push(sortedAmounts[Math.floor(i * (sortedAmounts.length - 1) / rangeCount)] as BigDecimal);
            }
            result.push(maxAmount);
        } else if (dimension === TransactionExplorerDataDimension.SourceAmountRangeEqualWidth
            || dimension === TransactionExplorerDataDimension.DestinationAmountRangeEqualWidth) {
            if (minAmount.equals(maxAmount)) {
                return [minAmount, maxAmount];
            }

            const width: BigDecimal = maxAmount.subtract(minAmount).divide(rangeCount); // (maxAmount - minAmount) / rangeCount

            for (let i = 0; i < rangeCount; i++) {
                result.push(minAmount.add(width.multiply(i))); // minAmount + i * width
            }
            result.push(maxAmount);
        } else if (dimension === TransactionExplorerDataDimension.SourceAmountRangeLogScale
            || dimension === TransactionExplorerDataDimension.DestinationAmountRangeLogScale) {
            const epsilon: number = 1e-9;

            const transform = (x: BigDecimal): BigDecimal => {
                if (x.isZero()) {
                    return x;
                }

                return x.sign().multiply(x.abs().add(epsilon).log()); // sign(x) * log(abs(x) + epsilon)
            };

            const inverse = (y: BigDecimal): BigDecimal => {
                if (y.isZero()) {
                    return y;
                }

                return y.sign().multiply(y.abs().exp().subtract(epsilon)); // sign(y) * (exp(abs(y)) - epsilon)
            };

            const transformed: BigDecimal[] = sortedAmounts.map(transform).sort((a, b) => a.compareTo(b));

            const tMin: BigDecimal = transformed[0] as BigDecimal;
            const tMax: BigDecimal = transformed[transformed.length - 1] as BigDecimal;

            if (tMin.equals(tMax)) {
                return [minAmount, maxAmount];
            }

            const width: BigDecimal = tMax.subtract(tMin).divide(rangeCount); // (tMax - tMin) / rangeCount

            result.push(minAmount);
            for (let i = 1; i < rangeCount; i++) {
                result.push(inverse(tMin.add(width.multiply(i)))); // inverse(tMin + i * width)
            }
            result.push(maxAmount);
        } else if (dimension === TransactionExplorerDataDimension.SourceAmountRangeStandardDeviation
            || dimension === TransactionExplorerDataDimension.DestinationAmountRangeStandardDeviation) {
            if (minAmount.equals(maxAmount)) {
                return [minAmount, maxAmount];
            }

            const averageAmountForVarianceCalculation: BigDecimal = mean(sortedAmounts, item => item).divide(AMOUNT_FACTOR);
            const { standardDeviation } = varianceAndStandardDeviation(sortedAmounts, averageAmountForVarianceCalculation, item => item.divide(AMOUNT_FACTOR));

            if (standardDeviation.isZero()) {
                return [minAmount, maxAmount];
            }

            const rawBreaks: BigDecimal[] = [];
            const halfCount = Math.floor(rangeCount / 2);

            if (rangeCount % 2 === 1) {
                for (let i = -halfCount; i <= halfCount; i++) {
                    rawBreaks.push(averageAmountForVarianceCalculation.add(standardDeviation.multiply(i)).multiply(AMOUNT_FACTOR));
                }
            } else {
                for (let i = -halfCount; i <= halfCount; i++) {
                    if (i === 0) {
                        continue;
                    }
                    rawBreaks.push(averageAmountForVarianceCalculation.add(standardDeviation.multiply(i - 0.5)).multiply(AMOUNT_FACTOR));
                }
                rawBreaks.sort((a, b) => a.compareTo(b));
            }

            const clipped = rawBreaks.map((v) => max(minAmount, min(maxAmount, v)))
                .filter((v, i, arr) => i === 0 || v.notEquals(arr[i - 1]));

            clipped[0] = minAmount;

            if (clipped[clipped.length - 1]?.notEquals(maxAmount)) {
                clipped.push(maxAmount);
            }

            return clipped;
        } else if (dimension === TransactionExplorerDataDimension.SourceAmountRangeNaturalBreaks
            || dimension === TransactionExplorerDataDimension.DestinationAmountRangeNaturalBreaks) {
            if (minAmount.equals(maxAmount)) {
                return [minAmount, maxAmount];
            }

            const n = sortedAmounts.length;
            const k = Math.min(rangeCount, n);

            const lowerClassLimits: number[][] = Array.from({ length: n + 1 }, () => new Array(k + 1).fill(0));
            const varianceCombinations: BigDecimal[][] = Array.from({ length: n + 1 }, () => new Array(k + 1).fill(BIG_DECIMAL_POSITIVE_INFINITY));

            for (let i = 1; i <= k; i++) {
                lowerClassLimits[1]![i] = 1;
                varianceCombinations[1]![i] = BIG_DECIMAL_ZERO;
            }

            for (let l = 2; l <= n; l++) {
                let sumZ: BigDecimal = BIG_DECIMAL_ZERO;
                let sumZ2: BigDecimal = BIG_DECIMAL_ZERO;

                for (let m = 1; m <= l; m++) {
                    const val = sortedAmounts[l - m] as BigDecimal;
                    sumZ = sumZ.add(val);
                    sumZ2 = sumZ2.add(val.pow(2));

                    const variance = sumZ2.subtract(sumZ.pow(2).divide(m));


                    if (m === l) {
                        for (let j = 1; j <= k; j++) {
                            if (variance.lessThan(varianceCombinations[l]![j]!)) {
                                lowerClassLimits[l]![j] = 1;
                                varianceCombinations[l]![j] = variance;
                            }
                        }
                    } else {
                        for (let j = 2; j <= k; j++) {
                            const combined = varianceCombinations[l - m]![j - 1]!.add(variance);
                            if (combined.lessThan(varianceCombinations[l]![j]!)) {
                                lowerClassLimits[l]![j] = l - m + 1;
                                varianceCombinations[l]![j] = combined;
                            }
                        }
                    }
                }
            }

            const breaks: BigDecimal[] = new Array(k + 1);
            breaks[k] = maxAmount;

            let currentK = k;
            let currentIdx = n;

            while (currentK >= 2) {
                const lowerIdx = lowerClassLimits[currentIdx]![currentK]!;
                breaks[currentK - 1] = sortedAmounts[lowerIdx - 1] as BigDecimal;
                currentIdx = lowerIdx - 1;
                currentK--;
            }

            breaks[0] = minAmount;
            return breaks;
        }

        return result;
    }

    function buildAllAmountRanges(categoryDimension: TransactionExplorerDataDimension, seriesDimension: TransactionExplorerDataDimension, filteredTransactionSourceAmountsInDefaultCurrency: BigDecimal[], filteredTransactionDestinationAmountsInDefaultCurrency: BigDecimal[], rangeCount: number): AmountRanges {
        const allAmountRanges: AmountRanges = {};

        if (categoryDimension.isSourceAmountRange || seriesDimension.isSourceAmountRange) {
            filteredTransactionSourceAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
            const sorteUniqueAmounts = filteredTransactionSourceAmountsInDefaultCurrency.filter((v, i, a) => i === 0 || v.notEquals(a[i - 1]));

            if (categoryDimension.isSourceAmountRange) {
                allAmountRanges.categorySourceAmountRanges = calculateAmountRanges(sorteUniqueAmounts, categoryDimension, rangeCount);
            }

            if (seriesDimension.isSourceAmountRange) {
                allAmountRanges.seriesSourceAmountRanges = calculateAmountRanges(sorteUniqueAmounts, seriesDimension, rangeCount);
            }
        }

        if (categoryDimension.isDestinationAmountRange || seriesDimension.isDestinationAmountRange) {
            filteredTransactionDestinationAmountsInDefaultCurrency.sort((a, b) => a.compareTo(b));
            const sorteUniqueAmounts = filteredTransactionDestinationAmountsInDefaultCurrency.filter((v, i, a) => i === 0 || v.notEquals(a[i - 1]));
            if (categoryDimension.isDestinationAmountRange) {
                allAmountRanges.categoryDestinationAmountRanges = calculateAmountRanges(sorteUniqueAmounts, categoryDimension, rangeCount);
            }

            if (seriesDimension.isDestinationAmountRange) {
                allAmountRanges.seriesDestinationAmountRanges = calculateAmountRanges(sorteUniqueAmounts, seriesDimension, rangeCount);
            }
        }

        return allAmountRanges;
    }

    return { buildAllAmountRanges };
}
