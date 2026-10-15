import { TRANSACTION_COMMENT_HINT_MIN_LENGTH, TRANSACTION_MAX_COMMENT_LENGTH } from '@/consts/transaction.ts';
import { TransactionQuickAddButtonActionType, TransactionType } from '@/core/transaction.ts';
import {
    getCurrentUnixTime,
    getTimezoneOffsetMinutes,
    getUtcOffsetByUtcOffsetMinutes
} from '@/lib/datetime.ts';
import {
    getExchangedAmountByRate,
    parseBigDecimal
} from '@/lib/numeral.ts';
import { Account } from '@/models/account.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import { computed } from 'vue';
import type { useTransactionEditPageBaseState } from './state.ts';
import { GeoLocationStatus, TransactionEditPageMode, TransactionEditPageType } from './types.ts';

interface Context {
    type: TransactionEditPageType;
    mode: ReturnType<typeof useTransactionEditPageBaseState>['mode'];
    isNormalTransactionTemplate: ReturnType<typeof useTransactionEditPageBaseState>['isNormalTransactionTemplate'];
    transaction: ReturnType<typeof useTransactionEditPageBaseState>['transaction'];
    isScheduledTransactionTemplate: ReturnType<typeof useTransactionEditPageBaseState>['isScheduledTransactionTemplate'];
    settingsStore: ReturnType<typeof useTransactionEditPageBaseState>['settingsStore'];
    allAccountsMap: ReturnType<typeof useTransactionEditPageBaseState>['allAccountsMap'];
    tt: ReturnType<typeof useTransactionEditPageBaseState>['tt'];
    defaultCurrency: ReturnType<typeof useTransactionEditPageBaseState>['defaultCurrency'];
    exchangeRatesStore: ReturnType<typeof useTransactionEditPageBaseState>['exchangeRatesStore'];
    getDisplayAmount: ReturnType<typeof useTransactionEditPageBaseState>['getDisplayAmount'];
    getAdaptiveAmountRate: ReturnType<typeof useTransactionEditPageBaseState>['getAdaptiveAmountRate'];
    allAccounts: ReturnType<typeof useTransactionEditPageBaseState>['allAccounts'];
    numeralSystem: ReturnType<typeof useTransactionEditPageBaseState>['numeralSystem'];
    getTimezoneDifferenceDisplayText: ReturnType<typeof useTransactionEditPageBaseState>['getTimezoneDifferenceDisplayText'];
    geoLocationStatus: ReturnType<typeof useTransactionEditPageBaseState>['geoLocationStatus'];
    formatNumberToLocalizedNumerals: ReturnType<typeof useTransactionEditPageBaseState>['formatNumberToLocalizedNumerals'];
}

// Owns presentation for the TransactionEditPageBase store.
export function useTransactionEditPageBasePresentation(context: Context) {
    const {
        type, mode, isNormalTransactionTemplate, transaction, isScheduledTransactionTemplate, settingsStore, allAccountsMap, tt,
        defaultCurrency, exchangeRatesStore, getDisplayAmount, getAdaptiveAmountRate, allAccounts, numeralSystem,
        getTimezoneDifferenceDisplayText, geoLocationStatus, formatNumberToLocalizedNumerals,
    } = context;

    const title = computed<string>(() => {
        if (type === TransactionEditPageType.Transaction) {
            if (mode.value === TransactionEditPageMode.Add) {
                return 'Add Transaction';
            } else if (mode.value === TransactionEditPageMode.Edit) {
                return 'Edit Transaction';
            } else {
                return 'Transaction Detail';
            }
        } else if (isNormalTransactionTemplate(transaction.value)) {
            if (mode.value === TransactionEditPageMode.Add) {
                return 'Add Transaction Template';
            } else if (mode.value === TransactionEditPageMode.Edit) {
                return 'Edit Transaction Template';
            }
        } else if (isScheduledTransactionTemplate(transaction.value)) {
            if (mode.value === TransactionEditPageMode.Add) {
                return 'Add Scheduled Transaction';
            } else if (mode.value === TransactionEditPageMode.Edit) {
                return 'Edit Scheduled Transaction';
            }
        }

        return '';
    });

    const saveButtonTitle = computed<string>(() => {
        if (mode.value === TransactionEditPageMode.Add) {
            return 'Add';
        } else {
            return 'Save';
        }
    });

    const quickSaveButtonTitle = computed<string>(() => {
        if (mode.value === TransactionEditPageMode.Add) {
            const quickAddActionType = TransactionQuickAddButtonActionType.valueOf(settingsStore.appSettings.quickAddButtonActionInMobileTransactionEditPage);

            if (quickAddActionType && quickAddActionType.type !== TransactionQuickAddButtonActionType.OpenMenu.type) {
                return quickAddActionType.name;
            } else {
                return 'Add';
            }
        } else {
            return 'Save';
        }
    });

    const cancelButtonTitle = computed<string>(() => {
        if (mode.value === TransactionEditPageMode.View) {
            return 'Close';
        } else {
            return 'Cancel';
        }
    });

    const sourceAmountName = computed<string>(() => {
        if (transaction.value.type === TransactionType.Expense) {
            return 'Expense Amount';
        } else if (transaction.value.type === TransactionType.Income) {
            return 'Income Amount';
        } else if (transaction.value.type === TransactionType.Transfer) {
            return 'Transfer Out Amount';
        } else {
            return 'Amount';
        }
    });

    const sourceAmountTitle = computed<string>(() => {
        const sourceAccount = allAccountsMap.value[transaction.value.sourceAccountId];
        const amountName = tt(sourceAmountName.value);

        if (!sourceAccount || sourceAccount.currency === defaultCurrency.value || !transaction.value.sourceAmount || transaction.value.hideAmount) {
            return amountName;
        }

        const fromExchangeRate = exchangeRatesStore.latestExchangeRateMap[sourceAccount.currency];
        const toExchangeRate = exchangeRatesStore.latestExchangeRateMap[defaultCurrency.value];

        if (!fromExchangeRate || !fromExchangeRate.rate || !toExchangeRate || !toExchangeRate.rate) {
            return amountName;
        }

        let amountInDefaultCurrency = getExchangedAmountByRate(parseBigDecimal(transaction.value.sourceAmount), fromExchangeRate.rate, toExchangeRate.rate);

        if (!amountInDefaultCurrency) {
            return amountName;
        }

        amountInDefaultCurrency = amountInDefaultCurrency.truncate();

        const displayAmountInDefaultCurrency = getDisplayAmount(amountInDefaultCurrency, transaction.value.hideAmount, defaultCurrency.value);
        return amountName + ` (${displayAmountInDefaultCurrency})`;
    });

    const sourceAccountTitle = computed<string>(() => {
        if (transaction.value.type === TransactionType.Expense || transaction.value.type === TransactionType.Income) {
            return 'Account';
        } else if (transaction.value.type === TransactionType.Transfer) {
            return 'Source Account';
        } else {
            return 'Account';
        }
    });

    const transferInAmountTitle = computed<string>(() => {
        const sourceAccount = allAccountsMap.value[transaction.value.sourceAccountId];
        const destinationAccount = allAccountsMap.value[transaction.value.destinationAccountId];

        if (!sourceAccount || !destinationAccount || sourceAccount.currency === destinationAccount.currency) {
            return tt('Transfer In Amount');
        }

        const fromExchangeRate = exchangeRatesStore.latestExchangeRateMap[sourceAccount.currency];
        const toExchangeRate = exchangeRatesStore.latestExchangeRateMap[destinationAccount.currency];
        const amountRate = getAdaptiveAmountRate(transaction.value.sourceAmount, transaction.value.destinationAmount, fromExchangeRate, toExchangeRate);

        if (!amountRate) {
            return tt('Transfer In Amount');
        }

        return tt('Transfer In Amount') + ` (${amountRate})`;
    });

    const sourceAccountName = computed<string>(() => {
        if (transaction.value.sourceAccountId) {
            return Account.findAccountNameById(allAccounts.value, transaction.value.sourceAccountId) || '';
        } else {
            return tt('None');
        }
    });

    const destinationAccountName = computed<string>(() => {
        if (transaction.value.destinationAccountId) {
            return Account.findAccountNameById(allAccounts.value, transaction.value.destinationAccountId) || '';
        } else {
            return tt('None');
        }
    });

    const sourceAccountCurrency = computed<string>(() => {
        const sourceAccount = allAccountsMap.value[transaction.value.sourceAccountId];

        if (sourceAccount) {
            return sourceAccount.currency;
        }

        return defaultCurrency.value;
    });

    const destinationAccountCurrency = computed<string>(() => {
        const destinationAccount = allAccountsMap.value[transaction.value.destinationAccountId];

        if (destinationAccount) {
            return destinationAccount.currency;
        }

        return defaultCurrency.value;
    });

    const transactionDisplayTimezone = computed<string>(() => {
        let utcOffsetMinutes: number = transaction.value.utcOffset;

        if (isScheduledTransactionTemplate(transaction.value) && transaction.value.timeZone) {
            utcOffsetMinutes = getTimezoneOffsetMinutes(getCurrentUnixTime(), transaction.value.timeZone);
        }

        const utcOffset = numeralSystem.value.replaceWesternArabicDigitsToLocalizedDigits(getUtcOffsetByUtcOffsetMinutes(utcOffsetMinutes));
        return `UTC${utcOffset}`;
    });

    const transactionTimezoneTimeDifference = computed<string>(() => {
        return getTimezoneDifferenceDisplayText(transaction.value.time, transaction.value.utcOffset);
    });

    const geoLocationStatusInfo = computed<string>(() => {
        if (geoLocationStatus.value === GeoLocationStatus.Success) {
            return '';
        } else if (geoLocationStatus.value === GeoLocationStatus.Getting) {
            return tt('Getting Location...');
        } else {
            return tt('No Location');
        }
    });

    const transactionDescriptionTitle = computed<string>(() => {
        if (!transaction.value.comment || transaction.value.comment.length < TRANSACTION_COMMENT_HINT_MIN_LENGTH) {
            return tt('Description');
        }

        if (transaction.value.comment.length > TRANSACTION_MAX_COMMENT_LENGTH) {
            return tt('Description') + ` (${tt('format.misc.charactersOverLimit', {
                count: formatNumberToLocalizedNumerals(transaction.value.comment.length - TRANSACTION_MAX_COMMENT_LENGTH)
            })})`;
        } else {
            return tt('Description') + ` (${tt('format.misc.charactersRemaining', {
                count: formatNumberToLocalizedNumerals(TRANSACTION_MAX_COMMENT_LENGTH - transaction.value.comment.length)
            })})`;
        }
    });

    const inputEmptyProblemMessage = computed<string | null>(() => {
        if (transaction.value.type === TransactionType.Expense) {
            if (!transaction.value.expenseCategoryId || transaction.value.expenseCategoryId === '') {
                return 'Transaction category cannot be blank';
            }

            if (!transaction.value.sourceAccountId || transaction.value.sourceAccountId === '') {
                return 'Transaction account cannot be blank';
            }
        } else if (transaction.value.type === TransactionType.Income) {
            if (!transaction.value.incomeCategoryId || transaction.value.incomeCategoryId === '') {
                return 'Transaction category cannot be blank';
            }

            if (!transaction.value.sourceAccountId || transaction.value.sourceAccountId === '') {
                return 'Transaction account cannot be blank';
            }
        } else if (transaction.value.type === TransactionType.Transfer) {
            if (!transaction.value.transferCategoryId || transaction.value.transferCategoryId === '') {
                return 'Transaction category cannot be blank';
            }

            if (!transaction.value.sourceAccountId || transaction.value.sourceAccountId === '') {
                return 'Source account cannot be blank';
            }

            if (!transaction.value.destinationAccountId || transaction.value.destinationAccountId === '') {
                return 'Destination account cannot be blank';
            }
        }

        if (type === TransactionEditPageType.Template && transaction.value instanceof TransactionTemplate) {
            if (!transaction.value.name) {
                return 'Template name cannot be blank';
            }
        }

        return null;
    });

    const inputIsEmpty = computed<boolean>(() => {
        return !!inputEmptyProblemMessage.value;
    });

    return {
        title, saveButtonTitle, quickSaveButtonTitle, cancelButtonTitle, sourceAmountName, sourceAmountTitle, sourceAccountTitle,
        transferInAmountTitle, sourceAccountName, destinationAccountName, sourceAccountCurrency, destinationAccountCurrency,
        transactionDisplayTimezone, transactionTimezoneTimeDifference, geoLocationStatusInfo, transactionDescriptionTitle,
        inputEmptyProblemMessage, inputIsEmpty,
    };
}
