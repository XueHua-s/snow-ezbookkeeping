import type { SetupProps } from '@/lib/ui/props.ts';
import {
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { Router } from 'framework7/types';
import type { useEditPageInitialization } from './initialization.ts';
import type { useEditPageLocation } from './location.ts';
import type { useEditPageState } from './state.ts';

interface Context {
    routeBackOnError: ReturnType<typeof useEditPageState>['routeBackOnError'];
    props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
    autoUploadPicture?: File;
    autoRecognizeClipboardText?: string;
}>;
    loadingError: ReturnType<typeof useEditPageState>['loadingError'];
    settingsStore: ReturnType<typeof useEditPageState>['settingsStore'];
    mode: ReturnType<typeof useEditPageState>['mode'];
    geoLocationStatus: ReturnType<typeof useEditPageState>['geoLocationStatus'];
    transaction: ReturnType<typeof useEditPageState>['transaction'];
    updateGeoLocation: ReturnType<typeof useEditPageLocation>['updateGeoLocation'];
    submitted: ReturnType<typeof useEditPageState>['submitted'];
    pageTypeAndMode: ReturnType<typeof useEditPageState>['pageTypeAndMode'];
    query: ReturnType<typeof useEditPageState>['query'];
    addByTemplateId: ReturnType<typeof useEditPageState>['addByTemplateId'];
    duplicateFromId: ReturnType<typeof useEditPageState>['duplicateFromId'];
    transactionsStore: ReturnType<typeof useEditPageState>['transactionsStore'];
    firstVisibleAccountId: ReturnType<typeof useEditPageState>['firstVisibleAccountId'];
    showConfirm: ReturnType<typeof useEditPageState>['showConfirm'];
    init: ReturnType<typeof useEditPageInitialization>['init'];
}

export function useEditPageLifecycle(context: Context) {
    const {
        routeBackOnError, props, loadingError, settingsStore, mode, geoLocationStatus, transaction, updateGeoLocation, submitted,
        pageTypeAndMode, query, addByTemplateId, duplicateFromId, transactionsStore, firstVisibleAccountId, showConfirm, init,
    } = context;

    function onPageAfterIn(): void {
        routeBackOnError(props.f7router, loadingError);

        if (settingsStore.appSettings.autoGetCurrentGeoLocation && mode.value === TransactionEditPageMode.Add
            && !geoLocationStatus.value && !transaction.value.geoLocation) {
            updateGeoLocation(false);
        }
    }

    function onPageBeforeOut(): void {
        if (submitted.value || pageTypeAndMode?.type !== TransactionEditPageType.Transaction || mode.value !== TransactionEditPageMode.Add || query['noTransactionDraft'] === 'true' || addByTemplateId.value || duplicateFromId.value) {
            return;
        }

        const initAmount: number | undefined = query['amount'] ? parseInt(query['amount']) : undefined;

        if (settingsStore.appSettings.autoSaveTransactionDraft === 'confirmation') {
            if (transactionsStore.isTransactionDraftModified(transaction.value, initAmount, query['categoryId'], query['accountId'], query['tagIds'], firstVisibleAccountId.value)) {
                showConfirm('Do you want to save this transaction draft?', () => {
                    transactionsStore.saveTransactionDraft(transaction.value, initAmount, query['categoryId'], query['accountId'], query['tagIds'], firstVisibleAccountId.value);
                }, () => {
                    transactionsStore.clearTransactionDraft();
                });
            } else {
                transactionsStore.clearTransactionDraft();
            }
        } else if (settingsStore.appSettings.autoSaveTransactionDraft === 'enabled') {
            transactionsStore.saveTransactionDraft(transaction.value, initAmount, query['categoryId'], query['accountId'], query['tagIds'], firstVisibleAccountId.value);
        }
    }

    init();

    return { onPageAfterIn, onPageBeforeOut };
}
