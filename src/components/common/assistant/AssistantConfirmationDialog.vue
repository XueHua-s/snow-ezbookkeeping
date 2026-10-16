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
.confirmation-dialog[open] { opacity: 1; transform: translateY(0) scale(1); }
.confirmation-dialog::backdrop {
    background: rgb(25 49 35 / 22%);
    backdrop-filter: blur(5px);
    opacity: 0;
    transition: opacity 180ms ease, display 180ms allow-discrete, overlay 180ms allow-discrete;
}
.confirmation-dialog[open]::backdrop { opacity: 1; }
@starting-style {
    .confirmation-dialog[open] { opacity: 0; transform: translateY(8px) scale(0.98); }
    .confirmation-dialog[open]::backdrop { opacity: 0; }
}
.confirmation-dialog p { margin: 0 0 16px; line-height: 1.7; }

.dialog-action {
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
.dialog-action:hover:not(:disabled) { background: color-mix(in srgb, var(--assistant-action, #278c62) 16%, var(--assistant-card, #fff)); }
.dialog-action:active:not(:disabled) { transform: scale(0.96); }
.dialog-action:disabled { opacity: 0.5; cursor: default; }
.dialog-action:focus-visible { outline: 2px solid var(--assistant-action, #278c62); outline-offset: 3px; }
.dialog-action svg { width: 20px; height: 20px; fill: currentColor; flex: none; }

@media (prefers-reduced-motion: reduce) {
    *, *::backdrop { animation: none !important; transition: none !important; }
}
</style>
