import { type CodeEditorExtraLib } from '@/components/desktop/CodeEditor.vue';
import { computed } from 'vue';
import type { useCustomChartState } from './state.ts';

interface Context {
    tt: ReturnType<typeof useCustomChartState>['tt'];
}

export function useCustomChartSamples(context: Context) {
    const { tt } = context;

    const customChartEditorExtraLibs = computed<CodeEditorExtraLib[]>(() => [{
        filePath: 'inmemory://model/ezbookkeeping-custom-chart.d.ts',
        content: `
/** ${tt('sample.insightsExplorerCustomChart.customChartTransactionDescription')} */
interface CustomChartTransaction {
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.id')} */
    id: string;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.unixTime')} */
    unixTime: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.year')} */
    year: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.month')} */
    month: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.dayOfMonth')} */
    dayOfMonth: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.dayOfWeek')} */
    dayOfWeek: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.hour')} */
    hour: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.minute')} */
    minute: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.second')} */
    second: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.utcOffset')} */
    utcOffset: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.type')} */
    type: string;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.primaryCategoryName')} */
    primaryCategoryName: string;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.secondaryCategoryName')} */
    secondaryCategoryName: string;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.sourceAccountName')} */
    sourceAccountName: string;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.destinationAccountName')} */
    destinationAccountName?: string;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.sourceCurrency')} */
    sourceCurrency: string;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.destinationCurrency')} */
    destinationCurrency?: string;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.sourceAmount')} */
    sourceAmount: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.destinationAmount')} */
    destinationAmount: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.sourceAmountInDefaultCurrency')} */
    sourceAmountInDefaultCurrency?: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.destinationAmountInDefaultCurrency')} */
    destinationAmountInDefaultCurrency?: number;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.tagNames')} */
    tagNames: string[];
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.geoLocation')} */
    geoLocation?: {
        latitude: number;
        longitude: number;
    };
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.comment')} */
    comment: string;
    /** ${tt('sample.insightsExplorerCustomChart.transactionField.inQueries')} */
    inQueries: string[];
}

/** ${tt('sample.insightsExplorerCustomChart.transactionTypeDescription')} */
declare const TransactionType: Readonly<{
    /** ${tt('Income')} */
    Income: string;
    /** ${tt('Expense')} */
    Expense: string;
    /** ${tt('Transfer')} */
    Transfer: string;
}>;

/** ${tt('sample.insightsExplorerCustomChart.chartColorsDescription')} */
declare const ChartColors: readonly string[];`
    }]);

    const sampleScript = computed<string>(() => `// ${tt('sample.insightsExplorerCustomChart.headerComment')}
/**
 * ${tt('sample.insightsExplorerCustomChart.functionDescription')}
 *
 * @param {CustomChartTransaction[]} transactions ${tt('sample.insightsExplorerCustomChart.transactionsFieldsDescription')}
 * @returns {object} ${tt('sample.insightsExplorerCustomChart.functionReturnDescription')}
 */
function buildChartOptions(transactions) {
    const transactionsByDate = new Map();

    transactions.forEach(transaction => {
        const date = transaction.year + '-'
            + String(transaction.month).padStart(2, '0') + '-'
            + String(transaction.dayOfMonth).padStart(2, '0');
        const dailyTransactions = transactionsByDate.get(date) || [];
        dailyTransactions.push(transaction);
        transactionsByDate.set(date, dailyTransactions);
    });

    const dailyData = Array.from(transactionsByDate.entries())
        .map(([date, dailyTransactions]) => ({
            date: date,
            year: dailyTransactions[0].year,
            month: dailyTransactions[0].month,
            dayOfMonth: dailyTransactions[0].dayOfMonth,
            totalAmount: sumTransactionAmounts(dailyTransactions)
        }))
        .sort((a, b) => a.date.localeCompare(b.date));

    const dailyDataMap = new Map(dailyData.map(item => [item.date, item]));
    const dates = dailyData.map(item => item.date);
    const totalAmounts = dailyData.map(item => item.totalAmount);

    const averageAmount = totalAmounts.length > 0
        ? totalAmounts.reduce((sum, amount) => sum + amount, 0) / totalAmounts.length
        : 0;

    const dayOverDayRates = dailyData.map(item => {
        const previousDate = new Date(Date.UTC(item.year, item.month - 1, item.dayOfMonth));
        previousDate.setUTCDate(previousDate.getUTCDate() - 1);

        const previousDateKey = previousDate.getUTCFullYear() + '-'
            + String(previousDate.getUTCMonth() + 1).padStart(2, '0') + '-'
            + String(previousDate.getUTCDate()).padStart(2, '0');
        const previousDayData = dailyDataMap.get(previousDateKey);

        if (!previousDayData || previousDayData.totalAmount === 0) {
            return null;
        }

        return (item.totalAmount - previousDayData.totalAmount) / previousDayData.totalAmount * 100;
    });

    const hasDayOverDayRates = dayOverDayRates.some(rate => rate !== null);
    const valueDimensions = [
        { name: 'date', type: 'ordinal', tooltip: false },
        { name: 'value', type: 'float', tooltip: false },
        { name: 'formattedValue', type: 'ordinal' }
    ];

    const series = [
        {
            name: ${JSON.stringify(tt('Total Amount'))},
            type: 'bar',
            yAxisIndex: 0,
            barMaxWidth: 40,
            itemStyle: { color: ChartColors[0] },
            dimensions: valueDimensions,
            encode: { x: 'date', y: 'formattedValue', tooltip: ['formattedValue'] },
            data: dailyData.map(item => [item.date, item.totalAmount, (item.totalAmount / 100).toFixed(2)])
        },
        {
            name: ${JSON.stringify(tt('Average Amount'))},
            type: 'line',
            yAxisIndex: 0,
            symbol: 'none',
            lineStyle: { type: 'dashed' },
            itemStyle: { color: ChartColors[1] },
            dimensions: valueDimensions,
            encode: { x: 'date', y: 'formattedValue', tooltip: ['formattedValue'] },
            data: dailyData.map(item => [item.date, averageAmount, (averageAmount / 100).toFixed(2)])
        }
    ];

    if (hasDayOverDayRates) {
        series.push({
            name: ${JSON.stringify(tt('Day-over-Day'))},
            type: 'line',
            yAxisIndex: 1,
            smooth: true,
            connectNulls: false,
            itemStyle: { color: ChartColors[2] },
            dimensions: valueDimensions,
            encode: { x: 'date', y: 'value', tooltip: ['formattedValue'] },
            data: dailyData.flatMap((item, index) => {
                const rate = dayOverDayRates[index];
                return rate !== null ? [[item.date, rate, rate.toFixed(2) + '%']] : [];
            })
        });
    }

    return {
        tooltip: {
            trigger: 'axis',
            axisPointer: { type: 'cross' }
        },
        legend: {
            top: 0,
            type: 'scroll'
        },
        grid: {
            top: 50,
            right: 50,
            bottom: 30,
            left: 50
        },
        xAxis: {
            type: 'category',
            data: dates
        },
        yAxis: [
            {
                type: 'value',
                name: ${JSON.stringify(tt('Amount'))},
                splitLine: {
                    lineStyle: { color: '#e1e6f2' }
                }
            },
            {
                type: 'value',
                name: ${JSON.stringify(tt('Day-over-Day'))},
                axisLabel: { formatter: '{value}%' },
                splitLine: { show: false }
            }
        ],
        series: series
    };
}

/**
 * ${tt('sample.insightsExplorerCustomChart.sumTransactionAmountsDescription')}
 *
 * @param {CustomChartTransaction[]} transactions ${tt('sample.insightsExplorerCustomChart.transactionsFieldsDescription')}
 * @returns {number} ${tt('sample.insightsExplorerCustomChart.sumTransactionAmountsReturnDescription')}
 */
function sumTransactionAmounts(transactions) {
    if (!Array.isArray(transactions) || transactions.length < 1) {
        return 0;
    }

    return transactions.reduce((sum, transaction) => {
        if (typeof transaction.sourceAmountInDefaultCurrency === 'number') {
            return sum + transaction.sourceAmountInDefaultCurrency;
        }

        return sum;
    }, 0);
}

/**
 * ${tt('sample.insightsExplorerCustomChart.allTransactionFieldsDescription')}
 * {string} id - ${tt('sample.insightsExplorerCustomChart.transactionField.id')}
 * {number} unixTime - ${tt('sample.insightsExplorerCustomChart.transactionField.unixTime')}
 * {number} year - ${tt('sample.insightsExplorerCustomChart.transactionField.year')}
 * {number} month - ${tt('sample.insightsExplorerCustomChart.transactionField.month')}
 * {number} dayOfMonth - ${tt('sample.insightsExplorerCustomChart.transactionField.dayOfMonth')}
 * {number} dayOfWeek - ${tt('sample.insightsExplorerCustomChart.transactionField.dayOfWeek')}
 * {number} hour - ${tt('sample.insightsExplorerCustomChart.transactionField.hour')}
 * {number} minute - ${tt('sample.insightsExplorerCustomChart.transactionField.minute')}
 * {number} second - ${tt('sample.insightsExplorerCustomChart.transactionField.second')}
 * {number} utcOffset - ${tt('sample.insightsExplorerCustomChart.transactionField.utcOffset')}
 * {string} type - ${tt('sample.insightsExplorerCustomChart.transactionField.type')}
 * {string} primaryCategoryName - ${tt('sample.insightsExplorerCustomChart.transactionField.primaryCategoryName')}
 * {string} secondaryCategoryName - ${tt('sample.insightsExplorerCustomChart.transactionField.secondaryCategoryName')}
 * {string} sourceAccountName - ${tt('sample.insightsExplorerCustomChart.transactionField.sourceAccountName')}
 * {string} destinationAccountName - ${tt('sample.insightsExplorerCustomChart.transactionField.destinationAccountName')}
 * {string} sourceCurrency - ${tt('sample.insightsExplorerCustomChart.transactionField.sourceCurrency')}
 * {string} destinationCurrency - ${tt('sample.insightsExplorerCustomChart.transactionField.destinationCurrency')}
 * {number} sourceAmount - ${tt('sample.insightsExplorerCustomChart.transactionField.sourceAmount')}
 * {number} destinationAmount - ${tt('sample.insightsExplorerCustomChart.transactionField.destinationAmount')}
 * {number} sourceAmountInDefaultCurrency - ${tt('sample.insightsExplorerCustomChart.transactionField.sourceAmountInDefaultCurrency')}
 * {number} destinationAmountInDefaultCurrency - ${tt('sample.insightsExplorerCustomChart.transactionField.destinationAmountInDefaultCurrency')}
 * {string[]} tagNames - ${tt('sample.insightsExplorerCustomChart.transactionField.tagNames')}
 * {string} geoLocation - ${tt('sample.insightsExplorerCustomChart.transactionField.geoLocation')}
 * {string} comment - ${tt('sample.insightsExplorerCustomChart.transactionField.comment')}
 * {string[]} inQueries - ${tt('sample.insightsExplorerCustomChart.transactionField.inQueries')}
 */

/**
 * ${tt('sample.insightsExplorerCustomChart.availableGlobalConstants')}
 * {TransactionType} TransactionType - ${tt('sample.insightsExplorerCustomChart.transactionTypeDescription')}
 * {string[]} ChartColors - ${tt('sample.insightsExplorerCustomChart.chartColorsDescription')}
 */`);

    return { customChartEditorExtraLibs, sampleScript };
}
