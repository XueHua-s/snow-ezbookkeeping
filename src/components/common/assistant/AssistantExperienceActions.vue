<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantVersionDTO } from '@/models/ai_assistant_conversation.ts';
import { assistantStateLabels } from './assistantPresentation.ts';
import AssistantCopyButton from './AssistantCopyButton.vue';
defineProps<{ content: string; streaming?: boolean; retry?: boolean; regenerate?: boolean; edit?: boolean; branch?: boolean; versions?: readonly AssistantVersionDTO[]; selectedVersionId?: string; contextVersionId?: string; hasMoreVersions?: boolean; disabledReason?: string }>();
const emit = defineEmits<{ retry: []; regenerate: []; edit: []; branch: []; selectVersion: [id: string]; loadVersions: [] }>();
const { tt } = useI18n();
function closeMenu(event: KeyboardEvent): void { const menu = event.currentTarget as HTMLDetailsElement; menu.open = false; menu.querySelector('summary')?.focus(); }
</script>
<template>
    <div class="experience-actions">
        <AssistantCopyButton v-if="content" :text="content" :streaming="streaming" />
        <button v-if="retry" type="button" :disabled="!!disabledReason" @click="emit('retry')">{{ tt('Retry') }}</button>
        <button v-if="regenerate" type="button" :disabled="!!disabledReason" @click="emit('regenerate')">{{ tt('Regenerate') }}</button>
        <label v-if="versions && versions.length > 1">{{ tt('Version') }}
            <select :value="selectedVersionId" :aria-label="tt('Answer version')" @change="emit('selectVersion', ($event.target as HTMLSelectElement).value)">
                <option v-for="version in versions" :key="version.id" :value="version.id">{{ version.attemptNumber }} · {{ tt(assistantStateLabels[version.state]) }}{{ version.id === contextVersionId ? ' · ' + tt('Used for next question') : '' }}</option>
            </select>
        </label>
        <button v-if="hasMoreVersions" type="button" @click="emit('loadVersions')">{{ tt('Older versions') }}</button>
        <details v-if="edit || branch" class="more" @keydown.esc="closeMenu">
            <summary>{{ tt('More') }}</summary>
            <div><button v-if="edit" type="button" :disabled="!!disabledReason" @click="emit('edit')">{{ tt('Edit and create branch') }}</button><button v-if="branch" type="button" :disabled="!!disabledReason" @click="emit('branch')">{{ tt('Create branch from here') }}</button></div>
        </details>
        <span v-if="disabledReason" class="reason">{{ tt(disabledReason) }}</span>
    </div>
</template>
<style scoped>
.experience-actions { display: flex; flex-wrap: wrap; align-items: center; gap: 2px 8px; font-size: .85em; margin-top: 8px; }
button, summary, select { width: auto; min-height: 44px; padding: 8px 12px; border-radius: 8px; border: 1px solid transparent; background: transparent; color: inherit; font: inherit; cursor: pointer; }
select { border-color: rgb(127 127 127 / 90%); max-width: 100%; } option { color: CanvasText; background: Canvas; }
button:hover, summary:hover { background: rgb(127 127 127 / 12%); } button:focus-visible, summary:focus-visible, select:focus-visible { outline: 2px solid currentColor; outline-offset: 2px; } button:disabled { opacity: .55; cursor: default; }
.more { position: relative; } .more[open] > div { display: flex; flex-direction: column; position: absolute; z-index: 3; inset-inline-end: 0; min-width: 210px; max-width: 80vw; background: Canvas; color: CanvasText; border: 1px solid; border-radius: 8px; padding: 4px; box-shadow: 0 5px 16px #0002; }
.reason { width: 100%; }
</style>
