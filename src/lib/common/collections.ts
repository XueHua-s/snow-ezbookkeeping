import {
    type GenericNameValue,
    type TypeAndDisplayName,
    type TypeAndName,
    entries,
    keys,
    keysIfValueEquals,
    values,
} from '@/core/base.ts';
import { isArray, isDefined, isNumber, isObject, isString } from './predicates.ts';

export function getFirstVisibleItem<T>(items: Record<string, T>[] | Record<string, Record<string, T>>, hiddenField: string): Record<string, T> | null {
    if (isArray(items) && items.length > 0) {
        const arr = items as Record<string, T>[];

        for (const item of arr) {
            if (hiddenField && item[hiddenField]) {
                continue;
            }

            return item;
        }
    } else if (isObject(items)) {
        const obj = items as Record<string, Record<string, T>>;

        for (const item of values(obj)) {
            if (hiddenField && item[hiddenField]) {
                continue;
            }

            return item;
        }
    }

    return null;
}

export function getItemByKeyValue<T>(src: Record<string, T>[] | Record<string, Record<string, T>>, value: T, keyField: string): Record<string, T> | null {
    if (isArray(src)) {
        const arr = src as Record<string, T>[];

        for (const item of arr) {
            if (item[keyField] === value) {
                return item;
            }
        }
    } else if (isObject(src)) {
        const obj = src as Record<string, Record<string, T>>;

        for (const item of values(obj)) {
            if (item[keyField] === value) {
                return item;
            }
        }
    }

    return null;
}

export function findNameByValue<T>(items: GenericNameValue<T>[], value: T): string | null {
    for (const item of items) {
        if (item.value === value) {
            return item.name;
        }
    }

    return null;
}

export function findNameByType(items: TypeAndName[], type: number): string | null {
    for (const item of items) {
        if (item.type === type) {
            return item.name;
        }
    }

    return null;
}

export function findDisplayNameByType(items: TypeAndDisplayName[], type: number): string | null {
    for (const item of items) {
        if (item.type === type) {
            return item.displayName;
        }
    }

    return null;
}

export function getNameByKeyValue<V, N>(src: Record<string, N | V>[] | Record<string, Record<string, N | V>>, value: V, keyField: string | null, nameField: string, defaultName?: N): N | undefined {
    if (isArray(src)) {
        const arr = src as Record<string, N | V>[];

        if (keyField) {
            for (const option of arr) {
                if (option[keyField] === value) {
                    return option[nameField] as N;
                }
            }
        } else if (isNumber(value)) {
            const index = value;

            if (arr[index]) {
                const option = arr[index];

                return option[nameField] as N;
            }
        }
    } else if (isObject(src)) {
        const obj = src as Record<string, Record<string, N | V>>;

        if (keyField) {
            for (const option of values(obj)) {
                if (option[keyField] === value) {
                    return option[nameField] as N;
                }
            }
        } else if (isString(value)) {
            const key = value;

            if (obj[key]) {
                const option = obj[key];

                return option[nameField] as N;
            }
        }
    }

    return defaultName;
}

export function arrayContainsFieldValue<T>(array: Record<string, T>[], fieldName: string, value: T): boolean {
    if (!value || !array || !array.length) {
        return false;
    }

    for (const item of array) {
        if (item[fieldName] === value) {
            return true;
        }
    }

    return false;
}

export function objectFieldToArrayItem(object: object): string[] {
    const ret: string[] = [];

    for (const field of keys(object)) {
        ret.push(field);
    }

    return ret;
}

export function mapObjectToArray<V, R>(object: Record<string | number | symbol, V>, mapFunc: (value: V, key: string | number | symbol, index: number) => R): R[] {
    const ret: R[] = [];
    let index = 0;

    for (const [key, value] of entries(object)) {
        const mappedValue = mapFunc(value, key, index++);

        if (isDefined(mappedValue)) {
            ret.push(mappedValue);
        }
    }

    return ret;
}

export function objectFieldWithValueToArrayItem<T>(object: Record<string, T>, value: T): string[] {
    const ret: string[] = [];

    for (const field of keysIfValueEquals(object, value)) {
        ret.push(field);
    }

    return ret;
}

export function objectValueToArrayItem<T>(object: Record<string, T>): T[] {
    const ret: T[] = [];

    for (const value of values(object)) {
        ret.push(value);
    }

    return ret;
}

export function arrayItemToObjectField<T>(array: string[], value: T): Record<string, T> {
    const ret: Record<string, T> = {};

    for (const item of array) {
        ret[item] = value;
    }

    return ret;
}

export function splitItemsToMap(str: string | undefined | null, separator: string): Record<string, boolean> {
    const ret: Record<string, boolean> = {};

    if (!str) {
        return ret;
    }

    const items = str.split(separator);

    for (const item of items) {
        if (item) {
            ret[item] = true;
        }
    }

    return ret;
}

export function countSplitItems(str: string | undefined | null, separator: string): number {
    if (!str) {
        return 0;
    }

    const items = str.split(separator);
    let count = 0;

    for (const item of items) {
        if (item) {
            count++;
        }
    }

    return count;
}

export function categorizedArrayToPlainArray<T>(object: Record<string, T[]>): T[] {
    const ret: T[] = [];

    for (const array of values(object)) {
        for (const item of array) {
            ret.push(item);
        }
    }

    return ret;
}

export function selectAll(filterItemIds: Record<string, boolean>, allItemsMap: { [key: string]: { id: string } }): void {
    for (const itemId of keys(filterItemIds)) {
        const item = allItemsMap[itemId];

        if (item) {
            filterItemIds[item.id] = false;
        }
    }
}

export function selectNone(filterItemIds: Record<string, boolean>, allItemsMap: { [key: string]: { id: string } }): void {
    for (const itemId of keys(filterItemIds)) {
        const item = allItemsMap[itemId];

        if (item) {
            filterItemIds[item.id] = true;
        }
    }
}

export function selectInvert(filterItemIds: Record<string, boolean>, allItemsMap: { [key: string]: { id: string } }): void {
    for (const itemId of keys(filterItemIds)) {
        const item = allItemsMap[itemId];

        if (item) {
            filterItemIds[item.id] = !filterItemIds[item.id];
        }
    }
}

export function isPrimaryItemHasSecondaryValue(primaryItem: Record<string, Record<string, unknown>[]>, primarySubItemsField: string, secondaryValueField: string | undefined, secondaryHiddenField: string | undefined, secondaryValue: unknown): boolean {
    const secondaryItems = primaryItem[primarySubItemsField];

    if (!secondaryItems || secondaryItems.length < 1) {
        return false;
    }

    for (const secondaryItem of secondaryItems) {
        if (secondaryHiddenField && secondaryItem[secondaryHiddenField]) {
            continue;
        }

        if (secondaryValueField && secondaryItem[secondaryValueField] === secondaryValue) {
            return true;
        } else if (!secondaryValueField && secondaryItem === secondaryValue) {
            return true;
        }
    }

    return false;
}

export function getPrimaryValueBySecondaryValue<T>(items: Record<string, Record<string, T>[]>[] | Record<string, Record<string, Record<string, T>[]>>, primarySubItemsField: string | undefined, primaryValueField: string | undefined, primaryHiddenField: string | undefined, secondaryValueField: string | undefined, secondaryHiddenField: string | undefined, secondaryValue: T): Record<string, T>[] | Record<string, Record<string, T>[]> | null | undefined {
    if (primarySubItemsField) {
        if (isArray(items)) {
            const arr = items as Record<string, Record<string, T>[]>[];

            for (const primaryItem of arr) {
                if (primaryHiddenField && primaryItem[primaryHiddenField]) {
                    continue;
                }

                if (isPrimaryItemHasSecondaryValue(primaryItem, primarySubItemsField, secondaryValueField, secondaryHiddenField, secondaryValue)) {
                    if (primaryValueField) {
                        return primaryItem[primaryValueField];
                    } else {
                        return primaryItem;
                    }
                }
            }
        } else {
            const obj = items as Record<string, Record<string, Record<string, T>[]>>;

            for (const primaryItem of values(obj)) {
                if (primaryHiddenField && primaryItem[primaryHiddenField]) {
                    continue;
                }

                if (isPrimaryItemHasSecondaryValue(primaryItem, primarySubItemsField, secondaryValueField, secondaryHiddenField, secondaryValue)) {
                    if (primaryValueField) {
                        return primaryItem[primaryValueField];
                    } else {
                        return primaryItem;
                    }
                }
            }
        }
    }

    return null;
}

export function arrangeArrayWithNewStartIndex<T>(array: T[], startIndex: number): T[] {
    if (startIndex <= 0 || startIndex >= array.length) {
        return array;
    }

    const newArray: T[] = [];

    for (let i = startIndex; i < array.length; i++) {
        newArray.push(array[i] as T);
    }

    for (let i = 0; i < startIndex; i++) {
        newArray.push(array[i] as T);
    }

    return newArray;
}
