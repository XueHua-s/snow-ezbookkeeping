<script setup lang="ts">
import { computed, shallowRef } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantMemoryDTO } from '@/models/ai_assistant_conversation.ts';
defineProps<{ items: readonly AssistantMemoryDTO[]; enabled: boolean; busy?: boolean; error?: string; byteLimit: number }>();
const emit = defineEmits<{ enabled: [value: boolean]; save: [item: AssistantMemoryDTO | undefined, text: string]; forget: [item: AssistantMemoryDTO]; forgetAll: []; source: [conversationId: string, messageId: string]; retry: []; close: [] }>();
const { tt } = useI18n();
const editing = shallowRef<AssistantMemoryDTO>();
const editorOpen = shallowRef(false);
const text = shallowRef('');
const bytes = computed(() => new TextEncoder().encode(text.value).length);
defineExpose({ finishEdit: () => { editorOpen.value = false; editing.value = undefined; text.value = ''; } });
function edit(item?: AssistantMemoryDTO): void { editing.value = item; text.value = item?.text ?? ''; editorOpen.value = true; }
</script>
<template>
    <section class="memories" @keydown.esc="editorOpen ? editorOpen = false : emit('close')">
        <header><h2>{{ tt('Long-term memory') }}</h2><button type="button" @click="emit('close')">{{ tt('Close') }}</button></header>
        <p>{{ tt('Memory helps future conversations use your preferences. Conversation history is saved separately.') }}</p>
        <label class="memory-setting-toggle"><input type="checkbox" :checked="enabled" :disabled="busy" @change="emit('enabled', ($event.target as HTMLInputElement).checked)" />{{ tt('Use and automatically organize long-term memory') }}</label>
        <p v-if="!enabled">{{ tt('Memory is off. Saved memories remain available to manage.') }}</p>
        <p v-if="error" role="alert">{{ tt(error) }} <button type="button" @click="emit('retry')">{{ tt('Retry') }}</button></p>
        <p v-if="busy" role="status">{{ tt('Confirming changes') }}</p>
        <button type="button" :disabled="busy" @click="edit()">{{ tt('Add memory') }}</button>
        <form v-if="editorOpen" @submit.prevent="emit('save', editing, text)"><label>{{ tt('Memory text') }}<textarea v-model="text" rows="4" /></label><p>{{ bytes }} / {{ byteLimit }} {{ tt('bytes') }}</p><p v-if="bytes > byteLimit" role="alert">{{ tt('Input exceeds the limit. Your text has been kept.') }}</p><button type="submit" :disabled="busy || !text.trim() || bytes > byteLimit">{{ tt('Save') }}</button><button type="button" @click="editorOpen = false">{{ tt('Cancel') }}</button></form>
        <p v-if="!items.length && !busy && !error">{{ tt('No saved memories') }}</p>
        <article v-for="item in items" :key="item.id"><p class="memory-text">{{ item.text }}</p><p>{{ tt(item.type) }} · {{ tt(item.status) }} · <time :datetime="item.updatedAt">{{ new Date(item.updatedAt).toLocaleString() }}</time></p><div><button v-for="source in item.sources" :key="source.messageId" type="button" @click="emit('source', source.conversationId, source.messageId)">{{ tt('View source') }}</button></div><button type="button" :disabled="busy" @click="edit(item)">{{ tt(item.status === 'candidate' ? 'Review and confirm' : 'Edit') }}</button><button type="button" :disabled="busy" @click="emit('forget', item)">{{ tt('Forget') }}</button></article>
        <p>{{ tt('Forgetting prevents future memory use and extraction from old sources. Original conversations remain until deleted.') }}</p>
        <button v-if="items.length" type="button" :disabled="busy" @click="emit('forgetAll')">{{ tt('Forget all memories') }}</button>
    </section>
</template>
<style scoped>
* { box-sizing: border-box; }
.memories { padding: 24px; max-width: 720px; margin-inline: auto; line-height: 1.65; } header { display: flex; justify-content: space-between; align-items: center; } h2 { font-size: 1.35em; } article { margin-block: 16px; padding-block: 12px; border-block-start: 1px solid rgb(127 127 127 / 50%); } .memory-text { white-space: pre-wrap; overflow-wrap: anywhere; } button { width: auto; min-height: 44px; padding: 8px 12px; font: inherit; background: transparent; color: inherit; border: 1px solid rgb(127 127 127 / 90%); border-radius: 8px; margin-inline-end: 8px; cursor: pointer; } textarea { width: 100%; color: inherit; font: inherit; background: transparent; border: 1px solid currentColor; border-radius: 8px; padding: 8px; } .memory-setting-toggle { display: flex; align-items: center; min-height: 44px; gap: 10px; } .memory-setting-toggle input { flex: 0 0 18px; width: 18px; height: 18px; min-height: 18px; margin: 0; } button:hover { background: rgb(127 127 127 / 10%); } button:focus-visible, textarea:focus-visible, input:focus-visible { outline: 2px solid currentColor; outline-offset: 2px; } button:disabled { opacity: .55; cursor: default; }</style>
