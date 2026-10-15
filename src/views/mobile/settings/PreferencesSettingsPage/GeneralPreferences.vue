<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
import type { usePreferencesSettingsPage } from '@/views/mobile/settings/PreferencesSettingsPage/usePreferencesSettingsPage.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof usePreferencesSettingsPage>>;

defineProps<{
    accountCategorysDisplayOrderContent: Scope['accountCategorysDisplayOrderContent'];
    chartColorSchemeContent: Scope['chartColorSchemeContent'];
}>();

const showAccountBalance = defineModel<Scope['showAccountBalance']>('showAccountBalance', { required: true });
const isAutoUpdateExchangeRatesData = defineModel<Scope['isAutoUpdateExchangeRatesData']>('isAutoUpdateExchangeRatesData', { required: true });

const { tt } = useI18n();

</script>

<template>
    <f7-list strong inset dividers class="settings-list">
        <f7-list-item>
            <template #after-title>
                {{ tt('Show Account Balance') }}
            </template>
            <template #after>
                <f7-toggle :checked="showAccountBalance" @toggle:change="showAccountBalance = $event"></f7-toggle>
            </template>
        </f7-list-item>
        <f7-list-item
            class="item-truncate-after-text"
            link="/settings/account_category_display_order">
            <template #after-title>
                <div class="item-actual-title">
                    <span>{{ tt('Account Category Order') }}</span>
                </div>
            </template>
            <template #after>
                <div>{{ accountCategorysDisplayOrderContent }}</div>
            </template>
        </f7-list-item>
        <f7-list-item
            class="item-truncate-after-text"
            link="/settings/chart_color_scheme">
            <template #after-title>
                <div class="item-actual-title">
                    <span>{{ tt('Chart Color Scheme') }}</span>
                </div>
            </template>
            <template #after>
                <div>{{ chartColorSchemeContent }}</div>
            </template>
        </f7-list-item>
        <f7-list-item>
            <template #after-title>
                {{ tt('Auto-update Exchange Rates Data') }}
            </template>
            <template #after>
                <f7-toggle :checked="isAutoUpdateExchangeRatesData" @toggle:change="isAutoUpdateExchangeRatesData = $event"></f7-toggle>
            </template>
        </f7-list-item>
    </f7-list>
</template>
