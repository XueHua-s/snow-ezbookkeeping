import { TransactionExplorerCustomChartDisplayLayout } from '@/core/explorer.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { TransactionExplorerQuery } from '@/models/explorer.ts';
import type { TransactionInsightDataItemWithQueryIndexes } from '@/models/transaction.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useExplorersStore } from '@/stores/explorer.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { computed, ref, useTemplateRef } from 'vue';
import { CustomChartAIState, type SnackBarType } from './types.ts';


interface Context {
    props: SetupProps<{
    disabled?: boolean;
    displayLayout: number;
    transactions: TransactionInsightDataItemWithQueryIndexes[];
    queries: TransactionExplorerQuery[];
    modelValue: string;
}>;
}

export function useCustomChartState(context: Context) {
    const { props } = context;

    const {
        tt,
        formatNumberToLocalizedNumeralsWithoutDigitGrouping
    } = useI18n();

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const explorersStore = useExplorersStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const sandboxMessageSignature: string = '#ezBookkeeping-sandbox-message#';

    const sandbox = useTemplateRef<HTMLIFrameElement>('sandbox');

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const snackbarForFailureReason = useTemplateRef<SnackBarType>('snackbarForFailureReason');

    const sandboxLoaded = ref<boolean>(false);

    const customScript = ref<string>('');

    const chartOptions = ref<Record<string, unknown>>({});

    const executingScript = ref<boolean>(false);

    const executionError = ref<string>('');

    const aiModeState = ref<CustomChartAIState>(CustomChartAIState.Idle);

    const userPrompt = ref<string>('');

    const cancelAIGeneratingUuid = ref<string | undefined>(undefined);

    const displayLayout = computed<TransactionExplorerCustomChartDisplayLayout>(() => TransactionExplorerCustomChartDisplayLayout.valueOf(props.displayLayout) ?? TransactionExplorerCustomChartDisplayLayout.Default);

    function isInAIMode(): boolean {
        return aiModeState.value !== CustomChartAIState.Idle;
    }

    function isAIGenerating(): boolean {
        return aiModeState.value === CustomChartAIState.Generating;
    }

    return {
        tt, formatNumberToLocalizedNumeralsWithoutDigitGrouping, settingsStore, userStore, explorersStore, exchangeRatesStore,
        sandboxMessageSignature, sandbox, snackbar, snackbarForFailureReason, sandboxLoaded, customScript, chartOptions,
        executingScript, executionError, aiModeState, userPrompt, cancelAIGeneratingUuid, displayLayout, isInAIMode,
        isAIGenerating,
    };
}
