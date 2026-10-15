<script setup lang="ts">
import LocaleSettingsSection from './UserBasicSettingPage/LocaleSettingsSection.vue';
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { SUPPORTED_IMAGE_EXTENSIONS } from '@/consts/file.ts';

import { isUserVerifyEmailEnabled } from '@/lib/server_settings.ts';

import {
    mdiAccount,
    mdiAccountEditOutline
} from '@mdi/js';
import { useUserBasicSettingPage } from './UserBasicSettingPage/useUserBasicSettingPage.ts';

const {
    tt, newProfile, oldProfile, emailVerified, loading, resending, saving, allAccounts, allVisibleAccounts,
    allVisibleCategorizedAccounts, allWeekDays, allCalendarDisplayTypes, allDateDisplayTypes, allLongDateFormats,
    allShortDateFormats, allLongTimeFormats, allShortTimeFormats, allFiscalYearFormats, allCurrencyDisplayTypes,
    allNumeralSystemTypes, allDecimalSeparators, allDigitGroupingSymbols, allDigitGroupingTypes, allCoordinateDisplayTypes,
    allExpenseAmountColorTypes, allIncomeAmountColorTypes, allTransactionEditScopeTypes, enableDisableOptions, languageTitle,
    supportDigitGroupingSymbol, inputIsNotChanged, inputIsInvalid, reset, avatarProvider, currentUserAvatar, save,
    updateAvatar, removeAvatar, resendVerifyEmail, showOpenAvatarDialog,
} = useUserBasicSettingPage();

</script>

<template>
    <v-row>
        <v-col cols="12">
            <v-card class="setting-items" :class="{ 'disabled': loading || saving }">
                <template #title>
                    <span>{{ tt('Basic Settings') }}</span>
                    <v-progress-circular indeterminate size="20" class="ms-3" v-if="loading"></v-progress-circular>
                </template>

                <v-card-text class="d-flex mt-2">
                    <v-avatar rounded="lg" variant="tonal" size="100" class="me-4 user-profile-avatar-icon"
                              :class="{ 'cursor-pointer': avatarProvider === 'internal', 'user-profile-avatar-icon-modifiable': avatarProvider === 'internal' }"
                              :color="currentUserAvatar ? 'rgba(0,0,0,0)' : 'primary'">
                        <v-img :src="currentUserAvatar" v-if="currentUserAvatar">
                            <template #placeholder>
                                <div class="d-flex align-center justify-center bg-light-primary">
                                    <v-icon color="primary" size="48" class="user-profile-avatar-placeholder" :icon="mdiAccount"/>
                                </div>
                            </template>
                        </v-img>
                        <v-icon size="48" class="user-profile-avatar-placeholder" :icon="mdiAccount" v-else-if="!currentUserAvatar"/>
                        <div class="avatar-edit-icon" v-if="avatarProvider === 'internal'">
                            <v-icon size="48" :icon="mdiAccountEditOutline"/>
                        </div>
                        <v-menu activator="parent" width="200" location="bottom" offset="14px" v-if="avatarProvider === 'internal'">
                            <v-list>
                                <v-list-item :disabled="saving" :title="tt('Update Avatar')" @click="showOpenAvatarDialog"></v-list-item>
                                <v-list-item :disabled="!currentUserAvatar || saving" :title="tt('Remove Avatar')" @click="removeAvatar"></v-list-item>
                            </v-list>
                        </v-menu>
                    </v-avatar>
                    <div class="d-flex flex-column justify-center gap-3">
                        <div class="d-flex text-body-large">
                            <span class="me-1">{{ tt('Username:') }}</span>
                            <v-skeleton-loader class="skeleton-no-margin" type="text" style="width: 100px" :loading="true" v-if="loading"></v-skeleton-loader>
                            <span v-if="!loading">{{ oldProfile.username }}</span>
                        </div>
                        <div class="d-flex text-body-large align-center" style="height: 40px;">
                            <span v-if="!loading && emailVerified">{{ tt('Email address is verified') }}</span>
                            <span v-if="!loading && !emailVerified">{{ tt('Email address is not verified') }}</span>
                            <v-btn class="ms-2 px-2" size="small" variant="text" :disabled="loading || resending"
                                   @click="resendVerifyEmail" v-if="isUserVerifyEmailEnabled() && !loading && !emailVerified">
                                {{ tt('Resend Validation Email') }}
                                <v-progress-circular indeterminate size="18" class="ms-2" v-if="resending"></v-progress-circular>
                            </v-btn>
                            <v-skeleton-loader class="skeleton-no-margin mt-2 mb-1" type="text" style="width: 160px" :loading="true" v-if="loading"></v-skeleton-loader>
                        </div>
                    </div>
                </v-card-text>

                <v-divider />

                <v-form>
                    <LocaleSettingsSection
    :loading="loading"
    :saving="saving"
    :allVisibleAccounts="allVisibleAccounts"
    :allVisibleCategorizedAccounts="allVisibleCategorizedAccounts"
    :allAccounts="allAccounts"
    :enableDisableOptions="enableDisableOptions"
    :allTransactionEditScopeTypes="allTransactionEditScopeTypes"
    v-model:newProfile="newProfile"
/>

                    <v-divider />

                    <v-card-text class="my-1 pa-0 text-body-medium">
                        <div class="setting-item">
                            <span>{{ languageTitle }}</span>
                            <v-spacer/>
                            <language-select class="ms-4" density="compact" max-width="400px"
                                             :disabled="loading || saving"
                                             :placeholder="languageTitle"
                                             :include-system-default="true"
                                             :use-model-value="true" v-model="newProfile.language" />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Default Currency') }}</span>
                            <v-spacer/>
                            <currency-select class="ms-4" density="compact" max-width="400px"
                                             :disabled="loading || saving"
                                             :placeholder="tt('Default Currency')"
                                             v-model="newProfile.defaultCurrency" />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('First Day of Week') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('First Day of Week')"
                                :items="allWeekDays"
                                v-model="newProfile.firstDayOfWeek"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Fiscal Year Start Date') }}</span>
                            <v-spacer/>
                            <fiscal-year-start-select
                                class="ms-4"
                                density="compact"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Fiscal Year Start Date')"
                                :numeral-system="newProfile.numeralSystem"
                                v-model="newProfile.fiscalYearStart"
                            />
                        </div>
                    </v-card-text>

                    <v-divider />

                    <v-card-text class="my-1 pa-0 text-body-medium">
                        <div class="setting-item">
                            <span>{{ tt('Calendar Display Type') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Calendar Display Type')"
                                :items="allCalendarDisplayTypes"
                                v-model="newProfile.calendarDisplayType"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Date Display Type') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Date Display Type')"
                                :items="allDateDisplayTypes"
                                v-model="newProfile.dateDisplayType"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Long Date Format') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Long Date Format')"
                                :items="allLongDateFormats"
                                v-model="newProfile.longDateFormat"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Short Date Format') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Short Date Format')"
                                :items="allShortDateFormats"
                                v-model="newProfile.shortDateFormat"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Long Time Format') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Long Time Format')"
                                :items="allLongTimeFormats"
                                v-model="newProfile.longTimeFormat"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Short Time Format') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Short Time Format')"
                                :items="allShortTimeFormats"
                                v-model="newProfile.shortTimeFormat"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Fiscal Year Format') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Fiscal Year Format')"
                                :items="allFiscalYearFormats"
                                v-model="newProfile.fiscalYearFormat"
                            />
                        </div>
                    </v-card-text>

                    <v-divider />

                    <v-card-text class="my-1 pa-0 text-body-medium">
                        <div class="setting-item">
                            <span>{{ tt('Currency Display Mode') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Currency Display Mode')"
                                :items="allCurrencyDisplayTypes"
                                v-model="newProfile.currencyDisplayType"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Numeral System') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Numeral System')"
                                :items="allNumeralSystemTypes"
                                v-model="newProfile.numeralSystem"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Digit Grouping') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Digit Grouping')"
                                :items="allDigitGroupingTypes"
                                v-model="newProfile.digitGrouping"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Digit Grouping Symbol') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving || !supportDigitGroupingSymbol"
                                :placeholder="tt('Digit Grouping Symbol')"
                                :items="allDigitGroupingSymbols"
                                v-model="newProfile.digitGroupingSymbol"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Decimal Separator') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Decimal Separator')"
                                :items="allDecimalSeparators"
                                v-model="newProfile.decimalSeparator"
                            />
                        </div>
                    </v-card-text>

                    <v-divider />

                    <v-card-text class="my-1 pa-0 text-body-medium">
                        <div class="setting-item">
                            <span>{{ tt('Geographic Location Format') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Geographic Location Format')"
                                :items="allCoordinateDisplayTypes"
                                v-model="newProfile.coordinateDisplayType"
                            />
                        </div>
                    </v-card-text>

                    <v-divider />

                    <v-card-text class="my-1 pa-0 text-body-medium">
                        <div class="setting-item">
                            <span>{{ tt('Expense Amount Color') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Expense Amount Color')"
                                :items="allExpenseAmountColorTypes"
                                v-model="newProfile.expenseAmountColor"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Income Amount Color') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="displayName"
                                item-value="type"
                                persistent-placeholder
                                max-width="400px"
                                :disabled="loading || saving"
                                :placeholder="tt('Income Amount Color')"
                                :items="allIncomeAmountColorTypes"
                                v-model="newProfile.incomeAmountColor"
                            />
                        </div>
                    </v-card-text>

                    <v-divider />

                    <v-card-text class="d-flex flex-wrap gap-4">
                        <v-btn :disabled="inputIsNotChanged || inputIsInvalid || saving" @click="save">
                            {{ tt('Save Changes') }}
                            <v-progress-circular indeterminate size="22" class="ms-2" v-if="saving"></v-progress-circular>
                        </v-btn>

                        <v-btn color="default" variant="tonal" @click="reset">
                            {{ tt('Reset') }}
                        </v-btn>
                    </v-card-text>
                </v-form>
            </v-card>
        </v-col>
    </v-row>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
    <input ref="avatarInput" type="file" style="display: none" :accept="SUPPORTED_IMAGE_EXTENSIONS" @change="updateAvatar($event)" />
</template>



<style>
.user-profile-avatar-icon .avatar-edit-icon {
    display: none;
    position: absolute;
    width: 100% !important;
    height: 100% !important;
    background-color: rgba(0, 0, 0, 0.4);
}

.user-profile-avatar-icon .avatar-edit-icon > i.v-icon {
    background-color: transparent;
    color: rgba(255, 255, 255, 0.8);
}

.user-profile-avatar-icon:hover .avatar-edit-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    vertical-align: middle;
}

.v-theme--dark {
    .user-profile-avatar-icon:hover .avatar-edit-icon {
        background-color: rgba(64, 64, 64, 0.6);
    }
}

.user-profile-avatar-icon-modifiable:hover .user-profile-avatar-placeholder {
    display: none;
}
</style>
