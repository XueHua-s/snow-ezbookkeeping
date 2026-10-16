// @env browser
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { reactive } from 'vue';
import { render } from 'vitest-browser-vue';
import { page } from 'vitest/browser';
import Workspace from './AssistantCloudWorkspace.vue';
import { fixtureConversation } from './experienceFixture.ts';
const mocks = vi.hoisted(() => ({ store: {} as Record<string, unknown> }));
vi.mock('@/stores/assistant.ts', () => ({ useAssistantStore: () => mocks.store }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));
afterEach(() => vi.restoreAllMocks());
beforeEach(() => {
    mocks.store = reactive({
        capabilities: { ownerScope: 'a', deletionEpoch: 1, generationAvailable: true,
            limits: { inputCharacters: 1000, inputBytes: 4000, titleCharacters: 100, memoryBytes: 1000 } },
        conversationId: '', local: { pending: [], drafts: [] }, activePending: [], draft: '',
        online: true, history: { items: [fixtureConversation] }, searchResults: { items: [] }, viewedVersions: {},
        memories: { items: [], enabled: true, memoryUseEpoch: 1, settingsVersion: 1 },
        readingPosition: vi.fn(), recoverBranchNavigation: vi.fn(), acknowledgeBranchNavigation: vi.fn(),
        initialize: vi.fn().mockResolvedValue(undefined), open: vi.fn().mockResolvedValue(undefined),
        attach: vi.fn(), detach: vi.fn(), setHistoryVisible: vi.fn(), execute: vi.fn(),
        exportData: vi.fn(), loadMemories: vi.fn().mockResolvedValue(undefined)
    });
});
describe.each(['desktop', 'mobile'] as const)('%s workspace native dialogs', variant => {
    it('restores Delete focus after cancellation and dismisses an old owner decision', async () => {
        await page.viewport(variant === 'desktop' ? 1280 : 390, 844);
        const screen = await render(Workspace, { props: { variant, conversationId: '', messageId: '' } });
        await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
        await page.getByRole('button', { name: 'Delete', exact: true }).click();
        await expect.element(page.getByRole('button', { name: 'Confirm', exact: true })).toBeVisible();
        await page.getByRole('button', { name: 'Cancel', exact: true }).click();
        await expect.poll(() => document.activeElement?.getAttribute('aria-label')).toBe('Delete');
        expect(mocks.store['execute']).not.toHaveBeenCalled();
        await page.getByRole('button', { name: 'Delete', exact: true }).click();
        (mocks.store['capabilities'] as { ownerScope: string }).ownerScope = 'b';
        await new Promise(requestAnimationFrame);
        expect(document.querySelector('.cloud-workspace')?.textContent).not.toContain(fixtureConversation.title);
        await expect.poll(() => document.querySelectorAll('dialog[open]').length).toBe(0);
        expect(mocks.store['execute']).not.toHaveBeenCalled();
        screen.unmount();
    });
    it('cancels a pending conversation deletion when the route changes', async () => {
        await page.viewport(variant === 'desktop' ? 1280 : 390, 844);
        const screen = await render(Workspace, { props: { variant, conversationId: '', messageId: '' } });
        await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
        await page.getByRole('button', { name: 'Delete', exact: true }).click();
        await expect.element(page.getByRole('button', { name: 'Confirm', exact: true })).toBeVisible();
        await screen.rerender({ messageId: 'new-anchor' });
        await expect.poll(() => document.querySelectorAll('dialog[open]').length).toBe(0);
        expect(mocks.store['execute']).not.toHaveBeenCalled();
        screen.unmount();
    });
    it('omits memory, export and bulk deletion while retaining conversation deletion', async () => {
        await page.viewport(variant === 'desktop' ? 1280 : 390, 844);
        const screen = await render(Workspace, { props: { variant, conversationId: '', messageId: '' } });
        await expect.element(page.getByRole('button', { name: 'New conversation', exact: true })).toBeVisible();
        expect(document.querySelector('.toolbar-more')).toBeNull();
        expect(document.body.textContent).not.toMatch(/Long-term memory|Export|Clear all assistant data/);
        await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
        await expect.element(page.getByRole('button', { name: 'Delete', exact: true })).toBeVisible();
        expect(document.body.textContent).not.toMatch(/Long-term memory|Export|Clear all assistant data/);
        expect(mocks.store['exportData']).not.toHaveBeenCalled();
        expect(mocks.store['loadMemories']).not.toHaveBeenCalled();
        screen.unmount();
    });
});

it('uses top-layer modality across resize and releases the application on close or unmount', async () => {
    const outside = document.createElement('button');
    outside.textContent = 'Application toolbar';
    document.body.append(outside);
    await page.viewport(390, 844);
    const screen = await render(Workspace, { props: { variant: 'mobile', conversationId: '', messageId: '' } });
    try {
        await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
        await expect.poll(() => document.querySelector('.cloud-history')?.matches(':modal')).toBe(true);
        outside.focus();
        expect(document.activeElement).not.toBe(outside);
        await page.viewport(1280, 900);
        await expect.poll(() => document.querySelector('.cloud-history')?.tagName).toBe('ASIDE');
        outside.focus();
        expect(document.activeElement).toBe(outside);
        await page.viewport(390, 844);
        await expect.poll(() => document.querySelector('.cloud-history')?.matches(':modal')).toBe(true);
        outside.focus();
        expect(document.activeElement).not.toBe(outside);
        await page.getByRole('button', { name: 'Close', exact: true }).click();
        await expect.poll(() => document.activeElement?.getAttribute('aria-label')).toBe('Conversation history');
        await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
        await expect.poll(() => document.querySelector('.cloud-history')?.matches(':modal')).toBe(true);
        screen.unmount();
        outside.focus();
        expect(document.activeElement).toBe(outside);
    } finally {
        screen.unmount();
        outside.remove();
    }
});
