<template>
    <markdown-renderer
        class="assistant-markdown"
        mode="streaming"
        :markdown="displayContent"
        :streamdown="streamdownOptions"
        :parserOptions="parserOptions"
        :components="customComponents"
    />
</template>

<script setup lang="ts">
import { computed, h, type FunctionalComponent } from 'vue';

import type { ParserOptions } from '@markdown-next/parser';
import { MarkdownRenderer } from '@markdown-next/vue';
import type { MarkdownComponents } from '@markdown-next/vue';

interface AssistantMarkdownContentProps {
    readonly content: string;
    readonly thinking?: string;
}

const props = defineProps<AssistantMarkdownContentProps>();

const displayContent = computed<string>(() => {
    const content = props.content || '';
    const thinking = props.thinking || '';

    if (thinking && content) {
        return `<think>\n${thinking}\n</think>\n\n${content}`;
    }

    if (thinking) {
        return `<think>\n${thinking}\n</think>`;
    }

    return content;
});

const parserOptions: ParserOptions = {
    extendedGrammar: ['gfm'],
    customTags: ['think']
};

const streamdownOptions = {
    parseIncompleteMarkdown: true
};

const ThinkBlock: FunctionalComponent = (_props, { slots }) => {
    return h('div', {
        class: 'assistant-think-block'
    }, slots['default'] ? slots['default']() : []);
};

const customComponents: MarkdownComponents = {
    think: ThinkBlock
};
</script>

<style scoped>
.assistant-markdown :deep(*) {
    box-sizing: border-box;
}

.assistant-markdown :deep(p) {
    margin: 0 0 10px;
    line-height: 1.55;
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
