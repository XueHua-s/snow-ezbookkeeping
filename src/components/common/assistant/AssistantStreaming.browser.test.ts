// @env browser
import { expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-vue';
import Messages from './AssistantExperienceMessageList.vue';
import { fixtureMessage } from './experienceFixture.ts';
import type { AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (key: string) => key }) }));

it('shows process summary before text and updates a live answer without replacing its DOM', async () => {
    const answer: AssistantRunSnapshot = { ...fixtureMessage.selectedAnswer!, content: '', processSummary: 'Checking the requested date range', state: 'running', references: [] };
    const screen = await render(Messages, { props: { messages: [{ ...fixtureMessage, selectedAnswer: answer }], busy: true } });
    expect(document.querySelector('.assistant-process')?.textContent).toContain('Checking the requested date range');
    expect(document.querySelector('.generation-indicator')).not.toBeNull();
    await screen.rerender({ messages: [{ ...fixtureMessage, selectedAnswer: { ...answer, content: 'First portion' } }] });
    await vi.waitFor(() => expect(document.querySelector('.assistant-markdown')?.textContent).toContain('First portion'));
    const content = document.querySelector('.answer-content');
    await screen.rerender({ messages: [{ ...fixtureMessage, selectedAnswer: { ...answer, content: 'First portion, followed by more text.' } }] });
    await vi.waitFor(() => expect(document.querySelector('.assistant-markdown')?.textContent).toContain('followed by more text'));
    expect(document.querySelector('.answer-content')).toBe(content);
    await screen.rerender({ messages: [{ ...fixtureMessage, selectedAnswer: { ...answer, content: 'First portion, followed by more text.', state: 'completed' } }], busy: false });
    expect(document.querySelector('.generation-indicator')).toBeNull();
    expect(document.querySelector('.answer-complete')?.textContent).toBe('Answer completed');
    screen.unmount();
});

it.each(['canceled', 'incomplete'] as const)('keeps partial text and removes activity immediately when %s', async state => {
    const answer: AssistantRunSnapshot = { ...fixtureMessage.selectedAnswer!, content: 'Saved partial answer', state: 'running', references: [] };
    const screen = await render(Messages, { props: { messages: [{ ...fixtureMessage, selectedAnswer: answer }] } });
    await screen.rerender({ messages: [{ ...fixtureMessage, selectedAnswer: { ...answer, state } }] });
    expect(document.querySelector('.generation-indicator')).toBeNull();
    expect(document.body.textContent).toContain('This answer was not completed. Partial text may be unreliable.');
    await vi.waitFor(() => expect(document.querySelector('.assistant-markdown')?.textContent).toContain('Saved partial answer'));
    screen.unmount();
});

it.each(['context_limit', 'execution_limit'] as const)('keeps an empty %s attempt actionable without claiming partial text was shown', async reasonCode => {
    const answer: AssistantRunSnapshot = { ...fixtureMessage.selectedAnswer!, content: '', processSummary: '', state: 'incomplete', reasonCode, references: [] };
    const screen = await render(Messages, { props: { messages: [{ ...fixtureMessage, selectedAnswer: answer }] } });
    expect(document.body.textContent).toContain('Your question is saved; you can retry here.');
    expect(document.body.textContent).not.toContain('Partial text may be unreliable.');
    expect(document.body.textContent).not.toContain('Generation timed out.');
    expect(document.body.textContent).not.toContain('The model service is unavailable.');
    expect(document.querySelector('.generation-indicator')).toBeNull();
    screen.unmount();
});
