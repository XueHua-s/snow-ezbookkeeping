import DataExportDialog from '@/components/desktop/DataExportDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import TrendsChart from '@/components/desktop/TrendsChart.vue';
import { type TextualYearMonth } from '@/core/datetime.ts';

export type SnackBarType = InstanceType<typeof SnackBar>;

export type TrendsChartType = InstanceType<typeof TrendsChart>;

export type DataExportDialogType = InstanceType<typeof DataExportDialog>;

export interface TransactionStatisticsProps {
    initAnalysisType?: string,
    initChartDataType?: string,
    initChartType?: string,
    initChartDateType?: string,
    initStartTime?: TextualYearMonth | '',
    initEndTime?: TextualYearMonth | '',
    initFilterAccountIds?: string,
    initFilterCategoryIds?: string,
    initTagFilter?: string,
    initKeyword?: string;
    initMatchMode?: string;
    initSortingType?: string,
    initTrendDateAggregationType?: string
    initAssetTrendsDateAggregationType?: string
}
