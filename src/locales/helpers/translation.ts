import {
    type TypeAndDisplayName,
    type TypeAndName
} from '@/core/base.ts';
import {
    isDefined,
    isString
} from '@/lib/common.ts';
import type { useHelpersLanguageInfo } from './language-info.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
    getCurrentLanguageTag: ReturnType<typeof useHelpersLanguageInfo>['getCurrentLanguageTag'];
}

// Owns translation for the helpers store.
export function useHelpersTranslation(context: Context) {
    const { t, getCurrentLanguageTag } = context;

    // public functions
    function translateIf(text: string | undefined, isTranslate?: boolean): string {
        if (!isDefined(text)) {
            return '';
        }

        if (isTranslate) {
            return t(text);
        }

        return text;
    }

    function joinMultiText(textArray: string[]): string {
        if (!textArray || !textArray.length) {
            return '';
        }

        const separator = t('format.misc.multiTextJoinSeparator');

        return textArray.join(separator);
    }

    function formatRange(start: string, end: string): string {
        return t('format.misc.startEndRange', { start: start, end: end });
    }

    function formatTypeAndNames(...typeAndName: TypeAndName[]): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];

        for (const item of typeAndName) {
            ret.push({
                type: item.type,
                displayName: t(item.name)
            });
        }

        return ret;
    }

    function getServerMultiLanguageConfigContent(multiLanguageConfig: Record<string, string>): string {
        if (!multiLanguageConfig) {
            return '';
        }

        const currentLanguage = getCurrentLanguageTag();

        if (isString(multiLanguageConfig[currentLanguage])) {
            return multiLanguageConfig[currentLanguage];
        }

        return multiLanguageConfig['default'] || '';
    }

    return { translateIf, joinMultiText, formatRange, formatTypeAndNames, getServerMultiLanguageConfigContent };
}
