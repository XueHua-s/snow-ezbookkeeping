<script setup lang="ts">
import { onBeforeUnmount, shallowRef } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
const props = defineProps<{ text: string; streaming?: boolean; code?: boolean }>();
const { tt } = useI18n();
const copied = shallowRef(false);
const fallback = shallowRef<string | null>(null);
let timer: ReturnType<typeof setTimeout> | undefined;
async function copy(): Promise<void> {
    const snapshot = props.text;
    try {
        if (!navigator.clipboard?.writeText) throw new Error('clipboard unavailable');
        await navigator.clipboard.writeText(snapshot);
        copied.value = true;
        fallback.value = null;
        clearTimeout(timer);
        timer = setTimeout(() => { copied.value = false; }, 2000);
    } catch {
        copied.value = false;
        fallback.value = snapshot;
    }
}
onBeforeUnmount(() => clearTimeout(timer));
</script>
<template>
    <span class="assistant-copy">
        <button type="button" :disabled="!text" :title="tt(code ? 'Copy code' : 'Copy Markdown')" @click="copy">
            {{ tt(copied ? (streaming ? 'Copied current content' : 'Copied') : (code ? 'Copy code' : 'Copy Markdown')) }}
        </button>
        <span v-if="fallback !== null" class="assistant-copy-fallback" role="status">
            {{ tt('Copy failed. Select the text below or retry.') }}
            <textarea readonly :value="fallback" :aria-label="tt('Text to copy')" @focus="($event.target as HTMLTextAreaElement).select()" />
            <button type="button" @click="copy">{{ tt('Retry') }}</button>
            <button type="button" @click="fallback = null">{{ tt('Close') }}</button>
        </span>
    </span>
</template>
<style scoped>
.assistant-copy { display: inline-flex; flex-wrap: wrap; }
button { width: auto; min-height: 44px; padding: 6px 12px; border: 1px solid transparent; border-radius: 8px; background: transparent; color: inherit; font: inherit; cursor: pointer; }
button:hover { background: rgb(127 127 127 / 12%); }
button:focus-visible, textarea:focus-visible { outline: 2px solid currentColor; outline-offset: 2px; }
button:disabled { opacity: .55; cursor: default; }
.assistant-copy-fallback { display: block; width: 100%; padding: 12px; border: 1px solid currentColor; }
textarea { display: block; width: 100%; min-height: 100px; color: inherit; background: transparent; }
</style>
