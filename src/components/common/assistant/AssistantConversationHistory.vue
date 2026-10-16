<script setup lang="ts">
import { nextTick, shallowRef } from 'vue';
import { mdiClose, mdiPlus, mdiPencilOutline, mdiArchiveOutline, mdiArchiveArrowUpOutline, mdiTrashCanOutline, mdiCheck, mdiMagnify, mdiRefresh } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantConversationDTO, AssistantSearchMatch } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{
    items: readonly AssistantConversationDTO[];
    matches?: readonly AssistantSearchMatch[];
    selectedId?: string;
    loading?: boolean;
    error?: string;
    hasMore?: boolean;
    archived?: boolean;
    titleLimit: number;
}>();
const emit = defineEmits<{
    select: [id: string, messageId?: string, versionId?: string];
    new: [];
    archived: [value: boolean];
    rename: [item: AssistantConversationDTO, title: string];
    archive: [item: AssistantConversationDTO];
    delete: [item: AssistantConversationDTO];
    retry: [];
    loadMore: [];
    close: [];
}>();
const { tt } = useI18n();
const query = defineModel<string>('query', { required: true });
const includeVersions = defineModel<boolean>('includeVersions', { required: true });
const editingId = shallowRef('');
const title = shallowRef('');
defineExpose({
    finishEdit: () => {
        editingId.value = '';
        title.value = '';
    }
});
function cancelRename(event: Event): void {
    const row = (event.target as HTMLElement).closest('.history-item');
    editingId.value = '';
    void nextTick(() => row?.querySelector<HTMLButtonElement>('.rename-action')?.focus());
}
async function startRename(item: AssistantConversationDTO, event: Event): Promise<void> {
    const row = (event.currentTarget as HTMLElement).closest('.history-item');
    editingId.value = item.id;
    title.value = item.title;
    await nextTick();
    row?.querySelector<HTMLInputElement>('input')?.focus();
}
function rename(item: AssistantConversationDTO): void {
    if (title.value.trim() && Array.from(title.value).length <= props.titleLimit)
        emit('rename', item, title.value);
}
</script>
<template>
    <nav
        class="history"
        :aria-label="tt('Conversation history')"
        @keydown.esc.stop.prevent="editingId ? cancelRename($event) : emit('close')"
    >
        <header>
            <strong>{{ tt('Conversation history') }}</strong
            ><button class="icon-action" type="button" :aria-label="tt('Close')" :title="tt('Close')" @click="emit('close')"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiClose" /></svg></button>
        </header>
        <button class="new icon-action" type="button" :aria-label="tt('New conversation')" :title="tt('New conversation')" @click="emit('new')"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiPlus" /></svg></button>
        <label class="search-field"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiMagnify" /></svg><input v-model="query" type="search" name="assistant-history-search" :aria-label="tt('Search history')" :placeholder="tt('Search history')" autocomplete="off" /></label>
        <label class="check"
            ><input
                type="checkbox"
                :checked="archived"
                @change="emit('archived', ($event.target as HTMLInputElement).checked)"
            />{{ tt('Include archived') }}</label
        >
        <label class="check"
            ><input v-model="includeVersions" type="checkbox" />{{
                tt('Include older answer versions')
            }}</label
        >
        <p v-if="loading" role="status">{{ tt('Loading') }}</p>
        <p v-if="error" role="alert">
            {{ tt(error) }} <button class="icon-action" type="button" :aria-label="tt('Retry')" :title="tt('Retry')" @click="emit('retry')"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiRefresh" /></svg></button>
        </p>
        <template v-if="query"
            ><p v-if="!loading && !error && !matches?.length">{{ tt('No matching conversations') }}</p>
            <button
                v-for="match in matches"
                :key="`${match.conversationId}:${match.messageId}:${match.versionId ?? ''}`"
                type="button"
                class="result"
                @click="emit('select', match.conversationId, match.messageId, match.versionId)"
            >
                <strong>{{ match.title }}</strong
                ><span>{{ match.snippet }}</span>
            </button></template
        >
        <template v-else
            ><p v-if="!items.length && !loading && !error">{{ tt('No conversations yet') }}</p>
            <section
                v-for="item in items"
                :key="item.id"
                class="history-item"
                :class="{ selected: selectedId === item.id }"
            >
                <button
                    type="button"
                    class="select"
                    :title="item.title"
                    :aria-current="selectedId === item.id ? 'page' : undefined"
                    @click="emit('select', item.id)"
                >
                    <strong>{{ item.title }}</strong
                    ><time :datetime="item.activityAt">{{
                        new Date(item.activityAt).toLocaleDateString()
                    }}</time
                    ><span v-if="item.activeRunId">{{ tt('Generating answer') }}</span
                    ><span v-else-if="item.lastRunState === 'failed'">{{
                        tt('Assistant request failed')
                    }}</span>
                </button>
                <form v-if="editingId === item.id" @submit.prevent="rename(item)">
                    <label>{{ tt('Conversation title') }}<input v-model="title" name="conversation-title" autocomplete="off" /></label
                    ><span>{{ Array.from(title).length }} / {{ titleLimit }}</span
                    ><button class="icon-action" type="submit" :aria-label="tt('Save')" :title="tt('Save')" :disabled="!title.trim() || Array.from(title).length > titleLimit">
                        <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiCheck" /></svg></button
                    ><button class="icon-action" type="button" :aria-label="tt('Cancel')" :title="tt('Cancel')" @click="cancelRename"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiClose" /></svg></button>
                </form>
                <div v-else class="item-actions">
                    <button class="icon-action rename-action" type="button" :aria-label="tt('Rename')" :title="tt('Rename')" @click="startRename(item, $event)">
                        <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiPencilOutline" /></svg>
                    </button>
                    <button class="icon-action" type="button" :aria-label="tt(item.archived ? 'Unarchive' : 'Archive')" :title="tt(item.archived ? 'Unarchive' : 'Archive')" @click="emit('archive', item)">
                        <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="item.archived ? mdiArchiveArrowUpOutline : mdiArchiveOutline" /></svg>
                    </button>
                    <button class="icon-action delete-action" type="button" :aria-label="tt('Delete')" :title="tt('Delete')" @click="emit('delete', item)">
                        <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiTrashCanOutline" /></svg>
                    </button>
                </div>
            </section>
        </template>
        <button v-if="hasMore" type="button" :disabled="loading" @click="emit('loadMore')">
            {{ tt('Load more') }}
        </button>
    </nav>
</template>
<style scoped>
* { box-sizing: border-box; }
.history {
    width: 100%;
    padding: 16px;
    overflow-y: auto;
    overscroll-behavior: contain;
    height: 100%;
    line-height: 1.5;
    background: var(--assistant-background, #fbfcf9);
}
header { display: flex; align-items: center; justify-content: space-between; }
button {
    min-height: 44px;
    padding: 10px 12px;
    color: inherit;
    font: inherit;
    background: transparent;
    border: 0;
    border-radius: 14px;
    cursor: pointer;
    transition: background-color 180ms ease, color 180ms ease, transform 180ms ease;
}
button:active:not(:disabled) { transform: scale(0.96); }
button:disabled { opacity: 0.45; cursor: default; }
svg { width: 20px; height: 20px; fill: currentColor; flex-shrink: 0; }
.icon-action { width: 44px; height: 44px; flex: 0 0 44px; padding: 10px; display: inline-flex; justify-content: center; align-items: center; min-width: 44px; color: var(--assistant-muted, #698778); }
.new { width: 100%; margin-block: 8px 12px; background: var(--assistant-accent-soft, #e5f4e9); color: var(--assistant-accent, #278c62); }
.search-field { display: flex; align-items: center; gap: 8px; padding-inline: 12px; background: var(--assistant-soft-surface, #f0f4ee); border-radius: 14px; color: var(--assistant-muted, #698778); margin-bottom: 10px; }
input:not([type="checkbox"]) { width: 100%; min-width: 0; min-height: 44px; padding: 8px; font: inherit; background: var(--assistant-soft-surface, #f0f4ee); color: inherit; border: 0; border-radius: 12px; }
.search-field input { padding-inline: 0; }
.check { display: flex; align-items: center; gap: 10px; min-height: 44px; font-size: 0.85em; color: var(--assistant-muted, #698778); cursor: pointer; }
.check input { width: 18px; height: 18px; accent-color: var(--assistant-accent, #278c62); }
.history-item { margin-top: 10px; padding: 4px; border-radius: 18px; background: var(--assistant-surface, #fff); box-shadow: 0 3px 14px rgb(30 61 45 / 4%); transition: background-color 200ms ease; }
.selected { background: var(--assistant-accent-soft, #e5f4e9); }
.select, .result { width: 100%; text-align: start; display: flex; flex-direction: column; gap: 4px; }
.select strong { max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-weight: 600; }
time, .select span { font-size: 0.8em; color: var(--assistant-muted, #698778); }
.item-actions { display: flex; justify-content: flex-end; gap: 2px; }
form { padding: 8px; }
form > span { display: block; margin-top: 4px; font-size: 0.8em; color: var(--assistant-muted, #698778); }
.result span { overflow-wrap: anywhere; font-size: 0.85em; color: var(--assistant-muted, #698778); }
@media (hover: hover) {
    button:hover:not(:disabled) { background: var(--assistant-accent-soft, #e5f4e9); color: var(--assistant-accent, #278c62); }
    .delete-action:hover:not(:disabled) { color: #b75f54; background: rgb(183 95 84 / 10%); }
}
button:focus-visible, input:focus-visible { outline: 2px solid var(--assistant-accent, #278c62); outline-offset: 2px; }
@media (prefers-reduced-motion: reduce) {
    button, .history-item { transition: none; }
    button:active:not(:disabled) { transform: none; }
}
</style>
