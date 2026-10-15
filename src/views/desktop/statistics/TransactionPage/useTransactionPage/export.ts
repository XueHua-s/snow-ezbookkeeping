import {
    CategoricalChartType,
    ExportMermaidChartType,
    StatisticsAnalysisType,
    TrendChartType
} from '@/core/statistics.ts';
import {
    parseBigDecimal
} from '@/lib/numeral.ts';
import type { useTransactionPageState } from './state.ts';

interface Context {
    analysisType: ReturnType<typeof useTransactionPageState>['analysisType'];
    categoricalAnalysisData: ReturnType<typeof useTransactionPageState>['categoricalAnalysisData'];
    query: ReturnType<typeof useTransactionPageState>['query'];
    dataExportDialog: ReturnType<typeof useTransactionPageState>['dataExportDialog'];
    tt: ReturnType<typeof useTransactionPageState>['tt'];
    defaultCurrency: ReturnType<typeof useTransactionPageState>['defaultCurrency'];
    formatAmountToWesternArabicNumeralsWithoutDigitGrouping: ReturnType<typeof useTransactionPageState>['formatAmountToWesternArabicNumeralsWithoutDigitGrouping'];
    formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping: ReturnType<typeof useTransactionPageState>['formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping'];
    trendsAnalysisData: ReturnType<typeof useTransactionPageState>['trendsAnalysisData'];
    monthlyTrendsChart: ReturnType<typeof useTransactionPageState>['monthlyTrendsChart'];
    assetTrendsData: ReturnType<typeof useTransactionPageState>['assetTrendsData'];
    dailyTrendsChart: ReturnType<typeof useTransactionPageState>['dailyTrendsChart'];
}

export function useTransactionPageExport(context: Context) {
    const {
        analysisType, categoricalAnalysisData, query, dataExportDialog, tt, defaultCurrency,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        trendsAnalysisData, monthlyTrendsChart, assetTrendsData, dailyTrendsChart,
    } = context;

    function exportResults(): void {
        if (analysisType.value === StatisticsAnalysisType.CategoricalAnalysis && categoricalAnalysisData.value && categoricalAnalysisData.value.items) {
            let supportedMermaidCharts: ExportMermaidChartType[] | undefined = undefined;

            if (query.value.categoricalChartType === CategoricalChartType.Pie.type) {
                supportedMermaidCharts = [ ExportMermaidChartType.PieChart ];
            }

            dataExportDialog.value?.open({
                headers: [
                    tt('Name'),
                    tt('Amount') + ` (${defaultCurrency.value})`,
                    tt('Proportion (%)')
                ],
                data: categoricalAnalysisData.value.items
                    .filter(item => !item.hidden)
                    .map(item => [
                        item.name,
                        formatAmountToWesternArabicNumeralsWithoutDigitGrouping(item.value, defaultCurrency.value),
                        formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping(parseBigDecimal(item.percent.toFixed(4)))
                    ]),
                supportedMermaidCharts: supportedMermaidCharts
            });
        } else if (analysisType.value === StatisticsAnalysisType.TrendAnalysis && trendsAnalysisData.value && trendsAnalysisData.value.items && monthlyTrendsChart.value) {
            const exportData = monthlyTrendsChart.value.exportData();
            let supportedMermaidCharts: ExportMermaidChartType[] | undefined = undefined;

            if (exportData.headers.length === 2 && query.value.trendChartType === TrendChartType.Column.type) {
                supportedMermaidCharts = [ ExportMermaidChartType.XYChartBar ];
            } else if (exportData.headers.length === 2 && query.value.trendChartType === TrendChartType.Area.type) {
                supportedMermaidCharts = [ ExportMermaidChartType.XYChartLine ];
            }

            dataExportDialog.value?.open({
                headers: exportData.headers || [],
                data: exportData.data || [],
                supportedMermaidCharts: supportedMermaidCharts
            });
        } else if (analysisType.value === StatisticsAnalysisType.AssetTrends && assetTrendsData.value && assetTrendsData.value.items && dailyTrendsChart.value) {
            const exportData = dailyTrendsChart.value.exportData();
            let supportedMermaidCharts: ExportMermaidChartType[] | undefined = undefined;

            if (exportData.headers.length === 2 && query.value.assetTrendsChartType === TrendChartType.Column.type) {
                supportedMermaidCharts = [ ExportMermaidChartType.XYChartBar ];
            } else if (exportData.headers.length === 2 && query.value.assetTrendsChartType === TrendChartType.Area.type) {
                supportedMermaidCharts = [ ExportMermaidChartType.XYChartLine ];
            }

            dataExportDialog.value?.open({
                headers: exportData.headers || [],
                data: exportData.data || [],
                supportedMermaidCharts: supportedMermaidCharts
            });
        }
    }

    return { exportResults };
}
