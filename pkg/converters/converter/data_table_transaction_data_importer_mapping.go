package converter

import (
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

func (c *DataTableTransactionDataImporter) buildTransactionTypeNameDbTypeMap() (map[string]models.TransactionDbType, error) {
	if c.transactionTypeMapping == nil {
		return nil, errs.ErrTransactionTypeInvalid
	}

	nameDbTypeMap := make(map[string]models.TransactionDbType, len(c.transactionTypeMapping))

	for name, transactionType := range c.transactionTypeMapping {
		if transactionType == models.TRANSACTION_TYPE_MODIFY_BALANCE {
			nameDbTypeMap[name] = models.TRANSACTION_DB_TYPE_MODIFY_BALANCE
		} else if transactionType == models.TRANSACTION_TYPE_INCOME {
			nameDbTypeMap[name] = models.TRANSACTION_DB_TYPE_INCOME
		} else if transactionType == models.TRANSACTION_TYPE_EXPENSE {
			nameDbTypeMap[name] = models.TRANSACTION_DB_TYPE_EXPENSE
		} else if transactionType == models.TRANSACTION_TYPE_TRANSFER {
			nameDbTypeMap[name] = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
		} else {
			return nil, errs.ErrTransactionTypeInvalid
		}
	}

	return nameDbTypeMap, nil
}

func (c *DataTableTransactionDataImporter) getTransactionDbType(nameDbTypeMap map[string]models.TransactionDbType, transactionTypeName string) (models.TransactionDbType, error) {
	transactionType, exists := nameDbTypeMap[transactionTypeName]

	if !exists {
		return 0, errs.ErrTransactionTypeInvalid
	}

	return transactionType, nil
}

func (c *DataTableTransactionDataImporter) getTransactionCategoryType(transactionType models.TransactionDbType) (models.TransactionCategoryType, error) {
	if transactionType == models.TRANSACTION_DB_TYPE_INCOME {
		return models.CATEGORY_TYPE_INCOME, nil
	} else if transactionType == models.TRANSACTION_DB_TYPE_EXPENSE {
		return models.CATEGORY_TYPE_EXPENSE, nil
	} else if transactionType == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		return models.CATEGORY_TYPE_TRANSFER, nil
	} else {
		return 0, errs.ErrTransactionTypeInvalid
	}
}

func (c *DataTableTransactionDataImporter) getTransactionCategory(categories map[string]map[string]*models.TransactionCategory, categoryName string, subCategoryName string) (*models.TransactionCategory, bool) {
	if len(categories) < 1 {
		return nil, false
	}

	subCategories, exists := categories[subCategoryName]

	if !exists || len(subCategories) < 1 {
		return nil, false
	}

	if categoryName == "" {
		for _, subCategory := range subCategories {
			if subCategory != nil {
				return subCategory, true
			}
		}
	}

	subCategory, exists := subCategories[categoryName]

	if !exists {
		for _, subCategory := range subCategories {
			if subCategory != nil {
				return subCategory, true
			}
		}
	}

	return subCategory, exists
}

func (c *DataTableTransactionDataImporter) addTag(user *models.User, tagName string, tagNamesMap map[string]bool, tagMap map[string]*models.TransactionTag, allNewTags []*models.TransactionTag, tagIds []string, tagNames []string) ([]*models.TransactionTag, []string, []string) {
	if tagName != "" && !tagNamesMap[tagName] {
		tag, exists := tagMap[tagName]

		if !exists {
			tag = c.createNewTransactionTagModel(user.Uid, tagName)
			allNewTags = append(allNewTags, tag)
			tagMap[tagName] = tag
		}

		if tag != nil {
			tagIds = append(tagIds, utils.Int64ToString(tag.TagId))
		}

		tagNames = append(tagNames, tagName)
		tagNamesMap[tagName] = true
	}

	return allNewTags, tagIds, tagNames
}

func (c *DataTableTransactionDataImporter) createNewAccountModel(uid int64, accountName string, currency string) *models.Account {
	return &models.Account{
		Uid:      uid,
		Name:     accountName,
		Currency: currency,
	}
}

func (c *DataTableTransactionDataImporter) createNewTransactionCategoryModel(uid int64, categoryName string, transactionCategoryType models.TransactionCategoryType) *models.TransactionCategory {
	return &models.TransactionCategory{
		Uid:  uid,
		Name: categoryName,
		Type: transactionCategoryType,
	}
}

func (c *DataTableTransactionDataImporter) createNewTransactionTagModel(uid int64, tagName string) *models.TransactionTag {
	return &models.TransactionTag{
		Uid:  uid,
		Name: tagName,
	}
}

// CreateNewImporterWithTypeNameMapping returns a new data table transaction data importer according to the specified arguments
func CreateNewImporterWithTypeNameMapping(transactionTypeMapping map[models.TransactionType]string, geoLocationSeparator string, geoLocationOrder TransactionGeoLocationOrder, transactionTagSeparator string) *DataTableTransactionDataImporter {
	return &DataTableTransactionDataImporter{
		transactionTypeMapping:  buildTransactionNameTypeMap(transactionTypeMapping),
		geoLocationSeparator:    geoLocationSeparator,
		geoLocationOrder:        geoLocationOrder,
		transactionTagSeparator: transactionTagSeparator,
	}
}

// CreateNewSimpleImporter returns a new data table transaction data importer according to the specified arguments
func CreateNewSimpleImporter(transactionTypeMapping map[string]models.TransactionType) *DataTableTransactionDataImporter {
	return &DataTableTransactionDataImporter{
		transactionTypeMapping: transactionTypeMapping,
	}
}

// CreateNewSimpleImporterWithTypeNameMapping returns a new data table transaction data importer according to the specified arguments
func CreateNewSimpleImporterWithTypeNameMapping(transactionTypeMapping map[models.TransactionType]string) *DataTableTransactionDataImporter {
	return &DataTableTransactionDataImporter{
		transactionTypeMapping: buildTransactionNameTypeMap(transactionTypeMapping),
	}
}

func buildTransactionNameTypeMap(transactionTypeMapping map[models.TransactionType]string) map[string]models.TransactionType {
	if transactionTypeMapping == nil {
		return nil
	}

	typeNameMap := make(map[string]models.TransactionType, len(transactionTypeMapping))

	for transactionType, name := range transactionTypeMapping {
		typeNameMap[name] = transactionType
	}

	return typeNameMap
}
