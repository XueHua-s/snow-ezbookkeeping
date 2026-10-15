import { detectFileEncoding } from '@/lib/file.ts';
import type { useImportDialogState } from './state.ts';

interface Context {
    isAIImageImport: ReturnType<typeof useImportDialogState>['isAIImageImport'];
    importImageFiles: ReturnType<typeof useImportDialogState>['importImageFiles'];
    importFile: ReturnType<typeof useImportDialogState>['importFile'];
    detectingFileEncoding: ReturnType<typeof useImportDialogState>['detectingFileEncoding'];
    autoDetectedFileEncoding: ReturnType<typeof useImportDialogState>['autoDetectedFileEncoding'];
    allSupportedEncodings: ReturnType<typeof useImportDialogState>['allSupportedEncodings'];
    loading: ReturnType<typeof useImportDialogState>['loading'];
    accountsStore: ReturnType<typeof useImportDialogState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useImportDialogState>['transactionCategoriesStore'];
    transactionTagsStore: ReturnType<typeof useImportDialogState>['transactionTagsStore'];
    snackbar: ReturnType<typeof useImportDialogState>['snackbar'];
    importTransactionCheckDataTab: ReturnType<typeof useImportDialogState>['importTransactionCheckDataTab'];
}

export function useImportDialogFiles(context: Context) {
    const {
        isAIImageImport, importImageFiles, importFile, detectingFileEncoding, autoDetectedFileEncoding, allSupportedEncodings,
        loading, accountsStore, transactionCategoriesStore, transactionTagsStore, snackbar, importTransactionCheckDataTab,
    } = context;

    function setImportFile(event: Event): void {
        if (!event || !event.target) {
            return;
        }

        const el = event.target as HTMLInputElement;

        if (!el.files || !el.files.length || !el.files[0]) {
            el.value = '';
            return;
        }

        if (isAIImageImport.value) {
            for (const file of el.files) {
                importImageFiles.value.push({
                    file: file,
                    previewUrl: URL.createObjectURL(file),
                    status: 'pending'
                });
            }

            el.value = '';
        } else {
            importFile.value = el.files[0] as File;
            detectingFileEncoding.value = false;
            autoDetectedFileEncoding.value = undefined;
            el.value = '';

            if (allSupportedEncodings.value) {
                detectingFileEncoding.value = true;

                detectFileEncoding(importFile.value).then(detectedEncoding => {
                    detectingFileEncoding.value = false;
                    autoDetectedFileEncoding.value = detectedEncoding;
                }).catch(() => {
                    detectingFileEncoding.value = false;
                    autoDetectedFileEncoding.value = undefined;
                });
            }
        }
    }

    function removeImportImageFile(index: number): void {
        const targetImageFile = importImageFiles.value[index];

        if (targetImageFile && targetImageFile.previewUrl) {
            URL.revokeObjectURL(targetImageFile.previewUrl);
        }

        importImageFiles.value.splice(index, 1);
    }

    function clearImportImageFiles(): void {
        for (const item of importImageFiles.value) {
            if (item.previewUrl) {
                URL.revokeObjectURL(item.previewUrl);
            }
        }

        importImageFiles.value.length = 0;
    }

    function reloadBasisData(): void {
        loading.value = true;

        Promise.allSettled([
            accountsStore.loadAllAccounts({ force: true }),
            transactionCategoriesStore.loadAllCategories({ force: true }),
            transactionTagsStore.loadAllTags({ force: true })
        ]).then(results => {
            loading.value = false;

            const isAllUpToDate = results.length === 3
                && results[0].status === 'rejected' && results[0].reason?.isUpToDate
                && results[1].status === 'rejected' && results[1].reason?.isUpToDate
                && results[2].status === 'rejected' && results[2].reason?.isUpToDate;

            // show info if all up to date
            if (isAllUpToDate) {
                snackbar.value?.showMessage('Data is up to date');
                return;
            }

            // show error if any
            for (const result of results) {
                if (result.status === 'rejected' && !result.reason?.isUpToDate) {
                    snackbar.value?.showError(result.reason);
                    return;
                }
            }

            // show info if one of them updated
            for (const result of results) {
                if (result.status === 'fulfilled') {
                    snackbar.value?.showMessage('Data has been updated');
                    importTransactionCheckDataTab.value?.updateAllTransactionsIsValid();
                    return;
                }
            }
        });
    }

    return { setImportFile, removeImportImageFile, clearImportImageFiles, reloadBasisData };
}
