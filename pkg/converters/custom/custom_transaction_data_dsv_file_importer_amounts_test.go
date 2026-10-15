package custom

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCustomTransactionDataDsvFileImporter_ParseAmountWithCustomFormat(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_tsv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ",", ".", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56\tE\t1.234,56"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(123456), allNewTransactions[0].Amount)
}

func TestCustomTransactionDataDsvFileImporter_ParseInvalidAmountWithCustomFormat(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_tsv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", ",", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56\tE\t1.234,56"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseInvalidAmountWithCustomFormat2(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_tsv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ",", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56\tE\t1.234,56"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParsePrimaryCategory(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_CATEGORY:         2,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"B": models.TRANSACTION_TYPE_MODIFY_BALANCE,
		"I": models.TRANSACTION_TYPE_INCOME,
		"E": models.TRANSACTION_TYPE_EXPENSE,
		"T": models.TRANSACTION_TYPE_TRANSFER,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 00:00:00,B,,123.45\n"+
			"2024-09-01 01:23:45,I,Test Category,0.12\n"+
			"2024-09-01 12:34:56,E,Test Category2,1.00\n"+
			"2024-09-01 23:59:59,T,Test Category3,0.05"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 4, len(allNewTransactions))
	assert.Equal(t, 1, len(allNewSubExpenseCategories))
	assert.Equal(t, 1, len(allNewSubIncomeCategories))
	assert.Equal(t, 1, len(allNewSubTransferCategories))

	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, "", allNewTransactions[0].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[1].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[1].Type)
	assert.Equal(t, int64(1725153825), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, "Test Category", allNewTransactions[1].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[2].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[2].Type)
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
	assert.Equal(t, "Test Category2", allNewTransactions[2].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[3].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, allNewTransactions[3].Type)
	assert.Equal(t, int64(1725235199), utils.GetUnixTimeFromTransactionTime(allNewTransactions[3].TransactionTime))
	assert.Equal(t, int64(5), allNewTransactions[3].Amount)
	assert.Equal(t, "Test Category3", allNewTransactions[3].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewSubExpenseCategories[0].Uid)
	assert.Equal(t, "Test Category2", allNewSubExpenseCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubIncomeCategories[0].Uid)
	assert.Equal(t, "Test Category", allNewSubIncomeCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubTransferCategories[0].Uid)
	assert.Equal(t, "Test Category3", allNewSubTransferCategories[0].Name)
}

func TestCustomTransactionDataDsvFileImporter_ParseValidAccountCurrency(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:         0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:         1,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME:             2,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY:         3,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:                   4,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME:     5,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY: 6,
		datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT:           7,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"B": models.TRANSACTION_TYPE_MODIFY_BALANCE,
		"T": models.TRANSACTION_TYPE_TRANSFER,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 01:23:45,B,Test Account,USD,123.45,,,\n"+
			"2024-09-01 12:34:56,T,Test Account,USD,1.23,Test Account2,EUR,1.10"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 2, len(allNewTransactions))
	assert.Equal(t, 2, len(allNewAccounts))

	assert.Equal(t, int64(1234567890), allNewAccounts[0].Uid)
	assert.Equal(t, "Test Account", allNewAccounts[0].Name)
	assert.Equal(t, "USD", allNewAccounts[0].Currency)

	assert.Equal(t, int64(1234567890), allNewAccounts[1].Uid)
	assert.Equal(t, "Test Account2", allNewAccounts[1].Name)
	assert.Equal(t, "EUR", allNewAccounts[1].Currency)
}

func TestCustomTransactionDataDsvFileImporter_ParseInvalidAccountCurrency(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:         0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:         1,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME:             2,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY:         3,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:                   4,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME:     5,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY: 6,
		datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT:           7,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"B": models.TRANSACTION_TYPE_MODIFY_BALANCE,
		"T": models.TRANSACTION_TYPE_TRANSFER,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 01:23:45,B,Test Account,USD,123.45,,,\n"+
			"2024-09-01 12:34:56,T,Test Account,CNY,1.23,Test Account2,EUR,1.10"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 01:23:45,B,Test Account,USD,123.45,,,\n"+
			"2024-09-01 12:34:56,T,Test Account2,CNY,1.23,Test Account,EUR,1.10"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseNotSupportedCurrency(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:         0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:         1,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME:             2,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY:         3,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:                   4,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME:     5,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY: 6,
		datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT:           7,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"B": models.TRANSACTION_TYPE_MODIFY_BALANCE,
		"T": models.TRANSACTION_TYPE_TRANSFER,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 01:23:45,B,Test Account,XXX,123.45,,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 01:23:45,T,Test Account,USD,123.45,Test Account2,XXX,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseValidAmount(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
		datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT:   3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"B": models.TRANSACTION_TYPE_MODIFY_BALANCE,
		"I": models.TRANSACTION_TYPE_INCOME,
		"E": models.TRANSACTION_TYPE_EXPENSE,
		"T": models.TRANSACTION_TYPE_TRANSFER,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 00:00:00,B,123.45000000,\n"+
			"2024-09-01 01:23:45,I,0.12000000,\n"+
			"2024-09-01 12:34:56,E,1.00000000,\n"+
			"2024-09-01 23:59:59,T,0.05000000,0.35000000"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 4, len(allNewTransactions))

	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)

	assert.Equal(t, int64(1234567890), allNewTransactions[1].Uid)
	assert.Equal(t, int64(1725153825), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(12), allNewTransactions[1].Amount)

	assert.Equal(t, int64(1234567890), allNewTransactions[2].Uid)
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
	assert.Equal(t, int64(100), allNewTransactions[2].Amount)

	assert.Equal(t, int64(1234567890), allNewTransactions[3].Uid)
	assert.Equal(t, int64(1725235199), utils.GetUnixTimeFromTransactionTime(allNewTransactions[3].TransactionTime))
	assert.Equal(t, int64(5), allNewTransactions[3].Amount)
	assert.Equal(t, int64(35), allNewTransactions[3].RelatedAccountAmount)
}

func TestCustomTransactionDataDsvFileImporter_ParseAmountWithSpaceDigitGroupingSymbol(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", " ", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	// normal space
	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 00:00:00,E,1 234,\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(123400), allNewTransactions[0].Amount)

	// no-break space (NBSP)
	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 00:00:00,E,1 234,\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(123400), allNewTransactions[0].Amount)

	// narrow no-break space (NNBSP)
	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 00:00:00,E,1 234,\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(123400), allNewTransactions[0].Amount)

	// figure space
	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 00:00:00,E,1 234,\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(123400), allNewTransactions[0].Amount)
}

func TestCustomTransactionDataDsvFileImporter_ParseInvalidAmount(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:     0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:     1,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME:         2,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:               3,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME: 4,
		datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT:       5,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
		"T": models.TRANSACTION_TYPE_TRANSFER,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,E,Test Account,123 45,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,T,Test Account,123.45,Test Account2,123 45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseNoAmount2(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:     0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:     1,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME:         2,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:               3,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME: 4,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
		"T": models.TRANSACTION_TYPE_TRANSFER,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,E,Test Account,123.45,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
	assert.Equal(t, int64(0), allNewTransactions[0].RelatedAccountAmount)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,T,Test Account,123.45,Test Account2"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
	assert.Equal(t, int64(12345), allNewTransactions[0].RelatedAccountAmount)
}
