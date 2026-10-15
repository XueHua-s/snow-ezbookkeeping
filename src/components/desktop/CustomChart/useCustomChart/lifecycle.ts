import { TransactionExplorerCustomChartDisplayLayout } from '@/core/explorer.ts';
import logger from '@/lib/logger.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import type { TransactionExplorerQuery } from '@/models/explorer.ts';
import type { TransactionInsightDataItemWithQueryIndexes } from '@/models/transaction.ts';
import { onMounted, onUnmounted, watch } from 'vue';
import type { useCustomChartData } from './data.ts';
import type { useCustomChartSamples } from './samples.ts';
import type { useCustomChartSandbox } from './sandbox.ts';
import type { useCustomChartState } from './state.ts';
import { type SandboxResponse } from './types.ts';

interface Context {
    sandbox: ReturnType<typeof useCustomChartState>['sandbox'];
    executingScript: ReturnType<typeof useCustomChartState>['executingScript'];
    snackbar: ReturnType<typeof useCustomChartState>['snackbar'];
    executionError: ReturnType<typeof useCustomChartState>['executionError'];
    tt: ReturnType<typeof useCustomChartState>['tt'];
    chartOptions: ReturnType<typeof useCustomChartState>['chartOptions'];
    reloadSandbox: ReturnType<typeof useCustomChartSandbox>['reloadSandbox'];
    customScript: ReturnType<typeof useCustomChartState>['customScript'];
    props: SetupProps<{
    disabled?: boolean;
    displayLayout: number;
    transactions: TransactionInsightDataItemWithQueryIndexes[];
    queries: TransactionExplorerQuery[];
    modelValue: string;
}>;
    sampleScript: ReturnType<typeof useCustomChartSamples>['sampleScript'];
    customChartTransactions: ReturnType<typeof useCustomChartData>['customChartTransactions'];
    emit: {
    (e: 'update:modelValue', value: string): void;
};
}

export function useCustomChartLifecycle(context: Context) {
    const { sandbox, executingScript, snackbar, executionError, tt, chartOptions, reloadSandbox, customScript, props, sampleScript, customChartTransactions, emit } = context;

    function onMessage(event: MessageEvent<SandboxResponse>): void {
        if (event.source !== sandbox.value?.contentWindow) {
            return;
        }

        executingScript.value = false;

        const data = event.data;

        if (data.knownError) {
            snackbar.value?.showError(data.knownError);
            executionError.value = tt(data.knownError);
        } else if (data.error) {
            logger.error('Failed to execute custom script: ' + data.error);
            snackbar.value?.showError('Failed to execute custom script');
            executionError.value = data.error;
        } else if (data.result) {
            try {
                chartOptions.value = JSON.parse(data.result) as Record<string, unknown>;
                executionError.value = '';
            } catch (error) {
                logger.error('Failed to build custom chart options', error);
                executionError.value = tt('Failed to execute custom script');
            }
        }

        reloadSandbox(false);
    }

    onMounted(() => {
        customScript.value = props.modelValue || sampleScript.value;
        reloadSandbox(true);
        window.addEventListener('message', onMessage);
    });

    onUnmounted(() => {
        window.removeEventListener('message', onMessage);
    });

    watch(() => props.displayLayout, (newValue, oldValue) => {
        if ((TransactionExplorerCustomChartDisplayLayout.valueOf(newValue)?.showChart || TransactionExplorerCustomChartDisplayLayout.valueOf(newValue)?.showChartData)
            && !(TransactionExplorerCustomChartDisplayLayout.valueOf(oldValue)?.showChart || TransactionExplorerCustomChartDisplayLayout.valueOf(oldValue)?.showChartData)) {
            reloadSandbox(true);
        }
    });

    watch(customChartTransactions, () => {
        reloadSandbox(true);
    });

    watch(customScript, value => {
        emit('update:modelValue', value);
    });

    watch(() => props.modelValue, value => {
        if (value !== customScript.value) {
            customScript.value = value;
            reloadSandbox(true);
        }
    });

    return {  };
}
