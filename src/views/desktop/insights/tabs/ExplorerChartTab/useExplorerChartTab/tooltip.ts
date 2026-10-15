import { type BigDecimal } from '@/core/numeral.ts';
import { isNumber } from '@/lib/common.ts';
import { BIG_DECIMAL_ZERO } from '@/lib/numeral.ts';
import type { useExplorerChartTabChartOptions } from './chart-options.ts';
import type { useExplorerChartTabSeries } from './series.ts';
import type { useExplorerChartTabState } from './state.ts';

interface Context {
    formatPercentToLocalizedNumerals: ReturnType<typeof useExplorerChartTabState>['formatPercentToLocalizedNumerals'];
    axisChartShowYearOverYear: ReturnType<typeof useExplorerChartTabChartOptions>['axisChartShowYearOverYear'];
    axisChartShowPeriodOverPeriod: ReturnType<typeof useExplorerChartTabChartOptions>['axisChartShowPeriodOverPeriod'];
    axisChartCategoryIndexYoYMap: ReturnType<typeof useExplorerChartTabSeries>['axisChartCategoryIndexYoYMap'];
    seriesDimensionTransactionExplorerDataMap: ReturnType<typeof useExplorerChartTabSeries>['seriesDimensionTransactionExplorerDataMap'];
}

export function useExplorerChartTabTooltip(context: Context) {
    const {
        formatPercentToLocalizedNumerals, axisChartShowYearOverYear, axisChartShowPeriodOverPeriod, axisChartCategoryIndexYoYMap,
        seriesDimensionTransactionExplorerDataMap,
    } = context;

    function formatDisplayChangeRate(current: BigDecimal, reference: BigDecimal): string {
        if (reference.isZero() && current.isZero()) {
            return formatPercentToLocalizedNumerals(0, 2, '<0.01');
        }

        if (reference.isZero()) {
            return '-';
        }

        const rate = current.subtract(reference).divide(reference).multiply(100).toDoubleNumber();
        return formatPercentToLocalizedNumerals(rate, 2, '<0.01');
    }

    function getAxisChartTooltipExtraColumnTotalValues(categoryIndex: number, totalValue: BigDecimal, visibleSeriesIds: string[]): string[] {
        const extraColumnValues: string[] = [];

        if (!axisChartShowYearOverYear.value && !axisChartShowPeriodOverPeriod.value) {
            return extraColumnValues;
        }

        if (axisChartShowYearOverYear.value) {
            const yoyReferenceIndex = axisChartCategoryIndexYoYMap.value[categoryIndex];
            let displayChangeRate = '-';

            if (isNumber(yoyReferenceIndex)) {
                let referenceTotalValue: BigDecimal = BIG_DECIMAL_ZERO;

                for (const seriesId of visibleSeriesIds) {
                    const seriesDimensionData = seriesDimensionTransactionExplorerDataMap.value[seriesId];

                    if (seriesDimensionData && seriesDimensionData.values) {
                        referenceTotalValue = referenceTotalValue.add(seriesDimensionData.values[yoyReferenceIndex] ?? BIG_DECIMAL_ZERO);
                    }
                }

                displayChangeRate = formatDisplayChangeRate(totalValue, referenceTotalValue);
            }

            extraColumnValues.push(displayChangeRate);
        }

        if (axisChartShowPeriodOverPeriod.value) {
            const popReferenceIndex = categoryIndex - 1;
            let displayChangeRate = '-';

            if (popReferenceIndex >= 0) {
                let referenceTotalValue: BigDecimal = BIG_DECIMAL_ZERO;

                for (const seriesId of visibleSeriesIds) {
                    const seriesDimensionData = seriesDimensionTransactionExplorerDataMap.value[seriesId];

                    if (seriesDimensionData && seriesDimensionData.values) {
                        referenceTotalValue = referenceTotalValue.add(seriesDimensionData.values[popReferenceIndex] ?? BIG_DECIMAL_ZERO);
                    }
                }

                displayChangeRate = formatDisplayChangeRate(totalValue, referenceTotalValue);
            }

            extraColumnValues.push(displayChangeRate);
        }

        return extraColumnValues;
    }

    function getAxisChartTooltipExtraColumnValues(seriesId: string, categoryIndex: number, currentValue: BigDecimal): string[] {
        const extraColumnValues: string[] = [];

        if (!axisChartShowYearOverYear.value && !axisChartShowPeriodOverPeriod.value) {
            return extraColumnValues;
        }

        const seriesDimensionData = seriesDimensionTransactionExplorerDataMap.value[seriesId];

        if (!seriesDimensionData || !seriesDimensionData.values) {
            return extraColumnValues;
        }

        const values = seriesDimensionData.values;

        if (axisChartShowYearOverYear.value) {
            const yoyReferenceIndex = axisChartCategoryIndexYoYMap.value[categoryIndex];
            let displayChangeRate = '-';

            if (isNumber(yoyReferenceIndex) && yoyReferenceIndex >= 0 && yoyReferenceIndex < values.length) {
                displayChangeRate = formatDisplayChangeRate(currentValue, values[yoyReferenceIndex] ?? BIG_DECIMAL_ZERO);
            }

            extraColumnValues.push(displayChangeRate);
        }

        if (axisChartShowPeriodOverPeriod.value) {
            const popReferenceIndex = categoryIndex - 1;
            let displayChangeRate = '-';

            if (popReferenceIndex >= 0 && popReferenceIndex < values.length) {
                displayChangeRate = formatDisplayChangeRate(currentValue, values[popReferenceIndex] ?? BIG_DECIMAL_ZERO);
            }

            extraColumnValues.push(displayChangeRate);
        }

        return extraColumnValues;
    }

    return { getAxisChartTooltipExtraColumnTotalValues, getAxisChartTooltipExtraColumnValues };
}
