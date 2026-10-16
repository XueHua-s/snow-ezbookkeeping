<script setup lang="ts">
import { mdiArrowTopRight, mdiPlus, mdiRefresh } from '@mdi/js';
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
    <p key="notice-1" v-if="store.conversationId && !store.detail && !store.loading" class="assistant-notice" role="status">
        {{ tt('This conversation does not exist or is not accessible.') }}
        <button class="assistant-action" type="button" @click="emit('navigate', '')" :aria-label="tt('New conversation')" :title="tt('New conversation')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiPlus" /></svg>
        </button>
    </p>
    <p key="notice-2" v-if="store.loading" class="assistant-notice" role="status">{{ tt('Loading') }}</p>
    <p key="notice-3" v-if="conversation?.sourceInvalidated" class="assistant-notice" role="status">
        {{
            tt(
                'Source data has changed or was removed. Old context will not be reused. Ask again to query current bills.'
            )
        }}
    </p>
    <p key="notice-4" v-if="conversation?.sourceConversationId" class="assistant-notice">
        {{
            tt(
                conversation.sourceInvalidated
                    ? 'The source is no longer available. Its context cannot be used.'
                    : 'Branched conversation'
            )
        }}
        <button
            class="assistant-action"
            v-if="!conversation.sourceInvalidated"
            type="button"
            @click="emit('navigate', conversation.sourceConversationId, conversation.sourceMessageId)"
         :aria-label="tt('View source')" :title="tt('View source')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiArrowTopRight" /></svg>
        </button>
    </p>

    <p key="notice-5" v-if="conversation?.pendingContinuation" class="assistant-notice">
        {{ tt('New branch created. Review your question below and send to start generating.') }}
    </p>
    <p key="notice-6" v-if="conversation?.pendingQuerySourceMessageId" class="assistant-notice">
        <button
            class="assistant-action"
            type="button"
            :disabled="!!disabledReason || store.busy"
            @click="emit('continueQuery')"
        >
            {{ tt('Continue this query') }}
        </button>
        <button
            class="assistant-action"
            type="button"
            :disabled="!!disabledReason || store.busy"
            @click="emit('restartQuery', conversation.pendingQuerySourceMessageId)"
         :aria-label="tt('Query these bills again')" :title="tt('Query these bills again')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
    </p>
    <p key="notice-7" v-if="conversation?.archived" class="assistant-notice">
        <button class="assistant-action" type="button" @click="emit('unarchive', conversation)">
            {{ tt('Unarchive and continue') }}
        </button>
    </p>
</template>
<style scoped src="./assistantAction.css"></style>
<style scoped src="./assistantNotice.css"></style>
