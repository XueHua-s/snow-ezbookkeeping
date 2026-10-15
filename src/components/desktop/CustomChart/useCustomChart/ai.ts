import { generateRandomUUID } from '@/lib/misc.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import type { TransactionExplorerQuery } from '@/models/explorer.ts';
import type { TransactionInsightDataItemWithQueryIndexes } from '@/models/transaction.ts';
import type { useCustomChartSandbox } from './sandbox.ts';
import type { useCustomChartState } from './state.ts';
import { CustomChartAIState } from './types.ts';

interface Context {
    props: SetupProps<{
    disabled?: boolean;
    displayLayout: number;
    transactions: TransactionInsightDataItemWithQueryIndexes[];
    queries: TransactionExplorerQuery[];
    modelValue: string;
}>;
    aiModeState: ReturnType<typeof useCustomChartState>['aiModeState'];
    userPrompt: ReturnType<typeof useCustomChartState>['userPrompt'];
    cancelAIGeneratingUuid: ReturnType<typeof useCustomChartState>['cancelAIGeneratingUuid'];
    explorersStore: ReturnType<typeof useCustomChartState>['explorersStore'];
    customScript: ReturnType<typeof useCustomChartState>['customScript'];
    reloadSandbox: ReturnType<typeof useCustomChartSandbox>['reloadSandbox'];
    snackbarForFailureReason: ReturnType<typeof useCustomChartState>['snackbarForFailureReason'];
    snackbar: ReturnType<typeof useCustomChartState>['snackbar'];
}

export function useCustomChartAi(context: Context) {
    const { props, aiModeState, userPrompt, cancelAIGeneratingUuid, explorersStore, customScript, reloadSandbox, snackbarForFailureReason, snackbar } = context;

    function switchToAIChartMode(): void {
        if (props.disabled || aiModeState.value !== CustomChartAIState.Idle) {
            return;
        }

        userPrompt.value = '';
        aiModeState.value = CustomChartAIState.Prompting;
    }

    function generateAICode(): void {
        if (props.disabled || aiModeState.value !== CustomChartAIState.Prompting || !userPrompt.value) {
            return;
        }

        const prompt = userPrompt.value.trim();

        if (!prompt) {
            return;
        }

        cancelAIGeneratingUuid.value = generateRandomUUID();
        aiModeState.value = CustomChartAIState.Generating;

        explorersStore.generateCustomChartCode({
            userPrompt: prompt,
            currentCode: customScript.value,
            cancelableUuid: cancelAIGeneratingUuid.value
        }).then(response => {
            aiModeState.value = CustomChartAIState.Idle;
            cancelAIGeneratingUuid.value = undefined;

            if (response.code) {
                customScript.value = response.code;
                reloadSandbox(true);
            } else if (response.failureReason) {
                snackbarForFailureReason.value?.showError(response.failureReason);
            }
        }).catch(error => {
            if (error.canceled) {
                return;
            }

            aiModeState.value = CustomChartAIState.Prompting;
            cancelAIGeneratingUuid.value = undefined;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function cancelAIGeneration(): void {
        if (!cancelAIGeneratingUuid.value) {
            return;
        }

        explorersStore.cancelGenerateCustomChartCode(cancelAIGeneratingUuid.value);
        aiModeState.value = CustomChartAIState.Prompting;
        cancelAIGeneratingUuid.value = undefined;

        snackbar.value?.showMessage('User Canceled');
    }

    function exitAIMode(): void {
        if (aiModeState.value == CustomChartAIState.Prompting) {
            aiModeState.value = CustomChartAIState.Idle;
        }
    }

    return { switchToAIChartMode, generateAICode, cancelAIGeneration, exitAIMode };
}
