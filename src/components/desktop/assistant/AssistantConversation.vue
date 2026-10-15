<script setup lang="ts">
import { useTemplateRef } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';
import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';
import SnackBar from '@/components/desktop/SnackBar.vue';
import AssistantCloudWorkspace from '@/components/common/assistant/AssistantCloudWorkspace.vue';
const route = useRoute();
const router = useRouter();
const referenceDialog = useTemplateRef<InstanceType<typeof EditDialog>>('referenceDialog');
const snackbar = useTemplateRef<InstanceType<typeof SnackBar>>('snackbar');
function navigate(id: string, messageId?: string, versionId?: string, replace = false) {
    const location = { path: '/assistant', query: { conversationId: id, messageId, versionId } };
    return replace ? router.replace(location) : router.push(location);
}
function openReference(id: string) {
    referenceDialog.value?.open({ id }).catch(error => { if (error && !error.processed) snackbar.value?.showError(error); });
}
function queryValue(key: string): string | undefined { const value = route.query[key]; return typeof value === 'string' ? value : undefined; }
</script>
<template>
    <AssistantCloudWorkspace variant="desktop" :conversation-id="queryValue('conversationId')" :message-id="queryValue('messageId')" :version-id="queryValue('versionId')" @navigate="navigate" @open-reference="openReference" />
    <EditDialog ref="referenceDialog" :type="TransactionEditPageType.Transaction" />
    <SnackBar ref="snackbar" />
</template>
