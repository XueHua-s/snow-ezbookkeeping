<script setup lang="ts">
import { computed, useId, shallowRef, useTemplateRef, watch, nextTick } from 'vue';
import { mdiArrowUp, mdiStop, mdiLoading } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantLimits, AssistantRunState } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{
    variant: 'desktop' | 'mobile';
    limits: Pick<AssistantLimits, 'inputCharacters' | 'inputBytes'>;
    generationState?: AssistantRunState;
    sending?: boolean;
    disabledReason?: string;
}>();
const model = defineModel<string>({ required: true });
const emit = defineEmits<{ send: []; stop: [] }>();
const { tt } = useI18n();
const input = useTemplateRef<HTMLTextAreaElement>('input');
const descriptionId = useId();
const composing = shallowRef(false);
const count = computed(() => Array.from(model.value).length);
const bytes = computed(() => new TextEncoder().encode(model.value).length);
const overLimit = computed(() => count.value > props.limits.inputCharacters || bytes.value > props.limits.inputBytes);
const running = computed(
    () =>
        !!props.generationState &&
        ['queued', 'preparing', 'running', 'cancel_requested'].includes(props.generationState)
);
const actionLabel = computed(() => tt(running.value
    ? props.generationState === 'cancel_requested' ? 'Stopping generation' : 'Stop generating'
    : props.sending ? 'Confirming delivery' : 'Send'));
const waiting = computed(() => props.sending || props.generationState === 'cancel_requested');
const canSend = computed(
    () => !!model.value.trim() && !overLimit.value && !props.disabledReason && !props.sending && !running.value
);
function send(): void {
    if (!canSend.value || composing.value) return;
    emit('send');
    if (props.variant === 'desktop') input.value?.focus();
}
function keydown(event: KeyboardEvent): void {
    if (
        event.key !== 'Enter' ||
        event.shiftKey ||
        props.variant === 'mobile' ||
        event.isComposing ||
        composing.value ||
        event.keyCode === 229
    )
        return;
    event.preventDefault();
    send();
}
watch(
    model,
    async () => {
        await nextTick();
        if (input.value) {
            input.value.style.height = 'auto';
            input.value.style.height = `${Math.min(input.value.scrollHeight, 192)}px`;
        }
    },
    { immediate: true }
);
</script>
<template>
    <section class="experience-composer" :class="variant">
        <div class="composer-row">
            <textarea
                class="composer-input"
                ref="input"
                v-model="model"
                rows="2"
                name="assistant-question"
                autocomplete="off"
                :aria-describedby="descriptionId"
                :aria-invalid="overLimit || undefined"
                :aria-label="tt('Ask your personal finance question')"
                :placeholder="tt('Ask your personal finance question')"
                @keydown="keydown"
                @compositionstart="composing = true"
                @compositionend="composing = false"
            />
            <button
                type="button"
                class="composer-action primary"
                :class="{ stopping: running }"
                :disabled="running ? generationState === 'cancel_requested' : !canSend"
                :aria-label="actionLabel"
                :title="actionLabel"
                @click="running ? emit('stop') : send()"
            >
                <Transition name="composer-icon" mode="out-in">
                    <svg :key="waiting ? 'waiting' : running ? 'stop' : 'send'" viewBox="0 0 24 24" aria-hidden="true" :class="{ spinning: waiting }">
                        <path :d="waiting ? mdiLoading : running ? mdiStop : mdiArrowUp" />
                    </svg>
                </Transition>
            </button>
        </div>
        <div class="composer-notes" :id="descriptionId">
            <span v-if="count >= limits.inputCharacters * 0.8 || bytes >= limits.inputBytes * 0.8" class="input-count">
                {{ count }} / {{ limits.inputCharacters }}
                <template v-if="bytes >= limits.inputBytes * 0.8">
                    · {{ bytes }} / {{ limits.inputBytes }} {{ tt('bytes') }}
                </template>
            </span>
            <span v-if="overLimit" role="alert">{{ tt('Input exceeds the limit. Your text has been kept.') }}</span>
            <span v-else-if="disabledReason" role="status">{{ tt(disabledReason) }}</span>
            <span v-else-if="running">
                {{ tt('You can write your next question while the answer is generating.') }}
            </span>
            <span v-else-if="variant === 'desktop'">{{ tt('Enter to send, Shift+Enter for a new line') }}</span>
        </div>
    </section>
</template>
<style scoped>
.experience-composer,
.composer-row,
.composer-input,
.composer-action {
    box-sizing: border-box;
}

.experience-composer {
    flex: 0 0 auto;
    padding: 12px 16px max(12px, env(safe-area-inset-bottom));
    background: var(--assistant-surface, transparent);
}
.composer-row {
    display: flex;
    align-items: flex-end;
    gap: 8px;
    max-width: 840px;
    margin-inline: auto;
    border: 1px solid var(--assistant-border, #e5ece4);
    background: var(--assistant-card, #fff);
    box-shadow: 0 6px 24px rgb(28 64 42 / 6%);
    transition: border-color 180ms ease, box-shadow 180ms ease;
    border-radius: 24px;
    padding: 10px;
}
.composer-input {
    resize: none;
    border: 0 !important;
    box-shadow: none !important;
    outline: none;
    min-width: 0;
    width: 100%;
    max-height: 192px;
    padding: 8px;
    background: transparent;
    color: inherit;
    font: inherit;
    line-height: 1.65;
}
.composer-action {
    width: auto;
    min-width: 44px;
    min-height: 44px;
    border: 0;
    border-radius: 15px;
    padding: 10px;
    font: inherit;
    color: inherit;
    background: var(--assistant-tint, #e8f3e9);
    display: inline-grid;
    place-items: center;
    transition: background 160ms ease, transform 160ms ease, opacity 160ms ease;
    cursor: pointer;
}
.primary {
    flex-shrink: 0;
    --action-color: var(--assistant-action, #278c62);
    background: var(--action-color);
    color: var(--assistant-on-action, #fff);
}
.composer-action:hover:not(:disabled) {
    background: var(--assistant-border, #e5ece4);
}
.primary:hover:not(:disabled),
.primary:active:not(:disabled) {
    background: color-mix(in srgb, var(--action-color) 65%, black);
}
.composer-action:active {
    transform: scale(0.94);
}
.composer-action:disabled {
    opacity: 0.55;
    cursor: default;
}
.composer-action:focus-visible {
    outline: 2px solid var(--assistant-action, #278c62);
    outline-offset: 3px;
}
.composer-row:focus-within {
    border-color: var(--assistant-action, #278c62);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--assistant-action, #278c62) 12%, transparent), 0 6px 24px rgb(28 64 42 / 6%);
}
.composer-input:focus,
.composer-input:focus-visible {
    outline: none;
    border: 0 !important;
}
.composer-notes {
    max-width: 820px;
    margin: 6px auto 0;
    display: flex;
    flex-wrap: wrap;
    gap: 6px 16px;
    font-size: 0.8em;
    line-height: 1.5;
    color: var(--assistant-muted, #718277);
}
.composer-action svg { width: 24px; height: 24px; fill: currentColor; }
.composer-input::placeholder { color: var(--assistant-muted, #718277); opacity: 0.85; }
.composer-notes:empty { display: none; }
.input-count { margin-inline-start: auto; font-variant-numeric: tabular-nums; }
.composer-icon-enter-active, .composer-icon-leave-active { transition: opacity 150ms ease, transform 150ms ease; }
.composer-icon-enter-from, .composer-icon-leave-to { opacity: 0; transform: scale(0.7); }
.spinning { animation: composer-spin 1s linear infinite; }
@keyframes composer-spin { to { transform: rotate(360deg); } }

@media (prefers-reduced-motion: reduce) {
    .composer-row, .composer-action, .composer-icon-enter-active, .composer-icon-leave-active { transition: none; }
    .spinning { animation: none; }
}
@media (max-height: 440px) and (orientation: landscape) {
    .mobile.experience-composer {
        padding-block: 4px max(4px, env(safe-area-inset-bottom));
    }
    .mobile .composer-row {
        padding: 4px;
    }
    .mobile .composer-input {
        height: 44px;
        min-height: 44px;
        max-height: 44px;
        overflow-y: auto;
    }
    .mobile .composer-notes {
        margin-top: 2px;
    }
}
</style>
