// @env browser
import { afterEach, expect, it } from 'vitest';
import { render } from 'vitest-browser-vue';
import { defineComponent, h, useTemplateRef } from 'vue';
import { useAssistantAutoScroll } from './useAssistantAutoScroll.ts';

afterEach(() => window.getSelection()?.removeAllRanges());

async function layout() {
    // Let ResizeObserver delivery and the resulting scroll/RAF settle.
    for (let frame = 0; frame < 4; frame++) await new Promise(requestAnimationFrame);
}

async function fixture() {
    let scrolling!: ReturnType<typeof useAssistantAutoScroll>;
    const screen = await render(defineComponent({
        setup() {
            scrolling = useAssistantAutoScroll(useTemplateRef('panel'), useTemplateRef('content'));
            return () => h('div', [
                h('p', { id: 'outside' }, 'Unrelated page text'),
                h('div', { ref: 'panel', style: 'height:300px;width:500px;overflow:auto;overflow-anchor:none' }, [
                    h('div', { ref: 'content' }, Array.from({ length: 12 }, (_, index) =>
                        h('section', { 'data-message-id': String(index), style: 'height:100px' }, `Message ${index}`)))
                ])
            ]);
        }
    }));
    const panel = screen.container.querySelector<HTMLElement>('[style*="overflow"]')!;
    const content = panel.firstElementChild as HTMLElement;
    await layout();
    expect(panel.scrollTop).toBe(900);
    return { scrolling, panel, content, screen };
}

function select(element: Element) {
    const range = document.createRange();
    range.selectNodeContents(element);
    window.getSelection()!.removeAllRanges();
    window.getSelection()!.addRange(range);
    expect(window.getSelection()!.toString()).not.toBe('');
}

it('pauses following for panel selection during content growth, then explicitly resumes', async () => {
    const { scrolling, panel, content } = await fixture();
    select(content.lastElementChild!);
    const top = panel.scrollTop;
    content.lastElementChild!.append(document.createElement('div'));
    (content.lastElementChild as HTMLElement).style.height = '400px';
    await layout();
    expect(panel.scrollTop).toBe(top);
    expect(scrolling.following.value).toBe(false);
    expect(scrolling.unread.value).toBe(true);
    window.getSelection()!.removeAllRanges();
    scrolling.scrollToLatest();
    await layout();
    expect(panel.scrollTop).toBe(1200);
    expect(scrolling.unread.value).toBe(false);
});

it('preserves selected reading position when only the panel resizes', async () => {
    const { scrolling, panel, content } = await fixture();
    select(content.lastElementChild!);
    panel.style.height = '200px';
    await layout();
    expect(panel.scrollTop).toBe(900);
    expect(scrolling.following.value).toBe(false);
});

it('keeps following when selection belongs to the surrounding page', async () => {
    const { scrolling, panel, content, screen } = await fixture();
    select(screen.container.querySelector('#outside')!);
    (content.lastElementChild as HTMLElement).style.height = '400px';
    await layout();
    expect(panel.scrollTop).toBe(1200);
    expect(scrolling.following.value).toBe(true);
    expect(scrolling.unread.value).toBe(false);
});

it('preserves a paused message anchor through prepend and late image layout', async () => {
    const { scrolling, panel, content } = await fixture();
    panel.scrollTop = 450;
    await layout();
    const before = scrolling.capture()!;
    expect(before.following).toBe(false);
    const older = document.createElement('section');
    older.dataset['messageId'] = 'older';
    older.style.height = '200px';
    content.prepend(older);
    await layout();
    expect(scrolling.capture()).toEqual(before);
    const image = document.createElement('img');
    image.style.display = 'block';
    content.firstElementChild!.after(image);
    image.src = 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="120" height="180"/>');
    await image.decode();
    await layout();
    expect(scrolling.capture()).toEqual(before);
    panel.style.height = '250px';
    await layout();
    expect(scrolling.capture()).toEqual(before);
    expect(panel.scrollTop).toBe(830);
});
