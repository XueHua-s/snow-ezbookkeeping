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

func TestCustomTransactionDataDsvFileImporter_ParseInvalidTime(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01T12:34:56,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"09/01/2024 12:34:56,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseTransactionWithoutType(t *testing.T) {
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
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,A,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrNotFoundTransactionDataInFile.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseInvalidType(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"B": 0,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,B,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTypeInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseTimeWithTimezone(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ssZ", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56-10:00,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725230096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56+00:00,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56+12:45,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725148196), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
}

func TestCustomTransactionDataDsvFileImporter_ParseTimeWithTimezone2(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ssZZ", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56-1000,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725230096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56+0000,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56+1245,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725148196), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
}

func TestCustomTransactionDataDsvFileImporter_ParseValidTimezone(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:     0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE: 1,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:     2,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:               3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,-10:00,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725230096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,+00:00,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,+12:45,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725148196), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
}

func TestCustomTransactionDataDsvFileImporter_ParseValidTimezone2(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:     0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE: 1,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:     2,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:               3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "ZZ", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,-1000,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725230096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,+0000,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,+1245,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725148196), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
}

func TestCustomTransactionDataDsvFileImporter_ParseInvalidTimezoneFormat(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:     0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE: 1,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:     2,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:               3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "z", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,CST,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrImportFileTransactionTimezoneFormatInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseInvalidTimezone(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:     0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE: 1,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:     2,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:               3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,Asia/Shanghai,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeZoneInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,-0700,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeZoneInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseInvalidTimezone2(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:     0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIMEZONE: 1,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:     2,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:               3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "ZZ", ".", "", "", "", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,Asia/Shanghai,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeZoneInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,0700,E,123.45"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeZoneInvalid.Message)
}
