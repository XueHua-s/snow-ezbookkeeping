import {
    MeridiemIndicator,
    WeekDay
} from '@/core/datetime.ts';
import type { LocaleDefaultSettings } from '@/core/setting.ts';
import {
    TextDirection
} from '@/core/text.ts';
import {
    getBrowserTimezoneName,
    getCurrentUnixTime,
    getTimezoneOffset,
    isPM
} from '@/lib/datetime.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import {
    getSessionCurrentLanguageKey,
    setSessionCurrentLanguageKey,
    setSessionLanguagePreview
} from '@/lib/settings.ts';
import {
    ALL_LANGUAGES
} from '@/locales/index.ts';
import moment from 'moment-timezone';
import type { useHelpersDateOptions } from './date-options.ts';
import type { useHelpersLanguageInfo } from './language-info.ts';
import type { useHelpersLanguage } from './language.ts';
import type { useHelpersState } from './state.ts';
import { type PreviewLanguageMessages, allLanguageActualRtlSetting, parsePreviewMessages, previewedLanguages } from './types.ts';

interface Context {
    getDefaultLanguage: ReturnType<typeof useHelpersLanguage>['getDefaultLanguage'];
    getLanguageInfo: ReturnType<typeof useHelpersLanguage>['getLanguageInfo'];
    locale: ReturnType<typeof useHelpersState>['locale'];
    getAllLongMonthNames: ReturnType<typeof useHelpersDateOptions>['getAllLongMonthNames'];
    getAllShortMonthNames: ReturnType<typeof useHelpersDateOptions>['getAllShortMonthNames'];
    getAllLongWeekdayNames: ReturnType<typeof useHelpersDateOptions>['getAllLongWeekdayNames'];
    getAllShortWeekdayNames: ReturnType<typeof useHelpersDateOptions>['getAllShortWeekdayNames'];
    getAllMinWeekdayNames: ReturnType<typeof useHelpersDateOptions>['getAllMinWeekdayNames'];
    t: ReturnType<typeof useHelpersState>['t'];
    getDefaultCurrency: ReturnType<typeof useHelpersLanguageInfo>['getDefaultCurrency'];
    getDefaultFirstDayOfWeek: ReturnType<typeof useHelpersLanguageInfo>['getDefaultFirstDayOfWeek'];
    setLocaleMessage: ReturnType<typeof useHelpersState>['setLocaleMessage'];
    settingsStore: ReturnType<typeof useHelpersState>['settingsStore'];
}

// Owns initialization for the helpers store.
export function useHelpersInitialization(context: Context) {
    const {
        getDefaultLanguage, getLanguageInfo, locale, getAllLongMonthNames, getAllShortMonthNames, getAllLongWeekdayNames,
        getAllShortWeekdayNames, getAllMinWeekdayNames, t, getDefaultCurrency, getDefaultFirstDayOfWeek, setLocaleMessage,
        settingsStore,
    } = context;

    function setLanguage(languageKey: string | null, force?: boolean): LocaleDefaultSettings | null {
        if (!languageKey) {
            languageKey = getDefaultLanguage();
            logger.info(`No specified language, use browser default language ${languageKey}`);
        }

        const languageInfo = getLanguageInfo(languageKey);

        if (!languageInfo) {
            languageKey = getDefaultLanguage();
            logger.warn(`Not found language ${languageKey}, use browser default language ${languageKey}`);
        }

        if (!force && locale.value === languageKey) {
            logger.info(`Current locale is already ${languageKey}`);
            return null;
        }

        logger.info(`Apply current language to ${languageKey}`);

        locale.value = languageKey;
        moment.updateLocale(languageKey, {
            months: getAllLongMonthNames(),
            monthsShort: getAllShortMonthNames(),
            weekdays: getAllLongWeekdayNames(),
            weekdaysShort: getAllShortWeekdayNames(),
            weekdaysMin: getAllMinWeekdayNames(),
            meridiem: function (hours) {
                if (isPM(hours)) {
                    return t(`datetime.${MeridiemIndicator.PM.name}.content`);
                } else {
                    return t(`datetime.${MeridiemIndicator.AM.name}.content`);
                }
            }
        });

        setSessionCurrentLanguageKey(languageKey);
        services.setLocale(languageKey);
        document.querySelector('html')?.setAttribute('lang', languageKey);

        if (document.querySelector('html')?.getAttribute('data-dir-mode') === 'static') {
            if (languageInfo && languageInfo.textDirection === TextDirection.LTR) {
                if (location.search.includes('rtl')) {
                    const url = new URL(window.location.href);
                    url.search = '';
                    url.hash = '#/';
                    window.location.replace(url.toString());
                }
            } else if (languageInfo && languageInfo.textDirection === TextDirection.RTL) {
                if (!location.search.includes('rtl')) {
                    const url = new URL(window.location.href);
                    url.searchParams.set('rtl', '');
                    url.hash = '#/';
                    window.location.replace(url.toString());
                }
            }
        } else {
            if (languageInfo && languageInfo.textDirection === TextDirection.LTR) {
                document.querySelector('html')?.removeAttribute('dir');
            } else if (languageInfo && languageInfo.textDirection === TextDirection.RTL) {
                document.querySelector('html')?.setAttribute('dir', 'rtl');
            }
        }

        const defaultCurrency = getDefaultCurrency();
        const defaultFirstDayOfWeekName = getDefaultFirstDayOfWeek();
        let defaultFirstDayOfWeek = WeekDay.DefaultFirstDay.type;

        if (WeekDay.parse(defaultFirstDayOfWeekName)) {
            defaultFirstDayOfWeek = (WeekDay.parse(defaultFirstDayOfWeekName) as WeekDay).type;
        }

        const localeDefaultSettings: LocaleDefaultSettings = {
            currency: defaultCurrency,
            firstDayOfWeek: defaultFirstDayOfWeek
        };

        return localeDefaultSettings;
    }

    function hasPreviewedConfiguration(languageKey: string): boolean {
        return !!previewedLanguages[languageKey];
    }

    function previewLanguage(json: string, textDirection: TextDirection): void {
        const messages: PreviewLanguageMessages = parsePreviewMessages(json);
        setSessionLanguagePreview(locale.value, {
            data: json,
            textDirection: textDirection
        });
        setLocaleMessage(locale.value, messages);
        allLanguageActualRtlSetting[locale.value] = textDirection === TextDirection.RTL;
        previewedLanguages[locale.value] = true;
        settingsStore.updateLocalizedDefaultSettings(setLanguage(locale.value, true));
    }

    function restoreLanguagePreview(): void {
        const languageInfo = ALL_LANGUAGES[locale.value];

        if (!languageInfo) {
            return;
        }

        setSessionLanguagePreview(locale.value, null);
        setLocaleMessage(locale.value, JSON.parse(JSON.stringify(languageInfo.content)));
        allLanguageActualRtlSetting[locale.value] = languageInfo.textDirection === 'rtl';
        delete previewedLanguages[locale.value];
        settingsStore.updateLocalizedDefaultSettings(setLanguage(locale.value, true));
    }

    function setTimeZone(timezone: string): void {
        if (timezone) {
            moment.tz.setDefault(timezone);
        } else {
            moment.tz.setDefault();
        }
    }

    function initLocale(lastUserLanguage?: string, timezone?: string): LocaleDefaultSettings | null {
        const sessionLanguageKey: string = getSessionCurrentLanguageKey();
        let localeDefaultSettings: LocaleDefaultSettings | null = null;

        if (lastUserLanguage && getLanguageInfo(lastUserLanguage)) {
            logger.info(`Last user language is ${lastUserLanguage}`);
            localeDefaultSettings = setLanguage(lastUserLanguage, true);
        } else if (sessionLanguageKey && getLanguageInfo(sessionLanguageKey)) {
            logger.info(`Session language is ${sessionLanguageKey}`);
            localeDefaultSettings = setLanguage(sessionLanguageKey, true);
        } else {
            localeDefaultSettings = setLanguage(null, true);
        }

        if (timezone) {
            logger.info(`Current timezone is ${timezone}`);
            setTimeZone(timezone);
        } else {
            logger.info(`No timezone is set, use browser default ${getTimezoneOffset(getCurrentUnixTime())} (${getBrowserTimezoneName()})`);
            setTimeZone('');
        }

        return localeDefaultSettings;
    }

    return { setLanguage, hasPreviewedConfiguration, previewLanguage, restoreLanguagePreview, setTimeZone, initLocale };
}
