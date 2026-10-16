import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useRootStore } from '@/stores/index.ts';
import { AssistantDataClearIncompleteError } from '@/stores/assistant/types.ts';
import { useUserStore } from '@/stores/user.ts';

import type { DataStatisticsResponse, DisplayDataStatistics } from '@/models/data_management.ts';

import { parseBigDecimal } from '@/lib/numeral.ts';

export function useDataManagementPageBase() {
    const { tt, formatBigDecimalToLocalizedNumerals } = useI18n();

    const userStore = useUserStore();
    const rootStore = useRootStore();

    const dataStatistics = ref<DataStatisticsResponse | null>(null);

    const displayDataStatistics = computed<DisplayDataStatistics | null>(() => {
        if (!dataStatistics.value) {
            return null;
        }

        return {
            totalTransactionCount: formatBigDecimalToLocalizedNumerals(parseBigDecimal(dataStatistics.value.totalTransactionCount)),
            totalAccountCount: formatBigDecimalToLocalizedNumerals(parseBigDecimal(dataStatistics.value.totalAccountCount)),
            totalTransactionCategoryCount: formatBigDecimalToLocalizedNumerals(parseBigDecimal(dataStatistics.value.totalTransactionCategoryCount)),
            totalTransactionTagCount: formatBigDecimalToLocalizedNumerals(parseBigDecimal(dataStatistics.value.totalTransactionTagCount)),
            totalTransactionPictureCount: formatBigDecimalToLocalizedNumerals(parseBigDecimal(dataStatistics.value.totalTransactionPictureCount)),
            totalExplorationCount: formatBigDecimalToLocalizedNumerals(parseBigDecimal(dataStatistics.value.totalExplorationCount)),
            totalTransactionTemplateCount: formatBigDecimalToLocalizedNumerals(parseBigDecimal(dataStatistics.value.totalTransactionTemplateCount)),
            totalScheduledTransactionCount: formatBigDecimalToLocalizedNumerals(parseBigDecimal(dataStatistics.value.totalScheduledTransactionCount)),
            totalCustomIconCount: formatBigDecimalToLocalizedNumerals(parseBigDecimal(dataStatistics.value.totalCustomIconCount))
        };
    });

    async function clearAllUserData(password: string, confirmNewOperation: () => Promise<boolean>) {
        try { return await rootStore.clearAllUserData({ password }); } catch (error) {
            if (!(error instanceof AssistantDataClearIncompleteError)) throw error;
            dataStatistics.value = await userStore.getUserDataStatistics();
            if (!await confirmNewOperation()) throw { processed: true };
            return rootStore.clearAllUserData({ password, newOperation: true });
        }
    }

    function getExportFileName(fileExtension: string): string {
        const nickname = userStore.currentUserNickname;

        if (nickname) {
            return tt('dataExport.exportFilename', {
                nickname: nickname
            }) + '.' + fileExtension;
        }

        return tt('dataExport.defaultExportFilename') + '.' + fileExtension;
    }

    return {
        // states
        dataStatistics,
        // computed states
        displayDataStatistics,
        // functions
        getExportFileName, clearAllUserData
    }
}
