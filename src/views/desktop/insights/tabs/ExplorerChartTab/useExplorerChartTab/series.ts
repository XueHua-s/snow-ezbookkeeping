import {
    entries,
    itemAndIndex
} from '@/core/base.ts';
import {
    TransactionExplorerChartType,
    TransactionExplorerDataDimension
} from '@/core/explorer.ts';
import { ChartSortingType } from '@/core/statistics.ts';
import { isDefined, isNumber } from '@/lib/common.ts';
import { parseDateTimeFromString } from '@/lib/datetime.ts';
import { BIG_DECIMAL_ZERO } from '@/lib/numeral.ts';
import { sortStatisticsItems } from '@/lib/statistics.ts';
import {
    type CategoriedTransactionExplorerDataItem
} from '@/stores/explorer.ts';
import { computed } from 'vue';
import type { useExplorerChartTabChartOptions } from './chart-options.ts';
import type { useExplorerChartTabFormatting } from './formatting.ts';
import type { useExplorerChartTabState } from './state.ts';
import { type CategoryDimensionData, type SeriesDimensionData, type SortableCategoriedTransactionExplorerDataItem } from './types.ts';

interface Context {
    currentExploration: ReturnType<typeof useExplorerChartTabState>['currentExploration'];
    explorersStore: ReturnType<typeof useExplorerChartTabState>['explorersStore'];
    getCategoriedDataDisplayName: ReturnType<typeof useExplorerChartTabFormatting>['getCategoriedDataDisplayName'];
    axisChartShowYearOverYear: ReturnType<typeof useExplorerChartTabChartOptions>['axisChartShowYearOverYear'];
}

export function useExplorerChartTabSeries(context: Context) {
    const { currentExploration, explorersStore, getCategoriedDataDisplayName, axisChartShowYearOverYear } = context;

    const categoryDimensionTransactionExplorerData = computed<CategoryDimensionData[]>(() => {
        if (currentExploration.value.chartType !== TransactionExplorerChartType.Pie.value
            && currentExploration.value.chartType !== TransactionExplorerChartType.Donut.value
            && currentExploration.value.chartType !== TransactionExplorerChartType.NightingaleRose.value
            && currentExploration.value.chartType !== TransactionExplorerChartType.CalendarHeatmap.value) {
            return [];
        }

        if (!explorersStore.categoriedTransactionExplorerData || !explorersStore.categoriedTransactionExplorerData.length) {
            return [];
        }

        const result: CategoryDimensionData[] = [];

        for (const categoriedData of explorersStore.categoriedTransactionExplorerData) {
            const data = categoriedData.data[0];

            if (!isDefined(data)) {
                continue;
            }

            const displayName = getCategoriedDataDisplayName(categoriedData);

            result.push({
                id: categoriedData.categoryId,
                dimension: categoriedData.categoryIdType,
                name: displayName,
                displayOrders: categoriedData.categoryDisplayOrders,
                value: data.value
            });
        }

        sortStatisticsItems(result, currentExploration.value.chartSortingType);

        return result;
    });

    const categoriedDataSortedByDisplayOrder = computed<SortableCategoriedTransactionExplorerDataItem[]>(() => {
        if (!explorersStore.categoriedTransactionExplorerData || !explorersStore.categoriedTransactionExplorerData.length) {
            return [];
        }

        const result: SortableCategoriedTransactionExplorerDataItem[] = [];
        const customSorting = currentExploration.value.chartType === TransactionExplorerChartType.Radar.value && currentExploration.value.seriesDimension === TransactionExplorerDataDimension.None.value;

        for (const categoriedData of explorersStore.categoriedTransactionExplorerData) {
            result.push({
                name: getCategoriedDataDisplayName(categoriedData),
                displayOrders: categoriedData.categoryDisplayOrders,
                value: customSorting ? (categoriedData.data[0]?.value ?? BIG_DECIMAL_ZERO) : BIG_DECIMAL_ZERO,
                originalItem: categoriedData
            });
        }

        sortStatisticsItems(result, customSorting ? currentExploration.value.chartSortingType : ChartSortingType.DisplayOrder.type);

        return result;
    });

    const categoriedNamesSortedByDisplayOrder = computed<string[]>(() => {
        if (!categoriedDataSortedByDisplayOrder.value || !categoriedDataSortedByDisplayOrder.value.length) {
            return [];
        }

        const result: string[] = [];

        for (const categoriedData of categoriedDataSortedByDisplayOrder.value) {
            result.push(categoriedData.name);
        }

        return result;
    });

    const seriesDimensionTransactionExplorerData = computed<SeriesDimensionData[]>(() => {
        if (!TransactionExplorerChartType.valueOf(currentExploration.value.chartType)?.seriesDimensionRequired) {
            return [];
        }

        if (!explorersStore.categoriedTransactionExplorerData || !explorersStore.categoriedTransactionExplorerData.length) {
            return [];
        }

        const result: SeriesDimensionData[] = [];
        const seriesDimensionDataMap: Record<string, SeriesDimensionData> = {};

        for (const categoriedData of explorersStore.categoriedTransactionExplorerData) {
            for (const seriesData of categoriedData.data) {
                const displayName = getCategoriedDataDisplayName(seriesData);
                let seriesDimensionData: SeriesDimensionData | undefined = seriesDimensionDataMap[seriesData.seriesId];

                if (!seriesDimensionData) {
                    seriesDimensionData = {
                        id: seriesData.seriesId,
                        dimension: seriesData.seriesIdType,
                        name: displayName,
                        displayOrders: seriesData.seriesDisplayOrders,
                        values: [],
                        value: BIG_DECIMAL_ZERO
                    };

                    seriesDimensionDataMap[seriesData.seriesId] = seriesDimensionData;
                    result.push(seriesDimensionData);
                }
            }
        }

        for (const categoriedData of categoriedDataSortedByDisplayOrder.value) {
            const seriesDataMap: Record<string, CategoriedTransactionExplorerDataItem> = {};

            for (const seriesData of categoriedData.originalItem.data) {
                seriesDataMap[seriesData.seriesId] = seriesData;
            }

            for (const seriesDimensionData of result) {
                const seriesData = seriesDataMap[seriesDimensionData.id];

                if (isDefined(seriesData)) {
                    seriesDimensionData.values.push(seriesData.value);
                    seriesDimensionData.value = seriesDimensionData.value.add(seriesData.value);
                } else {
                    seriesDimensionData.values.push(BIG_DECIMAL_ZERO);
                }
            }
        }

        sortStatisticsItems(result, currentExploration.value.chartSortingType);

        return result;
    });

    const seriesDimensionTransactionExplorerDataMap = computed<Record<string, SeriesDimensionData>>(() => {
        const result: Record<string, SeriesDimensionData> = {};

        for (const seriesDimensionData of seriesDimensionTransactionExplorerData.value) {
            result[seriesDimensionData.id] = seriesDimensionData;
        }

        return result;
    });

    const axisChartCategoryIndexYoYMap = computed<Record<number, number>>(() => {
        const result: Record<number, number> = {};

        if (!axisChartShowYearOverYear.value) {
            return result;
        }

        const dateKeyToIndex: Record<string, number> = {};
        const dateKeyToPreviousYearDateKey: Record<string, string> = {};

        for (const [item, categoryIndex] of itemAndIndex(categoriedDataSortedByDisplayOrder.value)) {
            const categoriedData = item.originalItem;
            const name = categoriedData.categoryId;
            const dimessionType = categoriedData.categoryIdType;
            const dimension = currentExploration.value.categoryDimension;

            if (dimension === TransactionExplorerDataDimension.DateTimeByYearMonthDay.value) {
                const dateTime = parseDateTimeFromString(name, dimessionType);

                if (dateTime) {
                    dateKeyToIndex[dateTime.getGregorianCalendarYearDashMonthDashDay()] = categoryIndex;
                    dateKeyToPreviousYearDateKey[dateTime.getGregorianCalendarYearDashMonthDashDay()] = dateTime.add(-1, 'years').getGregorianCalendarYearDashMonthDashDay();
                }
            } else if (dimension === TransactionExplorerDataDimension.DateTimeByYearMonth.value) {
                const dateTime = parseDateTimeFromString(name, dimessionType);

                if (dateTime) {
                    dateKeyToIndex[dateTime.getGregorianCalendarYearDashMonth()] = categoryIndex;
                    dateKeyToPreviousYearDateKey[dateTime.getGregorianCalendarYearDashMonth()] = dateTime.add(-1, 'years').getGregorianCalendarYearDashMonth();
                }
            } else if (dimension === TransactionExplorerDataDimension.DateTimeByYearQuarter.value) {
                const parts = name.split('-');
                const year = parts.length === 2 ? parseInt(parts[0] as string) : 0;
                const quarter = parts.length === 2 ? parseInt(parts[1] as string) : 0;

                dateKeyToIndex[`${year}-Q${quarter}`] = categoryIndex;
                dateKeyToPreviousYearDateKey[`${year}-Q${quarter}`] = `${year - 1}-Q${quarter}`;
            } else if (dimension === TransactionExplorerDataDimension.DateTimeByYear.value) {
                const year = parseInt(name);
                dateKeyToIndex[name] = categoryIndex;
                dateKeyToPreviousYearDateKey[name] = (year - 1).toString(10);
            } else if (dimension === TransactionExplorerDataDimension.DateTimeByFiscalYear.value) {
                const year = parseInt(name);
                dateKeyToIndex[name] = categoryIndex;
                dateKeyToPreviousYearDateKey[name] = (year - 1).toString(10);
            }
        }

        for (const [dateKey, previousYearDateKey] of entries(dateKeyToPreviousYearDateKey)) {
            const categoryIndex = dateKeyToIndex[dateKey];
            const previousYearCategoryIndex = dateKeyToIndex[previousYearDateKey];

            if (isNumber(categoryIndex) && isNumber(previousYearCategoryIndex)) {
                result[categoryIndex] = previousYearCategoryIndex;
            }
        }

        return result;
    });

    return {
        categoryDimensionTransactionExplorerData, categoriedDataSortedByDisplayOrder, categoriedNamesSortedByDisplayOrder,
        seriesDimensionTransactionExplorerData, seriesDimensionTransactionExplorerDataMap, axisChartCategoryIndexYoYMap,
    };
}
