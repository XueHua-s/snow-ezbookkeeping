import type { TypeAndName } from '@/core/base.ts';
import type { BigDecimal, NumeralSymbolType } from './types.ts';

export class NumeralSystem implements TypeAndName {
    private static readonly allInstances: NumeralSystem[] = [];
    private static readonly allInstancesByType: Record<number, NumeralSystem> = {};
    private static readonly allInstancesByTypeName: Record<string, NumeralSystem> = {};
    private static readonly allDigitsToWesternArabic: Record<string, number> = {};
    private static readonly allDigitsToNumeralSystem: Record<string, NumeralSystem> = {};

    public static readonly LanguageDefaultType: number = 0;
    public static readonly WesternArabicNumerals = new NumeralSystem(1, 'WesternArabicNumerals', 'Western Arabic Numerals', '\u0030');
    public static readonly EasternArabicNumerals = new NumeralSystem(2, 'EasternArabicNumerals', 'Eastern Arabic Numerals', '\u0660');
    public static readonly PersianDigits = new NumeralSystem(3, 'PersianDigits', 'Persian Digits', '\u06F0');
    public static readonly BurmeseNumerals = new NumeralSystem(4, 'BurmeseNumerals', 'Burmese Numerals', '\u1040');
    public static readonly DevanagariNumerals = new NumeralSystem(5, 'DevanagariNumerals', 'Devanagari Numerals', '\u0966');

    public static readonly Default = NumeralSystem.WesternArabicNumerals;

    public readonly type: number;
    public readonly typeName: string;
    public readonly name: string;
    public readonly digitZero: string;
    public readonly doubleDigitZero: string;
    public readonly textualAllDigits: string;
    private readonly allDigits: string[];
    private readonly digitsToWesternArabic: Record<string, number> = {};

    private constructor(type: number, typeName: string, name: string, digitZero: string) {
        this.type = type;
        this.typeName = typeName;
        this.name = name;
        this.digitZero = digitZero;
        this.doubleDigitZero = digitZero + digitZero;
        this.allDigits = [];
        this.digitsToWesternArabic = {};

        for (let i = 0; i < 10; i++) {
            const digit = String.fromCharCode(this.digitZero.charCodeAt(0) + i);
            this.allDigits.push(digit);
            this.digitsToWesternArabic[digit] = i;
            NumeralSystem.allDigitsToWesternArabic[digit] = i;
            NumeralSystem.allDigitsToNumeralSystem[digit] = this;
        }

        this.textualAllDigits = this.allDigits.join('');

        NumeralSystem.allInstances.push(this);
        NumeralSystem.allInstancesByType[type] = this;
        NumeralSystem.allInstancesByTypeName[typeName] = this;
    }

    public getAllDigits(): string[] {
        return this.allDigits.slice();
    }

    public isDigit(digit: string): boolean {
        return this.digitsToWesternArabic.hasOwnProperty(digit);
    }

    public getLocalizedDigit(digit: number): string {
        if (digit < 0 || digit > 9) {
            return '';
        }

        return this.allDigits[digit] as string;
    }

    public parseInt(value: string): number {
        if (!value) {
            return Number.NaN;
        }

        if (this.type === NumeralSystem.WesternArabicNumerals.type) {
            return parseInt(value, 10);
        } else {
            const westernArabicValue = this.replaceLocalizedDigitsToWesternArabicDigits(value);
            return parseInt(westernArabicValue, 10);
        }
    }

    public formatNumber(value: number): string {
        if (Number.isNaN(value) || !Number.isFinite(value)) {
            return value.toString();
        }

        if (this.type === NumeralSystem.WesternArabicNumerals.type) {
            return value.toString(10);
        }

        if (value === 0) {
            return this.digitZero;
        }

        return this.replaceWesternArabicDigitsToLocalizedDigits(value.toString(10));
    }

    public formatBigDecimal(value: BigDecimal): string {
        if (!value) {
            return '';
        }

        if (this.type === NumeralSystem.WesternArabicNumerals.type) {
            return value.toString();
        }

        if (value.isZero()) {
            return this.digitZero;
        }

        return this.replaceWesternArabicDigitsToLocalizedDigits(value.toString());
    }

    public replaceWesternArabicDigitsToLocalizedDigits(value: string): string {
        if (!value) {
            return '';
        }

        let result = '';

        for (let i = 0; i < value.length; i++) {
            const ch = value.charAt(i);

            if (NumeralSystem.WesternArabicNumerals.isDigit(ch)) {
                const digit = NumeralSystem.WesternArabicNumerals.digitsToWesternArabic[ch] as number;
                result += this.allDigits[digit] as string;
            } else {
                result += ch;
            }
        }

        return result;
    }

    public replaceLocalizedDigitsToWesternArabicDigits(value: string): string {
        if (!value) {
            return '';
        }

        let result = '';

        for (let i = 0; i < value.length; i++) {
            const ch = value.charAt(i);

            if (this.isDigit(ch)) {
                const digit = this.digitsToWesternArabic[ch] as number;
                result += NumeralSystem.WesternArabicNumerals.allDigits[digit] as string;
            } else {
                result += ch;
            }
        }

        return result;
    }

    public static values(): NumeralSystem[] {
        return NumeralSystem.allInstances;
    }

    public static valueOf(type: number): NumeralSystem | undefined {
        return NumeralSystem.allInstancesByType[type];
    }

    public static parse(typeName: string): NumeralSystem | undefined {
        return NumeralSystem.allInstancesByTypeName[typeName];
    }

    public static detect(digit: string): NumeralSystem | undefined {
        return NumeralSystem.allDigitsToNumeralSystem[digit];
    }

    public static toNumber(digit: string): number | undefined {
        return NumeralSystem.allDigitsToWesternArabic[digit];
    }
}

export class DecimalSeparator implements TypeAndName, NumeralSymbolType {
    private static readonly allInstances: DecimalSeparator[] = [];
    private static readonly allInstancesByType: Record<number, DecimalSeparator> = {};
    private static readonly allInstancesByTypeName: Record<string, DecimalSeparator> = {};

    public static readonly LanguageDefaultType: number = 0;
    public static readonly Dot = new DecimalSeparator(1, 'Dot', '.');
    public static readonly Comma = new DecimalSeparator(2, 'Comma', ',');

    public static readonly Default = DecimalSeparator.Dot;

    public readonly type: number;
    public readonly name: string;
    public readonly symbol: string;

    private constructor(type: number, name: string, symbol: string) {
        this.type = type;
        this.name = name;
        this.symbol = symbol;

        DecimalSeparator.allInstances.push(this);
        DecimalSeparator.allInstancesByType[type] = this;
        DecimalSeparator.allInstancesByTypeName[name] = this;
    }

    public static values(): DecimalSeparator[] {
        return DecimalSeparator.allInstances;
    }

    public static valueOf(type: number): DecimalSeparator | undefined {
        return DecimalSeparator.allInstancesByType[type];
    }

    public static parse(typeName: string): DecimalSeparator | undefined {
        return DecimalSeparator.allInstancesByTypeName[typeName];
    }
}

export class DigitGroupingSymbol implements TypeAndName, NumeralSymbolType {
    private static readonly allInstances: DigitGroupingSymbol[] = [];
    private static readonly allInstancesByType: Record<number, DigitGroupingSymbol> = {};
    private static readonly allInstancesByTypeName: Record<string, DigitGroupingSymbol> = {};

    public static readonly LanguageDefaultType: number = 0;
    public static readonly Dot = new DigitGroupingSymbol(1, 'Dot', '.');
    public static readonly Comma = new DigitGroupingSymbol(2, 'Comma', ',');
    public static readonly Space = new DigitGroupingSymbol(3, 'Space', ' ');
    public static readonly Apostrophe = new DigitGroupingSymbol(4, 'Apostrophe', '\'');

    public static readonly Default = DigitGroupingSymbol.Comma;

    public readonly type: number;
    public readonly name: string;
    public readonly symbol: string;

    private constructor(type: number, name: string, symbol: string) {
        this.type = type;
        this.name = name;
        this.symbol = symbol;

        DigitGroupingSymbol.allInstances.push(this);
        DigitGroupingSymbol.allInstancesByType[type] = this;
        DigitGroupingSymbol.allInstancesByTypeName[name] = this;
    }

    public static values(): DigitGroupingSymbol[] {
        return DigitGroupingSymbol.allInstances;
    }

    public static valueOf(type: number): DigitGroupingSymbol | undefined {
        return DigitGroupingSymbol.allInstancesByType[type];
    }

    public static parse(typeName: string): DigitGroupingSymbol | undefined {
        return DigitGroupingSymbol.allInstancesByTypeName[typeName];
    }
}

export class DigitGroupingType implements TypeAndName {
    private static readonly allInstances: DigitGroupingType[] = [];
    private static readonly allInstancesByType: Record<number, DigitGroupingType> = {};
    private static readonly allInstancesByTypeName: Record<string, DigitGroupingType> = {};

    public static readonly LanguageDefaultType: number = 0;
    public static readonly None = new DigitGroupingType(1, 'None', 'None', false,
        (numericChars: string[]) => {
            return numericChars.join('');
        }
    );
    public static readonly ThousandsSeparator = new DigitGroupingType(2, 'ThousandsSeparator', 'Thousands Separator', true,
        (numericChars: string[], digitGroupingSymbol: string) => {
            if (numericChars.length <= 3) {
                return numericChars.join('');
            }

            let ret = '';

            for (let i = numericChars.length - 1, j = 0; i >= 0; i--, j++) {
                if (j > 0 && j % 3 === 0) {
                    ret = digitGroupingSymbol + ret;
                }

                ret = numericChars[i] + ret;
            }

            return ret;
        }
    );
    public static readonly IndianNumberGrouping = new DigitGroupingType(3, 'IndianNumberGrouping', 'Indian Number Grouping', true,
        (numericChars: string[], digitGroupingSymbol: string) => {
            if (numericChars.length <= 3) {
                return numericChars.join('');
            }

            let ret = '';
            const length = numericChars.length;

            for (let i = length - 1, j = 0; i >= 0; i--, j++) {
                if (j === 3) {
                    ret = digitGroupingSymbol + ret;
                } else if (j > 3 && (j - 3) % 2 === 0) {
                    ret = digitGroupingSymbol + ret;
                }

                ret = numericChars[i] + ret;
            }

            return ret;
        }
    );

    public static readonly Default = DigitGroupingType.ThousandsSeparator;

    public readonly type: number;
    public readonly typeName: string;
    public readonly name: string;
    public readonly enabled: boolean;
    public readonly format: (numericChars: string[], digitGroupingSymbol: string) => string;

    private constructor(type: number, typeName: string, name: string, enabled: boolean, format: (numericChars: string[], digitGroupingSymbol: string) => string) {
        this.type = type;
        this.typeName = typeName;
        this.name = name;
        this.enabled = enabled;
        this.format = format;

        DigitGroupingType.allInstances.push(this);
        DigitGroupingType.allInstancesByType[type] = this;
        DigitGroupingType.allInstancesByTypeName[typeName] = this;
    }

    public static values(): DigitGroupingType[] {
        return DigitGroupingType.allInstances;
    }

    public static valueOf(type: number): DigitGroupingType | undefined {
        return DigitGroupingType.allInstancesByType[type];
    }

    public static parse(typeName: string): DigitGroupingType | undefined {
        return DigitGroupingType.allInstancesByTypeName[typeName];
    }
}
