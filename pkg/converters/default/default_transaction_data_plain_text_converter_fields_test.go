package _default

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

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseInvalidTime(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount\n"+
		"2024-09-01T12:34:56,Expense,Test Category,Test Account,123.45,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount\n"+
		"09/01/2024 12:34:56,Expense,Test Category,Test Account,123.45,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseInvalidType(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount\n"+
		"2024-09-01 12:34:56,Type,Test Category,Test Account,123.45,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTypeInvalid.Message)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseValidTimezone(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Timezone,Type,Sub Category,Account,Amount,Account2,Account2 Amount\n"+
		"2024-09-01 12:34:56,-10:00,Expense,Test Category,Test Account,123.45,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725230096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Timezone,Type,Sub Category,Account,Amount,Account2,Account2 Amount\n"+
		"2024-09-01 12:34:56,+00:00,Expense,Test Category,Test Account,123.45,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Timezone,Type,Sub Category,Account,Amount,Account2,Account2 Amount\n"+
		"2024-09-01 12:34:56,+12:45,Expense,Test Category,Test Account,123.45,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725148196), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseInvalidTimezone(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Timezone,Type,Sub Category,Account,Amount,Account2,Account2 Amount\n"+
		"2024-09-01 12:34:56,Asia/Shanghai,Expense,Test Category,Test Account,123.45,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeZoneInvalid.Message)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseValidAccountCurrency(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Account Currency,Amount,Account2,Account2 Currency,Account2 Amount\n"+
		"2024-09-01 01:23:45,Balance Modification,,Test Account,USD,123.45,,,\n"+
		"2024-09-01 12:34:56,Transfer,Test Category2,Test Account,USD,1.23,Test Account2,EUR,1.10"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

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

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseInvalidAccountCurrency(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Account Currency,Amount,Account2,Account2 Currency,Account2 Amount\n"+
		"2024-09-01 01:23:45,Balance Modification,,Test Account,USD,123.45,,,\n"+
		"2024-09-01 12:34:56,Transfer,Test Category3,Test Account,CNY,1.23,Test Account2,EUR,1.10"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Account Currency,Amount,Account2,Account2 Currency,Account2 Amount\n"+
		"2024-09-01 01:23:45,Balance Modification,,Test Account,USD,123.45,,,\n"+
		"2024-09-01 12:34:56,Transfer,Test Category3,Test Account2,CNY,1.23,Test Account,EUR,1.10"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseNotSupportedCurrency(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Account Currency,Amount,Account2,Account2 Currency,Account2 Amount\n"+
		"2024-09-01 01:23:45,Balance Modification,,Test Account,XXX,123.45,,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Account Currency,Amount,Account2,Account2 Currency,Account2 Amount\n"+
		"2024-09-01 01:23:45,Transfer,Test Category,Test Account,USD,123.45,Test Account2,XXX,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseInvalidAmount(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount\n"+
		"2024-09-01 12:34:56,Expense,Test Category,Test Account,123 45,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount\n"+
		"2024-09-01 12:34:56,Transfer,Test Category,Test Account,123.45,Test Account2,123 45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseNoAmount2(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2\n"+
		"2024-09-01 12:34:56,Expense,Test Category,Test Account,123.45,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
	assert.Equal(t, int64(0), allNewTransactions[0].RelatedAccountAmount)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2\n"+
		"2024-09-01 12:34:56,Transfer,Test Category,Test Account,123.45,Test Account2"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
	assert.Equal(t, int64(12345), allNewTransactions[0].RelatedAccountAmount)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseValidGeographicLocation(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount,Geographic Location\n"+
		"2024-09-01 12:34:56,Expense,Test Category,Test Account,123.45,,,123.45 45.56"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, 123.45, allNewTransactions[0].GeoLongitude)
	assert.Equal(t, 45.56, allNewTransactions[0].GeoLatitude)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseInvalidGeographicLocation(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount,Geographic Location\n"+
		"2024-09-01 12:34:56,Expense,Test Category,Test Account,123.45,,,1"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, float64(0), allNewTransactions[0].GeoLongitude)
	assert.Equal(t, float64(0), allNewTransactions[0].GeoLatitude)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount,Geographic Location\n"+
		"2024-09-01 12:34:56,Expense,Test Category,Test Account,123.45,,,a b"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrGeographicLocationInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount,Geographic Location\n"+
		"2024-09-01 12:34:56,Expense,Test Category,Test Account,123.45,,,1 "), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrGeographicLocationInvalid.Message)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseTag(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, allNewTags, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount,Tags\n"+
		"2024-09-01 00:00:00,Balance Modification,,Test Account,123.45,,,foo;;bar.;#test;hello\tworld;;"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 4, len(allNewTags))

	assert.Equal(t, int64(1234567890), allNewTags[0].Uid)
	assert.Equal(t, "foo", allNewTags[0].Name)

	assert.Equal(t, int64(1234567890), allNewTags[1].Uid)
	assert.Equal(t, "bar.", allNewTags[1].Name)

	assert.Equal(t, int64(1234567890), allNewTags[2].Uid)
	assert.Equal(t, "#test", allNewTags[2].Name)

	assert.Equal(t, int64(1234567890), allNewTags[3].Uid)
	assert.Equal(t, "hello\tworld", allNewTags[3].Name)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_ParseDescription(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Time,Type,Sub Category,Account,Amount,Account2,Account2 Amount,Description\n"+
		"2024-09-01 12:34:56,Expense,Test Category,Test Account,123.45,,,foo    bar\t#test"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "foo    bar\t#test", allNewTransactions[0].Comment)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_MissingFileHeader(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(""), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrNotFoundTransactionDataInFile.Message)
}

func TestDefaultTransactionDataCSVFileConverterParseImportedData_MissingRequiredColumn(t *testing.T) {
	importer := DefaultTransactionDataCSVFileConverter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing Time Column
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("Timezone,Type,Category,Sub Category,Account,Account Currency,Amount,Account2,Account2 Currency,Account2 Amount,Geographic Location,Tags,Description\n"+
		"+08:00,Balance Modification,,Test Sub Category,Test Account,CNY,123.45,,,,,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Type Column
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Category,Sub Category,Account,Account Currency,Amount,Account2,Account2 Currency,Account2 Amount,Geographic Location,Tags,Description\n"+
		"2024-09-01 00:00:00,+08:00,Test Category,Test Sub Category,Test Account,CNY,123.45,,,,,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Sub Category Column
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Type,Account,Account Currency,Amount,Account2,Account2 Currency,Account2 Amount,Geographic Location,Tags,Description\n"+
		"2024-09-01 00:00:00,+08:00,Balance Modification,Test Account,CNY,123.45,,,,,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Account Name Column
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Timezone,Type,Category,Sub Category,Account Currency,Amount,Account2,Account2 Currency,Account2 Amount,Geographic Location,Tags,Description\n"+
		"2024-09-01 00:00:00,+08:00,Balance Modification,,Test Sub Category,CNY,123.45,,,,,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Amount Column
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Timezone,Type,Category,Sub Category,Account,Account Currency,Account2,Account2 Currency,Account2 Amount,Geographic Location,Tags,Description\n"+
		"2024-09-01 00:00:00,+08:00,Balance Modification,,Test Sub Category,Test Account,CNY,,,,,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Account2 Name Column
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte("Time,Timezone,Type,Category,Sub Category,Account,Account Currency,Amount,Account2 Currency,Account2 Amount,Geographic Location,Tags,Description\n"+
		"2024-09-01 00:00:00,+08:00,Balance Modification,,Test Sub Category,Test Account,CNY,123.45,,,,,"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)
}
