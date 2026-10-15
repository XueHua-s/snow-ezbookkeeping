<script setup lang="ts">
import ConditionValueInput from './ExplorerQueryTab/ConditionValueInput.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import {
    TransactionExplorerConditionField,

    TransactionExplorerConditionRelation
} from '@/core/explorer.ts';

import {

    isArray
} from '@/lib/common.ts';

import AccountFilterSettingsDialog from '@/views/desktop/common/dialogs/AccountFilterSettingsDialog.vue';
import CategoryFilterSettingsDialog from '@/views/desktop/common/dialogs/CategoryFilterSettingsDialog.vue';
import {
    mdiCheck,
    mdiClose,
    mdiContentCopy,
    mdiDrag,
    mdiPencilOutline,
    mdiPlus,
    mdiTextBoxSearchOutline
} from '@mdi/js';
import { type ExplorerQueryTabProps, useExplorerQueryTab } from './ExplorerQueryTab/useExplorerQueryTab.ts';

const props = defineProps<ExplorerQueryTabProps>();

const {
    tt, getAllTransactionExplorerConditionOperators, formatNumberToLocalizedNumeralsWithoutDigitGrouping, currentCondition,
    showExpression, showFilterSourceAccountsDialog, showFilterDestinationAccountsDialog,
    showFilterTransactionCategoriesDialog, editingQuery, editingQueryName, queries, defaultCurrency, hasAnyAccount,
    hasAnyTransactionCategory, allTransactionExplorerConditionFields, allAvailableDayOfWeekOptions,
    allAvailableDayOfMonthOptions, allAvailableMonthOfYearOptions, allAvailableHourOfDayOptions,
    getFilteredAccountsDisplayContent, getFilteredTransactionCategoriesDisplayContent, getConditionStyle, addQuery,
    updateQueryName, cancelUpdateQueryName, duplicateQuery, removeQuery, clearAllQueries, addCondition, addSubCondition,
    removeCondition, updateItemRelation, updateConditionField, updateSourceAccount, updateDestinationAccount,
    updateTransactionCategories, getExpression,
} = useExplorerQueryTab(props);

</script>

<template>
    <v-card-subtitle class="px-4">
        <div class="title-and-toolbar d-flex">
            <v-btn color="default" variant="outlined"
                   :disabled="loading || disabled || !!editingQuery"
                   @click="addQuery">{{ tt('Add Query') }}</v-btn>
            <v-spacer />
            <v-btn color="secondary" variant="tonal"
                   :disabled="loading || disabled || !!editingQuery || queries.length < 1"
                   @click="clearAllQueries">{{ tt('Clear All') }}</v-btn>
        </div>
    </v-card-subtitle>
    <v-card-text class="pt-0 pb-0">
        <draggable-list
            item-key="id"
            handle=".drag-handle"
            ghost-class="dragging-item"
            v-model="queries"
        >
            <template #item="{ element, index }">
                <v-card border class="card-title-with-bg my-4">
                    <v-card-title class="d-flex align-center py-2 px-5">
                        <v-icon :icon="mdiTextBoxSearchOutline" size="20" />
                        <span class="query-name text-body-large ms-2" v-if="editingQuery !== element">{{ element.name || tt('format.misc.queryIndex', { index: formatNumberToLocalizedNumeralsWithoutDigitGrouping(index + 1) }) }}</span>
                        <div class="query-name-edit ms-2" v-if="editingQuery === element">
                            <v-text-field autofocus type="text" autocomplete="off" density="compact" variant="underlined"
                                          :disabled="loading || disabled"
                                          :placeholder="tt('format.misc.queryIndex', { index: formatNumberToLocalizedNumeralsWithoutDigitGrouping(index + 1) })"
                                          v-text-field-auto-width="{ minWidth: 20, maxWidth: 300, auxSpanId: `query-name-aux-span-${index + 1}-${element.id}` }"
                                          v-model="editingQueryName"
                                          @keyup.esc="cancelUpdateQueryName"
                                          @keyup.enter="updateQueryName(element)" />
                            <span :id="`query-name-aux-span-${index + 1}-${element.id}`" />
                        </div>
                        <v-btn class="ms-2" density="compact" color="primary" variant="text"
                               :aria-label="tt('Apply')" :disabled="loading || disabled" :icon="true"
                               @click="updateQueryName(element)"
                               v-if="editingQuery === element">
                            <v-icon :icon="mdiCheck" size="18" />
                            <v-tooltip activator="parent">{{ tt('Apply') }}</v-tooltip>
                        </v-btn>
                        <v-btn class="ms-1" density="compact" color="default" variant="text"
                               :aria-label="tt('Cancel')" :disabled="loading || disabled" :icon="true"
                               @click="cancelUpdateQueryName"
                               v-if="editingQuery === element">
                            <v-icon :icon="mdiClose" size="18" />
                            <v-tooltip activator="parent">{{ tt('Cancel') }}</v-tooltip>
                        </v-btn>
                        <v-btn class="ms-2" density="compact" color="default" variant="text"
                               :aria-label="tt('Modify Query Name')" :disabled="loading || disabled || !!editingQuery" :icon="true"
                               @click="editingQueryName = element.name; editingQuery = element"
                               v-if="!editingQuery || editingQuery !== element">
                            <v-icon :icon="mdiPencilOutline" size="18" />
                            <v-tooltip activator="parent">{{ tt('Modify Query Name') }}</v-tooltip>
                        </v-btn>
                        <v-btn class="ms-1" density="compact" color="default" variant="text"
                               :aria-label="tt('Duplicate')" :disabled="loading || disabled || !!editingQuery" :icon="true"
                               @click="duplicateQuery(element)"
                               v-if="!editingQuery || editingQuery !== element">
                            <v-icon :icon="mdiContentCopy" size="18" />
                            <v-tooltip activator="parent">{{ tt('Duplicate') }}</v-tooltip>
                        </v-btn>
                        <v-spacer />
                        <v-switch class="bidirectional-switch ms-2" color="secondary"
                                  :disabled="loading || disabled || !!editingQuery || !element.conditions || element.conditions.length < 1"
                                  :label="tt('Expression')"
                                  v-model="showExpression[element.id]"
                                  @click="showExpression[element.id] = !showExpression[element.id]">
                            <template #prepend>
                                <span>{{ tt('Editor') }}</span>
                            </template>
                        </v-switch>
                        <v-btn class="ms-2" density="compact" color="default" variant="text"
                               :aria-label="tt('Remove Query')" :disabled="loading || disabled || !!editingQuery || queries.length < 1 || (queries.length === 1 && (!element.conditions || element.conditions.length < 1))" :icon="true"
                               @click="removeQuery(element, index)">
                            <v-icon :icon="mdiClose" size="18" />
                            <v-tooltip activator="parent">{{ tt('Remove Query') }}</v-tooltip>
                        </v-btn>
                        <span class="ms-1 mb-1">
                            <v-icon :class="!loading && !disabled && !editingQuery && queries.length > 1 ? 'drag-handle' : 'disabled'"
                                    :aria-label="tt('Drag to Reorder')" :icon="mdiDrag"/>
                            <v-tooltip activator="parent" v-if="!loading && !disabled && !editingQuery && queries.length > 1">{{ tt('Drag to Reorder') }}</v-tooltip>
                        </span>
                    </v-card-title>

                    <v-divider />

                    <v-card-text>
                        <v-row>
                            <v-col cols="12">
                                <div class="text-center py-4" v-if="loading">
                                    <v-skeleton-loader class="skeleton-no-margin ms-3" type="text" :loading="true"></v-skeleton-loader>
                                </div>

                                <div class="text-body-medium text-center py-3" v-else-if="!loading && !element.conditions || element.conditions.length < 1">
                                    {{ tt('No conditions defined. All transactions will match.') }}
                                </div>

                                <div v-else-if="element.conditions && element.conditions.length > 0 && !showExpression[element.id]">
                                    <div :key="conditionIndex" v-for="(conditionWithRelation, conditionIndex) in element.conditions">
                                        <div class="d-flex overflow-x-auto align-center gap-2 mb-3"
                                             :style="getConditionStyle(element, conditionIndex)"
                                             v-if="conditionWithRelation.relation !== TransactionExplorerConditionRelation.SubEnd">
                                            <v-select
                                                disabled
                                                class="flex-0-0"
                                                width="140px"
                                                density="compact"
                                                item-title="displayName"
                                                item-value="value"
                                                :items="[{ value: TransactionExplorerConditionRelation.First, displayName: tt('WHERE') }]"
                                                :model-value="TransactionExplorerConditionRelation.First"
                                                v-if="conditionIndex < 1"
                                            />

                                            <v-select
                                                class="flex-0-0"
                                                width="140px"
                                                density="compact"
                                                item-title="displayName"
                                                item-value="value"
                                                :disabled="loading || disabled || !!editingQuery"
                                                :items="[
                                                    { value: TransactionExplorerConditionRelation.And, displayName: tt('AND') },
                                                    { value: TransactionExplorerConditionRelation.Or, displayName: tt('OR') },
                                                    { value: TransactionExplorerConditionRelation.AndSub, displayName: tt('AND SUB') },
                                                    { value: TransactionExplorerConditionRelation.OrSub, displayName: tt('OR SUB') }
                                                ]"
                                                :model-value="conditionWithRelation.relation"
                                                @update:model-value="updateItemRelation(element, conditionIndex, $event)"
                                                v-else-if="conditionIndex >= 1"
                                            />

                                            <v-select
                                                class="flex-0-0"
                                                density="compact"
                                                item-title="name"
                                                item-value="value"
                                                :disabled="loading || disabled || !!editingQuery"
                                                :items="allTransactionExplorerConditionFields"
                                                @update:model-value="updateConditionField(element, conditionIndex, TransactionExplorerConditionField.valueOf($event))"
                                                v-model="conditionWithRelation.condition.field"
                                            />

                                            <v-select
                                                class="flex-0-0"
                                                density="compact"
                                                item-title="name"
                                                item-value="value"
                                                :disabled="loading || disabled || !!editingQuery"
                                                :items="getAllTransactionExplorerConditionOperators(conditionWithRelation.getSupportedOperators())"
                                                v-model="conditionWithRelation.condition.operator"
                                            />

                                            <ConditionValueInput
    :loading="loading"
    :disabled="disabled"
    :editingQuery="editingQuery"
    :allAvailableDayOfWeekOptions="allAvailableDayOfWeekOptions"
    :allAvailableDayOfMonthOptions="allAvailableDayOfMonthOptions"
    :allAvailableMonthOfYearOptions="allAvailableMonthOfYearOptions"
    :allAvailableHourOfDayOptions="allAvailableHourOfDayOptions"
    :hasAnyTransactionCategory="hasAnyTransactionCategory"
    :getFilteredTransactionCategoriesDisplayContent="getFilteredTransactionCategoriesDisplayContent"
    :hasAnyAccount="hasAnyAccount"
    :getFilteredAccountsDisplayContent="getFilteredAccountsDisplayContent"
    :defaultCurrency="defaultCurrency"
    v-model:condition="conditionWithRelation.condition"
    v-model:currentCondition="currentCondition"
    v-model:showFilterTransactionCategoriesDialog="showFilterTransactionCategoriesDialog"
    v-model:showFilterSourceAccountsDialog="showFilterSourceAccountsDialog"
    v-model:showFilterDestinationAccountsDialog="showFilterDestinationAccountsDialog"
/>

                                            <v-btn color="default" density="compact" variant="text"
                                                   :aria-label="tt('Remove Condition')" :disabled="loading || disabled || !!editingQuery" :icon="true"
                                                   @click="removeCondition(element, conditionIndex)">
                                                <v-icon :icon="mdiClose" size="18" />
                                                <v-tooltip activator="parent">{{ tt('Remove Condition') }}</v-tooltip>
                                            </v-btn>
                                        </div>
                                        <div class="d-flex overflow-x-auto align-center gap-2 mb-3"
                                             :style="getConditionStyle(element, conditionIndex)"
                                             v-if="conditionWithRelation.relation === TransactionExplorerConditionRelation.SubEnd">
                                            <v-btn class="px-2" density="comfortable" color="primary" variant="text"
                                                   :prepend-icon="mdiPlus"
                                                   :disabled="loading || disabled || !!editingQuery"
                                                   @click="addSubCondition(element, conditionIndex)">
                                                {{ tt('Add Sub Condition') }}
                                            </v-btn>
                                        </div>
                                    </div>
                                </div>
                                <div v-else-if="element.conditions && element.conditions.length > 0 && showExpression[element.id]">
                                    <div class="w-100 code-container">
                                        <v-textarea class="w-100 always-cursor-text mb-4" :readonly="true"
                                                    :value="getExpression(element, index)"></v-textarea>
                                    </div>
                                </div>

                                <v-btn class="px-2" density="comfortable" color="primary" variant="text"
                                       :prepend-icon="mdiPlus"
                                       :disabled="loading || disabled || !!editingQuery || showExpression[element.id]"
                                       @click="addCondition(element)">
                                    {{ tt('Add Condition') }}
                                </v-btn>
                            </v-col>
                        </v-row>
                    </v-card-text>
                </v-card>
            </template>
        </draggable-list>
    </v-card-text>

    <account-filter-settings-dialog type="custom"
                                    :selected-account-ids="isArray(currentCondition?.value) ? currentCondition?.value as string[] : []"
                                    v-model:show="showFilterSourceAccountsDialog"
                                    @settings:change="updateSourceAccount" />

    <account-filter-settings-dialog type="custom"
                                    :selected-account-ids="isArray(currentCondition?.value) ? currentCondition?.value as string[] : []"
                                    v-model:show="showFilterDestinationAccountsDialog"
                                    @settings:change="updateDestinationAccount" />

    <category-filter-settings-dialog type="custom"
                                     :selected-category-ids="isArray(currentCondition?.value) ? currentCondition?.value as string[] : []"
                                     v-model:show="showFilterTransactionCategoriesDialog"
                                     @settings:change="updateTransactionCategories" />

    <snack-bar ref="snackbar" />
</template>



<style>
.query-name {
    white-space: pre;
}

.query-name-edit {
    display: flex;
    align-items: center;
    height: 36px;
}

.query-name-edit .v-field {
    font-size: 0.9375rem;
}

.query-name-edit .v-field__input {
    padding-top: 0;
}
</style>
