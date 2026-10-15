import type { SetupProps } from '@/lib/ui/props.ts';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { KNOWN_COLUMN_NAME_MAPPING, KNOWN_TRANSACTION_TYPE_NAME_MAPPING } from '@/consts/import_transaction.ts';
import { DEFAULT_PAGE_COUNTS } from '@/consts/page.ts';
import { type NameNumeralValue, type NameValue, type TypeAndDisplayName, entries, itemAndIndex, keys } from '@/core/base.ts';
import { type DateFormatOrder } from '@/core/datetime.ts';
import { KnownFileType } from '@/core/file.ts';
import { ImportTransactionColumnType, ImportTransactionDataMapping } from '@/core/import_transaction.ts';
import { KnownAmountFormat } from '@/core/numeral.ts';
import { KnownDateTimezoneFormat } from '@/core/timezone.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    findDisplayNameByType,
    getObjectOwnFieldCount,
    isNumber,
    isObjectEmpty
} from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import {
    openTextFileContent,
    startDownloadFile
} from '@/lib/ui/common.ts';
import { useI18n } from '@/locales/helpers.ts';
import {
    mdiContentSaveOutline,
    mdiFolderOpenOutline
} from '@mdi/js';
import { computed, ref, useTemplateRef, watch } from 'vue';

export type SnackBarType = InstanceType<typeof SnackBar>;

export interface ImportTransactionDefineColumnResult {
    includeHeader: boolean;
    columnMapping: Record<number, number>;
    transactionTypeMapping: Record<string, TransactionType>;
    timeFormat: string | undefined;
    timezoneFormat: string | undefined;
    amountDecimalSeparator: string | undefined;
    amountDigitGroupingSymbol: string | undefined;
    geoLocationSeparator: string;
    geoLocationOrder: string;
    tagSeparator: string;
}

export interface ImportTransactionDefineColumnMenu {
    prependIcon: string;
    title: string;
    disabled?: boolean;
    onClick: () => void;
}

// Coordinates ImportTransactionDefineColumnTab state and effects; called once from the component setup.
export function useImportTransactionDefineColumnTab(props: SetupProps<{
    parsedFileData?: string[][];
    parsedFileColumnSeparator?: string;
    disabled?: boolean;
}>) {
    const {
        tt,
        ti,
        getLongDateFormatOrder,
        getShortDateFormatOrder,
        getAllImportTransactionColumnTypes,
        formatNumberToLocalizedNumerals,
        getTablePageOptions
    } = useI18n();

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const currentPage = ref<number>(1);

    const countPerPage = ref<number>(10);

    const parsedFileDataColumnMapping = ref<ImportTransactionDataMapping>(ImportTransactionDataMapping.createEmpty(props.parsedFileColumnSeparator));

    const longDateFormatOrder = computed<DateFormatOrder>(() => getLongDateFormatOrder());

    const shortDateFormatOrder = computed<DateFormatOrder>(() => getShortDateFormatOrder());

    const allImportTransactionColumnTypes = computed<TypeAndDisplayName[]>(() => getAllImportTransactionColumnTypes());

    const menus = computed<ImportTransactionDefineColumnMenu[]>(() => [
        {
            prependIcon: mdiFolderOpenOutline,
            title: tt('Load Data Mapping File'),
            onClick: loadColumnMappingFile
        },
        {
            prependIcon: mdiContentSaveOutline,
            title: tt('Save Data Mapping File'),
            onClick: saveColumnMappingFile
        }
    ]);

    const allSeparators = computed<NameValue[]>(() => {
        const separators: NameValue[] = [
            {
                name: tt('Space'),
                value: ' '
            },
            {
                name: tt('Comma'),
                value: ','
            },
            {
                name: tt('Semicolon'),
                value: ';'
            },
            {
                name: tt('Tab'),
                value: '\t'
            },
            {
                name: tt('Vertical Bar'),
                value: '|'
            }
        ];

        return separators;
    });

    const parsedFileLinesHeaders = computed<object[]>(() => {
        let maxColumnCount = 0;

        if (props.parsedFileData) {
            for (const lineData of props.parsedFileData) {
                if (lineData.length > maxColumnCount) {
                    maxColumnCount = lineData.length;
                }
            }
        }

        const firstLine: string[] = props.parsedFileData && props.parsedFileData.length > 0 ? (props.parsedFileData[0] as string[]) : [];

        const headers: object[] = [];
        headers.push({ key: 'index', value: 'index', title: '#', sortable: true, nowrap: true });

        for (let i = 0; i < maxColumnCount; i++) {
            let title = `#${i + 1}`;

            if (parsedFileDataColumnMapping.value.includeHeader && firstLine && firstLine[i]) {
                title = firstLine[i] as string;
            }

            headers.push({ key: i.toString(), value: `column${i + 1}`, title: title, sortable: true, nowrap: true });
        }

        return headers;
    });

    const parsedFileLines = computed<Record<string, string>[] | undefined>(() => {
        if (!props.parsedFileData) {
            return undefined;
        }

        const allLines: Record<string, string>[] = [];
        const startIndex = parsedFileDataColumnMapping.value.includeHeader ? 1 : 0;

        for (let i = startIndex, index = 1; i < props.parsedFileData.length; i++, index++) {
            const line: Record<string, string> = {};
            const columns = props.parsedFileData[i] as string[];

            for (const [data, columnIndex] of itemAndIndex(columns)) {
                line['index'] = index.toString();
                line[`column${columnIndex + 1}`] = data;
            }

            allLines.push(line);
        }

        return allLines;
    });

    const parsedFileLinesTablePageOptions = computed<NameNumeralValue[]>(() => getTablePageOptions(DEFAULT_PAGE_COUNTS, parsedFileLines.value?.length, true, false));

    const parsedFileAllTransactionTypes = computed<string[]>(() => parsedFileDataColumnMapping.value.parseFileAllTransactionTypes(props.parsedFileData));

    const parsedFileValidMappedTransactionTypes = computed<Record<string, TransactionType>>(() => parsedFileDataColumnMapping.value.parseFileValidMappedTransactionTypes(props.parsedFileData));

    const parsedFileAutoDetectedTimeFormat = computed<string | undefined>(() => parsedFileDataColumnMapping.value.parseFileAutoDetectedTimeFormat(props.parsedFileData, longDateFormatOrder.value, shortDateFormatOrder.value));

    const parsedFileAutoDetectedTimezoneFormat = computed<string | undefined>(() => parsedFileDataColumnMapping.value.parseFileAutoDetectedTimezoneFormat(props.parsedFileData));

    const parsedFileAutoDetectedAmountFormat = computed<string | undefined>(() => parsedFileDataColumnMapping.value.parseFileAutoDetectedAmountFormat(props.parsedFileData));

    const displayFileCurrentTimezoneFormat = computed<string>(() => {
        const format = parsedFileDataColumnMapping.value.timezoneFormat || parsedFileAutoDetectedTimezoneFormat.value || '';

        if (format) {
            const knownFormat = KnownDateTimezoneFormat.valueOf(format);

            if (knownFormat) {
                return ti(knownFormat.name, knownFormat.nameSupportedI18n);
            }
        }

        return tt('Unknown');
    });

    const displayFileAutoDetectedTimezoneFormat = computed<string>(() => {
        const format = parsedFileAutoDetectedTimezoneFormat.value;

        if (format) {
            const knownFormat = KnownDateTimezoneFormat.valueOf(format);

            if (knownFormat) {
                return ti(knownFormat.name, knownFormat.nameSupportedI18n);
            }
        }

        return tt('Unknown');
    });

    const displayFileAutoDetectedAmountFormat = computed<string>(() => {
        const format = parsedFileAutoDetectedAmountFormat.value;

        if (format) {
            const knownFormat = KnownAmountFormat.valueOf(format);

            if (knownFormat) {
                return knownFormat.format;
            }
        }

        return tt('Unknown');
    });

    function getNormalizedKey(key: string): string {
        return key.toLowerCase().replaceAll(' ', '').replaceAll('_', '').replaceAll('-', '');
    }

    function getParseDataMappedColumnDisplayName(columnIndex: number): string {
        for (const [columnType, index] of entries(parsedFileDataColumnMapping.value.dataColumnMapping)) {
            if (index === columnIndex) {
                return findDisplayNameByType(allImportTransactionColumnTypes.value, parseInt(columnType)) || tt('Unspecified');
            }
        }

        return tt('Unspecified');
    }

    function toggleDataMappingColumn(columnIndex: number, columnType: number): void {
        parsedFileDataColumnMapping.value.toggleDataMappingColumn(columnIndex, columnType);

        if (columnType === ImportTransactionColumnType.TransactionType.type) {
            if (!parsedFileDataColumnMapping.value.isColumnMappingSet(ImportTransactionColumnType.TransactionType)) {
                parsedFileDataColumnMapping.value.transactionTypeMapping = {};
            } else {
                const allTransactionTypes = parsedFileAllTransactionTypes.value;

                for (const transactionTypeName of keys(parsedFileDataColumnMapping.value.transactionTypeMapping)) {
                    if (!allTransactionTypes.includes(transactionTypeName)) {
                        delete parsedFileDataColumnMapping.value.transactionTypeMapping[transactionTypeName];
                    }
                }
            }

            autoSetTransactionTypeMapping();
        }
    }

    function autoSetColumnMapping(): void {
        if (!props.parsedFileData) {
            return;
        }

        const firstLine: string[] = props.parsedFileData.length > 0 ? (props.parsedFileData[0] as string[]) : [];
        const displayColumnNamesMap: Record<string, ImportTransactionColumnType> = {};

        for (const column of ImportTransactionColumnType.values()) {
            displayColumnNamesMap[getNormalizedKey(tt(column.name))] = column;
        }

        for (const [columnName, index] of itemAndIndex(firstLine)) {
            const normalizedColumnName = getNormalizedKey(columnName);

            if (displayColumnNamesMap[normalizedColumnName]) {
                const columnType = displayColumnNamesMap[normalizedColumnName];

                if (!isNumber(parsedFileDataColumnMapping.value.dataColumnMapping[columnType.type])) {
                    parsedFileDataColumnMapping.value.dataColumnMapping[columnType.type] = index;
                }

                continue;
            }

            if (KNOWN_COLUMN_NAME_MAPPING[normalizedColumnName]) {
                const columnType = KNOWN_COLUMN_NAME_MAPPING[normalizedColumnName];

                if (!isNumber(parsedFileDataColumnMapping.value.dataColumnMapping[columnType.type])) {
                    parsedFileDataColumnMapping.value.dataColumnMapping[columnType.type] = index;
                }

                continue;
            }
        }
    }

    function autoSetTransactionTypeMapping(): void {
        if (!props.parsedFileData) {
            return;
        }

        const displayTransactinTypeNamesMap: Record<string, TransactionType> = {
            [getNormalizedKey(tt('Modify Balance'))]: TransactionType.ModifyBalance,
            [getNormalizedKey(tt('Income'))]: TransactionType.Income,
            [getNormalizedKey(tt('Expense'))]: TransactionType.Expense,
            [getNormalizedKey(tt('Transfer'))]: TransactionType.Transfer
        };

        for (const transactionTypeName of parsedFileAllTransactionTypes.value) {
            const normalizedTransactionTypeName = getNormalizedKey(transactionTypeName);

            if (displayTransactinTypeNamesMap[normalizedTransactionTypeName]) {
                const transactionType = displayTransactinTypeNamesMap[normalizedTransactionTypeName];
                parsedFileDataColumnMapping.value.transactionTypeMapping[transactionTypeName] = transactionType;
                continue;
            }

            if (KNOWN_TRANSACTION_TYPE_NAME_MAPPING[normalizedTransactionTypeName]) {
                const transactionType = KNOWN_TRANSACTION_TYPE_NAME_MAPPING[normalizedTransactionTypeName];
                parsedFileDataColumnMapping.value.transactionTypeMapping[transactionTypeName] = transactionType;
                continue;
            }
        }
    }

    function generateResult(): ImportTransactionDefineColumnResult | undefined {
        const columnMapping: Record<number, number> = parsedFileDataColumnMapping.value.dataColumnMapping;
        const transactionTypeMapping: Record<string, TransactionType> = parsedFileValidMappedTransactionTypes.value;
        const includeHeader: boolean = parsedFileDataColumnMapping.value.includeHeader;
        const geoLocationSeparator: string = parsedFileDataColumnMapping.value.geoLocationSeparator;
        const geoLocationOrder: string = parsedFileDataColumnMapping.value.geoLocationOrder;
        const tagSeparator: string = parsedFileDataColumnMapping.value.tagSeparator;

        let timeFormat: string | undefined = parsedFileDataColumnMapping.value.timeFormat;
        let timezoneFormat: string | undefined = parsedFileDataColumnMapping.value.timezoneFormat;
        let amountFormat: string | undefined = parsedFileDataColumnMapping.value.amountFormat;
        let amountDecimalSeparator: string | undefined = undefined;
        let amountDigitGroupingSymbol: string | undefined = undefined;

        if (!columnMapping
            || !isNumber(columnMapping[ImportTransactionColumnType.TransactionTime.type])
            || !isNumber(columnMapping[ImportTransactionColumnType.TransactionType.type])
            || !isNumber(columnMapping[ImportTransactionColumnType.Amount.type])) {
            snackbar.value?.showError('Missing transaction time, transaction type, or amount column mapping');
            return undefined;
        }

        if (!transactionTypeMapping || isObjectEmpty(transactionTypeMapping)) {
            snackbar.value?.showError('Transaction type mapping is not set');
            return undefined;
        }

        if (!parsedFileDataColumnMapping.value.timeFormat) {
            timeFormat = parsedFileAutoDetectedTimeFormat.value;
        }

        if (!parsedFileDataColumnMapping.value.timezoneFormat) {
            timezoneFormat = parsedFileAutoDetectedTimezoneFormat.value;
        }

        if (!parsedFileDataColumnMapping.value.amountFormat) {
            amountFormat = parsedFileAutoDetectedAmountFormat.value;
        }

        if (amountFormat) {
            const knownAmountFormat = KnownAmountFormat.valueOf(amountFormat);

            if (knownAmountFormat) {
                amountDecimalSeparator = knownAmountFormat.decimalSeparator.symbol;
                amountDigitGroupingSymbol = knownAmountFormat.digitGroupingSymbol?.symbol;
            }
        }

        if (!timeFormat) {
            snackbar.value?.showError('Transaction time format is not set');
            return undefined;
        }

        if (!amountDecimalSeparator) {
            snackbar.value?.showError('Transaction amount format is not set');
            return undefined;
        }

        return {
            includeHeader: includeHeader,
            columnMapping: columnMapping,
            transactionTypeMapping: transactionTypeMapping,
            timeFormat: timeFormat,
            timezoneFormat: timezoneFormat,
            amountDecimalSeparator: amountDecimalSeparator,
            amountDigitGroupingSymbol: amountDigitGroupingSymbol,
            geoLocationSeparator: geoLocationSeparator,
            geoLocationOrder: geoLocationOrder,
            tagSeparator: tagSeparator
        };
    }

    function reset(): void {
        parsedFileDataColumnMapping.value.reset(props.parsedFileColumnSeparator);
        currentPage.value = 1;
        countPerPage.value = 10;
    }

    function loadColumnMappingFile(): void {
        openTextFileContent({
            allowedExtensions: KnownFileType.JSON.contentType
        }).then(content => {
            const result = ImportTransactionDataMapping.parseFromJson(content);

            if (result) {
                parsedFileDataColumnMapping.value = result;
            } else {
                logger.error('Failed to parse data mapping file');
                snackbar.value?.showError('Data mapping file is invalid');
            }
        }).catch(error => {
            logger.error('Failed to open data mapping file', error);
            snackbar.value?.showError('Data mapping file is invalid');
        });
    }

    function saveColumnMappingFile(): void {
        const fileName = KnownFileType.JSON.formatFileName(tt('dataExport.defaultImportDataMappingFileName'));
        startDownloadFile(fileName, KnownFileType.JSON.createBlob(parsedFileDataColumnMapping.value.toJson()));
    }

    watch(() => props.parsedFileData, (newValue, oldValue) => {
        if (newValue && !oldValue && getObjectOwnFieldCount(parsedFileDataColumnMapping.value.dataColumnMapping) < 1) {
            autoSetColumnMapping();
            autoSetTransactionTypeMapping();
        }
    }, { immediate: true });

    return {
        tt, ti, formatNumberToLocalizedNumerals, snackbar, currentPage, countPerPage, parsedFileDataColumnMapping,
        allImportTransactionColumnTypes, menus, allSeparators, parsedFileLinesHeaders, parsedFileLines,
        parsedFileLinesTablePageOptions, parsedFileAllTransactionTypes, parsedFileValidMappedTransactionTypes,
        parsedFileAutoDetectedTimeFormat, parsedFileAutoDetectedAmountFormat, displayFileCurrentTimezoneFormat,
        displayFileAutoDetectedTimezoneFormat, displayFileAutoDetectedAmountFormat, getParseDataMappedColumnDisplayName,
        toggleDataMappingColumn, generateResult, reset, loadColumnMappingFile, saveColumnMappingFile,
    };
}
