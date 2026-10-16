// @env browser
import { afterAll, afterEach, beforeEach, expect, it, vi } from 'vitest';
import { nextTick, reactive } from 'vue';
import { config } from '@vue/test-utils';
import { render } from 'vitest-browser-vue';
import Notices from './AssistantConversationNotices.vue';
import Recovery from './AssistantRecoveryPanel.vue';
import Capacity from './AssistantCapacityNotice.vue';
import Copy from './AssistantCopyButton.vue';
import Composer from './AssistantExperienceComposer.vue';
import { fixtureConversation } from './experienceFixture.ts';

const mocks = vi.hoisted(() => ({ store: {} as Record<string, unknown> }));
vi.mock('@/stores/assistant.ts', () => ({ useAssistantStore: () => mocks.store }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));
const originalStubs = config.global.stubs;
afterAll(() => { config.global.stubs = originalStubs; });
afterEach(() => vi.restoreAllMocks());
config.global.stubs = { transition: false, 'transition-group': false };
beforeEach(() => {
    mocks.store = reactive({
        capabilities: { ownerScope: 'a', deletionEpoch: 1 }, conversationId: '1',
        detail: { conversation: { ...fixtureConversation, sourceConversationId: 'source', sourceMessageId: 'message', pendingQuerySourceMessageId: 'query', archived: true } },
        online: true, local: { pending: [], drafts: [] }, delivery: {}, draft: ''
    });
});

it.each(['cleared', 'replaced'] as const)('removes stale conversation actions immediately when %s', async change => {
    const navigate = vi.fn();
    const screen = await render(Notices, { props: { onNavigate: navigate } });
    const source = document.querySelector<HTMLButtonElement>('[aria-label="View source"]')!;
    source.click();
    expect(navigate).toHaveBeenLastCalledWith('source', 'message');
    mocks.store['conversationId'] = change === 'cleared' ? '' : '2';
    mocks.store['detail'] = change === 'cleared' ? undefined : { conversation: { ...fixtureConversation, id: '2' } };
    await nextTick();
    // No polling: assertions must run before any exit animation can finish.
    expect(source.isConnected).toBe(false);
    expect(document.body.textContent).not.toContain('Continue this query');
    expect(document.body.textContent).not.toContain('Unarchive and continue');
    screen.unmount();
});

it.each(['owner', 'epoch'] as const)('removes pending and draft text on %s changes without an exit delay', async change => {
    mocks.store['local'] = {
        pending: [{ clientRequestId: 'pending', text: 'private pending question' }],
        drafts: [{ id: 'draft', conversationId: '1', text: 'private draft question' }], reason: 'conflict'
    };
    const screen = await render(Recovery);
    expect(document.body.textContent).toContain('private pending question');
    expect(document.body.textContent).toContain('private draft question');
    mocks.store['capabilities'] = { ownerScope: change === 'owner' ? 'b' : 'a', deletionEpoch: change === 'epoch' ? 2 : 1 };
    mocks.store['local'] = { pending: [], drafts: [] };
    await nextTick();
    expect(document.body.textContent).not.toContain('private pending question');
    expect(document.body.textContent).not.toContain('private draft question');
    expect(document.querySelector('[aria-label="Restore this question"]')).toBeNull();
    screen.unmount();
});

it('removes obsolete capacity and edit actions immediately', async () => {
    const capacity = await render(Capacity, { props: { used: 90, reserved: 0, limit: 100, canContinue: true } });
    await capacity.rerender({ used: 0, canContinue: false });
    expect(document.querySelector('.capacity')).toBeNull();
    capacity.unmount();
    const composer = await render(Composer, { props: { modelValue: '', variant: 'desktop', limits: { inputCharacters: 1000, inputBytes: 4000 }, editingLabel: 'Old edit context' } });
    await composer.rerender({ editingLabel: '' });
    expect(document.body.textContent).not.toContain('Old edit context');
    expect(document.querySelector('[aria-label="Cancel"]')).toBeNull();
    composer.unmount();
});

it('clears dismissed clipboard fallback text immediately', async () => {
    vi.spyOn(navigator.clipboard, 'writeText').mockRejectedValue(new Error('denied'));
    const screen = await render(Copy, { props: { text: 'private fallback text' } });
    document.querySelector<HTMLButtonElement>('[aria-label="Copy Markdown"]')!.click();
    await vi.waitFor(() => expect(document.querySelector('textarea')?.value).toBe('private fallback text'));
    document.querySelector<HTMLButtonElement>('[aria-label="Close"]')!.click();
    await nextTick();
    expect(document.querySelector('textarea')).toBeNull();
    screen.unmount();
});
