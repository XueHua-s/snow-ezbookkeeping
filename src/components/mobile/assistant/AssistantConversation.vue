<script setup lang="ts">
import { shallowRef } from 'vue';
import type { Router } from 'framework7/types';
import AssistantCloudWorkspace from '@/components/common/assistant/AssistantCloudWorkspace.vue';
import { useI18n } from '@/locales/helpers.ts';
const props = defineProps<{ f7router: Router.Router }>();
const { tt } = useI18n();
const active = shallowRef(true);
const route = shallowRef(props.f7router.currentRoute.query ?? {});
function refreshRoute() {
    route.value = { ...(props.f7router.currentRoute.query ?? {}) };
}
function navigate(id: string, messageId?: string, versionId?: string, replace = false) {
    const query = new URLSearchParams({ conversationId: id });
    if (messageId) query.set('messageId', messageId);
    if (versionId) query.set('versionId', versionId);
    const current = props.f7router.currentRoute;
    const target = `/assistant?${query}`;
    // Internal identity/first-send reconciliation changes the current URL, not
    // the page. Reloading during a history transition leaves overlapping pages.
    if (replace && current.path === '/assistant') {
        props.f7router.updateCurrentUrl(target);
        refreshRoute();
        return;
    }
    if (current.path === '/assistant'
        && current.query?.['conversationId'] === id
        && (current.query?.['messageId'] ?? '') === (messageId ?? '')
        && (current.query?.['versionId'] ?? '') === (versionId ?? '')) {
        refreshRoute();
        return;
    }
    props.f7router.navigate(target, { reloadCurrent: replace, ignoreCache: true });
}
function openReference(id: string) {
    props.f7router.navigate(`/transaction/detail?id=${encodeURIComponent(id)}`);
}
</script>
<template>
    <f7-page
        class="assistant-cloud-page"
        :page-content="false"
        @page:beforein="
            refreshRoute();
            active = true;
        "
        @page:beforeout="active = false"
    >
        <f7-navbar :title="tt('AI Assistant')" :back-link="tt('Back')" />
        <div class="assistant-cloud-content">
            <AssistantCloudWorkspace
                v-if="active"
                variant="mobile"
                :conversation-id="route['conversationId']"
                :message-id="route['messageId']"
                :version-id="route['versionId']"
                @navigate="navigate"
                @open-reference="openReference"
            />
        </div>
    </f7-page>
</template>
<style scoped>
.assistant-cloud-content {
    position: absolute;
    top: var(--f7-navbar-height, 56px);
    inset-inline: 0;
    bottom: 0;
    min-height: 0;
    overflow: hidden;
}
</style>
