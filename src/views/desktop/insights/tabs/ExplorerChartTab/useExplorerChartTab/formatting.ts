import {
    entries
} from '@/core/base.ts';
import { Month, WeekDay } from '@/core/datetime.ts';
import {
    TransactionExplorerDataDimension,
    TransactionExplorerDataDimensionType
} from '@/core/explorer.ts';
import { NumeralSystem } from '@/core/numeral.ts';
import { isNumber, isString } from '@/lib/common.ts';
import { getCurrentDateTime, parseDateTimeFromString } from '@/lib/datetime.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import {
    type CategoriedInfo,
    type SeriesInfo,
    TransactionExplorerDimensionType
} from '@/stores/explorer.ts';
import type { useExplorerChartTabState } from './state.ts';

interface Context {
    formatNumberToLocalizedNumeralsWithoutDigitGrouping: ReturnType<typeof useExplorerChartTabState>['formatNumberToLocalizedNumeralsWithoutDigitGrouping'];
    currentExploration: ReturnType<typeof useExplorerChartTabState>['currentExploration'];
    tt: ReturnType<typeof useExplorerChartTabState>['tt'];
    formatDateTimeToShortDateTime: ReturnType<typeof useExplorerChartTabState>['formatDateTimeToShortDateTime'];
    formatDateTimeToShortDate: ReturnType<typeof useExplorerChartTabState>['formatDateTimeToShortDate'];
    formatDateTimeToGregorianLikeShortYearMonth: ReturnType<typeof useExplorerChartTabState>['formatDateTimeToGregorianLikeShortYearMonth'];
    formatDateTimeToGregorianLikeYearQuarter: ReturnType<typeof useExplorerChartTabState>['formatDateTimeToGregorianLikeYearQuarter'];
    formatDateTimeToGregorianLikeShortYear: ReturnType<typeof useExplorerChartTabState>['formatDateTimeToGregorianLikeShortYear'];
    formatGregorianYearToGregorianLikeFiscalYear: ReturnType<typeof useExplorerChartTabState>['formatGregorianYearToGregorianLikeFiscalYear'];
    getWeekdayLongName: ReturnType<typeof useExplorerChartTabState>['getWeekdayLongName'];
    getMonthdayShortName: ReturnType<typeof useExplorerChartTabState>['getMonthdayShortName'];
    getMonthLongName: ReturnType<typeof useExplorerChartTabState>['getMonthLongName'];
    getQuarterName: ReturnType<typeof useExplorerChartTabState>['getQuarterName'];
    formatDateTimeToShortTime: ReturnType<typeof useExplorerChartTabState>['formatDateTimeToShortTime'];
    getCurrencyName: ReturnType<typeof useExplorerChartTabState>['getCurrencyName'];
    formatAmountToLocalizedNumerals: ReturnType<typeof useExplorerChartTabState>['formatAmountToLocalizedNumerals'];
    defaultCurrency: ReturnType<typeof useExplorerChartTabState>['defaultCurrency'];
    formatRange: ReturnType<typeof useExplorerChartTabState>['formatRange'];
}

export function useExplorerChartTabFormatting(context: Context) {
    const {
        formatNumberToLocalizedNumeralsWithoutDigitGrouping, currentExploration, tt, formatDateTimeToShortDateTime,
        formatDateTimeToShortDate, formatDateTimeToGregorianLikeShortYearMonth, formatDateTimeToGregorianLikeYearQuarter,
        formatDateTimeToGregorianLikeShortYear, formatGregorianYearToGregorianLikeFiscalYear, getWeekdayLongName,
        getMonthdayShortName, getMonthLongName, getQuarterName, formatDateTimeToShortTime, getCurrencyName,
        formatAmountToLocalizedNumerals, defaultCurrency, formatRange,
    } = context;

    function getFormattedI18nParameters(i18nParameters: Record<string, unknown> | undefined): Record<string, string> | undefined {
        if (!i18nParameters) {
            return undefined;
        }

        const formattedParameters: Record<string, string> = {};

        for (const [key, value] of entries(i18nParameters)) {
            if (isNumber(value)) {
                formattedParameters[key] = formatNumberToLocalizedNumeralsWithoutDigitGrouping(value);
            } else if (isString(value)) {
                formattedParameters[key] = value;
            }
        }

        return formattedParameters;
    }

    function getCategoriedDataDisplayName(info: CategoriedInfo | SeriesInfo): string {
        let name: string = '';
        let needI18n: boolean | undefined = false;
        let i18nParameters: Record<string, unknown> | undefined = undefined;
        let dimessionType: TransactionExplorerDimensionType = TransactionExplorerDimensionType.Other;
        let dimessionValue: TransactionExplorerDataDimensionType = TransactionExplorerDataDimension.None.value;

        if ('categoryName' in info) {
            name = info.categoryName;
            needI18n = info.categoryNameNeedI18n;
            i18nParameters = getFormattedI18nParameters(info.categoryNameI18nParameters);
            dimessionType = info.categoryIdType;
            dimessionValue = currentExploration.value.categoryDimension;
        } else if ('seriesName' in info) {
            name = info.seriesName;
            needI18n = info.seriesNameNeedI18n;
            i18nParameters = getFormattedI18nParameters(info.seriesNameI18nParameters);
            dimessionType = info.seriesIdType;
            dimessionValue = currentExploration.value.seriesDimension;
        }

        const dimession = TransactionExplorerDataDimension.valueOf(dimessionValue);
        let displayName: string = name;

        // convert the name to i18n if needed
        if (needI18n && i18nParameters) {
            displayName = tt(name, i18nParameters);
        } else if (needI18n && !i18nParameters) {
            displayName = tt(name);
        }

        // convert the name to formatted date time if needed
        if (dimession === TransactionExplorerDataDimension.DateTime) {
            const dateTime = parseDateTimeFromString(name, dimessionType);
            displayName = dateTime ? formatDateTimeToShortDateTime(dateTime) : tt('Unknown');
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByYearMonthDay) {
            const dateTime = parseDateTimeFromString(name, dimessionType);
            displayName = dateTime ? formatDateTimeToShortDate(dateTime) : tt('Unknown');
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByYearMonth) {
            const dateTime = parseDateTimeFromString(name, dimessionType);
            displayName = dateTime ? formatDateTimeToGregorianLikeShortYearMonth(dateTime) : tt('Unknown');
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByYearQuarter) {
            const parts = name.split('-');
            const year = parts.length === 2 ? parts[0] : '';
            const quarter = parts.length === 2 ? parseInt(parts[1] as string) : 0;
            const quarterLastMonth = quarter * 3;
            const dateTime = year && quarterLastMonth ? parseDateTimeFromString(`${year}-${quarterLastMonth.toString(10).padStart(2, NumeralSystem.WesternArabicNumerals.digitZero)}`, TransactionExplorerDimensionType.YearMonth) : undefined;
            displayName = dateTime ? formatDateTimeToGregorianLikeYearQuarter(dateTime) : tt('Unknown');
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByYear) {
            const dateTime = parseDateTimeFromString(name, dimessionType);
            displayName = dateTime ? formatDateTimeToGregorianLikeShortYear(dateTime) : tt('Unknown');
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByFiscalYear) {
            displayName = formatGregorianYearToGregorianLikeFiscalYear(parseInt(name));
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByDayOfWeek) {
            const weekDay = WeekDay.parse(name);
            displayName = weekDay ? getWeekdayLongName(weekDay) : tt('Unknown');
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByDayOfMonth) {
            displayName = getMonthdayShortName(parseInt(name));
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByMonthOfYear) {
            const month = Month.valueOf(parseInt(name));
            displayName = month ? getMonthLongName(month.name) : tt('Unknown');
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByQuarterOfYear) {
            displayName = getQuarterName(parseInt(name));
        } else if (dimession === TransactionExplorerDataDimension.DateTimeByHourOfDay) {
            const dateTime = getCurrentDateTime().set({
                hour: parseInt(name),
                minute: 0,
                second: 0,
                millisecond: 0
            });
            displayName = formatDateTimeToShortTime(dateTime);
        } else if (dimession === TransactionExplorerDataDimension.SourceAccountCurrency || dimession === TransactionExplorerDataDimension.DestinationAccountCurrency) {
            if (!needI18n) {
                displayName = getCurrencyName(name);
            }
        }

        if (dimession === TransactionExplorerDataDimension.SourceAmount
            || dimession === TransactionExplorerDataDimension.DestinationAmount) {
            if (name !== '' && name !== 'none' && Number.isFinite(parseInt(name))) {
                displayName = formatAmountToLocalizedNumerals(parseBigDecimal(name), defaultCurrency.value);
            }
        }

        if (dimession?.isSourceAmountRange || dimession?.isDestinationAmountRange) {
            const rangeParts = name.split('|');

            if (rangeParts && rangeParts.length === 2 && Number.isFinite(parseInt(rangeParts[0] as string)) && Number.isFinite(parseInt(rangeParts[1] as string))) {
                const from = formatAmountToLocalizedNumerals(parseBigDecimal(rangeParts[0] as string), defaultCurrency.value);
                const to = formatAmountToLocalizedNumerals(parseBigDecimal(rangeParts[1] as string), defaultCurrency.value);
                displayName = formatRange(from, to);
            }
        }

        return displayName;
    }

    return { getCategoriedDataDisplayName };
}
