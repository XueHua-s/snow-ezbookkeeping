import type { SetupProps } from '@/lib/ui/props.ts';
import type { TransactionExplorerQuery } from '@/models/explorer.ts';
import type { TransactionInsightDataItemWithQueryIndexes } from '@/models/transaction.ts';
import { useCustomChartAi } from './useCustomChart/ai.ts';
import { useCustomChartData } from './useCustomChart/data.ts';
import { useCustomChartLifecycle } from './useCustomChart/lifecycle.ts';
import { useCustomChartSamples } from './useCustomChart/samples.ts';
import { useCustomChartSandbox } from './useCustomChart/sandbox.ts';
import { useCustomChartState } from './useCustomChart/state.ts';
export * from './useCustomChart/types.ts';

// Coordinates CustomChart state and effects; called once from the component setup.
export function useCustomChart(props: SetupProps<{
    disabled?: boolean;
    displayLayout: number;
    transactions: TransactionInsightDataItemWithQueryIndexes[];
    queries: TransactionExplorerQuery[];
    modelValue: string;
}>, emit: {
    (e: 'update:modelValue', value: string): void;
}) {
    const {
        tt, formatNumberToLocalizedNumeralsWithoutDigitGrouping, settingsStore, userStore, explorersStore, exchangeRatesStore,
        sandboxMessageSignature, sandbox, snackbar, snackbarForFailureReason, sandboxLoaded, customScript, chartOptions,
        executingScript, executionError, aiModeState, userPrompt, cancelAIGeneratingUuid, displayLayout, isInAIMode,
        isAIGenerating,
    } = useCustomChartState({ props });
    const { customChartEditorExtraLibs, sampleScript } = useCustomChartSamples({ tt });
    const { displayChartData, customChartTransactions } = useCustomChartData({ executingScript, tt, executionError, chartOptions, props, formatNumberToLocalizedNumeralsWithoutDigitGrouping, userStore, exchangeRatesStore });
    const { reloadSandbox, executeCustomScript } = useCustomChartSandbox({ tt, settingsStore, sandboxMessageSignature, sandboxLoaded, sandbox, displayLayout, props, executingScript, customChartTransactions, customScript });
    const { switchToAIChartMode, generateAICode, cancelAIGeneration, exitAIMode } = useCustomChartAi({ props, aiModeState, userPrompt, cancelAIGeneratingUuid, explorersStore, customScript, reloadSandbox, snackbarForFailureReason, snackbar });
    const {  } = useCustomChartLifecycle({ sandbox, executingScript, snackbar, executionError, tt, chartOptions, reloadSandbox, customScript, props, sampleScript, customChartTransactions, emit });

    return {
        tt, sandbox, snackbar, snackbarForFailureReason, sandboxLoaded, customScript, chartOptions, executingScript,
        executionError, aiModeState, userPrompt, displayLayout, customChartEditorExtraLibs, displayChartData, isInAIMode,
        isAIGenerating, executeCustomScript, switchToAIChartMode, generateAICode, cancelAIGeneration, exitAIMode,
    };
}
