package custom

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIsDelimiterSeparatedValuesFileType(t *testing.T) {
	assert.True(t, IsDelimiterSeparatedValuesFileType("custom_csv"))
	assert.True(t, IsDelimiterSeparatedValuesFileType("custom_tsv"))
	assert.True(t, IsDelimiterSeparatedValuesFileType("custom_ssv"))

	assert.False(t, IsDelimiterSeparatedValuesFileType("dsv"))
	assert.False(t, IsDelimiterSeparatedValuesFileType("csv"))
	assert.False(t, IsDelimiterSeparatedValuesFileType("tsv"))
	assert.False(t, IsDelimiterSeparatedValuesFileType("ssv"))
}

func TestCustomTransactionDataDsvFileParser_ParseDataLines(t *testing.T) {
	importer, err := CreateNewCustomTransactionDataDsvFileParser("custom_csv", "utf-8")
	assert.Nil(t, err)

	context := core.NewNullContext()

	allLines, err := importer.ParseDataLines(context, []byte(
		"2024-09-01 00:00:00,B,123.45\n"+
			"2024-09-01 01:23:45,I,0.12\n"))
	assert.Nil(t, err)

	assert.Equal(t, 2, len(allLines))

	assert.Equal(t, 3, len(allLines[0]))
	assert.Equal(t, "2024-09-01 00:00:00", allLines[0][0])
	assert.Equal(t, "B", allLines[0][1])
	assert.Equal(t, "123.45", allLines[0][2])

	assert.Equal(t, 3, len(allLines[1]))
	assert.Equal(t, "2024-09-01 01:23:45", allLines[1][0])
	assert.Equal(t, "I", allLines[1][1])
	assert.Equal(t, "0.12", allLines[1][2])

	importer, err = CreateNewCustomTransactionDataDsvFileParser("custom_tsv", "utf-8")
	assert.Nil(t, err)

	allLines, err = importer.ParseDataLines(context, []byte(
		"2024-09-01 12:34:56\tE\t1.00\n"+
			"2024-09-01 23:59:59\tT\t0.05"))
	assert.Nil(t, err)

	assert.Equal(t, 2, len(allLines))

	assert.Equal(t, 3, len(allLines[0]))
	assert.Equal(t, "2024-09-01 12:34:56", allLines[0][0])
	assert.Equal(t, "E", allLines[0][1])
	assert.Equal(t, "1.00", allLines[0][2])

	assert.Equal(t, 3, len(allLines[1]))
	assert.Equal(t, "2024-09-01 23:59:59", allLines[1][0])
	assert.Equal(t, "T", allLines[1][1])
	assert.Equal(t, "0.05", allLines[1][2])

	importer, err = CreateNewCustomTransactionDataDsvFileParser("custom_ssv", "utf-8")
	assert.Nil(t, err)

	allLines, err = importer.ParseDataLines(context, []byte(
		"2024-09-01 12:34:56;E;1.00\n"+
			"2024-09-01 23:59:59;T;0.05"))
	assert.Nil(t, err)

	assert.Equal(t, 2, len(allLines))

	assert.Equal(t, 3, len(allLines[0]))
	assert.Equal(t, "2024-09-01 12:34:56", allLines[0][0])
	assert.Equal(t, "E", allLines[0][1])
	assert.Equal(t, "1.00", allLines[0][2])

	assert.Equal(t, 3, len(allLines[1]))
	assert.Equal(t, "2024-09-01 23:59:59", allLines[1][0])
	assert.Equal(t, "T", allLines[1][1])
	assert.Equal(t, "0.05", allLines[1][2])
}

func TestCustomTransactionDataDsvFileImporter_MinimumValidData(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"B": models.TRANSACTION_TYPE_MODIFY_BALANCE,
		"I": models.TRANSACTION_TYPE_INCOME,
		"E": models.TRANSACTION_TYPE_EXPENSE,
		"T": models.TRANSACTION_TYPE_TRANSFER,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", ".", "", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 00:00:00,B,123.45\n"+
			"2024-09-01 01:23:45,I,0.12\n"+
			"2024-09-01 12:34:56,E,1.00\n"+
			"2024-09-01 23:59:59,T,0.05"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 4, len(allNewTransactions))
	assert.Equal(t, 1, len(allNewAccounts))
	assert.Equal(t, 1, len(allNewSubExpenseCategories))
	assert.Equal(t, 1, len(allNewSubIncomeCategories))
	assert.Equal(t, 1, len(allNewSubTransferCategories))
	assert.Equal(t, 0, len(allNewTags))

	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
	assert.Equal(t, "", allNewTransactions[0].OriginalSourceAccountName)
	assert.Equal(t, "", allNewTransactions[0].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[1].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[1].Type)
	assert.Equal(t, int64(1725153825), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(12), allNewTransactions[1].Amount)
	assert.Equal(t, "", allNewTransactions[1].OriginalSourceAccountName)
	assert.Equal(t, "", allNewTransactions[1].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[2].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[2].Type)
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
	assert.Equal(t, int64(100), allNewTransactions[2].Amount)
	assert.Equal(t, "", allNewTransactions[2].OriginalSourceAccountName)
	assert.Equal(t, "", allNewTransactions[2].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[3].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, allNewTransactions[3].Type)
	assert.Equal(t, int64(1725235199), utils.GetUnixTimeFromTransactionTime(allNewTransactions[3].TransactionTime))
	assert.Equal(t, int64(5), allNewTransactions[3].Amount)
	assert.Equal(t, "", allNewTransactions[3].OriginalSourceAccountName)
	assert.Equal(t, "", allNewTransactions[3].OriginalDestinationAccountName)
	assert.Equal(t, "", allNewTransactions[3].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewAccounts[0].Uid)
	assert.Equal(t, "", allNewAccounts[0].Name)
	assert.Equal(t, "CNY", allNewAccounts[0].Currency)

	assert.Equal(t, int64(1234567890), allNewSubExpenseCategories[0].Uid)
	assert.Equal(t, "", allNewSubExpenseCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubIncomeCategories[0].Uid)
	assert.Equal(t, "", allNewSubIncomeCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubTransferCategories[0].Uid)
	assert.Equal(t, "", allNewSubTransferCategories[0].Name)
}

func TestCustomTransactionDataDsvFileImporter_WithAllSupportedColumns(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:         0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE:     1,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:         2,
		datatable.TRANSACTION_DATA_TABLE_CATEGORY:                 3,
		datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY:             4,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME:             5,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY:         6,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:                   7,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME:     8,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY: 9,
		datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT:           10,
		datatable.TRANSACTION_DATA_TABLE_GEOGRAPHIC_LOCATION:      11,
		datatable.TRANSACTION_DATA_TABLE_TAGS:                     12,
		datatable.TRANSACTION_DATA_TABLE_DESCRIPTION:              13,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"Balance Modification": models.TRANSACTION_TYPE_MODIFY_BALANCE,
		"Income":               models.TRANSACTION_TYPE_INCOME,
		"Expense":              models.TRANSACTION_TYPE_EXPENSE,
		"Transfer":             models.TRANSACTION_TYPE_TRANSFER,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, true, "YYYY-MM-DD HH:mm:ss", "", ".", "", " ", "lonlat", ";")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags, err := importer.ParseImportedData(context, user, []byte(
		"\"Time\",\"Timezone\",\"Type\",\"Category\",\"Sub Category\",\"Account\",\"Account Currency\",\"Amount\",\"Account2\",\"Account2 Currency\",\"Account2 Amount\",\"Geographic Location\",\"Tags\",\"Description\"\n"+
			"\"2024-09-01 00:00:00\",\"+08:00\",\"Balance Modification\",\"\",\"\",\"Test Account\",\"CNY\",\"123.45\",\"\",\"\",\"\",\"\",\"\",\"\"\n"+
			"\"2024-09-01 01:23:45\",\"+08:00\",\"Income\",\"Test Category\",\"Test Sub Category\",\"Test Account\",\"CNY\",\"0.12\",\"\",\"\",\"\",\"123.450000 45.670000\",\"Test Tag;Test Tag2\",\"Hello World\"\n"+
			"\"2024-09-01 12:34:56\",\"+00:00\",\"Expense\",\"Test Category2\",\"Test Sub Category2\",\"Test Account\",\"CNY\",\"1.00\",\"\",\"\",\"\",\"\",\"Test Tag\",\"Foo#Bar\"\n"+
			"\"2024-09-01 23:59:59\",\"-05:00\",\"Transfer\",\"Test Category3\",\"Test Sub Category3\",\"Test Account\",\"CNY\",\"0.05\",\"Test Account2\",\"USD\",\"0.35\",\"\",\"Test Tag2\",\"foo\tbar\""), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 4, len(allNewTransactions))
	assert.Equal(t, 2, len(allNewAccounts))
	assert.Equal(t, 1, len(allNewSubExpenseCategories))
	assert.Equal(t, 1, len(allNewSubIncomeCategories))
	assert.Equal(t, 1, len(allNewSubTransferCategories))
	assert.Equal(t, 2, len(allNewTags))

	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725120000), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[0].OriginalSourceAccountName)
	assert.Equal(t, "", allNewTransactions[0].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[1].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[1].Type)
	assert.Equal(t, int64(1725125025), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(12), allNewTransactions[1].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[1].OriginalSourceAccountName)
	assert.Equal(t, "Test Sub Category", allNewTransactions[1].OriginalCategoryName)
	assert.Equal(t, 123.45, allNewTransactions[1].GeoLongitude)
	assert.Equal(t, 45.67, allNewTransactions[1].GeoLatitude)
	assert.Equal(t, "Hello World", allNewTransactions[1].Comment)

	assert.Equal(t, int64(1234567890), allNewTransactions[2].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[2].Type)
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
	assert.Equal(t, int64(100), allNewTransactions[2].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[2].OriginalSourceAccountName)
	assert.Equal(t, "Test Sub Category2", allNewTransactions[2].OriginalCategoryName)
	assert.Equal(t, "Foo#Bar", allNewTransactions[2].Comment)

	assert.Equal(t, int64(1234567890), allNewTransactions[3].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, allNewTransactions[3].Type)
	assert.Equal(t, int64(1725253199), utils.GetUnixTimeFromTransactionTime(allNewTransactions[3].TransactionTime))
	assert.Equal(t, int64(5), allNewTransactions[3].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[3].OriginalSourceAccountName)
	assert.Equal(t, "Test Account2", allNewTransactions[3].OriginalDestinationAccountName)
	assert.Equal(t, "Test Sub Category3", allNewTransactions[3].OriginalCategoryName)
	assert.Equal(t, "foo\tbar", allNewTransactions[3].Comment)

	assert.Equal(t, int64(1234567890), allNewAccounts[0].Uid)
	assert.Equal(t, "Test Account", allNewAccounts[0].Name)
	assert.Equal(t, "CNY", allNewAccounts[0].Currency)

	assert.Equal(t, int64(1234567890), allNewAccounts[1].Uid)
	assert.Equal(t, "Test Account2", allNewAccounts[1].Name)
	assert.Equal(t, "USD", allNewAccounts[1].Currency)

	assert.Equal(t, int64(1234567890), allNewSubExpenseCategories[0].Uid)
	assert.Equal(t, "Test Sub Category2", allNewSubExpenseCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubIncomeCategories[0].Uid)
	assert.Equal(t, "Test Sub Category", allNewSubIncomeCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubTransferCategories[0].Uid)
	assert.Equal(t, "Test Sub Category3", allNewSubTransferCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewTags[0].Uid)
	assert.Equal(t, "Test Tag", allNewTags[0].Name)

	assert.Equal(t, int64(1234567890), allNewTags[1].Uid)
	assert.Equal(t, "Test Tag2", allNewTags[1].Name)
}
