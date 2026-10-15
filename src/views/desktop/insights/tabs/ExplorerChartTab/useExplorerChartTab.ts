import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import { useExplorerChartTabChartOptions } from './useExplorerChartTab/chart-options.ts';
import { useExplorerChartTabExport } from './useExplorerChartTab/export.ts';
import { useExplorerChartTabFormatting } from './useExplorerChartTab/formatting.ts';
import { useExplorerChartTabSelection } from './useExplorerChartTab/selection.ts';
import { useExplorerChartTabSeries } from './useExplorerChartTab/series.ts';
import { useExplorerChartTabState } from './useExplorerChartTab/state.ts';
import { useExplorerChartTabTooltip } from './useExplorerChartTab/tooltip.ts';
export * from './useExplorerChartTab/types.ts';

// Coordinates ExplorerChartTab state and effects; called once from the component setup.
export function useExplorerChartTab(emit: {
    (e: 'click:transaction', value: TransactionInsightDataItem): void;
}) {
    const {
        tt, formatRange, getMonthLongName, getMonthdayShortName, getWeekdayLongName, getQuarterName, getCurrencyName,
        formatDateTimeToShortDateTime, formatDateTimeToShortDate, formatDateTimeToShortTime,
        formatDateTimeToGregorianLikeShortYear, formatDateTimeToGregorianLikeShortYearMonth,
        formatDateTimeToGregorianLikeYearQuarter, formatGregorianYearToGregorianLikeFiscalYear, formatAmountToLocalizedNumerals,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        formatNumberToLocalizedNumeralsWithoutDigitGrouping, formatPercentToLocalizedNumerals, explorersStore, axisChart,
        radarChart, hierarchyChart, heatmapChart, customChart, transactionListDialog, defaultCurrency,
        allTransactionExplorerDataDimensions, allTransactionExplorerValueMetrics, allTransactionExplorerChartTypes,
        allTransactionExplorerChartSortingTypes, allTransactionExplorerCustomChartDisplayLayouts,
        currentTransactionExplorerCategoryDimensionName, currentExploration, isUsingAmountRange, allAmountRangeCounts, isInAIMode,
        isAIGenerating,
    } = useExplorerChartTabState();
    const { getCategoriedDataDisplayName } = useExplorerChartTabFormatting({
        formatNumberToLocalizedNumeralsWithoutDigitGrouping, currentExploration, tt, formatDateTimeToShortDateTime,
        formatDateTimeToShortDate, formatDateTimeToGregorianLikeShortYearMonth, formatDateTimeToGregorianLikeYearQuarter,
        formatDateTimeToGregorianLikeShortYear, formatGregorianYearToGregorianLikeFiscalYear, getWeekdayLongName,
        getMonthdayShortName, getMonthLongName, getQuarterName, formatDateTimeToShortTime, getCurrencyName,
        formatAmountToLocalizedNumerals, defaultCurrency, formatRange,
    });
    const {
        axisChartDisplayType, hierarchyChartDisplayType, axisChartSmoothCurve, axisChartStacked, axisChart100PercentStacked,
        axisChartShowYearOverYear, axisChartShowPeriodOverPeriod, axisChartTooltipExtraColumnNames,
    } = useExplorerChartTabChartOptions({ currentExploration, tt });
    const {
        categoryDimensionTransactionExplorerData, categoriedDataSortedByDisplayOrder, categoriedNamesSortedByDisplayOrder,
        seriesDimensionTransactionExplorerData, seriesDimensionTransactionExplorerDataMap, axisChartCategoryIndexYoYMap,
    } = useExplorerChartTabSeries({ currentExploration, explorersStore, getCategoriedDataDisplayName, axisChartShowYearOverYear });
    const { getAxisChartTooltipExtraColumnTotalValues, getAxisChartTooltipExtraColumnValues } = useExplorerChartTabTooltip({
        formatPercentToLocalizedNumerals, axisChartShowYearOverYear, axisChartShowPeriodOverPeriod, axisChartCategoryIndexYoYMap,
        seriesDimensionTransactionExplorerDataMap,
    });
    const {
        updateChartType, updateCategoryDimensionType, switchToAIChartMode, onClickPieChartItem, onClickTrendChartItem,
        onClickHierarchyChartItem, onClickHeatmapChartItem, onClickCalendarHeatmapChartItem, onClickTransaction,
    } = useExplorerChartTabSelection({
        currentExploration, customChart, explorersStore, getCategoriedDataDisplayName, transactionListDialog,
        categoriedDataSortedByDisplayOrder, seriesDimensionTransactionExplorerData, emit,
    });
    const { buildExportResults } = useExplorerChartTabExport({
        currentExploration, tt, categoryDimensionTransactionExplorerData, formatAmountToWesternArabicNumeralsWithoutDigitGrouping,
        defaultCurrency, formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping, radarChart, axisChartDisplayType, axisChart,
        hierarchyChartDisplayType, hierarchyChart, heatmapChart,
    });

    return {
        tt, explorersStore, axisChart, radarChart, hierarchyChart, heatmapChart, customChart, transactionListDialog,
        defaultCurrency, allTransactionExplorerDataDimensions, allTransactionExplorerValueMetrics,
        allTransactionExplorerChartTypes, allTransactionExplorerChartSortingTypes,
        allTransactionExplorerCustomChartDisplayLayouts, currentTransactionExplorerCategoryDimensionName, currentExploration,
        isUsingAmountRange, allAmountRangeCounts, categoryDimensionTransactionExplorerData, categoriedNamesSortedByDisplayOrder,
        seriesDimensionTransactionExplorerData, axisChartDisplayType, hierarchyChartDisplayType, axisChartSmoothCurve,
        axisChartStacked, axisChart100PercentStacked, axisChartShowYearOverYear, axisChartShowPeriodOverPeriod,
        axisChartTooltipExtraColumnNames, isInAIMode, isAIGenerating, getAxisChartTooltipExtraColumnTotalValues,
        getAxisChartTooltipExtraColumnValues, updateChartType, updateCategoryDimensionType, switchToAIChartMode,
        onClickPieChartItem, onClickTrendChartItem, onClickHierarchyChartItem, onClickHeatmapChartItem,
        onClickCalendarHeatmapChartItem, onClickTransaction, buildExportResults,
    };
}
