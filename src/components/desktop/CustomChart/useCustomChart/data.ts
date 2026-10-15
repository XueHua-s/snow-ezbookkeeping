import { itemAndIndex } from '@/core/base.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TransactionType } from '@/core/transaction.ts';
import { parseDateTimeFromUnixTimeWithTimezoneOffset } from '@/lib/datetime.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import type { TransactionExplorerQuery } from '@/models/explorer.ts';
import type { TransactionInsightDataItemWithQueryIndexes } from '@/models/transaction.ts';
import { computed } from 'vue';
import type { useCustomChartState } from './state.ts';
import { type CustomChartTransaction } from './types.ts';

interface Context {
    executingScript: ReturnType<typeof useCustomChartState>['executingScript'];
    tt: ReturnType<typeof useCustomChartState>['tt'];
    executionError: ReturnType<typeof useCustomChartState>['executionError'];
    chartOptions: ReturnType<typeof useCustomChartState>['chartOptions'];
    props: SetupProps<{
    disabled?: boolean;
    displayLayout: number;
    transactions: TransactionInsightDataItemWithQueryIndexes[];
    queries: TransactionExplorerQuery[];
    modelValue: string;
}>;
    formatNumberToLocalizedNumeralsWithoutDigitGrouping: ReturnType<typeof useCustomChartState>['formatNumberToLocalizedNumeralsWithoutDigitGrouping'];
    userStore: ReturnType<typeof useCustomChartState>['userStore'];
    exchangeRatesStore: ReturnType<typeof useCustomChartState>['exchangeRatesStore'];
}

export function useCustomChartData(context: Context) {
    const { executingScript, tt, executionError, chartOptions, props, formatNumberToLocalizedNumeralsWithoutDigitGrouping, userStore, exchangeRatesStore } = context;

    const displayChartData = computed<string>(() => {
        if (executingScript.value) {
            return tt('Executing Script...');
        } else if (executionError.value) {
            return `// ${executionError.value}`;
        } else if (chartOptions.value) {
            return JSON.stringify(chartOptions.value, null, 2);
        } else {
            return tt('No Preview Result');
        }
    });

    const customChartTransactions = computed<CustomChartTransaction[]>(() => {
        const queryNamesMap: Record<number, string> = {};

        if (props.queries) {
            for (const [query, index] of itemAndIndex(props.queries)) {
                let queryName = query.name;

                if (!queryName) {
                    queryName = tt('format.misc.queryIndex', { index: formatNumberToLocalizedNumeralsWithoutDigitGrouping(index + 1) });
                }

                queryNamesMap[index] = queryName;
            }
        }

        return props.transactions.map(transaction => {
            const transactionTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transaction.utcOffset);
            const defaultCurrency = userStore.currentUserDefaultCurrency;
            let transactionType: string = '';
            let sourceAmountInDefaultCurrency: BigDecimal | undefined = parseBigDecimal(transaction.sourceAmount);
            let destinationAmountInDefaultCurrency = transaction.type === TransactionType.Transfer && transaction.destinationAccount ? parseBigDecimal(transaction.destinationAmount) : undefined;

            if (transaction.sourceAccount.currency !== defaultCurrency) {
                sourceAmountInDefaultCurrency = exchangeRatesStore.getExchangedAmount(sourceAmountInDefaultCurrency, transaction.sourceAccount.currency, defaultCurrency)?.truncate();
            }

            if (destinationAmountInDefaultCurrency && transaction.destinationAccount && transaction.destinationAccount.currency !== defaultCurrency) {
                destinationAmountInDefaultCurrency = exchangeRatesStore.getExchangedAmount(destinationAmountInDefaultCurrency, transaction.destinationAccount.currency, defaultCurrency)?.truncate();
            }

            if (transaction.type === TransactionType.Income) {
                transactionType = tt('Income');
            } else if (transaction.type === TransactionType.Expense) {
                transactionType = tt('Expense');
            } else if (transaction.type === TransactionType.Transfer) {
                transactionType = tt('Transfer');
            } else if (transaction.type === TransactionType.ModifyBalance) {
                transactionType = tt('Modify Balance');
            } else {
                transactionType = tt('Unknown');
            }

            const finalTransaction: CustomChartTransaction = {
                id: transaction.id,
                unixTime: transaction.time,
                year: transactionTime.getGregorianCalendarYear(),
                month: transactionTime.getGregorianCalendarMonth(),
                dayOfMonth: transactionTime.getGregorianCalendarDay(),
                dayOfWeek: transactionTime.getWeekDay().type,
                hour: transactionTime.getHour(),
                minute: transactionTime.getMinute(),
                second: transactionTime.getSecond(),
                utcOffset: transaction.utcOffset,
                type: transactionType,
                primaryCategoryName: transaction.primaryCategoryName,
                secondaryCategoryName: transaction.secondaryCategoryName,
                sourceAccountName: transaction.sourceAccountName,
                destinationAccountName: transaction.destinationAccountName,
                sourceCurrency: transaction.sourceAccount.currency,
                destinationCurrency: transaction.destinationAccount?.currency,
                sourceAmount: transaction.sourceAmount,
                destinationAmount: transaction.destinationAmount,
                sourceAmountInDefaultCurrency: sourceAmountInDefaultCurrency?.toDoubleNumber(),
                destinationAmountInDefaultCurrency: destinationAmountInDefaultCurrency?.toDoubleNumber(),
                tagNames: transaction.tags?.map(tag => tag.name) ?? [],
                geoLocation: transaction.geoLocation ? {
                    latitude: transaction.geoLocation.latitude,
                    longitude: transaction.geoLocation.longitude
                } : undefined,
                comment: transaction.comment || '',
                inQueries: transaction.queryIndexes.map(index => queryNamesMap[index] ?? tt('format.misc.queryIndex', { index: formatNumberToLocalizedNumeralsWithoutDigitGrouping(index + 1) }))
            };

            return finalTransaction;
        });
    });

    return { displayChartData, customChartTransactions };
}
