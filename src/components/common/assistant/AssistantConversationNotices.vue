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
    <p key="notice-1" v-if="store.conversationId && !store.detail && !store.loading" class="notice" role="status">
        {{ tt('This conversation does not exist or is not accessible.') }}
        <button class="workspace-action" type="button" @click="emit('navigate', '')" :aria-label="tt('New conversation')" :title="tt('New conversation')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiPlus" /></svg>
        </button>
    </p>
    <p key="notice-2" v-if="store.loading" class="notice" role="status">{{ tt('Loading') }}</p>
    <p key="notice-3" v-if="conversation?.sourceInvalidated" class="notice" role="status">
        {{
            tt(
                'Source data has changed or was removed. Old context will not be reused. Ask again to query current bills.'
            )
        }}
    </p>
    <p key="notice-4" v-if="conversation?.sourceConversationId" class="notice">
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
         :aria-label="tt('View source')" :title="tt('View source')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiArrowTopRight" /></svg>
        </button>
    </p>

    <p key="notice-5" v-if="conversation?.pendingContinuation" class="notice">
        {{ tt('New branch created. Review your question below and send to start generating.') }}
    </p>
    <p key="notice-6" v-if="conversation?.pendingQuerySourceMessageId" class="notice">
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
         :aria-label="tt('Query these bills again')" :title="tt('Query these bills again')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
    </p>
    <p key="notice-7" v-if="conversation?.archived" class="notice">
        <button class="workspace-action" type="button" @click="emit('unarchive', conversation)">
            {{ tt('Unarchive and continue') }}
        </button>
    </p>
</template>
<style scoped>
.notice {
    margin: 8px 16px;
    padding: 12px 16px;
    border-radius: 16px;
    background: var(--assistant-tint, #e8f3e9);
    color: var(--assistant-text, #233e30);
    font-size: 0.875rem;
    line-height: 1.6;
    max-height: 180px;
    overflow-y: auto;
}
/* Remove obsolete content immediately; only its arrival is animated. */
.notice { animation: notice-enter 180ms ease; }
@keyframes notice-enter { from { opacity: 0; transform: translateY(4px); } }

.workspace-action {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-width: 44px;
    min-height: 44px;
    margin: 4px 6px 0 0;
    padding: 8px 14px;
    border: 0;
    border-radius: 14px;
    color: var(--assistant-action, #278c62);
    background: var(--assistant-tint, #e8f3e9);
    font: inherit;
    cursor: pointer;
    transition: background 160ms ease, transform 160ms ease, opacity 160ms ease;
}
.workspace-action:hover:not(:disabled) { background: color-mix(in srgb, var(--assistant-action, #278c62) 16%, var(--assistant-card, #fff)); }
.workspace-action:active:not(:disabled) { transform: scale(0.96); }
.workspace-action:disabled { opacity: 0.5; cursor: default; }
.workspace-action:focus-visible { outline: 2px solid var(--assistant-action, #278c62); outline-offset: 3px; }
.workspace-action svg { width: 20px; height: 20px; fill: currentColor; flex: none; }

@media (prefers-reduced-motion: reduce) {
    *, *::backdrop { animation: none !important; transition: none !important; }
}
</style>
