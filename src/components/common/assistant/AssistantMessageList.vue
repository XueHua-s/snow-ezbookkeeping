<script setup lang="ts">
import { useTemplateRef } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import type { AIAssistantConversationMessage } from '@/views/base/assistant/types.ts';
import { useAssistantAutoScroll } from '@/views/base/assistant/useAssistantAutoScroll.ts';
import AssistantMarkdownContent from '@/components/common/AssistantMarkdownContent.vue';
import AssistantMessageEvidence from '@/components/common/AssistantMessageEvidence.vue';

defineProps<{ messages: readonly AIAssistantConversationMessage[]; enabled: boolean; variant: 'desktop' | 'mobile' }>();
const emit = defineEmits<{ openReference: [id: string] }>();
const { tt } = useI18n();
const panel = useTemplateRef<HTMLElement>('panel');
const content = useTemplateRef<HTMLElement>('content');
useAssistantAutoScroll(panel, content);
</script>

<template>
    <div ref="panel" class="assistant-messages-panel" :class="`assistant-messages-${variant}`">
        <div ref="content" class="assistant-message-list">
            <div v-if="!enabled" role="status">{{ tt('AI assistant is disabled') }}</div>
            <div class="assistant-empty-stage" v-else-if="!messages.length">
                <div class="assistant-empty-state"><slot name="empty-icon" /><div>{{ tt('Start a conversation or generate a summary to analyze your bills') }}</div></div>
                <slot name="empty" />
            </div>
            <div class="assistant-message-item" :class="`assistant-message-${message.role}`" :key="message.id" v-for="message in messages">
                <div class="assistant-message-bubble">
                    <div class="assistant-message-role">{{ message.role === 'user' ? tt('You') : tt('AI Assistant') }}</div>
                    <AssistantMarkdownContent class="assistant-message-content" :content="message.content" :thinking="message.thinking" />
                    <AssistantMessageEvidence :message="message" @open-reference="emit('openReference', $event)" />
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.assistant-messages-panel { min-height: 0; overflow-y: auto; }
.assistant-messages-desktop { flex: 1; padding: 20px; background: linear-gradient(180deg, rgba(var(--v-theme-surface), 0.98) 0%, rgba(var(--v-theme-surface), 1) 72%); }
.assistant-messages-mobile { height: clamp(220px, calc(100vh - 338px), 560px); padding-inline: 12px; }
.assistant-message-list { width: 100%; }
.assistant-empty-stage { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 180px; }
.assistant-empty-state { max-width: 660px; text-align: center; opacity: 0.65; margin-bottom: 18px; }
.assistant-message-item { display: flex; margin-bottom: 16px; }
.assistant-message-item:last-child { margin-bottom: 0; }
.assistant-message-user { justify-content: flex-end; }
.assistant-message-assistant { justify-content: flex-start; }
.assistant-message-bubble { max-width: min(82%, 860px); border-radius: 14px; padding: 12px 14px; background: rgba(var(--v-theme-surface), 1); border: 1px solid rgba(var(--v-theme-outline), 0.18); box-shadow: 0 2px 10px rgba(15, 23, 42, 0.03); }
.assistant-message-user .assistant-message-bubble { background: rgba(var(--v-theme-primary), 0.08); border-color: rgba(var(--v-theme-primary), 0.25); }
.assistant-message-role { font-size: 12px; opacity: 0.7; margin-bottom: 6px; }
.assistant-message-content { line-height: 1.55; }
.assistant-messages-mobile .assistant-message-item { margin-bottom: 12px; }
.assistant-messages-mobile .assistant-message-bubble { max-width: 85%; border-radius: 12px; padding: 10px 12px; border: 1px solid var(--f7-page-master-border-color); background: var(--f7-card-bg-color); }
.assistant-messages-mobile .assistant-message-user .assistant-message-bubble { background: rgba(var(--f7-theme-color-rgb), 0.08); }
@media (max-width: 960px) { .assistant-messages-desktop .assistant-message-bubble { max-width: 100%; } }
</style>
