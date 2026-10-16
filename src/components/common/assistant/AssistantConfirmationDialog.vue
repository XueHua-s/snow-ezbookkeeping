<script setup lang="ts">
import { nextTick, useId, onBeforeUnmount, shallowRef, useTemplateRef, watch } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
const props = defineProps<{ context: string; fallbackFocus?: () => void }>();
const { tt } = useI18n();
const dialog = useTemplateRef<HTMLDialogElement>('dialog');
const messageId = useId();
const text = shallowRef('');
let resolvePending: ((value: boolean) => void) | undefined;
let trigger: HTMLElement | null = null;
let generation = 0;

function finish(value: boolean, restoreFocus = true) {
    const ticket = ++generation;
    dialog.value?.close();
    resolvePending?.(value);
    resolvePending = undefined;
    const target = trigger;
    trigger = null;
    if (restoreFocus) {
        void nextTick(() => {
            if (ticket !== generation) return;
            if (target?.isConnected && target.getClientRects().length) target.focus();
            else props.fallbackFocus?.();
        });
    }
}

/** Replacing or invalidating a decision always declines the previous request. */
async function confirm(message: string): Promise<boolean> {
    finish(false, false);
    const active = document.activeElement as HTMLElement | null;
    trigger = active?.closest('details')?.querySelector('summary') ?? active;
    text.value = message;
    const ticket = generation;
    const result = new Promise<boolean>((resolve) => {
        resolvePending = resolve;
    });
    await nextTick();
    if (ticket === generation) dialog.value?.showModal();
    return result;
}
watch(
    () => props.context,
    () => finish(false, false),
    { flush: 'sync' }
);
onBeforeUnmount(() => finish(false, false));
defineExpose({ confirm });
</script>

<template>
    <dialog :aria-label="tt('Confirm')" :aria-describedby="messageId" ref="dialog" class="confirmation-dialog" @cancel.prevent="finish(false)">
        <p :id="messageId">{{ text }}</p>
        <button class="dialog-action" type="button" autofocus @click="finish(false)">{{ tt('Cancel') }}</button>
        <button class="dialog-action" type="button" @click="finish(true)">{{ tt('Confirm') }}</button>
    </dialog>
</template>

<style scoped>
.confirmation-dialog {
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
.dialog-action {
    color: inherit;
    font: inherit;
    min-height: 44px;
    padding: 8px;
    border: 1px solid rgb(127 127 127 / 90%);
    border-radius: 8px;
    background: transparent;
    cursor: pointer;
}
.dialog-action:focus-visible {
    outline: 2px solid currentColor;
    outline-offset: 2px;
}
</style>
