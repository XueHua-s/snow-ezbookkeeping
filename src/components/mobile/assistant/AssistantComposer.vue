<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
defineProps<{ disabled: boolean; requesting: boolean; canSend: boolean }>();
const model = defineModel<string>({ required: true });
const emit = defineEmits<{ send: [] }>();
const { tt } = useI18n();
</script>

<template>
    <f7-block class="assistant-input-block">
        <div class="assistant-input-row">
            <f7-input
                class="assistant-input"
                type="textarea"
                resizable
                :disabled="disabled"
                :placeholder="tt('Ask your personal finance question')"
                v-model:value="model">
            </f7-input>
            <f7-button type="button" class="assistant-send-button"
                       fill
                       :disabled="!canSend"
                       @click="emit('send')">
                {{ tt('Send') }}
            </f7-button>
        </div>
        <div class="assistant-requesting margin-top-half" v-if="requesting">
            <f7-preloader size="16"></f7-preloader>
        </div>
    </f7-block>
</template>

<style scoped>
.assistant-input-block {
    margin-top: 0;
    margin-bottom: 0;
    padding-top: 10px;
    padding-bottom: calc(env(safe-area-inset-bottom) + 8px);
    position: sticky;
    bottom: 0;
    z-index: 8;
    background: var(--f7-page-bg-color);
    border-top: 1px solid var(--f7-page-master-border-color);
}

.assistant-input-row {
    display: flex;
    align-items: flex-end;
    column-gap: 8px;
}

.assistant-input {
    flex: 1;
    min-width: 0;
    width: auto;
    background: var(--f7-card-bg-color);
    border: 1px solid var(--f7-page-master-border-color);
    border-radius: 12px;
    padding-inline: 8px;
}

.assistant-input :deep(textarea) {
    width: 100%;
    min-height: 42px;
}

.assistant-send-button {
    flex: 0 0 auto;
    width: auto;
    height: 42px;
    min-width: 72px;
}

.assistant-requesting {
    display: flex;
    justify-content: center;
}
</style>
