import type { TypeAndDisplayName } from '@/core/base.ts';
import type { DigitGroupingType, NumeralSystem } from './symbols.ts';

export interface BigDecimal {
    isZero(): boolean;
    isFinite(): boolean;
    isPositive(): boolean;
    isNegative(): boolean;
    isPositiveOrZero(): boolean;
    isNegativeOrZero(): boolean;
    isPositiveInfinity(): boolean;
    isNegativeInfinity(): boolean;
    isNaN(): boolean;
    equals(other: BigDecimal | number | undefined): boolean;
    notEquals(other: BigDecimal | number | undefined): boolean;
    compareTo(other: BigDecimal | number): number;
    greaterThan(other: BigDecimal | number): boolean;
    greaterThanOrEqual(other: BigDecimal | number): boolean;
    lessThan(other: BigDecimal | number): boolean;
    lessThanOrEqual(other: BigDecimal | number): boolean;
    between(min: BigDecimal | number, max: BigDecimal | number): boolean;
    add(other: BigDecimal | number): BigDecimal;
    subtract(other: BigDecimal | number): BigDecimal;
    multiply(other: BigDecimal | number): BigDecimal;
    divide(other: BigDecimal | number): BigDecimal;
    negate(): BigDecimal;
    sign(): BigDecimal;
    pow(exponent: BigDecimal | number): BigDecimal;
    sqrt(): BigDecimal;
    log(): BigDecimal;
    exp(): BigDecimal;
    abs(): BigDecimal;
    truncate(): BigDecimal;
    toSafeIntegerNumber(): number;
    toDoubleNumber(): number;
    toString(): string;
}

export type HiddenAmount = '***';

export interface NumberFormatOptions {
    readonly numeralSystem: NumeralSystem;
    readonly digitGrouping: DigitGroupingType;
    readonly digitGroupingSymbol: string;
    readonly decimalSeparator: string;
    readonly decimalNumberCount?: number;
    readonly trimTailZero?: boolean;
}

export interface BigDecimalWithSuffix {
    readonly value: BigDecimal;
    readonly suffix: string;
}

export interface NumeralSymbolType {
    readonly type: number;
    readonly name: string;
    readonly symbol: string;
}

export interface LocalizedNumeralSymbolType extends TypeAndDisplayName {
    readonly type: number;
    readonly symbol: string;
    readonly displayName: string;
}

export interface LocalizedDigitGroupingType extends TypeAndDisplayName {
    readonly type: number;
    readonly enabled: boolean;
    readonly displayName: string;
}
