<script setup lang="ts">
import { nextTick, onBeforeUnmount, useTemplateRef, watch } from 'vue';
import { useAssistantStore } from '@/stores/assistant.ts';
import { useI18n } from '@/locales/helpers.ts';
import { useAssistantReference } from '@/views/base/assistant/useAssistantReference.ts';
import type { AssistantMessageDTO } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{ context: string; messages: readonly AssistantMessageDTO[] }>();
const emit = defineEmits<{ openReference: [id: string]; retryQuestion: [message: AssistantMessageDTO] }>();
const store = useAssistantStore();
const { tt } = useI18n();
let referenceTrigger: HTMLElement | null = null;
const reference = useAssistantReference(
    (id) => store.getReferenceTransaction(id),
    (id) => {
        referenceNotice.value?.close();
        referenceTrigger = null;
        emit('openReference', id);
    }
);
function checkReference(...args: Parameters<typeof reference.check>) {
    referenceTrigger = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    return reference.check(...args);
}
function dismissReference() {
    const trigger = referenceTrigger;
    referenceTrigger = null;
    reference.dismiss();
    referenceNotice.value?.close();
    if (trigger?.isConnected) trigger.focus({ preventScroll: true });
}
const referenceLabels = {
    checking: 'Checking current bill',
    changed: 'This bill changed after the original query. The saved answer has not been updated.',
    unverified: 'The current bill amount cannot be compared exactly. Review the current details.',
    unavailable: 'This bill is no longer accessible. It may have been deleted or access changed.',
    failed: 'Unable to check the current bill. Retry when the connection is available.'
};
async function queryReferenceAgain() {
    const question = props.messages.find((message) => message.id === reference.selected.value?.answer.userMessageId);
    if (question) {
        reference.dismiss();
        emit('retryQuestion', question);
    }
}
watch(
    () => props.context,
    () => {
        referenceTrigger = null;
        reference.dismiss();
        referenceNotice.value?.close();
    },
    { flush: 'sync' }
);
const referenceNotice = useTemplateRef<HTMLDialogElement>('referenceNotice');
watch(reference.state, async () => {
    await nextTick();
    if (reference.state.value && !referenceNotice.value?.open) referenceNotice.value?.showModal();
    else if (!reference.state.value) referenceNotice.value?.close();
});
onBeforeUnmount(() => reference.dismiss());
defineExpose({ check: checkReference });
</script>
<template>
    <dialog
        ref="referenceNotice"
        class="export-dialog reference-notice"
        :aria-label="tt('View Details')"
        @cancel.prevent="dismissReference"
    >
        <p v-if="reference.state.value" :role="reference.state.value === 'checking' ? 'status' : 'alert'">
            {{ tt(referenceLabels[reference.state.value]) }}
        </p>
        <button
            class="reference-action"
            v-if="reference.state.value === 'changed' || reference.state.value === 'unverified'"
            type="button"
            @click="reference.viewCurrent"
        >
            {{ tt('View current bill') }}
        </button>
        <button
            class="reference-action"
            v-if="reference.state.value === 'changed' && !store.busy && store.capabilities?.generationAvailable"
            type="button"
            @click="queryReferenceAgain"
        >
            {{ tt('Ask this question again') }}
        </button>
        <button
            class="reference-action"
            v-if="reference.state.value === 'unavailable' || reference.state.value === 'failed'"
            type="button"
            @click="reference.retry"
        >
            {{ tt('Retry') }}
        </button>
        <button class="reference-action" type="button" @click="dismissReference">{{ tt('Close') }}</button>
    </dialog>
</template>
<style scoped>
.reference-notice {
    max-height: 80dvh;
    overflow-y: auto;
    overscroll-behavior: contain;
    max-width: min(90vw, 500px);
    border-radius: 16px;
    padding: 24px;
    background: var(--assistant-surface, Canvas);
    color: inherit;
}
.reference-action {
    color: inherit;
    font: inherit;
    min-height: 44px;
    padding: 8px;
    border: 1px solid rgb(127 127 127 / 90%);
    border-radius: 8px;
    background: transparent;
    cursor: pointer;
}
.reference-action:focus-visible {
    outline: 2px solid currentColor;
    outline-offset: 2px;
}
</style>
