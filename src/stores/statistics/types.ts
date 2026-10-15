import type { CategoricalChartSourceDataItem, TrendsChartSourceDataItem } from '@/core/chart.ts';
import type { ColorValue } from '@/core/color.ts';
import type { TextualYearMonth } from '@/core/datetime.ts';
import type { IconType } from '@/core/icon.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import type {
    TransactionAssetTrendsAnalysisDataAmount, TransactionAssetTrendsAnalysisDataItem, TransactionStatisticDataItemType,
    TransactionTrendsAnalysisDataAmount, TransactionTrendsAnalysisDataItem,
} from '@/models/transaction.ts';

export interface WritableTransactionCategoricalAnalysisData {
    value: BigDecimal;
    totalNonNegativeAmount: BigDecimal;
    items: Record<string, WritableTransactionCategoricalAnalysisDataItem>;
}

export interface WritableTransactionCategoricalAnalysisDataItem extends Record<string, unknown>, CategoricalChartSourceDataItem {
    name: string;
    type: TransactionStatisticDataItemType;
    id: string;
    icon: string;
    iconType: IconType;
    color: ColorValue;
    hidden: boolean;
    displayOrders: number[];
    value: BigDecimal;
    originalValue?: BigDecimal;
    originalCurrency?: string;
    percent?: number;
}

export interface WritableTransactionTrendsAnalysisDataItem extends Record<string, unknown>, TrendsChartSourceDataItem, TransactionTrendsAnalysisDataItem {
    name: string;
    type: TransactionStatisticDataItemType;
    id: string;
    icon: string;
    iconType: IconType;
    color: ColorValue;
    hidden: boolean;
    displayOrders: number[];
    value: BigDecimal;
    items: TransactionTrendsAnalysisDataAmount[];
}

export interface WritableTransactionAssetTrendsAnalysisDataItem extends Record<string, unknown>, TrendsChartSourceDataItem, TransactionAssetTrendsAnalysisDataItem {
    name: string;
    type: TransactionStatisticDataItemType;
    id: string;
    icon: string;
    iconType: IconType;
    color: ColorValue;
    hidden: boolean;
    displayOrders: number[];
    value: BigDecimal;
    items: TransactionAssetTrendsAnalysisDataAmount[];
}

export interface TransactionStatisticsPartialFilter {
    chartDataType?: number;
    categoricalChartType?: number;
    categoricalChartDateType?: number;
    categoricalChartStartTime?: number;
    categoricalChartEndTime?: number;
    trendChartType?: number;
    trendChartDateType?: number;
    trendChartStartYearMonth?: TextualYearMonth | '';
    trendChartEndYearMonth?: TextualYearMonth | '';
    assetTrendsChartType?: number;
    assetTrendsChartDateType?: number;
    assetTrendsChartStartTime?: number;
    assetTrendsChartEndTime?: number;
    filterAccountIds?: Record<string, boolean>;
    filterCategoryIds?: Record<string, boolean>;
    tagFilter?: string;
    keyword?: string;
    matchMode?: number;
    sortingType?: number;
}

export interface TransactionStatisticsFilter extends TransactionStatisticsPartialFilter {
    chartDataType: number;
    categoricalChartType: number;
    categoricalChartDateType: number;
    categoricalChartStartTime: number;
    categoricalChartEndTime: number;
    trendChartType: number;
    trendChartDateType: number;
    trendChartStartYearMonth: TextualYearMonth | '';
    trendChartEndYearMonth: TextualYearMonth | '';
    assetTrendsChartType: number;
    assetTrendsChartDateType: number;
    assetTrendsChartStartTime: number;
    assetTrendsChartEndTime: number;
    filterAccountIds: Record<string, boolean>;
    filterCategoryIds: Record<string, boolean>;
    tagFilter: string;
    keyword: string;
    matchMode: number;
    sortingType: number;
}
