<script setup lang="ts">
import PaginationButtons from '@/components/desktop/PaginationButtons.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { KnownDateTimeFormat } from '@/core/datetime.ts';
import { ImportTransactionColumnType } from '@/core/import_transaction.ts';
import { KnownAmountFormat } from '@/core/numeral.ts';
import { KnownDateTimezoneFormat } from '@/core/timezone.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    getObjectOwnFieldCount
} from '@/lib/common.ts';
import { focusTableScrollContainer } from '@/lib/ui/desktop.ts';
import {
    mdiCheck,
    mdiClose
} from '@mdi/js';
import { useImportTransactionDefineColumnTab } from './ImportTransactionDefineColumnTab/useImportTransactionDefineColumnTab.ts';

const props = defineProps<{
    parsedFileData?: string[][];
    parsedFileColumnSeparator?: string;
    disabled?: boolean;
}>();

const {
    tt, ti, formatNumberToLocalizedNumerals, currentPage, countPerPage, parsedFileDataColumnMapping,
    allImportTransactionColumnTypes, menus, allSeparators, parsedFileLinesHeaders, parsedFileLines,
    parsedFileLinesTablePageOptions, parsedFileAllTransactionTypes, parsedFileValidMappedTransactionTypes,
    parsedFileAutoDetectedTimeFormat, parsedFileAutoDetectedAmountFormat, displayFileCurrentTimezoneFormat,
    displayFileAutoDetectedTimezoneFormat, displayFileAutoDetectedAmountFormat, getParseDataMappedColumnDisplayName,
    toggleDataMappingColumn, generateResult, reset, loadColumnMappingFile, saveColumnMappingFile,
} = useImportTransactionDefineColumnTab(props);

defineExpose({
    menus,
    generateResult,
    reset,
    loadColumnMappingFile,
    saveColumnMappingFile
});

</script>

<template>
    <div class="import-transaction-table-container d-flex flex-column">
        <v-data-table
            fixed-header
            fixed-footer
            height="100%"
            density="compact"
            item-value="index"
            :class="{ 'import-transaction-table': true, 'disabled': !!disabled }"
            :disable-sort="true"
            :headers="parsedFileLinesHeaders"
            :items="parsedFileLines"
            :hover="true"
            :no-data-text="tt('No data to import')"
            v-model:items-per-page="countPerPage"
            v-model:page="currentPage"
            @click="focusTableScrollContainer"
        >
            <template #headers="{ columns }">
                <tr>
                    <th class="text-no-wrap" :key="column.key ?? undefined" v-for="column in columns">
                        <span v-if="!column.key || column.key === 'index'">{{ column.title }}</span>
                        <div class="py-1" v-if="column.key && column.key !== 'index'">
                            <span>{{ getParseDataMappedColumnDisplayName(parseInt(column.key)) }}</span>
                            <br/>
                            <span>({{ column.title }})</span>
                            <v-menu activator="parent" location="bottom" max-height="500">
                                <v-list>
                                    <v-list-item :key="columnType.type"
                                                 :append-icon="parsedFileDataColumnMapping.dataColumnMapping[columnType.type] === parseInt(column.key) ? mdiCheck : undefined"
                                                 v-for="columnType in allImportTransactionColumnTypes"
                                                 @click="toggleDataMappingColumn(parseInt(column.key), columnType.type)">
                                        <v-list-item-title class="cursor-pointer">
                                            {{ columnType.displayName }}
                                        </v-list-item-title>
                                    </v-list-item>
                                </v-list>
                            </v-menu>
                        </div>
                    </th>
                </tr>
            </template>
            <template #bottom>
            </template>
        </v-data-table>

        <div class="import-transaction-table-footer">
            <v-divider />
            <div class="title-and-toolbar d-flex text-body-large align-center text-no-wrap my-1 mx-3" v-if="parsedFileData">
                <v-btn color="secondary" density="compact" variant="outlined"
                       :append-icon="parsedFileDataColumnMapping.includeHeader ? mdiCheck : mdiClose"
                       @click="parsedFileDataColumnMapping.toggleIncludeHeader()">{{ tt('Include Header Line') }}</v-btn>
                <v-btn class="ms-2" color="secondary" density="compact" variant="outlined"
                       :disabled="!parsedFileDataColumnMapping || !parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.TransactionType) || !parsedFileAllTransactionTypes">
                    <span>{{ tt('Transaction Type Mapping') }}</span>
                    <span class="ms-1" v-if="parsedFileDataColumnMapping && parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.TransactionType) && parsedFileAllTransactionTypes">({{ formatNumberToLocalizedNumerals(getObjectOwnFieldCount(parsedFileValidMappedTransactionTypes)) || tt('None') }})</span>
                    <v-menu eager activator="parent" location="bottom" max-height="500"
                            :close-on-content-click="false">
                        <v-list class="pa-0">
                            <v-list-item class="pa-0">
                                <v-table class="transaction-types-popup-menu">
                                    <tbody>
                                    <tr :key="typeName"
                                        v-for="typeName in parsedFileAllTransactionTypes">
                                        <td>{{ typeName }}</td>
                                        <td>
                                            <v-btn-toggle class="toggle-buttons" density="compact" variant="outlined"
                                                          color="default" mandatory="force"
                                                          v-model="parsedFileDataColumnMapping.transactionTypeMapping[typeName]">
                                                <v-btn :value="undefined">{{ tt('None') }}</v-btn>
                                                <v-btn :value="TransactionType.ModifyBalance">{{ tt('Modify Balance') }}</v-btn>
                                                <v-btn :value="TransactionType.Income">{{ tt('Income') }}</v-btn>
                                                <v-btn :value="TransactionType.Expense">{{ tt('Expense') }}</v-btn>
                                                <v-btn :value="TransactionType.Transfer">{{ tt('Transfer') }}</v-btn>
                                            </v-btn-toggle>
                                        </td>
                                    </tr>
                                    </tbody>
                                </v-table>
                            </v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-btn class="ms-2" color="secondary" density="compact" variant="outlined"
                       :disabled="!parsedFileDataColumnMapping || !parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.TransactionTime)">
                    <span>{{ tt('Time Format') }}</span>
                    <span class="ms-1" v-if="parsedFileDataColumnMapping && parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.TransactionTime)">({{ parsedFileDataColumnMapping.timeFormat || parsedFileAutoDetectedTimeFormat || tt('Unknown') }})</span>
                    <v-menu eager activator="parent" location="bottom" max-height="500">
                        <v-list>
                            <v-list-item key="auto"
                                         :append-icon="parsedFileDataColumnMapping.timeFormat === '' ? mdiCheck : undefined"
                                         @click="parsedFileDataColumnMapping.timeFormat = ''">
                                <v-list-item-title class="cursor-pointer">
                                    <span>{{ tt('Auto detect') }}</span>
                                    <span class="ms-1" v-if="parsedFileAutoDetectedTimeFormat">({{ parsedFileAutoDetectedTimeFormat || tt('Unknown') }})</span>
                                </v-list-item-title>
                            </v-list-item>
                            <v-list-item :key="dateTimeFormat.format"
                                         :append-icon="parsedFileDataColumnMapping.timeFormat === dateTimeFormat.format ? mdiCheck : undefined"
                                         v-for="dateTimeFormat in KnownDateTimeFormat.values()"
                                         @click="parsedFileDataColumnMapping.timeFormat = dateTimeFormat.format">
                                <v-list-item-title class="cursor-pointer">
                                    {{ dateTimeFormat.format }}
                                </v-list-item-title>
                            </v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-btn class="ms-2" color="secondary" density="compact" variant="outlined"
                       v-if="parsedFileDataColumnMapping && parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.TransactionTimezone)">
                    <span>{{ tt('Timezone Format') }}</span>
                    <span class="ms-1" v-if="parsedFileDataColumnMapping && parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.TransactionTimezone)">({{ displayFileCurrentTimezoneFormat }})</span>
                    <v-menu eager activator="parent" location="bottom" max-height="500">
                        <v-list>
                            <v-list-item key="auto"
                                         :append-icon="parsedFileDataColumnMapping.timezoneFormat === '' ? mdiCheck : undefined"
                                         @click="parsedFileDataColumnMapping.timezoneFormat = ''">
                                <v-list-item-title class="cursor-pointer">
                                    <span>{{ tt('Auto detect') }}</span>
                                    <span class="ms-1">({{ displayFileAutoDetectedTimezoneFormat }})</span>
                                </v-list-item-title>
                            </v-list-item>
                            <v-list-item :key="timezoneFormat.value"
                                         :append-icon="parsedFileDataColumnMapping.timezoneFormat === timezoneFormat.value ? mdiCheck : undefined"
                                         v-for="timezoneFormat in KnownDateTimezoneFormat.values()"
                                         @click="parsedFileDataColumnMapping.timezoneFormat = timezoneFormat.value">
                                <v-list-item-title class="cursor-pointer">
                                    {{ ti(timezoneFormat.name, timezoneFormat.nameSupportedI18n) }}
                                </v-list-item-title>
                            </v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-btn class="ms-2" color="secondary" density="compact" variant="outlined"
                       :disabled="!parsedFileDataColumnMapping || !parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.Amount)">
                    <span>{{ tt('Amount Format') }}</span>
                    <span class="ms-1" v-if="parsedFileDataColumnMapping && parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.Amount)">({{ KnownAmountFormat.valueOf(parsedFileDataColumnMapping.amountFormat || parsedFileAutoDetectedAmountFormat || '')?.format || tt('Unknown') }})</span>
                    <v-menu eager activator="parent" location="bottom" max-height="500">
                        <v-list>
                            <v-list-item key="auto"
                                         :append-icon="parsedFileDataColumnMapping.amountFormat === '' ? mdiCheck : undefined"
                                         @click="parsedFileDataColumnMapping.amountFormat = ''">
                                <v-list-item-title class="cursor-pointer">
                                    <span>{{ tt('Auto detect') }}</span>
                                    <span class="ms-1">({{ displayFileAutoDetectedAmountFormat }})</span>
                                </v-list-item-title>
                            </v-list-item>
                            <v-list-item :key="amountFormat.type"
                                         :append-icon="parsedFileDataColumnMapping.amountFormat === amountFormat.type ? mdiCheck : undefined"
                                         v-for="amountFormat in KnownAmountFormat.values()"
                                         @click="parsedFileDataColumnMapping.amountFormat = amountFormat.type">
                                <v-list-item-title class="cursor-pointer">
                                    {{ amountFormat.format }}
                                </v-list-item-title>
                            </v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-btn class="ms-2" color="secondary" density="compact" variant="outlined"
                       v-if="parsedFileDataColumnMapping && parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.GeographicLocation)">
                    <span>{{ tt('Geographic Location Separator') }}</span>
                    <span class="ms-1" v-if="parsedFileDataColumnMapping.geoLocationOrder">({{ parsedFileDataColumnMapping.formatGeoLocation(tt('Latitude'), tt('Longitude')) }})</span>
                    <v-menu eager activator="parent" location="bottom" max-height="500"
                            :close-on-content-click="false">
                        <v-list class="pa-0">
                            <v-list-item class="pa-0">
                                <v-table class="transaction-types-popup-menu">
                                    <tbody>
                                    <tr :key="separator.value"
                                        v-for="separator in allSeparators">
                                        <td>{{ separator.name }} ({{separator.value}})</td>
                                        <td>
                                            <v-btn-toggle class="toggle-buttons" density="compact" variant="outlined"
                                                          color="default" mandatory="force" divided
                                                          v-model="parsedFileDataColumnMapping.geoLocationOrder"
                                                          v-if="parsedFileDataColumnMapping.geoLocationSeparator === separator.value">
                                                <v-btn value="latlon">{{ `${tt('Latitude')}${separator.value}${tt('Longitude')}` }}</v-btn>
                                                <v-btn value="lonlat">{{ `${tt('Longitude')}${separator.value}${tt('Latitude')}` }}</v-btn>
                                            </v-btn-toggle>
                                            <v-btn-group class="toggle-buttons" density="compact" variant="outlined"
                                                         color="default" mandatory="force" divided
                                                         v-if="parsedFileDataColumnMapping.geoLocationSeparator !== separator.value">
                                                <v-btn @click="parsedFileDataColumnMapping.setGeoLocationFormat(separator.value, 'latlon')">{{ `${tt('Latitude')}${separator.value}${tt('Longitude')}` }}</v-btn>
                                                <v-btn @click="parsedFileDataColumnMapping.setGeoLocationFormat(separator.value, 'lonlat')">{{ `${tt('Longitude')}${separator.value}${tt('Latitude')}` }}</v-btn>
                                            </v-btn-group>
                                        </td>
                                    </tr>
                                    </tbody>
                                </v-table>
                            </v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-btn class="ms-2" color="secondary" density="compact" variant="outlined"
                       v-if="parsedFileDataColumnMapping && parsedFileDataColumnMapping.isColumnMappingSet(ImportTransactionColumnType.Tags)">
                    <span>{{ tt('Transaction Tags Separator') }}</span>
                    <span class="ms-1" v-if="parsedFileDataColumnMapping.tagSeparator">({{ parsedFileDataColumnMapping.tagSeparator }})</span>
                    <v-menu eager activator="parent" location="bottom" max-height="500">
                        <v-list>
                            <v-list-item :key="separator.value"
                                         :append-icon="parsedFileDataColumnMapping.tagSeparator === separator.value ? mdiCheck : undefined"
                                         v-for="separator in allSeparators"
                                         @click="parsedFileDataColumnMapping.tagSeparator = separator.value">
                                <v-list-item-title class="cursor-pointer">
                                    {{ separator.name }} ({{separator.value}})
                                </v-list-item-title>
                            </v-list-item>
                        </v-list>
                    </v-menu>
                </v-btn>
                <v-spacer/>
                <span class="ms-2">{{ tt('Lines Per Page') }}</span>
                <v-select class="ms-2" density="compact" max-width="100"
                          item-title="name"
                          item-value="value"
                          :disabled="!!disabled"
                          :items="parsedFileLinesTablePageOptions"
                          v-model="countPerPage"
                />
                <pagination-buttons density="comfortable"
                                    :disabled="!!disabled"
                                    :totalPageCount="Math.ceil((parsedFileLines ? parsedFileLines.length : 0) / countPerPage)"
                                    v-model="currentPage"></pagination-buttons>
            </div>
        </div>
    </div>

    <snack-bar ref="snackbar" />
</template>
