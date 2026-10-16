// @vitest-environment happy-dom
import { afterEach, expect, it, vi } from 'vitest';
import { defineComponent, h, shallowRef } from 'vue';
import { mount } from '@vue/test-utils';
import { useAssistantAutoScroll } from './useAssistantAutoScroll.ts';
const observers = vi.hoisted(() => [] as (() => void)[]);
vi.mock('@vueuse/core', () => ({
    useResizeObserver: (_target: unknown, callback: () => void) => {
        observers.push(callback);
    },
}));
afterEach(() => {
    vi.restoreAllMocks();
    observers.length = 0;
});

it('cancels pending frames and ignores queued observer/restore work after unmount', () => {
    let scrolling!: ReturnType<typeof useAssistantAutoScroll>;
    let frameId = 0;
    const canceled = vi.spyOn(window, 'cancelAnimationFrame').mockImplementation(() => undefined);
    vi.spyOn(window, 'requestAnimationFrame').mockImplementation(() => ++frameId);
    const panel = document.createElement('div');
    const wrapper = mount(
        defineComponent({
            setup() {
                scrolling = useAssistantAutoScroll(shallowRef(panel), shallowRef(document.createElement('div')));
                return () => h('div');
            },
        }),
    );
    scrolling.scrollToLatest();
    scrolling.scrollToLatest();
    expect(canceled).toHaveBeenCalledWith(1);
    wrapper.unmount();
    expect(canceled).toHaveBeenCalledWith(2);
    panel.scrollTop = 17;
    for (const callback of observers) callback();
    scrolling.restore();
    scrolling.scrollToLatest();
    expect(panel.scrollTop).toBe(17);
    expect(frameId).toBe(2);
});
