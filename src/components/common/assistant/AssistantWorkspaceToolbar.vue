<script setup lang="ts">
import { useTemplateRef } from 'vue';
import { mdiHistory, mdiPlus, mdiDotsHorizontal } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantConversationDTO } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{
    variant: 'desktop' | 'mobile';
    conversation?: AssistantConversationDTO;
    memoryEnabled?: boolean;
}>();
const emit = defineEmits<{
    history: [];
    new: [];
    memories: [];
    summary: [];
    export: [];
    exportAll: [];
    clear: [];
    memoryMode: [enabled: boolean];
}>();
const { tt } = useI18n();
const menu = useTemplateRef<HTMLDetailsElement>('menu');
function closeMenu() {
    if (!menu.value?.open) return;
    menu.value.open = false;
    menu.value.querySelector('summary')?.focus();
}
function menuClick(event: MouseEvent) {
    if ((event.target as HTMLElement).closest('button')) closeMenu();
}
function focus() {
    menu.value?.querySelector('summary')?.focus();
}
defineExpose({ focus });
</script>

<template>
    <header class="cloud-toolbar" :class="`toolbar-${props.variant}`">
        <button class="toolbar-action toolbar-primary" type="button" :aria-label="tt('Conversation history')" :title="tt('Conversation history')" @click="emit('history')">
            <svg class="toolbar-icon" viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiHistory" /></svg>
            <span class="toolbar-label">{{ tt('Conversation history') }}</span>
        </button>
        <strong class="toolbar-title" :title="conversation?.title || tt('New conversation')">{{ conversation?.title || tt('New conversation') }}</strong>
        <button class="toolbar-action toolbar-primary" type="button" :aria-label="tt('New conversation')" :title="tt('New conversation')" @click="emit('new')">
            <svg class="toolbar-icon" viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiPlus" /></svg>
            <span class="toolbar-label">{{ tt('New conversation') }}</span>
        </button>
        <details ref="menu" class="toolbar-more" @keydown.esc.prevent="closeMenu">
            <summary class="toolbar-action toolbar-primary" :aria-label="tt('More')" :title="tt('More')">
                <svg class="toolbar-icon" viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiDotsHorizontal" /></svg>
            </summary>
            <div class="toolbar-menu" @click.capture="menuClick">
                <p class="full-title">{{ conversation?.title || tt('New conversation') }}</p>
                <label v-if="conversation" class="conversation-memory-setting">
                    <input
                        class="memory-checkbox"
                        type="checkbox"
                        :checked="conversation.memoryMode !== 'off'"
                        @change="emit('memoryMode', ($event.target as HTMLInputElement).checked)"
                    />
                    {{ tt('Allow memory in this conversation') }}
                </label>
                <p v-if="memoryEnabled === false" class="toolbar-notice">{{ tt('Memory is off globally.') }}</p>
                <button
                    class="toolbar-action menu-action"
                    type="button"
                    @click="emit('memories')"
                >
                    {{ tt('Long-term memory') }}
                </button>
                <button class="toolbar-action menu-action" type="button" @click="emit('summary')">
                    {{ tt('Generate AI Summary') }}
                </button>
                <button v-if="conversation" class="toolbar-action menu-action" type="button" @click="emit('export')">
                    {{ tt('Export') }}
                </button>
                <button class="toolbar-action menu-action" type="button" @click="emit('exportAll')">
                    {{ tt('Export all assistant data') }}
                </button>
                <button class="toolbar-action menu-action" type="button" @click="emit('clear')">
                    {{ tt('Clear all assistant data') }}
                </button>
            </div>
        </details>
    </header>
</template>

<style scoped>
.cloud-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    flex-wrap: nowrap;
    flex-shrink: 0;
    border-bottom: 1px solid rgb(127 127 127 / 30%);
}
.toolbar-title {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.full-title {
    margin: 8px 12px;
    overflow-wrap: anywhere;
    font-weight: 600;
}
.toolbar-more {
    position: relative;
}
.toolbar-menu {
    position: absolute;
    z-index: 3;
    inset-inline-end: 0;
    width: min(280px, calc(100vw - 32px));
    padding: 4px;
    max-height: min(70dvh, 480px);
    overflow-y: auto;
    overscroll-behavior: contain;
    background: var(--assistant-surface, Canvas);
    color: inherit;
    box-shadow: 0 4px 20px #0003;
    border-radius: 10px;
}
.toolbar-action {
    color: inherit;
    font: inherit;
    min-height: 44px;
    min-width: 44px;
    box-sizing: border-box;
    gap: 8px;
    padding: 8px;
    border: 1px solid rgb(127 127 127 / 90%);
    border-radius: 8px;
    background: transparent;
    cursor: pointer;
    width: auto;
    display: inline-flex;
    align-items: center;
    flex: 0 0 auto;
}
.menu-action {
    width: 100%;
    text-align: start;
}
.toolbar-action:focus-visible {
    outline: 2px solid currentColor;
    outline-offset: 2px;
}
.conversation-memory-setting {
    padding: 12px;
    display: inline-flex;
    align-items: center;
    gap: 8px;
}
.memory-checkbox {
    width: 18px;
    height: 18px;
    flex: 0 0 18px;
}
.toolbar-notice {
    padding: 8px 16px;
    margin: 0;
    font-size: 0.875rem;
}
.toolbar-primary {
    border-color: transparent;
    justify-content: center;
}
.toolbar-icon {
    width: 22px;
    height: 22px;
    flex: 0 0 22px;
    fill: currentColor;
}
.toolbar-action:hover {
    background: rgb(127 127 127 / 12%);
}
.toolbar-action:active {
    background: rgb(127 127 127 / 22%);
}
.toolbar-mobile {
    gap: 4px;
    padding: 4px 8px;
}
.toolbar-mobile .toolbar-label {
    display: none;
}
@media (max-width: 960px) {
    .toolbar-label {
        display: none;
    }
}
</style>
