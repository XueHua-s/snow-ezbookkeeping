<script setup lang="ts">
import TransactionBasicFields from './EditDialog/TransactionBasicFields.vue';
import MapView from '@/components/common/MapView.vue';
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { SUPPORTED_IMAGE_EXTENSIONS } from '@/consts/file.ts';

import { TransactionQuickAddButtonActionType, TransactionType } from '@/core/transaction.ts';
import {
    isSupportGetGeoLocationByClick
} from '@/lib/map/index.ts';
import {
    getMapProvider,
    isTransactionFromAITextRecognitionEnabled,
    isTransactionPicturesEnabled
} from '@/lib/server_settings.ts';

import {
    AfterSaveAction,
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import {
    mdiCheck,
    mdiDotsVertical,
    mdiEyeOffOutline,
    mdiEyeOutline,
    mdiFullscreen,
    mdiImagePlusOutline,
    mdiMagicStaff,
    mdiMapMarkerOutline,
    mdiMenuDown,
    mdiPencilOutline,
    mdiSwapHorizontal,
    mdiTrashCanOutline
} from '@mdi/js';
import { useEditDialog } from './EditDialog/useEditDialog.ts';
export type { TransactionEditOptions } from './EditDialog/useEditDialog.ts';

const props = defineProps<{
    type: TransactionEditPageType;
    persistent?: boolean;
    show?: boolean;
}>();

const {
    tt, mode, loading, recognizing, submitting, uploadingPicture, setGeoLocationByClickMap, transaction, allTimezones,
    allVisibleAccounts, allVisibleCategorizedAccounts, allCategories, hasVisibleExpenseCategories, hasVisibleIncomeCategories,
    hasVisibleTransferCategories, canAddTransactionPicture, title, saveButtonTitle, cancelButtonTitle, sourceAmountName,
    sourceAmountTitle, sourceAccountTitle, transferInAmountTitle, sourceAccountName, destinationAccountName,
    sourceAccountCurrency, destinationAccountCurrency, transactionDisplayTimezone, transactionTimezoneTimeDifference,
    geoLocationStatusInfo, transactionDescriptionTitle, inputEmptyProblemMessage, inputIsEmpty, updateTransactionTime,
    updateTransactionTimezone, swapTransactionData, getTransactionPictureUrl, showState, showPasteTextDialog, activeTab,
    originalTransactionEditable, geoMenuState, removingPictureId, pastedText, sourceAmountColor, isTransactionModified, open,
    save, recognizeText, recognizeFromClipboard, duplicate, edit, remove, cancel, updateGeoLocation,
    updateSpecifiedGeoLocation, clearGeoLocation, showOpenPictureDialog, viewOrRemovePicture, onSavingTag, onUploadPicture,
    onShowDateTimeError,
} = useEditDialog(props);

defineExpose({
    open
});

</script>

<template>
    <v-dialog width="1000" :persistent="isTransactionModified || recognizing" v-model="showState">
        <two-column-dialog-layout :disabled="loading || submitting || recognizing" :loading="loading"
                                  :title="tt(title)" :cancel-button-title="tt(cancelButtonTitle)"
                                  @cancel="cancel">
            <template #after-title>
                <v-btn density="compact" color="default" variant="text" class="ms-2" :icon="true"
                       :aria-label="tt('Edit')" :disabled="loading || submitting || recognizing"
                       v-if="mode === TransactionEditPageMode.View && originalTransactionEditable"
                       @click="edit">
                    <v-icon :icon="mdiPencilOutline" size="22"/>
                    <v-tooltip activator="parent">{{ tt('Edit') }}</v-tooltip>
                </v-btn>
                <v-btn density="compact" color="default" variant="text" class="ms-2" :icon="true"
                       :aria-label="tt('AI Clipboard Text Recognition')" :disabled="loading || submitting || recognizing"
                       v-if="mode !== TransactionEditPageMode.View && type === TransactionEditPageType.Transaction && activeTab === 'basicInfo' && isTransactionFromAITextRecognitionEnabled()"
                       @click="recognizeFromClipboard">
                    <v-icon :icon="mdiMagicStaff" size="20" v-if="!recognizing"/>
                    <v-tooltip activator="parent">{{ tt('AI Clipboard Text Recognition') }}</v-tooltip>
                    <v-progress-circular indeterminate size="22" v-if="recognizing"></v-progress-circular>
                </v-btn>
                <small class="ms-2 text-truncate" v-if="recognizing">{{ tt('AI can make mistakes. Check important info.') }}</small>
            </template>

            <template #toolbar>
                <v-btn density="compact" color="default" variant="text" class="ms-2" :icon="true"
                       :aria-label="tt('More')" :disabled="loading || submitting || recognizing"
                       v-if="mode !== TransactionEditPageMode.View && (activeTab === 'basicInfo' || (activeTab === 'map' && isSupportGetGeoLocationByClick()))">
                    <v-icon :icon="mdiDotsVertical" size="22" />
                    <v-menu activator="parent">
                        <v-list v-if="activeTab === 'basicInfo'">
                            <v-list-item :prepend-icon="mdiSwapHorizontal"
                                         :title="tt('Swap Account')"
                                         v-if="transaction.type === TransactionType.Transfer"
                                         @click="swapTransactionData(true, false)"></v-list-item>
                            <v-list-item :prepend-icon="mdiSwapHorizontal"
                                         :title="tt('Swap Amount')"
                                         v-if="transaction.type === TransactionType.Transfer"
                                         @click="swapTransactionData(false, true)"></v-list-item>
                            <v-list-item :prepend-icon="mdiSwapHorizontal"
                                         :title="tt('Swap Account and Amount')"
                                         v-if="transaction.type === TransactionType.Transfer"
                                         @click="swapTransactionData(true, true)"></v-list-item>
                            <v-divider v-if="transaction.type === TransactionType.Transfer" />
                            <v-list-item :prepend-icon="mdiEyeOutline"
                                         :title="tt('Show Amount')"
                                         v-if="transaction.hideAmount" @click="transaction.hideAmount = false"></v-list-item>
                            <v-list-item :prepend-icon="mdiEyeOffOutline"
                                         :title="tt('Hide Amount')"
                                         v-if="!transaction.hideAmount" @click="transaction.hideAmount = true"></v-list-item>
                        </v-list>
                        <v-list v-if="activeTab === 'map'">
                            <v-list-item key="setGeoLocationByClickMap" value="setGeoLocationByClickMap"
                                         :prepend-icon="mdiMapMarkerOutline"
                                         :disabled="!transaction.geoLocation" v-if="isSupportGetGeoLocationByClick()">
                                <v-list-item-title class="cursor-pointer" @click="setGeoLocationByClickMap = !setGeoLocationByClickMap; geoMenuState = false">
                                    <div class="d-flex align-center">
                                        <span>{{ tt('Click on Map to Set Geographic Location') }}</span>
                                        <v-spacer/>
                                        <v-icon :icon="mdiCheck" v-if="setGeoLocationByClickMap" />
                                    </div>
                                </v-list-item-title>
                            </v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
            </template>

            <template #content-left-column>
                <div class="px-4">
                    <v-tabs class="v-tabs-pill" direction="vertical" :class="{ 'readonly': type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit }"
                            :disabled="loading || submitting || recognizing" v-model="transaction.type">
                        <v-tab :value="TransactionType.Expense" :disabled="type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit && transaction.type !== TransactionType.Expense" v-if="transaction.type !== TransactionType.ModifyBalance">
                            <span>{{ tt('Expense') }}</span>
                        </v-tab>
                        <v-tab :value="TransactionType.Income" :disabled="type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit && transaction.type !== TransactionType.Income" v-if="transaction.type !== TransactionType.ModifyBalance">
                            <span>{{ tt('Income') }}</span>
                        </v-tab>
                        <v-tab :value="TransactionType.Transfer" :disabled="type === TransactionEditPageType.Transaction && mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit && transaction.type !== TransactionType.Transfer" v-if="transaction.type !== TransactionType.ModifyBalance">
                            <span>{{ tt('Transfer') }}</span>
                        </v-tab>
                        <v-tab :value="TransactionType.ModifyBalance" v-if="type === TransactionEditPageType.Transaction && transaction.type === TransactionType.ModifyBalance">
                            <span>{{ tt('Modify Balance') }}</span>
                        </v-tab>
                    </v-tabs>
                </div>
                <v-divider class="my-2"/>
                <div class="px-4">
                    <v-tabs direction="vertical" :disabled="loading || submitting || recognizing" v-model="activeTab">
                        <v-tab value="basicInfo">
                            <span>{{ tt('Basic Information') }}</span>
                        </v-tab>
                        <v-tab value="map" :disabled="!transaction.geoLocation" v-if="type === TransactionEditPageType.Transaction && !!getMapProvider()">
                            <span>{{ tt('Location on Map') }}</span>
                        </v-tab>
                        <v-tab value="pictures" :disabled="mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit && (!transaction.pictures || !transaction.pictures.length)" v-if="type === TransactionEditPageType.Transaction && isTransactionPicturesEnabled()">
                            <span>{{ tt('Pictures') }}</span>
                        </v-tab>
                    </v-tabs>
                </div>
            </template>

            <template #content-right-column>
                <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container"
                          v-model="activeTab">
                    <TransactionBasicFields
    :type="type"
    :loading="loading"
    :submitting="submitting"
    :recognizing="recognizing"
    :sourceAmountColor="sourceAmountColor"
    :sourceAccountCurrency="sourceAccountCurrency"
    :mode="mode"
    :sourceAmountTitle="sourceAmountTitle"
    :sourceAmountName="sourceAmountName"
    :destinationAccountCurrency="destinationAccountCurrency"
    :transferInAmountTitle="transferInAmountTitle"
    :hasVisibleExpenseCategories="hasVisibleExpenseCategories"
    :allCategories="allCategories"
    :hasVisibleIncomeCategories="hasVisibleIncomeCategories"
    :hasVisibleTransferCategories="hasVisibleTransferCategories"
    :allVisibleAccounts="allVisibleAccounts"
    :sourceAccountName="sourceAccountName"
    :sourceAccountTitle="sourceAccountTitle"
    :allVisibleCategorizedAccounts="allVisibleCategorizedAccounts"
    :destinationAccountName="destinationAccountName"
    :transactionDisplayTimezone="transactionDisplayTimezone"
    :transactionTimezoneTimeDifference="transactionTimezoneTimeDifference"
    :allTimezones="allTimezones"
    :geoLocationStatusInfo="geoLocationStatusInfo"
    :transactionDescriptionTitle="transactionDescriptionTitle"
    v-model:transaction="transaction"
    v-model:geoMenuState="geoMenuState"
    @updateTransactionTime="updateTransactionTime"
    @onShowDateTimeError="onShowDateTimeError"
    @updateTransactionTimezone="updateTransactionTimezone"
    @updateGeoLocation="updateGeoLocation"
    @clearGeoLocation="clearGeoLocation"
    @onSavingTag="onSavingTag"
/>
                    <v-window-item value="map">
                        <map-view ref="map" map-class="transaction-edit-map-view mb-3 mb-sm-0"
                                  :enable-zoom-control="true" :geo-location="transaction.geoLocation"
                                  @click="updateSpecifiedGeoLocation">
                            <template #error-title="{ mapSupported, mapDependencyLoaded }">
                                <span class="text-body-large" v-if="!mapSupported"><b>{{ tt('Unsupported Map Provider') }}</b></span>
                                <span class="text-body-large" v-else-if="!mapDependencyLoaded"><b>{{ tt('Cannot Initialize Map') }}</b></span>
                            </template>
                            <template #error-content>
                                <span class="text-body-large">
                                    {{ tt('Please refresh the page and try again. If the error persists, ensure that the server\'s map settings are correctly configured.') }}
                                </span>
                            </template>
                        </map-view>
                    </v-window-item>
                    <v-window-item value="pictures">
                        <v-row class="transaction-pictures align-content-start ma-0 pt-3" :class="{ 'readonly': submitting || uploadingPicture || removingPictureId }">
                            <v-col :key="picIdx" cols="6" md="3" v-for="(pictureInfo, picIdx) in transaction.pictures">
                                <v-avatar rounded="lg" variant="tonal" size="160"
                                          class="cursor-pointer transaction-picture"
                                          color="rgba(0,0,0,0)" @click="viewOrRemovePicture(pictureInfo)">
                                    <v-img :src="getTransactionPictureUrl(pictureInfo)">
                                        <template #placeholder>
                                            <div class="d-flex align-center justify-center bg-light-primary">
                                                <v-progress-circular color="grey-500" indeterminate size="48"></v-progress-circular>
                                            </div>
                                        </template>
                                        <template #error>
                                            <div class="d-flex align-center justify-center bg-light-primary">
                                                <span class="text-body-large">{{ tt('Failed to load image, please check whether the config "domain" and "root_url" are set correctly.') }}</span>
                                            </div>
                                        </template>
                                    </v-img>
                                    <div class="picture-control-icon" :class="{ 'show-control-icon': pictureInfo.pictureId === removingPictureId }">
                                        <v-icon size="64" :icon="mdiTrashCanOutline" :aria-label="tt('Remove Picture')" v-if="(mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit) && pictureInfo.pictureId !== removingPictureId"/>
                                        <v-progress-circular color="grey-500" indeterminate size="48" v-if="(mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit) && pictureInfo.pictureId === removingPictureId"></v-progress-circular>
                                        <v-icon size="64" :icon="mdiFullscreen" :aria-label="tt('View')" v-if="mode !== TransactionEditPageMode.Add && mode !== TransactionEditPageMode.Edit"/>
                                    </div>
                                </v-avatar>
                            </v-col>
                            <v-col cols="6" md="3" v-if="canAddTransactionPicture">
                                <v-avatar rounded="lg" variant="tonal" size="160"
                                          class="transaction-picture transaction-picture-add"
                                          :class="{ 'enabled': !submitting, 'cursor-pointer': !submitting }"
                                          color="rgba(0,0,0,0)" @click="showOpenPictureDialog">
                                    <v-tooltip activator="parent" v-if="!submitting">{{ tt('Add Picture') }}</v-tooltip>
                                    <v-icon class="transaction-picture-add-icon" size="56" :aria-label="tt('Add Picture')" :icon="mdiImagePlusOutline" v-if="!uploadingPicture"/>
                                    <v-progress-circular color="grey-500" indeterminate size="48" v-if="uploadingPicture"></v-progress-circular>
                                </v-avatar>
                            </v-col>
                        </v-row>
                    </v-window-item>
                </v-window>
            </template>

            <template #footer>
                <v-btn color="error" variant="tonal" :disabled="loading || submitting || recognizing"
                       v-if="mode === TransactionEditPageMode.View && originalTransactionEditable" @click="remove">
                    {{ tt('Delete') }}
                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                </v-btn>
                <v-spacer/>
                <v-tooltip :disabled="!inputIsEmpty" :text="inputEmptyProblemMessage ? tt(inputEmptyProblemMessage) : ''">
                    <template v-slot:activator="{ props }">
                        <div v-bind="props" class="d-inline-block">
                            <v-btn-group density="comfortable" v-if="mode === TransactionEditPageMode.Add || mode === TransactionEditPageMode.Edit">
                                <v-btn color="primary" :disabled="inputIsEmpty || loading || submitting || recognizing" @click="save(AfterSaveAction.GoBack)">
                                    {{ tt(saveButtonTitle) }}
                                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                                </v-btn>
                                <v-btn color="primary" density="compact"
                                       :aria-label="tt('More')" :disabled="inputIsEmpty || loading || submitting || recognizing" :icon="true"
                                       v-if="type === TransactionEditPageType.Transaction && mode === TransactionEditPageMode.Add">
                                    <v-icon :icon="mdiMenuDown" size="24" />
                                    <v-menu activator="parent">
                                        <v-list>
                                            <v-list-item :title="tt(TransactionQuickAddButtonActionType.SaveAndAddNewTransaction.name)"
                                                         @click="save(AfterSaveAction.StayWithNewTransaction)"></v-list-item>
                                            <v-list-item :title="tt(TransactionQuickAddButtonActionType.SaveAndKeepCurrentData.name)"
                                                         @click="save(AfterSaveAction.StayWithCurrentTransaction)"></v-list-item>
                                        </v-list>
                                    </v-menu>
                                </v-btn>
                            </v-btn-group>
                        </div>
                    </template>
                </v-tooltip>
                <v-btn-group variant="tonal" density="comfortable"
                             v-if="mode === TransactionEditPageMode.View && transaction.type !== TransactionType.ModifyBalance">
                    <v-btn :disabled="loading || submitting || recognizing"
                           @click="duplicate(false, false)">{{ tt('Duplicate') }}</v-btn>
                    <v-btn density="compact" :aria-label="tt('More')" :disabled="loading || submitting || recognizing" :icon="true">
                        <v-icon :icon="mdiMenuDown" size="24" />
                        <v-menu activator="parent">
                            <v-list>
                                <v-list-item :title="tt('Duplicate (With Time)')"
                                             @click="duplicate(true, false)"></v-list-item>
                                <v-list-item :title="tt('Duplicate (With Geographic Location)')"
                                             @click="duplicate(false, true)"
                                             v-if="transaction.geoLocation"></v-list-item>
                                <v-list-item :title="tt('Duplicate (With Time and Geographic Location)')"
                                             @click="duplicate(true, true)"
                                             v-if="transaction.geoLocation"></v-list-item>
                            </v-list>
                        </v-menu>
                    </v-btn>
                </v-btn-group>
            </template>
        </two-column-dialog-layout>
    </v-dialog>

    <v-dialog width="600" v-model="showPasteTextDialog">
        <one-column-dialog-layout content-class="pa-0" :disabled="recognizing"
                                  :title="tt('AI Clipboard Text Recognition')" :cancel-button-title="tt('Cancel')"
                                  @cancel="showPasteTextDialog = false; pastedText = '';">
            <template #toolbar>
                <v-btn class="me-2" density="comfortable" variant="outlined"
                       :disabled="!pastedText || !pastedText.trim() || recognizing"
                       @click="showPasteTextDialog = false; recognizeText(pastedText);">{{ tt('Recognize') }}</v-btn>
            </template>

            <template #content>
                <v-textarea no-resize persistent-placeholder
                            class="w-100 h-100 ps-4 always-cursor-text"
                            rows="10" autocomplete="off" density="compact" variant="plain" :rounded="false"
                            :disabled="recognizing"
                            :placeholder="tt('Click here to paste a transaction description')"
                            v-model="pastedText"></v-textarea>
            </template>
        </one-column-dialog-layout>
    </v-dialog>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
    <input ref="pictureInput" type="file" style="display: none" :accept="SUPPORTED_IMAGE_EXTENSIONS" @change="onUploadPicture($event)" />
</template>



<style>
.transaction-edit-amount .v-field__prepend-inner,
.transaction-edit-amount .v-field__append-inner,
.transaction-edit-amount .v-field__field > input {
    font-size: 1.25rem;
}

.transaction-edit-timezone.v-input input::placeholder {
    color: rgba(var(--v-theme-on-background), var(--v-high-emphasis-opacity)) !important;
    opacity: unset;
}

.transaction-edit-map-view {
    height: 289px;
}

@media (min-height: 620px) {
    .transaction-edit-map-view {
        height: 415px;
    }

    @media (min-width: 960px) {
        .transaction-pictures {
            min-height: 416px;
        }
    }
}

@media (min-height: 700px) {
    .transaction-edit-map-view {
        height: 513px;
    }

    @media (min-width: 960px) {
        .transaction-pictures {
            min-height: 514px;
        }
    }
}

.transaction-picture .picture-control-icon {
    display: none;
    position: absolute;
    width: 100% !important;
    height: 100% !important;
    background-color: rgba(0, 0, 0, 0.4);
}

.transaction-picture .picture-control-icon > i.v-icon {
    background-color: transparent;
    color: rgba(255, 255, 255, 0.8);
}

.transaction-picture:hover .picture-control-icon,
.transaction-picture .picture-control-icon.show-control-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    vertical-align: middle;
}

.transaction-picture:hover .transaction-picture-placeholder {
    display: none;
}

.transaction-picture-add {
    border: 2px dashed rgba(var(--v-theme-grey-500));

    .transaction-picture-add-icon {
        color: rgba(var(--v-theme-grey-500));
    }
}

.transaction-picture-add.enabled:hover {
    border: 2px dashed rgba(var(--v-theme-grey-700));

    .transaction-picture-add-icon {
        color: rgba(var(--v-theme-grey-700));
    }
}
</style>
