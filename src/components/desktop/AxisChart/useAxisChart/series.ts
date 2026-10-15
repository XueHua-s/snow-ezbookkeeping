import type { SetupProps } from '@/lib/ui/props.ts';
import { itemAndIndex } from '@/core/base.ts';
import { type AxisChartSourceDataItem, ChartValueType } from '@/core/chart.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { getDisplayColor } from '@/lib/color.ts';
import { isArray } from '@/lib/common.ts';
import { BIG_DECIMAL_NEGATIVE_INFINITY, BIG_DECIMAL_POSITIVE_INFINITY, BIG_DECIMAL_ZERO } from '@/lib/numeral.ts';
import { computed } from 'vue';
import type { useAxisChartPresentation } from './presentation.ts';
import { type AxisChartData, type AxisChartDataItem, type AxisChartDisplayType } from './types.ts';

interface Context {
    props: SetupProps<{
    class?: string;
    skeleton?: boolean;
    noAnimation?: boolean;
    type: AxisChartDisplayType;
    stacked?: boolean;
    hideLegend?: boolean;
    legendPosition?: 'top' | 'bottom';
    hideXAxisLabels?: boolean;
    hideXAxisLine?: boolean;
    hideYAxisLabels?: boolean;
    hideHorizontalGridLines?: boolean;
    hideLineSymbols?: boolean;
    noMargin?: boolean;
    smoothCurve?: boolean;
    oneHundredPercentStacked?: boolean;
    sortingType: number;
    showValue?: boolean;
    showTotalAmountInTooltip?: boolean;
    totalNameInTooltip?: string;
    categoryTypeName: string;
    allCategoryNames: string[];
    items: AxisChartSourceDataItem[];
    valueType: ChartValueType;
    translateName?: boolean;
    defaultCurrency?: string;
    useCustomColor?: boolean;
    enableClickItem?: boolean;
    tooltipExtraColumnNames?: string[];
    tooltipExtraColumnTotalValues?: (categoryIndex: number, totalValue: BigDecimal, visibleSeriesIds: string[]) => string[];
    tooltipExtraColumnValues?: (seriesId: string, categoryIndex: number, currentValue: BigDecimal) => string[];
}>;
    getItemName: ReturnType<typeof useAxisChartPresentation>['getItemName'];
    chartColors: ReturnType<typeof useAxisChartPresentation>['chartColors'];
    getDisplayValue: ReturnType<typeof useAxisChartPresentation>['getDisplayValue'];
}

export function useAxisChartSeries(context: Context) {
    const { props, getItemName, chartColors, getDisplayValue } = context;

    const axisChartData = computed<AxisChartData>(() => {
        const allSeries: AxisChartDataItem[] = [];
        const allOriginalData: BigDecimal[][] = [];
        const categoryTotalAmount: Record<number, BigDecimal> = {};
        let maxAmountOfAllData: BigDecimal = BIG_DECIMAL_ZERO;

        for (const item of props.items) {
            if (item.hidden) {
                continue;
            }

            if (!isArray(item.values)) {
                continue;
            }

            const allAmounts: BigDecimal[] = item.values;

            for (const [amount, categoryIndex] of itemAndIndex(allAmounts)) {
                let totalAmount: BigDecimal = categoryTotalAmount[categoryIndex] ?? BIG_DECIMAL_ZERO;
                totalAmount = totalAmount.add(amount);
                categoryTotalAmount[categoryIndex] = totalAmount;

                if (amount.greaterThan(maxAmountOfAllData)) {
                    maxAmountOfAllData = amount;
                }
            }
        }

        for (const item of props.items) {
            if (item.hidden) {
                continue;
            }

            if (!isArray(item.values)) {
                continue;
            }

            const allAmounts: BigDecimal[] = item.values;

            if (props.oneHundredPercentStacked) {
                for (const [amount, categoryIndex] of itemAndIndex(allAmounts)) {
                    const totalAmount: BigDecimal = categoryTotalAmount[categoryIndex] ?? BIG_DECIMAL_ZERO;
                    allAmounts[categoryIndex] = !totalAmount.isZero() ? amount.divide(totalAmount).multiply(100) : BIG_DECIMAL_ZERO;
                }
            }

            const finalItem: AxisChartDataItem = {
                id: item.id ?? getItemName(item.name),
                name: item.id ?? getItemName(item.name),
                itemStyle: {
                    color: getDisplayColor(props.useCustomColor && item.color ? item.color : chartColors.value[allSeries.length % chartColors.value.length]),
                },
                selected: true,
                type: 'line',
                smooth: props.smoothCurve,
                showSymbol: !props.hideLineSymbols,
                animation: props.noAnimation ? false : !props.skeleton,
                data: allAmounts.map(amount => amount.toDoubleNumber())
            };

            if (props.stacked) {
                finalItem.stack = 'a';
            } else if (item.id) {
                finalItem.stack = item.id;
            }

            if (props.type === 'line') {
                finalItem.areaStyle = undefined;
            } else if (props.type === 'area') {
                finalItem.areaStyle = {};
            } else if (props.type === 'column') {
                finalItem.type = 'bar';
            } else if (props.type === 'bubble') {
                finalItem.type = 'scatter';
                finalItem.symbolSize = (data: number): number => {
                    return Math.sqrt(data) / Math.sqrt(maxAmountOfAllData.toDoubleNumber()) * 80 + 5;
                }
            }

            allSeries.push(finalItem);
            allOriginalData.push(allAmounts);
        }

        return {
            allSeries: allSeries,
            allOriginalData: allOriginalData
        };
    });

    const yAxisWidth = computed<number>(() => {
        let maxValue: BigDecimal = BIG_DECIMAL_NEGATIVE_INFINITY;
        let minValue: BigDecimal = BIG_DECIMAL_POSITIVE_INFINITY;
        let width = 90;

        if (!axisChartData.value || !axisChartData.value.allOriginalData || !axisChartData.value.allOriginalData.length) {
            return width;
        }

        for (const seriesData of axisChartData.value.allOriginalData) {
            for (const value of seriesData) {
                if (value.greaterThan(maxValue)) {
                    maxValue = value;
                }

                if (value.lessThan(minValue)) {
                    minValue = value;
                }
            }
        }

        const maxValueText = getDisplayValue(maxValue);
        const minValueText = getDisplayValue(minValue);
        const maxLengthText = maxValueText.length > minValueText.length ? maxValueText : minValueText;

        const canvas = document.createElement('canvas');
        const context = canvas.getContext('2d');

        if (context) {
            context.font = '12px Arial';

            const textMetrics = context.measureText(maxLengthText);
            const actualWidth = Math.round(textMetrics.width) + 20;

            if (actualWidth >= 200) {
                width = 200;
            } if (actualWidth > 90) {
                width = actualWidth;
            }
        }

        return width;
    });

    return { axisChartData, yAxisWidth };
}
