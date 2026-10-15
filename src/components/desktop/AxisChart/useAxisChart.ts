import type { SetupProps } from '@/lib/ui/props.ts';
import { type AxisChartSourceDataItem, ChartValueType } from '@/core/chart.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { useAxisChartInteraction } from './useAxisChart/interaction.ts';
import { useAxisChartOptions } from './useAxisChart/options.ts';
import { useAxisChartPresentation } from './useAxisChart/presentation.ts';
import { useAxisChartSeries } from './useAxisChart/series.ts';
import { type AxisChartDisplayType } from './useAxisChart/types.ts';
export * from './useAxisChart/types.ts';

// Coordinates AxisChart state and effects; called once from the component setup.
export function useAxisChart(props: SetupProps<{
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
}>, emit: {
    (e: 'click', itemId: string, categoryIndex: number, item: AxisChartSourceDataItem): void;
}) {
    const {
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        selectedLegends, textDirection, isDarkMode, chartColors, finalClass, allItemsMap, getItemName, getDisplayValue,
    } = useAxisChartPresentation({ props });
    const { axisChartData, yAxisWidth } = useAxisChartSeries({ props, getItemName, chartColors, getDisplayValue });
    const { chartOptions } = useAxisChartOptions({ isDarkMode, axisChartData, allItemsMap, getItemName, props, getDisplayValue, selectedLegends, yAxisWidth, textDirection });
    const { clickItem, exportData, onLegendSelectChanged } = useAxisChartInteraction({
        props, allItemsMap, emit, axisChartData, getItemName, formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, selectedLegends,
    });

    return { finalClass, chartOptions, clickItem, exportData, onLegendSelectChanged };
}
