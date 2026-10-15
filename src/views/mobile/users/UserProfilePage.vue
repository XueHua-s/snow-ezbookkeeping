<script setup lang="ts">
import FormatSettingsSection from './UserProfilePage/FormatSettingsSection.vue';
import { NumeralSystem } from '@/core/numeral.ts';
import { TransactionEditScopeType } from '@/core/transaction.ts';
import { findDisplayNameByType } from '@/lib/common.ts';
import { isUserVerifyEmailEnabled } from '@/lib/server_settings.ts';
import { Account } from '@/models/account.ts';
import type { Router } from 'framework7/types';
import { useUserProfilePage } from './UserProfilePage/useUserProfilePage.ts';

const props = defineProps<{
    f7router: Router.Router;
}>();

const {
    tt, getEnableDisableOption, getCurrencyName, formatFiscalYearStartToGregorianLikeLongMonth, newProfile, emailVerified,
    loading, resending, saving, allAccounts, allVisibleAccounts, allVisibleCategorizedAccounts, allWeekDays,
    allCalendarDisplayTypes, allDateDisplayTypes, allLongDateFormats, allShortDateFormats, allLongTimeFormats,
    allShortTimeFormats, allFiscalYearFormats, allCurrencyDisplayTypes, allNumeralSystemTypes, allDecimalSeparators,
    allDigitGroupingSymbols, allDigitGroupingTypes, allCoordinateDisplayTypes, allExpenseAmountColorTypes,
    allIncomeAmountColorTypes, allTransactionEditScopeTypes, enableDisableOptions, languageTitle, supportDigitGroupingSymbol,
    inputInvalidProblemMessage, langAndRegionInputInvalidProblemMessage, extendInputInvalidProblemMessage, inputIsNotChanged,
    inputIsInvalid, langAndRegionInputIsInvalid, extendInputIsInvalid, currentPassword, currentNoPassword,
    showInputPasswordSheet, showAccountSheet, showEditableTransactionRangePopup, showLanguagePopup, showDefaultCurrencyPopup,
    showFirstDayOfWeekPopup, showFiscalYearStartSheet, showCalendarDisplayTypePopup, showDateDisplayTypePopup,
    showLongDateFormatPopup, showShortDateFormatPopup, showLongTimeFormatPopup, showShortTimeFormatPopup,
    showFiscalYearFormatPopup, showCurrencyDisplayTypePopup, showNumberSystemPopup, showDigitGroupingPopup,
    showDigitGroupingSymbolPopup, showDecimalSeparatorPopup, showCoordinateDisplayTypePopup, showExpenseAmountColorPopup,
    showIncomeAmountColorPopup, showMoreActionSheet, allLanguages, allCurrencies, currentLanguageName, currentDayOfWeekName,
    save, resendVerifyEmail, onPageAfterIn,
} = useUserProfilePage(props);

</script>

<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt('User Profile')"></f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }">
                <f7-link icon-f7="ellipsis" :class="{ 'disabled': !isUserVerifyEmailEnabled() || loading || emailVerified }" :aria-label="tt('More')" @click="showMoreActionSheet = true"></f7-link>
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': inputIsNotChanged || inputIsInvalid || saving }" :aria-label="tt('Save')" @click="save(currentNoPassword)"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="margin-vertical-half skeleton-text" v-if="loading">
            <f7-list-input label="Password" placeholder="Your password"></f7-list-input>
            <f7-list-input label="Confirm Password" placeholder="Re-enter the password"></f7-list-input>
            <f7-list-input label="E-mail" placeholder="Your email address"></f7-list-input>
            <f7-list-input label="Nickname" placeholder="Your nickname"></f7-list-input>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="loading">
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Default Account" title="Unspecified"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Use Last Reconciled Time" title="Disabled"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Editable Transaction Range" title="All" link="#"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="loading">
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Default Language" title="Language" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Default Currency" title="Currency" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="First Day of Week" title="Week Day" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Fiscal Year Start Date" title="January 1"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="loading">
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Calendar Display Type" title="Calendar" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Date Display Type" title="Calendar" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Long Date Format" title="YYYY-MM-DD" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Short Date Format" title="YYYY-MM-DD" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Long Time Format" title="HH:mm:ss" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Short Time Format" title="HH:mm" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Fiscal Year Format" title="FY YYYY" link="#"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="loading">
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Currency Display Mode" title="Currency Symbol" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Digit Grouping" title="Thousands Separator" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Digit Grouping Symbol" title="Comma (,)" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Decimal Separator" title="Dot (.)" link="#"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="loading">
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Geographic Location Format" title="Coordinate Format" link="#"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="loading">
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Expense Amount Color" title="Amount Color" link="#"></f7-list-item>
            <f7-list-item class="list-item-with-header-and-title list-item-no-item-after" header="Income Amount Color" title="Amount Color" link="#"></f7-list-item>
        </f7-list>

        <f7-list form strong inset dividers class="margin-vertical-half" v-if="!loading">
            <f7-list-input
                type="password"
                autocomplete="new-password"
                clear-button
                :label="tt('Password')"
                :placeholder="tt('Your password')"
                v-model:value="newProfile.password"
            ></f7-list-input>

            <f7-list-input
                type="password"
                autocomplete="new-password"
                clear-button
                :label="tt('Confirm Password')"
                :placeholder="tt('Re-enter the password')"
                v-model:value="newProfile.confirmPassword"
            ></f7-list-input>

            <f7-list-input
                type="email"
                autocomplete="email"
                clear-button
                :label="tt('E-mail') + ' ' + (emailVerified ? tt('(Verified)') : tt('(Not Verified)'))"
                :placeholder="tt('Your email address')"
                v-model:value="newProfile.email"
            ></f7-list-input>

            <f7-list-input
                type="text"
                autocomplete="nickname"
                clear-button
                :label="tt('Nickname')"
                :placeholder="tt('Your nickname')"
                v-model:value="newProfile.nickname"
            ></f7-list-input>

            <f7-list-item class="ebk-list-item-error-info" v-if="inputIsInvalid" :footer="tt(inputInvalidProblemMessage || '')"></f7-list-item>
        </f7-list>

        <f7-list form strong inset dividers class="margin-vertical" v-if="!loading">
            <f7-list-item
                class="list-item-with-header-and-title"
                link="#" no-chevron
                :class="{ 'disabled': !allVisibleAccounts.length }"
                :header="tt('Default Account')"
                :title="Account.findAccountNameById(allAccounts, newProfile.defaultAccountId, tt('Unspecified'))"
                @click="showAccountSheet = true"
            >
                <two-column-list-item-selection-sheet primary-key-field="id" primary-value-field="category"
                                                      primary-title-field="name"
                                                      primary-icon-field="icon" primary-icon-type-field="iconType" primary-icon-type="account"
                                                      primary-sub-items-field="accounts"
                                                      :primary-title-i18n="true"
                                                      secondary-key-field="id" secondary-value-field="id"
                                                      secondary-title-field="name"
                                                      secondary-icon-field="icon" secondary-icon-type-field="iconType" secondary-icon-type="account" secondary-color-field="color"
                                                      :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                                      :items="allVisibleCategorizedAccounts"
                                                      v-model:show="showAccountSheet"
                                                      v-model="newProfile.defaultAccountId">
                </two-column-list-item-selection-sheet>
            </f7-list-item>

            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                popover-open=".use-last-reconciled-time-popover-menu"
                :header="tt('Use Last Reconciled Time')"
                :title="getEnableDisableOption(newProfile.useLastReconciledTime)"
            >
                <f7-popover class="use-last-reconciled-time-popover-menu">
                    <f7-list dividers>
                        <f7-list-item link="#" no-chevron popover-close
                                      :title="option.displayName"
                                      :class="{ 'list-item-selected': newProfile.useLastReconciledTime === option.value, 'disabled': !option.value && (TransactionEditScopeType.valueOf(newProfile.transactionEditScope)?.needLastReconciledTime ?? false) }"
                                      :key="option.value"
                                      v-for="option in enableDisableOptions"
                                      @click="newProfile.useLastReconciledTime = option.value">
                            <template #after>
                                <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="newProfile.useLastReconciledTime === option.value"></f7-icon>
                            </template>
                        </f7-list-item>
                    </f7-list>
                </f7-popover>
            </f7-list-item>

            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('Editable Transaction Range')"
                :title="findDisplayNameByType(allTransactionEditScopeTypes, newProfile.transactionEditScope)"
                @click="showEditableTransactionRangePopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('Editable Transaction Range')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Date Range')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allTransactionEditScopeTypes"
                                           v-model:show="showEditableTransactionRangePopup"
                                           v-model="newProfile.transactionEditScope">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item class="ebk-list-item-error-info" v-if="extendInputIsInvalid" :footer="tt(extendInputInvalidProblemMessage || '')"></f7-list-item>
        </f7-list>

        <f7-list form strong inset dividers class="margin-vertical" v-if="!loading">
            <f7-list-item
                class="list-item-with-header-and-title list-item-no-item-after"
                link="#"
                :header="languageTitle"
                :title="currentLanguageName"
                @click="showLanguagePopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="languageTag" value-field="languageTag"
                                           title-field="nativeDisplayName" after-field="displayName"
                                           :title="languageTitle"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Language')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allLanguages"
                                           v-model:show="showLanguagePopup"
                                           v-model="newProfile.language">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item
                class="list-item-with-header-and-title list-item-no-item-after"
                link="#"
                :header="tt('Default Currency')"
                @click="showDefaultCurrencyPopup = true"
            >
                <template #title>
                    <f7-block class="no-padding no-margin">
                        <span>{{ getCurrencyName(newProfile.defaultCurrency) }}&nbsp;</span>
                        <small class="smaller">{{ newProfile.defaultCurrency }}</small>
                    </f7-block>
                </template>
                <list-item-selection-popup value-type="item"
                                           key-field="currencyCode" value-field="currencyCode"
                                           title-field="displayName" after-field="currencyCode"
                                           :title="tt('Default Currency')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Currency')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allCurrencies"
                                           v-model:show="showDefaultCurrencyPopup"
                                           v-model="newProfile.defaultCurrency">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('First Day of Week')"
                :title="currentDayOfWeekName"
                @click="showFirstDayOfWeekPopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('First Day of Week')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Date')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allWeekDays"
                                           v-model:show="showFirstDayOfWeekPopup"
                                           v-model="newProfile.firstDayOfWeek">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item
                link="#" no-chevron
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('Fiscal Year Start Date')"
                :title="formatFiscalYearStartToGregorianLikeLongMonth(newProfile.fiscalYearStart, NumeralSystem.valueOf(newProfile.numeralSystem) ?? NumeralSystem.Default)"
                @click="showFiscalYearStartSheet = true"
            >
                <fiscal-year-start-selection-sheet
                    :numeral-system="newProfile.numeralSystem"
                    v-model:show="showFiscalYearStartSheet"
                    v-model="newProfile.fiscalYearStart">
                </fiscal-year-start-selection-sheet>
            </f7-list-item>

        </f7-list>

        <FormatSettingsSection
    v-if="!loading"
    :loading="loading"
    :allCalendarDisplayTypes="allCalendarDisplayTypes"
    :allDateDisplayTypes="allDateDisplayTypes"
    :allLongDateFormats="allLongDateFormats"
    :allShortDateFormats="allShortDateFormats"
    :allLongTimeFormats="allLongTimeFormats"
    :allShortTimeFormats="allShortTimeFormats"
    :allFiscalYearFormats="allFiscalYearFormats"
    v-model:newProfile="newProfile"
    v-model:showCalendarDisplayTypePopup="showCalendarDisplayTypePopup"
    v-model:showDateDisplayTypePopup="showDateDisplayTypePopup"
    v-model:showLongDateFormatPopup="showLongDateFormatPopup"
    v-model:showShortDateFormatPopup="showShortDateFormatPopup"
    v-model:showLongTimeFormatPopup="showLongTimeFormatPopup"
    v-model:showShortTimeFormatPopup="showShortTimeFormatPopup"
    v-model:showFiscalYearFormatPopup="showFiscalYearFormatPopup"
/>

        <f7-list form strong inset dividers class="margin-vertical" v-if="!loading">
            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('Currency Display Mode')"
                :title="findDisplayNameByType(allCurrencyDisplayTypes, newProfile.currencyDisplayType)"
                @click="showCurrencyDisplayTypePopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('Currency Display Mode')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Currency Display Mode')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allCurrencyDisplayTypes"
                                           v-model:show="showCurrencyDisplayTypePopup"
                                           v-model="newProfile.currencyDisplayType">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('Numeral System')"
                :title="findDisplayNameByType(allNumeralSystemTypes, newProfile.numeralSystem)"
                @click="showNumberSystemPopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('Numeral System')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Numeral System')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allNumeralSystemTypes"
                                           v-model:show="showNumberSystemPopup"
                                           v-model="newProfile.numeralSystem">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('Digit Grouping')"
                :title="findDisplayNameByType(allDigitGroupingTypes, newProfile.digitGrouping)"
                @click="showDigitGroupingPopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('Digit Grouping')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Digit Grouping')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allDigitGroupingTypes"
                                           v-model:show="showDigitGroupingPopup"
                                           v-model="newProfile.digitGrouping">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :disabled="!supportDigitGroupingSymbol"
                :header="tt('Digit Grouping Symbol')"
                :title="findDisplayNameByType(allDigitGroupingSymbols, newProfile.digitGroupingSymbol)"
                @click="showDigitGroupingSymbolPopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('Digit Grouping Symbol')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Digit Grouping Symbol')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allDigitGroupingSymbols"
                                           v-model:show="showDigitGroupingSymbolPopup"
                                           v-model="newProfile.digitGroupingSymbol">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('Decimal Separator')"
                :title="findDisplayNameByType(allDecimalSeparators, newProfile.decimalSeparator)"
                @click="showDecimalSeparatorPopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('Decimal Separator')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Decimal Separator')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allDecimalSeparators"
                                           v-model:show="showDecimalSeparatorPopup"
                                           v-model="newProfile.decimalSeparator">
                </list-item-selection-popup>
            </f7-list-item>
        </f7-list>

        <f7-list form strong inset dividers class="margin-vertical" v-if="!loading">
            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('Geographic Location Format')"
                :title="findDisplayNameByType(allCoordinateDisplayTypes, newProfile.coordinateDisplayType)"
                @click="showCoordinateDisplayTypePopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('Geographic Location Format')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Geographic Location Format')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allCoordinateDisplayTypes"
                                           v-model:show="showCoordinateDisplayTypePopup"
                                           v-model="newProfile.coordinateDisplayType">
                </list-item-selection-popup>
            </f7-list-item>
        </f7-list>

        <f7-list form strong inset dividers class="margin-vertical" v-if="!loading">
            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('Expense Amount Color')"
                :title="findDisplayNameByType(allExpenseAmountColorTypes, newProfile.expenseAmountColor)"
                @click="showExpenseAmountColorPopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('Expense Amount Color')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Color')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allExpenseAmountColorTypes"
                                           v-model:show="showExpenseAmountColorPopup"
                                           v-model="newProfile.expenseAmountColor">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item
                link="#"
                class="list-item-with-header-and-title list-item-no-item-after"
                :header="tt('Income Amount Color')"
                :title="findDisplayNameByType(allIncomeAmountColorTypes, newProfile.incomeAmountColor)"
                @click="showIncomeAmountColorPopup = true"
            >
                <list-item-selection-popup value-type="item"
                                           key-field="type" value-field="type"
                                           title-field="displayName"
                                           :title="tt('Income Amount Color')"
                                           :enable-filter="true"
                                           :filter-placeholder="tt('Color')"
                                           :filter-no-items-text="tt('No results')"
                                           :items="allIncomeAmountColorTypes"
                                           v-model:show="showIncomeAmountColorPopup"
                                           v-model="newProfile.incomeAmountColor">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item class="ebk-list-item-error-info" v-if="langAndRegionInputIsInvalid" :footer="tt(langAndRegionInputInvalidProblemMessage || '')"></f7-list-item>
        </f7-list>

        <f7-actions close-by-outside-click close-on-escape :opened="showMoreActionSheet" @actions:closed="showMoreActionSheet = false">
            <f7-actions-group>
                <f7-actions-button :class="{ 'disabled': loading || resending }" @click="resendVerifyEmail"
                                   v-if="isUserVerifyEmailEnabled() && !loading && !emailVerified"
                >{{ tt('Resend Validation Email') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <password-input-sheet :title="tt('Modify Password')"
                              :hint="tt('Please enter your current password when modifying your password')"
                              :confirm-disabled="saving"
                              :cancel-disabled="saving"
                              v-model:show="showInputPasswordSheet"
                              v-model="currentPassword"
                              @password:confirm="save(true)">
        </password-input-sheet>
    </f7-page>
</template>
