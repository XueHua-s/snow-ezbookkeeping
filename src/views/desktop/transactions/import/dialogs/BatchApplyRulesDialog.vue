<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';
import {
    ImportTransactionReplaceRuleAction,
    ImportTransactionReplaceRuleActionType
} from '@/core/rule.ts';
import { TransactionTag } from '@/models/transaction_tag.ts';
import {
    mdiCheck,
    mdiClose,
    mdiContentSaveOutline,
    mdiDotsVertical,
    mdiDrag,
    mdiFolderOpenOutline,
    mdiFunctionVariant,
    mdiMenuDown,
    mdiPencilOutline,
    mdiPound,
    mdiRefresh
} from '@mdi/js';
import { useBatchApplyRulesDialog } from './BatchApplyRulesDialog/useBatchApplyRulesDialog.ts';

const {
    tt, allTagsWithGroupHeader, showState, loading, rules, editingRule, editingRuleName, defaultCurrency, allVisibleAccounts,
    allVisibleCategorizedAccounts, allCategories, allTimezones, allConditionFields, allActions, transactionTypes, amountSigns,
    canApply, isModified, getDefaultRuleName, getRuleName, getConditionOperators, isActionDisabledForRule, getSourceItems,
    isItemCondition, isAmountCondition, isRangeOperator, isEmptyTextConditionOperator, getActionCategoryType,
    getSelectedCategoryPrimaryName, getSelectedCategorySecondaryName, getAccountDisplayName, open, reload, addRule,
    removeRule, updateRuleName, cancelUpdateRuleName, updateConditionField, updateStringConditionValue,
    updateConditionOperator, getAmountConditionValue, updateAmountConditionValue, updateActionType, loadReplaceRuleFile,
    saveReplaceRuleFile, apply, cancel,
} = useBatchApplyRulesDialog();

defineExpose({
    open
});

</script>

<template>
    <v-dialog width="1000" :persistent="loading || isModified" v-model="showState">
        <one-column-dialog-layout content-class="pa-0" :title="tt('Batch Apply Rules')"
                                  :cancel-button-title="tt('Cancel')" @cancel="cancel">
            <template #after-title>
                <v-btn density="compact" color="default" variant="text" class="ms-2"
                       :aria-label="tt('Refresh')" :icon="true" :disabled="loading"
                       :loading="loading" @click="reload">
                    <template #loader>
                        <v-progress-circular indeterminate size="20" />
                    </template>
                    <v-icon :icon="mdiRefresh" size="22" />
                    <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                </v-btn>
                <v-btn density="compact" color="primary" variant="outlined" class="ms-2"
                       :disabled="loading || !!editingRule" @click="addRule">{{ tt('Add Rule') }}</v-btn>
            </template>

            <template #toolbar>
                <v-btn-group class="ms-2" density="compact" variant="outlined" color="primary"
                             :disabled="loading || !!editingRule || !canApply">
                    <v-btn :disabled="loading || !!editingRule || !canApply"
                           @click="apply('selected')">
                        {{ tt('Apply to Selected Transactions') }}
                    </v-btn>
                    <v-btn :aria-label="tt('More')" :icon="true"
                           :disabled="loading || !!editingRule || !canApply">
                        <v-icon :icon="mdiMenuDown" size="24" />
                        <v-menu activator="parent">
                            <v-list>
                                <v-list-item :title="tt('Apply to All Transactions')"
                                             @click="apply('all')"></v-list-item>
                            </v-list>
                        </v-menu>
                    </v-btn>
                </v-btn-group>
                <v-btn density="compact" color="default" variant="text" class="ms-2"
                       :aria-label="tt('More')" :disabled="loading || !!editingRule" :icon="true">
                    <v-icon :icon="mdiDotsVertical" />
                    <v-menu activator="parent" max-height="500">
                        <v-list>
                            <v-list-item :prepend-icon="mdiFolderOpenOutline" :title="tt('Load Replace Rule File')"
                                         @click="loadReplaceRuleFile()" />
                            <v-list-item :prepend-icon="mdiContentSaveOutline" :title="tt('Save Replace Rule File')"
                                         :disabled="rules.length < 1" @click="saveReplaceRuleFile()" />
                        </v-list>
                    </v-menu>
                </v-btn>
            </template>

            <template #content>
                <div ref="rulesContainer" class="px-4 pb-2">
                    <div class="text-body-medium text-center py-5" v-if="rules.length < 1">{{ tt('No rules') }}</div>
                    <draggable-list item-key="id" handle=".drag-handle" ghost-class="dragging-item"
                                    :disabled="loading || !!editingRule || rules.length < 2" v-model="rules">
                        <template #item="{ element, index }">
                            <v-card border class="rule-card card-title-with-bg my-4">
                                <v-card-title class="rule-card-title d-flex align-center py-2 px-5">
                                    <v-icon :icon="mdiFunctionVariant" size="20" />
                                    <span class="rule-name text-body-large ms-2" v-if="editingRule !== element">{{ getRuleName(element, index) }}</span>
                                    <div class="rule-name-edit ms-2" v-if="editingRule === element">
                                        <v-text-field autofocus type="text" autocomplete="off"
                                                      density="compact" variant="underlined"
                                                      :disabled="loading" :placeholder="getDefaultRuleName(index)"
                                                      v-text-field-auto-width="{ minWidth: 20, maxWidth: 300, auxSpanId: `rule-name-aux-span-${index + 1}-${element.id}` }"
                                                      v-model="editingRuleName" @keyup.esc="cancelUpdateRuleName"
                                                      @keyup.enter="updateRuleName(element)" />
                                        <span :id="`rule-name-aux-span-${index + 1}-${element.id}`" />
                                    </div>
                                    <v-btn class="ms-2" density="compact" color="primary" variant="text"
                                           :aria-label="tt('Apply')" :disabled="loading" :icon="true"
                                           @click="updateRuleName(element)" v-if="editingRule === element">
                                        <v-icon :icon="mdiCheck" size="18" />
                                        <v-tooltip activator="parent">{{ tt('Apply') }}</v-tooltip>
                                    </v-btn>
                                    <v-btn class="ms-1" density="compact" color="default" variant="text"
                                           :aria-label="tt('Cancel')" :disabled="loading" :icon="true"
                                           @click="cancelUpdateRuleName" v-if="editingRule === element">
                                        <v-icon :icon="mdiClose" size="18" />
                                        <v-tooltip activator="parent">{{ tt('Cancel') }}</v-tooltip>
                                    </v-btn>
                                    <v-btn class="ms-2" density="compact" color="default" variant="text"
                                           :aria-label="tt('Modify Rule Name')" :disabled="loading || !!editingRule" :icon="true"
                                           @click="editingRuleName = element.name; editingRule = element" v-if="editingRule !== element">
                                        <v-icon :icon="mdiPencilOutline" size="18" />
                                        <v-tooltip activator="parent">{{ tt('Modify Rule Name') }}</v-tooltip>
                                    </v-btn>
                                    <v-spacer />
                                    <v-btn density="compact" color="default" variant="text" :aria-label="tt('Delete')"
                                           :disabled="loading || !!editingRule" :icon="true" @click="removeRule(index)">
                                        <v-icon :icon="mdiClose" size="18" />
                                        <v-tooltip activator="parent">{{ tt('Delete') }}</v-tooltip>
                                    </v-btn>
                                    <span class="ms-1">
                                        <v-icon :class="!loading && !editingRule && rules.length > 1 ? 'drag-handle' : 'disabled'"
                                                :aria-label="tt('Drag to Reorder')" :icon="mdiDrag" />
                                        <v-tooltip activator="parent" v-if="!loading && !editingRule && rules.length > 1">{{ tt('Drag to Reorder') }}</v-tooltip>
                                    </span>
                                </v-card-title>
                                <v-divider />
                                <v-card-text>
                                    <div class="text-body-small text-medium-emphasis mb-2">{{ tt('Match Condition') }}</div>
                                    <div class="d-flex overflow-x-auto align-center gap-2 mb-5">
                                        <v-select class="flex-0-0" min-width="220" density="compact"
                                                  item-title="name" item-value="value"
                                                  :disabled="loading || !!editingRule"
                                                  :items="allConditionFields"
                                                  :model-value="element.conditionField"
                                                  @update:model-value="updateConditionField(element, $event)" />

                                        <v-select class="flex-0-0" min-width="140" density="compact"
                                                  item-title="name" item-value="value"
                                                  :disabled="loading || !!editingRule"
                                                  :items="getConditionOperators(element)"
                                                  :model-value="element.conditionOperator"
                                                  @update:model-value="updateConditionOperator(element, $event)" />

                                        <v-autocomplete class="flex-1-1" min-width="240" density="compact"
                                                        item-title="name" item-value="value"
                                                        persistent-placeholder
                                                        :disabled="loading || !!editingRule"
                                                        :no-data-text="tt('No results')"
                                                        :items="getSourceItems(element.conditionField)"
                                                        :model-value="element.conditionValue"
                                                        @update:model-value="updateStringConditionValue(element, $event)"
                                                        v-if="isItemCondition(element.conditionField)" />

                                        <div class="d-flex flex-1-1 align-center gap-2" style="min-width: 280px"
                                             v-else-if="isAmountCondition(element.conditionField)">
                                            <amount-input density="compact" :currency="defaultCurrency" :disabled="loading || !!editingRule"
                                                          :model-value="getAmountConditionValue(element, 0)"
                                                          @update:model-value="updateAmountConditionValue(element, 0, $event)" />
                                            <span class="ms-2 me-2" v-if="isRangeOperator(element.conditionOperator)">{{ tt('format.misc.rangeSeparator') }}</span>
                                            <amount-input density="compact" :currency="defaultCurrency" :disabled="loading || !!editingRule"
                                                          :model-value="getAmountConditionValue(element, 1)"
                                                          @update:model-value="updateAmountConditionValue(element, 1, $event)"
                                                          v-if="isRangeOperator(element.conditionOperator)" />
                                        </div>

                                        <v-text-field class="flex-1-1" min-width="240" density="compact" disabled
                                                      :placeholder="tt('None')"
                                                      v-else-if="isEmptyTextConditionOperator(element.conditionOperator)" />

                                        <v-text-field class="flex-1-1" min-width="240" autocomplete="off" density="compact"
                                                      :disabled="loading || !!editingRule"
                                                      :placeholder="tt('None')" :model-value="element.conditionValue"
                                                      @update:model-value="updateStringConditionValue(element, $event)" v-else />
                                    </div>

                                    <div class="text-body-small text-medium-emphasis mb-2">{{ tt('Action') }}</div>
                                    <div class="d-flex overflow-x-auto align-center gap-2">

                                        <v-select class="flex-0-0" min-width="220" density="compact" item-title="name" item-value="value"
                                                  :disabled="loading || !!editingRule" :items="allActions" :model-value="element.actionType"
                                                  @update:model-value="updateActionType(element, $event)">
                                            <template #selection>
                                                <span>{{ tt(ImportTransactionReplaceRuleAction.valueOf(element.actionType)?.name || '') }}</span>
                                            </template>
                                            <template #item="{ props, internalItem }">
                                                <v-list-item :disabled="isActionDisabledForRule(internalItem.raw.value, element)" v-bind="props">
                                                    <template #title>
                                                        <div class="text-truncate">{{ internalItem.raw.name }}</div>
                                                    </template>
                                                </v-list-item>
                                            </template>
                                        </v-select>

                                        <v-select class="flex-1-1" min-width="280" density="compact" item-title="displayName" item-value="type"
                                                  :disabled="loading || !!editingRule" :items="transactionTypes"
                                                  v-model="element.targetValue" v-if="element.actionType === ImportTransactionReplaceRuleActionType.SetTransactionType" />

                                        <two-column-select class="flex-1-1" min-width="280" density="compact"
                                                           primary-key-field="id" primary-value-field="id" primary-title-field="name"
                                                           primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                                                           primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                                           secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                                           secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                                                           secondary-hidden-field="hidden" :disabled="loading || !!editingRule"
                                                           :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                                           :show-selection-primary-text="true" :custom-selection-primary-text="getSelectedCategoryPrimaryName(element)"
                                                           :custom-selection-secondary-text="getSelectedCategorySecondaryName(element)"
                                                           :items="allCategories[getActionCategoryType(element.actionType)]"
                                                           v-model="element.targetValue"
                                                           v-if="element.actionType === ImportTransactionReplaceRuleActionType.SetExpenseCategory || element.actionType === ImportTransactionReplaceRuleActionType.SetIncomeCategory || element.actionType === ImportTransactionReplaceRuleActionType.SetTransferCategory" />

                                        <two-column-select class="flex-1-1" min-width="280" density="compact"
                                                           primary-key-field="id" primary-value-field="category" primary-title-field="name"
                                                           primary-footer-field="displayBalance" primary-icon-field="icon" primary-icon-type-field="iconType"
                                                           primary-icon-type="account" primary-sub-items-field="accounts" :primary-title-i18n="true"
                                                           secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                                           secondary-footer-field="displayBalance" secondary-icon-field="icon" secondary-icon-type-field="iconType"
                                                           secondary-icon-type="account" secondary-color-field="color"
                                                           :disabled="loading || !!editingRule || !allVisibleAccounts.length" :enable-filter="true"
                                                           :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                           :custom-selection-primary-text="getAccountDisplayName(element.targetValue)"
                                                           :items="allVisibleCategorizedAccounts" v-model="element.targetValue"
                                                           v-else-if="element.actionType === ImportTransactionReplaceRuleActionType.SetSourceAccount || element.actionType === ImportTransactionReplaceRuleActionType.SetDestinationAccount" />

                                        <v-autocomplete class="flex-1-1" min-width="280" density="compact" item-title="name" item-value="id"
                                                        persistent-placeholder chips :disabled="loading || !!editingRule"
                                                        :items="allTagsWithGroupHeader" :no-data-text="tt('No available tag')"
                                                        v-model="element.targetValue" v-else-if="element.actionType === ImportTransactionReplaceRuleActionType.AddTag || element.actionType === ImportTransactionReplaceRuleActionType.ReplaceTag">
                                            <template #chip="{ props, internalItem }">
                                                <v-chip :prepend-icon="mdiPound" :text="internalItem.title" v-bind="props" v-if="element.targetValue" />
                                            </template>
                                            <template #subheader="{ props }"><v-list-subheader class="text-body-small">{{ props['title'] }}</v-list-subheader></template>
                                            <template #item="{ props, internalItem }">
                                                <v-list-item :value="internalItem.value" v-bind="props"
                                                             v-if="internalItem.raw instanceof TransactionTag && !internalItem.raw.hidden">
                                                    <template #title>
                                                        <v-list-item-title>
                                                            <div class="d-flex align-center">
                                                                <v-icon size="20" start :icon="mdiPound" />
                                                                <span>{{ internalItem.title }}</span>
                                                            </div>
                                                        </v-list-item-title>
                                                    </template>
                                                </v-list-item>
                                            </template>
                                        </v-autocomplete>

                                        <v-select class="flex-1-1" min-width="280" density="compact" item-title="name" item-value="value"
                                                  :disabled="loading || !!editingRule" :items="amountSigns"
                                                  v-model="element.targetValue"
                                                  v-else-if="element.actionType === ImportTransactionReplaceRuleActionType.SetSourceAmount || element.actionType === ImportTransactionReplaceRuleActionType.SetDestinationAmount" />

                                        <v-autocomplete class="flex-1-1" min-width="280" density="compact"
                                                        item-title="displayNameWithUtcOffset" item-value="name"
                                                        persistent-placeholder auto-select-first
                                                        :disabled="loading || !!editingRule"
                                                        :items="allTimezones"
                                                        v-model="element.targetValue"
                                                        v-else-if="element.actionType === ImportTransactionReplaceRuleActionType.SetTimezone" />
                                    </div>
                                </v-card-text>
                            </v-card>
                        </template>
                    </draggable-list>
                </div>
            </template>
        </one-column-dialog-layout>
    </v-dialog>
    <snack-bar ref="snackbar" />
</template>



<style>
.rule-name {
    white-space: pre;
}

.rule-card-title {
    min-height: 52px;
}

.rule-name-edit {
    display: flex;
    align-items: center;
    height: 36px;
}

.rule-name-edit .v-field {
    font-size: 0.9375rem;
}

.rule-name-edit .v-field__input {
    padding-top: 0;
}
</style>
