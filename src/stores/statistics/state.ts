import {
    CategoricalChartType,
    ChartDataType,
    ChartSortingType,
    DEFAULT_ASSET_TRENDS_CHART_DATA_RANGE,
    DEFAULT_CATEGORICAL_CHART_DATA_RANGE,
    DEFAULT_TREND_CHART_DATA_RANGE,
    TrendChartType
} from '@/core/statistics.ts';
import { KeywordMatchMode } from '@/core/text.ts';
import type { TransactionStatisticAssetTrendsResponseItem, TransactionStatisticResponse, TransactionStatisticTrendsResponseItem } from '@/models/transaction.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useUserStore } from '@/stores/user.ts';
import type { ComputedRef, Ref, UnwrapRef } from 'vue';
import { computed, ref } from 'vue';
import type { TransactionStatisticsFilter } from './types.ts';
interface StatisticsState {
    settingsStore: ReturnType<typeof useSettingsStore>;
    userStore: ReturnType<typeof useUserStore>;
    accountsStore: ReturnType<typeof useAccountsStore>;
    transactionCategoriesStore: ReturnType<typeof useTransactionCategoriesStore>;
    exchangeRatesStore: ReturnType<typeof useExchangeRatesStore>;
    transactionStatisticsFilter: Ref<UnwrapRef<TransactionStatisticsFilter>>;
    transactionCategoryStatisticsData: Ref<UnwrapRef<TransactionStatisticResponse | null>>;
    transactionCategoryTrendsData: Ref<UnwrapRef<TransactionStatisticTrendsResponseItem[]>>;
    transactionAssetTrendsData: Ref<UnwrapRef<TransactionStatisticAssetTrendsResponseItem[]>>;
    transactionStatisticsStateInvalid: Ref<UnwrapRef<boolean>>;
    categoricalAnalysisChartDataCategory: ComputedRef<string>;
}



// Owns state for the statistics store.
export function useStatisticsState(): StatisticsState  {
    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const transactionStatisticsFilter = ref<TransactionStatisticsFilter>({
        chartDataType: ChartDataType.Default.type,
        categoricalChartType: CategoricalChartType.Default.type,
        categoricalChartDateType: DEFAULT_CATEGORICAL_CHART_DATA_RANGE.type,
        categoricalChartStartTime: 0,
        categoricalChartEndTime: 0,
        trendChartType: TrendChartType.Default.type,
        trendChartDateType: DEFAULT_TREND_CHART_DATA_RANGE.type,
        trendChartStartYearMonth: '',
        trendChartEndYearMonth: '',
        assetTrendsChartType: TrendChartType.Default.type,
        assetTrendsChartDateType: DEFAULT_ASSET_TRENDS_CHART_DATA_RANGE.type,
        assetTrendsChartStartTime: 0,
        assetTrendsChartEndTime: 0,
        filterAccountIds: {},
        filterCategoryIds: {},
        tagFilter: '',
        keyword: '',
        matchMode: KeywordMatchMode.Default.type,
        sortingType: ChartSortingType.Default.type
    });

    const transactionCategoryStatisticsData = ref<TransactionStatisticResponse | null>(null);

    const transactionCategoryTrendsData = ref<TransactionStatisticTrendsResponseItem[]>([]);

    const transactionAssetTrendsData = ref<TransactionStatisticAssetTrendsResponseItem[]>([]);

    const transactionStatisticsStateInvalid = ref<boolean>(true);

    const categoricalAnalysisChartDataCategory = computed<string>(() => {
        if (transactionStatisticsFilter.value.chartDataType === ChartDataType.OutflowsByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.InflowsByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeByAccount.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalAssets.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.AccountTotalLiabilities.type) {
            return 'account';
        } else if (transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseByPrimaryCategory.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.ExpenseBySecondaryCategory.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeByPrimaryCategory.type ||
            transactionStatisticsFilter.value.chartDataType === ChartDataType.IncomeBySecondaryCategory.type) {
            return 'category';
        } else {
            return '';
        }
    });

    return {
        settingsStore, userStore, accountsStore, transactionCategoriesStore, exchangeRatesStore, transactionStatisticsFilter,
        transactionCategoryStatisticsData, transactionCategoryTrendsData, transactionAssetTrendsData,
        transactionStatisticsStateInvalid, categoricalAnalysisChartDataCategory,
    };
}
