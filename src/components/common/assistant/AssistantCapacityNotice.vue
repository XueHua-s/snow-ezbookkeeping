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
    <aside v-if="visible" class="capacity" role="status">
        <strong>{{ tt(reason || 'Approaching conversation capacity') }}</strong>
        <div>
            <button class="assistant-action" v-if="canContinue" type="button" @click="emit('continue')">
                {{ tt('Continue in a new conversation') }}</button
            ><button class="assistant-action" type="button" :aria-label="tt('Conversation history')" :title="tt('Conversation history')" @click="emit('manage')"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiHistory" /></svg></button>
        </div>
    </aside>
</template>
<style scoped src="./assistantAction.css"></style>
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
/* Remove obsolete content immediately; only its arrival is animated. */
.capacity { animation: capacity-enter 180ms ease; }
@keyframes capacity-enter { from { opacity: 0; transform: translateY(4px); } }

@media (prefers-reduced-motion: reduce) {
    .capacity { animation: none; }
}
</style>
