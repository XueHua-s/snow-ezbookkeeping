<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { isDefined, objectFieldWithValueToArrayItem } from '@/lib/common.ts';
import {
    mdiCheck,
    mdiDotsVertical,
    mdiFilterOutline,
    mdiHelpCircleOutline,
    mdiImagePlusOutline,
    mdiRefresh,
    mdiTrashCanOutline
} from '@mdi/js';
import { ImportCustomFileFormatProcessMethod, useImportDialog } from './ImportDialog/useImportDialog.ts';
import ImportTransactionCheckDataTab from './tabs/ImportTransactionCheckDataTab.vue';
import ImportTransactionDefineColumnTab from './tabs/ImportTransactionDefineColumnTab.vue';
import ImportTransactionExecuteCustomScriptTab from './tabs/ImportTransactionExecuteCustomScriptTab.vue';
import ImportTransactionRecognizeImagesTab from './tabs/ImportTransactionRecognizeImagesTab.vue';

const {
    tt, formatNumberToLocalizedNumerals, importTransactionDefineColumnTab: importTransactionDefineColumnTabRef, importTransactionExecuteCustomScriptTab: importTransactionExecuteCustomScriptTabRef,
    importTransactionCheckDataTab: importTransactionCheckDataTabRef, allSupportedAdditionalOptions, showState, additionalOptionsMenuState, currentStep,
    importProcess, fileType, fileSubType, fileEncoding, autoDetectedFileEncoding, processCustomFileFormatMethod, importFile,
    importImageFiles, importData, importAdditionalOptions, importAIAdditionalPrompt, parsedFileData,
    parsedFileColumnSeparator, importTransactions, importedCount, loading, submitting, allSupportedImportFileCategoryAndTypes,
    allFileSubTypes, allSupportedEncodings, isCustomFileFormat, isImportDataFromTextbox, needAITextRecognition,
    isAIImageImport, needAIImageRecognition, supportedAdditionalOptions, supportedAIAdditionalPrompt, allSteps,
    supportedImportFileExtensions, displaySelectedAdditionalOptions, exportFileGuideDocumentUrl,
    exportFileGuideDocumentLanguageName, fileName, open, showOpenFileDialog, setImportFile, removeImportImageFile,
    reloadBasisData, cancelBatchRecognizeImages, parseData, batchApplyRules, submit, close,
} = useImportDialog();

defineExpose({
    open
});

defineProps<{
        persistent?: boolean;
    }>();
</script>

<template>
    <v-dialog :persistent="!!persistent || loading || submitting" v-model="showState">
        <one-column-dialog-layout :class="{ 'import-transaction-dialog-layout': currentStep === 'defineColumn' || currentStep === 'checkData' }"
                                  :content-class="currentStep === 'defineColumn' || currentStep === 'checkData' ? 'pa-0 import-transaction-dialog-content' : 'pa-0'"
                                  :title="tt('Import Transactions')" :cancel-button-title="tt('Cancel')"
                                  :disabled="loading || submitting"
                                  :loading="currentStep !== 'checkData' && loading"
                                  @cancel="close(currentStep === 'finalResult')">
            <template #after-title>
                <div class="ms-2 text-body-large text-truncate" v-if="currentStep === 'uploadFile' && exportFileGuideDocumentUrl">
                    <a :class="{ 'disabled': submitting }" :href="exportFileGuideDocumentUrl" target="_blank">
                        <v-icon :icon="mdiHelpCircleOutline" size="16" />
                        <span class="vertical-align-middle ms-1" v-if="isCustomFileFormat">{{ tt('How to import this file?') }}</span>
                        <span class="vertical-align-middle ms-1" v-if="!isCustomFileFormat">{{ tt('How to export this file?') }}</span>
                        <span class="vertical-align-middle ms-1" v-if="exportFileGuideDocumentLanguageName">[{{ exportFileGuideDocumentLanguageName }}]</span>
                    </a>
                </div>

                <v-btn density="compact" color="default" variant="text" class="ms-2"
                       :aria-label="tt('Refresh Accounts, Categories and Tags')" :icon="true" :disabled="loading"
                       :loading="loading"
                       @click="reloadBasisData"
                       v-if="currentStep === 'checkData'">
                    <template #loader>
                        <v-progress-circular indeterminate size="20"/>
                    </template>
                    <v-icon :icon="mdiRefresh" size="22" />
                    <v-tooltip activator="parent">{{ tt('Refresh Accounts, Categories and Tags') }}</v-tooltip>
                </v-btn>

                <v-btn class="ms-2" density="comfortable" color="primary" variant="outlined"
                       :disabled="loading || submitting || importTransactionCheckDataTabRef?.isEditing"
                       @click="batchApplyRules"
                       v-if="currentStep === 'checkData'">
                    {{ tt('Batch Apply Rules') }}
                </v-btn>
            </template>

            <template #toolbar>
                <v-btn class="ms-2 me-1" density="comfortable" variant="outlined" color="primary"
                       :disabled="loading || submitting || (!isImportDataFromTextbox && !isAIImageImport && !importFile) || (isImportDataFromTextbox && !importData) || (!isImportDataFromTextbox && !isAIImageImport && allSupportedEncodings && fileEncoding === 'auto' && !autoDetectedFileEncoding) || (!isImportDataFromTextbox && isAIImageImport && importImageFiles.length === 0)"
                       @click="parseData"
                       v-if="currentStep === 'defineColumn' || currentStep === 'executeCustomScript' || (currentStep === 'recognizeImages' && !submitting) || currentStep === 'uploadFile'">
                    {{ tt('Next') }}
                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                </v-btn>

                <v-btn class="ms-2 me-1" density="comfortable" variant="outlined" color="secondary"
                       :disabled="loading"
                       @click="cancelBatchRecognizeImages()"
                       v-if="currentStep === 'recognizeImages' && submitting">{{ tt('Cancel Recognition') }}</v-btn>

                <v-btn class="ms-2" density="comfortable" variant="outlined"
                       :disabled="submitting || importTransactionCheckDataTabRef?.isEditing || !importTransactionCheckDataTabRef?.canImport"
                       @click="submit"
                       v-if="currentStep === 'checkData'">
                    {{ (submitting && importProcess > 0 ? tt('format.misc.importingTransactions', { process: formatNumberToLocalizedNumerals(importProcess, 2) }) : tt('Import')) }}
                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                </v-btn>
                <v-btn density="compact" color="default" variant="text" class="ms-2"
                       :aria-label="tt('More')" :disabled="loading || submitting" :icon="true"
                       v-if="currentStep === 'defineColumn' && importTransactionDefineColumnTabRef?.menus">
                    <v-icon :icon="mdiDotsVertical" />
                    <v-menu activator="parent" max-height="500">
                        <v-list>
                            <v-list-item :key="index"
                                         :prepend-icon="menu.prependIcon"
                                         :title="menu.title"
                                         :disabled="menu.disabled"
                                         @click="menu.onClick()"
                                         v-for="(menu, index) in importTransactionDefineColumnTabRef.menus"/>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-btn density="compact" color="default" variant="text" class="ms-2"
                       :aria-label="tt('More')" :disabled="loading || submitting" :icon="true"
                       v-if="currentStep === 'executeCustomScript' && importTransactionExecuteCustomScriptTabRef?.menus">
                    <v-icon :icon="mdiDotsVertical" />
                    <v-menu activator="parent" max-height="500">
                        <v-list>
                            <v-list-item :key="index"
                                         :prepend-icon="menu.prependIcon"
                                         :title="menu.title"
                                         :disabled="menu.disabled"
                                         @click="menu.onClick()"
                                         v-for="(menu, index) in importTransactionExecuteCustomScriptTabRef.menus"/>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-btn density="compact" color="default" variant="text" class="ms-2"
                       :aria-label="tt('Filter')" :icon="true" :disabled="loading || submitting || importTransactionCheckDataTabRef?.isEditing"
                       v-if="currentStep === 'checkData' && importTransactionCheckDataTabRef?.filterMenus">
                    <v-icon :icon="mdiFilterOutline" />
                    <v-menu activator="parent" max-height="500">
                        <v-list>
                            <template :key="groupIndex" v-for="(group, groupIndex) in importTransactionCheckDataTabRef.filterMenus">
                                <v-divider class="my-2" v-if="groupIndex > 0" />
                                <v-list-subheader class="text-body-small" :title="group.title" v-if="group.title" />
                                <v-list-item :key="`menu_${groupIndex}_${index}`"
                                             :prepend-icon="menu.prependIcon"
                                             :title="menu.title"
                                             :subtitle="menu.subTitle"
                                             :append-icon="menu.appendIcon"
                                             :disabled="menu.disabled"
                                             @click="menu.onClick()"
                                             v-for="(menu, index) in group.items" />
                            </template>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-btn density="compact" color="default" variant="text" class="ms-2"
                       :aria-label="tt('More')" :icon="true" :disabled="loading || submitting"
                       v-if="currentStep === 'checkData' && importTransactionCheckDataTabRef?.toolMenus">
                    <v-icon :icon="mdiDotsVertical" />
                    <v-menu activator="parent" max-height="500">
                        <v-list>
                            <template :key="index" v-for="(menu, index) in importTransactionCheckDataTabRef.toolMenus">
                                <v-divider class="my-2" v-if="menu.divider" />
                                <v-list-item :prepend-icon="menu.prependIcon"
                                             :title="menu.title"
                                             :subtitle="menu.subTitle"
                                             :append-icon="menu.appendIcon"
                                             :disabled="menu.disabled"
                                             @click="menu.onClick()" />
                            </template>
                        </v-list>
                    </v-menu>
                </v-btn>
            </template>

            <template #subtitle>
                <v-divider class="mt-2" />
                <div class="cursor-default mt-3 mx-3 mb-md-2">
                    <steps-bar min-width="700" :always-horizontal="true" :clickable="false"
                               :steps="allSteps" :current-step="currentStep" />
                </div>
            </template>

            <template #content>
                <v-window class="disable-tab-transition"
                          :class="{ 'import-transaction-dialog-window': currentStep === 'defineColumn' || currentStep === 'checkData' }"
                          v-model="currentStep">
                    <v-window-item value="uploadFile">
                        <div class="pa-4">
                            <v-row>
                                <v-col cols="12" md="12">
                                    <two-column-select primary-key-field="displayCategoryName"
                                                       primary-value-field="displayCategoryName"
                                                       primary-title-field="displayCategoryName"
                                                       primary-sub-items-field="fileTypes"
                                                       secondary-key-field="type"
                                                       secondary-value-field="type"
                                                       secondary-title-field="displayName"
                                                       :disabled="submitting"
                                                       :enable-filter="true"
                                                       :filter-placeholder="tt('Find file type')"
                                                       :filter-no-items-text="tt('No available file type')"
                                                       :label="tt('File Type')"
                                                       :placeholder="tt('File Type')"
                                                       :items="allSupportedImportFileCategoryAndTypes"
                                                       :auto-update-menu-position="true"
                                                       v-model="fileType">
                                    </two-column-select>
                                </v-col>

                                <v-col cols="12" md="12" v-if="allFileSubTypes">
                                    <v-select
                                        item-title="displayName"
                                        item-value="type"
                                        :disabled="submitting"
                                        :label="tt('Format')"
                                        :placeholder="tt('Format')"
                                        :items="allFileSubTypes"
                                        v-model="fileSubType"
                                    />
                                </v-col>

                                <v-col cols="12" md="12" v-if="!isImportDataFromTextbox && allSupportedEncodings">
                                    <v-select
                                        item-title="displayName"
                                        item-value="encoding"
                                        :disabled="submitting"
                                        :label="tt('File Encoding')"
                                        :placeholder="tt('File Encoding')"
                                        :items="allSupportedEncodings"
                                        v-model="fileEncoding"
                                    />
                                </v-col>

                                <v-col cols="12" md="12" v-if="isCustomFileFormat">
                                    <v-select
                                        item-title="displayName"
                                        item-value="type"
                                        :disabled="submitting"
                                        :label="tt('Handling Method')"
                                        :placeholder="tt('Handling Method')"
                                        :items="[
                                        { displayName: tt('Column Mapping'), type: ImportCustomFileFormatProcessMethod.ColumnMapping },
                                        { displayName: tt('Custom Script'), type: ImportCustomFileFormatProcessMethod.CustomScript }
                                     ]"
                                        v-model="processCustomFileFormatMethod"
                                    />
                                </v-col>

                                <v-col cols="12" md="12" v-if="supportedAdditionalOptions">
                                    <v-select
                                        :disabled="submitting"
                                        :label="tt('Additional Options')"
                                        :placeholder="tt('Additional Options')"
                                        v-model="fileType"
                                        v-model:menu="additionalOptionsMenuState"
                                    >
                                        <template #selection>
                                            <span class="cursor-pointer">{{ displaySelectedAdditionalOptions }}</span>
                                        </template>

                                        <template #no-data>
                                            <v-list class="py-0" select-strategy="classic"
                                                    :selected="objectFieldWithValueToArrayItem(importAdditionalOptions, true as (boolean | undefined))">
                                                <template v-for="item in allSupportedAdditionalOptions">
                                                    <v-list-item class="text-body-medium" :key="item.key" :value="item.key"
                                                                 :append-icon="importAdditionalOptions[item.key] ? mdiCheck : undefined"
                                                                 @click="importAdditionalOptions[item.key] = !importAdditionalOptions[item.key]"
                                                                 v-if="isDefined(supportedAdditionalOptions[item.key])">{{ tt(item.name) }}</v-list-item>
                                                </template>
                                            </v-list>
                                        </template>
                                    </v-select>
                                </v-col>

                                <v-col cols="12" md="12" v-if="supportedAIAdditionalPrompt">
                                    <v-textarea
                                        type="text"
                                        autocomplete="off"
                                        persistent-placeholder
                                        rows="2"
                                        :disabled="submitting"
                                        :placeholder="tt('Additional Prompt')"
                                        v-model="importAIAdditionalPrompt"
                                    />
                                </v-col>

                                <v-col cols="12" md="12" v-if="!isImportDataFromTextbox && !isAIImageImport">
                                    <v-text-field
                                        readonly
                                        persistent-placeholder
                                        type="text"
                                        class="always-cursor-pointer"
                                        :disabled="submitting"
                                        :label="tt('Data File')"
                                        :placeholder="tt('format.misc.clickToSelectedFile', { extensions: supportedImportFileExtensions })"
                                        v-model="fileName"
                                        @click="showOpenFileDialog"
                                    />
                                </v-col>

                                <v-col cols="12" md="12" v-if="isImportDataFromTextbox">
                                    <v-textarea
                                        type="text"
                                        autocomplete="off"
                                        persistent-placeholder
                                        rows="5"
                                        :disabled="submitting"
                                        :placeholder="tt('Data to import')"
                                        v-model="importData"
                                    />
                                </v-col>

                                <v-col cols="12" md="12" v-if="isAIImageImport">
                                    <div class="text-body-large mb-2">{{ tt('Image Files') }}</div>
                                    <div class="import-transaction-images d-flex gap-2 overflow-x-auto">
                                        <div :key="picIdx" v-for="(imageItem, picIdx) in importImageFiles">
                                            <v-avatar rounded="lg" variant="tonal" size="120"
                                                      class="cursor-pointer import-image"
                                                      color="rgba(0,0,0,0)" @click="removeImportImageFile(picIdx)">
                                                <v-img :src="imageItem.previewUrl"></v-img>
                                                <div class="picture-control-icon">
                                                    <v-icon size="48" :aria-label="tt('Remove Picture')" :icon="mdiTrashCanOutline" />
                                                </div>
                                            </v-avatar>
                                        </div>
                                        <div>
                                            <v-avatar rounded="lg" variant="tonal" size="120"
                                                      class="import-image import-image-add"
                                                      :class="{ 'enabled': !submitting, 'cursor-pointer': !submitting }"
                                                      color="rgba(0,0,0,0)" @click="showOpenFileDialog">
                                                <v-tooltip activator="parent" v-if="!submitting">{{ tt('Add Picture') }}</v-tooltip>
                                                <v-icon class="import-image-add-icon" size="48" :aria-label="tt('Add Picture')" :icon="mdiImagePlusOutline" />
                                            </v-avatar>
                                        </div>
                                    </div>
                                </v-col>
                            </v-row>
                        </div>
                    </v-window-item>
                    <v-window-item value="defineColumn">
                        <import-transaction-define-column-tab
                            ref="importTransactionDefineColumnTab"
                            :parsed-file-data="parsedFileData"
                            :parsed-file-column-separator="parsedFileColumnSeparator"
                            :disabled="loading || submitting"
                        />
                    </v-window-item>
                    <v-window-item value="executeCustomScript">
                        <import-transaction-execute-custom-script-tab
                            ref="importTransactionExecuteCustomScriptTab"
                            :parsed-file-data="parsedFileData"
                            :disabled="loading || submitting"
                        />
                    </v-window-item>
                    <v-window-item value="recognizeImages">
                        <import-transaction-recognize-images-tab
                            ref="importTransactionRecognizeImagesTab"
                            :disabled="loading || submitting"
                            :submitting="submitting"
                            :import-images="importImageFiles"
                        />
                    </v-window-item>
                    <v-window-item value="checkData">
                        <import-transaction-check-data-tab
                            ref="importTransactionCheckDataTab"
                            :import-transactions="importTransactions"
                            :disabled="loading || submitting"
                        />
                    </v-window-item>
                    <v-window-item value="finalResult">
                        <div class="mx-4 my-4">
                            <v-alert type="success" color="success-darken-1" variant="tonal">{{ tt('Data Import Completed') }}</v-alert>
                            <div class="text-body-large my-4">{{ tt('format.misc.importTransactionResult', { count: formatNumberToLocalizedNumerals(importedCount || 0) }) }}</div>
                        </div>
                    </v-window-item>
                </v-window>
            </template>

            <template #footer v-if="currentStep === 'uploadFile' && (needAITextRecognition || needAIImageRecognition)">
                <div class="w-100 text-body-large text-truncate">
                    <span v-if="needAITextRecognition">
                        {{ tt('Uploaded text and personal data will be sent to the large language model, please be aware of potential privacy risks.') }}
                    </span>
                    <span v-if="needAIImageRecognition">
                        {{ tt('Uploaded image and personal data will be sent to the large language model, please be aware of potential privacy risks.') }}
                    </span>
                </div>
            </template>
        </one-column-dialog-layout>
    </v-dialog>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
    <input ref="fileInput" type="file" style="display: none" :accept="supportedImportFileExtensions" :multiple="isAIImageImport" @change="setImportFile($event)" />
</template>



<style>
.import-transaction-dialog-layout {
    max-height: calc(100dvh - 48px);

    .import-transaction-dialog-content {
        min-height: 0;
        overflow-y: hidden !important;

        .import-transaction-dialog-window {
            min-height: 0;

            > .v-window__container {
                height: auto;
            }

            .v-window-item--active {
                min-height: 0;
            }

            .import-transaction-table-container {
                max-height: calc(100dvh - 164px);
                min-height: 0;
                overflow-y: hidden !important;

                .import-transaction-table {
                    flex: 1 1 auto;
                    min-height: 0;
                    overflow: hidden;

                    > .v-table__wrapper {
                        height: 100% !important;
                        min-height: 0;
                        overflow-y: auto;
                    }
                }

                .import-transaction-table-footer {
                    flex: 0 0 auto;
                }
            }
        }
    }
}

.import-transaction-images {
    .import-image {
        .picture-control-icon {
            display: none;
            position: absolute;
            width: 100% !important;
            height: 100% !important;
            background-color: rgba(0, 0, 0, 0.4);
        }

        .picture-control-icon > i.v-icon {
            background-color: transparent;
            color: rgba(255, 255, 255, 0.8);
        }
    }

    .import-image:hover {
        .picture-control-icon {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            vertical-align: middle;
        }
    }

    .import-image-add {
        border: 2px dashed rgba(var(--v-theme-grey-500));
    }

    .import-image-add .import-image-add-icon {
        color: rgba(var(--v-theme-grey-500));
    }

    .import-image-add.enabled:hover {
        border: 2px dashed rgba(var(--v-theme-grey-700));
    }

    .import-image-add.enabled:hover .import-image-add-icon {
        color: rgba(var(--v-theme-grey-700));
    }
}
</style>
