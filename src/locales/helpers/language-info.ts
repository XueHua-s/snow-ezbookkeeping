import {
    entries
} from '@/core/base.ts';
import {
    TextDirection
} from '@/core/text.ts';
import {
    type LanguageInfo,
    type LanguageOption,
    ALL_LANGUAGES
} from '@/locales/index.ts';
import type { useHelpersLanguage } from './language.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    locale: ReturnType<typeof useHelpersState>['locale'];
    getLanguageInfo: ReturnType<typeof useHelpersLanguage>['getLanguageInfo'];
    getDefaultLanguage: ReturnType<typeof useHelpersLanguage>['getDefaultLanguage'];
    getLocaleMessage: ReturnType<typeof useHelpersState>['getLocaleMessage'];
    t: ReturnType<typeof useHelpersState>['t'];
    getLanguageDisplayName: ReturnType<typeof useHelpersLanguage>['getLanguageDisplayName'];
}

// Owns language info for the helpers store.
export function useHelpersLanguageInfo(context: Context) {
    const { locale, getLanguageInfo, getDefaultLanguage, getLocaleMessage, t, getLanguageDisplayName } = context;

    function getCurrentLanguageTag(): string {
        return locale.value;
    }

    function getCurrentLanguageInfo(): LanguageInfo {
        const languageInfo = getLanguageInfo(locale.value);

        if (languageInfo) {
            return languageInfo;
        }

        return getLanguageInfo(getDefaultLanguage()) as LanguageInfo;
    }

    function getCurrentLanguageDisplayName(): string {
        const currentLanguageInfo = getCurrentLanguageInfo();
        return currentLanguageInfo.displayName;
    }

    function getCurrentLanguageTextDirection(): TextDirection {
        const currentLanguageInfo = getCurrentLanguageInfo();

        if (currentLanguageInfo.textDirection === 'rtl') {
            return TextDirection.RTL;
        } else {
            return TextDirection.LTR;
        }
    }

    function getCurrentLanguageMessagesJson(): string {
        return JSON.stringify(getLocaleMessage(locale.value), null, 4);
    }

    function getDefaultCurrency(): string {
        return t('default.currency');
    }

    function getDefaultFirstDayOfWeek(): string {
        return t('default.firstDayOfWeek');
    }

    function getAllLanguageOptions(includeSystemDefault: boolean): LanguageOption[] {
        const ret: LanguageOption[] = [];

        for (const [languageKey, languageInfo] of entries(ALL_LANGUAGES)) {
            const displayName = languageInfo.displayName;
            const languageNameInCurrentLanguage = getLanguageDisplayName(languageInfo.name);

            ret.push({
                languageTag: languageKey,
                displayName: languageNameInCurrentLanguage,
                nativeDisplayName: displayName
            });
        }

        ret.sort(function (lang1, lang2) {
            return lang1.languageTag.localeCompare(lang2.languageTag);
        });

        if (includeSystemDefault) {
            ret.splice(0, 0, {
                languageTag: '',
                displayName: '',
                nativeDisplayName: t('System Default')
            });
        }

        return ret;
    }

    return {
        getCurrentLanguageTag, getCurrentLanguageInfo, getCurrentLanguageDisplayName, getCurrentLanguageTextDirection,
        getCurrentLanguageMessagesJson, getDefaultCurrency, getDefaultFirstDayOfWeek, getAllLanguageOptions,
    };
}
