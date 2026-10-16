<script setup lang="ts">
import { mdiChartLine, mdiChartDonut, mdiWalletOutline, mdiAutoFix } from '@mdi/js';
import { computed, watch, useTemplateRef, nextTick } from 'vue';
import AssistantConfirmationDialog from './AssistantConfirmationDialog.vue';
import { useAssistantWorkspacePanels } from '@/views/base/assistant/useAssistantWorkspacePanels.ts';
import { useAssistantWorkspaceLifecycle } from '@/views/base/assistant/useAssistantWorkspaceLifecycle.ts';
import AssistantConversationHistory from './AssistantConversationHistory.vue';
import AssistantExperienceMessageList from './AssistantExperienceMessageList.vue';
import AssistantExperienceComposer from './AssistantExperienceComposer.vue';
import AssistantWorkspaceToolbar from './AssistantWorkspaceToolbar.vue';
import AssistantConversationNotices from './AssistantConversationNotices.vue';
import AssistantRecoveryPanel from './AssistantRecoveryPanel.vue';
import AssistantCapacityNotice from './AssistantCapacityNotice.vue';
import AssistantReferenceDialog from './AssistantReferenceDialog.vue';
import { useAssistantCloudWorkspace } from '@/views/base/assistant/useAssistantCloudWorkspace.ts';
import type { AssistantConversationDTO } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{
    variant: 'desktop' | 'mobile';
    conversationId?: string;
    messageId?: string;
    versionId?: string;
}>();
const emit = defineEmits<{
    navigate: [id: string, messageId?: string, versionId?: string, replace?: boolean];
    openReference: [id: string];
}>();
const confirmationDialog = useTemplateRef<InstanceType<typeof AssistantConfirmationDialog>>('confirmationDialog');
const confirm = (message: string) => confirmationDialog.value?.confirm(message) ?? Promise.resolve(false);
const routeKey = computed(() => JSON.stringify([props.conversationId, props.messageId, props.versionId]));
const {
    store,
    tt,
    captureScope,
    removeLocalDraft,
    error,
    working,
    answerNotice,
    historyOpen,
    archived,
    query,
    historySearching,
    includeVersions,
    input,
    retryQuerySource,
    restartQuery,
    messages,
    disabledReason,
    run,
    refreshHistory,
    rename,
    archive,
    remove,
    send,
    continueConversation,
    retry,
    nextPage,
    version,
    useAnswer
} = useAssistantCloudWorkspace(confirm, () => routeKey.value);
const toolbar = useTemplateRef<InstanceType<typeof AssistantWorkspaceToolbar>>('toolbar');
const historyManager = useTemplateRef<InstanceType<typeof AssistantConversationHistory>>('historyManager');
async function renameAndClose(item: AssistantConversationDTO, title: string) {
    const current = captureScope();
    const receipt = await rename(item, title);
    if (receipt && current()) historyManager.value?.finishEdit();
}
const referenceDialog = useTemplateRef<InstanceType<typeof AssistantReferenceDialog>>('referenceDialog');
const messageList = useTemplateRef<InstanceType<typeof AssistantExperienceMessageList>>('messageList');
const versionCursors = new Map<string, string | undefined>();
const versionLoads = new Map<string, symbol>();
async function loadVersions(id: string) {
    if (versionLoads.has(id) || working.value || store.busy) return;
    const ticket = Symbol();
    versionLoads.set(id, ticket);
    const anchor = messageList.value?.captureAnchor();
    try {
        await run(async (current) => {
            const result = await store.loadVersions(id, versionCursors.get(id));
            if (result && current()) {
                versionCursors.set(id, result.nextCursor);
                answerNotice.value = tt('Earlier answers loaded. Use the arrows to browse.');
                await nextTick();
                if (current() && anchor) messageList.value?.restoreAnchor(anchor);
            }
        });
    } finally {
        if (versionLoads.get(id) === ticket) versionLoads.delete(id);
    }
}
const ownerKey = computed(() => JSON.stringify([store.capabilities?.ownerScope, store.capabilities?.deletionEpoch]));
const context = computed(() => JSON.stringify([ownerKey.value, store.conversationId, store.detail?.conversation.id, routeKey.value]));
watch(context, () => { versionCursors.clear(); versionLoads.clear(); }, { flush: 'sync' });
const { narrow, shellStyle, historyPanel, historyKey } = useAssistantWorkspacePanels({
    variant: () => props.variant,
    historyOpen,
    captureScope
});
const latestMessageId = computed(() => messages.value.find((message) => store.isLatestMessage(message.id))?.id ?? '');
const conversation = computed(() => store.detail?.conversation);
function navigate(id: string, messageId?: string, versionId?: string) {
    historyOpen.value = false;
    emit('navigate', id, messageId, versionId);
}
const { anchor } = useAssistantWorkspaceLifecycle(store, {
    route: () => props,
    navigate: (...args) => emit('navigate', ...args),
    restoreAnchor: (value) => messageList.value?.restoreAnchor(value),
    run
});
watch(historyOpen, (value) => store.setHistoryVisible(value));
async function removeConversation(item: AssistantConversationDTO) {
    const current = captureScope();
    await remove(item);
    await nextTick();
    if (current() && document.activeElement === document.body)
        historyPanel.value?.querySelector<HTMLButtonElement>('button')?.focus();
}
function continueQuery() {
    const id = conversation.value?.pendingQuerySourceMessageId;
    if (id)
        void run(() =>
            store.send({
                mode: 'next_page',
                content: tt('Continue finding the remaining bills matching these conditions.'),
                sourceMessageId: id
            })
        );
}
function restartFromSource(id: string) {
    retryQuerySource.value = id;
    void restartQuery();
}
async function latest() {
    const current = captureScope();
    if (store.windowMissingLatest || store.updatesAvailable) await run(() => store.loadLatest());
    await nextTick();
    if (current()) messageList.value?.scrollToLatest();
}
async function loadOlder() {
    const current = captureScope();
    const saved = messageList.value?.captureAnchor();
    await run(() => store.loadOlder());
    await nextTick();
    if (current()) messageList.value?.restoreAnchor(saved);
}
function fallbackFocus() {
    const button = historyPanel.value?.querySelector<HTMLButtonElement>('button');
    if (button) button.focus();
    else toolbar.value?.focus();
}
</script>
<template>
    <section class="cloud-workspace" :class="variant" :style="shellStyle">
        <!-- Identity changes discard the old history immediately, including its leave animation. -->
        <Transition :key="ownerKey" name="history-panel">
            <component
                :is="narrow ? 'dialog' : 'aside'"
                v-if="historyOpen"
                ref="historyPanel"
                class="cloud-history"
                :role="narrow ? 'dialog' : undefined"
                :aria-modal="narrow ? true : undefined"
                :aria-label="tt('Conversation history')"
                @keydown="historyKey"
                @cancel.prevent="historyOpen = false"
            >
                <AssistantConversationHistory
                    ref="historyManager"
                    :key="ownerKey"
                    v-if="store.capabilities"
                    :items="store.history.items"
                    :matches="working || historySearching ? [] : store.searchResults.items"
                    :selected-id="store.conversationId"
                    :loading="working || historySearching"
                    :error="error"
                    :archived="archived"
                    :title-limit="store.capabilities.limits.titleCharacters"
                    :has-more="!!(query ? store.searchResults.nextCursor : store.history.nextCursor)"
                    @select="navigate"
                    @new="navigate('')"
                    @close="historyOpen = false"
                    v-model:query="query"
                    v-model:include-versions="includeVersions"
                    @archived="
                        archived = $event;
                        refreshHistory();
                    "
                    @rename="renameAndClose"
                    @archive="archive"
                    @delete="removeConversation"
                    @retry="refreshHistory()"
                    @load-more="refreshHistory(query ? store.searchResults.nextCursor : store.history.nextCursor)"
                />
            </component>
        </Transition>
        <section class="cloud-main" :inert="historyOpen && narrow">
            <AssistantWorkspaceToolbar
                ref="toolbar"
                :variant="variant"
                :conversation="conversation"
                @history="historyOpen = !historyOpen"
                @new="navigate('')"
            />
            <AssistantReferenceDialog
                ref="referenceDialog"
                :context="context"
                :messages="messages"
                @open-reference="emit('openReference', $event)"
                @retry-question="retry($event, 'regenerate')"
            />
            <p v-if="answerNotice" class="notice" role="status">{{ answerNotice }}</p>
            <p v-if="error || store.error" class="notice" role="alert">
                {{ error || tt('Assistant request failed') }}
                <button class="workspace-action" type="button" @click="run(() => store.synchronize())">
                    {{ tt('Retry') }}
                </button>
            </p>
            <p v-if="error && retryQuerySource" class="notice">
                <button
                    class="workspace-action"
                    type="button"
                    :disabled="!!disabledReason || store.busy"
                    @click="restartQuery"
                >
                    {{ tt('Query these bills again') }}
                </button>
            </p>
            <p v-if="store.updatesAvailable || store.windowMissingLatest" class="notice">
                <button class="workspace-action" type="button" @click="latest">
                    {{ tt('New content · Back to latest') }}
                </button>
            </p>
            <AssistantRecoveryPanel
                @retry-persistence="run(() => store.retryLocalPersistence())"
                @reconcile="run(() => store.reconcilePending())"
                @retry-pending="run(() => store.retryPending($event))"
                @discard="run(() => store.discardRejected($event))"
                @restore="run(() => store.restorePending($event))"
                @remove-draft="removeLocalDraft"
            />
            <AssistantConversationNotices
                :disabled-reason="disabledReason"
                @navigate="navigate"
                @continue-query="continueQuery"
                @restart-query="restartFromSource"
                @unarchive="archive"
            />
            <AssistantExperienceMessageList
                ref="messageList"
                :messages="messages"
                :latest-message-id="latestMessageId"
                :has-older="store.detail?.hasMore"
                :loading-older="working"
                :busy="store.busy || working"
                :write-disabled-reason="disabledReason"
                :anchor="anchor"
                :examples="[
                    { text: tt('Compare this month with last month'), icon: mdiChartLine },
                    { text: tt('How much did I spend on dining?'), icon: mdiChartDonut },
                    { text: tt('Show my account balances'), icon: mdiWalletOutline },
                    { text: tt('Help me find ways to save'), icon: mdiAutoFix }
                ]"
                @example="store.setDraft"
                @load-older="loadOlder"
                @latest="latest"
                @anchor-change="store.savePosition"
                @open-reference="referenceDialog?.check"
                @retry="retry"
                @regenerate="retry($event, 'regenerate')"
                @select-version="version"
                @use-answer="useAnswer"
                @load-versions="loadVersions($event.id)"
                @next-page="nextPage"
            />
            <AssistantCapacityNotice
                v-if="conversation && store.capabilities"
                :used="conversation.usedBytes"
                :reserved="conversation.reservedBytes"
                :limit="store.capabilities.limits.conversationBytes"
                :reason="
                    conversation.turnCount >= store.capabilities.limits.conversationTurns * 0.8
                        ? 'Approaching conversation turn capacity'
                        : undefined
                "
                :can-continue="!store.busy && !!messages.length"
                @continue="continueConversation"
                @manage="historyOpen = true"
            />
            <AssistantExperienceComposer
                v-if="store.capabilities"
                v-model="input"
                :variant="variant"
                :limits="store.capabilities.limits"
                :generation-state="store.detail?.activeRun?.state"
                :sending="store.submitting"
                :disabled-reason="disabledReason"
                @send="send"
                @stop="run(() => store.stop())"
            />
        </section>
        <AssistantConfirmationDialog ref="confirmationDialog" :context="context" :fallback-focus="fallbackFocus" />
    </section>
</template>
<style scoped src="./assistant-workspace.css" />
