// @vitest-environment happy-dom
import { describe, it, expect, vi, afterEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import Composer from './AssistantExperienceComposer.vue';
import Copy from './AssistantCopyButton.vue';
import LegacyEvidence from '@/components/common/AssistantMessageEvidence.vue';
import Markdown from '@/components/common/AssistantMarkdownContent.vue';
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (key: string) => key, formatAmountToLocalizedNumeralsWithCurrency: (value: unknown) => String(value) }) }));
afterEach(() => vi.unstubAllGlobals());
const limits = { inputCharacters: 4000, inputBytes: 16384 };
describe('assistant experience input', () => {
    it('counts code points, retains exact text, rejects overflow and respects effective byte limits', async () => {
        const wrapper = mount(Composer, { props: { variant: 'desktop', limits, modelValue: '😀'.repeat(4000) } });
        expect(wrapper.get('button').attributes('disabled')).toBeUndefined();
        await wrapper.get('button').trigger('click');
        expect(wrapper.emitted('send')).toHaveLength(1);
        await wrapper.setProps({ modelValue: '😀'.repeat(4001) });
        expect(wrapper.get('button').attributes('disabled')).toBeDefined();
        expect(wrapper.get('textarea').element.value).toBe('😀'.repeat(4001));
        await wrapper.setProps({ modelValue: '你好', limits: { inputCharacters: 4000, inputBytes: 5 } });
        expect(wrapper.get('button').attributes('disabled')).toBeDefined();
        wrapper.unmount();
    });
    it('does not send IME confirmation or mobile return; accepts next draft during generation', async () => {
        const wrapper = mount(Composer, { props: { variant: 'desktop', limits, modelValue: '  中文\n ' } });
        await wrapper.get('textarea').trigger('keydown', { key: 'Enter', isComposing: true });
        await wrapper.get('textarea').trigger('keydown', { key: 'Enter', keyCode: 229 });
        expect(wrapper.emitted('send')).toBeUndefined();
        await wrapper.get('textarea').trigger('keydown', { key: 'Enter' });
        expect(wrapper.emitted('send')).toHaveLength(1);
        await wrapper.setProps({ variant: 'mobile', generationState: 'running' });
        await wrapper.get('textarea').setValue('Next question');
        expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['Next question']);
        await wrapper.get('textarea').trigger('keydown', { key: 'Enter' });
        expect(wrapper.emitted('send')).toHaveLength(1);
        await wrapper.get('button').trigger('click');
        expect(wrapper.emitted('stop')).toHaveLength(1);
        wrapper.unmount();
    });
});
describe('assistant original text copying', () => {
    it('copies click-time Markdown and reports success only after clipboard resolves', async () => {
        let resolve!: () => void;
        const writeText = vi.fn(() => new Promise<void>(done => { resolve = done; }));
        vi.stubGlobal('navigator', { clipboard: { writeText } });
        const original = '# 中文\n\n| A | B |\n|---|---|\n\n```js\n  x();\n```';
        const wrapper = mount(Copy, { props: { text: original, streaming: true } });
        await wrapper.get('button').trigger('click');
        expect(wrapper.text()).not.toContain('Copied current');
        await wrapper.setProps({ text: original + '\nMore' });
        resolve(); await flushPromises();
        expect(writeText).toHaveBeenCalledWith(original);
        expect(wrapper.text()).toContain('Copied current content');
        wrapper.unmount();
    });
    it('keeps a selectable snapshot after permission denial', async () => {
        vi.stubGlobal('navigator', { clipboard: { writeText: vi.fn().mockRejectedValue(new Error('denied')) } });
        const wrapper = mount(Copy, { props: { text: '  exact\n' } });
        await wrapper.get('button').trigger('click'); await flushPromises();
        expect(wrapper.get('textarea').element.value).toBe('  exact\n');
        expect(wrapper.text()).not.toContain('Copied');
        wrapper.unmount();
    });
});
describe('safe Markdown', () => {
    it('does not load external images or execute HTML and dangerous links', async () => {
        const wrapper = mount(Markdown, { props: { content: '<script>alert(1)</script>\n<img src=x onerror=alert(1)>\n\n[x](javascript:alert(1))\n\n![remote](https://example.com/track.png)\n\n[svg](data:image/svg+xml,bad)', thinking: '<img src=x onerror=alert(1)>' } });
        await flushPromises();
        expect(wrapper.findAll('script,img,svg')).toHaveLength(0);
        expect(wrapper.findAll('a').every(link => !/^(javascript|data):/.test(link.attributes('href') || ''))).toBe(true);
        expect(wrapper.get('details').attributes('open')).toBeUndefined();
        wrapper.unmount();
    });
    it('copies code from parser text with indentation and trailing newline', async () => {
        const writeText = vi.fn().mockResolvedValue(undefined);
        vi.stubGlobal('navigator', { clipboard: { writeText } });
        const wrapper = mount(Markdown, { props: { content: '```js\n  const x = 1;\n    x++;\n```' } });
        await flushPromises();
        await vi.waitFor(() => expect(wrapper.find('button').exists()).toBe(true));
        await wrapper.get('button').trigger('click');
        expect(writeText).toHaveBeenCalledWith('  const x = 1;\n    x++;\n');
        wrapper.unmount();
    });
});

describe('message-relative reading anchors', () => {
    it('preserves the same visible message when content is prepended', async () => {
        const { defineComponent, shallowRef } = await import('vue');
        const { useAssistantAutoScroll } = await import('@/views/base/assistant/useAssistantAutoScroll.ts');
        const panel = document.createElement('section');
        const content = document.createElement('div'); panel.append(content);
        const item = document.createElement('article'); item.dataset['messageId'] = '42'; content.append(item);
        let top = 20;
        Object.defineProperties(panel, { scrollHeight: { value: 2000 }, clientHeight: { value: 400 }, scrollTop: { value: 500, writable: true } });
        panel.getBoundingClientRect = () => ({ top: 0, bottom: 400 } as DOMRect);
        item.getBoundingClientRect = () => ({ top, bottom: top + 100 } as DOMRect);
        let scrolling!: ReturnType<typeof useAssistantAutoScroll>;
        const wrapper = mount(defineComponent({ setup() { scrolling = useAssistantAutoScroll(shallowRef(panel), shallowRef(content)); return () => null; } }));
        panel.dispatchEvent(new Event('scroll'));
        expect(scrolling.following.value).toBe(false);
        const anchor = scrolling.capture();
        expect(anchor).toEqual({ messageId: '42', offset: 20, following: false });
        top = 320;
        scrolling.restore(anchor);
        expect(panel.scrollTop).toBe(800);
        scrolling.scrollToLatest();
        expect(scrolling.following.value).toBe(true);
        wrapper.unmount();
    });
});


describe('legacy exact amount compatibility', () => {
    it('prefers decimal strings, refuses unsafe numeric fallback and preserves hidden amounts', () => {
        const wrapper = mount(LegacyEvidence, { props: { message: { id: '1', role: 'assistant', content: '', createdAt: 1, state: 'completed', references: [
            { id: '1', time: 1, type: 2, sourceAmount: 9007199254740992, sourceAmountMinor: '9007199254740993' },
            { id: '2', time: 1, type: 2, sourceAmount: 9007199254740992 },
            { id: '3', time: 1, type: 2, sourceAmount: 9007199254740992, sourceAmountMinor: '9223372036854775807', hideAmount: true }
        ] } } });
        expect(wrapper.text()).toContain('9007199254740993');
        expect(wrapper.text()).toContain('Amount cannot be displayed precisely');
        expect(wrapper.text()).not.toContain('9007199254740992');
        expect(wrapper.text()).not.toContain('9223372036854775807');
        wrapper.unmount();
    });
});

describe('server-authoritative conversation tail', () => {
    it('does not regenerate the last visible message when the loaded window is historical', async () => {
        const { default: MessageList } = await import('./AssistantExperienceMessageList.vue');
        const { default: Actions } = await import('./AssistantExperienceActions.vue');
        const { fixtureMessage } = await import('./experienceFixture.ts');
        const wrapper = mount(MessageList, { props: { messages: [fixtureMessage], latestMessageId: '' }, global: { stubs: { AssistantMarkdownContent: true, AssistantExperienceEvidence: true } } });
        const actions = wrapper.findAllComponents(Actions).at(-1)!;
        expect(actions.props('regenerate')).toBe(false);
        expect(actions.props('branch')).toBe(true);
        await wrapper.setProps({ latestMessageId: fixtureMessage.id });
        expect(actions.props('regenerate')).toBe(true);
        expect(actions.props('branch')).toBe(false);
        wrapper.unmount();
    });
});
