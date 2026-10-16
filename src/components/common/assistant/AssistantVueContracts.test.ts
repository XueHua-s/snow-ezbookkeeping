// @vitest-environment happy-dom
import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import MemoryManager from './AssistantMemoryManager.vue';
import CopyButton from './AssistantCopyButton.vue';
import type { AssistantMemoryDTO } from '@/models/ai_assistant_conversation.ts';
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (key: string) => key }) }));
let wrapper: VueWrapper | undefined;
afterEach(() => {
    wrapper?.unmount();
    wrapper = undefined;
    vi.useRealTimers();
    vi.unstubAllGlobals();
});

describe('memory source identity', () => {
    it('preserves conversation-only source buttons across reordering and emits their distinct destinations', async () => {
        const item: AssistantMemoryDTO = {
            schemaVersion: 1, id: 'm1', key: 'budget', text: 'Budget preference', type: 'preference',
            status: 'active', sourceMessageIds: [], sources: [
                { conversationId: '11', messageId: '', version: 1 },
                { conversationId: '22', messageId: '', version: 1 },
                { conversationId: '33', messageId: '44', version: 1 }
            ], createdAt: '2026-10-16T00:00:00Z', updatedAt: '2026-10-16T00:00:00Z', version: 1, explicit: true
        };
        wrapper = mount(MemoryManager, { props: { items: [item], enabled: true, byteLimit: 1000 } });
        const sources = () => wrapper!.findAll('button').filter(button => button.text() === 'View source');
        const original = sources().map(button => button.element);
        await wrapper.setProps({ items: [{ ...item, sources: [item.sources[1]!, item.sources[0]!, item.sources[2]!] }] });
        expect(sources()[0]!.element).toBe(original[1]);
        expect(sources()[1]!.element).toBe(original[0]);
        expect(sources()[2]!.element).toBe(original[2]);
        for (const button of sources()) await button.trigger('click');
        expect(wrapper.emitted('source')).toEqual([['22', ''], ['11', ''], ['33', '44']]);
    });
});

describe('clipboard lifecycle', () => {
    it('does not start a success timer after an outstanding clipboard write outlives its component', async () => {
        vi.useFakeTimers();
        let finish!: () => void;
        vi.stubGlobal('navigator', { clipboard: { writeText: vi.fn(() => new Promise<void>(resolve => { finish = resolve; })) } });
        wrapper = mount(CopyButton, { props: { text: 'copy me' } });
        await wrapper.get('button').trigger('click');
        wrapper.unmount();
        wrapper = undefined;
        finish();
        await flushPromises();
        expect(vi.getTimerCount()).toBe(0);
    });
    it('keeps the latest copy failure when an older successful write finishes later', async () => {
        let finish!: () => void;
        const writeText = vi.fn().mockImplementationOnce(() => new Promise<void>(resolve => { finish = resolve; }))
            .mockRejectedValueOnce(new Error('denied'));
        vi.stubGlobal('navigator', { clipboard: { writeText } });
        wrapper = mount(CopyButton, { props: { text: 'old' } });
        await wrapper.get('button').trigger('click');
        await wrapper.setProps({ text: 'new' });
        await wrapper.get('button').trigger('click');
        await flushPromises();
        finish();
        await flushPromises();
        expect(wrapper.get('textarea').element.value).toBe('new');
        expect(wrapper.text()).not.toContain('Copied');
    });
});
