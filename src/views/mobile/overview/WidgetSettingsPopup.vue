<script setup lang="ts">
import { TRANSACTION_MAX_AMOUNT, TRANSACTION_MIN_AMOUNT } from '@/consts/transaction.ts';
import {
    type MobileOverviewWidgetLayout
} from '@/core/overview_layout.ts';
import { getDisplayColor } from '@/lib/color.ts';
import AccountFilterSettingsPage from '@/views/mobile/settings/AccountFilterSettingsPage.vue';
import CategoryFilterSettingsPage from '@/views/mobile/settings/CategoryFilterSettingsPage.vue';
import TransactionTagFilterSettingsPage from '@/views/mobile/settings/TransactionTagFilterSettingsPage.vue';
import TransactionAmountFilterPage from '@/views/mobile/transactions/AmountFilterPage.vue';
import { useWidgetSettingsPopup } from './WidgetSettingsPopup/useWidgetSettingsPopup.ts';

const props = defineProps<{
    modelValue: MobileOverviewWidgetLayout | null;
    show: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: MobileOverviewWidgetLayout): void;
    (e: 'update:show', value: boolean): void;
}>();

const {
    tt, loading, widget, showFilterAccountsPopup, showFilterCategoriesPopup, showFilterTagsPopup, showAmountSheet,
    showAmountFilterPopup, customSelectedAccountIds, customSelectedCategoryIds, customTagFilter, customAmountValue,
    customAmountFilter, defaultCurrency, supportsSettings, widgetSettingsContext, selectedSettingOptions,
    selectedSettingValue, isSettingDisabled, getSettingValue, updateSettingValue, updateColorSettingValue,
    getSingleSettingDisplayName, isMultipleValueSelected, isLastSelectedMultipleValue, updateMultipleValue,
    openSettingSelection, updateSelectedSettingValue, updateAccountValue, updateCategoryValue, updateTagValue,
    updateAmountSettingValue, updateAmountValue, confirm, onPopoverOpen, onPopupOpen, onPopupClosed,
} = useWidgetSettingsPopup(props, emit);

</script>

<template>
    <f7-popup push swipe-to-close :opened="show" @popup:open="onPopupOpen" @popup:closed="onPopupClosed">
        <f7-page v-show="!showFilterAccountsPopup && !showFilterCategoriesPopup && !showFilterTagsPopup && !showAmountFilterPopup">
            <f7-navbar>
                <f7-nav-left>
                    <f7-link popup-close icon-f7="xmark" :aria-label="tt('Cancel')"></f7-link>
                </f7-nav-left>
                <f7-nav-title :title="tt('Widget Settings')"></f7-nav-title>
                <f7-nav-right>
                    <f7-link icon-f7="checkmark_alt" :aria-label="tt('Apply')" @click="confirm"></f7-link>
                </f7-nav-right>
            </f7-navbar>

            <f7-list strong inset dividers class="settings-list margin-top-half" :class="{ 'disabled': loading }" >
                <template :key="`${setting.settingName}-${index}`" v-for="(setting, index) in supportsSettings">
                    <template v-if="setting.settingType === 'customSelect' && setting.multiple && (!setting.condition || setting.condition(widget?.settings))">
                        <f7-list-item group-title>
                            <small>{{ tt(setting.displayName) }}</small>
                        </f7-list-item>
                        <f7-list-item checkbox :key="`${setting.settingName}-${option.value}`"
                                      :title="tt(option.name)"
                                      :value="option.value"
                                      :checked="isMultipleValueSelected(setting, option.value)"
                                      :disabled="option.disabled?.(widget?.settings, widgetSettingsContext) || isLastSelectedMultipleValue(setting, option.value)"
                                      v-for="option in setting.selectValues"
                                      @change="updateMultipleValue(setting, option.value, $event.target.checked)"></f7-list-item>
                    </template>

                    <f7-list-item v-else-if="setting.settingType === 'switch' && (!setting.condition || setting.condition(widget?.settings))">
                        <template #after-title>
                            {{ tt(setting.displayName) }}
                        </template>
                        <template #after>
                            <f7-toggle :checked="getSettingValue(setting.settingName) as boolean"
                                       @toggle:change="updateSettingValue(setting.settingName, $event)"></f7-toggle>
                        </template>
                    </f7-list-item>

                    <f7-list-input readonly type="colorpicker" class="list-color-picker-input"
                                   :color-picker-params="{
                                       modules: ['sb-spectrum', 'hue-slider', 'hex'],
                                       hexLabel: false,
                                       hexValueEditable: true
                                   }"
                                   :value="{ hex: getDisplayColor(getSettingValue(setting.settingName) as string) }"
                                   @colorpicker:change="updateColorSettingValue(setting.settingName, $event)"
                                   v-else-if="setting.settingType === 'color' && (!setting.condition || setting.condition(widget?.settings))">
                        <template #inner-start>
                            <div class="item-actual-title">
                                <span>{{ tt(setting.displayName) }}</span>
                            </div>
                        </template>
                        <template #inner-end>
                            <div class="color-picker-display">
                                <ItemIcon icon-type="fixed-f7" icon-id="app_fill" :color="getSettingValue(setting.settingName)"></ItemIcon>
                                <f7-icon class="color-picker-chevron" f7="chevron_up_chevron_down"></f7-icon>
                            </div>
                        </template>
                    </f7-list-input>

                    <f7-list-input type="text"
                                   autocomplete="off"
                                   :label="tt(setting.displayName)"
                                   :placeholder="setting.placeholder ? tt(setting.placeholder) : undefined"
                                   :value="getSettingValue(setting.settingName) as string"
                                   @input="updateSettingValue(setting.settingName, $event.target.value)"
                                   v-else-if="setting.settingType === 'textbox' && (!setting.condition || setting.condition(widget?.settings))"></f7-list-input>

                    <f7-list-item class="item-truncate-after-text"
                                  :class="{ 'up-down-chevron': setting.settingType !== 'accountSelect' && setting.settingType !== 'categorySelect' && setting.settingType !== 'tagSelect' && setting.settingType !== 'amount' && setting.settingType !== 'amountFilter' }"
                                  :no-chevron="setting.settingType === 'amount'"
                                  link="#" :disabled="isSettingDisabled(setting)"
                                  @click="openSettingSelection(setting, $event)"
                                  v-else-if="!setting.condition || setting.condition(widget?.settings)">
                        <template #after-title>
                            <div class="item-actual-title">
                                <span>{{ tt(setting.displayName) }}</span>
                            </div>
                        </template>
                        <template #after>
                            <f7-preloader v-if="loading && (setting.settingType === 'accountSelect' || setting.settingType === 'categorySelect' || setting.settingType === 'tagSelect') && (!setting.condition || setting.condition(widget?.settings))" />
                            <div v-else>{{ getSingleSettingDisplayName(setting) }}</div>
                        </template>
                    </f7-list-item>
                </template>
            </f7-list>

            <f7-popover class="widget-setting-selection-popover" @popover:open="onPopoverOpen">
                <f7-list dividers>
                    <f7-list-item link="#" no-chevron popover-close
                                  :title="option.name"
                                  :disabled="option.disabled?.(widget?.settings, widgetSettingsContext)"
                                  :class="{ 'list-item-selected': selectedSettingValue === option.value }"
                                  :key="option.value" v-for="option in selectedSettingOptions"
                                  @click="updateSelectedSettingValue(option.value)">
                        <template #after>
                            <f7-icon class="list-item-checked-icon" f7="checkmark_alt"
                                     v-if="selectedSettingValue === option.value"></f7-icon>
                        </template>
                    </f7-list-item>
                </f7-list>
            </f7-popover>
        </f7-page>

        <account-filter-settings-page disable-hidden-account
                                      v-model:custom-selected-account-ids="customSelectedAccountIds"
                                      @save="updateAccountValue"
                                      v-if="showFilterAccountsPopup" />

        <category-filter-settings-page v-model:custom-selected-category-ids="customSelectedCategoryIds"
                                       @save="updateCategoryValue"
                                       v-if="showFilterCategoriesPopup" />

        <transaction-tag-filter-settings-page v-model:custom-tag-filter="customTagFilter"
                                              @save="updateTagValue"
                                              v-if="showFilterTagsPopup" />

        <number-pad-sheet :min-value="TRANSACTION_MIN_AMOUNT"
                          :max-value="TRANSACTION_MAX_AMOUNT"
                          :currency="defaultCurrency"
                          v-model:show="showAmountSheet"
                          v-model="customAmountValue"
                          @update:model-value="updateAmountSettingValue" />

        <transaction-amount-filter-page show-all-option
                                        v-model:custom-amount-filter="customAmountFilter"
                                        @save="updateAmountValue"
                                        v-if="showAmountFilterPopup" />
    </f7-popup>
</template>



<style>
.widget-setting-selection-popover .popover-inner {
    max-height: 350px;
    overflow-y: auto;
}
</style>
