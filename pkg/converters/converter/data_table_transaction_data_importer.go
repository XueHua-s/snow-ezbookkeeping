package converter

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/validators"
	"sort"
	"strings"
	"time"
)

type TransactionGeoLocationOrder string

const (
	TRANSACTION_GEO_LOCATION_ORDER_LONGITUDE_LATITUDE TransactionGeoLocationOrder = "lonlat" // longitude first, then latitude
	TRANSACTION_GEO_LOCATION_ORDER_LATITUDE_LONGITUDE TransactionGeoLocationOrder = "latlon" // latitude first, then longitude
)

// DataTableTransactionDataImporter defines the structure of plain text data table importer for transaction data
type DataTableTransactionDataImporter struct {
	transactionTypeMapping  map[string]models.TransactionType
	geoLocationSeparator    string
	geoLocationOrder        TransactionGeoLocationOrder
	transactionTagSeparator string
}

// ParseImportedData returns the imported transaction data
func (c *DataTableTransactionDataImporter) ParseImportedData(ctx core.Context, user *models.User, dataTable datatable.TransactionDataTable, defaultTimezone *time.Location, additionalOptions TransactionDataImporterOptions, accountMap map[string]*models.Account, expenseCategoryMap map[string]map[string]*models.TransactionCategory, incomeCategoryMap map[string]map[string]*models.TransactionCategory, transferCategoryMap map[string]map[string]*models.TransactionCategory, tagMap map[string]*models.TransactionTag) (models.ImportedTransactionSlice, []*models.Account, []*models.TransactionCategory, []*models.TransactionCategory, []*models.TransactionCategory, []*models.TransactionTag, error) {
	if dataTable.TransactionRowCount() < 1 {
		log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse import data for user \"uid:%d\", because data table row count is less 1", user.Uid)
		return nil, nil, nil, nil, nil, nil, errs.ErrNotFoundTransactionDataInFile
	}

	nameDbTypeMap, err := c.buildTransactionTypeNameDbTypeMap()

	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	if !dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME) ||
		!dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE) ||
		!dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY) ||
		!dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME) ||
		!dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_AMOUNT) ||
		!dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME) {
		log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse import data for user \"uid:%d\", because missing essential columns in header row", user.Uid)
		return nil, nil, nil, nil, nil, nil, errs.ErrMissingRequiredFieldInHeaderRow
	}

	if accountMap == nil {
		accountMap = make(map[string]*models.Account)
	}

	if expenseCategoryMap == nil {
		expenseCategoryMap = make(map[string]map[string]*models.TransactionCategory)
	}

	if incomeCategoryMap == nil {
		incomeCategoryMap = make(map[string]map[string]*models.TransactionCategory)
	}

	if transferCategoryMap == nil {
		transferCategoryMap = make(map[string]map[string]*models.TransactionCategory)
	}

	if tagMap == nil {
		tagMap = make(map[string]*models.TransactionTag)
	}

	allNewTransactions := make(models.ImportedTransactionSlice, 0, dataTable.TransactionRowCount())
	allNewAccounts := make([]*models.Account, 0)
	allNewSubExpenseCategories := make([]*models.TransactionCategory, 0)
	allNewSubIncomeCategories := make([]*models.TransactionCategory, 0)
	allNewSubTransferCategories := make([]*models.TransactionCategory, 0)
	allNewTags := make([]*models.TransactionTag, 0)

	dataRowIterator := dataTable.TransactionRowIterator()
	dataRowIndex := 0

	for dataRowIterator.HasNext() {
		dataRowIndex++
		dataRow, err := dataRowIterator.Next(ctx, user)

		if err != nil {
			log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse data row \"index:%d\" for user \"uid:%d\", because %s", dataRowIndex, user.Uid, err.Error())
			return nil, nil, nil, nil, nil, nil, err
		}

		if !dataRow.IsValid() {
			continue
		}

		timezone := defaultTimezone

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE) &&
			dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE) != datatable.TRANSACTION_DATA_TABLE_TIMEZONE_NOT_AVAILABLE {
			transactionTimezone, err := utils.ParseFromTimezoneOffset(dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE))

			if err != nil {
				log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse time zone \"%s\" in data row \"index:%d\" for user \"uid:%d\", because %s", dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE), dataRowIndex, user.Uid, err.Error())
				return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrTransactionTimeZoneInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE))
			}

			timezone = transactionTimezone
		}

		transactionTime, err := utils.ParseFromLongDateTimeInTimeZone(dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME), timezone)

		if err != nil {
			log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse time \"%s\" in data row \"index:%d\" for user \"uid:%d\", because %s", dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME), dataRowIndex, user.Uid, err.Error())
			return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrTransactionTimeInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME))
		}

		transactionDbType, err := c.getTransactionDbType(nameDbTypeMap, dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE))

		if err != nil {
			log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse transaction type \"%s\" in data row \"index:%d\" for user \"uid:%d\", because %s", dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE), dataRowIndex, user.Uid, err.Error())
			return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.Or(err, errs.ErrTransactionTypeInvalid), dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE))
		}

		categoryId := int64(0)
		categoryName := ""
		subCategoryName := ""

		if transactionDbType != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
			transactionCategoryType, err := c.getTransactionCategoryType(transactionDbType)

			if err != nil {
				log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse transaction category type in data row \"index:%d\" for user \"uid:%d\", because %s", dataRowIndex, user.Uid, err.Error())
				return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.Or(err, errs.ErrTransactionTypeInvalid), dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE))
			}

			categoryName = dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_CATEGORY)
			subCategoryName = dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY)

			if transactionDbType == models.TRANSACTION_DB_TYPE_EXPENSE {
				subCategory, exists := c.getTransactionCategory(expenseCategoryMap, categoryName, subCategoryName)

				if !exists {
					subCategory = c.createNewTransactionCategoryModel(user.Uid, subCategoryName, transactionCategoryType)
					allNewSubExpenseCategories = append(allNewSubExpenseCategories, subCategory)

					if _, exists = expenseCategoryMap[subCategoryName]; !exists {
						expenseCategoryMap[subCategoryName] = make(map[string]*models.TransactionCategory)
					}

					expenseCategoryMap[subCategoryName][categoryName] = subCategory
				}

				categoryId = subCategory.CategoryId
			} else if transactionDbType == models.TRANSACTION_DB_TYPE_INCOME {
				subCategory, exists := c.getTransactionCategory(incomeCategoryMap, categoryName, subCategoryName)

				if !exists {
					subCategory = c.createNewTransactionCategoryModel(user.Uid, subCategoryName, transactionCategoryType)
					allNewSubIncomeCategories = append(allNewSubIncomeCategories, subCategory)

					if _, exists = incomeCategoryMap[subCategoryName]; !exists {
						incomeCategoryMap[subCategoryName] = make(map[string]*models.TransactionCategory)
					}

					incomeCategoryMap[subCategoryName][categoryName] = subCategory
				}

				categoryId = subCategory.CategoryId
			} else if transactionDbType == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
				subCategory, exists := c.getTransactionCategory(transferCategoryMap, categoryName, subCategoryName)

				if !exists {
					subCategory = c.createNewTransactionCategoryModel(user.Uid, subCategoryName, transactionCategoryType)
					allNewSubTransferCategories = append(allNewSubTransferCategories, subCategory)

					if _, exists = transferCategoryMap[subCategoryName]; !exists {
						transferCategoryMap[subCategoryName] = make(map[string]*models.TransactionCategory)
					}

					transferCategoryMap[subCategoryName][categoryName] = subCategory
				}

				categoryId = subCategory.CategoryId
			}
		}

		accountName := dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME)
		accountCurrency := user.DefaultCurrency

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY) && dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY) != "" {
			accountCurrency = dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY)

			if _, ok := validators.AllCurrencyNames[accountCurrency]; !ok {
				log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] account currency \"%s\" is not supported in data row \"index:%d\" for user \"uid:%d\"", accountCurrency, dataRowIndex, user.Uid)
				return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrAccountCurrencyInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY))
			}
		}

		account, exists := accountMap[accountName]

		if !exists {
			account = c.createNewAccountModel(user.Uid, accountName, accountCurrency)
			allNewAccounts = append(allNewAccounts, account)
			accountMap[accountName] = account
		}

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY) && dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY) != "" {
			if account.Name != "" && account.Currency != accountCurrency {
				log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] currency \"%s\" in data row \"index:%d\" not equals currency \"%s\" of the account for user \"uid:%d\"", accountCurrency, dataRowIndex, account.Currency, user.Uid)
				return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrAccountCurrencyInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY))
			}
		} else if exists {
			accountCurrency = account.Currency
		}

		amount, err := utils.ParseAmount(dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_AMOUNT))

		if err != nil {
			log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse acmount \"%s\" in data row \"index:%d\" for user \"uid:%d\", because %s", dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_AMOUNT), dataRowIndex, user.Uid, err.Error())
			return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrAmountInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_AMOUNT))
		}

		relatedAccountId := int64(0)
		relatedAccountAmount := int64(0)
		account2Name := ""
		account2Currency := ""

		if transactionDbType == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			account2Name = dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME)
			account2Currency = user.DefaultCurrency

			if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY) && dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY) != "" {
				account2Currency = dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY)

				if _, ok := validators.AllCurrencyNames[account2Currency]; !ok {
					log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] account2 currency \"%s\" is not supported in data row \"index:%d\" for user \"uid:%d\"", account2Currency, dataRowIndex, user.Uid)
					return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrAccountCurrencyInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY))
				}
			}

			account2, exists := accountMap[account2Name]

			if !exists {
				account2 = c.createNewAccountModel(user.Uid, account2Name, account2Currency)
				allNewAccounts = append(allNewAccounts, account2)
				accountMap[account2Name] = account2
			}

			if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY) && dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY) != "" {
				if account2.Name != "" && account2.Currency != account2Currency {
					log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] currency \"%s\" in data row \"index:%d\" not equals currency \"%s\" of the account2 for user \"uid:%d\"", account2Currency, dataRowIndex, account2.Currency, user.Uid)
					return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrAccountCurrencyInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY))
				}
			} else if exists {
				account2Currency = account2.Currency
			}

			relatedAccountId = account2.AccountId

			if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT) {
				relatedAccountAmount, err = utils.ParseAmount(dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT))

				if err != nil {
					log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse acmount2 \"%s\" in data row \"index:%d\" for user \"uid:%d\", because %s", dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT), dataRowIndex, user.Uid, err.Error())
					return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrAmountInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT))
				}
			} else if transactionDbType == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
				relatedAccountAmount = amount
			}
		}

		geoLongitude := float64(0)
		geoLatitude := float64(0)

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_GEOGRAPHIC_LOCATION) && c.geoLocationSeparator != "" {
			geoLocationItems := strings.Split(dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_GEOGRAPHIC_LOCATION), c.geoLocationSeparator)

			if len(geoLocationItems) == 2 {
				geoLocationFirstItem, err := utils.StringToFloat64(geoLocationItems[0])

				if err != nil {
					log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse geographic location \"%s\" in data row \"index:%d\" for user \"uid:%d\", because %s", dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_GEOGRAPHIC_LOCATION), dataRowIndex, user.Uid, err.Error())
					return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrGeographicLocationInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_GEOGRAPHIC_LOCATION))
				}

				geoLocationSecondItem, err := utils.StringToFloat64(geoLocationItems[1])

				if err != nil {
					log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] cannot parse geographic location \"%s\" in data row \"index:%d\" for user \"uid:%d\", because %s", dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_GEOGRAPHIC_LOCATION), dataRowIndex, user.Uid, err.Error())
					return nil, nil, nil, nil, nil, nil, errs.NewImportError(errs.ErrGeographicLocationInvalid, dataRowIterator.CurrentRowId(), dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_GEOGRAPHIC_LOCATION))
				}

				if c.geoLocationOrder == TRANSACTION_GEO_LOCATION_ORDER_LONGITUDE_LATITUDE {
					geoLongitude = geoLocationFirstItem
					geoLatitude = geoLocationSecondItem
				} else if c.geoLocationOrder == TRANSACTION_GEO_LOCATION_ORDER_LATITUDE_LONGITUDE {
					geoLatitude = geoLocationFirstItem
					geoLongitude = geoLocationSecondItem
				}
			}
		}

		var tagIds []string
		var tagNames []string
		tagNamesMap := make(map[string]bool)

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_TAGS) {
			var tagNameItems []string

			if c.transactionTagSeparator != "" {
				tagNameItems = strings.Split(dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TAGS), c.transactionTagSeparator)
			} else {
				tagNameItems = append(tagNameItems, dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_TAGS))
			}

			for i := 0; i < len(tagNameItems); i++ {
				tagName := tagNameItems[i]

				if tagName == "" {
					continue
				}

				allNewTags, tagIds, tagNames = c.addTag(user, tagName, tagNamesMap, tagMap, allNewTags, tagIds, tagNames)
			}
		}

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_PAYEE) && additionalOptions.IsPayeeAsTag() {
			payee := dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_PAYEE)

			if payee != "" {
				allNewTags, tagIds, tagNames = c.addTag(user, payee, tagNamesMap, tagMap, allNewTags, tagIds, tagNames)
			}
		}

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_MEMBER) && additionalOptions.IsMemberAsTag() {
			member := dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_MEMBER)

			if member != "" {
				allNewTags, tagIds, tagNames = c.addTag(user, member, tagNamesMap, tagMap, allNewTags, tagIds, tagNames)
			}
		}

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_PROJECT) && additionalOptions.IsProjectAsTag() {
			project := dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_PROJECT)

			if project != "" {
				allNewTags, tagIds, tagNames = c.addTag(user, project, tagNamesMap, tagMap, allNewTags, tagIds, tagNames)
			}
		}

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_MERCHANT) && additionalOptions.IsMerchantAsTag() {
			merchant := dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_MERCHANT)

			if merchant != "" {
				allNewTags, tagIds, tagNames = c.addTag(user, merchant, tagNamesMap, tagMap, allNewTags, tagIds, tagNames)
			}
		}

		description := ""

		if dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_DESCRIPTION) {
			description = dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_DESCRIPTION)
		}

		if description == "" && additionalOptions.IsPayeeAsDescription() && dataTable.HasColumn(datatable.TRANSACTION_DATA_TABLE_PAYEE) {
			description = dataRow.GetData(datatable.TRANSACTION_DATA_TABLE_PAYEE)
		}

		transaction := &models.ImportTransaction{
			Transaction: &models.Transaction{
				Uid:                  user.Uid,
				Type:                 transactionDbType,
				CategoryId:           categoryId,
				TransactionTime:      utils.GetMinTransactionTimeFromUnixTime(transactionTime.Unix()),
				TimezoneUtcOffset:    utils.GetTimezoneOffsetMinutes(transactionTime.Unix(), timezone),
				AccountId:            account.AccountId,
				Amount:               amount,
				HideAmount:           false,
				RelatedAccountId:     relatedAccountId,
				RelatedAccountAmount: relatedAccountAmount,
				Comment:              description,
				GeoLongitude:         geoLongitude,
				GeoLatitude:          geoLatitude,
				CreatedIp:            ctx.ClientIP(),
			},
			TagIds:                             tagIds,
			OriginalCategoryName:               subCategoryName,
			OriginalSourceAccountName:          accountName,
			OriginalSourceAccountCurrency:      accountCurrency,
			OriginalDestinationAccountName:     account2Name,
			OriginalDestinationAccountCurrency: account2Currency,
			OriginalTagNames:                   tagNames,
		}

		allNewTransactions = append(allNewTransactions, transaction)
	}

	if len(allNewTransactions) < 1 {
		log.Errorf(ctx, "[data_table_transaction_data_importer.ParseImportedData] no transaction data parsed for \"uid:%d\"", user.Uid)
		return nil, nil, nil, nil, nil, nil, errs.ErrNotFoundTransactionDataInFile
	}

	sort.Sort(allNewTransactions)

	return allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags, nil
}
