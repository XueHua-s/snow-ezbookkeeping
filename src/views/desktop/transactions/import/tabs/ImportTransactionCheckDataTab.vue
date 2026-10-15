<script setup lang="ts">
import ImportCategoryCell from './ImportTransactionCheckDataTab/ImportCategoryCell.vue';
import PaginationButtons from '@/components/desktop/PaginationButtons.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { TRANSACTION_MAX_COMMENT_LENGTH } from '@/consts/transaction.ts';

import { TransactionType } from '@/core/transaction.ts';

import { ImportTransaction } from '@/models/imported_transaction.ts';
import { TransactionTag } from '@/models/transaction_tag.ts';
import {
    mdiAlertOutline,
    mdiArrowRight,
    mdiCheck,
    mdiPencilOutline,
    mdiPound,
    mdiSelect,
    mdiSelectAll,
    mdiSelectInverse
} from '@mdi/js';
import BatchApplyRulesDialog from '../dialogs/BatchApplyRulesDialog.vue';
import BatchCreateDialog from '../dialogs/BatchCreateDialog.vue';
import BatchReplaceDialog from '../dialogs/BatchReplaceDialog.vue';
import { useImportTransactionCheckDataTab } from './ImportTransactionCheckDataTab/useImportTransactionCheckDataTab.ts';

const props = defineProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>();

const {
    tt, formatNumberToLocalizedNumerals, formatCoordinate, allTagsWithGroupHeader, editingTransaction, editingTags, filters,
    currentPage, countPerPage, showCustomDateRangeDialog, showCustomAmountFilterDialog, showCustomDescriptionDialog,
    currentAmountFilterType, currentAmountFilterValue1, currentAmountFilterValue2, currentDescriptionFilterValue,
    defaultCurrency, allVisibleAccounts, allVisibleCategorizedAccounts, allAccountsMap, allCategories, allCategoriesMap,
    allTagsMap, hasVisibleExpenseCategories, hasVisibleIncomeCategories, hasVisibleTransferCategories, isEditing, canImport,
    filterMenus, toolMenus, importTransactionHeaders, importTransactionsTablePageOptions, totalPageCount,
    selectedImportTransactionCount, selectedInvalidTransactionCount, anyButNotAllTransactionSelected, allTransactionSelected,
    isTagValid, getDisplayDateTime, isSameAsDefaultTimezoneOffsetMinutes, getDisplayTimezone,
    getTransactionSourceAccountCurrency, getTransactionDestinationAccountCurrency, getTransactionDisplayAmount,
    getTransactionDisplayDestinationAmount, getTransactionDisplaySourceAmountInDefaultCurrency,
    getTransactionDisplayDestinationAmountInDefaultCurrency, getSourceAccountTitle, getSourceAccountDisplayName,
    getDestinationAccountDisplayName, getTransactionDescriptionTooltip, importTransactionsFilter, selectAllValid,
    selectAllInvalid, selectAll, selectNone, selectInvert, selectAllInThisPage, selectNoneInThisPage, selectInvertInThisPage,
    editTransaction, updateAllTransactionsIsValid, showBatchApplyRulesDialog, changeCustomDateFilter,
    focusTableScrollContainerWhenNonEditing, onShowDateRangeError, reset, setCountPerPage,
} = useImportTransactionCheckDataTab(props);

defineExpose({
    filterMenus,
    toolMenus,
    isEditing,
    canImport,
    showBatchApplyRulesDialog,
    updateAllTransactionsIsValid,
    reset,
    setCountPerPage
});

</script>

<template>
    <div class="import-transaction-table-container d-flex flex-column">
        <v-data-table
            fixed-header
            fixed-footer
            height="100%"
            show-select
            multi-sort
            density="compact"
            item-value="index"
            :class="{ 'import-transaction-table': true, 'disabled': !!disabled }"
            :headers="importTransactionHeaders"
            :items="importTransactions"
            :hover="true"
            :search="JSON.stringify(filters)"
            :custom-filter="importTransactionsFilter"
            :no-data-text="tt('No data to import')"
            v-model:items-per-page="countPerPage"
            v-model:page="currentPage"
            @click="focusTableScrollContainerWhenNonEditing"
        >
            <template #header.data-table-select>
                <v-checkbox readonly class="always-cursor-pointer"
                            density="compact" width="28"
                            :disabled="!!disabled"
                            :indeterminate="anyButNotAllTransactionSelected"
                            v-model="allTransactionSelected"
                >
                    <v-menu activator="parent" location="bottom">
                        <v-list>
                            <v-list-item :prepend-icon="mdiSelectAll"
                                         :title="tt('Select All Valid Items')"
                                         :disabled="!!disabled"
                                         @click="selectAllValid"></v-list-item>
                            <v-list-item :prepend-icon="mdiSelectAll"
                                         :title="tt('Select All Invalid Items')"
                                         :disabled="!!disabled"
                                         @click="selectAllInvalid"></v-list-item>
                            <v-divider class="my-2"/>
                            <v-list-item :prepend-icon="mdiSelectAll"
                                         :title="tt('Select All')"
                                         :disabled="!!disabled"
                                         @click="selectAll"></v-list-item>
                            <v-list-item :prepend-icon="mdiSelect"
                                         :title="tt('Select None')"
                                         :disabled="!!disabled"
                                         @click="selectNone"></v-list-item>
                            <v-list-item :prepend-icon="mdiSelectInverse"
                                         :title="tt('Invert Selection')"
                                         :disabled="!!disabled"
                                         @click="selectInvert"></v-list-item>
                            <v-divider class="my-2"/>
                            <v-list-item :prepend-icon="mdiSelectAll"
                                         :title="tt('Select All on This Page')"
                                         :disabled="!!disabled"
                                         @click="selectAllInThisPage"></v-list-item>
                            <v-list-item :prepend-icon="mdiSelect"
                                         :title="tt('Select None on This Page')"
                                         :disabled="!!disabled"
                                         @click="selectNoneInThisPage"></v-list-item>
                            <v-list-item :prepend-icon="mdiSelectInverse"
                                         :title="tt('Invert Selection on This Page')"
                                         :disabled="!!disabled"
                                         @click="selectInvertInThisPage"></v-list-item>
                        </v-list>
                    </v-menu>
                </v-checkbox>
            </template>
            <template #item.data-table-select="{ item }">
                <v-checkbox density="compact"
                            :color="!item.valid ? 'error' : 'primary'"
                            :disabled="!!disabled"
                            v-model="item.selected"></v-checkbox>
            </template>
            <template #item.valid="{ item }">
                <v-icon size="small" :class="{ 'text-error': !item.valid }"
                        :aria-label="editingTransaction === item ? tt('Apply') : tt('Edit')"
                        :disabled="!!disabled"
                        :icon="editingTransaction === item ? mdiCheck : mdiPencilOutline"
                        @click="editTransaction(item)">
                </v-icon>
                <v-tooltip activator="parent" v-if="!disabled">{{ editingTransaction === item ? tt('Apply') : tt('Edit') }}</v-tooltip>
            </template>
            <template #item.time="{ item }">
                <div class="d-flex align-center">
                    <span>{{ getDisplayDateTime(item) }}</span>
                    <v-chip class="ms-1" variant="flat" color="grey" size="x-small"
                            v-if="!isSameAsDefaultTimezoneOffsetMinutes(item)">{{ getDisplayTimezone(item) }}</v-chip>
                </div>
            </template>
            <template #item.type="{ value }">
                <v-chip label color="secondary" variant="outlined" size="x-small" v-if="value === TransactionType.ModifyBalance">{{ tt('Modify Balance') }}</v-chip>
                <v-chip label class="text-income" variant="outlined" size="x-small" v-else-if="value === TransactionType.Income">{{ tt('Income') }}</v-chip>
                <v-chip label class="text-expense" variant="outlined" size="x-small" v-else-if="value === TransactionType.Expense">{{ tt('Expense') }}</v-chip>
                <v-chip label color="primary" variant="outlined" size="x-small" v-else-if="value === TransactionType.Transfer">{{ tt('Transfer') }}</v-chip>
                <v-chip label color="default" variant="outlined" size="x-small" v-else>{{ tt('Unknown') }}</v-chip>
            </template>
            <template #item.actualCategoryName="{ item }">
<ImportCategoryCell
    :editingTransaction="editingTransaction"
    :allCategoriesMap="allCategoriesMap"
    :disabled="disabled"
    :hasVisibleExpenseCategories="hasVisibleExpenseCategories"
    :allCategories="allCategories"
    :hasVisibleIncomeCategories="hasVisibleIncomeCategories"
    :hasVisibleTransferCategories="hasVisibleTransferCategories"
    :item="item"
    v-model:categoryId="item.categoryId"
/>
</template>
            <template #item.sourceAmount="{ item }">
                <div class="d-flex align-center" v-if="editingTransaction !== item">
                    <span>{{ getTransactionDisplayAmount(item) }}</span>
                    <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" v-if="item.type === TransactionType.Transfer && item.sourceAccountId !== item.destinationAccountId"></v-icon>
                    <span v-if="item.type === TransactionType.Transfer && item.sourceAccountId !== item.destinationAccountId">{{ getTransactionDisplayDestinationAmount(item) }}</span>
                    <v-tooltip activator="parent" v-if="(item.type !== TransactionType.Transfer && getTransactionSourceAccountCurrency(item) !== defaultCurrency) || (item.type === TransactionType.Transfer && getTransactionSourceAccountCurrency(item) !== defaultCurrency && getTransactionDestinationAccountCurrency(item) !== defaultCurrency)">
                        <span>{{ getTransactionDisplaySourceAmountInDefaultCurrency(item) }}</span>
                        <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" v-if="item.type === TransactionType.Transfer && item.sourceAccountId !== item.destinationAccountId && getTransactionSourceAccountCurrency(item) !== getTransactionDestinationAccountCurrency(item) && item.sourceAmount !== item.destinationAmount"></v-icon>
                        <span v-if="item.type === TransactionType.Transfer && item.sourceAccountId !== item.destinationAccountId && getTransactionSourceAccountCurrency(item) !== getTransactionDestinationAccountCurrency(item) && item.sourceAmount !== item.destinationAmount">{{ getTransactionDisplayDestinationAmountInDefaultCurrency(item) }}</span>
                    </v-tooltip>
                </div>
                <div class="d-flex align-center" :style="`width: ${item.type === TransactionType.Transfer && item.sourceAccountId !== item.destinationAccountId ? 250 : 100}px`" v-if="editingTransaction === item">
                    <amount-input density="compact" variant="plain"
                                  persistent-placeholder
                                  :currency="getTransactionSourceAccountCurrency(item)"
                                  :show-currency="true"
                                  :disabled="!!disabled"
                                  :placeholder="tt('Amount')"
                                  v-model="item.sourceAmount"/>
                    <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" v-if="item.type === TransactionType.Transfer && item.sourceAccountId !== item.destinationAccountId"></v-icon>
                    <amount-input density="compact" variant="plain"
                                  persistent-placeholder
                                  :currency="getTransactionDestinationAccountCurrency(item)"
                                  :show-currency="true"
                                  :disabled="!!disabled"
                                  :placeholder="tt('Transfer In Amount')"
                                  v-model="item.destinationAmount"
                                  v-if="item.type === TransactionType.Transfer && item.sourceAccountId !== item.destinationAccountId"/>
                </div>
            </template>
            <template #item.actualSourceAccountName="{ item }">
                <div class="d-flex align-center" v-if="editingTransaction !== item">
                    <span v-if="item.sourceAccountId && item.sourceAccountId !== '0' && allAccountsMap[item.sourceAccountId]">{{ allAccountsMap[item.sourceAccountId]?.name }}</span>
                    <div class="text-error font-italic" v-else>
                        <v-icon class="me-1" :icon="mdiAlertOutline"/>
                        <span>{{ item.originalSourceAccountName }}</span>
                    </div>
                    <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" v-if="item.type === TransactionType.Transfer"></v-icon>
                    <span v-if="item.type === TransactionType.Transfer && item.destinationAccountId && item.destinationAccountId !== '0' && allAccountsMap[item.destinationAccountId]">{{allAccountsMap[item.destinationAccountId]?.name }}</span>
                    <div class="text-error font-italic" v-else-if="item.type === TransactionType.Transfer && (!item.destinationAccountId || item.destinationAccountId === '0' || !allAccountsMap[item.destinationAccountId])">
                        <v-icon class="me-1" :icon="mdiAlertOutline"/>
                        <span>{{ item.originalDestinationAccountName }}</span>
                    </div>
                </div>
                <div class="d-flex align-center" :style="`width: ${item.type === TransactionType.Transfer ? 450 : 200}px`"  v-if="editingTransaction === item">
                    <two-column-select density="compact" variant="plain"
                                       primary-key-field="id" primary-value-field="category"
                                       primary-title-field="name" primary-footer-field="displayBalance"
                                       primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="account"
                                       primary-sub-items-field="accounts"
                                       :primary-title-i18n="true"
                                       secondary-key-field="id" secondary-value-field="id"
                                       secondary-title-field="name" secondary-footer-field="displayBalance"
                                       secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="account" secondary-color-field="color"
                                       :disabled="!!disabled || !allVisibleAccounts.length"
                                       :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                       :custom-selection-primary-text="getSourceAccountDisplayName(item)"
                                       :placeholder="getSourceAccountTitle(item)"
                                       :items="allVisibleCategorizedAccounts"
                                       v-model="item.sourceAccountId">
                    </two-column-select>
                    <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" v-if="item.type === TransactionType.Transfer"></v-icon>
                    <two-column-select density="compact" variant="plain"
                                       primary-key-field="id" primary-value-field="category"
                                       primary-title-field="name" primary-footer-field="displayBalance"
                                       primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="account"
                                       primary-sub-items-field="accounts"
                                       :primary-title-i18n="true"
                                       secondary-key-field="id" secondary-value-field="id"
                                       secondary-title-field="name" secondary-footer-field="displayBalance"
                                       secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="account" secondary-color-field="color"
                                       :disabled="!!disabled || !allVisibleAccounts.length"
                                       :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                       :custom-selection-primary-text="getDestinationAccountDisplayName(item)"
                                       :placeholder="tt('Destination Account')"
                                       :items="allVisibleCategorizedAccounts"
                                       v-model="item.destinationAccountId"
                                       v-if="item.type === TransactionType.Transfer">
                    </two-column-select>
                </div>
            </template>
            <template #item.geoLocation="{ item }">
                <span v-if="item.geoLocation">{{ `(${formatCoordinate(item.geoLocation)})` }}</span>
                <span v-else-if="!item.geoLocation">{{ tt('None') }}</span>
            </template>
            <template #item.tagIds="{ item }">
                <div v-if="editingTransaction !== item">
                    <v-chip class="transaction-tag" size="small"
                            :class="{ 'font-italic': !tagId || tagId === '0' || !allTagsMap[tagId] }"
                            :prepend-icon="tagId && tagId !== '0' && allTagsMap[tagId] ? mdiPound : mdiAlertOutline"
                            :color="tagId && tagId !== '0' && allTagsMap[tagId] ? 'default' : 'error'"
                            :text="tagId && tagId !== '0' && allTagsMap[tagId] ? allTagsMap[tagId].name : item.originalTagNames[index]"
                            :key="tagId"
                            v-for="(tagId, index) in item.tagIds"/>
                    <v-chip class="transaction-tag" size="small"
                            :text="tt('None')"
                            v-if="!item.tagIds || !item.tagIds.length"/>
                </div>
                <div style="width: 200px" v-if="editingTransaction === item">
                    <v-autocomplete
                        item-title="name"
                        item-value="id"
                        auto-select-first
                        persistent-placeholder
                        multiple
                        chips
                        closable-chips
                        density="compact" variant="plain"
                        :disabled="!!disabled"
                        :placeholder="tt('None')"
                        :items="allTagsWithGroupHeader"
                        :no-data-text="tt('No available tag')"
                        v-model="editingTags"
                    >
                        <template #chip="{ props, index }">
                            <v-chip :class="{ 'font-italic': !isTagValid(editingTags, index) }"
                                    :prepend-icon="isTagValid(editingTags, index) ? mdiPound : mdiAlertOutline"
                                    :color="isTagValid(editingTags, index) ? 'default' : 'error'"
                                    :text="isTagValid(editingTags, index) ? allTagsMap[editingTags[index] as string]?.name : item.originalTagNames[index]"
                                    v-bind="props"/>
                        </template>

                        <template #subheader="{ props }">
                            <v-list-subheader class="text-body-small">{{ props['title'] }}</v-list-subheader>
                        </template>

                        <template #item="{ props, internalItem }">
                            <v-list-item :value="internalItem.value" v-bind="props" v-if="internalItem.raw instanceof TransactionTag && !internalItem.raw.hidden">
                                <template #title>
                                    <v-list-item-title>
                                        <div class="d-flex align-center">
                                            <v-icon size="20" start :icon="mdiPound"/>
                                            <span>{{ internalItem.title }}</span>
                                        </div>
                                    </v-list-item-title>
                                </template>
                            </v-list-item>
                        </template>
                    </v-autocomplete>
                </div>
            </template>
            <template #item.comment="{ item }">
                <template v-if="editingTransaction !== item">
                    <span v-if="!item.comment || item.comment.length <= TRANSACTION_MAX_COMMENT_LENGTH">{{ item.comment || '' }}</span>
                    <div class="text-error font-italic" v-else-if="item.comment && item.comment.length > TRANSACTION_MAX_COMMENT_LENGTH">
                        <v-tooltip activator="parent">{{ getTransactionDescriptionTooltip(item) }}</v-tooltip>
                        <v-icon class="me-1" :icon="mdiAlertOutline"/>
                        <span>{{ item.comment }}</span>
                    </div>
                </template>
                <div v-if="editingTransaction === item">
                    <v-text-field style="width: calc(max(300px, 100%))" type="text" autocomplete="off"
                                  density="compact" variant="plain"
                                  persistent-placeholder
                                  :placeholder="tt('Description')"
                                  :disabled="!!disabled"
                                  v-model="item.comment">
                        <v-tooltip activator="parent" v-if="item.comment && item.comment.length > TRANSACTION_MAX_COMMENT_LENGTH">
                            {{ getTransactionDescriptionTooltip(item) }}
                        </v-tooltip>
                    </v-text-field>
                </div>
            </template>
            <template #bottom>
            </template>
        </v-data-table>

        <div class="import-transaction-table-footer">
            <v-divider />
            <div class="title-and-toolbar d-flex text-body-large align-center text-no-wrap my-1 mx-3" v-if="importTransactions">
                <span :class="{ 'text-error': selectedInvalidTransactionCount > 0 }">
                    {{ tt('format.misc.selectedCount', { count: formatNumberToLocalizedNumerals(selectedImportTransactionCount), totalCount: formatNumberToLocalizedNumerals(importTransactions.length) }) }}
                </span>
                <v-spacer v-if="importTransactions.length > 10"/>
                <span class="ms-2" v-if="importTransactions.length > 10">{{ tt('Transactions Per Page') }}</span>
                <v-select class="ms-2" density="compact" max-width="100"
                          item-title="name"
                          item-value="value"
                          :disabled="!!disabled"
                          :items="importTransactionsTablePageOptions"
                          v-model="countPerPage"
                          v-if="importTransactions.length > 10"
                />
                <pagination-buttons density="comfortable"
                                    :disabled="!!disabled"
                                    :totalPageCount="totalPageCount"
                                    v-model="currentPage"
                                    v-if="importTransactions.length > 10"></pagination-buttons>
            </div>
        </div>
    </div>

    <v-dialog width="640" v-model="showCustomAmountFilterDialog">
        <one-column-dialog-layout :title="tt('Filter Amount')" :cancel-button-title="tt('Cancel')"
                                  @cancel="showCustomAmountFilterDialog = false">
            <template #toolbar>
                <v-btn class="mx-2" density="comfortable" variant="outlined"
                       @click="showCustomAmountFilterDialog = false; filters.amount = currentAmountFilterType?.toTextualFilter(currentAmountFilterValue1, currentAmountFilterValue2) ?? null">{{ tt('Apply') }}</v-btn>
            </template>

            <template #content>
                <div class="w-100 mt-1 d-flex justify-center">
                    <div class="me-2 d-flex flex-column justify-center" v-if="currentAmountFilterType">
                        {{ tt(currentAmountFilterType.name) }}
                    </div>
                    <amount-input :currency="defaultCurrency"
                                  v-model="currentAmountFilterValue1"/>
                    <div class="ms-2 me-2 d-flex flex-column justify-center" v-if="currentAmountFilterType && currentAmountFilterType.paramCount === 2">
                        {{ tt('format.misc.rangeSeparator') }}
                    </div>
                    <amount-input :currency="defaultCurrency"
                                  v-model="currentAmountFilterValue2"
                                  v-if="currentAmountFilterType && currentAmountFilterType.paramCount === 2"/>
                </div>
            </template>
        </one-column-dialog-layout>
    </v-dialog>

    <v-dialog width="640" v-model="showCustomDescriptionDialog">
        <one-column-dialog-layout :title="tt('Filter Description')" :cancel-button-title="tt('Cancel')"
                                  @cancel="showCustomDescriptionDialog = false; currentDescriptionFilterValue = ''">
            <template #toolbar>
                <v-btn class="mx-2" density="comfortable" variant="outlined"
                       :disabled="!currentDescriptionFilterValue"
                       @click="showCustomDescriptionDialog = false; filters.description = currentDescriptionFilterValue">{{ tt('Apply') }}</v-btn>
            </template>

            <template #content>
                <div class="mt-2">
                    <v-text-field
                        type="text"
                        autocomplete="off"
                        persistent-placeholder
                        :label="tt('Description')"
                        :placeholder="tt('Description')"
                        v-model="currentDescriptionFilterValue"
                    />
                </div>
            </template>
        </one-column-dialog-layout>
    </v-dialog>

    <date-range-selection-dialog :title="tt('Custom Date Range')"
                                 :min-time="filters.minDatetime"
                                 :max-time="filters.maxDatetime"
                                 v-model:show="showCustomDateRangeDialog"
                                 @dateRange:change="changeCustomDateFilter"
                                 @error="onShowDateRangeError" />
    <batch-replace-dialog ref="batchReplaceDialog" />
    <batch-apply-rules-dialog ref="batchApplyRulesDialog" />
    <batch-create-dialog ref="batchCreateDialog" />
    <snack-bar ref="snackbar" />
</template>



<style>
.import-transaction-table > .v-table__wrapper > table {
    th:not(:last-child),
    td:not(:last-child) {
        width: auto !important;
        white-space: nowrap;
    }

    th:last-child,
    td:last-child {
        width: 100% !important;
    }
}

.import-transaction-table .v-text-field.v-input.v-input--density-compact:not(.v-textarea),
.import-transaction-table .v-autocomplete.v-input.v-input--density-compact:not(.v-textarea),
.import-transaction-table .v-select.v-input.v-input--density-compact:not(.v-textarea) {
    .v-field__input {
        min-height: inherit;
        padding-top: 4px;
    }
}

.import-transaction-table .amount-input.v-input.v-input--density-compact {
    .v-field__prepend-inner {
        padding-top: 3px;
    }

    .v-field__input {
        padding-inline-start: 0.2rem;
    }
}

.import-transaction-table .v-chip.transaction-tag {
    margin-inline-end: 4px;
    margin-top: 2px;
    margin-bottom: 2px;
}

.import-transaction-table .v-chip.transaction-tag > .v-chip__content {
    display: block;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
}

.import-transaction-table .v-text-field.v-input.v-input--density-compact .v-field__input {
    padding-top: 0;
}
</style>
