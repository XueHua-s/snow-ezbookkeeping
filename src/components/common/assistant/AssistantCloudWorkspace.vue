<script setup lang="ts">
import { computed, watch, useTemplateRef, nextTick } from 'vue';
import AssistantConfirmationDialog from './AssistantConfirmationDialog.vue';
import AssistantExportDialog from './AssistantExportDialog.vue';
import { useAssistantWorkspacePanels } from '@/views/base/assistant/useAssistantWorkspacePanels.ts';
import { useAssistantWorkspaceLifecycle } from '@/views/base/assistant/useAssistantWorkspaceLifecycle.ts';
import AssistantConversationHistory from './AssistantConversationHistory.vue';
import AssistantMemoryManager from './AssistantMemoryManager.vue';
import AssistantExperienceMessageList from './AssistantExperienceMessageList.vue';
import AssistantExperienceComposer from './AssistantExperienceComposer.vue';
import AssistantWorkspaceToolbar from './AssistantWorkspaceToolbar.vue';
import AssistantConversationNotices from './AssistantConversationNotices.vue';
import AssistantRecoveryPanel from './AssistantRecoveryPanel.vue';
import AssistantCapacityNotice from './AssistantCapacityNotice.vue';
import AssistantReferenceDialog from './AssistantReferenceDialog.vue';
import { useAssistantCloudWorkspace } from '@/views/base/assistant/useAssistantCloudWorkspace.ts';
import type { AssistantConversationDTO, AssistantMemoryDTO } from '@/models/ai_assistant_conversation.ts';
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
    historyOpen,
    memoryOpen,
    archived,
    query,
    historySearching,
    includeVersions,
    editing,
    input,
    summaryOpen,
    summaryPeriod,
    retryQuerySource,
    restartQuery,
    messages,
    disabledReason,
    run,
    refreshHistory,
    rename,
    archive,
    remove,
    clear,
    edit,
    send,
    branch,
    continueConversation,
    retry,
    nextPage,
    version,
    summary,
    showMemories,
    saveMemory,
    forget
} = useAssistantCloudWorkspace(confirm, () => routeKey.value);
const toolbar = useTemplateRef<InstanceType<typeof AssistantWorkspaceToolbar>>('toolbar');
const memoryManager = useTemplateRef<InstanceType<typeof AssistantMemoryManager>>('memoryManager');
const historyManager = useTemplateRef<InstanceType<typeof AssistantConversationHistory>>('historyManager');
async function saveMemoryAndClose(item: AssistantMemoryDTO | undefined, text: string) {
    const current = captureScope();
    const receipt = await saveMemory(item, text);
    if (receipt && current()) memoryManager.value?.finishEdit();
}
async function renameAndClose(item: AssistantConversationDTO, title: string) {
    const current = captureScope();
    const receipt = await rename(item, title);
    if (receipt && current()) historyManager.value?.finishEdit();
}
const referenceDialog = useTemplateRef<InstanceType<typeof AssistantReferenceDialog>>('referenceDialog');
const messageList = useTemplateRef<InstanceType<typeof AssistantExperienceMessageList>>('messageList');
const exportDialog = useTemplateRef<InstanceType<typeof AssistantExportDialog>>('exportDialog');
const versionCursors = new Map<string, string | undefined>();
async function loadVersions(id: string) {
    await run(async (current) => {
        const result = await store.loadVersions(id, versionCursors.get(id));
        if (current()) versionCursors.set(id, result?.nextCursor);
    });
}
const ownerKey = computed(() => JSON.stringify([store.capabilities?.ownerScope, store.capabilities?.deletionEpoch]));
const context = computed(() => JSON.stringify([ownerKey.value, store.conversationId, routeKey.value]));
watch(context, () => versionCursors.clear(), { flush: 'sync' });
const { narrow, shellStyle, historyPanel, openMemories, historyKey } = useAssistantWorkspacePanels({
    variant: () => props.variant,
    historyOpen,
    memoryOpen,
    showMemories,
    captureScope
});
const memoryBusy = computed(
    () =>
        working.value ||
        store.local.pending.some(
            (item) => !item.rejected && (item.action.includes('Memory') || item.action === 'memorySettings')
        )
);
const latestMessageId = computed(() => messages.value.find((message) => store.isLatestMessage(message.id))?.id ?? '');
const conversation = computed(() => store.detail?.conversation);
function navigate(id: string, messageId?: string, versionId?: string) {
    editing.value = undefined;
    historyOpen.value = false;
    memoryOpen.value = false;
    emit('navigate', id, messageId, versionId);
}
const { anchor } = useAssistantWorkspaceLifecycle(store, {
    route: () => props,
    navigate: (...args) => emit('navigate', ...args),
    restoreAnchor: (value) => messageList.value?.restoreAnchor(value),
    run
});
watch(historyOpen, (value) => store.setHistoryVisible(value));
watch(
    () => [store.capabilities?.memoryUseEpoch, store.capabilities?.settingsVersion],
    () => {
        if (
            memoryOpen.value &&
            store.memories &&
            (store.memories.memoryUseEpoch !== store.capabilities?.memoryUseEpoch ||
                store.memories.settingsVersion !== store.capabilities?.settingsVersion)
        )
            void run(() => store.loadMemories());
    }
);
async function removeConversation(item: AssistantConversationDTO) {
    const current = captureScope();
    await remove(item);
    await nextTick();
    if (current() && document.activeElement === document.body)
        historyPanel.value?.querySelector<HTMLButtonElement>('button')?.focus();
}
function setMemoryMode(enabled: boolean) {
    const value = conversation.value;
    if (value)
        void run(() =>
            store.execute('update', {
                id: value.id,
                expectedMetadataVersion: value.metadataVersion,
                expectedContextVersion: value.contextVersion,
                memoryMode: enabled ? 'inherit' : 'off'
            })
        );
}
function continueQuery() {
    const id = conversation.value?.pendingQuerySourceMessageId;
    if (id)
        void run(() =>
            store.send({
                mode: 'next_page',
                content: tt('Show the next page of the current query results'),
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
function openExport(item?: AssistantConversationDTO, all = false) {
    exportDialog.value?.open(item, all);
}
function fallbackFocus() {
    const button = historyPanel.value?.querySelector<HTMLButtonElement>('button');
    if (button) button.focus();
    else toolbar.value?.focus();
}
</script>
<template>
    <section class="cloud-workspace" :class="variant" :style="shellStyle">
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
                @export="openExport"
                @retry="refreshHistory()"
                @load-more="refreshHistory(query ? store.searchResults.nextCursor : store.history.nextCursor)"
            />
        </component>
        <section class="cloud-main" :inert="historyOpen && narrow">
            <AssistantWorkspaceToolbar
                ref="toolbar"
                :variant="variant"
                :conversation="conversation"
                :memory-enabled="store.capabilities?.memoryEnabled"
                @history="historyOpen = !historyOpen"
                @new="navigate('')"
                @memories="openMemories"
                @summary="summaryOpen = !summaryOpen"
                @export="openExport()"
                @export-all="openExport(undefined, true)"
                @clear="clear"
                @memory-mode="setMemoryMode"
            />
            <AssistantReferenceDialog
                ref="referenceDialog"
                :context="context"
                :messages="messages"
                @open-reference="emit('openReference', $event)"
                @retry-question="retry($event, 'regenerate')"
            />
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
            <template v-if="memoryOpen">
                <div ref="memoryRegion" class="cloud-memory">
                    <AssistantMemoryManager
                        ref="memoryManager"
                        :key="ownerKey"
                        v-if="store.capabilities"
                        :items="store.memories?.items ?? []"
                        :enabled="store.memories?.enabled ?? false"
                        :busy="memoryBusy"
                        :error="error"
                        :byte-limit="store.capabilities.limits.memoryBytes"
                        @close="memoryOpen = false"
                        @retry="showMemories"
                        @enabled="
                            run(() =>
                                store.execute('memorySettings', {
                                    enabled: $event,
                                    expectedSettingsVersion: store.memories?.settingsVersion ?? 0
                                })
                            )
                        "
                        @save="saveMemoryAndClose"
                        @forget="forget"
                        @forget-all="forget()"
                        @source="navigate"
                    />
                </div>
            </template>
            <template v-else>
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
                    :busy="store.busy"
                    :anchor="anchor"
                    :examples="[
                        tt('Compare this month with last month'),
                        tt('How much did I spend on dining?'),
                        tt('Show my account balances')
                    ]"
                    @example="store.setDraft"
                    @load-older="loadOlder"
                    @latest="latest"
                    @anchor-change="store.savePosition"
                    @open-reference="referenceDialog?.check"
                    @edit="edit"
                    @retry="retry"
                    @regenerate="retry($event, 'regenerate')"
                    @branch="branch"
                    @select-version="version"
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
                    @export="openExport()"
                    @manage="historyOpen = true"
                />
                <form v-if="summaryOpen" class="notice" @submit.prevent="summary">
                    <label class="summary-label">
                        {{ tt('Summary period') }}
                        <input class="workspace-input" v-model="summaryPeriod" name="summary-period" autocomplete="off" />
                    </label>
                    <button
                        class="workspace-action"
                        type="submit"
                        :disabled="!!disabledReason || store.busy || !summaryPeriod.trim()"
                    >
                        {{ tt('Generate AI Summary') }}
                    </button>
                    <button class="workspace-action" type="button" @click="summaryOpen = false">
                        {{ tt('Cancel') }}
                    </button>
                </form>
                <AssistantExperienceComposer
                    v-if="store.capabilities"
                    v-model="input"
                    :variant="variant"
                    :limits="store.capabilities.limits"
                    :generation-state="store.detail?.activeRun?.state"
                    :sending="store.submitting"
                    :disabled-reason="disabledReason"
                    :editing-label="editing ? tt('Edit and create branch') : undefined"
                    @send="send"
                    @stop="run(() => store.stop())"
                    @cancel-edit="editing = undefined"
                />
            </template>
        </section>
        <AssistantConfirmationDialog ref="confirmationDialog" :context="context" :fallback-focus="fallbackFocus" />
        <AssistantExportDialog ref="exportDialog" :context="context" />
    </section>
</template>
<style scoped src="./assistant-workspace.css" />
