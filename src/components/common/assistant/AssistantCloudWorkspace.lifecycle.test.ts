// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h, reactive } from 'vue';
import { flushPromises, shallowMount, type VueWrapper } from '@vue/test-utils';
import Workspace from './AssistantCloudWorkspace.vue';
import { fixtureConversation } from './experienceFixture.ts';
const mocks = vi.hoisted(() => ({ store: {} as Record<string, unknown> }));
vi.mock('@/stores/assistant.ts', () => ({ useAssistantStore: () => mocks.store }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));
let wrapper: VueWrapper | undefined;
const restoreAnchor = vi.fn();
const messageListStub = defineComponent({ setup(_, { expose }) {
    expose({ restoreAnchor, captureAnchor: () => ({ messageId: 'answer', offset: 12, following: false }) });
    return () => h('div');
} });
function deferred() {
    let resolve!: () => void;
    const promise = new Promise<void>(done => { resolve = done; });
    return { promise, resolve };
}
beforeEach(() => {
    mocks.store = reactive({
        capabilities: undefined, conversationId: '', local: { pending: [], drafts: [] },
        online: true, history: { items: [] }, searchResults: { items: [] }, viewedVersions: {},
        loadVersions: vi.fn().mockResolvedValue({ nextCursor: 'page-two' }), loadMemories: vi.fn().mockResolvedValue(undefined), activePending: [], busy: false, draft: '',
        readingPosition: vi.fn(), recoverBranchNavigation: vi.fn(), acknowledgeBranchNavigation: vi.fn(),
        initialize: vi.fn().mockResolvedValue(undefined), open: vi.fn().mockResolvedValue(undefined),
        attach: vi.fn(), detach: vi.fn(), setHistoryVisible: vi.fn()
    });
});
afterEach(() => { wrapper?.unmount(); wrapper = undefined; });
describe('workspace mount ownership', () => {
    it.each(['initialize', 'open'])('does not attach after unmount during %s', async stage => {
        const late = deferred();
        vi.mocked(mocks.store[stage] as () => Promise<void>).mockReturnValue(late.promise);
        wrapper = shallowMount(Workspace, { props: { variant: 'desktop', conversationId: 'old' } });
        await flushPromises();
        wrapper.unmount();
        wrapper = undefined;
        late.resolve();
        await flushPromises();
        expect(mocks.store['attach']).not.toHaveBeenCalled();
        expect(mocks.store['detach']).not.toHaveBeenCalled();
        if (stage === 'initialize') expect(mocks.store['open']).not.toHaveBeenCalled();
    });
});

it('keeps memory and bulk export controls out of the workspace across owner changes', async () => {
    mocks.store['capabilities'] = { ownerScope: 'owner-a', deletionEpoch: 1, limits: { titleCharacters: 100 } };
    wrapper = shallowMount(Workspace, {
        props: { variant: 'desktop', conversationId: '' },
        global: { stubs: { AssistantWorkspaceToolbar: false, AssistantExperienceMessageList: messageListStub } }
    });
    await flushPromises();
    for (const owner of ['owner-a', 'owner-b']) {
        (mocks.store['capabilities'] as { ownerScope: string }).ownerScope = owner;
        await flushPromises();
        expect(wrapper.findAll('.cloud-toolbar button').map(button => button.attributes('aria-label')))
            .toEqual(['Conversation history', 'New conversation']);
        expect(wrapper.find('.toolbar-more').exists()).toBe(false);
        expect(wrapper.find('assistant-memory-manager-stub').exists()).toBe(false);
        expect(wrapper.text()).not.toMatch(/Long-term memory|Export all assistant data|Clear all assistant data/);
    }
    expect(mocks.store['loadMemories']).not.toHaveBeenCalled();
});

it('does not restore an old route anchor when its delayed open finishes', async () => {
    const late = deferred();
    vi.mocked(mocks.store['open'] as () => Promise<void>).mockReturnValueOnce(late.promise);
    wrapper = shallowMount(Workspace, {
        props: { variant: 'desktop', conversationId: 'old', messageId: 'old-message' },
        global: { stubs: { AssistantExperienceMessageList: messageListStub } }
    });
    await flushPromises();
    await wrapper.setProps({ conversationId: 'new', messageId: 'new-message' });
    await flushPromises();
    expect(restoreAnchor).toHaveBeenCalledTimes(1);
    expect(restoreAnchor).toHaveBeenCalledWith({ messageId: 'new-message', offset: 0, following: false });
    late.resolve();
    await flushPromises();
    expect(restoreAnchor).toHaveBeenCalledTimes(1);
});

it('resets version pagination on route change and ignores an old page completion', async () => {
    wrapper = shallowMount(Workspace, {
        props: { variant: 'desktop', conversationId: 'one' },
        global: { stubs: { AssistantExperienceMessageList: messageListStub } }
    });
    await flushPromises();
    const load = () => wrapper!.findComponent(messageListStub).vm.$emit('loadVersions', { id: 'question' });
    load();
    await flushPromises();
    load();
    await flushPromises();
    expect(mocks.store['loadVersions']).toHaveBeenLastCalledWith('question', 'page-two');
    let finish!: (value: { nextCursor: string }) => void;
    vi.mocked(mocks.store['loadVersions'] as () => Promise<{ nextCursor: string }>).mockReturnValueOnce(new Promise(resolve => { finish = resolve; }));
    load();
    await wrapper.setProps({ conversationId: 'two' });
    finish({ nextCursor: 'obsolete-page' });
    await flushPromises();
    load();
    await flushPromises();
    expect(mocks.store['loadVersions']).toHaveBeenLastCalledWith('question', undefined);
});


it.each([false, true])('keeps search filters and pagination after remount (close before debounce: %s)', async closeEarly => {
    vi.useFakeTimers();
    try {
        mocks.store['capabilities'] = { ownerScope: 'owner-a', deletionEpoch: 1, limits: { titleCharacters: 100 } };
        mocks.store['searchResults'] = { items: [{ conversationId: 'c1', messageId: 'm1', title: 'Found', snippet: 'Matched budget' }], nextCursor: 'search-page' };
        mocks.store['history'] = { items: [], nextCursor: 'history-page' };
        mocks.store['search'] = vi.fn().mockResolvedValue(undefined);
        wrapper = shallowMount(Workspace, {
            props: { variant: 'desktop' },
            global: { stubs: { AssistantConversationHistory: false, AssistantWorkspaceToolbar: false, AssistantExperienceMessageList: messageListStub } }
        });
        await flushPromises();
        const button = (label: string) => wrapper!.findAll('button').find(item => (item.attributes('aria-label') || item.text()) === label)!;
        await button('Conversation history').trigger('click');
        await wrapper.get('input[type="search"]').setValue('budget');
        await wrapper.findAll('input[type="checkbox"]').at(-1)!.setValue(true);
        if (!closeEarly) await vi.advanceTimersByTimeAsync(300);
        await button('Close').trigger('click');
        await button('Conversation history').trigger('click');
        expect((wrapper.get('input[type="search"]').element as HTMLInputElement).value).toBe('budget');
        expect((wrapper.findAll('input[type="checkbox"]').at(-1)!.element as HTMLInputElement).checked).toBe(true);
        await vi.advanceTimersByTimeAsync(300);
        expect(wrapper.text()).toContain('Matched budget');
        await button('Load more').trigger('click');
        expect(mocks.store['search']).toHaveBeenLastCalledWith('budget', 'search-page', true, false);
    } finally {
        vi.useRealTimers();
    }
});


it.each(['owner', 'epoch', 'unmount'])('discards a debounced search on %s invalidation', async boundary => {
    vi.useFakeTimers();
    try {
        mocks.store['capabilities'] = { ownerScope: 'owner-a', deletionEpoch: 1, limits: { titleCharacters: 100 } };
        mocks.store['search'] = vi.fn().mockResolvedValue(undefined);
        wrapper = shallowMount(Workspace, {
            props: { variant: 'desktop' },
            global: { stubs: { AssistantConversationHistory: false, AssistantWorkspaceToolbar: false, AssistantExperienceMessageList: messageListStub } }
        });
        await flushPromises();
        const button = (label: string) => wrapper!.findAll('button').find(item => (item.attributes('aria-label') || item.text()) === label)!;
        await button('Conversation history').trigger('click');
        await wrapper.get('input[type="search"]').setValue('old private query');
        if (boundary === 'unmount') {
            wrapper.unmount();
            wrapper = undefined;
        } else {
            const capabilities = mocks.store['capabilities'] as { ownerScope: string; deletionEpoch: number };
            if (boundary === 'owner') capabilities.ownerScope = 'owner-b';
            else capabilities.deletionEpoch++;
            await flushPromises();
            await button('Conversation history').trigger('click');
            expect((wrapper!.get('input[type="search"]').element as HTMLInputElement).value).toBe('');
        }
        await vi.advanceTimersByTimeAsync(300);
        expect(mocks.store['search']).not.toHaveBeenCalled();
    } finally {
        vi.useRealTimers();
    }
});


it('cancels history title editing with Escape before closing the history panel', async () => {
    mocks.store['capabilities'] = { ownerScope: 'owner-a', deletionEpoch: 1, limits: { titleCharacters: 100 } };
    mocks.store['history'] = { items: [fixtureConversation] };
    wrapper = shallowMount(Workspace, {
        props: { variant: 'desktop' },
        global: { stubs: { AssistantConversationHistory: false, AssistantWorkspaceToolbar: false, AssistantExperienceMessageList: messageListStub } }
    });
    await flushPromises();
    const button = (label: string) => wrapper!.findAll('button').find(item => (item.attributes('aria-label') || item.text()) === label)!;
    await button('Conversation history').trigger('click');
    await button('Rename').trigger('click');
    await wrapper.get('form input').trigger('keydown', { key: 'Escape' });
    expect(wrapper.find('form').exists()).toBe(false);
    expect(wrapper.find('input[type="search"]').exists()).toBe(true);
    await wrapper.get('input[type="search"]').trigger('keydown', { key: 'Escape' });
    expect(wrapper.find('input[type="search"]').exists()).toBe(false);
});

it.each([true, false])('refreshes a pending filter after navigation (explicit close: %s)', async closeFirst => {
    vi.useFakeTimers();
    try {
        mocks.store['capabilities'] = { ownerScope: 'owner-a', deletionEpoch: 1, limits: { titleCharacters: 100 } };
        mocks.store['searchResults'] = { items: [{ conversationId: 'c1', messageId: 'm1', title: 'Old', snippet: 'Result A' }], nextCursor: 'cursor-a' };
        mocks.store['search'] = vi.fn().mockImplementation(async (query: string) => {
            mocks.store['searchResults'] = { items: [{ conversationId: 'c2', messageId: 'm2', title: 'New', snippet: `Result ${query}` }], nextCursor: `cursor-${query}` };
        });
        wrapper = shallowMount(Workspace, {
            props: { variant: 'desktop', conversationId: 'one' },
            global: { stubs: { AssistantConversationHistory: false, AssistantWorkspaceToolbar: false, AssistantExperienceMessageList: messageListStub } }
        });
        await flushPromises();
        const button = (label: string) => wrapper!.findAll('button').find(item => (item.attributes('aria-label') || item.text()) === label)!;
        await button('Conversation history').trigger('click');
        await wrapper.get('input[type="search"]').setValue('B');
        if (closeFirst) await button('Close').trigger('click');
        await wrapper.setProps({ conversationId: 'two' });
        expect(wrapper.find('input[type="search"]').exists()).toBe(false);
        await vi.advanceTimersByTimeAsync(300);
        expect(mocks.store['search']).not.toHaveBeenCalled();
        await button('Conversation history').trigger('click');
        expect(wrapper.text()).not.toContain('Result A');
        await vi.advanceTimersByTimeAsync(300);
        expect(mocks.store['search']).toHaveBeenLastCalledWith('B', undefined, false, false);
        expect(wrapper.text()).toContain('Result B');
        await button('Load more').trigger('click');
        expect(mocks.store['search']).toHaveBeenLastCalledWith('B', 'cursor-B', false, false);
    } finally {
        vi.useRealTimers();
    }
});

it('coalesces same-tick route props before opening a conversation', async () => {
    wrapper = shallowMount(Workspace, { props: { variant: 'desktop', conversationId: 'one', messageId: 'old-message' }, global: { stubs: { AssistantExperienceMessageList: messageListStub } } });
    await flushPromises();
    vi.mocked(mocks.store['open'] as () => Promise<void>).mockClear();
    await wrapper.setProps({ conversationId: 'two', messageId: 'new-message', versionId: 'new-version' });
    await flushPromises();
    expect(mocks.store['open']).toHaveBeenCalledTimes(1);
    expect(mocks.store['open']).toHaveBeenCalledWith('two', 'new-message', 'new-version');
});

it('replaces an inaccessible route after open clears its conversation', async () => {
    mocks.store['open'] = vi.fn().mockImplementation(async () => { mocks.store['conversationId'] = ''; });
    mocks.store['conversationId'] = 'missing';
    wrapper = shallowMount(Workspace, { props: { variant: 'desktop', conversationId: 'missing' }, global: { stubs: { AssistantExperienceMessageList: messageListStub } } });
    await flushPromises();
    expect(wrapper.emitted('navigate')).toContainEqual(['', undefined, undefined, true]);
});

it('reads fresh route intent after initialization helper-to-caller microtasks', async () => {
    const late = deferred();
    mocks.store['initialize'] = vi.fn().mockReturnValue(late.promise);
    wrapper = shallowMount(Workspace, {
        props: { variant: 'desktop', conversationId: 'old', messageId: 'old-message' },
        global: { stubs: { AssistantExperienceMessageList: messageListStub } }
    });
    late.resolve();
    queueMicrotask(() => { void wrapper!.setProps({ conversationId: 'new', messageId: 'new-message' }); });
    await flushPromises();
    expect(mocks.store['open']).toHaveBeenCalledTimes(1);
    expect(mocks.store['open']).toHaveBeenCalledWith('new', 'new-message', undefined);
    expect(mocks.store['attach']).toHaveBeenCalledTimes(1);
});

it('retains recovery listeners after an initial capability failure', async () => {
    mocks.store['initialize'] = vi.fn().mockRejectedValue(new Error('offline'));
    wrapper = shallowMount(Workspace, {
        props: { variant: 'desktop', conversationId: '' },
        global: { stubs: { AssistantExperienceMessageList: messageListStub } }
    });
    await flushPromises();
    expect(mocks.store['attach']).toHaveBeenCalledTimes(1);
    wrapper.unmount();
    wrapper = undefined;
    expect(mocks.store['detach']).toHaveBeenCalledTimes(1);
});

it('acknowledges branch recovery only after the target route is rendered', async () => {
    mocks.store['recoverBranchNavigation'] = vi.fn().mockReturnValue('branch');
    mocks.store['open'] = vi.fn().mockImplementation(async (id: string) => {
        mocks.store['conversationId'] = id;
        mocks.store['detail'] = { conversation: { ...fixtureConversation, id }, messages: [] };
    });
    wrapper = shallowMount(Workspace, {
        props: { variant: 'desktop', conversationId: 'source' },
        global: { stubs: { AssistantExperienceMessageList: messageListStub } }
    });
    await flushPromises();
    expect(wrapper.emitted('navigate')).toContainEqual(['branch', undefined, undefined, true]);
    expect(mocks.store['acknowledgeBranchNavigation']).not.toHaveBeenCalled();
    await wrapper.setProps({ conversationId: 'branch' });
    await flushPromises();
    expect(mocks.store['acknowledgeBranchNavigation']).toHaveBeenCalledExactlyOnceWith('branch');
});


it('deduplicates version loads, preserves the cursor on stale results and restores the reading anchor', async () => {
    wrapper = shallowMount(Workspace, { props: { variant: 'desktop', conversationId: 'one' }, global: { stubs: { AssistantExperienceMessageList: messageListStub } } });
    await flushPromises();
    const load = () => wrapper!.findComponent(messageListStub).vm.$emit('loadVersions', { id: 'question' });
    load();
    await flushPromises();
    expect(restoreAnchor).toHaveBeenCalledWith({ messageId: 'answer', offset: 12, following: false });
    let finish!: (result: undefined) => void;
    vi.mocked(mocks.store['loadVersions'] as () => Promise<unknown>).mockReturnValueOnce(new Promise(resolve => { finish = resolve; }));
    load(); load();
    expect(mocks.store['loadVersions']).toHaveBeenCalledTimes(2);
    finish(undefined);
    await flushPromises();
    load();
    await flushPromises();
    expect(mocks.store['loadVersions']).toHaveBeenLastCalledWith('question', 'page-two');
});

it('resets version cursors atomically when the same conversation is reopened', async () => {
    mocks.store['detail'] = { conversation: fixtureConversation, messages: [] };
    wrapper = shallowMount(Workspace, { props: { variant: 'desktop', conversationId: '1' }, global: { stubs: { AssistantExperienceMessageList: messageListStub } } });
    await flushPromises();
    const load = () => wrapper!.findComponent(messageListStub).vm.$emit('loadVersions', { id: 'question' });
    load(); await flushPromises();
    load(); await flushPromises();
    expect(mocks.store['loadVersions']).toHaveBeenLastCalledWith('question', 'page-two');
    mocks.store['detail'] = undefined;
    mocks.store['detail'] = { conversation: fixtureConversation, messages: [] };
    load(); await flushPromises();
    expect(mocks.store['loadVersions']).toHaveBeenLastCalledWith('question', undefined);
});
