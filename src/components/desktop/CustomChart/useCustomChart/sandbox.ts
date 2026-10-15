import type { SetupProps } from '@/lib/ui/props.ts';
import type { TransactionExplorerQuery } from '@/models/explorer.ts';
import type { TransactionInsightDataItemWithQueryIndexes } from '@/models/transaction.ts';
import type { useCustomChartData } from './data.ts';
import type { useCustomChartState } from './state.ts';
import { type SandboxRequest } from './types.ts';

interface Context {
    tt: ReturnType<typeof useCustomChartState>['tt'];
    settingsStore: ReturnType<typeof useCustomChartState>['settingsStore'];
    sandboxMessageSignature: ReturnType<typeof useCustomChartState>['sandboxMessageSignature'];
    sandboxLoaded: ReturnType<typeof useCustomChartState>['sandboxLoaded'];
    sandbox: ReturnType<typeof useCustomChartState>['sandbox'];
    displayLayout: ReturnType<typeof useCustomChartState>['displayLayout'];
    props: SetupProps<{
    disabled?: boolean;
    displayLayout: number;
    transactions: TransactionInsightDataItemWithQueryIndexes[];
    queries: TransactionExplorerQuery[];
    modelValue: string;
}>;
    executingScript: ReturnType<typeof useCustomChartState>['executingScript'];
    customChartTransactions: ReturnType<typeof useCustomChartData>['customChartTransactions'];
    customScript: ReturnType<typeof useCustomChartState>['customScript'];
}

export function useCustomChartSandbox(context: Context) {
    const { tt, settingsStore, sandboxMessageSignature, sandboxLoaded, sandbox, displayLayout, props, executingScript, customChartTransactions, customScript } = context;

    const sandboxBuildinScripts: string = `
<script>
window.TransactionType = {
    Income: '${tt('Income')}',
    Expense: '${tt('Expense')}',
    Transfer: '${tt('Transfer')}'
};

window.ChartColors = ${JSON.stringify(settingsStore.chartColorList.map(color => '#' + color))};

window.addEventListener('message', function (event) {
    if (!event.data || typeof event.data !== 'string' || event.data[0] !== '{' || event.data.indexOf('${sandboxMessageSignature}') <= 0) {
        return;
    }

    try {
        const data = JSON.parse(event.data);

        if (!data || !data.source || data.source !== '${sandboxMessageSignature}') {
            return;
        }

        const transactions = data.transactions;
        eval(data.code);

        if (window.buildChartOptions) {
            const result = window.buildChartOptions(transactions);
            window.parent.postMessage({ result: JSON.stringify(result) }, '*');
        } else {
            window.parent.postMessage({ knownError: 'No buildChartOptions function defined' }, '*');
        }
    } catch (error) {
        window.parent.postMessage({ error: error.message }, '*');
    }
});
<\/script>
`;

    function reloadSandbox(executeAfterLoaded: boolean): void {
        sandboxLoaded.value = false;

        if (sandbox.value) {
            sandbox.value.src = 'about:blank';
            sandbox.value.srcdoc = sandboxBuildinScripts;

            sandbox.value.onload = () => {
                sandboxLoaded.value = true;

                if (executeAfterLoaded && (displayLayout.value.showChart || displayLayout.value.showChartData)) {
                    executeCustomScript();
                }
            };
        }
    }

    function executeCustomScript(): void {
        if (!sandbox.value || props.disabled || executingScript.value) {
            return;
        }

        executingScript.value = true;

        const sandboxRequest: SandboxRequest = {
            source: sandboxMessageSignature,
            transactions: customChartTransactions.value || [],
            code: customScript.value + `\n\n;if (typeof buildChartOptions !== 'undefined') { window.buildChartOptions = buildChartOptions; }`
        };

        sandbox.value?.contentWindow?.postMessage(JSON.stringify(sandboxRequest), '*');
    }

    return { reloadSandbox, executeCustomScript };
}
