<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch, useTemplateRef, nextTick } from 'vue';
import AssistantConversationHistory from './AssistantConversationHistory.vue';
import AssistantMemoryManager from './AssistantMemoryManager.vue';
import AssistantExperienceMessageList from './AssistantExperienceMessageList.vue';
import AssistantExperienceComposer from './AssistantExperienceComposer.vue';
import AssistantCopyButton from './AssistantCopyButton.vue';
import AssistantCapacityNotice from './AssistantCapacityNotice.vue';
import { useAssistantReference } from '@/views/base/assistant/useAssistantReference.ts';
import { useAssistantCloudWorkspace } from '@/views/base/assistant/useAssistantCloudWorkspace.ts';
import type { AssistantConversationDTO, AssistantMemoryDTO } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{ variant: 'desktop' | 'mobile'; conversationId?: string; messageId?: string; versionId?: string }>();
const emit = defineEmits<{ navigate: [id: string, messageId?: string, versionId?: string, replace?: boolean]; openReference: [id: string] }>();
const confirmationDialog = useTemplateRef<HTMLDialogElement>('confirmationDialog');
const confirmationText = ref('');
let confirmationTrigger: HTMLElement | null = null;
let confirmationResolve: ((value: boolean) => void) | undefined;
async function confirm(message: string): Promise<boolean> {
    const active = document.activeElement as HTMLElement;
    confirmationTrigger = active?.closest('details')?.querySelector('summary') ?? active;
    confirmationText.value = message; await nextTick(); confirmationDialog.value?.showModal();
    return new Promise(resolve => { confirmationResolve = resolve; });
}
function resolveConfirmation(value: boolean) {
    confirmationDialog.value?.close(); confirmationResolve?.(value); confirmationResolve = undefined;
    void nextTick(() => {
        if (confirmationTrigger?.isConnected && confirmationTrigger.getClientRects().length) confirmationTrigger.focus();
        else (historyPanel.value?.querySelector<HTMLButtonElement>('button') ?? toolbarMore.value?.querySelector('summary'))?.focus();
    });
}
function closeMore(event: KeyboardEvent) { const details = event.currentTarget as HTMLDetailsElement; details.open = false; details.querySelector('summary')?.focus(); }
const { store, tt, error, working, historyOpen, memoryOpen, archived, query, editing, input, summaryOpen,
    summaryPeriod, retryQuerySource, restartQuery, messages, disabledReason, run, refreshHistory, search, rename, archive, remove, clear, edit, send, branch,
    continueConversation, retry, nextPage, version, summary, showMemories, saveMemory, forget } = useAssistantCloudWorkspace(confirm);
const memoryManager = useTemplateRef<InstanceType<typeof AssistantMemoryManager>>('memoryManager');
const historyManager = useTemplateRef<InstanceType<typeof AssistantConversationHistory>>('historyManager');
async function saveMemoryAndClose(item: AssistantMemoryDTO | undefined, text: string) { const receipt = await saveMemory(item, text); if (receipt) memoryManager.value?.finishEdit(); }
async function renameAndClose(item: AssistantConversationDTO, title: string) { const receipt = await rename(item, title); if (receipt) historyManager.value?.finishEdit(); }
let referenceTrigger: HTMLElement | null = null;
const reference = useAssistantReference(id => store.getReferenceTransaction(id), id => { referenceNotice.value?.close(); referenceTrigger = null; emit('openReference', id); });
function checkReference(...args: Parameters<typeof reference.check>) {
    referenceTrigger = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    return reference.check(...args);
}
function dismissReference() {
    const trigger = referenceTrigger; referenceTrigger = null;
    reference.dismiss(); referenceNotice.value?.close();
    if (trigger?.isConnected) trigger.focus({ preventScroll: true });
}
const referenceLabels = { checking: 'Checking current bill', changed: 'This bill changed after the original query. The saved answer has not been updated.', unverified: 'The current bill amount cannot be compared exactly. Review the current details.', unavailable: 'This bill is no longer accessible. It may have been deleted or access changed.', failed: 'Unable to check the current bill. Retry when the connection is available.' };
async function queryReferenceAgain() {
    const question = messages.value.find(message => message.id === reference.selected.value?.answer.userMessageId);
    if (question) { reference.dismiss(); await retry(question, 'regenerate'); }
}
watch([() => props.conversationId, () => props.messageId, () => props.versionId, () => store.conversationId, () => store.capabilities?.ownerScope], () => { referenceTrigger = null; reference.dismiss(); referenceNotice.value?.close(); }, { flush: 'sync' });
const referenceNotice = useTemplateRef<HTMLDialogElement>('referenceNotice');
watch(reference.state, async () => {
    await nextTick();
    if (reference.state.value && !referenceNotice.value?.open) referenceNotice.value?.showModal();
    else if (!reference.state.value) referenceNotice.value?.close();
});
const messageList = useTemplateRef<InstanceType<typeof AssistantExperienceMessageList>>('messageList');
const exportDialog = useTemplateRef<HTMLDialogElement>('exportDialog');
const exporting = ref(false);
const versionCursors = new Map<string, string | undefined>();
async function loadVersions(id: string) { await run(async () => { const result = await store.loadVersions(id, versionCursors.get(id)); versionCursors.set(id, result?.nextCursor); }); }
const exportId = ref<string>();
const exportTitle = ref('assistant');
const exportFormat = ref('markdown');
const exportAll = ref(false);
const exportVersions = ref(false);
watch(exportFormat, format => { if (format !== 'json') exportVersions.value = false; });
let exportController: AbortController | undefined;
let initialized = false;
let navigating = false;
const viewport = ref(0);
const narrow = ref(false);
const historyPanel = useTemplateRef<HTMLElement>('historyPanel');
const memoryRegion = useTemplateRef<HTMLElement>('memoryRegion');
const toolbarMore = useTemplateRef<HTMLDetailsElement>('toolbarMore');
let memoryTrigger: HTMLElement | null = null;
async function openMemories() {
    const active = document.activeElement as HTMLElement; const menu = toolbarMore.value;
    memoryTrigger = menu?.open ? menu.querySelector('summary') : active; if (menu) menu.open = false;
    await showMemories(); await nextTick(); const heading = memoryRegion.value?.querySelector('h2'); if (heading) { heading.tabIndex = -1; heading.focus(); }
}
watch(memoryOpen, async visible => { if (!visible) { await nextTick(); memoryTrigger?.focus(); } });
let historyTrigger: HTMLElement | null = null;
watch(historyOpen, async visible => {
    if (visible) { historyTrigger = document.activeElement as HTMLElement; await nextTick(); if (narrow.value) historyPanel.value?.querySelector<HTMLButtonElement>('button')?.focus(); }
    else { await nextTick(); historyTrigger?.focus(); }
});
function historyKey(event: KeyboardEvent) {
    if (event.key === 'Escape') { historyOpen.value = false; return; }
    if (!narrow.value || event.key !== 'Tab') return;
    const items = Array.from(historyPanel.value?.querySelectorAll<HTMLElement>('button:not(:disabled), input:not(:disabled), summary, [tabindex="0"]') ?? []).filter(item => {
        if (!item.getClientRects().length) return false;
        for (let parent = item.parentElement; parent && parent !== historyPanel.value; parent = parent.parentElement) {
            if (parent instanceof HTMLDetailsElement && !parent.open) {
                const summary = parent.querySelector(':scope > summary');
                if (!summary?.contains(item)) return false;
            }
        }
        return true;
    });
    if (!items.length) return;
    event.preventDefault();
    const current = items.indexOf(document.activeElement as HTMLElement);
    const target = current < 0 ? 0 : (current + (event.shiftKey ? -1 : 1) + items.length) % items.length;
    items[target]?.focus();
}
const shellStyle = computed(() => props.variant === 'mobile' && viewport.value ? { height: `${viewport.value}px` } : undefined);
const anchor = computed(() => props.messageId ? { messageId: props.versionId || props.messageId, offset: 0, following: false } : store.readingPosition());
const conversation = computed(() => store.detail?.conversation);
function updateViewport() { narrow.value = window.innerWidth < 900; viewport.value = Math.max(180, (window.visualViewport?.height ?? window.innerHeight) - 56); }
function navigate(id: string, messageId?: string, versionId?: string) { editing.value = undefined; historyOpen.value = false; memoryOpen.value = false; emit('navigate', id, messageId, versionId); }
async function openRoute() {
    navigating = true;
    const requested = props.conversationId ?? '';
    const recovered = store.recoverBranchNavigation(requested);
    const target = recovered ?? requested;
    if (recovered && recovered !== requested) emit('navigate', recovered, undefined, undefined, true);
    try {
        await run(() => store.open(target, recovered ? undefined : props.messageId, recovered ? undefined : props.versionId));
        if (props.conversationId === target && store.detail?.conversation.id === target) await store.acknowledgeBranchNavigation(target);
        await nextTick(); messageList.value?.restoreAnchor(anchor.value);
    } finally { navigating = false; }
}
watch(() => [props.conversationId, props.messageId, props.versionId], () => { if (initialized) void openRoute(); });
watch(() => store.conversationId, id => { if (initialized && !navigating && id !== (props.conversationId ?? '')) emit('navigate', id, undefined, undefined, true); });
watch(historyOpen, value => store.setHistoryVisible(value));
watch(() => [store.capabilities?.memoryUseEpoch, store.capabilities?.settingsVersion], () => {
    if (memoryOpen.value && store.memories && (store.memories.memoryUseEpoch !== store.capabilities?.memoryUseEpoch || store.memories.settingsVersion !== store.capabilities?.settingsVersion)) void run(() => store.loadMemories());
});
onMounted(async () => {
    updateViewport(); window.addEventListener('resize', updateViewport); window.visualViewport?.addEventListener('resize', updateViewport);
    await run(() => store.initialize()); initialized = true;
    if (props.conversationId === undefined && store.recentConversationId) emit('navigate', store.recentConversationId, undefined, undefined, true);
    else await openRoute();
    store.attach();
});
onBeforeUnmount(() => { reference.dismiss(); resolveConfirmation(false); store.detach(); store.setHistoryVisible(false); window.removeEventListener('resize', updateViewport); exportController?.abort(); window.visualViewport?.removeEventListener('resize', updateViewport); });
async function removeConversation(item: AssistantConversationDTO) { await remove(item); await nextTick(); if (document.activeElement === document.body) historyPanel.value?.querySelector<HTMLButtonElement>('button')?.focus(); }
async function removeLocalDraft(id: string) { if (await confirm(tt('Delete this local draft?'))) await store.removeDraft(id); }
function setMemoryMode(enabled: boolean) { const value = conversation.value; if (value) void run(() => store.execute('update', { id: value.id, expectedMetadataVersion: value.metadataVersion, expectedContextVersion: value.contextVersion, memoryMode: enabled ? 'inherit' : 'off' })); }
function continueQuery() { const id = conversation.value?.pendingQuerySourceMessageId; if (id) void run(() => store.send({ mode: 'next_page', content: tt('Show the next page of the current query results'), sourceMessageId: id })); }
async function latest() { if (store.windowMissingLatest || store.updatesAvailable) await run(() => store.loadLatest()); await nextTick(); messageList.value?.scrollToLatest(); }
async function loadOlder() { const saved = messageList.value?.captureAnchor(); await run(() => store.loadOlder()); await nextTick(); messageList.value?.restoreAnchor(saved); }
function openExport(item?: AssistantConversationDTO, all = false) {
    exportId.value = item?.id ?? store.conversationId; exportTitle.value = item?.title ?? conversation.value?.title ?? 'assistant';
    exportAll.value = all; exportVersions.value = false; exportDialog.value?.showModal();
}
async function download() {
    exporting.value = true; exportController = new AbortController();
    await run(async () => {
        const blob = await store.exportData({ id: exportAll.value ? undefined : exportId.value, all: exportAll.value, format: exportFormat.value, includeVersions: exportVersions.value, viewedVersionIds: !exportAll.value && exportId.value === store.conversationId && !exportVersions.value ? Object.values(store.viewedVersions).map(answer => answer.assistantMessageId) : undefined }, exportController?.signal);
        const url = URL.createObjectURL(blob); const link = document.createElement('a'); link.href = url;
        link.download = `${exportTitle.value.replace(/[\\/\u0000-\u001f\u007f]/g, '_').slice(0, 100)}.${exportFormat.value === 'json' ? 'json' : 'md'}`;
        link.click(); setTimeout(() => URL.revokeObjectURL(url), 1000); exportDialog.value?.close();
    }); exporting.value = false;
}
</script>
<template>
    <section class="cloud-workspace" :class="variant" :style="shellStyle">
        <aside v-if="historyOpen" ref="historyPanel" class="cloud-history" :role="narrow ? 'dialog' : undefined" :aria-modal="narrow ? true : undefined" :aria-label="tt('Conversation history')" @keydown="historyKey">
            <AssistantConversationHistory ref="historyManager" v-if="store.capabilities" :items="store.history.items" :matches="store.searchResults.items" :selected-id="store.conversationId" :loading="working" :error="error" :archived="archived" :title-limit="store.capabilities.limits.titleCharacters" :has-more="!!(query ? store.searchResults.nextCursor : store.history.nextCursor)"
                @select="navigate" @new="navigate('')" @close="historyOpen = false" @search="search" @archived="archived = $event; refreshHistory()" @rename="renameAndClose" @archive="archive" @delete="removeConversation" @export="openExport" @retry="refreshHistory()" @load-more="refreshHistory(query ? store.searchResults.nextCursor : store.history.nextCursor)" />
        </aside>
        <section class="cloud-main" :inert="historyOpen && narrow">
            <header class="cloud-toolbar"><button type="button" @click="historyOpen = !historyOpen">{{ tt('Conversation history') }}</button><strong>{{ conversation?.title || tt('New conversation') }}</strong><button type="button" @click="navigate('')">{{ tt('New conversation') }}</button><button v-if="variant === 'desktop'" type="button" @click="openMemories">{{ tt('Long-term memory') }}</button><details ref="toolbarMore" @keydown.esc.prevent="closeMore"><summary>{{ tt('More') }}</summary><div class="toolbar-menu" @click="($event.target as HTMLElement).closest('button') && toolbarMore && (toolbarMore.open = false)"><label v-if="conversation" class="conversation-memory-setting"><input type="checkbox" :checked="conversation.memoryMode !== 'off'" @change="setMemoryMode(($event.target as HTMLInputElement).checked)" />{{ tt('Allow memory in this conversation') }}</label><p v-if="store.capabilities && !store.capabilities.memoryEnabled" class="notice">{{ tt('Memory is off globally.') }}</p><button v-if="variant === 'mobile'" type="button" @click="openMemories">{{ tt('Long-term memory') }}</button><button type="button" @click="summaryOpen = !summaryOpen">{{ tt('Generate AI Summary') }}</button><button v-if="conversation" type="button" @click="openExport()">{{ tt('Export') }}</button><button type="button" @click="openExport(undefined, true)">{{ tt('Export all assistant data') }}</button><button type="button" @click="clear">{{ tt('Clear all assistant data') }}</button></div></details></header>
            <dialog ref="referenceNotice" class="export-dialog reference-notice" :aria-label="tt('View Details')" @cancel.prevent="dismissReference">
                <p v-if="reference.state.value" :role="reference.state.value === 'checking' ? 'status' : 'alert'">{{ tt(referenceLabels[reference.state.value]) }}</p>
                <button v-if="reference.state.value === 'changed' || reference.state.value === 'unverified'" type="button" @click="reference.viewCurrent">{{ tt('View current bill') }}</button>
                <button v-if="reference.state.value === 'changed' && !store.busy && store.capabilities?.generationAvailable" type="button" @click="queryReferenceAgain">{{ tt('Ask this question again') }}</button>
                <button v-if="reference.state.value === 'unavailable' || reference.state.value === 'failed'" type="button" @click="reference.retry">{{ tt('Retry') }}</button>
                <button type="button" @click="dismissReference">{{ tt('Close') }}</button>
            </dialog>
            <p v-if="error || store.error" class="notice" role="alert">{{ error || tt('Assistant request failed') }} <button type="button" @click="run(() => store.synchronize())">{{ tt('Retry') }}</button></p>
            <p v-if="error && retryQuerySource" class="notice"><button type="button" :disabled="!!disabledReason || store.busy" @click="restartQuery">{{ tt('Query these bills again') }}</button></p>
            <p v-if="store.updatesAvailable || store.windowMissingLatest" class="notice"><button type="button" @click="latest">{{ tt('New content · Back to latest') }}</button></p>
            <p v-if="store.legacyBackend" class="notice" role="status">{{ tt('Upgrade the server to use saved assistant conversations.') }}</p>
            <p v-if="!store.online" class="notice" role="status">{{ tt('Offline. Drafts are saved on this device.') }}</p>
            <p v-if="store.local.scope === 'page'" class="notice" role="status">{{ tt('Only kept on this page. Refreshing may lose unsent content.') }}<button type="button" @click="run(() => store.retryLocalPersistence())">{{ tt('Retry') }}</button></p>
            <div v-for="pending in store.local.pending.filter(item => !item.resolvedConversationId)" :key="pending.clientRequestId" class="notice" role="status"><span>{{ tt(pending.rejected || store.delivery[pending.clientRequestId] === 'unsent' ? 'Not delivered' : 'Confirming changes') }}</span><pre v-if="pending.text">{{ pending.text }}</pre><AssistantCopyButton v-if="pending.text" :text="pending.text" /><button type="button" @click="run(() => store.reconcilePending())">{{ tt('Check status') }}</button><button v-if="!pending.rejected && store.delivery[pending.clientRequestId] === 'unsent'" type="button" @click="run(() => store.retryPending(pending.clientRequestId))">{{ tt('Retry') }}</button><button v-if="pending.rejected" type="button" @click="run(() => store.discardRejected(pending.clientRequestId))">{{ tt('Discard rejected request') }}</button><button v-if="pending.text && !store.draft" type="button" @click="run(() => store.restorePending(pending.clientRequestId))">{{ tt('Restore this question') }}</button></div>
            <details v-if="store.local.drafts.length > 1 || store.local.reason" class="notice"><summary>{{ tt('Local drafts') }}</summary><article v-for="draft in store.local.drafts" :key="draft.id"><p>{{ draft.conflict ? tt('Conflicting draft — both copies have been kept') : draft.conversationId || tt('New conversation') }}</p><pre>{{ draft.text }}</pre><AssistantCopyButton :text="draft.text" /><button v-if="draft.conversationId === store.conversationId" type="button" @click="store.chooseDraft(draft.id)">{{ tt('Use this draft') }}</button><button type="button" @click="removeLocalDraft(draft.id)">{{ tt('Delete') }}</button></article></details>
            <template v-if="memoryOpen"><div ref="memoryRegion" class="cloud-memory"><AssistantMemoryManager ref="memoryManager" v-if="store.capabilities" :items="store.memories?.items ?? []" :enabled="store.memories?.enabled ?? false" :busy="working || store.local.pending.some(p => !p.rejected && (p.action.includes('Memory') || p.action === 'memorySettings'))" :error="error" :byte-limit="store.capabilities.limits.memoryBytes" @close="memoryOpen = false" @retry="showMemories" @enabled="run(() => store.execute('memorySettings', { enabled: $event, expectedSettingsVersion: store.memories?.settingsVersion ?? 0 }))" @save="saveMemoryAndClose" @forget="forget" @forget-all="forget()" @source="navigate" /></div></template>
            <template v-else>
                <p v-if="store.conversationId && !store.detail && !store.loading" class="notice" role="status">{{ tt('This conversation does not exist or is not accessible.') }}<button type="button" @click="navigate('')">{{ tt('New conversation') }}</button></p>
                <p v-if="store.loading" class="notice" role="status">{{ tt('Loading') }}</p>
                <p v-if="conversation?.sourceInvalidated" class="notice" role="status">{{ tt('Source data has changed or was removed. Old context will not be reused. Ask again to query current bills.') }}</p>
                <p v-if="conversation?.sourceConversationId" class="notice">{{ tt(conversation.sourceInvalidated ? 'The source is no longer available. Its context cannot be used.' : 'Branched conversation') }} <button v-if="!conversation.sourceInvalidated" type="button" @click="navigate(conversation.sourceConversationId, conversation.sourceMessageId)">{{ tt('View source') }}</button></p>

                <p v-if="conversation?.pendingContinuation" class="notice">{{ tt('New branch created. Review your question below and send to start generating.') }}</p>
                <p v-if="conversation?.pendingQuerySourceMessageId" class="notice"><button type="button" :disabled="!!disabledReason || store.busy" @click="continueQuery">{{ tt('Continue this query') }}</button><button type="button" :disabled="!!disabledReason || store.busy" @click="retryQuerySource = conversation.pendingQuerySourceMessageId; restartQuery()">{{ tt('Query these bills again') }}</button></p>
                <p v-if="conversation?.archived" class="notice"><button type="button" @click="archive(conversation)">{{ tt('Unarchive and continue') }}</button></p>
                <AssistantExperienceMessageList ref="messageList" :messages="messages" :latest-message-id="messages.find(message => store.isLatestMessage(message.id))?.id ?? ''" :has-older="store.detail?.hasMore" :loading-older="working" :busy="store.busy" :anchor="anchor" :examples="[tt('Compare this month with last month'), tt('How much did I spend on dining?'), tt('Show my account balances')]" @example="store.setDraft" @load-older="loadOlder" @latest="latest" @anchor-change="store.savePosition" @open-reference="checkReference" @edit="edit" @retry="retry" @regenerate="retry($event, 'regenerate')" @branch="branch" @select-version="version" @load-versions="loadVersions($event.id)" @next-page="nextPage" />
                <AssistantCapacityNotice v-if="conversation && store.capabilities" :used="conversation.usedBytes" :reserved="conversation.reservedBytes" :limit="store.capabilities.limits.conversationBytes" :reason="conversation.turnCount >= store.capabilities.limits.conversationTurns * .8 ? 'Approaching conversation turn capacity' : undefined" :can-continue="!store.busy && !!messages.length" @continue="continueConversation" @export="openExport()" @manage="historyOpen = true" />
                <form v-if="summaryOpen" class="notice" @submit.prevent="summary"><label>{{ tt('Summary period') }}<input v-model="summaryPeriod" /></label><button type="submit" :disabled="!!disabledReason || store.busy || !summaryPeriod.trim()">{{ tt('Generate AI Summary') }}</button><button type="button" @click="summaryOpen = false">{{ tt('Cancel') }}</button></form>
                <AssistantExperienceComposer v-if="store.capabilities" v-model="input" :variant="variant" :limits="store.capabilities.limits" :generation-state="store.detail?.activeRun?.state" :sending="store.submitting" :disabled-reason="disabledReason" :editing-label="editing ? tt('Edit and create branch') : undefined" @send="send" @stop="run(() => store.stop())" @cancel-edit="editing = undefined" />
            </template>
        </section>
        <dialog ref="confirmationDialog" class="export-dialog" @cancel.prevent="resolveConfirmation(false)"><p>{{ confirmationText }}</p><button type="button" autofocus @click="resolveConfirmation(false)">{{ tt('Cancel') }}</button><button type="button" @click="resolveConfirmation(true)">{{ tt('Confirm') }}</button></dialog>
        <dialog ref="exportDialog" class="export-dialog" @cancel="exportController?.abort()"><form @submit.prevent="download"><h2>{{ tt('Export') }}</h2><label>{{ tt('Format') }}<select v-model="exportFormat"><option value="markdown">Markdown</option><option value="json">JSON</option></select></label><label v-if="exportFormat === 'json'"><input v-model="exportVersions" type="checkbox" />{{ tt('Include older answer versions') }}</label><p v-if="error" role="alert">{{ error }}</p><button type="submit" :disabled="exporting">{{ tt(exporting ? 'Exporting' : 'Download') }}</button><button type="button" @click="exportController?.abort(); exportDialog?.close()">{{ tt('Cancel') }}</button></form></dialog>
    </section>
</template>
<style scoped>
.reference-notice { max-height: 80dvh; overflow-y: auto; }
.cloud-workspace { display: flex; min-height: 0; height: calc(100dvh - 150px); background: var(--assistant-surface, transparent); overflow: hidden; border: 1px solid rgb(127 127 127 / 25%); border-radius: 16px; }
.desktop { --assistant-surface: rgb(var(--v-theme-surface)); --assistant-action: rgb(var(--v-theme-primary)); }.mobile { --assistant-surface: var(--f7-page-bg-color); --assistant-action: var(--f7-theme-color); }.cloud-main { min-width: 0; min-height: 0; flex: 1; display: flex; flex-direction: column; }.cloud-history { width: 290px; flex-shrink: 0; border-inline-end: 1px solid rgb(127 127 127 / 30%); }.cloud-memory { min-height: 0; overflow-y: auto; }.cloud-toolbar { display: flex; align-items: center; gap: 8px; padding: 8px 12px; flex-wrap: wrap; border-bottom: 1px solid rgb(127 127 127 / 30%); }.cloud-toolbar strong { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }.cloud-toolbar details { position: relative; }.toolbar-menu { position: absolute; z-index: 3; inset-inline-end: 0; width: 240px; background: var(--assistant-surface, Canvas); color: inherit; box-shadow: 0 4px 20px #0003; border-radius: 10px; }.toolbar-menu button { width: 100%; text-align: start; }.notice { margin: 0; padding: 8px 16px; font-size: .875rem; max-height: 180px; overflow-y: auto; }.notice pre { white-space: pre-wrap; }button, summary, input, select { color: inherit; font: inherit; min-height: 44px; padding: 8px; border: 1px solid rgb(127 127 127 / 90%); border-radius: 8px; background: transparent; }button, summary { cursor: pointer; width: auto; display: inline-flex; align-items: center; } .cloud-toolbar > button { flex: 0 0 auto; } .conversation-memory-setting { padding: 12px; display: inline-flex; align-items: center; gap: 8px; }.conversation-memory-setting input[type="checkbox"] { width: 18px; height: 18px; min-height: 18px; flex: 0 0 18px; }.notice label { display: inline-flex; align-items: center; gap: 8px; } .notice input[type="checkbox"] { width: 18px; height: 18px; min-height: 18px; flex: 0 0 18px; margin: 0; }button:disabled { opacity: .55; cursor: default; }button:focus-visible, summary:focus-visible { outline: 2px solid currentColor; outline-offset: 2px; }.export-dialog { max-width: min(90vw, 500px); border-radius: 16px; padding: 24px; background: var(--assistant-surface, Canvas); color: inherit; }.export-dialog label { display: block; margin-block: 12px; }.mobile { height: calc(100dvh - 56px); border: 0; border-radius: 0; }.mobile .cloud-toolbar strong { flex-basis: 100%; order: -1; padding-inline: 8px; }.mobile .cloud-history { position: absolute; z-index: 5; inset: 0; width: 100%; background: var(--assistant-surface, Canvas); }.mobile .cloud-toolbar { gap: 4px; padding: 4px; }.mobile .cloud-toolbar button, .mobile summary { font-size: .8rem; }
@media (max-width: 899px) { .cloud-workspace { position: relative; }.cloud-history { position: absolute; z-index: 5; inset: 0; width: min(100%, 360px); background: var(--assistant-surface, Canvas); color: inherit; box-shadow: 6px 0 20px #0003; } }
@media (max-height: 440px) and (orientation: landscape) {
    .mobile .cloud-toolbar { flex-wrap: nowrap; }
    .mobile .cloud-toolbar strong { flex-basis: 0; order: 0; }
    .mobile .cloud-toolbar button, .mobile .cloud-toolbar summary { box-sizing: border-box; height: 44px; }
    .mobile :deep(.experience-messages) { padding-block: 8px; }
    .mobile :deep(.experience-composer) { padding-block: 4px max(4px, env(safe-area-inset-bottom)); }
    .mobile :deep(.composer-row) { padding: 4px; }
    .mobile :deep(.composer-row textarea) { height: 44px; min-height: 44px; max-height: 44px; overflow-y: auto; }
    .mobile :deep(.composer-notes) { margin-top: 2px; }
}
</style>
