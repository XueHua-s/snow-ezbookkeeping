import { DEFAULT_EXPENSE_CATEGORIES, DEFAULT_INCOME_CATEGORIES, DEFAULT_TRANSFER_CATEGORIES } from '@/consts/category.ts';
import { DEFAULT_DOCUMENT_LANGUAGE_FOR_IMPORT_FILE, SUPPORTED_DOCUMENT_LANGUAGES_FOR_IMPORT_FILE, SUPPORTED_IMPORT_FILE_CATEGORY_AND_TYPES } from '@/consts/file.ts';
import {
    type LocalizedAccountCategory,
    AccountCategory
} from '@/core/account.ts';
import {
    type TypeAndDisplayName
} from '@/core/base.ts';
import {
    type LocalizedPresetCategory,
    type LocalizedPresetSubCategory,
    type PresetCategory,
    ALL_CATEGORY_TYPES,
    CategoryType
} from '@/core/category.ts';
import {
    PresetAmountColor
} from '@/core/color.ts';
import {
    CurrencySortingType
} from '@/core/currency.ts';
import {
    type LocalizedImportFileCategoryAndTypes,
    type LocalizedImportFileDocument,
    type LocalizedImportFileType,
    type LocalizedImportFileTypeSubType,
    type LocalizedImportFileTypeSupportedEncodings
} from '@/core/file.ts';
import {
    IconType
} from '@/core/icon.ts';
import {
    isString
} from '@/lib/common.ts';
import {
    type LanguageInfo,
    ALL_LANGUAGES,
    DEFAULT_LANGUAGE
} from '@/locales/index.ts';
import type { LatestExchangeRateResponse, LocalizedLatestExchangeRate } from '@/models/exchange_rate.ts';
import type { useHelpersLanguageInfo } from './language-info.ts';
import type { useHelpersLanguage } from './language.ts';
import type { useHelpersPreferences } from './preferences.ts';
import type { useHelpersState } from './state.ts';

interface Context {
    t: ReturnType<typeof useHelpersState>['t'];
    getCurrencyName: ReturnType<typeof useHelpersPreferences>['getCurrencyName'];
    settingsStore: ReturnType<typeof useHelpersState>['settingsStore'];
    getCurrentLanguageTag: ReturnType<typeof useHelpersLanguageInfo>['getCurrentLanguageTag'];
    getLanguageDisplayName: ReturnType<typeof useHelpersLanguage>['getLanguageDisplayName'];
}

// Owns domain options for the helpers store.
export function useHelpersDomainOptions(context: Context) {
    const { t, getCurrencyName, settingsStore, getCurrentLanguageTag, getLanguageDisplayName } = context;

    function getAllExpenseIncomeAmountColors(categoryType: CategoryType): TypeAndDisplayName[] {
        const ret: TypeAndDisplayName[] = [];
        let defaultAmountName = '';

        if (categoryType === CategoryType.Expense) {
            defaultAmountName = PresetAmountColor.DefaultExpenseColor.name;
        } else if (categoryType === CategoryType.Income) { // income
            defaultAmountName = PresetAmountColor.DefaultIncomeColor.name;
        }

        if (defaultAmountName) {
            defaultAmountName = t('color.amount.' + defaultAmountName);
        }

        ret.push({
            type: PresetAmountColor.SystemDefaultType,
            displayName: t('System Default') + (defaultAmountName ? ` (${defaultAmountName})` : '')
        });

        const allPresetAmountColors = PresetAmountColor.values();

        for (const amountColor of allPresetAmountColors) {
            ret.push({
                type: amountColor.type,
                displayName: t('color.amount.' + amountColor.name)
            });
        }

        return ret;
    }

    function getAllAccountCategories(customAccountCategoryOrder: string): LocalizedAccountCategory[] {
        const ret: LocalizedAccountCategory[] = [];
        const allCategories = AccountCategory.values(customAccountCategoryOrder);

        for (const accountCategory of allCategories) {
            ret.push({
                type: accountCategory.type,
                displayName: t(accountCategory.name),
                defaultAccountIconId: accountCategory.defaultAccountIconId
            });
        }

        return ret;
    }

    function getAllTransactionDefaultCategories(categoryType: 0 | CategoryType, locale: string): Record<string, LocalizedPresetCategory[]> {
        const allCategories: Record<string, LocalizedPresetCategory[]> = {};
        const categoryTypes: CategoryType[] = [];

        if (categoryType === 0) {
            categoryTypes.push(...ALL_CATEGORY_TYPES);
        } else {
            categoryTypes.push(categoryType);
        }

        for (const categoryType of categoryTypes) {
            const categories: LocalizedPresetCategory[] = [];
            let defaultCategories: PresetCategory[] = [];

            if (categoryType === CategoryType.Income) {
                defaultCategories = DEFAULT_INCOME_CATEGORIES;
            } else if (categoryType === CategoryType.Expense) {
                defaultCategories = DEFAULT_EXPENSE_CATEGORIES;
            } else if (categoryType === CategoryType.Transfer) {
                defaultCategories = DEFAULT_TRANSFER_CATEGORIES;
            }

            for (const category of defaultCategories) {
                const submitCategory: LocalizedPresetCategory = {
                    name: t('category.' + category.name, {}, { locale: locale }),
                    type: categoryType,
                    icon: category.categoryIconId,
                    iconType: IconType.System,
                    color: category.color,
                    subCategories: []
                };

                for (const subCategory of category.subCategories) {
                    const submitSubCategory: LocalizedPresetSubCategory = {
                        name: t('category.' + subCategory.name, {}, { locale: locale }),
                        type: categoryType,
                        icon: subCategory.categoryIconId,
                        iconType: IconType.System,
                        color: subCategory.color
                    };

                    submitCategory.subCategories.push(submitSubCategory);
                }

                categories.push(submitCategory);
            }

            allCategories[`${categoryType}`] = categories;
        }

        return allCategories;
    }

    function getAllDisplayExchangeRates(exchangeRatesData?: LatestExchangeRateResponse): LocalizedLatestExchangeRate[] {
        const availableExchangeRates: LocalizedLatestExchangeRate[] = [];

        if (!exchangeRatesData || !exchangeRatesData.exchangeRates) {
            return availableExchangeRates;
        }

        for (const exchangeRate of exchangeRatesData.exchangeRates) {
            availableExchangeRates.push({
                currencyCode: exchangeRate.currency,
                currencyDisplayName: getCurrencyName(exchangeRate.currency),
                rate: exchangeRate.rate
            });
        }

        if (settingsStore.appSettings.currencySortByInExchangeRatesPage === CurrencySortingType.Name.type) {
            availableExchangeRates.sort(function (c1, c2) {
                return c1.currencyDisplayName.localeCompare(c2.currencyDisplayName);
            });
        } else if (settingsStore.appSettings.currencySortByInExchangeRatesPage === CurrencySortingType.CurrencyCode.type) {
            availableExchangeRates.sort(function (c1, c2) {
                return c1.currencyCode.localeCompare(c2.currencyCode);
            });
        } else if (settingsStore.appSettings.currencySortByInExchangeRatesPage === CurrencySortingType.ExchangeRate.type) {
            availableExchangeRates.sort(function (c1, c2) {
                const rate1 = parseFloat(c1.rate);
                const rate2 = parseFloat(c2.rate);

                if (rate1 > rate2) {
                    return 1;
                } else if (rate1 < rate2) {
                    return -1;
                } else {
                    return 0;
                }
            });
        }

        return availableExchangeRates;
    }

    function getAllSupportedImportFileCagtegoryAndTypes(supportAITextRecognition: boolean, supportAIImageRecognition: boolean): LocalizedImportFileCategoryAndTypes[] {
        const allSupportedImportFileCategoryAndTypes: LocalizedImportFileCategoryAndTypes[] = [];

        for (const categoryAndTypes of SUPPORTED_IMPORT_FILE_CATEGORY_AND_TYPES) {
            const localizedCategoryAndTypes: LocalizedImportFileCategoryAndTypes = {
                displayCategoryName: t(categoryAndTypes.categoryName),
                fileTypes: []
            };

            for (const fileType of categoryAndTypes.fileTypes) {
                if (fileType.needAITextRecognition && !supportAITextRecognition) {
                    continue;
                }

                if (fileType.needAIImageRecognition && !supportAIImageRecognition) {
                    continue;
                }

                let document: LocalizedImportFileDocument | undefined;

                if (fileType.document) {
                    let documentLanguage = '';
                    let documentDisplayLanguageName = '';
                    let documentAnchor = '';

                    if (fileType.document.supportMultiLanguages === true) {
                        documentLanguage = getCurrentLanguageTag();

                        if (SUPPORTED_DOCUMENT_LANGUAGES_FOR_IMPORT_FILE[documentLanguage] === documentLanguage) {
                            documentAnchor = t(`document.anchor.export_and_import.${fileType.document.anchor}`);
                        } else if (SUPPORTED_DOCUMENT_LANGUAGES_FOR_IMPORT_FILE[documentLanguage]) {
                            documentLanguage = SUPPORTED_DOCUMENT_LANGUAGES_FOR_IMPORT_FILE[documentLanguage] as string;
                            documentAnchor = t(`document.anchor.export_and_import.${fileType.document.anchor}`, {}, { locale: documentLanguage });
                        } else {
                            documentLanguage = DEFAULT_DOCUMENT_LANGUAGE_FOR_IMPORT_FILE;
                            documentAnchor = t(`document.anchor.export_and_import.${fileType.document.anchor}`, {}, { locale: documentLanguage });
                        }
                    } else if (isString(fileType.document.supportMultiLanguages) && ALL_LANGUAGES[fileType.document.supportMultiLanguages]) {
                        documentLanguage = fileType.document.supportMultiLanguages;
                        documentAnchor = fileType.document.anchor;
                    }

                    if (documentLanguage && documentLanguage !== getCurrentLanguageTag() && ALL_LANGUAGES[documentLanguage]) {
                        documentDisplayLanguageName = getLanguageDisplayName((ALL_LANGUAGES[documentLanguage] as LanguageInfo).name);
                    }

                    if (documentLanguage) {
                        documentLanguage = documentLanguage.replace(/-/g, '_');
                    }

                    if (documentAnchor) {
                        documentAnchor = documentAnchor.toLowerCase().replace(/ /g, '-');
                    }

                    if (documentLanguage === DEFAULT_LANGUAGE) {
                        documentLanguage = '';
                    }

                    document = {
                        language: documentLanguage,
                        displayLanguageName: documentDisplayLanguageName,
                        anchor: documentAnchor
                    };
                } else {
                    document = undefined;
                }

                const subTypes: LocalizedImportFileTypeSubType[] = [];

                if (fileType.subTypes) {
                    for (const subType of fileType.subTypes) {
                        const localizedSubType: LocalizedImportFileTypeSubType = {
                            type: subType.type,
                            displayName: t(subType.name),
                            extensions: subType.extensions
                        };

                        subTypes.push(localizedSubType);
                    }
                }

                const supportedEncodings: LocalizedImportFileTypeSupportedEncodings[] = [];

                if (fileType.supportedEncodings) {
                    for (const encoding of fileType.supportedEncodings) {
                        const localizedEncoding: LocalizedImportFileTypeSupportedEncodings = {
                            encoding: encoding,
                            displayName: t(`encoding.${encoding}`)
                        };

                        supportedEncodings.push(localizedEncoding);
                    }
                }

                const localizedFileType: LocalizedImportFileType = {
                    type: fileType.type,
                    displayName: t(fileType.name),
                    extensions: fileType.extensions,
                    subTypes: subTypes.length ? subTypes : undefined,
                    supportedEncodings: supportedEncodings.length ? supportedEncodings : undefined,
                    dataFromTextbox: fileType.dataFromTextbox,
                    needAITextRecognition: fileType.needAITextRecognition,
                    needAIImageRecognition: fileType.needAIImageRecognition,
                    supportedAdditionalOptions: fileType.supportedAdditionalOptions,
                    supportedAIAdditionalPrompt: fileType.supportedAIAdditionalPrompt,
                    document: document
                };

                localizedCategoryAndTypes.fileTypes.push(localizedFileType);
            }

            allSupportedImportFileCategoryAndTypes.push(localizedCategoryAndTypes);
        }

        return allSupportedImportFileCategoryAndTypes;
    }

    return {
        getAllExpenseIncomeAmountColors, getAllAccountCategories, getAllTransactionDefaultCategories, getAllDisplayExchangeRates,
        getAllSupportedImportFileCagtegoryAndTypes,
    };
}
