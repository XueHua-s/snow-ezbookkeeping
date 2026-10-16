import { defineStore } from 'pinia';

import { useSettingsStore } from './setting.ts';
import { useUserStore } from './user.ts';

import type {
    TokenGenerateAPIResponse,
    TokenGenerateMCPResponse,
    TokenRefreshResponse,
    TokenInfoResponse
} from '@/models/token.ts';

import { isObject } from '@/lib/common.ts';
import { getCurrentToken, getCurrentUserInfo } from '@/lib/userstate.ts';

import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';

/** Bootstrap data is usable only while its committed identity/token remains current.
 * Callers recheck at their own continuation before applying locale or notification state.
 */
export interface RefreshedSession extends TokenRefreshResponse {
    isCurrent(): boolean;
}

export const useTokensStore = defineStore('tokens', () => {
    const settingsStore = useSettingsStore();
    const userStore = useUserStore();

    function getAllTokens(): Promise<TokenInfoResponse[]> {
        return new Promise((resolve, reject) => {
            services.getTokens().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve session list' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to load token list', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve session list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    let refreshPromise: Promise<RefreshedSession> | undefined;
    function refreshTokenAndRevokeOldToken(): Promise<RefreshedSession> {
        if (refreshPromise) return refreshPromise;
        const refresh = async () => {
            const owner = getCurrentUserInfo()?.username;
            const response = await services.refreshToken();
            const data = response.data;
            const authorization = response.config.headers?.Authorization;
            const expectedToken = data.result?.newToken;
            if (getCurrentUserInfo()?.username !== owner
                || (expectedToken ? getCurrentToken() !== expectedToken : authorization !== `Bearer ${getCurrentToken()}`)) {
                throw { processed: true };
            }
            const committedToken = getCurrentToken();
            if (data && data.success && data.result) {
                settingsStore.setApplicationSettingsFromCloudSettings(data.result.applicationCloudSettings);
            }
            if (data && data.success && data.result && data.result.user && isObject(data.result.user)) {
                userStore.storeUserBasicInfo(data.result.user);
            }
            if (data.result?.oldTokenId) {
                // Revocation is best-effort after the replacement has been committed.
                // A failed cleanup must not strand callers or become an unhandled rejection.
                void revokeToken({ tokenId: data.result.oldTokenId, ignoreError: true }).catch(() => undefined);
            }
            return { ...data.result, isCurrent: () => getCurrentUserInfo()?.username === owner && getCurrentToken() === committedToken };
        };
        refreshPromise = refresh().finally(() => { refreshPromise = undefined; });
        return refreshPromise;
    }

    function generateToken<T extends 'api' | 'mcp'>({ type, expiresInSeconds, password }: { type: T, expiresInSeconds: number, password: string }): Promise<{ 'api': TokenGenerateAPIResponse, 'mcp': TokenGenerateMCPResponse }[T]> {
        return new Promise((resolve, reject) => {
            let promise = null;

            if (type === 'api') {
                promise = services.generateAPIToken({ expiresInSeconds, password });
            } else if (type === 'mcp') {
                promise = services.generateMCPToken({ expiresInSeconds, password });
            } else {
                reject({ message: 'An error occurred' });
                return;
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to generate token' });
                    return;
                }

                resolve(data.result as { 'api': TokenGenerateAPIResponse, 'mcp': TokenGenerateMCPResponse }[T]);
            }).catch(error => {
                logger.error('failed to generate token', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to generate token' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function revokeToken({ tokenId, ignoreError }: { tokenId: string, ignoreError?: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.revokeToken({ tokenId, ignoreError }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to logout from this session' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to revoke token', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to logout from this session' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function revokeAllTokens(): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.revokeAllTokens().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to logout all other sessions' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to revoke all tokens', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to logout all other sessions' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return {
        // functions
        getAllTokens,
        refreshTokenAndRevokeOldToken,
        generateToken,
        revokeToken,
        revokeAllTokens
    };
});
