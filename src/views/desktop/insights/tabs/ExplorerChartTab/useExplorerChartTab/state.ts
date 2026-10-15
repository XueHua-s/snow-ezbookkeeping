import {
    type NameNumeralValue,
    type NameValue,
    type TypeAndDisplayName
} from '@/core/base.ts';
import { findNameByValue } from '@/lib/common.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { InsightsExplorer } from '@/models/explorer.ts';
import {
    useExplorersStore
} from '@/stores/explorer.ts';
import { useUserStore } from '@/stores/user.ts';
import { computed, useTemplateRef } from 'vue';
import { type AxisChartType, type CustomChartType, type HeatMapChartType, type HierarchyChartType, type RadarChartType, type TransactionListDialogType } from './types.ts';


export function useExplorerChartTabState() {
    const {
        tt,
        formatRange,
        getAllStatisticsSortingTypes,
        getAllTransactionExplorerDataDimensions,
        getAllTransactionExplorerValueMetrics,
        getAllTransactionExplorerChartTypes,
        getAllTransactionExplorerCustomChartDisplayLayouts,
        getMonthLongName,
        getMonthdayShortName,
        getWeekdayLongName,
        getQuarterName,
        getCurrencyName,
        formatDateTimeToShortDateTime,
        formatDateTimeToShortDate,
        formatDateTimeToShortTime,
        formatDateTimeToGregorianLikeShortYear,
        formatDateTimeToGregorianLikeShortYearMonth,
        formatDateTimeToGregorianLikeYearQuarter,
        formatGregorianYearToGregorianLikeFiscalYear,
        formatAmountToLocalizedNumerals,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping,
        formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        formatNumberToLocalizedNumerals,
        formatNumberToLocalizedNumeralsWithoutDigitGrouping,
        formatPercentToLocalizedNumerals
    } = useI18n();

    const userStore = useUserStore();

    const explorersStore = useExplorersStore();

    const axisChart = useTemplateRef<AxisChartType>('axisChart');

    const radarChart = useTemplateRef<RadarChartType>('radarChart');

    const hierarchyChart = useTemplateRef<HierarchyChartType>('hierarchyChart');

    const heatmapChart = useTemplateRef<HeatMapChartType>('heatmapChart');

    const customChart = useTemplateRef<CustomChartType>('customChart');

    const transactionListDialog = useTemplateRef<TransactionListDialogType>('transactionListDialog');

    const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

    const allTransactionExplorerDataDimensions = computed<NameValue[]>(() => getAllTransactionExplorerDataDimensions());

    const allTransactionExplorerValueMetrics = computed<NameValue[]>(() => getAllTransactionExplorerValueMetrics());

    const allTransactionExplorerChartTypes = computed<NameValue[]>(() => getAllTransactionExplorerChartTypes());

    const allTransactionExplorerChartSortingTypes = computed<TypeAndDisplayName[]>(() => getAllStatisticsSortingTypes(true));

    const allTransactionExplorerCustomChartDisplayLayouts = computed<TypeAndDisplayName[]>(() => getAllTransactionExplorerCustomChartDisplayLayouts());

    const currentTransactionExplorerCategoryDimensionName = computed<string>(() => findNameByValue(allTransactionExplorerDataDimensions.value, currentExploration.value.categoryDimension) ?? tt('Unknown'));

    const currentExploration = computed<InsightsExplorer>(() => explorersStore.currentExploration);

    const isUsingAmountRange = computed<boolean>(() => explorersStore.isUsingAmountRange);

    const allAmountRangeCounts = computed<NameNumeralValue[]>(() => {
        const pageCounts: NameNumeralValue[] = [];

        for (let i = 3; i <= 20; i++) {
            pageCounts.push({ value: i, name: formatNumberToLocalizedNumerals(i) });
        }

        return pageCounts;
    });

    function isInAIMode(): boolean {
        return customChart.value?.isInAIMode() ?? false;
    }

    function isAIGenerating(): boolean {
        return customChart.value?.isAIGenerating() ?? false;
    }

    return {
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
    };
}
