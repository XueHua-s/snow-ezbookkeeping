<script setup lang="ts">
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
    <p v-if="store.legacyBackend" class="notice" role="status">
        {{ tt('Upgrade the server to use saved assistant conversations.') }}
    </p>
    <p v-if="!store.online" class="notice" role="status">
        {{ tt('Offline. Drafts are saved on this device.') }}
    </p>
    <p v-if="store.local.scope === 'page'" class="notice" role="status">
        {{ tt('Only kept on this page. Refreshing may lose unsent content.') }}
        <button class="recovery-action" type="button" @click="emit('retryPersistence')">{{ tt('Retry') }}</button>
    </p>
    <div v-for="pending in unresolvedPending" :key="pending.clientRequestId" class="notice" role="status">
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
        <button class="recovery-action" type="button" @click="emit('reconcile')">
            {{ tt('Check status') }}
        </button>
        <button
            class="recovery-action"
            v-if="!pending.rejected && store.delivery[pending.clientRequestId] === 'unsent'"
            type="button"
            @click="emit('retryPending', pending.clientRequestId)"
        >
            {{ tt('Retry') }}
        </button>
        <button
            class="recovery-action"
            v-if="pending.rejected"
            type="button"
            @click="emit('discard', pending.clientRequestId)"
        >
            {{ tt('Discard rejected request') }}
        </button>
        <button
            class="recovery-action"
            v-if="pending.text && !store.draft"
            type="button"
            @click="emit('restore', pending.clientRequestId)"
        >
            {{ tt('Restore this question') }}
        </button>
    </div>
    <details v-if="store.local.drafts.length > 1 || store.local.reason" class="notice">
        <summary class="recovery-action">{{ tt('Local drafts') }}</summary>
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
                class="recovery-action"
                v-if="draft.conversationId === store.conversationId"
                type="button"
                @click="store.chooseDraft(draft.id)"
            >
                {{ tt('Use this draft') }}
            </button>
            <button class="recovery-action" type="button" @click="emit('removeDraft', draft.id)">
                {{ tt('Delete') }}
            </button>
        </article>
    </details>
</template>
<style scoped>
.notice {
    margin: 0;
    padding: 8px 16px;
    font-size: 0.875rem;
    max-height: 180px;
    overflow-y: auto;
}
.recovery-text {
    white-space: pre-wrap;
    overflow-wrap: anywhere;
}
.recovery-action {
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
.recovery-action:focus-visible {
    outline: 2px solid currentColor;
    outline-offset: 2px;
}
</style>
