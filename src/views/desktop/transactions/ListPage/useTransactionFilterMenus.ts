import { nextTick, useTemplateRef, type Ref } from 'vue';
import type { VMenu } from 'vuetify/components/VMenu';
import { AmountFilterType } from '@/core/numeral.ts';
import { isString } from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import { scrollToSelectedItem } from '@/lib/ui/common.ts';

// Called in the header's setup so each reference belongs to the rendered menu.
export function useTransactionFilterMenus(
    amountFilter: () => unknown,
    currentAmountFilterType: Ref<string>,
    currentAmountFilterValue1: Ref<number>,
    currentAmountFilterValue2: Ref<number>
) {
    const timeFilterMenu = useTemplateRef<VMenu>('timeFilterMenu');
    function scrollTimeMenuToSelectedItem(opened: boolean): void {
        if (opened) scrollMenuToSelectedItem(timeFilterMenu.value);
    }
    const categoryFilterMenu = useTemplateRef<VMenu>('categoryFilterMenu');
    const amountFilterMenu = useTemplateRef<VMenu>('amountFilterMenu');
    const accountFilterMenu = useTemplateRef<VMenu>('accountFilterMenu');
    const tagFilterMenu = useTemplateRef<VMenu>('tagFilterMenu');
    function scrollCategoryMenuToSelectedItem(opened: boolean): void {
        if (opened) {
            scrollMenuToSelectedItem(categoryFilterMenu.value);
        }
    }

    function scrollAmountMenuToSelectedItem(opened: boolean): void {
        if (opened) {
            currentAmountFilterType.value = '';

            let amount1 = 0, amount2 = 0;

            const filter = amountFilter();
            if (isString(filter)) {
                try {
                    const filterItems = filter.split(':');
                    const amountCount = (AmountFilterType.valueOf(filterItems[0] as string)?.paramCount ?? 0);

                    if (filterItems.length === 2 && amountCount === 1) {
                        amount1 = parseInt(filterItems[1] as string);
                    } else if (filterItems.length === 3 && amountCount === 2) {
                        amount1 = parseInt(filterItems[1] as string);
                        amount2 = parseInt(filterItems[2] as string);
                    }
                } catch (ex) {
                    logger.warn('cannot parse amount from filter value, original value is ' + amountFilter(), ex);
                }
            }

            currentAmountFilterValue1.value = amount1;
            currentAmountFilterValue2.value = amount2;

            scrollMenuToSelectedItem(amountFilterMenu.value);
        }
    }

    function scrollAccountMenuToSelectedItem(opened: boolean): void {
        if (opened) {
            scrollMenuToSelectedItem(accountFilterMenu.value);
        }
    }

    function scrollTagMenuToSelectedItem(opened: boolean): void {
        if (opened) {
            scrollMenuToSelectedItem(tagFilterMenu.value);
        }
    }

    function scrollMenuToSelectedItem(menu: VMenu | null): void {
        nextTick(() => {
            scrollToSelectedItem(menu?.contentEl, 'div.v-list', 'div.v-list', 'div.v-list-item.list-item-selected');
        });
    }

    return {
        scrollTimeMenuToSelectedItem, scrollCategoryMenuToSelectedItem, scrollAmountMenuToSelectedItem,
        scrollAccountMenuToSelectedItem, scrollTagMenuToSelectedItem
    };
}
