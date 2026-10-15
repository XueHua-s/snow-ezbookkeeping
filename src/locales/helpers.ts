import { useHelpersAccountBalances } from './helpers/account-balances.ts';
import { useHelpersAmountFormatting } from './helpers/amount-formatting.ts';
import { useHelpersCalendarLabels } from './helpers/calendar-labels.ts';
import { useHelpersCalendarLocales } from './helpers/calendar-locales.ts';
import { useHelpersCurrencyOptions } from './helpers/currency-options.ts';
import { useHelpersDateFormatting } from './helpers/date-formatting.ts';
import { useHelpersDateOptions } from './helpers/date-options.ts';
import { useHelpersDatePatterns } from './helpers/date-patterns.ts';
import { useHelpersDisplay } from './helpers/display.ts';
import { useHelpersDomainOptions } from './helpers/domain-options.ts';
import { useHelpersErrors } from './helpers/errors.ts';
import { useHelpersFormatOptions } from './helpers/format-options.ts';
import { useHelpersInitialization } from './helpers/initialization.ts';
import { useHelpersLanguageInfo } from './helpers/language-info.ts';
import { useHelpersLanguage } from './helpers/language.ts';
import { useHelpersMonthNames } from './helpers/month-names.ts';
import { useHelpersNumeralOptions } from './helpers/numeral-options.ts';
import { useHelpersOptions } from './helpers/options.ts';
import { useHelpersPreferences } from './helpers/preferences.ts';
import { useHelpersState } from './helpers/state.ts';
import { useHelpersTranslation } from './helpers/translation.ts';
export * from './helpers/types.ts';






import {
    KeywordMatchMode
} from '@/core/text.ts';

import {
    type BigDecimal,
    type BigDecimalWithSuffix,
    type HiddenAmount,
    DecimalSeparator,
    DigitGroupingSymbol,
    DigitGroupingType,
    NumeralSystem
} from '@/core/numeral.ts';

import {
    CalendarDisplayType,
    CalendarType,
    DateDisplayType
} from '@/core/calendar.ts';

import {
    type DateTime,
    type TextualYearMonthDay,
    KnownDateTimeFormat,
    LongDateFormat,
    LongTimeFormat,
    ShortDateFormat,
    ShortTimeFormat
} from '@/core/datetime.ts';


import {
    CurrencyDisplayType,
    CurrencySortingType
} from '@/core/currency.ts';


import {
    type Coordinate,
    CoordinateDisplayType
} from '@/core/coordinate.ts';





import {
    AccountType,
    CreditCardAmountDisplayType
} from '@/core/account.ts';

import {
    CategoryType
} from '@/core/category.ts';

import {
    TransactionEditScopeType,
    TransactionQuickAddButtonActionType,
    TransactionQuickSaveButtonStyle
} from '@/core/transaction.ts';

import {
    ImportTransactionColumnType
} from '@/core/import_transaction.ts';

import {
    ImportTransactionReplaceRuleConditionField
} from '@/core/rule.ts';

import {
    ScheduledTemplateFrequencyType
} from '@/core/template.ts';

import {
    AccountBalanceTrendChartType,
    CategoricalChartType,
    ChartDataType,
    ChartSortingType,
    StatisticsAnalysisType,
    TrendChartType
} from '@/core/statistics.ts';

import {
    TransactionExplorerChartType,
    TransactionExplorerConditionField,
    TransactionExplorerConditionOperator,
    TransactionExplorerCustomChartDisplayLayout,
    TransactionExplorerDataDimension,
    TransactionExplorerValueMetric
} from '@/core/explorer.ts';






import {
    formatDateTime,
    formatGregorianCalendarYearDashMonthDashDay,
    parseDateTimeFromString
} from '@/lib/datetime.ts';


import {
    appendDigitGroupingSymbolAndDecimalSeparator
} from '@/lib/numeral.ts';








export function useI18n() {
    const { t, locale, getLocaleMessage, setLocaleMessage, settingsStore, userStore, exchangeRatesStore } = useHelpersState();
    const { getLanguageDisplayName, getDefaultLanguage, getLanguageInfo } = useHelpersLanguage({ t });
    const { translateError } = useHelpersErrors({ t });
    const { getDateTimeLocaleData, getChineseCalendarLocaleData, getPersianCalendarLocaleData } = useHelpersCalendarLocales({ locale });
    const {
        getLocalizedDateTimeType, getLocalizedDateTimeFormat, getLocalizedLongDateFormat, getLocalizedShortDateFormat,
        getLocalizedLongYearFormat, getLocalizedShortYearFormat, getLocalizedLongYearMonthFormat,
        getLocalizedShortYearMonthFormat, getLocalizedLongMonthDayFormat, getLocalizedShortMonthDayFormat,
        getLocalizedShortDayFormat, getLocalizedLongTimeFormat, getLocalizedShortTimeFormat,
    } = useHelpersDatePatterns({ t, userStore });
    const {
        getAllLocalizedDigits, getLocaleDefaultCalendarDisplayType, getLocaleDefaultDateDisplayType,
        getLocaleDefaultNumeralSystemType, getLocaleDefaultDecimalSeparator, getLocaleDefaultDigitGroupingSymbol,
        getLocaleDefaultDigitGroupingType, getLocaleFiscalYearFormatType, getCurrentCalendarDisplayType,
        getCurrentDateDisplayType, getCurrentNumeralSystemType, getCurrentDecimalSeparator, getCurrentDigitGroupingSymbol,
        getCurrentDigitGroupingType, getCurrentFiscalYearFormatType, getCurrencyName, getLongDateFormatOrder,
        getShortDateFormatOrder, isLongDateMonthAfterYear, isShortDateMonthAfterYear, isLongTime24HourFormat,
        isLongTimeMeridiemIndicatorFirst, isShortTime24HourFormat, isShortTimeMeridiemIndicatorFirst, isLongTimeHourTwoDigits,
        isLongTimeMinuteTwoDigits, isLongTimeSecondTwoDigits,
    } = useHelpersPreferences({ t, userStore, getLocalizedDateTimeType, getLocalizedLongTimeFormat });
    const { getAllMonthNames, getAllWeekdayNames, getMonthdayOrdinal, getMonthdayShortName, getMonthLastDayShortName, getQuarterName } = useHelpersMonthNames({ t });
    const {
        getDateTimeFormatOptions, getNumberFormatOptions, getCurrencyUnitName, getGregorianLikeCalendarType, formatYearQuarter,
        formatTimeRangeToGregorianLikeFiscalYearFormat, getCurrentCurrencyDisplayType, getCalendarAlternateDisplayDate,
    } = useHelpersFormatOptions({
        getCurrentNumeralSystemType, getCurrentDateDisplayType, getDateTimeLocaleData, getChineseCalendarLocaleData,
        getPersianCalendarLocaleData, getCurrentDigitGroupingType, getCurrentDecimalSeparator, getCurrentDigitGroupingSymbol, t,
        getQuarterName, userStore,
    });
    const {
        getParsedAmountNumber, getFormattedAmount, getFormattedAmountWithCurrency, getFormattedNumber, getFormattedBigDecimal,
        getFormattedPercentValue, getFormattedChartValue, getFormattedVolume, getFormattedExchangeRateAmount,
        getAdaptiveAmountRate, getAmountPrependAndAppendText,
    } = useHelpersAmountFormatting({ getNumberFormatOptions, userStore, getCurrentCurrencyDisplayType, getCurrentNumeralSystemType, getCurrencyName, getCurrencyUnitName });
    const {
        getAllCurrencyDisplayTypes, getAllLocalizedCalendarTypes, getAllImageUploadQualityTypes, getLocalizedDisplayNameAndType,
        getLocalizedNameValue, getLocalizedDisplayNameAndTypeWithSystemDefault, getLocalizedNumeralSeparatorFormats,
        getLocalizedChartDateAggregationTypeAndDisplayName,
    } = useHelpersOptions({ t, userStore, getFormattedAmountWithCurrency, getFormattedNumber });
    const {
        getCurrentLanguageTag, getCurrentLanguageInfo, getCurrentLanguageDisplayName, getCurrentLanguageTextDirection,
        getCurrentLanguageMessagesJson, getDefaultCurrency, getDefaultFirstDayOfWeek, getAllLanguageOptions,
    } = useHelpersLanguageInfo({ locale, getLanguageInfo, getDefaultLanguage, getLocaleMessage, t, getLanguageDisplayName });
    const { translateIf, joinMultiText, formatRange, formatTypeAndNames, getServerMultiLanguageConfigContent } = useHelpersTranslation({ t, getCurrentLanguageTag });
    const { getAllEnableDisableOptions, getAllCurrencies, getAllMeridiemIndicators } = useHelpersCurrencyOptions({ t, getCurrencyName });
    const {
        getAllLongMonthNames, getAllShortMonthNames, getAllLongWeekdayNames, getAllShortWeekdayNames, getAllMinWeekdayNames,
        getAllMonths, getAllWeekDays, getAllHours, getAvailableMonthDays, getLocalizedDateTimeFormats, getAllDateRanges,
        getAllRecentMonthDateRanges, getAllTimezones, getAllTimezoneTypesUsedForStatistics, getAllFiscalYearFormats,
    } = useHelpersDateOptions({
        getAllMonthNames, getAllWeekdayNames, t, getLocalizedShortTimeFormat, getDateTimeFormatOptions, getMonthdayShortName,
        getMonthLastDayShortName, getLocalizedDateTimeFormat, getCurrentCalendarDisplayType, getLocalizedLongYearMonthFormat,
        getCurrentNumeralSystemType, userStore, formatTimeRangeToGregorianLikeFiscalYearFormat,
    });
    const { getAllNumeralSystemTypes, getAllDigitGroupingTypes } = useHelpersNumeralOptions({ t });
    const {
        getAllExpenseIncomeAmountColors, getAllAccountCategories, getAllTransactionDefaultCategories, getAllDisplayExchangeRates,
        getAllSupportedImportFileCagtegoryAndTypes,
    } = useHelpersDomainOptions({ t, getCurrencyName, settingsStore, getCurrentLanguageTag, getLanguageDisplayName });
    const {
        getMonthShortName, getMonthLongName, getWeekdayShortName, getWeekdayLongName, getMultiMonthAndDayLongNames,
        getMultiMonthdayShortNames, getMultiWeekdayLongNames,
    } = useHelpersCalendarLabels({ t, getLocalizedLongMonthDayFormat, getDateTimeFormatOptions, joinMultiText, getMonthdayOrdinal, getAllWeekDays });
    const {
        formatGregorianTextualMonthDayToGregorianLikeLongMonthDay, formatDateTimeToGregorianLikeYearQuarter,
        formatYearQuarterToGregorianLikeYearQuarter, formatUnixTimeToGregorianLikeFiscalYear,
        formatGregorianYearToGregorianLikeFiscalYear, formatFiscalYearStartToGregorianLikeLongMonth, formatDateRange,
        getTimezoneDifferenceDisplayText, getCalendarAlternateDates, getCalendarAlternateDate,
    } = useHelpersDateFormatting({
        getGregorianLikeCalendarType, getLocalizedLongMonthDayFormat, getDateTimeFormatOptions, getLocalizedShortYearFormat,
        formatYearQuarter, getCurrentFiscalYearFormatType, userStore, formatTimeRangeToGregorianLikeFiscalYearFormat, t,
        formatRange, getLocalizedShortYearMonthFormat, getLocalizedShortDateFormat, getLocalizedShortMonthDayFormat,
        getCurrentNumeralSystemType, getCurrentCalendarDisplayType, getChineseCalendarLocaleData, getCalendarAlternateDisplayDate,
    });
    const { getCategorizedAccountsWithDisplayBalance } = useHelpersAccountBalances({ userStore, getFormattedAmountWithCurrency, exchangeRatesStore });
    const { getTablePageOptions, getFormattedCoordinate, getLocalizedFileEncodingName, getLocalizedOAuth2ProviderName, getLocalizedOAuth2LoginText } = useHelpersDisplay({ getCurrentNumeralSystemType, t, userStore, getNumberFormatOptions, getServerMultiLanguageConfigContent });
    const { setLanguage, hasPreviewedConfiguration, previewLanguage, restoreLanguagePreview, setTimeZone, initLocale } = useHelpersInitialization({
        getDefaultLanguage, getLanguageInfo, locale, getAllLongMonthNames, getAllShortMonthNames, getAllLongWeekdayNames,
        getAllShortWeekdayNames, getAllMinWeekdayNames, t, getDefaultCurrency, getDefaultFirstDayOfWeek, setLocaleMessage,
        settingsStore,
    });

    return {
        // common functions
        tt: t,
        ti: translateIf,
        te: translateError,
        joinMultiText,
        formatRange,
        formatTypeAndNames,
        getServerMultiLanguageConfigContent,
        // get current language info
        getCurrentLanguageTag,
        getCurrentLanguageInfo,
        getCurrentLanguageDisplayName,
        getCurrentLanguageTextDirection,
        getCurrentLanguageMessagesJson,
        // get localization default type
        getDefaultCurrency,
        getDefaultFirstDayOfWeek,
        // get all localized info of specified type
        getAllLanguageOptions,
        getAllEnableDisableOptions,
        getAllCurrencies,
        getAllMeridiemIndicators,
        getAllLongMonthNames,
        getAllShortMonthNames,
        getAllLongWeekdayNames,
        getAllShortWeekdayNames,
        getAllMinWeekdayNames,
        getAllMonths,
        getAllWeekDays,
        getAllHours,
        getAvailableMonthDays,
        getAllCalendarDisplayTypes: () => getAllLocalizedCalendarTypes(CalendarDisplayType.values(), CalendarDisplayType.parse(t('default.calendarDisplayType')), CalendarDisplayType.Default, CalendarDisplayType.LanguageDefaultType),
        getAllDateDisplayTypes: () => getAllLocalizedCalendarTypes(DateDisplayType.values(), DateDisplayType.parse(t('default.dateDisplayType')), DateDisplayType.Default, DateDisplayType.LanguageDefaultType),
        getAllLongDateFormats: (numeralSystem: NumeralSystem, calendarType: CalendarType) => getLocalizedDateTimeFormats<LongDateFormat>('longDate', LongDateFormat.all(), LongDateFormat.values(), 'longDateFormat', LongDateFormat.Default, numeralSystem, calendarType),
        getAllShortDateFormats: (numeralSystem: NumeralSystem, calendarType: CalendarType) => getLocalizedDateTimeFormats<ShortDateFormat>('shortDate', ShortDateFormat.all(), ShortDateFormat.values(), 'shortDateFormat', ShortDateFormat.Default, numeralSystem, calendarType),
        getAllLongTimeFormats: (numeralSystem: NumeralSystem) => getLocalizedDateTimeFormats<LongTimeFormat>('longTime', LongTimeFormat.all(), LongTimeFormat.values(), 'longTimeFormat', LongTimeFormat.Default, numeralSystem),
        getAllShortTimeFormats: (numeralSystem: NumeralSystem) => getLocalizedDateTimeFormats<ShortTimeFormat>('shortTime', ShortTimeFormat.all(), ShortTimeFormat.values(), 'shortTimeFormat', ShortTimeFormat.Default, numeralSystem),
        getAllFiscalYearFormats,
        getAllDateRanges,
        getAllRecentMonthDateRanges,
        getAllTimezones,
        getAllTimezoneTypesUsedForStatistics,
        getAllNumeralSystemTypes,
        getAllDecimalSeparators: () => getLocalizedNumeralSeparatorFormats(DecimalSeparator.values(), DecimalSeparator.parse(t('default.decimalSeparator')), DecimalSeparator.Default, DecimalSeparator.LanguageDefaultType),
        getAllDigitGroupingSymbols: () => getLocalizedNumeralSeparatorFormats(DigitGroupingSymbol.values(), DigitGroupingSymbol.parse(t('default.digitGroupingSymbol')), DigitGroupingSymbol.Default, DigitGroupingSymbol.LanguageDefaultType),
        getAllDigitGroupingTypes,
        getAllCurrencyDisplayTypes,
        getAllCurrencySortingTypes: () => getLocalizedDisplayNameAndType(CurrencySortingType.values()),
        getAllCoordinateDisplayTypes: () => getLocalizedDisplayNameAndTypeWithSystemDefault(CoordinateDisplayType.values(), CoordinateDisplayType.SystemDefaultType, CoordinateDisplayType.Default),
        getAllKeywordMatchModes: () => getLocalizedDisplayNameAndType(KeywordMatchMode.values()),
        getAllImageUploadQualityTypes,
        getAllExpenseAmountColors: () => getAllExpenseIncomeAmountColors(CategoryType.Expense),
        getAllIncomeAmountColors: () => getAllExpenseIncomeAmountColors(CategoryType.Income),
        getAllAccountCategories,
        getAllAccountTypes: () => getLocalizedDisplayNameAndType(AccountType.values()),
        getAllCreditCardAmountDisplayTypes: () => getLocalizedDisplayNameAndType(CreditCardAmountDisplayType.values()),
        getAllCategoricalChartTypes: (withDesktopOnlyChart?: boolean) => getLocalizedDisplayNameAndType(CategoricalChartType.values(!!withDesktopOnlyChart)),
        getAllTrendChartTypes: () => getLocalizedDisplayNameAndType(TrendChartType.values()),
        getAllAccountBalanceTrendChartTypes: () => getLocalizedDisplayNameAndType(AccountBalanceTrendChartType.values()),
        getAllStatisticsChartDataTypes: (analysisType: StatisticsAnalysisType, withDesktopOnlyChart?: boolean) => getLocalizedDisplayNameAndType(ChartDataType.values(analysisType, withDesktopOnlyChart)),
        getAllStatisticsSortingTypes: (useAlternativeName?: boolean) => getLocalizedDisplayNameAndType(ChartSortingType.values(), useAlternativeName),
        getAllStatisticsDateAggregationTypes: (analysisType: StatisticsAnalysisType, includeBillingCycle: boolean) => getLocalizedChartDateAggregationTypeAndDisplayName(analysisType, true, includeBillingCycle),
        getAllStatisticsDateAggregationTypesWithShortName: (analysisType: StatisticsAnalysisType, includeBillingCycle: boolean) => getLocalizedChartDateAggregationTypeAndDisplayName(analysisType, false, includeBillingCycle),
        getAllTransactionEditScopeTypes: (useLastReconciledTime: boolean) => getLocalizedDisplayNameAndType(TransactionEditScopeType.values(useLastReconciledTime)),
        getAllTransactionQuickSaveButtonStyles: () => getLocalizedDisplayNameAndType(TransactionQuickSaveButtonStyle.values()),
        getAllTransactionQuickAddButtonActionTypes: () => getLocalizedDisplayNameAndType(TransactionQuickAddButtonActionType.values()),
        getAllTransactionScheduledFrequencyTypes: () => getLocalizedDisplayNameAndType(ScheduledTemplateFrequencyType.values()),
        getAllImportTransactionColumnTypes: () => getLocalizedDisplayNameAndType(ImportTransactionColumnType.values()),
        getAllImportTransactionReplaceRuleConditionFields: () => getLocalizedNameValue(ImportTransactionReplaceRuleConditionField.values()),
        getAllTransactionDefaultCategories,
        getAllDisplayExchangeRates,
        getAllSupportedImportFileCagtegoryAndTypes,
        getAllTransactionExplorerConditionFields: () => getLocalizedNameValue(TransactionExplorerConditionField.values()),
        getAllTransactionExplorerConditionOperators: (operators?: TransactionExplorerConditionOperator[]) => getLocalizedNameValue(operators ?? TransactionExplorerConditionOperator.values()),
        getAllTransactionExplorerDataDimensions: (operators?: TransactionExplorerDataDimension[]) => getLocalizedNameValue(operators ?? TransactionExplorerDataDimension.values()),
        getAllTransactionExplorerValueMetrics: (operators?: TransactionExplorerValueMetric[]) => getLocalizedNameValue(operators ?? TransactionExplorerValueMetric.values()),
        getAllTransactionExplorerChartTypes: (operators?: TransactionExplorerChartType[]) => getLocalizedNameValue(operators ?? TransactionExplorerChartType.values()),
        getAllTransactionExplorerCustomChartDisplayLayouts: () => getLocalizedDisplayNameAndType(TransactionExplorerCustomChartDisplayLayout.values()),
        // get localized info
        getLanguageInfo,
        getEnableDisableOption: (value: boolean) => t(value ? 'Enabled' : 'Disabled'),
        getMonthShortName,
        getMonthLongName,
        getMonthdayOrdinal,
        getMonthdayShortName,
        getWeekdayShortName,
        getWeekdayLongName,
        getQuarterName,
        getMultiMonthAndDayLongNames,
        getMultiMonthdayShortNames,
        getMultiWeekdayLongNames,
        getAllLocalizedDigits,
        getLocaleDefaultCalendarDisplayType,
        getLocaleDefaultDateDisplayType,
        getLocaleDefaultNumeralSystemType,
        getLocaleDefaultDecimalSeparator,
        getLocaleDefaultDigitGroupingSymbol,
        getLocaleDefaultDigitGroupingType,
        getLocaleFiscalYearFormatType,
        getCurrentCalendarDisplayType,
        getCurrentDateDisplayType,
        getCurrentNumeralSystemType,
        getCurrentDecimalSeparator,
        getCurrentDigitGroupingSymbol,
        getCurrentDigitGroupingType,
        getCurrentFiscalYearFormatType,
        getCurrencyName,
        getLongDateFormatOrder,
        getShortDateFormatOrder,
        isLongDateMonthAfterYear,
        isShortDateMonthAfterYear,
        isLongTime24HourFormat,
        isLongTimeMeridiemIndicatorFirst,
        isShortTime24HourFormat,
        isShortTimeMeridiemIndicatorFirst,
        isLongTimeHourTwoDigits,
        isLongTimeMinuteTwoDigits,
        isLongTimeSecondTwoDigits,
        // format date time (by calendar display type) functions
        getCalendarDisplayLongYearFromDateTime: (dateTime: DateTime, numeralSystem?: NumeralSystem) => formatDateTime(dateTime, getLocalizedLongYearFormat(), getDateTimeFormatOptions({ calendarType: getCurrentCalendarDisplayType().primaryCalendarType, numeralSystem: numeralSystem })),
        getCalendarDisplayShortYearFromDateTime: (dateTime: DateTime, numeralSystem?: NumeralSystem) => formatDateTime(dateTime, getLocalizedShortYearFormat(), getDateTimeFormatOptions({ calendarType: getCurrentCalendarDisplayType().primaryCalendarType, numeralSystem: numeralSystem })),
        getCalendarDisplayShortMonthFromDateTime: (dateTime: DateTime, numeralSystem?: NumeralSystem) => formatDateTime(dateTime, 'MMM', getDateTimeFormatOptions({ calendarType: getCurrentCalendarDisplayType().primaryCalendarType, numeralSystem: numeralSystem })),
        getCalendarDisplayDayOfMonthFromDateTime: (dateTime: DateTime, numeralSystem?: NumeralSystem) => formatDateTime(dateTime, getLocalizedShortDayFormat(), getDateTimeFormatOptions({ calendarType: getCurrentCalendarDisplayType().primaryCalendarType, numeralSystem: numeralSystem })),
        // format date time (by date display type) functions
        parseDateTimeFromLongDateTime: (dateTime: string) => parseDateTimeFromString(dateTime, getLocalizedLongDateFormat() + ' ' + getLocalizedLongTimeFormat()),
        parseDateTimeFromShortDateTime: (dateTime: string) => parseDateTimeFromString(dateTime, getLocalizedShortDateFormat() + ' ' + getLocalizedShortTimeFormat()),
        formatDateTimeToLongDateTime: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedLongDateFormat() + ' ' + getLocalizedLongTimeFormat(), getDateTimeFormatOptions()),
        formatDateTimeToShortDateTime: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedShortDateFormat() + ' ' + getLocalizedShortTimeFormat(), getDateTimeFormatOptions()),
        formatDateTimeToLongDate: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedLongDateFormat(), getDateTimeFormatOptions()),
        formatDateTimeToShortDate: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedShortDateFormat(), getDateTimeFormatOptions()),
        formatDateTimeToLongMonthDay: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedLongMonthDayFormat(), getDateTimeFormatOptions()),
        formatDateTimeToShortMonthDay: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedShortMonthDayFormat(), getDateTimeFormatOptions()),
        formatDateTimeToLongTime: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedLongTimeFormat(), getDateTimeFormatOptions()),
        formatDateTimeToShortTime: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedShortTimeFormat(), getDateTimeFormatOptions()),
        formatGregorianTextualYearMonthDayToLongDate: (date: TextualYearMonthDay) => formatGregorianCalendarYearDashMonthDashDay(date, getLocalizedLongDateFormat(), getDateTimeFormatOptions()),
        // format date time (Gregorian calendar and Gregorian-like calendar) functions
        formatDateTimeToGregorianLikeLongYear: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedLongYearFormat(), getDateTimeFormatOptions({ calendarType: getGregorianLikeCalendarType() })),
        formatDateTimeToGregorianLikeShortYear: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedShortYearFormat(), getDateTimeFormatOptions({ calendarType: getGregorianLikeCalendarType() })),
        formatDateTimeToGregorianLikeLongYearMonth: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedLongYearMonthFormat(), getDateTimeFormatOptions({ calendarType: getGregorianLikeCalendarType() })),
        formatDateTimeToGregorianLikeShortYearMonth: (dateTime: DateTime) => formatDateTime(dateTime, getLocalizedShortYearMonthFormat(), getDateTimeFormatOptions({ calendarType: getGregorianLikeCalendarType() })),
        formatDateTimeToGregorianLikeLongMonth: (dateTime: DateTime) => formatDateTime(dateTime, 'MMMM', getDateTimeFormatOptions({ calendarType: getGregorianLikeCalendarType() })),
        formatDateTimeToGregorianLikeShortMonth: (dateTime: DateTime) => formatDateTime(dateTime, 'MMM', getDateTimeFormatOptions({ calendarType: getGregorianLikeCalendarType() })),
        formatGregorianTextualMonthDayToGregorianLikeLongMonthDay,
        formatDateTimeToGregorianLikeYearQuarter,
        formatYearQuarterToGregorianLikeYearQuarter,
        formatDateTimeToGregorianLikeFiscalYear: (dateTime: DateTime) => formatUnixTimeToGregorianLikeFiscalYear(dateTime.getUnixTime()),
        formatGregorianYearToGregorianLikeFiscalYear,
        formatFiscalYearStartToGregorianLikeLongMonth,
        // format date time (Gregorian calendar) functions
        formatDateTimeToGregorianDefaultDateTime: (dateTime: DateTime) => formatDateTime(dateTime, KnownDateTimeFormat.DefaultDateTime.format, getDateTimeFormatOptions({ numeralSystem: NumeralSystem.WesternArabicNumerals, calendarType: CalendarType.Gregorian })),
        // other format date time functions
        formatDateRange,
        getTimezoneDifferenceDisplayText,
        getCalendarAlternateDates,
        getCalendarAlternateDate,
        // format amount/number functions
        parseAmountFromLocalizedNumerals: (value: string) => getParsedAmountNumber(value),
        parseAmountFromWesternArabicNumerals: (value: string) => getParsedAmountNumber(value, NumeralSystem.WesternArabicNumerals),
        formatAmountToLocalizedNumerals: (value: BigDecimal, currencyCode?: string) => getFormattedAmount(value, undefined, undefined, currencyCode),
        formatAmountToWesternArabicNumerals: (value: BigDecimal, currencyCode?: string) => getFormattedAmount(value, NumeralSystem.WesternArabicNumerals, undefined, currencyCode),
        formatAmountToLocalizedNumeralsWithoutDigitGrouping: (value: BigDecimal, currencyCode?: string) => getFormattedAmount(value, undefined, DigitGroupingType.None, currencyCode),
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping: (value: BigDecimal, currencyCode?: string) => getFormattedAmount(value, NumeralSystem.WesternArabicNumerals, DigitGroupingType.None, currencyCode),
        formatAmountToLocalizedNumeralsWithCurrency: (value: BigDecimal | HiddenAmount | BigDecimalWithSuffix, currencyCode?: string | false, currencyDisplayType?: CurrencyDisplayType) => getFormattedAmountWithCurrency(value, currencyCode, currencyDisplayType),
        formatAmountToWesternArabicNumeralsWithCurrency: (value: BigDecimal | HiddenAmount | BigDecimalWithSuffix, currencyCode?: string | false, currencyDisplayType?: CurrencyDisplayType) => getFormattedAmountWithCurrency(value, currencyCode, currencyDisplayType, NumeralSystem.WesternArabicNumerals),
        formatBigDecimalToLocalizedNumerals: (value: BigDecimal, precision?: number) => getFormattedBigDecimal(value, undefined, undefined, precision),
        formatBigDecimalToLocalizedNumeralsWithoutDigitGrouping: (value: BigDecimal, precision?: number) => getFormattedBigDecimal(value, undefined, DigitGroupingType.None, precision),
        formatBigDecimalToWesternArabicNumerals: (value: BigDecimal, precision?: number) => getFormattedBigDecimal(value, NumeralSystem.WesternArabicNumerals, undefined, precision),
        formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping: (value: BigDecimal, precision?: number) => getFormattedBigDecimal(value, NumeralSystem.WesternArabicNumerals, DigitGroupingType.None, precision),
        formatNumberToLocalizedNumerals: (value: number, precision?: number) => getFormattedNumber(value, undefined, undefined, precision),
        formatNumberToLocalizedNumeralsWithoutDigitGrouping: (value: number, precision?: number) => getFormattedNumber(value, undefined, DigitGroupingType.None, precision),
        formatNumberToWesternArabicNumeralsWithoutDigitGrouping: (value: number, precision?: number) => getFormattedNumber(value, NumeralSystem.WesternArabicNumerals, DigitGroupingType.None, precision),
        formatPercentToLocalizedNumerals: (value: number, precision: number, lowPrecisionValue: string) => getFormattedPercentValue(value, precision, lowPrecisionValue),
        formatPercentToWesternArabicNumerals: (value: number, precision: number, lowPrecisionValue: string) => getFormattedPercentValue(value, precision, lowPrecisionValue, NumeralSystem.WesternArabicNumerals),
        formatChartValueToLocalizedNumerals: getFormattedChartValue,
        formatVolumeToLocalizedNumerals: getFormattedVolume,
        formatExchangeRateAmountToWesternArabicNumerals: (value: BigDecimal) => getFormattedExchangeRateAmount(value, NumeralSystem.WesternArabicNumerals),
        appendDigitGroupingSymbolAndDecimalSeparator: (value: string) => appendDigitGroupingSymbolAndDecimalSeparator(value, getNumberFormatOptions({})),
        getAdaptiveAmountRate,
        getAmountPrependAndAppendText,
        getCategorizedAccountsWithDisplayBalance,
        getTablePageOptions,
        // other format functions
        formatCoordinate: (value: Coordinate) => getFormattedCoordinate(value),
        getLocalizedFileEncodingName,
        getLocalizedOAuth2ProviderName,
        getLocalizedOAuth2LoginText,
        // localization setting functions
        setLanguage,
        hasPreviewedConfiguration,
        previewLanguage,
        restoreLanguagePreview,
        setTimeZone,
        initLocale
    };
}
