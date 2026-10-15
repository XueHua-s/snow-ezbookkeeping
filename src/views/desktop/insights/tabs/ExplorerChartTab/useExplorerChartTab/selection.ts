import {
    TransactionExplorerChartType,
    TransactionExplorerChartTypeValue,
    TransactionExplorerCustomChartDisplayLayout,
    TransactionExplorerDataDimension,
    TransactionExplorerDataDimensionType
} from '@/core/explorer.ts';
import { isDefined } from '@/lib/common.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { useExplorerChartTabFormatting } from './formatting.ts';
import type { useExplorerChartTabSeries } from './series.ts';
import type { useExplorerChartTabState } from './state.ts';
import { type CategoryDimensionData } from './types.ts';

interface Context {
    currentExploration: ReturnType<typeof useExplorerChartTabState>['currentExploration'];
    customChart: ReturnType<typeof useExplorerChartTabState>['customChart'];
    explorersStore: ReturnType<typeof useExplorerChartTabState>['explorersStore'];
    getCategoriedDataDisplayName: ReturnType<typeof useExplorerChartTabFormatting>['getCategoriedDataDisplayName'];
    transactionListDialog: ReturnType<typeof useExplorerChartTabState>['transactionListDialog'];
    categoriedDataSortedByDisplayOrder: ReturnType<typeof useExplorerChartTabSeries>['categoriedDataSortedByDisplayOrder'];
    seriesDimensionTransactionExplorerData: ReturnType<typeof useExplorerChartTabSeries>['seriesDimensionTransactionExplorerData'];
    emit: {
    (e: 'click:transaction', value: TransactionInsightDataItem): void;
};
}

export function useExplorerChartTabSelection(context: Context) {
    const {
        currentExploration, customChart, explorersStore, getCategoriedDataDisplayName, transactionListDialog,
        categoriedDataSortedByDisplayOrder, seriesDimensionTransactionExplorerData, emit,
    } = context;

    function updateChartType(chartTypeValue: TransactionExplorerChartTypeValue): void {
        if (currentExploration.value.chartType !== chartTypeValue) {
            currentExploration.value.chartType = chartTypeValue;
        }

        const chartType = TransactionExplorerChartType.valueOf(chartTypeValue);

        if (isDefined(chartType?.fixedCategoryDimension)) {
            currentExploration.value.categoryDimension = chartType.fixedCategoryDimension;
        }

        if (!chartType?.seriesDimensionRequired) {
            currentExploration.value.seriesDimension = TransactionExplorerDataDimension.None.value;
        }

        if (isDefined(chartType?.fixedSortingType)) {
            currentExploration.value.chartSortingType = chartType.fixedSortingType;
        }
    }

    function updateCategoryDimensionType(dimensionType: TransactionExplorerDataDimensionType): void {
        if (currentExploration.value.categoryDimension !== dimensionType) {
            currentExploration.value.categoryDimension = dimensionType;
        }

        if (currentExploration.value.seriesDimension === currentExploration.value.categoryDimension) {
            currentExploration.value.seriesDimension = TransactionExplorerDataDimension.None.value;
        }
    }

    function switchToAIChartMode(): void {
        if (currentExploration.value.chartType !== TransactionExplorerChartType.Custom.value) {
            currentExploration.value.chartType = TransactionExplorerChartType.Custom.value;
        }

        const displayLayout = TransactionExplorerCustomChartDisplayLayout.valueOf(currentExploration.value.customChartDisplayLayout) ?? TransactionExplorerCustomChartDisplayLayout.Default

        if (!displayLayout.showCode) {
            currentExploration.value.customChartDisplayLayout = TransactionExplorerCustomChartDisplayLayout.Default.type;
        }

        customChart.value?.switchToAIChartMode();
    }

    function showClickedTransactionList({ categoryId, seriesId, title }: { categoryId: string, seriesId?: string, title?: string }): void {
        const categoriedData = explorersStore.categoriedTransactions[categoryId];

        if (!categoriedData) {
            return;
        }

        const seriesData = seriesId ? categoriedData.trasactions[seriesId] : undefined;

        if (!title) {
            if (seriesData) {
                if (currentExploration.value.seriesDimension === TransactionExplorerDataDimension.None.value) {
                    title = getCategoriedDataDisplayName(categoriedData);
                } else {
                    title = getCategoriedDataDisplayName(seriesData);
                }
            } else {
                title = getCategoriedDataDisplayName(categoriedData);
            }
        }

        transactionListDialog.value?.open({
            title: title,
            categoryId: categoryId,
            seriesId: seriesId
        });
    }

    function onClickPieChartItem(item: Record<string, unknown>): void {
        if (!item || !('id' in item) || !('dimension' in item)) {
            return;
        }

        const data = (item as unknown) as CategoryDimensionData;
        showClickedTransactionList({
            categoryId: data.id
        });
    }

    function onClickTrendChartItem(itemId: string, categoryIndex: number): void {
        const categoryData = categoriedDataSortedByDisplayOrder.value[categoryIndex];

        if (categoryData) {
            showClickedTransactionList({
                categoryId: categoryData.originalItem.categoryId,
                seriesId: itemId
            });
        }
    }

    function onClickHierarchyChartItem(categoryIndex: number, seriesIndex: number): void {
        const categoryData = categoriedDataSortedByDisplayOrder.value[categoryIndex];
        const seriesData = seriesDimensionTransactionExplorerData.value[seriesIndex];

        if (categoryData && seriesData) {
            showClickedTransactionList({
                categoryId: categoryData.originalItem.categoryId,
                seriesId: seriesData.id,
                title: categoryData.name
            });
        }
    }

    function onClickHeatmapChartItem(categoryIndex: number, seriesIndex: number): void {
        const categoryData = categoriedDataSortedByDisplayOrder.value[categoryIndex];
        const seriesData = seriesDimensionTransactionExplorerData.value[seriesIndex];

        if (categoryData && seriesData) {
            showClickedTransactionList({
                categoryId: categoryData.originalItem.categoryId,
                seriesId: seriesData.id
            });
        }
    }

    function onClickCalendarHeatmapChartItem(date: string, displayDate: string): void {
        showClickedTransactionList({
            categoryId: date,
            title: displayDate
        });
    }

    function onClickTransaction(transaction: TransactionInsightDataItem): void {
        emit('click:transaction', transaction);
    }

    return {
        updateChartType, updateCategoryDimensionType, switchToAIChartMode, onClickPieChartItem, onClickTrendChartItem,
        onClickHierarchyChartItem, onClickHeatmapChartItem, onClickCalendarHeatmapChartItem, onClickTransaction,
    };
}
