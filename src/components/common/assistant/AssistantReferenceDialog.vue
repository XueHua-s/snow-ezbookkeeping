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
        class="reference-notice"
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
         :aria-label="tt('Ask this question again')" :title="tt('Ask this question again')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <button
            class="reference-action"
            v-if="reference.state.value === 'unavailable' || reference.state.value === 'failed'"
            type="button"
            @click="reference.retry"
         :aria-label="tt('Retry')" :title="tt('Retry')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <button class="reference-action" type="button" @click="dismissReference" :aria-label="tt('Close')" :title="tt('Close')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiClose" /></svg>
        </button>
    </dialog>
</template>
<style scoped>
.reference-notice {
    width: min(90vw, 440px);
    max-height: 85dvh;
    overflow-y: auto;
    overscroll-behavior: contain;
    overflow-wrap: anywhere;
    border: 1px solid var(--assistant-border, #e5ece4);
    border-radius: 24px;
    padding: 24px;
    background: var(--assistant-surface, #fbfcf9);
    color: var(--assistant-text, #233e30);
    box-shadow: 0 20px 70px rgb(35 62 48 / 15%);
    opacity: 0;
    transform: translateY(8px) scale(0.98);
    transition: opacity 180ms ease, transform 180ms ease, display 180ms allow-discrete, overlay 180ms allow-discrete;
}
.reference-notice[open] { opacity: 1; transform: translateY(0) scale(1); }
.reference-notice::backdrop {
    background: rgb(25 49 35 / 22%);
    backdrop-filter: blur(5px);
    opacity: 0;
    transition: opacity 180ms ease, display 180ms allow-discrete, overlay 180ms allow-discrete;
}
.reference-notice[open]::backdrop { opacity: 1; }
@starting-style {
    .reference-notice[open] { opacity: 0; transform: translateY(8px) scale(0.98); }
    .reference-notice[open]::backdrop { opacity: 0; }
}
.reference-notice p { margin: 0 0 16px; line-height: 1.7; }

.reference-action {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-width: 44px;
    min-height: 44px;
    margin: 4px 6px 0 0;
    padding: 8px 14px;
    border: 0;
    border-radius: 14px;
    color: var(--assistant-action, #278c62);
    background: var(--assistant-tint, #e8f3e9);
    font: inherit;
    cursor: pointer;
    transition: background 160ms ease, transform 160ms ease, opacity 160ms ease;
}
.reference-action:hover:not(:disabled) { background: color-mix(in srgb, var(--assistant-action, #278c62) 16%, var(--assistant-card, #fff)); }
.reference-action:active:not(:disabled) { transform: scale(0.96); }
.reference-action:disabled { opacity: 0.5; cursor: default; }
.reference-action:focus-visible { outline: 2px solid var(--assistant-action, #278c62); outline-offset: 3px; }
.reference-action svg { width: 20px; height: 20px; fill: currentColor; flex: none; }

@media (prefers-reduced-motion: reduce) {
    *, *::backdrop { animation: none !important; transition: none !important; }
}
</style>
