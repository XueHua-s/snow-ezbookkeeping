<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
defineProps<{ requesting: boolean; canRequestNextPage: boolean }>();
const emit = defineEmits<{ next: []; cancel: [] }>();
const { tt } = useI18n();
</script>

<template>
    <div class="assistant-session-actions" v-if="requesting || canRequestNextPage">
        <v-btn variant="tonal" color="secondary" v-if="canRequestNextPage" @click="emit('next')">{{ tt('Next Page') }}</v-btn>
        <v-btn variant="text" v-if="requesting" @click="emit('cancel')">{{ tt('Cancel') }}</v-btn>
    </div>
</template>

<style scoped>
.assistant-session-actions { display: flex; flex-wrap: wrap; gap: 8px; padding: 8px 18px; }
</style>
