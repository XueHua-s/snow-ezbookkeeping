import { ChartValueType } from '@/core/chart.ts';
import {
    TransactionExplorerChartType,
    TransactionExplorerValueMetric
} from '@/core/explorer.ts';
import { ExportMermaidChartType } from '@/core/statistics.ts';
import type { useExplorerChartTabChartOptions } from './chart-options.ts';
import type { useExplorerChartTabSeries } from './series.ts';
import type { useExplorerChartTabState } from './state.ts';

interface Context {
    currentExploration: ReturnType<typeof useExplorerChartTabState>['currentExploration'];
    tt: ReturnType<typeof useExplorerChartTabState>['tt'];
    categoryDimensionTransactionExplorerData: ReturnType<typeof useExplorerChartTabSeries>['categoryDimensionTransactionExplorerData'];
    formatAmountToWesternArabicNumeralsWithoutDigitGrouping: ReturnType<typeof useExplorerChartTabState>['formatAmountToWesternArabicNumeralsWithoutDigitGrouping'];
    defaultCurrency: ReturnType<typeof useExplorerChartTabState>['defaultCurrency'];
    formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping: ReturnType<typeof useExplorerChartTabState>['formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping'];
    radarChart: ReturnType<typeof useExplorerChartTabState>['radarChart'];
    axisChartDisplayType: ReturnType<typeof useExplorerChartTabChartOptions>['axisChartDisplayType'];
    axisChart: ReturnType<typeof useExplorerChartTabState>['axisChart'];
    hierarchyChartDisplayType: ReturnType<typeof useExplorerChartTabChartOptions>['hierarchyChartDisplayType'];
    hierarchyChart: ReturnType<typeof useExplorerChartTabState>['hierarchyChart'];
    heatmapChart: ReturnType<typeof useExplorerChartTabState>['heatmapChart'];
}

export function useExplorerChartTabExport(context: Context) {
    const {
        currentExploration, tt, categoryDimensionTransactionExplorerData, formatAmountToWesternArabicNumeralsWithoutDigitGrouping,
        defaultCurrency, formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping, radarChart, axisChartDisplayType, axisChart,
        hierarchyChartDisplayType, hierarchyChart, heatmapChart,
    } = context;

    function buildExportResults(): { headers: string[], data: string[][], supportedMermaidCharts?: ExportMermaidChartType[] } | undefined {
        if (currentExploration.value.chartType === TransactionExplorerChartType.Pie.value
            || currentExploration.value.chartType === TransactionExplorerChartType.Donut.value
            || currentExploration.value.chartType === TransactionExplorerChartType.NightingaleRose.value
            || currentExploration.value.chartType === TransactionExplorerChartType.CalendarHeatmap.value) {
            const valueMetric = TransactionExplorerValueMetric.valueOf(currentExploration.value.valueMetric);
            let supportedMermaidCharts: ExportMermaidChartType[] | undefined = undefined;

            if (currentExploration.value.chartType === TransactionExplorerChartType.Pie.value) {
                supportedMermaidCharts = [ ExportMermaidChartType.PieChart ];
            }

            return {
                headers: [
                    tt('Name'),
                    tt(valueMetric?.name ?? 'Unknown')
                ],
                data: categoryDimensionTransactionExplorerData.value.map(data => [
                    data.name,
                    valueMetric?.valueType === ChartValueType.Amount ? formatAmountToWesternArabicNumeralsWithoutDigitGrouping(data.value, defaultCurrency.value) : formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping(data.value)
                ]),
                supportedMermaidCharts: supportedMermaidCharts
            };
        } else if (currentExploration.value.chartType === TransactionExplorerChartType.Radar.value) {
            const results = radarChart.value?.exportData();

            if (!results) {
                return undefined;
            }

            return {
                headers: results.headers,
                data: results.data,
                supportedMermaidCharts: undefined
            };
        } else if (TransactionExplorerChartType.valueOf(currentExploration.value.chartType)?.seriesDimensionRequired && axisChartDisplayType.value) {
            const results = axisChart.value?.exportData();

            if (!results) {
                return undefined;
            }

            let supportedMermaidCharts: ExportMermaidChartType[] | undefined = undefined;

            if (results.headers.length === 2 &&
                (
                    currentExploration.value.chartType === TransactionExplorerChartType.ColumnStacked.value ||
                    currentExploration.value.chartType === TransactionExplorerChartType.Column100PercentStacked.value ||
                    currentExploration.value.chartType === TransactionExplorerChartType.ColumnGrouped.value
                )
            ) {
                supportedMermaidCharts = [ ExportMermaidChartType.XYChartBar ];
            } else if (results.headers.length === 2 &&
                (
                    currentExploration.value.chartType === TransactionExplorerChartType.AreaStacked.value ||
                    currentExploration.value.chartType === TransactionExplorerChartType.Area100PercentStacked.value ||
                    currentExploration.value.chartType === TransactionExplorerChartType.SmoothAreaStacked.value ||
                    currentExploration.value.chartType === TransactionExplorerChartType.SmoothArea100PercentStacked.value
                )
            ) {
                supportedMermaidCharts = [ ExportMermaidChartType.XYChartLine ];
            } else if (currentExploration.value.chartType === TransactionExplorerChartType.LineGrouped.value
                || currentExploration.value.chartType === TransactionExplorerChartType.SmoothLineGrouped.value) {
                supportedMermaidCharts = [ ExportMermaidChartType.XYChartLine ];
            }

            return {
                headers: results.headers,
                data: results.data,
                supportedMermaidCharts: supportedMermaidCharts
            };
        } else if (TransactionExplorerChartType.valueOf(currentExploration.value.chartType)?.seriesDimensionRequired && hierarchyChartDisplayType.value) {
            const results = hierarchyChart.value?.exportData();

            if (!results) {
                return undefined;
            }

            return {
                headers: results.headers,
                data: results.data,
                supportedMermaidCharts: undefined
            };
        } else if (TransactionExplorerChartType.valueOf(currentExploration.value.chartType)?.seriesDimensionRequired && currentExploration.value.chartType === TransactionExplorerChartType.Heatmap.value) {
            const results = heatmapChart.value?.exportData();

            if (!results) {
                return undefined;
            }

            return {
                headers: results.headers,
                data: results.data,
                supportedMermaidCharts: undefined
            };
        } else {
            return undefined;
        }
    }

    return { buildExportResults };
}
