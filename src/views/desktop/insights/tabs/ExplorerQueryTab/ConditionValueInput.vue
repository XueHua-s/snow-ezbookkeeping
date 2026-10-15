<script setup lang="ts">
import {
    EASTERNMOST_TIMEZONE_UTC_OFFSET,
    WESTERNMOST_TIMEZONE_UTC_OFFSET,
} from '@/consts/timezone.ts';
import {
    TransactionExplorerConditionField,
    TransactionExplorerConditionOperator
} from '@/core/explorer.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    arrayItemToObjectField
} from '@/lib/common.ts';
import { useI18n } from '@/locales/helpers.ts';
import {
    TransactionExplorerGeoLocationCondition
} from '@/models/explorer.ts';
import type { useExplorerQueryTab } from '@/views/desktop/insights/tabs/ExplorerQueryTab/useExplorerQueryTab.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useExplorerQueryTab>>;

defineProps<{
    loading: boolean | undefined;
    disabled: boolean | undefined;
    editingQuery: Scope['editingQuery'];
    allAvailableDayOfWeekOptions: Scope['allAvailableDayOfWeekOptions'];
    allAvailableDayOfMonthOptions: Scope['allAvailableDayOfMonthOptions'];
    allAvailableMonthOfYearOptions: Scope['allAvailableMonthOfYearOptions'];
    allAvailableHourOfDayOptions: Scope['allAvailableHourOfDayOptions'];
    hasAnyTransactionCategory: Scope['hasAnyTransactionCategory'];
    getFilteredTransactionCategoriesDisplayContent: Scope['getFilteredTransactionCategoriesDisplayContent'];
    hasAnyAccount: Scope['hasAnyAccount'];
    getFilteredAccountsDisplayContent: Scope['getFilteredAccountsDisplayContent'];
    defaultCurrency: Scope['defaultCurrency'];
}>();

const condition = defineModel<import('@/models/explorer.ts').TransactionExplorerConditionWithRelation['condition']>('condition', { required: true });
const currentCondition = defineModel<Scope['currentCondition']>('currentCondition', { required: true });
const showFilterTransactionCategoriesDialog = defineModel<Scope['showFilterTransactionCategoriesDialog']>('showFilterTransactionCategoriesDialog', { required: true });
const showFilterSourceAccountsDialog = defineModel<Scope['showFilterSourceAccountsDialog']>('showFilterSourceAccountsDialog', { required: true });
const showFilterDestinationAccountsDialog = defineModel<Scope['showFilterDestinationAccountsDialog']>('showFilterDestinationAccountsDialog', { required: true });

const { tt } = useI18n();

</script>

<template>
    <div class="d-flex w-100 flex-1-1" style="min-width: 280px;">
        <v-select
            multiple chips closable-chips
            density="compact"
            item-title="displayName"
            item-value="type"
            :disabled="loading || disabled || !!editingQuery"
            :placeholder="tt('None')"
            :items="allAvailableDayOfWeekOptions"
            v-model="condition.value as number[]"
            v-if="condition.field === TransactionExplorerConditionField.TransactionTimeDayOfWeek.value"
        >
            <template #item="{ props, internalItem }">
                <v-list-item :value="internalItem.value" v-bind="props">
                    <template #title>
                        <v-list-item-title>
                            <div class="d-flex align-center">{{ internalItem.title }}</div>
                        </v-list-item-title>
                    </template>
                </v-list-item>
            </template>
        </v-select>

        <v-select
            multiple chips closable-chips
            density="compact"
            item-title="displayName"
            item-value="type"
            :disabled="loading || disabled || !!editingQuery"
            :placeholder="tt('None')"
            :items="allAvailableDayOfMonthOptions"
            v-model="condition.value as number[]"
            v-else-if="condition.field === TransactionExplorerConditionField.TransactionTimeDayOfMonth.value"
        >
            <template #item="{ props, internalItem }">
                <v-list-item :value="internalItem.value" v-bind="props">
                    <template #title>
                        <v-list-item-title>
                            <div class="d-flex align-center">{{ internalItem.title }}</div>
                        </v-list-item-title>
                    </template>
                </v-list-item>
            </template>
        </v-select>

        <v-select
            multiple chips closable-chips
            density="compact"
            item-title="displayName"
            item-value="type"
            :disabled="loading || disabled || !!editingQuery"
            :placeholder="tt('None')"
            :items="allAvailableMonthOfYearOptions"
            v-model="condition.value as number[]"
            v-else-if="condition.field === TransactionExplorerConditionField.TransactionTimeMonthOfYear.value"
        >
            <template #item="{ props, internalItem }">
                <v-list-item :value="internalItem.value" v-bind="props">
                    <template #title>
                        <v-list-item-title>
                            <div class="d-flex align-center">{{ internalItem.title }}</div>
                        </v-list-item-title>
                    </template>
                </v-list-item>
            </template>
        </v-select>

        <v-select
            multiple chips closable-chips
            density="compact"
            item-title="displayName"
            item-value="type"
            :disabled="loading || disabled || !!editingQuery"
            :placeholder="tt('None')"
            :items="allAvailableHourOfDayOptions"
            v-model="condition.value as number[]"
            v-else-if="condition.field === TransactionExplorerConditionField.TransactionTimeHourOfDay.value"
        >
            <template #item="{ props, internalItem }">
                <v-list-item :value="internalItem.value" v-bind="props">
                    <template #title>
                        <v-list-item-title>
                            <div class="d-flex align-center">{{ internalItem.title }}</div>
                        </v-list-item-title>
                    </template>
                </v-list-item>
            </template>
        </v-select>

        <div class="d-flex w-100 align-center gap-2"
             v-else-if="condition.field === TransactionExplorerConditionField.TransactionTimezone.value">
            <number-input density="compact"
                          :disabled="loading || disabled || !!editingQuery"
                          :min-value="WESTERNMOST_TIMEZONE_UTC_OFFSET"
                          :max-value="EASTERNMOST_TIMEZONE_UTC_OFFSET"
                          :max-decimal-count="0"
                          v-model="condition.value[0]"
                          v-if="condition.operator === TransactionExplorerConditionOperator.MinuteOffsetBetween.value ||
                                condition.operator === TransactionExplorerConditionOperator.MinuteOffsetNotBetween.value"
            />
            <span class="ms-2 me-2"
                  v-if="condition.operator === TransactionExplorerConditionOperator.MinuteOffsetBetween.value ||
                        condition.operator === TransactionExplorerConditionOperator.MinuteOffsetNotBetween.value">{{ tt('format.misc.rangeSeparator') }}</span>
            <number-input density="compact"
                          :disabled="loading || disabled || !!editingQuery"
                          :min-value="WESTERNMOST_TIMEZONE_UTC_OFFSET"
                          :max-value="EASTERNMOST_TIMEZONE_UTC_OFFSET"
                          :max-decimal-count="0"
                          v-model="condition.value[1]"
                          v-if="condition.operator === TransactionExplorerConditionOperator.MinuteOffsetBetween.value ||
                                condition.operator === TransactionExplorerConditionOperator.MinuteOffsetNotBetween.value"
            />
        </div>

        <v-select
            multiple chips closable-chips
            density="compact"
            item-title="displayName"
            item-value="type"
            :disabled="loading || disabled || !!editingQuery"
            :placeholder="tt('None')"
            :items="[
                { type: TransactionType.Expense, displayName: tt('Expense') },
                { type: TransactionType.Income, displayName: tt('Income') },
                { type: TransactionType.Transfer, displayName: tt('Transfer') }
            ]"
            v-model="condition.value as number[]"
            v-else-if="condition.field === TransactionExplorerConditionField.TransactionType.value"
        >
            <template #item="{ props, internalItem }">
                <v-list-item :value="internalItem.value" v-bind="props">
                    <template #title>
                        <v-list-item-title>
                            <div class="d-flex align-center">{{ internalItem.title }}</div>
                        </v-list-item-title>
                    </template>
                </v-list-item>
            </template>
        </v-select>

        <v-text-field
            class="always-cursor-pointer text-field-truncate"
            density="compact"
            item-title="displayName"
            item-value="type"
            persistent-placeholder
            :readonly="true"
            :disabled="loading || disabled || !!editingQuery || !hasAnyTransactionCategory"
            :placeholder="tt('None')"
            :model-value="getFilteredTransactionCategoriesDisplayContent(arrayItemToObjectField(condition.value as string[], true))"
            @click="currentCondition = condition; showFilterTransactionCategoriesDialog = true"
            v-else-if="condition.field === TransactionExplorerConditionField.TransactionCategory.value"
        />

        <v-text-field
            class="always-cursor-pointer text-field-truncate"
            density="compact"
            item-title="displayName"
            item-value="type"
            persistent-placeholder
            :readonly="true"
            :disabled="loading || disabled || !!editingQuery || !hasAnyAccount"
            :placeholder="tt('None')"
            :model-value="getFilteredAccountsDisplayContent(arrayItemToObjectField(condition.value as string[], true))"
            @click="currentCondition = condition; showFilterSourceAccountsDialog = true"
            v-else-if="condition.field === TransactionExplorerConditionField.SourceAccount.value"
        />

        <v-text-field
            class="always-cursor-pointer text-field-truncate"
            density="compact"
            item-title="displayName"
            item-value="type"
            persistent-placeholder
            :readonly="true"
            :disabled="loading || disabled || !!editingQuery || !hasAnyAccount"
            :placeholder="tt('None')"
            :model-value="getFilteredAccountsDisplayContent(arrayItemToObjectField(condition.value as string[], true))"
            @click="currentCondition = condition; showFilterDestinationAccountsDialog = true"
            v-else-if="condition.field === TransactionExplorerConditionField.DestinationAccount.value"
        />

        <div class="d-flex w-100 align-center gap-2"
             v-else-if="condition.field === TransactionExplorerConditionField.SourceAmount.value ||
                        condition.field === TransactionExplorerConditionField.DestinationAmount.value">
            <amount-input density="compact"
                          :currency="defaultCurrency"
                          :disabled="loading || disabled || !!editingQuery"
                          v-model="condition.value[0]"
            />
            <span class="ms-2 me-2"
                  v-if="condition.operator === TransactionExplorerConditionOperator.Between.value ||
                        condition.operator === TransactionExplorerConditionOperator.NotBetween.value">{{ tt('format.misc.rangeSeparator') }}</span>
            <amount-input density="compact"
                          :currency="defaultCurrency"
                          :disabled="loading || disabled || !!editingQuery"
                          v-model="condition.value[1]"
                          v-if="condition.operator === TransactionExplorerConditionOperator.Between.value ||
                                condition.operator === TransactionExplorerConditionOperator.NotBetween.value"
            />
        </div>

        <div class="d-flex w-100 align-center gap-2"
             v-else-if="condition.field === TransactionExplorerConditionField.GeoLocation.value">
            <v-text-field disabled density="compact"
                          :placeholder="tt('None')"
                          v-if="condition.operator !== TransactionExplorerConditionOperator.LatitudeBetween.value &&
                                condition.operator !== TransactionExplorerConditionOperator.LatitudeNotBetween.value &&
                                condition.operator !== TransactionExplorerConditionOperator.LongitudeBetween.value &&
                                condition.operator !== TransactionExplorerConditionOperator.LongitudeNotBetween.value"
            />
            <number-input autocomplete="off" density="compact"
                          :disabled="loading || disabled || !!editingQuery"
                          :min-value="TransactionExplorerGeoLocationCondition.MIN_LATITUDE"
                          :max-value="TransactionExplorerGeoLocationCondition.MAX_LATITUDE"
                          :max-decimal-count="6"
                          v-model="condition.value[0]"
                          v-if="condition.operator === TransactionExplorerConditionOperator.LatitudeBetween.value ||
                                condition.operator === TransactionExplorerConditionOperator.LatitudeNotBetween.value"
            />
            <number-input density="compact"
                          :disabled="loading || disabled || !!editingQuery"
                          :min-value="TransactionExplorerGeoLocationCondition.MIN_LONGITUDE"
                          :max-value="TransactionExplorerGeoLocationCondition.MAX_LONGITUDE"
                          :max-decimal-count="6"
                          v-model="condition.value[2]"
                          v-if="condition.operator === TransactionExplorerConditionOperator.LongitudeBetween.value ||
                                condition.operator === TransactionExplorerConditionOperator.LongitudeNotBetween.value"
            />
            <span class="ms-2 me-2"
                  v-if="condition.operator === TransactionExplorerConditionOperator.LatitudeBetween.value ||
                        condition.operator === TransactionExplorerConditionOperator.LatitudeNotBetween.value ||
                        condition.operator === TransactionExplorerConditionOperator.LongitudeBetween.value ||
                        condition.operator === TransactionExplorerConditionOperator.LongitudeNotBetween.value">{{ tt('format.misc.rangeSeparator') }}</span>
            <number-input density="compact"
                          :disabled="loading || disabled || !!editingQuery"
                          :min-value="TransactionExplorerGeoLocationCondition.MIN_LATITUDE"
                          :max-value="TransactionExplorerGeoLocationCondition.MAX_LATITUDE"
                          :max-decimal-count="6"
                          v-model="condition.value[1]"
                          v-if="condition.operator === TransactionExplorerConditionOperator.LatitudeBetween.value ||
                                condition.operator === TransactionExplorerConditionOperator.LatitudeNotBetween.value"
            />
            <number-input density="compact"
                          :disabled="loading || disabled || !!editingQuery"
                          :min-value="TransactionExplorerGeoLocationCondition.MIN_LONGITUDE"
                          :max-value="TransactionExplorerGeoLocationCondition.MAX_LONGITUDE"
                          :max-decimal-count="6"
                          v-model="condition.value[3]"
                          v-if="condition.operator === TransactionExplorerConditionOperator.LongitudeBetween.value ||
                                condition.operator === TransactionExplorerConditionOperator.LongitudeNotBetween.value"
            />
        </div>

        <div class="d-flex w-100" v-else-if="condition.field === TransactionExplorerConditionField.TransactionTag.value">
            <v-text-field
                autocomplete="off"
                disabled
                persistent-placeholder
                density="compact"
                :placeholder="tt('None')"
                v-if="condition.field === TransactionExplorerConditionField.TransactionTag.value &&
                     (condition.operator === TransactionExplorerConditionOperator.IsEmpty.value || condition.operator === TransactionExplorerConditionOperator.IsNotEmpty.value)"
            />

            <transaction-tag-auto-complete
                density="compact"
                :disabled="loading || disabled || !!editingQuery"
                v-model="condition.value"
                v-else-if="condition.operator !== TransactionExplorerConditionOperator.IsEmpty.value && condition.operator !== TransactionExplorerConditionOperator.IsNotEmpty.value"
            />
        </div>

        <v-text-field disabled density="compact"
                      :placeholder="tt('None')"
                      v-else-if="condition.field === TransactionExplorerConditionField.Pictures.value"
        />

        <v-text-field disabled density="compact"
                      :placeholder="tt('None')"
                      v-else-if="(condition.field === TransactionExplorerConditionField.Description.value || condition.field === TransactionExplorerConditionField.DescriptionCaseInsensitive.value || condition.field === TransactionExplorerConditionField.DescriptionNormalized.value) &&
                                 condition.operator === TransactionExplorerConditionOperator.IsEmpty.value || condition.operator === TransactionExplorerConditionOperator.IsNotEmpty.value"
        />

        <v-text-field autocomplete="off" density="compact"
                      :disabled="loading || disabled || !!editingQuery"
                      :placeholder="tt('None')"
                      v-model="condition.value"
                      v-else-if="(condition.field === TransactionExplorerConditionField.Description.value || condition.field === TransactionExplorerConditionField.DescriptionCaseInsensitive.value || condition.field === TransactionExplorerConditionField.DescriptionNormalized.value) &&
                                 condition.operator !== TransactionExplorerConditionOperator.IsEmpty.value && condition.operator !== TransactionExplorerConditionOperator.IsNotEmpty.value"
        />
    </div>
</template>
