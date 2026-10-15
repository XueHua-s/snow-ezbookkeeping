import { type TypeAndName } from '@/core/base.ts';

export interface TransactionListDisplayTotalAmount {
    incomeIsZero: boolean;
    expenseIsZero: boolean;
    income: string;
    expense: string;
    incomeInDefaultCurrency: string;
    expenseInDefaultCurrency: string;
}

export class TransactionListPageType implements TypeAndName {
    private static readonly allInstances: TransactionListPageType[] = [];
    private static readonly allInstancesByType: Record<number, TransactionListPageType> = {};

    public static readonly List = new TransactionListPageType(0, 'Transaction List');
    public static readonly Calendar = new TransactionListPageType(1, 'Transaction Calendar');
    public static readonly Gallery = new TransactionListPageType(2, 'Transaction Gallery');

    public static readonly Default = TransactionListPageType.List;

    public readonly type: number;
    public readonly name: string;

    private constructor(type: number, name: string) {
        this.type = type;
        this.name = name;

        TransactionListPageType.allInstances.push(this);
        TransactionListPageType.allInstancesByType[type] = this;
    }

    public static values(): TransactionListPageType[] {
        return TransactionListPageType.allInstances;
    }

    public static valueOf(type: number): TransactionListPageType | undefined {
        return TransactionListPageType.allInstancesByType[type];
    }
}
