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
    getCurrentToken,
    getCurrentUserInfo,
    updateCurrentToken
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
    /** Bind identity-sensitive requests before Axios schedules its interceptors. */
    readonly expectedToken?: string | null;
    /** Bind assistant payloads to their owner while allowing same-owner refresh. */
    readonly expectedOwner?: string;
}

export type ApiResponsePromise<T> = Promise<AxiosResponse<ApiResponse<T>>>;

export type { AIAssistantChatStreamCallbacks } from '@/lib/types/assistant.ts';

const blockedRequests: ((failure?: unknown) => void)[] = [];
let tokenRefresh: ApiResponsePromise<TokenRefreshResponse> | undefined;

const cancelableRequests: Record<string, boolean> = {};

export const cancelableStreamingRequests: Record<string, AbortController> = {};

axios.defaults.baseURL = getBasePath() + BASE_API_URL_PATH;

axios.defaults.timeout = DEFAULT_API_TIMEOUT;

axios.interceptors.request.use((config: ApiRequestConfig) => {
    const token = getCurrentToken();
    if ('expectedOwner' in config && (!token || config.expectedOwner !== getCurrentUserInfo()?.username)) return Promise.reject({ processed: true });
    if ('expectedToken' in config && config.expectedToken !== token) return Promise.reject({ processed: true });

    if (token && !config.noAuth) {
        config.headers.Authorization = `Bearer ${token}`;
    }

    config.headers['X-Timezone-Offset'] = getTimezoneOffsetMinutes(getCurrentUnixTime());

    let timezoneName = getTimeZone();

    if (!timezoneName || timezoneName.trim().length < 1) {
        timezoneName = getBrowserTimezoneName();
    }

    config.headers['X-Timezone-Name'] = timezoneName;

    if (tokenRefresh && !config.ignoreBlocked) {
        const owner = getCurrentUserInfo()?.username;
        return new Promise((resolve, reject) => {
            blockedRequests.push(failure => {
                if (failure) { reject(failure); return; }
                const current = getCurrentToken();
                if (!current || getCurrentUserInfo()?.username !== owner
                    || ('expectedToken' in config && config.expectedToken !== current)) {
                    reject({ processed: true });
                    return;
                }
                if (!config.noAuth) config.headers.Authorization = `Bearer ${current}`;
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
            // Another tab may already have replaced/revoked the token used by
            // this request. Its late 401 must not log out the current session.
            const authorization = error.response.config.headers?.Authorization;
            const current = getCurrentToken();
            if (authorization && authorization !== `Bearer ${current}`) return Promise.reject(error);
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
        if (tokenRefresh) return tokenRefresh;
        const owner = getCurrentUserInfo()?.username;
        const refresh = async (): ApiResponsePromise<TokenRefreshResponse> => {
            // Only refresh is retryable. Never replay a queued accounting write.
            for (let attempt = 0; attempt < 3; attempt++) {
                const token = getCurrentToken();
                if (!token || getCurrentUserInfo()?.username !== owner) throw { processed: true };
                let response;
                try {
                    response = await axios.post<ApiResponse<TokenRefreshResponse>>('v1/tokens/refresh.json', {}, {
                        ignoreBlocked: true, expectedToken: token
                    } as ApiRequestConfig);
                } catch (error) {
                    if (!getCurrentToken() || getCurrentUserInfo()?.username !== owner) throw { processed: true };
                    if (getCurrentToken() !== token && attempt < 2) continue;
                    throw error;
                }
                if (!getCurrentToken() || getCurrentUserInfo()?.username !== owner) throw { processed: true };
                if (getCurrentToken() !== token) {
                    if (attempt < 2) continue;
                    throw { processed: true };
                }
                if (!response.data.success || !response.data.result) throw response.data;
                if (response.data.result.newToken) updateCurrentToken(response.data.result.newToken);
                return response;
            }
            throw { processed: true };
        };
        const settle = (failure?: unknown) => {
            tokenRefresh = undefined;
            for (const resume of blockedRequests.splice(0)) resume(failure);
        };
        tokenRefresh = refresh().then(response => {
            settle();
            return response;
        }, error => {
            settle(error);
            throw error;
        });
        return tokenRefresh;
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
