<script setup lang="ts">
import { computed, onBeforeUnmount, shallowRef } from 'vue';
import { mdiCheck, mdiClose, mdiContentCopy, mdiRefresh } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
const props = defineProps<{ text: string; streaming?: boolean; code?: boolean }>();
const { tt } = useI18n();
const copied = shallowRef(false);
const fallback = shallowRef<string | null>(null);
const label = computed(() => tt(copied.value ? (props.streaming ? 'Copied current content' : 'Copied') : (props.code ? 'Copy code' : 'Copy Markdown')));
let generation = 0;
let timer: ReturnType<typeof setTimeout> | undefined;
async function copy(): Promise<void> {
    const ticket = ++generation;
    clearTimeout(timer);
    copied.value = false;
    const snapshot = props.text;
    try {
        if (!navigator.clipboard?.writeText) throw new Error('clipboard unavailable');
        await navigator.clipboard.writeText(snapshot);
        if (ticket !== generation) return;
        copied.value = true;
        fallback.value = null;
        clearTimeout(timer);
        timer = setTimeout(() => {
            copied.value = false;
        }, 2000);
    } catch {
        if (ticket !== generation) return;
        copied.value = false;
        fallback.value = snapshot;
    }
}
onBeforeUnmount(() => {
    generation++;
    clearTimeout(timer);
});
</script>
<template>
    <span class="assistant-copy">
        <button
            type="button"
            :disabled="!text"
            :title="label"
            :aria-label="label"
            :class="{ copied }"
            @click="copy"
        >
            <Transition name="copy-feedback" mode="out-in">
                <svg :key="String(copied)" viewBox="0 0 24 24" aria-hidden="true"><path :d="copied ? mdiCheck : mdiContentCopy" /></svg>
            </Transition>
        </button>
        <span class="sr-only" role="status">{{ copied ? label : '' }}</span>
        <Transition name="copy-feedback">
        <span v-if="fallback !== null" class="assistant-copy-fallback" role="status">
            {{ tt('Copy failed. Select the text below or retry.') }}
            <textarea
                readonly
                name="assistant-copy-text"
                autocomplete="off"
                :value="fallback"
                :aria-label="tt('Text to copy')"
                @focus="($event.target as HTMLTextAreaElement).select()"
            />
            <button type="button" :aria-label="tt('Retry')" :title="tt('Retry')" @click="copy"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg></button>
            <button type="button" :aria-label="tt('Close')" :title="tt('Close')" @click="fallback = null"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiClose" /></svg></button>
        </span>
        </Transition>
    </span>
</template>
<style scoped>
.assistant-copy {
    display: inline-flex;
    flex-wrap: wrap;
}
button {
    width: 44px;
    display: inline-grid;
    place-items: center;
    min-height: 44px;
    padding: 10px;
    border: 1px solid transparent;
    border-radius: 14px;
    background: transparent;
    color: var(--assistant-muted, #6b8074);
    font: inherit;
    cursor: pointer;
    transition: background 180ms ease, color 180ms ease, transform 180ms ease;
}
button:hover {
    background: var(--assistant-tint, #e8f3e9);
    color: var(--assistant-action, #278c62);
}
button:focus-visible,
textarea:focus-visible {
    outline: 2px solid var(--assistant-action, #278c62);
    outline-offset: 2px;
}
button:disabled {
    opacity: 0.55;
    cursor: default;
}
.assistant-copy-fallback {
    display: block;
    width: 100%;
    padding: 12px;
    border: 1px solid var(--assistant-border, #e5ece3);
    border-radius: 16px;
    background: var(--assistant-card, #fff);
}
textarea {
    display: block;
    width: 100%;
    min-height: 100px;
    color: inherit;
    background: transparent;
    border: 1px solid var(--assistant-border, #e5ece3);
    border-radius: 10px;
    padding: 10px;
}
svg { width: 20px; height: 20px; fill: currentColor; }
button:active:not(:disabled) { transform: scale(.94); }
.copied { color: var(--assistant-action, #278c62); }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
.copy-feedback-enter-active, .copy-feedback-leave-active { transition: opacity 120ms ease, transform 120ms ease; }
.copy-feedback-enter-from, .copy-feedback-leave-to { opacity: 0; transform: translateY(3px); }
@media (prefers-reduced-motion: reduce) { button, .copy-feedback-enter-active, .copy-feedback-leave-active { transition: none; } }
</style>
