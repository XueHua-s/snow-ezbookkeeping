<script setup lang="ts">
import { findDisplayNameByType } from '@/lib/common.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { useUserProfilePage } from '@/views/mobile/users/UserProfilePage/useUserProfilePage.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useUserProfilePage>>;

defineProps<{
    loading: Scope['loading'];
    allCalendarDisplayTypes: Scope['allCalendarDisplayTypes'];
    allDateDisplayTypes: Scope['allDateDisplayTypes'];
    allLongDateFormats: Scope['allLongDateFormats'];
    allShortDateFormats: Scope['allShortDateFormats'];
    allLongTimeFormats: Scope['allLongTimeFormats'];
    allShortTimeFormats: Scope['allShortTimeFormats'];
    allFiscalYearFormats: Scope['allFiscalYearFormats'];
}>();

const newProfile = defineModel<Scope['newProfile']>('newProfile', { required: true });
const showCalendarDisplayTypePopup = defineModel<Scope['showCalendarDisplayTypePopup']>('showCalendarDisplayTypePopup', { required: true });
const showDateDisplayTypePopup = defineModel<Scope['showDateDisplayTypePopup']>('showDateDisplayTypePopup', { required: true });
const showLongDateFormatPopup = defineModel<Scope['showLongDateFormatPopup']>('showLongDateFormatPopup', { required: true });
const showShortDateFormatPopup = defineModel<Scope['showShortDateFormatPopup']>('showShortDateFormatPopup', { required: true });
const showLongTimeFormatPopup = defineModel<Scope['showLongTimeFormatPopup']>('showLongTimeFormatPopup', { required: true });
const showShortTimeFormatPopup = defineModel<Scope['showShortTimeFormatPopup']>('showShortTimeFormatPopup', { required: true });
const showFiscalYearFormatPopup = defineModel<Scope['showFiscalYearFormatPopup']>('showFiscalYearFormatPopup', { required: true });

const { tt } = useI18n();

</script>

<template>
    <f7-list form strong inset dividers class="margin-vertical" >
        <f7-list-item
            link="#"
            class="list-item-with-header-and-title list-item-no-item-after"
            :header="tt('Calendar Display Type')"
            :title="findDisplayNameByType(allCalendarDisplayTypes, newProfile.calendarDisplayType)"
            @click="showCalendarDisplayTypePopup = true"
        >
            <list-item-selection-popup value-type="item"
                                       key-field="type" value-field="type"
                                       title-field="displayName"
                                       :title="tt('Calendar Display Type')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Calendar Display Type')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allCalendarDisplayTypes"
                                       v-model:show="showCalendarDisplayTypePopup"
                                       v-model="newProfile.calendarDisplayType">
            </list-item-selection-popup>
        </f7-list-item>

        <f7-list-item
            link="#"
            class="list-item-with-header-and-title list-item-no-item-after"
            :header="tt('Date Display Type')"
            :title="findDisplayNameByType(allDateDisplayTypes, newProfile.dateDisplayType)"
            @click="showDateDisplayTypePopup = true"
        >
            <list-item-selection-popup value-type="item"
                                       key-field="type" value-field="type"
                                       title-field="displayName"
                                       :title="tt('Date Display Type')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Date Display Type')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allDateDisplayTypes"
                                       v-model:show="showDateDisplayTypePopup"
                                       v-model="newProfile.dateDisplayType">
            </list-item-selection-popup>
        </f7-list-item>

        <f7-list-item
            link="#"
            class="list-item-with-header-and-title list-item-no-item-after"
            :header="tt('Long Date Format')"
            :title="findDisplayNameByType(allLongDateFormats, newProfile.longDateFormat)"
            @click="showLongDateFormatPopup = true"
        >
            <list-item-selection-popup value-type="item"
                                       key-field="type" value-field="type"
                                       title-field="displayName"
                                       :title="tt('Long Date Format')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Long Date Format')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allLongDateFormats"
                                       v-model:show="showLongDateFormatPopup"
                                       v-model="newProfile.longDateFormat">
            </list-item-selection-popup>
        </f7-list-item>

        <f7-list-item
            link="#"
            class="list-item-with-header-and-title list-item-no-item-after"
            :header="tt('Short Date Format')"
            :title="findDisplayNameByType(allShortDateFormats, newProfile.shortDateFormat)"
            @click="showShortDateFormatPopup = true"
        >
            <list-item-selection-popup value-type="item"
                                       key-field="type" value-field="type"
                                       title-field="displayName"
                                       :title="tt('Short Date Format')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Short Date Format')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allShortDateFormats"
                                       v-model:show="showShortDateFormatPopup"
                                       v-model="newProfile.shortDateFormat">
            </list-item-selection-popup>
        </f7-list-item>

        <f7-list-item
            link="#"
            class="list-item-with-header-and-title list-item-no-item-after"
            :header="tt('Long Time Format')"
            :title="findDisplayNameByType(allLongTimeFormats, newProfile.longTimeFormat)"
            @click="showLongTimeFormatPopup = true"
        >
            <list-item-selection-popup value-type="item"
                                       key-field="type" value-field="type"
                                       title-field="displayName"
                                       :title="tt('Long Time Format')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Long Time Format')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allLongTimeFormats"
                                       v-model:show="showLongTimeFormatPopup"
                                       v-model="newProfile.longTimeFormat">
            </list-item-selection-popup>
        </f7-list-item>

        <f7-list-item
            link="#"
            class="list-item-with-header-and-title list-item-no-item-after"
            :header="tt('Short Time Format')"
            :title="findDisplayNameByType(allShortTimeFormats, newProfile.shortTimeFormat)"
            @click="showShortTimeFormatPopup = true"
        >
            <list-item-selection-popup value-type="item"
                                       key-field="type" value-field="type"
                                       title-field="displayName"
                                       :title="tt('Short Time Format')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Short Time Format')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allShortTimeFormats"
                                       v-model:show="showShortTimeFormatPopup"
                                       v-model="newProfile.shortTimeFormat">
            </list-item-selection-popup>
        </f7-list-item>

        <f7-list-item
            link="#"
            class="list-item-with-header-and-title list-item-no-item-after"
            :header="tt('Fiscal Year Format')"
            :title="findDisplayNameByType(allFiscalYearFormats, newProfile.fiscalYearFormat)"
            @click="showFiscalYearFormatPopup = true"
        >
            <list-item-selection-popup value-type="item"
                                       key-field="type" value-field="type"
                                       title-field="displayName"
                                       :title="tt('Fiscal Year Format')"
                                       :enable-filter="true"
                                       :filter-placeholder="tt('Fiscal Year Format')"
                                       :filter-no-items-text="tt('No results')"
                                       :items="allFiscalYearFormats"
                                       v-model:show="showFiscalYearFormatPopup"
                                       v-model="newProfile.fiscalYearFormat">
            </list-item-selection-popup>
        </f7-list-item>
    </f7-list>
</template>
