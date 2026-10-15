import { itemAndIndex } from '@/core/base.ts';
import { KnownFileType } from '@/core/file.ts';
import { ImportTransactionColumnType } from '@/core/import_transaction.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    replaceAll
} from '@/lib/common.ts';
import {
    getUtcOffsetByUtcOffsetMinutes,
    parseDateTimeFromUnixTimeWithTimezoneOffset
} from '@/lib/datetime.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { startDownloadFile } from '@/lib/ui/common.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { ImportTransaction } from '@/models/imported_transaction.ts';
import type { useImportTransactionCheckDataTabFormatting } from './formatting.ts';
import type { useImportTransactionCheckDataTabSelection } from './selection.ts';
import type { useImportTransactionCheckDataTabState } from './state.ts';

interface Context {
    filters: ReturnType<typeof useImportTransactionCheckDataTabState>['filters'];
    showCustomDateRangeDialog: ReturnType<typeof useImportTransactionCheckDataTabState>['showCustomDateRangeDialog'];
    props: SetupProps<{
    importTransactions?: ImportTransaction[]
    disabled?: boolean;
}>;
    selectedImportTransactionCount: ReturnType<typeof useImportTransactionCheckDataTabSelection>['selectedImportTransactionCount'];
    tt: ReturnType<typeof useImportTransactionCheckDataTabState>['tt'];
    getDisplayTransactionType: ReturnType<typeof useImportTransactionCheckDataTabFormatting>['getDisplayTransactionType'];
    allAccountsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allAccountsMap'];
    formatAmountToWesternArabicNumeralsWithoutDigitGrouping: ReturnType<typeof useImportTransactionCheckDataTabState>['formatAmountToWesternArabicNumeralsWithoutDigitGrouping'];
    formatNumberToWesternArabicNumeralsWithoutDigitGrouping: ReturnType<typeof useImportTransactionCheckDataTabState>['formatNumberToWesternArabicNumeralsWithoutDigitGrouping'];
    allCategoriesMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allCategoriesMap'];
    allTagsMap: ReturnType<typeof useImportTransactionCheckDataTabState>['allTagsMap'];
    formatDateTimeToGregorianDefaultDateTime: ReturnType<typeof useImportTransactionCheckDataTabState>['formatDateTimeToGregorianDefaultDateTime'];
}

export function useImportTransactionCheckDataTabExport(context: Context) {
    const {
        filters, showCustomDateRangeDialog, props, selectedImportTransactionCount, tt, getDisplayTransactionType, allAccountsMap,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, formatNumberToWesternArabicNumeralsWithoutDigitGrouping,
        allCategoriesMap, allTagsMap, formatDateTimeToGregorianDefaultDateTime,
    } = context;

    function changeCustomDateFilter(minTime: number, maxTime: number): void {
        filters.value.minDatetime = minTime;
        filters.value.maxDatetime = maxTime;
        showCustomDateRangeDialog.value = false;
    }

    function exportData(fileType: KnownFileType): void {
        if (!props.importTransactions || props.importTransactions.length < 1 || selectedImportTransactionCount.value < 1) {
            return;
        }

        let separator = ',';
        let tagSeparator = ';';

        if (fileType === KnownFileType.TSV) {
            separator = '\t';
        } else if (fileType === KnownFileType.SSV) {
            separator = ';';
            tagSeparator = ',';
        }

        const header = [
            tt(ImportTransactionColumnType.TransactionTime.name),
            tt(ImportTransactionColumnType.TransactionTimezone.name),
            tt(ImportTransactionColumnType.TransactionType.name),
            tt(ImportTransactionColumnType.Category.name),
            tt(ImportTransactionColumnType.AccountName.name),
            tt(ImportTransactionColumnType.AccountCurrency.name),
            tt(ImportTransactionColumnType.Amount.name),
            tt(ImportTransactionColumnType.RelatedAccountName.name),
            tt(ImportTransactionColumnType.RelatedAccountCurrency.name),
            tt(ImportTransactionColumnType.RelatedAmount.name),
            tt(ImportTransactionColumnType.GeographicLocation.name),
            tt(ImportTransactionColumnType.Tags.name),
            tt(ImportTransactionColumnType.Description.name)
        ].join(separator) + '\n';

        const transactions = props.importTransactions ?? [];
        const rows = transactions.filter(transaction => transaction.selected).map(transaction => {
            const transactionTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transaction.utcOffset);
            const type = getDisplayTransactionType(transaction);
            const accountName = transaction.sourceAccountId && transaction.sourceAccountId !== '0' && allAccountsMap.value[transaction.sourceAccountId] ? (allAccountsMap.value[transaction.sourceAccountId]?.name ?? transaction.originalSourceAccountName) : transaction.originalSourceAccountName;
            const amountCurrency = transaction.sourceAccountId && transaction.sourceAccountId !== '0' && allAccountsMap.value[transaction.sourceAccountId] ? (allAccountsMap.value[transaction.sourceAccountId]?.currency ?? transaction.originalSourceAccountCurrency) : transaction.originalSourceAccountCurrency;
            const amount = formatAmountToWesternArabicNumeralsWithoutDigitGrouping(parseBigDecimal(transaction.sourceAmount), amountCurrency);
            const geographicLocation = transaction.geoLocation ? `${formatNumberToWesternArabicNumeralsWithoutDigitGrouping(transaction.geoLocation.longitude)} ${formatNumberToWesternArabicNumeralsWithoutDigitGrouping(transaction.geoLocation.latitude)}` : '';
            let categoryName = transaction.categoryId && transaction.categoryId !== '0' && allCategoriesMap.value[transaction.categoryId] ? (allCategoriesMap.value[transaction.categoryId]?.name ?? transaction.originalCategoryName) : transaction.originalCategoryName;
            let relatedAccountName: string | undefined = undefined;
            let relatedAccountCurrency: string | undefined = undefined;
            let relatedAmount: string | undefined = undefined;

            if (transaction.type === TransactionType.ModifyBalance) {
                categoryName = '';
            } else if (transaction.type === TransactionType.Transfer) {
                relatedAccountName = transaction.destinationAccountId && transaction.destinationAccountId !== '0' && allAccountsMap.value[transaction.destinationAccountId] ? (allAccountsMap.value[transaction.destinationAccountId]?.name ?? transaction.originalDestinationAccountName) : transaction.originalDestinationAccountName;
                relatedAccountCurrency = transaction.destinationAccountId && transaction.destinationAccountId !== '0' && allAccountsMap.value[transaction.destinationAccountId] ? (allAccountsMap.value[transaction.destinationAccountId]?.currency ?? transaction.originalDestinationAccountCurrency) : transaction.originalDestinationAccountCurrency;
                relatedAmount = formatAmountToWesternArabicNumeralsWithoutDigitGrouping(parseBigDecimal(transaction.destinationAmount), relatedAccountCurrency);
            }

            const tagNames: string[] = [];

            if (transaction.tagIds && transaction.tagIds.length > 0) {
                for (const [tagId, index] of itemAndIndex(transaction.tagIds)) {
                    let tagName = '';

                    if (tagId && tagId !== '0' && allTagsMap.value[tagId]) {
                        tagName = allTagsMap.value[tagId]!.name;
                    } else if (transaction.originalTagNames && transaction.originalTagNames[index]) {
                        tagName = transaction.originalTagNames[index] as string;
                    }

                    if (tagName) {
                        tagName = replaceAll(tagName, separator, ' ');
                        tagName = replaceAll(tagName, tagSeparator, ' ');
                        tagNames.push(tagName);
                    }
                }
            }

            return [
                formatDateTimeToGregorianDefaultDateTime(transactionTime),
                getUtcOffsetByUtcOffsetMinutes(transaction.utcOffset),
                type,
                replaceAll(categoryName, separator, ' '),
                replaceAll(accountName, separator, ' '),
                amountCurrency,
                amount,
                replaceAll(relatedAccountName ?? '', separator, ' '),
                relatedAccountCurrency ?? '',
                relatedAmount ?? '',
                geographicLocation,
                tagNames.join(tagSeparator),
                replaceAll(transaction.comment || '', separator, ' ')
            ].join(separator);
        });

        startDownloadFile(fileType.formatFileName(tt('dataExport.defaultImportCheckResultFileName')), fileType.createBlob(header + rows.join('\n')));
    }

    return { changeCustomDateFilter, exportData };
}
