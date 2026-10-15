import type { SetupProps } from '@/lib/ui/props.ts';
import { itemAndIndex } from '@/core/base.ts';
import { type AxisChartSourceDataItem, ChartValueType } from '@/core/chart.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { BIG_DECIMAL_ZERO } from '@/lib/numeral.ts';
import type { ECElementEvent } from 'echarts/core';
import type { useAxisChartPresentation } from './presentation.ts';
import type { useAxisChartSeries } from './series.ts';
import { type AxisChartDisplayType } from './types.ts';

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
    allItemsMap: ReturnType<typeof useAxisChartPresentation>['allItemsMap'];
    emit: {
    (e: 'click', itemId: string, categoryIndex: number, item: AxisChartSourceDataItem): void;
};
    axisChartData: ReturnType<typeof useAxisChartSeries>['axisChartData'];
    getItemName: ReturnType<typeof useAxisChartPresentation>['getItemName'];
    formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping: ReturnType<typeof useAxisChartPresentation>['formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping'];
    formatAmountToWesternArabicNumeralsWithoutDigitGrouping: ReturnType<typeof useAxisChartPresentation>['formatAmountToWesternArabicNumeralsWithoutDigitGrouping'];
    selectedLegends: ReturnType<typeof useAxisChartPresentation>['selectedLegends'];
}

export function useAxisChartInteraction(context: Context) {
    const {
        props, allItemsMap, emit, axisChartData, getItemName, formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, selectedLegends,
    } = context;

    function clickItem(e: ECElementEvent): void {
        if (!props.enableClickItem || e.componentType !== 'series') {
            return;
        }

        const id = e.seriesId as string;
        const item = allItemsMap.value[id] as AxisChartSourceDataItem;
        const itemId = item?.id ?? '';
        const category = props.allCategoryNames[e.dataIndex];

        if (!item || !category) {
            return;
        }

        emit('click', itemId, e.dataIndex, item);
    }

    function exportData(): { headers: string[], data: string[][] } {
        const headers: string[] = [];
        const data: string[][] = [];

        headers.push(props.categoryTypeName);

        for (const series of axisChartData.value.allSeries) {
            const id = series.id;
            const name = allItemsMap.value[id] ? getItemName(allItemsMap.value[id].name) : id;
            headers.push(name);
        }

        for (const [categoryName, index] of itemAndIndex(props.allCategoryNames)) {
            const row: string[] = [];
            row.push(categoryName);
            row.push(...axisChartData.value.allOriginalData.map(item => {
                if (props.oneHundredPercentStacked) {
                    return formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping(item[index] ?? BIG_DECIMAL_ZERO);
                } else if (props.valueType === ChartValueType.Amount) {
                    return formatAmountToWesternArabicNumeralsWithoutDigitGrouping(item[index] ?? BIG_DECIMAL_ZERO, props.defaultCurrency);
                } else {
                    return formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping(item[index] ?? BIG_DECIMAL_ZERO);
                }
            }));
            data.push(row);
        }

        return {
            headers: headers,
            data: data
        };
    }

    function onLegendSelectChanged(e: { selected: Record<string, boolean> }): void {
        selectedLegends.value = e.selected;
    }

    return { clickItem, exportData, onLegendSelectChanged };
}
