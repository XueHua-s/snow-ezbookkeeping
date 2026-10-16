// @vitest-environment happy-dom
import { afterEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import Confirmation from './AssistantConfirmationDialog.vue';
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));
let wrapper: VueWrapper<InstanceType<typeof Confirmation>> | undefined;
afterEach(() => { wrapper?.unmount(); wrapper = undefined; });
describe('confirmation decision and focus lifecycle', () => {
    it('declines a replaced decision and resolves the replacement only on user confirmation', async () => {
        wrapper = mount(Confirmation, { props: { context: 'a' } });
        const first = wrapper.vm.confirm('First');
        const second = wrapper.vm.confirm('Second');
        expect(await first).toBe(false);
        await flushPromises();
        await wrapper.findAll('button')[1]!.trigger('click');
        expect(await second).toBe(true);
    });
    it.each(['context', 'unmount'])('declines a pending decision on %s invalidation', async boundary => {
        wrapper = mount(Confirmation, { props: { context: 'a' } });
        const decision = wrapper.vm.confirm('Private decision');
        if (boundary === 'context') await wrapper.setProps({ context: 'b' });
        else { wrapper.unmount(); wrapper = undefined; }
        expect(await decision).toBe(false);
    });
    it.each(['replacement', 'unmount'])('does not restore canceled focus after %s', async boundary => {
        const fallback = vi.fn();
        wrapper = mount(Confirmation, { props: { context: 'a', fallbackFocus: fallback } });
        const old = wrapper.vm.confirm('First');
        await flushPromises();
        wrapper.findAll('button')[0]!.element.click();
        if (boundary === 'replacement') void wrapper.vm.confirm('Second');
        else { wrapper.unmount(); wrapper = undefined; }
        expect(await old).toBe(false);
        await flushPromises();
        expect(fallback).not.toHaveBeenCalled();
    });
});
