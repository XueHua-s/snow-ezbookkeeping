package iif

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIIFTransactionDataFileParseImportedData_InvalidDataLines(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	//Missing Transaction Line
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"SPL\t09/01/2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Missing Transaction And Split Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Missing Split Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Missing Transaction End Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-123.45\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Missing Transaction End Line (following is another header)
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-123.45\n"+
			"!ACCNT\tNAME\tACCNTTYPE\n"+
			"ACCNT\tTest Account\tBANK\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Invalid Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-123.45\n"+
			"TEST\t\t\t\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Repeat Transaction Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Repeat Transaction End Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\t\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\t\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)
}

func TestIIFTransactionDataFileParseImportedData_InvalidHeaderLines(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing All Sample Lines
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Missing Transaction Sample Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Missing Split Sample Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Missing Transaction End Sample Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Missing Transaction End Sample Line (following is data line)
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)

	// Invalid Sample Line
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!TEST\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidIIFFile.Message)
}

func TestIIFTransactionDataFileParseImportedData_MissingRequiredColumn(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing Date Column
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tACCNT\tAMOUNT\t\n"+
			"!SPL\tACCNT\tAMOUNT\t\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\tTest Account\t123.45\n"+
			"SPL\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Account Column
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tAMOUNT\t\n"+
			"!SPL\tDATE\tAMOUNT\t\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\t123.45\n"+
			"SPL\t09/01/2024\t-123.45\n"+
			"ENDTRNS\t\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Amount Column
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\t\n"+
			"!SPL\tDATE\tACCNT\t\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\n"+
			"SPL\t09/01/2024\tTest Account2\n"+
			"ENDTRNS\t\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)
}
