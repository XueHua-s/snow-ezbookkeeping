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
    <TransitionGroup name="notice">
    <p key="notice-1" v-if="store.legacyBackend" class="notice" role="status">
        {{ tt('Upgrade the server to use saved assistant conversations.') }}
    </p>
    <p key="notice-2" v-if="!store.online" class="notice" role="status">
        {{ tt('Offline. Drafts are saved on this device.') }}
    </p>
    <p key="notice-3" v-if="store.local.scope === 'page'" class="notice" role="status">
        {{ tt('Only kept on this page. Refreshing may lose unsent content.') }}
        <button class="recovery-action" type="button" @click="emit('retryPersistence')" :aria-label="tt('Retry')" :title="tt('Retry')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
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
        <button class="recovery-action" type="button" @click="emit('reconcile')" :aria-label="tt('Check status')" :title="tt('Check status')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiSync" /></svg>
        </button>
        <button
            class="recovery-action"
            v-if="!pending.rejected && store.delivery[pending.clientRequestId] === 'unsent'"
            type="button"
            @click="emit('retryPending', pending.clientRequestId)"
         :aria-label="tt('Retry')" :title="tt('Retry')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <button
            class="recovery-action"
            v-if="pending.rejected"
            type="button"
            @click="emit('discard', pending.clientRequestId)"
         :aria-label="tt('Discard rejected request')" :title="tt('Discard rejected request')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiDeleteOutline" /></svg>
        </button>
        <button
            class="recovery-action"
            v-if="pending.text && !store.draft"
            type="button"
            @click="emit('restore', pending.clientRequestId)"
         :aria-label="tt('Restore this question')" :title="tt('Restore this question')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRestore" /></svg>
        </button>
    </div>
    <details key="notice-4" v-if="store.local.drafts.length > 1 || store.local.reason" class="notice">
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
             :aria-label="tt('Use this draft')" :title="tt('Use this draft')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiTextBoxEditOutline" /></svg>
        </button>
            <button class="recovery-action" type="button" @click="emit('removeDraft', draft.id)" :aria-label="tt('Delete')" :title="tt('Delete')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiDeleteOutline" /></svg>
        </button>
        </article>
    </details>
    </TransitionGroup>
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
.notice-enter-active, .notice-leave-active, .notice-move { transition: opacity 180ms ease, transform 180ms ease; }
.notice-enter-from, .notice-leave-to { opacity: 0; transform: translateY(4px); }

.recovery-action {
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
.recovery-action:hover:not(:disabled) { background: color-mix(in srgb, var(--assistant-action, #278c62) 16%, var(--assistant-card, #fff)); }
.recovery-action:active:not(:disabled) { transform: scale(0.96); }
.recovery-action:disabled { opacity: 0.5; cursor: default; }
.recovery-action:focus-visible { outline: 2px solid var(--assistant-action, #278c62); outline-offset: 3px; }
.recovery-action svg { width: 20px; height: 20px; fill: currentColor; flex: none; }
.recovery-text { white-space: pre-wrap; overflow-wrap: anywhere; font: inherit; }
summary.recovery-action { display: list-item; background: transparent; }

@media (prefers-reduced-motion: reduce) {
    *, *::backdrop { animation: none !important; transition: none !important; }
}
</style>
