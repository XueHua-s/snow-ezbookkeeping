<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';
import type { TransactionExplorerQuery } from '@/models/explorer.ts';
import type { TransactionInsightDataItemWithQueryIndexes } from '@/models/transaction.ts';
import {
    mdiCheck,
    mdiClose,
    mdiPlay,
    mdiStop
} from '@mdi/js';
import { CustomChartAIState, useCustomChart } from './CustomChart/useCustomChart.ts';

const props = defineProps<{
    disabled?: boolean;
    displayLayout: number;
    transactions: TransactionInsightDataItemWithQueryIndexes[];
    queries: TransactionExplorerQuery[];
    modelValue: string;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void;
}>();

const {
    tt, sandboxLoaded, customScript, chartOptions, executingScript, executionError, aiModeState, userPrompt, displayLayout,
    customChartEditorExtraLibs, displayChartData, isInAIMode, isAIGenerating, executeCustomScript, switchToAIChartMode,
    generateAICode, cancelAIGeneration, exitAIMode,
} = useCustomChart(props, emit);

defineExpose({
    isInAIMode,
    isAIGenerating,
    switchToAIChartMode
});

</script>

<template>
    <v-row :class="{ 'readonly': disabled }">
        <v-col cols="12" :md="displayLayout.showChart || displayLayout.showChartData ? 6 : 12" v-if="displayLayout.showCode">
            <div class="title-and-toolbar d-flex w-100 mb-1">
                <v-btn density="compact" variant="tonal" :prepend-icon="mdiPlay"
                       :disabled="disabled || !sandboxLoaded || executingScript" :loading="executingScript"
                       @click="executeCustomScript"
                       v-if="aiModeState === CustomChartAIState.Idle">
                    <template #loader>
                        <v-progress-circular indeterminate size="15" width="2" class="loading-icon-before-text" />
                        <span>{{ tt('Execute Custom Script') }}</span>
                    </template>
                    <span>{{ tt('Execute Custom Script') }}</span>
                </v-btn>
                <v-btn density="compact" variant="tonal" :prepend-icon="mdiCheck"
                       :disabled="disabled || aiModeState === CustomChartAIState.Generating || !userPrompt || !userPrompt.trim()"
                       :loading="aiModeState === CustomChartAIState.Generating"
                       @click="generateAICode"
                       v-if="aiModeState !== CustomChartAIState.Idle">
                    <template #loader>
                        <v-progress-circular indeterminate size="15" width="2" class="loading-icon-before-text" />
                        <span>{{ tt('Generate') }}</span>
                    </template>
                    {{ tt('Generate') }}
                </v-btn>
                <v-btn class="ms-2" density="compact" color="secondary" variant="tonal"
                       :prepend-icon="aiModeState === CustomChartAIState.Generating ? mdiStop : mdiClose"
                       @click="aiModeState === CustomChartAIState.Generating ? cancelAIGeneration() : exitAIMode()"
                       v-if="aiModeState !== CustomChartAIState.Idle">
                    {{ aiModeState === CustomChartAIState.Generating ? tt('Cancel Generation') : tt('Cancel') }}
                </v-btn>
            </div>
            <code-editor class="w-100" style="height: 620px" language="javascript"
                         :readonly="disabled" :rounded="true" :line-numbers="true"
                         :extra-libs="customChartEditorExtraLibs" v-model="customScript"
                         v-if="aiModeState === CustomChartAIState.Idle" />
            <v-textarea no-resize persistent-placeholder
                        class="w-100 always-cursor-text" style="height: 620px"
                        autocomplete="off" density="compact" variant="outlined"
                        :disabled="disabled || aiModeState === CustomChartAIState.Generating"
                        :placeholder="tt('Describe the chart you want')"
                        v-model="userPrompt"
                        v-if="aiModeState !== CustomChartAIState.Idle"></v-textarea>
        </v-col>
        <v-col cols="12" :md="displayLayout.showCode ? 6 : 12"
               :class="{ 'readonly': disabled || aiModeState === CustomChartAIState.Generating }"
               v-if="displayLayout.showChart">
            <div class="w-100 custom-chart-container">
                <div class="w-100 h-100 d-flex align-center justify-content-center" v-if="executionError">
                    <v-alert class="mb-3" density="compact" type="error" variant="tonal" :text="executionError" />
                </div>
                <v-chart autoresize :option="chartOptions" :update-options="{ notMerge: true }"
                         v-if="!executionError && chartOptions" />
            </div>
        </v-col>
        <v-col cols="12" :md="displayLayout.showCode ? 6 : 12"
               :class="{ 'readonly': disabled || aiModeState === CustomChartAIState.Generating }"
               v-else-if="displayLayout.showChartData">
            <div class="title-and-toolbar d-flex align-center w-100 mb-1">
                <span class="text-body-large">{{ tt('Chart Data') }}</span>
            </div>
            <div class="w-100">
                <code-editor class="w-100" style="height: 620px" language="json"
                             :readonly="true" :rounded="true" :model-value="displayChartData" />
            </div>
        </v-col>
    </v-row>

    <iframe id="sandbox" ref="sandbox" sandbox="allow-scripts" style="display: none;"></iframe>
    <snack-bar ref="snackbar" />
    <snack-bar ref="snackbarForFailureReason" not-auto-hide color="warning" variant="outlined" location="top center" />
</template>



<style scoped>
.custom-chart-container {
    width: 100%;
    height: 660px;
}
</style>
