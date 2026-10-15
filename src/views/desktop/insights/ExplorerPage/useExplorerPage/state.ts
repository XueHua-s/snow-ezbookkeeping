import type { TypeAndDisplayName } from '@/core/base.ts';
import { type LocalizedDateRange, type WeekDayValue, DateRange, DateRangeScene } from '@/core/datetime.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import { isEquals } from '@/lib/common.ts';
import {
    parseDateTimeFromUnixTime
} from '@/lib/datetime.ts';
import { useI18n } from '@/locales/helpers.ts';
import { InsightsExplorer, InsightsExplorerBasicInfo } from '@/models/explorer.ts';
import { type TransactionInsightDataItem } from '@/models/transaction.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { type TransactionExplorerFilter, useExplorersStore } from '@/stores/explorer.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useUserStore } from '@/stores/user.ts';
import {
    mdiHomeClockOutline,
    mdiInvoiceTextClockOutline
} from '@mdi/js';
import { computed, ref, useTemplateRef } from 'vue';
import { useRouter } from 'vue-router';
import {
    type ConfirmDialogType, type DataExportDialogType, type EditDialogType, type ExplorerChangeDisplayOrderDialogType,
    type ExplorerChartTabType, type ExplorerDataTableTabType, type ExplorerPageTabType, type JsonExportDialogType,
    type JsonImportDialogType, type RenameDialogType, type SnackBarType,
} from './types.ts';


export function useExplorerPageState() {
    const router = useRouter();

    const {
        tt,
        getAllDateRanges,
        getAllTimezoneTypesUsedForStatistics,
        formatDateTimeToLongDateTime,
        formatDateRange
    } = useI18n();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const explorersStore = useExplorersStore();

    const timezoneTypeIconMap = {
        [TimezoneTypeForStatistics.ApplicationTimezone.type]: mdiHomeClockOutline,
        [TimezoneTypeForStatistics.TransactionTimezone.type]: mdiInvoiceTextClockOutline
    };

    const renameDialog = useTemplateRef<RenameDialogType>('renameDialog');

    const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const dataExportDialog = useTemplateRef<DataExportDialogType>('dataExportDialog');

    const queryImportDialog = useTemplateRef<JsonImportDialogType>('queryImportDialog');

    const queryExportDialog = useTemplateRef<JsonExportDialogType>('queryExportDialog');

    const explorerDataTableTab = useTemplateRef<ExplorerDataTableTabType>('explorerDataTableTab');

    const explorerChartTab = useTemplateRef<ExplorerChartTabType>('explorerChartTab');

    const explorerChangeDisplayOrderDialog = useTemplateRef<ExplorerChangeDisplayOrderDialogType>('explorerChangeDisplayOrderDialog');

    const editDialog = useTemplateRef<EditDialogType>('editDialog');

    const loading = ref<boolean>(true);

    const initing = ref<boolean>(true);

    const updating = ref<boolean>(false);

    const clientSessionId = ref<string>('');

    const initExploration = ref<InsightsExplorer | null>(null);

    const isCurrentDataTableEditable = ref<boolean>(false);

    const activeTab = ref<ExplorerPageTabType>('query');

    const showCustomDateRangeDialog = ref<boolean>(false);

    const firstDayOfWeek = computed<WeekDayValue>(() => userStore.currentUserFirstDayOfWeek);

    const fiscalYearStart = computed<number>(() => userStore.currentUserFiscalYearStart);

    const isCustomChartInAIMode = computed<boolean>(() => explorerChartTab.value?.isInAIMode() ?? false);

    const isCustomChartAIGenerating = computed<boolean>(() => explorerChartTab.value?.isAIGenerating() ?? false);

    const queryExportFileName = computed<string>(() => {
        const nickname = userStore.currentUserNickname;

        if (nickname) {
            return tt('dataExport.insightsExplorerQueryFileName', {
                nickname: nickname
            });
        }

        return tt('dataExport.defaultInsightsExplorerQueryFileName');
    });

    const sampleQueryJson = computed<string>(() => `[
    {
        "id": "", // ${tt('sample.importInsightsExplorerQueries.queryIdDescription')}
        "name": "", // ${tt('sample.importInsightsExplorerQueries.queryNameDescription')}
        "conditions": [
            {
                "condition": { // ${tt('sample.importInsightsExplorerQueries.conditionDescription')}
                    "field": "", // ${tt('sample.importInsightsExplorerQueries.conditionFieldDescription')}
                    "operator": "", // ${tt('sample.importInsightsExplorerQueries.conditionOperatorDescription')}
                    "value": "" // ${tt('sample.importInsightsExplorerQueries.conditionValueDescription')}
                },
                "relation": "" // ${tt('sample.importInsightsExplorerQueries.conditionRelationDescription')}
            }
            // ${tt('sample.importInsightsExplorerQueries.additionalQueryDescription')}
        ]
    }
]`);

    const allExplorations = computed<InsightsExplorerBasicInfo[]>(() => explorersStore.allExplorationBasicInfos);

    const allVisibleExplorations = computed<InsightsExplorerBasicInfo[]>(() => {
        const ret: InsightsExplorerBasicInfo[] = [];
        let hasCurrentExploration = false;

        for (const exploration of explorersStore.allExplorationBasicInfos) {
            if (!exploration.hidden || (exploration.id && exploration.id === currentExploration.value.id)) {
                ret.push(exploration);

                if (exploration.id && exploration.id === currentExploration.value.id) {
                    hasCurrentExploration = true;
                }
            }
        }

        if (!hasCurrentExploration && currentExploration.value && currentExploration.value.id) {
            ret.push(InsightsExplorerBasicInfo.of(currentExploration.value));
        }

        return ret;
    });

    const currentFilter = computed<TransactionExplorerFilter>(() => explorersStore.transactionExplorerFilter);

    const currentExploration = computed<InsightsExplorer>(() => explorersStore.currentExploration);

    const filteredTransactionsInDataTable = computed<TransactionInsightDataItem[]>(() => explorersStore.filteredTransactionsInDataTable);

    const allDateRanges = computed<LocalizedDateRange[]>(() => getAllDateRanges(DateRangeScene.InsightsExplorer, { includeCustom: true }));

    const allTimezoneTypesUsedForDateRange = computed<TypeAndDisplayName[]>(() => getAllTimezoneTypesUsedForStatistics());

    const canShiftDateRange = computed<boolean>(() => currentFilter.value.dateRangeType !== DateRange.All.type);

    const displayQueryDateRangeName = computed<string>(() => formatDateRange(currentFilter.value.dateRangeType, currentFilter.value.startTime, currentFilter.value.endTime));

    const displayQueryStartTime = computed<string>(() => formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(currentFilter.value.startTime)));

    const displayQueryEndTime = computed<string>(() => formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(currentFilter.value.endTime)));

    const isCurrentExplorationModified = computed<boolean>(() => {
        if (!currentExploration.value.id) { // Add
            return !!initExploration.value && !isEquals(currentExploration.value.toCreateRequest(clientSessionId.value), initExploration.value.toCreateRequest(clientSessionId.value));
        } else { // Edit
            return !!initExploration.value && !isEquals(currentExploration.value.toModifyRequest(), initExploration.value.toModifyRequest());
        }
    });

    const allTabs = computed<{ name: string, value: ExplorerPageTabType }[]>(() => {
        return [
            {
                name: tt('Query'),
                value: 'query'
            },
            {
                name: tt('Data Table'),
                value: 'table'
            },
            {
                name: tt('Chart'),
                value: 'chart'
            }
        ];
    });

    function getFilterLinkUrl(): string {
        return `/insights/explorer?${explorersStore.getTransactionExplorerPageParams(currentExploration.value.id, activeTab.value)}`;
    }

    return {
        router, tt, accountsStore, transactionCategoriesStore, transactionTagsStore, explorersStore, timezoneTypeIconMap,
        renameDialog, confirmDialog, snackbar, dataExportDialog, queryImportDialog, queryExportDialog, explorerDataTableTab,
        explorerChartTab, explorerChangeDisplayOrderDialog, editDialog, loading, initing, updating, clientSessionId,
        initExploration, isCurrentDataTableEditable, activeTab, showCustomDateRangeDialog, firstDayOfWeek, fiscalYearStart,
        isCustomChartInAIMode, isCustomChartAIGenerating, queryExportFileName, sampleQueryJson, allExplorations,
        allVisibleExplorations, currentFilter, currentExploration, filteredTransactionsInDataTable, allDateRanges,
        allTimezoneTypesUsedForDateRange, canShiftDateRange, displayQueryDateRangeName, displayQueryStartTime,
        displayQueryEndTime, isCurrentExplorationModified, allTabs, getFilterLinkUrl,
    };
}
