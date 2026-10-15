import { Account } from './account.ts';

export class AccountWithDisplayBalance extends Account {
    public displayBalance: string;

    private constructor(account: Account, displayBalance: string) {
        super(
            account.id,
            account.name,
            account.parentId,
            account.category,
            account.type,
            account.icon,
            account.iconType,
            account.color,
            account.currency,
            account.balance,
            account.comment,
            account.displayOrder,
            account.visible,
            account.balanceTime,
            account.lastReconciledTime,
            account.creditCardStatementDate,
            account.creditCardLimit,
            account.isAsset,
            account.isLiability,
            account.subAccounts
        );

        this.displayBalance = displayBalance;
    }

    public static fromAccount(account: Account, displayBalance: string): AccountWithDisplayBalance {
        return new AccountWithDisplayBalance(account, displayBalance);
    }
}
