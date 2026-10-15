<script setup lang="ts">
import { onBeforeRouteLeave } from 'vue-router';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';
import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';
import SnackBar from '@/components/desktop/SnackBar.vue';
import AssistantMessageList from '@/components/common/assistant/AssistantMessageList.vue';
import AssistantComposer from '@/components/desktop/assistant/AssistantComposer.vue';
import AssistantActions from '@/components/desktop/assistant/AssistantActions.vue';

import { useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { getAIAssistantModelID } from '@/lib/server_settings.ts';
import { useAssistantPageBase } from '@/views/base/assistant/AssistantPageBase.ts';

import {
    mdiRobotOutline,
    mdiTextBoxCheckOutline,
    mdiDeleteOutline,
    mdiMessageTextOutline
} from '@mdi/js';

const { tt } = useI18n();
const aiAssistantModelID = getAIAssistantModelID();
const {
    enabled,
    messages,
    messageInput,
    requesting,
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

const snackbar = useTemplateRef<InstanceType<typeof SnackBar>>('snackbar');
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
                                   :disabled="!enabled || requesting"
                                   :loading="requesting"
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
                    <AssistantMessageList :messages="messages" :enabled="enabled" variant="desktop" @open-reference="openReference">
                        <template #empty-icon><v-icon :icon="mdiMessageTextOutline" size="32" class="mb-3" /></template>
                        <template #empty><AssistantComposer class="assistant-composer-inline" v-model="messageInput" :disabled="!enabled || requesting" :requesting="requesting" :can-send="canSendMessage" @send="sendChatMessage" /></template>
                    </AssistantMessageList>

                    <v-divider />

                    <AssistantActions :requesting="requesting" :can-request-next-page="canRequestNextPage" @next="sendNextPage" @cancel="cancelCurrentRequest" />

                    <AssistantComposer class="assistant-composer-docked" v-if="enabled && messages.length > 0" v-model="messageInput" :disabled="requesting" :requesting="requesting" :can-send="canSendMessage" @send="sendChatMessage" />
                </div>
            </v-card>
        </v-col>
    </v-row>

    <EditDialog ref="referenceDialog" :type="TransactionEditPageType.Transaction" />
    <SnackBar ref="snackbar" />
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
    height: clamp(520px, calc(100vh - 260px), 760px);
    overflow: hidden;
    flex: 0 0 auto;
}

.assistant-composer-inline { width: min(820px, 100%); }
.assistant-composer-docked { padding: 14px 18px 16px; background: rgba(var(--v-theme-surface), 0.98); border-top: 1px solid rgba(var(--v-theme-outline), 0.12); }
@media (max-width: 960px) { .assistant-page-row { min-height: 0; } .assistant-conversation-shell { height: 480px; } }
</style>
