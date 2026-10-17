// @env browser
import { expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-vue';
import Evidence from './AssistantExperienceEvidence.vue';
import { fixtureMessage } from './experienceFixture.ts';
import type { AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';

vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({
    tt: (key: string) => key,
    formatAmountToLocalizedNumeralsWithCurrency: (value: unknown, currency: string) => `${currency} ${String(value)}`
}) }));

const warning = 'Each page reads current data; edits or backfilled transactions during pagination may change results.';
const answer: AssistantRunSnapshot = {
    ...fixtureMessage.selectedAnswer!,
    pageableQueryId: 'query-2',
    queries: ['query-1', 'query-2'].map(queryId => ({
        queryId, tool: 'query_transactions', coverage: 'page', queriedAt: '2026-10-16T14:26:58Z',
        scope: { startTime: '2026-01-01T00:00:00Z', endTime: '2026-10-16T14:26:43Z', timezone: 'UTC', includesHidden: true, excludesDeleted: true },
        warnings: [warning]
    }))
};

it('keeps bill navigation and pagination without technical query details or repeated notices', async () => {
    const openReference = vi.fn();
    const nextPage = vi.fn();
    const screen = await render(Evidence, { props: { answer, canPage: true, onOpenReference: openReference, onNextPage: nextPage } });
    expect(document.querySelector('details, summary')).toBeNull();
    expect(document.body.textContent).not.toContain('Query Evidence');
    expect(document.body.textContent).not.toContain('2026-10-16T14:26:58Z');
    expect(document.body.textContent).not.toContain('2026-01-01T00:00:00Z');
    expect(document.body.textContent).not.toContain('Current page only');
    expect(document.querySelectorAll('[role="note"]')).toHaveLength(1);
    expect(document.body.textContent).toContain('More matching records are available.');
    await screen.getByRole('button', { name: /View Details/ }).click();
    expect(openReference).toHaveBeenCalledExactlyOnceWith('12');
    await screen.getByRole('button', { name: 'Continue finding matching bills' }).click();
    expect(nextPage).toHaveBeenCalledTimes(1);
    await screen.rerender({ canPage: false });
    expect(document.body.textContent).not.toContain('Continue finding matching bills');
    expect(document.body.textContent).toContain('More matching records are available.');
    screen.unmount();
});

it('keeps hidden amounts hidden and removes stale notices when the answer changes', async () => {
    const screen = await render(Evidence, { props: { answer: {
        ...answer, references: answer.references.map(reference => ({ ...reference, hideAmount: true }))
    } } });
    expect(document.body.textContent).not.toContain('128000');
    await screen.rerender({ answer: { ...answer, queries: [], references: [], pageableQueryId: undefined } });
    expect(document.querySelector('[role="note"]')).toBeNull();
    expect(document.querySelector('button')).toBeNull();
    expect(document.body.textContent).not.toContain('More matching records are available.');
    screen.unmount();
});

it('preserves page coverage on the last page without technical evidence or a continuation', async () => {
    const screen = await render(Evidence, { props: { answer: {
        ...answer, pageableQueryId: undefined,
        queries: answer.queries.map(query => ({ ...query, warnings: [] }))
    } } });
    expect(document.body.textContent).toContain('Displayed details cover this page only.');
    expect(document.body.textContent).not.toContain('More matching records are available.');
    expect(document.querySelector('details, summary')).toBeNull();
    await screen.rerender({ answer: {
        ...answer, pageableQueryId: undefined,
        queries: answer.queries.map(query => ({ ...query, coverage: 'all_matching', warnings: [] }))
    } });
    expect(document.body.textContent).not.toContain('Displayed details cover this page only.');
    screen.unmount();
});
