<script setup lang="ts">
import { onBeforeUnmount, shallowRef, useTemplateRef, watch } from 'vue';
import { useAssistantStore } from '@/stores/assistant.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantConversationDTO } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{ context: string }>();
const store = useAssistantStore();
const { tt } = useI18n();
const dialog = useTemplateRef<HTMLDialogElement>('dialog');
const exporting = shallowRef(false);
const format = shallowRef('markdown');
const includeVersions = shallowRef(false);
const error = shallowRef('');
let selection: { id: string; title: string; all: boolean; viewedVersionIds: string[] } | undefined;
let controller: AbortController | undefined;
let generation = 0;

function cancel() {
    generation++;
    controller?.abort();
    controller = undefined;
    selection = undefined;
    exporting.value = false;
    error.value = '';
    dialog.value?.close();
}
function open(item?: AssistantConversationDTO, all = false) {
    cancel();
    const id = item?.id ?? store.conversationId;
    selection = {
        id,
        title: item?.title ?? store.detail?.conversation.title ?? 'assistant',
        all,
        viewedVersionIds:
            id === store.conversationId
                ? Object.values(store.viewedVersions).map((answer) => answer.assistantMessageId)
                : []
    };
    includeVersions.value = false;
    dialog.value?.showModal();
}
async function download() {
    if (!selection || exporting.value) return;
    const target = selection;
    const ticket = generation;
    const selectedFormat = format.value;
    const versions = includeVersions.value;
    controller = new AbortController();
    exporting.value = true;
    error.value = '';
    try {
        const blob = await store.exportData(
            {
                id: target.all ? undefined : target.id,
                all: target.all,
                format: selectedFormat,
                includeVersions: versions,
                viewedVersionIds: !target.all && !versions ? target.viewedVersionIds : undefined
            },
            controller.signal
        );
        // Abort is advisory: a late response must not download after leaving this scope.
        if (ticket !== generation) return;
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `${target.title.replace(/[\\/\u0000-\u001f\u007f]/g, '_').slice(0, 100)}.${selectedFormat === 'json' ? 'json' : 'md'}`;
        link.click();
        setTimeout(() => URL.revokeObjectURL(url), 1000);
        cancel();
    } catch {
        if (ticket === generation) error.value = tt('Assistant request failed');
    } finally {
        if (ticket === generation) exporting.value = false;
    }
}
watch(format, (value) => {
    if (value !== 'json') includeVersions.value = false;
});
watch(() => props.context, cancel, { flush: 'sync' });
watch([() => store.capabilities?.ownerScope, () => store.capabilities?.deletionEpoch], cancel, { flush: 'sync' });
onBeforeUnmount(cancel);
defineExpose({ open });
</script>

<template>
    <dialog :aria-label="tt('Export')" ref="dialog" class="export-dialog" @cancel.prevent="cancel">
        <form @submit.prevent="download">
            <h2>{{ tt('Export') }}</h2>
            <label class="export-field">
                {{ tt('Format') }}
                <select name="export-format" v-model="format" class="export-input" :disabled="exporting">
                    <option value="markdown">Markdown</option>
                    <option value="json">JSON</option>
                </select>
            </label>
            <label v-if="format === 'json'" class="export-field">
                <input v-model="includeVersions" type="checkbox" :disabled="exporting" />
                {{ tt('Include older answer versions') }}
            </label>
            <p v-if="error" role="alert">{{ error }}</p>
            <button class="export-action" type="submit" :disabled="exporting">
                {{ tt(exporting ? 'Exporting' : 'Download') }}
            </button>
            <button class="export-action" type="button" @click="cancel">{{ tt('Cancel') }}</button>
        </form>
    </dialog>
</template>

<style scoped>
.export-dialog {
    max-width: min(90vw, 500px);
    max-height: 85dvh;
    overflow-y: auto;
    overscroll-behavior: contain;
    overflow-wrap: anywhere;
    border-radius: 16px;
    padding: 24px;
    background: var(--assistant-surface, Canvas);
    color: inherit;
}
.export-field {
    display: block;
    margin-block: 12px;
}
.export-input,
.export-input {
    background: var(--assistant-surface, Canvas);
}
.export-action {
    color: inherit;
    font: inherit;
    min-height: 44px;
    padding: 8px;
    border: 1px solid rgb(127 127 127 / 90%);
    border-radius: 8px;
    background: transparent;
}
.export-action {
    cursor: pointer;
}
.export-action:disabled {
    opacity: 0.55;
    cursor: default;
}
.export-action:focus-visible {
    outline: 2px solid currentColor;
    outline-offset: 2px;
}
</style>
