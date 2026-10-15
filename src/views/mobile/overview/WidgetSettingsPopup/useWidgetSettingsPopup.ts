import type { SetupProps } from '@/lib/ui/props.ts';
import { MOBILE_OVERVIEW_WIDGET_DEFINITIONS } from '@/consts/overview_layout.ts';
import {
    type MobileOverviewWidgetLayout,
    type OverviewWidgetCustomSelectSettingItem,
    type OverviewWidgetCustomSelectSettingValue,
    type OverviewWidgetSettingContext,
    type OverviewWidgetSettingItem,
    type OverviewWidgetSettingValue
} from '@/core/overview_layout.ts';
import { isAllAccountsChecked } from '@/lib/account.ts';
import { isAllCategoriesChecked } from '@/lib/category.ts';
import { arrayItemToObjectField, isArray, isDefined, isInteger, isObjectEmpty, isString } from '@/lib/common.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { cloneWidget, getOverviewWidgetSettingContext } from '@/lib/overview_layout.ts';
import { scrollToSelectedItem } from '@/lib/ui/common.ts';
import { type Framework7Dom, openPopover, useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useI18n } from '@/locales/helpers.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useUserStore } from '@/stores/user.ts';
import { computed, nextTick, ref } from 'vue';



// Coordinates WidgetSettingsPopup state and effects; called once from the component setup.
export function useWidgetSettingsPopup(props: SetupProps<{
    modelValue: MobileOverviewWidgetLayout | null;
    show: boolean;
}>, emit: {
    (e: 'update:modelValue', value: MobileOverviewWidgetLayout): void;
    (e: 'update:show', value: boolean): void;
}) {
    const {
        tt,
        formatAmountToLocalizedNumerals,
        formatNumberToLocalizedNumerals,
        getTablePageOptions
    } = useI18n();

    const { showToast } = useI18nUIComponents();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const loading = ref<boolean>(false);

    const widget = ref<MobileOverviewWidgetLayout | null>(null);

    const currentSettingItem = ref<OverviewWidgetSettingItem | undefined>(undefined);

    const showFilterAccountsPopup = ref<boolean>(false);

    const showFilterCategoriesPopup = ref<boolean>(false);

    const showFilterTagsPopup = ref<boolean>(false);

    const showAmountSheet = ref<boolean>(false);

    const showAmountFilterPopup = ref<boolean>(false);

    const customSelectedAccountIds = ref<string[]>([]);

    const customSelectedCategoryIds = ref<string[]>([]);

    const customTagFilter = ref<string>('');

    const customAmountValue = ref<number>(0);

    const customAmountFilter = ref<string>('');

    const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

    const hasAnyAccount = computed<boolean>(() => accountsStore.allPlainAccounts.length > 0);

    const hasAnyVisibleAccount = computed<boolean>(() => accountsStore.allVisibleAccountsCount > 0);

    const hasAnyTransactionCategory = computed<boolean>(() => !isObjectEmpty(transactionCategoriesStore.allTransactionCategoriesMap));

    const hasAnyAvailableTag = computed<boolean>(() => transactionTagsStore.allAvailableTagsCount > 0);

    const supportsSettings = computed<OverviewWidgetSettingItem[]>(() => widget.value ? MOBILE_OVERVIEW_WIDGET_DEFINITIONS[widget.value.type]?.supportsSettings ?? [] : []);

    const widgetSettingsContext = computed<OverviewWidgetSettingContext>(() => getOverviewWidgetSettingContext());

    const selectedSettingOptions = computed<OverviewWidgetCustomSelectSettingValue[]>(() => currentSettingItem.value ? getSettingOptions(currentSettingItem.value) : []);

    const selectedSettingValue = computed<OverviewWidgetSettingValue | undefined>(() => currentSettingItem.value ? getSettingValue(currentSettingItem.value.settingName) : undefined);

    function isSettingDisabled(setting: OverviewWidgetSettingItem): boolean {
        if (setting.settingType === 'accountSelect' && setting.disableHiddenAccounts) {
            return !hasAnyVisibleAccount.value;
        } else if (setting.settingType === 'accountSelect' && !setting.disableHiddenAccounts) {
            return !hasAnyAccount.value;
        } else if (setting.settingType === 'categorySelect') {
            return !hasAnyTransactionCategory.value;
        } else if (setting.settingType === 'tagSelect') {
            return !hasAnyAvailableTag.value;
        }

        return false;
    }

    function getSettingOptions(setting: OverviewWidgetSettingItem): OverviewWidgetCustomSelectSettingValue[] {
        if (setting.settingType === 'itemCountSelect') {
            return getTablePageOptions(setting.itemCountValues, undefined, false, true);
        } else if (setting.settingType === 'monthSelect') {
            return setting.monthValues.map(value => ({
                name: tt('format.misc.nMonths', { n: formatNumberToLocalizedNumerals(value) }),
                value: value
            }));
        } else if (setting.settingType === 'customSelect') {
            return setting.selectValues.map(option => ({
                name: tt(option.name),
                value: option.value,
                disabled: option.disabled
            }));
        }

        return [];
    }

    function getSettingValue(settingName: string): OverviewWidgetSettingValue | undefined {
        return widget.value?.settings[settingName];
    }

    function updateSettingValue(settingName: string, value: OverviewWidgetSettingValue): void {
        if (widget.value) {
            widget.value.settings[settingName] = value;
        }
    }

    function updateColorSettingValue(settingName: string, value: { hex: string }): void {
        if (widget.value && value && value.hex) {
            widget.value.settings[settingName] = value.hex.replace(/^#/, '').substring(0, 6).toLowerCase();
        }
    }

    function getSingleSettingDisplayName(setting: OverviewWidgetSettingItem): string {
        const value = getSettingValue(setting.settingName);

        if (setting.settingType === 'accountSelect') {
            if (!isArray(value) || value.length < 1) {
                return tt('All');
            }

            return isAllAccountsChecked(accountsStore.allVisiblePlainAccounts, arrayItemToObjectField(value as string[], true)) ? tt('All') : tt('Partial');
        } else if (setting.settingType === 'categorySelect') {
            if (!isArray(value) || value.length < 1) {
                return tt('All');
            }

            return isAllCategoriesChecked(transactionCategoriesStore.allTransactionCategories, arrayItemToObjectField(value as string[], true)) ? tt('All') : tt('Partial');
        } else if (setting.settingType === 'tagSelect' || setting.settingType === 'amountFilter') {
            return value ? tt('Custom') : tt('All');
        } else if (setting.settingType === 'amount') {
            return formatAmountToLocalizedNumerals(parseBigDecimal(isInteger(value) ? value : 0), defaultCurrency.value);
        }

        return getSettingOptions(setting).find(option => option.value === value)?.name ?? '';
    }

    function isMultipleValueSelected(setting: OverviewWidgetCustomSelectSettingItem, value: string | number): boolean {
        const selectedValues = getSettingValue(setting.settingName);
        return isArray(selectedValues) && selectedValues.includes(value);
    }

    function isLastSelectedMultipleValue(setting: OverviewWidgetCustomSelectSettingItem, value: string | number): boolean {
        const selectedValues = getSettingValue(setting.settingName);
        return isArray(selectedValues) && selectedValues.length === 1 && selectedValues[0] === value;
    }

    function updateMultipleValue(setting: OverviewWidgetCustomSelectSettingItem, value: string | number, checked: boolean): void {
        const currentValue = getSettingValue(setting.settingName);
        const selectedValues = isArray(currentValue) ? [...currentValue] : [];
        const index = selectedValues.indexOf(value);

        if (checked && index < 0) {
            if (isDefined(setting.allValue) && value === setting.allValue) {
                updateSettingValue(setting.settingName, [value]);
                return;
            }

            selectedValues.push(value);
        } else if (!checked && index >= 0) {
            if (selectedValues.length <= 1) {
                return;
            }

            selectedValues.splice(index, 1);
        }

        if (isDefined(setting.allValue) && value !== setting.allValue) {
            const allValueIndex = selectedValues.indexOf(setting.allValue);

            if (allValueIndex >= 0) {
                selectedValues.splice(allValueIndex, 1);
            }
        }

        updateSettingValue(setting.settingName, selectedValues);
    }

    function openSettingSelection(setting: OverviewWidgetSettingItem, event: MouseEvent): void {
        if (loading.value) {
            return;
        }

        const value = getSettingValue(setting.settingName);
        currentSettingItem.value = setting;

        if (setting.settingType === 'accountSelect') {
            customSelectedAccountIds.value = isArray(value) ? [...value] as string[] : [];
            showFilterAccountsPopup.value = true;
        } else if (setting.settingType === 'categorySelect') {
            customSelectedCategoryIds.value = isArray(value) ? [...value] as string[] : [];
            showFilterCategoriesPopup.value = true;
        } else if (setting.settingType === 'tagSelect') {
            customTagFilter.value = isString(value) ? value : '';
            showFilterTagsPopup.value = true;
        } else if (setting.settingType === 'amount') {
            customAmountValue.value = isInteger(value) ? value : 0;
            showAmountSheet.value = true;
        } else if (setting.settingType === 'amountFilter') {
            customAmountFilter.value = isString(value) ? value : '';
            showAmountFilterPopup.value = true;
        } else {
            nextTick(() => {
                openPopover('.widget-setting-selection-popover', event.currentTarget as HTMLElement);
            });
        }
    }

    function updateSelectedSettingValue(value: string | number): void {
        if (currentSettingItem.value) {
            updateSettingValue(currentSettingItem.value.settingName, value);
        }
    }

    function updateAccountValue(): void {
        if (currentSettingItem.value && currentSettingItem.value.settingType === 'accountSelect') {
            updateSettingValue(currentSettingItem.value.settingName, customSelectedAccountIds.value);
        }

        currentSettingItem.value = undefined;
        customSelectedAccountIds.value = [];
        showFilterAccountsPopup.value = false;
    }

    function updateCategoryValue(): void {
        if (currentSettingItem.value && currentSettingItem.value.settingType === 'categorySelect') {
            updateSettingValue(currentSettingItem.value.settingName, customSelectedCategoryIds.value);
        }

        currentSettingItem.value = undefined;
        customSelectedCategoryIds.value = [];
        showFilterCategoriesPopup.value = false;
    }

    function updateTagValue(): void {
        if (currentSettingItem.value && currentSettingItem.value.settingType === 'tagSelect') {
            updateSettingValue(currentSettingItem.value.settingName, customTagFilter.value);
        }

        currentSettingItem.value = undefined;
        customTagFilter.value = '';
        showFilterTagsPopup.value = false;
    }

    function updateAmountSettingValue(value: number): void {
        if (currentSettingItem.value?.settingType === 'amount') {
            updateSettingValue(currentSettingItem.value.settingName, value);
        }

        currentSettingItem.value = undefined;
        customAmountValue.value = 0;
        showAmountSheet.value = false;
    }

    function updateAmountValue(): void {
        if (currentSettingItem.value && currentSettingItem.value.settingType === 'amountFilter') {
            updateSettingValue(currentSettingItem.value.settingName, customAmountFilter.value);
        }

        currentSettingItem.value = undefined;
        customAmountFilter.value = '';
        showAmountFilterPopup.value = false;
    }

    function confirm(): void {
        if (widget.value) {
            emit('update:modelValue', widget.value);
        }

        close();
    }

    function close(): void {
        emit('update:show', false);
    }

    function onPopoverOpen(event: { $el: Framework7Dom }): void {
        scrollToSelectedItem(event.$el[0], '.popover-inner', '.popover-inner', 'li.list-item-selected');
    }

    function onPopupOpen(): void {
        if (!props.modelValue) {
            close();
            return;
        }

        widget.value = cloneWidget(props.modelValue);
        currentSettingItem.value = undefined;
        customSelectedAccountIds.value = [];
        customSelectedCategoryIds.value = [];
        customTagFilter.value = '';
        customAmountValue.value = 0;
        customAmountFilter.value = '';

        if (MOBILE_OVERVIEW_WIDGET_DEFINITIONS[widget.value.type]) {
            const defaultSettings = MOBILE_OVERVIEW_WIDGET_DEFINITIONS[widget.value.type]?.defaultSettings ?? {};
            widget.value.settings = { ...defaultSettings, ...widget.value.settings };
        }

        const promises: Promise<unknown>[] = [];

        if (supportsSettings.value.some(setting => setting.settingType === 'accountSelect')) {
            promises.push(accountsStore.loadAllAccounts({ force: false }));
        }

        if (supportsSettings.value.some(setting => setting.settingType === 'categorySelect')) {
            promises.push(transactionCategoriesStore.loadAllCategories({ force: false }));
        }

        if (supportsSettings.value.some(setting => setting.settingType === 'tagSelect')) {
            promises.push(transactionTagsStore.loadAllTags({ force: false }));
        }

        if (promises.length > 0) {
            loading.value = true;

            Promise.all(promises).then(() => {
                loading.value = false;
            }).catch(error => {
                loading.value = false;

                if (!error.processed) {
                    showToast(error.message || error);
                }
            });
        }
    }

    function onPopupClosed(): void {
        widget.value = null;
        currentSettingItem.value = undefined;
        showFilterAccountsPopup.value = false;
        showFilterCategoriesPopup.value = false;
        showFilterTagsPopup.value = false;
        showAmountSheet.value = false;
        showAmountFilterPopup.value = false;
        customSelectedAccountIds.value = [];
        customSelectedCategoryIds.value = [];
        customTagFilter.value = '';
        customAmountValue.value = 0;
        customAmountFilter.value = '';
        close();
    }

    return {
        tt, loading, widget, showFilterAccountsPopup, showFilterCategoriesPopup, showFilterTagsPopup, showAmountSheet,
        showAmountFilterPopup, customSelectedAccountIds, customSelectedCategoryIds, customTagFilter, customAmountValue,
        customAmountFilter, defaultCurrency, supportsSettings, widgetSettingsContext, selectedSettingOptions,
        selectedSettingValue, isSettingDisabled, getSettingValue, updateSettingValue, updateColorSettingValue,
        getSingleSettingDisplayName, isMultipleValueSelected, isLastSelectedMultipleValue, updateMultipleValue,
        openSettingSelection, updateSelectedSettingValue, updateAccountValue, updateCategoryValue, updateTagValue,
        updateAmountSettingValue, updateAmountValue, confirm, close, onPopoverOpen, onPopupOpen, onPopupClosed,
    };
}
