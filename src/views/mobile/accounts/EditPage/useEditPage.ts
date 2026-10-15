import type { SetupProps } from '@/lib/ui/props.ts';
import { itemAndIndex } from '@/core/base.ts';
import type { LocalizedCurrencyInfo } from '@/core/currency.ts';
import { isDefined } from '@/lib/common.ts';
import {
    getTimezoneOffsetMinutes,
    parseDateTimeFromUnixTimeWithTimezoneOffset
} from '@/lib/datetime.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { hideLoading, showLoading, useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { Account } from '@/models/account.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useAccountEditPageBase } from '@/views/base/accounts/AccountEditPageBase.ts';
import type { Router } from 'framework7/types';
import { computed, ref, watch } from 'vue';

export interface AccountContext {
    showIconSelectionSheet: boolean;
    showColorSelectionSheet: boolean;
    showCurrencyPopup: boolean;
    showCreditCardStatementDatePopup: boolean;
    showBalanceSheet: boolean;
    showBalanceDateTimeSheet: boolean;
    showLastReconciledTimeSheet: boolean;
    balanceDateTimeSheetMode: string;
    lastReconciledDateTimeSheetMode: string;
}

// Coordinates EditPage state and effects; called once from the component setup.
export function useEditPage(props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>) {
    const {
        tt,
        getAllCurrencies,
        getCurrencyName,
        formatDateTimeToLongDate,
        formatDateTimeToLongTime,
        formatAmountToLocalizedNumeralsWithCurrency
    } = useI18n();

    const { showAlert, showToast, routeBackOnError } = useI18nUIComponents();

    const {
        editAccountId,
        clientSessionId,
        loading,
        submitting,
        account,
        subAccounts,
        useLastReconciledTime,
        title,
        inputEmptyProblemMessage,
        inputIsEmpty,
        allAccountCategories,
        allAccountTypes,
        allAvailableMonthDays,
        getDefaultTimezoneOffsetMinutes,
        getAccountCreditCardStatementDate,
        getAccountCreditCardCreditLimitDisplayValue,
        updateAccountBalanceTime,
        updateAccountLastReconciledTime,
        isNewAccount,
        addSubAccount,
        setAccount
    } = useAccountEditPageBase();

    const accountsStore = useAccountsStore();

    const DEFAULT_ACCOUNT_CONTEXT: AccountContext = {
        showIconSelectionSheet: false,
        showColorSelectionSheet: false,
        showCurrencyPopup: false,
        showCreditCardStatementDatePopup: false,
        showBalanceSheet: false,
        showBalanceDateTimeSheet: false,
        showLastReconciledTimeSheet: false,
        balanceDateTimeSheetMode: 'time',
        lastReconciledDateTimeSheetMode: 'time'
    };

    const accountContext = ref<AccountContext>(Object.assign({}, DEFAULT_ACCOUNT_CONTEXT));

    const subAccountContexts = ref<AccountContext[]>([]);

    const subAccountToDelete = ref<Account | null>(null);

    const loadingError = ref<unknown | null>(null);

    const showAccountCategorySheet = ref<boolean>(false);

    const showAccountTypeSheet = ref<boolean>(false);

    const showMoreActionSheet = ref<boolean>(false);

    const showDeleteActionSheet = ref<boolean>(false);

    const showCreditCardLimitSheet = ref<boolean>(false);

    const allCurrencies = computed<LocalizedCurrencyInfo[]>(() => getAllCurrencies());

    const allCurrenciesWithNotSet = computed<LocalizedCurrencyInfo[]>(() => getAllCurrencies(true));

    function formatAccountDisplayBalance(selectedAccount: Account): string {
        const balance = parseBigDecimal(selectedAccount.balance);
        return formatAmountToLocalizedNumeralsWithCurrency(account.value.isLiability ? balance.negate() : balance, selectedAccount.currency);
    }

    function formatDate(unixTime?: number): string {
        if (!isDefined(unixTime)) {
            return '';
        }

        const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(unixTime, getTimezoneOffsetMinutes(unixTime));
        return formatDateTimeToLongDate(dateTime);
    }

    function formatTime(unixTime?: number): string {
        if (!isDefined(unixTime)) {
            return '';
        }

        const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(unixTime, getTimezoneOffsetMinutes(unixTime));
        return formatDateTimeToLongTime(dateTime);
    }

    function init(): void {
        const query = props.f7route.query;
        clientSessionId.value = generateRandomUUID();

        if (query['id']) {
            loading.value = true;

            editAccountId.value = query['id'];

            accountsStore.getAccount({
                accountId: editAccountId.value
            }).then(response => {
                setAccount(response);
                subAccountContexts.value = [];

                for (let i = 0; i < subAccounts.value.length; i++) {
                    subAccountContexts.value.push(Object.assign({}, DEFAULT_ACCOUNT_CONTEXT));
                }

                loading.value = false;
            }).catch(error => {
                if (error.processed) {
                    loading.value = false;
                } else {
                    loadingError.value = error;
                    showToast(error.message || error);
                }
            });
        } else {
            loading.value = false;
        }
    }

    function save(): void {
        const router = props.f7router;
        const problemMessage = inputEmptyProblemMessage.value;

        if (problemMessage) {
            showAlert(problemMessage);
            return;
        }

        submitting.value = true;
        showLoading(() => submitting.value);

        accountsStore.saveAccount({
            account: account.value,
            subAccounts: subAccounts.value,
            isEdit: !!editAccountId.value,
            clientSessionId: clientSessionId.value
        }).then(() => {
            submitting.value = false;
            hideLoading();

            if (!editAccountId.value) {
                showToast('You have added a new account');
            } else {
                showToast('You have saved this account');
            }

            router.back();
        }).catch(error => {
            submitting.value = false;
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function addSubAccountAndContext(): void {
        if (addSubAccount()) {
            subAccountContexts.value.push(Object.assign({}, DEFAULT_ACCOUNT_CONTEXT));
        }
    }

    function removeSubAccount(currentSubAccount: Account | null, confirm: boolean): void {
        if (!currentSubAccount) {
            showAlert('An error occurred');
            return;
        }

        if (!confirm) {
            subAccountToDelete.value = currentSubAccount;
            showDeleteActionSheet.value = true;
            return;
        }

        showDeleteActionSheet.value = false;
        subAccountToDelete.value = null;

        for (const [subAccount, index] of itemAndIndex(subAccounts.value)) {
            if (subAccount === currentSubAccount) {
                subAccounts.value.splice(index, 1);
                subAccountContexts.value.splice(index, 1);
            }
        }
    }

    function showBalanceDateTimeDialog(accountContext: AccountContext, sheetMode: string): void {
        accountContext.balanceDateTimeSheetMode = sheetMode;
        accountContext.showBalanceDateTimeSheet = true;
    }

    function showLastReconciledDateTimeDialog(accountContext: AccountContext, sheetMode: string): void {
        accountContext.lastReconciledDateTimeSheetMode = sheetMode;
        accountContext.showLastReconciledTimeSheet = true;
    }

    function onPageAfterIn(): void {
        routeBackOnError(props.f7router, loadingError);
    }

    watch(() => account.value.type, () => {
        if (subAccounts.value.length < 1) {
            addSubAccountAndContext();
        }
    });

    init();

    return {
        tt, getCurrencyName, editAccountId, loading, submitting, account, subAccounts, useLastReconciledTime, title, inputIsEmpty,
        allAccountCategories, allAccountTypes, allAvailableMonthDays, getDefaultTimezoneOffsetMinutes,
        getAccountCreditCardStatementDate, getAccountCreditCardCreditLimitDisplayValue, updateAccountBalanceTime,
        updateAccountLastReconciledTime, isNewAccount, accountContext, subAccountContexts, subAccountToDelete,
        showAccountCategorySheet, showAccountTypeSheet, showMoreActionSheet, showDeleteActionSheet, showCreditCardLimitSheet,
        allCurrencies, allCurrenciesWithNotSet, formatAccountDisplayBalance, formatDate, formatTime, init, save,
        addSubAccountAndContext, removeSubAccount, showBalanceDateTimeDialog, showLastReconciledDateTimeDialog, onPageAfterIn,
    };
}
