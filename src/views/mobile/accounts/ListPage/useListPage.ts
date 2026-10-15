import type { SetupProps } from '@/lib/ui/props.ts';
import { AccountCategory, CreditCardAmountDisplayType } from '@/core/account.ts';
import { TextDirection } from '@/core/text.ts';
import { getCurrentUnixTime } from '@/lib/datetime.ts';
import { hideLoading, onSwipeoutDeleted, showLoading, useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { Account, AccountShowingIds } from '@/models/account.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useRootStore } from '@/stores/index.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useAccountListPageBase } from '@/views/base/accounts/AccountListPageBase.ts';
import type { Router } from 'framework7/types';
import { computed, ref } from 'vue';



// Coordinates ListPage state and effects; called once from the component setup.
export function useListPage(props: SetupProps<{
    f7router: Router.Router;
}>) {
    const { tt, getCurrentLanguageTextDirection } = useI18n();

    const { showAlert, showConfirm, showToast, routeBackOnError } = useI18nUIComponents();

    const {
        loading,
        showHidden,
        displayOrderModified,
        showAccountBalance,
        customAccountCategoryOrder,
        useLastReconciledTime,
        allCategorizedAccountsMap,
        allAccountCount,
        maxCategoryAccountCount,
        accountCategoryTotalBalance,
        accountBalance,
        accountBalanceOrAvailableCredit
    } = useAccountListPageBase();

    const rootStore = useRootStore();

    const settingsStore = useSettingsStore();

    const accountsStore = useAccountsStore();

    const loadingError = ref<unknown | null>(null);

    const sortable = ref<boolean>(false);

    const accountForMoreActionSheet = ref<Account | null>(null);

    const accountToDelete = ref<Account | null>(null);

    const accountToClearTransactions = ref<Account | null>(null);

    const currentPasswordForClearData = ref<string>('');

    const updatingLastReconciledTime = ref<boolean>(false);

    const clearingData = ref<boolean>(false);

    const showAccountMoreActionSheet = ref<boolean>(false);

    const showMoreActionSheet = ref<boolean>(false);

    const showDeleteActionSheet = ref<boolean>(false);

    const showInputPasswordSheetForClearAllTransactions = ref<boolean>(false);

    const displayOrderSaving = ref<boolean>(false);

    const textDirection = computed<TextDirection>(() => getCurrentLanguageTextDirection());

    const showAvailableCreditForCreditCard = ref<boolean>(settingsStore.appSettings.defaultCreditCardAmountDisplayTypeInMobile === CreditCardAmountDisplayType.AvailableCredit.type);

    const firstShowingIds = computed<AccountShowingIds>(() => accountsStore.getFirstShowingIds(showHidden.value));

    const lastShowingIds = computed<AccountShowingIds>(() => accountsStore.getLastShowingIds(showHidden.value));

    const hasAnyVisibleAccount = computed<boolean>(() => accountsStore.allVisibleAccountsCount > 0);

    const noAvailableAccount = computed<boolean>(() => {
        if (showHidden.value) {
            return accountsStore.allAvailableAccountsCount < 1;
        } else {
            return accountsStore.allVisibleAccountsCount < 1;
        }
    });

    function hasAccount(accountCategory: AccountCategory, visibleOnly: boolean): boolean {
        return accountsStore.hasAccount(accountCategory, visibleOnly);
    }

    function hasVisibleSubAccount(account: Account): boolean {
        return accountsStore.hasVisibleSubAccount(showHidden.value, account);
    }

    function getAccountDomId(account: Account): string {
        return 'account_' + account.id;
    }

    function parseAccountIdFromDomId(domId: string): string | null {
        if (!domId || domId.indexOf('account_') !== 0) {
            return null;
        }

        return domId.substring(8); // account_
    }

    function init(): void {
        loading.value = true;

        accountsStore.loadAllAccounts({
            force: false
        }).then(() => {
            loading.value = false;
        }).catch(error => {
            if (error.processed) {
                loading.value = false;
            } else {
                loadingError.value = error;
                showToast(error.message || error);
            }
        });
    }

    function reload(done?: () => void): void {
        if (sortable.value) {
            done?.();
            return;
        }

        const force = !!done;

        accountsStore.loadAllAccounts({
            force: force
        }).then(() => {
            done?.();

            if (force) {
                showToast('Account list has been updated');
            }
        }).catch(error => {
            done?.();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function edit(account: Account): void {
        props.f7router.navigate('/account/edit?id=' + account.id);
    }

    function showMoreActionSheetForAccount(account: Account): void {
        accountForMoreActionSheet.value = account;
        showAccountMoreActionSheet.value = true;
    }

    function showReconciliationStatement(account: Account | null): void {
        if (!account) {
            showAlert('An error occurred');
            return;
        }

        props.f7router.navigate('/account/reconciliation_statements?accountId=' + account.id);
        showAccountMoreActionSheet.value = false;
        accountForMoreActionSheet.value = null;
    }

    function updateLastReconciledTime(account: Account | null): void {
        if (!account) {
            showAlert('An error occurred');
            return;
        }

        showConfirm('Are you sure you want to update the last reconciled time of this account to the current time?', () => {
            updatingLastReconciledTime.value = true;
            showLoading(() => updatingLastReconciledTime.value);

            accountsStore.updateAccountLastReconciledTime(account.id, getCurrentUnixTime()).then(() => {
                updatingLastReconciledTime.value = false;
                hideLoading();
                showToast('Last reconciled time have been updated');

                if (accountsStore.accountListStateInvalid && !loading.value) {
                    reload();
                }
            }).catch(error => {
                updatingLastReconciledTime.value = false;
                hideLoading();

                if (!error.processed) {
                    showToast(error.message || error);
                }
            });

            showAccountMoreActionSheet.value = false;
            accountForMoreActionSheet.value = null;
        });
    }

    function moveAllTransactions(account: Account | null): void {
        if (!account) {
            showAlert('An error occurred');
            return;
        }

        props.f7router.navigate('/account/move_all_transactions?fromAccountId=' + account.id);
        showAccountMoreActionSheet.value = false;
        accountForMoreActionSheet.value = null;
    }

    function showPasswordSheetForClearAllTransaction(account: Account | null): void {
        if (!account) {
            showAlert('An error occurred');
            return;
        }

        accountToClearTransactions.value = account;
        currentPasswordForClearData.value = '';
        showInputPasswordSheetForClearAllTransactions.value = true;
        showAccountMoreActionSheet.value = false;
        accountForMoreActionSheet.value = null;
    }

    function clearAllTransactions(password: string): void {
        if (!accountToClearTransactions.value) {
            showAlert('An error occurred');
            return;
        }

        clearingData.value = true;
        showLoading(() => clearingData.value);

        rootStore.clearAllUserTransactionsOfAccount({
            accountId: accountToClearTransactions.value.id,
            password: password
        }).then(() => {
            clearingData.value = false;
            currentPasswordForClearData.value = '';
            hideLoading();

            showInputPasswordSheetForClearAllTransactions.value = false;
            showToast('All transactions in this account have been cleared');
        }).catch(error => {
            clearingData.value = false;
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function hide(account: Account, hidden: boolean): void {
        showLoading();

        accountsStore.hideAccount({
            account: account,
            hidden: hidden
        }).then(() => {
            hideLoading();
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function remove(account: Account | null, confirm: boolean): void {
        if (!account) {
            showAlert('An error occurred');
            return;
        }

        if (!confirm) {
            accountToDelete.value = account;
            showDeleteActionSheet.value = true;
            return;
        }

        showDeleteActionSheet.value = false;
        accountToDelete.value = null;
        showLoading();

        accountsStore.deleteAccount({
            account: account,
            beforeResolve: (done) => {
                onSwipeoutDeleted(getAccountDomId(account), done);
            }
        }).then(() => {
            hideLoading();
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function setSortable(): void {
        if (sortable.value) {
            return;
        }

        showHidden.value = true;
        sortable.value = true;
        displayOrderModified.value = false;
    }

    function saveSortResult(): void {
        if (!displayOrderModified.value) {
            showHidden.value = false;
            sortable.value = false;
            return;
        }

        displayOrderSaving.value = true;
        showLoading();

        accountsStore.updateAccountDisplayOrders().then(() => {
            displayOrderSaving.value = false;
            hideLoading();

            showHidden.value = false;
            sortable.value = false;
            displayOrderModified.value = false;
        }).catch(error => {
            displayOrderSaving.value = false;
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function cancelSort(): void {
        if (!displayOrderModified.value) {
            showHidden.value = false;
            sortable.value = false;
            return;
        }

        displayOrderSaving.value = true;
        showLoading();

        accountsStore.loadAllAccounts({
            force: false
        }).then(() => {
            displayOrderSaving.value = false;
            hideLoading();

            showHidden.value = false;
            sortable.value = false;
            displayOrderModified.value = false;
        }).catch(error => {
            displayOrderSaving.value = false;
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function setAccountsIncludedInTotal(): void {
        props.f7router.navigate('/settings/filter/account?type=accountListTotalAmount');
    }

    function onSort(event: { el: { id: string }; from: number; to: number }): void {
        if (!event || !event.el || !event.el.id) {
            showToast('Unable to move account');
            return;
        }

        const id = parseAccountIdFromDomId(event.el.id);

        if (!id) {
            showToast('Unable to move account');
            return;
        }

        accountsStore.changeAccountDisplayOrder({
            accountId: id,
            from: event.from - 1, // first item in the list is title, so the index need minus one
            to: event.to - 1,
            updateListOrder: true,
            updateGlobalListOrder: true
        }).then(() => {
            displayOrderModified.value = true;
        }).catch(error => {
            showToast(error.message || error);
        });
    }

    function onPageAfterIn(): void {
        if (accountsStore.accountListStateInvalid && !loading.value) {
            reload();
        }

        routeBackOnError(props.f7router, loadingError);
    }

    init();

    return {
        tt, loading, showHidden, displayOrderModified, showAccountBalance, customAccountCategoryOrder, useLastReconciledTime,
        allCategorizedAccountsMap, allAccountCount, maxCategoryAccountCount, accountCategoryTotalBalance, accountBalance,
        accountBalanceOrAvailableCredit, sortable, accountForMoreActionSheet, accountToDelete, accountToClearTransactions,
        currentPasswordForClearData, clearingData, showAccountMoreActionSheet, showMoreActionSheet, showDeleteActionSheet,
        showInputPasswordSheetForClearAllTransactions, displayOrderSaving, textDirection, showAvailableCreditForCreditCard,
        firstShowingIds, lastShowingIds, hasAnyVisibleAccount, noAvailableAccount, hasAccount, hasVisibleSubAccount,
        getAccountDomId, reload, edit, showMoreActionSheetForAccount, showReconciliationStatement, updateLastReconciledTime,
        moveAllTransactions, showPasswordSheetForClearAllTransaction, clearAllTransactions, hide, remove, setSortable,
        saveSortResult, cancelSort, setAccountsIncludedInTotal, onSort, onPageAfterIn,
    };
}
