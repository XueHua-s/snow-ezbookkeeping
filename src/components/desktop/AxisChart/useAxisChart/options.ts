import type { SetupProps } from '@/lib/ui/props.ts';
import { itemAndIndex } from '@/core/base.ts';
import { type AxisChartSourceDataItem, ChartValueType } from '@/core/chart.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TextDirection } from '@/core/text.ts';
import { isNumber } from '@/lib/common.ts';
import { BIG_DECIMAL_ZERO, parseBigDecimal } from '@/lib/numeral.ts';
import { sortStatisticsItems } from '@/lib/statistics.ts';
import type { CallbackDataParams } from 'echarts/types/dist/shared';
import { computed } from 'vue';
import type { useAxisChartPresentation } from './presentation.ts';
import type { useAxisChartSeries } from './series.ts';
import { type AxisChartDisplayType, type AxisChartTooltipItem } from './types.ts';

interface Context {
    isDarkMode: ReturnType<typeof useAxisChartPresentation>['isDarkMode'];
    axisChartData: ReturnType<typeof useAxisChartSeries>['axisChartData'];
    allItemsMap: ReturnType<typeof useAxisChartPresentation>['allItemsMap'];
    getItemName: ReturnType<typeof useAxisChartPresentation>['getItemName'];
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
    getDisplayValue: ReturnType<typeof useAxisChartPresentation>['getDisplayValue'];
    selectedLegends: ReturnType<typeof useAxisChartPresentation>['selectedLegends'];
    yAxisWidth: ReturnType<typeof useAxisChartSeries>['yAxisWidth'];
    textDirection: ReturnType<typeof useAxisChartPresentation>['textDirection'];
}

export function useAxisChartOptions(context: Context) {
    const { isDarkMode, axisChartData, allItemsMap, getItemName, props, getDisplayValue, selectedLegends, yAxisWidth, textDirection } = context;

    const chartOptions = computed<object>(() => {
        return {
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'cross',
                    label: {
                        backgroundColor: isDarkMode.value ? '#333' : '#fff',
                        color: isDarkMode.value ? '#eee' : '#333'
                    },
                },
                backgroundColor: isDarkMode.value ? '#333' : '#fff',
                borderColor: isDarkMode.value ? '#333' : '#fff',
                textStyle: {
                    color: isDarkMode.value ? '#eee' : '#333'
                },
                formatter: (params: CallbackDataParams[]) => {
                    let tooltip = '';
                    let totalAmount: BigDecimal = BIG_DECIMAL_ZERO;
                    let actualDisplayItemCount = 0;
                    const displayItems: AxisChartTooltipItem[] = [];
                    const categoryIndex = params.length > 0 && params[0] ? (params[0].dataIndex ?? 0) : 0;
                    const visibleSeriesIds: string[] = [];

                    for (const param of params) {
                        if (!isNumber(param.seriesIndex)) {
                            continue;
                        }

                        const seriesIndex = param.seriesIndex;
                        const seriesData = axisChartData.value.allSeries[seriesIndex];

                        if (!seriesData) {
                            continue;
                        }

                        const seriesId = seriesData.id;
                        const name = allItemsMap.value[seriesId] ? getItemName(allItemsMap.value[seriesId].name) : seriesId;
                        const color = seriesData.itemStyle.color;
                        const displayOrders = allItemsMap.value[seriesId]?.displayOrders ?? [0];
                        const amount: BigDecimal = axisChartData.value.allOriginalData[seriesIndex]?.[categoryIndex] ?? BIG_DECIMAL_ZERO;

                        displayItems.push({
                            id: seriesId,
                            name: name,
                            color: color,
                            displayOrders: displayOrders,
                            value: amount
                        });

                        visibleSeriesIds.push(seriesId);
                        totalAmount = totalAmount.add(amount);
                    }

                    sortStatisticsItems(displayItems, props.sortingType);

                    const extraColumnValuesMap: Record<number, string[]> = {};
                    const extraColumnTotalValues: string[] = [];
                    const hasExtraColumnIndexes: Record<number, boolean> = {};

                    if (props.tooltipExtraColumnNames) {
                        if (props.tooltipExtraColumnValues) {
                            for (const [item, index] of itemAndIndex(displayItems)) {
                                const values = props.tooltipExtraColumnValues(item.id, categoryIndex, item.value);
                                extraColumnValuesMap[index] = values;

                                for (const [value, columnIndex] of itemAndIndex(values)) {
                                    if (value && value !== '-') {
                                        hasExtraColumnIndexes[columnIndex] = true;
                                    }
                                }
                            }
                        }

                        if (props.tooltipExtraColumnTotalValues) {
                            const values = props.tooltipExtraColumnTotalValues(categoryIndex, totalAmount, visibleSeriesIds);
                            extraColumnTotalValues.push(...values);

                            for (const [value, columnIndex] of itemAndIndex(values)) {
                                if (value && value !== '-') {
                                    hasExtraColumnIndexes[columnIndex] = true;
                                }
                            }
                        }
                    }

                    for (const [item, index] of itemAndIndex(displayItems)) {
                        if (displayItems.length === 1 || !item.value.isZero()) {
                            const value = getDisplayValue(item.value);
                            tooltip += '<tr><td><span class="chart-pointer" style="background-color: ' + item.color + '"></span>';
                            tooltip += `<span>${item.name}</span></td><td><span class="ms-5" style="float: inline-end">${value}</span></td>`;

                            if (props.tooltipExtraColumnNames) {
                                const values = extraColumnValuesMap[index] ?? [];

                                for (let i = 0; i < props.tooltipExtraColumnNames.length; i++) {
                                    if (!hasExtraColumnIndexes[i]) {
                                        continue;
                                    }

                                    const value = values[i] ?? '-';
                                    tooltip += `<td><span class="ms-5" style="float: inline-end">${value}</span></td>`;
                                }
                            }

                            tooltip += '</tr>';
                            actualDisplayItemCount++;
                        }
                    }

                    if (props.showTotalAmountInTooltip && !props.oneHundredPercentStacked) {
                        const displayTotalAmount = getDisplayValue(totalAmount);
                        let totalColumnCount = 2;

                        let totalTooltip = `<tr><td><span class="chart-pointer" style="background-color: ${isDarkMode.value ? '#eee' : '#333'}"></span>`
                            + `<span>${props.totalNameInTooltip}</span></td><td><span class="ms-5" style="float: inline-end">${displayTotalAmount}</span></td>`;

                        if (props.tooltipExtraColumnNames) {
                            for (let i = 0; i < props.tooltipExtraColumnNames.length; i++) {
                                if (!hasExtraColumnIndexes[i]) {
                                    continue;
                                }

                                const value = extraColumnTotalValues[i] ?? '-';
                                totalTooltip += `<td><span class="ms-5" style="float: inline-end">${value}</span></td>`;
                                totalColumnCount++;
                            }
                        }

                        totalTooltip += '</tr>';
                        totalTooltip += `<tr><td colspan="${totalColumnCount}" ${actualDisplayItemCount > 0 ? 'style="border-bottom: ' + (isDarkMode.value ? '#eee' : '#333') + ' dashed 1px"' : ''}></td></tr>`;
                        tooltip = totalTooltip + tooltip;
                    }

                    if (params.length && params[0] && params[0].name) {
                        let tooltipHeader = `<td>${params[0].name}</td><td></td>`;

                        if (props.tooltipExtraColumnNames) {
                            for (const [columnName, columnIndex] of itemAndIndex(props.tooltipExtraColumnNames)) {
                                if (!hasExtraColumnIndexes[columnIndex]) {
                                    continue;
                                }

                                tooltipHeader += `<td><span class="ms-5" style="float: inline-end">${columnName}</span></td>`;
                            }
                        }

                        tooltip = `<table class="chart-tooltip-table"><tbody><tr>${tooltipHeader}</tr>${tooltip}</tbody></table>`
                    }

                    return tooltip;
                }
            },
            legend: {
                show: !props.hideLegend,
                orient: 'horizontal',
                type: 'scroll',
                top: (!props.legendPosition || props.legendPosition === 'top') ? 0 : undefined,
                bottom: props.legendPosition === 'bottom' ? 5 : undefined,
                data: axisChartData.value.allSeries.map(item => item.name),
                selected: selectedLegends.value,
                textStyle: {
                    color: isDarkMode.value ? '#eee' : '#333'
                },
                formatter: (id: string) => allItemsMap.value[id] ? getItemName(allItemsMap.value[id].name) : id
            },
            grid: {
                left: props.noMargin ? 0 : (props.hideYAxisLabels ? 10 : yAxisWidth.value),
                top: props.noMargin ? 5 : (!props.hideLegend && (!props.legendPosition || props.legendPosition === 'top') ? 50 : 5),
                right: props.noMargin ? 0 : 10,
                bottom: props.noMargin ? 0 : ((props.hideXAxisLabels ? 10 : 30) + (!props.hideLegend && props.legendPosition === 'bottom' ? 25 : 0)),
            },
            xAxis: [
                {
                    type: 'category',
                    data: props.allCategoryNames,
                    boundaryGap: !props.hideYAxisLabels,
                    inverse: textDirection.value === TextDirection.RTL,
                    axisLine: {
                        show: !props.hideXAxisLine
                    },
                    axisTick: {
                        show: !props.hideXAxisLine
                    },
                    axisLabel: {
                        show: !props.hideXAxisLabels,
                        color: isDarkMode.value ? '#888' : '#666'
                    }
                }
            ],
            yAxis: [
                {
                    type: 'value',
                    min: props.oneHundredPercentStacked ? 0 : undefined,
                    max: props.oneHundredPercentStacked ? 100 : undefined,
                    axisLabel: {
                        show: !props.hideYAxisLabels && (props.showValue !== false || props.valueType !== ChartValueType.Amount),
                        color: isDarkMode.value ? '#888' : '#666',
                        formatter: (value: number) => {
                            return getDisplayValue(parseBigDecimal(value));
                        }
                    },
                    axisPointer: {
                        label: {
                            show: props.showValue !== false || props.valueType !== ChartValueType.Amount,
                            formatter: (params: CallbackDataParams) => {
                                return getDisplayValue(parseBigDecimal(params.value as number).truncate());
                            }
                        }
                    },
                    splitLine: {
                        show: !props.hideHorizontalGridLines,
                        lineStyle: {
                            color: isDarkMode.value ? '#4f4f4f' : '#e1e6f2',
                        }
                    }
                }
            ],
            series: axisChartData.value.allSeries
        };
    });

    return { chartOptions };
}
