import { DecimalSeparator, DigitGroupingSymbol } from './symbols.ts';

export class KnownAmountFormat {
    private static readonly allInstances: KnownAmountFormat[] = [];
    private static readonly allInstancesByType: Record<string, KnownAmountFormat> = {};

    public static readonly DotDecimalSeparator = new KnownAmountFormat('1234.56', DecimalSeparator.Dot, undefined, /^-?[0-9]+(\.[0-9]+)?$/);
    public static readonly CommaDecimalSeparator = new KnownAmountFormat('1234,56', DecimalSeparator.Comma, undefined, /^-?[0-9]+(,[0-9]+)?$/);
    public static readonly DotDecimalSeparatorWithCommaGroupingSymbol = new KnownAmountFormat('1,234.56', DecimalSeparator.Dot, DigitGroupingSymbol.Comma, /^-?([0-9]+,)*[0-9]+(\.[0-9]+)?$/);
    public static readonly CommaDecimalSeparatorWithDotGroupingSymbol = new KnownAmountFormat('1.234,56', DecimalSeparator.Comma, DigitGroupingSymbol.Dot, /^-?([0-9]+\.)*[0-9]+(,[0-9]+)?$/);
    public static readonly DotDecimalSeparatorWithSpaceGroupingSymbol = new KnownAmountFormat('1 234.56', DecimalSeparator.Dot, DigitGroupingSymbol.Space, /^-?([0-9]+([    ]))*[0-9]+(\.[0-9]+)?$/);
    public static readonly CommaDecimalSeparatorWithSpaceGroupingSymbol = new KnownAmountFormat('1 234,56', DecimalSeparator.Comma, DigitGroupingSymbol.Space, /^-?([0-9]+([    ]))*[0-9]+(,[0-9]+)?$/);
    public static readonly DotDecimalSeparatorWithApostropheGroupingSymbol = new KnownAmountFormat('1\'234.56', DecimalSeparator.Dot, DigitGroupingSymbol.Apostrophe, /^-?([0-9]+')*[0-9]+(\.[0-9]+)?$/);
    public static readonly CommaDecimalSeparatorWithApostropheGroupingSymbol = new KnownAmountFormat('1\'234,56', DecimalSeparator.Comma, DigitGroupingSymbol.Apostrophe, /^-?([0-9]+')*[0-9]+(,[0-9]+)?$/);

    public readonly format: string;
    public readonly decimalSeparator: DecimalSeparator;
    public readonly digitGroupingSymbol?: DigitGroupingSymbol;
    public readonly type: string;
    private readonly regex: RegExp;

    private constructor(format: string, decimalSeparator: DecimalSeparator, digitGroupingSymbol: DigitGroupingSymbol | undefined, regex: RegExp) {
        this.format = format;
        this.decimalSeparator = decimalSeparator;
        this.digitGroupingSymbol = digitGroupingSymbol;
        this.type = this.decimalSeparator.type + '-' + (this.digitGroupingSymbol ? this.digitGroupingSymbol.type : 0).toString();
        this.regex = regex;

        KnownAmountFormat.allInstances.push(this);
        KnownAmountFormat.allInstancesByType[this.type] = this;
    }

    public isValid(amount: string): boolean {
        return this.regex.test(amount);
    }

    public static values(): KnownAmountFormat[] {
        return KnownAmountFormat.allInstances;
    }

    public static valueOf(type: string): KnownAmountFormat | undefined {
        return KnownAmountFormat.allInstancesByType[type];
    }

    public static detect(amount: string): KnownAmountFormat[] | undefined {
        const result: KnownAmountFormat[] = [];

        for (const format of KnownAmountFormat.allInstances) {
            if (format.isValid(amount)) {
                result.push(format);
            }
        }

        return result.length > 0 ? result : undefined;
    }

    public static detectMulti(amounts: string[]): KnownAmountFormat[] | undefined {
        const detectedCounts: Record<string, number> = {};

        for (const amount of amounts) {
            const detectedFormats = KnownAmountFormat.detect(amount);

            if (detectedFormats) {
                for (const format of detectedFormats) {
                    detectedCounts[format.type] = (detectedCounts[format.type] || 0) + 1;
                }
            } else {
                return undefined;
            }
        }

        const result: KnownAmountFormat[] = [];

        for (const format of KnownAmountFormat.allInstances) {
            if (detectedCounts[format.type] === amounts.length) {
                result.push(format);
            }
        }

        return result.length > 0 ? result : undefined;
    }
}

export class AmountFilterType {
    private static readonly allInstances: AmountFilterType[] = [];
    private static readonly allInstancesByType: Record<string, AmountFilterType> = {};

    public static readonly GreaterThan = new AmountFilterType('gt', 'Greater than', 1,
        (amount: number, ...params: number[]) => {
            return params && params.length > 0 && amount > (params[0] as number);
        }
    );
    public static readonly LessThan = new AmountFilterType('lt', 'Less than', 1,
        (amount: number, ...params: number[]) => {
            return params && params.length > 0 && amount < (params[0] as number);
        }
    );
    public static readonly EqualTo = new AmountFilterType('eq', 'Equal to', 1,
        (amount: number, ...params: number[]) => {
            return params && params.length > 0 && amount === (params[0] as number);
        }
    );
    public static readonly NotEqualTo = new AmountFilterType('ne', 'Not equal to', 1,
        (amount: number, ...params: number[]) => {
            return params && params.length > 0 && amount !== (params[0] as number);
        }
    );
    public static readonly Between = new AmountFilterType('bt', 'Between', 2,
        (amount: number, ...params: number[]) => {
            return params && params.length > 1 && amount >= (params[0] as number) && amount <= (params[1] as number);
        }
    );
    public static readonly NotBetween = new AmountFilterType('nb', 'Not between', 2,
        (amount: number, ...params: number[]) => {
            return params && params.length > 1 && (amount < (params[0] as number) || amount > (params[1] as number));
        }
    );

    public readonly type: string;
    public readonly name: string;
    public readonly paramCount: number;
    private readonly matchFn: (amount: number, ...params: number[]) => boolean;

    private constructor(type: string, name: string, paramCount: number, matchFn: (amount: number, ...params: number[]) => boolean) {
        this.type = type;
        this.name = name;
        this.paramCount = paramCount;
        this.matchFn = matchFn;

        AmountFilterType.allInstances.push(this);
        AmountFilterType.allInstancesByType[type] = this;
    }

    public toTextualFilter(...params: number[]): string {
        if (this.paramCount === 1) {
            return `${this.type}:${params[0] ?? ''}`;
        } else if (this.paramCount === 2) {
            return `${this.type}:${params[0] ?? ''}:${params[1] ?? ''}`;
        } else {
            return '';
        }
    }

    public static parseTextualFilter(filter: string): { filterType: AmountFilterType, params: number[] } | undefined {
        const parts = filter.split(':');

        if (parts.length < 2) {
            return undefined;
        }

        const filterType = AmountFilterType.valueOf(parts[0] as string);

        if (!filterType) {
            return undefined;
        }

        if (parts.length - 1 !== filterType.paramCount) {
            return undefined;
        }

        const params: number[] = [];

        for (let i = 1; i < parts.length; i++) {
            const param = parseInt(parts[i] as string);

            if (Number.isNaN(param)) {
                return undefined;
            }

            params.push(param);
        }

        return {
            filterType: filterType,
            params: params
        };
    }

    public static match(filter: string, amount: number): boolean {
        const parts = filter.split(':');

        if (parts.length < 2) {
            return false;
        }

        const filterType = AmountFilterType.valueOf(parts[0] as string);

        if (!filterType) {
            return false;
        }

        if (parts.length - 1 !== filterType.paramCount) {
            return false;
        }

        const params: number[] = [];

        for (let i = 1; i < parts.length; i++) {
            const param = parseInt(parts[i] as string);

            if (Number.isNaN(param)) {
                return false;
            }

            params.push(param);
        }

        return filterType.matchFn(amount, ...params);
    }

    public static values(): AmountFilterType[] {
        return AmountFilterType.allInstances;
    }

    public static valueOf(type: string): AmountFilterType | undefined {
        return AmountFilterType.allInstancesByType[type];
    }
}
