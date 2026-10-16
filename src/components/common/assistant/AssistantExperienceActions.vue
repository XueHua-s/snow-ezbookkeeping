<script setup lang="ts">
import { computed } from 'vue';
import { mdiRefresh, mdiHistory, mdiChevronLeft, mdiChevronRight } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantVersionDTO } from '@/models/ai_assistant_conversation.ts';
import AssistantCopyButton from './AssistantCopyButton.vue';
const props = defineProps<{
    content: string;
    streaming?: boolean;
    retry?: boolean;
    regenerate?: boolean;
    versions?: readonly AssistantVersionDTO[];
    selectedVersionId?: string;
    selectedAttemptNumber?: number;
    canUseAnswer?: boolean;
    hasMoreVersions?: boolean;
    disabledReason?: string;
    writeDisabledReason?: string;
}>();
const emit = defineEmits<{
    retry: [];
    regenerate: [];
    selectVersion: [id: string];
    useAnswer: [];
    loadVersions: [];
}>();
const { tt } = useI18n();
const orderedVersions = computed(() => [...(props.versions ?? [])].sort((a, b) => a.attemptNumber - b.attemptNumber));
const attempt = computed(() => props.selectedAttemptNumber ?? orderedVersions.value.find(item => item.id === props.selectedVersionId)?.attemptNumber ?? 1);
const total = computed(() => Math.max(attempt.value, ...orderedVersions.value.map(item => item.attemptNumber)));
const earlier = computed(() => orderedVersions.value.findLast(item => item.attemptNumber < attempt.value));
const writeReason = computed(() => props.disabledReason || props.writeDisabledReason);
const later = computed(() => orderedVersions.value.find(item => item.attemptNumber > attempt.value));
</script>
<template>
    <div class="experience-actions">
        <AssistantCopyButton v-if="content" :text="content" :streaming="streaming" />
        <button v-if="retry" class="icon-action" type="button" :aria-label="tt('Retry')" :title="tt(writeReason || 'Retry')" :disabled="!!writeReason" @click="emit('retry')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <button v-if="regenerate" class="icon-action" type="button" :aria-label="tt('Regenerate')" :title="tt(writeReason || 'Regenerate')" :disabled="!!writeReason" @click="emit('regenerate')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <div v-if="total > 1" class="answer-navigation" role="group" :aria-label="tt('Answer versions')">
            <button class="icon-action" type="button" :aria-label="tt('Earlier answer')" :title="tt('Earlier answer')" :disabled="!!disabledReason || !earlier" @click="earlier && emit('selectVersion', earlier.id)">
                <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiChevronLeft" /></svg>
            </button>
            <span class="answer-position" aria-live="polite">{{ attempt }} / {{ total }}</span>
            <button class="icon-action" type="button" :aria-label="tt('Newer answer')" :title="tt('Newer answer')" :disabled="!!disabledReason || !later" @click="later && emit('selectVersion', later.id)">
                <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiChevronRight" /></svg>
            </button>
        </div>
        <button v-if="hasMoreVersions" class="text-action" type="button" :disabled="!!disabledReason" @click="emit('loadVersions')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiHistory" /></svg>{{ tt('Load earlier answers') }}
        </button>
        <button v-if="canUseAnswer" type="button" :disabled="!!writeReason" :title="writeReason ? tt(writeReason) : undefined" @click="emit('useAnswer')">{{ tt('Continue with this answer') }}</button>
    </div>
</template>
<style scoped>
.experience-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 2px;
    color: var(--assistant-muted, #6b8074);
    font-size: 0.85em;
    margin-top: 8px;
}
button {
    width: auto;
    min-height: 44px;
    padding: 8px 12px;
    border-radius: 14px;
    border: 1px solid transparent;
    background: transparent;
    color: inherit;
    font: inherit;
    cursor: pointer;
    transition: background 180ms ease, color 180ms ease, transform 180ms ease;
}
button:hover:not(:disabled) {
    background: var(--assistant-tint, #e8f3e9);
    color: var(--assistant-action, #278c62);
}
button:focus-visible {
    outline: 2px solid var(--assistant-action, #278c62);
    outline-offset: 2px;
}
button:disabled {
    opacity: 0.55;
    cursor: default;
}
.icon-action { width: 44px; display: inline-grid; place-items: center; padding: 10px; }
svg { width: 20px; height: 20px; fill: currentColor; }
button:active:not(:disabled) { transform: scale(.94); }
.answer-navigation, .text-action { display: inline-flex; align-items: center; gap: 4px; }
.answer-position { font-variant-numeric: tabular-nums; white-space: nowrap; }
@media (prefers-reduced-motion: reduce) { button { transition: none; } }
</style>
