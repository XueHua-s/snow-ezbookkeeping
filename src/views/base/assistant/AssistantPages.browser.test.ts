// @env browser
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render, cleanup } from 'vitest-browser-vue';
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

const api = vi.hoisted(() => Object.fromEntries(['capabilities','list','get','state','command','send','subscribe','cancel','memories','message'].map(key => [key, vi.fn()])));
vi.mock('@/lib/services/assistant_conversations.ts', () => ({ assistantConversationServices: api }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (key: string) => key, formatAmountToLocalizedNumeralsWithCurrency: (amount: unknown, currency: string) => `${currency} ${String(amount)}` }) }));
vi.mock('@/views/base/transactions/TransactionEditPageBase.ts', () => ({ TransactionEditPageType: { Transaction: 'transaction' } }));
vi.mock('@/views/desktop/transactions/list/dialogs/EditDialog.vue', () => ({ default: defineComponent({ setup(_, { expose }) { expose({ open: async () => undefined }); return () => h('div'); } }) }));
vi.mock('@/components/desktop/SnackBar.vue', () => ({ default: defineComponent({ setup: () => () => h('div') }) }));
import { createPinia, disposePinia, type Pinia } from 'pinia';
import { setupCloudAPI } from './testing/cloudFixture.ts';
import DesktopPage from '@/views/desktop/assistant/AssistantPage.vue';
import MobilePage from '@/views/mobile/assistant/AssistantPage.vue';
import { fixtureConversation, fixtureMessage } from '@/components/common/assistant/experienceFixture.ts';

Framework7.use(Framework7Vue);

let mobileStyle: HTMLStyleElement | undefined;
let pinia: Pinia;
let fixture: ReturnType<typeof setupCloudAPI>;
afterEach(() => {
    cleanup();
    disposePinia(pinia);
    mobileStyle?.remove(); mobileStyle = undefined;
});
beforeEach(() => {
    fixture = setupCloudAPI(api);
    pinia = createPinia();
});

async function openPage(platform: 'desktop' | 'mobile', initial = '/assistant?conversationId=1', waitForComposer = true): Promise<{ leave: () => Promise<void>; back: () => Promise<void>; go: (url: string) => Promise<void>; current: () => string }> {
    if (platform === 'desktop') {
        const router = createRouter({ history: createMemoryHistory(), routes: [
            { path: '/assistant', component: DesktopPage }, { path: '/away', component: defineComponent({ render: () => h('p', 'Away') }) }
        ] });
        await router.push('/away');
        await router.push(initial); await router.isReady();
        await render(RouterView, { global: {
            plugins: [pinia, router, createVuetify({ components })],
            components: { MainPageLayout: defineComponent({ setup(_, { slots }) { return () => h('main', slots['content']?.()); } }) }
        } });
        return { leave: async () => { await router.push('/away'); }, back: async () => { router.back(); }, go: async url => { await router.push(url); }, current: () => router.currentRoute.value.fullPath };
    }
    // The production desktop and mobile entries never share Framework7's global reset.
    mobileStyle = document.createElement('style');
    mobileStyle.textContent = framework7Styles;
    document.head.append(mobileStyle);
    const app = defineComponent({ setup() {
        return () => h(F7App, { theme: 'ios', routes: [
            { path: '/assistant', component: MobilePage },
            { path: '/away/', component: defineComponent({ render: () => h(F7Page, () => [h(F7Navbar, { title: 'Away' }), h('p', 'Away')]) }) }
        ] }, () => h(F7View, { main: true, url: '/away/' }));
    } });
    const screen = await render(app, { global: { plugins: [pinia, { install: registerComponents }] } });
    screen.container.style.height = '100vh';
    await expect.element(page.getByText('Away', { exact: true }).last()).toBeVisible();
    f7.views.main.router.navigate(initial, { animate: false });
    if (waitForComposer) await expect.element(page.getByRole('button', { name: 'Send', exact: true })).toBeVisible();
    return { leave: async () => { f7.views.main.router.navigate('/away/', { animate: false }); }, back: async () => { f7.views.main.router.back(undefined, { animate: false }); }, go: async url => { f7.views.main.router.navigate(url, { animate: false }); }, current: () => f7.views.main.router.currentRoute.url };
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


describe.each(['desktop', 'mobile'] as const)('%s real platform query routing', platform => {
    it('opens same-path conversation/message/version changes and returns through the router history', async () => {
        const response = (result: unknown) => ({ data: { success: true, result } });
        api['get']!.mockImplementation(async ({ id }: { id: string }) => response({
            conversation: { ...fixtureConversation, id },
            messages: [{ ...fixtureMessage, conversationId: id, content: `Question in ${id}` }], hasMore: false
        }));
        api['message']!.mockResolvedValue(response({ ...fixtureMessage.selectedAnswer!, conversationId: '2', assistantMessageId: 'version-2' }));
        const navigation = await openPage(platform);
        await expect.element(page.getByText('Question in 1', { exact: true })).toBeVisible();
        await navigation.go('/assistant?conversationId=2&messageId=2&versionId=version-2');
        await expect.element(page.getByText('Question in 2', { exact: true })).toBeVisible();
        expect(api['get']).toHaveBeenCalledWith({ id: '2', aroundMessageId: '2', limit: 30 });
        expect(api['message']).toHaveBeenCalledWith('version-2');
        await navigation.back();
        await expect.poll(navigation.current).toBe('/assistant?conversationId=1');
        await expect.element(page.getByText('Question in 1', { exact: true })).toBeVisible();
        expect(api['cancel']).not.toHaveBeenCalled();
    });
});


describe.each(['desktop', 'mobile'] as const)('%s first-send route replacement', platform => {
    it('replaces the empty route so back returns to the preceding page', async () => {
        const navigation = await openPage(platform, '/assistant?conversationId=');
        await page.getByPlaceholder('Ask your personal finance question').fill('First question');
        await page.getByRole('button', { name: 'Send', exact: true }).click();
        await expect.element(page.getByText('Verified cloud answer', { exact: true })).toBeVisible();
        await expect.poll(navigation.current).toBe('/assistant?conversationId=1');
        await navigation.back();
        await expect.element(page.getByText('Away', { exact: true }).last()).toBeVisible();
        expect(navigation.current().split('?')[0]).toMatch(/^\/away\/?$/);
        expect(api['send']).toHaveBeenCalledTimes(1);
    });
});


describe.each(['desktop', 'mobile'] as const)('%s inaccessible conversation route', platform => {
    it('replaces a 404 conversation with the explicit new-conversation route', async () => {
        api['get']!.mockRejectedValue({ response: { status: 404 } });
        const navigation = await openPage(platform, '/assistant?conversationId=missing');
        await expect.poll(navigation.current).toBe('/assistant?conversationId=');
        await expect.element(page.getByText('How can I help with your finances?', { exact: true })).toBeVisible();
    });
});


describe.each(['desktop', 'mobile'] as const)('%s initial offline recovery', platform => {
    it('keeps the store listener attached and restores composition without auto-sending', async () => {
        const connectivity = vi.spyOn(navigator, 'onLine', 'get').mockReturnValue(false);
        try {
            api['capabilities']!.mockRejectedValueOnce(new Error('offline'));
            await openPage(platform, '/assistant?conversationId=', false);
            const input = page.getByPlaceholder('Ask your personal finance question');
            await expect.poll(() => api['capabilities']!.mock.calls.length).toBe(1);
            await new Promise(requestAnimationFrame);
            connectivity.mockReturnValue(true);
            // Controlled online notification tests the registered listener; this is
            // not a native connectivity transition or a cross-tab identity event.
            window.dispatchEvent(new Event('online'));
            await expect.poll(() => api['capabilities']!.mock.calls.length).toBeGreaterThan(1);
            await input.fill('Recovered question');
            await expect.element(page.getByRole('button', { name: 'Send', exact: true })).toBeEnabled();
            expect(api['send']).not.toHaveBeenCalled();
        } finally {
            connectivity.mockRestore();
        }
    });
});


it('keeps real Framework7 controls in the viewport after repeated identity route replacements', async () => {
    await page.viewport(390, 844);
    const navigation = await openPage('mobile');
    const historyButton = page.getByRole('button', { name: 'Conversation history', exact: true });
    for (const [owner, id] of [['second-owner', '2'], ['third-owner', '3']] as const) {
        fixture.setOwner(owner);
        const detail = { ...fixture.getDetail(), conversation: { ...fixtureConversation, id } };
        api['get']!.mockImplementation(async ({ id: requested }: { id: string }) => {
            if (requested !== id) throw { response: { status: 404 } };
            return { data: { success: true, result: detail } };
        });
        api['list']!.mockResolvedValue({ data: { success: true, result: { items: [detail.conversation] } } });
        // Controlled identity notifications with a real Framework7 router/transition.
        // The application E2E separately proves native cross-tab StorageEvents.
        window.dispatchEvent(new StorageEvent('storage', { key: 'ebk_user_info' }));
        window.dispatchEvent(new StorageEvent('storage', { key: 'ebk_user_token' }));
        await expect.poll(navigation.current).toBe('/assistant?conversationId=');
        await historyButton.click();
        await page.getByRole('navigation', { name: 'Conversation history', exact: true })
            .getByRole('button', { name: new RegExp(fixtureConversation.title) }).click();
        await expect.poll(navigation.current).toBe(`/assistant?conversationId=${id}`);
        await expect.poll(() => document.querySelectorAll('.assistant-cloud-page.page-current').length).toBe(1);
        await expect.poll(() => {
            const button = document.querySelector<HTMLButtonElement>('[aria-label="Conversation history"]');
            return !!button && button.getBoundingClientRect().left >= 0;
        }).toBe(true);
    }
});
