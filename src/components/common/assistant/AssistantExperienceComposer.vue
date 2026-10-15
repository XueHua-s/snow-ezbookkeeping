<script setup lang="ts">
import { computed, shallowRef, useTemplateRef, watch, nextTick } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantLimits, AssistantRunState } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{
    variant: 'desktop' | 'mobile';
    limits: Pick<AssistantLimits, 'inputCharacters' | 'inputBytes'>;
    generationState?: AssistantRunState;
    sending?: boolean;
    disabledReason?: string;
    editingLabel?: string;
}>();
const model = defineModel<string>({ required: true });
const emit = defineEmits<{ send: []; stop: []; cancelEdit: [] }>();
const { tt } = useI18n();
const input = useTemplateRef<HTMLTextAreaElement>('input');
const composing = shallowRef(false);
const count = computed(() => Array.from(model.value).length);
const bytes = computed(() => new TextEncoder().encode(model.value).length);
const overLimit = computed(() => count.value > props.limits.inputCharacters || bytes.value > props.limits.inputBytes);
const running = computed(() => !!props.generationState && ['queued', 'preparing', 'running', 'cancel_requested'].includes(props.generationState));
const canSend = computed(() => !!model.value.trim() && !overLimit.value && !props.disabledReason && !props.sending && !running.value);
function send(): void {
    if (!canSend.value || composing.value) return;
    emit('send');
    if (props.variant === 'desktop') input.value?.focus();
}
function keydown(event: KeyboardEvent): void {
    if (event.key === 'Escape' && props.editingLabel) { emit('cancelEdit'); return; }
    if (event.key !== 'Enter' || event.shiftKey || props.variant === 'mobile' || event.isComposing || composing.value || event.keyCode === 229) return;
    event.preventDefault();
    send();
}
watch(model, async () => {
    await nextTick();
    if (input.value) { input.value.style.height = 'auto'; input.value.style.height = `${Math.min(input.value.scrollHeight, 192)}px`; }
}, { immediate: true });
</script>
<template>
    <section class="experience-composer">
        <div v-if="editingLabel" class="editing"><strong>{{ editingLabel }}</strong><button type="button" @click="emit('cancelEdit')">{{ tt('Cancel') }}</button></div>
        <div class="composer-row">
            <textarea ref="input" v-model="model" rows="2" :aria-label="tt('Ask your personal finance question')" :placeholder="tt('Ask your personal finance question')" @keydown="keydown" @compositionstart="composing = true" @compositionend="composing = false" />
            <button v-if="running" type="button" class="primary" :disabled="generationState === 'cancel_requested'" @click="emit('stop')">{{ tt(generationState === 'cancel_requested' ? 'Stopping generation' : 'Stop generating') }}</button>
            <button v-else type="button" class="primary" :disabled="!canSend" @click="send">{{ tt(sending ? 'Confirming delivery' : editingLabel ? 'Edit and create branch' : 'Send') }}</button>
        </div>
        <div class="composer-notes">
            <span>{{ count }} / {{ limits.inputCharacters }}<template v-if="bytes >= limits.inputBytes * .8"> · {{ bytes }} / {{ limits.inputBytes }} {{ tt('bytes') }}</template></span>
            <span v-if="overLimit" role="alert">{{ tt('Input exceeds the limit. Your text has been kept.') }}</span>
            <span v-else-if="disabledReason" role="status">{{ tt(disabledReason) }}</span>
            <span v-else-if="running">{{ tt('You can write your next question while the answer is generating.') }}</span>
            <span v-else-if="variant === 'desktop'">{{ tt('Enter to send, Shift+Enter for a new line') }}</span>
        </div>
    </section>
</template>
<style scoped>
* { box-sizing: border-box; }

.experience-composer { flex: 0 0 auto; padding: 12px 16px max(12px, env(safe-area-inset-bottom)); background: var(--assistant-surface, transparent); }
.composer-row { display: flex; align-items: flex-end; gap: 8px; max-width: 840px; margin-inline: auto; border: 1px solid rgb(127 127 127 / 90%); border-radius: 18px; padding: 10px; }
textarea { resize: none; border: 0 !important; box-shadow: none !important; outline: none; min-width: 0; width: 100%; max-height: 192px; padding: 8px; background: transparent; color: inherit; font: inherit; line-height: 1.65; }
button { width: auto; min-width: 44px; min-height: 44px; border: 0; border-radius: 10px; padding: 8px 12px; font: inherit; color: inherit; background: rgb(127 127 127 / 12%); cursor: pointer; }
.primary { flex-shrink: 0; background: var(--assistant-action, var(--f7-theme-color, rgb(var(--v-theme-primary, 51, 65, 85)))); color: rgb(var(--v-theme-on-primary, 255, 255, 255)); }
button:hover { filter: brightness(.9); } button:active { transform: translateY(1px); } button:disabled { opacity: .55; cursor: default; } button:focus-visible { outline: 2px solid currentColor; outline-offset: 3px; }
.composer-row:focus-within { outline: 2px solid var(--f7-theme-color, rgb(var(--v-theme-primary, 51, 65, 85))); outline-offset: 2px; }
textarea:focus, textarea:focus-visible { outline: none; border: 0 !important; }
.composer-notes, .editing { max-width: 820px; margin: 6px auto 0; display: flex; flex-wrap: wrap; gap: 6px 16px; font-size: .8em; line-height: 1.5; }
.editing { align-items: center; justify-content: space-between; }
@media (max-width: 440px) { .primary { max-width: 100px; } }
</style>
