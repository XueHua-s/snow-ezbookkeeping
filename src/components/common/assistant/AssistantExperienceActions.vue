<script setup lang="ts">
import { mdiRefresh, mdiHistory } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantVersionDTO } from '@/models/ai_assistant_conversation.ts';
import { assistantStateLabels } from './assistantPresentation.ts';
import AssistantCopyButton from './AssistantCopyButton.vue';
defineProps<{
    content: string;
    streaming?: boolean;
    retry?: boolean;
    regenerate?: boolean;
    versions?: readonly AssistantVersionDTO[];
    selectedVersionId?: string;
    contextVersionId?: string;
    hasMoreVersions?: boolean;
    disabledReason?: string;
}>();
const emit = defineEmits<{
    retry: [];
    regenerate: [];
    selectVersion: [id: string];
    loadVersions: [];
}>();
const { tt } = useI18n();
</script>
<template>
    <div class="experience-actions">
        <AssistantCopyButton v-if="content" :text="content" :streaming="streaming" />
        <button v-if="retry" class="icon-action" type="button" :aria-label="tt('Retry')" :title="tt('Retry')" :disabled="!!disabledReason" @click="emit('retry')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <button v-if="regenerate" class="icon-action" type="button" :aria-label="tt('Regenerate')" :title="tt('Regenerate')" :disabled="!!disabledReason" @click="emit('regenerate')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg>
        </button>
        <label v-if="versions && versions.length > 1"
            >{{ tt('Version') }}
            <select
                name="answer-version"
                :value="selectedVersionId"
                :aria-label="tt('Answer version')"
                @change="emit('selectVersion', ($event.target as HTMLSelectElement).value)"
            >
                <option v-for="version in versions" :key="version.id" :value="version.id">
                    {{ version.attemptNumber }} · {{ tt(assistantStateLabels[version.state])
                    }}{{ version.id === contextVersionId ? ' · ' + tt('Used for next question') : '' }}
                </option>
            </select>
        </label>
        <button v-if="hasMoreVersions" class="icon-action" :aria-label="tt('Older versions')" :title="tt('Older versions')" type="button" @click="emit('loadVersions')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiHistory" /></svg>
        </button>
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
button,
summary,
select {
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
select {
    border-color: var(--assistant-border, #e5ece3);
    max-width: 100%;
    background: var(--assistant-card, #fff);
}
option {
    color: CanvasText;
    background: Canvas;
}
button:hover,
summary:hover {
    background: var(--assistant-tint, #e8f3e9);
    color: var(--assistant-action, #278c62);
}
button:focus-visible,
summary:focus-visible,
select:focus-visible {
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
label { min-width: 0; display: flex; align-items: center; gap: 8px; margin-inline-start: auto; max-width: 100%; }
@media (prefers-reduced-motion: reduce) { button, select { transition: none; } }
</style>
