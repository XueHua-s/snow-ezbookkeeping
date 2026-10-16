// @env browser
import { expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-vue';
import { page, userEvent } from 'vitest/browser';
import Actions from './AssistantExperienceActions.vue';
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (key: string) => key }) }));
it('browses answers with keyboard buttons and provides visible focus and touch targets', async () => {
    await page.viewport(390, 844);
    const browse = vi.fn();
    const adopt = vi.fn();
    const screen = await render(Actions, { props: {
        content: '', disabledReason: undefined as string | undefined, selectedVersionId: '2', selectedAttemptNumber: 2, canUseAnswer: true,
        versions: [1, 2, 3].map(n => ({ id: String(n), attemptNumber: n, state: 'completed' as const, createdAt: '' })),
        onSelectVersion: browse, onUseAnswer: adopt
    } });
    await userEvent.tab();
    const earlier = screen.getByRole('button', { name: 'Earlier answer' });
    await expect.element(earlier).toHaveFocus();
    const button = document.activeElement as HTMLElement;
    expect(button.getBoundingClientRect().height).toBeGreaterThanOrEqual(44);
    expect(button.getBoundingClientRect().width).toBeGreaterThanOrEqual(44);
    expect(getComputedStyle(button).outlineStyle).not.toBe('none');
    await userEvent.keyboard('{Enter}');
    expect(browse).toHaveBeenCalledExactlyOnceWith('1');
    expect(adopt).not.toHaveBeenCalled();
    await userEvent.tab();
    await expect.element(screen.getByRole('button', { name: 'Newer answer' })).toHaveFocus();
    await userEvent.keyboard(' ');
    expect(browse).toHaveBeenLastCalledWith('3');
    await userEvent.tab();
    await userEvent.keyboard('{Enter}');
    expect(adopt).toHaveBeenCalledTimes(1);
    await screen.rerender({ disabledReason: 'Working' });
    for (const element of document.querySelectorAll<HTMLButtonElement>('button')) expect(element.disabled).toBe(true);
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(window.innerWidth);
    screen.unmount();
});
