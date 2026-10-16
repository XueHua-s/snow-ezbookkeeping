<script setup lang="ts">
import { mdiClose, mdiRefresh } from '@mdi/js';
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
        class="reference-notice assistant-dialog"
        :aria-label="tt('View Details')"
        @cancel.prevent="dismissReference"
    >
        <p v-if="reference.state.value" :role="reference.state.value === 'checking' ? 'status' : 'alert'">
            {{ tt(referenceLabels[reference.state.value]) }}
        </p>
        <button
            class="assistant-action"
            v-if="reference.state.value === 'changed' || reference.state.value === 'unverified'"
            type="button"
            @click="reference.viewCurrent"
        >
            {{ tt('View current bill') }}
        </button>
        <button
            class="assistant-action"
            v-if="reference.state.value === 'changed' && !store.busy && store.capabilities?.generationAvailable"
            type="button"
            @click="queryReferenceAgain"
         :aria-label="tt('Ask this question again')" :title="tt('Ask this question again')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <button
            class="assistant-action"
            v-if="reference.state.value === 'unavailable' || reference.state.value === 'failed'"
            type="button"
            @click="reference.retry"
         :aria-label="tt('Retry')" :title="tt('Retry')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <button class="assistant-action" type="button" @click="dismissReference" :aria-label="tt('Close')" :title="tt('Close')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiClose" /></svg>
        </button>
    </dialog>
</template>
<style scoped src="./assistantAction.css"></style>
<style scoped src="./assistantDialog.css"></style>
