<script setup lang="ts">
import { useTemplateRef, watch, nextTick } from 'vue';
import { mdiArrowDown, mdiArrowTopRight, mdiAutoFix } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantMessageDTO, AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';
import {
    useAssistantAutoScroll,
    type AssistantReadingAnchor
} from '@/views/base/assistant/useAssistantAutoScroll.ts';
import { assistantStateLabels, assistantReasonLabels } from './assistantPresentation.ts';
import AssistantMarkdownContent from '@/components/common/AssistantMarkdownContent.vue';
import AssistantExperienceActions from './AssistantExperienceActions.vue';
import AssistantExperienceEvidence from './AssistantExperienceEvidence.vue';
const props = defineProps<{
    messages: readonly AssistantMessageDTO[];
    latestMessageId?: string;
    hasOlder?: boolean;
    loadingOlder?: boolean;
    olderError?: string;
    anchor?: AssistantReadingAnchor;
    busy?: boolean;
    examples?: readonly { readonly text: string; readonly icon: string }[];
}>();
const emit = defineEmits<{
    loadOlder: [];
    latest: [];
    openReference: [id: string, answer: AssistantRunSnapshot];
    retry: [message: AssistantMessageDTO];
    regenerate: [message: AssistantMessageDTO];
    edit: [message: AssistantMessageDTO];
    branch: [message: AssistantMessageDTO];
    selectVersion: [message: AssistantMessageDTO, id: string];
    loadVersions: [message: AssistantMessageDTO];
    nextPage: [message: AssistantMessageDTO];
    example: [text: string];
    anchorChange: [anchor: AssistantReadingAnchor];
}>();
const { tt } = useI18n();
const panel = useTemplateRef<HTMLElement>('panel');
const content = useTemplateRef<HTMLElement>('content');
const scrolling = useAssistantAutoScroll(panel, content);
watch(
    () => props.anchor,
    async (anchor, _previous, onCleanup) => {
        let current = true;
        onCleanup(() => { current = false; });
        await nextTick();
        if (current) scrolling.restore(anchor);
    },
    { immediate: true }
);
function isLatest(message: AssistantMessageDTO, index: number): boolean {
    return props.latestMessageId === undefined
        ? index === props.messages.length - 1
        : message.id === props.latestMessageId;
}
function saveAnchor(): void {
    const anchor = scrolling.capture();
    if (anchor) emit('anchorChange', anchor);
}
defineExpose({
    scrollToLatest: scrolling.scrollToLatest,
    captureAnchor: scrolling.capture,
    restoreAnchor: scrolling.restore
});
</script>
<template>
    <section class="message-region">
        <div ref="panel" class="experience-messages" @scroll.passive="saveAnchor">
            <div ref="content" class="message-column">
                <button v-if="hasOlder" type="button" :disabled="loadingOlder" @click="emit('loadOlder')">
                    {{ tt(loadingOlder ? 'Loading' : 'Load older messages') }}
                </button>
                <p v-if="olderError" role="alert">
                    {{ tt(olderError) }}
                    <button type="button" @click="emit('loadOlder')">{{ tt('Retry') }}</button>
                </p>
                <div v-if="!messages.length" class="welcome">
                    <div class="welcome-tag"><span class="welcome-mark" aria-hidden="true"><svg viewBox="0 0 24 24"><path :d="mdiAutoFix" /></svg></span>{{ tt('Your personal finance companion') }}</div>
                    <h2>{{ tt('How can I help with your finances?') }}</h2>
                    <p class="welcome-intro">{{ tt('Understand your spending, explore your balances, and plan what comes next.') }}</p>
                    <div class="examples">
                        <button
                            v-for="example in examples"
                            :key="example.text"
                            type="button"
                            @click="emit('example', example.text)"
                        >
                            <span class="example-icon" aria-hidden="true"><svg viewBox="0 0 24 24"><path :d="example.icon" /></svg></span>
                            <span>{{ example.text }}</span>
                            <svg class="example-arrow" viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiArrowTopRight" /></svg>
                        </button>
                    </div>
                    <p class="welcome-disclosure">{{ tt('Conversations are saved on this server and sent to the configured model service when generating.') }}</p>
                </div>
                <section
                    v-for="(message, index) in messages"
                    :key="message.id"
                    :data-message-id="message.id"
                    class="turn"
                >
                    <div class="question">
                        <div class="question-text">{{ message.content }}</div>
                        <AssistantExperienceActions
                            :content="message.content"
                            edit
                            :disabled-reason="
                                busy ? 'An answer is still being prepared. You can wait for it or stop it to send a new message.' : undefined
                            "
                            @edit="emit('edit', message)"
                        />
                    </div>
                    <article
                        v-if="message.selectedAnswer"
                        :data-message-id="message.selectedAnswer.assistantMessageId"
                        class="answer"
                    >
                        <header class="answer-heading">
                            <span class="answer-avatar" aria-hidden="true"><svg viewBox="0 0 24 24"><path :d="mdiAutoFix" /></svg></span>
                            <span>{{ tt('AI Assistant') }}</span>
                            <span v-if="message.selectedAnswer.state === 'completed'" class="answer-complete">{{ tt('Answer completed') }}</span>
                        </header>
                        <AssistantMarkdownContent
                            v-if="message.selectedAnswer.content"
                            :content="message.selectedAnswer.content"
                            :thinking="message.selectedAnswer.processSummary"
                        />
                        <p v-if="message.selectedAnswer.state !== 'completed'" class="run-status" role="status" aria-live="polite">
                            {{ tt(assistantStateLabels[message.selectedAnswer.state]) }}
                        </p>
                        <p v-if="message.selectedAnswer.reasonCode" class="run-error" role="alert">
                            {{ tt(assistantReasonLabels[message.selectedAnswer.reasonCode]) }}
                        </p>
                        <p
                            v-if="
                                ['failed', 'incomplete', 'interrupted', 'canceled'].includes(
                                    message.selectedAnswer.state
                                )
                            "
                        >
                            {{ tt('This answer was not completed. Partial text may be unreliable.') }}
                        </p>
                        <AssistantExperienceEvidence
                            :answer="message.selectedAnswer"
                            :can-page="!busy"
                            :page-label="
                                !isLatest(message, index)
                                    ? tt('Continue this query in a new conversation')
                                    : undefined
                            "
                            @open-reference="emit('openReference', $event, message.selectedAnswer)"
                            @next-page="emit('nextPage', message)"
                        />
                        <AssistantExperienceActions
                            :content="message.selectedAnswer.content"
                            :streaming="
                                ['queued', 'preparing', 'running', 'cancel_requested'].includes(
                                    message.selectedAnswer.state
                                )
                            "
                            :retry="
                                isLatest(message, index) &&
                                ['failed', 'incomplete', 'interrupted', 'canceled'].includes(
                                    message.selectedAnswer.state
                                )
                            "
                            :regenerate="
                                isLatest(message, index) && message.selectedAnswer.state === 'completed'
                            "
                            :branch="!isLatest(message, index)"
                            :versions="message.versions"
                            :selected-version-id="message.selectedAnswer.assistantMessageId"
                            :context-version-id="message.selectedAssistantMessageId"
                            :has-more-versions="message.hasMoreVersions"
                            :disabled-reason="
                                busy ? 'An answer is still being prepared. You can wait for it or stop it to send a new message.' : undefined
                            "
                            @retry="emit('retry', message)"
                            @regenerate="emit('regenerate', message)"
                            @branch="emit('branch', message)"
                            @select-version="emit('selectVersion', message, $event)"
                            @load-versions="emit('loadVersions', message)"
                        />
                    </article>
                </section>
            </div>
        </div>
        <button
            v-if="!scrolling.following.value"
            type="button"
            class="latest"
            :aria-label="tt(scrolling.unread.value ? 'New content · Back to latest' : 'Back to latest')"
            :title="tt(scrolling.unread.value ? 'New content · Back to latest' : 'Back to latest')"
            @click="
                emit('latest');
                scrolling.scrollToLatest();
            "
        >
<svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiArrowDown" /></svg>
            <span v-if="scrolling.unread.value" class="unread-dot" aria-hidden="true" />
        </button>
    </section>
</template>
<style scoped>
* {
    box-sizing: border-box;
}
.message-region {
    min-height: 0;
    flex: 1;
    position: relative;
    display: flex;
    flex-direction: column;
}
.experience-messages {
    flex: 1;
    overflow-y: auto;
    min-height: 0;
    width: 100%;
    padding: 24px 16px;
    overscroll-behavior: contain;
}
.message-column {
    max-width: 800px;
    margin-inline: auto;
    font-size: 1rem;
    line-height: 1.7;
}
.turn {
    margin-bottom: 24px;
}
.question {
    max-width: 85%;
    width: fit-content;
    margin-inline-start: auto;
    margin-bottom: 24px;
}
.question-text {
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    background: var(--assistant-tint, #e8f3e9);
    padding: 12px 16px;
    border-radius: 20px 20px 6px 20px;
}
.answer {
    min-width: 0;
}
.run-status {
    font-size: 0.8em;
    margin: 8px 0;
}
button {
    min-height: 44px;
    padding: 8px 12px;
    background: transparent;
    color: inherit;
    font: inherit;
    border: 1px solid var(--assistant-border, #e5ece4);
    border-radius: 16px;
    cursor: pointer;
    transition: background 180ms ease, box-shadow 180ms ease, transform 180ms ease;
}
button:focus-visible {
    outline: 2px solid var(--assistant-action, #278c62);
    outline-offset: 2px;
}
.welcome { padding: 36px 12px 28px; max-width: 620px; margin-inline: auto; }
.welcome-tag { display: inline-flex; align-items: center; gap: 7px; padding: 7px 12px; border-radius: 24px; background: var(--assistant-tint, #e8f3e9); color: var(--assistant-action, #278c62); font-size: .75rem; font-weight: 600; }
.welcome-mark { display: grid; place-items: center; }
.welcome-mark svg { width: 17px; height: 17px; }
.answer-heading { display: flex; align-items: center; flex-wrap: wrap; gap: 9px; margin-bottom: 16px; font-size: .8rem; font-weight: 600; }
.answer-avatar { display: grid; place-items: center; flex-shrink: 0; width: 30px; height: 30px; border-radius: 11px; color: #fff; background: var(--assistant-action, #278c62); }
.answer-avatar svg { width: 18px; height: 18px; }
.answer-complete { color: var(--assistant-muted, #718277); font-weight: 400; font-size: .7rem; }
.welcome h2 { font-size: clamp(1.75rem, 4vw, 2.4rem); font-weight: 700; line-height: 1.35; letter-spacing: -.035em; margin: 24px 0 12px; text-wrap: balance; }
.welcome p { font-size: .875rem; color: var(--assistant-muted, #718277); line-height: 1.8; }
.welcome .welcome-disclosure { margin: 24px 0 0; font-size: .7rem; line-height: 1.7; }
.examples { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; margin-top: 28px; }
.examples button { position: relative; display: flex; flex-direction: column; align-items: flex-start; gap: 12px; text-align: start; padding: 16px; background: var(--assistant-card, #fff); box-shadow: 0 5px 16px #23452d05; font-size: .875rem; line-height: 1.55; }
.example-icon { width: 34px; height: 34px; display: grid; place-items: center; border-radius: 12px; background: var(--assistant-tint, #e8f3e9); color: var(--assistant-action, #278c62); }
.example-arrow { position: absolute; inset-inline-end: 14px; top: 20px; opacity: .4; }
svg { width: 21px; height: 21px; fill: currentColor; }
.latest { position: absolute; bottom: 12px; inset-inline-end: 24px; width: 46px; height: 46px; display: grid; place-items: center; padding: 10px; border-radius: 50%; background: var(--assistant-card, #fff); color: var(--assistant-action, #278c62); box-shadow: 0 5px 20px #23452d14; }
.unread-dot { position: absolute; top: 5px; inset-inline-end: 5px; width: 7px; height: 7px; border-radius: 50%; background: var(--assistant-action, #278c62); }
button:hover:not(:disabled) { background: var(--assistant-tint, #e8f3e9); }
button:active:not(:disabled) { transform: scale(.98); }
/* Reading anchors measure message geometry while navigation restores the list. */
.turn { animation: message-fade 220ms ease both; }
.welcome { animation: message-enter 220ms ease both; }
@keyframes message-fade { from { opacity: 0; } to { opacity: 1; } }
@keyframes message-enter { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 440px) { .welcome { padding: 12px 4px; } .examples { gap: 10px; } .examples button { padding: 14px; } }
@media (prefers-reduced-motion: reduce) { .turn, .welcome { animation: none; } button { transition: none; } }
</style>
