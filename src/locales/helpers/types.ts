import {
    entries
} from '@/core/base.ts';
import {
    TextDirection
} from '@/core/text.ts';
import {
    isArray,
    isDefined,
    isObject,
    isString
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import { isLanguagePreviewEnabled } from '@/lib/server_settings.ts';
import {
    getSessionLanguagePreview
} from '@/lib/settings.ts';
import {
    ALL_LANGUAGES,
    DEFAULT_LANGUAGE
} from '@/locales/index.ts';
import { reactive } from 'vue';

export interface PreviewLanguageMessages {
    [key: string]: string | PreviewLanguageMessages;
}

export interface LocalizedErrorParameter {
    readonly key: string;
    readonly localized: boolean;
    readonly value: string;
}

export interface LocalizedError {
    readonly message: string;
    readonly parameters?: LocalizedErrorParameter[];
}

export const allLanguageActualRtlSetting = reactive<Record<string, boolean>>({});

export const previewedLanguages = reactive<Record<string, boolean>>({});

export function parsePreviewMessages(json: string): PreviewLanguageMessages {
    const value: unknown = JSON.parse(json.replace(/^\uFEFF/, ''));

    function validate(value: unknown, path: string): asserts value is PreviewLanguageMessages {
        if (!value || !isObject(value) || isArray(value)) {
            throw new Error(`${path} must be a JSON object containing strings or nested objects.`);
        }

        for (const [key, message] of entries(value)) {
            if (key === '__proto__' || key === 'constructor' || key === 'prototype') {
                throw new Error(`Unsupported message key: ${path}.${key}`);
            }

            if (!isString(message)) {
                validate(message, `${path}.${key}`);
            }
        }
    }

    validate(value, 'Language configuration');
    return value;
}

export function getI18nOptions(): object {
    return {
        legacy: false,
        locale: DEFAULT_LANGUAGE,
        fallbackLocale: DEFAULT_LANGUAGE,
        formatFallbackMessages: true,
        messages: (function () {
            const messages: Record<string, object> = {};

            for (const [languageKey, languageInfo] of entries(ALL_LANGUAGES)) {
                messages[languageKey] = languageInfo.content;

                if (!isLanguagePreviewEnabled()) {
                    continue;
                }

                const preview = getSessionLanguagePreview(languageKey);

                if (preview && preview.data && preview.textDirection) {
                    try {
                        messages[languageKey] = parsePreviewMessages(preview.data);
                        allLanguageActualRtlSetting[languageKey] = preview.textDirection === TextDirection.RTL;
                        previewedLanguages[languageKey] = true;
                    } catch (ex) {
                        logger.warn(`failed to restore language preview for ${languageKey}`, ex);
                    }
                }
            }

            return messages;
        })()
    };
}

export function getRtlLocales(): Record<string, boolean> {
    for (const [languageKey, languageInfo] of entries(ALL_LANGUAGES)) {
        if (!isDefined(allLanguageActualRtlSetting[languageKey])) {
            allLanguageActualRtlSetting[languageKey] = languageInfo.textDirection === 'rtl';
        }
    }

    return allLanguageActualRtlSetting;
}
