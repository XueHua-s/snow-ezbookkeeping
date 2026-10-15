import { DEFAULT_ACCOUNT_COLOR } from '@/consts/color.ts';
import { AccountCategory, AccountType } from '@/core/account.ts';
import type { ColorValue } from '@/core/color.ts';
import { IconType } from '@/core/icon.ts';
import { compareAccounts, createAccountRequest, modifyAccountRequest } from './operations.ts';
import type { AccountCreateRequest, AccountInfoResponse, AccountModifyRequest } from './types.ts';

export class Account implements AccountInfoResponse {
    public id: string;
    public name: string;
    public parentId: string;
    public category: number;
    public type: number;
    public icon: string;
    public iconType: number;
    public color: ColorValue;
    public currency: string;
    public balanceTime?: number;
    public lastReconciledTime?: number;
    public comment: string;
    public creditCardStatementDate?: number;
    public displayOrder: number;
    public visible: boolean;
    public subAccounts?: Account[];

    private _initialBalance?: string;
    private _numericBalance?: number;
    private _initialCreditCardLimit?: string;
    private _numericCreditCardLimit?: number;

    private readonly _isAsset?: boolean;
    private readonly _isLiability?: boolean;

    protected constructor(id: string, name: string, parentId: string, category: number, type: number, icon: string, iconType: number, color: string, currency: string, initialBalance: string, comment: string, displayOrder: number, visible: boolean, balanceTime?: number, lastReconciledTime?: number, creditCardStatementDate?: number, initialCreditCardLimit?: string, isAsset?: boolean, isLiability?: boolean, subAccounts?: Account[]) {
        this.id = id;
        this.name = name;
        this.parentId = parentId;
        this.category = category;
        this.type = type;
        this.icon = icon;
        this.iconType = iconType;
        this.color = color;
        this.currency = currency;
        this.balanceTime = balanceTime;
        this.lastReconciledTime = lastReconciledTime;
        this.comment = comment;
        this.displayOrder = displayOrder;
        this.visible = visible;
        this.creditCardStatementDate = creditCardStatementDate;

        this._initialBalance = initialBalance;
        this._numericBalance = undefined;
        this._initialCreditCardLimit = initialCreditCardLimit;
        this._numericCreditCardLimit = undefined;

        this._isAsset = isAsset;
        this._isLiability = isLiability;

        if (typeof(subAccounts) !== 'undefined') {
            this.subAccounts = subAccounts;
        } else {
            this.subAccounts = undefined;
        }
    }

    public get balance(): string {
        if (typeof(this._numericBalance) !== 'undefined') {
            return this._numericBalance.toString(10);
        }

        return this._initialBalance ?? '0';
    }

    public set balance(value: string) {
        throw new Error('account balance cannot be assigned directly');
    }

    public get numericBalance(): number {
        if (typeof (this._numericBalance) !== 'undefined') {
            return this._numericBalance;
        }

        this._numericBalance = parseInt(this._initialBalance ?? '0', 10);

        if (!Number.isSafeInteger(this._numericBalance)) {
            this._numericBalance = 0;
        }

        return this._numericBalance;
    }

    public set numericBalance(value: number) {
        this._numericBalance = value;
    }

    public get creditCardLimit(): string {
        if (typeof(this._numericCreditCardLimit) !== 'undefined') {
            return this._numericCreditCardLimit.toString(10);
        }

        return this._initialCreditCardLimit ?? '0';
    }

    public set creditCardLimit(value: string) {
        throw new Error('account credit card limit cannot be assigned directly');
    }

    public get numericCreditCardLimit(): number {
        if (typeof(this._numericCreditCardLimit) !== 'undefined') {
            return this._numericCreditCardLimit;
        }

        this._numericCreditCardLimit = parseInt(this.creditCardLimit ?? '0', 10);

        if (!Number.isSafeInteger(this._numericCreditCardLimit) || this._numericCreditCardLimit < 0) {
            this._numericCreditCardLimit = 0;
        }

        return this._numericCreditCardLimit;
    }

    public set numericCreditCardLimit(value: number) {
        if (value < 0) {
            this._numericCreditCardLimit = 0;
        } else {
            this._numericCreditCardLimit = value;
        }
    }

    public get isAsset(): boolean {
        if (typeof(this._isAsset) !== 'undefined') {
            return this._isAsset;
        }

        const accountCategory = AccountCategory.valueOf(this.category);

        if (accountCategory) {
            return accountCategory.isAsset;
        }

        return false;
    }

    public get isLiability(): boolean {
        if (typeof(this._isLiability) !== 'undefined') {
            return this._isLiability;
        }

        const accountCategory = AccountCategory.valueOf(this.category);

        if (accountCategory) {
            return accountCategory.isLiability;
        }

        return false;
    }

    public get hidden(): boolean {
        return !this.visible;
    }

    public equals(other: Account): boolean {
        const isEqual = this.id === other.id &&
            this.name === other.name &&
            this.parentId === other.parentId &&
            this.category === other.category &&
            this.type === other.type &&
            this.icon === other.icon &&
            this.iconType === other.iconType &&
            this.color === other.color &&
            this.currency === other.currency &&
            this.balance === other.balance &&
            this.numericBalance === other.numericBalance &&
            this.balanceTime === other.balanceTime &&
            this.lastReconciledTime === other.lastReconciledTime &&
            this.comment === other.comment &&
            this.displayOrder === other.displayOrder &&
            this.visible === other.visible &&
            this.creditCardStatementDate === other.creditCardStatementDate &&
            this.creditCardLimit === other.creditCardLimit &&
            this.numericCreditCardLimit === other.numericCreditCardLimit;

        if (!isEqual) {
            return false;
        }

        if (this.subAccounts && other.subAccounts) {
            if (this.subAccounts.length !== other.subAccounts.length) {
                return false;
            }

            for (let i = 0; i < this.subAccounts.length; i++) {
                if (!(this.subAccounts[i] as Account).equals(other.subAccounts[i] as Account)) {
                    return false;
                }
            }
        } else if ((this.subAccounts && this.subAccounts.length) || (other.subAccounts && other.subAccounts.length)) {
            return false;
        }

        return true;
    }

    public compareTo(other: Account, accountCategoryDisplayOrders: Record<number, number>, allAccountsMap?: Record<string, Account>): number {
        return compareAccounts(this, other, accountCategoryDisplayOrders, allAccountsMap);
    }

    public fillFrom(other: Account): void {
        this.id = other.id;
        this.category = other.category;
        this.type = other.type;
        this.name = other.name;
        this.icon = other.icon;
        this.iconType = other.iconType;
        this.color = other.color;
        this.currency = other.currency;
        this.balanceTime = other.balanceTime;
        this.lastReconciledTime = other.lastReconciledTime;
        this.comment = other.comment;
        this.creditCardStatementDate = other.creditCardStatementDate;
        this.visible = other.visible;

        this._initialBalance = other._initialBalance;
        this._numericBalance = other._numericBalance;
        this._initialCreditCardLimit = other._initialCreditCardLimit;
        this._numericCreditCardLimit = other._numericCreditCardLimit;
    }

    public setSuitableIcon(oldCategory: number, newCategory: number): void {
        const allCategories = AccountCategory.values();

        for (const category of allCategories) {
            if (category.type === oldCategory) {
                if (this.icon !== category.defaultAccountIconId) {
                    return;
                } else {
                    break;
                }
            }
        }

        for (const category of allCategories) {
            if (category.type === newCategory) {
                this.icon = category.defaultAccountIconId;
                this.iconType = IconType.System;
            }
        }
    }

    public toCreateRequest(clientSessionId: string, subAccounts?: Account[], parentAccount?: Account): AccountCreateRequest {
        return createAccountRequest(this, clientSessionId, subAccounts, parentAccount);
    }

    public toModifyRequest(clientSessionId: string, subAccounts?: Account[], parentAccount?: Account): AccountModifyRequest {
        return modifyAccountRequest(this, clientSessionId, subAccounts, parentAccount);
    }

    public getAccountOrSubAccountId(subAccountId?: string): string | null {
        return this.getAccountOrSubAccount(subAccountId)?.id ?? null;
    }

    public isAccountOrSubAccountHidden(subAccountId?: string): boolean {
        return this.getAccountOrSubAccount(subAccountId)?.hidden ?? false;
    }

    public getAccountOrSubAccountComment(subAccountId?: string): string | null {
        return this.getAccountOrSubAccount(subAccountId)?.comment ?? null;
    }

    public getAccountOrSubAccount(subAccountId?: string): Account | null {
        if (this.type === AccountType.SingleAccount.type) {
            return this;
        } else if (this.type === AccountType.MultiSubAccounts.type && !subAccountId) {
            return this;
        } else if (this.type === AccountType.MultiSubAccounts.type && subAccountId) {
            if (!this.subAccounts || !this.subAccounts.length) {
                return null;
            }

            for (const subAccount of this.subAccounts) {
                if (subAccountId && subAccountId === subAccount.id) {
                    return subAccount;
                }
            }

            return null;
        } else {
            return null;
        }
    }

    public getSubAccount(subAccountId?: string): Account | null {
        if (!this.subAccounts || !this.subAccounts.length) {
            return null;
        }

        for (const subAccount of this.subAccounts) {
            if (subAccountId && subAccountId === subAccount.id) {
                return subAccount;
            }
        }

        return null;
    }

    public getSubAccountCurrencies(showHidden: boolean, subAccountId?: string): string[] {
        if (!this.subAccounts || !this.subAccounts.length) {
            return [];
        }

        const subAccountCurrenciesMap: Record<string, boolean> = {};
        const subAccountCurrencies: string[] = [];

        for (const subAccount of this.subAccounts) {
            if (!showHidden && subAccount.hidden) {
                continue;
            }

            if (subAccountId && subAccountId === subAccount.id) {
                return [subAccount.currency];
            } else {
                if (!subAccountCurrenciesMap[subAccount.currency]) {
                    subAccountCurrenciesMap[subAccount.currency] = true;
                    subAccountCurrencies.push(subAccount.currency);
                }
            }
        }

        return subAccountCurrencies;
    }

    public cloneSelf(): Account {
        return new Account(
            this.id,
            this.name,
            this.parentId,
            this.category,
            this.type,
            this.icon,
            this.iconType,
            this.color,
            this.currency,
            this.balance,
            this.comment,
            this.displayOrder,
            this.visible,
            this.balanceTime,
            this.lastReconciledTime,
            this.creditCardStatementDate,
            this.creditCardLimit,
            this.isAsset,
            this.isLiability
        );
    }

    public clone(): Account {
        return new Account(
            this.id,
            this.name,
            this.parentId,
            this.category,
            this.type,
            this.icon,
            this.iconType,
            this.color,
            this.currency,
            this.balance,
            this.comment,
            this.displayOrder,
            this.visible,
            this.balanceTime,
            this.lastReconciledTime,
            this.creditCardStatementDate,
            this.creditCardLimit,
            this.isAsset,
            this.isLiability,
            typeof(this.subAccounts) !== 'undefined' ? Account.cloneAccounts(this.subAccounts) : undefined);
    }

    public createNewSubAccount(currency: string, balanceTime: number): Account {
        return new Account(
            '', // id
            '', // name
            '', // parentId
            0, // category
            0, // type
            this.icon, // icon
            this.iconType, // iconType
            this.color, // color
            currency, // currency
            '0', // balance
            '', // comment
            0, // displayOrder
            true, // visible
            balanceTime, // balanceTime
            undefined, // lastReconciledTime
            0, // creditCardStatementDate
            undefined // creditCardLimit
        );
    }

    public static createNewAccount(accountCategory: AccountCategory, currency: string, balanceTime: number): Account {
        return new Account(
            '', // id
            '', // name
            '', // parentId
            accountCategory.type, // category
            AccountType.SingleAccount.type, // type
            accountCategory.defaultAccountIconId, // icon
            IconType.System, // iconType
            DEFAULT_ACCOUNT_COLOR, // color
            currency, // currency
            '0', // balance
            '', // comment
            0, // displayOrder
            true, // visible
            balanceTime, // balanceTime
            undefined, // lastReconciledTime
            0, // creditCardStatementDate
            undefined // creditCardLimit
        );
    }

    public static of(accountResponse: AccountInfoResponse): Account {
        return new Account(
            accountResponse.id,
            accountResponse.name,
            accountResponse.parentId,
            accountResponse.category,
            accountResponse.type,
            accountResponse.icon,
            accountResponse.iconType,
            accountResponse.color,
            accountResponse.currency,
            accountResponse.balance,
            accountResponse.comment,
            accountResponse.displayOrder,
            !accountResponse.hidden,
            undefined,
            accountResponse.lastReconciledTime,
            accountResponse.creditCardStatementDate,
            accountResponse.creditCardLimit,
            accountResponse.isAsset,
            accountResponse.isLiability,
            accountResponse.subAccounts ? Account.ofMulti(accountResponse.subAccounts) : undefined
        );
    }

    public static ofMulti(accountResponses: AccountInfoResponse[]): Account[] {
        const accounts: Account[] = [];

        for (const accountResponse of accountResponses) {
            accounts.push(Account.of(accountResponse));
        }

        return accounts;
    }

    public static findAccountNameById(accounts: Account[], accountId: string, defaultName?: string): string | undefined {
        for (const account of accounts) {
            if (account.id === accountId) {
                return account.name;
            }
        }

        return defaultName;
    }

    public static cloneAccounts(accounts: Account[]): Account[] {
        const clonedAccounts: Account[] = [];

        for (const account of accounts) {
            clonedAccounts.push(account.clone());
        }

        return clonedAccounts;
    }

    public static sortAccounts(accounts: Account[], accountCategoryDisplayOrders: Record<number, number>, allAccountsMap?: Record<string, Account>): Account[] {
        if (!accounts || !accounts.length) {
            return accounts;
        }

        return accounts.sort((account1, account2) => account1.compareTo(account2, accountCategoryDisplayOrders, allAccountsMap));
    }
}
