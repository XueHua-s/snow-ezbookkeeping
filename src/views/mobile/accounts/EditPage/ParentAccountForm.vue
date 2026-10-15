<script setup lang="ts">
import { ALL_ACCOUNT_COLORS } from '@/consts/color.ts';
import { ACCOUNT_CURRENCY_NOT_SET_VALUE } from '@/consts/currency.ts';
import { ALL_ACCOUNT_ICONS } from '@/consts/icon.ts';
import { TRANSACTION_MAX_AMOUNT } from '@/consts/transaction.ts';
import { AccountCategory} from '@/core/account.ts';
import { getAccountIconType } from '@/lib/icon.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { useEditPage } from '@/views/mobile/accounts/EditPage/useEditPage.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useEditPage>>;

defineProps<{
    loading: Scope['loading'];
    allCurrenciesWithNotSet: Scope['allCurrenciesWithNotSet'];
    getAccountCreditCardCreditLimitDisplayValue: Scope['getAccountCreditCardCreditLimitDisplayValue'];
    getAccountCreditCardStatementDate: Scope['getAccountCreditCardStatementDate'];
    allAvailableMonthDays: Scope['allAvailableMonthDays'];
    editAccountId: Scope['editAccountId'];
}>();

const account = defineModel<Scope['account']>('account', { required: true });
const accountContext = defineModel<Scope['accountContext']>('accountContext', { required: true });
const showCreditCardLimitSheet = defineModel<Scope['showCreditCardLimitSheet']>('showCreditCardLimitSheet', { required: true });

const { tt, getCurrencyName } = useI18n();

</script>

<template>
    <f7-list form strong inset dividers class="margin-vertical" >
        <f7-list-input
            type="text"
            autocomplete="off"
            clear-button
            :label="tt('Account Name')"
            :placeholder="tt('Your account name')"
            v-model:value="account.name"
        ></f7-list-input>

        <f7-list-item class="list-item-with-header-and-title list-item-with-multi-item">
            <template #default>
                <div class="grid grid-cols-2">
                    <div class="list-item-subitem no-chevron">
                        <a class="item-link" href="#" @click="accountContext.showIconSelectionSheet = true">
                            <div class="item-content">
                                <div class="item-inner">
                                    <div class="item-header">
                                        <span>{{ tt('Account Icon') }}</span>
                                    </div>
                                    <div class="item-title">
                                        <div class="list-item-custom-title no-padding">
                                            <ItemIcon :icon-type="getAccountIconType(account.iconType)" :icon-id="account.icon" :color="account.color"></ItemIcon>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </a>

                        <icon-selection-sheet :all-system-icon-infos="ALL_ACCOUNT_ICONS"
                                              :color="account.color"
                                              v-model:show="accountContext.showIconSelectionSheet"
                                              v-model:icon-type="account.iconType"
                                              v-model="account.icon"
                        ></icon-selection-sheet>
                    </div>
                    <div class="list-item-subitem no-chevron">
                        <a class="item-link" href="#" @click="accountContext.showColorSelectionSheet = true">
                            <div class="item-content">
                                <div class="item-inner">
                                    <div class="item-header">
                                        <span>{{ tt('Account Color') }}</span>
                                    </div>
                                    <div class="item-title">
                                        <div class="list-item-custom-title no-padding">
                                            <ItemIcon icon-type="fixed-f7" icon-id="app_fill" :color="account.color"></ItemIcon>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </a>

                        <color-selection-sheet :all-system-color-infos="ALL_ACCOUNT_COLORS"
                                               v-model:show="accountContext.showColorSelectionSheet"
                                               v-model="account.color"
                        ></color-selection-sheet>
                    </div>
                </div>
            </template>
        </f7-list-item>

        <f7-list-item
            class="list-item-with-header-and-title list-item-no-item-after"
            link="#"
            :header="tt('Default Currency')"
            @click="accountContext.showCurrencyPopup = true"
            v-if="account.category === AccountCategory.CreditCard.type"
        >
            <template #title>
                <div class="no-padding no-margin">
                    <span>{{ getCurrencyName(account.currency) }}&nbsp;</span>
                    <small class="smaller" v-if="account.currency !== ACCOUNT_CURRENCY_NOT_SET_VALUE">{{ account.currency }}</small>
                </div>
            </template>
            <list-item-selection-popup value-type="item"
                                       key-field="currencyCode" value-field="currencyCode"
                                       title-field="displayName" after-field="currencyCode"
                                       :title="tt('Currency Name')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Currency')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allCurrenciesWithNotSet"
                                       v-model:show="accountContext.showCurrencyPopup"
                                       v-model="account.currency">
            </list-item-selection-popup>
        </f7-list-item>

        <f7-list-item
            link="#" no-chevron
            class="list-item-with-header-and-title"
            :class="{ 'disabled': account.currency === '' || account.currency === ACCOUNT_CURRENCY_NOT_SET_VALUE }"
            :header="tt('Credit Limit')"
            :title="getAccountCreditCardCreditLimitDisplayValue(account.numericCreditCardLimit, account.currency)"
            v-if="account.category === AccountCategory.CreditCard.type"
            @click="showCreditCardLimitSheet = true"
        >
            <number-pad-sheet :min-value="0"
                              :max-value="TRANSACTION_MAX_AMOUNT"
                              :currency="account.currency"
                              v-model:show="showCreditCardLimitSheet"
                              v-model="account.numericCreditCardLimit"
            ></number-pad-sheet>
        </f7-list-item>

        <f7-list-item
            link="#"
            class="list-item-with-header-and-title list-item-no-item-after"
            :header="tt('Statement Date')"
            :title="getAccountCreditCardStatementDate(account.creditCardStatementDate)"
            v-if="account.category === AccountCategory.CreditCard.type"
            @click="accountContext.showCreditCardStatementDatePopup = true"
        >
            <list-item-selection-popup value-type="item"
                                       key-field="type" value-field="type"
                                       title-field="displayName"
                                       :title="tt('Statement Date')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Statement Date')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allAvailableMonthDays"
                                       v-model:show="accountContext.showCreditCardStatementDatePopup"
                                       v-model="account.creditCardStatementDate">
            </list-item-selection-popup>
        </f7-list-item>

        <f7-list-item :title="tt('Visible')" v-if="editAccountId">
            <template #after>
                <f7-toggle :checked="account.visible" @toggle:change="account.visible = $event"></f7-toggle>
            </template>
        </f7-list-item>

        <f7-list-input
            type="textarea"
            style="height: auto"
            autocomplete="off"
            :label="tt('Description')"
            :placeholder="tt('Your account description (optional)')"
            v-textarea-auto-size
            v-model:value="account.comment"
        ></f7-list-input>
    </f7-list>
</template>
