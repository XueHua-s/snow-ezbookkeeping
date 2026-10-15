// @env browser
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-vue';
import { page } from 'vitest/browser';
import { defineComponent, h } from 'vue';
import { createMemoryHistory, createRouter, RouterView } from 'vue-router';
import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import Framework7 from 'framework7/bundle';
import Framework7Vue, { f7, f7App as F7App, f7View as F7View, f7Page as F7Page, f7Navbar as F7Navbar } from 'framework7-vue/bundle';
// Framework7 exports this function at runtime but omits it from its bundled declarations.
// @ts-expect-error Upstream declaration omission, also used by mobile-main.ts.
import { registerComponents } from 'framework7-vue/bundle';
import 'vuetify/styles';
import framework7Styles from 'framework7/css/bundle?inline';

const api = vi.hoisted(() => Object.fromEntries(['capabilities','list','get','state','command','send','subscribe','cancel','memories'].map(key => [key, vi.fn()])));
vi.mock('@/lib/services/assistant_conversations.ts', () => ({ assistantConversationServices: api }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (key: string) => key, formatAmountToLocalizedNumeralsWithCurrency: (amount: unknown, currency: string) => `${currency} ${String(amount)}` }) }));
vi.mock('@/views/base/transactions/TransactionEditPageBase.ts', () => ({ TransactionEditPageType: { Transaction: 'transaction' } }));
vi.mock('@/views/desktop/transactions/list/dialogs/EditDialog.vue', () => ({ default: defineComponent({ setup(_, { expose }) { expose({ open: async () => undefined }); return () => h('div'); } }) }));
vi.mock('@/components/desktop/SnackBar.vue', () => ({ default: defineComponent({ setup: () => () => h('div') }) }));
import { createPinia } from 'pinia';
import { setupCloudAPI } from './testing/cloudFixture.ts';
import DesktopPage from '@/views/desktop/assistant/AssistantPage.vue';
import MobilePage from '@/views/mobile/assistant/AssistantPage.vue';

Framework7.use(Framework7Vue);

let mobileStyle: HTMLStyleElement | undefined;
afterEach(() => { mobileStyle?.remove(); mobileStyle = undefined; });
beforeEach(() => {
    setupCloudAPI(api);
});

async function openPage(platform: 'desktop' | 'mobile'): Promise<{ leave: () => Promise<void>; back: () => Promise<void> }> {
    if (platform === 'desktop') {
        const router = createRouter({ history: createMemoryHistory(), routes: [
            { path: '/assistant', component: DesktopPage }, { path: '/away', component: defineComponent({ render: () => h('p', 'Away') }) }
        ] });
        await router.push('/assistant?conversationId=1'); await router.isReady();
        await render(RouterView, { global: {
            plugins: [createPinia(), router, createVuetify({ components })],
            components: { MainPageLayout: defineComponent({ setup(_, { slots }) { return () => h('main', slots['content']?.()); } }) }
        } });
        return { leave: async () => { await router.push('/away'); }, back: async () => { await router.push('/assistant?conversationId=1'); } };
    }
    // The production desktop and mobile entries never share Framework7's global reset.
    mobileStyle = document.createElement('style');
    mobileStyle.textContent = framework7Styles;
    document.head.append(mobileStyle);
    const app = defineComponent({ setup() {
        return () => h(F7App, { theme: 'ios', routes: [
            { path: '/assistant', component: MobilePage },
            { path: '/away/', component: defineComponent({ render: () => h(F7Page, () => [h(F7Navbar, { title: 'Away' }), h('p', 'Away')]) }) }
        ] }, () => h(F7View, { main: true, url: '/assistant?conversationId=1' }));
    } });
    const screen = await render(app, { global: { plugins: [createPinia(), { install: registerComponents }] } });
    screen.container.style.height = '100vh';
    await expect.element(page.getByRole('button', { name: 'Send', exact: true })).toBeVisible();
    return { leave: async () => { f7.views.main.router.navigate('/away/', { animate: false }); }, back: async () => { f7.views.main.router.back(undefined, { animate: false }); } };
}

describe.each(['desktop', 'mobile'] as const)('%s real platform cloud controls', platform => {
    it('returns from real platform navigation to the same history and draft without canceling generation', async () => {
        await page.viewport(platform === 'desktop' ? 1280 : 390, 900);
        const navigation = await openPage(platform);
        const input = page.getByPlaceholder('Ask your personal finance question');
        await expect.element(page.getByText('October dining', { exact: true })).toBeVisible();
        await input.fill('A financial question');
        await page.getByRole('button', { name: 'Send', exact: true }).click();
        await expect.element(page.getByText('Verified cloud answer', { exact: true })).toBeVisible();
        expect(api['send']).toHaveBeenCalledTimes(1);
        await input.fill('Draft to keep');
        await navigation.leave();
        await expect.element(page.getByText('Away', { exact: true }).last()).toBeVisible();
        expect(api['cancel']).not.toHaveBeenCalled();
        await navigation.back();
        await expect.element(input).toHaveValue('Draft to keep');
        await expect.element(page.getByText('Verified cloud answer', { exact: true })).toBeVisible();
        expect(api['send']).toHaveBeenCalledTimes(1);
        await page.screenshot({ path: `.vitest/assistant-cloud-${platform}.png` });
    });
});
