<script setup lang="ts">
import { useTemplateRef, watch, nextTick } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import type { AssistantMessageDTO, AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';
import { useAssistantAutoScroll, type AssistantReadingAnchor } from '@/views/base/assistant/useAssistantAutoScroll.ts';
import { assistantStateLabels, assistantReasonLabels } from './assistantPresentation.ts';
import AssistantMarkdownContent from '@/components/common/AssistantMarkdownContent.vue';
import AssistantExperienceActions from './AssistantExperienceActions.vue';
import AssistantExperienceEvidence from './AssistantExperienceEvidence.vue';
const props = defineProps<{ messages: readonly AssistantMessageDTO[]; latestMessageId?: string; hasOlder?: boolean; loadingOlder?: boolean; olderError?: string; anchor?: AssistantReadingAnchor; busy?: boolean; examples?: readonly string[] }>();
const emit = defineEmits<{ loadOlder: []; latest: []; openReference: [id: string, answer: AssistantRunSnapshot]; retry: [message: AssistantMessageDTO]; regenerate: [message: AssistantMessageDTO]; edit: [message: AssistantMessageDTO]; branch: [message: AssistantMessageDTO]; selectVersion: [message: AssistantMessageDTO, id: string]; loadVersions: [message: AssistantMessageDTO]; nextPage: [message: AssistantMessageDTO]; example: [text: string]; anchorChange: [anchor: AssistantReadingAnchor] }>();
const { tt } = useI18n();
const panel = useTemplateRef<HTMLElement>('panel');
const content = useTemplateRef<HTMLElement>('content');
const scrolling = useAssistantAutoScroll(panel, content);
watch(() => props.anchor, async anchor => { await nextTick(); scrolling.restore(anchor); }, { immediate: true });
function isLatest(message: AssistantMessageDTO, index: number): boolean { return props.latestMessageId === undefined ? index === props.messages.length - 1 : message.id === props.latestMessageId; }
function saveAnchor(): void { const anchor = scrolling.capture(); if (anchor) emit('anchorChange', anchor); }
defineExpose({ scrollToLatest: scrolling.scrollToLatest, captureAnchor: scrolling.capture, restoreAnchor: scrolling.restore });

</script>
<template>
    <section class="message-region">
        <div ref="panel" class="experience-messages" @scroll.passive="saveAnchor">
            <div ref="content" class="message-column">
                <button v-if="hasOlder" type="button" :disabled="loadingOlder" @click="emit('loadOlder')">{{ tt(loadingOlder ? 'Loading' : 'Load older messages') }}</button>
                <p v-if="olderError" role="alert">{{ tt(olderError) }} <button type="button" @click="emit('loadOlder')">{{ tt('Retry') }}</button></p>
                <div v-if="!messages.length" class="welcome"><h2>{{ tt('How can I help with your finances?') }}</h2><p>{{ tt('Conversations are saved on this server and sent to the configured model service when generating.') }}</p><div class="examples"><button v-for="example in examples" :key="example" type="button" @click="emit('example', example)">{{ example }}</button></div></div>
                <section v-for="(message, index) in messages" :key="message.id" :data-message-id="message.id" class="turn">
                    <div class="question"><div class="question-text">{{ message.content }}</div><AssistantExperienceActions :content="message.content" edit :disabled-reason="busy ? 'Stop generation before editing or regenerating.' : undefined" @edit="emit('edit', message)" /></div>
                    <article v-if="message.selectedAnswer" :data-message-id="message.selectedAnswer.assistantMessageId" class="answer">
                        <AssistantMarkdownContent v-if="message.selectedAnswer.content" :content="message.selectedAnswer.content" :thinking="message.selectedAnswer.processSummary" />
                        <p class="run-status" role="status" aria-live="polite">{{ tt(assistantStateLabels[message.selectedAnswer.state]) }}</p>
                        <p v-if="message.selectedAnswer.reasonCode" class="run-error" role="alert">{{ tt(assistantReasonLabels[message.selectedAnswer.reasonCode]) }}</p>
                        <p v-if="['failed', 'incomplete', 'interrupted', 'canceled'].includes(message.selectedAnswer.state)">{{ tt('This answer was not completed. Partial text may be unreliable.') }}</p>
                        <AssistantExperienceEvidence :answer="message.selectedAnswer" :can-page="!busy" :page-label="!isLatest(message, index) ? tt('Continue this query in a new conversation') : undefined" @open-reference="emit('openReference', $event, message.selectedAnswer)" @next-page="emit('nextPage', message)" />
                        <AssistantExperienceActions :content="message.selectedAnswer.content" :streaming="['queued', 'preparing', 'running', 'cancel_requested'].includes(message.selectedAnswer.state)" :retry="isLatest(message, index) && ['failed', 'incomplete', 'interrupted', 'canceled'].includes(message.selectedAnswer.state)" :regenerate="isLatest(message, index) && message.selectedAnswer.state === 'completed'" :branch="!isLatest(message, index)" :versions="message.versions" :selected-version-id="message.selectedAnswer.assistantMessageId" :context-version-id="message.selectedAssistantMessageId" :has-more-versions="message.hasMoreVersions" :disabled-reason="busy ? 'Stop generation before editing or regenerating.' : undefined" @retry="emit('retry', message)" @regenerate="emit('regenerate', message)" @branch="emit('branch', message)" @select-version="emit('selectVersion', message, $event)" @load-versions="emit('loadVersions', message)" />
                    </article>
                </section>
            </div>
        </div>
        <button v-if="!scrolling.following.value" type="button" class="latest" @click="emit('latest'); scrolling.scrollToLatest()">{{ tt(scrolling.unread.value ? 'New content · Back to latest' : 'Back to latest') }}</button>
    </section>
</template>
<style scoped>
* { box-sizing: border-box; }
.message-region { min-height: 0; flex: 1; position: relative; display: flex; flex-direction: column; } .experience-messages { flex: 1; overflow-y: auto; min-height: 0; width: 100%; padding: 24px 16px; overscroll-behavior: contain; } .message-column { max-width: 800px; margin-inline: auto; font-size: 1rem; line-height: 1.7; } .turn { margin-bottom: 24px; } .question { max-width: 85%; width: fit-content; margin-inline-start: auto; margin-bottom: 24px; } .question-text { white-space: pre-wrap; overflow-wrap: anywhere; background: rgb(127 127 127 / 12%); padding: 12px 16px; border-radius: 18px; } .answer { min-width: 0; } .run-status { font-size: .8em; margin: 8px 0; } button { min-height: 44px; padding: 8px 12px; background: transparent; color: inherit; font: inherit; border: 1px solid rgb(127 127 127 / 90%); border-radius: 10px; cursor: pointer; } button:focus-visible { outline: 2px solid currentColor; outline-offset: 2px; } .welcome { padding-block: 48px 24px; text-align: center; } .welcome h2 { font-size: 1.5em; font-weight: 500; } .examples { display: flex; flex-wrap: wrap; justify-content: center; gap: 8px; } .latest { align-self: center; margin: 8px 0; flex-shrink: 0; background: Canvas; color: CanvasText; box-shadow: 0 3px 12px #0002; } @media (max-width: 440px) { .welcome { padding-top: 12px; } .examples { flex-direction: column; } }
</style>
