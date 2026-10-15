<script setup lang="ts">
import { TRANSACTION_MAX_AMOUNT, TRANSACTION_MIN_AMOUNT } from '@/consts/transaction.ts';
import { CategoryType } from '@/core/category.ts';
import { TemplateType } from '@/core/template.ts';
import {
    TransactionType
} from '@/core/transaction.ts';
import { getTransactionPrimaryCategoryName, getTransactionSecondaryCategoryName } from '@/lib/category.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { useI18n } from '@/locales/helpers.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import {
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { useEditPage } from '@/views/mobile/transactions/EditPage/useEditPage.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useEditPage>>;

defineProps<{
    loading: Scope['loading'];
    pageTypeAndMode: Scope['pageTypeAndMode'];
    sourceAmountClass: Scope['sourceAmountClass'];
    sourceAmountTitle: Scope['sourceAmountTitle'];
    getDisplayAmount: Scope['getDisplayAmount'];
    sourceAccountCurrency: Scope['sourceAccountCurrency'];
    destinationAmountClass: Scope['destinationAmountClass'];
    transferInAmountTitle: Scope['transferInAmountTitle'];
    destinationAccountCurrency: Scope['destinationAccountCurrency'];
    hasVisibleExpenseCategories: Scope['hasVisibleExpenseCategories'];
    mode: Scope['mode'];
    allCategories: Scope['allCategories'];
    hasVisibleIncomeCategories: Scope['hasVisibleIncomeCategories'];
    hasVisibleTransferCategories: Scope['hasVisibleTransferCategories'];
    allVisibleAccounts: Scope['allVisibleAccounts'];
    sourceAccountTitle: Scope['sourceAccountTitle'];
    sourceAccountName: Scope['sourceAccountName'];
    allVisibleCategorizedAccounts: Scope['allVisibleCategorizedAccounts'];
    destinationAccountName: Scope['destinationAccountName'];
    currentTimezoneOffsetMinutes: Scope['currentTimezoneOffsetMinutes'];
    transactionDisplayDate: Scope['transactionDisplayDate'];
    transactionDisplayTime: Scope['transactionDisplayTime'];
    transactionDateTimeSheetMode: Scope['transactionDateTimeSheetMode'];
    transactionDisplayScheduledFrequency: Scope['transactionDisplayScheduledFrequency'];
    transactionDisplayScheduledStartDate: Scope['transactionDisplayScheduledStartDate'];
    transactionDisplayScheduledEndDate: Scope['transactionDisplayScheduledEndDate'];
    transactionDisplayTimezone: Scope['transactionDisplayTimezone'];
    transactionDisplayTimezoneName: Scope['transactionDisplayTimezoneName'];
    transactionTimezoneTimeDifference: Scope['transactionTimezoneTimeDifference'];
    allTimezones: Scope['allTimezones'];
    geoLocationStatusInfo: Scope['geoLocationStatusInfo'];
    allTagsMap: Scope['allTagsMap'];
    showTransactionPictures: Scope['showTransactionPictures'];
    submitting: Scope['submitting'];
    uploadingPicture: Scope['uploadingPicture'];
    removingPictureId: Scope['removingPictureId'];
    getTransactionPictureUrl: Scope['getTransactionPictureUrl'];
    canAddTransactionPicture: Scope['canAddTransactionPicture'];
    transactionDescriptionTitle: Scope['transactionDescriptionTitle'];
}>();

const transaction = defineModel<Scope['transaction']>('transaction', { required: true });
const showSourceAmountSheet = defineModel<Scope['showSourceAmountSheet']>('showSourceAmountSheet', { required: true });
const showDestinationAmountSheet = defineModel<Scope['showDestinationAmountSheet']>('showDestinationAmountSheet', { required: true });
const showCategorySheet = defineModel<Scope['showCategorySheet']>('showCategorySheet', { required: true });
const showSourceAccountSheet = defineModel<Scope['showSourceAccountSheet']>('showSourceAccountSheet', { required: true });
const showDestinationAccountSheet = defineModel<Scope['showDestinationAccountSheet']>('showDestinationAccountSheet', { required: true });
const showTransactionDateTimeSheet = defineModel<Scope['showTransactionDateTimeSheet']>('showTransactionDateTimeSheet', { required: true });
const showTransactionScheduledFrequencySheet = defineModel<Scope['showTransactionScheduledFrequencySheet']>('showTransactionScheduledFrequencySheet', { required: true });
const showScheduledStartDateSheet = defineModel<Scope['showScheduledStartDateSheet']>('showScheduledStartDateSheet', { required: true });
const showScheduledEndDateSheet = defineModel<Scope['showScheduledEndDateSheet']>('showScheduledEndDateSheet', { required: true });
const showTimezonePopup = defineModel<Scope['showTimezonePopup']>('showTimezonePopup', { required: true });
const setGeoLocationByClickMap = defineModel<Scope['setGeoLocationByClickMap']>('setGeoLocationByClickMap', { required: true });
const showGeoLocationMapSheet = defineModel<Scope['showGeoLocationMapSheet']>('showGeoLocationMapSheet', { required: true });
const showTransactionTagSheet = defineModel<Scope['showTransactionTagSheet']>('showTransactionTagSheet', { required: true });
const showGeoLocationActionSheet = defineModel<Scope['showGeoLocationActionSheet']>('showGeoLocationActionSheet', { required: true });

const emit = defineEmits<{
    showDateTimeDialog: Parameters<Scope['showDateTimeDialog']>;
    updateTransactionTime: Parameters<Scope['updateTransactionTime']>;
    updateTransactionTimezone: Parameters<Scope['updateTransactionTimezone']>;
    viewOrRemovePicture: Parameters<Scope['viewOrRemovePicture']>;
    showOpenPictureDialog: Parameters<Scope['showOpenPictureDialog']>;
}>();

const showDateTimeDialog = (...args: Parameters<Scope['showDateTimeDialog']>) => emit('showDateTimeDialog', ...args);
const updateTransactionTime = (...args: Parameters<Scope['updateTransactionTime']>) => emit('updateTransactionTime', ...args);
const updateTransactionTimezone = (...args: Parameters<Scope['updateTransactionTimezone']>) => emit('updateTransactionTimezone', ...args);
const viewOrRemovePicture = (...args: Parameters<Scope['viewOrRemovePicture']>) => emit('viewOrRemovePicture', ...args);
const showOpenPictureDialog = (...args: Parameters<Scope['showOpenPictureDialog']>) => emit('showOpenPictureDialog', ...args);

const { tt, formatCoordinate } = useI18n();

</script>

<template>
    <f7-list form strong inset dividers class="margin-vertical-half" >
        <f7-list-input
            type="text"
            autocomplete="off"
            clear-button
            :label="tt('Template Name')"
            :placeholder="tt('Template Name')"
            v-model:value="transaction.name"
            v-if="pageTypeAndMode?.type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate"
        ></f7-list-input>

        <f7-list-item
            class="transaction-edit-amount"
            link="#" no-chevron
            :class="sourceAmountClass"
            :header="sourceAmountTitle"
            :title="getDisplayAmount(parseBigDecimal(transaction.sourceAmount), transaction.hideAmount, sourceAccountCurrency)"
            @click="showSourceAmountSheet = true"
        >
            <number-pad-sheet :min-value="TRANSACTION_MIN_AMOUNT"
                              :max-value="TRANSACTION_MAX_AMOUNT"
                              :currency="sourceAccountCurrency"
                              v-model:show="showSourceAmountSheet"
                              v-model="transaction.sourceAmount"
            ></number-pad-sheet>
        </f7-list-item>

        <f7-list-item
            class="transaction-edit-amount text-color-primary"
            link="#" no-chevron
            :class="destinationAmountClass"
            :header="transferInAmountTitle"
            :title="getDisplayAmount(parseBigDecimal(transaction.destinationAmount), transaction.hideAmount, destinationAccountCurrency)"
            @click="showDestinationAmountSheet = true"
            v-if="transaction.type === TransactionType.Transfer"
        >
            <number-pad-sheet :min-value="TRANSACTION_MIN_AMOUNT"
                              :max-value="TRANSACTION_MAX_AMOUNT"
                              :currency="destinationAccountCurrency"
                              v-model:show="showDestinationAmountSheet"
                              v-model="transaction.destinationAmount"
            ></number-pad-sheet>
        </f7-list-item>

        <f7-list-item
            class="list-item-with-header-and-title list-item-title-hide-overflow"
            key="expenseCategorySelection"
            link="#" no-chevron
            :class="{ 'disabled': !hasVisibleExpenseCategories, 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt('Category')"
            @click="showCategorySheet = true"
            v-if="transaction.type === TransactionType.Expense"
        >
            <template #title>
                <div class="list-item-custom-title" v-if="hasVisibleExpenseCategories">
                    <span>{{ getTransactionPrimaryCategoryName(transaction.expenseCategoryId, allCategories[CategoryType.Expense]) }}</span>
                    <f7-icon class="category-separate-icon icon-with-direction" f7="chevron_right"></f7-icon>
                    <span>{{ getTransactionSecondaryCategoryName(transaction.expenseCategoryId, allCategories[CategoryType.Expense]) }}</span>
                </div>
                <div class="list-item-custom-title" v-else-if="!hasVisibleExpenseCategories">
                    <span>{{ tt('None') }}</span>
                </div>
            </template>
            <tree-view-selection-sheet primary-key-field="id" primary-title-field="name"
                                       primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                                       primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                       secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                       secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                                       secondary-hidden-field="hidden"
                                       :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                       :items="allCategories[CategoryType.Expense]"
                                       v-model:show="showCategorySheet"
                                       v-model="transaction.expenseCategoryId">
            </tree-view-selection-sheet>
        </f7-list-item>

        <f7-list-item
            class="list-item-with-header-and-title list-item-title-hide-overflow"
            key="incomeCategorySelection"
            link="#" no-chevron
            :class="{ 'disabled': !hasVisibleIncomeCategories, 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt('Category')"
            @click="showCategorySheet = true"
            v-if="transaction.type === TransactionType.Income"
        >
            <template #title>
                <div class="list-item-custom-title" v-if="hasVisibleIncomeCategories">
                    <span>{{ getTransactionPrimaryCategoryName(transaction.incomeCategoryId, allCategories[CategoryType.Income]) }}</span>
                    <f7-icon class="category-separate-icon icon-with-direction" f7="chevron_right"></f7-icon>
                    <span>{{ getTransactionSecondaryCategoryName(transaction.incomeCategoryId, allCategories[CategoryType.Income]) }}</span>
                </div>
                <div class="list-item-custom-title" v-else-if="!hasVisibleIncomeCategories">
                    <span>{{ tt('None') }}</span>
                </div>
            </template>
            <tree-view-selection-sheet primary-key-field="id" primary-title-field="name"
                                       primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                                       primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                       secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                       secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                                       secondary-hidden-field="hidden"
                                       :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                       :items="allCategories[CategoryType.Income]"
                                       v-model:show="showCategorySheet"
                                       v-model="transaction.incomeCategoryId">
            </tree-view-selection-sheet>
        </f7-list-item>

        <f7-list-item
            class="list-item-with-header-and-title list-item-title-hide-overflow"
            key="transferCategorySelection"
            link="#" no-chevron
            :class="{ 'disabled': !hasVisibleTransferCategories, 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt('Category')"
            @click="showCategorySheet = true"
            v-if="transaction.type === TransactionType.Transfer"
        >
            <template #title>
                <div class="list-item-custom-title" v-if="hasVisibleTransferCategories">
                    <span>{{ getTransactionPrimaryCategoryName(transaction.transferCategoryId, allCategories[CategoryType.Transfer]) }}</span>
                    <f7-icon class="category-separate-icon icon-with-direction" f7="chevron_right"></f7-icon>
                    <span>{{ getTransactionSecondaryCategoryName(transaction.transferCategoryId, allCategories[CategoryType.Transfer]) }}</span>
                </div>
                <div class="list-item-custom-title" v-else-if="!hasVisibleTransferCategories">
                    <span>{{ tt('None') }}</span>
                </div>
            </template>
            <tree-view-selection-sheet primary-key-field="id" primary-title-field="name"
                                       primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="category" primary-color-field="color"
                                       primary-hidden-field="hidden" primary-sub-items-field="subCategories"
                                       secondary-key-field="id" secondary-value-field="id" secondary-title-field="name"
                                       secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="category" secondary-color-field="color"
                                       secondary-hidden-field="hidden"
                                       :enable-filter="true" :filter-placeholder="tt('Find category')" :filter-no-items-text="tt('No available category')"
                                       :items="allCategories[CategoryType.Transfer]"
                                       v-model:show="showCategorySheet"
                                       v-model="transaction.transferCategoryId">
            </tree-view-selection-sheet>
        </f7-list-item>

        <f7-list-item
            class="list-item-with-header-and-title"
            link="#" no-chevron
            :class="{ 'disabled': !allVisibleAccounts.length || (mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance), 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt(sourceAccountTitle)"
            :title="sourceAccountName"
            @click="showSourceAccountSheet = true"
        >
            <two-column-list-item-selection-sheet primary-key-field="id" primary-value-field="category"
                                                  primary-title-field="name" primary-footer-field="displayBalance"
                                                  primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="account"
                                                  primary-sub-items-field="accounts"
                                                  :primary-title-i18n="true"
                                                  secondary-key-field="id" secondary-value-field="id"
                                                  secondary-title-field="name" secondary-footer-field="displayBalance"
                                                  secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="account" secondary-color-field="color"
                                                  :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                  :items="allVisibleCategorizedAccounts"
                                                  v-model:show="showSourceAccountSheet"
                                                  v-model="transaction.sourceAccountId">
            </two-column-list-item-selection-sheet>
        </f7-list-item>

        <f7-list-item
            class="list-item-with-header-and-title"
            link="#" no-chevron
            :class="{ 'disabled': !allVisibleAccounts.length, 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt('Destination Account')"
            :title="destinationAccountName"
            v-if="transaction.type === TransactionType.Transfer"
            @click="showDestinationAccountSheet = true"
        >
            <two-column-list-item-selection-sheet primary-key-field="id" primary-value-field="category"
                                                  primary-title-field="name" primary-footer-field="displayBalance"
                                                  primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="account"
                                                  primary-sub-items-field="accounts"
                                                  :primary-title-i18n="true"
                                                  secondary-key-field="id" secondary-value-field="id"
                                                  secondary-title-field="name" secondary-footer-field="displayBalance"
                                                  secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="account" secondary-color-field="color"
                                                  :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                  :items="allVisibleCategorizedAccounts"
                                                  v-model:show="showDestinationAccountSheet"
                                                  v-model="transaction.destinationAccountId">
            </two-column-list-item-selection-sheet>
        </f7-list-item>

        <f7-list-item
            class="transaction-edit-datetime list-item-with-header-and-title"
            link="#" no-chevron
            :class="{ 'disabled': mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance, 'readonly': mode === TransactionEditPageMode.View && transaction.utcOffset === currentTimezoneOffsetMinutes }"
            v-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction"
        >
            <template #header>
                <div class="transaction-edit-datetime-header" @click="showDateTimeDialog('time')">{{ tt('Transaction Time') }}</div>
            </template>
            <template #title>
                <div class="transaction-edit-datetime-title">
                    <div @click="showDateTimeDialog('date')">{{ transactionDisplayDate }}</div>&nbsp;<div class="transaction-edit-datetime-time" @click="showDateTimeDialog('time')">{{ transactionDisplayTime }}</div>
                </div>
            </template>
            <date-time-selection-sheet :init-mode="transactionDateTimeSheetMode"
                                       :timezone-utc-offset="transaction.utcOffset"
                                       :model-value="transaction.time"
                                       v-model:show="showTransactionDateTimeSheet"
                                       @update:model-value="updateTransactionTime">
            </date-time-selection-sheet>
        </f7-list-item>

        <f7-list-item
            class="list-item-with-header-and-title"
            link="#" no-chevron
            :class="{ 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt('Scheduled Transaction Frequency')"
            :title="transactionDisplayScheduledFrequency"
            @click="showTransactionScheduledFrequencySheet = true"
            v-if="pageTypeAndMode?.type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type"
        >
            <schedule-frequency-sheet v-model:show="showTransactionScheduledFrequencySheet"
                                      v-model:type="transaction.scheduledFrequencyType"
                                      v-model="transaction.scheduledFrequency">
            </schedule-frequency-sheet>
        </f7-list-item>

        <f7-list-item
            class="transaction-edit-datetime list-item-with-header-and-title"
            link="#" no-chevron
            :class="{ 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt('Start Date')"
            :title="transactionDisplayScheduledStartDate"
            @click="showScheduledStartDateSheet = true"
            v-if="pageTypeAndMode?.type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type"
        >
            <date-selection-sheet v-model:show="showScheduledStartDateSheet"
                                  v-model="transaction.scheduledStartDate">
            </date-selection-sheet>
        </f7-list-item>

        <f7-list-item
            class="transaction-edit-datetime list-item-with-header-and-title"
            link="#" no-chevron
            :class="{ 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt('End Date')"
            :title="transactionDisplayScheduledEndDate"
            @click="showScheduledEndDateSheet = true"
            v-if="pageTypeAndMode?.type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type"
        >
            <date-selection-sheet v-model:show="showScheduledEndDateSheet"
                                  v-model="transaction.scheduledEndDate">
            </date-selection-sheet>
        </f7-list-item>

        <f7-list-item
            :no-chevron="mode === TransactionEditPageMode.View"
            link="#"
            class="list-item-with-header-and-title list-item-title-hide-overflow list-item-no-item-after"
            :class="{ 'disabled': mode === TransactionEditPageMode.Edit && transaction.type === TransactionType.ModifyBalance, 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt('Transaction Timezone')"
            v-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction || (pageTypeAndMode?.type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type)"
            @click="showTimezonePopup = true"
        >
            <template #title>
                <f7-block class="list-item-custom-title no-padding no-margin">
                    <span>{{ `(${transactionDisplayTimezone})` }}</span>
                    <span class="transaction-edit-timezone-name" v-if="transaction.timeZone || transaction.timeZone === ''">{{ transactionDisplayTimezoneName }}</span>
                    <span class="transaction-edit-timezone-name" v-else-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction && !transaction.timeZone && transaction.timeZone !== ''">{{ transactionTimezoneTimeDifference }}</span>
                </f7-block>
            </template>
            <list-item-selection-popup value-type="item"
                                       key-field="name" value-field="name"
                                       title-field="displayNameWithUtcOffset"
                                       :title="tt('Transaction Timezone')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Timezone')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allTimezones"
                                       :model-value="transaction.timeZone"
                                       v-model:show="showTimezonePopup"
                                       @update:model-value="updateTransactionTimezone">
            </list-item-selection-popup>
        </f7-list-item>

        <f7-list-item
            link="#" no-chevron
            class="list-item-with-header-and-title list-item-title-hide-overflow"
            :class="{ 'readonly': mode === TransactionEditPageMode.View && !transaction.geoLocation }"
            :header="tt('Geographic Location')"
            @click="showGeoLocationActionSheet = true"
            v-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction"
        >
            <template #title>
                <f7-block class="list-item-custom-title no-padding no-margin">
                    <span v-if="transaction.geoLocation">{{ `(${formatCoordinate(transaction.geoLocation)})` }}</span>
                    <span v-else-if="!transaction.geoLocation">{{ geoLocationStatusInfo }}</span>
                </f7-block>
            </template>

            <map-sheet :readonly="mode === TransactionEditPageMode.View"
                       v-model="transaction.geoLocation"
                       v-model:set-geo-location-by-click-map="setGeoLocationByClickMap"
                       v-model:show="showGeoLocationMapSheet">
            </map-sheet>
        </f7-list-item>

        <f7-list-item
            link="#" no-chevron
            :class="{ 'readonly': mode === TransactionEditPageMode.View }"
            :header="tt('Tags')"
            @click="showTransactionTagSheet = true"
        >
            <transaction-tag-selection-sheet :allow-add-new-tag="true" :enable-filter="true"
                                             v-model:show="showTransactionTagSheet"
                                             v-model="transaction.tagIds">
            </transaction-tag-selection-sheet>

            <template #footer>
                <f7-block class="margin-top-half no-padding no-margin" v-if="transaction.tagIds && transaction.tagIds.length">
                    <f7-chip media-text-color="var(--f7-chip-text-color)" class="transaction-edit-tag"
                             :text="allTagsMap[tagId]?.name ?? ''"
                             :key="tagId"
                             v-for="tagId in transaction.tagIds">
                        <template #media>
                            <f7-icon f7="number"></f7-icon>
                        </template>
                    </f7-chip>
                </f7-block>
                <f7-block class="margin-top-half no-padding no-margin" v-else-if="!transaction.tagIds || !transaction.tagIds.length">
                    <f7-chip class="transaction-edit-tag" :text="tt('None')">
                    </f7-chip>
                </f7-block>
            </template>
        </f7-list-item>

        <f7-list-item
            link="#" no-chevron
            :header="tt('Pictures')"
            v-if="showTransactionPictures || (transaction.pictures && transaction.pictures.length > 0)"
        >
            <template #footer>
                <f7-block class="margin-top-half no-padding no-margin" :class="{ 'readonly': submitting || uploadingPicture || removingPictureId }">
                    <swiper-container
                        :pagination="false"
                        :space-between="10"
                        :slides-per-view="'auto'"
                        class="transaction-pictures"
                    >
                        <swiper-slide class="transaction-picture-container" :key="picIdx"
                                      v-for="(pictureInfo, picIdx) in transaction.pictures"
                                      @click="viewOrRemovePicture(pictureInfo)">
                            <div class="transaction-picture">
                                <div class="display-flex justify-content-center align-items-center transaction-picture-control-backdrop"
                                     v-if="mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit">
                                    <f7-icon class="picture-control-icon picture-remove-icon" f7="trash" :aria-label="tt('Remove Picture')" v-if="pictureInfo.pictureId !== removingPictureId"></f7-icon>
                                    <f7-preloader color="white" :size="28" v-if="pictureInfo.pictureId === removingPictureId" />
                                </div>
                                <image-box style="height: 100%" alt="picture" :src="getTransactionPictureUrl(pictureInfo)">
                                    <template #error>
                                        {{ tt('Failed to load image, please check whether the config "domain" and "root_url" are set correctly.') }}
                                    </template>
                                </image-box>
                            </div>
                        </swiper-slide>
                        <swiper-slide @click="showOpenPictureDialog" v-if="canAddTransactionPicture">
                            <div class="display-flex justify-content-center align-items-center transaction-picture transaction-picture-add">
                                <f7-icon class="picture-control-icon" f7="plus" :aria-label="tt('Add Picture')" v-if="!uploadingPicture"></f7-icon>
                                <f7-preloader :size="28" v-if="uploadingPicture" />
                            </div>
                        </swiper-slide>
                    </swiper-container>
                </f7-block>
            </template>
        </f7-list-item>

        <f7-list-input
            type="textarea"
            class="transaction-edit-comment"
            style="height: auto"
            autocomplete="off"
            :class="{ 'readonly': mode === TransactionEditPageMode.View }"
            :label="transactionDescriptionTitle"
            :placeholder="mode !== TransactionEditPageMode.View ? tt('Your transaction description (optional)') : ''"
            v-textarea-auto-size
            v-model:value="transaction.comment"
        ></f7-list-input>
    </f7-list>
</template>
