import type { PartialRecord } from '@/core/base.ts';
import type { CategoricalChartSourceDataItem } from '@/core/chart.ts';
import type { ColorValue } from '@/core/color.ts';
import type { Year1BasedMonth, YearMonthDay } from '@/core/datetime.ts';
import type { IconType } from '@/core/icon.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import type { Account, AccountInfoResponse } from '@/models/account.ts';
import type { TransactionCategory, TransactionCategoryInfoResponse } from '@/models/transaction_category.ts';
import type { TransactionTagInfoResponse } from '@/models/transaction_tag.ts';
import type { TransactionGeoLocationResponse, TransactionInfoResponse } from './requests.ts';
import type { TransactionAmountsRequestType, TransactionPageWrapper, TransactionStatisticResponseItem } from './statistics.ts';

export interface SortableTransactionStatisticDataItem {
    readonly name: string;
    readonly displayOrders: number[];
    readonly value: BigDecimal;
}

export interface TransactionStatisticResponseItemWithInfo extends TransactionStatisticResponseItem {
    categoryId: string;
    accountId: string;
    relatedAccountId?: string;
    amount: string;
    account?: Account;
    primaryAccount?: Account;
    relatedAccount?: Account;
    relatedPrimaryAccount?: Account;
    relatedAccountType?: number;
    category?: TransactionCategory;
    primaryCategory?: TransactionCategory;
    amountInDefaultCurrency: BigDecimal | null;
}

export interface TransactionStatisticResponseWithInfo {
    readonly startTime: number;
    readonly endTime: number;
    readonly items: TransactionStatisticResponseItemWithInfo[];
}

export interface TransactionStatisticTrendsResponseItemWithInfo {
    readonly year: number;
    readonly month: number; // 1-based (1 = January, 12 = December)
    readonly items: TransactionStatisticResponseItemWithInfo[];
}

export interface TransactionStatisticAssetTrendsResponseItemWithInfo {
    readonly year: number;
    readonly month: number; // 1-based (1 = January, 12 = December)
    readonly day: number;
    readonly items: TransactionStatisticResponseItemWithInfo[];
}

export type TransactionStatisticDataItemType = 'category' | 'account' | 'total';

export interface TransactionStatisticDataItemBase extends SortableTransactionStatisticDataItem, CategoricalChartSourceDataItem {
    readonly name: string;
    readonly type: TransactionStatisticDataItemType;
    readonly id: string;
    readonly icon: string;
    readonly iconType: IconType;
    readonly color: ColorValue;
    readonly hidden: boolean;
    readonly displayOrders: number[];
    readonly value: BigDecimal;
}

export interface TransactionCategoricalOverviewAnalysisData {
    readonly totalIncome: BigDecimal;
    readonly totalExpense: BigDecimal;
    readonly items: TransactionCategoricalOverviewAnalysisDataItem[];
}

export enum TransactionCategoricalOverviewAnalysisDataItemType {
    IncomeByPrimaryCategory = 'incomeByPrimaryCategory',
    IncomeBySecondaryCategory = 'incomeBySecondaryCategory',
    IncomeByAccount = 'incomeByAccount',
    ExpenseByAccount = 'expenseByAccount',
    NetCashFlow = 'netCashFlow',
    ExpenseBySecondaryCategory = 'expenseBySecondaryCategory',
    ExpenseByPrimaryCategory = 'expenseByPrimaryCategory'
}

export interface TransactionCategoricalOverviewAnalysisDataItem extends SortableTransactionStatisticDataItem {
    readonly id: string;
    readonly name: string;
    readonly type: TransactionCategoricalOverviewAnalysisDataItemType;
    readonly displayOrders: number[];
    readonly hidden: boolean;
    readonly inflows: TransactionCategoricalOverviewAnalysisDataItemOutflowItem[];
    readonly outflows: TransactionCategoricalOverviewAnalysisDataItemOutflowItem[];
    value: BigDecimal;
    totalNonNegativeAmount: BigDecimal;
    includeInPercent?: boolean;
    percent?: number;
}

export interface TransactionCategoricalOverviewAnalysisDataItemOutflowItem {
    readonly relatedItem: TransactionCategoricalOverviewAnalysisDataItem;
    amount: BigDecimal;
}

export interface TransactionCategoricalAnalysisData {
    readonly value: BigDecimal;
    readonly items: TransactionCategoricalAnalysisDataItem[];
}

export interface TransactionCategoricalAnalysisDataItem extends Record<string, unknown>, TransactionStatisticDataItemBase {
    readonly percent: number;
    readonly originalValue?: BigDecimal;
    readonly originalCurrency?: string;
}

export interface TransactionTrendsAnalysisData {
    readonly items: TransactionTrendsAnalysisDataItem[];
}

export interface TransactionTrendsAnalysisDataItem extends Record<string, unknown>, TransactionStatisticDataItemBase {
    readonly items: TransactionTrendsAnalysisDataAmount[];
}

export interface TransactionTrendsAnalysisDataAmount extends Record<string, unknown>, Year1BasedMonth {
    readonly year: number;
    readonly month1base: number;
    readonly value: BigDecimal;
}

export interface TransactionAssetTrendsAnalysisData {
    readonly items: TransactionAssetTrendsAnalysisDataItem[];
}

export interface TransactionAssetTrendsAnalysisDataItem extends Record<string, unknown>, TransactionStatisticDataItemBase {
    readonly items: TransactionAssetTrendsAnalysisDataAmount[];
}

export interface TransactionAssetTrendsAnalysisDataAmount extends Record<string, unknown>, YearMonthDay {
    readonly year: number;
    readonly month: number;
    readonly day: number;
    readonly value: BigDecimal;
}

export interface TransactionInsightDataItem extends TransactionInfoResponse {
    readonly id: string;
    readonly time: number;
    readonly utcOffset: number;
    readonly type: number;
    readonly primaryCategory: TransactionCategoryInfoResponse;
    readonly primaryCategoryName: string;
    readonly secondaryCategory: TransactionCategoryInfoResponse;
    readonly secondaryCategoryName: string;
    readonly sourceAccount: AccountInfoResponse;
    readonly sourceAccountName: string;
    readonly destinationAccount?: AccountInfoResponse;
    readonly destinationAccountName?: string;
    readonly sourceAmount: number;
    readonly destinationAmount: number;
    readonly hideAmount: boolean;
    readonly tags: TransactionTagInfoResponse[];
    readonly comment: string;
    readonly geoLocation?: TransactionGeoLocationResponse;
}

export interface TransactionInsightDataItemWithQueryIndexes extends TransactionInsightDataItem {
    readonly queryIndexes: number[];
}

export type TransactionAmountsResponse = PartialRecord<TransactionAmountsRequestType, TransactionAmountsResponseItem>;

export interface TransactionAmountsResponseItem {
    readonly startTime: number;
    readonly endTime: number;
    readonly amounts: TransactionAmountsResponseItemAmountInfo[];
}

export interface TransactionAmountsResponseItemAmountInfo {
    readonly currency: string;
    readonly incomeAmount: string;
    readonly expenseAmount: string;
}

export interface TransactionDailyAmountsRequest {
    readonly startTime: number;
    readonly endTime: number;
    readonly useTransactionTimezone: boolean;
    readonly excludeAccountIds: string[];
    readonly excludeCategoryIds: string[];
    readonly tagFilter: string;
}

export interface TransactionDailyAmountsResponseItem {
    readonly date: string;
    readonly amounts: TransactionAmountsResponseItemAmountInfo[];
}

export type TransactionOverviewData = PartialRecord<TransactionAmountsRequestType, TransactionOverviewDataItem>;

export type TransactionOverviewDisplayTime = PartialRecord<TransactionAmountsRequestType, TransactionOverviewDisplayTimeItem>;

export interface TransactionOverviewDisplayTimeItem {
    readonly displayTime?: string;
    readonly startTime?: string;
    readonly endTime?: string;
}

export interface TransactionOverviewDataItem {
    readonly valid: boolean;
    readonly incomeAmount: BigDecimal;
    readonly expenseAmount: BigDecimal;
    readonly incompleteIncomeAmount: boolean;
    readonly incompleteExpenseAmount: boolean;
    readonly amounts?: TransactionAmountsResponseItemAmountInfo[];
}

export interface TransactionMonthlyIncomeAndExpenseData {
    readonly monthStartTime: number;
    readonly incomeAmount: BigDecimal;
    readonly expenseAmount: BigDecimal;
    readonly incompleteIncomeAmount: boolean;
    readonly incompleteExpenseAmount: boolean;
}

export const EMPTY_TRANSACTION_RESULT: TransactionPageWrapper = {
    items: [],
    totalCount: 0
}
