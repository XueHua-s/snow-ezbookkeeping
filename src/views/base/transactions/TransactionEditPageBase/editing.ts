import { TransactionType } from '@/core/transaction.ts';
import {
    isDefined
} from '@/lib/common.ts';
import {
    getBrowserTimezoneName
} from '@/lib/datetime.ts';
import {
    type SetTransactionOptions,
    setTransactionModelByTransaction
} from '@/lib/transaction.ts';
import type { RecognizedTransactionResponse } from '@/models/large_language_model.ts';
import { Transaction } from '@/models/transaction.ts';
import type { TransactionPictureInfoBasicResponse } from '@/models/transaction_picture_info.ts';
import { watch } from 'vue';
import type { useTransactionEditPageBaseState } from './state.ts';
import { AfterSaveAction, TransactionEditPageMode } from './types.ts';

interface Context {
    transaction: ReturnType<typeof useTransactionEditPageBaseState>['transaction'];
    allCategories: ReturnType<typeof useTransactionEditPageBaseState>['allCategories'];
    allCategoriesMap: ReturnType<typeof useTransactionEditPageBaseState>['allCategoriesMap'];
    allVisibleAccounts: ReturnType<typeof useTransactionEditPageBaseState>['allVisibleAccounts'];
    allAccountsMap: ReturnType<typeof useTransactionEditPageBaseState>['allAccountsMap'];
    allTagsMap: ReturnType<typeof useTransactionEditPageBaseState>['allTagsMap'];
    defaultAccountId: ReturnType<typeof useTransactionEditPageBaseState>['defaultAccountId'];
    createNewTransactionModel: ReturnType<typeof useTransactionEditPageBaseState>['createNewTransactionModel'];
    transactionDefaultType: number | undefined;
    geoLocationStatus: ReturnType<typeof useTransactionEditPageBaseState>['geoLocationStatus'];
    allTimezones: ReturnType<typeof useTransactionEditPageBaseState>['allTimezones'];
    isScheduledTransactionTemplate: ReturnType<typeof useTransactionEditPageBaseState>['isScheduledTransactionTemplate'];
    transactionsStore: ReturnType<typeof useTransactionEditPageBaseState>['transactionsStore'];
    mode: ReturnType<typeof useTransactionEditPageBaseState>['mode'];
    loading: ReturnType<typeof useTransactionEditPageBaseState>['loading'];
}

// Owns editing for the TransactionEditPageBase store.
export function useTransactionEditPageBaseEditing(context: Context) {
    const {
        transaction, allCategories, allCategoriesMap, allVisibleAccounts, allAccountsMap, allTagsMap, defaultAccountId,
        createNewTransactionModel, transactionDefaultType, geoLocationStatus, allTimezones, isScheduledTransactionTemplate,
        transactionsStore, mode, loading,
    } = context;

    function setTransactionModel(newTransaction: Transaction | null, options: SetTransactionOptions | undefined, setContextData: boolean): void {
        setTransactionModelByTransaction(
            transaction.value,
            newTransaction,
            allCategories.value,
            allCategoriesMap.value,
            allVisibleAccounts.value,
            allAccountsMap.value,
            allTagsMap.value,
            defaultAccountId.value,
            {
                time: options?.time,
                type: options?.type,
                categoryId: options?.categoryId,
                accountId: options?.accountId,
                destinationAccountId: options?.destinationAccountId,
                amount: options?.amount,
                destinationAmount: options?.destinationAmount,
                tagIds: options?.tagIds,
                comment: options?.comment
            },
            setContextData
        );
    }

    function updateTransactionModelFromRecognizedResponse(response: RecognizedTransactionResponse): void {
        const options: SetTransactionOptions = {
            type: response.type,
            time: response.time,
            categoryId: response.categoryId,
            accountId: response.sourceAccountId,
            destinationAccountId: response.destinationAccountId,
            amount: response.sourceAmount,
            destinationAmount: response.destinationAmount,
            tagIds: response.tagIds ? response.tagIds.join(',') : undefined,
            comment: response.comment
        };

        setTransactionModel(null, options, true);
    }

    function updateTransactionModelByAfterSaveAction(afterSaveAction: AfterSaveAction, initOptions?: SetTransactionOptions): void {
        if (afterSaveAction === AfterSaveAction.StayWithNewTransaction) {
            transaction.value = createNewTransactionModel(transactionDefaultType);
            setTransactionModel(null, initOptions, true);
            geoLocationStatus.value = null;
        } else if (afterSaveAction === AfterSaveAction.StayWithCurrentTransaction) {
            transaction.value.clearPictures();
        }
    }

    function updateTransactionTime(newTime: number): void {
        transaction.value.time = newTime;
        updateTransactionTimezone(transaction.value.timeZone ?? '');
    }

    function updateTransactionTimezone(timezoneName: string | null): void {
        const oldUtcOffset = transaction.value.utcOffset;

        if (!timezoneName) {
            timezoneName = ''
        }

        for (const timezone of allTimezones.value) {
            if (timezone.name === timezoneName) {
                if (isScheduledTransactionTemplate(transaction.value)) {
                    transaction.value.timeZone = timezone.name || getBrowserTimezoneName();
                } else {
                    transaction.value.timeZone = timezone.name;
                }

                transaction.value.utcOffset = timezone.utcOffsetMinutes;
                break;
            }
        }

        transaction.value.time = transaction.value.time - (transaction.value.utcOffset - oldUtcOffset) * 60;
    }

    function swapTransactionData(swapAccount: boolean, swapAmount: boolean): void {
        if (swapAccount) {
            const oldSourceAccountId = transaction.value.sourceAccountId;
            transaction.value.sourceAccountId = transaction.value.destinationAccountId;
            transaction.value.destinationAccountId = oldSourceAccountId;
        }

        if (swapAmount) {
            const oldSourceAmount = transaction.value.sourceAmount;
            transaction.value.sourceAmount = transaction.value.destinationAmount;
            transaction.value.destinationAmount = oldSourceAmount;
        }
    }

    function getTransactionPictureUrl(pictureInfo?: TransactionPictureInfoBasicResponse | null): string | undefined {
        return transactionsStore.getTransactionPictureUrl(pictureInfo);
    }

    watch(() => transaction.value.sourceAmount, (newValue, oldValue) => {
        if (mode.value === TransactionEditPageMode.View || loading.value) {
            return;
        }

        transactionsStore.setTransactionSuitableDestinationAmount(transaction.value, oldValue, newValue);
    });

    watch(() => transaction.value.destinationAmount, (newValue) => {
        if (mode.value === TransactionEditPageMode.View || loading.value) {
            return;
        }

        if (transaction.value.type === TransactionType.Expense || transaction.value.type === TransactionType.Income) {
            transaction.value.sourceAmount = newValue;
        }
    });

    // Watch for account changes and recalculate destination amount for transfers
    watch(() => [transaction.value.sourceAccountId, transaction.value.destinationAccountId], ([newSourceAccountId, newDestinationAccountId], [oldSourceAccountId, oldDestinationAccountId]) => {
        if (mode.value === TransactionEditPageMode.View || loading.value) {
            return;
        }

        if (transaction.value.type !== TransactionType.Transfer) {
            return;
        }

        // Only recalculate if accounts actually changed (skip initial watch call)
        if (isDefined(oldSourceAccountId) && isDefined(oldDestinationAccountId)) {
            if (newSourceAccountId === oldSourceAccountId && newDestinationAccountId === oldDestinationAccountId) {
                return;
            }
        }

        transactionsStore.setTransactionSuitableDestinationAmount(transaction.value, transaction.value.sourceAmount, transaction.value.sourceAmount, oldSourceAccountId, oldDestinationAccountId);
    });

    return {
        setTransactionModel, updateTransactionModelFromRecognizedResponse, updateTransactionModelByAfterSaveAction,
        updateTransactionTime, updateTransactionTimezone, swapTransactionData, getTransactionPictureUrl,
    };
}
