import { ACCOUNT_CURRENCY_NOT_SET_VALUE } from '@/consts/currency.ts';
import type { Account } from './account.ts';
import type { AccountCreateRequest, AccountModifyRequest } from './types.ts';
import { AccountType, AccountCategory } from '@/core/account.ts';

// Account ordering and wire-format rules stay independent of mutable model state.
export function compareAccounts(account: Account, other: Account, accountCategoryDisplayOrders: Record<number, number>, allAccountsMap?: Record<string, Account>): number {
        if (account.category !== other.category) {
            const thisCategoryDisplayOrder = accountCategoryDisplayOrders[account.category];
            const otherCategoryDisplayOrder = accountCategoryDisplayOrders[other.category];

            if (!thisCategoryDisplayOrder) {
                return 1;
            }

            if (!otherCategoryDisplayOrder) {
                return -1;
            }

            return thisCategoryDisplayOrder - otherCategoryDisplayOrder;
        }

        if (account.parentId === other.parentId) {
            return account.displayOrder - other.displayOrder;
        }

        if (account.id === other.parentId) {
            return -1;
        } else if (other.id === account.parentId) {
            return 1;
        }

        let thisDisplayOrder: number | null = account.displayOrder;
        let otherDisplayOrder: number | null = other.displayOrder;

        if (account.parentId && account.parentId !== '0') {
            if (allAccountsMap && allAccountsMap[account.parentId]) {
                thisDisplayOrder = (allAccountsMap[account.parentId] as Account).displayOrder;
            } else {
                thisDisplayOrder = null;
            }
        }

        if (other.parentId && other.parentId !== '0') {
            if (allAccountsMap && allAccountsMap[other.parentId]) {
                otherDisplayOrder = (allAccountsMap[other.parentId] as Account).displayOrder;
            } else {
                otherDisplayOrder = null;
            }
        }

        if (thisDisplayOrder !== null && otherDisplayOrder !== null) {
            return thisDisplayOrder - otherDisplayOrder;
        } else {
            return account.id.localeCompare(other.id);
        }
    }

export function createAccountRequest(account: Account, clientSessionId: string, subAccounts?: Account[], parentAccount?: Account): AccountCreateRequest {
        let subAccountCreateRequests: AccountCreateRequest[] | undefined = undefined;

        if (account.type === AccountType.MultiSubAccounts.type) {
            subAccountCreateRequests = [];

            if (!subAccounts) {
                subAccounts = account.subAccounts;
            }

            if (subAccounts) {
                for (const subAccount of subAccounts) {
                    subAccountCreateRequests.push(subAccount.toCreateRequest(clientSessionId, undefined, account));
                }
            }
        }

        return {
            name: account.name,
            category: parentAccount ? parentAccount.category : account.category,
            type: parentAccount ? AccountType.SingleAccount.type : account.type,
            icon: account.icon,
            iconType: account.iconType,
            color: account.color,
            currency: parentAccount || account.type === AccountType.SingleAccount.type || account.category === AccountCategory.CreditCard.type ? account.currency : ACCOUNT_CURRENCY_NOT_SET_VALUE,
            balance: parentAccount || account.type === AccountType.SingleAccount.type ? account.balance : '0',
            balanceTime: (parentAccount || account.type === AccountType.SingleAccount.type) && account.balanceTime ? account.balanceTime : 0,
            comment: account.comment,
            creditCardStatementDate: !parentAccount && account.category === AccountCategory.CreditCard.type ? account.creditCardStatementDate : undefined,
            creditCardLimit: !parentAccount && account.category === AccountCategory.CreditCard.type && account.numericCreditCardLimit > 0 ? account.creditCardLimit : undefined,
            subAccounts: !parentAccount ? subAccountCreateRequests : undefined,
            clientSessionId: !parentAccount ? clientSessionId : undefined
        };
    }

export function modifyAccountRequest(account: Account, clientSessionId: string, subAccounts?: Account[], parentAccount?: Account): AccountModifyRequest {
        let subAccountModifyRequests: AccountModifyRequest[] | undefined = undefined;

        if (account.type === AccountType.MultiSubAccounts.type) {
            subAccountModifyRequests = [];

            if (!subAccounts) {
                subAccounts = account.subAccounts;
            }

            if (subAccounts) {
                for (const subAccount of subAccounts) {
                    subAccountModifyRequests.push(subAccount.toModifyRequest(clientSessionId, undefined, account));
                }
            }
        }

        return {
            id: account.id || '0',
            name: account.name,
            category: parentAccount ? parentAccount.category : account.category,
            icon: account.icon,
            iconType: account.iconType,
            color: account.color,
            currency: (parentAccount && (!account.id || account.id === '0')) || (!parentAccount && account.category === AccountCategory.CreditCard.type) ? account.currency : undefined,
            balance: parentAccount && (!account.id || account.id === '0') ? account.balance : undefined,
            balanceTime: parentAccount && (!account.id || account.id === '0') ? account.balanceTime : undefined,
            lastReconciledTime: account.lastReconciledTime,
            comment: account.comment,
            creditCardStatementDate: !parentAccount && account.category === AccountCategory.CreditCard.type ? account.creditCardStatementDate : undefined,
            creditCardLimit: !parentAccount && account.category === AccountCategory.CreditCard.type && account.numericCreditCardLimit > 0 ? account.creditCardLimit : undefined,
            hidden: !account.visible,
            subAccounts: !parentAccount ? subAccountModifyRequests : undefined,
            clientSessionId: !parentAccount ? clientSessionId : undefined
        };
    }
