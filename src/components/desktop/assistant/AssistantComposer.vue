<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
import { mdiSend } from '@mdi/js';
const props = defineProps<{ disabled: boolean; requesting: boolean; canSend: boolean }>();
const model = defineModel<string>({ required: true });
const emit = defineEmits<{ send: [] }>();
const { tt } = useI18n();

function sendOnEnter(event: KeyboardEvent): void {
    // Enter confirms IME candidates before it becomes a submission gesture.
    // WebKit can clear isComposing on the final key while retaining code 229.
    if (event.isComposing || event.keyCode === 229) {
        return;
    }
    event.preventDefault();
    if (!props.disabled && props.canSend && !props.requesting) {
        emit('send');
    }
}
</script>

<template>
    <div class="assistant-composer">
        <v-textarea
            auto-grow
            rows="2"
            max-rows="6"
            class="assistant-input"
            variant="outlined"
            hide-details
            :disabled="disabled"
            :placeholder="tt('Ask your personal finance question')"
            v-model="model"
            @keydown.enter.exact="sendOnEnter"
        />
        <v-btn color="primary"
               class="assistant-send-button"
               :disabled="!canSend"
               :loading="requesting"
               @click="emit('send')">
            <v-icon :icon="mdiSend" size="20" class="me-2" />
            {{ tt('Send') }}
        </v-btn>
    </div>
</template>

<style scoped>
.assistant-composer {
    display: flex;
    align-items: flex-end;
    column-gap: 10px;
}

.assistant-input {
    flex: 1;
    min-width: 0;
}

.assistant-send-button {
    height: 44px;
    min-width: 106px;
}

@media (max-width: 960px) { .assistant-composer { flex-direction: column; align-items: stretch; } .assistant-send-button { width: 100%; } }
</style>
