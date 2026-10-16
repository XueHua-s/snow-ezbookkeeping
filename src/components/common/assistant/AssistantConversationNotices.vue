<script setup lang="ts">
import { computed } from 'vue';
import { useAssistantStore } from '@/stores/assistant.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantConversationDTO } from '@/models/ai_assistant_conversation.ts';
defineProps<{ disabledReason?: string }>();
const emit = defineEmits<{
    navigate: [id: string, messageId?: string];
    continueQuery: [];
    restartQuery: [sourceId: string];
    unarchive: [conversation: AssistantConversationDTO];
}>();
const store = useAssistantStore();
const { tt } = useI18n();
const conversation = computed(() => store.detail?.conversation);
</script>
<template>
    <p v-if="store.conversationId && !store.detail && !store.loading" class="notice" role="status">
        {{ tt('This conversation does not exist or is not accessible.') }}
        <button class="workspace-action" type="button" @click="emit('navigate', '')">
            {{ tt('New conversation') }}
        </button>
    </p>
    <p v-if="store.loading" class="notice" role="status">{{ tt('Loading') }}</p>
    <p v-if="conversation?.sourceInvalidated" class="notice" role="status">
        {{
            tt(
                'Source data has changed or was removed. Old context will not be reused. Ask again to query current bills.'
            )
        }}
    </p>
    <p v-if="conversation?.sourceConversationId" class="notice">
        {{
            tt(
                conversation.sourceInvalidated
                    ? 'The source is no longer available. Its context cannot be used.'
                    : 'Branched conversation'
            )
        }}
        <button
            class="workspace-action"
            v-if="!conversation.sourceInvalidated"
            type="button"
            @click="emit('navigate', conversation.sourceConversationId, conversation.sourceMessageId)"
        >
            {{ tt('View source') }}
        </button>
    </p>

    <p v-if="conversation?.pendingContinuation" class="notice">
        {{ tt('New branch created. Review your question below and send to start generating.') }}
    </p>
    <p v-if="conversation?.pendingQuerySourceMessageId" class="notice">
        <button
            class="workspace-action"
            type="button"
            :disabled="!!disabledReason || store.busy"
            @click="emit('continueQuery')"
        >
            {{ tt('Continue this query') }}
        </button>
        <button
            class="workspace-action"
            type="button"
            :disabled="!!disabledReason || store.busy"
            @click="emit('restartQuery', conversation.pendingQuerySourceMessageId)"
        >
            {{ tt('Query these bills again') }}
        </button>
    </p>
    <p v-if="conversation?.archived" class="notice">
        <button class="workspace-action" type="button" @click="emit('unarchive', conversation)">
            {{ tt('Unarchive and continue') }}
        </button>
    </p>
</template>
<style scoped>
.notice {
    margin: 0;
    padding: 8px 16px;
    font-size: 0.875rem;
    max-height: 180px;
    overflow-y: auto;
}
.workspace-action {
    color: inherit;
    font: inherit;
    min-height: 44px;
    padding: 8px;
    border: 1px solid rgb(127 127 127 / 90%);
    border-radius: 8px;
    background: transparent;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
}
.workspace-action:disabled {
    opacity: 0.55;
    cursor: default;
}
.workspace-action:focus-visible {
    outline: 2px solid currentColor;
    outline-offset: 2px;
}
</style>
