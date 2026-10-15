// @vitest-environment happy-dom
import type { PendingAssistantRequest } from './types/testing.ts';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h, nextTick, reactive, type Component } from 'vue';
import type { Router } from 'framework7/types';
import { mount as mountComponent, flushPromises, type VueWrapper, type DOMWrapper } from '@vue/test-utils';
import { createMemoryHistory, createRouter, RouterView } from 'vue-router';
import type { AIAssistantChatStreamChunk, AIAssistantQuery } from '@/models/large_language_model.ts';
import { ALL_LANGUAGES } from '@/locales/index.ts';

const controls = vi.hoisted(() => ({ open: vi.fn(), toast: vi.fn() }));
vi.mock('@/lib/services.ts', () => ({ default: { chatWithAIAssistantStream: vi.fn(), cancelRequest: vi.fn() } }));
vi.mock('@/lib/server_settings.ts', () => ({ isAIAssistantEnabled: () => true, getAIAssistantModelID: () => 'test-model' }));
vi.mock('@/lib/misc.ts', () => ({ generateRandomUUID: () => crypto.randomUUID() }));
vi.mock('@/stores/user.ts', () => ({ useUserStore: vi.fn() }));
vi.mock('@/lib/ui/mobile.ts', () => ({ useI18nUIComponents: () => ({ showToast: controls.toast }) }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({
    tt: (key: string, args?: Record<string, string | number>) => Object.entries(args ?? {}).reduce((text, [key, value]) => text.replace(`{${key}}`, String(value)), key),
    formatAmountToLocalizedNumeralsWithCurrency: (amount: unknown, currency: string) => `${currency} ${String(amount)}`
}) }));
vi.mock('@/views/base/transactions/TransactionEditPageBase.ts', () => ({ TransactionEditPageType: { Transaction: 'transaction' } }));
vi.mock('@/views/desktop/transactions/list/dialogs/EditDialog.vue', () => ({ default: defineComponent({ setup(_, { expose }) { expose({ open: controls.open }); return () => h('div'); } }) }));
vi.mock('@/components/desktop/SnackBar.vue', () => ({ default: defineComponent({ setup(_, { expose }) { expose({ showError: controls.toast }); return () => h('div'); } }) }));
import services from '@/lib/services.ts';
import { useUserStore } from '@/stores/user.ts';
import DesktopPage from '@/views/desktop/assistant/AssistantPage.vue';
import MobilePage from '@/views/mobile/assistant/AssistantPage.vue';
// Platform widgets are contract stubs; feature components and Markdown are mounted normally.
const stub = (tag = 'div') => defineComponent({
    inheritAttrs: false, props: ['disabled', 'modelValue', 'value'], emits: ['update:modelValue', 'update:value'],
    setup(props, { attrs, slots, emit }) {
        return () => h(tag, {
            ...attrs, disabled: props.disabled, value: props.modelValue ?? props.value,
            ...(tag === 'textarea' ? { onInput: (event: Event) => {
                const value = (event.target as HTMLTextAreaElement).value;
                emit('update:modelValue', value); emit('update:value', value);
            } } : {})
        }, tag === 'textarea' ? undefined : Object.values(slots).flatMap(slot => slot?.() ?? []));
    }
});
function text(root: { text(): string }): string { return root.text(); }
function button(root: VueWrapper, label: string): DOMWrapper<Element> {
    const found = root.findAll('button').find(button => button.text().trim() === label);
    expect(found, `Button ${label}`).toBeDefined(); return found!;
}
async function click(root: VueWrapper, label: string): Promise<void> {
    const target = button(root, label); expect(target.attributes('disabled')).toBeUndefined();
    await target.trigger('click');
}
const flush = flushPromises;
let requests: PendingAssistantRequest[];
let user: { currentUserBasicInfo: { username: string } | null };
let unmount: (() => void) | undefined;
const candidate = { tool: 'query_transactions', cursor: 'DO-NOT-RENDER-CURSOR' } as const;
const query: AIAssistantQuery = {
    queryId: 'q1', tool: 'aggregate_transactions', coverage: 'all_matching', queriedAt: '2026-10-14T14:03:12+08:00', warnings: [],
    scope: { startTime: '2020-01-01T00:00:00+08:00', endTime: '2021-01-01T00:00:00+08:00', timezone: 'Asia/Shanghai', currency: 'CNY', includesHidden: true, excludesDeleted: true, accountIds: ['1'], categoryIds: ['2'], tagIds: ['3'], tagMatch: 'all' }
};
function finish(chunk: Partial<AIAssistantChatStreamChunk> = {}): void { const req = requests.at(-1)!; req.callbacks?.onDone?.({ type: 'done', reply: 'Verified answer', queries: [query], ...chunk }); req.resolve(); }
async function mount(page: Component): Promise<{ root: VueWrapper; navigate: ReturnType<typeof vi.fn>; leave: () => Promise<void> }> {
    const navigate = vi.fn();
    const router = createRouter({ history: createMemoryHistory(), routes: [
        { path: '/', component: page, props: page === MobilePage ? { f7router: { navigate } as unknown as Router.Router } : {} },
        { path: '/away', component: defineComponent({ render: () => h('div', 'Away') }) }
    ] });
    await router.push('/'); await router.isReady();
    const components: Record<string, Component> = {};
    for (const name of ['main-page-layout','v-row','v-col','v-card','v-card-text','v-card-subtitle','v-divider','v-alert','v-chip','v-icon','f7-navbar','f7-block','f7-icon','f7-preloader']) components[name] = stub();
    for (const name of ['v-btn','f7-button']) components[name] = stub('button');
    for (const name of ['v-textarea','f7-input']) components[name] = stub('textarea');
    // Dispatch the public Framework7 lifecycle contract from a DOM interaction.
    components['f7-page'] = defineComponent({ emits: ['page:beforeout'], setup(_, { slots, emit }) {
        return () => h('div', [h('button', { onClick: () => emit('page:beforeout') }, 'Leave page'), slots['default']?.()]);
    } });
    const root = mountComponent(RouterView, { attachTo: document.body, global: { plugins: [router], components } });
    unmount = () => root.unmount();
    return { root, navigate, leave: async () => { await router.push('/away'); await flushPromises(); } };
}
beforeEach(() => {
    requests = []; user = reactive({ currentUserBasicInfo: { username: 'alice' } });
    controls.open.mockResolvedValue(undefined);
    vi.mocked(useUserStore).mockReturnValue(user as ReturnType<typeof useUserStore>);
    vi.mocked(services.chatWithAIAssistantStream).mockImplementation(args => new Promise<void>((resolve, reject) => requests.push({ ...args, resolve, reject })));
});
afterEach(() => { unmount?.(); unmount = undefined; document.body.innerHTML = ''; vi.unstubAllGlobals(); });

describe.each([{ name: 'desktop', page: DesktopPage }, { name: 'mobile', page: MobilePage }])('$name assistant page', ({ name, page }) => {
    it('renders progress, true range, coverage, read time, and explicitly requests the next page', async () => {
        const { root } = await mount(page);
        await click(root, 'Generate AI Summary');
        requests[0]!.callbacks?.onStatus?.('aggregating'); await nextTick();
        expect(text(root)).toContain('Calculating matching transactions');
        finish({ continuation: candidate }); await flush();
        for (const expected of ['Answer completed','Query Evidence','All matching records','2020-01-01','2021-01-01','End time excluded','Asia/Shanghai','Hidden records included','Deleted records excluded','Selected accounts: 1','Selected categories: 1','All selected tags: 1','Read at','2026-10-14T14:03:12+08:00']) expect(text(root)).toContain(expected);
        expect(text(root)).not.toContain(candidate.cursor); expect(requests).toHaveLength(1);
        await root.get('textarea').setValue('Unsent draft');
        await click(root, 'Next Page');
        expect((root.get('textarea').element as HTMLTextAreaElement).value).toBe('Unsent draft');
        expect(requests[1]!.req).toMatchObject({ message: 'Show the next page of the current query results', continuation: candidate });
        expect(root.findAll('button').some(button => button.text().trim() === 'Next Page')).toBe(false);
        finish({ queries: [{ ...query, tool: 'query_transactions', coverage: 'page' }] }); await flush();
        expect(text(root)).toContain('Current page only');
    });

    it('shows incomplete and failed outcomes with retained evidence and continuation', async () => {
        const { root } = await mount(page); await click(root, 'Generate AI Summary');
        requests[0]!.callbacks?.onReplyDelta?.('Partial evidence');
        requests[0]!.reject({ response: { data: { errorMessage: 'budget exhausted', context: { assistant: { state: 'incomplete', queries: [query], continuation: candidate } } } } }); await flush();
        expect(text(root)).toContain('Answer incomplete'); expect(text(root)).toContain('Partial text may be unreliable'); expect(text(root)).toContain('Query Evidence');
        await click(root, 'Next Page'); requests[1]!.reject({ response: { data: { errorMessage: 'failure', context: { assistant: { state: 'failed', queries: [] } } } } }); await flush();
        expect(text(root)).toContain('Assistant request failed');
    });

    it('cancels, keeps the partial answer visibly canceled, and permits a new question', async () => {
        const { root } = await mount(page); await click(root, 'Generate AI Summary');
        requests[0]!.callbacks?.onReplyDelta?.('Canceled partial'); await nextTick();
        await click(root, 'Cancel');
        expect(text(root)).toContain('Assistant request canceled'); expect(services.cancelRequest).toHaveBeenCalledWith(requests[0]!.cancelableUuid);
        finish({ continuation: candidate }); await flush(); expect(text(root)).toContain('Assistant request canceled');
        await root.get('textarea').setValue('A fresh question');
        await click(root, 'Send'); expect(requests[1]!.req).toMatchObject({ message: 'A fresh question', history: [] });
        expect(requests[1]!.req.continuation).toBeUndefined(); finish(); await flush();
    });

    it('masks both reference amounts and opens only a validated local detail target', async () => {
        const { root, navigate } = await mount(page); await click(root, 'Generate AI Summary');
        finish({ references: [{ id: '17', time: 1, type: 4, sourceAmount: 12345, destinationAmount: 67890, sourceAccountName: 'A', destinationAccountName: 'B', currency: 'CNY', destinationCurrency: 'USD', hideAmount: true }, { id: 'javascript:alert(1)', time: 1, type: 2, sourceAmount: 12345 }] }); await flush();
        const reference = root.get('.assistant-reference-item');
        expect(text(reference).replace(/\s+/g, ' ')).toContain('*** → ***'); expect(text(reference)).not.toContain('12345'); expect(text(reference)).not.toContain('67890');
        expect(text(root)).not.toContain('javascript:alert');
        await reference.trigger('click');
        if (name === 'desktop') expect(controls.open).toHaveBeenCalledWith({ id: '17' });
        else expect(navigate).toHaveBeenCalledWith('/transaction/detail?id=17');
    });

    it('does not render a destination amount on ordinary references with a zero compatibility field', async () => {
        const { root } = await mount(page); await click(root, 'Generate AI Summary');
        finish({ references: [{ id: '18', time: 1, type: 3, sourceAmount: 12345, destinationAmount: 0, currency: 'CNY' }, { id: '19', time: 1, type: 4, sourceAmount: 100, currency: 'CNY', destinationAccountName: 'B', destinationCurrency: 'USD' }] }); await flush();
        const reference = root.get('.assistant-reference-item');
        expect(text(reference)).toContain('CNY 12345'); expect(text(reference)).not.toContain('→');
        const transfer = root.findAll('.assistant-reference-item')[1]!;
        expect(text(transfer)).toContain('USD 0');
    });

    it('distinguishes unlimited transaction scope from current balances, and shows explicit comparison periods', async () => {
        const { root } = await mount(page); await click(root, 'Generate AI Summary');
        const allTime = { ...query.scope, startTime: null, endTime: null, accountIds: undefined };
        finish({ queries: [
            { ...query, queryId: 'all', tool: 'query_transactions', scope: allTime, coverage: 'page', warnings: ['Each page reads current data; edits or backfilled transactions during pagination may change results.'] },
            { ...query, queryId: 'balances', tool: 'get_account_balances', scope: allTime },
            { ...query, queryId: 'periods', scope: { ...query.scope, periods: [{ name: 'baseline', startTime: '2025-10-01T00:00:00+08:00', endTime: '2025-10-14T14:03:12+08:00', matchedCount: '6' }] } }
        ] }); await flush();
        expect(text(root)).toContain('No start limit'); expect(text(root)).toContain('No end limit'); expect(text(root)).toContain('All valid accounts');
        expect(text(root)).toContain('Current balances at query time'); expect(text(root)).toContain('2025-10-14T14:03:12+08:00'); expect(text(root)).toContain('Matching transactions: 6');
        expect(text(root)).toContain('edits or backfilled transactions');
    });

    it('retains a completed summary when the user follows up', async () => {
        const { root } = await mount(page); await click(root, 'Generate AI Summary'); finish(); await flush();
        await root.get('textarea').setValue('那上个月呢');
        await click(root, 'Send');
        expect(requests[1]!.req.history).toEqual([{ role: 'user', content: 'Generate AI Summary' }, { role: 'assistant', content: 'Verified answer' }]);
        finish(); await flush();
    });

    it('uses the real Markdown sanitizer for malicious reply/thinking and scope strings', async () => {
        const { root } = await mount(page); await click(root, 'Generate AI Summary');
        const attack = '<script>alert(1)</script>\n\n<img src="x" onerror="alert(1)">\n\n[unsafe](javascript:alert%281%29) [safe](https://example.com)';
        finish({ reply: attack, thinking: attack, queries: [{ ...query, scope: { ...query.scope, keyword: '<img onerror="attack" />' } }] }); await flush();
        await vi.waitFor(() => expect(root.find('a[href="https://example.com"]').exists()).toBe(true));
        expect(root.find('script').exists()).toBe(false);
        expect(root.find('[onerror]').exists()).toBe(false);
        expect(root.find('a[href^="javascript:"]').exists()).toBe(false);
        expect(text(root)).toContain('<img onerror="attack" />');
    });

    it.each(['clear', 'user', 'route', 'unmount'] as const)('cleans and cancels on %s, with stale callbacks ignored', async action => {
        const { root, leave } = await mount(page); await click(root, 'Generate AI Summary'); finish({ continuation: candidate }); await flush();
        await click(root, 'Next Page');
        if (action === 'clear') await click(root, 'Clear Conversation');
        if (action === 'user') user.currentUserBasicInfo = { username: 'bob' };
        if (action === 'route' && name === 'desktop') await leave();
        if (action === 'route' && name === 'mobile') await click(root, 'Leave page');
        if (action === 'unmount') { unmount!(); unmount = undefined; }
        await nextTick();
        expect(services.cancelRequest).toHaveBeenCalledWith(requests[1]!.cancelableUuid);
        finish({ reply: 'STALE SECRET', continuation: candidate }); await flush();
        expect(text(root)).not.toContain('STALE SECRET'); expect(text(root)).not.toContain('Next Page');
        if (action !== 'unmount' && !(action === 'route' && name === 'desktop')) expect(text(root)).toContain('Start a conversation');
    });
});

it('does not submit Enter while confirming an IME candidate', async () => {
    const { root } = await mount(DesktopPage);
    const input = root.get('textarea');
    await input.setValue('本月支出');
    await input.trigger('keydown', { key: 'Enter', isComposing: true });
    expect(requests).toHaveLength(0);
    // WebKit may signal the final composition key via keyCode 229 instead.
    await input.trigger('keydown', { key: 'Enter', keyCode: 229 });
    expect(requests).toHaveLength(0);
    await input.trigger('keydown', { key: 'Enter' });
    expect(requests).toHaveLength(1);
    expect(requests[0]!.req.message).toBe('本月支出');
    finish(); await flush();
});

it('provides every new assistant label or explicit English fallback in all supported locales', () => {
    const keys = [
        'Working on your question', 'Answer completed', 'Answer incomplete', 'Assistant request failed', 'Assistant request canceled',
        'Reading your ledger', 'Calculating matching transactions', 'Preparing your answer', 'This answer was not completed. Partial text may be unreliable.',
        'Query Evidence', 'All matching records', 'Current page only', 'End time excluded', 'Matching transactions: {count}', 'No start limit', 'No end limit',
        'Hidden records included', 'Hidden records excluded', 'Deleted records excluded', 'All valid accounts', 'Selected accounts: {count}', 'Selected categories: {count}',
        'All selected tags: {count}', 'Any selected tags: {count}', 'Transfer source side', 'Transfer destination side', 'Either transfer side',
        'Minimum amount', 'Maximum amount', 'Read at', 'Bills read during this query; individual conclusions may use other query evidence.',
        'Show the next page of the current query results', 'Each page reads current data; edits or backfilled transactions during pagination may change results.',
        'Current balances at query time', 'Next Page', 'Keyword', 'Account Balances', 'Uncategorized', 'Send'
    ];
    for (const language of Object.values(ALL_LANGUAGES)) for (const key of keys) expect((language.content as Record<string, string>)[key]).toBeTruthy();
});
