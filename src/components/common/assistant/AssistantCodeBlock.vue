<script setup lang="ts">
import { computed } from 'vue';
import AssistantCopyButton from './AssistantCopyButton.vue';
const props = defineProps<{ node?: { children?: { type: string; value?: string }[] }; children?: unknown; className?: string }>();
// HAST text comes from the parser, never from rendered DOM or innerHTML.
const code = computed(() => props.node?.children?.filter(child => child.type === 'text').map(child => child.value ?? '').join('') ?? (typeof props.children === 'string' ? props.children : ''));
</script>
<template><code :class="className">{{ code }}</code><AssistantCopyButton v-if="code.includes('\n') || className?.includes('language-')" :text="code" code /></template>
