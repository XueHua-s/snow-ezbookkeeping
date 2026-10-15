// @env browser
import { onBeforeUnmount, readonly, shallowRef, watch, nextTick } from 'vue';
import { useResizeObserver } from '@vueuse/core';
import type { ReadonlyElementRef } from './types.ts';
export interface AssistantReadingAnchor { messageId: string; offset: number; following: boolean }
/** Message-relative anchors survive prepended pages and changing Markdown layout. */
export function useAssistantAutoScroll(panel: ReadonlyElementRef, content: ReadonlyElementRef) {
    const following = shallowRef(true);
    const unread = shallowRef(false);
    let anchor: AssistantReadingAnchor | undefined;
    let programmatic = false;
    function capture(): AssistantReadingAnchor | undefined {
        const viewport = panel.value;
        if (!viewport) return undefined;
        const top = viewport.getBoundingClientRect().top;
        const items = Array.from(viewport.querySelectorAll<HTMLElement>('[data-message-id]'));
        const item = items.find(element => element.getBoundingClientRect().bottom > top);
        return item ? { messageId: item.dataset['messageId']!, offset: item.getBoundingClientRect().top - top, following: following.value } : undefined;
    }
    function scrollToLatest(): void {
        const viewport = panel.value;
        if (!viewport) return;
        programmatic = true;
        viewport.scrollTop = viewport.scrollHeight;
        following.value = true;
        unread.value = false;
        anchor = capture();
        requestAnimationFrame(() => { programmatic = false; });
    }
    function restore(value?: AssistantReadingAnchor): void {
        if (!value || !panel.value) return;
        following.value = value.following;
        anchor = value;
        if (value.following) { scrollToLatest(); return; }
        const item = Array.from(panel.value.querySelectorAll<HTMLElement>('[data-message-id]')).find(element => element.dataset['messageId'] === value.messageId);
        if (item) {
            programmatic = true;
            panel.value.scrollTop += item.getBoundingClientRect().top - panel.value.getBoundingClientRect().top - value.offset;
            requestAnimationFrame(() => { programmatic = false; });
        }
    }
    function onScroll(): void {
        if (programmatic || !panel.value) return;
        following.value = panel.value.scrollHeight - panel.value.clientHeight - panel.value.scrollTop <= 80;
        if (following.value) unread.value = false;
        anchor = capture();
    }
    function layoutChanged(): void {
        const selection = window.getSelection();
        if (following.value && !selection?.toString()) scrollToLatest();
        else { unread.value = true; restore(anchor); }
    }
    watch(panel, (element, previous) => { previous?.removeEventListener('scroll', onScroll); element?.addEventListener('scroll', onScroll, { passive: true }); }, { immediate: true });
    useResizeObserver(content, layoutChanged);
    useResizeObserver(panel, () => restore(anchor));
    onBeforeUnmount(() => panel.value?.removeEventListener('scroll', onScroll));
    async function preserveWhile(action: () => Promise<void>): Promise<void> { const saved = capture(); await action(); await nextTick(); restore(saved); }
    return { following: readonly(following), unread: readonly(unread), capture, restore, scrollToLatest, preserveWhile };
}
