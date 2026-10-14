<script setup lang="ts">
import { onBeforeRouteLeave } from 'vue-router';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';
import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';
import SnackBar from '@/components/desktop/SnackBar.vue';
import AssistantMarkdownContent from '@/components/common/AssistantMarkdownContent.vue';
import AssistantMessageEvidence from '@/components/common/AssistantMessageEvidence.vue';

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
type ScrollablePanelRef = HTMLElement | { $el?: Element };

const { tt } = useI18n();
const aiAssistantModelID = getAIAssistantModelID();
const {
    enabled,
    messages,
    messageInput,
    requesting,
    rendering,
    canSendMessage,
    clearConversation,
    disposeConversation,
    cancelCurrentRequest,
    canRequestNextPage,
    requestNextPage,
    sendMessage,
    generateSummary
} = useAssistantPageBase();

const referenceDialog = useTemplateRef<InstanceType<typeof EditDialog>>('referenceDialog');
onBeforeRouteLeave(disposeConversation);

const snackbar = useTemplateRef<SnackBarType>('snackbar');
const messagesPanel = useTemplateRef<ScrollablePanelRef>('messagesPanel');

function getMessagesPanelElement(): HTMLElement | null {
    const panel = messagesPanel.value;

    if (panel instanceof HTMLElement) {
        return panel;
    }

    if (panel && '$el' in panel && panel.$el instanceof HTMLElement) {
        return panel.$el;
    }

    return null;
}

watch(() => messages.value.map(message => `${message.id}:${message.thinking?.length || 0}:${message.content.length}:${message.references?.length || 0}:${message.state}:${message.progress}:${message.queries?.length || 0}`).join('|'), () => {
    nextTick(() => {
        const panel = getMessagesPanelElement();

        if (!panel) {
            return;
        }

        panel.scrollTop = panel.scrollHeight;
    });
});

function openReference(id: string): void {
    referenceDialog.value?.open({ id }).catch(error => {
        if (error && !error.processed) snackbar.value?.showError(error);
    });
}

function sendNextPage(): void {
    requestNextPage(tt('Show the next page of the current query results')).catch(error => {
        if (!error.processed) snackbar.value?.showError(error);
    });
}

function sendChatMessage(): void {
    sendMessage().catch(error => {
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function generateSummaryMessage(): void {
    generateSummary(tt('Generate AI Summary')).catch(error => {
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}
</script>

<template>
    <main-page-layout no-navbar>
        <template #content>
    <v-row class="assistant-page-row">
        <v-col cols="12" class="assistant-page-col">
            <v-card class="assistant-card">
                <template #title>
                    <div class="assistant-title-toolbar">
                        <div class="assistant-title-main">
                            <v-icon :icon="mdiRobotOutline" size="24" />
                            <span class="ms-2">{{ tt('AI Assistant') }}</span>
                            <v-chip class="ms-3" size="small" color="secondary" variant="tonal" v-if="aiAssistantModelID">{{ aiAssistantModelID }}</v-chip>
                        </div>

                        <div class="assistant-title-actions">
                            <v-btn color="secondary"
                                   variant="tonal"
                                   :disabled="!enabled || requesting || rendering"
                                   :loading="requesting || rendering"
                                   @click="generateSummaryMessage">
                                <v-icon :icon="mdiTextBoxCheckOutline" size="20" class="me-2" />
                                {{ tt('Generate AI Summary') }}
                            </v-btn>
                            <v-btn color="default"
                                   variant="text"
                                   :disabled="!messages.length"
                                   @click="clearConversation">
                                <v-icon :icon="mdiDeleteOutline" size="20" class="me-2" />
                                {{ tt('Clear Conversation') }}
                            </v-btn>
                        </div>
                    </div>
                </template>
                <v-card-subtitle>{{ tt('Private assistant for personal bills and bookkeeping suggestions') }}</v-card-subtitle>

                <v-divider />

                <div class="assistant-conversation-shell">
                    <v-card-text ref="messagesPanel" class="assistant-messages-panel">
                        <v-alert class="mb-4" type="warning" density="compact" variant="tonal" v-if="!enabled">
                            {{ tt('AI assistant is disabled') }}
                        </v-alert>

                        <div class="assistant-empty-stage" v-if="enabled && messages.length < 1">
                            <div class="assistant-empty-state">
                                <v-icon :icon="mdiMessageTextOutline" size="32" />
                                <div class="mt-3">{{ tt('Start a conversation or generate a summary to analyze your bills') }}</div>
                            </div>

                            <div class="assistant-composer assistant-composer-inline">
                                <v-textarea
                                    auto-grow
                                    rows="2"
                                    max-rows="6"
                                    class="assistant-input"
                                    variant="outlined"
                                    hide-details
                                    :disabled="!enabled || requesting || rendering"
                                    :placeholder="tt('Ask your personal finance question')"
                                    v-model="messageInput"
                                    @keydown.enter.exact.prevent="sendChatMessage"
                                />
                                <v-btn color="primary"
                                       class="assistant-send-button"
                                       :disabled="!enabled || !canSendMessage"
                                       :loading="requesting"
                                       @click="sendChatMessage">
                                    <v-icon :icon="mdiSend" size="20" class="me-2" />
                                    {{ tt('Send') }}
                                </v-btn>
                            </div>
                        </div>

                        <div class="assistant-message-list" v-else>
                            <div class="assistant-message-item"
                                 :class="`assistant-message-${message.role}`"
                                 :key="message.id"
                                 v-for="message in messages">
                                <div class="assistant-message-bubble">
                                    <div class="assistant-message-role">{{ message.role === 'user' ? tt('You') : tt('AI Assistant') }}</div>
                                    <assistant-markdown-content class="assistant-message-content"
                                                                :content="message.content"
                                                                :thinking="message.thinking" />

                                    <assistant-message-evidence :message="message" @open-reference="openReference" />
                                </div>
                            </div>
                        </div>
                    </v-card-text>

                    <v-divider />

                    <div class="assistant-session-actions" v-if="requesting || canRequestNextPage">
                        <v-btn variant="tonal" color="secondary" v-if="canRequestNextPage" @click="sendNextPage">{{ tt('Next Page') }}</v-btn>
                        <v-btn variant="text" v-if="requesting" @click="cancelCurrentRequest">{{ tt('Cancel') }}</v-btn>
                    </div>

                    <div class="assistant-composer assistant-composer-docked" v-if="enabled && messages.length > 0">
                        <v-textarea
                            auto-grow
                            rows="2"
                            max-rows="6"
                            class="assistant-input"
                            variant="outlined"
                            hide-details
                            :disabled="!enabled || requesting || rendering"
                            :placeholder="tt('Ask your personal finance question')"
                            v-model="messageInput"
                            @keydown.enter.exact.prevent="sendChatMessage"
                        />
                        <v-btn color="primary"
                               class="assistant-send-button"
                               :disabled="!enabled || !canSendMessage"
                               :loading="requesting"
                               @click="sendChatMessage">
                            <v-icon :icon="mdiSend" size="20" class="me-2" />
                            {{ tt('Send') }}
                        </v-btn>
                    </div>
                </div>
            </v-card>
        </v-col>
    </v-row>

        </template>
    </main-page-layout>

    <edit-dialog ref="referenceDialog" :type="TransactionEditPageType.Transaction" />
    <snack-bar ref="snackbar" />
</template>



<style scoped>
.assistant-page-row {
    min-height: calc(100vh - 188px);
}

.assistant-page-col {
    display: flex;
}

.assistant-card {
    width: 100%;
    display: flex;
    flex-direction: column;
    border-radius: 16px;
    overflow: hidden;
}

.assistant-title-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    row-gap: 10px;
    column-gap: 14px;
}

.assistant-title-main {
    display: flex;
    align-items: center;
    min-width: 0;
}

.assistant-title-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    flex-wrap: wrap;
    column-gap: 8px;
    row-gap: 8px;
}

.assistant-conversation-shell {
    display: flex;
    flex-direction: column;
    min-height: clamp(520px, calc(100vh - 260px), 760px);
    flex: 1;
}

.assistant-messages-panel {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 20px;
    background: linear-gradient(
        180deg,
        rgba(var(--v-theme-surface), 0.98) 0%,
        rgba(var(--v-theme-surface), 1) 72%
    );
}

.assistant-empty-stage {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100%;
}

.assistant-message-list {
    width: 100%;
}

.assistant-empty-state {
    max-width: 660px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    opacity: 0.65;
    margin-bottom: 18px;
}

.assistant-message-item {
    display: flex;
    margin-bottom: 16px;
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
    max-width: min(82%, 860px);
    border-radius: 14px;
    padding: 12px 14px;
    background: rgba(var(--v-theme-surface), 1);
    border: 1px solid rgba(var(--v-theme-outline), 0.18);
    box-shadow: 0 2px 10px rgba(15, 23, 42, 0.03);
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
    line-height: 1.55;
}

.assistant-session-actions { display: flex; flex-wrap: wrap; gap: 8px; padding: 8px 18px; }

.assistant-composer {
    display: flex;
    align-items: flex-end;
    column-gap: 10px;
}

.assistant-composer-inline {
    width: min(820px, 100%);
}

.assistant-composer-docked {
    padding: 14px 18px 16px;
    background: rgba(var(--v-theme-surface), 0.98);
    border-top: 1px solid rgba(var(--v-theme-outline), 0.12);
}

.assistant-input {
    flex: 1;
    min-width: 0;
}

.assistant-send-button {
    height: 44px;
    min-width: 106px;
}

@media (max-width: 960px) {
    .assistant-page-row {
        min-height: 0;
    }

    .assistant-conversation-shell {
        min-height: 480px;
    }

    .assistant-message-bubble {
        max-width: 100%;
    }

    .assistant-composer {
        flex-direction: column;
        align-items: stretch;
    }

    .assistant-send-button {
        width: 100%;
    }
}
</style>
