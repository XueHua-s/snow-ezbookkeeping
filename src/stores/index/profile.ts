import {
    isObject,
    isString
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import {
    updateCurrentToken
} from '@/lib/userstate.ts';
import type { ForgetPasswordRequest } from '@/models/forget_password.ts';
import type {
UserProfileUpdateRequest,
UserProfileUpdateResponse
} from '@/models/user.ts';
import type { useIndexSession } from './session.ts';

interface Context {
    userStore: ReturnType<typeof useIndexSession>['userStore'];
    accountsStore: ReturnType<typeof useIndexSession>['accountsStore'];
    overviewStore: ReturnType<typeof useIndexSession>['overviewStore'];
    statisticsStore: ReturnType<typeof useIndexSession>['statisticsStore'];
    exchangeRatesStore: ReturnType<typeof useIndexSession>['exchangeRatesStore'];
}

// Owns profile for the index store.
export function useIndexProfile(context: Context) {
    const { userStore, accountsStore, overviewStore, statisticsStore, exchangeRatesStore } = context;

    function requestResetPassword(req: ForgetPasswordRequest): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.requestResetPassword(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to send password reset email' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to send password reset email', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to send password reset email' });
                }
            });
        });
    }

    function resetPassword({ email, token, password }: { email: string, token: string, password: string }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.resetPassword({
                email,
                token,
                password
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to reset password' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to reset password', error);

                if (error && error.processed) {
                    reject(error);
                } else if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else {
                    reject({ message: 'Unable to reset password' });
                }
            });
        });
    }

    function updateUserProfile(req: UserProfileUpdateRequest): Promise<UserProfileUpdateResponse> {
        const userDefaultCurrency = userStore.currentUserDefaultCurrency;

        return new Promise((resolve, reject) => {
            services.updateProfile(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to update user profile' });
                    return;
                }

                if (data.result.newToken && isString(data.result.newToken)) {
                    updateCurrentToken(data.result.newToken);
                }

                if (data.result.user && isObject(data.result.user)) {
                    userStore.storeUserBasicInfo(data.result.user);
                }

                if (!accountsStore.accountListStateInvalid) {
                    accountsStore.updateAccountListInvalidState(true);
                }

                if (!overviewStore.transactionOverviewStateInvalid) {
                    overviewStore.updateTransactionOverviewInvalidState(true);
                }

                if (!statisticsStore.transactionStatisticsStateInvalid) {
                    statisticsStore.updateTransactionStatisticsInvalidState(true);
                }

                if (data.result.user && data.result.user.defaultCurrency !== userDefaultCurrency) {
                    exchangeRatesStore.resetLatestExchangeRates();
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to save user profile', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to update user profile' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function resendVerifyEmailByLoginedUser(): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.resendVerifyEmailByLoginedUser().then(response => {
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

    return { requestResetPassword, resetPassword, updateUserProfile, resendVerifyEmailByLoginedUser };
}
