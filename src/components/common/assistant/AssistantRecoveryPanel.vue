<script setup lang="ts">
import { mdiDeleteOutline, mdiRefresh, mdiRestore, mdiSync, mdiTextBoxEditOutline } from '@mdi/js';
import { computed } from 'vue';
import { useAssistantStore } from '@/stores/assistant.ts';
import { useI18n } from '@/locales/helpers.ts';
import AssistantCopyButton from './AssistantCopyButton.vue';
const emit = defineEmits<{
    retryPersistence: [];
    reconcile: [];
    retryPending: [id: string];
    discard: [id: string];
    restore: [id: string];
    removeDraft: [id: string];
}>();
const store = useAssistantStore();
const { tt } = useI18n();
const unresolvedPending = computed(() => store.local.pending.filter((item) => !item.resolvedConversationId));
</script>
<template>
    <p key="notice-1" v-if="store.legacyBackend" class="assistant-notice" role="status">
        {{ tt('Upgrade the server to use saved assistant conversations.') }}
    </p>
    <p key="notice-2" v-if="!store.online" class="assistant-notice" role="status">
        {{ tt('Offline. Drafts are saved on this device.') }}
    </p>
    <p key="notice-3" v-if="store.local.scope === 'page'" class="assistant-notice" role="status">
        {{ tt('Only kept on this page. Refreshing may lose unsent content.') }}
        <button class="assistant-action" type="button" @click="emit('retryPersistence')" :aria-label="tt('Retry')" :title="tt('Retry')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
    </p>
    <div v-for="pending in unresolvedPending" :key="pending.clientRequestId" class="assistant-notice" role="status">
        <span>
            {{
                tt(
                    pending.rejected || store.delivery[pending.clientRequestId] === 'unsent'
                        ? 'Not delivered'
                        : 'Confirming changes'
                )
            }}
        </span>
        <pre class="recovery-text" v-if="pending.text">{{ pending.text }}</pre>
        <AssistantCopyButton v-if="pending.text" :text="pending.text" />
        <button class="assistant-action" type="button" @click="emit('reconcile')" :aria-label="tt('Check status')" :title="tt('Check status')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiSync" /></svg>
        </button>
        <button
            class="assistant-action"
            v-if="!pending.rejected && store.delivery[pending.clientRequestId] === 'unsent'"
            type="button"
            @click="emit('retryPending', pending.clientRequestId)"
         :aria-label="tt('Retry')" :title="tt('Retry')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <button
            class="assistant-action"
            v-if="pending.rejected"
            type="button"
            @click="emit('discard', pending.clientRequestId)"
         :aria-label="tt('Discard rejected request')" :title="tt('Discard rejected request')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiDeleteOutline" /></svg>
        </button>
        <button
            class="assistant-action"
            v-if="pending.text && !store.draft"
            type="button"
            @click="emit('restore', pending.clientRequestId)"
         :aria-label="tt('Restore this question')" :title="tt('Restore this question')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRestore" /></svg>
        </button>
    </div>
    <details key="notice-4" v-if="store.local.drafts.length > 1 || store.local.reason" class="assistant-notice">
        <summary class="assistant-action">{{ tt('Local drafts') }}</summary>
        <article v-for="draft in store.local.drafts" :key="draft.id">
            <p>
                {{
                    draft.conflict
                        ? tt('Conflicting draft — both copies have been kept')
                        : draft.conversationId || tt('New conversation')
                }}
            </p>
            <pre class="recovery-text">{{ draft.text }}</pre>
            <AssistantCopyButton :text="draft.text" />
            <button
                class="assistant-action"
                v-if="draft.conversationId === store.conversationId"
                type="button"
                @click="store.chooseDraft(draft.id)"
             :aria-label="tt('Use this draft')" :title="tt('Use this draft')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiTextBoxEditOutline" /></svg>
        </button>
            <button class="assistant-action" type="button" @click="emit('removeDraft', draft.id)" :aria-label="tt('Delete')" :title="tt('Delete')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiDeleteOutline" /></svg>
        </button>
        </article>
    </details>
</template>
<style scoped src="./assistantAction.css"></style>
<style scoped src="./assistantNotice.css"></style>
<style scoped>
.recovery-text { white-space: pre-wrap; overflow-wrap: anywhere; font: inherit; }
summary.assistant-action { display: list-item; background: transparent; }
</style>
