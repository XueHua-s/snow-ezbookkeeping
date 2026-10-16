import { computed, nextTick, onBeforeUnmount, onMounted, watch } from 'vue';
import type { useAssistantStore } from '@/stores/assistant.ts';
import type { AssistantReadingAnchor } from './useAssistantAutoScroll.ts';

interface WorkspaceRoute {
    conversationId?: string;
    messageId?: string;
    versionId?: string;
}
interface LifecycleOptions {
    route: () => WorkspaceRoute;
    navigate: (id: string, messageId?: string, versionId?: string, replace?: boolean) => void;
    restoreAnchor: (anchor: AssistantReadingAnchor | undefined) => void;
    run: <T>(action: () => Promise<T>) => Promise<T | undefined>;
}

/** Owns route synchronization and the lifetime of the workspace's store attachment. */
export function useAssistantWorkspaceLifecycle(store: ReturnType<typeof useAssistantStore>, options: LifecycleOptions) {
    let disposed = false;
    let initialized = false;
    let attached = false;
    let navigating = false;
    let generation = 0;
    const anchor = computed(() => {
        const route = options.route();
        return route.messageId
            ? { messageId: route.versionId || route.messageId, offset: 0, following: false }
            : store.readingPosition();
    });

    async function openRoute() {
        const ticket = ++generation;
        const current = () => !disposed && ticket === generation;
        navigating = true;
        const route = options.route();
        const requested = route.conversationId ?? '';
        const recovered = store.recoverBranchNavigation(requested);
        const target = recovered ?? requested;
        if (recovered && recovered !== requested) options.navigate(recovered, undefined, undefined, true);
        try {
            await options.run(() =>
                store.open(target, recovered ? undefined : route.messageId, recovered ? undefined : route.versionId),
            );
            if (!current()) return;
            if (options.route().conversationId === target && store.detail?.conversation.id === target) {
                await store.acknowledgeBranchNavigation(target);
            }
            await nextTick();
            if (!current()) return;
            options.restoreAnchor(anchor.value);
            // The id watcher is suppressed while opening. A 404 can clear the id
            // during that interval, so reconcile it once this route read settles.
            if (target && !store.conversationId) options.navigate('', undefined, undefined, true);
        } finally {
            if (current()) navigating = false;
        }
    }
    watch(
        [
            () => options.route().conversationId,
            () => options.route().messageId,
            () => options.route().versionId,
            () => store.capabilities?.ownerScope,
            () => store.capabilities?.deletionEpoch,
        ],
        async () => {
            const ticket = ++generation;
            // Invalidate immediately, but do not read partially updated props or
            // session state from inside their synchronous mutation stack.
            if (!initialized || disposed) return;
            await nextTick();
            if (!disposed && ticket === generation) void openRoute();
        },
        { flush: 'sync' },
    );
    watch(
        () => store.conversationId,
        (id) => {
            if (initialized && !disposed && !navigating && id !== (options.route().conversationId ?? '')) {
                options.navigate(id, undefined, undefined, true);
            }
        },
    );
    onMounted(async () => {
        await options.run(() => store.initialize());
        if (disposed) return;
        initialized = true;
        if (options.route().conversationId === undefined && store.recentConversationId) {
            options.navigate(store.recentConversationId, undefined, undefined, true);
        } else {
            await openRoute();
        }
        if (!disposed) {
            store.attach();
            attached = true;
        }
    });
    onBeforeUnmount(() => {
        disposed = true;
        generation++;
        if (attached) store.detach();
        store.setHistoryVisible(false);
    });
    return { anchor };
}
