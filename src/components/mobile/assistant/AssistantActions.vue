<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
defineProps<{ requesting: boolean; canRequestNextPage: boolean }>();
const emit = defineEmits<{ next: []; cancel: [] }>();
const { tt } = useI18n();
</script>

<template>
    <f7-block class="assistant-session-actions" v-if="requesting || canRequestNextPage">
        <f7-button type="button" outline v-if="canRequestNextPage" @click="emit('next')">{{ tt('Next Page') }}</f7-button>
        <f7-button type="button" v-if="requesting" @click="emit('cancel')">{{ tt('Cancel') }}</f7-button>
    </f7-block>
</template>

<style scoped>
.assistant-session-actions { display: flex; flex-wrap: wrap; gap: 8px; padding: 8px 18px; }
</style>
