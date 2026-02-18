<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="d-flex align-center">
                        <v-icon :icon="mdiRobotOutline" size="24" />
                        <span class="ms-2">{{ tt('AI Assistant') }}</span>
                        <v-chip class="ms-3" size="small" color="secondary" variant="tonal" v-if="aiAssistantModelID">{{ aiAssistantModelID }}</v-chip>
                        <v-spacer />
                        <v-btn class="ms-2"
                               color="secondary"
                               variant="tonal"
                               :disabled="!enabled || requesting"
                               :loading="requesting"
                               @click="generateSummaryMessage">
                            <v-icon :icon="mdiTextBoxCheckOutline" size="20" class="me-2" />
                            {{ tt('Generate AI Summary') }}
                        </v-btn>
                        <v-btn class="ms-2"
                               color="default"
                               variant="text"
                               :disabled="requesting || !messages.length"
                               @click="clearConversation">
                            <v-icon :icon="mdiDeleteOutline" size="20" class="me-2" />
                            {{ tt('Clear Conversation') }}
                        </v-btn>
                    </div>
                </template>
                <v-card-subtitle>{{ tt('Private assistant for personal bills and bookkeeping suggestions') }}</v-card-subtitle>

                <v-divider />

                <v-card-text ref="messagesPanel" class="assistant-messages-panel">
                    <v-alert class="mb-4" type="warning" density="compact" variant="tonal" v-if="!enabled">
                        {{ tt('AI assistant is disabled') }}
                    </v-alert>

                    <div class="assistant-empty-state" v-if="enabled && messages.length < 1">
                        <v-icon :icon="mdiMessageTextOutline" size="28" />
                        <div class="mt-2">{{ tt('Start a conversation or generate a summary to analyze your bills') }}</div>
                    </div>

                    <div class="assistant-message-item"
                         :class="`assistant-message-${message.role}`"
                         :key="message.id"
                         v-for="message in messages">
                        <div class="assistant-message-bubble">
                            <div class="assistant-message-role">{{ message.role === 'user' ? tt('You') : tt('AI Assistant') }}</div>
                            <div class="assistant-message-content">{{ message.content }}</div>

                            <div class="assistant-message-references mt-3"
                                 v-if="message.references && message.references.length">
                                <div class="text-subtitle-2">{{ tt('Referenced Bills') }}</div>
                                <div class="assistant-reference-item mt-1"
                                     :key="reference.id + '-' + reference.time + '-' + reference.similarityScore"
                                     v-for="reference in message.references">
                                    <div class="assistant-reference-main">
                                        <span>{{ reference.timeText || '-' }}</span>
                                        <span>{{ reference.categoryName || tt('Uncategorized') }}</span>
                                        <span>{{ reference.sourceAccountName || tt('Account') }}</span>
                                    </div>
                                    <div class="assistant-reference-sub">
                                        <span>{{ formatAmountToLocalizedNumeralsWithCurrency(reference.sourceAmount, reference.currency || false) }}</span>
                                        <span v-if="reference.similarityScore">· {{ tt('Similarity') }} {{ reference.similarityScore }}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </v-card-text>

                <v-divider />

                <v-card-actions class="assistant-actions">
                    <v-textarea
                        auto-grow
                        rows="2"
                        max-rows="6"
                        class="assistant-input me-3"
                        variant="outlined"
                        :disabled="!enabled || requesting"
                        :placeholder="tt('Ask your personal finance question')"
                        v-model="messageInput"
                        @keydown.ctrl.enter.prevent="sendChatMessage"
                    />
                    <v-btn color="primary"
                           :disabled="!enabled || !canSendMessage"
                           :loading="requesting"
                           @click="sendChatMessage">
                        <v-icon :icon="mdiSend" size="20" class="me-2" />
                        {{ tt('Send') }}
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-col>
    </v-row>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';

import { nextTick, useTemplateRef, watch } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { getAIAssistantModelID } from '@/lib/server_settings.ts';
import { useAssistantPageBase } from '@/views/base/assistant/AssistantPageBase.ts';

import {
    mdiRobotOutline,
    mdiTextBoxCheckOutline,
    mdiDeleteOutline,
    mdiMessageTextOutline,
    mdiSend
} from '@mdi/js';

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();
const aiAssistantModelID = getAIAssistantModelID();
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

const snackbar = useTemplateRef<SnackBarType>('snackbar');
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
            snackbar.value?.showError(error);
        }
    });
}

function generateSummaryMessage(): void {
    generateSummary().catch(error => {
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}
</script>

<style scoped>
.assistant-messages-panel {
    min-height: clamp(260px, calc(100vh - 460px), 540px);
    max-height: clamp(260px, calc(100vh - 460px), 540px);
    overflow-y: auto;
}

.assistant-empty-state {
    height: 100%;
    min-height: 480px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    opacity: 0.65;
}

.assistant-message-item {
    display: flex;
    margin-bottom: 14px;
}

.assistant-message-item:last-child {
    margin-bottom: 0;
}

.assistant-message-user {
    justify-content: flex-end;
}

.assistant-message-assistant {
    justify-content: flex-start;
}

.assistant-message-bubble {
    max-width: min(84%, 860px);
    border-radius: 12px;
    padding: 12px 14px;
    background: rgba(var(--v-theme-surface), 1);
    border: 1px solid rgba(var(--v-theme-outline), 0.18);
}

.assistant-message-user .assistant-message-bubble {
    background: rgba(var(--v-theme-primary), 0.08);
    border-color: rgba(var(--v-theme-primary), 0.25);
}

.assistant-message-role {
    font-size: 12px;
    opacity: 0.7;
    margin-bottom: 6px;
}

.assistant-message-content {
    white-space: pre-wrap;
    line-height: 1.55;
}

.assistant-message-references {
    border-top: 1px solid rgba(var(--v-theme-outline), 0.12);
    padding-top: 10px;
}

.assistant-reference-item {
    border: 1px solid rgba(var(--v-theme-outline), 0.12);
    border-radius: 8px;
    padding: 8px 10px;
}

.assistant-reference-main,
.assistant-reference-sub {
    display: flex;
    column-gap: 8px;
    flex-wrap: wrap;
    font-size: 13px;
}

.assistant-reference-sub {
    opacity: 0.72;
    margin-top: 3px;
}

.assistant-actions {
    align-items: flex-end;
    padding: 16px;
}

.assistant-input {
    width: 100%;
}
</style>
