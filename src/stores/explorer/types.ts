import type { BigDecimal } from '@/core/numeral.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';

export enum TransactionExplorerDimensionType {
    DateTime = 'YYYY-MM-DD HH:mm:ss',
    YearMonthDay = 'YYYY-MM-DD',
    YearMonth = 'YYYY-MM',
    YearQuarter = 'YYYY-Q',
    Year = 'YYYY',
    TransactionType = 'transactionType',
    Category = 'category',
    Account = 'account',
    Amount = 'amount',
    Other = 'other'
}

export interface TransactionExplorerPartialFilter {
    dateRangeType?: number;
    startTime?: number;
    endTime?: number;
}

export interface TransactionExplorerFilter extends TransactionExplorerPartialFilter {
    dateRangeType: number;
    startTime: number;
    endTime: number;
}

export interface CategoriedInfo {
    categoryName: string;
    categoryNameNeedI18n?: boolean;
    categoryNameI18nParameters?: Record<string, unknown>;
    categoryId: string;
    categoryIdType: TransactionExplorerDimensionType;
    categoryDisplayOrders: number[];
}

export interface CategoriedTransactions extends CategoriedInfo {
    trasactions: Record<string, SeriesTransactions>;
}

export interface CategoriedTransactionExplorerData extends CategoriedInfo {
    data: CategoriedTransactionExplorerDataItem[];
}

export interface SeriesInfo {
    seriesName: string;
    seriesNameNeedI18n?: boolean;
    seriesNameI18nParameters?: Record<string, unknown>;
    seriesId: string;
    seriesIdType: TransactionExplorerDimensionType;
    seriesDisplayOrders: number[];
}

export interface SeriesTransactions extends SeriesInfo {
    trasactions: TransactionInsightDataItem[];
}

export interface CategoriedTransactionExplorerDataItem extends SeriesInfo {
    value: BigDecimal;
}

export interface AmountRanges {
    categorySourceAmountRanges?: BigDecimal[];
    categoryDestinationAmountRanges?: BigDecimal[];
    seriesSourceAmountRanges?: BigDecimal[];
    seriesDestinationAmountRanges?: BigDecimal[];
}

export interface TransactionInsightDataItemInQuery {
    queryIndex: number;
    queryName: string;
    transaction: TransactionInsightDataItem;
}

export interface InsightsExplorerTransactionStatisticData {
    totalCount: number;
    totalAmount: BigDecimal;
    totalIncome: BigDecimal;
    totalExpense: BigDecimal;
    netIncome: BigDecimal;
    averageAmount: BigDecimal;
    medianAmount: BigDecimal;
    minimumAmount: BigDecimal;
    maximumAmount: BigDecimal;
    p90Amount: BigDecimal;
    range: BigDecimal;
    interquartileRange: BigDecimal;
    medianToMeanRatio?: BigDecimal;
    top5AmountShare?: BigDecimal;
    transactionsFor80PercentAmount?: BigDecimal;
    variance?: BigDecimal;
    standardDeviation?: BigDecimal;
    coefficientOfVariation?: BigDecimal;
}
