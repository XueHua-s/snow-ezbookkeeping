import {
    BASE_API_URL_PATH,
    DEFAULT_API_TIMEOUT
} from '@/consts/api.ts';
import type { ApiResponse } from '@/core/api.ts';
import {
    getBrowserTimezoneName,
    getCurrentUnixTime,
    getTimezoneOffsetMinutes
} from '@/lib/datetime.ts';
import logger from '@/lib/logger.ts';
import {
    getTimeZone
} from '@/lib/settings.ts';
import {
    clearCurrentTokenAndUserInfo,
    getCurrentToken
} from '@/lib/userstate.ts';
import { getBasePath } from '@/lib/web.ts';
import type {
TokenRefreshResponse
} from '@/models/token.ts';
import axios,{ type AxiosRequestConfig,type AxiosRequestHeaders,type AxiosResponse } from 'axios';

export interface ApiRequestConfig extends AxiosRequestConfig {
    readonly headers: AxiosRequestHeaders;
    readonly noAuth?: boolean;
    readonly ignoreBlocked?: boolean;
    readonly ignoreError?: boolean;
    readonly timeout?: number;
    readonly cancelableUuid?: string;
}

export type ApiResponsePromise<T> = Promise<AxiosResponse<ApiResponse<T>>>;

export type { AIAssistantChatStreamCallbacks } from '@/lib/types/assistant.ts';

let needBlockRequest = false;

const blockedRequests: ((token: string | undefined) => void)[] = [];

const cancelableRequests: Record<string, boolean> = {};

export const cancelableStreamingRequests: Record<string, AbortController> = {};

axios.defaults.baseURL = getBasePath() + BASE_API_URL_PATH;

axios.defaults.timeout = DEFAULT_API_TIMEOUT;

axios.interceptors.request.use((config: ApiRequestConfig) => {
    const token = getCurrentToken();

    if (token && !config.noAuth) {
        config.headers.Authorization = `Bearer ${token}`;
    }

    config.headers['X-Timezone-Offset'] = getTimezoneOffsetMinutes(getCurrentUnixTime());

    let timezoneName = getTimeZone();

    if (!timezoneName || timezoneName.trim().length < 1) {
        timezoneName = getBrowserTimezoneName();
    }

    config.headers['X-Timezone-Name'] = timezoneName;

    if (needBlockRequest && !config.ignoreBlocked) {
        return new Promise(resolve => {
            blockedRequests.push(newToken => {
                if (newToken) {
                    config.headers.Authorization = `Bearer ${newToken}`;
                }

                resolve(config);
            });
        });
    }

    return config;
}, error => {
    return Promise.reject(error);
});

axios.interceptors.response.use(response => {
    if ('cancelableUuid' in response.config && response.config.cancelableUuid && cancelableRequests[response.config.cancelableUuid as string]) {
        logger.debug('Response canceled by user request, url: ' + response.config.url + ', cancelableUuid: ' + response.config.cancelableUuid);
        delete cancelableRequests[response.config.cancelableUuid as string];
        return Promise.reject({ canceled: true });
    }

    return response;
}, error => {
    if (error.response?.config && 'cancelableUuid' in error.response.config && error.response.config.cancelableUuid && cancelableRequests[error.response.config.cancelableUuid]) {
        logger.debug('Response canceled by user request, url: ' + error.response.config.url + ', cancelableUuid: ' + error.response.config.cancelableUuid);
        delete cancelableRequests[error.response.config.cancelableUuid];
        return Promise.reject({ canceled: true });
    }

    if (error.response && !error.response.config.ignoreError && error.response.data && error.response.data.errorCode) {
        const errorCode = error.response.data.errorCode;

        if (errorCode === 202001 // unauthorized access
            || errorCode === 202002 // current token is invalid
            || errorCode === 202003 // current token is expired
            || errorCode === 202004 // current token type is invalid
            || errorCode === 202005 // current token requires two-factor authorization
            || errorCode === 202006 // current token does not require two-factor authorization
            || errorCode === 202012 // token is empty
        ) {
            clearCurrentTokenAndUserInfo(false);
            location.reload();
            return Promise.reject({ processed: true });
        }
    }

    return Promise.reject(error);
});

export const transportServices = {
    setLocale: (locale: string) => {
        axios.defaults.headers.common['Accept-Language'] = locale;
    },
    refreshToken: (): ApiResponsePromise<TokenRefreshResponse> => {
        return new Promise((resolve) => {
            needBlockRequest = true;

            axios.post<ApiResponse<TokenRefreshResponse>>('v1/tokens/refresh.json', {}, {
                ignoreBlocked: true
            } as ApiRequestConfig).then(response => {
                const data = response.data;

                resolve(response);
                needBlockRequest = false;

                return data.result.newToken;
            }).then(newToken => {
                blockedRequests.forEach(func => func(newToken));
                blockedRequests.length = 0;
            });
        });
    },
    cancelRequest: (cancelableUuid: string) => {
        if (cancelableStreamingRequests[cancelableUuid]) {
            cancelableStreamingRequests[cancelableUuid].abort();
            delete cancelableStreamingRequests[cancelableUuid];
            return;
        }

        cancelableRequests[cancelableUuid] = true;
    }
};
