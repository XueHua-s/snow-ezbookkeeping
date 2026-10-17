// @env browser
import { expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-vue';
import { page } from 'vitest/browser';
import { mdiChartLine, mdiWalletOutline } from '@mdi/js';
import MessageList from './AssistantExperienceMessageList.vue';
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));

it('keeps each example icon and selected question together when examples are reordered', async () => {
    const balance = { text: 'Show balances', icon: mdiWalletOutline };
    const trend = { text: 'Compare months', icon: mdiChartLine };
    const selected = vi.fn();
    const screen = await render(MessageList, { props: { messages: [], examples: [balance, trend], onExample: selected } });
    try {
        await screen.rerender({ examples: [trend, balance] });
        const button = page.getByRole('button', { name: balance.text, exact: true });
        await button.click();
        expect(selected).toHaveBeenCalledExactlyOnceWith(balance.text);
        const balanceButton = [...document.querySelectorAll<HTMLButtonElement>('.examples button')].find(element => element.textContent?.includes(balance.text));
        expect(balanceButton?.querySelector('.example-icon path')?.getAttribute('d')).toBe(balance.icon);
    } finally {
        screen.unmount();
    }
});
