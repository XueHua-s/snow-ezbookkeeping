import {
    entries,
    keys
} from '@/core/base.ts';
import {
    type LanguageInfo,
    ALL_LANGUAGES,
    DEFAULT_LANGUAGE
} from '@/locales/index.ts';
import type { useHelpersState } from './state.ts';
import { allLanguageActualRtlSetting, previewedLanguages } from './types.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
}

// Owns language for the helpers store.
export function useHelpersLanguage(context: Context) {
    const { t } = context;

    // private functions
    function getLanguageDisplayName(languageName: string): string {
        return t(`language.${languageName}`);
    }

    function getDefaultLanguage(): string {
        if (!window || !window.navigator) {
            return DEFAULT_LANGUAGE;
        }

        let browserLanguage = window.navigator.browserLanguage || window.navigator.language;

        if (!browserLanguage) {
            return DEFAULT_LANGUAGE;
        }

        // try to match the full browser language tag with full language tag in i18n file
        if (ALL_LANGUAGES[browserLanguage]) {
            return browserLanguage;
        }

        // try to match the full browser language tag with language alias tags in i18n file
        let alternativeLanguage = getLanguageKeyFromLanguageAlias(browserLanguage);

        if (alternativeLanguage && ALL_LANGUAGES[alternativeLanguage]) {
            return alternativeLanguage;
        }

        const languageTagParts = browserLanguage.split('-');

        // maybe browser language is language-script-region format
        if (languageTagParts.length > 2) {
            // fallback to use language tag with language-script / language-region format
            browserLanguage = languageTagParts[0] + '-' + languageTagParts[1];

            // try to match language tag in language-script / language-region format with full language tag in i18n file
            if (ALL_LANGUAGES[browserLanguage]) {
                return browserLanguage;
            }

            // try to match language tag in language-script / language-region format with language alias tags in i18n file
            alternativeLanguage = getLanguageKeyFromLanguageAlias(browserLanguage);

            if (alternativeLanguage && ALL_LANGUAGES[alternativeLanguage]) {
                return alternativeLanguage;
            }
        }

        // fallback to use marco language tag
        if (languageTagParts.length > 1) {
            browserLanguage = languageTagParts[0] as string;

            // try to match marco language tag with full language tag in i18n file
            if (ALL_LANGUAGES[browserLanguage]) {
                return browserLanguage;
            }

            // try to match marco language tag with language alias tags in i18n file
            alternativeLanguage = getLanguageKeyFromLanguageAlias(browserLanguage);

            if (alternativeLanguage && ALL_LANGUAGES[alternativeLanguage]) {
                return alternativeLanguage;
            }
        }

        // fallback to match marco language tag with marco language tag in i18n file
        alternativeLanguage = getLanguageKeyFromMarcoLanguageTag(browserLanguage);

        if (alternativeLanguage && ALL_LANGUAGES[alternativeLanguage]) {
            return alternativeLanguage;
        }

        // fallback to use the default language
        return DEFAULT_LANGUAGE;
    }

    function getLanguageKeyFromLanguageAlias(alias: string): string | null {
        for (const [languageKey, languageInfo] of entries(ALL_LANGUAGES)) {
            if (languageKey.toLowerCase() === alias.toLowerCase()) {
                return languageKey;
            }

            const aliases = languageInfo.aliases;

            if (!aliases || aliases.length < 1) {
                continue;
            }

            for (const aliasName of aliases) {
                if (aliasName.toLowerCase() === alias.toLowerCase()) {
                    return languageKey;
                }
            }
        }

        return null;
    }

    function getLanguageKeyFromMarcoLanguageTag(languageTag: string): string | null {
        for (const languageKey of keys(ALL_LANGUAGES)) {
            if (languageKey.indexOf('-') < 0) {
                continue;
            }

            const marcoLanguageTag = languageKey.split('-')[0] as string;

            if (marcoLanguageTag.toLowerCase() === languageTag.toLowerCase()) {
                return languageKey;
            }
        }

        return null;
    }

    function getLanguageInfo(languageKey: string): LanguageInfo | undefined {
        const languageInfo = ALL_LANGUAGES[languageKey];

        if (languageInfo && previewedLanguages[languageKey]) {
            return { ...languageInfo, textDirection: allLanguageActualRtlSetting[languageKey] ? 'rtl' : 'ltr' };
        }

        return languageInfo;
    }

    return { getLanguageDisplayName, getDefaultLanguage, getLanguageInfo };
}
