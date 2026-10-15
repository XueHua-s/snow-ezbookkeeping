<script setup lang="ts">
import { onBeforeUnmount, shallowRef, watch } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantConversationDTO, AssistantSearchMatch } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{ items: readonly AssistantConversationDTO[]; matches?: readonly AssistantSearchMatch[]; selectedId?: string; loading?: boolean; error?: string; hasMore?: boolean; archived?: boolean; titleLimit: number }>();
const emit = defineEmits<{ select: [id: string, messageId?: string, versionId?: string]; new: []; search: [query: string, includeVersions: boolean]; archived: [value: boolean]; rename: [item: AssistantConversationDTO, title: string]; archive: [item: AssistantConversationDTO]; delete: [item: AssistantConversationDTO]; export: [item: AssistantConversationDTO]; retry: []; loadMore: []; close: [] }>();
const { tt } = useI18n();
const query = shallowRef('');
const includeVersions = shallowRef(false);
const editingId = shallowRef('');
const title = shallowRef('');
let debounce: ReturnType<typeof setTimeout> | undefined;
watch([query, includeVersions], () => { clearTimeout(debounce); debounce = setTimeout(() => emit('search', query.value, includeVersions.value), 300); });
onBeforeUnmount(() => clearTimeout(debounce));
defineExpose({ finishEdit: () => { editingId.value = ''; title.value = ''; } });
function rename(item: AssistantConversationDTO): void { if (title.value.trim() && Array.from(title.value).length <= props.titleLimit) emit('rename', item, title.value); }
</script>
<template>
    <nav class="history" :aria-label="tt('Conversation history')" @keydown.esc="editingId ? editingId = '' : emit('close')">
        <header><strong>{{ tt('Conversation history') }}</strong><button type="button" @click="emit('close')">{{ tt('Close') }}</button></header>
        <button class="new" type="button" @click="emit('new')">{{ tt('New conversation') }}</button>
        <label>{{ tt('Search history') }}<input v-model="query" type="search" /></label>
        <label class="check"><input type="checkbox" :checked="archived" @change="emit('archived', ($event.target as HTMLInputElement).checked)" />{{ tt('Include archived') }}</label>
        <label class="check"><input v-model="includeVersions" type="checkbox" />{{ tt('Include older answer versions') }}</label>
        <p v-if="loading" role="status">{{ tt('Loading') }}</p>
        <p v-if="error" role="alert">{{ tt(error) }} <button type="button" @click="emit('retry')">{{ tt('Retry') }}</button></p>
        <template v-if="query"><p v-if="!loading && !error && !matches?.length">{{ tt('No matching conversations') }}</p><button v-for="match in matches" :key="match.messageId + (match.versionId || '')" type="button" class="result" @click="emit('select', match.conversationId, match.messageId, match.versionId)"><strong>{{ match.title }}</strong><span>{{ match.snippet }}</span></button></template>
        <template v-else><p v-if="!items.length && !loading && !error">{{ tt('No conversations yet') }}</p>
            <section v-for="item in items" :key="item.id" class="history-item" :class="{ selected: selectedId === item.id }">
                <button type="button" class="select" :title="item.title" :aria-current="selectedId === item.id ? 'page' : undefined" @click="emit('select', item.id)"><strong>{{ item.title }}</strong><time :datetime="item.activityAt">{{ new Date(item.activityAt).toLocaleDateString() }}</time><span v-if="item.activeRunId">{{ tt('Generating answer') }}</span><span v-else-if="item.lastRunState === 'failed'">{{ tt('Assistant request failed') }}</span></button>
                <form v-if="editingId === item.id" @submit.prevent="rename(item)"><label>{{ tt('Conversation title') }}<input v-model="title" /></label><span>{{ Array.from(title).length }} / {{ titleLimit }}</span><button type="submit" :disabled="!title.trim() || Array.from(title).length > titleLimit">{{ tt('Save') }}</button><button type="button" @click="editingId = ''">{{ tt('Cancel') }}</button></form>
                <details v-else><summary>{{ tt('More') }}</summary><div class="item-actions"><button type="button" @click="editingId = item.id; title = item.title">{{ tt('Rename') }}</button><button type="button" @click="emit('archive', item)">{{ tt(item.archived ? 'Unarchive' : 'Archive') }}</button><button type="button" @click="emit('export', item)">{{ tt('Export') }}</button><button type="button" @click="emit('delete', item)">{{ tt('Delete') }}</button></div></details>
            </section>
        </template>
        <button v-if="hasMore" type="button" :disabled="loading" @click="emit('loadMore')">{{ tt('Load more') }}</button>
    </nav>
</template>
<style scoped>
* { box-sizing: border-box; }
.history { width: 100%; padding: 16px; overflow-y: auto; height: 100%; line-height: 1.5; } header { display: flex; align-items: center; justify-content: space-between; } button, summary { width: auto; min-height: 44px; padding: 8px; color: inherit; font: inherit; background: transparent; border: 1px solid transparent; border-radius: 8px; cursor: pointer; } button:hover, summary:hover, .selected { background: rgb(127 127 127 / 12%); } input { width: 100%; min-height: 44px; padding: 8px; font: inherit; background: transparent; color: inherit; border: 1px solid rgb(127 127 127 / 90%); border-radius: 8px; } .check { display: flex; align-items: center; gap: 8px; font-size: .85em; } .check input { width: 18px; } .new { width: 100%; margin-block: 8px; border-color: currentColor; } .history-item { margin-top: 8px; border-radius: 8px; } .select, .result { width: 100%; text-align: start; display: flex; flex-direction: column; } .select strong { max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; } time { font-size: .8em; } .item-actions { display: flex; flex-wrap: wrap; } button:focus-visible, summary:focus-visible, input:focus-visible { outline: 2px solid currentColor; outline-offset: 2px; }</style>
