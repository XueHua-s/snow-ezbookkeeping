<script setup lang="ts">
import { mdiArrowRight, mdiStop } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
defineProps<{ requesting: boolean; canRequestNextPage: boolean }>();
const emit = defineEmits<{ next: []; cancel: [] }>();
const { tt } = useI18n();
</script>

<template>
    <div class="assistant-session-actions" v-if="requesting || canRequestNextPage">
        <button v-if="canRequestNextPage" type="button" :aria-label="tt('Next Page')" :title="tt('Next Page')" @click="emit('next')"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiArrowRight" /></svg></button>
        <button v-if="requesting" type="button" :aria-label="tt('Cancel')" :title="tt('Cancel')" @click="emit('cancel')"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiStop" /></svg></button>
    </div>
</template>
<style scoped>
.assistant-session-actions { display: flex; flex-wrap: wrap; gap: 8px; padding: 8px 18px; }
button { display: grid; place-items: center; width: 44px; height: 44px; padding: 10px; border: none; border-radius: 14px; color: var(--assistant-action, #278c62); background: var(--assistant-tint, #e8f3e9); cursor: pointer; transition: background 180ms ease, transform 180ms ease; }
button:hover { background: var(--assistant-border, #dce9df); }
button:active { transform: scale(.94); }
button:focus-visible { outline: 2px solid var(--assistant-action, #278c62); outline-offset: 2px; }
svg { width: 20px; height: 20px; fill: currentColor; }
@media (prefers-reduced-motion: reduce) { button { transition: none; } }
</style>
