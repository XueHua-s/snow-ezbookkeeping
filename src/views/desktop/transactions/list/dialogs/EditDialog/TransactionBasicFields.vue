<script setup lang="ts">
import { CategoryType } from '@/core/category.ts';
import { TemplateType } from '@/core/template.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    getTransactionPrimaryCategoryName,
    getTransactionSecondaryCategoryName
} from '@/lib/category.ts';
import { useI18n } from '@/locales/helpers.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import {
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { useEditDialog } from '@/views/desktop/transactions/list/dialogs/EditDialog/useEditDialog.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useEditDialog>>;

type ParentProps = {
    type: TransactionEditPageType;
    persistent?: boolean;
    show?: boolean;
};

defineProps<{
    type: ParentProps['type'];
    loading: Scope['loading'];
    submitting: Scope['submitting'];
    recognizing: Scope['recognizing'];
    sourceAmountColor: Scope['sourceAmountColor'];
    sourceAccountCurrency: Scope['sourceAccountCurrency'];
    mode: Scope['mode'];
    sourceAmountTitle: Scope['sourceAmountTitle'];
    sourceAmountName: Scope['sourceAmountName'];
    destinationAccountCurrency: Scope['destinationAccountCurrency'];
    transferInAmountTitle: Scope['transferInAmountTitle'];
    hasVisibleExpenseCategories: Scope['hasVisibleExpenseCategories'];
    allCategories: Scope['allCategories'];
    hasVisibleIncomeCategories: Scope['hasVisibleIncomeCategories'];
    hasVisibleTransferCategories: Scope['hasVisibleTransferCategories'];
    allVisibleAccounts: Scope['allVisibleAccounts'];
    sourceAccountName: Scope['sourceAccountName'];
    sourceAccountTitle: Scope['sourceAccountTitle'];
    allVisibleCategorizedAccounts: Scope['allVisibleCategorizedAccounts'];
    destinationAccountName: Scope['destinationAccountName'];
    transactionDisplayTimezone: Scope['transactionDisplayTimezone'];
    transactionTimezoneTimeDifference: Scope['transactionTimezoneTimeDifference'];
    allTimezones: Scope['allTimezones'];
    geoLocationStatusInfo: Scope['geoLocationStatusInfo'];
    transactionDescriptionTitle: Scope['transactionDescriptionTitle'];
}>();

const transaction = defineModel<Scope['transaction']>('transaction', { required: true });
const geoMenuState = defineModel<Scope['geoMenuState']>('geoMenuState', { required: true });

const emit = defineEmits<{
    updateTransactionTime: Parameters<Scope['updateTransactionTime']>;
    onShowDateTimeError: Parameters<Scope['onShowDateTimeError']>;
    updateTransactionTimezone: Parameters<Scope['updateTransactionTimezone']>;
    updateGeoLocation: Parameters<Scope['updateGeoLocation']>;
    clearGeoLocation: Parameters<Scope['clearGeoLocation']>;
    onSavingTag: Parameters<Scope['onSavingTag']>;
}>();

const updateTransactionTime = (...args: Parameters<Scope['updateTransactionTime']>) => emit('updateTransactionTime', ...args);
const onShowDateTimeError = (...args: Parameters<Scope['onShowDateTimeError']>) => emit('onShowDateTimeError', ...args);
const updateTransactionTimezone = (...args: Parameters<Scope['updateTransactionTimezone']>) => emit('updateTransactionTimezone', ...args);
const updateGeoLocation = (...args: Parameters<Scope['updateGeoLocation']>) => emit('updateGeoLocation', ...args);
const clearGeoLocation = (...args: Parameters<Scope['clearGeoLocation']>) => emit('clearGeoLocation', ...args);
const onSavingTag = (...args: Parameters<Scope['onSavingTag']>) => emit('onSavingTag', ...args);

const { tt, formatCoordinate } = useI18n();

</script>

<template>
    <v-window-item value="basicInfo">
        <v-form class="my-4">
            <v-row>
                <v-col cols="12" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate">
                    <v-text-field
                        type="text"
                        autocomplete="off"
                        persistent-placeholder
                        :disabled="loading || submitting || recognizing"
                        :label="tt('Template Name')"
                        :placeholder="tt('Template Name')"
                        v-model="transaction.name"
                    />
                </v-col>
                <v-col cols="12" :md="transaction.type === TransactionType.Transfer ? 6 : 12">
                    <amount-input class="transaction-edit-amount font-weight-bold"
                                  :color="sourceAmountColor"
                                  :currency="sourceAccountCurrency"
                                  :show-currency="true"
                                  :readonly="mode === TransactionEditPageMode.View"
                                  :disabled="loading || submitting || recognizing"
                                  :persistent-placeholder="true"
                                  :hide="transaction.hideAmount"
                                  :label="sourceAmountTitle"
                                  :placeholder="tt(sourceAmountName)"
                                  :enable-formula="mode !== TransactionEditPageMode.View"
                                  v-model="transaction.sourceAmount"/>
                </v-col>
                <v-col cols="12" :md="6" v-if="transaction.type === TransactionType.Transfer">
                    <amount-input class="transaction-edit-amount font-weight-bold" color="primary"
                                  :currency="destinationAccountCurrency"
                                  :show-currency="true"
                                  :readonly="mode === TransactionEditPageMode.View"
                                  :disabled="loading || submitting || recognizing"
                                  :persistent-placeholder="true"
                                  :hide="transaction.hideAmount"
                                  :label="transferInAmountTitle"
                                  :placeholder="tt('Transfer In Amount')"
                                  :enable-formula="mode !== TransactionEditPageMode.View"
                                  v-model="transaction.destinationAmount"/>
                </v-col>
                <v-col cols="12" md="12" v-if="transaction.type === TransactionType.Expense">
                    <v-tooltip :disabled="hasVisibleExpenseCategories" :text="hasVisibleExpenseCategories ? '' : tt('No secondary expense categories are available')">
                        <template v-slot:activator="{ props }">
                            <div v-bind="props" class="d-block">
                                <two-column-select primary-key-field="id" primary-value-field="id" primary-title-field="name"
                                                   primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                                                   primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                                   secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                                   secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                                                   secondary-hidden-field="hidden"
                                                   :readonly="mode === TransactionEditPageMode.View"
                                                   :disabled="loading || submitting || recognizing || !hasVisibleExpenseCategories"
                                                   :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                                   :show-selection-primary-text="true"
                                                   :custom-selection-primary-text="getTransactionPrimaryCategoryName(transaction.expenseCategoryId, allCategories[CategoryType.Expense])"
                                                   :custom-selection-secondary-text="getTransactionSecondaryCategoryName(transaction.expenseCategoryId, allCategories[CategoryType.Expense])"
                                                   :label="tt('Category')" :placeholder="tt('Category')"
                                                   :items="allCategories[CategoryType.Expense] || []"
                                                   v-model="transaction.expenseCategoryId">
                                </two-column-select>
                            </div>
                        </template>
                    </v-tooltip>
                </v-col>
                <v-col cols="12" md="12" v-if="transaction.type === TransactionType.Income">
                    <v-tooltip :disabled="hasVisibleIncomeCategories" :text="hasVisibleIncomeCategories ? '' : tt('No secondary income categories are available')">
                        <template v-slot:activator="{ props }">
                            <div v-bind="props" class="d-block">
                                <two-column-select primary-key-field="id" primary-value-field="id" primary-title-field="name"
                                                   primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                                                   primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                                   secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                                   secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                                                   secondary-hidden-field="hidden"
                                                   :readonly="mode === TransactionEditPageMode.View"
                                                   :disabled="loading || submitting || recognizing || !hasVisibleIncomeCategories"
                                                   :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                                   :show-selection-primary-text="true"
                                                   :custom-selection-primary-text="getTransactionPrimaryCategoryName(transaction.incomeCategoryId, allCategories[CategoryType.Income])"
                                                   :custom-selection-secondary-text="getTransactionSecondaryCategoryName(transaction.incomeCategoryId, allCategories[CategoryType.Income])"
                                                   :label="tt('Category')" :placeholder="tt('Category')"
                                                   :items="allCategories[CategoryType.Income] || []"
                                                   v-model="transaction.incomeCategoryId">
                                </two-column-select>
                            </div>
                        </template>
                    </v-tooltip>
                </v-col>
                <v-col cols="12" md="12" v-if="transaction.type === TransactionType.Transfer">
                    <v-tooltip :disabled="hasVisibleTransferCategories" :text="hasVisibleTransferCategories ? '' : tt('No secondary transfer categories are available')">
                        <template v-slot:activator="{ props }">
                            <div v-bind="props" class="d-block">
                                <two-column-select primary-key-field="id" primary-value-field="id" primary-title-field="name"
                                                   primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                                                   primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                                   secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                                   secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                                                   secondary-hidden-field="hidden"
                                                   :readonly="mode === TransactionEditPageMode.View"
                                                   :disabled="loading || submitting || recognizing || !hasVisibleTransferCategories"
                                                   :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                                   :show-selection-primary-text="true"
                                                   :custom-selection-primary-text="getTransactionPrimaryCategoryName(transaction.transferCategoryId, allCategories[CategoryType.Transfer])"
                                                   :custom-selection-secondary-text="getTransactionSecondaryCategoryName(transaction.transferCategoryId, allCategories[CategoryType.Transfer])"
                                                   :label="tt('Category')" :placeholder="tt('Category')"
                                                   :items="allCategories[CategoryType.Transfer] || []"
                                                   v-model="transaction.transferCategoryId">
                                </two-column-select>
                            </div>
                        </template>
                    </v-tooltip>
                </v-col>
                <v-col cols="12" :md="transaction.type === TransactionType.Transfer ? 6 : 12">
                    <v-tooltip :disabled="!!allVisibleAccounts.length" :text="allVisibleAccounts.length ? '' : tt('No available account')">
                        <template v-slot:activator="{ props }">
                            <div v-bind="props" class="d-block">
                                <two-column-select primary-key-field="id" primary-value-field="category"
                                                   primary-title-field="name" primary-footer-field="displayBalance"
                                                   primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="account"
                                                   primary-sub-items-field="accounts"
                                                   :primary-title-i18n="true"
                                                   secondary-key-field="id" secondary-value-field="id"
                                                   secondary-title-field="name" secondary-footer-field="displayBalance"
                                                   secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="account" secondary-color-field="color"
                                                   :readonly="mode === TransactionEditPageMode.View"
                                                   :disabled="loading || submitting || recognizing || !allVisibleAccounts.length || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance)"
                                                   :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                   :custom-selection-primary-text="sourceAccountName"
                                                   :label="tt(sourceAccountTitle)"
                                                   :placeholder="tt(sourceAccountTitle)"
                                                   :items="allVisibleCategorizedAccounts"
                                                   v-model="transaction.sourceAccountId">
                                </two-column-select>
                            </div>
                        </template>
                    </v-tooltip>
                </v-col>
                <v-col cols="12" md="6" v-if="transaction.type === TransactionType.Transfer">
                    <v-tooltip :disabled="!!allVisibleAccounts.length" :text="allVisibleAccounts.length ? '' : tt('No available account')">
                        <template v-slot:activator="{ props }">
                            <div v-bind="props" class="d-block">
                                <two-column-select primary-key-field="id" primary-value-field="category"
                                                   primary-title-field="name" primary-footer-field="displayBalance"
                                                   primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="account"
                                                   primary-sub-items-field="accounts"
                                                   :primary-title-i18n="true"
                                                   secondary-key-field="id" secondary-value-field="id"
                                                   secondary-title-field="name" secondary-footer-field="displayBalance"
                                                   secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="account" secondary-color-field="color"
                                                   :readonly="mode === TransactionEditPageMode.View"
                                                   :disabled="loading || submitting || recognizing || !allVisibleAccounts.length"
                                                   :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                   :custom-selection-primary-text="destinationAccountName"
                                                   :label="tt('Destination Account')"
                                                   :placeholder="tt('Destination Account')"
                                                   :items="allVisibleCategorizedAccounts"
                                                   v-model="transaction.destinationAccountId">
                                </two-column-select>
                            </div>
                        </template>
                    </v-tooltip>
                </v-col>
                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Transaction">
                    <date-time-select
                        :readonly="mode === TransactionEditPageMode.View"
                        :disabled="loading || submitting || recognizing || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance)"
                        :label="tt('Transaction Time')"
                        :timezone-utc-offset="transaction.utcOffset"
                        :model-value="transaction.time"
                        @update:model-value="updateTransactionTime"
                        @error="onShowDateTimeError" />
                </v-col>
                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type">
                    <schedule-frequency-select
                        :readonly="mode === TransactionEditPageMode.View"
                        :disabled="loading || submitting || recognizing"
                        :label="tt('Scheduled Transaction Frequency')"
                        v-model:type="transaction.scheduledFrequencyType"
                        v-model="transaction.scheduledFrequency" />
                </v-col>
                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Transaction || (type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type)">
                    <v-autocomplete
                        class="transaction-edit-timezone"
                        item-title="displayNameWithUtcOffset"
                        item-value="name"
                        auto-select-first
                        persistent-placeholder
                        :readonly="mode === TransactionEditPageMode.View"
                        :disabled="loading || submitting || recognizing || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance)"
                        :label="tt('Transaction Timezone')"
                        :placeholder="!transaction.timeZone && transaction.timeZone !== '' ? `(${transactionDisplayTimezone}) ${type === TransactionEditPageType.Transaction ? transactionTimezoneTimeDifference : ''}` : tt('Timezone')"
                        :items="allTimezones"
                        :no-data-text="tt('No results')"
                        :model-value="transaction.timeZone"
                        @update:model-value="updateTransactionTimezone"
                    />
                </v-col>
                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type">
                    <date-select
                        :readonly="mode === TransactionEditPageMode.View"
                        :disabled="loading || submitting || recognizing"
                        :clearable="true"
                        :label="tt('Start Date')"
                        :no-data-text="tt('No limit')"
                        v-model="transaction.scheduledStartDate" />
                </v-col>
                <v-col cols="12" md="6" v-if="type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type">
                    <date-select
                        :readonly="mode === TransactionEditPageMode.View"
                        :disabled="loading || submitting || recognizing"
                        :clearable="true"
                        :label="tt('End Date')"
                        :no-data-text="tt('No limit')"
                        v-model="transaction.scheduledEndDate" />
                </v-col>
                <v-col cols="12" md="12" v-if="type === TransactionEditPageType.Transaction">
                    <v-select
                        persistent-placeholder
                        :readonly="mode === TransactionEditPageMode.View"
                        :disabled="loading || submitting || recognizing"
                        :label="tt('Geographic Location')"
                        v-model="transaction"
                        v-model:menu="geoMenuState"
                    >
                        <template #selection>
                            <span class="cursor-pointer" v-if="transaction.geoLocation">{{ `(${formatCoordinate(transaction.geoLocation)})` }}</span>
                            <span class="cursor-pointer" v-else-if="!transaction.geoLocation">{{ geoLocationStatusInfo }}</span>
                        </template>

                        <template #no-data>
                            <v-list class="py-0">
                                <v-list-item class="text-body-medium" v-if="mode !== TransactionEditPageMode.View" @click="updateGeoLocation(true)">{{ tt('Update Geographic Location') }}</v-list-item>
                                <v-list-item class="text-body-medium" v-if="mode !== TransactionEditPageMode.View" @click="clearGeoLocation">{{ tt('Clear Geographic Location') }}</v-list-item>
                            </v-list>
                        </template>
                    </v-select>
                </v-col>
                <v-col cols="12" md="12">
                    <transaction-tag-auto-complete
                        :readonly="mode === TransactionEditPageMode.View"
                        :disabled="loading || submitting || recognizing"
                        :show-label="true"
                        :allow-add-new-tag="true"
                        v-model="transaction.tagIds"
                        @tag:saving="onSavingTag"
                    />
                </v-col>
                <v-col cols="12" md="12">
                    <v-textarea
                        type="text"
                        autocomplete="off"
                        persistent-placeholder
                        rows="3"
                        :readonly="mode === TransactionEditPageMode.View"
                        :disabled="loading || submitting || recognizing"
                        :label="transactionDescriptionTitle"
                        :placeholder="tt('Your transaction description (optional)')"
                        v-model="transaction.comment"
                    />
                </v-col>
            </v-row>
        </v-form>
    </v-window-item>
</template>
