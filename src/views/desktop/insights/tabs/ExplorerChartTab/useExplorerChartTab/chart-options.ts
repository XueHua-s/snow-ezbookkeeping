import { type AxisChartDisplayType } from '@/components/desktop/AxisChart.vue';
import { type HierarchyChartDisplayType } from '@/components/desktop/HierarchyChart.vue';
import {
    TransactionExplorerChartType,
    TransactionExplorerDataDimension
} from '@/core/explorer.ts';
import { computed } from 'vue';
import type { useExplorerChartTabState } from './state.ts';

interface Context {
    currentExploration: ReturnType<typeof useExplorerChartTabState>['currentExploration'];
    tt: ReturnType<typeof useExplorerChartTabState>['tt'];
}

export function useExplorerChartTabChartOptions(context: Context) {
    const { currentExploration, tt } = context;

    const axisChartDisplayType = computed<AxisChartDisplayType | undefined>(() => {
        if (currentExploration.value.chartType === TransactionExplorerChartType.ColumnStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.Column100PercentStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.ColumnGrouped.value) {
            return 'column';
        } else if (currentExploration.value.chartType === TransactionExplorerChartType.LineGrouped.value
            || currentExploration.value.chartType === TransactionExplorerChartType.SmoothLineGrouped.value) {
            return 'line';
        } else if (currentExploration.value.chartType === TransactionExplorerChartType.AreaStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.Area100PercentStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.SmoothAreaStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.SmoothArea100PercentStacked.value) {
            return 'area';
        } else if (currentExploration.value.chartType === TransactionExplorerChartType.BubbleGrouped.value) {
            return 'bubble';
        } else {
            return undefined;
        }
    });

    const hierarchyChartDisplayType = computed<HierarchyChartDisplayType | undefined>(() => {
        if (currentExploration.value.chartType === TransactionExplorerChartType.Treemap.value) {
            return 'treemap';
        } else if (currentExploration.value.chartType === TransactionExplorerChartType.Sunburst.value) {
            return 'sunburst';
        } else {
            return undefined;
        }
    });

    const axisChartSmoothCurve = computed<boolean>(() => {
        return (currentExploration.value.chartType === TransactionExplorerChartType.SmoothLineGrouped.value
            || currentExploration.value.chartType === TransactionExplorerChartType.SmoothAreaStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.SmoothArea100PercentStacked.value);
    });

    const axisChartStacked = computed<boolean>(() => {
        return (currentExploration.value.chartType === TransactionExplorerChartType.ColumnStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.Column100PercentStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.AreaStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.Area100PercentStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.SmoothAreaStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.SmoothArea100PercentStacked.value);
    });

    const axisChart100PercentStacked = computed<boolean>(() => {
        return (currentExploration.value.chartType === TransactionExplorerChartType.Column100PercentStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.Area100PercentStacked.value
            || currentExploration.value.chartType === TransactionExplorerChartType.SmoothArea100PercentStacked.value);
    });

    const axisChartShowYearOverYear = computed<boolean>(() => {
        const dimession = currentExploration.value.categoryDimension;

        return dimession === TransactionExplorerDataDimension.DateTimeByYearMonthDay.value
            || dimession === TransactionExplorerDataDimension.DateTimeByYearMonth.value
            || dimession === TransactionExplorerDataDimension.DateTimeByYearQuarter.value
            || dimession === TransactionExplorerDataDimension.DateTimeByYear.value
            || dimession === TransactionExplorerDataDimension.DateTimeByFiscalYear.value;
    });

    const axisChartShowPeriodOverPeriod = computed<boolean>(() => {
        const dimession = currentExploration.value.categoryDimension;

        return dimession === TransactionExplorerDataDimension.DateTimeByYearMonthDay.value
            || dimession === TransactionExplorerDataDimension.DateTimeByYearMonth.value
            || dimession === TransactionExplorerDataDimension.DateTimeByYearQuarter.value;
    });

    const axisChartTooltipExtraColumnNames = computed<string[]>(() => {
        const extraColumnNames: string[] = [];
        const dimession = currentExploration.value.categoryDimension;

        if (axisChartShowYearOverYear.value) {
            extraColumnNames.push(tt('Year-over-Year'));
        }

        if (axisChartShowPeriodOverPeriod.value) {
            if (dimession === TransactionExplorerDataDimension.DateTimeByYearQuarter.value) {
                extraColumnNames.push(tt('Quarter-over-Quarter'));
            } else if (dimession === TransactionExplorerDataDimension.DateTimeByYearMonth.value) {
                extraColumnNames.push(tt('Month-over-Month'));
            } else if (dimession === TransactionExplorerDataDimension.DateTimeByYearMonthDay.value) {
                extraColumnNames.push(tt('Day-over-Day'));
            } else {
                extraColumnNames.push(tt('Period-over-Period'));
            }
        }

        return extraColumnNames;
    });

    return {
        axisChartDisplayType, hierarchyChartDisplayType, axisChartSmoothCurve, axisChartStacked, axisChart100PercentStacked,
        axisChartShowYearOverYear, axisChartShowPeriodOverPeriod, axisChartTooltipExtraColumnNames,
    };
}
