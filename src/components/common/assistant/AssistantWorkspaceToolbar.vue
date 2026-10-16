<script setup lang="ts">
import { useTemplateRef } from 'vue';
import { mdiHistory, mdiPlus } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantConversationDTO } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{
    variant: 'desktop' | 'mobile';
    conversation?: AssistantConversationDTO;
}>();
const emit = defineEmits<{ history: []; new: [] }>();
const { tt } = useI18n();
const historyButton = useTemplateRef<HTMLButtonElement>('historyButton');
defineExpose({ focus: () => historyButton.value?.focus() });
</script>

<template>
    <header class="cloud-toolbar" :class="`toolbar-${props.variant}`">
        <button ref="historyButton" class="toolbar-action" type="button" :aria-label="tt('Conversation history')" :title="tt('Conversation history')" @click="emit('history')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiHistory" /></svg>
        </button>
        <strong class="toolbar-title" :title="conversation?.title || tt('New conversation')">{{ conversation?.title || tt('New conversation') }}</strong>
        <button class="toolbar-action toolbar-new" type="button" :aria-label="tt('New conversation')" :title="tt('New conversation')" @click="emit('new')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiPlus" /></svg>
        </button>
    </header>
</template>

<style scoped>
.cloud-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    flex-shrink: 0;
}
.toolbar-title {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: center;
    font-size: 0.95rem;
    font-weight: 650;
}
.toolbar-action {
    color: var(--assistant-muted, #698778);
    font: inherit;
    width: 44px;
    height: 44px;
    border: 0;
    border-radius: 15px;
    background: var(--assistant-soft-surface, #f0f4ee);
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: 0 0 44px;
    transition: background-color 180ms ease, color 180ms ease, transform 180ms ease;
}
.toolbar-new {
    color: var(--assistant-accent, #278c62);
    background: var(--assistant-accent-soft, #e5f4e9);
}
.toolbar-action:focus-visible {
    outline: 2px solid var(--assistant-accent, #278c62);
    outline-offset: 3px;
}
.toolbar-action svg { width: 22px; height: 22px; fill: currentColor; }
@media (hover: hover) {
    .toolbar-action:hover { background: var(--assistant-accent-soft, #e5f4e9); color: var(--assistant-accent, #278c62); }
}
.toolbar-action:active { transform: scale(0.92); }
.toolbar-mobile { padding: 8px 12px; }
@media (prefers-reduced-motion: reduce) {
    .toolbar-action { transition: none; }
    .toolbar-action:active { transform: none; }
}
</style>
