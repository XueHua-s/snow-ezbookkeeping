// @env browser
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { reactive } from 'vue';
import { render } from 'vitest-browser-vue';
import { page } from 'vitest/browser';
import Workspace from './AssistantCloudWorkspace.vue';
const mocks = vi.hoisted(() => ({ store: {} as Record<string, unknown> }));
vi.mock('@/stores/assistant.ts', () => ({ useAssistantStore: () => mocks.store }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));
afterEach(() => vi.restoreAllMocks());
beforeEach(() => {
    mocks.store = reactive({
        capabilities: { ownerScope: 'a', deletionEpoch: 1, generationAvailable: true,
            limits: { inputCharacters: 1000, inputBytes: 4000, titleCharacters: 100, memoryBytes: 1000 } },
        conversationId: '', local: { pending: [], drafts: [] }, activePending: [], draft: '',
        online: true, history: { items: [] }, searchResults: { items: [] }, viewedVersions: {},
        memories: { items: [], enabled: true, memoryUseEpoch: 1, settingsVersion: 1 },
        readingPosition: vi.fn(), recoverBranchNavigation: vi.fn(), acknowledgeBranchNavigation: vi.fn(),
        initialize: vi.fn().mockResolvedValue(undefined), open: vi.fn().mockResolvedValue(undefined),
        attach: vi.fn(), detach: vi.fn(), setHistoryVisible: vi.fn(), execute: vi.fn(),
        exportData: vi.fn(), loadMemories: vi.fn().mockResolvedValue(undefined)
    });
});
describe.each(['desktop', 'mobile'] as const)('%s workspace native dialogs', variant => {
    it('restores More focus after cancellation and dismisses an old owner decision', async () => {
        await page.viewport(variant === 'desktop' ? 1280 : 390, 844);
        const screen = await render(Workspace, { props: { variant, conversationId: '', messageId: '' } });
        await page.getByLabelText('More', { exact: true }).click();
        await page.getByRole('button', { name: 'Clear all assistant data', exact: true }).click();
        await expect.element(page.getByRole('button', { name: 'Confirm', exact: true })).toBeVisible();
        await page.getByRole('button', { name: 'Cancel', exact: true }).click();
        await expect.poll(() => document.activeElement?.getAttribute('aria-label')).toBe('More');
        await page.getByLabelText('More', { exact: true }).click();
        await page.getByRole('button', { name: 'Clear all assistant data', exact: true }).click();
        (mocks.store['capabilities'] as { ownerScope: string }).ownerScope = 'b';
        await expect.poll(() => document.querySelectorAll('dialog[open]').length).toBe(0);
        expect(mocks.store['execute']).not.toHaveBeenCalled();
        screen.unmount();
    });
    it('cancels an open export when route props change and ignores its late download', async () => {
        let finish!: (value: Blob) => void;
        vi.mocked(mocks.store['exportData'] as () => Promise<Blob>).mockReturnValue(new Promise(resolve => { finish = resolve; }));
        const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);
        await page.viewport(variant === 'desktop' ? 1280 : 390, 844);
        const screen = await render(Workspace, { props: { variant, conversationId: '', messageId: '' } });
        await page.getByLabelText('More', { exact: true }).click();
        await page.getByRole('button', { name: 'Export all assistant data', exact: true }).click();
        await page.getByRole('button', { name: 'Download', exact: true }).click();
        await screen.rerender({ messageId: 'new-anchor' });
        await expect.poll(() => document.querySelectorAll('dialog[open]').length).toBe(0);
        finish(new Blob(['old owner data']));
        await new Promise(resolve => setTimeout(resolve, 0));
        expect(click).not.toHaveBeenCalled();
        click.mockRestore();
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
