// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h } from 'vue';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { createPinia } from 'pinia';
import { createMemoryHistory, createRouter, RouterView } from 'vue-router';
import { setupCloudAPI } from './testing/cloudFixture.ts';
const api = vi.hoisted(() => Object.fromEntries(['capabilities','list','get','state','command','send','subscribe','cancel','memories'].map(key => [key, vi.fn()])));
vi.mock('@/lib/services/assistant_conversations.ts', () => ({ assistantConversationServices: api }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (key: string) => key, formatAmountToLocalizedNumeralsWithCurrency: (amount: unknown, currency: string) => `${currency} ${String(amount)}` }) }));
vi.mock('@/views/base/transactions/TransactionEditPageBase.ts', () => ({ TransactionEditPageType: { Transaction: 'transaction' } }));
vi.mock('@/views/desktop/transactions/list/dialogs/EditDialog.vue', () => ({ default: defineComponent({ setup(_, { expose }) { expose({ open: async () => undefined }); return () => h('div'); } }) }));
vi.mock('@/components/desktop/SnackBar.vue', () => ({ default: defineComponent({ setup: () => () => h('div') }) }));
import DesktopPage from '@/views/desktop/assistant/AssistantPage.vue';
import { useAssistantStore } from '@/stores/assistant.ts';
let wrapper: VueWrapper | undefined;
let fixture: ReturnType<typeof setupCloudAPI>;
beforeEach(() => { fixture = setupCloudAPI(api); });
afterEach(() => { wrapper?.unmount(); wrapper = undefined; document.body.innerHTML = ''; });
async function openPage() {
    const router = createRouter({ history: createMemoryHistory(), routes: [
        { path: '/assistant', component: DesktopPage }, { path: '/away', component: defineComponent({ render: () => h('p', 'Away') }) }
    ] });
    await router.push('/assistant?conversationId=1'); await router.isReady();
    const pinia = createPinia();
    wrapper = mount(RouterView, { attachTo: document.body, global: { plugins: [router, pinia], components: {
        MainPageLayout: defineComponent({ setup(_, { slots }) { return () => h('main', slots['content']?.()); } })
    } } });
    await flushPromises();
    return { router, store: useAssistantStore(pinia), root: wrapper };
}
describe('cloud assistant route lifecycle', () => {
    it('keeps cloud history, local draft and reading position through route removal and return', async () => {
        const { router, store, root } = await openPage();
        await vi.waitFor(() => expect(root.text()).toContain('October dining'));
        await root.get('textarea').setValue('Unsent follow-up'); await flushPromises();
        store.savePosition({ messageId: '3', offset: 31, following: false });
        await router.push('/away'); await flushPromises();
        expect(api['cancel']).not.toHaveBeenCalled();
        expect(store.draft).toBe('Unsent follow-up'); expect(store.readingPosition()?.offset).toBe(31);
        await router.push('/assistant?conversationId=1'); await flushPromises();
        await vi.waitFor(() => expect(root.text()).toContain('October dining')); expect(root.get('textarea').element.value).toBe('Unsent follow-up');
    });
    it('retains history and writable draft when generation is disabled', async () => {
        fixture.setGeneration(false);
        const { root } = await openPage();
        await root.get('textarea').setValue('Keep this question');
        expect(root.text()).toContain('AI assistant is disabled'); await vi.waitFor(() => expect(root.text()).toContain('October dining'));
        expect(root.findAll('button').find(button => button.text() === 'Send')?.attributes('disabled')).toBeDefined();
        expect(api['send']).not.toHaveBeenCalled();
    });
    it('submits Unicode text intact only after IME confirmation and does not send client history', async () => {
        const { root } = await openPage(); const input = root.get('textarea');
        await input.setValue('  本月支出\n'); await input.trigger('keydown', { key: 'Enter', isComposing: true });
        await input.trigger('keydown', { key: 'Enter', keyCode: 229 }); expect(api['send']).not.toHaveBeenCalled();
        await input.trigger('keydown', { key: 'Enter' }); await flushPromises();
        expect(api['send']).toHaveBeenCalledTimes(1);
        expect(api['send']!.mock.calls[0]![0]).toMatchObject({ content: '  本月支出\n', conversationId: '1', expectedContextVersion: 1 });
        expect(api['send']!.mock.calls[0]![0]).not.toHaveProperty('history');
        await vi.waitFor(() => expect(root.text()).toContain('Verified cloud answer'));
    });
});
