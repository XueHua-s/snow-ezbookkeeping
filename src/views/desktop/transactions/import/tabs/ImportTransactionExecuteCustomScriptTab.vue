<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';
import {
    mdiPlay
} from '@mdi/js';
import { useImportTransactionExecuteCustomScriptTab } from './ImportTransactionExecuteCustomScriptTab/useImportTransactionExecuteCustomScriptTab.ts';

const props = defineProps<{
    parsedFileData?: string[][];
    disabled?: boolean;
}>();

const {
    tt, formatNumberToLocalizedNumerals, sandboxLoaded, customScript, previewResult, executingScript, previewCount,
    previewCounts, customScriptEditorExtraLibs, displayPreviewResult, menus, executeCustomScript, generateResult, reset,
} = useImportTransactionExecuteCustomScriptTab(props);

defineExpose({
    menus,
    generateResult,
    reset
});

</script>

<template>
    <v-row class="ma-3">
        <v-col cols="12" md="6">
            <div class="title-and-toolbar d-flex w-100 mb-1">
                <v-btn density="compact" variant="tonal" :prepend-icon="mdiPlay"
                       :disabled="disabled || !sandboxLoaded || executingScript" :loading="executingScript"
                       @click="executeCustomScript()">
                    <template #loader>
                        <v-progress-circular indeterminate size="15" width="2" class="loading-icon-before-text" />
                        <span>{{ tt('Execute Custom Script') }}</span>
                    </template>
                    <span>{{ tt('Execute Custom Script') }}</span>
                </v-btn>
            </div>
            <code-editor class="w-100" style="height: 360px" language="javascript"
                         :readonly="disabled" :rounded="true" :line-numbers="true"
                         :extra-libs="customScriptEditorExtraLibs" v-model="customScript" />
        </v-col>
        <v-col cols="12" md="6">
            <div class="title-and-toolbar d-flex w-100 mb-1">
                <v-btn density="compact" color="default" variant="text"
                       :disabled="disabled || !sandboxLoaded || executingScript || !previewResult">
                    <span>{{ tt('format.misc.previewCount', { count: previewCount > 0 ? formatNumberToLocalizedNumerals(previewCount) : tt('All') }) }}</span>
                    <v-menu activator="parent">
                        <v-list>
                            <v-list-item :key="count.value" :title="count.name"
                                         v-for="count in previewCounts"
                                         @click="previewCount = count.value"></v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
            </div>
            <div class="w-100" style="height: 360px">
                <code-editor language="json" :readonly="true" :rounded="true" :model-value="displayPreviewResult" />
            </div>
        </v-col>
    </v-row>

    <iframe id="sandbox" ref="sandbox" sandbox="allow-scripts" style="display: none;"></iframe>
    <snack-bar ref="snackbar" />
</template>
