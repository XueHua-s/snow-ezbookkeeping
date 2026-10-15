<script setup lang="ts">
import { TransactionEditScopeType } from '@/core/transaction.ts';
import { useI18n } from '@/locales/helpers.ts';
import { Account } from '@/models/account.ts';
import type { useUserBasicSettingPage } from '@/views/desktop/users/UserBasicSettingPage/useUserBasicSettingPage.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useUserBasicSettingPage>>;

defineProps<{
    loading: Scope['loading'];
    saving: Scope['saving'];
    allVisibleAccounts: Scope['allVisibleAccounts'];
    allVisibleCategorizedAccounts: Scope['allVisibleCategorizedAccounts'];
    allAccounts: Scope['allAccounts'];
    enableDisableOptions: Scope['enableDisableOptions'];
    allTransactionEditScopeTypes: Scope['allTransactionEditScopeTypes'];
}>();

const newProfile = defineModel<Scope['newProfile']>('newProfile', { required: true });

const { tt } = useI18n();

</script>

<template>
    <v-card-text class="my-1 pa-0 text-body-medium">
        <div class="setting-item">
            <span>{{ tt('Nickname') }}</span>
            <v-spacer/>
            <v-text-field
                class="ms-4"
                density="compact"
                type="text"
                autocomplete="nickname"
                persistent-placeholder
                max-width="400px"
                :disabled="loading || saving"
                :placeholder="tt('Your nickname')"
                v-model="newProfile.nickname"
            />
        </div>

        <div class="setting-item">
            <span>{{ tt('E-mail') }}</span>
            <v-spacer/>
            <v-text-field
                class="ms-4"
                density="compact"
                type="email"
                autocomplete="email"
                persistent-placeholder
                max-width="400px"
                :disabled="loading || saving"
                :placeholder="tt('Your email address')"
                v-model="newProfile.email"
            />
        </div>

        <div class="setting-item">
            <span>{{ tt('Default Account') }}</span>
            <v-spacer/>
            <two-column-select class="ms-4" density="compact" max-width="600px"
                               primary-key-field="id" primary-value-field="category"
                               primary-title-field="name"
                               primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="account"
                               primary-sub-items-field="accounts"
                               :primary-title-i18n="true"
                               secondary-key-field="id" secondary-value-field="id"
                               secondary-title-field="name"
                               secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="account" secondary-color-field="color"
                               :disabled="loading || saving || !allVisibleAccounts.length"
                               :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                               :placeholder="tt('Default Account')"
                               :items="allVisibleCategorizedAccounts"
                               :no-item-text="Account.findAccountNameById(allAccounts, newProfile.defaultAccountId, tt('Unspecified'))"
                               v-model="newProfile.defaultAccountId">
            </two-column-select>
        </div>

        <div class="setting-item">
            <span>{{ tt('Use Last Reconciled Time') }}</span>
            <v-spacer/>
            <v-select
                class="ms-4"
                density="compact"
                item-title="displayName"
                item-value="value"
                persistent-placeholder
                :disabled="loading || saving"
                max-width="400px"
                :placeholder="tt('Use Last Reconciled Time')"
                :items="enableDisableOptions"
                v-model="newProfile.useLastReconciledTime"
            >
                <template #item="{ props, internalItem }">
                    <v-list-item :disabled="!internalItem.raw.value && (TransactionEditScopeType.valueOf(newProfile.transactionEditScope)?.needLastReconciledTime ?? false)" v-bind="props">
                        <template #title>
                            <div class="text-truncate">{{ internalItem.raw.displayName }}</div>
                        </template>
                    </v-list-item>
                </template>
            </v-select>
        </div>

        <div class="setting-item">
            <span>{{ tt('Editable Transaction Range') }}</span>
            <v-spacer/>
            <v-select
                class="ms-4"
                density="compact"
                item-title="displayName"
                item-value="type"
                persistent-placeholder
                max-width="400px"
                :disabled="loading || saving"
                :placeholder="tt('Editable Transaction Range')"
                :items="allTransactionEditScopeTypes"
                v-model="newProfile.transactionEditScope"
            />
        </div>
    </v-card-text>
</template>
