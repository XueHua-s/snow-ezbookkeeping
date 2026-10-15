// @env browser
import type { PendingAssistantRequest } from './types/testing.ts';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-vue';
import { page, userEvent } from 'vitest/browser';
import { defineComponent, h, reactive } from 'vue';
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

vi.mock('@/lib/services.ts', () => ({ default: { chatWithAIAssistantStream: vi.fn(), cancelRequest: vi.fn() } }));
vi.mock('@/lib/server_settings.ts', () => ({ isAIAssistantEnabled: () => true, getAIAssistantModelID: () => 'test-model' }));
vi.mock('@/stores/user.ts', () => ({ useUserStore: () => reactive({ currentUserBasicInfo: { username: 'isolated-test' } }) }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({
    tt: (key: string) => key,
    formatAmountToLocalizedNumeralsWithCurrency: (amount: unknown, currency: string) => `${currency} ${String(amount)}`
}) }));
vi.mock('@/lib/ui/mobile.ts', () => ({ useI18nUIComponents: () => ({ showToast: vi.fn() }) }));
vi.mock('@/views/base/transactions/TransactionEditPageBase.ts', () => ({ TransactionEditPageType: { Transaction: 'transaction' } }));
vi.mock('@/views/desktop/transactions/list/dialogs/EditDialog.vue', () => ({ default: defineComponent({ setup(_, { expose }) { expose({ open: async () => undefined }); return () => h('div'); } }) }));
vi.mock('@/components/desktop/SnackBar.vue', () => ({ default: defineComponent({ setup(_, { expose }) { expose({ showError: vi.fn() }); return () => h('div'); } }) }));
import services from '@/lib/services.ts';
import DesktopPage from '@/views/desktop/assistant/AssistantPage.vue';
import MobilePage from '@/views/mobile/assistant/AssistantPage.vue';

Framework7.use(Framework7Vue);
let requests: PendingAssistantRequest[];
let mobileStyle: HTMLStyleElement | undefined;
afterEach(() => { mobileStyle?.remove(); mobileStyle = undefined; });
beforeEach(() => {
    requests = [];
    vi.mocked(services.chatWithAIAssistantStream).mockImplementation(args => new Promise<void>((resolve, reject) => requests.push({ ...args, resolve, reject })));
});

async function openPage(platform: 'desktop' | 'mobile'): Promise<() => Promise<void>> {
    if (platform === 'desktop') {
        const router = createRouter({ history: createMemoryHistory(), routes: [
            { path: '/', component: DesktopPage }, { path: '/away', component: defineComponent({ render: () => h('p', 'Away') }) }
        ] });
        await router.push('/'); await router.isReady();
        await render(RouterView, { global: {
            plugins: [router, createVuetify({ components })],
            components: { MainPageLayout: defineComponent({ setup(_, { slots }) { return () => h('main', slots['content']?.()); } }) }
        } });
        return async () => { await router.push('/away'); };
    }
    // The production desktop and mobile entries never share Framework7's global reset.
    mobileStyle = document.createElement('style');
    mobileStyle.textContent = framework7Styles;
    document.head.append(mobileStyle);
    const app = defineComponent({ setup() {
        return () => h(F7App, { theme: 'ios', routes: [
            { path: '/', component: MobilePage },
            { path: '/away/', component: defineComponent({ render: () => h(F7Page, () => [h(F7Navbar, { title: 'Away' }), h('p', 'Away')]) }) }
        ] }, () => h(F7View, { main: true, url: '/' }));
    } });
    const screen = await render(app, { global: { plugins: [{ install: registerComponents }] } });
    screen.container.style.height = '100vh';
    await expect.element(page.getByRole('button', { name: 'Generate AI Summary' })).toBeVisible();
    return async () => { f7.views.main.router.navigate('/away/', { animate: false }); };
}

function finish(): void {
    const request = requests.at(-1)!;
    request.callbacks?.onDone?.({ type: 'done', mode: 'chat', reply: 'Verified answer', continuation: { tool: 'query_transactions', cursor: 'opaque' } });
    request.resolve();
}

describe.each(['desktop', 'mobile'] as const)('%s real platform controls', platform => {
    it('sends, scrolls rendered content, explicitly paginates, cancels, and disposes on real navigation', async () => {
        await page.viewport(platform === 'desktop' ? 1280 : 390, 900);
        const leave = await openPage(platform);
        const input = page.getByPlaceholder('Ask your personal finance question');
        await input.fill('A financial question');
        if (platform === 'desktop') {
            await userEvent.keyboard('{Shift>}{Enter}{/Shift}');
            expect(requests).toHaveLength(0);
            await userEvent.keyboard('{Enter}');
        } else {
            await page.getByRole('button', { name: 'Send', exact: true }).click();
        }
        await vi.waitFor(() => expect(requests).toHaveLength(1));
        requests[0]!.callbacks?.onReplyDelta?.('A long response. '.repeat(400));
        await expect.element(page.getByRole('button', { name: 'Cancel', exact: true })).toBeVisible();
        await vi.waitFor(() => {
            const panel = document.querySelector<HTMLElement>('.assistant-messages-panel')!;
            expect(panel.scrollHeight).toBeGreaterThan(panel.clientHeight);
            expect(Math.abs(panel.scrollHeight - panel.clientHeight - panel.scrollTop)).toBeLessThan(2);
        });
        finish();
        await input.fill('Draft to keep');
        await page.getByRole('button', { name: 'Next Page', exact: true }).click();
        await expect.element(input).toHaveValue('Draft to keep');
        expect(requests[1]!.req.continuation?.cursor).toBe('opaque');
        await page.getByRole('button', { name: 'Cancel', exact: true }).click();
        expect(services.cancelRequest).toHaveBeenCalledWith(requests[1]!.cancelableUuid);
        await expect.element(page.getByText('Assistant request canceled', { exact: true })).toBeVisible();
        finish();
        await page.screenshot({ path: `.vitest/assistant-${platform}.png` });
        await input.fill('Another question');
        await page.getByRole('button', { name: 'Send', exact: true }).click();
        await vi.waitFor(() => expect(requests).toHaveLength(3));
        await leave();
        await vi.waitFor(() => expect(services.cancelRequest).toHaveBeenCalledWith(requests[2]!.cancelableUuid));
        finish();
        await vi.waitFor(() => expect(document.body.textContent).not.toContain('Another question'));
    });
});
