// @env browser
import { onBeforeUnmount, readonly, shallowRef, watch } from 'vue';
import { useResizeObserver } from '@vueuse/core';
import type { ReadonlyElementRef } from './types.ts';
export interface AssistantReadingAnchor {
    messageId: string;
    offset: number;
    following: boolean;
}
/** Message-relative anchors survive prepended pages and changing Markdown layout. */
export function useAssistantAutoScroll(panel: ReadonlyElementRef, content: ReadonlyElementRef) {
    const following = shallowRef(true);
    const unread = shallowRef(false);
    let anchor: AssistantReadingAnchor | undefined;
    let programmatic = false;
    let disposed = false;
    let frame: number | undefined;
    function markProgrammatic(): void {
        programmatic = true;
        if (frame !== undefined) cancelAnimationFrame(frame);
        frame = requestAnimationFrame(() => {
            frame = undefined;
            programmatic = false;
        });
    }
    function capture(): AssistantReadingAnchor | undefined {
        const viewport = panel.value;
        if (!viewport) return undefined;
        const top = viewport.getBoundingClientRect().top;
        const items = Array.from(viewport.querySelectorAll<HTMLElement>('[data-message-id]'));
        const item = items.find((element) => element.getBoundingClientRect().bottom > top);
        return item
            ? {
                  messageId: item.dataset['messageId']!,
                  offset: item.getBoundingClientRect().top - top,
                  following: following.value,
              }
            : undefined;
    }
    function scrollToLatest(): void {
        const viewport = panel.value;
        if (disposed || !viewport) return;
        markProgrammatic();
        viewport.scrollTop = viewport.scrollHeight;
        following.value = true;
        unread.value = false;
        anchor = capture();
    }
    function restore(value?: AssistantReadingAnchor): void {
        if (disposed) return;
        // A newly opened conversation has no reading anchor: it starts at latest,
        // rather than inheriting the previous conversation's paused follow state.
        if (!value) {
            anchor = undefined;
            following.value = true;
            unread.value = false;
            scrollToLatest();
            return;
        }
        if (!panel.value) return;
        following.value = value.following;
        anchor = value;
        if (value.following) {
            scrollToLatest();
            return;
        }
        const item = Array.from(panel.value.querySelectorAll<HTMLElement>('[data-message-id]')).find(
            (element) => element.dataset['messageId'] === value.messageId,
        );
        if (item) {
            markProgrammatic();
            panel.value.scrollTop +=
                item.getBoundingClientRect().top - panel.value.getBoundingClientRect().top - value.offset;
        }
    }
    function onScroll(): void {
        if (disposed || programmatic || !panel.value) return;
        following.value = panel.value.scrollHeight - panel.value.clientHeight - panel.value.scrollTop <= 80;
        if (following.value) unread.value = false;
        anchor = capture();
    }
    function layoutChanged(contentChanged: boolean): void {
        const viewport = panel.value;
        if (disposed || !viewport) return;
        const selection = viewport.ownerDocument.getSelection();
        const selecting = selection && !selection.isCollapsed &&
            Array.from({ length: selection.rangeCount }, (_, index) => selection.getRangeAt(index))
                .some((range) => range.intersectsNode(viewport));
        if (selecting) {
            // Selection pauses follow just like reading older content. Preserve
            // the pre-layout anchor; restoring a following anchor jumps to latest.
            following.value = false;
            const reading = anchor ?? capture();
            anchor = reading ? { ...reading, following: false } : undefined;
        }
        if (following.value) scrollToLatest();
        else {
            if (contentChanged) unread.value = true;
            if (anchor) restore(anchor);
        }
    }
    watch(
        panel,
        (element, previous) => {
            previous?.removeEventListener('scroll', onScroll);
            element?.addEventListener('scroll', onScroll, { passive: true });
        },
        { immediate: true },
    );
    useResizeObserver(content, () => layoutChanged(true));
    useResizeObserver(panel, () => layoutChanged(false));
    onBeforeUnmount(() => {
        disposed = true;
        panel.value?.removeEventListener('scroll', onScroll);
        if (frame !== undefined) cancelAnimationFrame(frame);
    });
    return { following: readonly(following), unread: readonly(unread), capture, restore, scrollToLatest };
}
