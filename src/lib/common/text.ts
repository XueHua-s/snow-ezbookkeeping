

export function replaceAll(value: string, originalValue: string, targetValue: string): string {
    // Escape special characters in originalValue to safely use it in a regex pattern.
    // This ensures that characters like . (dot), * (asterisk), +, ?, etc. are treated literally,
    // rather than as special regex symbols.
    const escapedOriginalValue = originalValue.replace(/([.*+?^=!:${}()|\-/\\])/g, '\\$1');

    return value.replace(new RegExp(escapedOriginalValue, 'g'), targetValue);
}

export function removeAll(value: string, originalValue: string): string {
    return replaceAll(value, originalValue, '');
}

export function limitText(value: string, maxLength: number): string {
    let length = 0;

    for (let i = 0; i < value.length; i++) {
        const c = value.charCodeAt(i);

        if ((c >= 0x0001 && c <= 0x007e) || (0xff60 <= c && c <= 0xff9f)) {
            length++;
        } else {
            length += 2;
        }
    }

    if (length <= maxLength || maxLength <= 3) {
        return value;
    }

    return value.substring(0, maxLength - 3) + '...';
}

export function base64encode(arrayBuffer: ArrayBuffer): string {
    return btoa(String.fromCharCode.apply(null, Array.from(new Uint8Array(arrayBuffer))));
}

export function base64decode(str: string): string {
    if (!str) {
        return '';
    }

    return atob(str);
}

export function arrayBufferToString(arrayBuffer: ArrayBuffer): string {
    return String.fromCharCode.apply(null, Array.from(new Uint8Array(arrayBuffer)));
}

export function stringToArrayBuffer(str: string): ArrayBuffer {
    return Uint8Array.from(str, c => c.charCodeAt(0)).buffer;
}
