// @env browser
import { useResizeObserver } from '@vueuse/core';
import type { ReadonlyElementRef } from './types.ts';

/** Follow rendered content, including asynchronous Markdown and image layout.
 * VueUse disconnects both observers when the owning component is disposed.
 */
export function useAssistantAutoScroll(panel: ReadonlyElementRef, content: ReadonlyElementRef): void {
    function scrollToEnd(): void {
        if (panel.value) {
            panel.value.scrollTop = panel.value.scrollHeight;
        }
    }
    useResizeObserver(content, scrollToEnd);
    useResizeObserver(panel, scrollToEnd);
}
