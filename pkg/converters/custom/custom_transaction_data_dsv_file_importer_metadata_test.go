package custom

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCustomTransactionDataDsvFileImporter_ParseValidGeographicLocation(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:    0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:    1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:              2,
		datatable.TRANSACTION_DATA_TABLE_GEOGRAPHIC_LOCATION: 3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", ";", "lonlat", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,E,123.45,123.45;45.56"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, 123.45, allNewTransactions[0].GeoLongitude)
	assert.Equal(t, 45.56, allNewTransactions[0].GeoLatitude)
}

func TestCustomTransactionDataDsvFileImporter_ParseInvalidGeographicLocation(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:    0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:    1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:              2,
		datatable.TRANSACTION_DATA_TABLE_GEOGRAPHIC_LOCATION: 3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", " ", "lonlat", "")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,E,123.45,,,1"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, float64(0), allNewTransactions[0].GeoLongitude)
	assert.Equal(t, float64(0), allNewTransactions[0].GeoLatitude)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"2024-09-01 12:34:56,E,123.45,a b"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrGeographicLocationInvalid.Message)
}

func TestCustomTransactionDataDsvFileImporter_ParseTag(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
		datatable.TRANSACTION_DATA_TABLE_TAGS:             3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
		"E": models.TRANSACTION_TYPE_EXPENSE,
	}
	importer, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", ";")
	assert.Nil(t, err)

	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, allNewTags, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 00:00:00,E,123.45,foo;;bar.;#test;hello\tworld;;"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

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

func TestCustomTransactionDataDsvFileImporter_ParseTagWithoutSeparator(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
		datatable.TRANSACTION_DATA_TABLE_TAGS:             3,
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

	_, _, _, _, _, allNewTags, err := importer.ParseImportedData(context, user, []byte(
		"2024-09-01 00:00:00,E,123.45,foo;;bar.;#test;hello\tworld;;"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 1, len(allNewTags))

	assert.Equal(t, int64(1234567890), allNewTags[0].Uid)
	assert.Equal(t, "foo;;bar.;#test;hello	world;;", allNewTags[0].Name)
}

func TestCustomTransactionDataDsvFileImporter_ParseDescription(t *testing.T) {
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
		datatable.TRANSACTION_DATA_TABLE_DESCRIPTION:      3,
	}
	transactionTypeMapping := map[string]models.TransactionType{
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
		"2024-09-01 12:34:56,T,123.45,foo    bar\t#test"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "foo    bar\t#test", allNewTransactions[0].Comment)
}

func TestCustomTransactionDataDsvFileImporter_InvalidSeparator(t *testing.T) {
	transactionTypeMapping := map[string]models.TransactionType{
		"B": models.TRANSACTION_TYPE_MODIFY_BALANCE,
	}
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	_, err := CreateNewCustomTransactionDataDsvFileImporter("test", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.EqualError(t, err, errs.ErrImportFileTypeNotSupported.Message)
}

func TestCustomTransactionDataDsvFileImporter_InvalidFileEncoding(t *testing.T) {
	transactionTypeMapping := map[string]models.TransactionType{
		"B": models.TRANSACTION_TYPE_MODIFY_BALANCE,
	}
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           2,
	}
	_, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "ascii", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.EqualError(t, err, errs.ErrImportFileEncodingNotSupported.Message)
}

func TestCustomTransactionDataDsvFileImporter_MissingRequiredColumn(t *testing.T) {
	transactionTypeMapping := map[string]models.TransactionType{
		"B": models.TRANSACTION_TYPE_MODIFY_BALANCE,
	}

	// Missing Time Column
	columnIndexMapping := map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 0,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           1,
	}
	_, err := CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Type Column
	columnIndexMapping = map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT:           1,
	}
	_, err = CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Amount Column
	columnIndexMapping = map[datatable.TransactionDataTableColumn]int{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME: 0,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE: 1,
	}
	_, err = CreateNewCustomTransactionDataDsvFileImporter("custom_csv", "utf-8", columnIndexMapping, transactionTypeMapping, false, "YYYY-MM-DD HH:mm:ss", "", ".", "", "", "", "")
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)
}
