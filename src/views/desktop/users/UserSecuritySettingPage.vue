<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import {
    getOAuth2Provider,
    isAPITokenEnabled,
    isMCPServerEnabled,
    isOAuth2Enabled
} from '@/lib/server_settings.ts';
import {
    mdiRefresh
} from '@mdi/js';
import { VTextField } from 'vuetify/components/VTextField';
import UnlinkThirdPartyLoginDialog from './settings/dialogs/UnlinkThirdPartyLoginDialog.vue';
import UserGenerateTokenDialog from './settings/dialogs/UserGenerateTokenDialog.vue';
import { useUserSecuritySettingPage } from './UserSecuritySettingPage/useUserSecuritySettingPage.ts';

const {
    tt, newPasswordInput, confirmPasswordInput, externalAuths, currentPassword, newPassword, confirmPassword,
    updatingPassword, loadingExternalAuth, loadingSession, loggingInByOAuth2, oauth2LinkUrl, thirdPartyLoginList, sessions,
    updatePassword, reloadExternalAuth, unlinkExternalAuth, generateToken, reloadSessions, revokeSession, revokeAllSessions,
} = useUserSecuritySettingPage();

</script>

<template>
    <v-row>
        <v-col cols="12">
            <v-card :class="{ 'disabled': updatingPassword }" :title="tt('Modify Password')">
                <v-form>
                    <v-card-text class="pt-2">
                        <span class="text-body-large">{{ tt('After changing the password, other devices will be logged out. Please use the new password to log in on other devices.') }}</span>
                    </v-card-text>

                    <v-divider/>

                    <v-card-text class="mt-1">
                        <v-row>
                            <v-col cols="12" md="6">
                                <v-text-field
                                    autocomplete="current-password"
                                    ref="currentPasswordInput"
                                    type="password"
                                    :disabled="updatingPassword"
                                    :label="tt('Current Password')"
                                    :placeholder="tt('Current Password')"
                                    v-model="currentPassword"
                                    @keyup.enter="newPasswordInput?.focus()"
                                />
                            </v-col>
                        </v-row>

                        <v-row>
                            <v-col cols="12" md="6">
                                <v-text-field
                                    autocomplete="new-password"
                                    ref="newPasswordInput"
                                    type="password"
                                    :disabled="updatingPassword"
                                    :label="tt('New Password')"
                                    :placeholder="tt('New Password')"
                                    v-model="newPassword"
                                    @keyup.enter="confirmPasswordInput?.focus()"
                                />
                            </v-col>
                        </v-row>

                        <v-row>
                            <v-col cols="12" md="6">
                                <v-text-field
                                    ref="confirmPasswordInput"
                                    type="password"
                                    :disabled="updatingPassword"
                                    :label="tt('Confirm Password')"
                                    :placeholder="tt('Re-enter the password')"
                                    v-model="confirmPassword"
                                    @keyup.enter="updatePassword"
                                />
                            </v-col>
                        </v-row>
                    </v-card-text>

                    <v-divider/>

                    <v-card-text class="d-flex flex-wrap gap-4">
                        <v-btn :disabled="!newPassword || !confirmPassword || updatingPassword" @click="updatePassword">
                            {{ tt('Save Changes') }}
                            <v-progress-circular indeterminate size="22" class="ms-2" v-if="updatingPassword"></v-progress-circular>
                        </v-btn>
                    </v-card-text>
                </v-form>
            </v-card>
        </v-col>

        <v-col cols="12" v-if="isOAuth2Enabled() && (loadingExternalAuth || (thirdPartyLoginList && thirdPartyLoginList.length))">
            <v-card :class="{ 'disabled': loadingExternalAuth }">
                <template #title>
                    <div class="d-flex align-center">
                        <span>{{ tt('Third-Party Login') }}</span>
                        <v-btn density="compact" color="default" variant="text" size="24" class="ms-2"
                               :aria-label="tt('Refresh')" :icon="true" :loading="loadingExternalAuth" @click="reloadExternalAuth(false)">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                    </div>
                </template>

                <v-table class="table-striped text-no-wrap" :hover="!loadingExternalAuth">
                    <thead>
                    <tr>
                        <th>{{ tt('Type') }}</th>
                        <th>{{ tt('Username') }}</th>
                        <th>{{ tt('Linked Time') }}</th>
                        <th class="text-right">{{ tt('Operation') }}</th>
                    </tr>
                    </thead>
                    <tbody>
                    <tr :key="itemIdx"
                        v-for="itemIdx in (loadingExternalAuth && (!externalAuths || externalAuths.length < 1) ? [ 1 ] : [])">
                        <td class="px-0" colspan="4">
                            <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                        </td>
                    </tr>

                    <tr :key="thirdPartyLogin.externalAuthType"
                        v-for="thirdPartyLogin in thirdPartyLoginList">
                        <td>
                            <v-icon start :icon="thirdPartyLogin.icon"/>
                            {{ thirdPartyLogin.displayName }}
                        </td>
                        <td>{{ thirdPartyLogin.externalUsername }}</td>
                        <td>{{ thirdPartyLogin.createdAt }}</td>
                        <td class=" text-right">
                            <v-btn density="comfortable" variant="tonal"
                                   :disabled="loggingInByOAuth2"
                                   :href="oauth2LinkUrl"
                                   @click="loggingInByOAuth2 = true"
                                   v-if="!thirdPartyLogin.linked && isOAuth2Enabled() && getOAuth2Provider() === thirdPartyLogin.externalAuthType">
                                {{ tt('Link') }}
                                <v-progress-circular indeterminate size="22" class="ms-2" v-if="loggingInByOAuth2"></v-progress-circular>
                            </v-btn>
                            <v-btn density="comfortable" color="error" variant="tonal"
                                   :disabled="loadingExternalAuth"
                                   @click="unlinkExternalAuth(thirdPartyLogin)"
                                   v-if="thirdPartyLogin.linked">
                                {{ tt('Unlink') }}
                            </v-btn>
                        </td>
                    </tr>
                    </tbody>
                </v-table>
            </v-card>
        </v-col>

        <v-col cols="12">
            <v-card :class="{ 'disabled': loadingSession }">
                <template #title>
                    <div class="d-flex align-center">
                        <span>{{ tt('Device & Sessions') }}</span>
                        <v-btn class="ms-3" density="compact" color="default" variant="outlined"
                               @click="generateToken" v-if="isAPITokenEnabled() || isMCPServerEnabled()">{{ tt('Generate Token') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24" class="ms-2"
                               :aria-label="tt('Refresh')" :icon="true" :loading="loadingSession" @click="reloadSessions(false)">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                    </div>
                </template>

                <v-table class="table-striped text-no-wrap" :hover="!loadingSession">
                    <thead>
                    <tr>
                        <th>{{ tt('Type') }}</th>
                        <th>{{ tt('Device Info') }}</th>
                        <th>{{ tt('Last Activity Time') }}</th>
                        <th class="text-right">
                            <v-btn density="comfortable" color="error" variant="tonal"
                                   :disabled="sessions.length < 2 || loadingSession"
                                   @click="revokeAllSessions">
                                {{ tt('Logout All') }}
                            </v-btn>
                        </th>
                    </tr>
                    </thead>
                    <tbody>
                    <tr :key="itemIdx"
                        v-for="itemIdx in (loadingSession && (!sessions || sessions.length < 1) ? [ 1, 2, 3 ] : [])">
                        <td class="px-0" colspan="4">
                            <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                        </td>
                    </tr>

                    <tr :key="session.tokenId"
                        v-for="session in sessions">
                        <td>
                            <v-icon start :icon="session.icon"/>
                            {{ tt(session.deviceName) }}
                        </td>
                        <td>{{ session.deviceInfo }}</td>
                        <td>{{ session.lastSeenDateTime }}</td>
                        <td class="text-right">
                            <v-btn density="comfortable" color="error" variant="tonal"
                                   :disabled="session.isCurrent || loadingSession"
                                   @click="revokeSession(session)">
                                {{ tt('Log Out') }}
                            </v-btn>
                        </td>
                    </tr>
                    </tbody>
                </v-table>
            </v-card>
        </v-col>
    </v-row>

    <unlink-third-party-login-dialog ref="unlinkThirdPartyLoginDialog" />
    <user-generate-token-dialog ref="generateTokenDialog" />
    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>
