import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import * as Vue from 'vue';
import { readFileSync } from 'node:fs';
import { parse, compileScript } from '@vue/compiler-sfc';
import { compile } from '@vue/compiler-dom';
import { createRenderer, defineComponent, h, nextTick, reactive, type Component } from 'vue';
import type { Router } from 'framework7/types';
import type { AIAssistantChatRequest, AIAssistantChatStreamChunk, AIAssistantQuery } from '@/models/large_language_model.ts';
import type { AIAssistantChatStreamCallbacks } from '@/lib/assistant_stream.ts';
import { ALL_LANGUAGES } from '@/locales/index.ts';

const controls = vi.hoisted(() => ({ leave: undefined as (() => void) | undefined, open: vi.fn(), toast: vi.fn() }));
vi.mock('vue-router', () => ({ onBeforeRouteLeave: (callback: () => void) => { controls.leave = callback; } }));
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
import AssistantMessageEvidence from '@/components/common/AssistantMessageEvidence.vue';
import AssistantMarkdownContent from '@/components/common/AssistantMarkdownContent.vue';

// Vitest's Node plugin emits SSR templates. Compile the actual SFC templates for
// Vue's custom host, keeping their imported setup/composable code and sanitizer.
for (const [component, file] of [[DesktopPage, 'src/views/desktop/assistant/AssistantPage.vue'], [MobilePage, 'src/views/mobile/assistant/AssistantPage.vue'], [AssistantMessageEvidence, 'src/components/common/AssistantMessageEvidence.vue'], [AssistantMarkdownContent, 'src/components/common/AssistantMarkdownContent.vue']] as const) {
    const { descriptor } = parse(readFileSync(file, 'utf8'), { filename: file });
    const bindings = compileScript(descriptor, { id: file }).bindings;
    const { code } = compile(descriptor.template!.content, { mode: 'function', prefixIdentifiers: true, bindingMetadata: bindings });
    component.render = new Function('Vue', code)(Vue);
}

// Vue's real render/lifecycle/events in an isolated in-memory host. No DOM package,
// network, database, source-text assertions, or replacement of the safe Markdown renderer.
interface HostNode { type: string; text: string; props: Record<string, unknown>; children: HostNode[]; parent: HostNode | null }
function node(type: string, text = ''): HostNode { return { type, text, props: {}, children: [], parent: null }; }
const renderer = createRenderer<HostNode, HostNode>({
    createElement: type => node(type), createText: text => node('#text', text), createComment: text => node('#comment', text),
    setText: (node, text) => { node.text = text; }, setElementText: (node, text) => { node.text = text; node.children = []; },
    parentNode: node => node.parent, nextSibling: node => node.parent?.children[node.parent.children.indexOf(node) + 1] ?? null,
    patchProp: (node, key, _old, value) => { node.props[key] = value; },
    insert(node, parent, anchor) {
        if (node.parent) node.parent.children.splice(node.parent.children.indexOf(node), 1);
        node.parent = parent;
        const index = anchor ? parent.children.indexOf(anchor) : -1;
        parent.children.splice(index < 0 ? parent.children.length : index, 0, node);
    },
    remove(node) { if (node.parent) node.parent.children.splice(node.parent.children.indexOf(node), 1); node.parent = null; }
});
const stub = (tag = 'div') => defineComponent({
    inheritAttrs: false, props: ['disabled'], emits: ['update:modelValue', 'update:value'],
    setup(props, { attrs, slots, emit }) { return () => h(tag, { ...attrs, disabled: props.disabled, ...(tag === 'input' ? { onInput: (event: { target: { value: string } }) => { emit('update:modelValue', event.target.value); emit('update:value', event.target.value); } } : {}) }, Object.values(slots).flatMap(slot => slot?.() ?? [])); }
});
function descendants(root: HostNode): HostNode[] { return [root, ...root.children.flatMap(descendants)]; }
function text(root: HostNode): string { return (root.type === '#comment' ? '' : root.text) + root.children.map(text).join(' '); }
function button(root: HostNode, label: string): HostNode { const found = descendants(root).find(node => node.type === 'button' && text(node).trim() === label); expect(found, `Button ${label}`).toBeDefined(); return found!; }
async function click(root: HostNode, label: string): Promise<void> {
    const target = button(root, label); expect(target.props['disabled']).toBeFalsy();
    (target.props['onClick'] as () => void)(); await nextTick();
}
async function flush(): Promise<void> { for (let i = 0; i < 6; i++) { await new Promise(resolve => setTimeout(resolve, 0)); await nextTick(); } }
type Pending = { req: AIAssistantChatRequest; cancelableUuid?: string; callbacks?: AIAssistantChatStreamCallbacks; resolve: () => void; reject: (error: unknown) => void };
let requests: Pending[];
let user: { currentUserBasicInfo: { username: string } | null };
let unmount: (() => void) | undefined;
const candidate = { tool: 'query_transactions', cursor: 'DO-NOT-RENDER-CURSOR' } as const;
const query: AIAssistantQuery = {
    queryId: 'q1', tool: 'aggregate_transactions', coverage: 'all_matching', queriedAt: '2026-10-14T14:03:12+08:00', warnings: [],
    scope: { startTime: '2020-01-01T00:00:00+08:00', endTime: '2021-01-01T00:00:00+08:00', timezone: 'Asia/Shanghai', currency: 'CNY', includesHidden: true, excludesDeleted: true, accountIds: ['1'], categoryIds: ['2'], tagIds: ['3'], tagMatch: 'all' }
};
function finish(chunk: Partial<AIAssistantChatStreamChunk> = {}): void { const req = requests.at(-1)!; req.callbacks?.onDone?.({ type: 'done', reply: 'Verified answer', queries: [query], ...chunk }); req.resolve(); }
function mount(page: Component): { root: HostNode; navigate: ReturnType<typeof vi.fn> } {
    const root = node('root'); const navigate = vi.fn();
    const app = renderer.createApp(page, page === MobilePage ? { f7router: { navigate } as unknown as Router.Router } : {});
    for (const name of ['main-page-layout','v-row','v-col','v-card','v-card-text','v-card-subtitle','v-divider','v-alert','v-chip','v-icon','f7-page','f7-navbar','f7-block','f7-icon','f7-preloader']) app.component(name, stub());
    for (const name of ['v-btn','f7-button']) app.component(name, stub('button'));
    for (const name of ['v-textarea','f7-input']) app.component(name, stub('input'));
    app.provide(Vue.ssrContextKey, {});
    app.mount(root); unmount = () => app.unmount(); return { root, navigate };
}
beforeEach(() => {
    requests = []; user = reactive({ currentUserBasicInfo: { username: 'alice' } }); controls.leave = undefined;
    vi.stubGlobal('HTMLElement', class {});
    controls.open.mockResolvedValue(undefined);
    vi.mocked(useUserStore).mockReturnValue(user as ReturnType<typeof useUserStore>);
    vi.mocked(services.chatWithAIAssistantStream).mockImplementation(args => new Promise<void>((resolve, reject) => requests.push({ ...args, resolve, reject })));
});
afterEach(() => { unmount?.(); unmount = undefined; vi.unstubAllGlobals(); });

describe.each([{ name: 'desktop', page: DesktopPage }, { name: 'mobile', page: MobilePage }])('$name assistant page', ({ name, page }) => {
    it('renders progress, true range, coverage, read time, and explicitly requests the next page', async () => {
        const { root } = mount(page);
        await click(root, 'Generate AI Summary');
        requests[0]!.callbacks?.onStatus?.('aggregating'); await nextTick();
        expect(text(root)).toContain('Calculating matching transactions');
        finish({ continuation: candidate }); await flush();
        for (const expected of ['Answer completed','Query Evidence','All matching records','2020-01-01','2021-01-01','End time excluded','Asia/Shanghai','Hidden records included','Deleted records excluded','Selected accounts: 1','Selected categories: 1','All selected tags: 1','Read at','2026-10-14T14:03:12+08:00']) expect(text(root)).toContain(expected);
        expect(text(root)).not.toContain(candidate.cursor); expect(requests).toHaveLength(1);
        await click(root, 'Next Page');
        expect(requests[1]!.req).toMatchObject({ message: 'Show the next page of the current query results', continuation: candidate });
        expect(descendants(root).some(node => node.type === 'button' && text(node).trim() === 'Next Page')).toBe(false);
        finish({ queries: [{ ...query, tool: 'query_transactions', coverage: 'page' }] }); await flush();
        expect(text(root)).toContain('Current page only');
    });

    it('shows incomplete and failed outcomes with retained evidence and continuation', async () => {
        const { root } = mount(page); await click(root, 'Generate AI Summary');
        requests[0]!.callbacks?.onReplyDelta?.('Partial evidence');
        requests[0]!.reject({ response: { data: { errorMessage: 'budget exhausted', context: { assistant: { state: 'incomplete', queries: [query], continuation: candidate } } } } }); await flush();
        expect(text(root)).toContain('Answer incomplete'); expect(text(root)).toContain('Partial text may be unreliable'); expect(text(root)).toContain('Query Evidence');
        await click(root, 'Next Page'); requests[1]!.reject({ response: { data: { errorMessage: 'failure', context: { assistant: { state: 'failed', queries: [] } } } } }); await flush();
        expect(text(root)).toContain('Assistant request failed');
    });

    it('cancels, keeps the partial answer visibly canceled, and permits a new question', async () => {
        const { root } = mount(page); await click(root, 'Generate AI Summary');
        requests[0]!.callbacks?.onReplyDelta?.('Canceled partial'); await nextTick();
        await click(root, 'Cancel');
        expect(text(root)).toContain('Assistant request canceled'); expect(services.cancelRequest).toHaveBeenCalledWith(requests[0]!.cancelableUuid);
        finish({ continuation: candidate }); await flush(); expect(text(root)).toContain('Assistant request canceled');
        const input = descendants(root).find(node => node.type === 'input')!;
        (input.props['onInput'] as (event: unknown) => void)({ target: { value: 'A fresh question' } }); await nextTick();
        await click(root, 'Send'); expect(requests[1]!.req).toMatchObject({ message: 'A fresh question', history: [] });
        expect(requests[1]!.req.continuation).toBeUndefined(); finish(); await flush();
    });

    it('masks both reference amounts and opens only a validated local detail target', async () => {
        const { root, navigate } = mount(page); await click(root, 'Generate AI Summary');
        finish({ references: [{ id: '17', time: 1, type: 4, sourceAmount: 12345, destinationAmount: 67890, sourceAccountName: 'A', destinationAccountName: 'B', currency: 'CNY', destinationCurrency: 'USD', hideAmount: true }, { id: 'javascript:alert(1)', time: 1, type: 2, sourceAmount: 12345 }] }); await flush();
        const reference = descendants(root).find(node => node.type === 'button' && String(node.props['class']).includes('assistant-reference-item'))!;
        expect(text(reference).replace(/\s+/g, ' ')).toContain('*** → ***'); expect(text(reference)).not.toContain('12345'); expect(text(reference)).not.toContain('67890');
        expect(text(root)).not.toContain('javascript:alert');
        (reference.props['onClick'] as () => void)(); await nextTick();
        if (name === 'desktop') expect(controls.open).toHaveBeenCalledWith({ id: '17' });
        else expect(navigate).toHaveBeenCalledWith('/transaction/detail?id=17');
    });

    it('does not render a destination amount on ordinary references with a zero compatibility field', async () => {
        const { root } = mount(page); await click(root, 'Generate AI Summary');
        finish({ references: [{ id: '18', time: 1, type: 3, sourceAmount: 12345, destinationAmount: 0, currency: 'CNY' }, { id: '19', time: 1, type: 4, sourceAmount: 100, currency: 'CNY', destinationAccountName: 'B', destinationCurrency: 'USD' }] }); await flush();
        const reference = descendants(root).find(node => node.type === 'button' && String(node.props['class']).includes('assistant-reference-item'))!;
        expect(text(reference)).toContain('CNY 12345'); expect(text(reference)).not.toContain('→');
        const transfer = descendants(root).filter(node => node.type === 'button' && String(node.props['class']).includes('assistant-reference-item'))[1]!;
        expect(text(transfer)).toContain('USD 0');
    });

    it('distinguishes unlimited transaction scope from current balances, and shows explicit comparison periods', async () => {
        const { root } = mount(page); await click(root, 'Generate AI Summary');
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
        const { root } = mount(page); await click(root, 'Generate AI Summary'); finish(); await flush();
        const input = descendants(root).find(node => node.type === 'input')!;
        (input.props['onInput'] as (event: unknown) => void)({ target: { value: '那上个月呢' } }); await nextTick();
        await click(root, 'Send');
        expect(requests[1]!.req.history).toEqual([{ role: 'user', content: 'Generate AI Summary' }, { role: 'assistant', content: 'Verified answer' }]);
        finish(); await flush();
    });

    it('uses the real Markdown sanitizer for malicious reply/thinking and scope strings', async () => {
        const { root } = mount(page); await click(root, 'Generate AI Summary');
        const attack = '<script>alert(1)</script>\n\n<img src="x" onerror="alert(1)">\n\n[unsafe](javascript:alert%281%29) [safe](https://example.com)';
        finish({ reply: attack, thinking: attack, queries: [{ ...query, scope: { ...query.scope, keyword: '<img onerror="attack" />' } }] }); await flush();
        await vi.waitFor(() => expect(descendants(root).some(node => node.type === 'a' && node.props['href'] === 'https://example.com')).toBe(true), { timeout: 1000, interval: 10 });
        const nodes = descendants(root);
        expect(nodes.some(node => node.type === 'script')).toBe(false);
        expect(nodes.some(node => 'onerror' in node.props || 'onError' in node.props)).toBe(false);
        expect(nodes.some(node => node.props['href'] && String(node.props['href']).startsWith('javascript:'))).toBe(false);
        expect(nodes.some(node => node.type === 'a' && node.props['href'] === 'https://example.com')).toBe(true);
        expect(text(root)).toContain('<img onerror="attack" />');
    });

    it.each(['clear', 'user', 'route', 'unmount'] as const)('cleans and cancels on %s, with stale callbacks ignored', async action => {
        const { root } = mount(page); await click(root, 'Generate AI Summary'); finish({ continuation: candidate }); await flush();
        await click(root, 'Next Page');
        if (action === 'clear') await click(root, 'Clear Conversation');
        if (action === 'user') user.currentUserBasicInfo = { username: 'bob' };
        if (action === 'route' && name === 'desktop') controls.leave!();
        if (action === 'route' && name === 'mobile') (descendants(root).find(node => typeof node.props['onPage:beforeout'] === 'function')!.props['onPage:beforeout'] as () => void)();
        if (action === 'unmount') { unmount!(); unmount = undefined; }
        await nextTick();
        expect(services.cancelRequest).toHaveBeenCalledWith(requests[1]!.cancelableUuid);
        finish({ reply: 'STALE SECRET', continuation: candidate }); await flush();
        expect(text(root)).not.toContain('STALE SECRET'); expect(text(root)).not.toContain('Next Page');
        if (action !== 'unmount') expect(text(root)).toContain('Start a conversation');
    });
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
