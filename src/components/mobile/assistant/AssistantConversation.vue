<script setup lang="ts">
import type { Router } from 'framework7/types';
import AssistantMessageList from '@/components/common/assistant/AssistantMessageList.vue';
import AssistantComposer from '@/components/mobile/assistant/AssistantComposer.vue';
import AssistantActions from '@/components/mobile/assistant/AssistantActions.vue';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useAssistantPageBase } from '@/views/base/assistant/AssistantPageBase.ts';

const props = defineProps<{ f7router: Router.Router }>();

const { tt } = useI18n();
const { showToast } = useI18nUIComponents();
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

function openReference(id: string): void {
    props.f7router.navigate(`/transaction/detail?id=${encodeURIComponent(id)}`);
}

function sendNextPage(): void {
    requestNextPage(tt('Show the next page of the current query results')).catch(error => {
        if (!error.processed) showToast(error.message || error);
    });
}

function sendChatMessage(): void {
    sendMessage().catch(error => {
        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function generateSummaryMessage(): void {
    generateSummary(tt('Generate AI Summary')).catch(error => {
        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}
</script>

<template>
    <f7-page @page:beforeout="disposeConversation">
        <f7-navbar :title="tt('AI Assistant')" :back-link="tt('Back')"></f7-navbar>

        <f7-block strong inset v-if="!enabled">
            {{ tt('AI assistant is disabled') }}
        </f7-block>

        <template v-else>
            <f7-block class="assistant-top-actions display-flex justify-content-space-between align-items-center">
                <f7-button type="button" class="assistant-top-action-btn" fill
                           :disabled="requesting"
                           @click="generateSummaryMessage">
                    {{ tt('Generate AI Summary') }}
                </f7-button>
                <f7-button type="button" class="assistant-top-action-btn" outline
                           :disabled="!messages.length"
                           @click="clearConversation">
                    {{ tt('Clear Conversation') }}
                </f7-button>
            </f7-block>

            <AssistantMessageList :messages="messages" :enabled="enabled" variant="mobile" @open-reference="openReference">
                <template #empty-icon><f7-icon f7="chat_bubble_2" class="margin-bottom-half" /></template>
            </AssistantMessageList>

            <AssistantActions :requesting="requesting" :can-request-next-page="canRequestNextPage" @next="sendNextPage" @cancel="cancelCurrentRequest" />

            <AssistantComposer v-model="messageInput" :disabled="requesting" :requesting="requesting" :can-send="canSendMessage" @send="sendChatMessage" />
        </template>
    </f7-page>
</template>

<style scoped>
.assistant-top-actions {
    margin-top: 10px;
    margin-bottom: 8px;
    column-gap: 8px;
}

.assistant-top-action-btn {
    flex: 1;
    min-width: 0;
}

</style>
