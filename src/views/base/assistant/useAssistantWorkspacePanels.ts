import { computed, nextTick, onBeforeUnmount, onMounted, shallowRef, useTemplateRef, watch, type Ref } from 'vue';

interface PanelOptions {
    variant: () => 'desktop' | 'mobile';
    historyOpen: Ref<boolean>;
    memoryOpen: Ref<boolean>;
    showMemories: () => Promise<unknown>;
    captureScope: () => () => boolean;
}

/** Owns responsive panel measurements, keyboard containment and focus restoration. */
export function useAssistantWorkspacePanels(options: PanelOptions) {
    const viewport = shallowRef(0);
    const narrow = shallowRef(false);
    const historyPanel = useTemplateRef<HTMLElement>('historyPanel');
    const memoryRegion = useTemplateRef<HTMLElement>('memoryRegion');
    let memoryTrigger: HTMLElement | null = null;
    async function openMemories() {
        const current = options.captureScope();
        const active = document.activeElement as HTMLElement;
        memoryTrigger = active;
        await options.showMemories();
        await nextTick();
        if (!current()) return;
        const heading = memoryRegion.value?.querySelector('h2');
        if (heading) {
            heading.tabIndex = -1;
            heading.focus();
        }
    }
    watch(options.memoryOpen, async (visible) => {
        const current = options.captureScope();
        if (!visible) {
            await nextTick();
            if (current() && memoryTrigger?.isConnected) memoryTrigger.focus();
        }
    });
    watch(historyPanel, (panel) => {
        if (panel instanceof HTMLDialogElement && !panel.open) panel.showModal();
    }, { flush: 'post' });
    let historyTrigger: HTMLElement | null = null;
    watch(options.historyOpen, async (visible) => {
        const current = options.captureScope();
        if (visible) {
            historyTrigger = document.activeElement as HTMLElement;
            await nextTick();
            if (current() && narrow.value) historyPanel.value?.querySelector<HTMLButtonElement>('button')?.focus();
        } else {
            await nextTick();
            if (current() && historyTrigger?.isConnected) historyTrigger.focus();
        }
    });
    function historyKey(event: KeyboardEvent) {
        if (event.key === 'Escape') {
            options.historyOpen.value = false;
            return;
        }
        if (!narrow.value || event.key !== 'Tab') return;
        const items = Array.from(
            historyPanel.value?.querySelectorAll<HTMLElement>(
                'button:not(:disabled), input:not(:disabled), summary, [tabindex="0"]'
            ) ?? []
        ).filter((item) => {
            if (!item.getClientRects().length) return false;
            for (
                let parent = item.parentElement;
                parent && parent !== historyPanel.value;
                parent = parent.parentElement
            ) {
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
    const shellStyle = computed(() =>
        options.variant() === 'mobile' && viewport.value ? { height: `${viewport.value}px` } : undefined
    );
    function updateViewport() {
        narrow.value = window.innerWidth < 900;
        viewport.value = Math.max(180, (window.visualViewport?.height ?? window.innerHeight) - 56);
    }
    onMounted(() => {
        updateViewport();
        window.addEventListener('resize', updateViewport);
        window.visualViewport?.addEventListener('resize', updateViewport);
    });
    onBeforeUnmount(() => {
        window.removeEventListener('resize', updateViewport);
        window.visualViewport?.removeEventListener('resize', updateViewport);
    });
    return { narrow, shellStyle, historyPanel, openMemories, historyKey };
}
