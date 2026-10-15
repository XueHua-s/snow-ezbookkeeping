import {
    AMAP_JAVASCRIPT_URL,
    BAIDU_MAP_JAVASCRIPT_URL,
    BASE_AMAP_API_PROXY_URL_PATH,
    BASE_PROXY_URL_PATH,
    BASE_QRCODE_PATH,
    GOOGLE_MAP_JAVASCRIPT_URL
} from '@/consts/api.ts';
import {
    isBoolean
} from '@/lib/common.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import {
    getAmapApplicationKey,
    getBaiduMapAK,
    getGoogleMapAPIKey
} from '@/lib/server_settings.ts';
import {
    getCurrentToken
} from '@/lib/userstate.ts';
import { getBasePath } from '@/lib/web.ts';


export const urlsServices = {
    generateOAuth2LoginUrl: (platform: 'mobile' | 'desktop', clientSessionId: string): string => {
        return `${getBasePath()}/oauth2/login?platform=${platform}&client_session_id=${clientSessionId}`;
    },
    generateOAuth2LinkUrl: (platform: 'mobile' | 'desktop', clientSessionId: string): string => {
        return `${getBasePath()}/oauth2/login?platform=${platform}&client_session_id=${clientSessionId}&token=${getCurrentToken()}`;
    },
    generateQrCodeUrl: (qrCodeName: string): string => {
        return `${getBasePath()}${BASE_QRCODE_PATH}/${qrCodeName}.png`;
    },
    getMapProxyTileImageAndAnnotationImageUrlPatterns(): string[] {
        return [
            `.*${BASE_PROXY_URL_PATH}/map/tile/[^/]+/[^/]+/[^/]+\\.png\\?provider=[^&]+.*$`,
            `.*${BASE_PROXY_URL_PATH}/map/annotation/[^/]+/[^/]+/[^/]+\\.png\\?provider=[^&]+.*$`
        ];
    },
    generateMapProxyTileImageUrl: (mapProvider: string, language: string): string => {
        const token = getCurrentToken();
        let url = `${getBasePath()}${BASE_PROXY_URL_PATH}/map/tile/{z}/{x}/{y}.png?provider=${mapProvider}&token=${token}`;

        if (language) {
            url = url + `&language=${language}`;
        }

        return url;
    },
    generateMapProxyAnnotationImageUrl: (mapProvider: string, language: string): string => {
        const token = getCurrentToken();
        let url = `${getBasePath()}${BASE_PROXY_URL_PATH}/map/annotation/{z}/{x}/{y}.png?provider=${mapProvider}&token=${token}`;

        if (language) {
            url = url + `&language=${language}`;
        }

        return url;
    },
    generateGoogleMapJavascriptUrl: (language: string | undefined, callbackFnName: string): string => {
        let url = `${GOOGLE_MAP_JAVASCRIPT_URL}?key=${getGoogleMapAPIKey()}&libraries=core,marker&callback=${callbackFnName}`;

        if (language) {
            url = url + `&language=${language}`;
        }

        return url;
    },
    generateBaiduMapJavascriptUrl: (callbackFnName: string): string => {
        return `${BAIDU_MAP_JAVASCRIPT_URL}&ak=${getBaiduMapAK()}&callback=${callbackFnName}`;
    },
    generateAmapJavascriptUrl: (callbackFnName: string): string => {
        return `${AMAP_JAVASCRIPT_URL}&key=${getAmapApplicationKey()}&plugin=AMap.ToolBar&callback=${callbackFnName}`;
    },
    generateAmapApiInternalProxyUrl: (): string => {
        return `${window.location.origin}${getBasePath()}${BASE_AMAP_API_PROXY_URL_PATH}`;
    },
    getInternalAvatarUrlWithToken(avatarUrl: string, disableBrowserCache?: boolean | string): string {
        if (!avatarUrl) {
            return avatarUrl;
        }

        const params: string[] = [];
        params.push('token=' + getCurrentToken());

        if (disableBrowserCache) {
            if (isBoolean(disableBrowserCache)) {
                params.push('_nocache=' + generateRandomUUID());
            } else {
                params.push('_nocache=' + disableBrowserCache);
            }
        }

        if (avatarUrl.indexOf('?') >= 0) {
            return avatarUrl + '&' + params.join('&');
        } else {
            return avatarUrl + '?' + params.join('&');
        }
    },
    getUserCustomIconUrlWithToken(iconId: string | number): string {
        return `${getBasePath()}/icons/${iconId}.png?token=${getCurrentToken()}`;
    },
    getTransactionPictureUrlWithToken(pictureUrl: string, disableBrowserCache?: boolean | string): string {
        if (!pictureUrl) {
            return pictureUrl;
        }

        const params: string[] = [];
        params.push('token=' + getCurrentToken());

        if (disableBrowserCache) {
            if (isBoolean(disableBrowserCache)) {
                params.push('_nocache=' + generateRandomUUID());
            } else {
                params.push('_nocache=' + disableBrowserCache);
            }
        }

        if (pictureUrl.indexOf('?') >= 0) {
            return pictureUrl + '&' + params.join('&');
        } else {
            return pictureUrl + '?' + params.join('&');
        }
    }
};
