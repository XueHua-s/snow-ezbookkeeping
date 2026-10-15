import { OAUTH2_PROVIDER_DISPLAY_NAME } from '@/consts/oauth2.ts';
import {
    type NameNumeralValue
} from '@/core/base.ts';
import {
    type Coordinate,
    type CoordinateFormatOptions,
    CoordinateDisplayType
} from '@/core/coordinate.ts';
import {
    DigitGroupingType,
    NumeralSystem
} from '@/core/numeral.ts';
import {
    formatCoordinate
} from '@/lib/coordinate.ts';
import type { useHelpersFormatOptions } from './format-options.ts';
import type { useHelpersPreferences } from './preferences.ts';
import type { useHelpersState } from './state.ts';
import type { useHelpersTranslation } from './translation.ts';

interface Context {
    getCurrentNumeralSystemType: ReturnType<typeof useHelpersPreferences>['getCurrentNumeralSystemType'];
    t: ReturnType<typeof useHelpersState>['t'];
    userStore: ReturnType<typeof useHelpersState>['userStore'];
    getNumberFormatOptions: ReturnType<typeof useHelpersFormatOptions>['getNumberFormatOptions'];
    getServerMultiLanguageConfigContent: ReturnType<typeof useHelpersTranslation>['getServerMultiLanguageConfigContent'];
}

// Owns display for the helpers store.
export function useHelpersDisplay(context: Context) {
    const { getCurrentNumeralSystemType, t, userStore, getNumberFormatOptions, getServerMultiLanguageConfigContent } = context;

    function getTablePageOptions(availableCountPerPage: number[], totalCount: number | undefined, includeAll: boolean, alwaysAllCount: boolean): NameNumeralValue[] {
        const numeralSystem = getCurrentNumeralSystemType();
        const pageOptions: NameNumeralValue[] = [];

        if (!alwaysAllCount && (!totalCount || totalCount < 1)) {
            if (includeAll) {
                pageOptions.push({value: -1, name: t('All')});
            }

            return pageOptions;
        }

        if (!totalCount) {
            totalCount = 0;
        }

        for (const count of availableCountPerPage) {
            if (!alwaysAllCount && (totalCount < count)) {
                break;
            }

            pageOptions.push({ value: count, name: numeralSystem.formatNumber(count) });
        }

        if (includeAll) {
            pageOptions.push({value: -1, name: t('All')});
        }

        return pageOptions;
    }

    function getFormattedCoordinate(value: Coordinate, coordinateDisplayType?: CoordinateDisplayType): string {
        if (!coordinateDisplayType) {
            coordinateDisplayType = CoordinateDisplayType.valueOf(userStore.currentUserCoordinateDisplayType) ?? CoordinateDisplayType.Default;
        }

        const numberFormatOptions = getNumberFormatOptions({
            digitGrouping: DigitGroupingType.None,
            numeralSystem: NumeralSystem.WesternArabicNumerals
        });

        const options: CoordinateFormatOptions = {
            coordinateDisplayType: coordinateDisplayType,
            numberFormatOptions: numberFormatOptions
        };

        return formatCoordinate(value, options);
    }

    function getLocalizedFileEncodingName(encoding: string): string {
        return t(`encoding.${encoding}`);
    }

    function getLocalizedOAuth2ProviderName(oauth2Provider: string, oidcDisplayNames: Record<string, string>): string {
        if (oauth2Provider === 'oidc') {
            const providerDisplayName = getServerMultiLanguageConfigContent(oidcDisplayNames);

            if (providerDisplayName) {
                return providerDisplayName;
            } else {
                return 'Connect ID';
            }
        } else {
            const providerDisplayName = OAUTH2_PROVIDER_DISPLAY_NAME[oauth2Provider];

            if (providerDisplayName) {
                return providerDisplayName;
            } else {
                return 'OAuth 2.0';
            }
        }
    }

    function getLocalizedOAuth2LoginText(oauth2Provider: string, oidcDisplayNames: Record<string, string>): string {
        if (oauth2Provider === 'oidc') {
            const providerDisplayName = getServerMultiLanguageConfigContent(oidcDisplayNames);

            if (providerDisplayName) {
                return t('format.misc.loginWithCustomProvider', { name: providerDisplayName });
            } else {
                return t('Log in with Connect ID');
            }
        } else {
            const providerDisplayName = OAUTH2_PROVIDER_DISPLAY_NAME[oauth2Provider];

            if (providerDisplayName) {
                return t('format.misc.loginWithCustomProvider', { name: providerDisplayName });
            } else {
                return t('Log in with OAuth 2.0');
            }
        }
    }

    return { getTablePageOptions, getFormattedCoordinate, getLocalizedFileEncodingName, getLocalizedOAuth2ProviderName, getLocalizedOAuth2LoginText };
}
