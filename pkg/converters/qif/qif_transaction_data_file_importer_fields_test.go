package qif

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQIFTransactionDataFileParseImportedData_ParseAmountWithThousandsSeparator(t *testing.T) {
	importer := QifYearMonthDayTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"D2024-09-01\n"+
			"T-123,456.78\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 1, len(allNewTransactions))

	assert.Equal(t, int64(12345678), allNewTransactions[0].Amount)
}

func TestQIFTransactionDataFileParseImportedData_ParseInvalidAmount(t *testing.T) {
	importer := QifYearMonthDayTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"D2024-09-01\n"+
			"T-123 45\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}

func TestQIFTransactionDataFileParseImportedData_ParseAccountType(t *testing.T) {
	importer := QifYearMonthDayTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!Type:Cash\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!Type:CCard\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!Type:Oth A\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!Type:Oth L\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
}

func TestQIFTransactionDataFileParseImportedData_ParseAccount(t *testing.T) {
	importer := QifYearMonthDayTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!Account\n"+
			"NTest Account\n"+
			"^\n"+
			"!Type:Bank\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, 1, len(allNewAccounts))

	assert.Equal(t, "Test Account", allNewTransactions[0].OriginalSourceAccountName)

	assert.Equal(t, int64(1234567890), allNewAccounts[0].Uid)
	assert.Equal(t, "Test Account", allNewAccounts[0].Name)
	assert.Equal(t, "CNY", allNewAccounts[0].Currency)
}

func TestQIFTransactionDataFileParseImportedData_ParseAmountWithLeadingPlusSign(t *testing.T) {
	importer := QifYearMonthDayTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"D2024-09-01\n"+
			"T+123.45\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
}

func TestQIFTransactionDataFileParseImportedData_ParseSubCategory(t *testing.T) {
	importer := QifYearMonthDayTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, allNewSubExpenseCategories, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"LTest Category:Sub Category\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, 1, len(allNewSubExpenseCategories))

	assert.Equal(t, "Sub Category", allNewTransactions[0].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewSubExpenseCategories[0].Uid)
	assert.Equal(t, "Sub Category", allNewSubExpenseCategories[0].Name)
}

func TestQIFTransactionDataFileParseImportedData_ParseDescription(t *testing.T) {
	importer := QifYearMonthDayTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"PTest\n"+
			"Mfoo    bar\t#test\n"+
			"^\n"+
			"D2024-09-02\n"+
			"T-234.56\n"+
			"PTest2\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(allNewTransactions))
	assert.Equal(t, "foo    bar\t#test", allNewTransactions[0].Comment)
	assert.Equal(t, "", allNewTransactions[1].Comment)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"PTest\n"+
			"Mfoo    bar\t#test\n"+
			"^\n"+
			"D2024-09-02\n"+
			"T-234.56\n"+
			"PTest2\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions.WithPayeeAsDescription(), nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(allNewTransactions))
	assert.Equal(t, "foo    bar\t#test", allNewTransactions[0].Comment)
	assert.Equal(t, "Test2", allNewTransactions[1].Comment)
}

func TestQIFTransactionDataFileParseImportedData_WithAdditionalOptions(t *testing.T) {
	importer := QifYearMonthDayTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"PTest2\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, 0, len(allNewTransactions[0].OriginalTagNames))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"D2024-09-01\n"+
			"T-123.45\n"+
			"PTest2\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions.WithPayeeAsTag(), nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, 1, len(allNewTransactions[0].OriginalTagNames))
	assert.Equal(t, "Test2", allNewTransactions[0].OriginalTagNames[0])
}

func TestQIFTransactionDataFileParseImportedData_MissingRequiredFields(t *testing.T) {
	importer := QifYearMonthDayTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing Time Field
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"T-123.45\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingTransactionTime.Message)

	// Missing Amount Field
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!Type:Bank\n"+
			"D2024-09-01\n"+
			"^\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}
