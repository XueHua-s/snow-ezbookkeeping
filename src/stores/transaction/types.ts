import type { TextualYearMonth } from '@/core/datetime.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import type { Transaction } from '@/models/transaction.ts';

export interface TransactionListPartialFilter {
    dateType?: number;
    maxTime?: number;
    minTime?: number;
    type?: number;
    categoryIds?: string;
    accountIds?: string;
    tagFilter?: string;
    amountFilter?: string;
    keyword?: string;
    matchMode?: number;
}

export interface TransactionListFilter extends TransactionListPartialFilter {
    dateType: number;
    maxTime: number;
    minTime: number;
    type: number;
    categoryIds: string;
    accountIds: string;
    tagFilter: string;
    amountFilter: string;
    keyword: string;
    matchMode: number;
}

export interface TransactionTotalAmount {
    transactionCount: number;
    expense: BigDecimal;
    incompleteExpense: boolean;
    income: BigDecimal;
    incompleteIncome: boolean;
}

export interface TransactionDailyTotalAmounts {
    inflowOutflowDailyTotalAmounts: Record<string, TransactionTotalAmount>;
    incomeExpenseDailyTotalAmounts: Record<string, TransactionTotalAmount>;
}

export interface TransactionMonthList {
    readonly year: number;
    readonly month: number; // 1-based (1 = January, 12 = December)
    readonly yearDashMonth: TextualYearMonth;
    opened: boolean;
    readonly items: Transaction[];
    readonly inflowOutflowTotalAmount: TransactionTotalAmount;
    readonly inflowOutflowDailyTotalAmounts: Record<string, TransactionTotalAmount>;
    readonly incomeExpenseTotalAmount: TransactionTotalAmount;
    readonly incomeExpenseDailyTotalAmounts: Record<string, TransactionTotalAmount>;
}
