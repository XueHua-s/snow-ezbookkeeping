<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
const props = defineProps<{
    used: number;
    reserved: number;
    limit: number;
    reason?: string;
    canContinue?: boolean;
}>();
const emit = defineEmits<{ continue: []; export: []; manage: [] }>();
const { tt } = useI18n();
const visible = computed(() => !!props.reason || props.used + props.reserved >= props.limit * 0.8);
</script>
<template>
    <aside v-if="visible" class="capacity" role="status">
        <strong>{{ tt(reason || 'Approaching conversation capacity') }}</strong>
        <p>
            {{ tt('Used') }}: {{ used }} · {{ tt('Reserved for generation') }}: {{ reserved }} / {{ limit }}
            {{ tt('bytes') }}
        </p>
        <div>
            <button v-if="canContinue" type="button" @click="emit('continue')">
                {{ tt('Continue in a new conversation') }}</button
            ><button type="button" @click="emit('export')">{{ tt('Export') }}</button
            ><button type="button" @click="emit('manage')">{{ tt('Manage assistant data') }}</button>
        </div>
    </aside>
</template>
<style scoped>
.capacity {
    padding: 16px;
    border: 1px solid currentColor;
    border-radius: 12px;
    margin: 12px 16px;
    line-height: 1.6;
}
button {
    min-height: 44px;
    margin-inline-end: 8px;
    padding: 8px 12px;
    color: inherit;
    background: transparent;
    border: 1px solid currentColor;
    border-radius: 8px;
    font: inherit;
    cursor: pointer;
}
button:focus-visible {
    outline: 2px solid currentColor;
    outline-offset: 3px;
}
</style>
