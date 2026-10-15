import type { LocalizedPresetCategory } from '@/core/category.ts';
import {
    isObject,
    isString
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services,{ type ApiResponsePromise } from '@/lib/services.ts';
import {
    clearCurrentSessionToken,
    clearCurrentTokenAndUserInfo,
    clearWebAuthnConfig,
    getUserAppLockState,
    hasUserAppLockState,
    updateCurrentToken
} from '@/lib/userstate.ts';
import type { AuthResponse, RegisterResponse } from '@/models/auth_response.ts';
import type {
User,
UserLoginRequest,
UserResendVerifyEmailRequest,
UserVerifyEmailResponse
} from '@/models/user.ts';
import type { useIndexSession } from './session.ts';

interface Context {
    settingsStore: ReturnType<typeof useIndexSession>['settingsStore'];
    userStore: ReturnType<typeof useIndexSession>['userStore'];
    resetAllStates: ReturnType<typeof useIndexSession>['resetAllStates'];
}

// Owns authorization for the index store.
export function useIndexAuthorization(context: Context) {
    const { settingsStore, userStore, resetAllStates } = context;

    function authorize(req: UserLoginRequest): Promise<AuthResponse> {
        return new Promise((resolve, reject) => {
            services.authorize(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result || !data.result.token) {
                    reject({ message: 'Unable to log in' });
                    return;
                }

                if (data.result.need2FA) {
                    resolve(data.result);
                    return;
                }

                if (settingsStore.appSettings.applicationLock || hasUserAppLockState()) {
                    const appLockState = getUserAppLockState();

                    if (!appLockState || appLockState.username !== data.result.user?.username) {
                        clearCurrentTokenAndUserInfo(true);
                        settingsStore.setEnableApplicationLock(false);
                        settingsStore.setEnableApplicationLockWebAuthn(false);
                        clearWebAuthnConfig();
                    }
                }

                settingsStore.setApplicationSettingsFromCloudSettings(data.result.applicationCloudSettings);

                updateCurrentToken(data.result.token);

                if (data.result.user && isObject(data.result.user)) {
                    userStore.storeUserBasicInfo(data.result.user);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to login', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to log in' });
                }
            });
        });
    }

    function authorize2FA({ token, passcode, recoveryCode }: { token: string, passcode: string | null, recoveryCode: string | null }): Promise<AuthResponse> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<AuthResponse>;

            if (passcode) {
                promise = services.authorize2FA({
                    passcode: passcode,
                    token: token
                });
            } else if (recoveryCode) {
                promise = services.authorize2FAByBackupCode({
                    recoveryCode: recoveryCode,
                    token: token
                });
            } else {
                reject({ message: 'An error occurred' });
                return;
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result || !data.result.token) {
                    reject({ message: 'Unable to verify' });
                    return;
                }

                if (settingsStore.appSettings.applicationLock || hasUserAppLockState()) {
                    const appLockState = getUserAppLockState();

                    if (!appLockState || appLockState.username !== data.result.user?.username) {
                        clearCurrentTokenAndUserInfo(true);
                        settingsStore.setEnableApplicationLock(false);
                        settingsStore.setEnableApplicationLockWebAuthn(false);
                        clearWebAuthnConfig();
                    }
                }

                settingsStore.setApplicationSettingsFromCloudSettings(data.result.applicationCloudSettings);

                updateCurrentToken(data.result.token);

                if (data.result.user && isObject(data.result.user)) {
                    userStore.storeUserBasicInfo(data.result.user);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to verify 2fa', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to verify' });
                }
            });
        });
    }

    function authorizeOAuth2({ password, passcode, callbackToken }: { password?: string, passcode?: string, callbackToken: string }): Promise<AuthResponse> {
        return new Promise((resolve, reject) => {
            services.authorizeOAuth2({
                password,
                passcode,
                callbackToken
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result || !data.result.token) {
                    reject({ message: 'Unable to log in' });
                    return;
                }

                if (settingsStore.appSettings.applicationLock || hasUserAppLockState()) {
                    const appLockState = getUserAppLockState();

                    if (!appLockState || appLockState.username !== data.result.user?.username) {
                        clearCurrentTokenAndUserInfo(true);
                        settingsStore.setEnableApplicationLock(false);
                        settingsStore.setEnableApplicationLockWebAuthn(false);
                        clearWebAuthnConfig();
                    }
                }

                settingsStore.setApplicationSettingsFromCloudSettings(data.result.applicationCloudSettings);

                updateCurrentToken(data.result.token);

                if (data.result.user && isObject(data.result.user)) {
                    userStore.storeUserBasicInfo(data.result.user);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to login', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to log in' });
                }
            });
        });
    }

    function register({ user, presetCategories }: { user: User, presetCategories?: LocalizedPresetCategory[] }): Promise<RegisterResponse> {
        return new Promise((resolve, reject) => {
            services.register(user.toRegisterRequest(presetCategories)).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to sign up' });
                    return;
                }

                if (settingsStore.appSettings.applicationLock) {
                    settingsStore.setEnableApplicationLock(false);
                    settingsStore.setEnableApplicationLockWebAuthn(false);
                    clearWebAuthnConfig();
                }

                if (data.result.token && isString(data.result.token)) {
                    updateCurrentToken(data.result.token);
                }

                if (data.result.user && isObject(data.result.user)) {
                    userStore.storeUserBasicInfo(data.result.user);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to sign up', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to sign up' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function lock(): void {
        clearCurrentSessionToken();
        resetAllStates(false);
    }

    function logout(): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.logout().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to logout' });
                    return;
                }

                clearCurrentTokenAndUserInfo(true);
                clearWebAuthnConfig();
                resetAllStates(true);

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to log out', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to logout' });
                }
            });
        });
    }

    function forceLogout(): void {
        clearCurrentTokenAndUserInfo(true);
        clearWebAuthnConfig();
        resetAllStates(true);
    }

    function verifyEmail({ token, requestNewToken }: { token: string, requestNewToken: boolean }): Promise<UserVerifyEmailResponse> {
        return new Promise((resolve, reject) => {
            services.verifyEmail({
                token,
                requestNewToken
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to verify email' });
                    return;
                }

                if (data.result.newToken && isString(data.result.newToken)) {
                    updateCurrentToken(data.result.newToken);
                }

                if (data.result.user && isObject(data.result.user)) {
                    userStore.storeUserBasicInfo(data.result.user);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to verify email', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to verify email' });
                }
            });
        });
    }

    function resendVerifyEmailByUnloginUser(req: UserResendVerifyEmailRequest): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.resendVerifyEmailByUnloginUser(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to resend validation email' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to resend verify email', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to resend validation email' });
                }
            });
        });
    }

    return { authorize, authorize2FA, authorizeOAuth2, register, lock, logout, forceLogout, verifyEmail, resendVerifyEmailByUnloginUser };
}
