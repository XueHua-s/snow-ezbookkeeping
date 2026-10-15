import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { KnownErrorCode } from '@/consts/api.ts';
import { itemAndIndex, reversedItemAndIndex } from '@/core/base.ts';
import { isEquals } from '@/lib/common.ts';
import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import {
    getOIDCCustomDisplayNames,
    isOAuth2Enabled
} from '@/lib/server_settings.ts';
import { parseSessionInfo } from '@/lib/session.ts';
import { useI18n } from '@/locales/helpers.ts';
import { type TokenInfoResponse, SessionDeviceType, SessionInfo } from '@/models/token.ts';
import { type UserExternalAuthInfoResponse } from '@/models/user_external_auth.ts';
import { useRootStore } from '@/stores/index.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTokensStore } from '@/stores/token.ts';
import { useUserExternalAuthStore } from '@/stores/userExternalAuth.ts';
import UnlinkThirdPartyLoginDialog from '@/views/desktop/users/settings/dialogs/UnlinkThirdPartyLoginDialog.vue';
import UserGenerateTokenDialog from '@/views/desktop/users/settings/dialogs/UserGenerateTokenDialog.vue';
import {
    mdiCellphone,
    mdiConsole,
    mdiCreationOutline,
    mdiDevices,
    mdiGithub,
    mdiLinkVariant,
    mdiTablet,
    mdiTelevision,
    mdiWatch
} from '@mdi/js';
import { computed, ref, useTemplateRef } from 'vue';
import { VTextField } from 'vuetify/components/VTextField';

export type UnlinkThirdPartyLoginDialogType = InstanceType<typeof UnlinkThirdPartyLoginDialog>;

export type UserGenerateTokenDialogType = InstanceType<typeof UserGenerateTokenDialog>;

export type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;

export type SnackBarType = InstanceType<typeof SnackBar>;

// Coordinates UserSecuritySettingPage state and effects; called once from the component setup.
export function useUserSecuritySettingPage() {
    class DesktopPageLinkedThirdPartyLogin {
        public readonly externalAuthType: string;
        public readonly icon: string;
        public readonly displayName: string;
        public readonly linked: boolean;
        public readonly externalUsername: string;
        public readonly createdAt: string;

        public constructor(externalAuthInfoResponse: UserExternalAuthInfoResponse) {
            this.externalAuthType = externalAuthInfoResponse.externalAuthType;
            this.linked = externalAuthInfoResponse.linked;
            this.externalUsername = externalAuthInfoResponse.externalUsername ? externalAuthInfoResponse.externalUsername : '-';
            this.createdAt = externalAuthInfoResponse.createdAt ? formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(externalAuthInfoResponse.createdAt)) : '-';

            if (externalAuthInfoResponse.externalAuthCategory === 'oauth2') {
                this.displayName = getLocalizedOAuth2ProviderName(externalAuthInfoResponse.externalAuthType, getOIDCCustomDisplayNames());

                if (externalAuthInfoResponse.externalAuthType === 'github') {
                    this.icon = mdiGithub;
                } else {
                    this.icon = mdiLinkVariant;
                }
            } else {
                this.displayName = externalAuthInfoResponse.externalAuthType;
                this.icon = mdiLinkVariant;
            }
        }
    }

    class DesktopPageSessionInfo extends SessionInfo {
        public readonly icon: string;
        public readonly lastSeenDateTime: string;

        public constructor(sessionInfo: SessionInfo) {
            super(sessionInfo.tokenId, sessionInfo.isCurrent, sessionInfo.deviceType, sessionInfo.deviceInfo, sessionInfo.deviceName, sessionInfo.lastSeen);
            this.icon = this.getTokenIcon(sessionInfo.deviceType);
            this.lastSeenDateTime = sessionInfo.lastSeen ? formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(sessionInfo.lastSeen)) : '-';
        }

        private getTokenIcon(deviceType: SessionDeviceType): string {
            if (deviceType === SessionDeviceType.Phone) {
                return mdiCellphone;
            } else if (deviceType === SessionDeviceType.Wearable) {
                return mdiWatch;
            } else if (deviceType === SessionDeviceType.Tablet) {
                return mdiTablet;
            } else if (deviceType === SessionDeviceType.TV) {
                return mdiTelevision;
            } else if (deviceType === SessionDeviceType.Api) {
                return mdiConsole;
            } else if (deviceType === SessionDeviceType.MCP) {
                return mdiCreationOutline;
            } else {
                return mdiDevices;
            }
        }
    }

    const {
        tt,
        formatDateTimeToLongDateTime,
        getLocalizedOAuth2ProviderName,
        setLanguage
    } = useI18n();

    const rootStore = useRootStore();

    const settingsStore = useSettingsStore();

    const userExternalAuthStore = useUserExternalAuthStore();

    const tokensStore = useTokensStore();

    const newPasswordInput = useTemplateRef<VTextField>('newPasswordInput');

    const confirmPasswordInput = useTemplateRef<VTextField>('confirmPasswordInput');

    const unlinkThirdPartyLoginDialog = useTemplateRef<UnlinkThirdPartyLoginDialogType>('unlinkThirdPartyLoginDialog');

    const generateTokenDialog = useTemplateRef<UserGenerateTokenDialogType>('generateTokenDialog');

    const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const externalAuths = ref<UserExternalAuthInfoResponse[]>([]);

    const tokens = ref<TokenInfoResponse[]>([]);

    const currentPassword = ref<string>('');

    const newPassword = ref<string>('');

    const confirmPassword = ref<string>('');

    const updatingPassword = ref<boolean>(false);

    const loadingExternalAuth = ref<boolean>(true);

    const loadingSession = ref<boolean>(true);

    const loggingInByOAuth2 = ref<boolean>(false);

    const oauth2ClientSessionId = ref<string>(generateRandomUUID());

    const oauth2LinkUrl = computed<string>(() => rootStore.generateOAuth2LinkUrl('desktop', oauth2ClientSessionId.value));

    const thirdPartyLoginList = computed<DesktopPageLinkedThirdPartyLogin[]>(() => {
        const ret: DesktopPageLinkedThirdPartyLogin[] = [];

        if (!externalAuths.value) {
            return ret;
        }

        for (const externalAuth of externalAuths.value) {
            ret.push(new DesktopPageLinkedThirdPartyLogin(externalAuth));
        }

        return ret;
    });

    const sessions = computed<DesktopPageSessionInfo[]>(() => {
        const sessions: DesktopPageSessionInfo[] = [];

        if (!tokens.value) {
            return sessions;
        }

        for (const token of tokens.value) {
            const sessionInfo = parseSessionInfo(token);
            sessions.push(new DesktopPageSessionInfo(sessionInfo));
        }

        return sessions;
    });

    const inputProblemMessage = computed<string | null>(() => {
        if (!newPassword.value && !confirmPassword.value) {
            return 'Nothing has been modified';
        } else if (!newPassword.value && confirmPassword.value) {
            return 'New password cannot be blank';
        } else if (newPassword.value && !confirmPassword.value) {
            return 'Password confirmation cannot be blank';
        } else if (newPassword.value && confirmPassword.value && newPassword.value !== confirmPassword.value) {
            return 'Password and password confirmation do not match';
        } else {
            return null;
        }
    });

    function init(): void {
        reloadExternalAuth(true);
        reloadSessions(true);
    }

    function updatePassword(): void {
        const problemMessage = inputProblemMessage.value;

        if (problemMessage) {
            snackbar.value?.showMessage(problemMessage);
            return;
        }

        updatingPassword.value = true;

        rootStore.updateUserProfile({
            password: newPassword.value,
            oldPassword: currentPassword.value
        }).then(response => {
            updatingPassword.value = false;
            currentPassword.value = '';
            newPassword.value = '';
            confirmPassword.value = '';

            if (response.user) {
                const localeDefaultSettings = setLanguage(response.user.language);
                settingsStore.updateLocalizedDefaultSettings(localeDefaultSettings);
            }

            reloadSessions(true);

            snackbar.value?.showMessage('Your profile has been successfully updated');
        }).catch(error => {
            updatingPassword.value = false;
            currentPassword.value = '';

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function reloadExternalAuth(silent?: boolean): void {
        if (!isOAuth2Enabled()) {
            return;
        }

        loadingExternalAuth.value = true;

        userExternalAuthStore.getExternalAuths().then(response => {
            if (!silent) {
                if (isEquals(externalAuths.value, response)) {
                    snackbar.value?.showMessage('Third-party login list is up to date');
                } else {
                    snackbar.value?.showMessage('Third-party login list has been updated');
                }
            }

            externalAuths.value = response;
            loadingExternalAuth.value = false;
        }).catch(error => {
            loadingExternalAuth.value = false;

            if (error.error && error.error.errorCode === KnownErrorCode.ApiNotFound) {
                externalAuths.value = [];
            } else if (!error.processed) {
                externalAuths.value = [];
                snackbar.value?.showError(error);
            }
        });
    }

    function unlinkExternalAuth(thirdPartyLogin: DesktopPageLinkedThirdPartyLogin): void {
        if (!isOAuth2Enabled()) {
            return;
        }

        unlinkThirdPartyLoginDialog.value?.open(thirdPartyLogin.externalAuthType).then(() => {
            reloadExternalAuth(true);
        });
    }

    function generateToken(): void {
        generateTokenDialog.value?.open().then(() => {
            reloadSessions(true);
        });
    }

    function reloadSessions(silent?: boolean): void {
        loadingSession.value = true;

        tokensStore.getAllTokens().then(response => {
            if (!silent) {
                if (isEquals(tokens.value, response)) {
                    snackbar.value?.showMessage('Session list is up to date');
                } else {
                    snackbar.value?.showMessage('Session list has been updated');
                }
            }

            tokens.value = response;
            loadingSession.value = false;
        }).catch(error => {
            loadingSession.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function revokeSession(session: SessionInfo): void {
        confirmDialog.value?.open('Are you sure you want to logout from this session?').then(() => {
            loadingSession.value = true;

            tokensStore.revokeToken({
                tokenId: session.tokenId
            }).then(() => {
                loadingSession.value = false;

                for (const [token, index] of itemAndIndex(tokens.value)) {
                    if (token.tokenId === session.tokenId) {
                        tokens.value.splice(index, 1);
                    }
                }
            }).catch(error => {
                loadingSession.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    function revokeAllSessions(): void {
        if (sessions.value.length < 2) {
            return;
        }

        confirmDialog.value?.open('Are you sure you want to logout all other sessions?').then(() => {
            loadingSession.value = true;

            tokensStore.revokeAllTokens().then(() => {
                loadingSession.value = false;

                for (const [token, index] of reversedItemAndIndex(tokens.value)) {
                    if (!token.isCurrent) {
                        tokens.value.splice(index, 1);
                    }
                }

                snackbar.value?.showMessage('You have logged out all other sessions');
            }).catch(error => {
                loadingSession.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    init();

    return {
        tt, newPasswordInput, confirmPasswordInput, unlinkThirdPartyLoginDialog, generateTokenDialog, confirmDialog, snackbar,
        externalAuths, currentPassword, newPassword, confirmPassword, updatingPassword, loadingExternalAuth, loadingSession,
        loggingInByOAuth2, oauth2LinkUrl, thirdPartyLoginList, sessions, updatePassword, reloadExternalAuth, unlinkExternalAuth,
        generateToken, reloadSessions, revokeSession, revokeAllSessions,
    };
}
