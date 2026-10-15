import type { BigDecimal } from '@/core/numeral.ts';
import { Decimal } from 'decimal.js';


export class DecimalJSBigDecimal implements BigDecimal {
    public static readonly ZERO = new DecimalJSBigDecimal(0);
    public static readonly ONE = new DecimalJSBigDecimal(1);
    public static readonly NEGATIVE_ONE = new DecimalJSBigDecimal(-1);
    public static readonly NaN = new DecimalJSBigDecimal(Number.NaN);
    public static readonly POSITIVE_INFINITY = new DecimalJSBigDecimal(Number.POSITIVE_INFINITY);
    public static readonly NEGATIVE_INFINITY = new DecimalJSBigDecimal(Number.NEGATIVE_INFINITY);

    private readonly value: Decimal;

    private constructor(value: string | number | Decimal) {
        this.value = new Decimal(value);
    }

    public isZero(): boolean {
        return this.value.isZero();
    }

    public isFinite(): boolean {
        return this.value.isFinite();
    }

    public isPositive(): boolean {
        return this.value.greaterThan(0);
    }

    public isNegative(): boolean {
        return this.value.lessThan(0);
    }

    public isPositiveOrZero(): boolean {
        return this.value.greaterThanOrEqualTo(0);
    }

    public isNegativeOrZero(): boolean {
        return this.value.lessThanOrEqualTo(0);
    }

    public isPositiveInfinity(): boolean {
        return this.value.isPositive() && !this.value.isFinite();
    }

    public isNegativeInfinity(): boolean {
        return this.value.isNegative() && !this.value.isFinite();
    }

    public isNaN(): boolean {
        return this.value.isNaN();
    }

    public equals(other: BigDecimal | number | undefined): boolean {
        if (other === undefined) {
            return false;
        } else if (isBigDecimal(other)) {
            return this.value.equals((other as DecimalJSBigDecimal).value);
        } else {
            return this.value.equals(other);
        }
    }

    public notEquals(other: BigDecimal | number | undefined): boolean {
        if (other === undefined) {
            return true;
        } else if (isBigDecimal(other)) {
            return !this.value.equals((other as DecimalJSBigDecimal).value);
        } else {
            return !this.value.equals(other);
        }
    }

    public compareTo(other: BigDecimal | number): number {
        if (isBigDecimal(other)) {
            return this.value.comparedTo((other as DecimalJSBigDecimal).value);
        } else {
            return this.value.comparedTo(other);
        }
    }

    public greaterThan(other: BigDecimal | number): boolean {
        if (isBigDecimal(other)) {
            return this.value.greaterThan((other as DecimalJSBigDecimal).value);
        } else {
            return this.value.greaterThan(other);
        }
    }

    public greaterThanOrEqual(other: BigDecimal | number): boolean {
        if (isBigDecimal(other)) {
            return this.value.greaterThanOrEqualTo((other as DecimalJSBigDecimal).value);
        } else {
            return this.value.greaterThanOrEqualTo(other);
        }
    }

    public lessThan(other: BigDecimal | number): boolean {
        if (isBigDecimal(other)) {
            return this.value.lessThan((other as DecimalJSBigDecimal).value);
        } else {
            return this.value.lessThan(other);
        }
    }

    public lessThanOrEqual(other: BigDecimal | number): boolean {
        if (isBigDecimal(other)) {
            return this.value.lessThanOrEqualTo((other as DecimalJSBigDecimal).value);
        } else {
            return this.value.lessThanOrEqualTo(other);
        }
    }

    public between(min: BigDecimal | number, max: BigDecimal | number): boolean {
        const minValue = isBigDecimal(min) ? (min as DecimalJSBigDecimal).value : min;
        const maxValue = isBigDecimal(max) ? (max as DecimalJSBigDecimal).value : max;
        return this.value.greaterThanOrEqualTo(minValue) && this.value.lessThanOrEqualTo(maxValue);
    }

    public add(other: BigDecimal | number): BigDecimal {
        if (isBigDecimal(other)) {
            return new DecimalJSBigDecimal(this.value.add((other as DecimalJSBigDecimal).value));
        } else {
            return new DecimalJSBigDecimal(this.value.add(other));
        }
    }

    public subtract(other: BigDecimal | number): BigDecimal {
        if (isBigDecimal(other)) {
            return new DecimalJSBigDecimal(this.value.sub((other as DecimalJSBigDecimal).value));
        } else {
            return new DecimalJSBigDecimal(this.value.sub(other));
        }
    }

    public multiply(other: BigDecimal | number): BigDecimal {
        if (isBigDecimal(other)) {
            return new DecimalJSBigDecimal(this.value.times((other as DecimalJSBigDecimal).value));
        } else {
            return new DecimalJSBigDecimal(this.value.times(other));
        }
    }

    public divide(other: BigDecimal | number): BigDecimal {
        if (isBigDecimal(other)) {
            return new DecimalJSBigDecimal(this.value.dividedBy((other as DecimalJSBigDecimal).value));
        } else {
            return new DecimalJSBigDecimal(this.value.dividedBy(other));
        }
    }

    public negate(): BigDecimal {
        return new DecimalJSBigDecimal(this.value.negated());
    }

    public sign(): BigDecimal {
        if (this.value.isZero()) {
            return DecimalJSBigDecimal.ZERO;
        }

        if (this.value.isPositive()) {
            return DecimalJSBigDecimal.ONE;
        } else if (this.value.isNegative()) {
            return DecimalJSBigDecimal.NEGATIVE_ONE;
        } else {
            return DecimalJSBigDecimal.ZERO;
        }
    }

    public pow(exponent: BigDecimal | number): BigDecimal {
        if (isBigDecimal(exponent)) {
            return new DecimalJSBigDecimal(this.value.pow((exponent as DecimalJSBigDecimal).value));
        } else {
            return new DecimalJSBigDecimal(this.value.pow(exponent));
        }
    }

    public sqrt(): BigDecimal {
        return new DecimalJSBigDecimal(this.value.sqrt());
    }

    public log(): BigDecimal {
        return new DecimalJSBigDecimal(this.value.ln());
    }

    public exp(): BigDecimal {
        return new DecimalJSBigDecimal(this.value.exp());
    }

    public abs(): BigDecimal {
        return new DecimalJSBigDecimal(this.value.abs());
    }

    public truncate(): BigDecimal {
        return new DecimalJSBigDecimal(this.value.trunc());
    }

    public toSafeIntegerNumber(): number {
        const num = this.value.toNumber();

        if (!Number.isSafeInteger(num)) {
            throw new Error(`big decimal value \"${num}\" cannot be converted to a safe integer number`);
        }

        return num;
    }

    public toDoubleNumber(): number {
        return this.value.toNumber();
    }

    public toString(): string {
        return this.value.toFixed();
    }

    public static of(value: number): BigDecimal {
        return new DecimalJSBigDecimal(value);
    }

    public static parse(value: string): BigDecimal {
        return new DecimalJSBigDecimal(value);
    }
}

export const BIG_DECIMAL_ZERO: BigDecimal = DecimalJSBigDecimal.ZERO;

export const BIG_DECIMAL_POSITIVE_INFINITY: BigDecimal = DecimalJSBigDecimal.POSITIVE_INFINITY;

export const BIG_DECIMAL_NEGATIVE_INFINITY: BigDecimal = DecimalJSBigDecimal.NEGATIVE_INFINITY;

export function parseBigDecimal(value: string | number): BigDecimal {
    if (typeof value === 'number') {
        return DecimalJSBigDecimal.of(value);
    } else if (typeof value === 'string') {
        return DecimalJSBigDecimal.parse(value);
    } else {
        return DecimalJSBigDecimal.NaN;
    }
}

export function isBigDecimal(val: unknown): val is BigDecimal {
    return val instanceof DecimalJSBigDecimal;
}
