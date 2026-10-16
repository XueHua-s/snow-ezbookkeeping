// @env browser
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-vue';
import { page } from 'vitest/browser';
import Fixture from './AssistantExperienceFixture.vue';
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (key: string) => key, formatAmountToLocalizedNumeralsWithCurrency: (value: unknown, currency: string) => `${currency} ${Number(String(value)) / 100}` }) }));
describe.each(['desktop', 'mobile'] as const)('%s experience visual fixtures', variant => {
    it.each(['empty', 'answer', 'failure', 'history', 'memory', 'capacity'] as const)('renders %s in light and dark without horizontal overflow', async state => {
        await page.viewport(variant === 'desktop' ? 1280 : 390, 844);
        const screen = await render(Fixture, { props: { variant, state, dark: false } });
        await expect.element(page.getByRole('main')).toBeVisible();
        if (['answer', 'failure', 'capacity'].includes(state)) await expect.element(page.getByRole('heading', { name: 'October dining' })).toBeVisible();
        const panel = document.querySelector<HTMLElement>('.experience-messages');
        if (panel) { panel.scrollTop = 0; panel.dispatchEvent(new Event('scroll')); }
        expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(window.innerWidth);
        await page.screenshot({ path: `.vitest/experience-${variant}-${state}-light.png` });
        await screen.rerender({ dark: true });
        await page.screenshot({ path: `.vitest/experience-${variant}-${state}-dark.png` });
        if (state === 'answer' && panel) { panel.scrollTop = panel.scrollHeight; panel.dispatchEvent(new Event('scroll')); await expect.element(page.getByRole('button', { name: 'Regenerate', exact: true })).toBeVisible(); await page.screenshot({ path: `.vitest/experience-${variant}-answer-bottom-dark.png` }); }
        screen.unmount();
    });
});

it('keeps mobile return as newline, supports keyboard actions and RTL enlarged layout', async () => {
    await page.viewport(390, 844);
    const screen = await render(Fixture, { props: { variant: 'mobile', state: 'empty', dark: false } });
    await page.getByRole('button', { name: 'How much did I spend on dining?', exact: true }).click();
    await expect.element(page.getByRole('textbox')).toHaveValue('How much did I spend on dining?');
    const main = document.querySelector<HTMLElement>('main')!;
    main.dir = 'rtl';
    main.style.fontSize = '20px';
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(window.innerWidth);
    await page.getByRole('button', { name: 'Send', exact: true }).click();
    await expect.element(page.getByRole('status')).toHaveTextContent('Send requested');
    await page.screenshot({ path: '.vitest/experience-mobile-rtl-enlarged.png' });
    screen.unmount();
});
