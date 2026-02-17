<template>
    <f7-page>
        <f7-navbar :title="tt('AI Assistant')" :back-link="tt('Back')"></f7-navbar>

        <f7-block strong inset v-if="!enabled">
            {{ tt('AI assistant is disabled') }}
        </f7-block>

        <template v-else>
            <f7-block class="assistant-top-actions display-flex justify-content-space-between align-items-center">
                <f7-button small fill
                           :disabled="requesting"
                           @click="generateSummaryMessage">
                    {{ tt('Generate AI Summary') }}
                </f7-button>
                <f7-button small
                           :disabled="requesting || !messages.length"
                           @click="clearConversation">
                    {{ tt('Clear Conversation') }}
                </f7-button>
            </f7-block>

            <f7-block ref="messagesPanel" class="assistant-message-list no-margin">
                <div class="assistant-empty-state" v-if="messages.length < 1">
                    <f7-icon f7="chat_bubble_2"></f7-icon>
                    <div class="margin-top-half">{{ tt('Start a conversation or generate a summary to analyze your bills') }}</div>
                </div>

                <div class="assistant-message-item"
                     :class="`assistant-message-${message.role}`"
                     :key="message.id"
                     v-for="message in messages">
                    <div class="assistant-message-bubble">
                        <div class="assistant-message-role">{{ message.role === 'user' ? tt('You') : tt('AI Assistant') }}</div>
                        <div class="assistant-message-content">{{ message.content }}</div>

                        <div class="assistant-message-references margin-top" v-if="message.references && message.references.length">
                            <div class="assistant-reference-title">{{ tt('Referenced Bills') }}</div>
                            <div class="assistant-reference-item margin-top-half"
                                 :key="reference.id + '-' + reference.time + '-' + reference.similarityScore"
                                 v-for="reference in message.references">
                                <div>{{ reference.timeText || '-' }}</div>
                                <div>{{ reference.categoryName || tt('Uncategorized') }}</div>
                                <div>{{ formatAmountToLocalizedNumeralsWithCurrency(reference.sourceAmount, reference.currency || false) }}</div>
                            </div>
                        </div>
                    </div>
                </div>
            </f7-block>

            <f7-block class="assistant-input-block">
                <f7-textarea
                    class="assistant-input"
                    type="textarea"
                    resizable
                    :disabled="requesting"
                    :placeholder="tt('Ask your personal finance question')"
                    v-model:value="messageInput">
                </f7-textarea>
                <div class="display-flex justify-content-space-between align-items-center margin-top-half">
                    <f7-preloader size="18" v-if="requesting"></f7-preloader>
                    <f7-button fill
                               :disabled="!canSendMessage"
                               @click="sendChatMessage">
                        {{ tt('Send') }}
                    </f7-button>
                </div>
            </f7-block>
        </template>
    </f7-page>
</template>

<script setup lang="ts">
import { nextTick, useTemplateRef, watch } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useAssistantPageBase } from '@/views/base/assistant/AssistantPageBase.ts';

const { tt, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();
const { showToast } = useI18nUIComponents();
const {
    enabled,
    messages,
    messageInput,
    requesting,
    canSendMessage,
    clearConversation,
    sendMessage,
    generateSummary
} = useAssistantPageBase();

const messagesPanel = useTemplateRef<HTMLElement>('messagesPanel');

watch(() => messages.value.length, () => {
    nextTick(() => {
        if (!messagesPanel.value) {
            return;
        }

        messagesPanel.value.scrollTop = messagesPanel.value.scrollHeight;
    });
});

function sendChatMessage(): void {
    sendMessage().catch(error => {
        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function generateSummaryMessage(): void {
    generateSummary().catch(error => {
        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}
</script>

<style scoped>
.assistant-top-actions {
    margin-top: var(--f7-block-margin-vertical);
    margin-bottom: 0;
}

.assistant-message-list {
    height: calc(100vh - 290px);
    overflow-y: auto;
}

.assistant-empty-state {
    min-height: 220px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    opacity: 0.66;
    text-align: center;
}

.assistant-message-item {
    display: flex;
    margin-bottom: 12px;
}

.assistant-message-user {
    justify-content: flex-end;
}

.assistant-message-assistant {
    justify-content: flex-start;
}

.assistant-message-bubble {
    max-width: 85%;
    border-radius: 12px;
    padding: 10px 12px;
    border: 1px solid var(--f7-page-master-border-color);
    background: var(--f7-card-bg-color);
}

.assistant-message-user .assistant-message-bubble {
    background: rgba(var(--f7-theme-color-rgb), 0.08);
}

.assistant-message-role {
    opacity: 0.7;
    font-size: var(--f7-label-font-size);
    margin-bottom: 5px;
}

.assistant-message-content {
    white-space: pre-wrap;
    line-height: 1.45;
}

.assistant-message-references {
    border-top: 1px solid var(--f7-page-master-border-color);
    padding-top: 8px;
    font-size: var(--f7-label-font-size);
}

.assistant-reference-item {
    opacity: 0.75;
}

.assistant-input-block {
    margin-top: 0;
    padding-top: 8px;
}

.assistant-input {
    background: var(--f7-card-bg-color);
    border: 1px solid var(--f7-page-master-border-color);
    border-radius: 12px;
    padding-inline: 8px;
}
</style>
