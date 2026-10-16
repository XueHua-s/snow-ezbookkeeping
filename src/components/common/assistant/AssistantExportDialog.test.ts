// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { reactive } from 'vue';
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import ExportDialog from './AssistantExportDialog.vue';
const mocks = vi.hoisted(() => ({ store: {} as Record<string, unknown> }));
vi.mock('@/stores/assistant.ts', () => ({ useAssistantStore: () => mocks.store }));
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));
let wrapper: VueWrapper<InstanceType<typeof ExportDialog>> | undefined;
let click: ReturnType<typeof vi.spyOn>;
beforeEach(() => {
    mocks.store = reactive({
        conversationId: 'original', capabilities: { ownerScope: 'a', deletionEpoch: 1 },
        detail: { conversation: { title: 'Original' } }, viewedVersions: { question: { assistantMessageId: 'version-a' } },
        exportData: vi.fn().mockResolvedValue(new Blob(['private']))
    });
    click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);
    vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:test');
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined);
    wrapper = mount(ExportDialog, { props: { context: 'route-a' } });
});
afterEach(() => { wrapper?.unmount(); wrapper = undefined; });
describe('export dialog authority', () => {
    it.each(['owner', 'epoch', 'route'])('invalidates an open selection on %s change before submit', async boundary => {
        wrapper!.vm.open(undefined, true);
        if (boundary === 'route') await wrapper!.setProps({ context: 'route-b' });
        else {
            const capabilities = mocks.store['capabilities'] as { ownerScope: string; deletionEpoch: number };
            if (boundary === 'owner') capabilities.ownerScope = 'b';
            else capabilities.deletionEpoch++;
        }
        await wrapper!.get('form').trigger('submit');
        expect(mocks.store['exportData']).not.toHaveBeenCalled();
        expect(click).not.toHaveBeenCalled();
    });
    it.each(['cancel', 'owner', 'route', 'unmount'])('rejects a noncooperative late response after %s', async boundary => {
        let finish!: (value: Blob) => void;
        vi.mocked(mocks.store['exportData'] as () => Promise<Blob>).mockReturnValue(new Promise(resolve => { finish = resolve; }));
        wrapper!.vm.open();
        await wrapper!.get('form').trigger('submit');
        if (boundary === 'cancel') await wrapper!.findAll('button')[1]!.trigger('click');
        if (boundary === 'owner') (mocks.store['capabilities'] as { ownerScope: string }).ownerScope = 'b';
        if (boundary === 'route') await wrapper!.setProps({ context: 'route-b' });
        if (boundary === 'unmount') { wrapper!.unmount(); wrapper = undefined; }
        finish(new Blob(['private']));
        await flushPromises();
        expect(click).not.toHaveBeenCalled();
        expect(URL.createObjectURL).not.toHaveBeenCalled();
    });
    it('captures target and displayed version at opening and downloads the captured selection', async () => {
        wrapper!.vm.open();
        mocks.store['viewedVersions'] = { question: { assistantMessageId: 'version-b' } };
        await wrapper!.get('form').trigger('submit');
        await flushPromises();
        expect(mocks.store['exportData']).toHaveBeenCalledWith(expect.objectContaining({ id: 'original', viewedVersionIds: ['version-a'] }), expect.any(AbortSignal));
        expect(click).toHaveBeenCalledTimes(1);
    });
});
