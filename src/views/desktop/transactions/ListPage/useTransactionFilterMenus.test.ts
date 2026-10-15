// @vitest-environment happy-dom
import { mount } from '@vue/test-utils';
import { defineComponent, h, nextTick, ref } from 'vue';
import { describe, expect, it, vi } from 'vitest';
import { scrollToSelectedItem } from '@/lib/ui/common.ts';
import { useTransactionFilterMenus } from './useTransactionFilterMenus.ts';

vi.mock('@/lib/ui/common.ts', () => ({ scrollToSelectedItem: vi.fn() }));

function mountMenus() {
    const filter = ref<unknown>('bt:100:200');
    const type = ref('eq');
    const first = ref(0);
    const second = ref(0);
    let menus!: ReturnType<typeof useTransactionFilterMenus>;
    const elements = new Map<string, HTMLElement>();
    const Menu = defineComponent({
        props: { name: { type: String, required: true } },
        setup(props, { expose }) {
            const contentEl = document.createElement('div');
            elements.set(props.name, contentEl);
            expose({ contentEl });
            return () => h('div');
        }
    });
    const wrapper = mount(defineComponent({
        setup() {
            menus = useTransactionFilterMenus(() => filter.value, type, first, second);
            return () => h('div', ['time', 'category', 'amount', 'account', 'tag'].map(name =>
                h(Menu, { name, ref: `${name}FilterMenu` })
            ));
        }
    }));
    return { wrapper, menus, elements, filter, type, first, second };
}

describe('transaction filter menu ownership', () => {
    it('scrolls the selected item in each mounted header menu', async () => {
        const { wrapper, menus, elements } = mountMenus();
        try {
            for (const [name, open] of [
                ['time', menus.scrollTimeMenuToSelectedItem],
                ['category', menus.scrollCategoryMenuToSelectedItem],
                ['amount', menus.scrollAmountMenuToSelectedItem],
                ['account', menus.scrollAccountMenuToSelectedItem],
                ['tag', menus.scrollTagMenuToSelectedItem]
            ] as const) {
                open(true);
                await nextTick();
                expect(scrollToSelectedItem).toHaveBeenLastCalledWith(
                    elements.get(name), 'div.v-list', 'div.v-list', 'div.v-list-item.list-item-selected'
                );
            }
        } finally {
            wrapper.unmount();
        }
    });

    it('restores amount inputs from the current query only when opened', () => {
        const { wrapper, menus, filter, type, first, second } = mountMenus();
        try {
            menus.scrollAmountMenuToSelectedItem(true);
            expect([type.value, first.value, second.value]).toEqual(['', 100, 200]);
            filter.value = 'gt:350';
            menus.scrollAmountMenuToSelectedItem(false);
            expect([first.value, second.value]).toEqual([100, 200]);
            menus.scrollAmountMenuToSelectedItem(true);
            expect([first.value, second.value]).toEqual([350, 0]);
            filter.value = undefined;
            menus.scrollAmountMenuToSelectedItem(true);
            expect([first.value, second.value]).toEqual([0, 0]);
        } finally {
            wrapper.unmount();
        }
    });
});
