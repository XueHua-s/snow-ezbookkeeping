<script setup lang="ts">
import TransactionForm from './EditPage/TransactionForm.vue';
import { SUPPORTED_IMAGE_EXTENSIONS } from '@/consts/file.ts';

import { TemplateType } from '@/core/template.ts';
import {
    TransactionQuickAddButtonActionType,
    TransactionQuickSaveButtonStyle,
    TransactionType
} from '@/core/transaction.ts';

import {
    getMapProvider,
    isTransactionFromAITextRecognitionEnabled,
    isTransactionPicturesEnabled
} from '@/lib/server_settings.ts';
import { isiOS } from '@/lib/ui/mobile.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import {
    AfterSaveAction,
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { Router } from 'framework7/types';
import { useEditPage } from './EditPage/useEditPage.ts';

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
    autoUploadPicture?: File;
    autoRecognizeClipboardText?: string;
}>();

const {
    pageTypeAndMode, tt, mode, loading, recognizing, submitting, uploadingPicture, setGeoLocationByClickMap, transaction,
    currentTimezoneOffsetMinutes, allTimezones, allVisibleAccounts, allVisibleCategorizedAccounts, allCategories, allTagsMap,
    hasVisibleExpenseCategories, hasVisibleIncomeCategories, hasVisibleTransferCategories, canAddTransactionPicture, title,
    quickSaveButtonTitle, sourceAmountTitle, sourceAccountTitle, transferInAmountTitle, sourceAccountName,
    destinationAccountName, sourceAccountCurrency, destinationAccountCurrency, transactionDisplayTimezone,
    transactionTimezoneTimeDifference, geoLocationStatusInfo, transactionDescriptionTitle, inputIsEmpty,
    updateTransactionTime, updateTransactionTimezone, swapTransactionData, getDisplayAmount, getTransactionPictureUrl,
    isSupportClipboard, removingPictureId, pastedText, transactionDateTimeSheetMode, showQuickSavePopover, showTimezonePopup,
    showGeoLocationActionSheet, showMoreActionSheet, showSourceAmountSheet, showDestinationAmountSheet, showCategorySheet,
    showSourceAccountSheet, showDestinationAccountSheet, showTransactionDateTimeSheet, showTransactionScheduledFrequencySheet,
    showScheduledStartDateSheet, showScheduledEndDateSheet, showGeoLocationMapSheet, showTransactionTagSheet,
    showTransactionPictures, showAITextRecognitionSheet, quickSaveButtonStyleType, quickSaveButtonFloatingPosition,
    sourceAmountClass, destinationAmountClass, transactionDisplayDate, transactionDisplayTime, transactionDisplayTimezoneName,
    transactionPictures, transactionThumbs, transactionDisplayScheduledFrequency, transactionDisplayScheduledStartDate,
    transactionDisplayScheduledEndDate, save, quickSave, recognizeText, recognizeFromClipboard, pasteAmount,
    updateGeoLocation, clearGeoLocation, showDateTimeDialog, showOpenPictureDialog, viewOrRemovePicture, duplicate,
    onUploadPicture, onPageAfterIn, onPageBeforeOut,
} = useEditPage(props);

</script>

<template>
    <f7-page @page:afterin="onPageAfterIn" @page:beforeout="onPageBeforeOut">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt(title)"></f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }" v-if="mode !== TransactionEditPageMode.View || transaction.type !== TransactionType.ModifyBalance">
                <f7-link icon-f7="ellipsis" :aria-label="tt('More')" @click="showMoreActionSheet = true"></f7-link>
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': inputIsEmpty || submitting || recognizing }" :aria-label="tt('Save')" @click="save(AfterSaveAction.GoBack)" v-if="mode !== TransactionEditPageMode.View"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-block :class="{ 'subnav-segmented-bar': true, 'disabled': loading }">
            <f7-segmented strong round :class="{ 'readonly': pageTypeAndMode?.type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit }">
                <f7-button round :text="tt('Expense')" :active="transaction.type === TransactionType.Expense"
                           :disabled="pageTypeAndMode?.type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit && transaction.type !== TransactionType.Expense"
                           v-if="transaction.type !== TransactionType.ModifyBalance"
                           @click="transaction.type = TransactionType.Expense"></f7-button>
                <f7-button round :text="tt('Income')" :active="transaction.type === TransactionType.Income"
                           :disabled="pageTypeAndMode?.type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit && transaction.type !== TransactionType.Income"
                           v-if="transaction.type !== TransactionType.ModifyBalance"
                           @click="transaction.type = TransactionType.Income"></f7-button>
                <f7-button round :text="tt('Transfer')" :active="transaction.type === TransactionType.Transfer"
                           :disabled="pageTypeAndMode?.type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit && transaction.type !== TransactionType.Transfer"
                           v-if="transaction.type !== TransactionType.ModifyBalance"
                           @click="transaction.type = TransactionType.Transfer"></f7-button>
                <f7-button round :text="tt('Modify Balance')" :active="transaction.type === TransactionType.ModifyBalance"
                           v-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction && transaction.type === TransactionType.ModifyBalance"></f7-button>
            </f7-segmented>
        </f7-block>

        <f7-list strong inset dividers class="margin-vertical-half skeleton-text" v-if="loading">
            <f7-list-input label="Template Name" placeholder="Template Name" v-if="pageTypeAndMode?.type === TransactionEditPageType.Template"></f7-list-input>
            <f7-list-item
                class="transaction-edit-amount ebk-large-amount"
                header="Expense Amount" title="0.00">
            </f7-list-item>
            <f7-list-item
                class="transaction-edit-amount ebk-large-amount"
                header="Transfer In Amount" title="0.00" v-if="transaction.type === TransactionType.Transfer">
            </f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-title-hide-overflow" header="Category" title="Category Names" v-if="transaction.type !== TransactionType.ModifyBalance"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title" header="Account" title="Account Name"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title" header="Destination Account" title="Account Name" v-if="transaction.type === TransactionType.Transfer"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title" header="Transaction Time" title="YYYY/MM/DD HH:mm:ss" v-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title" header="Scheduled Transaction Frequency" title="Every XXXXX" v-if="pageTypeAndMode?.type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-title-hide-overflow list-item-no-item-after" header="Transaction Timezone" title="(UTC XX:XX) System Default" link="#" :no-chevron="mode === TransactionEditPageMode.View" v-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction || (pageTypeAndMode?.type === TransactionEditPageType.Template && transaction instanceof TransactionTemplate && transaction.templateType === TemplateType.Schedule.type)"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-title-hide-overflow" header="Geographic Location" title="No Location" v-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction"></f7-list-item>
            <f7-list-item header="Tags">
                <template #footer>
                    <f7-block class="margin-top-half no-padding no-margin">
                        <f7-chip class="transaction-edit-tag" text="None"></f7-chip>
                    </f7-block>
                </template>
            </f7-list-item>
            <f7-list-input type="textarea" label="Description" placeholder="Your transaction description (optional)"></f7-list-input>
        </f7-list>

        <TransactionForm
    v-else-if="!loading"
    :loading="loading"
    :pageTypeAndMode="pageTypeAndMode"
    :sourceAmountClass="sourceAmountClass"
    :sourceAmountTitle="sourceAmountTitle"
    :getDisplayAmount="getDisplayAmount"
    :sourceAccountCurrency="sourceAccountCurrency"
    :destinationAmountClass="destinationAmountClass"
    :transferInAmountTitle="transferInAmountTitle"
    :destinationAccountCurrency="destinationAccountCurrency"
    :hasVisibleExpenseCategories="hasVisibleExpenseCategories"
    :mode="mode"
    :allCategories="allCategories"
    :hasVisibleIncomeCategories="hasVisibleIncomeCategories"
    :hasVisibleTransferCategories="hasVisibleTransferCategories"
    :allVisibleAccounts="allVisibleAccounts"
    :sourceAccountTitle="sourceAccountTitle"
    :sourceAccountName="sourceAccountName"
    :allVisibleCategorizedAccounts="allVisibleCategorizedAccounts"
    :destinationAccountName="destinationAccountName"
    :currentTimezoneOffsetMinutes="currentTimezoneOffsetMinutes"
    :transactionDisplayDate="transactionDisplayDate"
    :transactionDisplayTime="transactionDisplayTime"
    :transactionDateTimeSheetMode="transactionDateTimeSheetMode"
    :transactionDisplayScheduledFrequency="transactionDisplayScheduledFrequency"
    :transactionDisplayScheduledStartDate="transactionDisplayScheduledStartDate"
    :transactionDisplayScheduledEndDate="transactionDisplayScheduledEndDate"
    :transactionDisplayTimezone="transactionDisplayTimezone"
    :transactionDisplayTimezoneName="transactionDisplayTimezoneName"
    :transactionTimezoneTimeDifference="transactionTimezoneTimeDifference"
    :allTimezones="allTimezones"
    :geoLocationStatusInfo="geoLocationStatusInfo"
    :allTagsMap="allTagsMap"
    :showTransactionPictures="showTransactionPictures"
    :submitting="submitting"
    :uploadingPicture="uploadingPicture"
    :removingPictureId="removingPictureId"
    :getTransactionPictureUrl="getTransactionPictureUrl"
    :canAddTransactionPicture="canAddTransactionPicture"
    :transactionDescriptionTitle="transactionDescriptionTitle"
    v-model:transaction="transaction"
    v-model:showSourceAmountSheet="showSourceAmountSheet"
    v-model:showDestinationAmountSheet="showDestinationAmountSheet"
    v-model:showCategorySheet="showCategorySheet"
    v-model:showSourceAccountSheet="showSourceAccountSheet"
    v-model:showDestinationAccountSheet="showDestinationAccountSheet"
    v-model:showTransactionDateTimeSheet="showTransactionDateTimeSheet"
    v-model:showTransactionScheduledFrequencySheet="showTransactionScheduledFrequencySheet"
    v-model:showScheduledStartDateSheet="showScheduledStartDateSheet"
    v-model:showScheduledEndDateSheet="showScheduledEndDateSheet"
    v-model:showTimezonePopup="showTimezonePopup"
    v-model:setGeoLocationByClickMap="setGeoLocationByClickMap"
    v-model:showGeoLocationMapSheet="showGeoLocationMapSheet"
    v-model:showTransactionTagSheet="showTransactionTagSheet"
    v-model:showGeoLocationActionSheet="showGeoLocationActionSheet"
    @showDateTimeDialog="showDateTimeDialog"
    @updateTransactionTime="updateTransactionTime"
    @updateTransactionTimezone="updateTransactionTimezone"
    @viewOrRemovePicture="viewOrRemovePicture"
    @showOpenPictureDialog="showOpenPictureDialog"
/>

        <f7-actions close-by-outside-click close-on-escape :opened="showGeoLocationActionSheet" @actions:closed="showGeoLocationActionSheet = false">
            <f7-actions-group>
                <f7-actions-button v-if="mode !== TransactionEditPageMode.View" @click="updateGeoLocation(true)">{{ tt('Update Geographic Location') }}</f7-actions-button>
                <f7-actions-button v-if="mode !== TransactionEditPageMode.View" @click="clearGeoLocation">{{ tt('Clear Geographic Location') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="!!getMapProvider()">
                <f7-actions-button :class="{ 'disabled': !transaction.geoLocation }" @click="setGeoLocationByClickMap = false; showGeoLocationMapSheet = true">{{ tt('Show on the map') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-actions close-by-outside-click close-on-escape :opened="showMoreActionSheet" @actions:closed="showMoreActionSheet = false">
            <f7-actions-group v-if="mode !== TransactionEditPageMode.View && pageTypeAndMode?.type === TransactionEditPageType.Transaction && isTransactionFromAITextRecognitionEnabled()">
                <f7-actions-button @click="recognizeFromClipboard">{{ tt('AI Clipboard Text Recognition') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="mode !== TransactionEditPageMode.View && transaction.type === TransactionType.Transfer">
                <f7-actions-button @click="swapTransactionData(true, false)">{{ tt('Swap Account') }}</f7-actions-button>
                <f7-actions-button @click="swapTransactionData(false, true)">{{ tt('Swap Amount') }}</f7-actions-button>
                <f7-actions-button @click="swapTransactionData(true, true)">{{ tt('Swap Account and Amount') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="mode !== TransactionEditPageMode.View">
                <f7-actions-button v-if="isSupportClipboard && !isiOS()" @click="pasteAmount('sourceAmount')">{{ tt('Paste Amount') }}</f7-actions-button>
                <f7-actions-button v-if="isSupportClipboard && !isiOS() && transaction.type === TransactionType.Transfer" @click="pasteAmount('destinationAmount')">{{ tt('Paste Destination Amount') }}</f7-actions-button>
                <f7-actions-button v-if="transaction.hideAmount" @click="transaction.hideAmount = false">{{ tt('Show Amount') }}</f7-actions-button>
                <f7-actions-button v-if="!transaction.hideAmount" @click="transaction.hideAmount = true">{{ tt('Hide Amount') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction && (mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit) && isTransactionPicturesEnabled() && !showTransactionPictures">
                <f7-actions-button @click="showTransactionPictures = true">{{ tt('Add Picture') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="pageTypeAndMode?.type === TransactionEditPageType.Transaction && mode === TransactionEditPageMode.View && transaction.type !== TransactionType.ModifyBalance">
                <f7-actions-button @click="duplicate(false, false)">{{ tt('Duplicate') }}</f7-actions-button>
                <f7-actions-button @click="duplicate(true, false)">{{ tt('Duplicate (With Time)') }}</f7-actions-button>
                <f7-actions-button @click="duplicate(false, true)" v-if="transaction.geoLocation">{{ tt('Duplicate (With Geographic Location)') }}</f7-actions-button>
                <f7-actions-button @click="duplicate(true, true)" v-if="transaction.geoLocation">{{ tt('Duplicate (With Time and Geographic Location)') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <template #fixed v-if="quickSaveButtonStyleType === TransactionQuickSaveButtonStyle.BottomLeftFloating.type || quickSaveButtonStyleType === TransactionQuickSaveButtonStyle.BottomCenterFloating.type || quickSaveButtonStyleType === TransactionQuickSaveButtonStyle.BottomRightFloating.type">
            <f7-fab id="quick-save-button" :class="{ 'disabled': inputIsEmpty || submitting || recognizing }" :position="quickSaveButtonFloatingPosition"
                    :text="tt(quickSaveButtonTitle)"
                    @click="quickSave" v-if="mode !== TransactionEditPageMode.View">
            </f7-fab>
        </template>

        <f7-toolbar id="quick-save-button" class="compact-tabbar" tabbar bottom v-if="quickSaveButtonStyleType === TransactionQuickSaveButtonStyle.BottomFixed.type && mode !== TransactionEditPageMode.View">
            <f7-link :class="{ 'disabled': inputIsEmpty || submitting || recognizing }" @click="quickSave">
                <span class="tabbar-primary-link">{{ tt(quickSaveButtonTitle) }}</span>
            </f7-link>
        </f7-toolbar>

        <f7-popover class="quick-save-popover" target-el="#quick-save-button"
                    v-model:opened="showQuickSavePopover">
            <f7-list>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt(TransactionQuickAddButtonActionType.SaveAndGoBack.name)"
                              @click="save(AfterSaveAction.GoBack)"></f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt(TransactionQuickAddButtonActionType.SaveAndAddNewTransaction.name)"
                              @click="save(AfterSaveAction.StayWithNewTransaction)"></f7-list-item>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tt(TransactionQuickAddButtonActionType.SaveAndKeepCurrentData.name)"
                              @click="save(AfterSaveAction.StayWithCurrentTransaction)"></f7-list-item>
            </f7-list>
        </f7-popover>

        <a-i-text-recognition-sheet :initial-text="pastedText" v-model:show="showAITextRecognitionSheet" @text:confirm="recognizeText" />
        <f7-photo-browser ref="pictureBrowser" type="popup" navbar-of-text="/"
                          :navbar-show-count="true" :exposition="false"
                          :photos="transactionPictures" :thumbs="transactionThumbs" />
        <input ref="pictureInput" type="file" style="display: none" :accept="`${SUPPORTED_IMAGE_EXTENSIONS};capture=camera`" @change="onUploadPicture($event)" />
    </f7-page>
</template>



<style>
.category-separate-icon.icon {
    margin-inline-start: 5px;
    margin-inline-end: 5px;
    font-size: var(--ebk-category-separate-icon-font-size);
    line-height: 16px;
    color: var(--f7-color-gray-tint);
}

.transaction-edit-amount {
    line-height: 53px;
}

.transaction-edit-amount .item-title {
    font-weight: bolder;
}

.transaction-edit-amount .item-header {
    padding-top: calc(var(--f7-typography-padding) / 2);
}

.transaction-edit-datetime .item-title {
    width: 100%;
}

.transaction-edit-datetime .item-title > .item-header > .transaction-edit-datetime-header {
    display: block;
    width: 100%;
}

.transaction-edit-datetime .item-title > .transaction-edit-datetime-title {
    display: flex;
    width: 100%;
}

.transaction-edit-datetime .item-title > .transaction-edit-datetime-title > .transaction-edit-datetime-time {
    flex-grow: 1;
    overflow: hidden;
    text-overflow: ellipsis;
}

.transaction-edit-timezone-name {
    padding-inline-start: 4px;
}

.transaction-edit-tag {
    --f7-chip-bg-color: var(--ebk-transaction-tag-chip-bg-color);
    margin-inline-end: 4px;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
}

.chip.transaction-edit-tag .chip-media+.chip-label {
    margin-inline-start: 0;
}

.chip.transaction-edit-tag .chip-media i.icon {
    font-size: calc(var(--f7-chip-media-size) - 12px);
    height: calc(var(--f7-chip-media-size) - 12px);
}

.transaction-pictures {
    height: var(--ebk-transaction-picture-size);
}

.transaction-pictures .transaction-picture-container {
    width: var(--ebk-transaction-picture-size);
    height: var(--ebk-transaction-picture-size);
}

.transaction-pictures .transaction-picture-container .transaction-picture {
    width: 100%;
    height: 100%;
}

.transaction-pictures .transaction-picture-container .transaction-picture .transaction-picture-control-backdrop {
    width: 100%;
    height: 100%;
    position: absolute;
    z-index: 10;
    background-color: rgba(0, 0, 0, 0.4);
    border-radius: 8px;
}

.transaction-pictures .transaction-picture-container .transaction-picture .picture-control-icon {
    z-index: 15;
    font-size: var(--ebk-transaction-picture-add-icon-size);
}

.transaction-pictures .transaction-picture-container .transaction-picture .picture-remove-icon {
    background-color: transparent;
    color: rgba(255, 255, 255, 0.8);
    font-size: var(--ebk-transaction-picture-remove-icon-size);
}

.transaction-pictures .transaction-picture-container .transaction-picture img {
    object-fit: cover;
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    border-radius: 8px;
}

.transaction-pictures .transaction-picture-add {
    width: calc(var(--ebk-transaction-picture-size) - 2px);
    height: calc(var(--ebk-transaction-picture-size) - 4px);
    border: 2px dashed #ccc;
    border-radius: 8px;
}
</style>
