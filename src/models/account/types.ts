import type { BigDecimal, BigDecimalWithSuffix, HiddenAmount } from '@/core/numeral.ts';
import type { Account } from './account.ts';
import type { AccountWithDisplayBalance } from './display.ts';

export interface AccountCreateRequest {
    readonly name: string;
    readonly category: number;
    readonly type: number;
    readonly icon: string;
    readonly iconType: number;
    readonly color: string;
    readonly currency: string;
    readonly balance: string;
    readonly balanceTime: number;
    readonly comment: string;
    readonly creditCardStatementDate?: number;
    readonly creditCardLimit?: string;
    readonly subAccounts?: AccountCreateRequest[];
    readonly clientSessionId?: string;
}

export interface AccountModifyRequest {
    readonly id: string;
    readonly name: string;
    readonly category: number;
    readonly icon: string;
    readonly iconType: number;
    readonly color: string;
    readonly currency?: string;
    readonly balance?: string;
    readonly balanceTime?: number;
    readonly lastReconciledTime?: number;
    readonly comment: string;
    readonly creditCardStatementDate?: number;
    readonly creditCardLimit?: string;
    readonly hidden: boolean;
    readonly subAccounts?: AccountModifyRequest[];
    readonly clientSessionId?: string;
}

export interface AccountUpdateLastReconciledTimeRequest {
    readonly id: string;
    readonly lastReconciledTime: number;
}

export interface AccountInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly parentId: string;
    readonly category: number;
    readonly type: number;
    readonly icon: string;
    readonly iconType: number;
    readonly color: string;
    readonly currency: string;
    readonly balance: string;
    readonly lastReconciledTime?: number;
    readonly comment: string;
    readonly creditCardStatementDate?: number;
    readonly creditCardLimit?: string;
    readonly displayOrder: number;
    readonly isAsset?: boolean;
    readonly isLiability?: boolean;
    readonly hidden: boolean;
    readonly subAccounts?: AccountInfoResponse[];
}

export interface AccountHideRequest {
    readonly id: string;
    readonly hidden: boolean;
}

export interface AccountMoveRequest {
    readonly newDisplayOrders: AccountNewDisplayOrderRequest[];
}

export interface AccountNewDisplayOrderRequest {
    readonly id: string;
    readonly displayOrder: number;
}

export interface AccountDeleteRequest {
    readonly id: string;
}

export interface AccountBalance {
    readonly balance: BigDecimal;
    readonly category: number;
    readonly creditCardLimit?: {
        readonly amount: BigDecimal;
        readonly currency: string;
        readonly shareByCount: number;
    };
    readonly isAsset: boolean;
    readonly isLiability: boolean;
    readonly currency: string;
}

export interface AccountDisplayBalance {
    readonly balance: BigDecimal | HiddenAmount | BigDecimalWithSuffix;
    readonly currency: string;
}

export interface CategorizedAccount {
    readonly category: number;
    readonly name: string;
    readonly icon: string;
    readonly accounts: Account[];
}

export class CategorizedAccountWithDisplayBalance {
    public category: number;
    public name: string;
    public icon: string;
    public accounts: AccountWithDisplayBalance[];
    public displayBalance: string;

    private constructor(category: number, name: string, icon: string, accounts: AccountWithDisplayBalance[], displayBalance: string) {
        this.category = category;
        this.name = name;
        this.icon = icon;
        this.accounts = accounts;
        this.displayBalance = displayBalance;
    }

    public static of(categorizedAccount: CategorizedAccount, accounts: AccountWithDisplayBalance[], displayBalance: string): CategorizedAccountWithDisplayBalance {
        return new CategorizedAccountWithDisplayBalance(categorizedAccount.category, categorizedAccount.name, categorizedAccount.icon, accounts, displayBalance);
    }
}

export interface AccountShowingIds {
    readonly accounts: Record<number, string>;
    readonly subAccounts: Record<string, string>;
}
