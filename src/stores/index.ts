import { defineStore } from 'pinia';
import { useIndexAuthorization } from './index/authorization.ts';
import { useIndexDataManagement } from './index/data-management.ts';
import { useIndexProfile } from './index/profile.ts';
import { useIndexSession } from './index/session.ts';




export const useRootStore = defineStore('root', () => {
    const {
        settingsStore, userStore, accountsStore, transactionCategoriesStore, transactionTagsStore, overviewStore, statisticsStore,
        exchangeRatesStore, userCustomIconsStore, currentNotification, resetAllStates, setNotificationContent,
        generateOAuth2LoginUrl, generateOAuth2LinkUrl,
    } = useIndexSession();
    const { authorize, authorize2FA, authorizeOAuth2, register, lock, logout, forceLogout, verifyEmail, resendVerifyEmailByUnloginUser } = useIndexAuthorization({ settingsStore, userStore, resetAllStates });
    const { requestResetPassword, resetPassword, updateUserProfile, resendVerifyEmailByLoginedUser } = useIndexProfile({ userStore, accountsStore, overviewStore, statisticsStore, exchangeRatesStore });
    const { clearAllUserTransactionsOfAccount, clearAllUserTransactions, clearAllUserData } = useIndexDataManagement({ accountsStore, overviewStore, statisticsStore, transactionCategoriesStore, transactionTagsStore, userCustomIconsStore });

    return {
        // states
        currentNotification,
        // functions
        setNotificationContent,
        generateOAuth2LoginUrl,
        generateOAuth2LinkUrl,
        authorize,
        authorize2FA,
        authorizeOAuth2,
        register,
        lock,
        logout,
        forceLogout,
        verifyEmail,
        resendVerifyEmailByUnloginUser,
        requestResetPassword,
        resetPassword,
        updateUserProfile,
        resendVerifyEmailByLoginedUser,
        clearAllUserTransactionsOfAccount,
        clearAllUserTransactions,
        clearAllUserData
    };
});
