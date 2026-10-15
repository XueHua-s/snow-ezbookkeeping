<template>
    <div class="assistant-markdown-container">
    <details v-if="thinking" class="assistant-process"><summary>{{ tt('Process summary') }}</summary><p>{{ thinking }}</p></details>
    <markdown-renderer
        class="assistant-markdown"
        mode="streaming"
        :markdown="content"
        :streamdown="streamdownOptions"
        :parserOptions="parserOptions"
        :components="customComponents"
    />
    </div>
</template>

<script setup lang="ts">
import { h, type FunctionalComponent, type VNodeChild } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import AssistantCodeBlock from '@/components/common/assistant/AssistantCodeBlock.vue';

import type { ParserOptions } from '@markdown-next/parser';
import { MarkdownRenderer } from '@markdown-next/vue';
import type { MarkdownComponents } from '@markdown-next/vue';

interface AssistantMarkdownContentProps {
    readonly content: string;
    readonly thinking?: string;
}

defineProps<AssistantMarkdownContentProps>();

const { tt } = useI18n();

const parserOptions: ParserOptions = {
    extendedGrammar: ['gfm'],
    customTags: []
};

const streamdownOptions = {
    parseIncompleteMarkdown: true
};

function safeLink(value: unknown): string | undefined {
    if (typeof value !== 'string' || /[\u0000-\u0020]/.test(value)) return undefined;
    try { const url = new URL(value); return ['https:', 'http:', 'mailto:'].includes(url.protocol) ? value : undefined; } catch { return undefined; }
}
const SafeLink: FunctionalComponent<{ href?: string; children?: VNodeChild }> = (props) => h(safeLink(props['href']) ? 'a' : 'span', safeLink(props['href']) ? { href: safeLink(props['href']), target: '_blank', rel: 'noopener noreferrer' } : {}, [props['children']]);
const SafeImage: FunctionalComponent<{ src?: string; alt?: string }> = (props) => h(safeLink(props['src']) ? 'a' : 'span', safeLink(props['src']) ? { href: safeLink(props['src']), target: '_blank', rel: 'noopener noreferrer' } : {}, String(props['alt'] || tt('External image')));
const customComponents: MarkdownComponents = { a: SafeLink, img: SafeImage, code: AssistantCodeBlock };
</script>

<style scoped>
.assistant-markdown { overflow-wrap: anywhere; min-width: 0; color: inherit !important; background: transparent !important; font-family: inherit !important; font-size: inherit !important; }
.assistant-markdown :deep(th), .assistant-markdown :deep(pre) { background: rgb(127 127 127 / 12%) !important; color: inherit !important; }
.assistant-markdown :deep(blockquote), .assistant-markdown :deep(h6) { color: inherit !important; }
.assistant-markdown :deep(a) { color: inherit; text-decoration: underline; }
.assistant-markdown :deep(table) { display: block; overflow-x: auto; max-width: 100%; font-variant-numeric: tabular-nums; }
.assistant-process { margin-block-end: 12px; font-size: .85em; }
.assistant-process summary { cursor: pointer; min-height: 44px; padding-block: 10px; }
.assistant-markdown :deep(*) {
    box-sizing: border-box;
}

.assistant-markdown :deep(p) {
    margin: 0 0 10px;
    line-height: 1.7;
}

.assistant-markdown :deep(p:last-child) {
    margin-bottom: 0;
}

.assistant-markdown :deep(ul),
.assistant-markdown :deep(ol) {
    margin: 0 0 10px;
    padding-inline-start: 20px;
}

.assistant-markdown :deep(li + li) {
    margin-top: 4px;
}

.assistant-markdown :deep(code) {
    color: inherit !important;
    padding: 1px 4px;
    border-radius: 4px;
    font-size: 0.9em;
    background: rgba(127, 127, 127, 0.14);
}

.assistant-markdown :deep(pre) {
    margin: 0 0 10px;
    padding: 10px;
    border-radius: 8px;
    overflow-x: auto;
    background: rgba(127, 127, 127, 0.12);
}

.assistant-markdown :deep(pre code) {
    padding: 0;
    background: transparent;
}

.assistant-markdown :deep(.assistant-think-block) {
    margin: 0 0 12px;
    padding: 10px 12px;
    border-left: 3px solid rgba(127, 127, 127, 0.58);
    border-radius: 4px;
    background: rgba(127, 127, 127, 0.12);
    color: inherit;
    opacity: 0.85;
}

.assistant-markdown :deep(.assistant-think-block p:last-child) {
    margin-bottom: 0;
}
</style>
