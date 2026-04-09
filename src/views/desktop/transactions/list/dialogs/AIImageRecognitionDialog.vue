<template>
    <v-dialog width="900" :persistent="loading || recognizing || imageItems.length > 0" v-model="showState" @paste="onPaste">
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <h4 class="text-h4">{{ tt('AI Image Recognition') }}</h4>
            </template>

            <v-card-text class="d-flex flex-column flex-md-row flex-grow-1 overflow-y-auto" style="height: 480px">
                <!-- Left: Image thumbnails area -->
                <div class="w-50 h-100 border position-relative me-2 d-flex flex-column"
                     @dragenter.prevent="onDragEnter"
                     @dragover.prevent
                     @dragleave.prevent="onDragLeave"
                     @drop.prevent="onDrop">

                    <!-- Empty state / drag overlay -->
                    <div class="d-flex w-100 h-100 justify-center align-center justify-content-center text-center px-4 position-absolute"
                         style="z-index: 10; top: 0; left: 0;"
                         :class="{ 'dropzone': true, 'dropzone-dark': isDarkMode, 'dropzone-blurry-bg': isDragOver }"
                         v-if="imageItems.length === 0 || isDragOver">
                        <div class="d-inline-flex flex-column" v-if="!loading && imageItems.length === 0 && !isDragOver">
                            <h3 class="pa-2">{{ tt('You can drag and drop, paste or click to select receipt or transaction images') }}</h3>
                            <span class="pa-2">{{ tt('Uploaded image and personal data will be sent to the large language model, please be aware of potential privacy risks.') }}</span>
                            <v-btn class="mt-4 mx-auto" variant="outlined" @click="showOpenImageDialog">
                                {{ tt('Select Images') }}
                            </v-btn>
                        </div>
                        <h3 class="pa-2" v-else-if="isDragOver">{{ tt('Release to load image') }}</h3>
                    </div>

                    <!-- Thumbnail grid -->
                    <div class="flex-grow-1 overflow-y-auto pa-2" v-if="imageItems.length > 0"
                         @click.self="showOpenImageDialog">
                        <div class="d-flex flex-wrap gap-2">
                            <div v-for="(item, index) in imageItems" :key="index"
                                 class="image-thumb-wrapper position-relative"
                                 :class="{
                                     'image-thumb-processing': item.status === 'processing',
                                     'image-thumb-done': item.status === 'done',
                                     'image-thumb-error': item.status === 'error'
                                 }">
                                <v-img :src="item.src" width="120" height="90" cover class="rounded border" />
                                <v-btn v-if="!recognizing" icon size="x-small" color="error" variant="flat"
                                       class="position-absolute" style="top: -6px; right: -6px; z-index: 2;"
                                       @click="removeImage(index)">
                                    <v-icon size="14">{{ mdiClose }}</v-icon>
                                </v-btn>
                                <div v-if="item.status === 'processing'" class="image-thumb-overlay d-flex align-center justify-center">
                                    <v-progress-circular indeterminate size="24" color="white" />
                                </div>
                                <div v-else-if="item.status === 'done'" class="image-thumb-overlay-icon">
                                    <v-icon color="success" size="20">{{ mdiCheckCircle }}</v-icon>
                                </div>
                                <div v-else-if="item.status === 'error'" class="image-thumb-overlay-icon">
                                    <v-icon color="error" size="20">{{ mdiAlertCircle }}</v-icon>
                                </div>
                            </div>
                            <!-- Add more button -->
                            <div v-if="!recognizing" class="image-thumb-add d-flex align-center justify-center rounded border cursor-pointer"
                                 @click="showOpenImageDialog">
                                <v-icon size="32" color="grey">{{ mdiPlus }}</v-icon>
                            </div>
                        </div>
                    </div>

                    <!-- Image count badge -->
                    <div v-if="imageItems.length > 0" class="pa-2 border-t text-caption text-center">
                        {{ tt('{count} image(s) selected', { count: imageItems.length }) }}
                    </div>
                </div>

                <!-- Right: Results area -->
                <div class="w-50 h-100 border ms-2 d-flex flex-column">
                    <div v-if="results.length === 0" class="d-flex w-100 h-100 align-center justify-center text-center px-4">
                        <span class="text-grey" v-if="!recognizing">{{ tt('Recognition results will appear here') }}</span>
                        <div v-else class="d-flex flex-column align-center">
                            <v-progress-circular indeterminate size="40" class="mb-3" />
                            <span>{{ tt('Recognizing image {current} of {total}...', { current: recognizingIndex + 1, total: imageItems.length }) }}</span>
                        </div>
                    </div>

                    <div v-else class="flex-grow-1 overflow-y-auto">
                        <v-list density="compact" class="pa-0">
                            <v-list-item v-for="(item, index) in results" :key="index"
                                         :class="{ 'bg-red-lighten-5': !item.success }">
                                <template #prepend>
                                    <v-checkbox-btn v-if="item.success"
                                                      :model-value="selectedIndices.has(index)"
                                                      @update:model-value="toggleSelect(index, $event)"
                                                      density="compact" hide-details />
                                    <v-icon v-else color="error" size="20" class="me-2">{{ mdiAlertCircle }}</v-icon>
                                </template>
                                <v-list-item-title v-if="item.success && item.result" class="text-body-2">
                                    <v-chip size="x-small" :color="getTypeColor(item.result.type)" class="me-1">
                                        {{ getTypeName(item.result.type) }}
                                    </v-chip>
                                    <span v-if="item.result.sourceAmount">{{ formatAmount(item.result.sourceAmount) }}</span>
                                </v-list-item-title>
                                <v-list-item-title v-else class="text-body-2 text-error">
                                    #{{ item.index + 1 }}: {{ item.error || tt('Failed') }}
                                </v-list-item-title>
                                <v-list-item-subtitle v-if="item.success && item.result" class="text-caption">
                                    {{ item.result.comment || '' }}
                                </v-list-item-subtitle>
                            </v-list-item>
                        </v-list>
                    </div>

                    <div v-if="results.length > 0" class="pa-2 border-t text-caption text-center">
                        {{ tt('Recognition complete') }}: {{ successCount }}/{{ results.length }}
                    </div>
                </div>
            </v-card-text>

            <v-card-text>
                <div class="w-100 d-flex justify-center flex-wrap mt-sm-1 mt-md-2 gap-4">
                    <!-- Single image mode: Recognize button -->
                    <v-btn v-if="results.length === 0"
                           :disabled="loading || recognizing || imageItems.length === 0"
                           @click="recognize">
                        {{ tt('Recognize') }}
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="recognizing"></v-progress-circular>
                    </v-btn>
                    <!-- Multi result: Import Selected -->
                    <v-btn v-if="results.length > 0 && successCount > 0"
                           color="primary"
                           :disabled="selectedIndices.size === 0"
                           @click="importSelected">
                        {{ tt('Import Selected') }} ({{ selectedIndices.size }})
                    </v-btn>
                    <v-btn color="secondary" variant="tonal" :disabled="loading"
                           @click="cancelRecognize" v-if="recognizing && cancelRecognizingUuid">{{ tt('Cancel Recognition') }}</v-btn>
                    <v-btn color="secondary" variant="tonal" :disabled="loading || recognizing"
                           @click="cancel" v-if="!recognizing || !cancelRecognizingUuid">{{ tt('Cancel') }}</v-btn>
                </div>
            </v-card-text>
        </v-card>
    </v-dialog>

    <snack-bar ref="snackbar" />
    <input ref="imageInput" type="file" style="display: none" multiple :accept="SUPPORTED_IMAGE_EXTENSIONS" @change="openImages($event)" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, useTemplateRef } from 'vue';
import { useTheme } from 'vuetify';

import { useI18n } from '@/locales/helpers.ts';

import { useTransactionsStore } from '@/stores/transaction.ts';

import { KnownFileType } from '@/core/file.ts';
import { TransactionType } from '@/core/transaction.ts';
import { ThemeType } from '@/core/theme.ts';
import { SUPPORTED_IMAGE_EXTENSIONS } from '@/consts/file.ts';

import type { RecognizedReceiptImageResponse, RecognizedReceiptImageResultItem } from '@/models/large_language_model.ts';

import { generateRandomUUID } from '@/lib/misc.ts';
import { compressJpgImage } from '@/lib/ui/common.ts';
import logger from '@/lib/logger.ts';

import {
    mdiClose,
    mdiPlus,
    mdiCheckCircle,
    mdiAlertCircle
} from '@mdi/js';

type SnackBarType = InstanceType<typeof SnackBar>;

interface ImageItem {
    file: File;
    src: string;
    status: 'pending' | 'processing' | 'done' | 'error';
}

const theme = useTheme();

const { tt } = useI18n();

const transactionsStore = useTransactionsStore();

const snackbar = useTemplateRef<SnackBarType>('snackbar');
const imageInput = useTemplateRef<HTMLInputElement>('imageInput');

let resolveFunc: ((responses: RecognizedReceiptImageResponse[]) => void) | null = null;
let rejectFunc: ((reason?: unknown) => void) | null = null;

const showState = ref<boolean>(false);
const loading = ref<boolean>(false);
const recognizing = ref<boolean>(false);
const recognizingIndex = ref<number>(0);
const cancelRecognizingUuid = ref<string | undefined>(undefined);
const imageItems = ref<ImageItem[]>([]);
const results = ref<RecognizedReceiptImageResultItem[]>([]);
const selectedIndices = ref<Set<number>>(new Set());
const isDragOver = ref<boolean>(false);

const isDarkMode = computed<boolean>(() => theme.global.name.value === ThemeType.Dark);
const successCount = computed<number>(() => results.value.filter(r => r.success).length);

function loadImages(files: File[]): void {
    loading.value = true;

    const promises = files.map(file =>
        compressJpgImage(file, 1280, 1280, 0.8).then(blob => {
            const compressedFile = KnownFileType.JPG.createFileFromBlob(blob, "image");
            const src = URL.createObjectURL(blob);
            imageItems.value.push({ file: compressedFile, src, status: 'pending' });
        }).catch(error => {
            logger.error('failed to compress image', error);
        })
    );

    Promise.all(promises).then(() => {
        loading.value = false;
    }).catch(() => {
        loading.value = false;
        snackbar.value?.showError('Unable to load image');
    });
}

function removeImage(index: number): void {
    if (imageItems.value[index]?.src) {
        URL.revokeObjectURL(imageItems.value[index].src);
    }
    imageItems.value.splice(index, 1);
    // Reset results if images changed
    results.value = [];
    selectedIndices.value.clear();
}

function open(): Promise<RecognizedReceiptImageResponse[]> {
    showState.value = true;
    loading.value = false;
    recognizing.value = false;
    recognizingIndex.value = 0;
    cancelRecognizingUuid.value = undefined;
    imageItems.value = [];
    results.value = [];
    selectedIndices.value.clear();

    return new Promise((resolve, reject) => {
        resolveFunc = resolve;
        rejectFunc = reject;
    });
}

function showOpenImageDialog(): void {
    if (loading.value || recognizing.value) {
        return;
    }

    imageInput.value?.click();
}

function openImages(event: Event): void {
    if (!event || !event.target) {
        return;
    }

    const el = event.target as HTMLInputElement;

    if (!el.files || !el.files.length) {
        return;
    }

    const files = Array.from(el.files) as File[];
    el.value = '';

    loadImages(files);
}

function recognize(): void {
    if (loading.value || recognizing.value || imageItems.value.length === 0) {
        return;
    }

    cancelRecognizingUuid.value = generateRandomUUID();
    recognizing.value = true;
    recognizingIndex.value = 0;
    results.value = [];
    selectedIndices.value.clear();

    // Mark all as processing
    for (const item of imageItems.value) {
        item.status = 'processing';
    }

    const files = imageItems.value.map(item => item.file);

    transactionsStore.recognizeReceiptImages({
        imageFiles: files,
        cancelableUuid: cancelRecognizingUuid.value
    }).then(response => {
        results.value = response.results;

        // Update image statuses and auto-select successful results
        for (const resultItem of response.results) {
            const imageItem = imageItems.value[resultItem.index];
            if (imageItem) {
                imageItem.status = resultItem.success ? 'done' : 'error';
            }

            if (resultItem.success) {
                selectedIndices.value.add(resultItem.index);
            }
        }

        recognizing.value = false;
        cancelRecognizingUuid.value = undefined;
    }).catch(error => {
        if (error.canceled) {
            return;
        }

        // Mark all as error
        for (const item of imageItems.value) {
            if (item.status === 'processing') {
                item.status = 'error';
            }
        }

        recognizing.value = false;
        cancelRecognizingUuid.value = undefined;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function toggleSelect(index: number, value: boolean): void {
    if (value) {
        selectedIndices.value.add(index);
    } else {
        selectedIndices.value.delete(index);
    }
    // Trigger reactivity
    selectedIndices.value = new Set(selectedIndices.value);
}

function importSelected(): void {
    const selected: RecognizedReceiptImageResponse[] = [];

    for (const index of selectedIndices.value) {
        const item = results.value[index];
        if (item?.success && item.result) {
            selected.push(item.result);
        }
    }

    if (selected.length === 0) {
        snackbar.value?.showError('No results to import');
        return;
    }

    resolveFunc?.(selected);
    showState.value = false;
}

function cancelRecognize(): void {
    if (!cancelRecognizingUuid.value) {
        return;
    }

    transactionsStore.cancelRecognizeReceiptImage(cancelRecognizingUuid.value);
    recognizing.value = false;
    cancelRecognizingUuid.value = undefined;

    for (const item of imageItems.value) {
        if (item.status === 'processing') {
            item.status = 'pending';
        }
    }

    snackbar.value?.showMessage('User Canceled');
}

function cancel(): void {
    rejectFunc?.();
    showState.value = false;
    cleanup();
}

function cleanup(): void {
    for (const item of imageItems.value) {
        if (item.src) {
            URL.revokeObjectURL(item.src);
        }
    }
    loading.value = false;
    recognizing.value = false;
    cancelRecognizingUuid.value = undefined;
    imageItems.value = [];
    results.value = [];
    selectedIndices.value.clear();
}

function getTypeColor(type: number): string {
    switch (type) {
        case TransactionType.Income: return 'success';
        case TransactionType.Expense: return 'error';
        case TransactionType.Transfer: return 'info';
        default: return 'grey';
    }
}

function getTypeName(type: number): string {
    switch (type) {
        case TransactionType.Income: return tt('Income');
        case TransactionType.Expense: return tt('Expense');
        case TransactionType.Transfer: return tt('Transfer');
        default: return '';
    }
}

function formatAmount(amount: number): string {
    return (amount / 100).toFixed(2);
}

function onDragEnter(): void {
    if (loading.value || recognizing.value) {
        return;
    }

    isDragOver.value = true;
}

function onDragLeave(): void {
    isDragOver.value = false;
}

function onDrop(event: DragEvent): void {
    if (loading.value || recognizing.value) {
        return;
    }

    isDragOver.value = false;

    if (event.dataTransfer && event.dataTransfer.files && event.dataTransfer.files.length) {
        const files = Array.from(event.dataTransfer.files) as File[];
        loadImages(files);
    }
}

function onPaste(event: ClipboardEvent) {
    if (!event.clipboardData) {
        event.preventDefault();
        return;
    }

    const files: File[] = [];

    for (let i = 0; i < event.clipboardData.items.length; i++) {
        const item = event.clipboardData.items[i];

        if (item && item.type.startsWith('image/')) {
            const file = item.getAsFile();
            if (file) {
                files.push(file);
            }
        }
    }

    if (files.length > 0) {
        loadImages(files);
        event.preventDefault();
    }
}

defineExpose({
    open
});
</script>

<style>
.dropzone {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    pointer-events: none;
    border-radius: 8px;
    z-index: 10;

    h3, span {
        color: rgb(var(--v-theme-on-grey-200)) !important;
        text-shadow: -1px -1px 0 #fff, 1px -1px 0 #fff, -1px 1px 0 #fff, 1px 1px 0 #fff;
    }

    &.dropzone-dark {
        h3, span {
            color: rgb(var(--v-theme-on-grey-100)) !important;
            text-shadow: -1px -1px 0 #000, 1px -1px 0 #000, -1px 1px 0 #000, 1px 1px 0 #000;
        }
    }
}

.dropzone-blurry-bg {
    /* stylelint-disable property-no-vendor-prefix */
    -webkit-backdrop-filter: blur(6px);
    backdrop-filter: blur(6px);
}

.dropzone-dragover {
    border: 6px dashed rgba(var(--v-border-color),var(--v-border-opacity));
}

.image-thumb-wrapper {
    width: 120px;
    height: 90px;
    flex-shrink: 0;
}

.image-thumb-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.4);
    border-radius: 4px;
}

.image-thumb-overlay-icon {
    position: absolute;
    bottom: 2px;
    right: 2px;
}

.image-thumb-add {
    width: 120px;
    height: 90px;
    flex-shrink: 0;
    border-style: dashed !important;
}

.image-thumb-error {
    border: 2px solid rgb(var(--v-theme-error));
    border-radius: 4px;
}
</style>
