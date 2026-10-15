import {
    keys,
    keysIfValueEquals
} from '@/core/base.ts';


// eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
export function isFunction(val: unknown): val is Function {
    return typeof(val) === 'function';
}

export function isDefined<T>(val: T | null | undefined): val is T {
    return val !== null && typeof(val) !== 'undefined';
}

export function isObject(val: unknown): val is object {
    return val !== null && typeof(val) === 'object' && !isArray(val);
}

export function isArray<T>(val: unknown): val is T[] {
    if (isFunction(Array.isArray)) {
        return Array.isArray(val);
    }

    return Object.prototype.toString.call(val) === '[object Array]';
}

export function isString(val: unknown): val is string {
    return typeof(val) === 'string';
}

export function isNumber(val: unknown): val is number {
    return typeof(val) === 'number';
}

export function isInteger(val: unknown): val is number {
    return Number.isInteger(val);
}

export function isBoolean(val: unknown): val is boolean {
    return typeof(val) === 'boolean';
}

export function isTextualUUID(val: unknown): val is string {
    if (!isString(val)) {
        return false;
    }

    const parts = val.split('-');

    if (parts.length !== 5) {
        return false;
    }

    const expectedLengths = [8, 4, 4, 4, 12];

    for (let i = 0; i < parts.length; i++) {
        if (parts[i]!.length !== expectedLengths[i]) {
            return false;
        }

        for (let j = 0; j < parts[i]!.length; j++) {
            const char = parts[i]![j]!;
            if (!((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F'))) {
                return false;
            }
        }
    }

    return true;
}

export function isHextualColor(val: unknown): val is string {
    if (!isString(val)) {
        return false;
    }

    const hexColorRegex = /^[0-9a-fA-F]{6}$/;
    return hexColorRegex.test(val);
}

export function isYearMonth(val: unknown): val is string {
    if (!isString(val)) {
        return false;
    }

    const items = val.split('-');

    if (items.length !== 2) {
        return false;
    }

    return !!parseInt(items[0] as string) && !!parseInt(items[1] as string);
}

export function isEquals<T>(obj1: T, obj2: T): boolean {
    if (obj1 === obj2) {
        return true;
    }

    if (isArray(obj1) && isArray(obj2)) {
        const arr1 = obj1;
        const arr2 = obj2;

        if (arr1.length !== arr2.length) {
            return false;
        }

        for (let i = 0; i < arr1.length; i++) {
            if (!isEquals(arr1[i], arr2[i])) {
                return false;
            }
        }

        return true;
    } else if (isObject(obj1) && isObject(obj2)) {
        const keys1 = Object.keys(obj1);
        const keys2 = Object.keys(obj2);

        if (keys1.length !== keys2.length) {
            return false;
        }

        const keyExistsMap2 = new Map<string, boolean>();

        for (let i = 0; i < keys2.length; i++) {
            const key = keys2[i] as string;

            keyExistsMap2.set(key, true);
        }

        for (let i = 0; i < keys1.length; i++) {
            const key = keys1[i] as string;

            if (!keyExistsMap2.get(key)) {
                return false;
            }

            if (!isEquals((obj1 as Record<string, unknown>)[key], (obj2 as Record<string, unknown>)[key])) {
                return false;
            }
        }

        return true;
    } else {
        return obj1 === obj2;
    }
}

export function isYearMonthEquals(val1: string, val2: string): boolean {
    const items1 = val1.split('-');
    const items2 = val2.split('-');

    if (items1.length !== 2 || items2.length !== 2) {
        return false;
    }

    return (!!parseInt(items1[0] as string) && !!parseInt(items1[1] as string)) && (parseInt(items1[0] as string) === parseInt(items2[0] as string)) && (parseInt(items1[1] as string) === parseInt(items2[1] as string));
}

export function isArray1SubsetOfArray2<T>(array1: T[], array2: T[]): boolean {
    if (array1.length > array2.length) {
        return false;
    }

    const array2ValuesMap: Map<T, boolean> = new Map<T, boolean>();

    for (let i = 0; i < array2.length; i++) {
        array2ValuesMap.set(array2[i] as T, true);
    }

    for (let i = 0; i < array1.length; i++) {
        if (!array2ValuesMap.get(array1[i] as T)) {
            return false;
        }
    }

    return true;
}

export function isObjectEmpty(obj: object): boolean {
    if (!obj) {
        return true;
    }

    for (const _ of keys(obj)) {
        return false;
    }

    return true;
}

export function ofObject<T>(object: T): T {
    return object;
}

export function getNumberValue(value: unknown, defaultValue: number): number {
    if (isString(value)) {
        return parseInt(value, 10);
    } else if (isNumber(value)) {
        return value;
    } else {
        return defaultValue;
    }
}

export function normalizeInteger(value: unknown, defaultValue: number, min: number, max: number): number {
    if (!isInteger(value)) {
        return defaultValue;
    }

    return Math.min(max, Math.max(min, Math.round(value)));
}

export function sortNumbersArray(array: number[]): number[] {
    return array.sort(function (num1, num2) {
        return num1 - num2;
    });
}

export function getObjectOwnFieldCount(object: object): number {
    let count = 0;

    if (!object || !isObject(object)) {
        return count;
    }

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    for (const _ of keys(object)) {
        count++;
    }

    return count;
}

export function getObjectOwnFieldWithValueCount(object: object, value: unknown): number {
    let count = 0;

    if (!object || !isObject(object)) {
        return count;
    }

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    for (const _ of keysIfValueEquals(object, value)) {
        count++;
    }

    return count;
}
