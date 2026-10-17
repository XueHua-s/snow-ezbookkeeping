import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { AccountCategory, AccountType } from '@/core/account.ts';
import { DateRange, DateRangeScene, type LocalizedDateRange, type TimeRangeAndDateType } from '@/core/datetime.ts';
import { DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_DESKTOP } from '@/core/statistics.ts';
import { isNumber } from '@/lib/common.ts';
import {
    getCurrentUnixTime,
    getDateRangeByBillingCycleDateType,
    getDateRangeByDateType,
    getDateRangeByLastReconciledTimeRangeDateType
} from '@/lib/datetime.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { Account } from '@/models/account.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { useAccountListPageBase } from '@/views/base/accounts/AccountListPageBase.ts';
import ClearAllTransactionsDialog from '@/views/desktop/accounts/list/dialogs/ClearAllTransactionsDialog.vue';
import EditDialog from '@/views/desktop/accounts/list/dialogs/EditDialog.vue';
import MoveAllTransactionsDialog from '@/views/desktop/accounts/list/dialogs/MoveAllTransactionsDialog.vue';
import ReconciliationStatementDialog from '@/views/desktop/accounts/list/dialogs/ReconciliationStatementDialog.vue';
import { computed, ref, useTemplateRef, watch } from 'vue';
import { useDisplay } from 'vuetify';

export type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;

export type SnackBarType = InstanceType<typeof SnackBar>;

export type EditDialogType = InstanceType<typeof EditDialog>;

export type ReconciliationStatementDialogType = InstanceType<typeof ReconciliationStatementDialog>;

export type MoveAllTransactionsDialogType = InstanceType<typeof MoveAllTransactionsDialog>;

export type ClearAllTransactionsDialogType = InstanceType<typeof ClearAllTransactionsDialog>;

// Coordinates ListPage state and effects; called once from the component setup.
export function useListPage() {
    const { lgAndUp } = useDisplay();

    const { tt, getAllDateRanges, getCurrencyName, joinMultiText } = useI18n();

    const {
        loading,
        showHidden,
        displayOrderModified,
        showAccountBalance,
        customAccountCategoryOrder,
        defaultAccountCategory,
        firstDayOfWeek,
        fiscalYearStart,
        useLastReconciledTime,
        allAccounts,
        allCategorizedAccountsMap,
        allAccountCount,
        totalAssets,
        canShowAvailableCredit,
        accountCategoryTotalBalance,
        accountBalance,
        accountAvailableCredit
    } = useAccountListPageBase();

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const editDialog = useTemplateRef<EditDialogType>('editDialog');

    const reconciliationStatementDialog = useTemplateRef<ReconciliationStatementDialogType>('reconciliationStatementDialog');

    const moveAllTransactionsDialog = useTemplateRef<MoveAllTransactionsDialogType>('moveAllTransactionsDialog');

    const clearAllTransactionsDialog = useTemplateRef<ClearAllTransactionsDialogType>('clearAllTransactionsDialog');

    const activeAccountCategoryType = ref<number>(defaultAccountCategory.value.type);

    const activeTab = ref<string>('accountPage');

    const activeSubAccount = ref<Record<string, string>>({});

    const accountToShowReconciliationStatement = ref<Account | null>(null);

    const alwaysShowNav = ref<boolean>(lgAndUp.value);

    const showNav = ref<boolean>(lgAndUp.value);

    const showAccountsIncludedInTotalDialog = ref<boolean>(false);

    const showCustomDateRangeDialog = ref<boolean>(false);

    const hideAccountCategoriesWithoutAccounts = computed<boolean>(() => settingsStore.appSettings.hideCategoriesWithoutAccounts);

    const hasAnyVisibleAccount = computed<boolean>(() => accountsStore.allVisibleAccountsCount > 0);

    const activeAccountCategory = computed<AccountCategory | undefined>(() => AccountCategory.valueOf(activeAccountCategoryType.value));

    const activeAccountCategoryTotalBalance = computed<string>(() => accountCategoryTotalBalance(activeAccountCategory.value, false));

    const activeCreditCardTotalAvailabeCredit = computed<string>(() => accountCategoryTotalBalance(activeAccountCategory.value, true));

    const activeAccountCategoryVisibleAccountCount = computed<number>(() => {
        if (!activeAccountCategory.value) {
            return 0;
        }

        const categorizedAccounts = allCategorizedAccountsMap.value[activeAccountCategory.value.type];

        if (!categorizedAccounts || !categorizedAccounts.accounts || !categorizedAccounts.accounts.length) {
            return 0;
        }

        if (showHidden.value) {
            return categorizedAccounts.accounts.length;
        }

        let visibleCount = 0;

        for (const account of categorizedAccounts.accounts) {
            if (!account.hidden) {
                visibleCount++;
            }
        }

        return visibleCount;
    });

    function reload(force: boolean): void {
        loading.value = true;

        accountsStore.loadAllAccounts({
            force: force
        }).then(() => {
            loading.value = false;
            displayOrderModified.value = false;

            if (allAccounts.value) {
                for (const account of allAccounts.value) {
                    if (account.type === AccountType.MultiSubAccounts.type && !activeSubAccount.value[account.id]) {
                        activeSubAccount.value[account.id] = '';
                    }
                }
            }

            if (force) {
                snackbar.value?.showMessage('Account list has been updated');
            }
        }).catch(error => {
            loading.value = false;

            if (error && error.isUpToDate) {
                displayOrderModified.value = false;
            }

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function hasAccount(accountCategory: AccountCategory): boolean {
        return accountsStore.hasAccount(accountCategory, !showHidden.value);
    }

    function accountCurrency(account: Account): string | null {
        if (account.type === AccountType.SingleAccount.type) {
            return getCurrencyName(account.currency);
        } else if (account.type === AccountType.MultiSubAccounts.type) {
            const subAccountCurrencies = account.getSubAccountCurrencies(showHidden.value, activeSubAccount.value[account.id])
                .map(currencyCode => getCurrencyName(currencyCode));
            return joinMultiText(subAccountCurrencies);
        } else {
            return null;
        }
    }

    function accountReconciliationStatementDateRanges(account: Account): LocalizedDateRange[] {
        return getAllDateRanges(DateRangeScene.Normal, {
            includeCustom: true,
            includeBillingCycle: !!accountsStore.getAccountStatementDate(account.id),
            includeLastReconciledTimeRange: userStore.currentUserUseLastReconciledTime && !!account.lastReconciledTime
        });
    }

    function add(): void {
        editDialog.value?.open({
            category: activeAccountCategoryType.value
        }).then(result => {
            if (result && result.message) {
                snackbar.value?.showMessage(result.message);
            }
        }).catch(error => {
            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }

    function edit(account: Account): void {
        editDialog.value?.open({
            id: account.id,
            currentAccount: account
        }).then(result => {
            if (result && result.message) {
                snackbar.value?.showMessage(result.message);
            }

            if (accountsStore.accountListStateInvalid && !loading.value) {
                reload(false);
            }
        }).catch(error => {
            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }

    function showReconciliationStatementDialog(account: Account, dateRangeType?: number): void {
        if (!isNumber(dateRangeType)) {
            const defualtDateRange = DateRange.valueOf(settingsStore.appSettings.reconciliationStatementButtonDefaultDateRangeTypeInDesktop);

            if (!defualtDateRange) {
                dateRangeType = DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_DESKTOP.type;
            } else if (defualtDateRange.isBillingCycle && !accountsStore.getAccountStatementDate(account.id)) {
                dateRangeType = DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_DESKTOP.type;
            } else if (defualtDateRange.isLastReconciledTimeRange && (!userStore.currentUserUseLastReconciledTime || !account.lastReconciledTime)) {
                dateRangeType = DEFAULT_RECONCILIATION_STATEMENT_DATE_RANGE_IN_DESKTOP.type;
            } else {
                dateRangeType = defualtDateRange.type;
            }
        }

        if (!isNumber(dateRangeType) || dateRangeType === DateRange.Custom.type) {
            accountToShowReconciliationStatement.value = account;
            showCustomDateRangeDialog.value = true;
            return;
        }

        let dateRange: TimeRangeAndDateType | null = null;

        if (DateRange.isBillingCycle(dateRangeType)) {
            dateRange = getDateRangeByBillingCycleDateType(dateRangeType, firstDayOfWeek.value, fiscalYearStart.value, accountsStore.getAccountStatementDate(account.id));
        } else if (DateRange.isLastReconciledTimeRange(dateRangeType)) {
            dateRange = getDateRangeByLastReconciledTimeRangeDateType(dateRangeType, account.lastReconciledTime);
        } else {
            dateRange = getDateRangeByDateType(dateRangeType, firstDayOfWeek.value, fiscalYearStart.value);
        }

        if (!dateRange) {
            return;
        }

        reconciliationStatementDialog.value?.open({
            accountId: account.id,
            startTime: dateRange.minTime,
            endTime: dateRange.maxTime
        });
    }

    function updateLastReconciledTime(account: Account): void {
        confirmDialog.value?.open('Are you sure you want to update the last reconciled time of this account to the current time?').then(() => {
            loading.value = true;

            accountsStore.updateAccountLastReconciledTime(account.id, getCurrentUnixTime()).then(() => {
                loading.value = false;
                snackbar.value?.showMessage('Last reconciled time have been updated');

                if (accountsStore.accountListStateInvalid && !loading.value) {
                    reload(false);
                }

            }).catch(error => {
                loading.value = false;

                if (error) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    function moveAllTransactions(account: Account): void {
        moveAllTransactionsDialog.value?.open(account).then(() => {
            snackbar.value?.showMessage('All transactions in this account have been moved.');

            if (accountsStore.accountListStateInvalid && !loading.value) {
                reload(false);
            }
        });
    }

    function clearAllTransactions(account: Account): void {
        clearAllTransactionsDialog.value?.open(account).then(() => {
            snackbar.value?.showMessage('All transactions in this account have been cleared');

            if (accountsStore.accountListStateInvalid && !loading.value) {
                reload(false);
            }
        });
    }

    function hide(account: Account, targetAccount: Account, hidden: boolean): void {
        loading.value = true;

        accountsStore.hideAccount({
            account: targetAccount,
            hidden: hidden
        }).then(() => {
            if (hidden && !showHidden.value && activeSubAccount.value[account.id]) {
                activeSubAccount.value[account.id] = '';
            }

            loading.value = false;
        }).catch(error => {
            loading.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function remove(account: Account): void {
        if (activeSubAccount.value[account.id]) {
            const subAccount: Account | null = account.getSubAccount(activeSubAccount.value[account.id]);

            if (!subAccount) {
                snackbar.value?.showMessage('Unable to delete this sub-account');
                return;
            }

            confirmDialog.value?.open('Are you sure you want to delete this sub-account?').then(() => {
                loading.value = true;

                accountsStore.deleteSubAccount({
                    subAccount: subAccount
                }).then(() => {
                    activeSubAccount.value[account.id] = '';
                    loading.value = false;
                }).catch(error => {
                    loading.value = false;

                    if (!error.processed) {
                        snackbar.value?.showError(error);
                    }
                });
            });
        } else {
            confirmDialog.value?.open('Are you sure you want to delete this account?').then(() => {
                loading.value = true;

                accountsStore.deleteAccount({
                    account: account
                }).then(() => {
                    loading.value = false;
                }).catch(error => {
                    loading.value = false;

                    if (!error.processed) {
                        snackbar.value?.showError(error);
                    }
                });
            });
        }
    }

    function saveSortResult(): void {
        if (!displayOrderModified.value) {
            return;
        }

        loading.value = true;

        accountsStore.updateAccountDisplayOrders().then(() => {
            loading.value = false;
            displayOrderModified.value = false;
        }).catch(error => {
            loading.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function onMove(event: { moved: { element: { id: string }, oldIndex: number, newIndex: number } }): void {
        if (!event || !event.moved) {
            return;
        }

        const moveEvent = event.moved;

        if (!moveEvent.element || !moveEvent.element.id) {
            snackbar.value?.showMessage('Unable to move account');
            return;
        }

        accountsStore.changeAccountDisplayOrder({
            accountId: moveEvent.element.id,
            from: moveEvent.oldIndex,
            to: moveEvent.newIndex,
            updateListOrder: false,
            updateGlobalListOrder: true
        }).then(() => {
            displayOrderModified.value = true;
        }).catch(error => {
            snackbar.value?.showError(error);
        });
    }

    function onCustomDateRangeChanged(minUnixTime: number, maxUnixTime: number): void {
        if (!accountToShowReconciliationStatement.value) {
            snackbar.value?.showMessage('An error occurred');
            return;
        }

        showCustomDateRangeDialog.value = false;

        reconciliationStatementDialog.value?.open({
            accountId: accountToShowReconciliationStatement.value.id,
            startTime: minUnixTime,
            endTime: maxUnixTime
        });

        accountToShowReconciliationStatement.value = null;
    }

    function onShowDateRangeError(message: string): void {
        snackbar.value?.showError(message);
    }

    watch(lgAndUp, (newValue) => {
        alwaysShowNav.value = newValue;

        if (!showNav.value) {
            showNav.value = newValue;
        }
    });

    reload(false);

    return {
        tt, loading, showHidden, displayOrderModified, showAccountBalance, customAccountCategoryOrder, useLastReconciledTime,
        allCategorizedAccountsMap, allAccountCount, totalAssets, canShowAvailableCredit,
        accountCategoryTotalBalance, accountBalance, accountAvailableCredit, confirmDialog, snackbar, editDialog,
        reconciliationStatementDialog, moveAllTransactionsDialog, clearAllTransactionsDialog, activeAccountCategoryType,
        activeTab, activeSubAccount, showAccountsIncludedInTotalDialog, showCustomDateRangeDialog,
        hideAccountCategoriesWithoutAccounts, hasAnyVisibleAccount, activeAccountCategory, activeAccountCategoryTotalBalance,
        activeCreditCardTotalAvailabeCredit, activeAccountCategoryVisibleAccountCount, reload, hasAccount, accountCurrency,
        accountReconciliationStatementDateRanges, add, edit, showReconciliationStatementDialog, updateLastReconciledTime,
        moveAllTransactions, clearAllTransactions, hide, remove, saveSortResult, onMove, onCustomDateRangeChanged,
        onShowDateRangeError,
    };
}
