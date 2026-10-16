<script setup lang="ts">
import { computed } from 'vue';
import { mdiHistory } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
const props = defineProps<{
    used: number;
    reserved: number;
    limit: number;
    reason?: string;
    canContinue?: boolean;
}>();
const emit = defineEmits<{ continue: []; manage: [] }>();
const { tt } = useI18n();
const visible = computed(() => !!props.reason || props.used + props.reserved >= props.limit * 0.8);
</script>
<template>
    <Transition name="capacity">
    <aside v-if="visible" class="capacity" role="status">
        <strong>{{ tt(reason || 'Approaching conversation capacity') }}</strong>
        <div>
            <button v-if="canContinue" type="button" @click="emit('continue')">
                {{ tt('Continue in a new conversation') }}</button
            ><button type="button" :aria-label="tt('Conversation history')" :title="tt('Conversation history')" @click="emit('manage')"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiHistory" /></svg></button>
        </div>
    </aside>
    </Transition>
</template>
<style scoped>
.capacity {
    padding: 16px;
    border: 1px solid var(--assistant-border, #e5ece4);
    background: var(--assistant-card, #fff);
    color: var(--assistant-text, #233e30);
    border-radius: 18px;
    margin: 12px 16px;
    line-height: 1.6;
}
.capacity p { color: var(--assistant-muted, #718277); font-size: 0.8125rem; margin: 6px 0; }
.capacity-enter-active, .capacity-leave-active { transition: opacity 180ms ease, transform 180ms ease; }
.capacity-enter-from, .capacity-leave-to { opacity: 0; transform: translateY(4px); }

button {
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
button:hover:not(:disabled) { background: color-mix(in srgb, var(--assistant-action, #278c62) 16%, var(--assistant-card, #fff)); }
button:active:not(:disabled) { transform: scale(0.96); }
button:disabled { opacity: 0.5; cursor: default; }
button:focus-visible { outline: 2px solid var(--assistant-action, #278c62); outline-offset: 3px; }
button svg { width: 20px; height: 20px; fill: currentColor; flex: none; }

@media (prefers-reduced-motion: reduce) {
    *, *::backdrop { animation: none !important; transition: none !important; }
}
</style>
