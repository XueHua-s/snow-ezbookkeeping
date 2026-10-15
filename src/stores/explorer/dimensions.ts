import { AccountCategory } from '@/core/account.ts';
import {
    TransactionExplorerDataDimension,
    TransactionExplorerValueMetric
} from '@/core/explorer.ts';
import { type BigDecimal, NumeralSystem } from '@/core/numeral.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    isDefined,
    isNumber
} from '@/lib/common.ts';
import {
    getFiscalYearFromUnixTime,
    getUtcOffsetByUtcOffsetMinutes,
    parseDateTimeFromUnixTime,
    parseDateTimeFromUnixTimeWithTimezoneOffset
} from '@/lib/datetime.ts';
import {
    parseBigDecimal
} from '@/lib/numeral.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { useExplorerState } from './state.ts';
import { type CategoriedInfo, TransactionExplorerDimensionType } from './types.ts';

interface Context {
    userStore: ReturnType<typeof useExplorerState>['userStore'];
    currentExploration: ReturnType<typeof useExplorerState>['currentExploration'];
    accountsStore: ReturnType<typeof useExplorerState>['accountsStore'];
    settingsStore: ReturnType<typeof useExplorerState>['settingsStore'];
    currencyDisplayOrders: ReturnType<typeof useExplorerState>['currencyDisplayOrders'];
    exchangeRatesStore: ReturnType<typeof useExplorerState>['exchangeRatesStore'];
}

// Owns dimensions for the explorer store.
export function useExplorerDimensions(context: Context) {
    const { userStore, currentExploration, accountsStore, settingsStore, currencyDisplayOrders, exchangeRatesStore } = context;

    function getDataCategoryInfo(timezoneUsedForDateRange: number, dimension: TransactionExplorerDataDimension, sourceAmountRanges: BigDecimal[] | undefined, destinationAmountRanges: BigDecimal[] | undefined, queryName: string, queryIndex: number, transaction: TransactionInsightDataItem): CategoriedInfo {
        const defaultCurrency = userStore.currentUserDefaultCurrency;
        let transactionTimeUtfOffset: number | undefined = undefined;

        if (timezoneUsedForDateRange === TimezoneTypeForStatistics.TransactionTimezone.type) {
            transactionTimeUtfOffset = transaction.utcOffset;
        }

        if (dimension === TransactionExplorerDataDimension.None) {
            const valueMetric = TransactionExplorerValueMetric.valueOf(currentExploration.value.valueMetric);
            return {
                categoryName: valueMetric?.name ?? 'Unknown',
                categoryNameNeedI18n: true,
                categoryId: 'none',
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [1]
            };
        } else if (dimension === TransactionExplorerDataDimension.Query) {
            if (queryName) {
                return {
                    categoryName: queryName,
                    categoryId: (queryIndex + 1).toString(10),
                    categoryIdType: TransactionExplorerDimensionType.Other,
                    categoryDisplayOrders: [queryIndex + 1]
                };
            } else {
                return {
                    categoryName: `format.misc.queryIndex`,
                    categoryNameNeedI18n: true,
                    categoryNameI18nParameters: {
                        index: queryIndex + 1
                    },
                    categoryId: (queryIndex + 1).toString(10),
                    categoryIdType: TransactionExplorerDimensionType.Other,
                    categoryDisplayOrders: [queryIndex + 1]
                };
            }
        } else if (dimension === TransactionExplorerDataDimension.DateTime) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);
            const textualDateTime = `${dateTime.getGregorianCalendarYearDashMonthDashDay()} ${dateTime.getHour().toString(10).padStart(2, NumeralSystem.WesternArabicNumerals.digitZero)}:${dateTime.getMinute().toString(10).padStart(2, NumeralSystem.WesternArabicNumerals.digitZero)}:${dateTime.getSecond().toString(10).padStart(2, NumeralSystem.WesternArabicNumerals.digitZero)}`;

            return {
                categoryName: textualDateTime,
                categoryId: textualDateTime,
                categoryIdType: TransactionExplorerDimensionType.DateTime,
                categoryDisplayOrders: [dateTime.getUnixTime()]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByYearMonthDay) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);
            const yearMonthDay = dateTime.getGregorianCalendarYearDashMonthDashDay();

            return {
                categoryName: yearMonthDay,
                categoryId: yearMonthDay,
                categoryIdType: TransactionExplorerDimensionType.YearMonthDay,
                categoryDisplayOrders: [dateTime.getUnixTime()]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByYearMonth) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);
            const yearMonth = dateTime.getGregorianCalendarYearDashMonth();

            return {
                categoryName: yearMonth,
                categoryId: yearMonth,
                categoryIdType: TransactionExplorerDimensionType.YearMonth,
                categoryDisplayOrders: [dateTime.getUnixTime()]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByYearQuarter) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);
            const yearQuarter = `${dateTime.getGregorianCalendarYear().toString(10)}-${dateTime.getGregorianCalendarQuarter().toString(10)}`;

            return {
                categoryName: yearQuarter,
                categoryId: yearQuarter,
                categoryIdType: TransactionExplorerDimensionType.YearQuarter,
                categoryDisplayOrders: [dateTime.getUnixTime()]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByYear) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);

            return {
                categoryName: dateTime.getGregorianCalendarYear().toString(10),
                categoryId: dateTime.getGregorianCalendarYear().toString(10),
                categoryIdType: TransactionExplorerDimensionType.Year,
                categoryDisplayOrders: [dateTime.getUnixTime()]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByFiscalYear) {
            const fiscalYear = getFiscalYearFromUnixTime(transaction.time, userStore.currentUserFiscalYearStart, transactionTimeUtfOffset).toString(10);

            return {
                categoryName: fiscalYear,
                categoryId: fiscalYear,
                categoryIdType: TransactionExplorerDimensionType.Year,
                categoryDisplayOrders: [parseInt(fiscalYear)]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByDayOfWeek) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);

            return {
                categoryName: dateTime.getWeekDay().name,
                categoryId: dateTime.getWeekDay().type.toString(10),
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [dateTime.getWeekDay().getDisplayOrder(userStore.currentUserFirstDayOfWeek)]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByDayOfMonth) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);

            return {
                categoryName: dateTime.getGregorianCalendarDay().toString(10),
                categoryId: dateTime.getGregorianCalendarDay().toString(10),
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [dateTime.getGregorianCalendarDay()]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByMonthOfYear) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);

            return {
                categoryName: dateTime.getGregorianCalendarMonth().toString(10),
                categoryId: dateTime.getGregorianCalendarMonth().toString(10),
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [dateTime.getGregorianCalendarMonth()]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByQuarterOfYear) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);

            return {
                categoryName: dateTime.getGregorianCalendarQuarter().toString(10),
                categoryId: dateTime.getGregorianCalendarQuarter().toString(10),
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [dateTime.getGregorianCalendarQuarter()]
            };
        } else if (dimension === TransactionExplorerDataDimension.DateTimeByHourOfDay) {
            const dateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);

            return {
                categoryName: dateTime.getHour().toString(10),
                categoryId: dateTime.getHour().toString(10),
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [dateTime.getHour()]
            };
        } else if (dimension === TransactionExplorerDataDimension.TimezoneOffset) {
            return {
                categoryName: getUtcOffsetByUtcOffsetMinutes(transaction.utcOffset),
                categoryId: transaction.utcOffset.toString(10),
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [transaction.utcOffset]
            };
        } else if (dimension === TransactionExplorerDataDimension.TransactionType) {
            let transactionTypeName = 'Unknown';

            if (transaction.type === TransactionType.ModifyBalance) {
                transactionTypeName = 'Modify Balance';
            } else if (transaction.type === TransactionType.Income) {
                transactionTypeName = 'Income';
            } else if (transaction.type === TransactionType.Expense) {
                transactionTypeName = 'Expense';
            } else if (transaction.type === TransactionType.Transfer) {
                transactionTypeName = 'Transfer';
            }

            return {
                categoryName: transactionTypeName,
                categoryNameNeedI18n: true,
                categoryId: transaction.type.toString(10),
                categoryIdType: TransactionExplorerDimensionType.TransactionType,
                categoryDisplayOrders: [transaction.type]
            };
        } else if (dimension === TransactionExplorerDataDimension.SourceAccount) {
            const primaryAccount = accountsStore.allAccountsMap[transaction.sourceAccount.parentId] ?? transaction.sourceAccount;
            const primaryAccountCategoryDisplayOrder: number = settingsStore.accountCategoryDisplayOrders[primaryAccount.category] || Number.MAX_SAFE_INTEGER;

            return {
                categoryName: transaction.sourceAccountName || 'Unknown',
                categoryNameNeedI18n: !transaction.sourceAccountName,
                categoryId: transaction.sourceAccountId || 'unknown',
                categoryIdType: TransactionExplorerDimensionType.Account,
                categoryDisplayOrders: [primaryAccountCategoryDisplayOrder, primaryAccount.displayOrder, transaction.sourceAccount.displayOrder]
            };
        } else if (dimension === TransactionExplorerDataDimension.SourceAccountCategory) {
            const accountCategory = AccountCategory.valueOf(transaction.sourceAccount.category);
            const accountCategoryDisplayOrder: number = settingsStore.accountCategoryDisplayOrders[accountCategory?.type ?? 0] || Number.MAX_SAFE_INTEGER;

            return {
                categoryName: accountCategory?.name || 'Unknown',
                categoryNameNeedI18n: true,
                categoryId: accountCategory?.type.toString(10) || 'unknown',
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [accountCategoryDisplayOrder]
            };
        } else if (dimension === TransactionExplorerDataDimension.SourceAccountCurrency) {
            return {
                categoryName: transaction.sourceAccount.currency || 'Unknown',
                categoryNameNeedI18n: !transaction.sourceAccount.currency,
                categoryId: transaction.sourceAccount.currency || 'unknown',
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [currencyDisplayOrders[transaction.sourceAccount.currency] || Number.MAX_SAFE_INTEGER]
            };
        }  else if (dimension === TransactionExplorerDataDimension.DestinationAccount) {
            const primaryAccount = accountsStore.allAccountsMap[transaction.destinationAccount?.parentId ?? ''] ?? transaction.destinationAccount;
            const primaryAccountCategoryDisplayOrder: number = settingsStore.accountCategoryDisplayOrders[primaryAccount?.category || 0] || Number.MAX_SAFE_INTEGER;

            return {
                categoryName: transaction.type === TransactionType.Transfer ? (transaction.destinationAccountName || 'Unknown') : 'None',
                categoryNameNeedI18n: transaction.type !== TransactionType.Transfer || !transaction.destinationAccountName,
                categoryId: transaction.type === TransactionType.Transfer ? (transaction.destinationAccountId || 'unknown') : 'none',
                categoryIdType: TransactionExplorerDimensionType.Account,
                categoryDisplayOrders: transaction.type === TransactionType.Transfer && primaryAccount && transaction.destinationAccount ? [primaryAccountCategoryDisplayOrder, primaryAccount.displayOrder, transaction.destinationAccount.displayOrder] : [0]
            };
        } else if (dimension === TransactionExplorerDataDimension.DestinationAccountCategory) {
            const accountCategory = transaction.type === TransactionType.Transfer && transaction.destinationAccount ? AccountCategory.valueOf(transaction.destinationAccount.category) : undefined;
            const accountCategoryDisplayOrder: number = settingsStore.accountCategoryDisplayOrders[accountCategory?.type ?? 0] || Number.MAX_SAFE_INTEGER;

            return {
                categoryName: transaction.type === TransactionType.Transfer ? (accountCategory?.name || 'Unknown') : 'None',
                categoryNameNeedI18n: true,
                categoryId: transaction.type === TransactionType.Transfer ? (accountCategory?.name || 'unknown') : 'none',
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: transaction.type === TransactionType.Transfer ? [accountCategoryDisplayOrder] : [Number.MAX_SAFE_INTEGER]
            };
        } else if (dimension === TransactionExplorerDataDimension.DestinationAccountCurrency) {
            return {
                categoryName: transaction.type === TransactionType.Transfer ? (transaction.destinationAccount?.currency || 'Unknown') : 'None',
                categoryNameNeedI18n: transaction.type !== TransactionType.Transfer || !transaction.destinationAccount?.currency,
                categoryId: transaction.type === TransactionType.Transfer ? (transaction.destinationAccount?.currency || 'unknown') : 'none',
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: transaction.type === TransactionType.Transfer ? [currencyDisplayOrders[transaction.destinationAccount?.currency ?? ''] || Number.MAX_SAFE_INTEGER] : [Number.MAX_SAFE_INTEGER]
            };
        } else if (dimension === TransactionExplorerDataDimension.PrimaryCategory) {
            return {
                categoryName: transaction.primaryCategory.name,
                categoryId: transaction.primaryCategory.id,
                categoryIdType: TransactionExplorerDimensionType.Category,
                categoryDisplayOrders: [transaction.primaryCategory.type, transaction.primaryCategory.displayOrder]
            };
        } else if (dimension === TransactionExplorerDataDimension.SecondaryCategory) {
            return {
                categoryName: transaction.secondaryCategory.name,
                categoryId: transaction.categoryId,
                categoryIdType: TransactionExplorerDimensionType.Category,
                categoryDisplayOrders: [transaction.primaryCategory.type, transaction.primaryCategory.displayOrder, transaction.secondaryCategory.displayOrder]
            };
        } else if (dimension === TransactionExplorerDataDimension.SourceAmount || dimension === TransactionExplorerDataDimension.DestinationAmount) {
            if (dimension === TransactionExplorerDataDimension.DestinationAmount && transaction.type !== TransactionType.Transfer) {
                return {
                    categoryName: 'None',
                    categoryNameNeedI18n: true,
                    categoryId: 'none',
                    categoryIdType: TransactionExplorerDimensionType.Other,
                    categoryDisplayOrders: [Number.MAX_SAFE_INTEGER]
                };
            }

            const amount: BigDecimal = dimension === TransactionExplorerDataDimension.SourceAmount ? parseBigDecimal(transaction.sourceAmount) : parseBigDecimal(transaction.destinationAmount);
            const account = dimension === TransactionExplorerDataDimension.SourceAmount ? transaction.sourceAccount : transaction.destinationAccount;
            let amountInDefaultCurrency: BigDecimal = amount;

            if (!account) {
                return {
                    categoryName: 'Unknown',
                    categoryNameNeedI18n: true,
                    categoryId: 'unknown',
                    categoryIdType: TransactionExplorerDimensionType.Other,
                    categoryDisplayOrders: [Number.MAX_SAFE_INTEGER]
                };
            }

            if (account.currency !== defaultCurrency) {
                const exchangedAmount = exchangeRatesStore.getExchangedAmount(amount, account.currency, defaultCurrency);

                if (exchangedAmount) {
                    amountInDefaultCurrency = exchangedAmount.truncate();
                } else {
                    return {
                        categoryName: 'Unknown',
                        categoryNameNeedI18n: true,
                        categoryId: 'unknown',
                        categoryIdType: TransactionExplorerDimensionType.Other,
                        categoryDisplayOrders: [Number.MAX_SAFE_INTEGER]
                    };
                }
            }

            return {
                categoryName: amountInDefaultCurrency.toString(),
                categoryId: amountInDefaultCurrency.toString(),
                categoryIdType: TransactionExplorerDimensionType.Amount,
                categoryDisplayOrders: [amountInDefaultCurrency.toDoubleNumber()]
            };
        } else if (dimension.isSourceAmountRange || dimension.isDestinationAmountRange) {
            const isSourceAmount = dimension.isSourceAmountRange;

            if (dimension.isDestinationAmountRange && transaction.type !== TransactionType.Transfer) {
                return {
                    categoryName: 'None',
                    categoryNameNeedI18n: true,
                    categoryId: 'none',
                    categoryIdType: TransactionExplorerDimensionType.Other,
                    categoryDisplayOrders: [Number.MAX_SAFE_INTEGER]
                };
            }

            const amount: BigDecimal = dimension.isSourceAmountRange ? parseBigDecimal(transaction.sourceAmount) : parseBigDecimal(transaction.destinationAmount);
            const account = dimension.isSourceAmountRange ? transaction.sourceAccount : transaction.destinationAccount;
            let amountInDefaultCurrency: BigDecimal = amount;

            if (!account) {
                return {
                    categoryName: 'Unknown',
                    categoryNameNeedI18n: true,
                    categoryId: 'unknown',
                    categoryIdType: TransactionExplorerDimensionType.Other,
                    categoryDisplayOrders: [Number.MAX_SAFE_INTEGER]
                };
            }

            if (account.currency !== defaultCurrency) {
                const exchangedAmount = exchangeRatesStore.getExchangedAmount(amount, account.currency, defaultCurrency);

                if (exchangedAmount) {
                    amountInDefaultCurrency = exchangedAmount.truncate();
                } else {
                    return {
                        categoryName: 'Unknown',
                        categoryNameNeedI18n: true,
                        categoryId: 'unknown',
                        categoryIdType: TransactionExplorerDimensionType.Other,
                        categoryDisplayOrders: [Number.MAX_SAFE_INTEGER]
                    };
                }
            }

            const amountRanges: BigDecimal[] = isSourceAmount ? (sourceAmountRanges ?? []) : (destinationAmountRanges ?? []);
            let matchAmountRangeMin: BigDecimal | undefined = undefined;
            let matchAmountRangeMax: BigDecimal | undefined = undefined;
            let matchAmountRangeIndex: number | undefined = undefined;

            for (let i = 1; i < amountRanges.length; i++) {
                const amountRangeMin = amountRanges[i - 1] as BigDecimal;
                const amountRangeMax = amountRanges[i] as BigDecimal;

                if (amountInDefaultCurrency.lessThan(amountRangeMin)) {
                    continue;
                }

                if (amountInDefaultCurrency.greaterThan(amountRangeMax)) {
                    continue;
                }

                if (i < amountRanges.length - 1 && amountInDefaultCurrency.equals(amountRangeMax)) {
                    continue;
                }

                matchAmountRangeMin = amountRangeMin;
                matchAmountRangeMax = amountRangeMax;
                matchAmountRangeIndex = i - 1;
            }

            if (matchAmountRangeMin && matchAmountRangeMax && isNumber(matchAmountRangeIndex)) {
                return {
                    categoryName: `${matchAmountRangeMin.toString()}|${matchAmountRangeMax.toString()}`,
                    categoryId: matchAmountRangeIndex.toString(10),
                    categoryIdType: TransactionExplorerDimensionType.Other,
                    categoryDisplayOrders: [matchAmountRangeIndex]
                };
            } else {
                return {
                    categoryName: 'Other',
                    categoryNameNeedI18n: true,
                    categoryId: 'other',
                    categoryIdType: TransactionExplorerDimensionType.Other,
                    categoryDisplayOrders: [Number.MAX_SAFE_INTEGER]
                };
            }
        } else {
            return {
                categoryName: '',
                categoryId: '',
                categoryIdType: TransactionExplorerDimensionType.Other,
                categoryDisplayOrders: [0]
            };
        }
    }

    return { getDataCategoryInfo };
}
